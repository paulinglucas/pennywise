package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type GoalRepository interface {
	List(ctx context.Context, userID string, page, perPage int) ([]models.Goal, int, error)
	Create(ctx context.Context, goal *models.Goal) error
	GetByID(ctx context.Context, userID, id string) (*models.Goal, error)
	Update(ctx context.Context, goal *models.Goal) (bool, error)
	SoftDelete(ctx context.Context, userID, id string) (bool, error)
	Reorder(ctx context.Context, userID string, rankings []queries.GoalRanking) error
	NextPriorityRank(ctx context.Context, userID string) (int, error)
}

func (h *AppHandler) ListGoals(w http.ResponseWriter, r *http.Request, params ListGoalsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	goals, total, err := h.goals.List(r.Context(), userID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list goals", requestID)
		return
	}

	data := make([]GoalResponse, len(goals))
	for i, g := range goals {
		data[i] = goalToResponse(g)
	}

	WriteJSON(w, http.StatusOK, GoalListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "name is required", requestID)
		return
	}

	goal, err := h.newGoalModel(r.Context(), userID, req)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to determine priority rank", requestID)
		return
	}

	if err := h.goals.Create(r.Context(), goal); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create goal", requestID)
		return
	}

	created, err := h.goals.GetByID(r.Context(), userID, goal.ID)
	if err != nil || created == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve goal", requestID)
		return
	}

	newData := goalJSON(created)
	h.recordAudit(r.Context(), userID, "goal", created.ID, "create", nil, &newData)

	WriteJSON(w, http.StatusCreated, goalToResponse(*created))
}

func (h *AppHandler) GetGoal(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	goal, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get goal", requestID)
		return
	}
	if goal == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, goalToResponse(*goal))
}

func (h *AppHandler) UpdateGoal(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	goal, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get goal", requestID)
		return
	}
	if goal == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	prevData := goalJSON(goal)

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	applyGoalUpdates(goal, req)

	if _, err := h.goals.Update(r.Context(), goal); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update goal", requestID)
		return
	}

	updated, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil || updated == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve goal", requestID)
		return
	}

	newData := goalJSON(updated)
	h.recordAudit(r.Context(), userID, "goal", updated.ID, "update", &prevData, &newData)

	WriteJSON(w, http.StatusOK, goalToResponse(*updated))
}

func (h *AppHandler) DeleteGoal(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	goal, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get goal", requestID)
		return
	}
	if goal == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	prevData := goalJSON(goal)

	found, err := h.goals.SoftDelete(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete goal", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	h.recordAudit(r.Context(), userID, "goal", id.String(), "delete", &prevData, nil)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppHandler) ReorderGoals(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req GoalReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if len(req.Rankings) == 0 {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "rankings must not be empty", requestID)
		return
	}

	rankings := make([]queries.GoalRanking, len(req.Rankings))
	for i, r := range req.Rankings {
		rankings[i] = queries.GoalRanking{
			ID:   r.Id.String(),
			Rank: r.PriorityRank,
		}
	}

	if err := h.goals.Reorder(r.Context(), userID, rankings); err != nil {
		if err == queries.ErrGoalNotFound {
			WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "One or more goal IDs not found", requestID)
			return
		}
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to reorder goals", requestID)
		return
	}

	goals, total, err := h.goals.List(r.Context(), userID, 1, 100)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list goals", requestID)
		return
	}

	data := make([]GoalResponse, len(goals))
	for i, g := range goals {
		data[i] = goalToResponse(g)
	}

	WriteJSON(w, http.StatusOK, GoalListResponse{
		Data:       data,
		Pagination: paginationMeta(1, 100, total),
	})
}

