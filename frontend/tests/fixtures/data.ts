import type { CampaignResponse, CharacterResponse, SessionLogEntry } from '../../src/api/types';
export const timestamp = '2026-01-01T12:00:00Z';
export const user = { id: 'fixture-user', name: 'Fixture Player', email: 'player@example.test' };
export const campaign = (id = 'alpha'): CampaignResponse => ({ id, name: `Adventure ${id}`, description: 'An isolated adventure', genre: 'fantasy', tone: 'hopeful', themes: ['discovery'], status: 'active', rules_mode: 'narrative', created_by: user.id, created_at: timestamp, updated_at: timestamp });
export const character = (id = 'alpha'): CharacterResponse => ({ id: 'hero', campaign_id: id, user_id: user.id, name: 'Mira', description: 'A scout', stats: {}, hp: 10, max_hp: 10, experience: 0, level: 1, status: 'active', abilities: [] });
export const history = (id = 'alpha'): SessionLogEntry[] => [
  { turn_number: 1, player_input: '', input_type: 'narrative', llm_response: `The ${id} gate opens.`, choices: ['Enter the gate'], created_at: timestamp },
  { turn_number: 2, player_input: 'Enter the gate', input_type: 'action', llm_response: `The ${id} keeper welcomes you.`, created_at: timestamp },
];
export const result = { narrative: 'You find a silver key.', state_changes: [], resolution_events: [], choices: [{ id: 'inspect', text: 'Inspect the key' }] };
export const profile = { genre: 'fantasy', tone: 'hopeful', themes: ['discovery'], world_type: 'frontier', danger_level: 'low', political_complexity: 'low' };
export const heroProfile = { name: 'Mira', concept: 'Scout', background: 'Wanderer', personality: 'Curious', motivations: ['Explore'], strengths: ['Perception'], weaknesses: ['Haste'] };
