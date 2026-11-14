package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// Sign in with ChatGPT endpoints and constants. See
// https://developers.openai.com/siwc/token-sharing-open-source/sign-in.
const (
	ChatGPTAuthBaseURL         = "https://auth.openai.com"
	ChatGPTAuthorizePath       = "/api/accounts/authorize"
	ChatGPTTokenPath           = "/api/accounts/oauth/token"
	ChatGPTResource            = "https://api.openai.com/v1"
	ChatGPTPlanScope           = "chatgpt.tokens.use.direct"
	ChatGPTDynamicClientID     = "dynamic_agent_client"
	chatGPTRefreshBeforeExpiry = 5 * time.Minute
)

// ChatGPTCredentials is the credential record written by cmd/chatgpt-login, in
// the format OpenAI's Sign in with ChatGPT guide describes.
type ChatGPTCredentials struct {
	Email          string    `json:"email"`
	Issuer         string    `json:"issuer"`
	Subject        string    `json:"subject"`
	ClientID       string    `json:"client_id"`
	ExtAgentHostID string    `json:"ext_agent_host_id"`
	IDToken        string    `json:"id_token"`
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token"`
	TokenType      string    `json:"token_type"`
	ExpiresIn      int       `json:"expires_in"`
	Scopes         []string  `json:"scopes"`
	SavedAt        time.Time `json:"saved_at"`
}

// ExpiresAt returns when the access token expires.
func (c ChatGPTCredentials) ExpiresAt() time.Time {
	return c.SavedAt.Add(time.Duration(c.ExpiresIn) * time.Second)
}

// PlanUsageGranted reports whether the grant allows using the ChatGPT plan.
func (c ChatGPTCredentials) PlanUsageGranted() bool {
	return slices.Contains(c.Scopes, ChatGPTPlanScope)
}

// ReadChatGPTCredentials reads a credential file.
func ReadChatGPTCredentials(path string) (ChatGPTCredentials, error) {
	var creds ChatGPTCredentials
	data, err := os.ReadFile(path)
	if err != nil {
		return creds, err
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return creds, fmt.Errorf("parse %s: %w", path, err)
	}
	return creds, nil
}

// WriteChatGPTCredentials writes a credential file atomically with mode 0600.
func WriteChatGPTCredentials(path string, creds ChatGPTCredentials) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".chatgpt-auth-*.json")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// ChatGPTTokenStore hands out a current access token and refreshes it, saving
// the rotated refresh token back to the credential file.
type ChatGPTTokenStore struct {
	path     string
	tokenURL string
	client   *http.Client
	now      func() time.Time

	mu    sync.Mutex
	creds ChatGPTCredentials
}

// LoadChatGPTTokenStore loads the credential file at path.
func LoadChatGPTTokenStore(path string) (*ChatGPTTokenStore, error) {
	creds, err := ReadChatGPTCredentials(path)
	if err != nil {
		return nil, fmt.Errorf("chatgpt provider unavailable: %w (run cmd/chatgpt-login and copy the file here)", err)
	}
	switch {
	case creds.ClientID == "" || creds.ClientID == ChatGPTDynamicClientID:
		return nil, fmt.Errorf("chatgpt provider unavailable: %s has no issued client_id; sign in again", path)
	case creds.RefreshToken == "":
		return nil, fmt.Errorf("chatgpt provider unavailable: %s has no refresh token; sign in again", path)
	case !creds.PlanUsageGranted():
		return nil, fmt.Errorf("chatgpt provider unavailable: the sign-in in %s did not grant ChatGPT plan usage (%s)", path, ChatGPTPlanScope)
	}
	return &ChatGPTTokenStore{
		path:     path,
		tokenURL: ChatGPTAuthBaseURL + ChatGPTTokenPath,
		client:   &http.Client{Timeout: 30 * time.Second},
		now:      time.Now,
		creds:    creds,
	}, nil
}

// AccessToken returns an access token, refreshing it first when it expires
// within five minutes.
func (s *ChatGPTTokenStore) AccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.creds.AccessToken != "" && s.now().Before(s.creds.ExpiresAt().Add(-chatGPTRefreshBeforeExpiry)) {
		return s.creds.AccessToken, nil
	}
	if err := s.refreshLocked(ctx); err != nil {
		return "", err
	}
	return s.creds.AccessToken, nil
}

// RefreshAfterReject refreshes after the API rejected rejected. If another
// request already refreshed, it returns the newer token without refreshing.
func (s *ChatGPTTokenStore) RefreshAfterReject(ctx context.Context, rejected string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.creds.AccessToken != rejected {
		return s.creds.AccessToken, nil
	}
	if err := s.refreshLocked(ctx); err != nil {
		return "", err
	}
	return s.creds.AccessToken, nil
}

type chatGPTTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

var chatGPTUnusableRefreshCodes = map[string]bool{
	"invalid_grant": true, "invalid_refresh_token": true, "token_expired": true,
	"refresh_token_expired": true, "refresh_token_invalidated": true, "refresh_token_reused": true,
	"invalid_client": true,
}

func (s *ChatGPTTokenStore) refreshLocked(ctx context.Context) error {
	form := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {s.creds.ClientID},
		"refresh_token": {s.creds.RefreshToken},
		"resource":      {ChatGPTResource},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return &ErrTimeout{URL: s.tokenURL, Err: err}
		}
		return &ErrConnection{URL: s.tokenURL, Err: fmt.Errorf("refresh chatgpt token: %w", err)}
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	var tok chatGPTTokenResponse
	_ = json.Unmarshal(body, &tok)
	if resp.StatusCode != http.StatusOK || tok.AccessToken == "" {
		code := tok.Error
		if code == "" {
			code = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if chatGPTUnusableRefreshCodes[tok.Error] {
			return &ErrAuth{URL: s.tokenURL, StatusCode: resp.StatusCode,
				Err: fmt.Errorf("chatgpt sign-in is no longer valid (%s); run cmd/chatgpt-login again", code)}
		}
		if resp.StatusCode >= 500 {
			return &ErrTransient{URL: s.tokenURL, StatusCode: resp.StatusCode, Err: fmt.Errorf("refresh chatgpt token: %s", code)}
		}
		return &ErrAuth{URL: s.tokenURL, StatusCode: resp.StatusCode, Err: fmt.Errorf("refresh chatgpt token: %s", code)}
	}

	next := s.creds
	next.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		next.RefreshToken = tok.RefreshToken
	}
	if tok.IDToken != "" {
		next.IDToken = tok.IDToken
	}
	if tok.TokenType != "" {
		next.TokenType = tok.TokenType
	}
	next.ExpiresIn = tok.ExpiresIn
	if tok.Scope != "" {
		next.Scopes = strings.Fields(tok.Scope)
	}
	next.SavedAt = s.now().UTC()
	if err := WriteChatGPTCredentials(s.path, next); err != nil {
		// The old refresh token may already be spent, so keep the new set in
		// memory even when it cannot be saved.
		s.creds = next
		return fmt.Errorf("save refreshed chatgpt credentials to %s: %w", s.path, err)
	}
	s.creds = next
	return nil
}
