export interface CampaignAttributesValue {
  readonly genre: string;
  readonly settingStyle: string;
  readonly tone: string;
}

interface CampaignAttributesFormProps {
  readonly value: CampaignAttributesValue;
  readonly onChange: (value: CampaignAttributesValue) => void;
  readonly onContinue: (value: CampaignAttributesValue) => void | Promise<void>;
  readonly onBack?: () => void;
  readonly continueLabel?: string;
  readonly continueLoadingLabel?: string;
  readonly backLabel?: string;
  readonly errorMessage?: string | null;
  readonly isLoading?: boolean;
}

interface AttributeOption {
  readonly label: string;
  readonly value: string;
  readonly description: string;
}

const genreOptions: readonly AttributeOption[] = [
  { label: 'Fantasy', value: 'Fantasy', description: 'Magic, mythic threats, lost kingdoms, and ancient ruins.' },
  { label: 'Sci-Fi', value: 'Sci-Fi', description: 'Advanced technology, strange worlds, and speculative futures.' },
  { label: 'Horror', value: 'Horror', description: 'Fear, dread, and survival against overwhelming darkness.' },
  { label: 'Historical', value: 'Historical', description: 'Ground the campaign in a recognizable real-world era.' },
  { label: 'Modern', value: 'Modern', description: 'Contemporary life colliding with mystery, danger, or the uncanny.' },
  { label: 'Post-Apocalyptic', value: 'Post-Apocalyptic', description: 'Scarcity, collapse, and rebuilding after catastrophe.' },
  { label: 'Steampunk', value: 'Steampunk', description: 'Industrial wonder, strange engines, and clockwork ambition.' },
];

const settingStyleOptions: readonly AttributeOption[] = [
  { label: 'Open Wilderness', value: 'Open Wilderness', description: 'Frontier travel, scattered settlements, and untamed regions.' },
  { label: 'Urban Sprawl', value: 'Urban Sprawl', description: 'Dense cities, district politics, and layered social factions.' },
  { label: 'Island Archipelago', value: 'Island Archipelago', description: 'Sea lanes, isolated cultures, and contested ports.' },
  { label: 'Underground', value: 'Underground', description: 'Caverns, buried cities, and claustrophobic depths.' },
  { label: 'War-Torn Kingdom', value: 'War-Torn Kingdom', description: 'Broken borders, military pressure, and unstable loyalties.' },
  { label: 'Peaceful Realm', value: 'Peaceful Realm', description: 'Calm on the surface, with tension hidden under stability.' },
];

const toneOptions: readonly AttributeOption[] = [
  { label: 'Gritty and Dark', value: 'Gritty and Dark', description: 'Hard choices, costly victories, and constant danger.' },
  { label: 'Light-Hearted', value: 'Light-Hearted', description: 'Warmth, optimism, and room for playful turns.' },
  { label: 'Epic and Grand', value: 'Epic and Grand', description: 'Mythic stakes, sweeping scale, and larger-than-life action.' },
  { label: 'Mysterious', value: 'Mysterious', description: 'Hidden motives, buried truths, and slow-burn discovery.' },
  { label: 'Humorous', value: 'Humorous', description: 'Comic beats, absurdity, and a lighter dramatic touch.' },
  {
    label: 'Tense and Suspenseful',
    value: 'Tense and Suspenseful',
    description: 'Pressure, uncertainty, and danger that keeps tightening.',
  },
];

export {
  genreOptions as campaignGenreOptions,
  settingStyleOptions as campaignSettingStyleOptions,
  toneOptions as campaignToneOptions,
};
export type { CampaignAttributesFormProps };

type CampaignAttributeField = keyof CampaignAttributesValue;

