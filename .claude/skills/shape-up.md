# Shape Up Workflow for NTX

Shape Up methodology adapted for solo/small team open source development.

## File Locations

```
docs/
  pitches/           # Shaped work ready for betting
  cycles/            # Active and past cycle tracking
    TRACKER.md       # Current cycle status
```

## The Process

```
SHAPE → BET → BUILD → SHIP → COOLDOWN
                              ↓
                         (next cycle)
```

## Phase 1: SHAPE (Write a Pitch)

Before building anything significant, write a pitch in `docs/pitches/`.

### Pitch Template

```markdown
# [Feature Name]

## Problem
What job is the user trying to accomplish?
When [situation], I want to [motivation], so I can [outcome].

## Appetite
How much time does this deserve?
- 1 week: Small enhancement
- 2 weeks: New feature
- 4 weeks: Significant capability

## Solution
Fat-marker sketch - key screens, core flows, data relationships.
No implementation details. No edge cases.

## Rabbit Holes
What could explode scope? List and explicitly avoid.

## No-Gos
What's explicitly out of scope for this pitch?
```

### Example Pitch (docs/pitches/realized-pnl-fifo.md)

```markdown
# Realized P/L with FIFO

## Problem
When I sell shares, I want to see my actual profit/loss using FIFO cost basis,
so I can track my real performance and calculate taxes.

## Appetite
2 weeks

## Solution
- Match sells to buys using FIFO
- Store lot assignments in database
- Show realized P/L in transactions view
- Add `ntx realized` command for summary

## Rabbit Holes
- Partial lot matching (handle explicitly)
- Corporate actions affecting cost basis (defer to later pitch)
- Multiple cost basis methods (FIFO only for now)

## No-Gos
- LIFO/average cost methods
- Tax form generation
- Broker-specific lot matching rules
```

## Phase 2: BET (Decide What to Build)

At cycle start, review pitches and decide:

- **Bet**: Build this cycle
- **Pass**: Maybe later
- **Kill**: Not building this

Record bets in `docs/cycles/TRACKER.md`.

### Betting Criteria

1. Does it serve a real job?
2. Does it fit the appetite?
3. Is this the most important thing?
4. Can we actually ship it?

## Phase 3: BUILD (Execute)

### Hill Chart Progress

Track work as uphill (figuring out) or downhill (executing):

```
        ●  ← Stuck = problem
       /
      /
     /  ●  ← Good progress  
    /  /
───●──●─────
Uphill  Downhill
```

Update `docs/cycles/TRACKER.md` with hill positions.

### Scope Cutting

When running out of time:
1. Cut from edges, not core
2. Ask: "Does removing this still accomplish the job?"
3. Nice-to-haves go first
4. Ship the job, not the feature list

## Phase 4: SHIP

### Definition of Done

- Core job can be accomplished
- Works end-to-end
- No blocking bugs
- Documented in README/CHANGELOG

### What Done is NOT

- Every edge case handled
- Perfect code
- Comprehensive tests
- Every enhancement included

## Phase 5: COOLDOWN

After shipping, take 1 week for:
- Bug fixes from shipped work
- Technical debt (limited)
- Exploring ideas for next cycle
- No scheduled deliverables

## Cycle Rhythm

For NTX (solo/small team):

```
Weeks 1-4: BUILD
Week 5: COOLDOWN
Week 6: SHAPE next cycle's pitches, BET
```

## Quick Reference

### Should I Build This?

```
1. Can I state the job in one sentence? No → Write pitch first
2. Does it fit in 2-4 weeks? No → Narrow scope
3. Is this the most important job? No → Consider other pitches
4. Can I ship something useful? No → Reshape
```

### Feature Creep Detection

For every proposed addition:
1. Does this serve the original job?
2. Would users notice if we skipped it?
3. Can this be a separate pitch?

If answers are No/No/Yes → Cut it.

### Kill Criteria

Kill mid-cycle if:
- Job assumption proved wrong
- Technical approach unworkable
- Can't fit appetite after cuts

Killing is discipline, not failure.

## Commands

```bash
# Create new pitch
touch docs/pitches/feature-name.md

# Check cycle status
cat docs/cycles/TRACKER.md

# Start new cycle
cp docs/cycles/TEMPLATE.md docs/cycles/2026-q1-cycle.md
```
