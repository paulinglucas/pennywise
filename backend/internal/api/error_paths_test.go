package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

var errMock = errors.New("mock error")

type mockUserRepo struct {
	getByEmailFn func(ctx context.Context, email string) (*models.User, error)
	getByIDFn    func(ctx context.Context, id string) (*models.User, error)
	countFn      func(ctx context.Context) (int, error)
	createFn     func(ctx context.Context, user *models.User) error
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &models.User{ID: id, Email: "test@test.com", Name: "Test"}, nil
}
func (m *mockUserRepo) CountUsers(ctx context.Context) (int, error) {
	if m.countFn != nil {
		return m.countFn(ctx)
	}
	return 1, nil
}
func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

type mockAccountRepo struct {
	listFn    func(ctx context.Context, userID string, page, perPage int) ([]models.Account, int, error)
	createFn  func(ctx context.Context, account *models.Account) error
	getByIDFn func(ctx context.Context, userID, id string) (*models.Account, error)
	updateFn  func(ctx context.Context, account *models.Account) (bool, error)
	deleteFn  func(ctx context.Context, userID, id string) (bool, error)
}

func (m *mockAccountRepo) List(ctx context.Context, userID string, page, perPage int) ([]models.Account, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockAccountRepo) Create(ctx context.Context, account *models.Account) error {
	if m.createFn != nil {
		return m.createFn(ctx, account)
	}
	return nil
}
func (m *mockAccountRepo) GetByID(ctx context.Context, userID, id string) (*models.Account, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockAccountRepo) Update(ctx context.Context, account *models.Account) (bool, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, account)
	}
	return true, nil
}
func (m *mockAccountRepo) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}

type mockTransactionRepo struct {
	listFn       func(ctx context.Context, userID string, filter queries.TransactionFilter, page, perPage int) ([]models.Transaction, int, error)
	createFn     func(ctx context.Context, txn *models.Transaction, tags []string) error
	getByIDFn    func(ctx context.Context, userID, id string) (*models.Transaction, error)
	updateFn     func(ctx context.Context, txn *models.Transaction, tags []string) (bool, error)
	deleteFn     func(ctx context.Context, userID, id string) (bool, error)
	bulkCreateFn func(ctx context.Context, txns []models.Transaction) (int, []queries.BulkCreateError)
	listCatsFn   func(ctx context.Context, userID string) ([]string, error)
}

