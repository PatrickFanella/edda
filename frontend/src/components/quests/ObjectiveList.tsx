import { useMemo } from 'react';

import type { QuestObjectiveResponse } from '../../api/types';
import { cn } from '../../lib/cn';

interface ObjectiveListProps {
  readonly objectives: readonly QuestObjectiveResponse[];
  readonly className?: string;
}

export function ObjectiveList({ objectives, className }: ObjectiveListProps) {
  const sortedObjectives = useMemo(
    () => [...objectives].sort((left, right) => left.order_index - right.order_index),
    [objectives],
  );

  if (sortedObjectives.length === 0) {
    return (
      <div className={cn('border border-dashed border-sapphire/15 bg-charcoal/50 px-4 py-5 text-sm text-pewter', className)}>
        No objectives yet.
      </div>
    );
  }

  return (
    <ol className={cn('space-y-3', className)}>
      {sortedObjectives.map((objective, index) => {
        const completed = objective.completed;

        return (
          <li
            key={objective.id}
            className={cn(
              'flex items-start gap-3 border-2 px-4 py-3 text-sm transition-all duration-200',
              completed
                ? 'border-jade/30 bg-jade/5 text-jade hover:border-jade/50'
                : 'border-gold/15 bg-charcoal text-champagne/80 hover:border-gold/30',
            )}
          >
            <span
              aria-hidden="true"
              className={cn(
                'mt-0.5 inline-flex size-6 shrink-0 items-center justify-center border text-xs font-semibold',
                completed
                  ? 'border-jade/50 bg-jade/20 text-jade'
                  : 'border-gold/20 bg-charcoal text-pewter',
              )}
            >
              {completed ? '✓' : index + 1}
            </span>
            <div className="min-w-0 flex-1">
              <p className={cn('leading-6', completed ? 'line-through decoration-jade/70' : '')}>{objective.description}</p>
              <p className="mt-1 text-xs font-medium uppercase tracking-[0.2em] text-pewter">{completed ? 'Completed' : 'In progress'}</p>
            </div>
          </li>
        );
      })}
    </ol>
  );
}
