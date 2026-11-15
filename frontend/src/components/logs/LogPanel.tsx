import type { NarrativeEntryItem } from '../narrative/NarrativeEntry';
import { HudPanel } from '../layout/HudPanel';

interface LogPanelProps {
  readonly entries: NarrativeEntryItem[];
  readonly streamingEntry: NarrativeEntryItem | null;
  readonly isLoading: boolean;
  readonly error: string | null;
}

export function LogPanel({ entries, streamingEntry, isLoading, error }: LogPanelProps) {
  const allEntries = streamingEntry ? [...entries, streamingEntry] : entries;

  return (
    <HudPanel title="Logs" accent="scene" bodyClassName="flex flex-col gap-2 font-mono text-sm">
      {error && (
        <p className="text-ruby">{error}</p>
      )}
      {allEntries.length === 0 && !isLoading && (
        <p className="italic text-pewter">No log entries yet.</p>
      )}
      {allEntries.map((entry) => (
        <div key={entry.id} className="flex min-w-0 flex-col gap-1 text-champagne/75 sm:flex-row sm:gap-3">
          <span className="break-all text-pewter/60 sm:shrink-0">{entry.timestamp}</span>
          <span className="text-gold/70">[{entry.kind}]</span>
          <span className="min-w-0 wrap-break-word">{entry.text}</span>
        </div>
      ))}
      {isLoading && (
        <p className="animate-pulse text-pewter">...</p>
      )}
    </HudPanel>
  );
}
