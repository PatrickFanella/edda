import { useLayoutEffect, useRef } from 'react';

interface TabBarTab<TTab extends string> {
  readonly key: TTab;
  readonly label: string;
  readonly activeTone?: string;
  readonly hoverTone?: string;
  readonly badge?: boolean;
}

interface TabBarProps<TTab extends string> {
  readonly idPrefix: string;
  readonly tabs: readonly TabBarTab<TTab>[];
  readonly activeTab: TTab;
  readonly onChange: (tab: TTab) => void;
}

const DEFAULT_ACTIVE = 'bg-ruby text-champagne';
const DEFAULT_HOVER = 'border border-pewter/20 bg-charcoal text-champagne/70 hover:border-pewter hover:text-pewter hover:bg-pewter/5';

export function TabBar<TTab extends string>({ idPrefix, tabs, activeTab, onChange }: TabBarProps<TTab>) {
  const buttons = useRef(new Map<TTab, HTMLButtonElement>());
  const focusRequest = useRef<TTab | null>(null);

  // A newly mounted panel may autofocus its content. Tab activation must keep
  // keyboard focus on the selected tab until the player chooses to leave it.
  useLayoutEffect(() => {
    if (focusRequest.current === activeTab) {
      buttons.current.get(activeTab)?.focus();
    }
    focusRequest.current = null;
  }, [activeTab]);

  return (
    <div
      role="tablist"
      aria-label="Campaign play sections"
      aria-orientation="horizontal"
      className="mx-auto flex max-h-12 w-full flex-nowrap justify-between gap-2 overflow-x-auto overflow-y-hidden bg-charcoal p-1"
    >
      {tabs.map((tab) => {
        const isActive = tab.key === activeTab;

        return (
          <button
            key={tab.key}
            type="button"
            role="tab"
            id={`${idPrefix}-tab-${tab.key}`}
            aria-controls={`${idPrefix}-panel-${tab.key}`}
            tabIndex={isActive ? 0 : -1}
            ref={(button) => {
              if (button) buttons.current.set(tab.key, button);
              else buttons.current.delete(tab.key);
            }}
            onKeyDown={(event) => {
              const index = tabs.findIndex((candidate) => candidate.key === tab.key);
              let next: number;
              switch (event.key) {
                case 'ArrowRight': next = (index + 1) % tabs.length; break;
                case 'ArrowLeft': next = (index - 1 + tabs.length) % tabs.length; break;
                case 'Home': next = 0; break;
                case 'End': next = tabs.length - 1; break;
                default: return;
              }
              event.preventDefault();
              const target = tabs[next].key;
              buttons.current.get(target)?.focus();
              focusRequest.current = target;
              onChange(target);
            }}
            aria-selected={isActive}
            onClick={() => {
              focusRequest.current = tab.key;
              onChange(tab.key);
            }}
            className={[
              'hud-tab-button relative shrink-0 px-4 text-center text-[0.75rem] font-semibold uppercase tracking-[0.12em] transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-pewter focus:ring-offset-2 focus:ring-offset-obsidian',
              isActive
                ? (tab.activeTone ?? DEFAULT_ACTIVE)
                : (tab.hoverTone ?? DEFAULT_HOVER),
            ].join(' ')}
          >
            {tab.label}
            {tab.badge && !isActive ? (
              <span className="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full border border-charcoal bg-jade animate-pulse" aria-label="New discovery" />
            ) : null}
          </button>
        );
      })}
    </div>
  );
}
