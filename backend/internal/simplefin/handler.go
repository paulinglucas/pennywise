package simplefin

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

const minSyncInterval = 30 * time.Minute

type accountsCache struct {
	mu        sync.Mutex
	accounts  []simplefinAccountResponse
	fetchedAt time.Time
	userID    string
}

type Handler struct {
	repo          *SQLiteSimplefinRepository
	client        *Client
	syncService   *SyncService
	encryptionKey []byte
	cache         accountsCache
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
	r.Post("/dismiss", handler.DismissAccount)
	r.Delete("/dismiss/{simplefinId}", handler.UndismissAccount)
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
		accounts[i] = linkedAccountResponse{
			AccountID:   la.AccountID,
			SimplefinID: la.SimplefinID,
			AccountName: la.AccountName,
			Institution: la.Institution,
			AccountType: la.AccountType,
		}
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
	AccountType string `json:"account_type"`
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

	accounts := h.getCachedAccounts(userID)
	if accounts == nil {
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

		resp, err := h.client.FetchAccounts(r.Context(), username, password, baseURL, nil)
		if err != nil {
			slog.Warn("failed to fetch SimpleFIN accounts", slog.Any("error", err))
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Failed to fetch accounts from SimpleFIN"})
			return
		}

		accounts = deduplicateAccounts(resp.Accounts)
		h.setCachedAccounts(userID, accounts)
	}

	dismissed, err := h.repo.GetDismissedAccountIDs(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get dismissed accounts", slog.Any("error", err))
		dismissed = nil
	}

	writeJSON(w, http.StatusOK, map[string]any{"accounts": accounts, "dismissed": dismissed})
}

func (h *Handler) getCachedAccounts(userID string) []simplefinAccountResponse {
	h.cache.mu.Lock()
	defer h.cache.mu.Unlock()
	if h.cache.userID == userID && time.Since(h.cache.fetchedAt) < minSyncInterval && len(h.cache.accounts) > 0 {
		return h.cache.accounts
	}
	return nil
}

func (h *Handler) setCachedAccounts(userID string, accounts []simplefinAccountResponse) {
	h.cache.mu.Lock()
	defer h.cache.mu.Unlock()
	h.cache.accounts = accounts
	h.cache.fetchedAt = time.Now()
	h.cache.userID = userID
}

type linkRequest struct {
	SimplefinID    string   `json:"simplefin_id"`
	AccountType    string   `json:"account_type"`
	Name           string   `json:"name"`
	Institution    string   `json:"institution"`
	Balance        string   `json:"balance"`
	Currency       string   `json:"currency"`
	InterestRate   *float64 `json:"interest_rate,omitempty"`
	LoanTermMonths *int     `json:"loan_term_months,omitempty"`
	PurchasePrice  *float64 `json:"purchase_price,omitempty"`
	PurchaseDate   *string  `json:"purchase_date,omitempty"`
	DownPaymentPct *float64 `json:"down_payment_pct,omitempty"`
}

func (h *Handler) LinkAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req linkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.SimplefinID == "" || req.AccountType == "" || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "simplefin_id, account_type, and name are required"})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	balance, _ := strconv.ParseFloat(req.Balance, 64)

	mortgageFields := &MortgageFields{
		InterestRate:   req.InterestRate,
		LoanTermMonths: req.LoanTermMonths,
		PurchasePrice:  req.PurchasePrice,
		PurchaseDate:   req.PurchaseDate,
		DownPaymentPct: req.DownPaymentPct,
	}

	accountID, err := h.repo.CreateAccountWithLink(r.Context(), userID, req.Name, req.Institution, req.AccountType, currency, req.SimplefinID, balance, mortgageFields)
	if err != nil {
		slog.Warn("failed to create and link account", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to link account"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "linked", "account_id": accountID})
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

type dismissRequest struct {
	SimplefinID string `json:"simplefin_id"`
}

func (h *Handler) DismissAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req dismissRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.SimplefinID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "simplefin_id is required"})
		return
	}

	if err := h.repo.DismissAccount(r.Context(), userID, req.SimplefinID); err != nil {
		slog.Error("failed to dismiss account", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to dismiss"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "dismissed"})
}

func (h *Handler) UndismissAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	simplefinID := chi.URLParam(r, "simplefinId")

	if err := h.repo.UndismissAccount(r.Context(), userID, simplefinID); err != nil {
		slog.Error("failed to undismiss account", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to undismiss"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "undismissed"})
}

type syncResponse struct {
	Updated              int    `json:"updated"`
	Errors               int    `json:"errors"`
	TransactionsImported int    `json:"transactions_imported"`
	Message              string `json:"message"`
}

func (h *Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	conn, err := h.repo.GetConnection(r.Context(), userID)
	if err != nil || conn == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "SimpleFIN not connected"})
		return
	}

	if conn.LastSyncAt != nil && time.Since(*conn.LastSyncAt) < minSyncInterval {
		hasUnsynced, _ := h.repo.HasUnsyncedLinkedAccounts(r.Context(), userID)
		if !hasUnsynced {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "Please wait at least 30 minutes between syncs"})
			return
		}
	}

	accessURL, err := pennywisecrypto.Decrypt(h.encryptionKey, conn.AccessURL)
	if err != nil {
		slog.Error("failed to decrypt access URL", slog.Any("error", err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal error"})
		return
	}

	result, err := h.syncService.SyncUser(r.Context(), userID, accessURL, conn.LastSyncAt)
	if err != nil {
		slog.Error("sync failed", slog.Any("error", err))
		_ = h.repo.UpdateSyncError(r.Context(), userID, err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Sync failed: " + err.Error()})
		return
	}

	_ = h.repo.UpdateSyncSuccess(r.Context(), userID)

	writeJSON(w, http.StatusOK, syncResponse{
		Updated:              result.Updated,
		Errors:               result.Errors,
		TransactionsImported: result.TransactionsImported,
		Message: "Sync complete",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func deduplicateAccounts(raw []Account) []simplefinAccountResponse {
	type entry struct {
		resp    simplefinAccountResponse
		balance float64
	}

	seen := make(map[string]entry)
	order := make([]string, 0, len(raw))

	for _, a := range raw {
		key := a.Name
		bal, _ := strconv.ParseFloat(a.Balance, 64)

		existing, exists := seen[key]
		if !exists {
			order = append(order, key)
		}

		if !exists || (existing.balance == 0 && bal != 0) {
			seen[key] = entry{
				resp: simplefinAccountResponse{
					ID:          a.ID,
					Name:        a.Name,
					Institution: a.Org.Name,
					Balance:     a.Balance,
					Currency:    a.Currency,
				},
				balance: bal,
			}
		}
	}

	result := make([]simplefinAccountResponse, 0, len(order))
	for _, key := range order {
		result = append(result, seen[key].resp)
	}
	return result
}

func jwtAuth(secret []byte) func(http.Handler) http.Handler {
	return middleware.JWTAuth(secret)
}
