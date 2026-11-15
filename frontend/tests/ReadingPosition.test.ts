import { afterEach, expect, it, vi } from 'vitest';
import { readOffset, readSelection, rememberOffset, rememberSelection } from '../src/components/journal/readingPosition';

afterEach(() => { vi.restoreAllMocks(); });

it('retains newer chapter preferences when persisted values exist but subsequent writes fail', () => {
  const campaign = 'storage-write-failure';
  rememberSelection(campaign, 'first', 'revision-1');
  rememberOffset(campaign, 'first', 'revision-1', 0.2);
  const write = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => { throw new Error('storage full'); });

  rememberSelection(campaign, 'second', 'revision-2');
  rememberOffset(campaign, 'first', 'revision-1', 0.7);

  expect(readSelection(campaign)).toEqual({ chapter: 'second', revision: 'revision-2' });
  expect(readOffset(campaign, 'first', 'revision-1')).toBe(0.7);

  // Successful writes remove the fallback so future persisted updates can win.
  write.mockRestore();
  rememberSelection(campaign, 'third', 'revision-3');
  rememberOffset(campaign, 'first', 'revision-1', 0.8);
  localStorage.setItem(`edda:chapter-reader:${campaign}`, JSON.stringify({ chapter: 'fourth', revision: 'revision-4' }));
  localStorage.setItem(`edda:chapter-offset:${campaign}:first:revision-1`, '0.9');
  expect(readSelection(campaign)).toEqual({ chapter: 'fourth', revision: 'revision-4' });
  expect(readOffset(campaign, 'first', 'revision-1')).toBe(0.9);
});
