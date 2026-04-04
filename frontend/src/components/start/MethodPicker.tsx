interface MethodOption<TMethod extends string = string> {
  readonly value: TMethod;
  readonly title: string;
  readonly description: string;
  readonly eyebrow?: string;
}

interface MethodPickerProps<TMethod extends string = string> {
  readonly title: string;
  readonly description: string;
  readonly methods: readonly MethodOption<TMethod>[];
  readonly selectedMethod: TMethod | null;
  readonly onSelectMethod: (method: TMethod) => void;
  readonly onContinue: (method: TMethod) => void | Promise<void>;
  readonly onBack?: () => void;
  readonly continueLabel?: string;
  readonly continueLoadingLabel?: string;
  readonly backLabel?: string;
  readonly helperText?: string;
  readonly errorMessage?: string | null;
  readonly isLoading?: boolean;
}

export type { MethodOption, MethodPickerProps };

export function MethodPicker<TMethod extends string = string>({
  title,
  description,
  methods,
  selectedMethod,
  onSelectMethod,
  onContinue,
  onBack,
  continueLabel = 'Continue',
  continueLoadingLabel = 'Continuing…',
  backLabel = 'Back',
  helperText,
  errorMessage = null,
  isLoading = false,
}: MethodPickerProps<TMethod>) {
  const hasSelection = selectedMethod !== null;

  function handleContinue() {
    if (!selectedMethod || isLoading) {
      return;
    }

    void onContinue(selectedMethod);
  }

  return (
    <section className="deco-corners deco-pattern space-y-6 border-2 border-gold/20 bg-charcoal p-6">
      <header className="space-y-3 border-b-2 border-gold/20 pb-5">
        <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Start game</p>
        <div className="space-y-2">
          <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">{title}</h2>
          <p className="max-w-2xl text-sm leading-7 text-champagne/70">{description}</p>
        </div>
      </header>

      <div className="grid gap-4 md:grid-cols-2">
        {methods.map((method) => {
          const isSelected = selectedMethod === method.value;

          return (
            <button
              key={method.value}
              type="button"
              onClick={() => {
                onSelectMethod(method.value);
              }}
              aria-pressed={isSelected}
              className={[
                'border-2 px-5 py-5 text-left transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gold/60',
                isSelected
                  ? 'border-gold bg-gold/10'
                  : 'border-gold/20 bg-charcoal hover:border-gold/40 hover:bg-charcoal/80 hover:-translate-y-0.5',
              ].join(' ')}
            >
              {method.eyebrow ? (
                <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-pewter/80">{method.eyebrow}</p>
              ) : null}
              <div className="mt-2 flex items-start justify-between gap-4">
                <div className="space-y-2">
                  <h3 className="font-heading text-lg font-semibold uppercase tracking-wide text-champagne">{method.title}</h3>
                  <p className="text-sm leading-6 text-champagne/70">{method.description}</p>
                </div>
                <span
                  className={[
                    'mt-1 inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border text-xs font-semibold',
                    isSelected
                      ? 'border-gold bg-gold text-obsidian'
                      : 'border-gold/30 bg-obsidian text-pewter',
                  ].join(' ')}
                  aria-hidden="true"
                >
                  {isSelected ? '✓' : ''}
                </span>
              </div>
            </button>
          );
        })}
      </div>

      {helperText ? <p className="text-sm leading-6 text-pewter">{helperText}</p> : null}

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
          disabled={!hasSelection || isLoading}
          className="inline-flex items-center justify-center bg-ruby px-5 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition-all duration-200 hover:bg-ruby-light hover:scale-[1.02] focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter disabled:hover:scale-100"
        >
          {isLoading ? continueLoadingLabel : continueLabel}
        </button>
      </div>
    </section>
  );
}
