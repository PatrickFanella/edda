package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// stubQuerier implements the subset of statedb.Querier used by campaign handlers.
// Methods not relevant to the test panic so missing coverage is obvious.
type stubQuerier struct {
	statedb.Querier

	campaigns    []statedb.Campaign
	campaignByID map[string]statedb.Campaign
	createErr    error
	listErr      error
}

func (s *stubQuerier) ListCampaignsByUser(_ context.Context, _ pgtype.UUID) ([]statedb.Campaign, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.campaigns, nil
}

func (s *stubQuerier) CreateCampaign(_ context.Context, arg statedb.CreateCampaignParams) (statedb.Campaign, error) {
	if s.createErr != nil {
		return statedb.Campaign{}, s.createErr
	}
	id := uuid.New()
	c := statedb.Campaign{
		ID:          dbutil.ToPgtype(id),
		Name:        arg.Name,
		Description: arg.Description,
		Genre:       arg.Genre,
		Tone:        arg.Tone,
		Themes:      arg.Themes,
		Status:      arg.Status,
		CreatedBy:   arg.CreatedBy,
	}
	return c, nil
}

func (s *stubQuerier) GetCampaignByID(_ context.Context, id pgtype.UUID) (statedb.Campaign, error) {
	uid := dbutil.FromPgtype(id).String()
	c, ok := s.campaignByID[uid]
	if !ok {
		return statedb.Campaign{}, context.DeadlineExceeded // stand-in for pgx.ErrNoRows
	}
	return c, nil
}

// newTestRouter builds a chi router with auth middleware and campaign routes
// wired to the given CampaignHandlers.
func newTestRouter(h *CampaignHandlers) *chi.Mux {
	r := chi.NewRouter()
	authMW := auth.NewNoOpMiddleware(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	r.Use(authMW.Authenticate)
	r.Route("/campaigns", func(r chi.Router) {
		r.Get("/", h.ListCampaigns)
		r.Post("/", h.CreateCampaign)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetCampaign)
		})
	})
	return r
}

func TestListCampaigns_Success(t *testing.T) {
	cID := uuid.New()
	q := &stubQuerier{
		campaigns: []statedb.Campaign{
			{
				ID:     dbutil.ToPgtype(cID),
				Name:   "Test Campaign",
				Status: "active",
				Themes: []string{"dark", "mystery"},
			},
		},
	}
	h := &CampaignHandlers{Queries: q, Logger: log.Default()}
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp api.CampaignListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Campaigns) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(resp.Campaigns))
	}
	if resp.Campaigns[0].Name != "Test Campaign" {
		t.Errorf("expected name %q, got %q", "Test Campaign", resp.Campaigns[0].Name)
	}
	if len(resp.Campaigns[0].Themes) != 2 {
		t.Errorf("expected 2 themes, got %d", len(resp.Campaigns[0].Themes))
	}
}

func TestCreateCampaign_Success(t *testing.T) {
	q := &stubQuerier{}
	h := &CampaignHandlers{Queries: q, Logger: log.Default()}
	router := newTestRouter(h)

	body, _ := json.Marshal(api.CampaignCreateRequest{
		Name:  "New Campaign",
		Genre: "fantasy",
		Tone:  "dark",
	})
	req := httptest.NewRequest(http.MethodPost, "/campaigns", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp api.CampaignResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "New Campaign" {
		t.Errorf("expected name %q, got %q", "New Campaign", resp.Name)
	}
	if resp.Status != "active" {
		t.Errorf("expected status %q, got %q", "active", resp.Status)
	}
}

func TestGetCampaign_NotFound(t *testing.T) {
	q := &stubQuerier{
		campaignByID: map[string]statedb.Campaign{},
	}
	h := &CampaignHandlers{Queries: q, Logger: log.Default()}
	router := newTestRouter(h)

	missingID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/campaigns/"+missingID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp["error"] != "campaign not found" {
		t.Errorf("expected error %q, got %q", "campaign not found", resp["error"])
	}
}

func TestGetCampaign_InvalidID(t *testing.T) {
	q := &stubQuerier{campaignByID: map[string]statedb.Campaign{}}
	h := &CampaignHandlers{Queries: q, Logger: log.Default()}
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/campaigns/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

// campaignRulesDB exercises the follow-up rules write without a real pool.
type campaignRulesDB struct {
	exec func(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (d campaignRulesDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return d.exec(ctx, sql, args...)
}
func (d campaignRulesDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}
func (d campaignRulesDB) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}

func TestCreateCampaign_RulesModeWriteFailure(t *testing.T) {
	for _, tc := range []struct {
		name string
		tag  string
		err  error
	}{
		{"write error", "", errors.New("rules write failed")},
		{"missing row", "UPDATE 0", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := &CampaignHandlers{Queries: &stubQuerier{}, Logger: log.Default(), Pool: campaignRulesDB{exec: func(context.Context, string, ...any) (pgconn.CommandTag, error) {
				return pgconn.NewCommandTag(tc.tag), tc.err
			}}}
			rec := httptest.NewRecorder()
			newTestRouter(h).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/campaigns", bytes.NewBufferString(`{"name":"Test","rules_mode":"crunch"}`)))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("got %d: %s, want 500", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateCampaign_RulesModeResponseUsesSavedRow(t *testing.T) {
	q := &stubQuerier{campaignByID: map[string]statedb.Campaign{}}
	h := &CampaignHandlers{Queries: q, Logger: log.Default(), Pool: campaignRulesDB{exec: func(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
		id := args[1].(pgtype.UUID)
		q.campaignByID[dbutil.FromPgtype(id).String()] = statedb.Campaign{ID: id, Name: "Test", Status: "active", RulesMode: "crunch"}
		return pgconn.NewCommandTag("UPDATE 1"), nil
	}}}
	rec := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/campaigns", bytes.NewBufferString(`{"name":"Test","rules_mode":"crunch"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("got %d: %s", rec.Code, rec.Body.String())
	}
	var resp api.CampaignResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.RulesMode != "crunch" {
		t.Fatalf("got rules_mode %q", resp.RulesMode)
	}
}

func TestCreateCampaign_RulesModeRequiresPersistence(t *testing.T) {
	h := &CampaignHandlers{Queries: &stubQuerier{}, Logger: log.Default()}
	rec := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/campaigns", bytes.NewBufferString(`{"name":"Test","rules_mode":"crunch"}`)))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("got %d: %s, want 500", rec.Code, rec.Body.String())
	}
}
