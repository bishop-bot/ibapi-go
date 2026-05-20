package ibapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// SessionService handles session-related API calls.
type SessionService struct {
	client *Client
}

// AuthStatusResponse represents the response from /iserver/auth/status.
type AuthStatusResponse struct {
	Authenticated bool       `json:"authenticated"`
	Competing    bool       `json:"competing"`
	Connected    bool       `json:"connected"`
	Message      string     `json:"message"`
	MAC          string     `json:"MAC"`
	ServerInfo   ServerInfo `json:"serverInfo,omitempty"`
	HardwareInfo string     `json:"hardware_info,omitempty"`
	Fail         string     `json:"fail,omitempty"`
}

// ServerInfo contains server version and name information.
type ServerInfo struct {
	ServerName    string `json:"serverName"`
	ServerVersion string `json:"serverVersion"`
}

// InitSessionRequest represents the request body for /iserver/auth/ssodh/init.
type InitSessionRequest struct {
	Publish bool `json:"publish"`
	Compete bool `json:"compete"`
}

// TickResponse represents the response from /tickle.
type TickResponse struct {
	Session    string       `json:"session"`
	SSOExpires int64        `json:"ssoExpires"`
	Collision  bool         `json:"collision"`
	UserID     int64        `json:"userId"`
	HMDS       HMDSInfo     `json:"hmds"`
	Iserver    IserverInfo  `json:"iserver"`
}

// HMDSInfo contains historical market data service info.
type HMDSInfo struct {
	Error string `json:"error"`
}

// IserverInfo contains iserver auth status.
type IserverInfo struct {
	AuthStatus AuthStatusResponse `json:"authStatus"`
}

// ValidateResponse represents the response from /sso/validate.
type ValidateResponse struct {
	UserID                int64           `json:"USER_ID"`
	UserName              string          `json:"USER_NAME"`
	Result                bool            `json:"RESULT"`
	AuthTime              int64           `json:"AUTH_TIME"`
	SFEnabled             bool            `json:"SF_ENABLED"`
	IsFreeTrial           bool            `json:"IS_FREE_TRIAL"`
	Credential            string          `json:"CREDENTIAL"`
	IP                    string          `json:"IP"`
	Expires               int64           `json:"EXPIRES"`
	QualifiedForMobileAuth *bool           `json:"QUALIFIED_FOR_MOBILE_AUTH,omitempty"`
	LandingApp            string          `json:"LANDING_APP"`
	IsMaster              bool            `json:"IS_MASTER"`
	LastAccessed          int64           `json:"lastAccessed"`
	LoginType             int            `json:"LOGIN_TYPE"`
	PaperUserName         string          `json:"PAPER_USER_NAME,omitempty"`
	Features              Features        `json:"features"`
	Region                string          `json:"region,omitempty"`
}

// Features contains supported feature flags.
type Features struct {
	Env          string `json:"env"`
	WLMS         bool   `json:"wlms"`
	Realtime     bool   `json:"realtime"`
	Bond         bool   `json:"bond"`
	OptionChains bool   `json:"optionChains"`
	Calendar     bool   `json:"calendar"`
	NewMF        bool   `json:"newMf"`
}

// AuthStatus returns the current authentication status to the Brokerage system.
// Market Data and Trading is not possible if not authenticated.
func (s *SessionService) AuthStatus(ctx context.Context) (*AuthStatusResponse, error) {
	body := map[string]interface{}{}
	return s.authStatus(ctx, body)
}

// authStatus performs the auth status check with optional body parameters.
func (s *SessionService) authStatus(ctx context.Context, body map[string]interface{}) (*AuthStatusResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.BuildURL("/v1/api/iserver/auth/status"), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result AuthStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// InitSession initializes the brokerage session.
// This is essential for using all endpoints besides /portfolio,
// including access to trading and market data.
func (s *SessionService) InitSession(ctx context.Context, publish, compete bool) (*AuthStatusResponse, error) {
	body := InitSessionRequest{
		Publish: publish,
		Compete: compete,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.BuildURL("/v1/api/iserver/auth/ssodh/init"), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result AuthStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Ping pings the server to prevent the session from timing out.
// If the gateway has not received any requests for several minutes
// an open session will automatically timeout.
// It is expected to call this endpoint approximately every 60 seconds
// to maintain the connection to the brokerage session.
func (s *SessionService) Ping(ctx context.Context) (*TickResponse, error) {
	body := map[string]interface{}{}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.BuildURL("/v1/api/tickle"), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TickResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Update client auth state
	if result.Session != "" {
		s.client.auth = &AuthState{
			SessionID:  result.Session,
			SSOExpires: result.SSOExpires,
			UserID:     result.UserID,
		}
	}

	return &result, nil
}

// ValidateSSO validates the current session for the SSO user.
// This endpoint is only valid for Client Portal Gateway and OAuth 2.0 clients.
func (s *SessionService) ValidateSSO(ctx context.Context) (*ValidateResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/sso/validate"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Logout logs the user out of the gateway session.
// Any further activity requires re-authentication.
func (s *SessionService) Logout(ctx context.Context) error {
	body := map[string]interface{}{}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.BuildURL("/v1/api/logout"), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Clear auth state
	s.client.auth = nil

	return nil
}