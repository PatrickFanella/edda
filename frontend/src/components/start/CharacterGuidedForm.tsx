import { type ChangeEvent, type FormEvent, useMemo, useState } from 'react';

export interface CharacterGuidedFormData {
  readonly name: string;
  readonly race: string;
  readonly class: string;
  readonly background: string;
  readonly alignment: string;
}

export interface CharacterGuidedFormProps {
  readonly value: CharacterGuidedFormData;
  readonly onChange: (value: CharacterGuidedFormData) => void;
  readonly onSubmit: (value: CharacterGuidedFormData) => void;
  readonly onBack?: () => void;
  readonly isSubmitting?: boolean;
  readonly errorMessage?: string | null;
  readonly continueLabel?: string;
  readonly backLabel?: string;
}

type CharacterGuidedField = keyof CharacterGuidedFormData;
type CharacterGuidedErrors = Partial<Record<CharacterGuidedField, string>>;

interface SelectFieldConfig {
  readonly key: Exclude<CharacterGuidedField, 'name'>;
  readonly label: string;
  readonly placeholder: string;
  readonly options: readonly string[];
}

export const CHARACTER_RACE_OPTIONS = [
  'Human',
  'Elf',
  'Dwarf',
  'Halfling',
  'Gnome',
  'Half-Elf',
  'Half-Orc',
  'Tiefling',
  'Dragonborn',
] as const;

export const CHARACTER_CLASS_OPTIONS = [
  'Barbarian',
  'Bard',
  'Cleric',
  'Druid',
  'Fighter',
  'Monk',
  'Paladin',
  'Ranger',
  'Rogue',
  'Sorcerer',
  'Warlock',
  'Wizard',
] as const;

export const CHARACTER_BACKGROUND_OPTIONS = [
  'Acolyte',
  'Charlatan',
  'Criminal',
  'Entertainer',
  'Folk Hero',
  'Guild Artisan',
  'Hermit',
  'Noble',
  'Outlander',
  'Sage',
  'Sailor',
  'Soldier',
  'Urchin',
] as const;

export const CHARACTER_ALIGNMENT_OPTIONS = [
  'Lawful Good',
  'Neutral Good',
  'Chaotic Good',
  'Lawful Neutral',
  'True Neutral',
  'Chaotic Neutral',
  'Lawful Evil',
  'Neutral Evil',
  'Chaotic Evil',
] as const;

const SELECT_FIELDS: readonly SelectFieldConfig[] = [
  {
    key: 'race',
    label: 'Race',
    placeholder: 'Choose a race',
    options: CHARACTER_RACE_OPTIONS,
  },
  {
    key: 'class',
    label: 'Class',
    placeholder: 'Choose a class',
    options: CHARACTER_CLASS_OPTIONS,
  },
  {
    key: 'background',
    label: 'Background',
    placeholder: 'Choose a background',
    options: CHARACTER_BACKGROUND_OPTIONS,
  },
  {
    key: 'alignment',
    label: 'Alignment',
    placeholder: 'Choose an alignment',
    options: CHARACTER_ALIGNMENT_OPTIONS,
  },
] as const;

function validateGuidedCharacter(value: CharacterGuidedFormData): CharacterGuidedErrors {
  const errors: CharacterGuidedErrors = {};

  if (value.name.trim().length === 0) {
    errors.name = 'Required';
  }

  for (const field of SELECT_FIELDS) {
    if (value[field.key].trim().length === 0) {
      errors[field.key] = 'Choose one';
    }
  }

  return errors;
}

function fieldHasValue(value: string): boolean {
  return value.trim().length > 0;
}

