import { useContext } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getCampaignCharacter } from '../../api/characters';
import { CampaignContext } from '../../context/CampaignContext';
import { cn } from '../../lib/cn';
import { HudPanel } from '../layout/HudPanel';
import { AbilityList } from './AbilityList';
import { FeatBrowser } from './FeatBrowser';
import { SkillTree } from './SkillTree';

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
  const statEntries = Object.entries(character.stats);

  return (
    <div className={cn('grid gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(18rem,0.95fr)]', className)}>
      <div className="space-y-6">
        <HudPanel title="Character" accent="vitals" className="deco-corners deco-corners-jade deco-pattern" bodyClassName="p-6">
          <div className="flex flex-wrap items-start justify-between gap-4 border-b border-jade/15 pb-5">
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

          <div className="mt-5 grid gap-5 xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)]">
            <div className="space-y-4">
              <VitalMeter label="HP" value={character.hp} max={character.max_hp} tone="jade" />
              <XPProgressLine level={character.level} experience={character.experience} />

              <dl className="space-y-2 border-t border-jade/15 pt-4 text-sm text-champagne/70">
                <InfoLine label="Level" value={String(character.level)} />
                <InfoLine label="Status" value={humanizeInlineValue(character.status)} />
                <InfoLine
                  label="Location"
                  value={character.current_location_id ? character.current_location_id : 'Location not recorded'}
                  mono={Boolean(character.current_location_id)}
                />
                <InfoLine label="Character ID" value={character.id} mono />
              </dl>
            </div>

            <div className="border-l border-jade/15 pl-5">
              <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-jade/80">Stats</p>
              {statEntries.length > 0 ? (
                <dl className="mt-3 grid gap-x-5 gap-y-2 sm:grid-cols-2">
                  {statEntries.map(([key, value]) => (
                    <div key={key} className="flex items-baseline justify-between gap-3 border-b border-jade/10 pb-2 text-sm">
                      <dt className="text-pewter/80">{humanizeStatKey(key)}</dt>
                      <dd className="font-semibold text-champagne">{formatStatValue(value)}</dd>
                    </div>
                  ))}
                </dl>
              ) : (
                <p className="mt-3 text-sm leading-6 text-pewter">No stats have been recorded for this character yet.</p>
              )}
            </div>
          </div>
        </HudPanel>
      </div>

      <div className="space-y-6">
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

function XPProgressLine({ level, experience }: { readonly level: number; readonly experience: number }) {
  const threshold = nextLevelThreshold(level);
  const prevThreshold = level > 1 ? nextLevelThreshold(level - 1) : 0;
  const xpInLevel = experience - prevThreshold;
  const xpNeeded = threshold - prevThreshold;
  const pct = xpNeeded > 0 ? Math.min(100, Math.round((xpInLevel / xpNeeded) * 100)) : 100;

  return (
    <div>
      <div className="flex items-baseline justify-between gap-3">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">Experience</p>
        <p className="text-sm font-semibold text-sapphire">
          {experience} <span className="text-pewter">/ {threshold}</span>
        </p>
      </div>
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

function VitalMeter({
  label,
  value,
  max,
  tone,
}: {
  readonly label: string;
  readonly value: number;
  readonly max: number;
  readonly tone: 'jade' | 'sapphire';
}) {
  const pct = max > 0 ? Math.min(100, Math.round((value / max) * 100)) : 0;

  return (
    <div>
      <div className="flex items-baseline justify-between gap-3">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{label}</p>
        <p className={cn('text-sm font-semibold', tone === 'jade' ? 'text-jade' : 'text-sapphire')}>
          {value} <span className="text-pewter">/ {max}</span>
        </p>
      </div>
      <div className="mt-2 h-2 w-full overflow-hidden bg-midnight/50">
        <div
          className={cn('h-full transition-all duration-500', tone === 'jade' ? 'bg-jade' : 'bg-sapphire')}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}

function InfoLine({ label, value, mono = false }: { readonly label: string; readonly value: string; readonly mono?: boolean }) {
  return (
    <div className="flex items-baseline justify-between gap-4">
      <dt className="text-pewter">{label}</dt>
      <dd className={cn('max-w-[65%] truncate text-right text-champagne/80', mono ? 'font-mono text-xs' : '')}>{value}</dd>
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

function humanizeStatKey(key: string): string {
  return key
    .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function formatStatValue(value: unknown): string {
  if (typeof value === 'number') {
    return Number.isFinite(value) ? value.toLocaleString() : String(value);
  }

  if (typeof value === 'string') {
    return value.trim().length > 0 ? value : '—';
  }

  if (typeof value === 'boolean') {
    return value ? 'True' : 'False';
  }

  if (value === null || value === undefined) {
    return '—';
  }

  if (Array.isArray(value)) {
    const parts = value.map((item) => formatStatValue(item)).filter((item) => item !== '—');
    return parts.length > 0 ? parts.join(', ') : '—';
  }

  return JSON.stringify(value);
}
