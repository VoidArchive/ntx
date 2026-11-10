---
status: accepted
---
# Personal state lives in SQLite behind one owner token

Holdings, transactions, gates, rules and bank figures the owner types are rows in the same SQLite database as the filings, and the RPCs that read or write them require one secret, `NTX_OWNER_TOKEN`, from the API environment, pasted once into the browser. We chose this over local-first browser storage because the engine evaluates gates in Go against the filings (ADR 0001), so holdings must reach the API; because the owner reads NTX from a laptop and a phone; and because ADR 0003 forbids a user model, which one token honours while a login would not. The cost is that a leaked token exposes the holdings until the secret is rotated, and that a second owner needs a second deployment. Public RPCs never read the personal tables, so a visitor without the token sees the same NTX as before.