export function CharacterGuidedForm({
  value,
  onChange,
  onSubmit,
  onBack,
  isSubmitting = false,
  errorMessage,
  continueLabel = 'Continue to confirmation',
  backLabel = 'Back',
}: CharacterGuidedFormProps) {
  const [fieldErrors, setFieldErrors] = useState<CharacterGuidedErrors>({});

  const errorCount = useMemo(() => Object.keys(fieldErrors).length, [fieldErrors]);

  function updateField<Key extends CharacterGuidedField>(key: Key, nextValue: CharacterGuidedFormData[Key]) {
    onChange({
      ...value,
      [key]: nextValue,
    });

    if (fieldErrors[key] && fieldHasValue(String(nextValue))) {
      setFieldErrors((current) => {
        const nextErrors = { ...current };
        delete nextErrors[key];
        return nextErrors;
      });
    }
  }

  function handleNameChange(event: ChangeEvent<HTMLInputElement>) {
    updateField('name', event.target.value);
  }

  function handleSelectChange(field: Exclude<CharacterGuidedField, 'name'>) {
    return (event: ChangeEvent<HTMLSelectElement>) => {
      updateField(field, event.target.value);
    };
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const nextErrors = validateGuidedCharacter(value);
    setFieldErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    onSubmit({
      ...value,
      name: value.name.trim(),
    });
  }

  return (
    <div className="deco-corners deco-corners-jade deco-pattern border-2 border-jade/20 bg-charcoal p-6">
      <div className="mb-6 space-y-2">
        <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-jade">Step 5 · Guided character</p>
        <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">Build your adventurer</h2>
        <p className="max-w-2xl text-sm leading-6 text-champagne/70">
          Mirror the TUI flow: lock in name, race, class, background, and alignment before the final confirmation step.
        </p>
      </div>

      <form className="space-y-6" onSubmit={handleSubmit} noValidate>
        {errorCount > 0 ? (
          <div className="border border-ruby/40 bg-ruby/10 px-4 py-3 text-sm text-ruby">
            Fix {errorCount} {errorCount === 1 ? 'field' : 'fields'} to continue.
          </div>
        ) : null}

        {errorMessage ? (
          <div className="border border-ruby/40 bg-ruby/10 px-4 py-3 text-sm text-ruby">
            {errorMessage}
          </div>
        ) : null}

        <div className="grid gap-6 md:grid-cols-2">
          <label className="space-y-2 md:col-span-2">
            <span className="text-sm font-medium text-champagne/80">Character name</span>
            <input
              type="text"
              value={value.name}
              onChange={handleNameChange}
              aria-invalid={fieldErrors.name ? 'true' : 'false'}
              aria-describedby={fieldErrors.name ? 'character-guided-name-error' : undefined}
              className="w-full border-2 border-gold/20 bg-obsidian px-4 py-3 text-sm text-champagne transition-all duration-200 placeholder:text-pewter/60 focus:border-jade focus:outline-none focus:ring-2 focus:ring-jade/40"
              placeholder="Aela Nightwind"
              autoComplete="off"
            />
            {fieldErrors.name ? (
              <p id="character-guided-name-error" className="text-sm text-ruby">
                {fieldErrors.name}
              </p>
            ) : (
              <p className="text-xs text-pewter">Required.</p>
            )}
          </label>

          {SELECT_FIELDS.map((field) => {
            const errorId = `character-guided-${field.key}-error`;
            const selectedValue = value[field.key];
            const fieldError = fieldErrors[field.key];

            return (
              <label key={field.key} className="space-y-2">
                <span className="text-sm font-medium text-champagne/80">{field.label}</span>
                <select
                  value={selectedValue}
                  onChange={handleSelectChange(field.key)}
                  aria-invalid={fieldError ? 'true' : 'false'}
                  aria-describedby={fieldError ? errorId : undefined}
                  className="w-full border-2 border-gold/20 bg-obsidian px-4 py-3 text-sm text-champagne transition-all duration-200 focus:border-jade focus:outline-none focus:ring-2 focus:ring-jade/40"
                >
                  <option value="" disabled>
                    {field.placeholder}
                  </option>
                  {field.options.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </select>
                {fieldError ? (
                  <p id={errorId} className="text-sm text-ruby">
                    {fieldError}
                  </p>
                ) : null}
              </label>
            );
          })}
        </div>

        <div className="flex flex-col gap-3 border-t-2 border-jade/20 pt-6 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-pewter">The next step reviews both the campaign pitch and this character build.</p>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            {onBack ? (
              <button
                type="button"
                onClick={onBack}
                disabled={isSubmitting}
                className="inline-flex items-center justify-center border border-gold/30 px-4 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition hover:border-gold hover:text-gold focus:outline-none focus:ring-2 focus:ring-gold focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:border-gold/10 disabled:text-pewter"
              >
                {backLabel}
              </button>
            ) : null}
            <button
              type="submit"
              disabled={isSubmitting}
              className="inline-flex items-center justify-center bg-jade px-5 py-3 text-sm font-semibold uppercase tracking-wide text-obsidian transition hover:bg-jade-light focus:outline-none focus:ring-2 focus:ring-jade focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter"
            >
              {isSubmitting ? 'Saving character…' : continueLabel}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
