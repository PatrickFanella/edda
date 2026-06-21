# HUD Style Rollout Implementation Plan

> **For agentic workers:** Execute this plan task-by-task. Recommended path:
> dispatch a fresh subagent per task, review each result with `review-quality`,
> then continue. For complex multi-agent splits, use
> `parallel-feature-development`, `team-composition-patterns`, and
> `team-communication-protocols`. Steps use checkbox (`- [ ]`) syntax for
> tracking.

**Goal:** Adapt every frontend page and play submenu to the new Edda RPG HUD style, including consistent exploration/dialogue/combat/replay/setup states and corrected text/button baseline alignment.

**Architecture:** Promote the campaign play HUD work into shared semantic primitives first, then migrate pages/panels in small vertical slices. Avoid one-off Tailwind button fixes by centralizing baseline, padding, panel, tab, badge, and state styling in semantic HUD utilities/components.

**Tech Stack:** React 19, TypeScript, Tailwind utility classes, custom CSS in `frontend/src/index.css`, Vite, `pnpm build` for verification.

---

## Current State

The campaign play route already has the target direction: fixed widescreen HUD shell, headline campaign title, top mode/day bar, transcript-first main view, right state sidebar, compact bottom tabs, semantic colors, and baseline-corrected labels/badges.

The rest of the frontend still mixes older AppShell/cards/buttons with newer HUD utilities. Several submenu controls still use ad hoc `px-* py-* leading-*` combinations, causing the recurring optical issue where labels appear too high or bottom padding appears larger.

## Design Rules for the Rollout

- **Gold is accent, not default chrome.** Use pewter for neutral frame/chrome, jade for vitals/success, sapphire for facts/objectives/navigation, ruby for danger/combat, gold for inventory/reward/brand moments.
- **Buttons use semantic classes.** Prefer `hud-btn`, `hud-text-button`, `hud-tab-button`, `hud-icon-btn`, and new variants over raw `py-*` per component.
- **Text labels need baseline compensation.** Uppercase HUD labels and badges should use shared baseline classes instead of per-element nudges.
- **Pages keep their purpose.** Do not make login/register/create/list pages look identical to campaign play; adapt them to the same HUD language with page-appropriate density.
- **States are explicit.** Surfaces should declare visual state: exploration, dialogue, combat, replay, setup, auth, danger, empty, loading, error.
- **Small commits.** Commit after each task or page group.

---

## File Structure and Ownership

### Shared Styling / Primitives

- Modify: `frontend/src/index.css`
  - Owns semantic HUD utility classes.
  - Add missing state variants: setup, replay, auth, empty, loading, error.
  - Add final baseline utilities for menu/filter/chip/mini buttons.
- Modify: `frontend/src/components/layout/AppShell.tsx`
  - Extend `variant` support beyond `default | game` if needed, or add a `tone` prop for non-play HUD pages.
  - Preserve existing default behavior while adding HUD page framing.
- Modify: `frontend/src/components/layout/TabBar.tsx`
  - Keep compact intrinsic tab buttons with spacing grown by the container, not button width.
  - Use shared baseline utilities only.
- Create: `frontend/src/components/layout/HudPanel.tsx`
  - Shared panel wrapper for title, accent, state, optional actions, and empty/error/loading content.
- Create: `frontend/src/components/layout/HudButton.tsx`
  - Optional thin wrapper for common button/link variants if CSS-only usage becomes repetitive.
  - Keep this shallow; do not abstract every button if class utilities are enough.

### Routed Pages

- Modify: `frontend/src/pages/CampaignPlayPage.tsx`
  - Extract local HUD helpers once shared primitives exist.
  - Keep current layout intact.
- Modify: `frontend/src/pages/CampaignListPage.tsx`
  - Convert dashboard cards/actions to HUD cards/buttons.
- Modify: `frontend/src/pages/CampaignCreatePage.tsx`
  - Convert startup wizard shell and controls to setup-state HUD.
