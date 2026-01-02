# Pitch: TUI Dashboard

## Problem

When I run `ntx` with no arguments, I want to see my portfolio at a glance in a beautiful terminal UI, so I can quickly check my holdings without switching to a web browser.

The CLAUDE.md specifies `ntx` with no args should launch TUI, but the `tui/` directory is empty. Power users and developers prefer terminal interfaces for quick access.

## Appetite

**2 weeks**

Why: Bubbletea is well-documented, lipgloss already in use for CLI styling. Focused scope with three views only.

## Solution

### Overview

Build a Bubbletea-based interactive dashboard that launches when `ntx` is run without subcommands. Three switchable views with keyboard navigation.

### Key Elements

**Holdings View (Default)**
Table showing: Symbol, Qty, Avg Cost, Current Price, Unrealized P&L, % Change. Sorted by value descending. Color-coded gains (green) and losses (red).

**Transactions View**
Recent transactions list with type, symbol, quantity, price, date. Scrollable with `j/k`.

**Summary View**
Portfolio metrics: Total invested, current value, unrealized P&L, realized P&L (once implemented), total return %.

**Navigation**
- `1/2/3` or `h/t/s` - Switch views (Holdings/Transactions/Summary)
- `j/k` or arrows - Scroll
- `s` - Sync prices from NEPSE
- `r` - Refresh display
- `q` or `Esc` - Quit

### Flow

1. User runs `ntx` (no arguments)
2. TUI launches, shows Holdings view
3. User navigates with keyboard
4. Press `s` to fetch latest prices
5. View updates with new P&L calculations
6. `q` exits cleanly

## Rabbit Holes

- **Complex filtering/sorting**: Defer. First version shows all holdings sorted by value. Add filtering in future cycle.
- **Real-time price streaming**: NEPSE has no public streaming API. Manual sync via `s` key is sufficient.
- **Mouse support**: Skip entirely. Keyboard-only is idiomatic for TUI apps.
- **Responsive layouts**: Use fixed-width columns. Terminal width detection can come later.
- **Help overlay**: Simple `?` shows keybindings in footer, not a modal.

## No-Gos

- Charts or sparklines (significant complexity, save for v2)
- Transaction entry/editing via TUI (CLI handles import)
- Multi-portfolio switching
- Custom themes or color configuration
- Persistent layout preferences

## Technical Notes

Dependencies to add:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - Table, viewport components
- `github.com/charmbracelet/lipgloss` - Already available for styling

File structure:
```
cmd/ntx/tui/
  tui.go        # Main model, Update, View
  holdings.go   # Holdings view component
  transactions.go
  summary.go
  styles.go     # Shared lipgloss styles
```

## Open Questions

- Should sync show a loading spinner or block?
- Include last sync timestamp in footer?

## Success Criteria

- [ ] `ntx` (no args) launches interactive TUI
- [ ] Holdings view shows all positions with P&L
- [ ] Can switch between three views
- [ ] `s` syncs prices and updates display
- [ ] `q` exits cleanly
- [ ] Colors indicate gains/losses
