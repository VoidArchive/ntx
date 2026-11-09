---
name: NTX
description: A dark stock report for NEPSE companies, one gunmetal card per question, molten orange as the single accent.
colors:
  primary: "oklch(0.72 0.19 45)"
  primary-foreground: "oklch(0.17 0.012 250)"
  background: "oklch(0.17 0.012 250)"
  card: "oklch(0.215 0.012 250)"
  popover: "oklch(0.235 0.012 250)"
  secondary: "oklch(0.27 0.012 250)"
  muted: "oklch(0.27 0.012 250)"
  accent: "oklch(0.3 0.012 250)"
  border: "oklch(0.3 0.012 250)"
  foreground: "oklch(0.87 0.008 240)"
  muted-foreground: "oklch(0.66 0.01 240)"
  positive: "oklch(0.74 0.15 155)"
  negative: "oklch(0.7 0.18 22)"
  caution: "oklch(0.8 0.13 85)"
  sidebar: "oklch(0.17 0.012 250)"
  sidebar-accent: "oklch(0.235 0.012 250)"
  sidebar-border: "oklch(0.27 0.012 250)"
  chart-2: "oklch(0.7 0.01 240)"
typography:
  display:
    fontFamily: "Source Serif 4 Variable, Georgia, serif"
    fontSize: "2.25rem"
    fontWeight: 500
    lineHeight: 1.1
    letterSpacing: "-0.025em"
  headline:
    fontFamily: "Source Serif 4 Variable, Georgia, serif"
    fontSize: "1.5rem"
    fontWeight: 500
    lineHeight: 1.33
    letterSpacing: "0em"
  claim:
    fontFamily: "Source Serif 4 Variable, Georgia, serif"
    fontSize: "1.25rem"
    fontWeight: 400
    lineHeight: 1.375
    letterSpacing: "0em"
  figure:
    fontFamily: "Source Serif 4 Variable, Georgia, serif"
    fontSize: "1.875rem"
    fontWeight: 500
    lineHeight: 1.2
    letterSpacing: "0em"
  lede:
    fontFamily: "Figtree Variable, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 400
    lineHeight: 1.625
    letterSpacing: "0em"
  body:
    fontFamily: "Figtree Variable, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.43
    letterSpacing: "0em"
  meta:
    fontFamily: "Figtree Variable, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 400
    lineHeight: 1.33
    letterSpacing: "0em"
  list-label:
    fontFamily: "Figtree Variable, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 600
    lineHeight: 1.33
    letterSpacing: "0.05em"
  axis-label:
    fontFamily: "Figtree Variable, system-ui, sans-serif"
    fontSize: "13px"
    fontWeight: 600
    lineHeight: 1
    letterSpacing: "0.18em"
rounded:
  sm: "4px"
  md: "6px"
  lg: "8px"
  xl: "12px"
  full: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "12px"
  lg: "16px"
  xl: "24px"
  2xl: "32px"
components:
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    padding: "24px"
  card-wide:
    backgroundColor: "{colors.card}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.lg}"
    padding: "32px"
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "0 12px"
    height: "36px"
  button-secondary:
    backgroundColor: "{colors.secondary}"
    textColor: "{colors.foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "0 12px"
    height: "36px"
  chip-section:
    backgroundColor: "transparent"
    textColor: "{colors.muted-foreground}"
    typography: "{typography.meta}"
    rounded: "{rounded.full}"
    padding: "4px 10px"
  chip-section-all-hold:
    backgroundColor: "transparent"
    textColor: "{colors.positive}"
    typography: "{typography.meta}"
    rounded: "{rounded.full}"
    padding: "4px 10px"
  chip-section-none-hold:
    backgroundColor: "transparent"
    textColor: "{colors.negative}"
    typography: "{typography.meta}"
    rounded: "{rounded.full}"
    padding: "4px 10px"
  toggle-range:
    backgroundColor: "{colors.background}"
    textColor: "{colors.muted-foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "4px 10px"
  toggle-range-selected:
    backgroundColor: "{colors.foreground}"
    textColor: "{colors.background}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "4px 10px"
  monogram-tile:
    backgroundColor: "color-mix(in oklab, oklch(0.72 0.19 45) 15%, transparent)"
    textColor: "{colors.primary}"
    typography: "{typography.figure}"
    rounded: "{rounded.lg}"
    size: "80px"
  nav-section-active:
    backgroundColor: "{colors.sidebar-accent}"
    textColor: "{colors.foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "6px 8px 6px 12px"
  nav-section:
    backgroundColor: "transparent"
    textColor: "{colors.muted-foreground}"
    typography: "{typography.body}"
    rounded: "{rounded.md}"
    padding: "6px 8px 6px 12px"
  snowflake:
    textColor: "{colors.primary}"
    typography: "{typography.axis-label}"
    width: "340px"
