package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type TransactionRepository interface {
	List(ctx context.Context, userID string, filter queries.TransactionFilter, page, perPage int) ([]models.Transaction, int, error)
	Create(ctx context.Context, txn *models.Transaction, tags []string) error
	GetByID(ctx context.Context, userID, id string) (*models.Transaction, error)
	Update(ctx context.Context, txn *models.Transaction, tags []string) (bool, error)
	SoftDelete(ctx context.Context, userID, id string) (bool, error)
	BulkCreate(ctx context.Context, txns []models.Transaction) (int, []queries.BulkCreateError)
	ListCategories(ctx context.Context, userID string) ([]string, error)
}

func (h *AppHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	categories, err := h.transactions.ListCategories(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list categories", requestID)
		return
	}

	if categories == nil {
		categories = []string{}
	}

	WriteJSON(w, http.StatusOK, CategoriesResponse{Categories: categories})
}

type AuditLogWriter interface {
	Record(ctx context.Context, entry *models.AuditLog) error
}

type FailedRequestWriter interface {
	Write(ctx context.Context, entry *models.FailedRequest) error
}

func (h *AppHandler) ListTransactions(w http.ResponseWriter, r *http.Request, params ListTransactionsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)
	filter := buildTransactionFilter(params)

	txns, total, err := h.transactions.List(r.Context(), userID, filter, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list transactions", requestID)
		return
	}

	data := make([]TransactionResponse, len(txns))
	for i, txn := range txns {
		data[i] = transactionToResponse(txn)
	}

	WriteJSON(w, http.StatusOK, TransactionListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if req.Category == "" || req.AccountId.String() == "" {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "category and account_id are required", requestID)
		return
	}

	txn := newTransactionModel(userID, req)
	var tags []string
	if req.Tags != nil {
		tags = *req.Tags
	}

	if err := h.transactions.Create(r.Context(), txn, tags); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create transaction", requestID)
		return
	}

	created, err := h.transactions.GetByID(r.Context(), userID, txn.ID)
	if err != nil || created == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve transaction", requestID)
		return
	}

	newData := transactionJSON(created)
	h.recordAudit(r.Context(), userID, "transaction", created.ID, "create", nil, &newData)

	WriteJSON(w, http.StatusCreated, transactionToResponse(*created))
}

func (h *AppHandler) GetTransaction(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	txn, err := h.transactions.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get transaction", requestID)
		return
	}
	if txn == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction not found", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, transactionToResponse(*txn))
}

func (h *AppHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	txn, err := h.transactions.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get transaction", requestID)
		return
	}
	if txn == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction not found", requestID)
		return
	}

	prevData := transactionJSON(txn)

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	tags := applyTransactionUpdates(txn, req)

	if _, err := h.transactions.Update(r.Context(), txn, tags); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update transaction", requestID)
		return
	}

	updated, err := h.transactions.GetByID(r.Context(), userID, id.String())
	if err != nil || updated == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve transaction", requestID)
		return
	}

	newData := transactionJSON(updated)
	h.recordAudit(r.Context(), userID, "transaction", updated.ID, "update", &prevData, &newData)

	WriteJSON(w, http.StatusOK, transactionToResponse(*updated))
}

func (h *AppHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	txn, err := h.transactions.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get transaction", requestID)
		return
	}
	if txn == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction not found", requestID)
		return
	}

	prevData := transactionJSON(txn)

	found, err := h.transactions.SoftDelete(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete transaction", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction not found", requestID)
		return
	}

	h.recordAudit(r.Context(), userID, "transaction", id.String(), "delete", &prevData, nil)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppHandler) ImportTransactions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	file, accountID, ok := extractImportForm(w, r, requestID)
	if !ok {
		return
	}
	defer func() { _ = file.Close() }()

	txns, parseErrors := parseCSV(file, userID, accountID)
	importErrors := csvErrorsToResponse(parseErrors)

	imported := 0
	if len(txns) > 0 {
		count, bulkErrors := h.transactions.BulkCreate(r.Context(), txns)
		imported = count
		importErrors = append(importErrors, bulkErrorsToResponse(bulkErrors)...)
	}

	WriteJSON(w, http.StatusCreated, ImportResponse{
		Imported: imported,
		Errors:   importErrors,
	})
}

func extractImportForm(w http.ResponseWriter, r *http.Request, requestID string) (multipart.File, string, bool) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid multipart form", requestID)
		return nil, "", false
	}

	accountID := r.FormValue("account_id")
	if accountID == "" {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "account_id is required", requestID)
		return nil, "", false
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "file is required", requestID)
		return nil, "", false
	}

	return file, accountID, true
}

