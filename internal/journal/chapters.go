package journal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

const chapterPromptVersion = "faithful-chapter-v1"
const chapterKnowledgeBoundary = "player-visible-transcript-through-to-turn-v1"

var (
	// ErrChapterBusy means another request holds the chapter's generation lease.
	ErrChapterBusy = errors.New("chapter generation is already in progress")
	// ErrChapterRequest means a request key was reused for a different source.
	ErrChapterRequest = errors.New("request key belongs to a different chapter")
)

const chapterColumns = `id::text, campaign_id::text, from_turn, to_turn, source_hash, source_turns, knowledge_boundary, rules_version, rules_mode, created_at`
const revisionColumns = `id::text, chapter_id::text, request_id::text, revision, status, content, error, prompt_version, created_at`

type scanner interface{ Scan(...any) error }

func scanChapter(row scanner) (api.ChapterResponse, error) {
	var c api.ChapterResponse
	var source []byte
	err := row.Scan(&c.ID, &c.CampaignID, &c.FromTurn, &c.ToTurn, &c.SourceHash, &source, &c.KnowledgeBoundary, &c.RulesVersion, &c.RulesMode, &c.CreatedAt)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(source, &c.SourceTurns)
	c.Revisions = []api.ChapterRevisionResponse{}
	return c, err
}

func scanRevision(row scanner) (api.ChapterRevisionResponse, error) {
	var r api.ChapterRevisionResponse
	err := row.Scan(&r.ID, &r.ChapterID, &r.RequestID, &r.Revision, &r.Status, &r.Content, &r.Error, &r.PromptVersion, &r.CreatedAt)
	return r, err
}

