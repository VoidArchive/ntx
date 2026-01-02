# Pitch: Web Dashboard (Foundation)

**Status: PASSED** - Not betting this cycle. Ship after core CLI features stabilize.

## Problem

When I want to visualize my portfolio with charts and share it with family, I want a web interface, so I can get a richer view than terminal allows.

SvelteKit scaffold exists with Tailwind and shadcn configured. ConnectRPC server runs. Proto TypeScript bindings generated. But only boilerplate exists.

## Appetite

**4 weeks**

## Solution

### Overview

Build foundational web dashboard: holdings table, summary cards, basic charts. Connect to ntxd via Connect-Web.

### Key Elements

**Holdings Table**
Sortable, filterable table showing all positions with P&L. Click to expand details.

**Summary Cards**
Total value, unrealized P&L, realized P&L, today's change. Color-coded.

**Charts**
- Pie chart: Sector allocation
- Line chart: Portfolio value over time (requires daily snapshots)

**Technical**
- Connect-Web for RPC calls
- Tailwind + shadcn components
- Dark mode support

## Why Passed

1. **Missing data**: No daily portfolio snapshots yet. Line chart needs this.
2. **Realized P&L not ready**: Summary cards would be incomplete.
3. **TUI provides similar value**: For solo developer, TUI might be enough.
4. **Scope creep risk**: Web dashboards tend to grow unbounded.

## Re-pitch Criteria

- Realized P&L and dividend tracking shipped
- Daily snapshot mechanism implemented (simple cron or on-sync)
- Clear use case beyond "looks nice" - sharing? mobile access?

## Notes for Future Shaping

- Consider mobile-first if the job is "check portfolio on phone"
- PWA might be better than full web app for offline access
- Authentication story if ever multi-user (probably not)