type importErrorEntry = struct {
	Message string `json:"message"`
	Row     int    `json:"row"`
}

func csvErrorsToResponse(errors []csvParseError) []importErrorEntry {
	result := make([]importErrorEntry, len(errors))
	for i, e := range errors {
		result[i] = importErrorEntry{Message: e.Message, Row: e.Row}
	}
	return result
}

func bulkErrorsToResponse(errors []queries.BulkCreateError) []importErrorEntry {
	result := make([]importErrorEntry, len(errors))
	for i, e := range errors {
		result[i] = importErrorEntry{Message: e.Message, Row: e.Row}
	}
	return result
}

func (h *AppHandler) recordAudit(ctx context.Context, userID, entityType, entityID, action string, prevData, newData *string) {
	if h.auditLog == nil {
		return
	}
	_ = h.auditLog.Record(ctx, &models.AuditLog{
		UserID:       userID,
		EntityType:   entityType,
		EntityID:     entityID,
		Action:       action,
		PreviousData: prevData,
		NewData:      newData,
	})
}

func newTransactionModel(userID string, req CreateTransactionRequest) *models.Transaction {
	currency := USD
	if req.Currency != nil {
		currency = *req.Currency
	}
	isRecurring := false
	if req.IsRecurring != nil {
		isRecurring = *req.IsRecurring
	}
	txn := &models.Transaction{
		ID:          uuid.New().String(),
		UserID:      userID,
		AccountID:   req.AccountId.String(),
		Type:        string(req.Type),
		Category:    req.Category,
		Amount:      float64(req.Amount),
		Currency:    string(currency),
		Date:        req.Date.Time,
		Notes:       req.Notes,
		IsRecurring: isRecurring,
	}
	if req.GroupId != nil {
		s := req.GroupId.String()
		txn.GroupID = &s
	}
	return txn
}

func transactionToResponse(txn models.Transaction) TransactionResponse {
	tags := txn.Tags
	if tags == nil {
		tags = []string{}
	}
	resp := TransactionResponse{
		Id:          ParseID(txn.ID),
		UserId:      ParseID(txn.UserID),
		AccountId:   ParseID(txn.AccountID),
		Type:        TransactionType(txn.Type),
		Category:    txn.Category,
		Amount:      float32(txn.Amount),
		Currency:    Currency(txn.Currency),
		Date:        dateToOpenAPI(txn.Date),
		IsRecurring: txn.IsRecurring,
		Notes:       txn.Notes,
		Tags:        tags,
		CreatedAt:   txn.CreatedAt,
		UpdatedAt:   txn.UpdatedAt,
	}
	if txn.RecurringTransactionID != nil {
		id := ParseID(*txn.RecurringTransactionID)
		resp.RecurringTransactionId = &id
	}
	if txn.GroupID != nil {
		id := ParseID(*txn.GroupID)
		resp.GroupId = &id
	}
	return resp
}

func applyTransactionUpdates(txn *models.Transaction, req UpdateTransactionRequest) []string {
	if req.Type != nil {
		txn.Type = string(*req.Type)
	}
	if req.Category != nil {
		txn.Category = *req.Category
	}
	if req.Amount != nil {
		txn.Amount = float64(*req.Amount)
	}
	if req.Currency != nil {
		txn.Currency = string(*req.Currency)
	}
	if req.Date != nil {
		txn.Date = req.Date.Time
	}
	if req.Notes != nil {
		txn.Notes = req.Notes
	}
	if req.AccountId != nil {
		txn.AccountID = req.AccountId.String()
	}
	if req.IsRecurring != nil {
		txn.IsRecurring = *req.IsRecurring
	}
	if req.GroupId != nil {
		s := req.GroupId.String()
		txn.GroupID = &s
	}

	if req.Tags != nil {
		return *req.Tags
	}
	return txn.Tags
}

func buildTransactionFilter(params ListTransactionsParams) queries.TransactionFilter {
	var filter queries.TransactionFilter

	if params.AccountId != nil {
		s := params.AccountId.String()
		filter.AccountID = &s
	}
	if params.Category != nil {
		filter.Category = params.Category
	}
	if params.Type != nil {
		s := string(*params.Type)
		filter.Type = &s
	}
	if params.DateFrom != nil {
		t := params.DateFrom.Time
		filter.DateFrom = &t
	}
	if params.DateTo != nil {
		t := params.DateTo.Time
		filter.DateTo = &t
	}
	if params.AmountMin != nil {
		f := float64(*params.AmountMin)
		filter.AmountMin = &f
	}
	if params.AmountMax != nil {
		f := float64(*params.AmountMax)
		filter.AmountMax = &f
	}
	if params.Tags != nil {
		filter.Tags = strings.Split(*params.Tags, ",")
	}
	if params.Search != nil {
		filter.Search = params.Search
	}
	if params.GroupId != nil {
		s := params.GroupId.String()
		filter.GroupID = &s
	}

	return filter
}

