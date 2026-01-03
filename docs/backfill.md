# Backfill Optimization Summary

## Problem
Initial backfill processed ~600 companies sequentially:
- ~2,400 API calls (4 per company)
- No parallelism
- ~40 minutes estimated
- Rate limiting only on errors

## Solution: Parallel Processing with Worker Pool

Implemented 5-worker pool for all backfill functions:

### Key Changes

1. **Concurrency Control**
   - Added `sync` package import
   - `semaphore` channel limits to 5 concurrent workers
   - `sync.WaitGroup` waits for all workers
   - `sync.Mutex` protects shared counters

2. **Progress Reporting**
   - Buffered `progressChan` (size = total companies)
   - Dedicated goroutine for printing (no race conditions)
   - Real-time progress updates

3. **Rate Limiting**
   - 200ms sleep after each worker completes
   - Proactive throttling (not reactive)
   - Respects NEPSE API limits

### Parallel Functions

```go
// Pattern repeated in all 4 backfill functions:
func backfillXxx(ctx, client, queries, securities) {
    wg := sync.WaitGroup{}
    semaphore := make(chan struct{}, 5)
    progressChan := make(chan string, total)

    // Progress printer goroutine
    go func() {
        for msg := range progressChan {
            fmt.Println(msg)
        }
    }()

    for i, sec := range securities {
        wg.Add(1)
        go func(idx int, s nepse.Security) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            // Do work here
            // Send progress updates
            time.Sleep(200 * time.Millisecond)
        }(i, sec)
    }

    wg.Wait()
    close(progressChan)
}
```

## Performance Improvement

| Metric | Before | After | Improvement |
|---------|---------|--------|-------------|
| Concurrency | 1 | 5 | **5x** |
| API calls/sec | ~0.5 | ~2.5 | **5x** |
| Estimated time | ~40 min | ~8-10 min | **4x** |
| Rate limiting | Reactive | Proactive | Safer |

## Usage

```bash
# Full backfill (all data types, parallel)
./bin/ntx backfill

# Backfill only prices
./bin/ntx backfill --prices

# Backfill prices + reports
./bin/ntx backfill --prices --reports
```

## Output Example

```
=== Backfilling Prices (parallel, 5 workers) ===
[1/600] ADBL fetching (2024-01-01 to 2025-01-03)...
[2/600] NABIL fetching (2024-01-01 to 2025-01-03)...
[3/600] NBL fetching (2024-01-01 to 2025-01-03)...
[4/600] EBL fetching (2024-01-01 to 2025-01-03)...
[5/600] SCB fetching (2024-01-01 to 2025-01-03)...
[1/600] ADBL done (251 records)
[6/600] NICL fetching...
...
Prices: 145,230 new records, 12 skipped, 5 errors
```

## Safety Features

1. **Error Isolation**: One failed request doesn't block others
2. **Rate Limiting**: 200ms sleep per request = ~2.5 req/s = 150 req/min
3. **Safe Concurrency**: 5 workers won't overwhelm NEPSE API
4. **Progress Tracking**: Real-time updates, no race conditions
5. **Graceful Degradation**: Logs errors, continues processing

## Tradeoffs

### What We Gained
- ✅ **4x faster** backfill
- ✅ Real-time progress updates
- ✅ Proactive rate limiting
- ✅ No race conditions
- ✅ Minimal code changes

### What We Kept
- ✅ Same API endpoints (no batch API needed)
- ✅ Same data structures
- ✅ Same error handling patterns
- ✅ Idempotent operations (can re-run safely)

### What We Accept
- ⚠️ 5x load on NEPSE API (still within safe limits)
- ⚠️ Slightly more complex code (goroutines, channels)
- ⚠️ Requires go-sqlite to handle concurrent writes

## Next Steps

1. **Test backfill** on small subset first:
   ```bash
   # Run with first 10 companies to verify
   ./bin/ntx backfill  # Edit code to limit securities
   ```

2. **Monitor for rate limits**: Watch for 429 errors

3. **Adjust if needed**:
   - If no rate limits → increase to 10 workers
   - If rate limits → decrease to 3 workers or increase sleep to 300ms

4. **After backfill complete**, worker sync jobs handle updates automatically


## Issue 
Nepse has 613 companies, 542 active. But when doing the backfill my companies table had 719 rows, something not right. 
We also don't need the Promoter share companies.. need to filter on that as well.. we can reduce the API calls that way. Also, we don't need the companies that are closed or merged. It's useless data that can and should filtered. 
When doing the API call i have found that, mutual funds are also being filled.
we don't need mutual fund as well. Have to code review it deeply. 
The active companies I have found to be are 323. That's our target goal. 

Remove mutual funds as sectors we don't need it. 

Removing Mutual fund, it will be 264, that's our Target. 

264 are the listed companies, what's a concrete way to filter this? 
A json file for 264 companies, with it's description, logo_url and sector, feels okay here, the json can be updated with backfill.. We can then simplify the database call here. 

Simplify, simplify. 
Do one thing well. One thing and One thing really well. Instead of doing many things. Features creep is getting to us. We are going to be the companies best screener.
Not mutual fund, and no any fucking thing that everyone does. no promoter share, no mutual fund, no bonds or debenture. Companies and companies alone. 
Json might be infecient, i think we need to fix our backfill. 
