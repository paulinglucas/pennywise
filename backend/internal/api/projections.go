package api

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

type scenarioConfig struct {
	name       Scenario
	returnRate float64
}

var defaultScenarios = []scenarioConfig{
	{name: Best, returnRate: 0.10},
	{name: Average, returnRate: 0.07},
	{name: Worst, returnRate: 0.04},
}

type projectionState struct {
	investments float64
	cash        float64
	debts       []debtState
}

type debtState struct {
	balance      float64
	monthlyRate  float64
	minPayment   float64
}

func (h *AppHandler) ComputeProjection(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req ProjectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	nwResult, err := h.dashboard.GetNetWorth(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get net worth", requestID)
		return
	}

	debtRows, err := h.dashboard.GetDebtsSummary(r.Context(), userID, time.Now())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get debts", requestID)
		return
	}

	monthlySavings := h.estimateMonthlySavings(r, userID, req)

	state := buildProjectionState(nwResult, debtRows)
	scenarios := buildScenarios(req, state, monthlySavings)

	WriteJSON(w, http.StatusOK, ProjectionResponse{Scenarios: scenarios})
}

func buildProjectionState(nw queries.NetWorthResult, debtRows []queries.DebtRow) projectionState {
	debts := make([]debtState, 0, len(debtRows))
	for _, dr := range debtRows {
		if dr.Balance <= 0 {
			continue
		}
		annualRate := 0.0
		if dr.InterestRate != nil {
			annualRate = *dr.InterestRate / 100
		}
		debts = append(debts, debtState{
			balance:     dr.Balance,
			monthlyRate: annualRate / 12,
			minPayment:  dr.MonthlyPayment,
		})
	}

	sort.Slice(debts, func(i, j int) bool {
		return debts[i].monthlyRate > debts[j].monthlyRate
	})

	return projectionState{
		investments: nw.AssetTotal,
		cash:        nw.CashTotal,
		debts:       debts,
	}
}

func (h *AppHandler) estimateMonthlySavings(r *http.Request, userID string, req ProjectionRequest) float64 {
	cashFlow, err := h.dashboard.GetCashFlowThisMonth(r.Context(), userID, time.Now())
	if err != nil {
		cashFlow = 0
	}

	baseSavings := cashFlow
	if baseSavings < 0 {
		baseSavings = 0
	}

	if req.MonthlySavingsAdjustment != nil {
		adjustment := float64(*req.MonthlySavingsAdjustment) / 100
		baseSavings = baseSavings * (1 + adjustment)
	}

	return baseSavings
}

type projectionScenario = struct {
	DataPoints []struct {
		Date  openapi_types.Date `json:"date"`
		Value float32            `json:"value"`
	} `json:"data_points"`
	DebtFreeDate    *openapi_types.Date `json:"debt_free_date,omitempty"`
	MillionaireDate *openapi_types.Date `json:"millionaire_date,omitempty"`
	Scenario        Scenario            `json:"scenario"`
}

func buildScenarios(req ProjectionRequest, state projectionState, monthlySavings float64) []projectionScenario {
	configs := resolveScenarioConfigs(req)
	result := make([]projectionScenario, len(configs))

	extraDebt := 0.0
	if req.ExtraDebtPayment != nil {
		extraDebt = float64(*req.ExtraDebtPayment)
	}

	now := time.Now()

	for i, cfg := range configs {
		stateCopy := copyState(state)
		points, millDate, debtDate := projectScenario(now, &stateCopy, monthlySavings, cfg.returnRate, extraDebt, req)
		result[i].Scenario = cfg.name
		result[i].DataPoints = points
		result[i].MillionaireDate = millDate
		result[i].DebtFreeDate = debtDate
	}

	return result
}

func copyState(s projectionState) projectionState {
	debts := make([]debtState, len(s.debts))
	copy(debts, s.debts)
	return projectionState{
		investments: s.investments,
		cash:        s.cash,
		debts:       debts,
	}
}

