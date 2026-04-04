import type { ItemResponse } from '../../api/types';
import { cn } from '../../lib/cn';

interface ItemCardProps {
  readonly item: ItemResponse;
  readonly className?: string;
}

const RARITY_TONE: Record<string, string> = {
  common: 'border-pewter/30 bg-pewter/10 text-pewter',
  uncommon: 'border-jade/30 bg-jade/10 text-jade',
  rare: 'border-sapphire/40 bg-sapphire/10 text-sapphire',
  epic: 'border-ruby/40 bg-ruby/10 text-ruby',
  legendary: 'border-gold/40 bg-gold/10 text-gold',
};

export function ItemCard({ item, className }: ItemCardProps) {
  const propertyEntries = Object.entries(item.properties ?? {});
  const rarityTone = RARITY_TONE[item.rarity.toLowerCase()] ?? 'border-pewter/30 bg-pewter/10 text-pewter';

  return (
    <article
      className={cn(
        'deco-pattern border-2 bg-charcoal p-5 transition-all duration-200 hover:-translate-y-0.5',
        item.equipped
          ? 'border-jade/50 bg-jade/5'
          : 'border-gold/20 hover:border-gold/40',
        className,
      )}
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="font-heading text-base font-semibold uppercase tracking-wide text-champagne">{item.name}</h3>
            {item.equipped ? <StatusBadge tone="equipped">Equipped</StatusBadge> : null}
            <StatusBadge tone="quantity">{`Qty ${item.quantity}`}</StatusBadge>
          </div>
          <div className="flex flex-wrap gap-2">
            <StatusBadge tone="neutral">{humanizeToken(item.item_type)}</StatusBadge>
            <span className={cn('inline-flex rounded-sm border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em]', rarityTone)}>
              {humanizeToken(item.rarity)}
            </span>
          </div>
        </div>
      </div>

      <p className="mt-4 text-sm leading-6 text-champagne/70">{item.description || 'No description provided.'}</p>

      <div className="mt-5 space-y-3">
        <div className="flex items-center gap-3">
          <p className="font-heading text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter/80">Properties</p>
          <div className="h-px flex-1 bg-gradient-to-r from-pewter/30 to-transparent" />
        </div>

        {propertyEntries.length > 0 ? (
          <dl className="grid gap-2 sm:grid-cols-2">
            {propertyEntries.map(([key, value]) => (
              <div key={key} className="border border-pewter/15 bg-charcoal/80 px-3 py-2">
                <dt className="text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter/80">{humanizeToken(key)}</dt>
                <dd className="mt-1 break-words text-sm leading-5 text-champagne">{formatCompactValue(value)}</dd>
              </div>
            ))}
          </dl>
        ) : (
          <p className="border border-dashed border-pewter/15 bg-charcoal/50 px-3 py-3 text-sm text-pewter">
            No item properties.
          </p>
        )}
      </div>
    </article>
  );
}

function StatusBadge({ children, tone }: { readonly children: string; readonly tone: 'equipped' | 'quantity' | 'neutral' }) {
  const toneClassName =
    tone === 'equipped'
      ? 'border-jade/40 bg-jade/10 text-jade'
      : tone === 'quantity'
        ? 'border-gold/30 bg-gold/10 text-gold'
        : 'border-pewter/30 bg-pewter/10 text-pewter';

  return <span className={cn('inline-flex rounded-sm border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.2em]', toneClassName)}>{children}</span>;
}

function humanizeToken(value: string): string {
  return value
    .replace(/[_-]+/g, ' ')
    .trim()
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

function formatCompactValue(value: unknown, depth = 0): string {
  if (value === null) {
    return 'null';
  }

  if (typeof value === 'string') {
    const normalized = value.trim();
    if (!normalized) {
      return '""';
    }

    return normalized.length > 48 ? `${normalized.slice(0, 45)}…` : normalized;
  }

  if (typeof value === 'number' || typeof value === 'boolean' || typeof value === 'bigint') {
    return String(value);
  }

  if (Array.isArray(value)) {
    if (value.length === 0) {
      return '[]';
    }

    if (depth >= 2) {
      return '[…]';
    }

    const renderedItems = value.slice(0, 3).map((entry) => formatCompactValue(entry, depth + 1));
    return `[${renderedItems.join(', ')}${value.length > 3 ? ', …' : ''}]`;
  }

  if (typeof value === 'object') {
    const entries = Object.entries(value);
    if (entries.length === 0) {
      return '{}';
    }

    if (depth >= 2) {
      return '{…}';
    }

    const renderedEntries = entries
      .slice(0, 3)
      .map(([key, nestedValue]) => `${key}: ${formatCompactValue(nestedValue, depth + 1)}`);

    return `{ ${renderedEntries.join(', ')}${entries.length > 3 ? ', …' : ''} }`;
  }

  return String(value);
}
