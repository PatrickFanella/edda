import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { listChapters } from '../../api/journal';
import type { ChapterResponse } from '../../api/types';
import { readSelection } from '../journal/readingPosition';

interface Props {
  readonly campaignId: string;
  readonly disabled: boolean;
  readonly onDownload: (revisionIds: string[], format: 'markdown' | 'json') => void;
}

export function ChapterExportOptions(props: Props) {
  const [open, setOpen] = useState(false);
  return <details className="border border-gold/20 p-3" onToggle={event => setOpen(event.currentTarget.open)}>
    <summary className="min-h-11 cursor-pointer py-2 text-sm font-semibold uppercase tracking-wide">Saved Chapters</summary>
    {open && <ChapterSelection {...props} />}
  </details>;
}

function ChapterSelection({ campaignId, disabled, onDownload }: Props) {
  const query = useQuery({ queryKey: ['campaign', campaignId, 'journal', 'chapters'], queryFn: () => listChapters(campaignId), retry: false });
  if (query.isPending) return <p role="status">Loading saved chapters…</p>;
  if (query.isError) return <p role="alert">Could not load saved chapters. <button type="button" className="min-h-11 border border-gold/30 px-3 py-2 text-sm" onClick={() => void query.refetch()}>Retry loading chapters</button></p>;
  return <SavedChapterSelection campaignId={campaignId} chapters={query.data} disabled={disabled} onDownload={onDownload} />;
}

function SavedChapterSelection({ campaignId, chapters, disabled, onDownload }: Props & { readonly chapters: ChapterResponse[] }) {
  const [reader] = useState(() => readSelection(campaignId));
  const [selected, setSelected] = useState<Record<string, string>>({});
  const [format, setFormat] = useState<'markdown' | 'json'>('markdown');
  const ordered = [...chapters].sort((a, b) => a.created_at.localeCompare(b.created_at) || a.from_turn - b.from_turn || a.to_turn - b.to_turn || a.id.localeCompare(b.id));
  const available = ordered.filter(chapter => chapter.revisions.some(revision => revision.status === 'completed'));
  const selectedIds = available.flatMap(chapter => chapter.revisions.filter(revision => revision.status === 'completed' && revision.id === selected[chapter.id]).map(revision => revision.id));
  if (available.length === 0) return <p className="text-sm text-pewter">No completed saved chapters. Create a chapter in Journal first.</p>;
  return <div className="space-y-3">
    <p className="text-sm text-pewter">Select saved chapters and revisions. Includes captured source history, coverage, rules provenance and attribution. Downloads the saved prose without generating it again.</p>
    <fieldset disabled={disabled} className="space-y-3">
      <legend className="text-sm text-gold">Chapters to export</legend>
      {available.map(chapter => {
        const revisions = chapter.revisions.filter(revision => revision.status === 'completed').sort((a, b) => a.revision - b.revision);
        const preferred = revisions.find(revision => chapter.id === reader?.chapter && revision.id === reader.revision) ?? revisions.at(-1)!;
        const value = selected[chapter.id] ?? preferred.id;
        const label = `Chapter ${ordered.indexOf(chapter) + 1} · Turns ${chapter.from_turn}–${chapter.to_turn}`;
        return <div key={chapter.id} className="space-y-1">
          <label className="flex min-h-11 items-center gap-3 text-sm">
            <input type="checkbox" checked={chapter.id in selected} onChange={event => setSelected(previous => {
              const next = { ...previous };
              if (event.target.checked) next[chapter.id] = value; else delete next[chapter.id];
              return next;
            })} />{label}
          </label>
          <label className="block text-xs text-pewter">Revision for {label}
            <select className="mt-1 min-h-11 w-full bg-midnight p-2 text-sm text-champagne" disabled={!(chapter.id in selected)} value={value} onChange={event => setSelected(previous => ({ ...previous, [chapter.id]: event.target.value }))}>
              {revisions.map(revision => <option key={revision.id} value={revision.id}>Revision {revision.revision}</option>)}
            </select>
          </label>
        </div>;
      })}
      <label className="block text-sm text-pewter">Chapter file format
        <select className="mt-1 min-h-11 w-full bg-midnight p-2 text-champagne" value={format} onChange={event => setFormat(event.target.value as 'markdown' | 'json')}>
          <option value="markdown">Markdown</option><option value="json">JSON</option>
        </select>
      </label>
      <button type="button" className="min-h-11 border border-gold/30 px-3 py-2 text-sm hover:border-gold disabled:opacity-40" disabled={selectedIds.length === 0 || selectedIds.length > 100} onClick={() => onDownload(selectedIds, format)}>Download selected chapters</button>
      {selectedIds.length > 100 && <p role="alert">Select at most 100 chapters per download.</p>}
    </fieldset>
  </div>;
}
