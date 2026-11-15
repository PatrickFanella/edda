import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, it } from 'vitest';
import type { CharacterSheetProjection, CheckReceipt as Receipt } from '../src/api/types';
import { CheckReceipt } from '../src/components/character/CheckReceipt';
import { RulesSheet } from '../src/components/character/RulesSheet';
import fixtureJSON from './fixtures/srd-check-sheet.json?raw';

// Generated and checked by Go TestCheckSheetFixture, not calculated in React.
const fixture = JSON.parse(fixtureJSON) as {
  sheet: CharacterSheetProjection; receipts: Receipt[];
};

it('inspects server-resolved sheet actions with each die and sourced modifier', async () => {
  render(<RulesSheet sheet={fixture.sheet} receipts={fixture.receipts} />);
  const receipts = screen.getByRole('region', { name: 'Check receipts' });
  const summary = within(receipts).getByText('athletics check: 20 · success');
  await userEvent.click(summary);
  const inspected = within(summary.closest('details')!);
  expect(inspected.getByText('d20: 3 (discarded)')).toBeVisible();
  expect(inspected.getByText('d20: 15 (kept)')).toBeVisible();
  expect(inspected.getByText('strength: 3')).toBeVisible();
  expect(inspected.getByText('Recorded skill proficiency: 2')).toBeVisible();
  expect(inspected.getByText('Total: 20')).toBeVisible();
  expect(inspected.getByText('srd-5.2.1:content/character/skill-proficiencies')).toBeVisible();
  expect(inspected.getByText('srd-5.2.1:content/d20-tests/advantage-disadvantage')).toBeVisible();
  await userEvent.click(within(receipts).getByText('constitution saving throw: 8 · failure'));
  expect(within(receipts).getByText('d20: 17 (discarded)')).toBeVisible();
  expect(within(receipts).getByText('d20: 4 (kept)')).toBeVisible();
  expect(within(receipts).getByText('DC: 12')).toBeVisible();
});

it('displays a known outcome with a secret DC and an undisclosed outcome without guessing', async () => {
  const { unmount } = render(<CheckReceipt receipt={fixture.receipts[1]} />);
  await userEvent.click(screen.getByText('athletics check: 20 · success'));
  expect(screen.getByText('DC undisclosed')).toBeVisible();
  expect(screen.getByText('success', { exact: true })).toBeVisible();
  expect(screen.queryByText(/DC: 17/)).not.toBeInTheDocument();
  unmount();
  render(<CheckReceipt receipt={fixture.receipts[3]} />);
  await userEvent.click(screen.getByText('perception check: 13 · Outcome not revealed'));
  expect(screen.getByText('DC undisclosed')).toBeVisible();
  expect(screen.getByText('Outcome not revealed', { exact: true })).toBeVisible();
  expect(screen.queryByText(/success|failure/i)).not.toBeInTheDocument();
});

it('uses the server total and kept flags without recomputing dice or deciding the outcome', async () => {
  render(<CheckReceipt receipt={{ ...fixture.receipts[1], total: 42, outcome: 'failure' }} />);
  await userEvent.click(screen.getByText('athletics check: 42 · failure'));
  expect(screen.getByText('Total: 42')).toBeVisible();
  expect(screen.getByText('failure', { exact: true })).toBeVisible();
  expect(screen.getByText('d20: 15 (kept)')).toBeVisible();
});
