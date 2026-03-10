package simplefin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/db"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

var testSecret = []byte("test-jwt-secret")
var testEncKey = pennywisecrypto.DeriveKey("test-encryption-key")

func setupHandlerTest(t *testing.T) (*Handler, *SQLiteSimplefinRepository, *httptest.Server) {
	t.Helper()

	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	_, err = database.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash) VALUES ('u1', 'test@test.com', 'Test', 'hash')`)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type) VALUES
		('a1', 'u1', 'Checking', 'My Bank', 'checking'),
		('a2', 'u1', 'Savings', 'My Bank', 'savings')`)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value) VALUES
		('asset1', 'u1', 'a1', 'Checking Balance', 'liquid', 1000.00),
		('asset2', 'u1', 'a2', 'Savings Balance', 'liquid', 5000.00)`)
	require.NoError(t, err)

	sfinServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"errors": [],
			"accounts": [
				{"id": "sfin-001", "name": "Checking", "currency": "USD", "balance": "1500.00", "balance-date": 1709856000, "org": {"name": "My Bank", "id": "mybank", "domain": "mybank.com"}},
				{"id": "sfin-002", "name": "Savings", "currency": "USD", "balance": "5200.50", "balance-date": 1709856000, "org": {"name": "My Bank", "id": "mybank", "domain": "mybank.com"}}
			]
		}`))
	}))
	t.Cleanup(sfinServer.Close)

	repo := NewSimplefinRepository(database)
	client := NewClient(nil)
	syncSvc := NewSyncService(client, repo, testEncKey)
	handler := NewHandler(repo, client, syncSvc, testEncKey)

	return handler, repo, sfinServer
}

func authedRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()

	var reqBody bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&reqBody).Encode(body))
	}

	req := httptest.NewRequest(method, path, &reqBody)
	req.Header.Set("Content-Type", "application/json")

	claims := &middleware.Claims{
		UserID: "u1",
		Email:  "test@test.com",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "pennywise",
		},
	}
	tokenStr, err := middleware.SignToken(claims, testSecret)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})

	return req
}

func TestSetupAndStatus(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	claimServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("https://testuser:testpass@bridge.example.com/simplefin"))
	}))
	defer claimServer.Close()

	token := base64.StdEncoding.EncodeToString([]byte(claimServer.URL))
	req := authedRequest(t, http.MethodPost, "/setup", setupRequest{SetupToken: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = authedRequest(t, http.MethodGet, "/status", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var status map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&status))
	assert.Equal(t, true, status["connected"])
}

func TestStatusNotConnected(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	req := authedRequest(t, http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var status statusResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&status))
	assert.False(t, status.Connected)
}

func TestLinkAndListAccounts(t *testing.T) {
	handler, repo, sfinServer := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	accessURL := testAccessURL(sfinServer.URL)
	encrypted, err := pennywisecrypto.Encrypt(testEncKey, accessURL)
	require.NoError(t, err)
	require.NoError(t, repo.SaveConnection(context.Background(), "u1", encrypted))

	req := authedRequest(t, http.MethodPost, "/link", linkRequest{AccountID: "a1", SimplefinID: "sfin-001"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	req = authedRequest(t, http.MethodGet, "/status", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var status map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&status))
	linked := status["linked_accounts"].([]interface{})
	assert.Len(t, linked, 1)
}

func TestUnlinkAccount(t *testing.T) {
	handler, repo, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	require.NoError(t, repo.LinkAccount(context.Background(), "u1", "a1", "sfin-001"))

	req := authedRequest(t, http.MethodDelete, "/link/a1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTriggerSync(t *testing.T) {
	handler, repo, sfinServer := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	accessURL := testAccessURL(sfinServer.URL)
	encrypted, err := pennywisecrypto.Encrypt(testEncKey, accessURL)
	require.NoError(t, err)
	require.NoError(t, repo.SaveConnection(context.Background(), "u1", encrypted))
	require.NoError(t, repo.LinkAccount(context.Background(), "u1", "a1", "sfin-001"))
	require.NoError(t, repo.LinkAccount(context.Background(), "u1", "a2", "sfin-002"))

	req := authedRequest(t, http.MethodPost, "/sync", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp syncResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, 2, resp.Updated)
	assert.Equal(t, 0, resp.Errors)
}

func TestDisconnect(t *testing.T) {
	handler, repo, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	require.NoError(t, repo.SaveConnection(context.Background(), "u1", "encrypted-url"))
	require.NoError(t, repo.LinkAccount(context.Background(), "u1", "a1", "sfin-001"))

	req := authedRequest(t, http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	conn, err := repo.GetConnection(context.Background(), "u1")
	require.NoError(t, err)
	assert.Nil(t, conn)
}

func TestStatusWithLastSync(t *testing.T) {
	handler, repo, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	require.NoError(t, repo.SaveConnection(context.Background(), "u1", "enc"))
	require.NoError(t, repo.UpdateSyncSuccess(context.Background(), "u1"))

	req := authedRequest(t, http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var status map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&status))
	assert.Equal(t, true, status["connected"])
	assert.NotNil(t, status["last_sync_at"])
}

func TestSetupClaimFailed(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	claimServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer claimServer.Close()

	token := base64.StdEncoding.EncodeToString([]byte(claimServer.URL))
	req := authedRequest(t, http.MethodPost, "/setup", setupRequest{SetupToken: token})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetupEmptyToken(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	req := authedRequest(t, http.MethodPost, "/setup", setupRequest{SetupToken: ""})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetupInvalidBody(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	req := authedRequest(t, http.MethodPost, "/setup", nil)
	req.Body = http.NoBody
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLinkMissingFields(t *testing.T) {
	handler, repo, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	require.NoError(t, repo.SaveConnection(context.Background(), "u1", "enc"))

	req := authedRequest(t, http.MethodPost, "/link", linkRequest{AccountID: "", SimplefinID: "sfin-001"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestLinkNonexistentAccount(t *testing.T) {
	handler, repo, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	require.NoError(t, repo.SaveConnection(context.Background(), "u1", "enc"))

	req := authedRequest(t, http.MethodPost, "/link", linkRequest{AccountID: "nonexistent", SimplefinID: "sfin-001"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListSimplefinAccounts(t *testing.T) {
	handler, repo, sfinServer := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	accessURL := testAccessURL(sfinServer.URL)
	encrypted, err := pennywisecrypto.Encrypt(testEncKey, accessURL)
	require.NoError(t, err)
	require.NoError(t, repo.SaveConnection(context.Background(), "u1", encrypted))

	req := authedRequest(t, http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	accounts := resp["accounts"].([]interface{})
	assert.Len(t, accounts, 2)
}

func TestListSimplefinAccountsNotConnected(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	req := authedRequest(t, http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSyncNotConnected(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	req := authedRequest(t, http.MethodPost, "/sync", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUnauthenticatedRequest(t *testing.T) {
	handler, _, _ := setupHandlerTest(t)
	router := Routes(handler, testSecret)

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