- Modify: `frontend/src/pages/ReplayPage.tsx`
  - Convert replay layout to replay-state HUD.
- Modify: `frontend/src/pages/LoginPage.tsx`
  - Convert auth card to auth-state HUD without using the full play shell.
- Modify: `frontend/src/pages/RegisterPage.tsx`
  - Match login auth-state HUD.

### Campaign Play Submenus / Panels

- Modify: `frontend/src/components/world/WorldPanel.tsx`
  - World subtabs: map, facts, codex, relationships.
- Modify: `frontend/src/components/map/MapPanel.tsx`
  - Map module + selected location close button.
- Modify: `frontend/src/components/facts/FactsPanel.tsx`
  - Facts panel + empty/loading/error states.
- Modify: `frontend/src/components/codex/CodexPanel.tsx`
  - Codex internal tabs for languages/cultures/beliefs/economies.
- Modify: `frontend/src/components/quests/QuestPanel.tsx`
  - Type/status filters + clear filters button.
- Modify: `frontend/src/components/inventory/InventoryPanel.tsx`
  - Inventory header, item cards, empty state.
- Modify: `frontend/src/components/character/CharacterSheet.tsx`
  - Character cards/meters/stat panels.
- Modify: `frontend/src/components/npcs/NPCPanel.tsx`
  - NPC expandable rows and panel messages.
- Modify: `frontend/src/components/journal/JournalPanel.tsx`
  - Section toggles, summarize/save/delete buttons.
- Modify: `frontend/src/components/logs/LogPanel.tsx` if present, or current log component path if different.
- Modify: `frontend/src/components/combat/CombatView.tsx`
  - Combat state panel composition.
- Modify: `frontend/src/components/combat/CombatActionBar.tsx`
  - Combat action buttons.

### Replay Components

- Modify: `frontend/src/components/replay/ReplayControls.tsx`
- Modify: `frontend/src/components/replay/ReplayNarrative.tsx`
- Modify: `frontend/src/components/replay/ReplaySidebar.tsx`
- Modify: `frontend/src/components/replay/ReplayTimeline.tsx`

### Startup Wizard Components

- Modify: `frontend/src/components/start/MethodPicker.tsx`
- Modify: `frontend/src/components/start/ProposalPicker.tsx`
- Modify: `frontend/src/components/start/RulesModeStep.tsx`
- Modify: `frontend/src/components/start/ChatStep.tsx`
- Modify: `frontend/src/components/start/ConfirmationPanel.tsx`
- Modify: `frontend/src/components/start/CharacterGuidedForm.tsx`
- Modify: `frontend/src/components/start/CampaignAttributesForm.tsx`

---

## Task 1: Finalize Shared HUD Tokens and Baseline Utilities

**Files:**
- Modify: `frontend/src/index.css`
- Modify: `frontend/src/components/layout/TabBar.tsx`
- Modify: `frontend/src/components/layout/UserMenu.tsx`
- Modify: `frontend/src/components/audio/AudioControls.tsx`

- [ ] **Step 1: Audit existing HUD classes**

Run:

```bash
rg "hud-|game-hud-|py-|leading-" frontend/src/index.css frontend/src/components/layout frontend/src/components/audio
```

Expected: identify current baseline classes and any remaining ad hoc button padding in shared controls.

- [ ] **Step 2: Add/confirm semantic utility classes**

Ensure `frontend/src/index.css` includes these utility groups:

```css
.hud-btn { display: inline-flex; align-items: center; justify-content: center; border: 2px solid; padding: 0.62rem 1rem 0.38rem; line-height: 1; text-transform: uppercase; }
.hud-text-button { line-height: 1; padding-top: 0.62rem; padding-bottom: 0.38rem; }
.hud-tab-button { line-height: 1; padding-top: 0.5rem; padding-bottom: 0.28rem; }
.hud-icon-btn { height: 2rem; width: 2rem; padding: 0; }
.hud-baseline-badge { line-height: 1; padding-top: 0.32rem; padding-bottom: 0.16rem; }
.hud-label { font-family: 'Marcellus', serif; font-size: 0.75rem; font-weight: 600; letter-spacing: 0.2em; text-transform: uppercase; }
```

