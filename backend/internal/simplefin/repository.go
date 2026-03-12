package simplefin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/models"
)

type LinkedAccount struct {
	AccountID   string
	SimplefinID string
	AccountName string
	Institution string
	AccountType string
	Currency    string
}

type MortgageFields struct {
	InterestRate   *float64
	LoanTermMonths *int
	PurchasePrice  *float64
	PurchaseDate   *string
	DownPaymentPct *float64
}

type SQLiteSimplefinRepository struct {
	db *sql.DB
}

func NewSimplefinRepository(db *sql.DB) *SQLiteSimplefinRepository {
	return &SQLiteSimplefinRepository{db: db}
}

func (r *SQLiteSimplefinRepository) SaveConnection(ctx context.Context, userID, encryptedAccessURL string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO simplefin_connections (id, user_id, access_url)
		 VALUES (?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET access_url=excluded.access_url, sync_error=NULL, updated_at=datetime('now')`,
		uuid.New().String(), userID, encryptedAccessURL,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetConnection(ctx context.Context, userID string) (*models.SimplefinConnection, error) {
	var conn models.SimplefinConnection
	err := r.db.QueryRowContext(ctx, //nolint:gosec // parameterized query
		`SELECT id, user_id, access_url, last_sync_at, sync_error, created_at, updated_at
		 FROM simplefin_connections WHERE user_id = ?`,
		userID,
	).Scan(&conn.ID, &conn.UserID, &conn.AccessURL, &conn.LastSyncAt, &conn.SyncError, &conn.CreatedAt, &conn.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *SQLiteSimplefinRepository) GetAllConnections(ctx context.Context) ([]models.SimplefinConnection, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, access_url, last_sync_at, sync_error, created_at, updated_at
		 FROM simplefin_connections`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var conns []models.SimplefinConnection
	for rows.Next() {
		var conn models.SimplefinConnection
		if err := rows.Scan(&conn.ID, &conn.UserID, &conn.AccessURL, &conn.LastSyncAt, &conn.SyncError, &conn.CreatedAt, &conn.UpdatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	return conns, rows.Err()
}

func (r *SQLiteSimplefinRepository) DeleteConnection(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query, not injectable
		`UPDATE accounts SET simplefin_id = NULL, updated_at = datetime('now')
		 WHERE user_id = ? AND simplefin_id IS NOT NULL`,
		userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query, not injectable
		`DELETE FROM simplefin_connections WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteSimplefinRepository) LinkAccount(ctx context.Context, userID, accountID, simplefinID string) error {
	result, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query, not injectable
		`UPDATE accounts SET simplefin_id = ?, updated_at = datetime('now')
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		simplefinID, accountID, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SQLiteSimplefinRepository) UnlinkAccount(ctx context.Context, userID, accountID string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE accounts SET simplefin_id = NULL, updated_at = datetime('now')
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		accountID, userID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetLinkedAccounts(ctx context.Context, userID string) ([]LinkedAccount, error) {
	rows, err := r.db.QueryContext(ctx, //nolint:gosec // parameterized query
		`SELECT id, simplefin_id, name, institution, account_type, currency
		 FROM accounts
		 WHERE user_id = ? AND simplefin_id IS NOT NULL AND deleted_at IS NULL
		 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var linked []LinkedAccount
	for rows.Next() {
		var la LinkedAccount
		if err := rows.Scan(&la.AccountID, &la.SimplefinID, &la.AccountName, &la.Institution, &la.AccountType, &la.Currency); err != nil {
			return nil, err
		}
		linked = append(linked, la)
	}
	return linked, rows.Err()
}

func (r *SQLiteSimplefinRepository) UpdateSyncSuccess(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE simplefin_connections SET last_sync_at = datetime('now'), sync_error = NULL, updated_at = datetime('now')
		 WHERE user_id = ?`,
		userID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) UpdateSyncError(ctx context.Context, userID, syncError string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE simplefin_connections SET sync_error = ?, updated_at = datetime('now')
		 WHERE user_id = ?`,
		syncError, userID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetAssetForAccount(ctx context.Context, userID, accountID string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_id, name, asset_type, current_value, currency, metadata, created_at, updated_at
		 FROM assets
		 WHERE user_id = ? AND account_id = ? AND deleted_at IS NULL`,
		userID, accountID,
	).Scan(&asset.ID, &asset.UserID, &asset.AccountID, &asset.Name, &asset.AssetType, &asset.CurrentValue, &asset.Currency, &asset.Metadata, &asset.CreatedAt, &asset.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *SQLiteSimplefinRepository) UpdateAccountBalance(ctx context.Context, accountID string, balance float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE accounts SET original_balance = ? WHERE id = ? AND original_balance IS NULL`,
		balance, accountID,
	)
	if err != nil {
		return err
	}

	if err := r.backfillInitialBalance(ctx, tx, accountID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE accounts SET current_balance = ?, updated_at = datetime('now') WHERE id = ?`,
		balance, accountID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO account_balance_history (id, account_id, balance, recorded_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		uuid.New().String(), accountID, balance,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteSimplefinRepository) backfillInitialBalance(ctx context.Context, tx *sql.Tx, accountID string) error {
	var exists bool
	err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM account_balance_history WHERE account_id = ?)`,
		accountID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	var originalBalance sql.NullFloat64
	err = tx.QueryRowContext(ctx,
		`SELECT original_balance FROM accounts WHERE id = ?`,
		accountID,
	).Scan(&originalBalance)
	if err != nil {
		return err
	}

	if !originalBalance.Valid {
		return nil
	}

	var startDate sql.NullString
	err = tx.QueryRowContext(ctx,
		`SELECT MIN(ah.recorded_at) FROM asset_history ah
		 JOIN assets a ON a.id = ah.asset_id
		 WHERE a.account_id = ?`,
		accountID,
	).Scan(&startDate)
	if err != nil {
		return err
	}

	recordedAt := startDate.String
	if !startDate.Valid || recordedAt == "" {
		err = tx.QueryRowContext(ctx,
			`SELECT created_at FROM accounts WHERE id = ?`,
			accountID,
		).Scan(&recordedAt)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO account_balance_history (id, account_id, balance, recorded_at)
		 VALUES (?, ?, ?, ?)`,
		uuid.New().String(), accountID, originalBalance.Float64, recordedAt,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetDebtGoalForAccount(ctx context.Context, accountID string) (*models.Goal, error) {
	var goal models.Goal
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, goal_type, target_amount, current_amount, deadline, linked_account_id, priority_rank, created_at, updated_at
		 FROM goals
		 WHERE linked_account_id = ? AND goal_type = 'debt_payoff' AND deleted_at IS NULL`,
		accountID,
	).Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.GoalType, &goal.TargetAmount, &goal.CurrentAmount, &goal.Deadline, &goal.LinkedAccountID, &goal.PriorityRank, &goal.CreatedAt, &goal.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *SQLiteSimplefinRepository) UpdateDebtBalance(ctx context.Context, goalID string, newBalance float64) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE goals SET current_amount = ?, updated_at = datetime('now') WHERE id = ?`,
		newBalance, goalID,
	)
	return err
}

func accountTypeToAssetType(accountType string) string {
	switch accountType {
	case "checking", "savings", "hysa", "venmo", "hsa":
		return "liquid"
	case "brokerage":
		return "brokerage"
	case "retirement_401k", "retirement_ira", "retirement_roth_ira", "rollover_ira":
		return "retirement"
	case "crypto_wallet":
		return "speculative"
	default:
		return "other"
	}
}

func (r *SQLiteSimplefinRepository) CreateAccountWithLink(ctx context.Context, userID, name, institution, accountType, currency, simplefinID string, balance float64, mortgage *MortgageFields) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	accountID := uuid.New().String()

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active, simplefin_id,
		 interest_rate, loan_term_months, purchase_price, purchase_date, down_payment_pct)
		 VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?)`,
		accountID, userID, name, institution, accountType, currency, simplefinID,
		mortgage.InterestRate, mortgage.LoanTermMonths, mortgage.PurchasePrice, mortgage.PurchaseDate, mortgage.DownPaymentPct,
	)
	if err != nil {
		return "", err
	}

	if isDebtAccount(accountType) {
		absBalance := balance
		if absBalance < 0 {
			absBalance = -absBalance
		}
		_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
			`UPDATE accounts SET original_balance = ?, current_balance = ? WHERE id = ?`,
			absBalance, absBalance, accountID,
		)
		if err != nil {
			return "", err
		}

		goalID := uuid.New().String()
		_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
			`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, linked_account_id, priority_rank)
			 VALUES (?, ?, ?, 'debt_payoff', ?, ?, ?, (SELECT COALESCE(MAX(priority_rank), 0) + 1 FROM goals WHERE user_id = ?))`,
			goalID, userID, "Pay off "+name, absBalance, absBalance, accountID, userID,
		)
		if err != nil {
			return "", err
		}

		_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
			`INSERT INTO account_balance_history (id, account_id, balance, recorded_at)
			 VALUES (?, ?, ?, datetime('now'))`,
			uuid.New().String(), accountID, absBalance,
		)
		if err != nil {
			return "", err
		}

		if accountType == "mortgage" {
			if err := r.createMortgageAsset(ctx, tx, userID, accountID, name, currency, mortgage); err != nil {
				return "", err
			}
		}
	} else {
		assetType := accountTypeToAssetType(accountType)
		assetID := uuid.New().String()
		absBalance := balance
		if absBalance < 0 {
			absBalance = -absBalance
		}

		_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
			`INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value, currency)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			assetID, userID, accountID, name, assetType, absBalance, currency,
		)
		if err != nil {
			return "", err
		}

		_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
			`INSERT INTO asset_history (id, asset_id, value, recorded_at)
			 VALUES (?, ?, ?, datetime('now'))`,
			uuid.New().String(), assetID, absBalance,
		)
		if err != nil {
			return "", err
		}
	}

	return accountID, tx.Commit()
}

