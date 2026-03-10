package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type AssetRepository interface {
	List(ctx context.Context, userID string, page, perPage int) ([]models.Asset, int, error)
	Create(ctx context.Context, asset *models.Asset) error
	GetByID(ctx context.Context, userID, id string) (*models.Asset, error)
	Update(ctx context.Context, asset *models.Asset, prevValue float64) (bool, error)
	SoftDelete(ctx context.Context, userID, id string) (bool, error)
	GetHistory(ctx context.Context, userID, assetID string, since *time.Time) ([]models.AssetHistory, error)
	GetAllocation(ctx context.Context, userID string) ([]queries.AllocationRow, error)
	GetAllocationOverTime(ctx context.Context, userID string, since *time.Time) ([]queries.AllocationSnapshot, error)
	GetLinkedAccounts(ctx context.Context, accountIDs []string) (map[string]queries.LinkedAccountRow, error)
}

func (h *AppHandler) ListAssets(w http.ResponseWriter, r *http.Request, params ListAssetsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	assets, total, err := h.assets.List(r.Context(), userID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list assets", requestID)
		return
	}

	allocation, err := h.assets.GetAllocation(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to compute allocation", requestID)
		return
	}

	linkedAccounts, err := h.resolveLinkedAccounts(r.Context(), assets)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to resolve linked accounts", requestID)
		return
	}

	data := make([]AssetResponse, len(assets))
	for i, a := range assets {
		data[i] = assetToResponse(a)
		attachLinkedAccount(&data[i], a, linkedAccounts)
	}

	WriteJSON(w, http.StatusOK, AssetListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
		Summary:    buildPortfolioSummary(allocation),
	})
}

func (h *AppHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if req.Name == "" {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "name is required", requestID)
		return
	}

	asset := newAssetModel(userID, req)

	if err := h.assets.Create(r.Context(), asset); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create asset", requestID)
		return
	}

	created, err := h.assets.GetByID(r.Context(), userID, asset.ID)
	if err != nil || created == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve asset", requestID)
		return
	}

	newData := assetJSON(created)
	h.recordAudit(r.Context(), userID, "asset", created.ID, "create", nil, &newData)

	WriteJSON(w, http.StatusCreated, assetToResponse(*created))
}

func (h *AppHandler) GetAsset(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	asset, err := h.assets.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get asset", requestID)
		return
	}
	if asset == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Asset not found", requestID)
		return
	}

	history, err := h.assets.GetHistory(r.Context(), userID, id.String(), nil)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get asset history", requestID)
		return
	}

	resp := assetToResponse(*asset)
	historyEntries := historyToEntries(id, history)
	resp.History = &historyEntries

	linkedAccounts, err := h.resolveLinkedAccounts(r.Context(), []models.Asset{*asset})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to resolve linked account", requestID)
		return
	}
	attachLinkedAccount(&resp, *asset, linkedAccounts)

	WriteJSON(w, http.StatusOK, resp)
}

func (h *AppHandler) UpdateAsset(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	asset, err := h.assets.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get asset", requestID)
		return
	}
	if asset == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Asset not found", requestID)
		return
	}

	prevData := assetJSON(asset)
	prevValue := asset.CurrentValue

	var req UpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	applyAssetUpdates(asset, req)

	if _, err := h.assets.Update(r.Context(), asset, prevValue); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update asset", requestID)
		return
	}

	updated, err := h.assets.GetByID(r.Context(), userID, id.String())
	if err != nil || updated == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve asset", requestID)
		return
	}

	newData := assetJSON(updated)
	h.recordAudit(r.Context(), userID, "asset", updated.ID, "update", &prevData, &newData)

	WriteJSON(w, http.StatusOK, assetToResponse(*updated))
}

func (h *AppHandler) DeleteAsset(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	asset, err := h.assets.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get asset", requestID)
		return
	}
	if asset == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Asset not found", requestID)
		return
	}

	prevData := assetJSON(asset)

	found, err := h.assets.SoftDelete(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete asset", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Asset not found", requestID)
		return
	}

	h.recordAudit(r.Context(), userID, "asset", id.String(), "delete", &prevData, nil)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppHandler) GetAssetHistory(w http.ResponseWriter, r *http.Request, id IdParam, params GetAssetHistoryParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	since := periodToTime(params.Period)

	history, err := h.assets.GetHistory(r.Context(), userID, id.String(), since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get asset history", requestID)
		return
	}
	if history == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Asset not found", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, AssetHistoryResponse{
		AssetId: id,
		Entries: historyToEntries(id, history),
	})
}

func (h *AppHandler) GetAssetAllocation(w http.ResponseWriter, r *http.Request, params GetAssetAllocationParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	since := allocationPeriodToTime(params.Period)

	snapshots, err := h.assets.GetAllocationOverTime(r.Context(), userID, since)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get allocation data", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, AllocationResponse{
		Snapshots: snapshotsToResponse(snapshots),
	})
}

func snapshotsToResponse(snapshots []queries.AllocationSnapshot) []struct {
	Allocations []AllocationEntry  `json:"allocations"`
	Date        openapi_types.Date `json:"date"`
} {
	result := make([]struct {
		Allocations []AllocationEntry  `json:"allocations"`
		Date        openapi_types.Date `json:"date"`
	}, len(snapshots))

	for i, snap := range snapshots {
		date, _ := time.Parse("2006-01-02", snap.Date)
		var totalValue float64
		for _, a := range snap.Allocations {
			totalValue += a.TotalValue
		}
		entries := make([]AllocationEntry, len(snap.Allocations))
		for j, a := range snap.Allocations {
			pct := float32(0)
			if totalValue > 0 {
				pct = float32(a.TotalValue / totalValue * 100)
			}
			entries[j] = AllocationEntry{
				AssetType:  AssetType(a.AssetType),
				Value:      float32(a.TotalValue),
				Percentage: pct,
			}
		}
		result[i] = struct {
			Allocations []AllocationEntry  `json:"allocations"`
			Date        openapi_types.Date `json:"date"`
		}{
			Date:        openapi_types.Date{Time: date},
			Allocations: entries,
		}
	}
	return result
}

