package simplefin

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaimToken(t *testing.T) {
	claimServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write([]byte("https://testuser:testpass@bridge.example.com/simplefin"))
	}))
	defer claimServer.Close()

	token := base64.StdEncoding.EncodeToString([]byte(claimServer.URL))

	client := NewClient(nil)
	accessURL, err := client.ClaimToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, "https://testuser:testpass@bridge.example.com/simplefin", accessURL)
}

func TestClaimTokenInvalidBase64(t *testing.T) {
	client := NewClient(nil)
	_, err := client.ClaimToken(context.Background(), "not-valid-base64!@#")
	assert.Error(t, err)
}

func TestClaimTokenForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	token := base64.StdEncoding.EncodeToString([]byte(server.URL))
	client := NewClient(nil)
	_, err := client.ClaimToken(context.Background(), token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already claimed")
}

func TestFetchAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "testuser" || pass != "testpass" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		assert.Equal(t, "/simplefin/accounts", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"errors": [],
			"accounts": [
				{
					"id": "acc-001",
					"name": "Checking",
					"currency": "USD",
					"balance": "1234.56",
					"available-balance": "1200.00",
					"balance-date": 1709856000,
					"org": {
						"domain": "mybank.com",
						"name": "My Bank",
						"id": "mybank"
					},
					"transactions": [
						{
							"id": "txn-001",
							"posted": 1709856000,
							"amount": "-25.50",
							"description": "WALMART",
							"payee": "Walmart",
							"memo": "",
							"pending": false
						}
					]
				},
				{
					"id": "acc-002",
					"name": "Savings",
					"currency": "USD",
					"balance": "5000.00",
					"balance-date": 1709856000,
					"org": {
						"domain": "mybank.com",
						"name": "My Bank",
						"id": "mybank"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	accessURL := server.URL + "/simplefin"
	client := NewClient(nil)
	startDate := int64(1709000000)
	resp, err := client.FetchAccounts(context.Background(), "testuser", "testpass", accessURL, &startDate)
	require.NoError(t, err)
	assert.Empty(t, resp.Errors)
	require.Len(t, resp.Accounts, 2)

	assert.Equal(t, "acc-001", resp.Accounts[0].ID)
	assert.Equal(t, "Checking", resp.Accounts[0].Name)
	assert.Equal(t, "1234.56", resp.Accounts[0].Balance)
	assert.Equal(t, "My Bank", resp.Accounts[0].Org.Name)
	require.Len(t, resp.Accounts[0].Transactions, 1)
	assert.Equal(t, "txn-001", resp.Accounts[0].Transactions[0].ID)
	assert.Equal(t, "-25.50", resp.Accounts[0].Transactions[0].Amount)

	assert.Equal(t, "acc-002", resp.Accounts[1].ID)
	assert.Equal(t, "5000.00", resp.Accounts[1].Balance)
}

func TestFetchAccountsNilStartDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.URL.Query().Get("start-date"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [], "accounts": []}`))
	}))
	defer server.Close()

	client := NewClient(nil)
	resp, err := client.FetchAccounts(context.Background(), "user", "pass", server.URL, nil)
	require.NoError(t, err)
	assert.Empty(t, resp.Accounts)
}

func TestFetchAccountsBadCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	accessURL := server.URL + "/simplefin"
	client := NewClient(nil)
	_, err := client.FetchAccounts(context.Background(), "bad", "creds", accessURL, nil)
	assert.Error(t, err)
}

func TestParseAccessURL(t *testing.T) {
	username, password, baseURL, err := ParseAccessURL("https://myuser:mypass@beta-bridge.simplefin.org/simplefin")
	require.NoError(t, err)
	assert.Equal(t, "myuser", username)
	assert.Equal(t, "mypass", password)
	assert.Equal(t, "https://beta-bridge.simplefin.org/simplefin", baseURL)
}

func TestParseAccessURLNoCredentials(t *testing.T) {
	_, _, _, err := ParseAccessURL("https://beta-bridge.simplefin.org/simplefin")
	assert.Error(t, err)
}

func TestParseAccessURLMissingPassword(t *testing.T) {
	_, _, _, err := ParseAccessURL("https://useronly@beta-bridge.simplefin.org/simplefin")
	assert.Error(t, err)
}

func TestFetchAccountsPaymentRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPaymentRequired)
	}))
	defer server.Close()

	client := NewClient(nil)
	_, err := client.FetchAccounts(context.Background(), "user", "pass", server.URL, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscription lapsed")
}

func TestClaimTokenUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	token := base64.StdEncoding.EncodeToString([]byte(server.URL))
	client := NewClient(nil)
	_, err := client.ClaimToken(context.Background(), token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}

func TestFetchAccountsUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(nil)
	_, err := client.FetchAccounts(context.Background(), "user", "pass", server.URL, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status")
}
