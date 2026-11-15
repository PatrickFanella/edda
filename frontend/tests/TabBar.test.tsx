import { useState } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, it } from 'vitest';
import { TabBar } from '../src/components/layout/TabBar';

const tabs = ['Narrative', 'Character', 'Inventory', 'Quests', 'NPCs', 'World', 'Journal', 'Logs'].map(key => ({ key, label: key }));
function Sections() {
  const [active, setActive] = useState('Narrative');
  return <><TabBar idPrefix="fixture" tabs={tabs} activeTab={active} onChange={setActive} />
    {tabs.map(tab => <div key={tab.key} role="tabpanel" id={`fixture-panel-${tab.key}`} aria-labelledby={`fixture-tab-${tab.key}`} hidden={active !== tab.key} tabIndex={0}>{tab.label} content</div>)}
    <button>After panels</button></>;
}
it('uses one tab stop, automatic arrow selection, wraparound, Home and End for all sections', async () => {
  render(<Sections />); const user = userEvent.setup(); await user.tab();
  const assertSelected = (name: string) => {
    const tab = screen.getByRole('tab', { name });
    expect(tab).toHaveFocus(); expect(tab).toHaveAttribute('aria-selected', 'true');
    expect(screen.getAllByRole('tab').filter(t => t.tabIndex === 0)).toEqual([tab]);
    const panel = screen.getByRole('tabpanel', { name });
    expect(panel.id).toBe(tab.getAttribute('aria-controls'));
    expect(panel).toHaveAttribute('aria-labelledby', tab.id);
  };
  assertSelected('Narrative');
  for (const name of tabs.slice(1).map(t => t.label)) { await user.keyboard('{ArrowRight}'); assertSelected(name); }
  await user.keyboard('{ArrowRight}'); assertSelected('Narrative');
  await user.keyboard('{ArrowLeft}'); assertSelected('Logs');
  await user.keyboard('{Home}'); assertSelected('Narrative');
  await user.keyboard('{End}'); assertSelected('Logs');
  await user.tab(); expect(screen.getByRole('tabpanel', { name: 'Logs' })).toHaveFocus();
  await user.tab(); expect(screen.getByRole('button', { name: 'After panels' })).toHaveFocus();
  await user.tab({ shift: true }); await user.tab({ shift: true }); assertSelected('Logs');
});
it('keeps ids stable across click, keyboard and controlled selection changes', async () => {
  render(<Sections />); const user = userEvent.setup();
  const ids = screen.getAllByRole('tab').map(tab => [tab.id, tab.getAttribute('aria-controls')]);
  await user.click(screen.getByRole('tab', { name: 'World' }));
  expect(screen.getByRole('tabpanel', { name: 'World' })).toBeVisible();
  await user.keyboard('{ArrowLeft}'); expect(screen.getByRole('tab', { name: 'NPCs' })).toHaveFocus();
  expect(screen.getAllByRole('tab').map(tab => [tab.id, tab.getAttribute('aria-controls')])).toEqual(ids);
});
