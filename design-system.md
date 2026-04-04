# Game Master — Design System

> Art Deco RPG: opulent geometry meets tabletop fantasy.

## Philosophy

The UI evokes an ornate leather-bound campaign book — gold-framed panels, sharp geometric
precision, and dramatic lighting on obsidian black. Every surface is an architectural facade;
every heading is an inscription carved into stone. The style says **"the Game Master's chamber"**,
not a generic SaaS dashboard.

### Core Principles

1. **Geometry as decoration** — sharp corners, stepped edges, L-bracket corner ornaments.
2. **Contrast as drama** — obsidian black vs metallic gold; no muddy middle ground.
3. **Symmetry and ceremony** — centered headings, balanced columns, decorative dividers.
4. **Verticality** — tall proportions, stacked elements, upward aspiration.
5. **Material luxury** — gold glows, subtle crosshatch texture, no flat drop shadows.

---

## Color Tokens

| Token        | Hex       | Role                                             |
| ------------ | --------- | ------------------------------------------------ |
| `obsidian`   | `#0A0A0A` | Page background                                  |
| `charcoal`   | `#141414` | Card / panel surfaces                            |
| `champagne`  | `#F2F0E4` | Primary text (warm cream, never pure white)       |
| `gold`       | `#D4AF37` | Primary accent — borders, labels, GM badge, CTA  |
| `gold-light` | `#F2E8C4` | Hover / active gold brightening                  |
| `gold-dark`  | `#B8901A` | Pressed / deep gold                              |
| `midnight`   | `#1E3D59` | Secondary depth, inactive states                 |
| `pewter`     | `#888888` | Muted / secondary text                           |
| `jade`       | `#4ECCA3` | Player accent, HP, active status, equipped        |
| `ruby`       | `#AB2346` | Error, failed, epic rarity, danger                |
| `sapphire`   | `#5B8FB9` | Rare rarity, informational accent                 |

### Narrative Role Colors

| Role     | Accent    | Usage                            |
| -------- | --------- | -------------------------------- |
| Player   | `jade`    | Player action entries            |
| GM       | `gold`    | Game Master narration            |
| System   | `pewter`  | Notices, mechanical messages     |

### Item Rarity Colors

| Rarity    | Token      |
| --------- | ---------- |
| Common    | `pewter`   |
| Uncommon  | `jade`     |
| Rare      | `sapphire` |
| Epic      | `ruby`     |
| Legendary | `gold`     |

---

## Typography

| Role     | Font             | Classes                                            |
| -------- | ---------------- | -------------------------------------------------- |
| Headings | Marcellus (serif)| `font-heading text-champagne uppercase tracking-[0.2em]` |
| Body     | Josefin Sans     | `font-sans text-champagne/80`                      |
| Labels   | Josefin Sans     | `text-xs font-semibold uppercase tracking-[0.2em] text-gold` |

- H1: `text-4xl` or larger, uppercase, tracking-[0.15em]
- H2: `text-2xl`, uppercase, tracking-[0.12em]
- Body: `text-sm leading-7`
- Muted: `text-pewter`

---

## Borders & Radius

- **Radius**: `rounded-none` everywhere. Art Deco uses sharp edges exclusively.
- **Pill badges**: `rounded-sm` (2px softening) — never `rounded-full`.
- **Borders**: 1px gold at 20–30% opacity (`border-gold/20`). Intensify to 100% on hover.
- **Corner decorations**: `.deco-corners` CSS class adds L-brackets at top-left and bottom-right.

---

## Shadows & Effects

- No traditional drop shadows. Use gold **glows**:
  - `shadow-[0_0_15px_rgba(212,175,55,0.08)]` — panels
  - `shadow-[0_0_20px_rgba(212,175,55,0.15)]` — hover / focus
- Subtle diagonal crosshatch background texture at 3% opacity.
- Radial gold gradient from top-center on page body.

---

## Buttons

| Variant   | Default State                          | Hover State                             |
| --------- | -------------------------------------- | --------------------------------------- |
| Primary   | `bg-gold text-obsidian font-semibold`  | `bg-gold-light shadow-gold-glow`        |
| Secondary | `border-gold/30 text-champagne`        | `border-gold text-gold`                 |
| Danger    | `border-ruby/30 text-ruby`             | `border-ruby bg-ruby/10`                |

All buttons: `rounded-none`, uppercase, `tracking-[0.15em]`, min height `h-12`.
Focus: `ring-2 ring-gold ring-offset-2 ring-offset-obsidian`.

---

## Cards / Panels

```
bg-charcoal border border-gold/20 rounded-none
hover: border-gold/60, -translate-y-0.5
```

Corner decorations (`.deco-corners`) on primary panels.
Header separator: `border-b border-gold/20`.

---

## Inputs

- Transparent background, bottom border only: `border-b-2 border-gold bg-transparent`.
- Focus: border brightens to `gold-light`, subtle gold glow below.
- Labels: uppercase, small, gold, above input.

---

## Animations

- **Timing**: `duration-300` standard, `duration-500` theatrical.
- **Easing**: `ease-out` for mechanical Art Deco feel.
- **Hover**: cards lift `-translate-y-0.5`, border opacity intensifies.
- **Loading dots**: gold bouncing dots (not cyan).

---

## Accessibility

- Gold `#D4AF37` on obsidian `#0A0A0A`: ~7:1 contrast ratio (WCAG AA).
- Champagne `#F2F0E4` on obsidian: ~16:1 contrast ratio.
- Pewter `#888888` on obsidian: ~4.5:1 (acceptable for secondary text).
- All decorative elements use `aria-hidden="true"`.
- Touch targets minimum 48px height.

---

## File References

- **Tailwind tokens**: `frontend/tailwind.config.ts` + `frontend/src/index.css` `@theme` block
- **Global styles**: `frontend/src/index.css`
- **Fonts**: loaded via Google Fonts in `frontend/index.html`
