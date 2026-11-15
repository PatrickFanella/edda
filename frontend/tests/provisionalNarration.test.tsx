import { act, render, screen } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import { getSessionHistory } from '../src/api/campaigns';
import { NarrativePanel } from '../src/components/narrative/NarrativePanel';
import { useCampaign } from '../src/hooks/useCampaign';
import { useNarrative } from '../src/hooks/useNarrative';
import { result, timestamp } from './fixtures/data';

vi.mock('../src/api/campaigns', () => ({ getSessionHistory: vi.fn() }));
vi.mock('../src/hooks/useCampaign', () => ({ useCampaign: vi.fn() }));

class Socket extends EventTarget {
  static OPEN = 1;
  static CONNECTING = 0;
  static instances: Socket[] = [];
  readyState = 0;
  actions: { input: string; action_id: string }[] = [];
  constructor() { super(); Socket.instances.push(this); }
  open() { this.readyState = 1; this.dispatchEvent(new Event('open')); }
  close() { this.readyState = 3; this.dispatchEvent(new Event('close')); }
  send(raw: string) { this.actions.push(JSON.parse(raw).payload); }
  frame(type: string, payload: unknown) {
    this.dispatchEvent(new MessageEvent('message', { data: JSON.stringify({ type, payload, timestamp }) }));
  }
}

it.each(['error', 'disconnect', 'malformed'] as const)('discards provisional text on %s, reuses the action on retry and commits only its result', async (failure) => {
  vi.useFakeTimers(); Socket.instances = []; vi.stubGlobal('WebSocket', Socket);
  vi.mocked(getSessionHistory).mockResolvedValue({ entries: [] });
  vi.mocked(useCampaign).mockReturnValue({ campaignId: 'alpha' } as ReturnType<typeof useCampaign>);
  function Harness() {
    const narrative = useNarrative();
    return <>
      <button onClick={() => narrative.sendAction('Search')}>Send</button>
      <output aria-label="Committed GM count">{narrative.entries.filter(entry => entry.kind === 'gm').length}</output>
      <NarrativePanel entries={narrative.entries} streamingEntry={narrative.streamingEntry} isLoading={narrative.isLoading} />
    </>;
  }
  render(<Harness />);
  await act(() => vi.advanceTimersByTimeAsync(0));
  let socket = Socket.instances[0]; act(() => socket.open());
  act(() => screen.getByText('Send').click());
  const first = socket.actions[0];
  act(() => socket.frame('chunk', { text: 'You enter ' }));
  expect(screen.getByLabelText('Provisional narration')).toHaveTextContent('You enter');
  act(() => socket.frame('chunk', { text: 'the castle.' }));
  expect(screen.getByLabelText('Provisional narration')).toHaveTextContent('You enter the castle.');
  expect(screen.getByText('Provisional · not saved')).toBeVisible();
  expect(screen.getByLabelText('Committed GM count')).toHaveTextContent('0');
  act(() => socket.frame('status', { stage: 'narrative_reset', description: 'Correcting tools' }));
  expect(screen.queryByText('You enter the castle.')).not.toBeInTheDocument();
  act(() => socket.frame('chunk', { text: 'A replacement preview.' }));
  act(() => {
    if (failure === 'error') socket.frame('error', { error: 'Recover the committed action', action_id: first.action_id });
    else if (failure === 'disconnect') socket.close();
    else socket.dispatchEvent(new MessageEvent('message', { data: '{' }));
  });
  expect(screen.queryByLabelText('Provisional narration')).not.toBeInTheDocument();
  expect(screen.getByLabelText('Committed GM count')).toHaveTextContent('0');
  if (failure === 'disconnect') {
    await act(() => vi.advanceTimersByTimeAsync(1000));
    await act(() => vi.advanceTimersByTimeAsync(0));
    socket = Socket.instances[1]; act(() => socket.open());
  }
  act(() => screen.getByText('Send').click());
  expect(socket.actions.at(-1)).toEqual(first);
  expect(screen.queryByText('A replacement preview.')).not.toBeInTheDocument();
  act(() => socket.frame('chunk', { text: 'Recovered preview.' }));
  act(() => socket.frame('result', { ...result, recovered: true, action_id: first.action_id, turn_number: 4 }));
  expect(screen.queryByLabelText('Provisional narration')).not.toBeInTheDocument();
  expect(screen.queryByText('Recovered preview.')).not.toBeInTheDocument();
  expect(screen.getByLabelText('Committed GM count')).toHaveTextContent('1');
  expect(screen.getByText(result.narrative)).toBeVisible();
});
