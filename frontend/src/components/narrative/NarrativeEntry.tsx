import type { ResolutionEvent, StateChange } from '../../api/types';
import { cn } from '../../lib/cn';
import { EffectReceipt } from '../character/EffectReceipt';
import { SpellReceipt } from '../character/SpellReceipt';
import { HealingReceipt } from '../character/HealingReceipt';
import { AttackReceipt } from '../character/AttackReceipt';
import { CheckReceipt } from '../character/CheckReceipt';
import type { CheckReceipt as Receipt } from '../../api/types';
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
  readonly resolutionEvents?: ResolutionEvent[];
  readonly choices?: NarrativeChoiceOption[];
  readonly isStreaming?: boolean;
}

interface NarrativeEntryProps {
  readonly entry: NarrativeEntryItem;
  readonly className?: string;
}

const ENTRY_STYLES: Record<NarrativeEntryKind, { shell: string; badge: string; accent: string; defaultSpeaker: string }> = {
  player: {
    shell: 'border-l-2 border-jade/40 pl-4 text-champagne/85',
    badge: 'bg-jade/15 text-jade border border-jade/20',
    accent: 'text-jade',
    defaultSpeaker: 'You',
  },
  gm: {
    shell: 'text-champagne',
    badge: 'bg-gold/15 text-gold border border-gold/20',
    accent: 'text-gold',
    defaultSpeaker: 'Edda',
  },
  system: {
    shell: 'border-l-2 border-pewter/30 pl-4 text-pewter',
    badge: 'bg-pewter/15 text-pewter border border-pewter/20',
    accent: 'text-pewter',
    defaultSpeaker: 'System',
  },
};

export function NarrativeEntry({ entry, className }: NarrativeEntryProps) {
  const styles = ENTRY_STYLES[entry.kind];
  const speaker = entry.speaker?.trim() || styles.defaultSpeaker;
  const displayText = stripDanglingChoicesMarker(entry.text);
  const hasText = displayText.trim().length > 0;
  const showStreamingState = entry.isStreaming && !hasText;
  const timestampLabel = formatTimestamp(entry.timestamp);

  return (
    <article
      className={cn(
        'narrative-entry shrink-0 py-2',
        styles.shell,
        className,
      )}
      data-entry-id={entry.id}
      aria-label={entry.isStreaming ? 'Provisional narration' : `${speaker} entry`}
      aria-live={entry.isStreaming ? 'polite' : undefined}
    >
      <p className={cn('mb-2 text-sm font-semibold', styles.accent, entry.kind === 'gm' ? 'sr-only' : '')}>{speaker}</p>

      {entry.isStreaming ? <p className="mb-2 text-xs text-pewter">Provisional · not saved</p> : null}

      <div className="space-y-4">
        {showStreamingState ? <LoadingIndicator label={`${speaker} is responding`} detail="Streaming the next beat…" /> : null}

        {hasText ? (
          <div className={cn('space-y-5', entry.kind === 'gm' ? 'narrative-prose' : 'text-base leading-7')}>
            {displayText.split(/\n\s*\n/).map((paragraph, index) => <p key={index} className="whitespace-pre-wrap wrap-break-word">{paragraph}</p>)}
            {entry.isStreaming ? <span className="inline-block h-4 w-2 animate-pulse bg-gold/80 align-middle" /> : null}
          </div>
        ) : null}

        {!entry.isStreaming ? <details className="text-sm text-pewter">
          <summary className="inline-flex min-h-11 cursor-pointer items-center py-2 focus:outline-none focus-visible:ring-2 focus-visible:ring-gold">Turn details</summary>
          <time dateTime={entry.timestamp} className="block py-2">{timestampLabel}</time>
        {entry.stateChanges && entry.stateChanges.length > 0 ? (
          <div className="flex flex-wrap gap-2 border-t border-white/10 pt-3">
            {entry.stateChanges.map((change, index) => (
              <span
                key={`${change.entity_type}-${change.entity_id}-${change.change_type}-${index}`}
                className="hud-baseline-badge rounded-sm border border-pewter/15 bg-pewter/5 px-3 text-[11px] font-medium uppercase tracking-[0.18em] text-champagne/70"
              >
                {change.change_type}
              </span>
            ))}
          </div>
        ) : null}

        {entry.resolutionEvents && entry.resolutionEvents.length > 0 ? (
          <div className="space-y-2 border-t border-white/10 pt-3">
            {entry.resolutionEvents.map((event, index) => (
              <ResolutionEventBadge key={`${event.type}-${event.label}-${index}`} event={event} />
            ))}
          </div>
        ) : null}
        </details> : null}
      </div>
    </article>
  );
}

