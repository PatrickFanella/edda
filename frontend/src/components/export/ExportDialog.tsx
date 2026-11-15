import { useCallback, useEffect, useRef, useState } from 'react';

import { apiFetchBlob } from '../../api/backend';
import { cn } from '../../lib/cn';
import { ChapterExportOptions } from './ChapterExportOptions';

type ExportFormat = 'json' | 'transcript' | 'character' | 'chapters';

interface ExportDialogProps {
  readonly open: boolean;
  readonly campaignId: string;
  readonly onClose: () => void;
}

async function downloadExport(campaignId: string, format: ExportFormat, query = '') {
  const blob = await apiFetchBlob(`/campaigns/${campaignId}/export/${format}${query ? `?${query}` : ''}`, {
    method: 'GET',
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `campaign-${format}.${format === 'json' || new URLSearchParams(query).get('format') === 'json' ? 'json' : 'md'}`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

const exportOptions: { format: ExportFormat; label: string; description: string }[] = [
  { format: 'json', label: 'JSON Data', description: 'Full campaign data including all NPCs, locations, quests, and session history.' },
  { format: 'transcript', label: 'Session Transcript', description: 'Markdown document of all player/GM exchanges.' },
  { format: 'character', label: 'Character Sheet', description: 'Markdown character sheet with stats, abilities, and inventory.' },
];

export function ExportDialog({ open, campaignId, onClose }: ExportDialogProps) {
  const dialogRef = useRef<HTMLDialogElement | null>(null);
  const [loadingFormat, setLoadingFormat] = useState<ExportFormat | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const el = dialogRef.current;
    if (!el) return;
    if (open && !el.open) {
      el.showModal();
    } else if (!open && el.open) {
      el.close();
    }
  }, [open]);

  const handleCancel = useCallback(() => {
    onClose();
  }, [onClose]);

  const handleDownload = useCallback(
    async (format: ExportFormat, query = '') => {
      setLoadingFormat(format);
      setError(null);
      try {
        await downloadExport(campaignId, format, query);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Export failed');
      } finally {
        setLoadingFormat(null);
      }
    },
    [campaignId],
  );

  if (!open) return null;

  return (
    <dialog
      ref={dialogRef}
      onCancel={handleCancel}
      className="m-auto max-h-[90dvh] max-w-lg overflow-y-auto border-2 border-gold/30 bg-charcoal p-0 text-champagne backdrop:bg-obsidian/80"
    >
      <div className="space-y-4 p-6">
        <h2 className="font-heading text-lg font-semibold uppercase tracking-wide">Export Campaign</h2>
        <p className="text-sm leading-6 text-pewter">Choose a format to download your campaign data.</p>

        <div className="space-y-3">
          {exportOptions.map((opt) => {
            const isLoading = loadingFormat === opt.format;
            const isDisabled = loadingFormat !== null;
            return (
              <button
                key={opt.format}
                type="button"
                disabled={isDisabled}
                onClick={() => void handleDownload(opt.format)}
                className={cn(
                  'w-full border px-4 py-3 text-left transition',
                  isDisabled && !isLoading
                    ? 'border-pewter/20 opacity-50 cursor-not-allowed'
                    : 'border-gold/20 hover:border-gold hover:bg-gold/5 cursor-pointer',
                )}
              >
                <p className="text-sm font-semibold uppercase tracking-wide text-champagne">
                  {isLoading ? `Exporting ${opt.label}...` : opt.label}
                </p>
                <p className="mt-1 text-xs text-pewter">{opt.description}</p>
              </button>
            );
          })}
        </div>

        <ChapterExportOptions key={campaignId} campaignId={campaignId} disabled={loadingFormat !== null} onDownload={(ids, format) => {
          const query = new URLSearchParams({ format });
          ids.forEach(id => query.append('revision_id', id));
          void handleDownload('chapters', query.toString());
        }} />

        {error ? <p role="alert" className="text-xs text-ruby">{error}</p> : null}
      </div>

      <div className="flex justify-end border-t border-gold/15 px-6 py-4">
        <button
          type="button"
          onClick={onClose}
          className="min-h-11 min-w-11 border border-pewter/30 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-champagne/70 transition hover:border-pewter hover:text-champagne"
        >
          Close
        </button>
      </div>
    </dialog>
  );
}
