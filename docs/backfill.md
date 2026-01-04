# Backfill Strategy

## Scope

**261 companies.** That's it.

Filter criteria:
- `instrumentType == "Equity"` 
- `status == "Active"`

This excludes:
- Mutual funds
- Promoter shares
- Bonds/debentures
- Closed/merged companies

## How to Filter

NEPSE API returns all securities. Filter in Go before processing:

```go
// Only process equity companies
for _, sec := range securities {
    if sec.InstrumentType != "Equity" {
        continue
    }
    if sec.Status != "Active" {
        continue
    }
    // Process this company
}
```

Don't store what you don't need. The database should only ever have 261 rows in `companies`.

## Backfill Process

For each of the 261 companies:
1. Company details (name, sector)
2. Fundamentals (PE, EPS, book value)
3. Historical prices (OHLCV)

### Rate Limiting

- 5 concurrent workers
- 200ms delay between requests
- ~2.5 requests/second

### Estimated Time

| Data | API calls | Time |
|------|-----------|------|
| Companies | 261 | ~2 min |
| Fundamentals | 261 | ~2 min |
| Prices | 261 | ~2 min |
| **Total** | ~783 | **~6-8 min** |

## After Backfill

Background worker handles daily updates:
- New prices during market hours
- Fundamentals refresh daily
- Company list rarely changes