func transactionJSON(txn *models.Transaction) string {
	data, _ := json.Marshal(txn)
	return string(data)
}

type csvParseError struct {
	Row     int
	Message string
}

func parseCSV(reader io.Reader, userID, accountID string) ([]models.Transaction, []csvParseError) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true

	header, err := csvReader.Read()
	if err != nil {
		return nil, []csvParseError{{Row: 1, Message: "Failed to read CSV header"}}
	}

	colIndex := mapCSVColumns(header)
	if colIndex["date"] < 0 || colIndex["amount"] < 0 || colIndex["type"] < 0 || colIndex["category"] < 0 {
		return nil, []csvParseError{{Row: 1, Message: "CSV must have date, amount, type, and category columns"}}
	}

	var txns []models.Transaction
	var errs []csvParseError
	row := 1

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		row++
		if err != nil {
			errs = append(errs, csvParseError{Row: row, Message: "Failed to parse row"})
			continue
		}

		txn, parseErr := parseCSVRow(record, colIndex, userID, accountID, row)
		if parseErr != nil {
			errs = append(errs, *parseErr)
			continue
		}
		txns = append(txns, *txn)
	}

	return txns, errs
}

func mapCSVColumns(header []string) map[string]int {
	idx := map[string]int{
		"date": -1, "amount": -1, "type": -1, "category": -1, "notes": -1, "currency": -1,
	}
	for i, col := range header {
		normalized := strings.ToLower(strings.TrimSpace(col))
		if _, exists := idx[normalized]; exists {
			idx[normalized] = i
		}
	}
	return idx
}

func parseCSVRow(record []string, colIndex map[string]int, userID, accountID string, row int) (*models.Transaction, *csvParseError) {
	if colIndex["date"] >= len(record) || colIndex["amount"] >= len(record) || colIndex["type"] >= len(record) || colIndex["category"] >= len(record) {
		return nil, &csvParseError{Row: row, Message: "Row has fewer columns than header"}
	}

	date, amount, txnType, category, msg := parseRequiredCSVFields(record, colIndex)
	if msg != "" {
		return nil, &csvParseError{Row: row, Message: msg}
	}

	return &models.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		AccountID: accountID,
		Type:      txnType,
		Category:  category,
		Amount:    amount,
		Currency:  csvOptionalField(record, colIndex, "currency", "USD"),
		Date:      date,
		Notes:     csvOptionalFieldPtr(record, colIndex, "notes"),
	}, nil
}

func parseRequiredCSVFields(record []string, colIndex map[string]int) (time.Time, float64, string, string, string) {
	dateStr := strings.TrimSpace(record[colIndex["date"]])
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, 0, "", "", "Invalid date format, expected YYYY-MM-DD"
	}

	amountStr := strings.TrimSpace(record[colIndex["amount"]])
	var amount float64
	if _, err := parseAmount(amountStr, &amount); err != nil {
		return time.Time{}, 0, "", "", "Invalid amount"
	}

	txnType := strings.TrimSpace(record[colIndex["type"]])
	if txnType != "expense" && txnType != "deposit" {
		return time.Time{}, 0, "", "", "Type must be expense or deposit"
	}

	category := strings.TrimSpace(record[colIndex["category"]])
	if category == "" {
		return time.Time{}, 0, "", "", "Category is required"
	}

	return date, amount, txnType, category, ""
}

func csvOptionalField(record []string, colIndex map[string]int, key, defaultVal string) string {
	if colIndex[key] >= 0 && colIndex[key] < len(record) {
		if v := strings.TrimSpace(record[colIndex[key]]); v != "" {
			return v
		}
	}
	return defaultVal
}

func csvOptionalFieldPtr(record []string, colIndex map[string]int, key string) *string {
	if colIndex[key] >= 0 && colIndex[key] < len(record) {
		if v := strings.TrimSpace(record[colIndex[key]]); v != "" {
			return &v
		}
	}
	return nil
}

func parseAmount(s string, out *float64) (bool, error) {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, "$", "")
	s = strings.TrimSpace(s)

	val, err := json.Number(s).Float64()
	if err != nil {
		return false, err
	}
	*out = val
	return true, nil
}

func dateToOpenAPI(t time.Time) openapi_types.Date {
	return openapi_types.Date{Time: t}
}
