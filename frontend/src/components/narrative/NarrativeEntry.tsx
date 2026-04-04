import type { StateChange } from '../../api/types';
import { cn } from '../../lib/cn';
import { LoadingIndicator } from './LoadingIndicator';

export type NarrativeEntryKind = 'player' | 'gm' | 'system';

export interface NarrativeChoiceOption {
  readonly id: string;
  readonly text: string;
}

export interface NarrativeEntryItem {
  readonly id: string;
  readonly kind: NarrativeEntryKind;
  readonly text: string;
  readonly timestamp: string;
  readonly speaker?: string;
  readonly stateChanges?: StateChange[];
  readonly choices?: NarrativeChoiceOption[];
  readonly isStreaming?: boolean;
}

interface NarrativeEntryProps {
  readonly entry: NarrativeEntryItem;
  readonly className?: string;
}

const ENTRY_STYLES: Record<NarrativeEntryKind, { shell: string; badge: string; accent: string; defaultSpeaker: string }> = {
  player: {
    shell: 'border-jade/30 bg-jade/5 text-champagne hover:border-jade/50',
    badge: 'bg-jade/15 text-jade border border-jade/20',
    accent: 'text-jade',
    defaultSpeaker: 'You',
  },
  gm: {
    shell: 'border-gold/30 bg-gold/5 text-champagne hover:border-gold/50',
    badge: 'bg-gold/15 text-gold border border-gold/20',
    accent: 'text-gold',
    defaultSpeaker: 'Game Master',
  },
  system: {
    shell: 'border-pewter/30 bg-pewter/5 text-champagne/90 hover:border-pewter/50',
    badge: 'bg-pewter/15 text-pewter border border-pewter/20',
    accent: 'text-pewter',
    defaultSpeaker: 'System',
  },
};

export function NarrativeEntry({ entry, className }: NarrativeEntryProps) {
  const styles = ENTRY_STYLES[entry.kind];
  const speaker = entry.speaker?.trim() || styles.defaultSpeaker;
  const hasText = entry.text.trim().length > 0;
  const showStreamingState = entry.isStreaming && !hasText;
  const timestampLabel = formatTimestamp(entry.timestamp);

  return (
    <article
      className={cn(
        'border-2 px-4 py-4 transition-all duration-200 sm:px-5',
        styles.shell,
        className,
      )}
      aria-live={entry.isStreaming ? 'polite' : undefined}
    >
      <header className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="flex items-center gap-3">
          <span className={cn('inline-flex rounded-sm px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.2em]', styles.badge)}>
            {entry.kind}
          </span>
          <p className={cn('text-sm font-semibold', styles.accent)}>{speaker}</p>
        </div>
        <p className="text-xs uppercase tracking-[0.2em] text-pewter">{timestampLabel}</p>
      </header>

      <div className="mt-4 space-y-4">
        {showStreamingState ? <LoadingIndicator label={`${speaker} is responding`} detail="Streaming the next beat…" /> : null}

        {hasText ? (
          <p className="whitespace-pre-wrap break-words text-sm leading-7 text-inherit">
            {entry.text}
            {entry.isStreaming ? <span className="ml-2 inline-block h-4 w-2 animate-pulse bg-gold/80 align-middle" /> : null}
          </p>
        ) : null}

        {entry.stateChanges && entry.stateChanges.length > 0 ? (
          <div className="flex flex-wrap gap-2 border-t border-white/10 pt-3">
            {entry.stateChanges.map((change, index) => (
              <span
                key={`${change.entity_type}-${change.entity_id}-${change.change_type}-${index}`}
                className="rounded-sm border border-pewter/15 bg-pewter/5 px-3 py-1 text-[11px] font-medium uppercase tracking-[0.18em] text-champagne/70"
              >
                {change.change_type}
              </span>
            ))}
          </div>
        ) : null}
      </div>
    </article>
  );
}

function formatTimestamp(timestamp: string): string {
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime())) {
    return timestamp;
  }

  return parsed.toLocaleTimeString([], {
    hour: 'numeric',
    minute: '2-digit',
  });
}
