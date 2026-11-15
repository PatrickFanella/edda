import { useId, useRef, useState, type FormEvent, type KeyboardEvent } from 'react';

import { flushSync } from 'react-dom';

import { matchSuggestions, type ActionSuggestion } from '../../api/suggestions';
import { cn } from '../../lib/cn';

interface PlayerInputProps {
  readonly suggestions?: readonly ActionSuggestion[];
  readonly onSendAction: (input: string) => boolean | Promise<boolean>;
  readonly disabled?: boolean;
  readonly isLoading?: boolean;
  readonly placeholder?: string;
  readonly className?: string;
  readonly autoFocus?: boolean;
  readonly preparesAction?: boolean;
}

export function PlayerInput({
  onSendAction,
  suggestions = [],
  disabled = false,
  isLoading = false,
  placeholder = 'Describe what you do next…',
  className,
  autoFocus = false,
  preparesAction = false,
}: PlayerInputProps) {
  const [value, setValue] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [focused, setFocused] = useState(false);
  const [atEnd, setAtEnd] = useState(true);
  const [composing, setComposing] = useState(false);
  const [dismissed, setDismissed] = useState(false);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const ghostRef = useRef<HTMLSpanElement>(null);
  const hintId = useId();

  const isDisabled = disabled || isLoading || isSubmitting;
  const matches = matchSuggestions(suggestions, value);
  // Word matches can remain useful as buttons, but only a literal continuation
  // can be ghost text. Never replace arbitrary prose on Tab.
  const completion = focused && atEnd && !composing && !dismissed && !isDisabled
    ? matches.find(candidate => candidate.text.length > value.length && candidate.text.toLocaleLowerCase().startsWith(value.toLocaleLowerCase()))
    : undefined;

  function insertSuggestion(text: string) {
    flushSync(() => { setValue(text); setDismissed(true); });
    const input = inputRef.current;
    input?.focus({ preventScroll: true });
    input?.setSelectionRange(text.length, text.length);
  }

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
    if (event.nativeEvent.isComposing || event.nativeEvent.keyCode === 229 || composing) return;
    if (event.key === 'Escape') {
      setDismissed(true);
    }
    if (event.key === 'Tab' && !event.shiftKey && !event.ctrlKey && !event.altKey && !event.metaKey && completion) {
      event.preventDefault();
      insertSuggestion(completion.text);
      return;
    }
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
        <span className="relative block bg-obsidian">
        {completion && <span ref={ghostRef} aria-hidden="true" data-testid="action-completion" className="action-completion pointer-events-none absolute inset-0 overflow-hidden whitespace-pre-wrap break-words border-2 border-transparent px-4 py-3 text-sm leading-6 text-pewter/60">
          <span className="invisible">{value}</span>{completion.text.slice(value.length)}
        </span>}
        <textarea
          ref={inputRef}
          aria-describedby={completion ? hintId : undefined}
          onFocus={() => setFocused(true)}
          onBlur={() => setFocused(false)}
          onSelect={event => setAtEnd(event.currentTarget.selectionStart === value.length && event.currentTarget.selectionEnd === value.length)}
          onCompositionStart={() => setComposing(true)}
          onCompositionEnd={() => setComposing(false)}
          onScroll={event => { if (ghostRef.current) ghostRef.current.scrollTop = event.currentTarget.scrollTop; }}
          value={value}
          onChange={(event) => {
            setValue(event.target.value);
            setDismissed(false);
            setAtEnd(event.target.selectionStart === event.target.value.length && event.target.selectionEnd === event.target.value.length);
          }}
          onKeyDown={handleKeyDown}
          rows={2}
          autoFocus={autoFocus}
          disabled={isDisabled}
          placeholder={completion ? '' : placeholder}
          className="relative block bg-transparent min-h-16 max-h-28 overflow-y-auto w-full resize-none border-2 border-gold/20 px-4 py-3 text-sm leading-6 text-champagne transition-all duration-200 placeholder:text-pewter/60 focus:border-gold focus:outline-none disabled:cursor-not-allowed disabled:border-gold/10 disabled:bg-charcoal disabled:text-pewter"
        />
        </span>
      </label>
      {completion && <p id={hintId} className="sr-only">Completion: {completion.text}. Press Tab to insert or Escape to dismiss. This does not send an action.</p>}
      {suggestions.length > 0 && !isDisabled && <div aria-label="Action suggestions" className="mt-2 max-h-24 overflow-y-auto">
        {matches.map(candidate => <button key={candidate.id} type="button" className="block min-h-11 w-full py-2 text-left text-sm text-pewter hover:text-gold" onClick={() => insertSuggestion(candidate.text)}>
          <span className="block">{candidate.text}</span><small className="block text-xs">{candidate.reason}</small>
        </button>)}
      </div>}
      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm text-pewter">Press Enter to {preparesAction ? 'prepare' : 'act'}. Use Shift+Enter for a line break.</p>
        <button
          type="submit"
          disabled={isDisabled || value.trim().length === 0}
          className="hud-text-button inline-flex items-center justify-center bg-ruby px-5 text-sm font-semibold uppercase tracking-[0.15em] text-champagne transition-all duration-200 hover:bg-ruby-light hover:scale-[1.02] focus:outline-none focus:ring-2 focus:ring-ruby focus:ring-offset-2 focus:ring-offset-obsidian disabled:cursor-not-allowed disabled:bg-charcoal disabled:text-pewter disabled:hover:scale-100"
        >
          {isLoading || isSubmitting ? (preparesAction ? 'Preparing…' : 'Sending…') : preparesAction ? 'Prepare action' : 'Send action'}
        </button>
      </div>
    </form>
  );
}
