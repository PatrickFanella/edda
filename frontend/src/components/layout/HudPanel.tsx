import type { ReactNode } from 'react';

export type HudPanelAccent =
  | 'exploration'
  | 'dialogue'
  | 'combat'
  | 'vitals'
  | 'objective'
  | 'inventory'
  | 'scene'
  | 'setup'
  | 'replay'
  | 'auth'
  | 'error'
  | 'loading'
  | 'empty';

const panelClassByAccent: Record<HudPanelAccent, string> = {
  exploration: 'game-hud-panel-exploration',
  dialogue: 'game-hud-panel-dialogue',
  combat: 'game-hud-panel-combat',
  vitals: 'game-hud-panel-vitals',
  objective: 'game-hud-panel-objective',
  inventory: 'game-hud-panel-inventory',
  scene: 'game-hud-panel-scene',
  setup: 'game-hud-panel-setup',
  replay: 'game-hud-panel-replay',
  auth: 'game-hud-panel-auth',
  error: 'game-hud-panel-error',
  loading: 'game-hud-panel-loading',
  empty: 'game-hud-panel-empty',
};

const labelClassByAccent: Record<HudPanelAccent, string> = {
  exploration: 'hud-label-gold',
  dialogue: 'hud-label-sapphire',
  combat: 'hud-label-ruby',
  vitals: 'hud-label-jade',
  objective: 'hud-label-sapphire',
  inventory: 'hud-label-gold',
  scene: 'hud-label-pewter',
  setup: 'hud-label-sapphire',
  replay: 'hud-label-pewter',
  auth: 'hud-label-gold',
  error: 'hud-label-ruby',
  loading: 'hud-label-sapphire',
  empty: 'hud-label-pewter',
};

interface HudPanelProps {
  readonly title?: string;
  readonly accent: HudPanelAccent;
  readonly actions?: ReactNode;
  readonly children: ReactNode;
  readonly className?: string;
  readonly bodyClassName?: string;
}

export function HudPanel({ title, accent, actions, children, className, bodyClassName }: HudPanelProps) {
  const panelClass = panelClassByAccent[accent];
  const labelClass = labelClassByAccent[accent];
  const hasHeader = Boolean(title || actions);

  return (
    <section className={`game-hud-panel ${panelClass} border-2 bg-obsidian/65 p-3.5 ${className ?? ''}`.trim()}>
      {hasHeader ? (
        <header className="flex items-start justify-between gap-3">
          {title ? <h3 className={`font-heading hud-label ${labelClass}`}>{title}</h3> : <span />}
          {actions ? <div>{actions}</div> : null}
        </header>
      ) : null}
      <div className={`${hasHeader ? 'mt-2.5' : ''} ${bodyClassName ?? ''}`.trim()}>{children}</div>
    </section>
  );
}
