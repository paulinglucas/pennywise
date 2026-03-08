package api

import (
	"context"
	"math"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

type DashboardRepository interface {
	GetNetWorth(ctx context.Context, userID string) (queries.NetWorthResult, error)
	GetCashFlowThisMonth(ctx context.Context, userID string, now time.Time) (float64, error)
	GetSpendingByCategory(ctx context.Context, userID string, since time.Time, until time.Time) ([]queries.SpendingRow, error)
	GetDebtsSummary(ctx context.Context, userID string, now time.Time) ([]queries.DebtRow, error)
	GetNetWorthHistory(ctx context.Context, userID string, since time.Time, includeSinceDate bool) ([]queries.NetWorthDataPoint, error)
	PingDB(ctx context.Context) error
}

func (h *AppHandler) GetDashboard(w http.ResponseWriter, r *http.Request, params GetDashboardParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())
	now := time.Now()

	nwResult, err := h.dashboard.GetNetWorth(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get net worth", requestID)
		return
	}

	cashFlow, err := h.dashboard.GetCashFlowThisMonth(r.Context(), userID, now)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get cash flow", requestID)
		return
	}

	since, until := spendingPeriodToRange(params.SpendingPeriod, now)
	spendingRows, err := h.dashboard.GetSpendingByCategory(r.Context(), userID, since, until)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get spending", requestID)
		return
	}

	debtRows, err := h.dashboard.GetDebtsSummary(r.Context(), userID, now)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get debts", requestID)
		return
	}

	netWorth := nwResult.AssetTotal + nwResult.CashTotal - nwResult.DebtTotal
	spending := buildSpendingResponse(spendingRows)
	debts := buildDebtsResponse(debtRows)

	WriteJSON(w, http.StatusOK, DashboardResponse{
		NetWorth:           float32(netWorth),
		CashFlowThisMonth:  float32(cashFlow),
		SpendingByCategory: spending,
		DebtsSummary:       debts,
	})
}

func spendingPeriodToRange(period *GetDashboardParamsSpendingPeriod, now time.Time) (time.Time, time.Time) {
	if period == nil {
		return now.AddDate(0, 0, -30), now
	}
	switch *period {
	case GetDashboardParamsSpendingPeriodN7d:
		return now.AddDate(0, 0, -7), now
	case GetDashboardParamsSpendingPeriodN90d:
		return now.AddDate(0, 0, -90), now
	case GetDashboardParamsSpendingPeriodN1y:
		return now.AddDate(-1, 0, 0), now
	default:
		return now.AddDate(0, 0, -30), now
	}
}

func buildSpendingResponse(rows []queries.SpendingRow) []struct {
	Amount     float32 `json:"amount"`
	Category   string  `json:"category"`
	Percentage float32 `json:"percentage"`
} {
	var totalSpending float64
	for _, row := range rows {
		totalSpending += row.Amount
	}

	result := make([]struct {
		Amount     float32 `json:"amount"`
		Category   string  `json:"category"`
		Percentage float32 `json:"percentage"`
	}, len(rows))

	for i, row := range rows {
		pct := 0.0
		if totalSpending > 0 {
			pct = (row.Amount / totalSpending) * 100
		}
		result[i] = struct {
			Amount     float32 `json:"amount"`
			Category   string  `json:"category"`
			Percentage float32 `json:"percentage"`
		}{
			Amount:     float32(row.Amount),
			Category:   row.Category,
			Percentage: float32(math.Round(pct*100) / 100),
		}
	}
	return result
}

func buildDebtsResponse(rows []queries.DebtRow) []struct {
	AccountId       openapi_types.UUID  `json:"account_id"`
	Balance         float32             `json:"balance"`
	MonthlyPayment  float32             `json:"monthly_payment"`
	MonthsRemaining *int                `json:"months_remaining,omitempty"`
	Name            string              `json:"name"`
	OriginalBalance *float32            `json:"original_balance,omitempty"`
	PayoffDate      *openapi_types.Date `json:"payoff_date,omitempty"`
} {
	result := make([]struct {
		AccountId       openapi_types.UUID  `json:"account_id"`
		Balance         float32             `json:"balance"`
		MonthlyPayment  float32             `json:"monthly_payment"`
		MonthsRemaining *int                `json:"months_remaining,omitempty"`
		Name            string              `json:"name"`
		OriginalBalance *float32            `json:"original_balance,omitempty"`
		PayoffDate      *openapi_types.Date `json:"payoff_date,omitempty"`
	}, len(rows))

	for i, row := range rows {
		result[i].AccountId = ParseID(row.AccountID)
		result[i].Balance = float32(row.Balance)
		result[i].MonthlyPayment = float32(row.MonthlyPayment)
		result[i].Name = row.Name

		if row.OriginalBalance != nil {
			ob := float32(*row.OriginalBalance)
			result[i].OriginalBalance = &ob
		}

		if row.MonthlyPayment > 0 && row.Balance > 0 {
			months := int(math.Ceil(row.Balance / row.MonthlyPayment))
			result[i].MonthsRemaining = &months
			payoffDate := openapi_types.Date{Time: time.Now().AddDate(0, months, 0)}
			result[i].PayoffDate = &payoffDate
		}
	}
	return result
}

func (h *AppHandler) GetNetWorthHistory(w http.ResponseWriter, r *http.Request, params GetNetWorthHistoryParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	since, includeSinceDate := periodToSince(params.Period)

	points, err := h.dashboard.GetNetWorthHistory(r.Context(), userID, since, includeSinceDate)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get net worth history", requestID)
		return
	}

	dataPoints := make([]struct {
		Date  openapi_types.Date `json:"date"`
		Value float32            `json:"value"`
	}, len(points))
	for i, p := range points {
		dataPoints[i].Date = p.Date
		dataPoints[i].Value = float32(p.Value)
	}

	WriteJSON(w, http.StatusOK, NetWorthHistoryResponse{DataPoints: dataPoints})
}

func periodToSince(period *GetNetWorthHistoryParamsPeriod) (time.Time, bool) {
	now := time.Now()
	if period == nil {
		return now.AddDate(-1, 0, 0), true
	}
	switch *period {
	case GetNetWorthHistoryParamsPeriodN1m:
		return now.AddDate(0, -1, 0), true
	case GetNetWorthHistoryParamsPeriodN5y:
		return now.AddDate(-5, 0, 0), true
	case GetNetWorthHistoryParamsPeriodAll:
		return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), false
	default:
		return now.AddDate(-1, 0, 0), true
	}
}
