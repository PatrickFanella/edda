import { render, screen } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import coverage from '../src/data/srd-coverage.json';
import { RulesModeStep } from '../src/components/start/RulesModeStep';

it('uses the checked server-coverage snapshot for the Crunch setup boundary', () => {
  render(<RulesModeStep selectedMode="crunch" onSelectMode={vi.fn()} onContinue={vi.fn()} />);

  const crunch = screen.getByRole('button', { name: /Crunch.*Action Surge/ });
  expect(crunch).toHaveTextContent(coverage.supported_scope);
  expect(crunch).toHaveTextContent('This is a partial slice; Edda does not claim Full Rules support.');
});
