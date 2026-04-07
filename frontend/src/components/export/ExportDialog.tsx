import { useCallback, useEffect, useRef, useState } from 'react';

import { cn } from '../../lib/cn';

type ExportFormat = 'json' | 'transcript' | 'character';

interface ExportDialogProps {
  readonly open: boolean;
  readonly campaignId: string;
  readonly onClose: () => void;
}

const API_BASE = import.meta.env.VITE_API_BASE ?? '/api/v1';

async function downloadExport(campaignId: string, format: ExportFormat) {
  const token = localStorage.getItem('gm_token');
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}/campaigns/${campaignId}/export/${format}`, { headers });
  if (!res.ok) {
    throw new Error(`Export failed: ${res.statusText}`);
  }

  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `campaign-${format}.${format === 'json' ? 'json' : 'md'}`;
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
    async (format: ExportFormat) => {
      setLoadingFormat(format);
      setError(null);
      try {
        await downloadExport(campaignId, format);
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
      className="m-auto max-w-lg border-2 border-gold/30 bg-charcoal p-0 text-champagne backdrop:bg-obsidian/80"
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

        {error ? <p className="text-xs text-ruby">{error}</p> : null}
      </div>

      <div className="flex justify-end border-t border-gold/15 px-6 py-4">
        <button
          type="button"
          onClick={onClose}
          className="border border-pewter/30 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-champagne/70 transition hover:border-pewter hover:text-champagne"
        >
          Close
        </button>
      </div>
    </dialog>
  );
}
