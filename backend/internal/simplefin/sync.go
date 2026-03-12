package simplefin

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

	"github.com/google/uuid"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type SyncResult struct {
	UserID               string
	Updated              int
	Errors               int
	TransactionsImported int
	Error                error
}

type SyncService struct {
	client        *Client
	repo          *SQLiteSimplefinRepository
	encryptionKey []byte
}

func NewSyncService(client *Client, repo *SQLiteSimplefinRepository, encryptionKey []byte) *SyncService {
	return &SyncService{
		client:        client,
		repo:          repo,
		encryptionKey: encryptionKey,
	}
}

var debtAccountTypes = map[string]bool{
	"credit_card": true,
	"mortgage":    true,
	"credit_line": true,
}

func isDebtAccount(accountType string) bool {
	return debtAccountTypes[accountType]
}

func computeStartDate(lastSyncAt *time.Time) *int64 {
	var t time.Time
	if lastSyncAt != nil {
		t = lastSyncAt.Add(-24 * time.Hour)
	} else {
		t = time.Now().AddDate(0, 0, -90)
	}
	ts := t.Unix()
	return &ts
}

func (s *SyncService) SyncUser(ctx context.Context, userID, accessURL string, lastSyncAt *time.Time) (*SyncResult, error) {
	username, password, baseURL, err := ParseAccessURL(accessURL)
	if err != nil {
		return nil, fmt.Errorf("parse access URL: %w", err)
	}

	startDate := computeStartDate(lastSyncAt)
	resp, err := s.client.FetchAccounts(ctx, username, password, baseURL, startDate)
	if err != nil {
		return nil, fmt.Errorf("fetch SimpleFIN accounts: %w", err)
	}

	linked, err := s.repo.GetLinkedAccounts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get linked accounts: %w", err)
	}

	simplefinToAccount := make(map[string]LinkedAccount, len(linked))
	for _, la := range linked {
		simplefinToAccount[la.SimplefinID] = la
	}

	result := &SyncResult{UserID: userID}

	for _, sfinAccount := range resp.Accounts {
		la, ok := simplefinToAccount[sfinAccount.ID]
		if !ok {
			continue
		}

		balance, err := strconv.ParseFloat(sfinAccount.Balance, 64)
		if err != nil {
			slog.Warn("failed to parse balance", slog.String("account", sfinAccount.Name), slog.String("balance", sfinAccount.Balance)) //nolint:gosec // logged values are from trusted SimpleFIN API
			result.Errors++
			continue
		}
		balance = math.Abs(balance)

		if isDebtAccount(la.AccountType) {
			s.syncDebtBalance(ctx, la, balance, result)
		} else {
			s.syncAssetAccount(ctx, userID, la, balance, result)
		}

		s.syncTransactions(ctx, userID, la.AccountID, la.Currency, sfinAccount.Transactions, result)
	}

	return result, nil
}

func (s *SyncService) syncDebtBalance(ctx context.Context, la LinkedAccount, balance float64, result *SyncResult) {
	if err := s.repo.UpdateAccountBalance(ctx, la.AccountID, balance); err != nil {
		slog.Warn("failed to update account balance", slog.String("account_id", la.AccountID), slog.Any("error", err)) //nolint:gosec // account_id is internal DB value
		result.Errors++
		return
	}

	goal, err := s.repo.GetDebtGoalForAccount(ctx, la.AccountID)
	if err != nil {
		slog.Warn("failed to get debt goal", slog.String("account_id", la.AccountID), slog.Any("error", err)) //nolint:gosec // account_id is internal DB value
	}
	if goal != nil {
		if err := s.repo.UpdateDebtBalance(ctx, goal.ID, balance); err != nil {
			slog.Warn("failed to update debt goal balance", slog.String("goal_id", goal.ID), slog.Any("error", err)) //nolint:gosec // goal_id is internal DB value
		}
	}

	result.Updated++
}

func (s *SyncService) syncAssetAccount(ctx context.Context, userID string, la LinkedAccount, balance float64, result *SyncResult) {
	asset, err := s.repo.GetAssetForAccount(ctx, userID, la.AccountID)
	if err != nil {
		slog.Warn("failed to get asset for account", slog.String("account_id", la.AccountID), slog.Any("error", err)) //nolint:gosec // account_id is internal DB value
		result.Errors++
		return
	}
	if asset == nil {
		return
	}

	if math.Abs(asset.CurrentValue-balance) < 0.005 {
		return
	}

	if err := s.repo.UpdateAssetValue(ctx, asset.ID, balance); err != nil {
		slog.Warn("failed to update asset value", slog.String("asset_id", asset.ID), slog.Any("error", err)) //nolint:gosec // asset_id is internal DB value
		result.Errors++
		return
	}

	result.Updated++
}

func (s *SyncService) syncTransactions(ctx context.Context, userID, accountID, currency string, sfinTxns []Transaction, result *SyncResult) {
	if len(sfinTxns) == 0 {
		return
	}

	var txns []models.Transaction
	for _, st := range sfinTxns {
		if st.Pending {
			continue
		}

		txn := mapSimplefinTransaction(st, userID, accountID, currency)
		txns = append(txns, txn)
	}

	if len(txns) == 0 {
		return
	}

	imported, err := s.repo.BulkCreateSyncedTransactions(ctx, txns)
	if err != nil {
		slog.Warn("failed to import transactions", slog.String("account_id", accountID), slog.Any("error", err)) //nolint:gosec // account_id is internal DB value
		result.Errors++
		return
	}
	result.TransactionsImported += imported
}

func mapSimplefinTransaction(st Transaction, userID, accountID, currency string) models.Transaction {
	amount, _ := strconv.ParseFloat(st.Amount, 64)

	txnType := "deposit"
	if amount < 0 {
		txnType = "expense"
	}

	description := st.Description
	if description == "" {
		description = st.Payee
	}
	var notes *string
	if description != "" {
		notes = &description
	}

	externalID := st.ID

	return models.Transaction{
		ID:         uuid.New().String(),
		UserID:     userID,
		AccountID:  accountID,
		Type:       txnType,
		Category:   categorizeTransaction(description),
		Amount:     math.Abs(amount),
		Currency:   currency,
		Date:       time.Unix(st.Posted, 0),
		Notes:      notes,
		ExternalID: &externalID,
	}
}

func (s *SyncService) SyncAll(ctx context.Context) []SyncResult {
	connections, err := s.repo.GetAllConnections(ctx)
	if err != nil {
		slog.Error("failed to list SimpleFIN connections", slog.Any("error", err))
		return nil
	}

	var results []SyncResult
	for _, conn := range connections {
		accessURL, err := pennywisecrypto.Decrypt(s.encryptionKey, conn.AccessURL)
		if err != nil {
			slog.Error("failed to decrypt access URL", slog.String("user_id", conn.UserID), slog.Any("error", err))
			syncErr := "failed to decrypt access URL"
			_ = s.repo.UpdateSyncError(ctx, conn.UserID, syncErr)
			results = append(results, SyncResult{UserID: conn.UserID, Error: err})
			continue
		}

		result, err := s.SyncUser(ctx, conn.UserID, accessURL, conn.LastSyncAt)
		if err != nil {
			slog.Error("sync failed", slog.String("user_id", conn.UserID), slog.Any("error", err))
			_ = s.repo.UpdateSyncError(ctx, conn.UserID, err.Error())
			results = append(results, SyncResult{UserID: conn.UserID, Error: err})
			continue
		}

		_ = s.repo.UpdateSyncSuccess(ctx, conn.UserID)
		results = append(results, *result)
	}

	return results
}
