import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ChapterReader } from './ChapterReader';
import { generateChapter, listChapters, type GenerateChapterRequest } from '../../api/journal';

export function ChapterSection({ campaignId }: { readonly campaignId: string }) {
  const client = useQueryClient();
  const [message, setMessage] = useState('');
  const queryKey = ['campaign', campaignId, 'journal', 'chapters'];
  const chapters = useQuery({
    queryKey,
    queryFn: () => listChapters(campaignId),
    refetchInterval: query => query.state.data?.some(c => c.revisions.some(r => r.status === 'generating')) ? 2000 : false,
  });
  const mutation = useMutation({
    mutationFn: (request: GenerateChapterRequest) => generateChapter(campaignId, request),
    onSuccess: (result, request) => {
      // A server response makes the request durable; failed/in-progress
      // revisions expose their key through the list, including after reload.
      const key = requestKey(campaignId, request.chapter_id);
      if (pendingRequest(key) === request.request_id) rememberRequest(key);
      setMessage('message' in result ? 'No completed turns available for a new chapter.' : result.status === 'generating' ? 'Chapter generation is in progress.' : result.status === 'failed' ? 'Generation failed. You can retry; play remains available.' : 'Chapter saved.');
      void client.invalidateQueries({ queryKey: ['campaign', campaignId, 'journal', 'chapters'] });
    },
    onError: () => {
      setMessage('Could not finish the request. Retry uses the same request key; play remains available.');
      void client.invalidateQueries({ queryKey: ['campaign', campaignId, 'journal', 'chapters'] });
    },
  });
  function generate(chapterId?: string, requestId?: string) {
    const key = requestKey(campaignId, chapterId);
    const request: GenerateChapterRequest = { request_id: requestId ?? pendingRequest(key) ?? crypto.randomUUID(), chapter_id: chapterId };
    rememberRequest(key, request.request_id);
    setMessage('');
    mutation.mutate(request);
  }
  return (
    <section className="space-y-3" aria-label="Chapters">
      <p className="text-sm text-pewter">Create a chapter from the next completed turns. Chapters use only the recorded player-visible story.</p>
      <button type="button" className="hud-btn hud-text-button" disabled={mutation.isPending} onClick={() => generate()}>
        {mutation.isPending ? 'Generating chapter...' : 'Generate next chapter'}
      </button>
      {message && <p role="status" className="text-sm text-pewter">{message}</p>}
      {chapters.isPending && <p className="text-sm text-pewter">Loading chapters...</p>}
      {chapters.isError && <p role="alert">Could not load chapters. <button type="button" className="hud-btn hud-text-button" onClick={() => void chapters.refetch()}>Retry loading chapters</button></p>}
      {chapters.data?.length === 0 && <p className="text-sm text-pewter">No chapters yet.</p>}
      {chapters.data && <ChapterReader key={campaignId} campaignId={campaignId} chapters={chapters.data} />}
      {chapters.data?.map(chapter => (
        <article key={chapter.id} className="space-y-3 border-2 border-midnight/40 bg-obsidian p-5">
          <h3 className="font-heading text-sm text-gold">Generation · Turns {chapter.from_turn} – {chapter.to_turn}</h3>
          <p className="text-xs text-pewter">{chapter.rules_version} · Player knowledge through turn {chapter.to_turn}</p>
          {chapter.revisions.map(revision => (
            <div key={revision.id} className="space-y-2">
              <p className="text-xs text-pewter">Revision {revision.revision} · {revision.status}</p>
              {revision.error && <p className="text-sm text-ruby">{revision.error}</p>}
              {revision.status !== 'completed' && <button type="button" className="hud-btn hud-text-button" disabled={mutation.isPending} onClick={() => generate(chapter.id, revision.request_id)}>Retry revision {revision.revision}</button>}
            </div>
          ))}
          <button type="button" className="hud-btn hud-text-button" disabled={mutation.isPending} onClick={() => generate(chapter.id)}>Regenerate chapter</button>
        </article>
      ))}
    </section>
  );
}

// Uncertain delivery must survive panel unmounts and reloads in this tab. Keep
// only request identities, never chapter prose. Blocked storage uses a process
// fallback so navigating between Play, Notes and Chronicle still retains keys.
const storageFallback = new Map<string, string>();
function requestKey(campaignId: string, chapterId?: string): string {
  return `edda:chapter-request:${campaignId}:${chapterId ?? 'next'}`;
}
function pendingRequest(key: string): string | undefined {
  try { return sessionStorage.getItem(key) ?? storageFallback.get(key); }
  catch { return storageFallback.get(key); }
}
function rememberRequest(key: string, id?: string) {
  try { if (id) sessionStorage.setItem(key, id); else sessionStorage.removeItem(key); storageFallback.delete(key); }
  catch { if (id) storageFallback.set(key, id); else storageFallback.delete(key); }
}
