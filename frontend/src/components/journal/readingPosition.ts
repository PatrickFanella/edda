// Reader preferences contain identities and relative offsets, never story prose.
const fallback = new Map<string, string>();
function read(key: string): string | null {
  try { return fallback.get(key) ?? localStorage.getItem(key) ?? null; }
  catch { return fallback.get(key) ?? null; }
}
function write(key: string, value: string) {
  fallback.set(key, value);
  try { localStorage.setItem(key, value); fallback.delete(key); } catch { /* In-memory retention remains available. */ }
}
export function readSelection(campaignId: string): { chapter: string; revision: string } | null {
  try {
    const value: unknown = JSON.parse(read(`edda:chapter-reader:${campaignId}`) ?? 'null');
    if (value && typeof value === 'object' && 'chapter' in value && typeof value.chapter === 'string' && 'revision' in value && typeof value.revision === 'string') return { chapter: value.chapter, revision: value.revision };
  } catch { /* Corrupt preferences must not block reading. */ }
  return null;
}
export function rememberSelection(campaignId: string, chapter: string, revision: string) {
  write(`edda:chapter-reader:${campaignId}`, JSON.stringify({ chapter, revision }));
}
function offsetKey(campaignId: string, chapter: string, revision: string) {
  return `edda:chapter-offset:${campaignId}:${chapter}:${revision}`;
}
export function readOffset(campaignId: string, chapter: string, revision: string): number {
  const offset = Number(read(offsetKey(campaignId, chapter, revision)));
  return Number.isFinite(offset) ? Math.min(1, Math.max(0, offset)) : 0;
}
export function rememberOffset(campaignId: string, chapter: string, revision: string, offset: number) {
  if (Number.isFinite(offset)) write(offsetKey(campaignId, chapter, revision), String(Math.min(1, Math.max(0, offset))));
}