function ResolutionEventBadge({ event }: { readonly event: ResolutionEvent }) {
  if (event.type === 'srd_effect' && event.details.receipt) return <EffectReceipt receipt={event.details.receipt as import('../../api/types').EffectReceipt} />;
  if (event.type === 'srd_spell' && event.details.receipt) return <SpellReceipt receipt={event.details.receipt as import('../../api/types').SpellReceipt} />;
  if (event.type === 'srd_consumable' && event.details.receipt) return <HealingReceipt receipt={event.details.receipt as import('../../api/types').HealingReceipt} />;
  if (event.type === 'srd_attack' && event.details.receipt) return <AttackReceipt receipt={event.details.receipt as import('../../api/types').AttackReceipt} />;
  if (event.type === 'srd_check' && event.details.receipt) return <CheckReceipt receipt={event.details.receipt as Receipt} />;
  if (event.type === 'skill_check') {
    return <SkillCheckBadge event={event} />;
  }

  return null;
}

function SkillCheckBadge({ event }: { readonly event: ResolutionEvent }) {
  const success = event.outcome === 'success';
  const skill = stringDetail(event.details.skill) ?? event.label.replace(/ check$/i, '');
  const total = numberDetail(event.details.total);
  const dc = numberDetail(event.details.dc) ?? numberDetail(event.details.difficulty);
  const roll = numberDetail(event.details.roll);
  const modifier = numberDetail(event.details.modifier);
  const margin = numberDetail(event.details.margin);

  return (
    <div className={cn(
      'flex flex-col gap-1 rounded-sm border px-3 py-2 text-xs sm:flex-row sm:items-center sm:justify-between',
      success ? 'border-jade/25 bg-jade/10 text-jade' : 'border-ruby/25 bg-ruby/10 text-ruby',
    )}>
      <span className="font-semibold uppercase tracking-[0.16em]">
        {skill} check: {success ? 'success' : 'failure'}
      </span>
      <span className="font-mono text-[11px] uppercase tracking-[0.12em] text-champagne/75">
        {formatCheckMath({ roll, modifier, total, dc, margin })}
      </span>
    </div>
  );
}

function formatCheckMath(values: { roll?: number; modifier?: number; total?: number; dc?: number; margin?: number }): string {
  const parts: string[] = [];
  if (values.roll !== undefined) {
    const modifier = values.modifier ?? 0;
    const sign = modifier >= 0 ? '+' : '';
    parts.push(`roll ${values.roll}${sign}${modifier}`);
  }
  if (values.total !== undefined) {
    parts.push(`total ${values.total}`);
  }
  if (values.dc !== undefined) {
    parts.push(`DC ${values.dc}`);
  }
  if (values.margin !== undefined) {
    const sign = values.margin >= 0 ? '+' : '';
    parts.push(`${sign}${values.margin}`);
  }
  return parts.join(' · ');
}

function stringDetail(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim().length > 0 ? value : undefined;
}

function numberDetail(value: unknown): number | undefined {
  return typeof value === 'number' && Number.isFinite(value) ? value : undefined;
}

function stripDanglingChoicesMarker(text: string): string {
  const lines = text.split('\n');
  while (lines.length > 0) {
    const trimmed = lines[lines.length - 1].trim().toLowerCase().replace(/[’]/g, "'");
    if (trimmed === '' || trimmed === '**choices:**' || trimmed === 'choices:' || trimmed === 'options:') {
      lines.pop();
      continue;
    }
    break;
  }
  return lines.join('\n').trim();
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
