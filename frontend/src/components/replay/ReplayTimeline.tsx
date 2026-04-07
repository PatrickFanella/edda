import { useCallback, useState } from 'react';

interface ReplayTimelineProps {
  readonly currentTurnIndex: number;
  readonly totalTurns: number;
  readonly seekTo: (turnIndex: number) => void;
}

export function ReplayTimeline({ currentTurnIndex, totalTurns, seekTo }: ReplayTimelineProps) {
  const [hoveredTurn, setHoveredTurn] = useState<number | null>(null);
  const [hoverX, setHoverX] = useState(0);

  const handleClick = useCallback(
    (event: React.MouseEvent<HTMLDivElement>) => {
      if (totalTurns === 0) return;
      const rect = event.currentTarget.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const ratio = x / rect.width;
      const turnIndex = Math.round(ratio * (totalTurns - 1));
      seekTo(turnIndex);
    },
    [totalTurns, seekTo],
  );

  const handleMouseMove = useCallback(
    (event: React.MouseEvent<HTMLDivElement>) => {
      if (totalTurns === 0) return;
      const rect = event.currentTarget.getBoundingClientRect();
      const x = event.clientX - rect.left;
      const ratio = x / rect.width;
      const turnIndex = Math.round(ratio * (totalTurns - 1));
      setHoveredTurn(turnIndex);
      setHoverX(x);
    },
    [totalTurns],
  );

  const handleMouseLeave = useCallback(() => {
    setHoveredTurn(null);
  }, []);

  const progressPercent = totalTurns > 1 ? (currentTurnIndex / (totalTurns - 1)) * 100 : 0;

  return (
    <div className="border-2 border-gold/20 bg-charcoal px-5 py-4">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter">Timeline</span>
        <span className="text-[11px] uppercase tracking-[0.2em] text-pewter">
          {totalTurns} {totalTurns === 1 ? 'turn' : 'turns'}
        </span>
      </div>

      <div
        className="relative h-8 cursor-pointer select-none"
        onClick={handleClick}
        onMouseMove={handleMouseMove}
        onMouseLeave={handleMouseLeave}
        role="slider"
        aria-valuenow={currentTurnIndex}
        aria-valuemin={0}
        aria-valuemax={totalTurns - 1}
        aria-label="Replay timeline"
        tabIndex={0}
      >
        {/* Track background */}
        <div className="absolute inset-x-0 top-1/2 h-1 -translate-y-1/2 bg-obsidian/60 border border-gold/10" />

        {/* Filled track */}
        <div
          className="absolute top-1/2 left-0 h-1 -translate-y-1/2 bg-gold/40"
          style={{ width: `${progressPercent}%` }}
        />

        {/* Turn marker dots */}
        {totalTurns > 0 &&
          totalTurns <= 100 &&
          Array.from({ length: totalTurns }, (_, i) => {
            const left = totalTurns > 1 ? (i / (totalTurns - 1)) * 100 : 50;
            return (
              <div
                key={i}
                className="absolute top-1/2 h-2 w-2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-pewter/40"
                style={{ left: `${left}%` }}
              />
            );
          })}

        {/* Current position marker */}
        {totalTurns > 0 && (
          <div
            className="absolute top-1/2 h-4 w-4 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-gold bg-gold/80 shadow-[0_0_8px_rgba(212,175,55,0.4)] transition-[left] duration-150"
            style={{ left: `${progressPercent}%` }}
          />
        )}

        {/* Hover tooltip */}
        {hoveredTurn !== null && (
          <div
            className="pointer-events-none absolute -top-8 -translate-x-1/2 rounded-sm border border-gold/30 bg-obsidian px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-gold"
            style={{ left: `${hoverX}px` }}
          >
            Turn {hoveredTurn + 1}
          </div>
        )}
      </div>
    </div>
  );
}
