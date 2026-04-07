import { useCallback, useEffect, useRef, useState } from 'react';

const LS_AMBIENT_VOL = 'gm_audio_ambient_vol';
const LS_MUSIC_VOL = 'gm_audio_music_vol';
const LS_SFX_VOL = 'gm_audio_sfx_vol';
const LS_AMBIENT_MUTE = 'gm_audio_ambient_mute';
const LS_MUSIC_MUTE = 'gm_audio_music_mute';
const LS_SFX_MUTE = 'gm_audio_sfx_mute';

const CROSSFADE_MS = 500;
const FADE_STEPS = 20;

function readVolume(key: string, fallback: number): number {
  try {
    const raw = localStorage.getItem(key);
    if (raw !== null) {
      const parsed = Number(raw);
      if (!Number.isNaN(parsed) && parsed >= 0 && parsed <= 100) return parsed;
    }
  } catch {
    // localStorage unavailable
  }
  return fallback;
}

function readMute(key: string): boolean {
  try {
    return localStorage.getItem(key) === 'true';
  } catch {
    return false;
  }
}

function persist(key: string, value: string): void {
  try {
    localStorage.setItem(key, value);
  } catch {
    // localStorage unavailable
  }
}

function fadeOut(audio: HTMLAudioElement, durationMs: number): Promise<void> {
  return new Promise((resolve) => {
    const startVol = audio.volume;
    if (startVol <= 0) {
      audio.pause();
      resolve();
      return;
    }
    const stepMs = durationMs / FADE_STEPS;
    const decrement = startVol / FADE_STEPS;
    let step = 0;
    const interval = setInterval(() => {
      step++;
      audio.volume = Math.max(0, startVol - decrement * step);
      if (step >= FADE_STEPS) {
        clearInterval(interval);
        audio.pause();
        audio.volume = startVol;
        resolve();
      }
    }, stepMs);
  });
}

function fadeIn(audio: HTMLAudioElement, targetVol: number, durationMs: number): void {
  audio.volume = 0;
  const stepMs = durationMs / FADE_STEPS;
  const increment = targetVol / FADE_STEPS;
  let step = 0;
  const interval = setInterval(() => {
    step++;
    audio.volume = Math.min(targetVol, increment * step);
    if (step >= FADE_STEPS) {
      clearInterval(interval);
    }
  }, stepMs);
}

export interface UseAudioResult {
  ambientVolume: number;
  musicVolume: number;
  sfxVolume: number;
  ambientMuted: boolean;
  musicMuted: boolean;
  sfxMuted: boolean;
  setAmbientVolume: (v: number) => void;
  setMusicVolume: (v: number) => void;
  setSfxVolume: (v: number) => void;
  toggleAmbientMute: () => void;
  toggleMusicMute: () => void;
  toggleSfxMute: () => void;
  playAmbient: (src: string) => void;
  playMusic: (src: string) => void;
  playSfx: (src: string) => void;
  stopAll: () => void;
  userInteracted: boolean;
  requestInteraction: () => void;
}

