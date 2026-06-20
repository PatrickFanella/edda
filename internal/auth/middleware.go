package auth

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
)

const AuthCookieName = "gm_token"

var errInvalidAuthorizationHeader = errors.New("invalid authorization header format")

// AuthMiddleware defines middleware that authenticates a request and enriches context.
type AuthMiddleware interface {
	Authenticate(next http.Handler) http.Handler
}

// ---------------------------------------------------------------------------
// NoOpMiddleware — injects a hardcoded default user (used by TUI)
// ---------------------------------------------------------------------------

// NoOpMiddleware injects a default user ID into every request context.
type NoOpMiddleware struct {
	defaultUserID uuid.UUID
}

// NewNoOpMiddleware creates an auth middleware that always authenticates as defaultUserID.
func NewNoOpMiddleware(defaultUserID uuid.UUID) AuthMiddleware {
	return &NoOpMiddleware{defaultUserID: defaultUserID}
}

// Authenticate sets the default user ID in request context and passes through.
func (m *NoOpMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithUser(r.Context(), m.defaultUserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------------------------------------------------------------------------
// JWTMiddleware — validates Bearer tokens from the Authorization header
// ---------------------------------------------------------------------------

// JWTMiddleware validates JWT tokens and injects the user ID into context.
type JWTMiddleware struct {
	secret string
}

// NewJWTMiddleware creates an auth middleware that validates JWT Bearer tokens.
func NewJWTMiddleware(secret string) AuthMiddleware {
	return &JWTMiddleware{secret: secret}
}

// Authenticate reads the Authorization header, validates the JWT, and sets
// the user ID in request context. Returns 401 if the token is missing or invalid.
func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := tokenFromRequest(r)
		if err != nil {
			if errors.Is(err, errInvalidAuthorizationHeader) {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		userID, err := ValidateToken(token, m.secret)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := WithUser(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tokenFromRequest(r *http.Request) (string, error) {
	return TokenFromRequest(r)
}
