import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, it, vi } from 'vitest';
import { RulesSheet } from '../src/components/character/RulesSheet';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { CharacterSheet } from '../src/components/character/CharacterSheet';
import { getCampaignCharacter } from '../src/api/characters';
import { character } from './fixtures/data';
import type { CharacterSheetProjection } from '../src/api/types';
import fixture from './fixtures/srd-check-sheet.json';
it('displays server-derived values and inspectable terms without a client XP formula', async () => {
 const value = { value: 5, derivation: [{ label: 'Strength', value: 3, source_id: 'srd-5.2.1:content/character/ability-modifiers' }, { label: 'Proficiency', value: 2, source_id: 'srd-5.2.1:content/character/proficiency' }] };
 const sheet: CharacterSheetProjection = { coverage: { rules_version: '5.2.1', status: 'partial', full_support: false, qualified_register_entries: 0, register_entries: 62, planned_character_scopes: 344, planned_world_scopes: 335, planned_spell_item_scopes: 753, supported_scope: 'Fixture-only supported scope.' }, rules_version: 'srd-5.2.1', scope: 'fixture', abilities: {}, proficiency: value, saves: {}, skills: { athletics: value }, hp: { ...value, value: 9 }, max_hp: { ...value, value: 12 }, defenses: {}, resources: [{ id: 'second-wind', label: 'Second Wind', current: 1, maximum: 2, source_id: 'srd-5.2.1:content/fighter/second-wind' }], capabilities: [{ id: 'attack', label: 'Attack', playable: false, source_id: 'srd-5.2.1:content/actions/attack' }], progression: { current_xp: 125, level_xp: 0, next_level_xp: 300, source_id: 'srd-5.2.1:content/character/level-advancement' } };
 render(<RulesSheet sheet={sheet} />);
 expect(screen.getByText('HP 9 / 12 · XP 125 / 300')).toBeVisible();
 expect(screen.getByText('Second Wind: 1 / 2')).toBeVisible();
 expect(screen.getByText('Attack (read only)')).toBeVisible();
 expect(screen.getByText(/344 character\/advancement scopes, 335 world scopes and 753 spell\/item source families/)).toBeVisible();
 expect(screen.getByText(/Edda does not claim Full Rules support/)).toBeVisible();
 await userEvent.click(screen.getAllByText('+5')[0]);
 expect(screen.getAllByText('Strength: 3')[0]).toBeVisible();
 expect(screen.getAllByText('srd-5.2.1:content/character/proficiency')[0]).toBeVisible();
});

vi.mock('../src/api/characters', () => ({ getCampaignCharacter: vi.fn() }));
it.each(['legacy', 'srd-5.2.1'])('keeps the %s character adapter readable', async (version) => {
 vi.mocked(getCampaignCharacter).mockResolvedValue({ ...character(), rules_version: version, sheet_unavailable: version === 'srd-5.2.1' ? 'Unsupported fixture build' : undefined });
 render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><CharacterSheet campaignId="alpha" /></QueryClientProvider>);
 if (version === 'legacy') { expect(await screen.findByText('Experience')).toBeVisible(); expect(screen.queryByText('Unsupported fixture build')).not.toBeInTheDocument(); }
 else { expect(await screen.findByText('Unsupported fixture build')).toBeVisible(); expect(screen.getByText(/A scout · HP 10 \/ 10/)).toBeVisible(); expect(screen.queryByText('Experience')).not.toBeInTheDocument(); }
});

it('describes supported Action Surge separately from read-only Tactical Mind', () => {
 const sheet = structuredClone(fixture.sheet) as CharacterSheetProjection;
 sheet.capabilities.push(
  { id: 'action-surge', label: 'Action Surge', playable: true, source_id: 'srd-5.2.1:content/fighter/action-surge' },
  { id: 'tactical-mind', label: 'Tactical Mind', playable: false, source_id: 'srd-5.2.1:content/fighter/tactical-mind' },
 );
 render(<RulesSheet sheet={sheet} />);
 expect(screen.getByText('Action Surge (supported)')).toBeVisible();
 expect(screen.getByText('Tactical Mind (read only)')).toBeVisible();
 expect(screen.getByText(/Action Surge supports one extra Action except Magic on your owned active ordinary encounter turn/)).toBeVisible();
});
