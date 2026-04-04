export interface CampaignProfilePreview {
  readonly genre: string;
  readonly tone: string;
  readonly themes: readonly string[];
  readonly world_type: string;
  readonly danger_level: string;
  readonly political_complexity: string;
}

export interface CampaignProposalOption {
  readonly name: string;
  readonly summary: string;
  readonly profile: CampaignProfilePreview;
}

interface ProposalPickerProps {
  readonly proposals: readonly CampaignProposalOption[];
  readonly selectedProposalKey: string | null;
  readonly onSelectProposal: (proposal: CampaignProposalOption, proposalKey: string) => void;
  readonly onContinue: (proposal: CampaignProposalOption) => void | Promise<void>;
  readonly onBack?: () => void;
  readonly getProposalKey?: (proposal: CampaignProposalOption, index: number) => string;
  readonly title?: string;
  readonly description?: string;
  readonly continueLabel?: string;
  readonly continueLoadingLabel?: string;
  readonly backLabel?: string;
  readonly emptyStateTitle?: string;
  readonly emptyStateMessage?: string;
  readonly errorMessage?: string | null;
  readonly isLoading?: boolean;
}

export type { ProposalPickerProps };

export function ProposalPicker({
  proposals,
  selectedProposalKey,
  onSelectProposal,
  onContinue,
  onBack,
  getProposalKey = defaultProposalKey,
  title = 'Choose your campaign',
  description = 'Review the generated campaign proposals and pick the one you want to carry into character creation.',
  continueLabel = 'Continue with proposal',
  continueLoadingLabel = 'Continuing…',
  backLabel = 'Back',
  emptyStateTitle = 'No proposals ready',
  emptyStateMessage = 'We could not generate campaign options from the selected attributes. Go back and adjust the frame before trying again.',
  errorMessage = null,
  isLoading = false,
}: ProposalPickerProps) {
  const selectedProposal = proposals.find((proposal, index) => getProposalKey(proposal, index) === selectedProposalKey) ?? null;

  function handleContinue() {
    if (!selectedProposal || isLoading) {
      return;
    }

    void onContinue(selectedProposal);
  }

  if (proposals.length === 0) {
    return (
      <section className="deco-corners deco-pattern space-y-6 border-2 border-gold/20 bg-charcoal p-6">
        <header className="space-y-3 border-b-2 border-gold/20 pb-5">
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Campaign proposals</p>
          <div className="space-y-2">
            <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">{emptyStateTitle}</h2>
            <p className="max-w-2xl text-sm leading-7 text-champagne/70">{emptyStateMessage}</p>
          </div>
        </header>

        {errorMessage ? (
          <div className="border border-ruby/40 bg-ruby/10 px-4 py-3 text-sm text-ruby">{errorMessage}</div>
        ) : null}

        {onBack ? (
          <div className="flex justify-start border-t border-gold/20 pt-5">
            <button
              type="button"
              onClick={onBack}
              className="inline-flex items-center justify-center border border-gold/30 px-4 py-3 text-sm font-semibold uppercase tracking-wide text-champagne/80 transition hover:border-gold hover:text-champagne focus:outline-none focus:ring-2 focus:ring-gold/60"
            >
              {backLabel}
            </button>
          </div>
        ) : null}
      </section>
    );
  }

  return (
    <section className="deco-corners deco-pattern space-y-6 border-2 border-gold/20 bg-charcoal p-6">
      <header className="space-y-3 border-b-2 border-gold/20 pb-5">
        <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Campaign proposals</p>
        <div className="space-y-2">
          <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">{title}</h2>
          <p className="max-w-2xl text-sm leading-7 text-champagne/70">{description}</p>
        </div>
      </header>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.5fr)_minmax(18rem,1fr)]">
        <div className="space-y-4">
          {proposals.map((proposal, index) => {
            const proposalKey = getProposalKey(proposal, index);
            const isSelected = proposalKey === selectedProposalKey;

            return (
              <button
                key={proposalKey}
                type="button"
                onClick={() => {
                  onSelectProposal(proposal, proposalKey);
                }}
                aria-pressed={isSelected}
                className={[
                  'w-full border-2 px-5 py-5 text-left transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gold/60',
                  isSelected
                    ? 'border-gold bg-gold/10'
                    : 'border-gold/20 bg-charcoal hover:border-gold/40 hover:bg-charcoal/80 hover:-translate-y-0.5',
                ].join(' ')}
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="space-y-3">
                    <div>
                      <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">Proposal {index + 1}</p>
                      <h3 className="font-heading mt-2 text-xl font-semibold uppercase tracking-wide text-champagne">{proposal.name}</h3>
                    </div>
                    <p className="text-sm leading-7 text-champagne/70">{proposal.summary}</p>
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

        <aside className="border-2 border-gold/20 bg-charcoal p-5">
          {selectedProposal ? (
            <SelectedProposalPreview proposal={selectedProposal} />
          ) : (
            <div className="flex h-full min-h-64 flex-col items-center justify-center border border-dashed border-gold/15 bg-obsidian px-6 text-center">
              <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-pewter/80">Awaiting selection</p>
              <p className="mt-3 max-w-sm text-sm leading-7 text-pewter">
                Select a proposal to inspect its genre, world shape, and campaign pressures before continuing.
              </p>
            </div>
          )}
        </aside>
      </div>

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
          disabled={!selectedProposal || isLoading}
          className="inline-flex items-center justify-center bg-ruby px-5 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition hover:bg-ruby-light focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter"
        >
          {isLoading ? continueLoadingLabel : continueLabel}
        </button>
      </div>
    </section>
  );
}

function SelectedProposalPreview({ proposal }: { readonly proposal: CampaignProposalOption }) {
  const metadata = [
    { label: 'Genre', value: proposal.profile.genre || 'Unspecified' },
    { label: 'Tone', value: proposal.profile.tone || 'Unspecified' },
    { label: 'World type', value: proposal.profile.world_type || 'Unspecified' },
    { label: 'Danger', value: proposal.profile.danger_level || 'Unspecified' },
    {
      label: 'Politics',
      value: proposal.profile.political_complexity || 'Unspecified',
    },
    {
      label: 'Themes',
      value: proposal.profile.themes.length > 0 ? proposal.profile.themes.join(', ') : 'Unspecified',
    },
  ];

  return (
    <div className="space-y-4">
      <div>
        <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Selected proposal</p>
        <h3 className="font-heading mt-2 text-xl font-semibold uppercase tracking-wide text-champagne">{proposal.name}</h3>
        <p className="mt-3 text-sm leading-7 text-champagne/70">{proposal.summary}</p>
      </div>
      <dl className="space-y-3 text-sm text-champagne/70">
        {metadata.map((item) => (
          <div key={item.label} className="border border-gold/15 bg-obsidian px-4 py-3">
            <dt className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{item.label}</dt>
            <dd className="mt-2 leading-6 text-champagne">{item.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
}

function defaultProposalKey(proposal: CampaignProposalOption, index: number): string {
  return `${proposal.name}-${index}`;
}
