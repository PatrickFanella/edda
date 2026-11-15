import { useState } from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, it, vi } from 'vitest';
import { ConfirmationPanel } from '../src/components/start/ConfirmationPanel';
import { CharacterGuidedForm, type CharacterGuidedFormData } from '../src/components/start/CharacterGuidedForm';

const blank: CharacterGuidedFormData = { name: '', race: '', class: '', background: '', alignment: '' };

function Harness({ submit }: { submit: (value: CharacterGuidedFormData) => void }) {
  const [value, setValue] = useState(blank);
  return <CharacterGuidedForm value={value} onChange={setValue} onSubmit={submit} requiresSupportedMartialBuild />;
}

it('keeps guided inputs editable and blocks an incomplete Crunch build before confirmation', async () => {
  const submit = vi.fn();
  render(<Harness submit={submit} />);
  const user = userEvent.setup();
  await user.type(screen.getByLabelText('Character name'), 'Mira');
  await user.selectOptions(screen.getByLabelText('Race'), 'Elf');
  await user.selectOptions(screen.getByLabelText('Class'), 'Fighter');
  await user.selectOptions(screen.getByLabelText('Background'), 'Soldier');
  await user.selectOptions(screen.getByLabelText('Alignment'), 'Lawful Good');
  await user.click(screen.getByRole('button', { name: 'Continue to confirmation' }));
  expect(submit).not.toHaveBeenCalled();
  expect(screen.getByText('Crunch currently supports the Human Soldier Fighter build.')).toBeVisible();
  expect(screen.getByLabelText('Character name')).toHaveValue('Mira');
  await user.selectOptions(screen.getByLabelText('Race'), 'Human');
  await user.click(screen.getByRole('button', { name: 'Continue to confirmation' }));
  expect(submit).toHaveBeenCalledWith({ name: 'Mira', race: 'Human', class: 'Fighter', background: 'Soldier', alignment: 'Lawful Good' });
});


it('leaves Back usable and disables Begin when confirmation still requires choices',()=>{
 render(<ConfirmationPanel campaignName="Test" campaignProfile={{genre:'fantasy',tone:'hopeful',themes:[]}} characterProfile={{name:'Mira'}} onBegin={vi.fn()} onBack={vi.fn()} canBegin={false} incompleteMessage="Choose a supported origin." />);
 expect(screen.getByRole('button',{name:'Begin adventure'})).toBeDisabled();
 expect(screen.getByRole('button',{name:'Back'})).toBeEnabled();
 expect(screen.getByText('Choose a supported origin.')).toBeVisible();
});
