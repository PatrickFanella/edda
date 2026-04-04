import type { CampaignProfile, CharacterProfile, OpeningSceneResponse } from '../api/types';

export interface GuidedCharacterAttributes {
  readonly name: string;
  readonly race: string;
  readonly class: string;
  readonly background: string;
  readonly alignment: string;
  readonly traits?: readonly string[];
}

export interface StartupPlaySeed {
  readonly campaignName: string;
  readonly campaignSummary: string;
  readonly openingScene: OpeningSceneResponse;
  readonly seededAt: string;
}

const CLASS_MOTIVATIONS: Record<string, readonly string[]> = {
  Barbarian: ['Prove strength in battle', 'Protect the tribe'],
  Bard: ['Collect epic tales', 'Inspire others through art'],
  Cleric: ['Serve the divine', 'Heal the wounded'],
  Druid: ['Preserve the natural order', 'Guard the wilds'],
  Fighter: ['Protect the weak', 'Master the art of combat'],
  Monk: ['Achieve inner perfection', 'Uphold monastic traditions'],
  Paladin: ['Uphold a sacred oath', 'Smite the wicked'],
  Ranger: ['Guard the frontier', 'Hunt dangerous beasts'],
  Rogue: ['Seek fortune and thrills', 'Uncover hidden truths'],
  Sorcerer: ['Understand innate power', 'Prove worthy of the gift'],
  Warlock: ['Fulfill the pact', 'Gain forbidden knowledge'],
  Wizard: ['Seek arcane knowledge', 'Unlock the mysteries of magic'],
};

const RACE_STRENGTHS: Record<string, readonly string[]> = {
  Human: ['Versatile adaptability'],
  Elf: ['Keen perception', 'Grace under pressure'],
  Dwarf: ['Unyielding resilience', 'Stonework intuition'],
  Halfling: ['Remarkable luck', 'Nimble evasion'],
  Gnome: ['Inventive ingenuity', 'Arcane curiosity'],
  'Half-Elf': ['Social adaptability', 'Dual heritage insight'],
  'Half-Orc': ['Savage endurance', 'Intimidating presence'],
  Tiefling: ['Infernal resilience', 'Force of personality'],
  Dragonborn: ['Draconic breath weapon', 'Commanding presence'],
};

const CLASS_STRENGTHS: Record<string, readonly string[]> = {
  Barbarian: ['Raw physical power'],
  Bard: ['Silver tongue'],
  Cleric: ['Divine channeling'],
  Druid: ['Nature magic'],
  Fighter: ['Combat expertise'],
  Monk: ['Unarmed discipline'],
  Paladin: ['Righteous resolve'],
  Ranger: ['Wilderness tracking'],
  Rogue: ['Stealth and cunning'],
  Sorcerer: ['Innate spellcasting'],
  Warlock: ['Eldritch invocations'],
  Wizard: ['Scholarly spellcraft'],
};

const BACKGROUND_WEAKNESSES: Record<string, readonly string[]> = {
  Acolyte: ['Naïve idealism', 'Blind faith'],
  Charlatan: ['Compulsive dishonesty', 'Difficulty forming trust'],
  Criminal: ['Trust issues', 'Haunted by past deeds'],
  Entertainer: ['Craves attention', 'Fear of being forgotten'],
  'Folk Hero': ['Stubborn pride', 'Unrealistic expectations'],
  'Guild Artisan': ['Obsessive perfectionism', 'Material attachment'],
  Hermit: ['Social awkwardness', 'Distrust of authority'],
  Noble: ['Sense of entitlement', 'Sheltered worldview'],
  Outlander: ['Distrust of civilization', 'Blunt to a fault'],
  Sage: ['Absent-minded', 'Overthinks simple problems'],
  Sailor: ['Restless on land', 'Rough manners'],
  Soldier: ['Rigid discipline', 'Haunted by war'],
  Urchin: ['Deep-seated insecurity', 'Hoarding instinct'],
};

const DEFAULT_MOTIVATIONS = ['Seek adventure', 'Find a place in the world'] as const;
const DEFAULT_STRENGTHS = ['Determined spirit'] as const;
const DEFAULT_WEAKNESSES = ['Untested resolve'] as const;

export function summarizeCampaignProfile(profile: CampaignProfile): string {
  const parts: string[] = [];

  if (profile.tone !== '' || profile.genre !== '') {
    parts.push(`${profile.tone} ${profile.genre}`.trim());
  }
  if (profile.themes.length > 0) {
    parts.push(`themes of ${profile.themes.join(', ')}`);
  }
  if (profile.world_type !== '') {
    parts.push(`set in a ${profile.world_type} world`);
  }
  if (profile.danger_level !== '') {
    parts.push(`with ${profile.danger_level} danger`);
  }
  if (profile.political_complexity !== '') {
    parts.push(`and ${profile.political_complexity} politics`);
  }

  if (parts.length === 0) {
    return 'A new adventure awaits.';
  }

  return `${parts.join(' ').trim()}.`;
}

export function buildCharacterProfileFromGuidedAttributes(attributes: GuidedCharacterAttributes): CharacterProfile {
  const concept = `${attributes.race.toLowerCase()} ${attributes.class.toLowerCase()}`.trim();
  const traits = (attributes.traits ?? []).filter((trait) => trait.trim().length > 0);

  let personality = attributes.alignment;
  if (traits.length > 0) {
    personality = `${personality}; ${traits.join('; ')}`;
  }

  return {
    name: attributes.name.trim(),
    concept,
    background: attributes.background,
    personality,
    motivations: lookupStrings(CLASS_MOTIVATIONS, attributes.class, DEFAULT_MOTIVATIONS),
    strengths: mergeStrengths(attributes.race, attributes.class),
    weaknesses: lookupStrings(BACKGROUND_WEAKNESSES, attributes.background, DEFAULT_WEAKNESSES),
  };
}

function lookupStrings(
  source: Record<string, readonly string[]>,
  key: string,
  fallback: readonly string[],
): string[] {
  return [...(source[key] ?? fallback)];
}

function mergeStrengths(race: string, characterClass: string): string[] {
  return [
    ...lookupStrings(RACE_STRENGTHS, race, DEFAULT_STRENGTHS),
    ...lookupStrings(CLASS_STRENGTHS, characterClass, DEFAULT_STRENGTHS),
  ];
}
