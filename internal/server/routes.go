package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/export"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

func NewRouterWithProvider(logger *log.Logger, gameEngine engine.GameEngine, queries statedb.Querier, provider llm.Provider, pool *pgxpool.Pool, defaultUserID uuid.UUID, cfg config.Config, saveStore *saves.Store) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(loggingMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://gm.subcult.tv", "http://gm.subcult.tv", "https://edda.subcult.tv"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	campaignH := &handlers.CampaignHandlers{Engine: gameEngine, Queries: queries, Logger: logger, Pool: pool}
	charH := &handlers.CharacterHandlers{Queries: queries, Logger: logger, Pool: pool}
	worldH := &handlers.WorldHandlers{Queries: queries, Logger: logger}
	actionH := &handlers.ActionHandlers{Engine: gameEngine, Queries: queries, Provider: provider, Logger: logger}
	startupH := handlers.NewStartupHandlers(provider, queries, logger, pool)
	registerAPIRoutes(logger, r, campaignH, charH, worldH, actionH, startupH, pool, defaultUserID, cfg, saveStore, provider)
	return r
}

func registerAPIRoutes(logger *log.Logger, r chi.Router, campaignH *handlers.CampaignHandlers, charH *handlers.CharacterHandlers, worldH *handlers.WorldHandlers, actionH *handlers.ActionHandlers, startupH *handlers.StartupHandlers, pool *pgxpool.Pool, defaultUserID uuid.UUID, cfg config.Config, saveStore *saves.Store, provider llm.Provider) {
	var authMW auth.AuthMiddleware
	if cfg.Server.JWTSecret != "" {
		authMW = auth.NewJWTMiddleware(cfg.Server.JWTSecret)
	} else {
		authMW = auth.NewNoOpMiddleware(defaultUserID)
	}

	r.Route("/api", func(r chi.Router) {
		r.Use(disableCacheMiddleware)
		r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(logger, w, http.StatusOK, map[string]any{"status": "ok", "engine_ready": campaignH != nil && campaignH.Engine != nil})
		})
		r.Route("/v1", func(r chi.Router) {
			if cfg.Server.JWTSecret != "" {
				r.Route("/auth", func(r chi.Router) {
					authH := auth.NewAuthHandlers(auth.NewDBAuthQuerier(pool), cfg.Server.JWTSecret)
					r.Post("/register", authH.Register)
					r.Post("/login", authH.Login)
					r.Post("/logout", func(w http.ResponseWriter, req *http.Request) {
						auth.ClearSessionCookie(w, req)
						w.WriteHeader(http.StatusNoContent)
					})
				})
			}
			r.Group(func(r chi.Router) {
				r.Use(authMW.Authenticate)
				if cfg.Server.JWTSecret != "" {
					authH := auth.NewAuthHandlers(auth.NewDBAuthQuerier(pool), cfg.Server.JWTSecret)
					r.Get("/auth/me", authH.Me)
				}
				r.Route("/campaigns", func(r chi.Router) {
					r.Get("/", campaignH.ListCampaigns)
					r.Post("/", campaignH.CreateCampaign)
					r.Route("/start", func(r chi.Router) {
						r.Post("/campaign-interview", startupH.StartCampaignInterview)
						r.Post("/campaign-interview/{sessionID}", startupH.StepCampaignInterview)
						r.Post("/proposals", startupH.GenerateCampaignProposals)
						r.Post("/name", startupH.GenerateCampaignName)
						r.Post("/character-interview", startupH.StartCharacterInterview)
						r.Post("/character-interview/{sessionID}", startupH.StepCharacterInterview)
						r.Post("/world", startupH.BuildWorld)
					})
					r.Route("/{id}", func(r chi.Router) {
						r.Get("/", campaignH.GetCampaign)
						r.Put("/", campaignH.UpdateCampaign)
						r.Delete("/", campaignH.DeleteCampaign)
						r.Get("/history", campaignH.GetSessionHistory)
						r.Get("/character", charH.GetCharacter)
						r.Get("/character/inventory", charH.GetCharacterInventory)
						r.Get("/character/abilities", charH.GetCharacterAbilities)
						r.Get("/character/feats", charH.GetCharacterFeats)
						r.Get("/character/skills", charH.GetCharacterSkills)
						r.Get("/locations", worldH.ListLocations)
						r.Get("/locations/{lid}", worldH.GetLocation)
						r.Get("/npcs/encountered", worldH.ListEncounteredNPCs)
						r.Get("/npcs", worldH.ListNPCs)
						r.Get("/npcs/{nid}/dialogue", worldH.GetNPCDialogue)
						r.Get("/npcs/{nid}", worldH.GetNPC)
						r.Get("/quests", worldH.ListQuests)
						r.Get("/quests/{qid}", worldH.GetQuest)
						r.Get("/quests/{qid}/notes", worldH.ListQuestNotes)
						r.Post("/quests/{qid}/notes", worldH.CreateQuestNote)
						r.Delete("/quests/{qid}/notes/{noteID}", worldH.DeleteQuestNote)
						r.Get("/quests/{qid}/history", worldH.ListQuestHistory)
						r.Get("/facts", worldH.ListKnownFacts)
						r.Get("/relationships", worldH.ListAwareRelationships)
						r.Get("/codex/languages", worldH.ListKnownLanguages)
						r.Get("/codex/cultures", worldH.ListKnownCultures)
						r.Get("/codex/beliefs", worldH.ListKnownBeliefSystems)
						r.Get("/codex/economies", worldH.ListKnownEconomicSystems)
						r.Get("/map", worldH.GetMapData)
						r.Post("/action", actionH.ProcessAction)
						r.Get("/ws", actionH.HandleWebSocket)
						savesH := saves.NewHandlers(saveStore)
						r.Post("/saves", savesH.ManualSave)
						r.Get("/saves", savesH.ListSaves)
						r.Post("/start-over", savesH.StartOver)
						r.Get("/time", savesH.GetTime)
						exportH := export.NewHandlers(export.NewStore(pool))
						r.Get("/export/json", exportH.ExportJSON)
						r.Get("/export/transcript", exportH.ExportTranscript)
						r.Get("/export/character", exportH.ExportCharacterSheet)
						journalStore := journal.NewStore(pool)
						journalH := journal.NewHandlersWithSummarizer(journalStore, journal.NewSummarizer(provider, journalStore))
						r.Route("/journal", func(r chi.Router) {
							r.Get("/summaries", journalH.ListSummaries)
							r.Get("/entries", journalH.ListEntries)
							r.Post("/entries", journalH.CreateEntry)
							r.Delete("/entries/{eid}", journalH.DeleteEntry)
							r.Post("/summarize", journalH.Summarize)
						})
					})
				})
			})
		})
	})
}

func disableCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0, s-maxage=0")
		headers.Set("Pragma", "no-cache")
		headers.Set("Expires", "0")
		headers.Set("Surrogate-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(logger *log.Logger, w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logger.Errorf("encode json response: %v", err)
	}
}

func loggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			logger.Infof("%s %s status=%d bytes=%d duration=%s request_id=%s", r.Method, r.URL.Path, ww.Status(), ww.BytesWritten(), time.Since(start).Round(time.Millisecond), middleware.GetReqID(r.Context()))
		})
	}
}
