---
status: accepted
---
# The story is computed in Go, never in the browser

Claims, checks and strips are produced by the `story` package in the API from SQLite rows, and returned by `GetCompanyStory`. The browser only renders. We chose this over computing in SvelteKit because sector medians need every company's rows, the templates and thresholds must be unit-tested against real filings, and the same engine later evaluates personal break conditions against new filings. The cost is a proto change for every new claim kind.
