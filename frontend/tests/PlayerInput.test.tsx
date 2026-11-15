import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, it, vi } from 'vitest';
import { PlayerInput } from '../src/components/narrative/PlayerInput';
it('retains a rejected draft and clears it only after acceptance', async () => {
  const send = vi.fn().mockReturnValueOnce(false).mockReturnValueOnce(true);
  render(<PlayerInput onSendAction={send} />);
  const user = userEvent.setup(); const input = screen.getByRole('textbox');
  await user.type(input, 'Open the gate{Enter}');
  await waitFor(() => expect(input).toBeEnabled());
  expect(input).toHaveValue('Open the gate');
  expect(send).toHaveBeenCalledTimes(1);
  await user.keyboard('{Enter}');
  await waitFor(() => expect(input).toHaveValue(''));
  expect(send).toHaveBeenCalledTimes(2);
});
it('keeps Shift+Enter as a newline and blocks actions during a turn', async () => {
  const send = vi.fn(); const user = userEvent.setup();
  const { rerender } = render(<PlayerInput onSendAction={send} />);
  await user.type(screen.getByRole('textbox'), 'Look{Shift>}{Enter}{/Shift}around');
  expect(screen.getByRole('textbox')).toHaveValue('Look\naround');
  rerender(<PlayerInput onSendAction={send} isLoading />);
  expect(screen.getByRole('textbox')).toBeDisabled(); expect(send).not.toHaveBeenCalled();
});
it('matches suggestions locally and selects an editable draft without submitting', async () => {
  const send = vi.fn();
  render(<PlayerInput onSendAction={send} suggestions={[
    { id: 'clue', text: 'Investigate the blue seal', reason: 'A known clue, with an uncertain outcome.', source_ids: ['known-fact:seal'] },
    { id: 'skill', text: 'Look around using Perception', reason: 'Your sheet supports this check.', source_ids: ['srd:skill'] },
  ]} />);
  await userEvent.type(screen.getByRole('textbox'), 'I inv');
  expect(screen.queryByRole('button', { name: /Look around/ })).not.toBeInTheDocument();
  await userEvent.click(screen.getByRole('button', { name: /Investigate the blue seal/ }));
  expect(screen.getByRole('textbox')).toHaveValue('Investigate the blue seal');
  expect(send).not.toHaveBeenCalled();
  await userEvent.type(screen.getByRole('textbox'), ' carefully');
  expect(screen.getByRole('textbox')).toHaveValue('Investigate the blue seal carefully');
});

const completion = [{ id: 'clue', text: 'Investigate the blue seal', reason: 'A known clue.', source_ids: ['known-fact:seal'] }];
it('accepts only a visible prefix completion with Tab, then keeps the draft editable', async () => {
  const send = vi.fn(); const user = userEvent.setup();
  render(<PlayerInput onSendAction={send} suggestions={completion} />);
  const input = screen.getByRole('textbox');
  await user.type(input, 'Inv');
  expect(screen.getByTestId('action-completion')).toHaveTextContent('Investigate the blue seal');
  expect(input).toHaveAccessibleDescription(/Press Tab to insert or Escape to dismiss/);
  await user.keyboard('{Tab}');
  expect(input).toHaveValue(completion[0].text); expect(input).toHaveFocus();
  expect(send).not.toHaveBeenCalled();
  await user.keyboard(' carefully');
  expect(input).toHaveValue('Investigate the blue seal carefully');
  await user.tab(); expect(input).not.toHaveFocus();
});
it('leaves Tab and Shift+Tab traversal alone after dismissal, non-prefix matches and selection edits', async () => {
  const user = userEvent.setup(); const send = vi.fn();
  render(<PlayerInput onSendAction={send} suggestions={completion} />);
  const input = screen.getByRole('textbox');
  await user.type(input, 'Inv'); await user.keyboard('{Escape}');
  expect(screen.queryByTestId('action-completion')).not.toBeInTheDocument();
  await user.tab(); expect(input).not.toHaveFocus();
  await user.clear(input); await user.type(input, 'I inv');
  expect(screen.queryByTestId('action-completion')).not.toBeInTheDocument();
  await user.tab(); expect(input).not.toHaveFocus();
  await user.clear(input); await user.type(input, 'Inv');
  fireEvent.select(input, { target: { selectionStart: 0, selectionEnd: 2 } });
  expect(screen.queryByTestId('action-completion')).not.toBeInTheDocument();
  await user.tab(); expect(input).toHaveValue('Inv');
  await user.click(input); await user.keyboard('{End}');
  await user.tab({ shift: true }); expect(input).not.toHaveFocus();
  expect(send).not.toHaveBeenCalled();
});
it('does not accept or send while an IME composition is active', async () => {
  const send = vi.fn(); const user = userEvent.setup();
  render(<PlayerInput onSendAction={send} suggestions={completion} />);
  const input = screen.getByRole('textbox'); await user.type(input, 'Inv');
  fireEvent.compositionStart(input);
  expect(screen.queryByTestId('action-completion')).not.toBeInTheDocument();
  fireEvent.keyDown(input, { key: 'Tab', isComposing: true });
  fireEvent.keyDown(input, { key: 'Enter', keyCode: 229 });
  expect(input).toHaveValue('Inv'); expect(send).not.toHaveBeenCalled();
  fireEvent.compositionEnd(input);
  await user.keyboard('{Tab}'); expect(input).toHaveValue(completion[0].text);
});
it('inserts a suggestion by an explicit button and restores the composer focus', async () => {
  const send = vi.fn(); const user = userEvent.setup();
  render(<PlayerInput onSendAction={send} suggestions={completion} />);
  await user.click(screen.getByRole('button', { name: /Investigate the blue seal/ }));
  expect(screen.getByRole('textbox')).toHaveFocus();
  expect(screen.getByRole('textbox')).toHaveValue(completion[0].text);
  expect(send).not.toHaveBeenCalled();
});
