interface TabBarTab<TTab extends string> {
  readonly key: TTab;
  readonly label: string;
  readonly activeTone?: string;
  readonly hoverTone?: string;
}

interface TabBarProps<TTab extends string> {
  readonly tabs: readonly TabBarTab<TTab>[];
  readonly activeTab: TTab;
  readonly onChange: (tab: TTab) => void;
}

const DEFAULT_ACTIVE = 'bg-ruby text-champagne';
const DEFAULT_HOVER = 'border border-gold/20 bg-charcoal text-champagne/70 hover:border-gold hover:text-gold hover:bg-gold/5';

export function TabBar<TTab extends string>({ tabs, activeTab, onChange }: TabBarProps<TTab>) {
  return (
    <div
      role="tablist"
      aria-label="Campaign play sections"
      className="flex flex-wrap gap-2 border-2 border-gold/20 bg-charcoal p-2"
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
              'px-4 py-2 text-sm font-semibold uppercase tracking-[0.15em] transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-gold focus:ring-offset-2 focus:ring-offset-obsidian',
              isActive
                ? (tab.activeTone ?? DEFAULT_ACTIVE)
                : (tab.hoverTone ?? DEFAULT_HOVER),
            ].join(' ')}
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
