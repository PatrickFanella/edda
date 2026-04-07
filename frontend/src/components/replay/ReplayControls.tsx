import { useState } from 'react';

import { cn } from '../../lib/cn';

const SPEEDS = [1, 2, 4];

interface ReplayControlsProps {
  readonly isPlaying: boolean;
  readonly playbackSpeed: number;
  readonly currentTurnIndex: number;
  readonly totalTurns: number;
  readonly play: () => void;
  readonly pause: () => void;
  readonly setSpeed: (speed: number) => void;
  readonly nextTurn: () => void;
  readonly prevTurn: () => void;
}

export function ReplayControls({
  isPlaying,
  playbackSpeed,
  currentTurnIndex,
  totalTurns,
  play,
  pause,
  setSpeed,
  nextTurn,
  prevTurn,
}: ReplayControlsProps) {
  const [copied, setCopied] = useState(false);

  function handleCopyLink() {
    try {
      void navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API may not be available
    }
  }

  return (
    <div className="flex flex-wrap items-center justify-between gap-4 border-2 border-gold/20 bg-charcoal px-5 py-4">
      {/* Transport controls */}
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={prevTurn}
          disabled={currentTurnIndex <= 0}
          className="inline-flex items-center justify-center border border-gold/20 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-champagne transition-all duration-200 hover:border-gold hover:text-gold disabled:opacity-30 disabled:cursor-not-allowed"
          aria-label="Previous turn"
        >
          Prev
        </button>

        <button
          type="button"
          onClick={isPlaying ? pause : play}
          disabled={totalTurns === 0}
          className="inline-flex items-center justify-center border-2 border-ruby/40 bg-ruby/10 px-5 py-2 text-sm font-semibold uppercase tracking-wide text-ruby transition-all duration-200 hover:border-ruby hover:bg-ruby/20 disabled:opacity-30 disabled:cursor-not-allowed"
          aria-label={isPlaying ? 'Pause' : 'Play'}
        >
          {isPlaying ? 'Pause' : 'Play'}
        </button>

        <button
          type="button"
          onClick={nextTurn}
          disabled={currentTurnIndex >= totalTurns - 1}
          className="inline-flex items-center justify-center border border-gold/20 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-champagne transition-all duration-200 hover:border-gold hover:text-gold disabled:opacity-30 disabled:cursor-not-allowed"
          aria-label="Next turn"
        >
          Next
        </button>
      </div>

      {/* Speed selector */}
      <div className="flex items-center gap-2">
        <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter">Speed</span>
        {SPEEDS.map((speed) => (
          <button
            key={speed}
            type="button"
            onClick={() => setSpeed(speed)}
            className={cn(
              'inline-flex items-center justify-center border px-3 py-1.5 text-xs font-semibold uppercase tracking-wide transition-all duration-200',
              speed === playbackSpeed
                ? 'border-gold bg-gold/15 text-gold'
                : 'border-gold/20 text-champagne/60 hover:border-gold/40 hover:text-champagne',
            )}
          >
            {speed}x
          </button>
        ))}
      </div>

      {/* Turn counter */}
      <p className="text-sm font-medium tracking-wide text-champagne/70">
        <span className="text-gold">{totalTurns > 0 ? currentTurnIndex + 1 : 0}</span>
        <span className="mx-1 text-pewter">/</span>
        <span>{totalTurns}</span>
        <span className="ml-2 text-[11px] uppercase tracking-[0.2em] text-pewter">turns</span>
      </p>

      {/* Copy link button */}
      <button
        type="button"
        onClick={handleCopyLink}
        className="inline-flex items-center justify-center border border-sapphire/30 px-3 py-2 text-xs font-semibold uppercase tracking-wide text-sapphire transition-all duration-200 hover:border-sapphire hover:text-sapphire"
      >
        {copied ? 'Copied!' : 'Copy Link'}
      </button>
    </div>
  );
}
