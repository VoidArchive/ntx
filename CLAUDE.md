# NTX Development Guide

## Product

Thesis and vocabulary: `PRODUCT.md`, `GLOSSARY.md`. Decisions: `docs/adr/`. Current build: spec issue #20 and its sub-issues. Visual work starts from the page structure in the spec, never from a style roll.

## Code Rules

- Early returns, no `else` blocks
- const over let (except Svelte runes).
- try/catch only at API boundaries (load functions, form actions).
- No `any` type
- Comments explain why, not what

Judgement rules the reviewer enforces: `CODING_STANDARDS.md`.

## Workflow

Build features as vertical slices: migration + SQLC query, proto RPC, handler, UI. Commands: `mise tasks`. Regenerate after any `.proto` change (`mise run proto`) or SQL query change (`mise run sqlc`); SQLC queries need a `-- name: X :one` comment.

## Stack

ConnectRPC · SQLite + SQLC · SvelteKit · shadcn/svelte. Web requests never hit go-nepse; all reads come from the database:

```
go-nepse --> Background Worker --> SQLite --> ConnectRPC --> SvelteKit
```

Data gotchas the schema does not confess (annualized EPS, year-to-date profit, fiscal-year formats, which syncs run): `docs/data.md`. Row counts and freshness: `mise run data-check`.

NEPSE trades 11:00-15:05 NPT, Sunday-Thursday.

## Agent skills

### Issue tracker

Issues live in GitHub Issues for VoidArchive/ntx, via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default vocabulary: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` and `docs/adr/` at the repo root. Research notes: `docs/research/`. See `docs/agents/domain.md`.