export function CampaignAttributesForm({
  value,
  onChange,
  onContinue,
  onBack,
  continueLabel = 'Generate proposals',
  continueLoadingLabel = 'Generating…',
  backLabel = 'Back',
  errorMessage = null,
  isLoading = false,
}: CampaignAttributesFormProps) {
  const missingFields = getMissingFields(value);

  function updateField(field: CampaignAttributeField, nextValue: string) {
    onChange({
      ...value,
      [field]: nextValue,
    });
  }

  function handleContinue() {
    if (missingFields.length > 0 || isLoading) {
      return;
    }

    void onContinue(value);
  }

  return (
    <section className="deco-corners deco-pattern space-y-6 border-2 border-gold/20 bg-charcoal p-6">
      <header className="space-y-3 border-b-2 border-gold/20 pb-5">
        <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Campaign attributes</p>
        <div className="space-y-2">
          <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">Choose the campaign frame</h2>
          <p className="max-w-2xl text-sm leading-7 text-champagne/70">
            Pick the core genre, world shape, and tone. These selections seed the proposal generator before world-building hands off into play.
          </p>
        </div>
      </header>

      <div className="space-y-6">
        <AttributeSection
          title="Genre"
          description="What broad fiction tradition should the campaign draw from?"
          options={genreOptions}
          selectedValue={value.genre}
          onSelect={(nextValue) => {
            updateField('genre', nextValue);
          }}
        />
        <AttributeSection
          title="Setting style"
          description="What kind of space should the campaign mostly unfold in?"
          options={settingStyleOptions}
          selectedValue={value.settingStyle}
          onSelect={(nextValue) => {
            updateField('settingStyle', nextValue);
          }}
        />
        <AttributeSection
          title="Tone"
          description="How should the campaign feel moment to moment?"
          options={toneOptions}
          selectedValue={value.tone}
          onSelect={(nextValue) => {
            updateField('tone', nextValue);
          }}
        />
      </div>

      {missingFields.length > 0 ? (
        <div className="border border-gold/40 bg-gold/10 px-4 py-3 text-sm text-gold">
          Select a value for {formatMissingFields(missingFields)} before continuing.
        </div>
      ) : null}

      {errorMessage ? (
        <div className="border border-ruby/40 bg-ruby/10 px-4 py-3 text-sm text-ruby">{errorMessage}</div>
      ) : null}

      <div className="flex flex-col-reverse gap-3 border-t-2 border-gold/20 pt-5 sm:flex-row sm:items-center sm:justify-between">
        {onBack ? (
          <button
            type="button"
            onClick={onBack}
            className="inline-flex items-center justify-center border border-gold/30 px-4 py-3 text-sm font-semibold uppercase tracking-wide text-champagne/80 transition hover:border-gold hover:text-champagne focus:outline-none focus:ring-2 focus:ring-gold/60"
          >
            {backLabel}
          </button>
        ) : (
          <div />
        )}
        <button
          type="button"
          onClick={handleContinue}
          disabled={missingFields.length > 0 || isLoading}
          className="inline-flex items-center justify-center bg-ruby px-5 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition hover:bg-ruby-light focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter"
        >
          {isLoading ? continueLoadingLabel : continueLabel}
        </button>
      </div>
    </section>
  );
}

function AttributeSection({
  title,
  description,
  options,
  selectedValue,
  onSelect,
}: {
  readonly title: string;
  readonly description: string;
  readonly options: readonly AttributeOption[];
  readonly selectedValue: string;
  readonly onSelect: (value: string) => void;
}) {
  return (
    <section className="space-y-3 border-2 border-gold/15 bg-charcoal/80 p-5">
      <div className="space-y-1">
        <h3 className="font-heading text-lg font-semibold uppercase tracking-wide text-champagne">{title}</h3>
        <p className="text-sm leading-6 text-pewter">{description}</p>
      </div>
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {options.map((option) => {
          const isSelected = selectedValue === option.value;

          return (
            <button
              key={option.value}
              type="button"
              onClick={() => {
                onSelect(option.value);
              }}
              aria-pressed={isSelected}
              className={[
                'border-2 px-4 py-4 text-left transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gold/60',
                isSelected
                  ? 'border-gold bg-gold/10'
                  : 'border-gold/15 bg-obsidian hover:border-gold/40 hover:bg-charcoal hover:-translate-y-0.5',
              ].join(' ')}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="space-y-2">
                  <p className="font-heading text-base font-semibold text-champagne">{option.label}</p>
                  <p className="text-sm leading-6 text-champagne/70">{option.description}</p>
                </div>
                <span
                  aria-hidden="true"
                  className={[
                    'mt-1 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-xs font-semibold',
                    isSelected
                      ? 'border-gold bg-gold text-obsidian'
                      : 'border-gold/30 bg-obsidian text-pewter',
                  ].join(' ')}
                >
                  {isSelected ? '✓' : ''}
                </span>
              </div>
            </button>
          );
        })}
      </div>
    </section>
  );
}

function getMissingFields(value: CampaignAttributesValue): CampaignAttributeField[] {
  const missingFields: CampaignAttributeField[] = [];

  if (!value.genre.trim()) {
    missingFields.push('genre');
  }
  if (!value.settingStyle.trim()) {
    missingFields.push('settingStyle');
  }
  if (!value.tone.trim()) {
    missingFields.push('tone');
  }

  return missingFields;
}

function formatMissingFields(fields: readonly CampaignAttributeField[]): string {
  const labels = fields.map((field) => {
    switch (field) {
      case 'genre':
        return 'genre';
      case 'settingStyle':
        return 'setting style';
      case 'tone':
        return 'tone';
      default: {
        const exhaustiveField: never = field;
        return exhaustiveField;
      }
    }
  });

  if (labels.length === 1) {
    return labels[0];
  }

  if (labels.length === 2) {
    return `${labels[0]} and ${labels[1]}`;
  }

  return `${labels.slice(0, -1).join(', ')}, and ${labels.at(-1)}`;
}
