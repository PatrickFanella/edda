import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getCampaignCharacterInventory } from '../../api/characters';
import type { ItemResponse } from '../../api/types';
import { useCampaign } from '../../hooks/useCampaign';
import { ItemCard } from './ItemCard';

export function InventoryPanel() {
  const { campaign, campaignId } = useCampaign();
  const activeCampaignId = campaignId?.trim() ?? '';

  const inventoryQuery = useQuery({
    queryKey: ['campaign', activeCampaignId, 'character', 'inventory'],
    queryFn: () => getCampaignCharacterInventory(activeCampaignId),
    enabled: activeCampaignId.length > 0,
  });

  const items = useMemo(() => sortInventory(inventoryQuery.data ?? []), [inventoryQuery.data]);
  const equippedCount = items.filter((item) => item.equipped).length;

  if (!activeCampaignId) {
    return (
      <section className="border border-ruby/40 bg-ruby/10 p-6 text-sm text-ruby">
        No active campaign is selected, so inventory cannot load.
      </section>
    );
  }

  if (inventoryQuery.isPending) {
    return (
      <section className="border border-gold/20 bg-charcoal p-6 text-sm text-champagne/70">
        Loading inventory…
      </section>
    );
  }

  if (inventoryQuery.isError) {
    return (
      <section className="border border-ruby/40 bg-ruby/10 p-6 text-sm text-ruby">
        {queryErrorMessage(inventoryQuery.error)}
      </section>
    );
  }

  if (items.length === 0) {
    return (
      <section className="border border-dashed border-pewter/15 bg-charcoal p-8 text-center text-champagne/70">
        <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-pewter/80">Inventory empty</p>
        <p className="mt-3 text-sm leading-6 text-pewter">
          {campaign?.name ?? 'This campaign'} has no items yet. Loot, rewards, and equipped gear will appear here.
        </p>
      </section>
    );
  }

  return (
    <section className="space-y-5">
      <div className="flex flex-wrap items-end justify-between gap-4 border-2 border-gold/20 bg-charcoal px-5 py-4">
        <div>
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Inventory</p>
          <h2 className="font-heading mt-2 text-lg font-semibold uppercase tracking-[0.1em] text-champagne">{campaign?.name ?? 'Campaign'} gear and carried items</h2>
          <p className="mt-1 text-sm text-pewter">Equipped items stay surfaced first. Properties show backend state without flattening away details.</p>
        </div>
        <dl className="grid grid-cols-2 gap-3 text-sm text-champagne/70 sm:min-w-56">
          <SummaryStat label="Items" value={String(items.length)} />
          <SummaryStat label="Equipped" value={String(equippedCount)} />
        </dl>
      </div>

      <div className="grid gap-4 xl:grid-cols-2">
        {items.map((item) => (
          <ItemCard key={item.id} item={item} />
        ))}
      </div>
    </section>
  );
}

function SummaryStat({ label, value }: { readonly label: string; readonly value: string }) {
  return (
    <div className="border border-gold/15 bg-charcoal/80 px-3 py-3 text-right transition-all duration-200 hover:border-gold/30">
      <dt className="text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter/80">{label}</dt>
      <dd className="mt-1 text-lg font-semibold text-champagne">{value}</dd>
    </div>
  );
}

function queryErrorMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Unable to load inventory.';
}

function sortInventory(items: ItemResponse[]): ItemResponse[] {
  return [...items].sort((left, right) => {
    if (left.equipped !== right.equipped) {
      return left.equipped ? -1 : 1;
    }

    const rarityDelta = rarityRank(right.rarity) - rarityRank(left.rarity);
    if (rarityDelta !== 0) {
      return rarityDelta;
    }

    return left.name.localeCompare(right.name);
  });
}

function rarityRank(rarity: string): number {
  switch (rarity.toLowerCase()) {
    case 'legendary':
      return 5;
    case 'epic':
      return 4;
    case 'rare':
      return 3;
    case 'uncommon':
      return 2;
    case 'common':
      return 1;
    default:
      return 0;
  }
}
