import { useState, type FormEvent, type KeyboardEvent } from 'react';

import { cn } from '../../lib/cn';

interface PlayerInputProps {
  readonly onSendAction: (input: string) => boolean | Promise<boolean>;
  readonly disabled?: boolean;
  readonly isLoading?: boolean;
  readonly placeholder?: string;
  readonly className?: string;
  readonly autoFocus?: boolean;
}

export function PlayerInput({
  onSendAction,
  disabled = false,
  isLoading = false,
  placeholder = 'Describe what you do next…',
  className,
  autoFocus = false,
}: PlayerInputProps) {
  const [value, setValue] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isDisabled = disabled || isLoading || isSubmitting;

  async function handleSubmit(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();
    if (isDisabled || value.trim().length === 0) {
      return;
    }

    setIsSubmitting(true);
    try {
      const accepted = await Promise.resolve(onSendAction(value));
      if (accepted) {
        setValue('');
      }
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
    <form
      className={cn(
        'border-2 border-gold/20 bg-charcoal p-4 transition-all duration-200',
        className,
      )}
      onSubmit={(event) => {
        void handleSubmit(event);
      }}
    >
      <label className="flex flex-col gap-3">
        <span className="text-xs font-semibold uppercase tracking-[0.2em] text-pewter">Your move</span>
        <textarea
          value={value}
          onChange={(event) => {
            setValue(event.target.value);
          }}
          onKeyDown={handleKeyDown}
          rows={3}
          autoFocus={autoFocus}
          disabled={isDisabled}
          placeholder={placeholder}
          className="min-h-28 w-full resize-none border-2 border-gold/20 bg-obsidian px-4 py-3 text-sm leading-6 text-champagne transition-all duration-200 placeholder:text-pewter/60 focus:border-gold focus:outline-none disabled:cursor-not-allowed disabled:border-gold/10 disabled:bg-charcoal disabled:text-pewter"
        />
      </label>
      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-pewter">Press Enter to act. Use Shift+Enter for a line break.</p>
        <button
          type="submit"
          disabled={isDisabled}
          className="inline-flex items-center justify-center bg-ruby px-5 py-3 text-sm font-semibold uppercase tracking-[0.15em] text-champagne transition-all duration-200 hover:bg-ruby-light hover:scale-[1.02] focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter disabled:hover:scale-100"
        >
          {isLoading || isSubmitting ? 'Sending…' : 'Send action'}
        </button>
      </div>
    </form>
  );
}
