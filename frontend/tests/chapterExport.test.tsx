import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, expect, test, vi } from 'vitest';
import { ChapterExportOptions } from '../src/components/export/ChapterExportOptions';
import { rememberSelection } from '../src/components/journal/readingPosition';
import { listChapters } from '../src/api/journal';
import { chapterExportFixture } from './fixtures/chapterExport';

vi.mock('../src/api/journal', () => ({ listChapters: vi.fn() }));
const first = chapterExportFixture(1), second = chapterExportFixture(2);
beforeEach(() => vi.mocked(listChapters).mockResolvedValue([second, first]));
function setup(campaignId = 'alpha', disabled = false) {
  const onDownload = vi.fn();
  const view = render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><ChapterExportOptions campaignId={campaignId} disabled={disabled} onDownload={onDownload} /></QueryClientProvider>);
  return { ...view, onDownload, user: userEvent.setup() };
}

test('exports only checked chapters, defaulting to the saved reader revision and latest completed elsewhere', async () => {
  rememberSelection('alpha', first.id, first.revisions[0].id);
  const { user, onDownload } = setup();
  await user.click(screen.getByText('Saved Chapters'));
  const download = await screen.findByRole('button', { name: 'Download selected chapters' });
  expect(download).toBeDisabled();
  const checks = screen.getAllByRole('checkbox');
  expect(checks).toHaveLength(2);
  await user.click(checks[0]);
  expect(screen.getByRole('combobox', { name: /Revision for Chapter 1/ })).toHaveValue(first.revisions[0].id);
  await user.click(checks[1]);
  const revision = screen.getByRole('combobox', { name: /Revision for Chapter 2/ });
  expect(revision).toHaveValue(second.revisions[1].id);
  expect(revision.querySelectorAll('option')).toHaveLength(2);
  await user.click(download);
  expect(onDownload).toHaveBeenLastCalledWith([first.revisions[0].id, second.revisions[1].id], 'markdown');
  await user.click(checks[0]);
  await user.selectOptions(revision, second.revisions[0].id);
  await user.selectOptions(screen.getByLabelText('Chapter file format'), 'json');
  await user.click(download);
  expect(onDownload).toHaveBeenLastCalledWith([second.revisions[0].id], 'json');
});

test('offers explicit recovery when chapter loading fails', async () => {
  vi.mocked(listChapters).mockRejectedValueOnce(new Error('offline'));
  const { user } = setup();
  await user.click(screen.getByText('Saved Chapters'));
  expect(await screen.findByRole('alert')).toHaveTextContent('Could not load saved chapters');
  await user.click(screen.getByRole('button', { name: 'Retry loading chapters' }));
  await waitFor(() => expect(screen.getAllByRole('checkbox')).toHaveLength(2));
});

test('campaigns without completed chapters cannot export', async () => {
  vi.mocked(listChapters).mockResolvedValue([{ ...first, revisions: first.revisions.slice(2) }]);
  const { user, onDownload } = setup();
  await user.click(screen.getByText('Saved Chapters'));
  expect(await screen.findByText(/No completed saved chapters/)).toBeVisible();
  expect(screen.queryByRole('button', { name: 'Download selected chapters' })).toBeNull();
  expect(onDownload).not.toHaveBeenCalled();
});
