# NTX Development Guide

## Project Goal

Build the best NEPSE stock screener with dense, useful company pages. Web-only, no CLI.

## How Claude Should Help

The human is learning by building. Claude's role:

1. **Explain** - Answer questions about concepts, not write code
2. **Review** - Point out issues in code the human wrote
3. **Hint** - When stuck, give direction, not solutions
4. **Never write code** - Unless explicitly asked to fix a specific bug

The human writes all the code. 

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
make migrate-create NAME=add_companies
# Write migration SQL
# Write SQLC query
make sqlc

# 2. Proto
# Add GetCompany RPC to proto/ntx/v1/company.proto
make proto

# 3. Backend
# Implement handler in cmd/ntxd/

# 4. Frontend
# Build the UI that calls the RPC
```

## Quick Commands

```bash
# Backend + Frontend dev (hot reload)
make dev

# Regenerate after proto changes
make proto

# Regenerate after SQL query changes  
make sqlc

# Database migrations
make migrate-create NAME=name
make migrate
make migrate-down

# Test
make test

# Lint
make lint
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

- **Proto changes need regeneration** - `make proto` after any .proto change
- **SQLC queries need comments** - `-- name: GetCompany :one`
- **Market hours** - NEPSE is 11:00-15:05 NPT, Sun-Thu
