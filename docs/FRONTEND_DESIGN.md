# NTX Frontend Design Brief

## Vision

Build a NEPSE stock storyteller, not a screener.

When someone visits a company page, they should feel like they're reading a story about that company - not staring at a spreadsheet. The data should unfold as a narrative, one insight at a time.

**Inspiration**: [stocktapper.com](https://stocktapper.com)

---

## The Problem with Stock Screeners

Most screeners dump everything on screen:

```
Price: 450 | Change: +2.5% | EPS: 25.6 | P/E: 17.5 | Book Value: 180 | Volume: 50,000 | 52W High: 600 | 52W Low: 380 | Market Cap: 45B | ...
```

This is **data**, not **insight**. Users have to do the mental work of figuring out what it means.

---

## The Storytelling Approach

Instead of showing everything at once, **reveal the story progressively** as users scroll. Each section answers one question and leads to the next.

### Story Structure

```
1. WHO is this company?
   ↓
2. WHERE is the price now? (context, not just number)
   ↓
3. HOW did it get here? (price journey)
   ↓
4. WHY is it valued this way? (fundamentals)
   ↓
5. WHAT does this mean? (the verdict)
```

---

## Page Flow: Company Story

### Act 1: The Introduction

> **NABIL**
> Nabil Bank Limited
>
> Nepal's first private sector bank, established in 1984.
> Commercial Banking · Active

Simple. Clean. No numbers yet. Just context.

---

### Act 2: The Current State

> **Rs. 450**
>
> *Trading 25% below its 52-week high of Rs. 600*
> *but 18% above its low of Rs. 380*

One number. One insight. The user immediately knows: "It's closer to the bottom than the top."

---

### Act 3: The Journey (Price Chart)

Full-width area chart. Let it breathe.

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                         ╱╲                                  │
│                        ╱  ╲     ← "Hit 600 in Chaitra"      │
│                       ╱    ╲                                │
│          ╱╲__________╱      ╲___                            │
│         ╱                       ╲____                       │
│   _____╱                             ╲_____  ← "You are here│
│                                                             │
│   Baisakh    Jestha    Ashadh    Shrawan    Bhadra         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Below the chart, a sentence:**

> "After peaking in Chaitra 2080, NABIL has pulled back 25% over 4 months on lower trading volumes. The decline mirrors the broader banking sector correction."

Not just a chart. A story about the chart.

---

### Act 4: The Fundamentals Story

Don't show a table. Show a narrative with supporting visuals.

#### Section: Earnings

> **Earnings are growing faster than the stock price.**
>
> EPS increased 15% this year while the price fell 25%.
> This means the stock is getting *cheaper* relative to what it earns.

```
P&L Bar Chart (simple, 4-5 years)

   2077    2078    2079    2080
    ██      ██      ██      ████
    ██      ██      ████    ████
    ██      ████    ████    ████
    ████    ████    ████    ████
```

> "Four consecutive years of profit growth."

---

#### Section: Valuation

> **How expensive is it?**
>
> P/E Ratio: 17.5
>
> *Cheaper than 70% of commercial banks.*

Radar chart comparing to sector average:

```
              EPS
               ●
              /|\
             / | \
            /  |  \
     P/E  ●----+----●  Book Value
            \  |  /
             \ | /
              \|/
               ●
          Profit Growth

     ── NABIL (solid)
     -- Sector Avg (dashed)
```

> "NABIL scores above sector average on earnings and book value, but trades at a lower valuation - suggesting the market may be underpricing its fundamentals."

---

### Act 5: The Verdict

> **The Story So Far**
>
> NABIL is a consistent performer trading at a discount to its recent highs. With growing earnings and a below-average P/E, the pullback may present an opportunity - or reflect concerns about the banking sector's near-term outlook.
>
> *This is not investment advice.*

---

## Design Principles

### 1. One Idea Per Screen

Don't cram. Let each insight have its moment.

```
BAD:  Price + Change + Volume + 52W Range + EPS + P/E + ... (all visible)
GOOD: Price → scroll → Chart → scroll → Earnings → scroll → Valuation
```

### 2. Words Before Numbers

Lead with the insight, support with data.

```
BAD:  "P/E: 17.5"
GOOD: "Cheaper than 70% of peers" (P/E: 17.5)
```

### 3. Charts That Speak

Never show a chart without explaining what it means.

```
BAD:  [Chart]
GOOD: [Chart]
      "After a volatile first half, the stock has stabilized around Rs. 450"
```

### 4. Progressive Disclosure

Reveal complexity gradually. Start simple, add depth as users scroll.

### 5. Generous Whitespace

Let the story breathe. Dense data feels like work. Spaced content feels like reading.

---

## Charts Specification

### Price History (Area Chart)

**Data**: `price.getPriceHistory({ symbol, days: 365 })`

- Full width, generous height
- Gradient fill (brand color → transparent)
- Subtle grid lines
- Key points annotated (52W high, 52W low, significant moves)
- Time range toggles: 1M | 3M | 6M | 1Y
- Hover: Show OHLC + Volume for that day

---

### Profit & Loss (Bar Chart)

**Data**: `company.getFundamentals({ symbol })` → `history[].profitAmount`

- Vertical bars, one per fiscal year
- Color: Green if growth YoY, Red if decline
- Show 4-5 years max
- Label each bar with the fiscal year (2077, 2078, etc.)
- Hover: Show exact profit amount + YoY change %

---

### Fundamentals (Radar Chart)

**Data**: Computed from `getFundamentals()` normalized against sector

**Axes** (5 points):
- EPS (higher = better)
- Book Value (higher = better)
- Profit Growth (higher = better)
- P/E Ratio (lower = better, invert for display)
- Price to Book (context dependent)

**Display**:
- Company: Solid filled polygon
- Sector Average: Dashed outline polygon
- Normalize all values to 0-100 scale for comparison

---

## API Reference

### CompanyService

```typescript
// List companies (for screener/search)
company.listCompanies({
  sector?: Sector,
  query?: string,
  limit?: number,
  offset?: number
}) → { companies: Company[] }

// Get single company
company.getCompany({ symbol: string }) → { company: Company }

// Get fundamentals history
company.getFundamentals({ symbol: string }) → {
  latest: Fundamental,
  history: Fundamental[]
}
```

### PriceService

```typescript
// Latest price
price.getPrice({ symbol: string }) → { price: Price }

// Price history (default 365 days)
price.getPriceHistory({
  symbol: string,
  days?: number
}) → { prices: Price[] }
```

---

## Data Models

```typescript
interface Company {
  id: number
  name: string
  symbol: string
  status: 'ACTIVE' | 'SUSPENDED' | 'DELISTED'
  email?: string
  website?: string
  sector: Sector
  instrumentType: 'EQUITY'
}

interface Price {
  businessDate: string    // "2080-05-15"
  open?: number
  high?: number
  low?: number
  close?: number
  ltp?: number            // last traded price
  previousClose?: number
  change?: number
  changePercent?: number
  volume?: number
  turnover?: number
  trades?: number
}

interface Fundamental {
  fiscalYear: string      // "2080/81"
  quarter?: string        // "Q1", "Q2", "Q3", "Q4"
  eps?: number            // earnings per share
  peRatio?: number
  bookValue?: number
  paidUpCapital?: number
  profitAmount?: number
}

type Sector =
  | 'COMMERCIAL_BANK'
  | 'DEVELOPMENT_BANK'
  | 'FINANCE'
  | 'MICROFINANCE'
  | 'LIFE_INSURANCE'
  | 'NON_LIFE_INSURANCE'
  | 'HYDROPOWER'
  | 'MANUFACTURING'
  | 'HOTEL'
  | 'TRADING'
  | 'INVESTMENT'
  | 'OTHERS'
```

---

## Story Generation

The app needs to generate narratives, insights, and verdicts from raw data. **You decide the approach.**

### The Challenge

The backend provides numbers:
- Price: 450
- 52W High: 600
- 52W Low: 380
- EPS: 25.6 (was 22.3 last year)
- P/E: 17.5 (sector avg: 22)

The frontend needs to turn this into:
> "Trading 25% below 52-week high despite 15% earnings growth. At a P/E of 17.5, it's cheaper than 70% of banks."

### Approach 1: Rule-Based Templates

Write conditional logic that maps data patterns to sentences.

**How it works:**
- Define thresholds (e.g., "if > 20% below 52W high, say 'significant pullback'")
- Create sentence templates with placeholders
- Combine multiple signals into overall verdicts

**Pros:**
- Fast (computed client-side, no API calls)
- Free (no external services)
- Predictable (same data always produces same story)
- Easy to debug and adjust

**Cons:**
- Limited vocabulary (can feel repetitive)
- Manual effort to cover edge cases
- Won't catch nuanced situations

**Example patterns to handle:**
- Price near high vs near low vs mid-range
- Earnings growing vs declining vs flat
- Valuation cheap vs expensive vs fair vs sector
- Volume spike vs normal vs declining
- Combining signals: cheap + growing = opportunity, expensive + slowing = caution

---

### Approach 2: LLM-Generated

Send data to an AI model and let it write the narrative.

**How it works:**
- Collect all data points for a company
- Send to Claude/GPT with a prompt asking for analysis
- Cache the response (regenerate daily or on data change)

**Pros:**
- Natural, varied language
- Can catch nuanced patterns
- Handles edge cases gracefully
- Can adjust tone/style via prompt

**Cons:**
- Latency (API call required)
- Cost (per-request pricing)
- Needs caching strategy
- Less predictable output

---

### Approach 3: Hybrid

Use templates for common patterns, LLM for deeper analysis.

**How it works:**
- Quick verdicts (price context, simple ratios) → templates
- "Full Analysis" or "Deep Dive" section → LLM-generated
- Cache LLM responses aggressively

---

### Recommended Starting Point

Start with **rule-based templates**. Cover these scenarios:

**Price Story:**
- Where is it in the 52W range? (near high / near low / middle)
- Recent trend? (up X% in Y days / down / flat)
- Volume story? (higher than average / drying up)

**Earnings Story:**
- YoY growth rate (surging / solid / modest / declining)
- Trend over multiple years (consecutive growth / turnaround / deteriorating)

**Valuation Story:**
- P/E vs sector average (undervalued / premium / fair)
- Price-to-book context
- Earnings yield vs alternatives

**Overall Verdict:**
Combine 2-3 signals into one sentence:
- Cheap + Growing = "Potential opportunity"
- Expensive + Slowing = "Proceed with caution"
- Near lows + Stable earnings = "Market may be overreacting"
- Near highs + Accelerating growth = "Momentum intact"

You can cover 80% of cases with ~20-30 well-crafted templates. The stories won't be poetry, but they'll be accurate and useful.

---

### Structure Suggestion

```
lib/
  story/
    price.ts        # price position, trend, volume insights
    earnings.ts     # EPS growth, profit trend
    valuation.ts    # P/E context, sector comparison
    verdict.ts      # combine signals → overall narrative
    index.ts        # main entry: generateStory(data) → Story
```

Each module exports functions that take data and return strings. The verdict module combines them into a cohesive narrative.

---

## Pages

### `/company` - The Screener

Grid of company cards. Each card is a mini-story:

```
┌────────────────────────────┐
│  NABIL                     │
│  Nabil Bank Limited        │
│                            │
│  Rs. 450  ↑ 2.5%          │
│                            │
│  "25% below 52W high,      │
│   but earnings growing"    │
└────────────────────────────┘
```

- Filter by sector
- Search by name/symbol
- Cards link to full story

### `/company/[symbol]` - The Story

The full narrative experience described above.

---

## Tech Stack

- **Framework**: SvelteKit
- **UI Components**: shadcn-svelte
- **Charts**: Chart.js (via shadcn)
- **Styling**: Tailwind CSS
- **API**: ConnectRPC

---

## Remember

> We're not building a terminal for traders.
> We're building a storyteller for people who want to understand a company in 60 seconds.
> We should use tailwind css and shadcn-svelte when possible. Style block are only allowed if it makes sense. 
> Design should use #FBF7EB as background and 

Every design decision should answer: **"Does this help tell the story?"**

