import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getSessionHistory } from '../../api/campaigns';
import type { ChapterResponse, ChapterRevisionResponse } from '../../api/types';
import { readOffset, readSelection, rememberOffset, rememberSelection } from './readingPosition';

function completedRevisions(chapter: ChapterResponse) {
  return chapter.revisions.filter(revision => revision.status === 'completed').sort((a, b) => a.revision - b.revision);
}

export function ChapterReader({ campaignId, chapters }: { readonly campaignId: string; readonly chapters: readonly ChapterResponse[] }) {
  const [selection, setSelection] = useState(() => readSelection(campaignId));
  const ordered = [...chapters].sort((a, b) => a.created_at.localeCompare(b.created_at) || a.from_turn - b.from_turn || a.to_turn - b.to_turn || a.id.localeCompare(b.id));
  const chapter = ordered.find(item => item.id === selection?.chapter) ?? ordered[0];
  const revisions = chapter ? completedRevisions(chapter) : [];
  const revision = revisions.find(item => item.id === selection?.revision) ?? revisions.at(-1);
  const history = useQuery({ queryKey: ['campaign', campaignId, 'history'], queryFn: () => getSessionHistory(campaignId), retry: false });
  useEffect(() => {
    if (chapter && revision) rememberSelection(campaignId, chapter.id, revision.id);
  }, [campaignId, chapter, revision]);
  const uncovered = history.data?.entries.filter(turn => !ordered.some(item => completedRevisions(item).length > 0 &&
    turn.turn_number >= item.from_turn && turn.turn_number <= item.to_turn &&
    new Date(turn.created_at).getTime() <= new Date(item.created_at).getTime()));
  function selectChapter(id: string) {
    const next = ordered.find(item => item.id === id);
    if (!next) return;
    setSelection({ chapter: id, revision: completedRevisions(next).at(-1)?.id ?? '' });
  }
  if (!chapter) return null;
  const index = ordered.findIndex(item => item.id === chapter.id);
  return <section aria-label="Chapter reader" className="min-w-0 space-y-3 border border-gold/20 bg-obsidian p-3 sm:p-5">
    <p role="status" className="text-sm text-pewter">{history.isError ? 'Freshness unavailable. Saved chapters remain readable.' : history.isPending ? 'Checking covered play…' : uncovered?.length ? `${uncovered.length} completed ${uncovered.length === 1 ? 'turn awaits' : 'turns await'} a saved chapter.` : 'All current completed play is covered by saved chapters.'}</p>
    <label className="block text-sm text-pewter">Chapter
      <select className="mt-1 min-h-11 w-full max-w-full bg-midnight p-2 text-champagne" value={chapter.id} onChange={event => selectChapter(event.target.value)}>
        {ordered.map((item, i) => <option key={item.id} value={item.id}>Chapter {i + 1} · Turns {item.from_turn}–{item.to_turn}</option>)}
      </select>
    </label>
    <nav aria-label="Chapter navigation" className="flex flex-wrap justify-between gap-2">
      <button type="button" className="hud-btn hud-text-button disabled:opacity-40" disabled={index === 0} onClick={() => selectChapter(ordered[index - 1].id)}>Previous chapter</button>
      <span className="self-center text-sm text-pewter">{index + 1} / {ordered.length}</span>
      <button type="button" className="hud-btn hud-text-button disabled:opacity-40" disabled={index === ordered.length - 1} onClick={() => selectChapter(ordered[index + 1].id)}>Next chapter</button>
    </nav>
    {revisions.length > 1 && <label className="block text-sm text-pewter">Saved revision
      <select className="mt-1 min-h-11 w-full bg-midnight p-2 text-champagne" value={revision?.id} onChange={event => setSelection({ chapter: chapter.id, revision: event.target.value })}>
        {revisions.map(item => <option key={item.id} value={item.id}>Revision {item.revision}</option>)}
      </select>
    </label>}
    {revision ? <ReadingPane key={`${chapter.id}:${revision.id}`} campaignId={campaignId} chapter={chapter} revision={revision} /> : <p>No completed revision yet. Retry generation below; other chapters remain readable.</p>}
  </section>;
}

function ReadingPane({ campaignId, chapter, revision }: { readonly campaignId: string; readonly chapter: ChapterResponse; readonly revision: ChapterRevisionResponse }) {
  const pane = useRef<HTMLDivElement>(null);
  const sources = useRef<HTMLDetailsElement>(null);
  useLayoutEffect(() => {
    const element = pane.current;
    if (element) element.scrollTop = readOffset(campaignId, chapter.id, revision.id) * Math.max(0, element.scrollHeight - element.clientHeight);
  }, [campaignId, chapter.id, revision.id]);
  return <div ref={pane} role="region" aria-label="Chapter prose" tabIndex={0} className="h-[min(52dvh,32rem)] min-h-40 overflow-y-auto overscroll-contain px-1 focus-visible:outline-gold" onScroll={event => {
    const element = event.currentTarget;
    const range = element.scrollHeight - element.clientHeight;
    if (range > 0) rememberOffset(campaignId, chapter.id, revision.id, element.scrollTop / range);
  }}>
    <article className="mx-auto max-w-prose space-y-5 pb-5 narrative-prose text-base leading-8 text-champagne/90 sm:text-lg">
      <h3 className="font-heading text-gold">Turns {chapter.from_turn} – {chapter.to_turn}</h3>
      <p className="font-body text-xs leading-6 text-pewter">{chapter.rules_version} · Revision {revision.revision} · Player knowledge through turn {chapter.to_turn}</p>
      <a className="inline-flex min-h-11 items-center text-sm text-gold underline" href={`#chapter-sources-${chapter.id}`} onClick={event => {
        event.preventDefault();
        if (sources.current) { sources.current.open = true; sources.current.focus({ preventScroll: true }); sources.current.scrollIntoView({ block: 'nearest' }); }
      }}>Read source history</a>
      {revision.content.split(/\n\s*\n/).map((paragraph, i) => <p key={i} className="whitespace-pre-wrap wrap-break-word">{paragraph}</p>)}
      <details ref={sources} id={`chapter-sources-${chapter.id}`} tabIndex={-1} className="border-t border-gold/20 pt-4 font-body text-sm leading-7">
        <summary className="min-h-11 cursor-pointer py-2 text-gold">Captured source history</summary>
        <p className="my-3 text-pewter">This immutable source records the play used for this chapter. It may predate restored or restarted play.</p>
        {chapter.source_turns.map(turn => <section key={turn.log_id} id={`chapter-source-${chapter.id}-${turn.log_id}`} className="my-4 space-y-2 border-b border-pewter/20 pb-3">
          <h4 className="text-gold">Turn {turn.turn_number}</h4>
          <p className="whitespace-pre-wrap wrap-break-word">You: {turn.player_input || '(opening narration)'}</p>
          <p className="whitespace-pre-wrap wrap-break-word">{turn.narrative}</p>
          <small className="block break-all text-pewter">Source {turn.log_id}{turn.attempt_id ? ` · Action ${turn.attempt_id}` : ''}</small>
        </section>)}
        <p className="break-all text-xs text-pewter">{chapter.knowledge_boundary} · {chapter.source_hash}</p>
      </details>
    </article>
  </div>;
}
