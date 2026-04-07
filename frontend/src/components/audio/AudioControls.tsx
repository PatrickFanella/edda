import { useState } from 'react';

import type { UseAudioResult } from '../../hooks/useAudio';

type AudioControlsProps = UseAudioResult;

function VolumeIcon({ muted, className = '' }: { readonly muted: boolean; readonly className?: string }) {
  if (muted) {
    return (
      <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
        <line x1="23" y1="9" x2="17" y2="15" />
        <line x1="17" y1="9" x2="23" y2="15" />
      </svg>
    );
  }

  return (
    <svg className={className} width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
      <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
      <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
    </svg>
  );
}

function SpeakerIcon({ className = '' }: { readonly className?: string }) {
  return (
    <svg className={className} width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
      <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
      <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
    </svg>
  );
}

interface LayerRowProps {
  readonly label: string;
  readonly muted: boolean;
  readonly volume: number;
  readonly onToggleMute: () => void;
  readonly onVolumeChange: (v: number) => void;
}

function LayerRow({ label, muted, volume, onToggleMute, onVolumeChange }: LayerRowProps) {
  return (
    <div className="flex items-center gap-3">
      <span className="w-16 text-xs font-semibold uppercase tracking-wide text-champagne/70">{label}</span>
      <button
        type="button"
        onClick={onToggleMute}
        className={`flex h-7 w-7 shrink-0 items-center justify-center border transition-colors ${
          muted
            ? 'border-pewter/30 text-pewter hover:border-pewter hover:text-pewter'
            : 'border-gold/40 text-gold hover:border-gold'
        }`}
        aria-label={`${muted ? 'Unmute' : 'Mute'} ${label}`}
      >
        <VolumeIcon muted={muted} />
      </button>
      <input
        type="range"
        min={0}
        max={100}
        value={volume}
        onChange={(e) => onVolumeChange(Number(e.target.value))}
        className="h-1.5 flex-1 cursor-pointer appearance-none bg-charcoal accent-gold [&::-webkit-slider-runnable-track]:h-1.5 [&::-webkit-slider-runnable-track]:rounded [&::-webkit-slider-runnable-track]:bg-midnight [&::-webkit-slider-thumb]:mt-[-3px] [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-gold"
        aria-label={`${label} volume`}
      />
      <span className="w-8 text-right text-xs tabular-nums text-pewter">{volume}</span>
    </div>
  );
}

export function AudioControls(props: AudioControlsProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="inline-flex items-center justify-center border-2 border-gold/30 px-3 py-2 text-champagne transition-all duration-200 hover:border-gold hover:text-gold focus:outline-none focus:ring-2 focus:ring-gold focus:ring-offset-2 focus:ring-offset-obsidian"
        aria-label="Toggle audio controls"
        aria-expanded={expanded}
      >
        <SpeakerIcon />
      </button>

      {expanded && (
        <div className="absolute right-0 top-full z-50 mt-2 w-72 border-2 border-gold/20 bg-obsidian p-4 shadow-lg">
          <h3 className="mb-3 font-heading text-xs font-semibold uppercase tracking-[0.15em] text-gold">
            Audio
          </h3>

          {!props.userInteracted && (
            <button
              type="button"
              onClick={props.requestInteraction}
              className="mb-3 w-full border border-gold/40 bg-gold/5 px-3 py-2 text-xs text-gold transition-colors hover:bg-gold/10"
            >
              Click to enable audio
            </button>
          )}

          <div className="space-y-3">
            <LayerRow
              label="Ambient"
              muted={props.ambientMuted}
              volume={props.ambientVolume}
              onToggleMute={props.toggleAmbientMute}
              onVolumeChange={props.setAmbientVolume}
            />
            <LayerRow
              label="Music"
              muted={props.musicMuted}
              volume={props.musicVolume}
              onToggleMute={props.toggleMusicMute}
              onVolumeChange={props.setMusicVolume}
            />
            <LayerRow
              label="SFX"
              muted={props.sfxMuted}
              volume={props.sfxVolume}
              onToggleMute={props.toggleSfxMute}
              onVolumeChange={props.setSfxVolume}
            />
          </div>
        </div>
      )}
    </div>
  );
}
