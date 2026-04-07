export interface AudioAssetMap {
  ambient: Record<string, string>; // location_type -> audio file path
  music: Record<string, string>; // game_state/mood -> audio file path
  sfx: Record<string, string>; // tool_name -> audio file path
}

export const audioAssets: AudioAssetMap = {
  ambient: {
    tavern: '/audio/ambient/tavern.mp3',
    forest: '/audio/ambient/forest.mp3',
    dungeon: '/audio/ambient/dungeon.mp3',
    city: '/audio/ambient/city.mp3',
    cave: '/audio/ambient/cave.mp3',
    ocean: '/audio/ambient/ocean.mp3',
    default: '/audio/ambient/wind.mp3',
  },
  music: {
    exploration: '/audio/music/exploration.mp3',
    combat: '/audio/music/combat.mp3',
    tavern: '/audio/music/tavern.mp3',
    mystery: '/audio/music/mystery.mp3',
    victory: '/audio/music/victory.mp3',
    default: '/audio/music/exploration.mp3',
  },
  sfx: {
    roll_dice: '/audio/sfx/dice-roll.mp3',
    skill_check: '/audio/sfx/dice-roll.mp3',
    initiate_combat: '/audio/sfx/sword-draw.mp3',
    resolve_combat: '/audio/sfx/victory-fanfare.mp3',
    add_item: '/audio/sfx/item-pickup.mp3',
    level_up: '/audio/sfx/level-up.mp3',
    add_experience: '/audio/sfx/xp-gain.mp3',
    create_quest: '/audio/sfx/quest-start.mp3',
    move_player: '/audio/sfx/footsteps.mp3',
  },
};

export function getAmbientTrack(locationType: string): string {
  return audioAssets.ambient[locationType.toLowerCase()] ?? audioAssets.ambient.default;
}

export function getMusicTrack(state: string): string {
  return audioAssets.music[state.toLowerCase()] ?? audioAssets.music.default;
}

export function getSfxTrack(toolName: string): string | null {
  return audioAssets.sfx[toolName] ?? null;
}