func resolveScenarioConfigs(req ProjectionRequest) []scenarioConfig {
	if req.ReturnRate != nil {
		customRate := float64(*req.ReturnRate) / 100
		return []scenarioConfig{
			{name: Best, returnRate: customRate + 0.03},
			{name: Average, returnRate: customRate},
			{name: Worst, returnRate: math.Max(customRate-0.03, 0.01)},
		}
	}
	return defaultScenarios
}

func projectScenario(
	start time.Time,
	state *projectionState,
	monthlySavings, annualReturn, extraDebtPayment float64,
	req ProjectionRequest,
) ([]struct {
	Date  openapi_types.Date `json:"date"`
	Value float32            `json:"value"`
}, *openapi_types.Date, *openapi_types.Date) {
	monthlyReturn := annualReturn / 12
	totalMonths := req.YearsToProject * 12

	events := parseOneTimeEvents(req)

	points := make([]struct {
		Date  openapi_types.Date `json:"date"`
		Value float32            `json:"value"`
	}, 0, totalMonths+1)

	var millionaireDate *openapi_types.Date
	var debtFreeDate *openapi_types.Date

	for month := 0; month <= totalMonths; month++ {
		currentDate := start.AddDate(0, month, 0)

		for _, ev := range events {
			if sameMonth(currentDate, ev.date) {
				state.cash += ev.amount
			}
		}

		if month > 0 {
			state.investments *= (1 + monthlyReturn)

			freedPayments := applyDebtPayments(state, extraDebtPayment)

			investable := monthlySavings + freedPayments
			if investable > 0 {
				state.investments += investable
			}

			if debtFreeDate == nil && len(state.debts) > 0 && allDebtsPaid(state.debts) {
				d := openapi_types.Date{Time: currentDate}
				debtFreeDate = &d
			}
		}

		netWorth := state.investments + state.cash - totalDebt(state.debts)

		if month%3 == 0 || month == totalMonths {
			points = append(points, struct {
				Date  openapi_types.Date `json:"date"`
				Value float32            `json:"value"`
			}{
				Date:  openapi_types.Date{Time: currentDate},
				Value: float32(math.Round(netWorth*100) / 100),
			})
		}

		if millionaireDate == nil && netWorth >= 1_000_000 {
			d := openapi_types.Date{Time: currentDate}
			millionaireDate = &d
		}
	}

	return points, millionaireDate, debtFreeDate
}

func applyDebtPayments(state *projectionState, extraPayment float64) float64 {
	remainingExtra := extraPayment
	freedPayments := 0.0

	for i := range state.debts {
		d := &state.debts[i]
		if d.balance <= 0 {
			freedPayments += d.minPayment
			continue
		}

		interest := d.balance * d.monthlyRate
		payment := d.minPayment + remainingExtra
		remainingExtra = 0

		principal := payment - interest
		if principal > d.balance {
			remainingExtra = principal - d.balance
			principal = d.balance
		}
		if principal < 0 {
			principal = 0
		}

		d.balance -= principal
		if d.balance < 0.01 {
			d.balance = 0
		}
	}

	return freedPayments
}

func allDebtsPaid(debts []debtState) bool {
	for _, d := range debts {
		if d.balance > 0 {
			return false
		}
	}
	return true
}

func totalDebt(debts []debtState) float64 {
	total := 0.0
	for _, d := range debts {
		total += d.balance
	}
	return total
}

type oneTimeEvent struct {
	date   time.Time
	amount float64
}

func parseOneTimeEvents(req ProjectionRequest) []oneTimeEvent {
	if req.OneTimeEvents == nil {
		return nil
	}
	events := make([]oneTimeEvent, len(*req.OneTimeEvents))
	for i, ev := range *req.OneTimeEvents {
		amount := float64(ev.Amount)
		if ev.Type == ProjectionRequestOneTimeEventsTypeExpense {
			amount = -amount
		}
		events[i] = oneTimeEvent{date: ev.Date.Time, amount: amount}
	}
	return events
}

func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}
