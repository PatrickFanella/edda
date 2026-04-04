import { useEffect, useRef, useState, type FormEvent, type KeyboardEvent } from 'react';

import { cn } from '../../lib/cn';

export type ChatRole = 'assistant' | 'user' | 'system';

export interface ChatTranscriptEntry {
  readonly id: string;
  readonly role: ChatRole;
  readonly content: string;
}

interface ChatStepProps {
  readonly title: string;
  readonly description: string;
  readonly transcript: readonly ChatTranscriptEntry[];
  readonly onSubmitMessage: (message: string) => void | Promise<void>;
  readonly onBack?: () => void;
  readonly submitLabel?: string;
  readonly submitLoadingLabel?: string;
  readonly backLabel?: string;
  readonly inputLabel?: string;
  readonly inputPlaceholder?: string;
  readonly emptyStateTitle?: string;
  readonly emptyStateMessage?: string;
  readonly helperText?: string;
  readonly errorMessage?: string | null;
  readonly isLoading?: boolean;
  readonly autoFocus?: boolean;
}

export type { ChatStepProps };

export function ChatStep({
  title,
  description,
  transcript,
  onSubmitMessage,
  onBack,
  submitLabel = 'Send reply',
  submitLoadingLabel = 'Sending…',
  backLabel = 'Back',
  inputLabel = 'Your reply',
  inputPlaceholder = 'Describe the kind of world you want to play in…',
  emptyStateTitle = 'Interview not started',
  emptyStateMessage = 'The conversation will appear here once the interviewer opens with the first prompt.',
  helperText = 'Press Enter to send. Use Shift+Enter for a line break.',
  errorMessage = null,
  isLoading = false,
  autoFocus = false,
}: ChatStepProps) {
  const [draft, setDraft] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const endRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [isLoading, transcript]);

  const isDisabled = isLoading || isSubmitting;

  async function handleSubmit(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();

    const message = draft.trim();
    if (!message || isDisabled) {
      return;
    }

    setIsSubmitting(true);
    try {
      await Promise.resolve(onSubmitMessage(message));
      setDraft('');
    } finally {
      setIsSubmitting(false);
    }
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      void handleSubmit();
    }
  }

  return (
    <section className="grid gap-6 xl:grid-cols-[minmax(0,1.8fr)_minmax(18rem,1fr)]">
      <div className="border-2 border-gold/20 bg-charcoal">
        <header className="space-y-3 border-b-2 border-gold/20 px-6 py-5">
          <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold">Campaign interview</p>
          <div className="space-y-2">
            <h2 className="font-heading text-2xl font-semibold uppercase tracking-wide text-champagne">{title}</h2>
            <p className="max-w-2xl text-sm leading-7 text-champagne/70">{description}</p>
          </div>
        </header>

        <div className="flex max-h-[34rem] min-h-80 flex-col gap-4 overflow-y-auto px-4 py-4 sm:px-5" role="log" aria-live="polite" aria-busy={isLoading ? 'true' : 'false'}>
          {transcript.length > 0 ? (
            transcript.map((entry) => <TranscriptBubble key={entry.id} entry={entry} />)
          ) : (
            <div className="flex min-h-64 flex-1 flex-col items-center justify-center border border-dashed border-gold/15 bg-charcoal/50 px-6 text-center">
              <p className="font-heading text-sm font-semibold uppercase tracking-[0.2em] text-pewter/80">{emptyStateTitle}</p>
              <p className="mt-3 max-w-md text-sm leading-7 text-pewter">{emptyStateMessage}</p>
            </div>
          )}

          {isLoading ? (
            <div className="flex justify-start">
              <div className="max-w-2xl rounded-bl-md border border-gold/20 bg-gold/10 px-4 py-3 text-sm text-champagne">
                <p className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-gold/80">Guide</p>
                <p className="mt-2 leading-6 text-champagne">Thinking through the next question…</p>
              </div>
            </div>
          ) : null}
          <div ref={endRef} aria-hidden="true" />
        </div>
      </div>

      <aside className="space-y-4">
        <form
          className="border-2 border-gold/20 bg-charcoal p-5"
          onSubmit={(event) => {
            void handleSubmit(event);
          }}
        >
          <label className="flex flex-col gap-3">
            <span className="font-heading text-xs font-semibold uppercase tracking-[0.2em] text-pewter">{inputLabel}</span>
            <textarea
              value={draft}
              onChange={(event) => {
                setDraft(event.target.value);
              }}
              onKeyDown={handleKeyDown}
              rows={7}
              autoFocus={autoFocus}
              disabled={isDisabled}
              placeholder={inputPlaceholder}
              className="min-h-44 w-full resize-none border-2 border-gold/20 bg-obsidian px-4 py-3 text-sm leading-6 text-champagne transition-all duration-200 placeholder:text-pewter/60 focus:border-gold focus:outline-none focus:ring-2 focus:ring-gold/40 disabled:cursor-not-allowed disabled:border-gold/10 disabled:bg-charcoal disabled:text-pewter"
            />
          </label>

          <div className="mt-4 space-y-3">
            <p className="text-sm leading-6 text-pewter">{helperText}</p>
            <div className="flex flex-col-reverse gap-3 sm:flex-row sm:items-center sm:justify-between">
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
                type="submit"
                disabled={isDisabled || draft.trim().length === 0}
                className="inline-flex items-center justify-center bg-ruby px-5 py-3 text-sm font-semibold uppercase tracking-wide text-champagne transition hover:bg-ruby-light focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter"
              >
                {isDisabled ? submitLoadingLabel : submitLabel}
              </button>
            </div>
          </div>
        </form>

        {errorMessage ? (
          <div className="border border-ruby/40 bg-ruby/10 px-4 py-3 text-sm text-ruby">{errorMessage}</div>
        ) : null}

        <section className="border-2 border-gold/20 bg-charcoal p-5">
          <h3 className="font-heading text-lg font-semibold uppercase tracking-wide text-champagne">Interview flow</h3>
          <ul className="mt-4 space-y-3 text-sm leading-6 text-champagne/70">
            <li>The guide can start the conversation before you type anything.</li>
            <li>Short answers work. The interview will keep drilling into missing details.</li>
            <li>Use back if you want to switch creation methods without losing the surrounding shell.</li>
          </ul>
        </section>
      </aside>
    </section>
  );
}

function TranscriptBubble({ entry }: { readonly entry: ChatTranscriptEntry }) {
  const isUser = entry.role === 'user';
  const isSystem = entry.role === 'system';

  return (
    <div className={cn('flex', isUser ? 'justify-end' : isSystem ? 'justify-center' : 'justify-start')}>
      <article
        className={cn(
          'max-w-2xl px-4 py-3 text-sm',
          isUser
            ? 'rounded-br-md border border-gold/20 bg-charcoal text-champagne'
            : isSystem
              ? 'border border-pewter/30 bg-pewter/10 text-champagne/90 shadow-none'
              : 'rounded-bl-md border border-gold/20 bg-gold/10 text-champagne',
        )}
      >
        <p
          className={cn(
            'text-xs font-semibold uppercase tracking-[0.24em]',
            isUser ? 'text-pewter' : isSystem ? 'text-pewter/80' : 'text-gold/80',
          )}
        >
          {isUser ? 'You' : isSystem ? 'System' : 'Guide'}
        </p>
        <p className="mt-2 whitespace-pre-wrap leading-6">{entry.content}</p>
      </article>
    </div>
  );
}
