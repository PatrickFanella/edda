import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { listQuestHistory } from '../../api/quests';
import type { QuestHistoryEntry } from '../../api/types';

interface QuestHistoryProps {
  readonly campaignId: string;
  readonly questId: string;
}

function relativeTime(iso: string): string {
  try {
    const diff = Date.now() - new Date(iso).getTime();
    const seconds = Math.floor(diff / 1000);
    if (seconds < 60) return 'just now';
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    const days = Math.floor(hours / 24);
    return `${days}d ago`;
  } catch {
    return iso;
  }
}

function parseSnapshot(snapshot: string): Record<string, unknown> | null {
  try {
    return JSON.parse(snapshot) as Record<string, unknown>;
  } catch {
    return null;
  }
}

function snapshotSummary(snapshot: string): string {
  const parsed = parseSnapshot(snapshot);
  if (!parsed) return snapshot.slice(0, 80);

  const parts: string[] = [];
  if (typeof parsed.status === 'string') parts.push(`Status: ${parsed.status}`);
  if (typeof parsed.title === 'string') parts.push(parsed.title);
  if (Array.isArray(parsed.objectives)) parts.push(`${parsed.objectives.length} objectives`);
  return parts.length > 0 ? parts.join(' · ') : snapshot.slice(0, 80);
}

export function QuestHistory({ campaignId, questId }: QuestHistoryProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  const historyQuery = useQuery({
    queryKey: ['campaign', campaignId, 'quest-history', questId],
    queryFn: () => listQuestHistory(campaignId, questId),
  });

  const entries: QuestHistoryEntry[] = historyQuery.data ?? [];

  return (
    <div className="mt-4 space-y-3 border-t-2 border-sapphire/15 pt-4">
      <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-sapphire">History</p>

      {historyQuery.isPending ? (
        <p className="text-sm text-pewter">Loading history...</p>
      ) : historyQuery.isError ? (
        <p className="text-sm text-ruby">Failed to load history.</p>
      ) : entries.length === 0 ? (
        <p className="text-sm text-pewter">No history entries yet.</p>
      ) : (
        <ul className="space-y-2">
          {entries.map((entry) => {
            const isExpanded = expandedId === entry.id;
            const parsed = parseSnapshot(entry.snapshot);

            return (
              <li key={entry.id} className="border border-sapphire/15 bg-obsidian/50">
                <button
                  type="button"
                  onClick={() => setExpandedId(isExpanded ? null : entry.id)}
                  className="flex w-full items-center justify-between px-4 py-3 text-left transition hover:bg-sapphire/5"
                >
                  <span className="text-sm text-champagne/80">{snapshotSummary(entry.snapshot)}</span>
                  <span className="ml-3 shrink-0 text-xs text-pewter">{relativeTime(entry.created_at)}</span>
                </button>

                {isExpanded ? (
                  <div className="border-t border-sapphire/10 px-4 py-3">
                    <pre className="overflow-x-auto text-xs leading-5 text-champagne/60">
                      {parsed ? JSON.stringify(parsed, null, 2) : entry.snapshot}
                    </pre>
                  </div>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
