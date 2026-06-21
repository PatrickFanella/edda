package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestTokenFromRequest_PrecedenceAndFallback(t *testing.T) {
	userID := uuid.New()
	token, err := GenerateToken(userID, "secret", DefaultTokenTTL)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	tests := []struct {
		name    string
		req     *http.Request
		wantTok string
		wantErr bool
	}{
		{name: "rest cookie rejected", req: func() *http.Request {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
			r.AddCookie(&http.Cookie{Name: AuthCookieName, Value: token})
			return r
		}(), wantErr: true},
		{name: "ws cookie accepted", req: func() *http.Request {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/1/ws", nil)
			r.Header.Set("Upgrade", "websocket")
			r.AddCookie(&http.Cookie{Name: AuthCookieName, Value: token})
			return r
		}(), wantTok: token},
		{name: "malformed auth no fallback", req: func() *http.Request {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/1/ws", nil)
			r.Header.Set("Authorization", "Bearer")
			r.AddCookie(&http.Cookie{Name: AuthCookieName, Value: token})
			r.Header.Set("Upgrade", "websocket")
			return r
		}(), wantErr: true},
		{name: "invalid auth no fallback", req: func() *http.Request {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/1/ws", nil)
			r.Header.Set("Authorization", "Basic abc")
			r.AddCookie(&http.Cookie{Name: AuthCookieName, Value: token})
			r.Header.Set("Upgrade", "websocket")
			return r
		}(), wantErr: true},
		{name: "bearer precedence over cookie", req: func() *http.Request {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/campaigns/1/ws", nil)
			r.Header.Set("Authorization", "Bearer "+token)
			r.AddCookie(&http.Cookie{Name: AuthCookieName, Value: "other"})
			r.Header.Set("Upgrade", "websocket")
			return r
		}(), wantTok: token},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TokenFromRequest(tc.req)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got token %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("token from request: %v", err)
			}
			if got != tc.wantTok {
				t.Fatalf("token = %q, want %q", got, tc.wantTok)
			}
		})
	}
}

func TestSessionCookieIssueAndClear(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	IssueSessionCookie(rec, req, "tok")
	issue := rec.Result().Cookies()[0]
	if issue.Path != "/" || !issue.HttpOnly || !issue.Secure || issue.SameSite != http.SameSiteLaxMode || issue.MaxAge <= 0 {
		t.Fatalf("issue cookie attrs = %+v", issue)
	}

	rec = httptest.NewRecorder()
	ClearSessionCookie(rec, req)
	clear := rec.Result().Cookies()[0]
	if clear.Path != "/" || !clear.HttpOnly || !clear.Secure || clear.SameSite != http.SameSiteLaxMode || clear.MaxAge != -1 {
		t.Fatalf("clear cookie attrs = %+v", clear)
	}
}

func TestSecureCookieFromRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if SecureCookieFromRequest(req) {
		t.Fatal("expected insecure request")
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	if !SecureCookieFromRequest(req) {
		t.Fatal("expected forwarded https to be secure")
	}
}

func TestUserFromContext(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		if userID, ok := UserFromContext(nil); ok || userID != uuid.Nil {
			t.Fatalf("expected missing user, got %s %v", userID, ok)
		}
	})

	t.Run("nil", func(t *testing.T) {
		if userID, ok := UserFromContext(WithUser(context.Background(), uuid.Nil)); ok || userID != uuid.Nil {
			t.Fatalf("expected nil user, got %s %v", userID, ok)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		ctx := contextWithValueForTest("not-a-uuid")
		if userID, ok := UserFromContext(ctx); ok || userID != uuid.Nil {
			t.Fatalf("expected wrong type to be rejected, got %s %v", userID, ok)
		}
	})

	t.Run("valid", func(t *testing.T) {
		userID := uuid.New()
		if got, ok := UserFromContext(WithUser(context.Background(), userID)); !ok || got != userID {
			t.Fatalf("expected %s, got %s %v", userID, got, ok)
		}
	})
}

func contextWithValueForTest(v any) context.Context {
	return context.WithValue(context.Background(), sessionContextKey{}, v)
}
