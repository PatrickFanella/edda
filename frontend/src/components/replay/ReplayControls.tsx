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
    <div className="flex flex-wrap items-center justify-between gap-4">
      {/* Transport controls */}
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={prevTurn}
          disabled={currentTurnIndex <= 0}
          className="hud-btn hud-icon-btn border border-pewter/30 text-pewter transition-colors hover:border-sapphire hover:text-sapphire disabled:cursor-not-allowed disabled:opacity-30"
          aria-label="Previous turn"
        >
          <span aria-hidden="true">‹</span>
        </button>

        <button
          type="button"
          onClick={isPlaying ? pause : play}
          disabled={totalTurns === 0}
          className="hud-btn hud-text-button border border-sapphire/40 bg-sapphire/10 px-5 text-sm font-semibold uppercase tracking-wide text-sapphire transition-colors hover:border-sapphire hover:bg-sapphire/15 disabled:cursor-not-allowed disabled:opacity-30"
          aria-label={isPlaying ? 'Pause' : 'Play'}
        >
          {isPlaying ? 'Pause' : 'Play'}
        </button>

        <button
          type="button"
          onClick={nextTurn}
          disabled={currentTurnIndex >= totalTurns - 1}
          className="hud-btn hud-icon-btn border border-pewter/30 text-pewter transition-colors hover:border-sapphire hover:text-sapphire disabled:cursor-not-allowed disabled:opacity-30"
          aria-label="Next turn"
        >
          <span aria-hidden="true">›</span>
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
              'hud-tab-button px-3 text-xs font-semibold uppercase tracking-wide transition-colors duration-200',
              speed === playbackSpeed
                ? 'border-sapphire bg-sapphire/15 text-sapphire'
                : 'border-pewter/20 text-pewter/70 hover:border-sapphire/40 hover:text-sapphire',
            )}
          >
            {speed}x
          </button>
        ))}
      </div>

      {/* Turn counter */}
      <p className="text-sm font-medium tracking-wide text-champagne/70">
        <span className="text-sapphire">{totalTurns > 0 ? currentTurnIndex + 1 : 0}</span>
        <span className="mx-1 text-pewter">/</span>
        <span>{totalTurns}</span>
        <span className="ml-2 text-[11px] uppercase tracking-[0.2em] text-pewter">turns</span>
      </p>

      {/* Copy link button */}
      <button
        type="button"
        onClick={handleCopyLink}
        className="hud-btn hud-text-button border border-sapphire/30 px-3 text-xs font-semibold uppercase tracking-wide text-sapphire transition-colors hover:border-sapphire hover:text-sapphire"
      >
        {copied ? 'Copied!' : 'Copy Link'}
      </button>
    </div>
  );
}