func (m *mockTransactionRepo) List(ctx context.Context, userID string, filter queries.TransactionFilter, page, perPage int) ([]models.Transaction, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, filter, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockTransactionRepo) Create(ctx context.Context, txn *models.Transaction, tags []string) error {
	if m.createFn != nil {
		return m.createFn(ctx, txn, tags)
	}
	return nil
}
func (m *mockTransactionRepo) GetByID(ctx context.Context, userID, id string) (*models.Transaction, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockTransactionRepo) Update(ctx context.Context, txn *models.Transaction, tags []string) (bool, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, txn, tags)
	}
	return true, nil
}
func (m *mockTransactionRepo) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}
func (m *mockTransactionRepo) BulkCreate(ctx context.Context, txns []models.Transaction) (int, []queries.BulkCreateError) {
	if m.bulkCreateFn != nil {
		return m.bulkCreateFn(ctx, txns)
	}
	return len(txns), nil
}
func (m *mockTransactionRepo) ListCategories(ctx context.Context, userID string) ([]string, error) {
	if m.listCatsFn != nil {
		return m.listCatsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockTransactionRepo) BulkCategorize(_ context.Context, _ string, _ []queries.CategoryUpdate) (int, error) {
	return 0, nil
}

type mockAssetRepo struct {
	listFn             func(ctx context.Context, userID string, page, perPage int) ([]models.Asset, int, error)
	createFn           func(ctx context.Context, asset *models.Asset) error
	getByIDFn          func(ctx context.Context, userID, id string) (*models.Asset, error)
	updateFn           func(ctx context.Context, asset *models.Asset, prevValue float64) (bool, error)
	deleteFn           func(ctx context.Context, userID, id string) (bool, error)
	getHistoryFn       func(ctx context.Context, userID, assetID string, since *time.Time) ([]models.AssetHistory, error)
	getAllocationFn    func(ctx context.Context, userID string) ([]queries.AllocationRow, error)
	getAllocOverTimeFn func(ctx context.Context, userID string, since *time.Time) ([]queries.AllocationSnapshot, error)
	getLinkedFn        func(ctx context.Context, accountIDs []string) (map[string]queries.LinkedAccountRow, error)
}

func (m *mockAssetRepo) List(ctx context.Context, userID string, page, perPage int) ([]models.Asset, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockAssetRepo) Create(ctx context.Context, asset *models.Asset) error {
	if m.createFn != nil {
		return m.createFn(ctx, asset)
	}
	return nil
}
func (m *mockAssetRepo) GetByID(ctx context.Context, userID, id string) (*models.Asset, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockAssetRepo) Update(ctx context.Context, asset *models.Asset, prevValue float64) (bool, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, asset, prevValue)
	}
	return true, nil
}
func (m *mockAssetRepo) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}
func (m *mockAssetRepo) GetHistory(ctx context.Context, userID, assetID string, since *time.Time) ([]models.AssetHistory, error) {
	if m.getHistoryFn != nil {
		return m.getHistoryFn(ctx, userID, assetID, since)
	}
	return nil, nil
}
func (m *mockAssetRepo) GetAllocation(ctx context.Context, userID string) ([]queries.AllocationRow, error) {
	if m.getAllocationFn != nil {
		return m.getAllocationFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockAssetRepo) GetAllocationOverTime(ctx context.Context, userID string, since *time.Time) ([]queries.AllocationSnapshot, error) {
	if m.getAllocOverTimeFn != nil {
		return m.getAllocOverTimeFn(ctx, userID, since)
	}
	return nil, nil
}
func (m *mockAssetRepo) GetLinkedAccounts(ctx context.Context, accountIDs []string) (map[string]queries.LinkedAccountRow, error) {
	if m.getLinkedFn != nil {
		return m.getLinkedFn(ctx, accountIDs)
	}
	return map[string]queries.LinkedAccountRow{}, nil
}

type mockGoalRepo struct {
	listFn     func(ctx context.Context, userID string, page, perPage int) ([]models.Goal, int, error)
	createFn   func(ctx context.Context, goal *models.Goal) error
	getByIDFn  func(ctx context.Context, userID, id string) (*models.Goal, error)
	updateFn   func(ctx context.Context, goal *models.Goal) (bool, error)
	deleteFn   func(ctx context.Context, userID, id string) (bool, error)
	reorderFn  func(ctx context.Context, userID string, rankings []queries.GoalRanking) error
	nextRankFn func(ctx context.Context, userID string) (int, error)
}

func (m *mockGoalRepo) List(ctx context.Context, userID string, page, perPage int) ([]models.Goal, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockGoalRepo) Create(ctx context.Context, goal *models.Goal) error {
	if m.createFn != nil {
		return m.createFn(ctx, goal)
	}
	return nil
}
func (m *mockGoalRepo) GetByID(ctx context.Context, userID, id string) (*models.Goal, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockGoalRepo) Update(ctx context.Context, goal *models.Goal) (bool, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, goal)
	}
	return true, nil
}
func (m *mockGoalRepo) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}
func (m *mockGoalRepo) Reorder(ctx context.Context, userID string, rankings []queries.GoalRanking) error {
	if m.reorderFn != nil {
		return m.reorderFn(ctx, userID, rankings)
	}
	return nil
}
func (m *mockGoalRepo) NextPriorityRank(ctx context.Context, userID string) (int, error) {
	if m.nextRankFn != nil {
		return m.nextRankFn(ctx, userID)
	}
	return 1, nil
}

type mockGoalContribRepo struct {
	createFn  func(ctx context.Context, contrib *models.GoalContribution) error
	listFn    func(ctx context.Context, userID, goalID string, page, perPage int) ([]models.GoalContribution, int, error)
	getByIDFn func(ctx context.Context, userID, id string) (*models.GoalContribution, error)
	deleteFn  func(ctx context.Context, userID, id string) (bool, error)
}

func (m *mockGoalContribRepo) Create(ctx context.Context, contrib *models.GoalContribution) error {
	if m.createFn != nil {
		return m.createFn(ctx, contrib)
	}
	return nil
}
func (m *mockGoalContribRepo) ListByGoal(ctx context.Context, userID, goalID string, page, perPage int) ([]models.GoalContribution, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, goalID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockGoalContribRepo) GetByID(ctx context.Context, userID, id string) (*models.GoalContribution, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockGoalContribRepo) Delete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}

type mockRecurringRepo struct {
	listFn    func(ctx context.Context, userID string, page, perPage int) ([]models.RecurringTransaction, int, error)
	createFn  func(ctx context.Context, rec *models.RecurringTransaction) error
	getByIDFn func(ctx context.Context, userID, id string) (*models.RecurringTransaction, error)
	updateFn  func(ctx context.Context, rec *models.RecurringTransaction) (bool, error)
	deleteFn  func(ctx context.Context, userID, id string) (bool, error)
}

func (m *mockRecurringRepo) List(ctx context.Context, userID string, page, perPage int) ([]models.RecurringTransaction, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockRecurringRepo) Create(ctx context.Context, rec *models.RecurringTransaction) error {
	if m.createFn != nil {
		return m.createFn(ctx, rec)
	}
	return nil
}
func (m *mockRecurringRepo) GetByID(ctx context.Context, userID, id string) (*models.RecurringTransaction, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockRecurringRepo) Update(ctx context.Context, rec *models.RecurringTransaction) (bool, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, rec)
	}
	return true, nil
}
func (m *mockRecurringRepo) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}