---

# Design System: NTX

## Overview

**Creative North Star: "The Night Report"**

NTX reads like a printed stock report laid on a dark desk: a gunmetal night ground, one raised gunmetal card per question, fog silver type, and a single molten orange accent that marks the thing the reader is meant to look at (the snowflake, the primary button, the active section, the price line). Serif headings and serif claim sentences carry what the filings say; a humanist sans carries the figures, labels and chrome. Nothing on the page recommends; colour that reads as a verdict (green, red) is spent only where the data itself has a direction.

The world is dark by default (`mode-watcher` sets `defaultMode="dark"`). A light variant exists under `:root` with the same roles; its values live in the sidecar (`.impeccable/design.json`, `colorMeta.*.light`), not here, so the frontmatter stays one source of truth for the shipped default.

**Scope note.** Every route is in this world (2026-11-09): the company page and its header card, snowflake and sidebar; the home page; the company list, hydro and mutual fund list as a header card over content cards; the fund detail as the company header's twin over a holdings card. A list route keeps its filters inside the header card and its table in one card that scrolls sideways on a phone, clipping at the card edge.

**Key Characteristics:**
- Dark gunmetal ground, cards one step lighter, hairline borders one step lighter again; depth is tonal plus one soft shadow.
- One accent, molten orange; green and red are reserved for direction and verdict.
- Serif for headings, claim sentences and the headline price; sans for everything else; tabular numerals everywhere.
- Sections are numbered report cards: number + title, a verdict chip, one claim sentence, a two-column checks list.
- Icons are Lucide line glyphs at 14-16px, used only to mark verdicts and actions.

## Colors

A near-monochrome blue-grey ramp (hue 240-250, chroma 0.003-0.012) with one warm accent; semantic greens and reds sit on the same lightness as the accent so they read as peers, not alarms.

### Primary
- **Molten Orange** (`colors.primary`): the only accent. Snowflake fill and outline, the primary button, the active section's left bar in the sidebar, the sparkline and its fading gradient, the "Share price" dot, the monogram tile's letters and 15% tint, text selection, focus ring, and the 1D/1Y change when it is exactly flat. Text on it is the ground colour (`colors.primary-foreground`), never white.

### Semantic
- **Up Green** (`colors.positive`): a price change above zero; a check that holds (icon, and chip text when every evaluated check in a section holds). Also tints a rising price chart (`stroke-positive`, `fill-positive/15`).
- **Down Red** (`colors.negative`): a price change below zero; chip text when no evaluated check in a section holds; the falling price chart. Doubles as `destructive`.
- **Caution Amber** (`colors.caution`): defined and themed, but the company page does not use it; "Watch" verdicts render in `muted-foreground`. Treat it as reserved, not as a third verdict colour.

### Neutral
- **Gunmetal Night** (`colors.background`, also `colors.sidebar`): the page and sidebar ground.
- **Raised Gunmetal** (`colors.card`): every report card and the header card.
- **Popover Gunmetal** (`colors.popover`, same value as `colors.sidebar-accent`): menus, dialogs, and the active sidebar link's fill.
- **Secondary / Muted** (`colors.secondary`, `colors.muted`): the secondary button, the AI-prompt `<pre>` panel, the snowflake's outer halo ring, scrollbar thumb.
- **Accent Fill** (`colors.accent`, same value as `colors.border`): hover fills on ghost and outline buttons, hairline borders, snowflake rings and spokes, card dividers.
- **Fog Silver** (`colors.foreground`): primary type; also the selected range toggle's fill (with the ground as its text).
- **Fog Dim** (`colors.muted-foreground`): secondary type, section numbers, breadcrumbs, provenance lines, inactive nav, Watch and Needs-more-history icons, unevaluated axis labels.
- **Chart Grey** (`colors.chart-2`): the neutral chart series; `chart-3`/`chart-4`/`chart-5` exist for multi-series charts and event markers but do not appear in the rebuilt header or overview.

