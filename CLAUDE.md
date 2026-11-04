# NTX Development Guide

## Project Goal

Build the best NEPSE stock screener with dense, useful company pages.
The Company page should tell a story.

## Code Rules

- Early returns, no `else` blocks
- const over let (except Svelte runes).
- try/catch only at API boundaries (load functions, form actions).
- No `any` type
- Comments explain why, not what

## Development Workflow

### Feature Cycle (Full Vertical Slice)

Build features in complete cycles, not layers. For each feature:

1. **Database** - migration + SQLC query
2. **Proto** - define the RPC
3. **Backend** - implement handler
4. **Frontend** - build the UI

Keep the context fresh

### Example: Adding "Get Company" feature

```bash
# 1. Database
mise run migrate-create add_companies
# Write migration SQL
# Write SQLC query
mise run sqlc

# 2. Proto
# Add GetCompany RPC to proto/ntx/v1/company.proto
mise run proto

# 3. Backend
# Implement handler in cmd/ntxd/

# 4. Frontend
# Build the UI that calls the RPC
```

## Quick Commands

```bash
# Backend + Frontend dev (hot reload)
mise run dev

# Regenerate after proto changes
mise run proto

# Regenerate after SQL query changes  
mise run sqlc

# Database migrations
mise run migrate-create name
mise run migrate-up
mise run migrate-down

# Test
mise run test

# Lint
mise run lint
```

## Stack

| Component | Library |
|-----------|---------|
| RPC | ConnectRPC |
| Database | SQLite + SQLC |
| Web | SvelteKit |
| UI | shadcn/svelte |


## Data Flow

```
go-nepse --> Background Worker --> SQLite --> ConnectRPC --> SvelteKit
                                     ^
                                     |
                              Single source of truth
```

Web requests never hit go-nepse directly. All reads come from the database.

## Common Gotchas

- **Proto changes need regeneration** - `mise run proto` after any .proto change
- **SQLC queries need comments** - `-- name: GetCompany :one`
- **Market hours** - NEPSE is 11:00-15:05 NPT, Sun-Thu

## Agent skills

### Issue tracker

Issues live in GitHub Issues for VoidArchive/ntx, via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
