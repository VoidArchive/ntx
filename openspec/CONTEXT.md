Project NTX. 
```
Strategic sequencing (do this in order)
Solidify domain correctness
Transactions
Corporate actions
P/L math
CLI (non-TUI)
Import CSV
Show summaries
TUI
Tables
Trends
Drill-downs
AI explanations
Text output first
Local API
Enables web UI
Web UI
Only after logic is battle-tested
```


Below is the **consolidated CONTEXT** for **NTX**.  
This is the architectural ground truth. Deviating from this reintroduces fragility.

---

# NTX — Architecture Context (Canonical)

## Project intent

NTX is a **local-first NEPSE stock analyzer** with:

- portfolio management
    
- analytics (P/L, WAC, time-series)
    
- Meroshare CSV ingestion
    
- AI-based explanations (not predictions)
    
- multiple clients: CLI, TUI, Web
    

It must survive:

- NepalStock API breakage
    
- UI rewrites
    
- future expansion (web, mobile, API)
    
- offline operation
    

---

## Core principles (non-negotiable)

1. **Local-first**
    
    - Works fully offline
        
    - Remote APIs are enhancements, not dependencies
        
2. **UI is disposable**
    
    - Logic never lives in CLI/TUI/Web
        
    - All interfaces are thin clients
        
3. **Unstable upstream isolation**
    
    - NepalStock reverse-engineered auth is treated as volatile
        
    - Breakage must be contained to one adapter
        
4. **Typed API boundary**
    
    - Multiple clients require a stable contract
        
    - HTML is never the API
        
5. **Permissive licensing**
    
    - `go-nepse`: MIT
        
    - `ntx`: MIT
        
    - No GPL anywhere
        

---

## Architectural style

**Hexagonal architecture (Ports & Adapters)**  
combined with selective use of `internal/` for protection.

Hexagonal answers **dependency direction**.  
`internal/` answers **visibility**.  
They solve different problems and are used together.

---

## Dependency invariant (must always hold)

```
apps → ports → core
apps → adapters → ports
core → NOTHING
```

If this invariant breaks, the architecture has failed.

---

## Monorepo (required)

Single repo, single source of truth, atomic refactors.

---

## Canonical folder structure

```
ntx/
├─ proto/                       # ConnectRPC contracts (shared)
│   └─ ntx/v1/
│       ├─ portfolio.proto
│       ├─ analytics.proto
│       └─ ai.proto

├─ core/                        # PURE domain (no I/O, no UI, no deps)
│   ├─ portfolio/               # holdings, transactions, WAC, P/L
│   ├─ analytics/               # returns, risk, aggregates
│   ├─ timeseries/              # price series, rollups
│   ├─ ai/                      # insight generation (logic only)
│   └─ domain.go

├─ ports/                       # Interfaces the core depends on
│   ├─ marketdata.go
│   ├─ transactionsource.go
│   ├─ repository.go
│   └─ ai_provider.go

├─ adapters/                    # Replaceable implementations
│   ├─ nepse/                   # uses go-nepse (volatile)
│   ├─ meroshare/               # CSV ingestion
│   ├─ sqlite/                  # local persistence
│   ├─ duckdb/                  # optional analytics backend
│   ├─ ai_openai/               # or local LLM adapter
│   └─ cache/                   # cache + singleflight

├─ internal/                    # Application glue (NOT business logic)
│   ├─ composition/             # dependency wiring (composition root)
│   ├─ config/                  # env, flags, files
│   └─ logging/

├─ apps/                        # Delivery mechanisms
│   ├─ cli/
│   │   └─ main.go              # Cobra / text output
│   ├─ tui/
│   │   └─ main.go              # Bubble Tea (thin)
│   ├─ api/
│   │   └─ main.go              # ConnectRPC server (localhost)
│   └─ canary/
│       └─ main.go              # NepalStock breakage detector

├─ web/                         # Web client (JS/TS)
│   └─ src/

├─ go.mod
├─ README.md
└─ LICENSE
```

---

## Role of each layer

### `core/`

- Deterministic finance logic
    
- Domain types you own
    
- Fully testable
    
- Imports nothing
    

### `ports/`

- Interfaces only
    
- Describe _what the core needs_, not how
    

### `adapters/`

- Concrete implementations
    
- NepalStock, CSV, DB, AI, cache
    
- Safe to replace or delete
    

### `internal/composition/`

- Dependency wiring only
    
- Chooses which adapter fulfills which port
    
- One file per app (`api`, `cli`, `tui`)
    

### `apps/`

- Entry points
    
- Zero business logic
    
- UI concerns only
    

---

## Web integration (correct model)

- Web consumes **ConnectRPC**
    
- Browser talks to **local NTX API**
    
- Same API can serve TUI, CLI, AI, future mobile
    

```
Web → ConnectRPC → apps/api → ports → core
```

HTMX is optional and peripheral only.  
It is never the primary interface.

---

## Data & storage

- SQLite first (local-first)
    
- DuckDB optional later for analytics
    
- Cache + `singleflight` to prevent thundering herd
    
- Always serve stale data if upstream fails
    

---

## AI usage (bounded)

AI is used to:

- explain portfolio behavior
    
- summarize analytics
    
- surface anomalies
    

AI does **not**:

- predict prices
    
- give buy/sell signals
    
- override deterministic logic
    

---

## Upstream risk strategy (NepalStock)

- Wire protocol isolated in adapter
    
- Domain never sees raw payloads
    
- Canary detects breakage
    
- Offline mode preserves value
    
- Adapter replacement is cheap
    

---

## Package naming rules

- Name by **domain meaning**, not format
    
    - `meroshare`, not `csv`
        
    - `analytics`, not `graphs`
        
- `internal/` only for glue, never logic
    
- `pkg/` avoided unless public reuse is proven
    

---

## Licensing (final)

- go-nepse → MIT
    
- ntx → MIT
    
- All shared parsers / contracts → MIT
    

No GPL anywhere in the stack.

---

## One invariant to remember

**Logic that lives below the UI outlives every interface.**

NTX is a **core engine with many clients**, not a UI product.