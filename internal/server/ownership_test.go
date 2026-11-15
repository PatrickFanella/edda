package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

type ownershipQueries struct {
	statedb.Querier // Unexpected downstream queries panic, making a missing guard fail.
	campaigns       map[pgtype.UUID]statedb.Campaign
	lookupErr       error
	mutations       int
}

func (q *ownershipQueries) GetCampaignByID(_ context.Context, id pgtype.UUID) (statedb.Campaign, error) {
	if q.lookupErr != nil {
		return statedb.Campaign{}, q.lookupErr
	}
	c, ok := q.campaigns[id]
	if !ok {
		return c, pgx.ErrNoRows
	}
	return c, nil
}
func (q *ownershipQueries) UpdateCampaign(_ context.Context, arg statedb.UpdateCampaignParams) (statedb.Campaign, error) {
	q.mutations++
	c := q.campaigns[arg.ID]
	c.Name = arg.Name
	q.campaigns[arg.ID] = c
	return c, nil
}
func (q *ownershipQueries) DeleteCampaign(_ context.Context, id pgtype.UUID) error {
	q.mutations++
	delete(q.campaigns, id)
	return nil
}
func (q *ownershipQueries) GetQuestByID(_ context.Context, arg statedb.GetQuestByIDParams) (statedb.Quest, error) {
	return statedb.Quest{}, pgx.ErrNoRows
}
func (q *ownershipQueries) GetNPCByID(_ context.Context, arg statedb.GetNPCByIDParams) (statedb.Npc, error) {
	return statedb.Npc{}, pgx.ErrNoRows
}
func (q *ownershipQueries) GetLocationByID(_ context.Context, arg statedb.GetLocationByIDParams) (statedb.Location, error) {
	return statedb.Location{}, pgx.ErrNoRows
}

type ownershipEngine struct {
	engine.GameEngine
	turns atomic.Int32
}

func (e *ownershipEngine) ProcessTurn(context.Context, uuid.UUID, string) (*engine.TurnResult, error) {
	e.turns.Add(1)
	return &engine.TurnResult{Narrative: "Owner turn"}, nil
}
func (e *ownershipEngine) ProcessTurnStream(context.Context, uuid.UUID, string) (<-chan engine.StreamEvent, error) {
	e.turns.Add(1)
	ch := make(chan engine.StreamEvent, 1)
	ch <- engine.StreamEvent{Type: "result", Result: &engine.TurnResult{Narrative: "Owner stream"}}
	close(ch)
	return ch, nil
}

