import { useEffect, useRef, useState } from 'react';

import type { SessionLogEntry } from '../../api/types';
import { HudPanel } from '../layout/HudPanel';
import { cn } from '../../lib/cn';

interface ReplayNarrativeProps {
  readonly visibleEntries: SessionLogEntry[];
  readonly currentTurnIndex: number;
  readonly isPlaying: boolean;
}

const CHARS_PER_TICK = 30;
const TICK_INTERVAL_MS = 33; // ~30fps

export function ReplayNarrative({ visibleEntries, currentTurnIndex }: ReplayNarrativeProps) {
  const [typedText, setTypedText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const prevTurnRef = useRef(-1);

  // Typewriter effect for the latest entry's llm_response
  useEffect(() => {
    if (visibleEntries.length === 0) return;

    const latestEntry = visibleEntries[visibleEntries.length - 1];
    if (!latestEntry) return;

    // Only animate when the turn index changes
    if (prevTurnRef.current === currentTurnIndex) return;
    prevTurnRef.current = currentTurnIndex;

    const fullText = latestEntry.llm_response;
    if (!fullText) {
      setTypedText('');
      setIsTyping(false);
      return;
    }

    setIsTyping(true);
    setTypedText('');

    let charIndex = 0;
    const timer = setInterval(() => {
      charIndex += CHARS_PER_TICK;
      if (charIndex >= fullText.length) {
        setTypedText(fullText);
        setIsTyping(false);
        clearInterval(timer);
      } else {
        setTypedText(fullText.slice(0, charIndex));
      }
    }, TICK_INTERVAL_MS);

    return () => clearInterval(timer);
  }, [currentTurnIndex, visibleEntries]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [typedText, visibleEntries.length]);

  if (visibleEntries.length === 0) {
    return (
      <HudPanel accent="empty" title="Archive Log" bodyClassName="text-center">
        <p className="text-sm text-champagne/50">No turns to display. Press Play to begin the replay.</p>
      </HudPanel>
    );
  }

  return (
    <HudPanel accent="replay" title="Narrative Feed" bodyClassName="max-h-[600px] overflow-y-auto pr-1">
      <div ref={scrollRef} className="space-y-4">
      {visibleEntries.map((entry, index) => {
        const isLatest = index === visibleEntries.length - 1;
        const gmText = isLatest && isTyping ? typedText : entry.llm_response;

        return (
          <div key={`${entry.turn_number}-${index}`} className="space-y-3">
            {/* Player input */}
            {entry.player_input && (
              <div className="flex justify-end">
                <div className="max-w-[80%] border border-sapphire/25 bg-sapphire/5 px-4 py-3">
                  <div className="mb-1 flex items-center justify-end gap-2">
                    <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-sapphire">You</span>
                    <span className="hud-baseline-badge inline-flex rounded-sm border border-sapphire/20 bg-sapphire/10 px-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-sapphire">
                      player
                    </span>
                  </div>
                  <p className="text-right text-sm leading-7 text-champagne">{entry.player_input}</p>
                </div>
              </div>
            )}

            {/* GM response */}
            {gmText && (
              <div className="border border-pewter/25 bg-pewter/5 px-4 py-3">
                <div className="mb-1 flex items-center gap-2">
                  <span className="hud-baseline-badge inline-flex rounded-sm border border-pewter/20 bg-pewter/10 px-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-pewter">
                    gm
                  </span>
                  <span className="text-[11px] font-semibold uppercase tracking-[0.2em] text-pewter">Game Master</span>
                  <span className="ml-auto text-[10px] uppercase tracking-[0.2em] text-pewter">
                    Turn {entry.turn_number}
                  </span>
                </div>
                <p className={cn('whitespace-pre-wrap text-sm leading-7 text-champagne')}>
                  {gmText}
                  {isLatest && isTyping && (
                    <span className="ml-1 inline-block h-4 w-2 animate-pulse bg-sapphire/80 align-middle" />
                  )}
                </p>
              </div>
            )}
          </div>
        );
      })}
      </div>
    </HudPanel>
  );
}