### Named Rules
**The One Flame Rule.** Orange is a pointer, not a palette. On any screen it marks at most: the snowflake, one primary button, one active nav item, one price line. If a second orange element is a decoration rather than a pointer, remove it.

**The Direction Rule.** Green and red mean one of two things only: a price moved that way, or a published check holds / fails. They never colour headings, backgrounds, chips at rest, or anything a reader could mistake for advice.

**The One Encoding Rule.** Every comparison chart on the page reads the same way, and no chart carries its own legend: orange is this company (or its latest value), Fog at 25% is everyone else (or its past values), a dashed Fog line is the median, and the check icon beside the question carries the verdict. A chart that needs a second colour or a legend is two charts.

**The Ground-on-Accent Rule.** Text on orange, green or red is the ground colour (`*-foreground` tokens), not white. The accent glows against the night; it does not carry a white label.

## Typography

**Display / Heading Font:** Source Serif 4 Variable (with Georgia, serif), self-hosted via `@fontsource-variable/source-serif-4`.
**Body / UI Font:** Figtree Variable (with system-ui, sans-serif), self-hosted via `@fontsource-variable/figtree`.
**Numerals:** `font-variant-numeric: tabular-nums` on `body`; every figure, table and date already lines up without extra classes.

The serif speaks for the company and the filings; the sans speaks for NTX. `h1`-`h3` are serif, weight 500, `text-wrap: balance` by base style, but list labels inside cards override back to sans (see `list-label`).

### Scale as used
- **display** (36px desktop, 30px phone; serif 500, tight tracking): the company name in the header card.
- **headline** (24px, serif 500): every section title, with its number in `muted-foreground` and an 8px gap before the title; also "SYM Stock Overview".
- **figure** (30px, serif 500, tabular): the headline price "Rs 552.00" and the monogram initials in the tile.
- **claim** (20px, serif 400, snug 1.375 leading, max-width 48rem): the one sentence the filings support under each section head. Numbers inside it are dotted-underlined links to their evidence; period and "filing" trailers drop to `meta` size, weight 400.
- **lede** (18px, sans 400, relaxed 1.625 leading): the Overview's two opening sentences; the only large sans paragraph, deliberately not serif so the Overview reads as summary rather than claim.
- **body** (14px, sans 400): checks lists, sidebar links, chips, toggles, table cells, breadcrumbs, the 1D/1Y change.
- **meta** (12px, sans 400): provenance lines, check rule text, footer, the `(Q1 2083/84)` period trailers.
- **list-label** (12px, sans 600, uppercase, 0.05em tracking, `muted-foreground`): the "HOLDS" / "WATCH" headings over the Overview lists. These head a list that follows them; they are never placed above a title as an eyebrow.
- **axis-label** (13px, sans 600, uppercase, 0.18em tracking): the five snowflake axis names, set on arcs outside the ring.
- Sidebar wordmark: "NTX" in serif, 20px desktop, 18px phone, tight tracking. No logo image.

### Named Rules
**The Two Voices Rule.** Serif for headings, claim sentences and the headline price; sans for figures, labels, lists, chrome. A serif figure is only allowed where it is the page's single headline number.

**The Linked Number Rule.** A number inside a claim sentence is a link to its evidence: dotted underline, 4px offset, orange on hover. Prose that cannot point at a row does not get a number.

## Layout

- **Shell:** desktop (`lg`, 1024px and up) is a two-column flex: a 224px sticky sidebar (full viewport height, 12px horizontal padding, 16px vertical, right hairline border) and a fluid main column. Below `lg` the sidebar becomes a 48px sticky top bar (ground at 80% with backdrop blur, bottom hairline) whose menu button opens the same nav in a left-anchored dialog.
- **Page column:** `max-width 80rem` (1280px), centred, 16px horizontal padding, 24px vertical. The content column is the only grid; there is no page-level 12-column system.
- **Card rhythm:** cards stack with a 24px gap (`mt-6`); each has `scroll-margin-top: 24px` so anchor jumps land with breathing room. Inside a card: 24px padding on phone, 32px from `sm` (640px).
- **Header card:** two columns from `lg`, `1fr` and a 420px price column; stacks on phone. The monogram tile (80px) sits 20px left of the name block.
- **Overview card:** two columns from `lg`, `1fr` and a 340px snowflake column, 32px gap; the snowflake centres on phone and left-aligns in its column on desktop.
- **Section anatomy, top to bottom:** title row (title left, chip right, wraps on phone) → claim sentence 12px below → checks list 20px below as a two-column grid from `sm` (32px column gap, 12px row gap) → the section's figures and tables 24px below (32px for the valuation chart and strips).
- **Sidebar nav:** site links (Home, Companies, Hydro, Mutual funds) at 14px, 2px apart; on a company page an "On this page" list follows 24px below with the numbered sections, the number in a 12px-wide column so titles align.
- **Footer of the page:** a 12px provenance line, 32px below the last card, with 8px horizontal padding so it aligns to card text.

