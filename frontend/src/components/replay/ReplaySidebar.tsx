import { useQuery } from '@tanstack/react-query';

import { getCampaignCharacter } from '../../api/characters';

interface ReplaySidebarProps {
  readonly campaignId: string;
}

export function ReplaySidebar({ campaignId }: ReplaySidebarProps) {
  const { data: character, isLoading, isError } = useQuery({
    queryKey: ['character', campaignId],
    queryFn: () => getCampaignCharacter(campaignId),
    enabled: campaignId.length > 0,
  });

  return (
    <aside className="border-2 border-jade/20 bg-charcoal p-5">
      <h2 className="mb-4 font-heading text-sm font-semibold uppercase tracking-[0.2em] text-jade">
        Character
      </h2>

      {isLoading && (
        <p className="text-xs text-champagne/50">Loading character data...</p>
      )}

      {isError && (
        <p className="text-xs text-ruby/70">Unable to load character.</p>
      )}

      {character && (
        <div className="space-y-4">
          <div>
            <p className="text-lg font-semibold text-champagne">{character.name}</p>
            <p className="text-xs uppercase tracking-[0.2em] text-pewter">Level {character.level}</p>
          </div>

          <div className="border-t border-jade/15 pt-3">
            <dl className="space-y-2 text-sm">
              <div className="flex items-center justify-between">
                <dt className="font-medium text-jade/80">HP</dt>
                <dd className="text-champagne">
                  <span className="text-jade">{character.hp}</span>
                  <span className="mx-1 text-pewter">/</span>
                  <span>{character.max_hp}</span>
                </dd>
              </div>
              <div className="flex items-center justify-between">
                <dt className="font-medium text-jade/80">XP</dt>
                <dd className="text-champagne">{character.experience}</dd>
              </div>
              <div className="flex items-center justify-between">
                <dt className="font-medium text-jade/80">Status</dt>
                <dd className="text-champagne capitalize">{character.status}</dd>
              </div>
            </dl>
          </div>

          <div className="border-t border-jade/15 pt-3">
            <p className="text-[10px] uppercase tracking-[0.18em] text-pewter/60">
              Character state shown as of latest save
            </p>
          </div>
        </div>
      )}
    </aside>
  );
}