type mockAlertRepo struct {
	listFn     func(ctx context.Context, userID string, page, perPage int) ([]models.Alert, int, error)
	markReadFn func(ctx context.Context, userID, id string) (bool, error)
}

func (m *mockAlertRepo) List(ctx context.Context, userID string, page, perPage int) ([]models.Alert, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockAlertRepo) MarkRead(ctx context.Context, userID, id string) (bool, error) {
	if m.markReadFn != nil {
		return m.markReadFn(ctx, userID, id)
	}
	return true, nil
}

type mockGroupRepo struct {
	createFn  func(ctx context.Context, group *models.TransactionGroup) error
	getByIDFn func(ctx context.Context, userID, id string) (*models.TransactionGroup, error)
	updateFn  func(ctx context.Context, group *models.TransactionGroup) (bool, error)
	deleteFn  func(ctx context.Context, userID, id string) (bool, error)
	listFn    func(ctx context.Context, userID string, page, perPage int) ([]models.TransactionGroup, int, error)
	membersFn func(ctx context.Context, userID, groupID string) ([]models.Transaction, error)
}

func (m *mockGroupRepo) Create(ctx context.Context, group *models.TransactionGroup) error {
	if m.createFn != nil {
		return m.createFn(ctx, group)
	}
	return nil
}
func (m *mockGroupRepo) GetByID(ctx context.Context, userID, id string) (*models.TransactionGroup, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, userID, id)
	}
	return nil, nil
}
func (m *mockGroupRepo) Update(ctx context.Context, group *models.TransactionGroup) (bool, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, group)
	}
	return true, nil
}
func (m *mockGroupRepo) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, userID, id)
	}
	return true, nil
}
func (m *mockGroupRepo) List(ctx context.Context, userID string, page, perPage int) ([]models.TransactionGroup, int, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, perPage)
	}
	return nil, 0, nil
}
func (m *mockGroupRepo) ListMembers(ctx context.Context, userID, groupID string) ([]models.Transaction, error) {
	if m.membersFn != nil {
		return m.membersFn(ctx, userID, groupID)
	}
	return nil, nil
}

type mockDashboardRepo struct {
	getNetWorthFn  func(ctx context.Context, userID string) (queries.NetWorthResult, error)
	getCashFlowFn  func(ctx context.Context, userID string, now time.Time) (float64, error)
	getSpendingFn  func(ctx context.Context, userID string, since time.Time, until time.Time) ([]queries.SpendingRow, error)
	getDebtsFn     func(ctx context.Context, userID string, now time.Time) ([]queries.DebtRow, error)
	getNWHistoryFn func(ctx context.Context, userID string, since time.Time, includeSinceDate bool) ([]queries.NetWorthDataPoint, error)
	pingFn         func(ctx context.Context) error
}

func (m *mockDashboardRepo) GetNetWorth(ctx context.Context, userID string) (queries.NetWorthResult, error) {
	if m.getNetWorthFn != nil {
		return m.getNetWorthFn(ctx, userID)
	}
	return queries.NetWorthResult{}, nil
}
func (m *mockDashboardRepo) GetCashFlowThisMonth(ctx context.Context, userID string, now time.Time) (float64, error) {
	if m.getCashFlowFn != nil {
		return m.getCashFlowFn(ctx, userID, now)
	}
	return 0, nil
}
func (m *mockDashboardRepo) GetSpendingByCategory(ctx context.Context, userID string, since time.Time, until time.Time) ([]queries.SpendingRow, error) {
	if m.getSpendingFn != nil {
		return m.getSpendingFn(ctx, userID, since, until)
	}
	return nil, nil
}
func (m *mockDashboardRepo) GetDebtsSummary(ctx context.Context, userID string, now time.Time) ([]queries.DebtRow, error) {
	if m.getDebtsFn != nil {
		return m.getDebtsFn(ctx, userID, now)
	}
	return nil, nil
}
func (m *mockDashboardRepo) GetNetWorthHistory(ctx context.Context, userID string, since time.Time, includeSinceDate bool) ([]queries.NetWorthDataPoint, error) {
	if m.getNWHistoryFn != nil {
		return m.getNWHistoryFn(ctx, userID, since, includeSinceDate)
	}
	return nil, nil
}
func (m *mockDashboardRepo) PingDB(ctx context.Context) error {
	if m.pingFn != nil {
		return m.pingFn(ctx)
	}
	return nil
}

type mockAuditWriter struct{}

func (m *mockAuditWriter) Record(_ context.Context, _ *models.AuditLog) error { return nil }

type mockDLQWriter struct{}

func (m *mockDLQWriter) Write(_ context.Context, _ *models.FailedRequest) error { return nil }