### Named Rules
**The One Column of Cards Rule.** The company page is a single vertical stack of full-width cards; nothing sits beside a card except the sidebar. Two-column layouts live inside cards, never between them.

**The Breathing Anchor Rule.** Every section is an anchor target with 24px scroll margin; the sidebar highlights the topmost visible section using an observer with a `-10% 0 -60% 0` root margin so long sections stay active while read.

## Elevation & Depth

Depth is tonal first, shadow second. The ground is the darkest step; cards sit one step lighter; popovers and the active nav fill one step lighter again; borders are a fourth step and never darker than the surface they outline.

- **Card shadow** (`--shadow-card`: `0 1px 2px oklch(0 0 0 / 0.25), 0 8px 24px -12px oklch(0 0 0 / 0.5)`): the one shadow in the system, on every report card and the header card. It is soft and downward, an ambient lift, never an offset block.
- **Hairlines** (`colors.border`, 1px): sidebar edge, mobile bar edge, chip outlines, toggle outlines, snowflake rings and spokes. Cards themselves have no border; the shadow and the tonal step separate them.
- **Tints over surfaces:** translucency is used for emphasis fills, not for glass: the monogram tile is orange at 15%, the sparkline gradient runs orange 35% to 0, the snowflake blob is orange at 78%, the halo ring is muted at 50%. The only blur is the mobile top bar (ground at 80% plus a small backdrop blur).
- **Focus:** a 2px ring in the accent colour, 2px outside the element, on `:focus-visible` only.

### Named Rules
**The Tonal Step Rule.** Each layer is exactly one lightness step above the one it rests on (ground 0.17 → card 0.215 → popover/active 0.235 → secondary 0.27 → border 0.3). Do not introduce a surface between steps or skip one.

**The Soft Lift Rule.** The only shadow is `--shadow-card`. No hard offset shadows, no inner shadows, no glows other than the orange tints listed above.

## Shapes

- **Cards and tiles:** 8px radius (`rounded.lg`): report cards, the header card, the monogram tile, the AI-prompt panel.
- **Controls:** 6px radius (`rounded.md`): buttons, range toggles, sidebar section links. Site nav links use 4px (`rounded.sm`).
- **Pills:** fully round (`rounded.full`): the section chip, the "Share price" dot, snowflake axis points.
- **Lines:** the sparkline is 1.75px with round joins and non-scaling stroke; the snowflake outline is 2px with round joins; rings and spokes are 1px, the outer ring 1.5px.
- **The snowflake blob** is a closed Catmull-Rom spline through five axis points (tension 1/6), so the shape is organic, never a polygon.
- **Active state is a bar, not a fill alone:** the active section link carries a 2px left border in orange plus the popover fill; inactive links have a transparent 2px border so text does not shift.

### Named Rules
**The 8/6 Rule.** Surfaces are 8px, controls are 6px, counters are round. There is no 12px+ radius in the shipped company page even though `rounded.xl` is defined.

## Components

### Report card (`card`, `card-wide`)
Shipped as the `card-report` utility in `layout.css`; pages add `mt-6` between cards.
A raised gunmetal surface, 8px corners, `--shadow-card`, 24px padding on phone and 32px from `sm`, stacked 24px apart with a 24px scroll margin. It has no border and no header strip; its first line is always a serif title.

### Section head (the anatomy every numbered card follows)
1. **Title row:** `headline` serif title with its number in `muted-foreground` and 8px right margin, flex-spaced against the chip; wraps on phone with a 12px gap.
2. **Chip:** see below.
3. **Claim sentence:** `claim` serif, 12px below, max-width 48rem; when there is nothing to claim, the empty-state sentence is the same size in `muted-foreground`. Valuation omits it: its figures row says the same thing, and a sentence whose links point a few hundred pixels down the same card reads as a dead link.
4. **Checks list:** 20px below, two columns from `sm`, each item a 16px verdict icon (4px top offset, `positive` for holds, `muted-foreground` for watch and pending) beside the check's name in weight 500, " · Verdict" in `muted-foreground`, and the rule text on its own line at `meta` size.

