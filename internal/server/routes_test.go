package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/config"
)

func TestNewRouterWithProvider_PreservesPublicRoutes(t *testing.T) {
	t.Parallel()

	router := NewRouterWithProvider(log.New(io.Discard), nil, nil, nil, nil, uuid.Nil, config.Config{}, nil)

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{http.MethodGet, "/healthz", http.StatusOK},
		{http.MethodGet, "/api/healthz", http.StatusOK},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != tc.want {
				t.Fatalf("%s %s status = %d, want %d", tc.method, tc.path, rec.Code, tc.want)
			}
		})
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/campaigns", nil))
	if rec.Code == http.StatusNotFound {
		t.Fatal("GET /api/v1/campaigns returned 404, want registered route")
	}
}

func TestNewRouterWithProvider_AuthModeSwitchesByConfig(t *testing.T) {
	t.Parallel()

	logger := log.New(io.Discard)
	defaultUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	t.Run("auth enabled", func(t *testing.T) {
		t.Parallel()
		cfg := config.Config{}
		cfg.Server.JWTSecret = "test-secret"
		router := NewRouterWithProvider(logger, nil, nil, nil, nil, defaultUserID, cfg, nil)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
		if rec.Code == http.StatusNotFound {
			t.Fatal("GET /api/v1/auth/me returned 404 in auth-enabled mode")
		}

		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/campaigns", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("GET /api/v1/campaigns status = %d, want %d without auth token", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("auth disabled", func(t *testing.T) {
		t.Parallel()
		router := NewRouterWithProvider(logger, nil, nil, nil, nil, defaultUserID, config.Config{}, nil)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("GET /api/v1/auth/me status = %d, want %d when JWT is disabled", rec.Code, http.StatusNotFound)
		}

		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/campaigns", nil))
		if rec.Code == http.StatusNotFound {
			t.Fatal("GET /api/v1/campaigns returned 404 in no-op auth mode")
		}
	})
}

func TestNewRouterWithProvider_CORSAndRequestIDStillWork(t *testing.T) {
	t.Parallel()

	router := NewRouterWithProvider(log.New(io.Discard), nil, nil, nil, nil, uuid.Nil, config.Config{}, nil)
	mux, ok := router.(*chi.Mux)
	if !ok {
		t.Fatalf("NewRouterWithProvider() type = %T, want *chi.Mux", router)
	}

	mux.Get("/panic", func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})
	panicRec := httptest.NewRecorder()
	router.ServeHTTP(panicRec, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if panicRec.Code != http.StatusInternalServerError {
		t.Fatalf("GET /panic status = %d, want %d", panicRec.Code, http.StatusInternalServerError)
	}

	preflightReq := httptest.NewRequest(http.MethodOptions, "/api/v1/campaigns/", nil)
	preflightReq.Header.Set("Origin", "http://localhost:3000")
	preflightReq.Header.Set("Access-Control-Request-Method", http.MethodGet)
	preflightRes := httptest.NewRecorder()
	router.ServeHTTP(preflightRes, preflightReq)
	if got := preflightRes.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("OPTIONS allow-origin = %q, want %q", got, "http://localhost:3000")
	}

	mux.Get("/request-id", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(middleware.GetReqID(r.Context())))
	})
	requestIDRec := httptest.NewRecorder()
	router.ServeHTTP(requestIDRec, httptest.NewRequest(http.MethodGet, "/request-id", nil))
	if strings.TrimSpace(requestIDRec.Body.String()) == "" {
		t.Fatal("request ID middleware did not populate request context")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/healthz", nil))
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("GET /api/healthz body = %q, want status ok", rec.Body.String())
	}
}
