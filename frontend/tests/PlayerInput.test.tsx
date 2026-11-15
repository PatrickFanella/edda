import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, it, vi } from 'vitest';
import { PlayerInput } from '../src/components/narrative/PlayerInput';
it('retains a rejected draft and clears it only after acceptance', async () => {
  const send = vi.fn().mockReturnValueOnce(false).mockReturnValueOnce(true);
  render(<PlayerInput onSendAction={send} />);
  const user = userEvent.setup(); const input = screen.getByRole('textbox');
  await user.type(input, 'Open the gate{Enter}');
  expect(input).toHaveValue('Open the gate');
  await user.keyboard('{Enter}');
  expect(input).toHaveValue(''); expect(send).toHaveBeenCalledTimes(2);
});
it('keeps Shift+Enter as a newline and blocks actions during a turn', async () => {
  const send = vi.fn(); const user = userEvent.setup();
  const { rerender } = render(<PlayerInput onSendAction={send} />);
  await user.type(screen.getByRole('textbox'), 'Look{Shift>}{Enter}{/Shift}around');
  expect(screen.getByRole('textbox')).toHaveValue('Look\naround');
  rerender(<PlayerInput onSendAction={send} isLoading />);
  expect(screen.getByRole('textbox')).toBeDisabled(); expect(send).not.toHaveBeenCalled();
});
