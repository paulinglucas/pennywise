package queries

import (
	"context"
	"database/sql"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/observability"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

type NetWorthResult struct {
	AssetTotal float64
	CashTotal  float64
	DebtTotal  float64
}

type SpendingRow struct {
	Category string
	Amount   float64
}

type DebtRow struct {
	AccountID       string
	Name            string
	Balance         float64
	MonthlyPayment  float64
	OriginalBalance *float64
	InterestRate    *float64
}

type NetWorthDataPoint struct {
	Date  openapi_types.Date
	Value float64
}

func (r *DashboardRepository) GetNetWorth(ctx context.Context, userID string) (NetWorthResult, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_net_worth", time.Since(start)) }()

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
		`SELECT COALESCE(SUM(COALESCE(a.current_balance, 0)), 0)
		 FROM accounts a
		 WHERE a.user_id = ? AND a.deleted_at IS NULL
		   AND a.account_type IN ('checking', 'savings', 'other')`,
		userID,
	).Scan(&result.CashTotal)
	if err != nil {
		return result, err
	}

	err = r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(
		   COALESCE(a.current_balance, a.original_balance, 0)
		 ), 0)
		 FROM accounts a
		 WHERE a.user_id = ? AND a.deleted_at IS NULL
		   AND a.account_type IN ('credit_card', 'mortgage', 'credit_line')`,
		userID,
	).Scan(&result.DebtTotal)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *DashboardRepository) GetCashFlowThisMonth(ctx context.Context, userID string, now time.Time) (float64, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_cash_flow", time.Since(start)) }()

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	var deposits, expenses float64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'deposit' AND category NOT IN ('transfer', 'cash') THEN amount ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN type = 'expense' AND category NOT IN ('transfer', 'cash') THEN amount ELSE 0 END), 0)
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

func (r *DashboardRepository) GetSpendingByCategory(ctx context.Context, userID string, since time.Time, until time.Time) ([]SpendingRow, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_spending_by_category", time.Since(start)) }()

	rows, err := r.db.QueryContext(ctx,
		`SELECT category, SUM(amount) as total
		 FROM transactions
		 WHERE user_id = ? AND type = 'expense' AND deleted_at IS NULL
		   AND category NOT IN ('transfer', 'cash')
		   AND date >= ? AND date < ?
		 GROUP BY category
		 ORDER BY total DESC`,
		userID, since.Format("2006-01-02"), until.Format("2006-01-02"),
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
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_debts_summary", time.Since(start)) }()

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.name,
		        COALESCE(g.current_amount, 0) as balance,
		        COALESCE((
		            SELECT SUM(t.amount) FROM transactions t
		            WHERE t.user_id = a.user_id AND t.account_id = a.id
		              AND t.category = 'transfer' AND t.deleted_at IS NULL
		              AND t.date >= ? AND t.date < ?
		        ), 0) as monthly_payment,
		        a.original_balance,
		        a.interest_rate
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
		if err := rows.Scan(&row.AccountID, &row.Name, &row.Balance, &row.MonthlyPayment, &row.OriginalBalance, &row.InterestRate); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *DashboardRepository) GetNetWorthHistory(ctx context.Context, userID string, since time.Time, includeSinceDate bool) ([]NetWorthDataPoint, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_net_worth_history", time.Since(start)) }()

	points, err := r.getHistoricalPoints(ctx, userID, since, includeSinceDate)
	if err != nil {
		return nil, err
	}

	todayPoint, err := r.getCurrentNetWorthPoint(ctx, userID)
	if err != nil {
		return nil, err
	}

	return appendTodayPoint(points, todayPoint), nil
}

func (r *DashboardRepository) getHistoricalPoints(ctx context.Context, userID string, since time.Time, includeSinceDate bool) ([]NetWorthDataPoint, error) {
	sinceStr := since.Format("2006-01-02")
	includeSince := 0
	if includeSinceDate {
		includeSince = 1
	}
	rows, err := r.db.QueryContext(ctx,
		`WITH asset_snap_dates AS (
		   SELECT DISTINCT DATE(ah.recorded_at) as snap_date
		   FROM asset_history ah
		   JOIN assets a ON a.id = ah.asset_id
		   WHERE a.user_id = ? AND a.deleted_at IS NULL
		     AND ah.recorded_at >= ?
		 ),
		 cash_snap_dates AS (
		   SELECT DISTINCT DATE(abh.recorded_at) as snap_date
		   FROM account_balance_history abh
		   JOIN accounts a ON a.id = abh.account_id
		   WHERE a.user_id = ? AND a.deleted_at IS NULL
		     AND a.account_type IN ('checking', 'savings', 'other')
		     AND abh.recorded_at >= ?
		 ),
		 debt_snap_dates AS (
		   SELECT DISTINCT DATE(abh.recorded_at) as snap_date
		   FROM account_balance_history abh
		   JOIN accounts a ON a.id = abh.account_id
		   WHERE a.user_id = ? AND a.deleted_at IS NULL
		     AND a.account_type IN ('credit_card', 'mortgage', 'credit_line')
		     AND abh.recorded_at >= ?
		 ),
		 since_date AS (
		   SELECT ? as snap_date WHERE ? = 1
		 ),
		 all_dates AS (
		   SELECT snap_date FROM since_date
		   UNION
		   SELECT snap_date FROM asset_snap_dates
		   UNION
		   SELECT snap_date FROM cash_snap_dates
		   UNION
		   SELECT snap_date FROM debt_snap_dates
		 )
		 SELECT
		   ad.snap_date,
		   COALESCE((
		     SELECT SUM(latest_val) FROM (
		       SELECT (
		         SELECT ah.value FROM asset_history ah
		         WHERE ah.asset_id = a.id AND DATE(ah.recorded_at) <= ad.snap_date
		         ORDER BY ah.recorded_at DESC LIMIT 1
		       ) as latest_val
		       FROM assets a
		       WHERE a.user_id = ? AND a.deleted_at IS NULL
		     )
		   ), 0)
		     + COALESCE((
		         SELECT SUM(COALESCE(
		           (SELECT abh.balance FROM account_balance_history abh
		            WHERE abh.account_id = ca.id AND DATE(abh.recorded_at) <= ad.snap_date
		            ORDER BY abh.recorded_at DESC LIMIT 1),
		           CASE WHEN DATE(ca.created_at) <= ad.snap_date THEN COALESCE(ca.current_balance, 0) END,
		           0
		         ))
		         FROM accounts ca
		         WHERE ca.user_id = ? AND ca.deleted_at IS NULL
		           AND ca.account_type IN ('checking', 'savings', 'other')
		       ), 0)
		     - COALESCE((
		         SELECT SUM(COALESCE(
		           (SELECT abh.balance FROM account_balance_history abh
		            WHERE abh.account_id = a2.id AND DATE(abh.recorded_at) <= ad.snap_date
		            ORDER BY abh.recorded_at DESC LIMIT 1),
		           CASE WHEN DATE(a2.created_at) <= ad.snap_date THEN a2.original_balance END,
		           0
		         ))
		         FROM accounts a2
		         WHERE a2.user_id = ? AND a2.deleted_at IS NULL
		           AND a2.account_type IN ('credit_card', 'mortgage', 'credit_line')
		       ), 0)
		   as net_worth
		 FROM all_dates ad
		 ORDER BY ad.snap_date ASC`,
		userID, sinceStr, userID, sinceStr, userID, sinceStr, sinceStr, includeSince, userID, userID, userID,
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

func (r *DashboardRepository) getCurrentNetWorthPoint(ctx context.Context, userID string) (NetWorthDataPoint, error) {
	nw, err := r.GetNetWorth(ctx, userID)
	if err != nil {
		return NetWorthDataPoint{}, err
	}

	today := time.Now()
	return NetWorthDataPoint{
		Date:  openapi_types.Date{Time: time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)},
		Value: nw.AssetTotal + nw.CashTotal - nw.DebtTotal,
	}, nil
}

func appendTodayPoint(points []NetWorthDataPoint, today NetWorthDataPoint) []NetWorthDataPoint {
	todayStr := today.Date.Format("2006-01-02")
	if len(points) > 0 {
		lastDate := points[len(points)-1].Date.Format("2006-01-02")
		if lastDate == todayStr {
			points[len(points)-1].Value = today.Value
			return points
		}
	}
	return append(points, today)
}

func (r *DashboardRepository) PingDB(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
