# NTX

> The best company snapshot for NEPSE

A focused, information-dense stock snapshot for Nepal Stock Exchange. 

## What NTX Does
- Stock screener with fundamental filters
- Company pages that open with what the data says, period and source beside every figure
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
# Run API and web with hot reload
mise run dev

# After proto changes
mise run proto

# After SQL query changes
mise run sqlc
```

## License

MIT
