import { useCallback, useMemo, useState } from 'react';

import type { QuestResponse } from '../../api/types';

interface PinnedObjectivesProps {
  readonly quests: QuestResponse[];
}

const STORAGE_KEY = 'gm_pinned_objectives';

function readPinnedIds(): string[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((id): id is string => typeof id === 'string') : [];
  } catch {
    return [];
  }
}

function writePinnedIds(ids: string[]): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
}

interface PinnedObjective {
  questTitle: string;
  objectiveId: string;
  objectiveText: string;
  completed: boolean;
}

export function PinnedObjectives({ quests }: PinnedObjectivesProps) {
  const [pinnedIds, setPinnedIds] = useState<string[]>(readPinnedIds);

  const pinnedObjectives = useMemo<PinnedObjective[]>(() => {
    const result: PinnedObjective[] = [];
    const idSet = new Set(pinnedIds);

    for (const quest of quests) {
      for (const obj of quest.objectives) {
        if (idSet.has(obj.id)) {
          result.push({
            questTitle: quest.title,
            objectiveId: obj.id,
            objectiveText: obj.description,
            completed: obj.completed,
          });
        }
      }
    }

    return result;
  }, [quests, pinnedIds]);

  const togglePin = useCallback((objectiveId: string) => {
    setPinnedIds((prev) => {
      const next = prev.includes(objectiveId) ? prev.filter((id) => id !== objectiveId) : [...prev, objectiveId];
      writePinnedIds(next);
      return next;
    });
  }, []);

  if (pinnedObjectives.length === 0) {
    return null;
  }

  return (
    <div className="sticky top-0 z-10 border-2 border-gold/30 bg-charcoal/95 px-5 py-3 backdrop-blur">
      <p className="mb-2 font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Pinned objectives</p>
      <ul className="space-y-1.5">
        {pinnedObjectives.map((obj) => (
          <li key={obj.objectiveId} className="flex items-center gap-3 text-sm">
            <span className={obj.completed ? 'text-jade' : 'text-gold'}>
              {obj.completed ? '\u2713' : '\u25CB'}
            </span>
            <span className="text-pewter">{obj.questTitle}</span>
            <span className="text-champagne/80">{obj.objectiveText}</span>
            <button
              type="button"
              onClick={() => togglePin(obj.objectiveId)}
              className="ml-auto shrink-0 text-xs text-pewter transition hover:text-gold"
            >
              Unpin
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

/** Small button to pin/unpin an objective from outside this component. */
export function PinObjectiveButton({
  objectiveId,
}: {
  readonly objectiveId: string;
}) {
  const [pinned, setPinned] = useState(() => readPinnedIds().includes(objectiveId));

  function toggle() {
    const ids = readPinnedIds();
    const next = ids.includes(objectiveId) ? ids.filter((id) => id !== objectiveId) : [...ids, objectiveId];
    writePinnedIds(next);
    setPinned(next.includes(objectiveId));
  }

  return (
    <button
      type="button"
      onClick={toggle}
      className={`text-xs transition ${pinned ? 'text-gold' : 'text-pewter hover:text-gold'}`}
      title={pinned ? 'Unpin objective' : 'Pin objective'}
    >
      {pinned ? '\u2605' : '\u2606'}
    </button>
  );
}
