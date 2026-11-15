import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { expect, it, vi } from 'vitest';
import { ReturnReminder } from '../src/components/narrative/ReturnReminder';
import { getReturnReminder, type ReturnReminder as Reminder } from '../src/api/reminder';

vi.mock('../src/api/reminder', () => ({ getReturnReminder: vi.fn() }));
const data: Reminder = {
  scene: { text: 'North gate', source_id: 'known-location:gate' },
  objective: { text: 'Find the keeper', source_id: 'quest:keeper' },
  known_facts: [{ text: 'The blue seal is broken.', source_id: 'known-fact:seal' }],
};
function mount(campaignId = 'alpha') {
  return render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><ReturnReminder key={campaignId} campaignId={campaignId} /></QueryClientProvider>);
}
it('quotes the provided scene, active quest and known facts, and can dismiss/reopen', async () => {
  vi.mocked(getReturnReminder).mockResolvedValue(data);
  mount();
  await userEvent.click(screen.getByRole('button', { name: 'Where was I?' }));
  expect(await screen.findByText(/North gate/)).toBeVisible();
  expect(screen.getByText(/Find the keeper/)).toBeVisible();
  expect(screen.getByRole('list', { name: 'Known facts' })).toHaveTextContent('The blue seal is broken.');
  await userEvent.click(screen.getByRole('button', { name: 'Dismiss reminder' }));
  expect(screen.getByText(/North gate/)).not.toBeVisible();
  expect(screen.getByRole('button', { name: 'Where was I?' })).toHaveAttribute('aria-expanded', 'false');
  await userEvent.click(screen.getByRole('button', { name: 'Where was I?' }));
  expect(screen.getByText(/North gate/)).toBeVisible();
});
it('allows dismissal while delayed and exposes retry after failure', async () => {
  let fail!: (error: Error) => void;
  vi.mocked(getReturnReminder).mockReturnValueOnce(new Promise((_resolve, reject) => { fail = reject; })).mockResolvedValue({ known_facts: [] });
  mount();
  await userEvent.click(screen.getByRole('button', { name: 'Where was I?' }));
  expect(screen.getByRole('status')).toHaveTextContent('You can keep playing.');
  await userEvent.click(screen.getByRole('button', { name: 'Dismiss reminder' }));
  await act(async () => fail(new Error('provider unavailable')));
  await userEvent.click(screen.getByRole('button', { name: 'Where was I?' }));
  // Reopening retries a failed read; an empty result never invents missing context.
  expect(await screen.findByText(/No current location is known yet/)).toBeVisible();
  expect(screen.getByText(/No active quest is recorded/)).toBeVisible();
  expect(screen.getByText('No current known facts are recorded.')).toBeVisible();
});
it('reports safe failure and retries without displaying raw errors', async () => {
  vi.mocked(getReturnReminder).mockRejectedValueOnce(new Error('private diagnostic')).mockResolvedValueOnce(data);
  mount('beta');
  await userEvent.click(screen.getByRole('button', { name: 'Where was I?' }));
  await userEvent.click(await screen.findByRole('button', { name: 'Retry reminder' }));
  expect(await screen.findByText(/North gate/)).toBeVisible();
  expect(screen.queryByText(/private diagnostic/)).not.toBeInTheDocument();
  expect(getReturnReminder).toHaveBeenCalledWith('beta');
});
