import { useEffect, type ReactNode } from 'react';

import { UserMenu } from './UserMenu';

interface AppShellProps {
  readonly title: string;
  readonly description: string;
  readonly actions?: ReactNode;
  readonly userMenuActions?: ReactNode;
  readonly children: ReactNode;
  readonly variant?: 'default' | 'game';
}

export function AppShell({ title, description, actions, userMenuActions, children, variant = 'default' }: AppShellProps) {
  const isGame = variant === 'game';
  useEffect(() => {
    if (!isGame) return;
    const viewport = window.visualViewport;
    const root = document.documentElement;
    const previous = root.style.getPropertyValue('--game-viewport-height');
    const resize = () => {
      // Pinch zoom should retain native panning rather than compressing the UI.
      if (viewport && viewport.scale === 1) {
        root.style.setProperty('--game-viewport-height', `${viewport.height}px`);
      } else {
        root.style.removeProperty('--game-viewport-height');
      }
    };
    resize();
    viewport?.addEventListener('resize', resize);
    window.addEventListener('resize', resize);
    return () => {
      viewport?.removeEventListener('resize', resize);
      window.removeEventListener('resize', resize);
      if (previous) root.style.setProperty('--game-viewport-height', previous);
      else root.style.removeProperty('--game-viewport-height');
    };
  }, [isGame]);

  return (
    <main className={`${isGame ? 'game-app-shell bg-obsidian p-2 text-[17px]' : 'min-h-screen bg-obsidian px-6 py-16'} text-champagne`}>
      <div
        className={[
          'mx-auto flex w-full flex-col border-2 border-pewter/20 bg-charcoal',
          isGame
            ? 'game-app-frame gap-2 p-2'
            : 'deco-corners deco-pattern max-w-5xl gap-8 p-8',
        ].join(' ')}
      >
        <header className={`flex shrink-0 flex-col border-b border-pewter/20 ${isGame ? 'game-app-header gap-2 bg-obsidian/80 px-3 py-2' : 'gap-6 pb-6'}`}>
          <div className="flex items-center justify-between gap-3">
            <p className={`${isGame ? 'text-[1rem] leading-none' : 'text-sm'} font-heading font-semibold uppercase tracking-[0.24em] text-pewter`}>
              edda<span className={isGame ? 'hidden sm:inline' : undefined}> - game master</span>
            </p>
            <div className="flex items-center gap-2">
              {isGame && actions ? <div className="flex items-center gap-2">{actions}</div> : null}
              <UserMenu actions={userMenuActions} />
            </div>
          </div>
          <div className={`flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between ${isGame ? 'hidden' : ''}`}>
            <div className="space-y-2">
              <h1 className={`font-heading font-semibold uppercase tracking-[0.12em] text-champagne ${isGame ? 'text-xl sm:text-2xl' : 'text-3xl sm:text-4xl'}`}>{title}</h1>
              <p className={`${isGame ? 'max-w-4xl text-xs leading-5' : 'max-w-2xl text-sm leading-7 sm:text-base'} text-champagne/70`}>{description}</p>
            </div>
            {actions ? <div className="flex shrink-0 items-center gap-3">{actions}</div> : null}
          </div>
        </header>
        <section className={isGame ? 'game-app-content min-h-0 flex-1' : undefined}>{children}</section>
      </div>
    </main>
  );
}
