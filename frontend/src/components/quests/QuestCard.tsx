import type { QuestResponse } from '../../api/types';
import { cn } from '../../lib/cn';
import { ObjectiveList } from './ObjectiveList';

interface QuestCardProps {
  readonly quest: QuestResponse;
  readonly className?: string;
}

const questTypeLabels: Record<string, string> = {
  short_term: 'Short term',
  medium_term: 'Medium term',
  long_term: 'Long term',
};

const questTypeTones: Record<string, string> = {
  short_term: 'border-jade/30 bg-jade/10 text-jade',
  medium_term: 'border-gold/30 bg-gold/10 text-gold',
  long_term: 'border-ruby/30 bg-ruby/10 text-ruby',
};

const questStatusLabels: Record<string, string> = {
  active: 'Active',
  completed: 'Completed',
  failed: 'Failed',
  abandoned: 'Abandoned',
};

const questStatusTones: Record<string, string> = {
  active: 'border-jade/30 bg-jade/10 text-jade',
  completed: 'border-gold/30 bg-gold/10 text-gold',
  failed: 'border-ruby/30 bg-ruby/10 text-ruby',
  abandoned: 'border-pewter/30 bg-pewter/10 text-pewter',
};

function formatLabel(value: string, labels: Record<string, string>): string {
  return labels[value] ?? value.replaceAll('_', ' ');
}

function progressLabel(quest: QuestResponse): string {
  if (quest.objectives.length === 0) {
    return 'No objectives';
  }

  const completedCount = quest.objectives.filter((objective) => objective.completed).length;
  return `${completedCount}/${quest.objectives.length} objectives complete`;
}

export function QuestCard({ quest, className }: QuestCardProps) {
  return (
    <article className={cn('deco-corners deco-corners-sapphire deco-pattern border-2 border-sapphire/25 bg-charcoal p-5 transition-all duration-200 hover:border-sapphire/45 hover:-translate-y-0.5', className)}>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="font-heading text-lg font-semibold uppercase tracking-[0.08em] text-champagne">{quest.title}</h3>
            {quest.parent_quest_id ? (
              <span className="rounded-sm border border-pewter/30 bg-pewter/10 px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter">
                Subquest
              </span>
            ) : null}
          </div>
          <p className="max-w-3xl text-sm leading-6 text-champagne/70">{quest.description || 'No quest briefing added yet.'}</p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <span
            className={cn(
              'rounded-sm border px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em]',
              questTypeTones[quest.quest_type] ?? 'border-pewter/30 bg-pewter/10 text-pewter',
            )}
          >
            {formatLabel(quest.quest_type, questTypeLabels)}
          </span>
          <span
            className={cn(
              'rounded-sm border px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em]',
              questStatusTones[quest.status] ?? 'border-pewter/30 bg-pewter/10 text-pewter',
            )}
          >
            {formatLabel(quest.status, questStatusLabels)}
          </span>
        </div>
      </div>

      <dl className="mt-5 grid gap-4 border-t-2 border-sapphire/20 pt-5 text-sm text-champagne/70 sm:grid-cols-3">
        <MetadataItem label="Progress" value={progressLabel(quest)} />
        <MetadataItem label="Objectives" value={String(quest.objectives.length)} />
        <MetadataItem label="Parent quest" value={quest.parent_quest_id ? 'Linked' : 'Standalone'} />
      </dl>

      <div className="mt-5 space-y-3">
        <div>
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-sapphire">Objectives</p>
          <p className="mt-1 text-sm text-pewter">Track completion state for the current quest chain.</p>
        </div>
        <ObjectiveList objectives={quest.objectives} />
      </div>
    </article>
  );
}

function MetadataItem({ label, value }: { readonly label: string; readonly value: string }) {
  return (
    <div className="space-y-1">
      <dt className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{label}</dt>
      <dd className="leading-6 text-champagne/80">{value}</dd>
    </div>
  );
}