func newAssetModel(userID string, req CreateAssetRequest) *models.Asset {
	currency := USD
	if req.Currency != nil {
		currency = *req.Currency
	}

	asset := &models.Asset{
		ID:           uuid.New().String(),
		UserID:       userID,
		Name:         req.Name,
		AssetType:    string(req.AssetType),
		CurrentValue: float64(req.CurrentValue),
		Currency:     string(currency),
	}

	if req.AccountId != nil {
		s := req.AccountId.String()
		asset.AccountID = &s
	}

	if req.Metadata != nil {
		data, _ := json.Marshal(req.Metadata)
		s := string(data)
		asset.Metadata = &s
	}

	return asset
}

func assetToResponse(a models.Asset) AssetResponse {
	resp := AssetResponse{
		Id:           ParseID(a.ID),
		UserId:       ParseID(a.UserID),
		Name:         a.Name,
		AssetType:    AssetType(a.AssetType),
		CurrentValue: float32(a.CurrentValue),
		Currency:     Currency(a.Currency),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}

	if a.AccountID != nil {
		id := ParseID(*a.AccountID)
		resp.AccountId = &id
	}

	if a.Metadata != nil {
		var meta map[string]interface{}
		if err := json.Unmarshal([]byte(*a.Metadata), &meta); err == nil {
			resp.Metadata = &meta
		}
	}

	return resp
}

func applyAssetUpdates(asset *models.Asset, req UpdateAssetRequest) {
	if req.Name != nil {
		asset.Name = *req.Name
	}
	if req.AssetType != nil {
		asset.AssetType = string(*req.AssetType)
	}
	if req.CurrentValue != nil {
		asset.CurrentValue = float64(*req.CurrentValue)
	}
	if req.Currency != nil {
		asset.Currency = string(*req.Currency)
	}
	if req.AccountId != nil {
		s := req.AccountId.String()
		asset.AccountID = &s
	}
	if req.Metadata != nil {
		data, _ := json.Marshal(req.Metadata)
		s := string(data)
		asset.Metadata = &s
	}
}

func historyToEntries(assetID openapi_types.UUID, history []models.AssetHistory) []AssetHistoryEntry {
	entries := make([]AssetHistoryEntry, len(history))
	for i, h := range history {
		entries[i] = AssetHistoryEntry{
			Id:         ParseID(h.ID),
			AssetId:    assetID,
			Value:      float32(h.Value),
			RecordedAt: h.RecordedAt,
		}
	}
	return entries
}

func buildPortfolioSummary(allocation []queries.AllocationRow) PortfolioSummary {
	var totalValue float64
	for _, a := range allocation {
		totalValue += a.TotalValue
	}

	entries := make([]AllocationEntry, len(allocation))
	for i, a := range allocation {
		pct := float32(0)
		if totalValue > 0 {
			pct = float32(a.TotalValue / totalValue * 100)
		}
		entries[i] = AllocationEntry{
			AssetType:  AssetType(a.AssetType),
			Value:      float32(a.TotalValue),
			Percentage: pct,
		}
	}

	return PortfolioSummary{
		TotalValue: float32(totalValue),
		Allocation: entries,
	}
}

func (h *AppHandler) resolveLinkedAccounts(ctx context.Context, assets []models.Asset) (map[string]queries.LinkedAccountRow, error) {
	var accountIDs []string
	for _, a := range assets {
		if a.AccountID != nil {
			accountIDs = append(accountIDs, *a.AccountID)
		}
	}
	if len(accountIDs) == 0 {
		return map[string]queries.LinkedAccountRow{}, nil
	}
	return h.assets.GetLinkedAccounts(ctx, accountIDs)
}

func attachLinkedAccount(resp *AssetResponse, asset models.Asset, linked map[string]queries.LinkedAccountRow) {
	if asset.AccountID == nil {
		return
	}
	row, ok := linked[*asset.AccountID]
	if !ok {
		return
	}
	summary := LinkedAccountSummary{
		Id:          ParseID(row.AccountID),
		Name:        row.Name,
		AccountType: AccountType(row.AccountType),
		Institution: &row.Institution,
	}
	if row.Balance != nil {
		bal := float32(*row.Balance)
		summary.Balance = &bal
	}
	resp.LinkedAccount = &summary
}

func assetJSON(asset *models.Asset) string {
	data, _ := json.Marshal(asset)
	return string(data)
}

func periodToTime[T ~string](period *T) *time.Time {
	if period == nil {
		return nil
	}
	now := time.Now()
	var since time.Time
	switch string(*period) {
	case "1m":
		since = now.AddDate(0, -1, 0)
	case "3m":
		since = now.AddDate(0, -3, 0)
	case "6m":
		since = now.AddDate(0, -6, 0)
	case "1y":
		since = now.AddDate(-1, 0, 0)
	case "all":
		return nil
	default:
		since = now.AddDate(0, -6, 0)
	}
	return &since
}

func allocationPeriodToTime(period *GetAssetAllocationParamsPeriod) *time.Time {
	return periodToTime(period)
}
