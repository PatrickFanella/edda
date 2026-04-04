import type { CharacterAbility } from '../../api/types';
import { cn } from '../../lib/cn';

interface AbilityListProps {
  readonly abilities: CharacterAbility[];
  readonly className?: string;
  readonly title?: string;
}

export function AbilityList({ abilities, className, title = 'Abilities' }: AbilityListProps) {
  return (
    <section className={cn('border-2 border-jade/20 bg-charcoal p-6', className)}>
      <div>
        <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-jade">Character</p>
        <h3 className="font-heading mt-2 text-lg font-semibold uppercase tracking-[0.1em] text-champagne">{title}</h3>
        <p className="mt-2 text-sm leading-6 text-pewter">Capabilities and special actions currently attached to this character.</p>
      </div>

      {abilities.length === 0 ? (
        <div className="mt-5 border border-dashed border-jade/15 bg-charcoal/50 px-4 py-6 text-sm text-pewter">
          No abilities are available yet. The character can still act through the narrative tab while new capabilities are discovered.
        </div>
      ) : (
        <ul className="mt-5 space-y-3">
          {abilities.map((ability) => (
            <li key={ability.name} className="border border-jade/20 bg-charcoal/80 p-4 transition-all duration-200 hover:border-sapphire/30">
              <h4 className="text-base font-semibold text-champagne">{ability.name}</h4>
              <p className="mt-2 text-sm leading-6 text-champagne/70">
                {ability.description && ability.description.trim().length > 0
                  ? ability.description
                  : 'No ability description has been recorded yet.'}
              </p>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