Add missing state classes:

```css
.game-hud-panel-setup { color: #5B8FB9; border-color: rgba(91, 143, 185, 0.28); }
.game-hud-panel-replay { color: #888888; border-color: rgba(136, 136, 136, 0.28); }
.game-hud-panel-auth { color: #D4AF37; border-color: rgba(212, 175, 55, 0.24); }
.game-hud-panel-error { color: #AB2346; border-color: rgba(171, 35, 70, 0.42); }
.game-hud-panel-loading { color: #5B8FB9; border-color: rgba(91, 143, 185, 0.24); }
.game-hud-panel-empty { color: #888888; border-color: rgba(136, 136, 136, 0.22); }
```

- [ ] **Step 3: Remove shared-control one-offs**

Replace ad hoc baseline fixes in shared controls with semantic classes:

```tsx
// TabBar buttons
className="hud-tab-button relative shrink-0 px-4 text-center text-[0.75rem] font-semibold uppercase tracking-[0.12em] ..."

// Text HUD buttons
className="hud-btn hud-btn-primary"

// Icon HUD buttons
className="hud-btn hud-btn-primary hud-icon-btn"
```

- [ ] **Step 4: Verify build**

Run:

```bash
pnpm --dir frontend build
```

Expected: build passes; Vite chunk warning may remain.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/index.css frontend/src/components/layout/TabBar.tsx frontend/src/components/layout/UserMenu.tsx frontend/src/components/audio/AudioControls.tsx
git commit -m "style: finalize HUD control utilities"
```

---

## Task 2: Extract Shared HUD Panel Primitive

**Files:**
- Create: `frontend/src/components/layout/HudPanel.tsx`
- Modify: `frontend/src/pages/CampaignPlayPage.tsx`

- [ ] **Step 1: Create `HudPanel`**

Create `frontend/src/components/layout/HudPanel.tsx`:

```tsx
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

