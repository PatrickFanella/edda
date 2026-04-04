export interface CampaignProfileSummary {
  readonly genre: string;
  readonly tone: string;
  readonly themes: readonly string[];
  readonly world_type?: string;
  readonly danger_level?: string;
  readonly political_complexity?: string;
}

export interface CharacterProfileSummary {
  readonly name: string;
  readonly concept?: string;
  readonly background?: string;
  readonly personality?: string;
  readonly motivations?: readonly string[];
  readonly strengths?: readonly string[];
  readonly weaknesses?: readonly string[];
}

export interface ConfirmationPanelProps {
  readonly campaignName: string;
  readonly campaignSummary?: string | null;
  readonly campaignProfile: CampaignProfileSummary;
  readonly characterProfile: CharacterProfileSummary;
  readonly onBegin: () => void;
  readonly onBack: () => void;
  readonly isBeginning?: boolean;
  readonly errorMessage?: string | null;
  readonly beginLabel?: string;
  readonly backLabel?: string;
}

interface SummaryRowProps {
  readonly label: string;
  readonly value?: string | null;
  readonly fallback?: string;
}

interface SummaryListProps {
  readonly label: string;
  readonly values?: readonly string[];
}

function SummaryRow({ label, value, fallback = 'Not provided yet.' }: SummaryRowProps) {
  const hasValue = value !== undefined && value !== null && value.trim().length > 0;

  return (
    <div className="grid gap-1 sm:grid-cols-[8rem_minmax(0,1fr)] sm:gap-4">
      <dt className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{label}</dt>
      <dd className={hasValue ? 'text-sm leading-6 text-champagne' : 'text-sm leading-6 text-pewter'}>
        {hasValue ? value : fallback}
      </dd>
    </div>
  );
}

function SummaryList({ label, values }: SummaryListProps) {
  if (!values || values.length === 0) {
    return <SummaryRow label={label} />;
  }

  return <SummaryRow label={label} value={values.join(', ')} />;
}

export function ConfirmationPanel({
  campaignName,
  campaignSummary,
  campaignProfile,
  characterProfile,
  onBegin,
  onBack,
  isBeginning = false,
  errorMessage,
  beginLabel = 'Begin adventure',
  backLabel = 'Back',
}: ConfirmationPanelProps) {
  return (
    <div className="deco-corners deco-pattern space-y-6 border-2 border-gold/20 bg-charcoal p-6">
      <div className="space-y-2">
        <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-gold">Step 6 · Confirmation</p>
        <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">Review the campaign and character</h2>
        <p className="max-w-2xl text-sm leading-6 text-champagne/70">
          Confirm the world hook and hero summary, then hand off into world-building and the opening scene.
        </p>
      </div>

      <div className="grid gap-6 xl:grid-cols-2">
        <section className="border-2 border-midnight/40 bg-obsidian p-5 transition-all duration-200 hover:border-midnight/60">
          <div className="mb-4 space-y-1">
            <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-sapphire">Campaign</p>
            <h3 className="font-heading text-xl font-semibold uppercase tracking-wide text-champagne">{campaignName}</h3>
          </div>
          <dl className="space-y-4">
            <SummaryRow label="Summary" value={campaignSummary} fallback="A campaign summary has not been generated yet." />
            <SummaryRow label="Genre" value={campaignProfile.genre} />
            <SummaryRow label="Tone" value={campaignProfile.tone} />
            <SummaryList label="Themes" values={campaignProfile.themes} />
            <SummaryRow label="World type" value={campaignProfile.world_type} />
            <SummaryRow label="Danger" value={campaignProfile.danger_level} />
            <SummaryRow label="Politics" value={campaignProfile.political_complexity} />
          </dl>
        </section>

        <section className="border-2 border-midnight/40 bg-obsidian p-5 transition-all duration-200 hover:border-midnight/60">
          <div className="mb-4 space-y-1">
            <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-jade">Character</p>
            <h3 className="font-heading text-xl font-semibold uppercase tracking-wide text-champagne">{characterProfile.name}</h3>
          </div>
          <dl className="space-y-4">
            <SummaryRow label="Concept" value={characterProfile.concept} />
            <SummaryRow label="Background" value={characterProfile.background} />
            <SummaryRow label="Personality" value={characterProfile.personality} />
            <SummaryList label="Motivations" values={characterProfile.motivations} />
            <SummaryList label="Strengths" values={characterProfile.strengths} />
            <SummaryList label="Weaknesses" values={characterProfile.weaknesses} />
          </dl>
        </section>
      </div>

      {errorMessage ? (
        <div className="border border-ruby/40 bg-ruby/10 px-4 py-3 text-sm text-ruby">
          {errorMessage}
        </div>
      ) : null}

      <div className="flex flex-col gap-3 border-t-2 border-gold/20 pt-6 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-pewter">Beginning the adventure should kick off world-building and route straight into the opening scene.</p>
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <button
            type="button"
            onClick={onBack}
            disabled={isBeginning}
            className="inline-flex items-center justify-center border border-gold/30 px-4 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition hover:border-gold hover:text-gold focus:outline-none focus:ring-2 focus:ring-gold focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:border-gold/10 disabled:text-pewter"
          >
            {backLabel}
          </button>
          <button
            type="button"
            onClick={onBegin}
            disabled={isBeginning}
            className="inline-flex items-center justify-center bg-ruby px-5 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition hover:bg-ruby-light focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter"
          >
            {isBeginning ? 'Starting adventure…' : beginLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