type mockHandlerConfig struct {
	users     *mockUserRepo
	accounts  *mockAccountRepo
	txns      *mockTransactionRepo
	groups    *mockGroupRepo
	assets    *mockAssetRepo
	goals     *mockGoalRepo
	contribs  *mockGoalContribRepo
	recurring *mockRecurringRepo
	alerts    *mockAlertRepo
	dashboard *mockDashboardRepo
}

func buildMockHandler(cfg mockHandlerConfig) *api.AppHandler {
	users := cfg.users
	if users == nil {
		users = &mockUserRepo{}
	}
	accounts := cfg.accounts
	if accounts == nil {
		accounts = &mockAccountRepo{}
	}
	txns := cfg.txns
	if txns == nil {
		txns = &mockTransactionRepo{}
	}
	groups := cfg.groups
	if groups == nil {
		groups = &mockGroupRepo{}
	}
	assets := cfg.assets
	if assets == nil {
		assets = &mockAssetRepo{}
	}
	goals := cfg.goals
	if goals == nil {
		goals = &mockGoalRepo{}
	}
	contribs := cfg.contribs
	if contribs == nil {
		contribs = &mockGoalContribRepo{}
	}
	recurring := cfg.recurring
	if recurring == nil {
		recurring = &mockRecurringRepo{}
	}
	alerts := cfg.alerts
	if alerts == nil {
		alerts = &mockAlertRepo{}
	}
	dash := cfg.dashboard
	if dash == nil {
		dash = &mockDashboardRepo{}
	}
	return api.NewAppHandler(users, accounts, txns, groups, assets, goals, contribs, recurring, alerts, dash, &mockAuditWriter{}, &mockDLQWriter{}, testSecret)
}

func mockRequest(method, path, body string) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	ctx := middleware.WithUserID(req.Context(), "u1")
	return req.WithContext(ctx)
}