func ownershipFixture(t *testing.T) (http.Handler, *ownershipQueries, *ownershipEngine, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	q := &ownershipQueries{campaigns: map[pgtype.UUID]statedb.Campaign{db.ToPgUUID(c): {ID: db.ToPgUUID(c), CreatedBy: db.ToPgUUID(a), Name: "Private campaign"}}}
	e := &ownershipEngine{}
	cfg := config.Config{}
	cfg.Server.JWTSecret = "ownership-test-secret"
	return NewRouterWithProvider(log.New(io.Discard), e, q, nil, nil, a, cfg, nil), q, e, a, b, c
}
func ownershipToken(t *testing.T, user uuid.UUID) string {
	t.Helper()
	token, err := auth.GenerateToken(user, "ownership-test-secret", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}
func ownershipRequest(t *testing.T, h http.Handler, user uuid.UUID, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+ownershipToken(t, user))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCampaignOwnershipEveryRoute(t *testing.T) {
	h, q, e, a, b, c := ownershipFixture(t)
	// Walk the production router so additions to the campaign subtree are covered.
	count := 0
	err := chi.Walk(h.(chi.Routes), func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if !strings.Contains(route, "/campaigns/{id}") {
			return nil
		}
		count++
		path := strings.ReplaceAll(route, "{id}", c.String())
		for _, key := range []string{"lid", "nid", "qid", "noteID", "eid"} {
			path = strings.ReplaceAll(path, "{"+key+"}", uuid.NewString())
		}
		t.Run(method+" "+route, func(t *testing.T) {
			for _, user := range []uuid.UUID{b, a} {
				// Both users are foreign to this ID: B sees A's existing campaign;
				// A sees an absent ID. Neither response discloses existence.
				p := path
				if user == a {
					p = strings.ReplaceAll(p, c.String(), uuid.NewString())
				}
				rec := ownershipRequest(t, h, user, method, p, `{"input":"look","content":"note"}`)
				if rec.Code != http.StatusNotFound || rec.Body.String() != "{\"error\":\"campaign not found\"}\n" {
					t.Fatalf("got %d %s", rec.Code, rec.Body.String())
				}
			}
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count < 40 {
		t.Fatalf("only covered %d campaign routes", count)
	}
	if q.mutations != 0 || e.turns.Load() != 0 {
		t.Fatal("denied requests changed state")
	}
}

func TestCampaignOwnershipOwnerAndFailures(t *testing.T) {
	h, q, e, a, _, c := ownershipFixture(t)
	base := "/api/v1/campaigns/" + c.String()
	for _, tc := range []struct {
		method, path, body string
		want               int
	}{
		{"GET", base, "", 200}, {"PUT", base, `{"name":"Updated"}`, 200}, {"POST", base + "/action", `{"input":"look"}`, 200},
		{"GET", "/api/v1/campaigns/not-a-uuid", "", 400},
	} {
		rec := ownershipRequest(t, h, a, tc.method, tc.path, tc.body)
		if rec.Code != tc.want {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
	if e.turns.Load() != 1 || q.mutations != 1 {
		t.Fatal("owner actions did not run")
	}
	q.lookupErr = errors.New("database offline")
	if rec := ownershipRequest(t, h, a, "DELETE", base, ""); rec.Code != 500 {
		t.Fatalf("lookup failure: %d", rec.Code)
	}
	q.lookupErr = nil
	if rec := ownershipRequest(t, h, a, "DELETE", base, ""); rec.Code != 204 {
		t.Fatalf("owner delete: %d", rec.Code)
	}
}

func TestCampaignOwnershipNestedResources(t *testing.T) {
	h, _, _, a, _, c := ownershipFixture(t)
	base := "/api/v1/campaigns/" + c.String()
	id := uuid.NewString()
	for _, tc := range []struct{ method, path string }{
		{"GET", "/quests/" + id}, {"GET", "/quests/" + id + "/notes"}, {"POST", "/quests/" + id + "/notes"}, {"DELETE", "/quests/" + id + "/notes/" + uuid.NewString()}, {"GET", "/quests/" + id + "/history"},
		{"GET", "/locations/" + id}, {"GET", "/npcs/" + id}, {"GET", "/npcs/" + id + "/dialogue"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			rec := ownershipRequest(t, h, a, tc.method, base+tc.path, `{"content":"note"}`)
			if rec.Code != 404 {
				t.Fatalf("got %d %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCampaignOwnershipWebSocket(t *testing.T) {
	h, _, e, a, b, c := ownershipFixture(t)
	srv := httptest.NewServer(h)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/campaigns/" + c.String() + "/ws"
	for _, user := range []uuid.UUID{b, a} {
		// Browser WebSockets authenticate with the session cookie.
		headers := http.Header{}
		headers.Set("Cookie", auth.AuthCookieName+"="+ownershipToken(t, user))
		conn, resp, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: headers})
		if user == b {
			if err == nil {
				_ = conn.CloseNow()
				t.Fatal("foreign upgrade accepted")
			}
			if resp == nil || resp.StatusCode != 404 {
				t.Fatalf("foreign upgrade response: %v %v", resp, err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = conn.CloseNow() }()
		if err := wsjson.Write(ctx, conn, map[string]any{"type": "action", "payload": map[string]string{"input": "look"}}); err != nil {
			t.Fatal(err)
		}
		var result map[string]any
		if err := wsjson.Read(ctx, conn, &result); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(fmt.Sprint(result), "Owner stream") {
			t.Fatalf("unexpected stream: %v", result)
		}
	}
	if e.turns.Load() != 1 {
		t.Fatalf("turns: %d", e.turns.Load())
	}
}

func TestCampaignOwnershipNoOpMode(t *testing.T) {
	_, q, e, a, b, c := ownershipFixture(t)
	for _, user := range []uuid.UUID{a, b} {
		h := NewRouterWithProvider(log.New(io.Discard), e, q, nil, nil, user, config.Config{}, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/campaigns/"+c.String(), nil))
		want := 200
		if user == b {
			want = 404
		}
		if rec.Code != want {
			t.Fatalf("default user %s: %d, want %d", user, rec.Code, want)
		}
	}
}