const PANEL_CLASS: Record<HudPanelAccent, string> = {
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

const LABEL_CLASS: Record<HudPanelAccent, string> = {
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

export function HudPanel({ title, accent, actions, children, className = '', bodyClassName = '' }: HudPanelProps) {
  return (
    <section className={`game-hud-panel ${PANEL_CLASS[accent]} border-2 bg-obsidian/65 p-3.5 ${className}`}>
      {title || actions ? (
        <header className="flex items-center justify-between gap-3">
          {title ? <h3 className={`hud-label ${LABEL_CLASS[accent]}`}>{title}</h3> : <span />}
          {actions}
        </header>
      ) : null}
      <div className={`${title || actions ? 'mt-2.5' : ''} ${bodyClassName}`}>{children}</div>
    </section>
  );
}
```

- [ ] **Step 2: Replace local `HudSection` in play page**

In `frontend/src/pages/CampaignPlayPage.tsx`, import `HudPanel` and replace local `HudSection` usage.

```tsx
import { HudPanel } from '../components/layout/HudPanel';
```

Replace:

```tsx
<HudSection title="Vitals" accent="vitals">...</HudSection>
```

with:

```tsx
<HudPanel title="Vitals" accent="vitals">...</HudPanel>
```

Delete the local `HudSection` function after all references are replaced.

- [ ] **Step 3: Verify build**

```bash
pnpm --dir frontend build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/layout/HudPanel.tsx frontend/src/pages/CampaignPlayPage.tsx
git commit -m "refactor: extract shared HUD panel"
```

---

## Task 3: Fix Submenu Button Baselines in Campaign Play Panels

**Files:**
- Modify: `frontend/src/components/world/WorldPanel.tsx`
- Modify: `frontend/src/components/map/MapPanel.tsx`
- Modify: `frontend/src/components/codex/CodexPanel.tsx`
- Modify: `frontend/src/components/quests/QuestPanel.tsx`
- Modify: `frontend/src/components/journal/JournalPanel.tsx`
- Modify: `frontend/src/components/combat/CombatActionBar.tsx`

- [ ] **Step 1: Replace internal tab/filter buttons with `hud-tab-button`**

Apply this pattern to world tabs, codex tabs, quest filters, replay speed tabs, and similar compact menu controls:

```tsx
className={cn(
  'hud-tab-button border px-4 text-[0.75rem] font-semibold uppercase tracking-[0.12em] transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-pewter focus:ring-offset-2 focus:ring-offset-obsidian',
  isActive ? 'bg-ruby text-champagne' : 'border-pewter/20 bg-charcoal text-champagne/70 hover:border-pewter hover:text-pewter',
)}
```

- [ ] **Step 2: Replace action buttons with `hud-btn`**

Apply this pattern to `Close`, `Clear filters`, `Summarize`, `Save entry`, combat send/cancel, and destructive controls:

```tsx
className="hud-btn hud-btn-secondary"
```

Use danger variant for destructive actions:

```tsx
className="hud-btn hud-btn-danger"
```

- [ ] **Step 3: Replace badge-like submenu labels with `hud-baseline-badge`**

Apply this pattern to small uppercase badges inside cards:

```tsx
className="hud-baseline-badge inline-flex rounded-sm border border-pewter/15 bg-pewter/5 px-3 text-[11px] font-medium uppercase tracking-[0.18em] text-champagne/70"
```

- [ ] **Step 4: Verify with targeted grep**

Run:

```bash
rg "py-[0-9]|leading-" frontend/src/components/world frontend/src/components/map frontend/src/components/codex frontend/src/components/quests frontend/src/components/journal frontend/src/components/combat
```

Expected: remaining `py-*` should be for layout containers/inputs, not uppercase buttons/badges.

- [ ] **Step 5: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/components/world frontend/src/components/map frontend/src/components/codex frontend/src/components/quests frontend/src/components/journal frontend/src/components/combat
git commit -m "style: align campaign submenu controls"
```

---

## Task 4: Adapt World/Facts/Codex/Map Panels to HUD Modules

**Files:**
- Modify: `frontend/src/components/world/WorldPanel.tsx`
- Modify: `frontend/src/components/map/MapPanel.tsx`
- Modify: `frontend/src/components/facts/FactsPanel.tsx`
- Modify: `frontend/src/components/codex/CodexPanel.tsx`

- [ ] **Step 1: Wrap each panel in `HudPanel`**

Use semantic accents:

```tsx
<HudPanel title="World map" accent="exploration">...</HudPanel>
<HudPanel title="World facts" accent="objective">...</HudPanel>
<HudPanel title="Codex" accent="scene">...</HudPanel>
```

- [ ] **Step 2: Convert loading/error/empty states to HUD panel states**

Use:

```tsx
<HudPanel title="Loading" accent="loading">...</HudPanel>
<HudPanel title="Unavailable" accent="error">...</HudPanel>
<HudPanel title="No records" accent="empty">...</HudPanel>
```

- [ ] **Step 3: Preserve existing data behavior**

Do not change API calls, query keys, sorting, filtering, or empty-state conditions.

- [ ] **Step 4: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/components/world frontend/src/components/map frontend/src/components/facts frontend/src/components/codex
git commit -m "style: adapt world panels to HUD modules"
```

---

## Task 5: Adapt Character, Inventory, Quests, NPCs, Journal, Logs

**Files:**
- Modify: `frontend/src/components/character/CharacterSheet.tsx`
- Modify: `frontend/src/components/inventory/InventoryPanel.tsx`
- Modify: `frontend/src/components/quests/QuestPanel.tsx`
- Modify: `frontend/src/components/npcs/NPCPanel.tsx`
- Modify: `frontend/src/components/journal/JournalPanel.tsx`
- Modify: log panel component path used by `CampaignPlayPage.tsx`

- [ ] **Step 1: Apply semantic panel accents**

Use:

```tsx
<HudPanel title="Character" accent="vitals">...</HudPanel>
<HudPanel title="Inventory" accent="inventory">...</HudPanel>
<HudPanel title="Quests" accent="objective">...</HudPanel>
<HudPanel title="NPCs" accent="dialogue">...</HudPanel>
<HudPanel title="Journal" accent="scene">...</HudPanel>
<HudPanel title="Logs" accent="scene">...</HudPanel>
```

- [ ] **Step 2: Convert repeated panel messages**

Replace local `PanelMessage` variants with `HudPanel` empty/loading/error states where possible. Keep local helpers only if they encode unique content structure.

- [ ] **Step 3: Baseline-fix all local buttons/badges**

Use `hud-btn`, `hud-tab-button`, or `hud-baseline-badge`. Do not add new raw `py-*` to uppercase controls.

- [ ] **Step 4: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/components/character frontend/src/components/inventory frontend/src/components/quests frontend/src/components/npcs frontend/src/components/journal
git commit -m "style: adapt campaign detail panels to HUD"
```

---

## Task 6: Adapt Replay Page to Replay HUD State

**Files:**
- Modify: `frontend/src/pages/ReplayPage.tsx`
- Modify: `frontend/src/components/replay/ReplayControls.tsx`
- Modify: `frontend/src/components/replay/ReplayNarrative.tsx`
- Modify: `frontend/src/components/replay/ReplaySidebar.tsx`
- Modify: `frontend/src/components/replay/ReplayTimeline.tsx`

- [ ] **Step 1: Add replay tone to page shell**

Replay should feel like archived playback, not live campaign play. Use pewter/sapphire as primary colors.

- [ ] **Step 2: Convert replay controls**

Use `hud-icon-btn` for previous/play/next where icon-only, and `hud-tab-button` for speeds.

- [ ] **Step 3: Convert replay timeline and sidebar**

Wrap with `HudPanel accent="replay"` or `accent="scene"`.

- [ ] **Step 4: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/pages/ReplayPage.tsx frontend/src/components/replay
git commit -m "style: adapt replay page to HUD"
```

---

## Task 7: Adapt Campaign List Dashboard

**Files:**
- Modify: `frontend/src/pages/CampaignListPage.tsx`

- [ ] **Step 1: Convert campaign cards to HUD cards**

Use `game-hud-panel game-hud-panel-scene` or `HudPanel accent="scene"` for campaign records.

- [ ] **Step 2: Convert dashboard actions**

Use:

```tsx
className="hud-btn hud-btn-primary"
className="hud-btn hud-btn-secondary"
className="hud-btn hud-btn-danger"
```

- [ ] **Step 3: Keep list page less dense than play page**

Do not use the fixed 16:9 game shell. Use the default page scroll, but with HUD cards/chrome.

- [ ] **Step 4: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/pages/CampaignListPage.tsx
git commit -m "style: adapt campaign dashboard to HUD"
```

---

## Task 8: Adapt Startup Wizard to Setup HUD State

**Files:**
- Modify: `frontend/src/pages/CampaignCreatePage.tsx`
- Modify: `frontend/src/components/start/MethodPicker.tsx`
- Modify: `frontend/src/components/start/ProposalPicker.tsx`
- Modify: `frontend/src/components/start/RulesModeStep.tsx`
- Modify: `frontend/src/components/start/ChatStep.tsx`
- Modify: `frontend/src/components/start/ConfirmationPanel.tsx`
- Modify: `frontend/src/components/start/CharacterGuidedForm.tsx`
- Modify: `frontend/src/components/start/CampaignAttributesForm.tsx`

- [ ] **Step 1: Set setup-state visual language**

Use sapphire as the setup/progression accent. Reserve gold for selected/high-value options.

- [ ] **Step 2: Normalize all wizard buttons**

Replace raw button padding with `hud-btn` or `hud-text-button`.

- [ ] **Step 3: Normalize option cards**

Use `game-hud-panel game-hud-panel-setup` for selectable wizard modules.

- [ ] **Step 4: Preserve form semantics**

Do not alter controlled inputs, submit handlers, validation rules, or API payload shapes.

- [ ] **Step 5: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/pages/CampaignCreatePage.tsx frontend/src/components/start
git commit -m "style: adapt startup wizard to HUD"
```

---

## Task 9: Adapt Auth Pages to Auth HUD State

**Files:**
- Modify: `frontend/src/pages/LoginPage.tsx`
- Modify: `frontend/src/pages/RegisterPage.tsx`

- [ ] **Step 1: Keep auth pages standalone**

Do not wrap auth pages in authenticated `AppShell`. Use the same dark HUD language with an auth card.

- [ ] **Step 2: Apply auth panel state**

Use `game-hud-panel game-hud-panel-auth` for the form card.

- [ ] **Step 3: Normalize auth submit buttons**

Use `hud-btn hud-btn-primary` or ruby primary if the existing brand flow requires it.

- [ ] **Step 4: Build and commit**

```bash
pnpm --dir frontend build
git add frontend/src/pages/LoginPage.tsx frontend/src/pages/RegisterPage.tsx
git commit -m "style: adapt auth pages to HUD"
```

---

## Task 10: Final UI Consistency Sweep

**Files:**
- Modify only files with remaining ad hoc controls discovered by grep.

- [ ] **Step 1: Search for remaining button baseline risks**

Run:

```bash
rg "inline-flex.*uppercase|uppercase.*py-|rounded-sm.*uppercase|tracking-\[" frontend/src
```

Review every remaining match. If it is a button, tab, badge, chip, or label, convert it to a semantic HUD utility.

- [ ] **Step 2: Search for gold overuse**

Run:

```bash
rg "gold" frontend/src/pages frontend/src/components frontend/src/index.css
```

Expected: gold remains for brand, inventory/reward, GM narration, and selected highlights; neutral chrome should usually be pewter.

- [ ] **Step 3: Run full frontend build**

```bash
pnpm --dir frontend build
```

- [ ] **Step 4: Optional lint check**

Run:

```bash
pnpm --dir frontend lint
```

Expected: lint may fail due to known pre-existing issues. If it fails, confirm whether failures are unrelated to touched files before proceeding.

- [ ] **Step 5: Commit final sweep**

```bash
git status --short
git add frontend/src
git commit -m "style: complete HUD consistency sweep"
```

---

## Verification Checklist

- [ ] Campaign play still uses fixed widescreen HUD layout.
- [ ] Exploration/dialogue/combat modes remain visually distinct.
- [ ] Scene scan remains pinned to sidebar bottom.
- [ ] Vitals/objective/inventory remain in sidebar top stack.
- [ ] Bottom tabs are compact, centered, and do not create a horizontal scrollbar at target widescreen size.
- [ ] World submenu tabs (`map`, `facts`, `codex`, `relationships`) have corrected baseline alignment.
- [ ] Codex tabs (`languages`, `cultures`, `beliefs`, `economies`) have corrected baseline alignment.
- [ ] Quest filters and clear button have corrected baseline alignment.
- [ ] Journal, replay, combat, and wizard buttons have corrected baseline alignment.
- [ ] Gold is not the default border/text color everywhere.
- [ ] `pnpm --dir frontend build` passes.

## Plan Self-Review

- **Spec coverage:** Covers all routed pages, campaign play submenus, state variants, semantic tokens/utilities, and submenu button baseline fixes requested by the user.
- **Placeholder scan:** No TBD/fill-later implementation steps. Each task has concrete files, commands, and class patterns.
- **Type consistency:** `HudPanelAccent` includes every accent referenced later in the plan.