func (h *AppHandler) newGoalModel(ctx context.Context, userID string, req CreateGoalRequest) (*models.Goal, error) {
	var rank int
	if req.PriorityRank != nil {
		rank = *req.PriorityRank
	} else {
		nextRank, err := h.goals.NextPriorityRank(ctx, userID)
		if err != nil {
			return nil, err
		}
		rank = nextRank
	}

	goal := &models.Goal{
		ID:            uuid.New().String(),
		UserID:        userID,
		Name:          req.Name,
		GoalType:      string(req.GoalType),
		TargetAmount:  float64(req.TargetAmount),
		CurrentAmount: 0,
		PriorityRank:  rank,
	}

	if req.CurrentAmount != nil {
		goal.CurrentAmount = float64(*req.CurrentAmount)
	}

	if req.Deadline != nil {
		t := req.Deadline.Time
		goal.Deadline = &t
	}

	if req.LinkedAccountId != nil {
		s := req.LinkedAccountId.String()
		goal.LinkedAccountID = &s
	}

	return goal, nil
}

func goalToResponse(g models.Goal) GoalResponse {
	resp := GoalResponse{
		Id:            ParseID(g.ID),
		UserId:        ParseID(g.UserID),
		Name:          g.Name,
		GoalType:      GoalType(g.GoalType),
		TargetAmount:  float32(g.TargetAmount),
		CurrentAmount: float32(g.CurrentAmount),
		PriorityRank:  g.PriorityRank,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}

	if g.Deadline != nil {
		d := openapi_types.Date{Time: *g.Deadline}
		resp.Deadline = &d
	}

	if g.LinkedAccountID != nil {
		id := ParseID(*g.LinkedAccountID)
		resp.LinkedAccountId = &id
	}

	computeGoalProjections(&resp, g)

	return resp
}

func computeGoalProjections(resp *GoalResponse, g models.Goal) {
	remaining := g.TargetAmount - g.CurrentAmount
	if remaining <= 0 {
		onTrack := true
		resp.OnTrack = &onTrack
		zero := float32(0)
		resp.RequiredMonthlyContribution = &zero
		return
	}

	if g.Deadline == nil {
		return
	}

	now := time.Now()
	monthsLeft := monthsBetween(now, *g.Deadline)

	if monthsLeft <= 0 {
		onTrack := false
		resp.OnTrack = &onTrack
		return
	}

	requiredMonthly := float32(remaining / float64(monthsLeft))
	resp.RequiredMonthlyContribution = &requiredMonthly

	completionDate := estimateCompletionDate(now, g.CurrentAmount, g.TargetAmount, g.CreatedAt)
	if completionDate != nil {
		d := openapi_types.Date{Time: *completionDate}
		resp.ProjectedCompletionDate = &d
	}

	if completionDate != nil {
		onTrack := !completionDate.After(*g.Deadline)
		resp.OnTrack = &onTrack
	} else {
		onTrack := false
		resp.OnTrack = &onTrack
	}
}

func monthsBetween(from, to time.Time) float64 {
	years := float64(to.Year() - from.Year())
	months := float64(to.Month() - from.Month())
	days := float64(to.Day() - from.Day())
	return years*12 + months + days/30.0
}

func estimateCompletionDate(now time.Time, current, target float64, createdAt time.Time) *time.Time {
	if current <= 0 {
		return nil
	}

	elapsed := monthsBetween(createdAt, now)
	if elapsed <= 0 {
		return nil
	}

	monthlyRate := current / elapsed
	if monthlyRate <= 0 {
		return nil
	}

	remaining := target - current
	monthsNeeded := remaining / monthlyRate

	completion := now.AddDate(0, int(math.Ceil(monthsNeeded)), 0)
	return &completion
}

func applyGoalUpdates(goal *models.Goal, req UpdateGoalRequest) {
	if req.Name != nil {
		goal.Name = *req.Name
	}
	if req.GoalType != nil {
		goal.GoalType = string(*req.GoalType)
	}
	if req.TargetAmount != nil {
		goal.TargetAmount = float64(*req.TargetAmount)
	}
	if req.CurrentAmount != nil {
		goal.CurrentAmount = float64(*req.CurrentAmount)
	}
	if req.Deadline != nil {
		t := req.Deadline.Time
		goal.Deadline = &t
	}
	if req.LinkedAccountId != nil {
		s := req.LinkedAccountId.String()
		goal.LinkedAccountID = &s
	}
	if req.PriorityRank != nil {
		goal.PriorityRank = *req.PriorityRank
	}
}

func goalJSON(goal *models.Goal) string {
	data, _ := json.Marshal(goal)
	return string(data)
}
