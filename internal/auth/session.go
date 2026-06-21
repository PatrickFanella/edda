package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type sessionContextKey struct{}

// WithUser stores an authenticated user ID in context.
func WithUser(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, userID)
}

// UserFromContext extracts an authenticated user ID from context.
// A uuid.Nil value is treated as unauthenticated and returns ok=false.
func UserFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	switch userID := ctx.Value(sessionContextKey{}).(type) {
	case uuid.UUID:
		if userID == uuid.Nil {
			return uuid.Nil, false
		}
		return userID, true
	default:
		return uuid.Nil, false
	}
}

// TokenFromRequest extracts the bearer token for REST requests, or cookie token
// for websocket upgrade requests when no Authorization header is present.
func TokenFromRequest(r *http.Request) (string, error) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader != "" {
		token, found := strings.CutPrefix(authHeader, "Bearer ")
		if !found || token == "" {
			return "", errInvalidAuthorizationHeader
		}
		return token, nil
	}

	if !isWebSocketUpgradeRequest(r) {
		return "", http.ErrNoCookie
	}

	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return "", err
	}
	if token := strings.TrimSpace(cookie.Value); token != "" {
		return token, nil
	}
	return "", http.ErrNoCookie
}

// SecureCookieFromRequest reports whether session cookies should be marked Secure.
func SecureCookieFromRequest(r *http.Request) bool {
	if r != nil && r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}

// IssueSessionCookie sets the browser-sendable auth cookie.
func IssueSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   SecureCookieFromRequest(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(DefaultTokenTTL.Seconds()),
	})
}

// ClearSessionCookie clears the browser-sendable auth cookie used for same-origin websocket auth.
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   SecureCookieFromRequest(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func isWebSocketUpgradeRequest(r *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket")
}
