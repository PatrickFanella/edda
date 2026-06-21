import type { ReactNode } from 'react';

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

  return (
    <main className={`${isGame ? 'min-h-screen overflow-x-hidden bg-obsidian px-2 py-2 text-[17px] xl:h-screen xl:overflow-hidden' : 'min-h-screen bg-obsidian px-6 py-16'} text-champagne`}>
      <div
        className={[
          'mx-auto flex w-full flex-col border-2 border-pewter/20 bg-charcoal',
          isGame
            ? 'min-h-[calc(100vh-1rem)] gap-2 p-2 xl:h-[calc(100vh-1rem)] xl:max-w-[min(calc(100vw-1rem),calc((100vh-1rem)*16/9))] xl:overflow-hidden'
            : 'deco-corners deco-pattern max-w-5xl gap-8 p-8',
        ].join(' ')}
      >
        <header className={`flex shrink-0 flex-col border-b border-pewter/20 ${isGame ? 'gap-2 bg-obsidian/80 px-3 py-2' : 'gap-6 pb-6'}`}>
          <div className="flex items-center justify-between gap-3">
            <p className={`${isGame ? 'text-[1rem] leading-none' : 'text-sm'} font-heading font-semibold uppercase tracking-[0.24em] text-pewter`}>edda - game master</p>
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
        <section className={isGame ? 'min-h-0 flex-1' : undefined}>{children}</section>
      </div>
    </main>
  );
}
