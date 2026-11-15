import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { expect, it, vi } from 'vitest';
import { ActionPreviewPanel } from '../src/components/character/ActionPreviewPanel';
import { previewCharacterAction } from '../src/api/characters';
import type { ActionPreview, CharacterSheetProjection } from '../src/api/types';
import fixture from './fixtures/srd-check-sheet.json';

vi.mock('../src/api/characters', () => ({ previewCharacterAction: vi.fn() }));
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
  expect(screen.queryByRole('button', { name: /confirm|roll|submit/i })).not.toBeInTheDocument();
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
