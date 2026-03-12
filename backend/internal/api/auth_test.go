package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamespsullivan/pennywise/internal/api"
	"github.com/jamespsullivan/pennywise/internal/db"
	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/dlq"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

var testSecret = []byte("test-secret-key-at-least-32-bytes-long")

var precomputedHash string

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("pennywise"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	precomputedHash = string(hash)
}

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	err = db.Migrate(database)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00001-0000-0000-0000-000000000001",
		"james@example.com",
		"James",
		precomputedHash,
	)
	require.NoError(t, err)

	return database
}

func setupRouter(t *testing.T) (*sql.DB, http.Handler) {
	t.Helper()
	database := setupTestDB(t)
	userRepo := queries.NewUserRepository(database)
	accountRepo := queries.NewAccountRepository(database)
	txnRepo := queries.NewTransactionRepository(database)
	assetRepo := queries.NewAssetRepository(database)
	goalRepo := queries.NewGoalRepository(database)
	goalContribRepo := queries.NewGoalContributionRepository(database)
	recurringRepo := queries.NewRecurringRepository(database)
	alertRepo := queries.NewAlertRepository(database)
	dashboardRepo := queries.NewDashboardRepository(database)
	auditRepo := queries.NewAuditLogRepository(database)
	dlqWriter := dlq.NewFailedRequestWriter(database)
	groupRepo := queries.NewTransactionGroupRepository(database)
	handler := api.NewAppHandler(userRepo, accountRepo, txnRepo, groupRepo, assetRepo, goalRepo, goalContribRepo, recurringRepo, alertRepo, dashboardRepo, auditRepo, dlqWriter, testSecret)
	handler.SetBcryptCost(bcrypt.MinCost)

	validator, err := middleware.Validation(api.OpenAPISpec, "/api/v1")
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(validator)

	authMiddleware := middleware.Auth(testSecret, api.CookieAuthScopes)

	return database, api.HandlerWithOptions(handler, api.ChiServerOptions{
		BaseRouter:  router,
		BaseURL:     "/api/v1",
		Middlewares: []api.MiddlewareFunc{authMiddleware},
	})
}

func TestPostAuthLogin_ValidCredentials(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"james@example.com","password":"pennywise"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "james@example.com", resp.User.Email)
	assert.Equal(t, "James", resp.User.Name)
	assert.NotEmpty(t, resp.User.Id)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.True(t, cookies[0].HttpOnly)
	assert.True(t, cookies[0].Secure)
	assert.Equal(t, http.SameSiteStrictMode, cookies[0].SameSite)
	assert.NotEmpty(t, cookies[0].Value)
}

func TestPostAuthLogin_WrongPassword(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"james@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp["error"]["code"])
}

func TestPostAuthLogin_NonexistentEmail(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"nobody@example.com","password":"pennywise"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPostAuthLogin_InvalidRequestBody(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `not json`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostAuthLogout_ClearsCookie(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	loginBody := `{"email":"james@example.com","password":"pennywise"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)
	tokenCookie := loginRec.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(tokenCookie)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.True(t, cookies[0].MaxAge < 0)
}

func TestGetAuthMe_WithValidToken(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	loginBody := `{"email":"james@example.com","password":"pennywise"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	tokenCookie := loginRec.Result().Cookies()[0]

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.AddCookie(tokenCookie)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)

	assert.Equal(t, http.StatusOK, meRec.Code)

	var resp api.UserResponse
	require.NoError(t, json.Unmarshal(meRec.Body.Bytes(), &resp))
	assert.Equal(t, "james@example.com", resp.Email)
	assert.Equal(t, "James", resp.Name)
}

func TestGetAuthMe_WithoutToken_Returns401(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestPostAuthRegister_Success(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"newuser@example.com","password":"securepassword123","name":"New User"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "newuser@example.com", resp.User.Email)
	assert.Equal(t, "New User", resp.User.Name)
	assert.NotEmpty(t, resp.User.Id)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "token", cookies[0].Name)
	assert.True(t, cookies[0].HttpOnly)
}

func TestPostAuthRegister_DuplicateEmail(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"james@example.com","password":"securepassword123","name":"Duplicate"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestPostAuthRegister_MaxUsersReached(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)

	for i := 2; i <= 10; i++ {
		_, err := database.ExecContext(context.Background(),
			`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("usr-max-%02d", i),
			fmt.Sprintf("user%d@example.com", i),
			fmt.Sprintf("User %d", i),
			precomputedHash,
		)
		require.NoError(t, err)
	}

	body := `{"email":"eleventh@example.com","password":"securepassword123","name":"Eleventh"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestPostAuthRegister_ShortPassword(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"short@example.com","password":"abc","name":"Short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostAuthRegister_AutoLoginAfterRegister(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"autologin@example.com","password":"securepassword123","name":"Auto"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	tokenCookie := rec.Result().Cookies()[0]
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.AddCookie(tokenCookie)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)

	assert.Equal(t, http.StatusOK, meRec.Code)
	var resp api.UserResponse
	require.NoError(t, json.Unmarshal(meRec.Body.Bytes(), &resp))
	assert.Equal(t, "autologin@example.com", resp.Email)
}

func TestPostAuthRegister_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostAuthRegister_MissingName_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"noname@example.com","password":"securepassword123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostAuthLogin_MissingPassword_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"james@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRequestID_PresentOnEveryResponse(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"email":"james@example.com","password":"pennywise"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}
