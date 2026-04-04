import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getCampaignQuest, listCampaignQuests } from '../../api/quests';
import type { QuestResponse } from '../../api/types';
import { cn } from '../../lib/cn';
import { QuestCard } from './QuestCard';

interface QuestPanelProps {
  readonly campaignId: string;
  readonly className?: string;
}

interface FilterOption {
  readonly value: string;
  readonly label: string;
}

const questTypeOptions: readonly FilterOption[] = [
  { value: '', label: 'All types' },
  { value: 'short_term', label: 'Short term' },
  { value: 'medium_term', label: 'Medium term' },
  { value: 'long_term', label: 'Long term' },
] as const;

const questStatusOptions: readonly FilterOption[] = [
  { value: '', label: 'All statuses' },
  { value: 'active', label: 'Active' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'abandoned', label: 'Abandoned' },
] as const;

function queryErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to load campaign quests.';
}

async function loadCampaignQuests(campaignId: string): Promise<QuestResponse[]> {
  const quests = await listCampaignQuests(campaignId);

  if (quests.length === 0) {
    return [];
  }

  return Promise.all(quests.map((quest) => getCampaignQuest(campaignId, quest.id)));
}

export function QuestPanel({ campaignId, className }: QuestPanelProps) {
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const questsQuery = useQuery({
    queryKey: ['campaign', campaignId, 'quests'],
    queryFn: () => loadCampaignQuests(campaignId),
    enabled: campaignId.trim().length > 0,
  });

  const filteredQuests = useMemo(() => {
    const quests = questsQuery.data ?? [];

    return quests.filter((quest) => {
      if (typeFilter && quest.quest_type !== typeFilter) {
        return false;
      }

      if (statusFilter && quest.status !== statusFilter) {
        return false;
      }

      return true;
    });
  }, [questsQuery.data, statusFilter, typeFilter]);

  if (campaignId.trim().length === 0) {
    return <QuestPanelMessage className={className} tone="error" title="Missing campaign" message="Select a campaign before opening the quest board." />;
  }

  if (questsQuery.isPending) {
    return <QuestPanelMessage className={className} tone="default" title="Loading quests" message="Gathering active, completed, and archived quest threads for this campaign." />;
  }

  if (questsQuery.isError) {
    return <QuestPanelMessage className={className} tone="error" title="Quest board unavailable" message={queryErrorMessage(questsQuery.error)} />;
  }

  const quests = questsQuery.data;
  const hasFilters = typeFilter !== '' || statusFilter !== '';

  if (quests.length === 0) {
    return <QuestPanelMessage className={className} tone="default" title="No quests yet" message="Quest hooks, long-term arcs, and subquests will appear here once the campaign creates them." />;
  }

  return (
    <section className={cn('space-y-5 border-2 border-sapphire/20 bg-charcoal p-5', className)}>
      <div className="flex flex-col gap-4 border-b-2 border-sapphire/20 pb-5 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-sapphire">Quests</p>
          <h2 className="font-heading mt-2 text-xl font-semibold uppercase tracking-[0.1em] text-champagne">Campaign quest board</h2>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-pewter">
            Filter quest threads by type or status. Objective progress comes from each quest record, not placeholder frontend state.
          </p>
        </div>

        <div className="grid gap-3 sm:grid-cols-2">
          <FilterField label="Type" value={typeFilter} options={questTypeOptions} onChange={setTypeFilter} />
          <FilterField label="Status" value={statusFilter} options={questStatusOptions} onChange={setStatusFilter} />
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border border-sapphire/15 bg-charcoal/80 px-4 py-3 text-sm text-champagne/70">
        <p>
          Showing <span className="font-semibold text-champagne">{filteredQuests.length}</span> of{' '}
          <span className="font-semibold text-champagne">{quests.length}</span> quests.
        </p>
        {hasFilters ? (
          <button
            type="button"
            onClick={() => {
              setTypeFilter('');
              setStatusFilter('');
            }}
            className="inline-flex items-center justify-center border border-sapphire/20 px-3 py-2 text-xs font-semibold uppercase tracking-[0.2em] text-champagne/80 transition hover:border-sapphire hover:text-sapphire focus:outline-none focus:ring-2 focus:ring-sapphire focus:ring-offset-2 focus:ring-offset-obsidian"
          >
            Clear filters
          </button>
        ) : null}
      </div>

      {filteredQuests.length === 0 ? (
        <QuestPanelMessage
          tone="default"
          title="No matching quests"
          message="No quest currently matches the selected filters. Clear a filter to see the full campaign board."
        />
      ) : (
        <div className="space-y-4">
          {filteredQuests.map((quest) => (
            <QuestCard key={quest.id} quest={quest} />
          ))}
        </div>
      )}
    </section>
  );
}

interface FilterFieldProps {
  readonly label: string;
  readonly value: string;
  readonly options: readonly FilterOption[];
  readonly onChange: (nextValue: string) => void;
}

function FilterField({ label, value, options, onChange }: FilterFieldProps) {
  return (
    <label className="space-y-2 text-sm text-champagne/70">
      <span className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{label}</span>
      <select
        value={value}
        onChange={(event) => {
          onChange(event.target.value);
        }}
        className="w-full border border-sapphire/20 bg-obsidian px-3 py-2.5 text-sm text-champagne outline-none transition focus:border-sapphire"
      >
        {options.map((option) => (
          <option key={`${label}-${option.value || 'all'}`} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

interface QuestPanelMessageProps {
  readonly title: string;
  readonly message: string;
  readonly tone: 'default' | 'error';
  readonly className?: string;
}

function QuestPanelMessage({ title, message, tone, className }: QuestPanelMessageProps) {
  return (
    <section
      className={cn(
        'border p-6',
        tone === 'error'
          ? 'border-ruby/40 bg-ruby/10 text-ruby'
          : 'border-sapphire/20 bg-charcoal text-champagne/70',
        className,
      )}
    >
      <h2 className="font-heading text-lg font-semibold uppercase tracking-[0.1em] text-champagne">{title}</h2>
      <p className="mt-3 text-sm leading-6">{message}</p>
    </section>
  );
}
