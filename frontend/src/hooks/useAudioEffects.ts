import { useEffect, useRef } from 'react';

import type { StateChange } from '../api/types';
import type { TurnResponseWithChoices } from './useWebSocket';
import { getAmbientTrack, getMusicTrack, getSfxTrack } from '../lib/audioAssets';
import type { UseAudioResult } from './useAudio';

/**
 * Connects the audio system to game events.
 *
 * Watches `latestResult` for state_changes and triggers the appropriate audio
 * layer transitions (ambient crossfade on location change, music swap on combat
 * state, SFX on item/quest/xp/level events).
 */
export function useAudioEffects(
  latestResult: TurnResponseWithChoices | null,
  combatActive: boolean,
  audio: UseAudioResult,
): void {
  const prevResultRef = useRef<TurnResponseWithChoices | null>(null);
  const prevCombatRef = useRef<boolean>(false);
  const victoryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Track combat state transitions for music switching
  useEffect(() => {
    if (combatActive === prevCombatRef.current) return;
    const wasCombat = prevCombatRef.current;
    prevCombatRef.current = combatActive;

    if (combatActive) {
      audio.playMusic(getMusicTrack('combat'));
    } else if (wasCombat) {
      // Combat just ended — brief victory then exploration
      audio.playMusic(getMusicTrack('victory'));
      if (victoryTimeoutRef.current) clearTimeout(victoryTimeoutRef.current);
      victoryTimeoutRef.current = setTimeout(() => {
        audio.playMusic(getMusicTrack('exploration'));
        victoryTimeoutRef.current = null;
      }, 5000);
    }
  }, [combatActive, audio]);

  // Cleanup victory timeout on unmount
  useEffect(() => {
    return () => {
      if (victoryTimeoutRef.current) clearTimeout(victoryTimeoutRef.current);
    };
  }, []);

  // Process turn result state_changes for ambient, SFX, and quest completion
  useEffect(() => {
    if (!latestResult || latestResult === prevResultRef.current) return;
    prevResultRef.current = latestResult;

    const changes = latestResult.state_changes ?? [];
    if (changes.length === 0) return;

    // --- Ambient: location change ---
    const locationChange = changes.find(
      (c: StateChange) => c.entity_type === 'location' && c.change_type === 'updated',
    );
    if (locationChange) {
      const locationType =
        (locationChange.details?.location_type as string) ??
        (locationChange.details?.type as string) ??
        '';
      if (locationType) {
        audio.playAmbient(getAmbientTrack(locationType));
      }
    }

    // --- Music: quest completion triggers brief victory ---
    const questCompletion = changes.find(
      (c: StateChange) => c.entity_type === 'quest' && c.change_type === 'completed',
    );
    if (questCompletion && !combatActive) {
      audio.playMusic(getMusicTrack('victory'));
      if (victoryTimeoutRef.current) clearTimeout(victoryTimeoutRef.current);
      victoryTimeoutRef.current = setTimeout(() => {
        audio.playMusic(getMusicTrack('exploration'));
        victoryTimeoutRef.current = null;
      }, 5000);
    }

    // --- SFX: map state_changes to sound effects ---
    for (const change of changes) {
      let sfxKey: string | null = null;

      if (change.entity_type === 'item' && change.change_type === 'created') {
        sfxKey = 'add_item';
      } else if (change.entity_type === 'quest' && change.change_type === 'created') {
        sfxKey = 'create_quest';
      } else if (change.entity_type === 'player_character') {
        if (change.change_type === 'level') {
          sfxKey = 'level_up';
        } else if (change.change_type === 'xp' || change.change_type === 'experience') {
          sfxKey = 'add_experience';
        }
      } else if (change.entity_type === 'combat') {
        if (change.change_type === 'started') {
          sfxKey = 'initiate_combat';
        } else if (change.change_type === 'ended' || change.change_type === 'resolved') {
          sfxKey = 'resolve_combat';
        }
      } else if (change.entity_type === 'location' && change.change_type === 'moved') {
        sfxKey = 'move_player';
      }

      if (sfxKey) {
        const track = getSfxTrack(sfxKey);
        if (track) {
          audio.playSfx(track);
        }
      }
    }
  }, [latestResult, combatActive, audio]);
}
