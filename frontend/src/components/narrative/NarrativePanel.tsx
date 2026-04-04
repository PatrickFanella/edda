import { useEffect, useMemo, useRef, type ReactNode } from 'react';

import { cn } from '../../lib/cn';
import { NarrativeEntry, type NarrativeEntryItem, type NarrativeEntryKind } from './NarrativeEntry';

interface StreamingNarrativeEntry {
  readonly text?: string;
  readonly timestamp?: string;
  readonly speaker?: string;
  readonly kind?: NarrativeEntryKind;
}

interface NarrativePanelProps {
  readonly entries: NarrativeEntryItem[];
  readonly streamingEntry?: StreamingNarrativeEntry | null;
  readonly isLoading?: boolean;
  readonly className?: string;
  readonly emptyState?: ReactNode;
}

export function NarrativePanel({
  entries,
  streamingEntry = null,
  isLoading = false,
  className,
  emptyState,
}: NarrativePanelProps) {
  const endRef = useRef<HTMLDivElement | null>(null);

  const activeStreamingEntry = useMemo<NarrativeEntryItem | null>(() => {
    if (streamingEntry) {
      return {
        id: 'streaming-entry',
        kind: streamingEntry.kind ?? 'gm',
        text: streamingEntry.text ?? '',
        timestamp: streamingEntry.timestamp ?? new Date().toISOString(),
        speaker: streamingEntry.speaker,
        isStreaming: true,
      };
    }

    if (!isLoading) {
      return null;
    }

    return {
      id: 'streaming-entry',
      kind: 'gm',
      text: '',
      timestamp: new Date().toISOString(),
      speaker: 'Game Master',
      isStreaming: true,
    };
  }, [isLoading, streamingEntry]);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [entries, activeStreamingEntry]);

  const hasEntries = entries.length > 0 || activeStreamingEntry !== null;

  return (
    <section
      className={cn(
        'deco-corners deco-pattern border-2 border-gold/20 bg-charcoal',
        className,
      )}
    >
      <div className="flex items-center justify-between border-b-2 border-gold/20 px-5 py-4">
        <div>
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Narrative</p>
          <p className="mt-1 text-sm text-pewter">Live scene log with player actions and GM responses.</p>
        </div>
      </div>

      <div
        role="log"
        aria-live="polite"
        aria-busy={activeStreamingEntry ? 'true' : 'false'}
        className="flex max-h-[34rem] min-h-80 flex-col gap-4 overflow-y-auto px-4 py-4 sm:px-5"
      >
        {hasEntries ? (
          <>
            {entries.map((entry) => (
              <NarrativeEntry key={entry.id} entry={entry} />
            ))}
            {activeStreamingEntry ? <NarrativeEntry entry={activeStreamingEntry} /> : null}
          </>
        ) : (
          emptyState ?? (
            <div className="flex min-h-64 flex-1 flex-col items-center justify-center border border-dashed border-gold/15 bg-charcoal/50 px-6 text-center">
              <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-pewter/80">Awaiting first move</p>
              <p className="mt-3 max-w-md text-sm leading-7 text-pewter">
                Send an action to start the scene. New beats, system notices, and suggested choices will collect here.
              </p>
            </div>
          )
        )}
        <div ref={endRef} aria-hidden="true" />
      </div>
    </section>
  );
}
