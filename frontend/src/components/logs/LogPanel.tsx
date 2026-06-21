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
        <div key={entry.id} className="flex gap-3 text-champagne/75">
          <span className="shrink-0 text-pewter/60">{entry.timestamp}</span>
          <span className="text-gold/70">[{entry.kind}]</span>
          <span>{entry.text}</span>
        </div>
      ))}
      {isLoading && (
        <p className="animate-pulse text-pewter">...</p>
      )}
    </HudPanel>
  );
}
