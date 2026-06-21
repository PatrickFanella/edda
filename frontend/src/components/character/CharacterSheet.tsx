import { useContext } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getCampaignCharacter } from '../../api/characters';
import { CampaignContext } from '../../context/CampaignContext';
import { cn } from '../../lib/cn';
import { HudPanel } from '../layout/HudPanel';
import { AbilityList } from './AbilityList';
import { FeatBrowser } from './FeatBrowser';
import { SkillTree } from './SkillTree';
import { StatsBlock } from './StatsBlock';

interface CharacterSheetProps {
  readonly campaignId?: string;
  readonly className?: string;
}

export function CharacterSheet({ campaignId, className }: CharacterSheetProps) {
  const campaign = useContext(CampaignContext);
  const activeCampaignId = campaignId ?? campaign?.campaignId ?? '';

  const characterQuery = useQuery({
    queryKey: ['campaign', activeCampaignId, 'character'],
    queryFn: () => getCampaignCharacter(activeCampaignId),
    enabled: activeCampaignId.length > 0,
  });

  if (!activeCampaignId) {
    return (
      <HudPanel title="Character" accent="empty" className={className}>
        <p className="text-sm leading-6 text-pewter">Open a campaign before loading the character sheet.</p>
      </HudPanel>
    );
  }

  if (characterQuery.isPending) {
    return (
      <HudPanel title="Character" accent="loading" className={className}>
        <p className="text-sm leading-6 text-pewter">Fetching the latest character sheet…</p>
      </HudPanel>
    );
  }

  if (characterQuery.isError) {
    return (
      <HudPanel title="Character" accent="error" className={className}>
        <p className="text-sm leading-6 text-ruby">{queryErrorMessage(characterQuery.error)}</p>
      </HudPanel>
    );
  }

  const character = characterQuery.data;

  return (
    <div className={cn('grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.95fr)]', className)}>
      <div className="space-y-6">
        <HudPanel title="Character" accent="vitals" className="deco-corners deco-corners-jade deco-pattern" bodyClassName="p-6">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="space-y-3">
              <div>
                <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-jade">Character</p>
                <h2 className="font-heading mt-2 text-2xl font-semibold uppercase tracking-[0.1em] text-champagne">{character.name}</h2>
              </div>
              <p className="max-w-2xl text-sm leading-7 text-champagne/70">
                {character.description.trim().length > 0 ? character.description : 'No character description has been recorded yet.'}
              </p>
            </div>
            <div className="inline-flex rounded-sm border border-jade/30 bg-jade/10 px-3 py-1 text-xs font-semibold uppercase tracking-[0.2em] text-jade">
              {humanizeInlineValue(character.status)}
            </div>
          </div>

          <div className="mt-6 grid gap-4 sm:grid-cols-3">
            <MetricCard label="HP" value={`${character.hp} / ${character.max_hp}`} accent="text-jade" />
            <MetricCard label="Level" value={String(character.level)} accent="text-gold" />
            <XPProgressCard level={character.level} experience={character.experience} />
          </div>
        </HudPanel>

        <StatsBlock stats={character.stats} />
      </div>

      <div className="space-y-6">
        <HudPanel title="Current standing" accent="vitals" bodyClassName="p-6">
          <dl className="mt-4 space-y-3 text-sm text-champagne/70">
            <SummaryRow label="Status" value={humanizeInlineValue(character.status)} />
            <SummaryRow label="Character ID" value={character.id} mono />
            <SummaryRow
              label="Current location"
              value={character.current_location_id ? character.current_location_id : 'Location not recorded'}
              mono={Boolean(character.current_location_id)}
            />
          </dl>
        </HudPanel>

        <AbilityList abilities={character.abilities} />

        {campaign?.campaign?.rules_mode === 'crunch' && (
          <>
            <FeatBrowser campaignId={activeCampaignId} />
            <SkillTree campaignId={activeCampaignId} />
          </>
        )}
      </div>
    </div>
  );
}

/** Returns cumulative XP needed to reach the next level. Mirrors progression/leveling.go. */
function nextLevelThreshold(level: number): number {
  return 1000 * level * (level + 1) / 2;
}

function XPProgressCard({ level, experience }: { readonly level: number; readonly experience: number }) {
  const threshold = nextLevelThreshold(level);
  const prevThreshold = level > 1 ? nextLevelThreshold(level - 1) : 0;
  const xpInLevel = experience - prevThreshold;
  const xpNeeded = threshold - prevThreshold;
  const pct = xpNeeded > 0 ? Math.min(100, Math.round((xpInLevel / xpNeeded) * 100)) : 100;

  return (
    <div className="rounded-none border border-jade/20 bg-charcoal/80 p-4 transition-all duration-200 hover:border-jade/40">
      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">Experience</p>
      <p className="mt-2 text-lg font-semibold tracking-tight text-sapphire">
        {experience} <span className="text-sm text-pewter">/ {threshold}</span>
      </p>
      <div className="mt-2 h-2 w-full overflow-hidden bg-midnight/50">
        <div
          className="h-full bg-sapphire transition-all duration-500"
          style={{ width: `${pct}%` }}
        />
      </div>
      <p className="mt-1 text-[10px] uppercase tracking-[0.15em] text-pewter/70">{pct}% to level {level + 1}</p>
    </div>
  );
}

function MetricCard({
  label,
  value,
  accent,
}: {
  readonly label: string;
  readonly value: string;
  readonly accent: string;
}) {
  return (
    <div className="rounded-none border border-jade/20 bg-charcoal/80 p-4 transition-all duration-200 hover:border-jade/40">
      <p className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{label}</p>
      <p className={cn('mt-3 text-2xl font-semibold tracking-tight text-champagne', accent)}>{value}</p>
    </div>
  );
}

function SummaryRow({
  label,
  value,
  mono = false,
}: {
  readonly label: string;
  readonly value: string;
  readonly mono?: boolean;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-pewter">{label}</dt>
      <dd className={cn('max-w-[60%] text-right text-champagne/80', mono ? 'font-mono text-xs leading-6 text-champagne/60' : '')}>{value}</dd>
    </div>
  );
}

function queryErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to load the character sheet.';
}

function humanizeInlineValue(value: string): string {
  return value
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}
