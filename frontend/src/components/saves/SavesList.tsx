import { useQuery } from '@tanstack/react-query';

import { listSavePoints } from '../../api/campaigns';
import type { SavePointResponse } from '../../api/types';

interface SavesListProps {
  readonly campaignId: string;
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function SaveBadge({ isAuto }: { readonly isAuto: boolean }) {
  if (isAuto) {
    return (
      <span className="rounded-sm border border-pewter/30 bg-pewter/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-pewter">
        Auto
      </span>
    );
  }
  return (
    <span className="rounded-sm border border-gold/30 bg-gold/10 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-[0.15em] text-gold">
      Manual
    </span>
  );
}

function SaveCard({ save }: { readonly save: SavePointResponse }) {
  return (
    <div className="flex items-center justify-between gap-4 border-2 border-gold/15 bg-charcoal px-4 py-3">
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-semibold text-champagne">{save.name}</p>
        <p className="mt-0.5 text-xs text-pewter">
          Turn {save.turn_number} &middot; {formatDate(save.created_at)}
        </p>
      </div>
      <SaveBadge isAuto={save.is_auto} />
    </div>
  );
}

export function SavesList({ campaignId }: SavesListProps) {
  const { data, isPending, isError, error } = useQuery({
    queryKey: ['campaign', campaignId, 'saves'],
    queryFn: () => listSavePoints(campaignId),
    enabled: campaignId.length > 0,
  });

  if (isPending) {
    return (
      <div className="border border-gold/20 bg-charcoal p-4 text-sm text-champagne/70">
        Loading saves...
      </div>
    );
  }

  if (isError) {
    return (
      <div className="border border-ruby/40 bg-ruby/10 p-4 text-sm text-ruby">
        {error instanceof Error ? error.message : 'Failed to load saves.'}
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="flex min-h-24 flex-col items-center justify-center border border-dashed border-gold/15 bg-charcoal/50 px-6 text-center">
        <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-pewter/80">No saves yet</p>
        <p className="mt-2 max-w-md text-xs leading-5 text-pewter">
          Use the Save button to create a manual save point. Auto-saves are created after each turn.
        </p>
      </div>
    );
  }

  return (
    <section className="space-y-2">
      <h3 className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold/80">
        Save Points ({data.length})
      </h3>
      <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
        {data.map((save) => (
          <SaveCard key={save.id} save={save} />
        ))}
      </div>
    </section>
  );
}