func TestListAccounts_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Account, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListAccounts(rec, mockRequest(http.MethodGet, "/", ""), api.ListAccountsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateAccount_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			createFn: func(_ context.Context, _ *models.Account) error {
				return errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"Test","institution":"Bank","account_type":"checking"}`
	h.CreateAccount(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateAccount_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Account, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"Test","institution":"Bank","account_type":"checking"}`
	h.CreateAccount(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateAccount_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	acct := &models.Account{ID: "a1", UserID: "u1", Name: "Checking", Institution: "Bank", AccountType: "checking"}
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Account, error) {
				return acct, nil
			},
			updateFn: func(_ context.Context, _ *models.Account) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"Updated"}`
	h.UpdateAccount(rec, mockRequest(http.MethodPut, "/", body), api.ParseID("a1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListTransactions_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			listFn: func(_ context.Context, _ string, _ queries.TransactionFilter, _, _ int) ([]models.Transaction, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListTransactions(rec, mockRequest(http.MethodGet, "/", ""), api.ListTransactionsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateTransaction_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			createFn: func(_ context.Context, _ *models.Transaction, _ []string) error {
				return errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"type":"expense","category":"food","amount":10,"date":"2025-06-15","account_id":"a0000001-0000-0000-0000-000000000001"}`
	h.CreateTransaction(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateTransaction_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"type":"expense","category":"food","amount":10,"date":"2025-06-15","account_id":"a0000001-0000-0000-0000-000000000001"}`
	h.CreateTransaction(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTransaction_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetTransaction(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("t1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransaction_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateTransaction(rec, mockRequest(http.MethodPut, "/", `{"category":"new"}`), api.ParseID("t1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransaction_UpdateError_Returns500(t *testing.T) {
	t.Parallel()
	txn := &models.Transaction{ID: "t1", UserID: "u1", AccountID: "a1", Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: time.Now()}
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return txn, nil
			},
			updateFn: func(_ context.Context, _ *models.Transaction, _ []string) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateTransaction(rec, mockRequest(http.MethodPut, "/", `{"category":"new"}`), api.ParseID("t1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransaction_RetrievalAfterUpdateError_Returns500(t *testing.T) {
	t.Parallel()
	callCount := 0
	txn := &models.Transaction{ID: "t1", UserID: "u1", AccountID: "a1", Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: time.Now()}
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				callCount++
				if callCount == 1 {
					return txn, nil
				}
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateTransaction(rec, mockRequest(http.MethodPut, "/", `{"category":"new"}`), api.ParseID("t1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteTransaction_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteTransaction(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("t1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteTransaction_SoftDeleteError_Returns500(t *testing.T) {
	t.Parallel()
	txn := &models.Transaction{ID: "t1", UserID: "u1", AccountID: "a1", Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: time.Now()}
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return txn, nil
			},
			deleteFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteTransaction(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("t1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteTransaction_SoftDeleteNotFound_Returns404(t *testing.T) {
	t.Parallel()
	txn := &models.Transaction{ID: "t1", UserID: "u1", AccountID: "a1", Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: time.Now()}
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return txn, nil
			},
			deleteFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteTransaction(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("t1"))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListCategories_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			listCatsFn: func(_ context.Context, _ string) ([]string, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListCategories(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListAssets_ListError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Asset, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListAssets(rec, mockRequest(http.MethodGet, "/", ""), api.ListAssetsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListAssets_AllocationError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getAllocationFn: func(_ context.Context, _ string) ([]queries.AllocationRow, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListAssets(rec, mockRequest(http.MethodGet, "/", ""), api.ListAssetsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListAssets_LinkedAccountsError_Returns500(t *testing.T) {
	t.Parallel()
	acctID := "a1"
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Asset, int, error) {
				return []models.Asset{{ID: "x1", AccountID: &acctID}}, 1, nil
			},
			getLinkedFn: func(_ context.Context, _ []string) (map[string]queries.LinkedAccountRow, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListAssets(rec, mockRequest(http.MethodGet, "/", ""), api.ListAssetsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateAsset_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			createFn: func(_ context.Context, _ *models.Asset) error {
				return errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"Test Asset","asset_type":"liquid","current_value":1000}`
	h.CreateAsset(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateAsset_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"Test Asset","asset_type":"liquid","current_value":1000}`
	h.CreateAsset(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetAsset_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetAsset(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetAsset_HistoryError_Returns500(t *testing.T) {
	t.Parallel()
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD"}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return asset, nil
			},
			getHistoryFn: func(_ context.Context, _, _ string, _ *time.Time) ([]models.AssetHistory, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetAsset(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetAsset_LinkedAccountError_Returns500(t *testing.T) {
	t.Parallel()
	acctID := "a1"
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD", AccountID: &acctID}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return asset, nil
			},
			getLinkedFn: func(_ context.Context, _ []string) (map[string]queries.LinkedAccountRow, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetAsset(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateAsset_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateAsset(rec, mockRequest(http.MethodPut, "/", `{"name":"new"}`), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateAsset_UpdateError_Returns500(t *testing.T) {
	t.Parallel()
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD", CurrentValue: 100}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return asset, nil
			},
			updateFn: func(_ context.Context, _ *models.Asset, _ float64) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateAsset(rec, mockRequest(http.MethodPut, "/", `{"name":"new"}`), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteAsset_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteAsset(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteAsset_SoftDeleteError_Returns500(t *testing.T) {
	t.Parallel()
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD"}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return asset, nil
			},
			deleteFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteAsset(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteAsset_NotFoundAfterSoftDelete_Returns404(t *testing.T) {
	t.Parallel()
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD"}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return asset, nil
			},
			deleteFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteAsset(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("x1"))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetAssetHistory_Error_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getHistoryFn: func(_ context.Context, _, _ string, _ *time.Time) ([]models.AssetHistory, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetAssetHistory(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("x1"), api.GetAssetHistoryParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetAssetAllocation_Error_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getAllocOverTimeFn: func(_ context.Context, _ string, _ *time.Time) ([]queries.AllocationSnapshot, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetAssetAllocation(rec, mockRequest(http.MethodGet, "/", ""), api.GetAssetAllocationParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListTransactionGroups_ListError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.TransactionGroup, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListTransactionGroups(rec, mockRequest(http.MethodGet, "/", ""), api.ListTransactionGroupsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListTransactionGroups_ListMembersError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.TransactionGroup, int, error) {
				return []models.TransactionGroup{{ID: "g1", UserID: "u1", Name: "G"}}, 1, nil
			},
			membersFn: func(_ context.Context, _, _ string) ([]models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListTransactionGroups(rec, mockRequest(http.MethodGet, "/", ""), api.ListTransactionGroupsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateTransactionGroup_CreateError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			createFn: func(_ context.Context, _ *models.TransactionGroup) error {
				return errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"G","members":[{"type":"expense","category":"a","amount":10,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"},{"type":"expense","category":"b","amount":20,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"}]}`
	h.CreateTransactionGroup(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateTransactionGroup_MemberCreateError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			createFn: func(_ context.Context, _ *models.Transaction, _ []string) error {
				return errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"G","members":[{"type":"expense","category":"a","amount":10,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"},{"type":"expense","category":"b","amount":20,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"}]}`
	h.CreateTransactionGroup(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateTransactionGroup_MemberGetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"G","members":[{"type":"expense","category":"a","amount":10,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"},{"type":"expense","category":"b","amount":20,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"}]}`
	h.CreateTransactionGroup(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreateTransactionGroup_GroupGetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	txn := &models.Transaction{ID: "t1", UserID: "u1", AccountID: "a1", Type: "expense", Category: "x", Amount: 10, Currency: "USD", Date: time.Now()}
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Transaction, error) {
				return txn, nil
			},
		},
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"name":"G","members":[{"type":"expense","category":"a","amount":10,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"},{"type":"expense","category":"b","amount":20,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"}]}`
	h.CreateTransactionGroup(rec, mockRequest(http.MethodPost, "/", body))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTransactionGroup_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetTransactionGroup(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTransactionGroup_MembersError_Returns500(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return group, nil
			},
			membersFn: func(_ context.Context, _, _ string) ([]models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetTransactionGroup(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransactionGroup_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateTransactionGroup(rec, mockRequest(http.MethodPut, "/", `{"name":"new"}`), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransactionGroup_UpdateError_Returns500(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return group, nil
			},
			updateFn: func(_ context.Context, _ *models.TransactionGroup) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateTransactionGroup(rec, mockRequest(http.MethodPut, "/", `{"name":"new"}`), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteTransactionGroup_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteTransactionGroup(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteTransactionGroup_SoftDeleteError_Returns500(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return group, nil
			},
			deleteFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteTransactionGroup(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteTransactionGroup_SoftDeleteNotFound_Returns404(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return group, nil
			},
			deleteFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.DeleteTransactionGroup(rec, mockRequest(http.MethodDelete, "/", ""), api.ParseID("g1"))
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListAlerts_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		alerts: &mockAlertRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Alert, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListAlerts(rec, mockRequest(http.MethodGet, "/", ""), api.ListAlertsParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMarkAlertRead_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		alerts: &mockAlertRepo{
			markReadFn: func(_ context.Context, _, _ string) (bool, error) {
				return false, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.MarkAlertRead(rec, mockRequest(http.MethodPost, "/", ""), api.ParseID("al1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetDashboard_NetWorthError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			getNetWorthFn: func(_ context.Context, _ string) (queries.NetWorthResult, error) {
				return queries.NetWorthResult{}, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetDashboard(rec, mockRequest(http.MethodGet, "/", ""), api.GetDashboardParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetDashboard_CashFlowError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			getCashFlowFn: func(_ context.Context, _ string, _ time.Time) (float64, error) {
				return 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetDashboard(rec, mockRequest(http.MethodGet, "/", ""), api.GetDashboardParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetDashboard_SpendingError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			getSpendingFn: func(_ context.Context, _ string, _, _ time.Time) ([]queries.SpendingRow, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetDashboard(rec, mockRequest(http.MethodGet, "/", ""), api.GetDashboardParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetDashboard_DebtsError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			getDebtsFn: func(_ context.Context, _ string, _ time.Time) ([]queries.DebtRow, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetDashboard(rec, mockRequest(http.MethodGet, "/", ""), api.GetDashboardParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_UserError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		users: &mockUserRepo{
			getByIDFn: func(_ context.Context, _ string) (*models.User, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_AccountsError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Account, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_TransactionsError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			listFn: func(_ context.Context, _ string, _ queries.TransactionFilter, _, _ int) ([]models.Transaction, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_AssetsError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Asset, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_GoalsError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		goals: &mockGoalRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Goal, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_RecurringError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		recurring: &mockRecurringRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.RecurringTransaction, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportData_AlertsError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		alerts: &mockAlertRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Alert, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportData(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestExportCsv_RepoError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		txns: &mockTransactionRepo{
			listFn: func(_ context.Context, _ string, _ queries.TransactionFilter, _, _ int) ([]models.Transaction, int, error) {
				return nil, 0, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ExportCsv(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestComputeProjection_NetWorthError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			getNetWorthFn: func(_ context.Context, _ string) (queries.NetWorthResult, error) {
				return queries.NetWorthResult{}, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ComputeProjection(rec, mockRequest(http.MethodPost, "/", `{"years_to_project":10}`))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetHealth_DBError_Returns503(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			pingFn: func(_ context.Context) error {
				return errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetHealth(rec, mockRequest(http.MethodGet, "/", ""))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp api.HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.Unhealthy, resp.Status)
	assert.Equal(t, api.Disconnected, resp.Database)
}

func TestGetNetWorthHistory_Error_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		dashboard: &mockDashboardRepo{
			getNWHistoryFn: func(_ context.Context, _ string, _ time.Time, _ bool) ([]queries.NetWorthDataPoint, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.GetNetWorthHistory(rec, mockRequest(http.MethodGet, "/", ""), api.GetNetWorthHistoryParams{})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestPaginationDefaults_PerPageCapped(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	pp := 200
	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transactions?per_page=%d", pp), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 100, resp.Pagination.PerPage)
}

func TestImportTransactions_CSVRowFewerColumnsThanHeader(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount,notes,currency\n2025-06-15,expense\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.ImportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Imported)
	assert.NotEmpty(t, resp.Errors)
}

func TestImportTransactions_CSVInvalidAmount(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount\n2025-06-15,expense,food,not-a-number\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.ImportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Imported)
	assert.NotEmpty(t, resp.Errors)
}

func TestImportTransactions_CSVInvalidType(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount\n2025-06-15,transfer,food,42.99\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.ImportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Imported)
	assert.NotEmpty(t, resp.Errors)
}

func TestImportTransactions_CSVEmptyCategory(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount\n2025-06-15,expense,,42.99\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.ImportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 0, resp.Imported)
	assert.NotEmpty(t, resp.Errors)
}

func TestUpdateTransactionGroup_SyncMembersError_Returns500(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return group, nil
			},
			membersFn: func(_ context.Context, _, _ string) ([]models.Transaction, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"members":[{"type":"expense","category":"a","amount":10,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"}]}`
	h.UpdateTransactionGroup(rec, mockRequest(http.MethodPut, "/", body), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransactionGroup_FinalMembersListError_Returns500(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	callCount := 0
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				return group, nil
			},
			membersFn: func(_ context.Context, _, _ string) ([]models.Transaction, error) {
				callCount++
				if callCount > 1 {
					return nil, errMock
				}
				return nil, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	body := `{"members":[{"type":"expense","category":"a","amount":10,"date":"2025-01-01","account_id":"a0000001-0000-0000-0000-000000000001"}]}`
	h.UpdateTransactionGroup(rec, mockRequest(http.MethodPut, "/", body), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateTransactionGroup_FinalGetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	group := &models.TransactionGroup{ID: "g1", UserID: "u1", Name: "G"}
	getCount := 0
	h := buildMockHandler(mockHandlerConfig{
		groups: &mockGroupRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.TransactionGroup, error) {
				getCount++
				if getCount > 1 {
					return nil, errMock
				}
				return group, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateTransactionGroup(rec, mockRequest(http.MethodPut, "/", `{"name":"new"}`), api.ParseID("g1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateAsset_RetrievalAfterUpdateError_Returns500(t *testing.T) {
	t.Parallel()
	callCount := 0
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD", CurrentValue: 100}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				callCount++
				if callCount > 1 {
					return nil, errMock
				}
				return asset, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateAsset(rec, mockRequest(http.MethodPut, "/", `{"name":"new"}`), api.ParseID("x1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListTransactions_FilterByGroupIDParam(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	groupBody := fmt.Sprintf(`{
		"name": "Test Group",
		"members": [
			{"type":"expense","category":"food","amount":10,"date":"2026-03-08","account_id":"%s"},
			{"type":"expense","category":"drinks","amount":5,"date":"2026-03-08","account_id":"%s"}
		]
	}`, txnTestAccountID, txnTestAccountID)
	groupReq := authedRequest(http.MethodPost, "/api/v1/transaction-groups", groupBody, cookie)
	groupRec := httptest.NewRecorder()
	router.ServeHTTP(groupRec, groupReq)
	require.Equal(t, http.StatusCreated, groupRec.Code)

	var group api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(groupRec.Body.Bytes(), &group))

	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transactions?group_id=%s", group.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	for _, txn := range resp.Data {
		require.NotNil(t, txn.GroupId)
		assert.Equal(t, group.Id.String(), txn.GroupId.String())
	}
}

func TestSyncGroupMembers_UpdateExistingMember(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	groupBody := fmt.Sprintf(`{
		"name": "Sync Test Group",
		"members": [
			{"type":"expense","category":"food","amount":10,"date":"2026-03-08","account_id":"%s"},
			{"type":"expense","category":"drinks","amount":5,"date":"2026-03-08","account_id":"%s"}
		]
	}`, txnTestAccountID, txnTestAccountID)
	groupReq := authedRequest(http.MethodPost, "/api/v1/transaction-groups", groupBody, cookie)
	groupRec := httptest.NewRecorder()
	router.ServeHTTP(groupRec, groupReq)
	require.Equal(t, http.StatusCreated, groupRec.Code)

	var group api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(groupRec.Body.Bytes(), &group))

	member1ID := group.Members[0].Id
	member2ID := group.Members[1].Id

	updateBody := fmt.Sprintf(`{
		"members": [
			{"id":"%s","type":"expense","category":"food-updated","amount":25,"date":"2026-03-09","account_id":"%s"},
			{"id":"%s","type":"expense","category":"drinks","amount":5,"date":"2026-03-08","account_id":"%s"}
		]
	}`, member1ID, txnTestAccountID, member2ID, txnTestAccountID)

	updateReq := authedRequest(http.MethodPut, "/api/v1/transaction-groups/"+group.Id.String(), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var updated api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &updated))
	assert.Equal(t, "Sync Test Group", updated.Name)
}

func TestSyncGroupMembers_DeleteMember(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	groupBody := fmt.Sprintf(`{
		"name": "Delete Member Group",
		"members": [
			{"type":"expense","category":"a","amount":10,"date":"2026-03-08","account_id":"%s"},
			{"type":"expense","category":"b","amount":20,"date":"2026-03-08","account_id":"%s"},
			{"type":"expense","category":"c","amount":30,"date":"2026-03-08","account_id":"%s"}
		]
	}`, txnTestAccountID, txnTestAccountID, txnTestAccountID)
	groupReq := authedRequest(http.MethodPost, "/api/v1/transaction-groups", groupBody, cookie)
	groupRec := httptest.NewRecorder()
	router.ServeHTTP(groupRec, groupReq)
	require.Equal(t, http.StatusCreated, groupRec.Code)

	var group api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(groupRec.Body.Bytes(), &group))

	member1ID := group.Members[0].Id
	member2ID := group.Members[1].Id

	updateBody := fmt.Sprintf(`{
		"members": [
			{"id":"%s","type":"expense","category":"a","amount":10,"date":"2026-03-08","account_id":"%s"},
			{"id":"%s","type":"expense","category":"b","amount":20,"date":"2026-03-08","account_id":"%s"}
		]
	}`, member1ID, txnTestAccountID, member2ID, txnTestAccountID)

	updateReq := authedRequest(http.MethodPut, "/api/v1/transaction-groups/"+group.Id.String(), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)
}

func TestSyncGroupMembers_AddNewMember(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	groupBody := fmt.Sprintf(`{
		"name": "Add Member Group",
		"members": [
			{"type":"expense","category":"a","amount":10,"date":"2026-03-08","account_id":"%s"},
			{"type":"expense","category":"b","amount":20,"date":"2026-03-08","account_id":"%s"}
		]
	}`, txnTestAccountID, txnTestAccountID)
	groupReq := authedRequest(http.MethodPost, "/api/v1/transaction-groups", groupBody, cookie)
	groupRec := httptest.NewRecorder()
	router.ServeHTTP(groupRec, groupReq)
	require.Equal(t, http.StatusCreated, groupRec.Code)

	var group api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(groupRec.Body.Bytes(), &group))

	member1ID := group.Members[0].Id
	member2ID := group.Members[1].Id

	updateBody := fmt.Sprintf(`{
		"members": [
			{"id":"%s","type":"expense","category":"a","amount":10,"date":"2026-03-08","account_id":"%s"},
			{"id":"%s","type":"expense","category":"b","amount":20,"date":"2026-03-08","account_id":"%s"},
			{"type":"expense","category":"c-new","amount":30,"date":"2026-03-08","account_id":"%s"}
		]
	}`, member1ID, txnTestAccountID, member2ID, txnTestAccountID, txnTestAccountID)

	updateReq := authedRequest(http.MethodPut, "/api/v1/transaction-groups/"+group.Id.String(), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)
}

func TestUpdateAccount_GetByIDError_Returns500(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Account, error) {
				return nil, errMock
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateAccount(rec, mockRequest(http.MethodPut, "/", `{"name":"X"}`), api.ParseID("a1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateAccount_RetrievalAfterUpdateError_Returns500(t *testing.T) {
	t.Parallel()
	acct := &models.Account{ID: "a1", UserID: "u1", Name: "Checking", Institution: "Bank", AccountType: "checking"}
	callCount := 0
	h := buildMockHandler(mockHandlerConfig{
		accounts: &mockAccountRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Account, error) {
				callCount++
				if callCount > 1 {
					return nil, errMock
				}
				return acct, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.UpdateAccount(rec, mockRequest(http.MethodPut, "/", `{"name":"X"}`), api.ParseID("a1"))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAttachLinkedAccount_NoAccountID(t *testing.T) {
	t.Parallel()
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			listFn: func(_ context.Context, _ string, _, _ int) ([]models.Asset, int, error) {
				return []models.Asset{{ID: "x1", Name: "No Linked", AssetType: "liquid", Currency: "USD"}}, 1, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	h.ListAssets(rec, mockRequest(http.MethodGet, "/", ""), api.ListAssetsParams{})
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetAssetHistory_DefaultPeriodFallback(t *testing.T) {
	t.Parallel()
	calledWithPeriod := false
	asset := &models.Asset{ID: "x1", UserID: "u1", Name: "A", AssetType: "liquid", Currency: "USD"}
	h := buildMockHandler(mockHandlerConfig{
		assets: &mockAssetRepo{
			getByIDFn: func(_ context.Context, _, _ string) (*models.Asset, error) {
				return asset, nil
			},
			getHistoryFn: func(_ context.Context, _, _ string, since *time.Time) ([]models.AssetHistory, error) {
				if since != nil {
					calledWithPeriod = true
				}
				return []models.AssetHistory{}, nil
			},
		},
	})
	rec := httptest.NewRecorder()
	invalid := api.GetAssetHistoryParamsPeriod("invalid")
	h.GetAssetHistory(rec, mockRequest(http.MethodGet, "/", ""), api.ParseID("x1"), api.GetAssetHistoryParams{Period: &invalid})
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, calledWithPeriod)
}