export function useAudio(): UseAudioResult {
  const [ambientVolume, setAmbientVolumeState] = useState(() => readVolume(LS_AMBIENT_VOL, 40));
  const [musicVolume, setMusicVolumeState] = useState(() => readVolume(LS_MUSIC_VOL, 30));
  const [sfxVolume, setSfxVolumeState] = useState(() => readVolume(LS_SFX_VOL, 60));
  const [ambientMuted, setAmbientMuted] = useState(() => readMute(LS_AMBIENT_MUTE));
  const [musicMuted, setMusicMuted] = useState(() => readMute(LS_MUSIC_MUTE));
  const [sfxMuted, setSfxMuted] = useState(() => readMute(LS_SFX_MUTE));
  const [userInteracted, setUserInteracted] = useState(false);

  const ambientRef = useRef<HTMLAudioElement | null>(null);
  const musicRef = useRef<HTMLAudioElement | null>(null);
  const ambientSrcRef = useRef<string>('');
  const musicSrcRef = useRef<string>('');

  // Keep refs for current volume/mute so callbacks always see latest values
  const ambientVolRef = useRef(ambientVolume);
  const musicVolRef = useRef(musicVolume);
  const sfxVolRef = useRef(sfxVolume);
  const ambientMutedRef = useRef(ambientMuted);
  const musicMutedRef = useRef(musicMuted);
  const sfxMutedRef = useRef(sfxMuted);
  const userInteractedRef = useRef(false);

  ambientVolRef.current = ambientVolume;
  musicVolRef.current = musicVolume;
  sfxVolRef.current = sfxVolume;
  ambientMutedRef.current = ambientMuted;
  musicMutedRef.current = musicMuted;
  sfxMutedRef.current = sfxMuted;
  userInteractedRef.current = userInteracted;

  // Listen for first user interaction
  useEffect(() => {
    if (userInteracted) return;

    const handler = () => {
      setUserInteracted(true);
    };

    document.addEventListener('click', handler, { once: true });
    document.addEventListener('keydown', handler, { once: true });

    return () => {
      document.removeEventListener('click', handler);
      document.removeEventListener('keydown', handler);
    };
  }, [userInteracted]);

  // Sync volume/mute changes to active audio elements
  useEffect(() => {
    if (ambientRef.current) {
      ambientRef.current.volume = ambientMuted ? 0 : ambientVolume / 100;
    }
  }, [ambientVolume, ambientMuted]);

  useEffect(() => {
    if (musicRef.current) {
      musicRef.current.volume = musicMuted ? 0 : musicVolume / 100;
    }
  }, [musicVolume, musicMuted]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      try { ambientRef.current?.pause(); } catch { /* noop */ }
      try { musicRef.current?.pause(); } catch { /* noop */ }
      ambientRef.current = null;
      musicRef.current = null;
    };
  }, []);

  const setAmbientVolume = useCallback((v: number) => {
    const clamped = Math.max(0, Math.min(100, v));
    setAmbientVolumeState(clamped);
    persist(LS_AMBIENT_VOL, String(clamped));
  }, []);

  const setMusicVolume = useCallback((v: number) => {
    const clamped = Math.max(0, Math.min(100, v));
    setMusicVolumeState(clamped);
    persist(LS_MUSIC_VOL, String(clamped));
  }, []);

  const setSfxVolume = useCallback((v: number) => {
    const clamped = Math.max(0, Math.min(100, v));
    setSfxVolumeState(clamped);
    persist(LS_SFX_VOL, String(clamped));
  }, []);

  const toggleAmbientMute = useCallback(() => {
    setAmbientMuted((prev) => {
      const next = !prev;
      persist(LS_AMBIENT_MUTE, String(next));
      return next;
    });
  }, []);

  const toggleMusicMute = useCallback(() => {
    setMusicMuted((prev) => {
      const next = !prev;
      persist(LS_MUSIC_MUTE, String(next));
      return next;
    });
  }, []);

  const toggleSfxMute = useCallback(() => {
    setSfxMuted((prev) => {
      const next = !prev;
      persist(LS_SFX_MUTE, String(next));
      return next;
    });
  }, []);

  const playAmbient = useCallback((src: string) => {
    if (!userInteractedRef.current) return;
    if (ambientSrcRef.current === src && ambientRef.current && !ambientRef.current.paused) return;

    const effectiveVol = ambientMutedRef.current ? 0 : ambientVolRef.current / 100;

    const startNew = () => {
      const audio = new Audio(src);
      audio.loop = true;
      audio.volume = 0;
      ambientRef.current = audio;
      ambientSrcRef.current = src;
      audio.play().then(() => {
        fadeIn(audio, effectiveVol, CROSSFADE_MS);
      }).catch(() => {
        // Audio file missing or autoplay blocked — graceful degradation
      });
    };

    if (ambientRef.current && !ambientRef.current.paused) {
      const old = ambientRef.current;
      fadeOut(old, CROSSFADE_MS).then(startNew);
    } else {
      startNew();
    }
  }, []);

  const playMusic = useCallback((src: string) => {
    if (!userInteractedRef.current) return;
    if (musicSrcRef.current === src && musicRef.current && !musicRef.current.paused) return;

    const effectiveVol = musicMutedRef.current ? 0 : musicVolRef.current / 100;

    try { musicRef.current?.pause(); } catch { /* noop */ }

    const audio = new Audio(src);
    audio.loop = true;
    audio.volume = effectiveVol;
    musicRef.current = audio;
    musicSrcRef.current = src;
    audio.play().catch(() => {
      // Audio file missing or autoplay blocked — graceful degradation
    });
  }, []);

  const playSfx = useCallback((src: string) => {
    if (!userInteractedRef.current) return;
    if (sfxMutedRef.current) return;

    try {
      const audio = new Audio(src);
      audio.volume = sfxVolRef.current / 100;
      audio.play().catch(() => {
        // Audio file missing — graceful degradation
      });
    } catch {
      // Constructor failed — graceful degradation
    }
  }, []);

  const stopAll = useCallback(() => {
    try { ambientRef.current?.pause(); } catch { /* noop */ }
    try { musicRef.current?.pause(); } catch { /* noop */ }
    ambientRef.current = null;
    musicRef.current = null;
    ambientSrcRef.current = '';
    musicSrcRef.current = '';
  }, []);

  const requestInteraction = useCallback(() => {
    setUserInteracted(true);
  }, []);

  return {
    ambientVolume,
    musicVolume,
    sfxVolume,
    ambientMuted,
    musicMuted,
    sfxMuted,
    setAmbientVolume,
    setMusicVolume,
    setSfxVolume,
    toggleAmbientMute,
    toggleMusicMute,
    toggleSfxMute,
    playAmbient,
    playMusic,
    playSfx,
    stopAll,
    userInteracted,
    requestInteraction,
  };
}
