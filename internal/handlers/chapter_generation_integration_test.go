//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

type chapterModel struct {
	mu       sync.Mutex
	calls    [][]llm.Message
	complete func(context.Context) (*llm.Response, error)
}

func (m *chapterModel) Complete(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	m.mu.Lock()
	m.calls = append(m.calls, messages)
	fn := m.complete
	m.mu.Unlock()
	if len(tools) != 0 {
		return nil, errors.New("unexpected tools")
	}
	if fn != nil {
		return fn(ctx)
	}
	return &llm.Response{Content: "Ada tried the gate. It remained closed."}, nil
}
func (m *chapterModel) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("unexpected stream")
}
func (m *chapterModel) count() int { m.mu.Lock(); defer m.mu.Unlock(); return len(m.calls) }
func (m *chapterModel) set(fn func(context.Context) (*llm.Response, error)) {
	m.mu.Lock()
	m.complete = fn
	m.mu.Unlock()
}

func TestIntegrationDurableChapters(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	campaign, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Chapters", Status: "active", CreatedBy: db.ToPgUUID(testUserID), RulesVersion: "srd-5.2.1"})
	if err != nil {
		t.Fatal(err)
	}
	cid := db.FromPgUUID(campaign.ID)
	t.Cleanup(func() {
		if _, err := testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, cid); err != nil {
			t.Error(err)
		}
	})
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := testPool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	// Private tools and unrevealed lore are poison markers, never model input.
	exec(`INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,1,'Try the gate','game_action','The gate remained closed.','[{"secret":"HIDDEN_TOOL"}]'),($1,2,'Search ahead','game_action','UNCOMMITTED_NARRATION','[]')`, cid)
	exec(`INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status,final_narrative) VALUES($1,$2,1,'Try the gate','completed','The gate remained closed.')`, uuid.New(), cid)
	exec(`INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status) VALUES($1,$2,2,'Search ahead','pending_narration')`, uuid.New(), cid)
	store := journal.NewStore(testPool)
	if _, err := store.CreateSummary(ctx, cid, 1, 900, "LEGACY_SUMMARY_MUST_NOT_DEFINE_COVERAGE"); err != nil {
		t.Fatal(err)
	}
	model := &chapterModel{}
	makeRouter := func() http.Handler {
		return server.NewRouterWithProvider(log.New(io.Discard), &stubIntegrationEngine{}, q, model, testPool, testUserID, config.Config{}, nil)
	}
	router := makeRouter()
	base := "/api/v1/campaigns/" + cid.String() + "/journal/chapters"
	request := func(handler http.Handler, body string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest("POST", base, strings.NewReader(body)))
		return rec
	}
	generate := func(key, chapter string) api.ChapterRevisionResponse {
		t.Helper()
		body, _ := json.Marshal(api.GenerateChapterRequest{RequestID: key, ChapterID: chapter})
		rec := request(router, string(body))
		if rec.Code != 200 {
			t.Fatalf("generate=%d %s", rec.Code, rec.Body.String())
		}
		return decodeJSON[api.ChapterRevisionResponse](t, rec)
	}
	list := func() []api.ChapterResponse {
		t.Helper()
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", base, nil))
		if rec.Code != 200 {
			t.Fatalf("list=%d %s", rec.Code, rec.Body.String())
		}
		return decodeJSON[[]api.ChapterResponse](t, rec)
	}
	for _, body := range []string{`{`, `{}`, `{"request_id":"bad"}`, `{"request_id":"` + uuid.NewString() + `","from_turn":1}`, `{"request_id":"` + uuid.NewString() + `"} {}`} {
		if rec := request(router, body); rec.Code != 400 {
			t.Fatalf("invalid request %s = %d", body, rec.Code)
		}
	}
	key := uuid.NewString()
	first := generate(key, "")
	if first.Status != "completed" || first.Revision != 1 {
		t.Fatalf("first=%+v", first)
	}
	chapters := list()
	if len(chapters) != 1 || chapters[0].FromTurn != 1 || chapters[0].ToTurn != 1 || chapters[0].RulesVersion != "srd-5.2.1" || chapters[0].KnowledgeBoundary == "" || chapters[0].SourceHash == "" || len(chapters[0].SourceTurns) != 1 || chapters[0].SourceTurns[0].AttemptID == "" {
		t.Fatalf("chapters=%+v", chapters)
	}
	original := chapters[0]
	model.mu.Lock()
	prompt := model.calls[0][1].Content
	system := model.calls[0][0].Content
	model.mu.Unlock()
	for _, secret := range []string{"HIDDEN_TOOL", "UNCOMMITTED_NARRATION", "LEGACY_SUMMARY"} {
		if strings.Contains(prompt, secret) {
			t.Fatalf("private/uncommitted source leaked: %s", secret)
		}
	}
	if !strings.Contains(prompt, "The gate remained closed.") || !strings.Contains(system, "Do not invent events, dialogue") {
		t.Fatal("missing source or fidelity instructions")
	}
	// A fresh service instance plus a newly completed turn must still replay the key.
	exec(`UPDATE turn_attempts SET status='completed',final_narrative='You found a path.' WHERE campaign_id=$1 AND turn_number=2`, cid)
	exec(`UPDATE session_logs SET llm_response='You found a path.' WHERE campaign_id=$1 AND turn_number=2`, cid)
	router = makeRouter()
	replay := generate(key, "")
	if replay.ID != first.ID || model.count() != 1 {
		t.Fatalf("retry duplicated generation: %+v calls=%d", replay, model.count())
	}
	// Regeneration reads the frozen source even if legacy transcript data is edited.
	exec(`UPDATE session_logs SET llm_response='LATER_EDIT_NOT_SOURCE' WHERE campaign_id=$1 AND turn_number=1`, cid)
	second := generate(uuid.NewString(), first.ChapterID)
	if second.Revision != 2 || second.ChapterID != first.ChapterID {
		t.Fatalf("regeneration=%+v", second)
	}
	model.mu.Lock()
	regenPrompt := model.calls[1][1].Content
	model.mu.Unlock()
	if strings.Contains(regenPrompt, "LATER_EDIT") || strings.Contains(regenPrompt, "found a path") {
		t.Fatal("regeneration source changed")
	}
	if got := list()[0]; got.SourceHash != original.SourceHash || len(got.Revisions) != 2 || got.Revisions[0].Content != first.Content {
		t.Fatalf("source/original revision changed: %+v", got)
	}
	// Failure is durable and sanitized. Same-key retry completes the same revision.
	model.set(func(context.Context) (*llm.Response, error) { return nil, errors.New("PRIVATE_PROVIDER_ERROR") })
	nextKey := uuid.NewString()
	failed := generate(nextKey, "")
	if failed.Status != "failed" || strings.Contains(failed.Error, "PRIVATE_PROVIDER") || failed.ChapterID == first.ChapterID {
		t.Fatalf("failure=%+v", failed)
	}
	model.set(nil)
	recovered := generate(nextKey, failed.ChapterID)
	if recovered.ID != failed.ID || recovered.Revision != 1 || recovered.Status != "completed" {
		t.Fatalf("recovery=%+v", recovered)
	}
	if rec := request(router, `{"request_id":"`+key+`","chapter_id":"`+failed.ChapterID+`"}`); rec.Code != 409 {
		t.Fatalf("key rebinding=%d", rec.Code)
	}
	if rec := request(router, `{"request_id":"`+uuid.NewString()+`","chapter_id":"`+uuid.NewString()+`"}`); rec.Code != 404 {
		t.Fatalf("unknown chapter=%d", rec.Code)
	}
	noMore := request(router, `{"request_id":"`+uuid.NewString()+`"}`)
	if noMore.Code != 200 || !strings.Contains(noMore.Body.String(), "no completed turns") {
		t.Fatalf("empty range=%d %s", noMore.Code, noMore.Body.String())
	}
	// Legacy summary survives, and chapter work never writes the transcript.
	summaries, err := store.ListSummaries(ctx, cid)
	if err != nil || len(summaries) != 1 || summaries[0].ToTurn != 900 {
		t.Fatalf("legacy summaries=%+v %v", summaries, err)
	}
	var transcript string
	if err := testPool.QueryRow(ctx, `SELECT llm_response FROM session_logs WHERE campaign_id=$1 AND turn_number=1`, cid).Scan(&transcript); err != nil || transcript != "LATER_EDIT_NOT_SOURCE" {
		t.Fatalf("transcript mutated: %s %v", transcript, err)
	}
	// Another campaign's otherwise valid parent cannot select this chapter.
	other, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Other", Status: "active", CreatedBy: db.ToPgUUID(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	defer exec(`DELETE FROM campaigns WHERE id=$1`, other.ID)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest("POST", "/api/v1/campaigns/"+db.FromPgUUID(other.ID).String()+"/journal/chapters", strings.NewReader(`{"request_id":"`+uuid.NewString()+`","chapter_id":"`+first.ChapterID+`"}`)))
	if rec.Code != 404 {
		t.Fatalf("foreign chapter=%d %s", rec.Code, rec.Body.String())
	}

}

