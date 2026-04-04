import { cn } from '../../lib/cn';
import type { NarrativeChoiceOption } from './NarrativeEntry';

interface ChoiceListProps {
  readonly choices: NarrativeChoiceOption[];
  readonly onSelectChoice: (choiceText: string) => void | Promise<void>;
  readonly disabled?: boolean;
  readonly className?: string;
  readonly title?: string;
}

export function ChoiceList({
  choices,
  onSelectChoice,
  disabled = false,
  className,
  title = 'Suggested choices',
}: ChoiceListProps) {
  if (choices.length === 0) {
    return null;
  }

  return (
    <section className={cn('space-y-3', className)} aria-label={title}>
      <div className="flex items-center gap-3">
        <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">{title}</p>
        <div className="h-px flex-1 bg-gradient-to-r from-gold/40 to-transparent" />
      </div>
      <div className="grid gap-2 sm:grid-cols-2">
        {choices.map((choice) => (
          <button
            key={choice.id}
            type="button"
            disabled={disabled}
            onClick={() => {
              void onSelectChoice(choice.text);
            }}
            className="group flex min-h-14 cursor-pointer items-start justify-between gap-3 border-2 border-gold/15 bg-charcoal px-4 py-3 text-left text-sm text-champagne/80 transition-all duration-200 hover:border-gold/50 hover:bg-gold/5 hover:translate-x-1 focus:outline-none focus:ring-2 focus:ring-gold/40 disabled:cursor-not-allowed disabled:border-gold/10 disabled:text-pewter disabled:hover:translate-x-0"
          >
            <span className="leading-6">{choice.text}</span>
            <span className="mt-1 text-xs uppercase tracking-[0.2em] text-gold/60 transition group-hover:text-gold">
              Choose
            </span>
          </button>
        ))}
      </div>
    </section>
  );
}
