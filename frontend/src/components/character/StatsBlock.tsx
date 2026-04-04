import { cn } from '../../lib/cn';

interface StatsBlockProps {
  readonly stats: Record<string, unknown>;
  readonly className?: string;
  readonly title?: string;
}

export function StatsBlock({ stats, className, title = 'Stats' }: StatsBlockProps) {
  const entries = Object.entries(stats);

  return (
    <section className={cn('border-2 border-jade/20 bg-charcoal p-6', className)}>
      <div className="flex items-center gap-3">
        <div>
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-jade">Character</p>
          <h3 className="font-heading mt-2 text-lg font-semibold uppercase tracking-[0.1em] text-champagne">{title}</h3>
        </div>
      </div>

      {entries.length === 0 ? (
        <div className="mt-5 border border-dashed border-jade/15 bg-charcoal/50 px-4 py-6 text-sm text-pewter">
          No stats have been recorded for this character yet.
        </div>
      ) : (
        <dl className="mt-5 grid gap-4 sm:grid-cols-2">
          {entries.map(([key, value]) => (
            <div key={key} className="border border-jade/20 bg-charcoal/80 p-4 transition-all duration-200 hover:border-jade/40">
              <dt className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{humanizeStatKey(key)}</dt>
              <dd className="mt-3 text-lg font-semibold text-champagne">{formatStatValue(value)}</dd>
            </div>
          ))}
        </dl>
      )}
    </section>
  );
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