### Section chip (`chip-section`, `-all-hold`, `-none-hold`)
A hairline pill, 10px horizontal and 4px vertical padding, 12px weight 500 text, a 14px Lucide icon with a 6px gap. Four states, all with transparent fill: no checks (`CircleDashed`, muted text, muted border); needs more history (`CircleDashed`, muted); some hold (`CircleCheck`, foreground text, muted border); all hold (`CircleCheck`, green text, green border at 40%); none hold (`CircleCheck`, red text, red border at 40%). Hovering shows every check and verdict in the title tooltip.

### Header card (StoryOpening)
A `card` with two columns from `lg`. Left: breadcrumb (`body`, muted, "Stocks / Sector" with slashes), the 80px monogram tile (first letters of the first two alphabetic words, serif 30px, orange on 15% orange), the `display` name, "NEPSE:SYM Stock Report" in `lede`-sized muted sans, a `body` muted line with market cap and "Rating X, n of m checks hold", the primary button 24px below, and a `meta` provenance line 24px below that. Right: a "Share price" label with an orange 8px dot, the `figure` price in serif, the one price chart on the page (layerchart area, 176px tall, coloured by the Direction Rule, y-axis fitted to the period's range, result and notice dates as dashed marks), then one row with the range toggle on the left and the 1D / 1Y changes on the right in `body` tabular, label muted, value coloured by the Direction Rule (orange when flat). There is no second price chart and no volume chart; the header is the price's only home.

### Primary button (`button-primary`)
shadcn `default` variant at size `sm`: 36px tall, 12px horizontal padding, 6px radius, 14px weight 500, orange fill with ground-coloured text and a 16px leading Lucide icon; hover drops the fill to 90%. One per header card ("AI Research"). The `secondary` variant (`button-secondary`, `colors.secondary` fill, hover 80%) is used inside the AI dialog for Copy. `outline`, `ghost` and `link` variants exist in the shadcn map but are not placed on the company page.

### Range toggle (`toggle-range`, `-selected`)
A row of 6px-radius hairline buttons, 6px apart, 14px weight 500, 10px/4px padding (12px/6px from `sm`). Selected inverts: Fog Silver fill, ground text, foreground border. Unselected hovers to a 50% foreground border. This is the only inverted control in the system.

### Valuation figures row
A `<dl>` of four figures (PE, PB, EPS, Dividend yield), two columns on phone and four from `sm`: `meta` muted label, `headline`-sized serif value in tabular numerals, and the sector median (or "no sector figures") in `meta` muted beneath. One provenance line follows. No bars: the figures only state, the question charts below compare.

### Question chart (`RatioHistory`, `PeerBars`)
A `lede`-sized serif question that is the published check reworded ("Is the PE below its own three-year median?"), the verdict icon and word in `body` muted directly under it, then the chart 24px below. Two per Valuation card, side by side from `lg`, each carrying the claim anchor the API links to (`strip-pe`, `strip-pb`). Both obey the One Encoding Rule:
- **Ratio history:** up to twelve quarterly columns, 176px tall, 8px apart, oldest six hidden on phone; the latest column orange, the rest Fog 25%, 2px top radius; values above each column at 11px (latest in weight 500), periods beneath at 10px; the company's own median as one dashed horizontal line with its value in a small ground-filled label at the right end.
- **Peer bars:** a 3-column grid (4.5rem symbol, bar, 3.5rem value), 12px-tall bars 6px apart, sorted ascending; this company in orange and weight 500, peers Fog 25% with muted symbols that link to their pages; the sector median as a dashed vertical line drawn through every row with its value labelled above the column. Shows the nearest five each way when a sector has more than twelve peers, with a "Show all" text button.

### Home header card
The company header's twin for the market: left, "Nepal Stock Exchange" with the session's Bikram Sambat and Gregorian dates in `body` muted, the `display` "NEPSE", the one-sentence session summary in `claim` serif (close, direction by the Direction Rule, turnover, profitable share), a `meta` provenance line. Right, when a session is recorded: "NEPSE index" with the orange dot, the `figure` close with its 1D change, the advance / decline / unchanged bar (green, red, `muted`, 8px tall, segments separated by a 2px card-coloured line) with its legend, and a `<dl>` of session figures (turnover, shares traded, transactions, scrips, market cap when known) in two columns on phone and four from `sm`. Sectors, Filed recently and Latest session follow as unnumbered report cards: the sector table keeps its first column sticky on `card`, the filings list stacks symbol over sentence on phone and runs in two columns from `lg`.

### List route (companies, hydro, mutual funds)
A header card with the `display` title, the `claim` sentence and the route's filters or view toggle, then one content card per table or map. Table symbols are sans weight 500, never serif; rows hover to `accent` at 40%; a sticky first column sits on `card`. The fund grid is the exception: each fund is its own report card, so the grid is not wrapped in another. The fund detail header mirrors the company header (monogram, name, "SYM Fund Report", manager line, sentence, provenance; right, "NAV per unit" with the orange dot, the figure, from-par change, and the stats list) and the holdings list draws each sector's share as a Fog bar scaled to the largest sector, with no colour legend.

### Snowflake
A 340px-wide SVG (viewBox 340×340, centre 170, outer radius 118) with five axes drawn clockwise from the top: Value, Past, Health, Dividend, Ownership. Each axis is the share of its evaluated checks that hold, drawn out from a 14% core so a zero still shows a sprout. Rings at 25/50/75/100% in `border` (the outer at 1.5px), spokes in `border`, a 12px halo ring just outside at `muted` 50%. The shape is an orange Catmull-Rom blob at 78% opacity with a 2px orange outline; axis points are 3px circles, ground-filled, orange-stroked, each with a title tooltip. Labels ride arcs outside the ring (22px out above, 34px out below, reversed in the lower half so text stays upright) in `axis-label`; an axis with no evaluated checks sets its label in `muted-foreground`. A "Snowflake" caption in weight 500 and a `body` muted explanation sit under it; the explanation always ends "Not a recommendation."

### Sidebar and section nav (`nav-section`, `-active`)
A 224px column on the ground colour with a right hairline. Wordmark "NTX" in serif, search 16px below, nav 16px below that, theme toggle pinned to the bottom right. Site links: 14px, 8px/6px padding, 4px radius, current page in weight 500 Fog Silver, others muted, hover fills `sidebar-accent`. Section links: 6px radius, 2px left border, 12px left padding, the number in a 12px column at `meta` size; active shows the orange bar, the `sidebar-accent` fill and weight 500; inactive is muted and brightens on hover. Sections the page does not render (Plants for non-hydro companies) are omitted, not disabled.

### Overview lists
"HOLDS" and "WATCH" in `list-label` head two plain lists, 8px apart, each item a 16px verdict icon beside the rule sentence at `body`. Pending checks are summarised in one `body` muted sentence below, not listed.

### Motion
Route changes cross-fade through the View Transitions API (150ms ease-out out, 150ms ease-in in with a 50ms delay) and are disabled under `prefers-reduced-motion`. Interactive colour changes use Tailwind's default `transition-colors`. Nothing else animates.

## Do's and Don'ts

**Do**
- Open every section with its number, its serif title, its chip and the one sentence the filings support, in that order.
- Spend orange only as a pointer: snowflake, primary button, active section, price line, flat change.
- Colour green or red only for price direction and check verdicts, and set text on them in the ground colour.
- Keep every figure in tabular numerals and every claim number as a dotted link to its evidence.
- Separate cards by the tonal step and `--shadow-card`; separate items inside cards with hairlines.
- Use Lucide line icons at 14-16px for verdicts and actions only.
- End any snowflake or rating explanation with "Not a recommendation."

**Don't**
- Don't give a card a border, a header strip, or a second shadow.
- Don't set a label above a title as an eyebrow or kicker; uppercase labels head lists (HOLDS, WATCH) or ride the snowflake's arcs, nothing else.
- Don't introduce a third verdict colour; Watch and Needs-more-history are `muted-foreground`, and `caution` stays reserved until the build uses it.
- Don't draw the snowflake as a polygon, add a sixth axis, or weight an axis; the API's verdicts are the rule.
- Don't put white text on orange, green or red.
- Don't use a radius other than 8px for surfaces, 6px for controls (4px for site nav links), or full for pills.
