---
version: 1
slug: "apps-web-src-routes-company-symbol-page-svelte"
primary_target: "apps/web/src/routes/company/[symbol]/+page.svelte"
related_targets: ["apps/web/src/lib/components/company/StoryOpening.svelte","apps/web/src/lib/components/Sidebar.svelte"]
---

# Company page

Scope: /company/[symbol], the product. Mode: Operate (the visitor reads one company and forms a view). Audience: the maintainer and a small circle of NEPSE holders, mostly outside market hours, laptop or phone.

## Direction contract

THESIS: A stock report, not a data page: one dark report card per question, in the Simply Wall St reading order (valuation, past performance, financial health, dividend, ownership, funds, plants), each opening with the sentence the filings support and a chip counting the checks that hold. Refused: the eyebrow-plus-hairline editorial scroll of the previous build, and any buy/sell call.

OWN-WORLD: Gunmetal night ground (near-black blue-grey) with raised gunmetal cards, fog silver type, molten orange as the single accent on the snowflake fill, the primary button, active nav and the price when flat. Green and red only for price and verdict direction. Figtree for UI and figures, Source Serif 4 for headings and claims. Rounded 8px cards with soft offset shadows, 6px buttons, tabular numerals everywhere.

STORY: The reader sees the name, price and one-year line, then the snowflake tells them in one glance where the company is strong and weak across five axes derived from the eight published checks. Each numbered section proves its axis with figures and provenance. They leave with their own view, never a recommendation.

FIRST VIEWPORT: Full-width header card: monogram tile, name, "NEPSE:SYM Stock Report", market cap; AI research as the orange primary button; right side the latest close with 1D/1Y change over a one-year sparkline. Below, the left sticky numbered section nav and the Overview card: claims and the holds/watch lists on the left, the orange snowflake with curved axis labels and its one-line summary on the right.

FORM: Report card with sticky section nav (first on the list; the brief pins the composition to the reference). seed: none (unavailable)

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
