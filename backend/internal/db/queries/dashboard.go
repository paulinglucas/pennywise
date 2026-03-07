package queries

import (
	"context"
	"database/sql"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

type NetWorthResult struct {
	AssetTotal float64
	DebtTotal  float64
}

type SpendingRow struct {
	Category string
	Amount   float64
}

type DebtRow struct {
	AccountID      string
	Name           string
	Balance        float64
	MonthlyPayment float64
}

type NetWorthDataPoint struct {
	Date  openapi_types.Date
	Value float64
}

func (r *DashboardRepository) GetNetWorth(ctx context.Context, userID string) (NetWorthResult, error) {
	var result NetWorthResult

	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(current_value), 0)
		 FROM assets WHERE user_id = ? AND deleted_at IS NULL`,
		userID,
	).Scan(&result.AssetTotal)
	if err != nil {
		return result, err
	}

	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(g.current_amount), 0)
		 FROM goals g
		 WHERE g.user_id = ? AND g.goal_type = 'debt_payoff' AND g.deleted_at IS NULL`,
		userID,
	).Scan(&result.DebtTotal)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *DashboardRepository) GetCashFlowThisMonth(ctx context.Context, userID string, now time.Time) (float64, error) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var deposits, expenses float64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'deposit' THEN amount ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		 FROM transactions
		 WHERE user_id = ? AND deleted_at IS NULL
		   AND date >= ? AND date < ?`,
		userID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"),
	).Scan(&deposits, &expenses)
	if err != nil {
		return 0, err
	}

	return deposits - expenses, nil
}

func (r *DashboardRepository) GetSpendingByCategory(ctx context.Context, userID string, now time.Time) ([]SpendingRow, error) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	rows, err := r.db.QueryContext(ctx,
		`SELECT category, SUM(amount) as total
		 FROM transactions
		 WHERE user_id = ? AND type = 'expense' AND deleted_at IS NULL
		   AND date >= ? AND date < ?
		 GROUP BY category
		 ORDER BY total DESC`,
		userID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []SpendingRow
	for rows.Next() {
		var row SpendingRow
		if err := rows.Scan(&row.Category, &row.Amount); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *DashboardRepository) GetDebtsSummary(ctx context.Context, userID string, now time.Time) ([]DebtRow, error) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.name,
		        COALESCE(g.current_amount, 0) as balance,
		        COALESCE((
		            SELECT SUM(t.amount) FROM transactions t
		            WHERE t.user_id = a.user_id AND t.account_id = a.id
		              AND t.type = 'expense' AND t.deleted_at IS NULL
		              AND t.date >= ? AND t.date < ?
		        ), 0) as monthly_payment
		 FROM accounts a
		 LEFT JOIN goals g ON g.linked_account_id = a.id AND g.goal_type = 'debt_payoff' AND g.deleted_at IS NULL
		 WHERE a.user_id = ? AND a.deleted_at IS NULL
		   AND a.account_type IN ('credit_card', 'mortgage', 'credit_line')
		 ORDER BY balance DESC`,
		monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"), userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []DebtRow
	for rows.Next() {
		var row DebtRow
		if err := rows.Scan(&row.AccountID, &row.Name, &row.Balance, &row.MonthlyPayment); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *DashboardRepository) GetNetWorthHistory(ctx context.Context, userID string, since time.Time) ([]NetWorthDataPoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DATE(ah.recorded_at) as snap_date, SUM(ah.value) as total_value
		 FROM asset_history ah
		 JOIN assets a ON a.id = ah.asset_id
		 WHERE a.user_id = ? AND a.deleted_at IS NULL
		   AND ah.recorded_at >= ?
		 GROUP BY snap_date
		 ORDER BY snap_date ASC`,
		userID, since.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []NetWorthDataPoint
	for rows.Next() {
		var dateStr string
		var value float64
		if err := rows.Scan(&dateStr, &value); err != nil {
			return nil, err
		}
		parsed, parseErr := time.Parse("2006-01-02", dateStr)
		if parseErr != nil {
			return nil, parseErr
		}
		result = append(result, NetWorthDataPoint{
			Date:  openapi_types.Date{Time: parsed},
			Value: value,
		})
	}
	return result, rows.Err()
}

func (r *DashboardRepository) PingDB(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
