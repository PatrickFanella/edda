import { useCallback, useEffect, useRef, useState } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getSessionHistory } from '../api/campaigns';
import type { SessionLogEntry } from '../api/types';

export interface UseReplayEngineResult {
  currentTurnIndex: number;
  totalTurns: number;
  isPlaying: boolean;
  playbackSpeed: number;
  currentEntry: SessionLogEntry | null;
  visibleEntries: SessionLogEntry[];
  play: () => void;
  pause: () => void;
  setSpeed: (speed: number) => void;
  seekTo: (turnIndex: number) => void;
  nextTurn: () => void;
  prevTurn: () => void;
}

export function useReplayEngine(campaignId: string): UseReplayEngineResult {
  const { data } = useQuery({
    queryKey: ['session-history', campaignId],
    queryFn: () => getSessionHistory(campaignId),
    enabled: campaignId.length > 0,
  });

  const entries = data?.entries ?? [];
  const totalTurns = entries.length;

  const [currentTurnIndex, setCurrentTurnIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackSpeed, setPlaybackSpeed] = useState(1);

  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const clearTimer = useCallback(() => {
    if (timerRef.current !== null) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
  }, []);

  // Auto-advance timer
  useEffect(() => {
    clearTimer();

    if (!isPlaying || totalTurns === 0) return;

    const interval = 3000 / playbackSpeed;

    timerRef.current = setInterval(() => {
      setCurrentTurnIndex((prev) => {
        const next = prev + 1;
        if (next >= totalTurns) {
          setIsPlaying(false);
          return prev;
        }
        return next;
      });
    }, interval);

    return clearTimer;
  }, [isPlaying, playbackSpeed, totalTurns, clearTimer]);

  const play = useCallback(() => {
    if (totalTurns === 0) return;
    // If at the end, restart from beginning
    if (currentTurnIndex >= totalTurns - 1) {
      setCurrentTurnIndex(0);
    }
    setIsPlaying(true);
  }, [currentTurnIndex, totalTurns]);

  const pause = useCallback(() => {
    setIsPlaying(false);
  }, []);

  const setSpeed = useCallback((speed: number) => {
    setPlaybackSpeed(speed);
  }, []);

  const seekTo = useCallback(
    (turnIndex: number) => {
      const clamped = Math.max(0, Math.min(turnIndex, totalTurns - 1));
      setCurrentTurnIndex(clamped);
    },
    [totalTurns],
  );

  const nextTurn = useCallback(() => {
    setCurrentTurnIndex((prev) => Math.min(prev + 1, totalTurns - 1));
  }, [totalTurns]);

  const prevTurn = useCallback(() => {
    setCurrentTurnIndex((prev) => Math.max(prev - 1, 0));
  }, []);

  const currentEntry = totalTurns > 0 ? entries[currentTurnIndex] ?? null : null;
  const visibleEntries = entries.slice(0, currentTurnIndex + 1);

  return {
    currentTurnIndex,
    totalTurns,
    isPlaying,
    playbackSpeed,
    currentEntry,
    visibleEntries,
    play,
    pause,
    setSpeed,
    seekTo,
    nextTurn,
    prevTurn,
  };
}
