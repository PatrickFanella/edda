//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	campaignexport "git.subcult.tv/subculture-collective/edda/internal/export"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

func TestIntegrationSavedChapterExport(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := testPool.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	create := func(name string) string {
		t.Helper()
		campaign, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: name, Status: "active", CreatedBy: db.ToPgUUID(testUserID), RulesVersion: "srd-5.2.1"})
		if err != nil {
			t.Fatal(err)
		}
		cid := db.FromPgUUID(campaign.ID).String()
		exec(`INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,1,'Try the gate','game_action','The gate remained closed.','[{"secret":"HIDDEN_TOOL"}]')`, cid)
		exec(`INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status,final_narrative) VALUES($1,$2,1,'Try the gate','completed','The gate remained closed.')`, uuid.New(), cid)
		return cid
	}
	cid, foreign := create("Export fixture"), create("Foreign export fixture")
	exec(`UPDATE campaigns SET description='HIDDEN_CAMPAIGN_DESCRIPTION',rules_mode='light' WHERE id=$1`, cid)
	exec(`INSERT INTO world_facts(campaign_id,fact,category,source) VALUES($1,'HIDDEN_LORE','secret','worldgen')`, cid)
	model := &chapterModel{}
	const secret = "isolated-export-test-secret"
	cfg := config.Config{Server: config.ServerConfig{JWTSecret: secret}}
	router := server.NewRouterWithProvider(log.New(io.Discard), &stubIntegrationEngine{}, q, model, testPool, testUserID, cfg, nil)
	token, err := auth.GenerateToken(testUserID, secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	outsider, err := auth.GenerateToken(uuid.New(), secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, body, bearer string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	generate := func(campaign, chapter string) api.ChapterRevisionResponse {
		t.Helper()
		body, err := json.Marshal(api.GenerateChapterRequest{RequestID: uuid.NewString(), ChapterID: chapter})
		if err != nil {
			t.Fatal(err)
		}
		rec := request("POST", "/api/v1/campaigns/"+campaign+"/journal/chapters", string(body), token)
		if rec.Code != 200 {
			t.Fatalf("generation = %d %s", rec.Code, rec.Body.String())
		}
		return decodeJSON[api.ChapterRevisionResponse](t, rec)
	}
	first := generate(cid, "")
	// A different saved revision must never silently replace the selected prose.
	model.set(func(context.Context) (*llm.Response, error) {
		return &llm.Response{Content: "NEWER_UNSELECTED_PROSE"}, nil
	})
	second := generate(cid, first.ChapterID)
	foreignRevision := generate(foreign, "")
	exec(`INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response) VALUES($1,2,'Continue along the wall','game_action','SECOND_CHAPTER_SOURCE')`, cid)
	model.set(func(context.Context) (*llm.Response, error) {
		return &llm.Response{Content: "UNSELECTED_CHAPTER_PROSE"}, nil
	})
	nextChapter := generate(cid, "")
	pendingID := uuid.NewString()
	exec(`INSERT INTO chapter_revisions(id,chapter_id,campaign_id,request_id,revision,status,content,error,prompt_version,lease_token,lease_until) VALUES($1,$2,$3,$4,3,'failed','FAILED_PROSE','FAILED_DETAILS','fixture', $5, now())`, pendingID, first.ChapterID, cid, uuid.New(), uuid.New())
	// Source snapshot survives transcript replacement and current rules changes.
	exec(`UPDATE session_logs SET player_input='LATER_INPUT',llm_response='LATER_NARRATION' WHERE campaign_id=$1`, cid)
	exec(`INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response) VALUES($1,3,'Unfinished action','game_action','PENDING_NARRATION')`, cid)
	exec(`INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status) VALUES($1,$2,3,'Unfinished action','pending_narration')`, uuid.New(), cid)
	exec(`UPDATE campaigns SET rules_mode='narrative' WHERE id=$1`, cid)
	reader := request("GET", "/api/v1/campaigns/"+cid+"/journal/chapters", "", token)
	if reader.Code != 200 {
		t.Fatal(reader.Body.String())
	}
	expected := decodeJSON[[]api.ChapterResponse](t, reader)[0]
	expected.Revisions = []api.ChapterRevisionResponse{first}
	calls := model.count()
	base := "/api/v1/campaigns/" + cid + "/export/chapters"
	selected := "revision_id=" + first.ID
	for _, format := range []string{"json", "markdown"} {
		t.Run(format, func(t *testing.T) {
			rec := request("GET", base+"?"+selected+"&format="+format, "", token)
			if rec.Code != 200 {
				t.Fatalf("export = %d %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Header().Get("Cache-Control"), "no-store") || !strings.Contains(rec.Header().Get("Content-Disposition"), "attachment;") {
				t.Fatal("download/cache headers missing")
			}
			body := rec.Body.String()
			for _, poison := range []string{"HIDDEN_TOOL", "HIDDEN_LORE", "HIDDEN_CAMPAIGN_DESCRIPTION", "LATER_INPUT", "LATER_NARRATION", "PENDING_NARRATION", "NEWER_UNSELECTED_PROSE", "FAILED_PROSE", "FAILED_DETAILS", "SECOND_CHAPTER_SOURCE", "UNSELECTED_CHAPTER_PROSE", nextChapter.ID, foreignRevision.ID} {
				if strings.Contains(body, poison) {
					t.Errorf("export leaked %s", poison)
				}
			}
			if format == "json" {
				got := decodeJSON[campaignexport.ChapterExport](t, rec)
				if got.CampaignID != cid || !reflect.DeepEqual(got.Chapters, []api.ChapterResponse{expected}) {
					t.Fatalf("export differs from selected reader snapshot: %+v", got)
				}
				if got.Attribution != srd521.Attribution || got.Changes == "" {
					t.Fatal("missing attribution/change notice")
				}
			} else {
				for _, text := range []string{first.Content, "The gate remained closed.", first.ID, first.RequestID, expected.SourceHash, expected.SourceTurns[0].LogID, expected.SourceTurns[0].AttemptID, expected.KnowledgeBoundary, expected.RulesVersion, expected.RulesMode, first.PromptVersion, srd521.Attribution, "Turns 1–1"} {
					if text == "" || !strings.Contains(body, text) {
						t.Errorf("markdown missing %q", text)
					}
				}
			}
		})
	}
	// Selecting two chapters includes both frozen ranges and only selected prose.
	recBoth := request("GET", base+"?"+selected+"&revision_id="+nextChapter.ID+"&format=json", "", token)
	if recBoth.Code != 200 {
		t.Fatalf("multi-chapter export=%d %s", recBoth.Code, recBoth.Body.String())
	}
	both := decodeJSON[campaignexport.ChapterExport](t, recBoth)
	if len(both.Chapters) != 2 || !reflect.DeepEqual(both.Chapters[0], expected) || both.Chapters[1].ID != nextChapter.ChapterID || both.Chapters[1].FromTurn != 2 || both.Chapters[1].Revisions[0].Content != nextChapter.Content {
		t.Fatalf("multi-chapter selection changed: %+v", both)
	}
	// Valid multi-revision selection is explicit and retains both saved versions.
	rec := request("GET", base+"?"+selected+"&revision_id="+second.ID+"&format=json", "", token)
	if rec.Code != 200 || len(decodeJSON[campaignexport.ChapterExport](t, rec).Chapters[0].Revisions) != 2 {
		t.Fatal("explicit revision selection failed")
	}
	for _, tc := range []struct {
		name, query, bearer string
		want                int
	}{
		{"anonymous", selected, "", 401},
		{"other owner", selected, outsider, 404},
		{"empty selection", "", token, 400},
		{"oversized selection", strings.Repeat(selected+"&", 101), token, 400},
		{"invalid id", "revision_id=invalid", token, 400},
		{"nil id", "revision_id=" + uuid.Nil.String(), token, 400},
		{"duplicate", selected + "&" + selected, token, 400},
		{"missing", "revision_id=" + uuid.NewString(), token, 404},
		{"foreign revision", "revision_id=" + foreignRevision.ID, token, 404},
		{"mixed foreign", selected + "&revision_id=" + foreignRevision.ID, token, 404},
		{"failed revision", "revision_id=" + pendingID, token, 404},
		{"unsupported format", selected + "&format=html", token, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := request("GET", base+"?"+tc.query, "", tc.bearer)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want %d: %s", rec.Code, tc.want, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), first.Content) || rec.Header().Get("Content-Disposition") != "" {
				t.Fatal("partial export on rejected request")
			}
		})
	}
	// Unfinished revisions are not readable exports either.
	exec(`UPDATE chapter_revisions SET status='generating' WHERE id=$1`, pendingID)
	if rec := request("GET", base+"?revision_id="+pendingID, "", token); rec.Code != 404 {
		t.Fatalf("generating export=%d", rec.Code)
	}
	if model.count() != calls {
		t.Fatal("export invoked generation")
	}
	var revisions int
	if err := testPool.QueryRow(ctx, `SELECT count(*) FROM chapter_revisions WHERE campaign_id=$1`, cid).Scan(&revisions); err != nil || revisions != 4 {
		t.Fatalf("export mutated revisions: count=%d err=%v", revisions, err)
	}
}
