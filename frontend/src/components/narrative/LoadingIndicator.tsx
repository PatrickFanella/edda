import { cn } from '../../lib/cn';

interface LoadingIndicatorProps {
  readonly label?: string;
  readonly detail?: string;
  readonly className?: string;
}

export function LoadingIndicator({
  label = 'Game Master is thinking',
  detail = 'Composing the next beat…',
  className,
}: LoadingIndicatorProps) {
  return (
    <div
      role="status"
      aria-live="polite"
      className={cn('flex items-center gap-3 text-sm text-champagne', className)}
    >
      <div className="flex items-center gap-1.5" aria-hidden="true">
        <span className="h-2.5 w-2.5 bg-gold/90 [animation-delay:-0.3s] animate-bounce" />
        <span className="h-2.5 w-2.5 bg-gold/80 [animation-delay:-0.15s] animate-bounce" />
        <span className="h-2.5 w-2.5 bg-gold/70 animate-bounce" />
      </div>
      <div className="space-y-0.5">
        <p className="font-medium text-champagne">{label}</p>
        <p className="text-xs uppercase tracking-[0.2em] text-gold/65">{detail}</p>
      </div>
    </div>
  );
}
