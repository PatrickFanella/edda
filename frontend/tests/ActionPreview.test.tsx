import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { expect, it, vi } from 'vitest';
import { ActionPreviewPanel } from '../src/components/character/ActionPreviewPanel';
import { confirmCharacterAction, readActionReceipt } from '../src/api/actions';
import { getCampaignCharacterInventory, previewCharacterAction } from '../src/api/characters';
import type { ActionPreview, CharacterSheetProjection } from '../src/api/types';
import fixture from './fixtures/srd-check-sheet.json';

vi.mock('../src/api/actions', () => ({ confirmCharacterAction: vi.fn(), readActionReceipt: vi.fn().mockResolvedValue(null), publishConfirmedAction: vi.fn() }));
vi.mock('../src/api/characters', () => ({ previewCharacterAction: vi.fn(), getCampaignCharacterInventory: vi.fn().mockResolvedValue([]) }));
const sheet = fixture.sheet as CharacterSheetProjection;
const ready: ActionPreview = {
  intent: { actor_id: 'hero', kind: 'ability-check', ability: 'strength', skill: 'athletics' },
  rules_version: 'srd-5.2.1', expected_revision: 'actor-v1:fixture', available: true,
  modifier: sheet.skills.athletics, targets: [], options: ['ordinary'], costs: [],
  cost_status: 'requires-adjudication', source_ids: ['srd-5.2.1:content/d20-tests/ability-checks'],
};
function mount(campaignId = 'alpha', actorId = 'hero') {
  return render(<QueryClientProvider client={new QueryClient()}><ActionPreviewPanel key={`${campaignId}:${actorId}`} campaignId={campaignId} actorId={actorId} sheet={sheet} /></QueryClientProvider>);
}
it('previews a selected sheet skill without confirming or performing an action', async () => {
  vi.mocked(previewCharacterAction).mockResolvedValue(ready);
  mount();
  await userEvent.selectOptions(screen.getByLabelText('Character action'), 'skill:athletics');
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  expect(await screen.findByText('Preview ready. No action has been performed.')).toBeVisible();
  expect(previewCharacterAction).toHaveBeenCalledWith('alpha', { actor_id: 'hero', kind: 'ability-check', skill: 'athletics' });
  expect(screen.getByText('Modifier: 5')).toBeVisible();
  expect(screen.getByText(/No direct resource cost/)).toHaveTextContent('Action and time costs require adjudication.');
  expect(confirmCharacterAction).not.toHaveBeenCalled();
  expect(screen.getByRole('button', { name: 'Confirm action' })).toBeEnabled();
  await userEvent.click(screen.getByText('Inspect preview sources'));
  expect(screen.getByText(/actor-v1:fixture/)).toBeVisible();
  await userEvent.selectOptions(screen.getByLabelText('Character action'), 'saving-throw:constitution');
  expect(screen.queryByText('Preview ready. No action has been performed.')).not.toBeInTheDocument();
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  expect(previewCharacterAction).toHaveBeenLastCalledWith('alpha', { actor_id: 'hero', kind: 'saving-throw', ability: 'constitution' });
});
it('shows safe unavailable reasons for sheet entries and allows retry after network failure', async () => {
  vi.mocked(previewCharacterAction).mockResolvedValueOnce({ ...ready, available: false, modifier: undefined, reason: 'This sheet entry does not support action previews yet.' }).mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce(ready);
  mount();
  await userEvent.selectOptions(screen.getByLabelText('Character action'), 'entry:attack');
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  expect(await screen.findByText('This sheet entry does not support action previews yet.')).toBeVisible();
  expect(screen.queryByText('Modifier: 5')).not.toBeInTheDocument();
  await userEvent.selectOptions(screen.getByLabelText('Character action'), 'skill:athletics');
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  expect(await screen.findByRole('alert')).toHaveTextContent('Try the preview again');
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  expect(await screen.findByText('Preview ready. No action has been performed.')).toBeVisible();
});
it('blocks overlapping selections while a preview is pending', async () => {
  let finish!: (value: ActionPreview) => void;
  vi.mocked(previewCharacterAction).mockReturnValue(new Promise(resolve => { finish = resolve; }));
  mount();
  await userEvent.selectOptions(screen.getByLabelText('Character action'), 'skill:athletics');
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  expect(screen.getByLabelText('Character action')).toBeDisabled();
  expect(screen.getByRole('button', { name: 'Preparing preview…' })).toBeDisabled();
  finish(ready);
  expect(await screen.findByText('Preview ready. No action has been performed.')).toBeVisible();
});

