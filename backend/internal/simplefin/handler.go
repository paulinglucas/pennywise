package simplefin

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

type Handler struct {
	repo          *SQLiteSimplefinRepository
	client        *Client
	syncService   *SyncService
	encryptionKey []byte
}

func NewHandler(repo *SQLiteSimplefinRepository, client *Client, syncService *SyncService, encryptionKey []byte) *Handler {
	return &Handler{
		repo:          repo,
		client:        client,
		syncService:   syncService,
		encryptionKey: encryptionKey,
	}
}

func Routes(handler *Handler, jwtSecret []byte) chi.Router {
	r := chi.NewRouter()
	r.Use(jwtAuth(jwtSecret))
	r.Post("/setup", handler.Setup)
	r.Get("/status", handler.Status)
	r.Delete("/", handler.Disconnect)
	r.Get("/accounts", handler.ListSimplefinAccounts)
	r.Post("/link", handler.LinkAccount)
	r.Delete("/link/{accountId}", handler.UnlinkAccount)
	r.Post("/sync", handler.TriggerSync)
	return r
}

type setupRequest struct {
	SetupToken string `json:"setup_token"`
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.SetupToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "setup_token is required"})
		return
	}

	accessURL, err := h.client.ClaimToken(r.Context(), req.SetupToken)
	if err != nil {
		slog.Warn("failed to claim SimpleFIN token", slog.Any("error", err))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Failed to claim token. It may have already been used."})
		return
	}

	encrypted, err := pennywisecrypto.Encrypt(h.encryptionKey, accessURL)
	if err != nil {
		slog.Error("failed to encrypt access URL", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal error"})
		return
	}

	if err := h.repo.SaveConnection(r.Context(), userID, encrypted); err != nil {
		slog.Error("failed to save SimpleFIN connection", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save connection"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "connected"})
}

type statusResponse struct {
	Connected  bool    `json:"connected"`
	LastSyncAt *string `json:"last_sync_at,omitempty"`
	SyncError  *string `json:"sync_error,omitempty"`
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	conn, err := h.repo.GetConnection(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get SimpleFIN connection", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal error"})
		return
	}

	if conn == nil {
		writeJSON(w, http.StatusOK, statusResponse{Connected: false})
		return
	}

	resp := statusResponse{Connected: true, SyncError: conn.SyncError}
	if conn.LastSyncAt != nil {
		ts := conn.LastSyncAt.Format("2006-01-02T15:04:05Z")
		resp.LastSyncAt = &ts
	}

	linked, err := h.repo.GetLinkedAccounts(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get linked accounts", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal error"})
		return
	}

	type fullStatus struct {
		statusResponse
		LinkedAccounts []linkedAccountResponse `json:"linked_accounts"`
	}

	accounts := make([]linkedAccountResponse, len(linked))
	for i, la := range linked {
		accounts[i] = linkedAccountResponse(la)
	}

	writeJSON(w, http.StatusOK, fullStatus{
		statusResponse: resp,
		LinkedAccounts: accounts,
	})
}

type linkedAccountResponse struct {
	AccountID   string `json:"account_id"`
	SimplefinID string `json:"simplefin_id"`
	AccountName string `json:"account_name"`
	Institution string `json:"institution"`
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	if err := h.repo.DeleteConnection(r.Context(), userID); err != nil {
		slog.Error("failed to delete SimpleFIN connection", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to disconnect"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type simplefinAccountResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Institution string `json:"institution"`
	Balance     string `json:"balance"`
	Currency    string `json:"currency"`
}

func (h *Handler) ListSimplefinAccounts(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	conn, err := h.repo.GetConnection(r.Context(), userID)
	if err != nil || conn == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SimpleFIN not connected"})
		return
	}

	accessURL, err := pennywisecrypto.Decrypt(h.encryptionKey, conn.AccessURL)
	if err != nil {
		slog.Error("failed to decrypt access URL", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal error"})
		return
	}

	username, password, baseURL, err := ParseAccessURL(accessURL)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Invalid stored access URL"})
		return
	}

	resp, err := h.client.FetchAccounts(r.Context(), username, password, baseURL)
	if err != nil {
		slog.Warn("failed to fetch SimpleFIN accounts", slog.Any("error", err))
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to fetch accounts from SimpleFIN"})
		return
	}

	accounts := make([]simplefinAccountResponse, len(resp.Accounts))
	for i, a := range resp.Accounts {
		accounts[i] = simplefinAccountResponse{
			ID:          a.ID,
			Name:        a.Name,
			Institution: a.Org.Name,
			Balance:     a.Balance,
			Currency:    a.Currency,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"accounts": accounts})
}

type linkRequest struct {
	AccountID   string `json:"account_id"`
	SimplefinID string `json:"simplefin_id"`
}

func (h *Handler) LinkAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req linkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.AccountID == "" || req.SimplefinID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "account_id and simplefin_id are required"})
		return
	}

	if err := h.repo.LinkAccount(r.Context(), userID, req.AccountID, req.SimplefinID); err != nil {
		slog.Warn("failed to link account", slog.Any("error", err))
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "Account not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "linked"})
}

func (h *Handler) UnlinkAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	accountID := chi.URLParam(r, "accountId")

	if err := h.repo.UnlinkAccount(r.Context(), userID, accountID); err != nil {
		slog.Error("failed to unlink account", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to unlink"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "unlinked"})
}

type syncResponse struct {
	Updated int    `json:"updated"`
	Errors  int    `json:"errors"`
	Message string `json:"message"`
}

func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	conn, err := h.repo.GetConnection(r.Context(), userID)
	if err != nil || conn == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SimpleFIN not connected"})
		return
	}

	accessURL, err := pennywisecrypto.Decrypt(h.encryptionKey, conn.AccessURL)
	if err != nil {
		slog.Error("failed to decrypt access URL", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal error"})
		return
	}

	result, err := h.syncService.SyncUser(r.Context(), userID, accessURL)
	if err != nil {
		slog.Error("sync failed", slog.Any("error", err))
		_ = h.repo.UpdateSyncError(r.Context(), userID, err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Sync failed: " + err.Error()})
		return
	}

	_ = h.repo.UpdateSyncSuccess(r.Context(), userID)

	writeJSON(w, http.StatusOK, syncResponse{
		Updated: result.Updated,
		Errors:  result.Errors,
		Message: "Sync complete",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func jwtAuth(secret []byte) func(http.Handler) http.Handler {
	return middleware.JWTAuth(secret)
}
