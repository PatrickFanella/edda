import type { ReactNode } from 'react';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { listKnownLanguages, listKnownCultures, listKnownBeliefSystems, listKnownEconomicSystems } from '../../api/codex';
import type { LanguageResponse, CultureResponse, BeliefSystemResponse, EconomicSystemResponse } from '../../api/types';
import { cn } from '../../lib/cn';
import { HudPanel } from '../layout/HudPanel';

interface CodexPanelProps {
  readonly campaignId: string;
  readonly className?: string;
}

type CodexSection = 'languages' | 'cultures' | 'beliefs' | 'economies';

const codexSections: readonly { key: CodexSection; label: string }[] = [
  { key: 'languages', label: 'Languages' },
  { key: 'cultures', label: 'Cultures' },
  { key: 'beliefs', label: 'Beliefs' },
  { key: 'economies', label: 'Economies' },
] as const;

function queryErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to load codex data.';
}

export function CodexPanel({ campaignId, className }: CodexPanelProps) {
  const [activeSection, setActiveSection] = useState<CodexSection>('languages');
  const enabled = campaignId.trim().length > 0;

  const languagesQuery = useQuery({
    queryKey: ['campaign', campaignId, 'codex-languages'],
    queryFn: () => listKnownLanguages(campaignId),
    enabled,
  });

  const culturesQuery = useQuery({
    queryKey: ['campaign', campaignId, 'codex-cultures'],
    queryFn: () => listKnownCultures(campaignId),
    enabled,
  });

  const beliefsQuery = useQuery({
    queryKey: ['campaign', campaignId, 'codex-beliefs'],
    queryFn: () => listKnownBeliefSystems(campaignId),
    enabled,
  });

  const economiesQuery = useQuery({
    queryKey: ['campaign', campaignId, 'codex-economies'],
    queryFn: () => listKnownEconomicSystems(campaignId),
    enabled,
  });

  if (!enabled) {
    return <PanelMessage className={className} title="Missing campaign" message="Select a campaign before viewing the codex." />;
  }

  return (
    <HudPanel title="Codex" accent="scene" className={cn(className)} bodyClassName="space-y-5">
      <div className="flex flex-wrap gap-2">
        {codexSections.map((section) => (
          <button
            key={section.key}
            type="button"
            onClick={() => setActiveSection(section.key)}
            className={cn(
              'hud-tab-button px-4 text-sm font-semibold uppercase tracking-[0.15em] transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-jade focus:ring-offset-2 focus:ring-offset-obsidian',
              activeSection === section.key
                ? 'bg-jade text-obsidian'
                : 'border border-jade/20 bg-charcoal text-champagne/70 hover:border-jade hover:text-jade hover:bg-jade/5',
            )}
          >
            {section.label}
          </button>
        ))}
      </div>

      {activeSection === 'languages' && (
        <CodexListSection<LanguageResponse>
          query={languagesQuery}
          emptyMessage="No languages discovered yet."
          renderItem={(item) => (
            <CodexEntry key={item.id} name={item.name} description={item.description} />
          )}
        />
      )}

      {activeSection === 'cultures' && (
        <CodexListSection<CultureResponse>
          query={culturesQuery}
          emptyMessage="No cultures discovered yet."
          renderItem={(item) => (
            <CodexEntry key={item.id} name={item.name} />
          )}
        />
      )}

      {activeSection === 'beliefs' && (
        <CodexListSection<BeliefSystemResponse>
          query={beliefsQuery}
          emptyMessage="No belief systems discovered yet."
          renderItem={(item) => (
            <CodexEntry key={item.id} name={item.name} />
          )}
        />
      )}

      {activeSection === 'economies' && (
        <CodexListSection<EconomicSystemResponse>
          query={economiesQuery}
          emptyMessage="No economic systems discovered yet."
          renderItem={(item) => (
            <CodexEntry key={item.id} name={item.name} />
          )}
        />
      )}
    </HudPanel>
  );
}

interface CodexListSectionProps<T> {
  readonly query: { isPending: boolean; isError: boolean; error: unknown; data: T[] | undefined };
  readonly emptyMessage: string;
  readonly renderItem: (item: T) => ReactNode;
}

function CodexListSection<T>({ query, emptyMessage, renderItem }: CodexListSectionProps<T>) {
  if (query.isPending) {
    return <HudPanel title="Loading" accent="loading">Loading...</HudPanel>;
  }

  if (query.isError) {
    return <HudPanel title="Unavailable" accent="error">{queryErrorMessage(query.error)}</HudPanel>;
  }

  const items = query.data ?? [];

  if (items.length === 0) {
    return <HudPanel title="No records" accent="empty">{emptyMessage}</HudPanel>;
  }

  return <div className="space-y-3">{items.map(renderItem)}</div>;
}

function CodexEntry({ name, description }: { readonly name: string; readonly description?: string }) {
  return (
    <div className="border border-jade/15 bg-obsidian px-5 py-4 transition-all duration-200 hover:border-jade/30">
      <h4 className="font-heading text-sm font-semibold uppercase tracking-wide text-champagne">{name}</h4>
      {description && <p className="mt-2 text-sm leading-6 text-pewter">{description}</p>}
    </div>
  );
}

interface PanelMessageProps {
  readonly title: string;
  readonly message: string;
  readonly className?: string;
}

function PanelMessage({ title, message, className }: PanelMessageProps) {
  return (
    <HudPanel title={title} accent="empty" className={cn(className)}>
      <p className="text-sm leading-6 text-pewter">{message}</p>
    </HudPanel>
  );
}
