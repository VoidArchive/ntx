# Coding standards

Judgement rules for review. Mechanical rules live in the linters: eslint enforces `svelte/prefer-const`, `no-else-return`, `no-explicit-any` and the banned-word check on Svelte text; golangci-lint's revive `indent-error-flow` flags an `else` after a return, and `scripts/banned_words.sh` checks Go string literals.

- **A rendered figure shows its period and source, or prints "not recorded".** A reviewer rejects a bare number on any page. Decision: `docs/adr/0002-no-figure-without-provenance.md`.
- **Claims read as facts, never as advice.** A sentence generated from data states the value, the comparison and the period. Words that recommend (buy, sell, bullish, bearish, target, signal, score) have no place in user-facing copy; the banned-word check in lint catches the words, the reviewer catches the tone.
- **Story logic lives in the Go `story` package.** A computed sentence, verdict or threshold appearing in Svelte is a defect. Decision: `docs/adr/0001-story-computed-in-backend.md`.
- **Templates are tested with exact strings** on fixtures captured from real rows, so a wording change is a deliberate test change.
- **Early returns are the default shape.** Handle the exceptional case and return, so the happy path reads top to bottom without nesting. The linters catch an `else` after a return; the reviewer catches nesting that should have been a guard.
- **try/catch only at API boundaries** (load functions, form actions, handlers). Everywhere else an error propagates; a catch in the middle hides the failure from the boundary that can report it.
- **Comments explain why, not what.** The code already says what it does; a comment earns its line by recording a reason the code cannot show (a data quirk, a decision, a workaround).
