package simplefin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AccountsResponse struct {
	Errors   []string  `json:"errors"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Currency         string `json:"currency"`
	Balance          string `json:"balance"`
	AvailableBalance string `json:"available-balance"`
	BalanceDate      int64  `json:"balance-date"`
	Org              Org    `json:"org"`
}

type Org struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	ID     string `json:"id"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{httpClient: httpClient}
}

func (c *Client) ClaimToken(ctx context.Context, setupToken string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(setupToken)
	if err != nil {
		return "", fmt.Errorf("decode setup token: %w", err)
	}

	claimURL := string(decoded)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claimURL, nil)
	if err != nil {
		return "", fmt.Errorf("create claim request: %w", err)
	}

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL from base64-decoded setup token provided by user
	if err != nil {
		return "", fmt.Errorf("claim request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("token already claimed or invalid (403)")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read claim response: %w", err)
	}

	return string(body), nil
}

func (c *Client) FetchAccounts(ctx context.Context, username, password, baseURL string) (*AccountsResponse, error) {
	endpoint := baseURL + "/accounts?balances-only=1"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(username, password)

	resp, err := c.httpClient.Do(req) //nolint:gosec // URL from stored SimpleFIN access URL
	if err != nil {
		return nil, fmt.Errorf("fetch accounts: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("access revoked or bad credentials (403)")
	}
	if resp.StatusCode == http.StatusPaymentRequired {
		return nil, fmt.Errorf("SimpleFIN subscription lapsed (402)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result AccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func ParseAccessURL(accessURL string) (username, password, baseURL string, err error) {
	parsed, err := url.Parse(accessURL)
	if err != nil {
		return "", "", "", fmt.Errorf("parse access URL: %w", err)
	}

	if parsed.User == nil {
		return "", "", "", fmt.Errorf("access URL has no credentials")
	}

	username = parsed.User.Username()
	password, _ = parsed.User.Password()
	if username == "" || password == "" {
		return "", "", "", fmt.Errorf("access URL missing username or password")
	}

	parsed.User = nil
	baseURL = parsed.String()

	return username, password, baseURL, nil
}
