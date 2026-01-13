# NTX

> The best company snapshot for NEPSE

A focused, information-dense stock snapshot for Nepal Stock Exchange. 

## What NTX Does
- Stock screener with fundamental filters
- Company pages with dense, actionable data
- Market overview (indices, top movers)

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go + ConnectRPC |
| Database | SQLite + SQLC |
| Frontend | SvelteKit + Tailwind + shadcn |
| Data | go-nepse |

## Development

```bash
# Run backend (hot reload)
make dev

# Run frontend
make dev-web

# After proto changes
make proto

# After SQL query changes
make sqlc
```

## Project Structure

```
ntx/
├── cmd/ntxd/           # Server binary
├── internal/           # Go packages
├── gen/                # Generated code (proto)
├── proto/              # Proto definitions
└── web/                # SvelteKit frontend
```

## License

MIT
