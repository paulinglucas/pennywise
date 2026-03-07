package api

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

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

	currentNetWorth := nwResult.AssetTotal - nwResult.DebtTotal

	monthlySavings := h.estimateMonthlySavings(r, userID, req)

	scenarios := buildScenarios(req, currentNetWorth, monthlySavings)

	WriteJSON(w, http.StatusOK, ProjectionResponse{Scenarios: scenarios})
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

func buildScenarios(req ProjectionRequest, currentNetWorth, monthlySavings float64) []struct {
	DataPoints []struct {
		Date  openapi_types.Date `json:"date"`
		Value float32            `json:"value"`
	} `json:"data_points"`
	MillionaireDate *openapi_types.Date `json:"millionaire_date,omitempty"`
	Scenario        Scenario            `json:"scenario"`
} {
	configs := resolveScenarioConfigs(req)

	result := make([]struct {
		DataPoints []struct {
			Date  openapi_types.Date `json:"date"`
			Value float32            `json:"value"`
		} `json:"data_points"`
		MillionaireDate *openapi_types.Date `json:"millionaire_date,omitempty"`
		Scenario        Scenario            `json:"scenario"`
	}, len(configs))

	now := time.Now()

	for i, cfg := range configs {
		points, millDate := projectScenario(now, currentNetWorth, monthlySavings, cfg.returnRate, req)
		result[i].Scenario = cfg.name
		result[i].DataPoints = points
		result[i].MillionaireDate = millDate
	}

	return result
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

func projectScenario(start time.Time, netWorth, monthlySavings, annualReturn float64, req ProjectionRequest) ([]struct {
	Date  openapi_types.Date `json:"date"`
	Value float32            `json:"value"`
}, *openapi_types.Date) {
	monthlyReturn := annualReturn / 12
	totalMonths := req.YearsToProject * 12

	events := parseOneTimeEvents(req)

	points := make([]struct {
		Date  openapi_types.Date `json:"date"`
		Value float32            `json:"value"`
	}, 0, totalMonths+1)

	value := netWorth
	var millionaireDate *openapi_types.Date

	for month := 0; month <= totalMonths; month++ {
		currentDate := start.AddDate(0, month, 0)

		for _, ev := range events {
			if sameMonth(currentDate, ev.date) {
				value += ev.amount
			}
		}

		if month > 0 {
			value = value*(1+monthlyReturn) + monthlySavings
		}

		if month%3 == 0 || month == totalMonths {
			points = append(points, struct {
				Date  openapi_types.Date `json:"date"`
				Value float32            `json:"value"`
			}{
				Date:  openapi_types.Date{Time: currentDate},
				Value: float32(math.Round(value*100) / 100),
			})
		}

		if millionaireDate == nil && value >= 1_000_000 {
			d := openapi_types.Date{Time: currentDate}
			millionaireDate = &d
		}
	}

	return points, millionaireDate
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
