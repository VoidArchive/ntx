# Coding standards

Judgement rules for review. Mechanical rules live in the linters (see #2).

- **A rendered figure shows its period and source, or prints "not recorded".** A reviewer rejects a bare number on any page. Decision: `docs/adr/0002-no-figure-without-provenance.md`.
- **Claims read as facts, never as advice.** A sentence generated from data states the value, the comparison and the period. Words that recommend (buy, sell, bullish, bearish, target, signal, score) have no place in user-facing copy; the banned-word check in lint catches the words, the reviewer catches the tone.
- **Story logic lives in the Go `story` package.** A computed sentence, verdict or threshold appearing in Svelte is a defect. Decision: `docs/adr/0001-story-computed-in-backend.md`.
- **Templates are tested with exact strings** on fixtures captured from real rows, so a wording change is a deliberate test change.
