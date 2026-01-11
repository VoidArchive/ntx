# ShareSansar NAV Extraction - Debug Summary

## What Works

### 1. Fetching Open-End Fund List ✅
```
GET https://www.sharesansar.com/mutual-fund-navs?type=2
```
- Returns 12 open-end funds with structured NAV data
- Saved to `data/funds.json`
- **This gives us the basic NAV numbers already!**

### 2. Image URL Pattern ✅
```
https://content.sharesansar.com/photos/shares/announcement/{timestamp}-{filename}.jpg
```

## What Doesn't Work

### Company Announcements API ❌
```
POST https://www.sharesansar.com/company-announcements
```

**Attempts:**
1. Simple POST → 419 CSRF error
2. Session + CSRF from `input[name=_token]` → 419 error
3. Session + CSRF from `meta[name=_token]` → Returns empty `{"data":[]}`
4. Added `company`, `symbol`, `sector` params → Still empty

**The API expects these params (from JS):**
```javascript
data: {
    "company": $("#companyid").html(),  // e.g., "698"
    "symbol": $("#symbol").html(),       // e.g., "NIBLSF"
    "sector": $("#sector").html(),       // e.g., "Mutual Fund"
}
headers: {
    'X-CSRF-Token': $('meta[name=_token]').attr('content')
}
```

**Possible issues:**
- Missing cookie/session state
- Additional hidden params in DataTables request
- Server-side validation we're not replicating

## Simpler Solution: Claude Vision

Instead of fighting the API, just:
1. Manually get announcement URLs (or scrape with Playwright)
2. Download the images
3. Use Claude Code with vision to extract data

**Benefits:**
- No OCR library needed (EasyOCR = 1.5GB PyTorch)
- Claude can read Nepali + translate + structure in one prompt
- More accurate than traditional OCR for complex tables
- Cost: ~$0.015/image × 12 funds = $0.18/month

## Files Created

```
scripts/
├── main.py           # Current scraper (partially working)
├── data/
│   └── funds.json    # ✅ 12 open-end funds with NAV data
├── images/           # (empty - image download not working yet)
└── DEBUG_SUMMARY.md  # This file
```

## Next Steps

### Option A: Fix the API (Complex)
- Use Playwright to actually render the page and intercept requests
- Debug the exact headers/cookies needed

### Option B: Claude Vision (Simple) ✅ Recommended
1. Get image URLs manually or via simple page scrape
2. Download images
3. Run Claude Code with a prompt like:

```
Read this Nepali mutual fund NAV report image and extract:
- Fund name
- Report date (convert Nepali date to English)
- NAV per unit (प्रति इकाई खुद मूल्य)
- Total units
- Net assets
- Top 10 holdings with company name and value

Return as JSON.
```

### Option C: Hybrid
- Use the working API for basic NAV data (already have this!)
- Only use Vision for detailed portfolio holdings when needed