func (r *SQLiteSimplefinRepository) createMortgageAsset(ctx context.Context, tx *sql.Tx, userID, accountID, name, currency string, mortgage *MortgageFields) error {
	if mortgage == nil || mortgage.PurchasePrice == nil {
		return nil
	}

	assetID := uuid.New().String()
	_, err := tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value, currency)
		 VALUES (?, ?, ?, ?, 'real_estate', ?, ?)`,
		assetID, userID, accountID, name+" (Home)", *mortgage.PurchasePrice, currency,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO asset_history (id, asset_id, value, recorded_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		uuid.New().String(), assetID, *mortgage.PurchasePrice,
	)
	return err
}

func (r *SQLiteSimplefinRepository) BulkCreateSyncedTransactions(ctx context.Context, txns []models.Transaction) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	imported := 0
	for _, t := range txns {
		result, err := tx.ExecContext(ctx, //nolint:gosec // parameterized query
			`INSERT OR IGNORE INTO transactions (id, user_id, account_id, type, category, amount, currency, date, notes, is_recurring, external_id)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`,
			t.ID, t.UserID, t.AccountID, t.Type, t.Category, t.Amount, t.Currency, t.Date.Format("2006-01-02"), t.Notes, t.ExternalID,
		)
		if err != nil {
			return imported, err
		}
		rows, _ := result.RowsAffected()
		imported += int(rows)
	}

	return imported, tx.Commit()
}

func (r *SQLiteSimplefinRepository) DismissAccount(ctx context.Context, userID, simplefinID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO dismissed_simplefin_accounts (id, user_id, simplefin_id) VALUES (?, ?, ?)`,
		uuid.New().String(), userID, simplefinID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) UndismissAccount(ctx context.Context, userID, simplefinID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM dismissed_simplefin_accounts WHERE user_id = ? AND simplefin_id = ?`,
		userID, simplefinID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetDismissedAccountIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT simplefin_id FROM dismissed_simplefin_accounts WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *SQLiteSimplefinRepository) UpdateAssetValue(ctx context.Context, assetID string, newValue float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE assets SET current_value = ?, updated_at = datetime('now') WHERE id = ?`,
		newValue, assetID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO asset_history (id, asset_id, value, recorded_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		uuid.New().String(), assetID, newValue,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
