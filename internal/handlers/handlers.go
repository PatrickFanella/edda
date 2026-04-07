package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/PatrickFanella/game-master/internal/engine"
	"github.com/PatrickFanella/game-master/internal/llm"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// DBTX is a minimal database interface for raw SQL execution.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// Handlers holds shared dependencies for all HTTP handlers.
type Handlers struct {
	Engine          engine.GameEngine
	Queries         statedb.Querier
	Logger          *log.Logger
	Provider        llm.Provider
	Pool            DBTX
	startupSessions *startupSessionStore
}

// New creates a Handlers with the given dependencies.
func New(eng engine.GameEngine, queries statedb.Querier, logger *log.Logger, providers ...llm.Provider) *Handlers {
	if logger == nil {
		logger = log.Default()
	}
	var provider llm.Provider
	if len(providers) > 0 {
		provider = providers[0]
	}
	return &Handlers{
		Engine:          eng,
		Queries:         queries,
		Logger:          logger,
		Provider:        provider,
		startupSessions: newStartupSessionStore(),
	}
}

// NewWithPool creates a Handlers with a database pool for raw SQL operations.
func NewWithPool(eng engine.GameEngine, queries statedb.Querier, logger *log.Logger, pool DBTX, providers ...llm.Provider) *Handlers {
	h := New(eng, queries, logger, providers...)
	h.Pool = pool
	return h
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Errorf("writeJSON encode: %v", err)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// campaignIDFromURL extracts and parses the campaign ID from the {id} URL parameter.
func campaignIDFromURL(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}
