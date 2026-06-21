interface TabBarTab<TTab extends string> {
  readonly key: TTab;
  readonly label: string;
  readonly activeTone?: string;
  readonly hoverTone?: string;
  readonly badge?: boolean;
}

interface TabBarProps<TTab extends string> {
  readonly tabs: readonly TabBarTab<TTab>[];
  readonly activeTab: TTab;
  readonly onChange: (tab: TTab) => void;
}

const DEFAULT_ACTIVE = 'bg-ruby text-champagne';
const DEFAULT_HOVER = 'border border-pewter/20 bg-charcoal text-champagne/70 hover:border-pewter hover:text-pewter hover:bg-pewter/5';

export function TabBar<TTab extends string>({ tabs, activeTab, onChange }: TabBarProps<TTab>) {
  return (
    <div
      role="tablist"
      aria-label="Campaign play sections"
      className="mx-auto flex max-h-12 w-full flex-nowrap justify-between gap-2 overflow-x-auto overflow-y-hidden bg-charcoal p-1"
    >
      {tabs.map((tab) => {
        const isActive = tab.key === activeTab;

        return (
          <button
            key={tab.key}
            type="button"
            role="tab"
            aria-selected={isActive}
            onClick={() => onChange(tab.key)}
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
