package ibapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	if client.baseURL != "https://localhost:5000" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://localhost:5000")
	}

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("https://custom.local:8080"),
		WithInsecureSkipVerify(true),
		WithTimeout(60*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	if client.baseURL != "https://custom.local:8080" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://custom.local:8080")
	}

	if !client.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}

func TestNewClientWithCustomHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 45 * time.Second}
	client, err := NewClient(WithHTTPClient(customClient))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	if client.httpClient != customClient {
		t.Error("httpClient should be the custom client")
	}
}

func TestBuildURL(t *testing.T) {
	client, _ := NewClient(WithBaseURL("https://localhost:5000"))

	tests := []struct {
		path     string
		expected string
	}{
		{"/v1/api/iserver/auth/status", "https://localhost:5000/v1/api/iserver/auth/status"},
		{"/v1/api/tickle", "https://localhost:5000/v1/api/tickle"},
		{"/v1/api/sso/validate", "https://localhost:5000/v1/api/sso/validate"},
	}

	for _, tt := range tests {
		got := client.BuildURL(tt.path)
		if got != tt.expected {
			t.Errorf("BuildURL(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestClientClose(t *testing.T) {
	client, _ := NewClient()
	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with error message",
			err:      &APIError{Message: "specific error"},
			expected: "specific error",
		},
		{
			name:     "with status",
			err:      &APIError{Status: "bad request"},
			expected: "bad request",
		},
		{
			name:     "with code only",
			err:      &APIError{Code: 500},
			expected: "API error: code=500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestHTTPError(t *testing.T) {
	err := &HTTPError{
		HTTPStatusCode: 400,
		APIError:       APIError{Message: "invalid request"},
	}

	expected := "invalid request"
	if got := err.Error(); got != expected {
		t.Errorf("HTTPError.Error() = %q, want %q", got, expected)
	}
}

// Mock server for testing
type mockServer struct {
	*httptest.Server
	handler http.HandlerFunc
}

func newMockServer(handler http.HandlerFunc) *mockServer {
	return &mockServer{
		Server: httptest.NewTLSServer(handler),
	}
}

func (m *mockServer) Client() *Client {
	client, _ := NewClient(
		WithBaseURL(m.URL),
		WithInsecureSkipVerify(true),
	)
	return client
}

func TestSessionAuthStatus(t *testing.T) {
	expected := &AuthStatusResponse{
		Authenticated: true,
		Connected:    true,
		Message:      "success",
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v1/api/iserver/auth/status" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/iserver/auth/status")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Session().AuthStatus(context.Background())
	if err != nil {
		t.Fatalf("AuthStatus() error = %v", err)
	}

	if resp.Authenticated != expected.Authenticated {
		t.Errorf("Authenticated = %v, want %v", resp.Authenticated, expected.Authenticated)
	}
	if resp.Connected != expected.Connected {
		t.Errorf("Connected = %v, want %v", resp.Connected, expected.Connected)
	}
}

func TestSessionPing(t *testing.T) {
	expected := &TickResponse{
		Session:    "test-session-id",
		SSOExpires: 3600,
		UserID:     12345,
		Collision:  false,
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/api/tickle" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/tickle")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Session().Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	if resp.Session != expected.Session {
		t.Errorf("Session = %q, want %q", resp.Session, expected.Session)
	}
	if resp.UserID != expected.UserID {
		t.Errorf("UserID = %d, want %d", resp.UserID, expected.UserID)
	}
}

func TestSessionValidateSSO(t *testing.T) {
	expected := &ValidateResponse{
		UserID:   12345,
		UserName: "testuser",
		Result:   true,
		Expires:  3600,
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}
		if r.URL.Path != "/v1/api/sso/validate" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/sso/validate")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Session().ValidateSSO(context.Background())
	if err != nil {
		t.Fatalf("ValidateSSO() error = %v", err)
	}

	if resp.UserName != expected.UserName {
		t.Errorf("UserName = %q, want %q", resp.UserName, expected.UserName)
	}
	if !resp.Result {
		t.Error("Result should be true")
	}
}

func TestSessionInitSession(t *testing.T) {
	expected := &AuthStatusResponse{
		Authenticated: true,
		Connected:    true,
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v1/api/iserver/auth/ssodh/init" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/iserver/auth/ssodh/init")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Session().InitSession(context.Background(), true, false)
	if err != nil {
		t.Fatalf("InitSession() error = %v", err)
	}

	if !resp.Authenticated {
		t.Error("Authenticated should be true")
	}
}

func TestContractSecdefSearch(t *testing.T) {
	expected := []SecDefSearchResult{
		{
			Conid:       "12345",
			Symbol:      "AAPL",
			CompanyName: "Apple Inc.",
		},
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}
		expectedPath := "/v1/api/iserver/secdef/search"
		if r.URL.Path != expectedPath {
			t.Errorf("path = %q, want %q", r.URL.Path, expectedPath)
		}
		r.ParseForm()
		if r.FormValue("symbol") != "AAPL" {
			t.Errorf("symbol = %q, want %q", r.FormValue("symbol"), "AAPL")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Contract().SecdefSearch(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("SecdefSearch() error = %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("len(resp) = %d, want 1", len(resp))
	}
	if resp[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", resp[0].Symbol, "AAPL")
	}
}

func TestMarketDataHistoricalData(t *testing.T) {
	expected := &HistoricalDataResponse{
		Symbol: "AAPL",
		Data: []HistoricalDataBar{
			{O: 150.0, H: 151.0, L: 149.0, C: 150.5, V: 1000000, T: 1234567890},
		},
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/api/iserver/marketdata/history" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/iserver/marketdata/history")
		}
		r.ParseForm()
		if r.FormValue("conid") != "12345" {
			t.Errorf("conid = %q, want %q", r.FormValue("conid"), "12345")
		}
		if r.FormValue("period") != "1d" {
			t.Errorf("period = %q, want %q", r.FormValue("period"), "1d")
		}
		if r.FormValue("bar") != "5 mins" {
			t.Errorf("bar = %q, want %q", r.FormValue("bar"), "5 mins")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.MarketData().HistoricalData(context.Background(), HistoricalDataRequest{
		Conid:    "12345",
		Period:   "1d",
		Bar:      "5 mins",
	})
	if err != nil {
		t.Fatalf("HistoricalData() error = %v", err)
	}

	if resp.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", resp.Symbol, "AAPL")
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
}

func TestScannerParams(t *testing.T) {
	expected := &ScannerParams{
		ScanTypeList: []ScanType{
			{DisplayName: "Top % Gainers", Code: "gainers"},
		},
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/api/iserver/scanner/params" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/iserver/scanner/params")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Scanner().Params(context.Background())
	if err != nil {
		t.Fatalf("Params() error = %v", err)
	}

	if len(resp.ScanTypeList) != 1 {
		t.Fatalf("len(ScanTypeList) = %d, want 1", len(resp.ScanTypeList))
	}
}

func TestScannerRun(t *testing.T) {
	expected := &ScannerResponse{
		Contracts: []ScannerContract{
			{Symbol: "AAPL", ConID: 12345},
		},
	}

	m := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/v1/api/iserver/scanner/run" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/v1/api/iserver/scanner/run")
		}

		json.NewEncoder(w).Encode(expected)
	})
	defer m.Server.Close()

	client := m.Client()
	defer client.Close()

	resp, err := client.Scanner().Run(context.Background(), ScannerRunRequest{
		Instrument: "STK",
		Location:   "STK.US.Major",
		Type:       "gainers",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(resp.Contracts) != 1 {
		t.Fatalf("len(Contracts) = %d, want 1", len(resp.Contracts))
	}
}