// ListChapters returns immutable sources and every revision, including failed attempts.
func (s *Store) ListChapters(ctx context.Context, campaignID uuid.UUID) ([]api.ChapterResponse, error) {
	rows, err := s.db.Query(ctx, `SELECT `+chapterColumns+` FROM chapters WHERE campaign_id=$1 ORDER BY created_at, from_turn`, campaignID)
	if err != nil {
		return nil, err
	}
	chapters := []api.ChapterResponse{}
	for rows.Next() {
		c, err := scanChapter(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		chapters = append(chapters, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Close the source query before loading revisions, also with a one-connection pool.
	for i := range chapters {
		rows, err := s.db.Query(ctx, `SELECT `+revisionColumns+` FROM chapter_revisions WHERE chapter_id=$1 ORDER BY revision`, chapters[i].ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			r, err := scanRevision(rows)
			if err != nil {
				rows.Close()
				return nil, err
			}
			chapters[i].Revisions = append(chapters[i].Revisions, r)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}
	return chapters, nil
}

// reserveChapter serializes only source selection and lease reservation. Inference
// happens after commit, so no campaign row lock is held while the model runs.
func (s *Store) reserveChapter(ctx context.Context, campaignID, requestID uuid.UUID, chapterID string) (api.ChapterResponse, api.ChapterRevisionResponse, uuid.UUID, error) {
	var c api.ChapterResponse
	var r api.ChapterRevisionResponse
	beginner, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return c, r, uuid.Nil, errors.New("chapter transactions unavailable")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return c, r, uuid.Nil, err
	}
	defer func() {
		cleanup, cancel := chapterCleanupContext()
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	var version, mode string
	if err := tx.QueryRow(ctx, `SELECT rules_version, rules_mode FROM campaigns WHERE id=$1 FOR UPDATE`, campaignID).Scan(&version, &mode); err != nil {
		return c, r, uuid.Nil, err
	}
	var live bool
	r, err = scanRevision(tx.QueryRow(ctx, `SELECT `+revisionColumns+` FROM chapter_revisions WHERE campaign_id=$1 AND request_id=$2`, campaignID, requestID))
	if err == nil {
		if chapterID != "" && chapterID != r.ChapterID {
			return c, r, uuid.Nil, ErrChapterRequest
		}
		c, err = scanChapter(tx.QueryRow(ctx, `SELECT `+chapterColumns+` FROM chapters WHERE id=$1 AND campaign_id=$2`, r.ChapterID, campaignID))
		if err != nil {
			return c, r, uuid.Nil, err
		}
		if r.Status == "completed" {
			return c, r, uuid.Nil, tx.Commit(ctx)
		}
		if err = tx.QueryRow(ctx, `SELECT lease_until > now() FROM chapter_revisions WHERE id=$1`, r.ID).Scan(&live); err != nil {
			return c, r, uuid.Nil, err
		}
		if r.Status == "generating" && live {
			return c, r, uuid.Nil, tx.Commit(ctx)
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return c, r, uuid.Nil, err
	} else {
		if chapterID != "" {
			c, err = scanChapter(tx.QueryRow(ctx, `SELECT `+chapterColumns+` FROM chapters WHERE id=$1 AND campaign_id=$2`, chapterID, campaignID))
		} else {
			// A failed first generation keeps its source reserved for the next request.
			c, err = scanChapter(tx.QueryRow(ctx, `SELECT `+chapterColumns+` FROM chapters c WHERE campaign_id=$1 AND EXISTS (SELECT 1 FROM session_logs l WHERE l.campaign_id=c.campaign_id AND l.id=(c.source_turns->-1->>'log_id')::uuid) AND NOT EXISTS (SELECT 1 FROM chapter_revisions r WHERE r.chapter_id=c.id AND r.status='completed') ORDER BY from_turn LIMIT 1`, campaignID))
			if errors.Is(err, pgx.ErrNoRows) {
				c, err = captureChapter(ctx, tx, campaignID, version, mode)
			}
		}
		if err != nil {
			return c, r, uuid.Nil, err
		}
		if c.ID == "" {
			return c, r, uuid.Nil, tx.Commit(ctx)
		}
	}
	// One live generation per source, even when concurrent callers use different keys.
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM chapter_revisions WHERE chapter_id=$1 AND status='generating' AND lease_until>now() AND request_id<>$2)`, c.ID, requestID).Scan(&live); err != nil {
		return c, r, uuid.Nil, err
	}
	if live {
		return c, r, uuid.Nil, ErrChapterBusy
	}
	if _, err := tx.Exec(ctx, `UPDATE chapter_revisions SET status='failed',error='Generation interrupted; retry this revision.' WHERE chapter_id=$1 AND status='generating' AND lease_until<=now()`, c.ID); err != nil {
		return c, r, uuid.Nil, err
	}
	token := uuid.New()
	if r.ID == "" {
		r, err = scanRevision(tx.QueryRow(ctx, `INSERT INTO chapter_revisions(chapter_id,campaign_id,request_id,revision,status,prompt_version,lease_token,lease_until) SELECT $1,$2,$3,COALESCE(MAX(revision),0)+1,'generating',$4,$5,now()+interval '2 minutes' FROM chapter_revisions WHERE chapter_id=$1 RETURNING `+revisionColumns, c.ID, campaignID, requestID, chapterPromptVersion, token))
	} else {
		r, err = scanRevision(tx.QueryRow(ctx, `UPDATE chapter_revisions SET status='generating',error='',lease_token=$2,lease_until=now()+interval '2 minutes' WHERE id=$1 RETURNING `+revisionColumns, r.ID, token))
	}
	if err != nil {
		return c, r, uuid.Nil, err
	}
	return c, r, token, tx.Commit(ctx)
}

func captureChapter(ctx context.Context, tx pgx.Tx, campaignID uuid.UUID, version, mode string) (api.ChapterResponse, error) {
	var c api.ChapterResponse
	var after int
	if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(to_turn),0) FROM chapters c WHERE campaign_id=$1 AND EXISTS (SELECT 1 FROM session_logs l WHERE l.campaign_id=c.campaign_id AND l.id=(c.source_turns->-1->>'log_id')::uuid)`, campaignID).Scan(&after); err != nil {
		return c, err
	}
	// A session log is the canonical durable player transcript. Legacy logs have no
	// turn_attempt; modern logs must have a completed receipt and no pending attempt.
	rows, err := tx.Query(ctx, `SELECT l.id::text,COALESCE(a.id::text,''),l.turn_number,l.player_input,l.llm_response,
  (length(trim(l.llm_response))>0 AND NOT EXISTS (SELECT 1 FROM turn_attempts p WHERE p.campaign_id=l.campaign_id AND p.turn_number=l.turn_number AND p.status IN ('pending_mechanics','pending_narration')) AND
   (a.id IS NOT NULL OR NOT EXISTS (SELECT 1 FROM turn_attempts p WHERE p.campaign_id=l.campaign_id AND p.turn_number=l.turn_number))) AS completed
  FROM session_logs l LEFT JOIN turn_attempts a ON a.campaign_id=l.campaign_id AND a.turn_number=l.turn_number AND a.status='completed'
  WHERE l.campaign_id=$1 AND l.turn_number>$2 ORDER BY l.turn_number LIMIT 100`, campaignID, after)
	if err != nil {
		return c, err
	}
	turns := []api.ChapterTurn{}
	for rows.Next() {
		var t api.ChapterTurn
		var completed bool
		if err := rows.Scan(&t.LogID, &t.AttemptID, &t.TurnNumber, &t.PlayerInput, &t.Narrative, &completed); err != nil {
			rows.Close()
			return c, err
		}
		if !completed || t.TurnNumber != after+len(turns)+1 {
			break
		}
		turns = append(turns, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return c, err
	}
	if len(turns) == 0 {
		return c, nil
	}
	source, hash, err := chapterSource(turns, version, mode)
	if err != nil {
		return c, err
	}
	return scanChapter(tx.QueryRow(ctx, `INSERT INTO chapters(campaign_id,from_turn,to_turn,source_hash,source_turns,knowledge_boundary,rules_version,rules_mode) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING `+chapterColumns, campaignID, turns[0].TurnNumber, turns[len(turns)-1].TurnNumber, hash, source, chapterKnowledgeBoundary, version, mode))
}

func chapterSource(turns []api.ChapterTurn, version, mode string) ([]byte, string, error) {
	source, err := json.Marshal(turns)
	if err != nil {
		return nil, "", err
	}
	canonical, err := json.Marshal(struct {
		Boundary, Version, Mode string
		Turns                   []api.ChapterTurn
	}{chapterKnowledgeBoundary, version, mode, turns})
	if err != nil {
		return nil, "", err
	}
	return source, fmt.Sprintf("%x", sha256.Sum256(canonical)), nil
}

func (s *Store) finishChapter(ctx context.Context, r api.ChapterRevisionResponse, token uuid.UUID, content, failure string) (api.ChapterRevisionResponse, error) {
	// Source reservation holds this same campaign lock. Finalization must join
	// that order so a next-range request cannot select an unfinished source and
	// then allocate a new revision after its worker concurrently completes it.
	beginner, ok := s.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return api.ChapterRevisionResponse{}, errors.New("chapter transactions unavailable")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return api.ChapterRevisionResponse{}, err
	}
	defer func() {
		cleanup, cancel := chapterCleanupContext()
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	var campaignID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT c.id FROM campaigns c JOIN chapter_revisions r ON r.campaign_id=c.id WHERE r.id=$1 FOR UPDATE OF c`, r.ID).Scan(&campaignID); err != nil {
		return api.ChapterRevisionResponse{}, err
	}
	status := "completed"
	if failure != "" {
		status = "failed"
	}
	result, err := scanRevision(tx.QueryRow(ctx, `UPDATE chapter_revisions SET status=$3,content=$4,error=$5 WHERE id=$1 AND lease_token=$2 AND status='generating' AND lease_until>now() RETURNING `+revisionColumns, r.ID, token, status, content, failure))
	if errors.Is(err, pgx.ErrNoRows) {
		return result, ErrChapterBusy
	}
	if err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}

// Bounded cleanup persists cancellation/failure after the request context is gone.
func chapterCleanupContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