it('confirms once and retries uncertain delivery with the same action identity', async () => {
  vi.mocked(previewCharacterAction).mockResolvedValue(ready);
  vi.mocked(confirmCharacterAction).mockRejectedValueOnce(new Error('offline')).mockResolvedValueOnce({ action_id: 'saved', turn_number: 1, narrative: 'Recorded check.', state_changes: [], resolution_events: [], combat_active: false });
  mount('retry-campaign');
  await userEvent.selectOptions(screen.getByLabelText('Character action'), 'skill:athletics');
  await userEvent.click(screen.getByRole('button', { name: 'Preview action' }));
  await userEvent.click(await screen.findByRole('button', { name: 'Confirm action' }));
  expect(await screen.findByRole('alert')).toHaveTextContent('ID is retained');
  const request = vi.mocked(confirmCharacterAction).mock.calls[0][1];
  expect(request.expected_revision).toBe(ready.expected_revision);
  expect(request.intent).toEqual(ready.intent);
  await userEvent.click(screen.getByRole('button', { name: 'Retry same action' }));
  expect(await screen.findByText('Action saved.')).toBeVisible();
  expect(vi.mocked(confirmCharacterAction).mock.calls[1][1]).toEqual(request);
});
it('reconciles a saved receipt after remount instead of rolling again', async () => {
  const request = { action_id: 'retained', expected_revision: ready.expected_revision, intent: ready.intent };
  sessionStorage.setItem('edda:prepared-action:reload:hero', JSON.stringify(request));
  vi.mocked(readActionReceipt).mockResolvedValueOnce({ result: { action_id: 'retained', turn_number: 2, narrative: 'Saved before disconnect.', state_changes: [], resolution_events: [], combat_active: false } });
  mount('reload');
  expect(await screen.findByText('Action saved.')).toBeVisible();
  expect(readActionReceipt).toHaveBeenCalledWith('reload', 'retained');
  expect(confirmCharacterAction).not.toHaveBeenCalled();
});

it('selects an equipped longsword and known target, then confirms only on request', async () => {
 vi.mocked(getCampaignCharacterInventory).mockResolvedValueOnce([{ id:'sword',campaign_id:'martial',name:'Longsword',description:'',rarity:'common',item_type:'weapon',properties:{srd521_id:'longsword'},equipped:true,quantity:1 }]);
 vi.mocked(previewCharacterAction).mockResolvedValueOnce({ ...ready,intent:{actor_id:'hero',kind:'martial-attack',entry_id:'sword'},targets:[{id:'foe',label:'Practice foe'}],costs:[{resource_id:'attack-action',amount:1}],cost_status:'fixture-only' });
 vi.mocked(confirmCharacterAction).mockResolvedValueOnce({action_id:'saved',turn_number:2,narrative:'Longsword hit, 8 slashing.',state_changes:[],resolution_events:[],combat_active:false});
 mount('martial');
 await screen.findByRole('option',{name:'Longsword (one hand)'});
 await userEvent.selectOptions(screen.getByLabelText('Character action'),'martial-attack:sword');
 await userEvent.click(screen.getByRole('button',{name:'Preview action'}));
 expect(await screen.findByRole('button',{name:'Confirm action'})).toBeDisabled();
 await userEvent.selectOptions(screen.getByLabelText('Attack target'),'foe');
 expect(confirmCharacterAction).not.toHaveBeenCalled();
 await userEvent.click(screen.getByRole('button',{name:'Confirm action'}));
 expect(await screen.findByText('Action saved.')).toBeVisible();
 expect(vi.mocked(confirmCharacterAction).mock.calls[0][1]).toMatchObject({target_id:'foe',intent:{kind:'martial-attack',entry_id:'sword'}});
});

it('previews a potion quantity and effect and confirms self-use only', async () => {
 vi.mocked(getCampaignCharacterInventory).mockResolvedValueOnce([{ id:'potion',campaign_id:'healing',name:'Potion of Healing',description:'',rarity:'common',item_type:'consumable',properties:{srd521_id:'potion-of-healing'},equipped:false,quantity:2 }]);
 vi.mocked(previewCharacterAction).mockResolvedValueOnce({...ready,intent:{actor_id:'hero',kind:'consumable-use',entry_id:'potion'},modifier:undefined,targets:[{id:'hero',label:'Mira (self)'}],options:['Quantity: 2; regain 2d4 + 2 HP, capped at 12 (current 1); self-use only'],costs:[{resource_id:'item:potion',amount:1},{resource_id:'bonus-action',amount:1}],cost_status:'fixture-only'});
 vi.mocked(confirmCharacterAction).mockResolvedValueOnce({action_id:'saved',turn_number:2,narrative:'Mira regains 6 HP.',state_changes:[],resolution_events:[],combat_active:false});
 mount('healing');
 await screen.findByRole('option',{name:'Potion of Healing (quantity 2, self)'});
 await userEvent.selectOptions(screen.getByLabelText('Character action'),'consumable-use:potion');
 await userEvent.click(screen.getByRole('button',{name:'Preview action'}));
 expect(await screen.findByText(/Roll options: Quantity: 2; regain 2d4/)).toBeVisible();
 expect(confirmCharacterAction).not.toHaveBeenCalled();
 await userEvent.click(screen.getByRole('button',{name:'Confirm action'}));
 expect(await screen.findByText('Action saved.')).toBeVisible();
 expect(vi.mocked(confirmCharacterAction).mock.calls[0][1]).toMatchObject({target_id:'hero',intent:{kind:'consumable-use',entry_id:'potion'}});
});
