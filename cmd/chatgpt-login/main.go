// Command chatgpt-login signs in with ChatGPT and writes the credential file
// that Edda's chatgpt provider reads. It follows OpenAI's "Sign in with
// ChatGPT" flow for open-source apps: a loopback callback on 127.0.0.1, PKCE,
// and dynamic client registration on first sign-in.
//
// The callback must reach this process, so run it where your browser is, or
// forward the port: ssh -L 1455:127.0.0.1:1455 <host>, then open the printed
// link in the local browser. Copy the written file to the Edda server; the
// server refreshes it from then on.
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

const (
	appName      = "Edda"
	callbackPath = "/auth/callback"
	scopes       = "openid profile email offline_access resource.invoke " + llm.ChatGPTPlanScope
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "chatgpt-login:", err)
		os.Exit(1)
	}
}

func run() error {
	home, _ := os.UserHomeDir()
	out := flag.String("out", filepath.Join(home, ".config", "edda", "chatgpt-auth.json"), "credential file to write")
	port := flag.Int("port", 1455, "loopback callback port")
	hostID := flag.String("host-id", "", "ext_agent_host_id for this host (default: reuse the one in -out, or create one)")
	noBrowser := flag.Bool("no-browser", false, "print the sign-in link instead of opening a browser")
	wait := flag.Duration("timeout", 10*time.Minute, "how long to wait for the browser callback")
	flag.Parse()

	existing, err := llm.ReadChatGPTCredentials(*out)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if *hostID == "" {
		*hostID = existing.ExtAgentHostID
	}
	if *hostID == "" {
		*hostID = "urn:uuid:" + newUUID()
	}

	state, nonce, verifier := randomToken(), randomToken(), randomToken()
	challenge := sha256.Sum256([]byte(verifier))
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d%s", *port, callbackPath)

	q := url.Values{
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"scope":                 {scopes},
		"resource":              {llm.ChatGPTResource},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge_method": {"S256"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"ext_agent_host_id":     {*hostID},
	}
	clientID := existing.ClientID
	if clientID == "" || clientID == llm.ChatGPTDynamicClientID {
		clientID = ""
		q.Set("client_id", llm.ChatGPTDynamicClientID)
		q.Set("agent_name_hint", appName)
	} else {
		q.Set("client_id", clientID)
		if existing.IDToken != "" {
			q.Set("id_token_hint", existing.IDToken)
		}
		if existing.Email != "" {
			q.Set("login_hint", existing.Email)
		}
	}
	authURL := llm.ChatGPTAuthBaseURL + llm.ChatGPTAuthorizePath + "?" + q.Encode()

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		return fmt.Errorf("listen for the callback: %w", err)
	}
	type callback struct {
		code, clientID string
		err            error
	}
	done := make(chan callback, 1)
	srv := &http.Server{ReadHeaderTimeout: 10 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != callbackPath {
			http.NotFound(w, r)
			return
		}
		v := r.URL.Query()
		if v.Get("state") != state {
			// A callback from an older attempt; keep waiting for this one.
			fmt.Fprintln(os.Stderr, "Ignored a callback from a different sign-in attempt; still waiting.")
			http.Error(w, "this callback belongs to a different sign-in attempt; use the latest link", http.StatusBadRequest)
			return
		}
		var cb callback
		switch {
		case v.Get("error") != "":
			cb.err = fmt.Errorf("sign-in failed: %s %s", v.Get("error"), v.Get("error_description"))
		case v.Get("code") == "":
			cb.err = errors.New("callback has no authorization code")
		default:
			cb.code, cb.clientID = v.Get("code"), v.Get("client_id")
		}
		if cb.err != nil {
			http.Error(w, cb.err.Error(), http.StatusBadRequest)
		} else {
			_, _ = io.WriteString(w, "Signed in. You can close this tab and return to the terminal.\n")
		}
		select {
		case done <- cb:
		default:
		}
	})}
	go func() { _ = srv.Serve(listener) }()
	defer func() { _ = srv.Close() }()

	// The link can carry an ID token hint, so it is shown only on the terminal.
	fmt.Fprintf(os.Stderr, "Open this link to sign in with ChatGPT:\n\n%s\n\n", authURL)
	if !*noBrowser {
		openBrowser(authURL)
	}

	var cb callback
	select {
	case cb = <-done:
	case <-time.After(*wait):
		return errors.New("timed out waiting for the browser callback")
	}
	if cb.err != nil {
		return cb.err
	}

	switch {
	case clientID == "" && (cb.clientID == "" || cb.clientID == llm.ChatGPTDynamicClientID):
		return errors.New("registration incomplete: the callback did not include an issued client_id")
	case clientID == "":
		clientID = cb.clientID
		// Keep the registration even if the code exchange fails, so the next
		// attempt reuses this client instead of registering another one.
		if err := llm.WriteChatGPTCredentials(*out, llm.ChatGPTCredentials{ClientID: clientID, ExtAgentHostID: *hostID}); err != nil {
			return err
		}
	case cb.clientID != "" && cb.clientID != clientID:
		return fmt.Errorf("callback client_id %q does not match the saved registration", cb.clientID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tok, err := exchangeCode(ctx, clientID, cb.code, verifier, redirectURI)
	if err != nil {
		return err
	}
	claims, err := verifyIDToken(ctx, tok.IDToken, clientID, nonce)
	if err != nil {
		return fmt.Errorf("verify ID token: %w", err)
	}
	if existing.Subject != "" && existing.ClientID == clientID && claims.Subject != existing.Subject {
		return errors.New("signed in as a different ChatGPT account than the saved registration")
	}

	creds := llm.ChatGPTCredentials{
		Email:          claims.Email,
		Issuer:         claims.Issuer,
		Subject:        claims.Subject,
		ClientID:       clientID,
		ExtAgentHostID: *hostID,
		IDToken:        tok.IDToken,
		AccessToken:    tok.AccessToken,
		RefreshToken:   tok.RefreshToken,
		TokenType:      tok.TokenType,
		ExpiresIn:      tok.ExpiresIn,
		Scopes:         strings.Fields(tok.Scope),
		SavedAt:        time.Now().UTC(),
	}
	if err := llm.WriteChatGPTCredentials(*out, creds); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Signed in as %s. Wrote %s.\n", claims.Email, *out)
	if !creds.PlanUsageGranted() {
		return fmt.Errorf("the sign-in did not grant ChatGPT plan usage (%s); run again and allow it", llm.ChatGPTPlanScope)
	}
	return printModels(ctx, tok.AccessToken)
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func exchangeCode(ctx context.Context, clientID, code, verifier, redirectURI string) (*tokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code":          {code},
		"code_verifier": {verifier},
		"redirect_uri":  {redirectURI},
		"resource":      {llm.ChatGPTResource},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, llm.ChatGPTAuthBaseURL+llm.ChatGPTTokenPath, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	var tok tokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tok); err != nil {
		return nil, fmt.Errorf("exchange code: HTTP %d: %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || tok.AccessToken == "" {
		if tok.Error == "invalid_grant" {
			return nil, errors.New("the authorization code was rejected (invalid_grant); run chatgpt-login again")
		}
		return nil, fmt.Errorf("exchange code: HTTP %d: %s %s", resp.StatusCode, tok.Error, tok.ErrorDescription)
	}
	if tok.IDToken == "" {
		return nil, errors.New("token response has no ID token")
	}
	return &tok, nil
}

type idClaims struct {
	Email string `json:"email"`
	Nonce string `json:"nonce"`
	jwt.RegisteredClaims
}

// verifyIDToken checks the ID token's signature against OpenAI's published
// keys, and its issuer, audience, expiry and nonce.
func verifyIDToken(ctx context.Context, raw, clientID, nonce string) (*idClaims, error) {
	keys, err := fetchJWKS(ctx)
	if err != nil {
		return nil, err
	}
	var claims idClaims
	_, err = jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		key, ok := keys[kid]
		if !ok {
			return nil, fmt.Errorf("no published key with kid %q", kid)
		}
		return key, nil
	},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(llm.ChatGPTAuthBaseURL),
		jwt.WithAudience(clientID),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if claims.Nonce != nonce {
		return nil, errors.New("nonce does not match this sign-in attempt")
	}
	if claims.Subject == "" {
		return nil, errors.New("ID token has no subject")
	}
	return &claims, nil
}

func fetchJWKS(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	var discovery struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := getJSON(ctx, llm.ChatGPTAuthBaseURL+"/.well-known/openid-configuration", "", &discovery); err != nil {
		return nil, err
	}
	if discovery.JWKSURI == "" {
		return nil, errors.New("openid configuration has no jwks_uri")
	}
	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := getJSON(ctx, discovery.JWKSURI, "", &set); err != nil {
		return nil, err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		n, errN := base64.RawURLEncoding.DecodeString(k.N)
		e, errE := base64.RawURLEncoding.DecodeString(k.E)
		if errN != nil || errE != nil {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	if len(keys) == 0 {
		return nil, errors.New("no RSA keys in the published key set")
	}
	return keys, nil
}

// printModels lists the models this ChatGPT account can use, for
// llm.chatgpt.model.
func printModels(ctx context.Context, accessToken string) error {
	var list struct {
		Models []struct {
			Slug        string `json:"slug"`
			DisplayName string `json:"display_name"`
			Visibility  string `json:"visibility"`
		} `json:"models"`
	}
	if err := getJSON(ctx, llm.ChatGPTResource+"/models", accessToken, &list); err != nil {
		return fmt.Errorf("list models: %w", err)
	}
	fmt.Fprintln(os.Stderr, "\nModels this account can use (set one as EDDA_LLM_CHATGPT_MODEL):")
	for _, m := range list.Models {
		if m.Visibility == "list" {
			fmt.Fprintf(os.Stderr, "  %-28s %s\n", m.Slug, m.DisplayName)
		}
	}
	return nil
}

func getJSON(ctx context.Context, rawURL, bearer string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("GET %s: HTTP %d: %s", rawURL, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(v)
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func openBrowser(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	_ = cmd.Start()
}