func TestIntegrationChapterConcurrentRetryAndLease(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	campaign, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Concurrent chapters", Status: "active", CreatedBy: db.ToPgUUID(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	cid := db.FromPgUUID(campaign.ID)
	defer func() {
		if _, err := testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, cid); err != nil {
			t.Error(err)
		}
	}()
	if _, err := testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response) VALUES($1,1,'wait','game_action','The gate stayed shut.')`, cid); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	model := &chapterModel{complete: func(context.Context) (*llm.Response, error) {
		close(entered)
		<-release
		return &llm.Response{Content: "Stale response"}, nil
	}}
	store := journal.NewStore(testPool)
	service := journal.NewSummarizer(model, store)
	key := uuid.New()
	done := make(chan error, 1)
	go func() { _, err := service.GenerateChapter(ctx, cid, key, ""); done <- err }()
	<-entered
	// Always release this test's model even if an assertion fails.
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	pending, err := journal.NewSummarizer(model, store).GenerateChapter(ctx, cid, key, "")
	if err != nil || pending.Status != "generating" || model.count() != 1 {
		t.Fatalf("concurrent retry=%+v %v", pending, err)
	}
	if _, err := service.GenerateChapter(ctx, cid, uuid.New(), ""); !errors.Is(err, journal.ErrChapterBusy) {
		t.Fatalf("competing key=%v", err)
	}
	// Simulate a dead worker's lease expiring, with no wall-clock sleep.
	if _, err := testPool.Exec(ctx, `UPDATE chapter_revisions SET lease_until=now()-interval '1 second' WHERE request_id=$1`, key); err != nil {
		t.Fatal(err)
	}
	model.set(nil)
	recovered, err := journal.NewSummarizer(model, store).GenerateChapter(ctx, cid, key, "")
	if err != nil || recovered.ID != pending.ID || recovered.Status != "completed" {
		t.Fatalf("lease recovery=%+v %v", recovered, err)
	}
	close(release)
	if err := <-done; !errors.Is(err, journal.ErrChapterBusy) {
		t.Fatal("stale worker overwrote the reclaimed revision")
	}
	chapters, err := store.ListChapters(ctx, cid)
	if err != nil || len(chapters) != 1 || len(chapters[0].Revisions) != 1 || chapters[0].Revisions[0].Content == "Stale response" {
		t.Fatalf("fencing=%+v %v", chapters, err)
	}

	// Start Over retains immutable chapters, but reused turn numbers belong to
	// new log identities and must not inherit earlier coverage.
	if err := saves.NewStore(testPool).StartOver(ctx, cid); err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response) VALUES($1,1,'look','game_action','A different beginning.')`, cid); err != nil {
		t.Fatal(err)
	}
	restarted, err := service.GenerateChapter(ctx, cid, uuid.New(), "")
	if err != nil || restarted.ChapterID == chapters[0].ID {
		t.Fatalf("reset coverage=%+v %v", restarted, err)
	}
	all, err := store.ListChapters(ctx, cid)
	if err != nil || len(all) != 2 || all[0].SourceHash == all[1].SourceHash {
		t.Fatalf("reset snapshots=%+v %v", all, err)
	}

	// Cancellation is durably isolated; an empty response is also retryable.
	model.set(func(context.Context) (*llm.Response, error) { return nil, context.Canceled })
	canceled, err := service.GenerateChapter(ctx, cid, uuid.New(), chapters[0].ID)
	if err != nil || canceled.Status != "failed" {
		t.Fatalf("canceled=%+v %v", canceled, err)
	}
	model.set(func(context.Context) (*llm.Response, error) { return &llm.Response{Content: " "}, nil })
	empty, err := service.GenerateChapter(ctx, cid, uuid.New(), chapters[0].ID)
	if err != nil || empty.Status != "failed" {
		t.Fatalf("empty=%+v %v", empty, err)
	}
}
