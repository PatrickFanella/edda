import { useCallback, useEffect, useRef } from 'react';

import { cn } from '../../lib/cn';

interface ConfirmationDialogProps {
  readonly open: boolean;
  readonly title: string;
  readonly message: string;
  readonly confirmLabel?: string;
  readonly cancelLabel?: string;
  readonly destructive?: boolean;
  readonly onConfirm: () => void;
  readonly onCancel: () => void;
}

export function ConfirmationDialog({
  open,
  title,
  message,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  destructive = false,
  onConfirm,
  onCancel,
}: ConfirmationDialogProps) {
  const dialogRef = useRef<HTMLDialogElement | null>(null);

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
    onCancel();
  }, [onCancel]);

  if (!open) return null;

  return (
    <dialog
      ref={dialogRef}
      onCancel={handleCancel}
      className="m-auto max-w-md border-2 border-gold/30 bg-charcoal p-0 text-champagne backdrop:bg-obsidian/80"
    >
      <div className="space-y-4 p-6">
        <h2 className="font-heading text-lg font-semibold uppercase tracking-wide">{title}</h2>
        <p className="text-sm leading-6 text-pewter">{message}</p>
      </div>
      <div className="flex justify-end gap-3 border-t border-gold/15 px-6 py-4">
        <button
          type="button"
          onClick={onCancel}
          className="border border-pewter/30 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-champagne/70 transition hover:border-pewter hover:text-champagne"
        >
          {cancelLabel}
        </button>
        <button
          type="button"
          onClick={onConfirm}
          className={cn(
            'px-4 py-2 text-sm font-semibold uppercase tracking-wide transition',
            destructive
              ? 'border border-ruby/40 bg-ruby/10 text-ruby hover:bg-ruby hover:text-champagne'
              : 'border border-gold/40 bg-gold/10 text-gold hover:bg-gold hover:text-obsidian',
          )}
        >
          {confirmLabel}
        </button>
      </div>
    </dialog>
  );
}
