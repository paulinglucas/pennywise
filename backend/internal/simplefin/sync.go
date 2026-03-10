package simplefin

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
)

type SyncResult struct {
	UserID  string
	Updated int
	Errors  int
	Error   error
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

func (s *SyncService) SyncUser(ctx context.Context, userID, accessURL string) (*SyncResult, error) {
	username, password, baseURL, err := ParseAccessURL(accessURL)
	if err != nil {
		return nil, fmt.Errorf("parse access URL: %w", err)
	}

	resp, err := s.client.FetchAccounts(ctx, username, password, baseURL)
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

		asset, err := s.repo.GetAssetForAccount(ctx, userID, la.AccountID)
		if err != nil {
			slog.Warn("failed to get asset for account", slog.String("account_id", la.AccountID), slog.Any("error", err))
			result.Errors++
			continue
		}
		if asset == nil {
			continue
		}

		if math.Abs(asset.CurrentValue-balance) < 0.005 {
			continue
		}

		if err := s.repo.UpdateAssetValue(ctx, asset.ID, balance); err != nil {
			slog.Warn("failed to update asset value", slog.String("asset_id", asset.ID), slog.Any("error", err)) //nolint:gosec // asset_id is internal DB value
			result.Errors++
			continue
		}

		result.Updated++
	}

	return result, nil
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

		result, err := s.SyncUser(ctx, conn.UserID, accessURL)
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
