# Stock-page survey: what makes a company page tell a story

Date: 2026-11-04. Purpose: input to the NTX redesign. The owner's complaint is that every
mock "looks like a broker website or nepalstock.com". This document records what the most
distinctive company pages and screeners worldwide actually do, verified on the products
themselves where possible, and maps the mechanisms onto the data NTX already has
(`fundamentals`: eps, pe_ratio, book_value, paid_up_capital, profit_amount by fiscal_year/quarter;
`ownership`: promoter_percent, public_percent; `corporate_actions`: bonus_percentage,
right_percentage, cash_dividend by fiscal_year; daily OHLCV; sector stats; mutual-fund NAV and holdings).

Method note. Pages were read live on 2026-11-04 via HTTP fetch. Several sites (Simply Wall St,
Stockopedia, Trendlyne, Macrotrends, Status Invest, Minkabu, Naver, Xueqiu, moomoo, Koyfin app,
Chukul, nepalstock.com) block non-browser clients or render only in JavaScript. For those, claims
come from the vendor's own help centre, open-source model repo, or indexed vendor pages, and are
marked **[docs]** or **[unverified]**. No listicle was used as a source for a product claim.

---

## 1. Product notes, by region

### Global / US

**Simply Wall St** — [snowflake help](https://support.simplywall.st/hc/en-us/articles/360001740916-How-does-the-Snowflake-work) [docs], [open analysis model](https://raw.githubusercontent.com/simplywallst/company-analysis-model/master/MODEL.markdown), [Narratives feature page](https://www.simplywall.st/features/community-narratives).
Leads with the Snowflake: five axes (Value, Future, Past, Health, Dividends), each 0–6 passed checks; "the greater the total number of successful checks the larger the snowflake" and the more green. Every check is a yes/no sentence from a public model, e.g. "Is the PE ratio less than the market average but still greater than 0?", "Have Earnings Per Share (EPS) increased in past 5 years?", "Has the dividend been volatile in the past 10 years?", "Are dividends paid well covered by Net Profit?" Narratives are user-written theses that attach a fair value and "Revenue growth p.a" to a stock and are shown next to the current price as "undervalued/overvalued %". Paid for full detail. Light theme.
*Borrow:* a published, boolean checklist computed from statements. NTX has EPS/book-value/profit history and dividends by fiscal year, enough for Past, Value (PE, PB vs sector) and Dividend axes; Future and Health need data NTX lacks.

**Stockopedia** — [styles article](https://stockopedia.com/academy/articles/a-style-classification-for-every-stock) and [StockRank Style ratio page](https://www.stockopedia.com/ratios/stockrank-style-5541) [docs; page blocked].
Header of every StockReport carries a 0–100 StockRank (Quality, Value, Momentum) and a one-phrase *style* label: Super Stock (all three ranks > 60), High Flyer, Contrarian ("good & cheap, but declining"), Turnaround; losing styles Sucker Stock (all < 40), Momentum Trap, Value Trap, Falling Star; else Neutral. Paid.
*Borrow:* a named archetype derived from two or three ranks. NTX can compute Value (PE, PB percentile within sector) and Momentum (1y/6m return percentile) today; a "Quality" proxy is EPS stability + ROE (profit_amount / (book_value × shares)).

**Morningstar** — [live quote page](https://www.morningstar.com/stocks/xnas/aapl/quote) (fetched), [help centre](https://www.morningstar.com/help-center/stocks/stock-quote-page).
First screen: star rating (locked), then "Price vs Fair Value" with one sentence — "AAPL is trading within a range we consider fairly valued." — then locked Fair Value, Uncertainty, 1-Star/5-Star Price, Economic Moat, Capital Allocation. Then "Bulls Say, Bears Say" as two short paragraphs. Analyst-written, mostly paywalled. Light, airy.
*Borrow:* the one-sentence verdict pattern and the Bulls/Bears pairing. NTX cannot do analyst fair value, but can state "trades at X× book, vs sector median Y×" as the equivalent single sentence.

**StockAnalysis.com** — [AAPL overview](https://stockanalysis.com/stocks/aapl/).
Quote, then a compact 18-cell stats grid (Market Cap … Earnings Date), then an "About" block, then two auto-written paragraphs: "Financial Performance" ("Earnings were $132.18 billion, an increase of 18.00%") and "Analyst Summary" ("According to 43 analysts, the average rating for AAPL stock is 'Buy.'"). Sub-pages: Financials, Forecast, Statistics, Metrics, Dividends, History. Light/dark toggle. Mostly free.
*Borrow:* short templated prose paragraph beneath the grid. Fully feasible from NTX quarterly EPS/profit.

**Finviz** — [AAPL quote](https://finviz.com/quote.ashx?t=AAPL).
A ~100-cell stat table (valuation, growth, ownership, margins, performance 1W–10Y) above news and insider table. No narrative. Free with Elite upsell.
*Borrow:* nothing stylistically — this is the "terminal grid" look the owner is trying to escape. The useful idea is colouring each cell by direction, not the grid itself.

**TradingView symbol page** — [NASDAQ-AAPL](https://www.tradingview.com/symbols/NASDAQ-AAPL/).
Chart first, then "Key facts today" (AI news summaries), Key stats with tooltips, Financials cards ("Revenue to profit conversion", "Debt level and coverage"), community ideas, Technicals/Analyst gauges ("Neutral"), Seasonals, auto-generated FAQ. Very long single column.
*Borrow:* financial *cards with a sentence title* ("Revenue to profit conversion") instead of table headers.

**Koyfin** — [Snapshot help](https://koyfin.com/help/global-equities-fundamentals-valuatiion) [docs; app blocked].
Distinctive is the **Percentile Rank snapshot**: a stock's metrics scored 0–100 against a peer cohort; the help example is "AAPL's P/E is higher than 71% of the US tech stocks". Paid tiers, dark-mode terminal.
*Borrow:* percentile-within-sector for every ratio. NTX has sector stats; with 112 hydropower companies the percentile is actually meaningful.

**Macrotrends** — [PE pages](https://www.macrotrends.net/stocks/charts/MAC/macerich/pe-ratio) [indexed; blocked live].
Each metric page opens with a templated sentence ("… PE ratio as of June 14, 2025 is 53.46"), a long chart, then a table of Date / Stock Price / TTM EPS / PE Ratio. One metric per page, 10+ years. Free, ad-heavy.
*Borrow:* the "one metric, whole history, one sentence" page unit.

**TIKR**, **Fiscal.ai (ex-FinChat)**, **Quartr** — [tikr.com](https://www.tikr.com/), [quartr.com](https://quartr.com/), Fiscal.ai rate-limited (third-party reviews only).
TIKR: 30-year financials, ownership, transcripts, dense terminal. Quartr: first-party IR material (calls, slides, transcripts) where "every output is traceable to its source". Fiscal.ai: segment KPIs plus AI copilot [unverified]. All paid/pro.
*Borrow:* Quartr's provenance idea — every number on NTX linking to the quarterly report it came from — fits a trust-through-transparency product.

**Public.com**, **Robinhood** — [public.com/stocks/aapl](https://public.com/stocks/aapl), [robinhood AAPL](https://robinhood.com/us/en/stocks/AAPL/).
Both lead with a buy widget / sign-up CTA, then About, Key statistics, Analyst ratings, News, "People also own". Robinhood has a "Stock Snapshot" prose block that restates the stat grid. Minimal, white, conversion-first.
*Do not borrow:* this is the broker template.

### India (closest analogue to NEPSE)

**Screener.in** — [RELIANCE consolidated](https://www.screener.in/company/RELIANCE/consolidated/).
First screen: name, price, a 9-cell ratio box (Market Cap, Current Price, High/Low, Stock P/E, Book Value, Dividend Yield, ROCE, ROE, Face Value), chart with a *Price-to-PE overlay*, then **Pros / Cons** with the disclaimer "The pros and cons are machine generated." Sample cons: "Company has a low return on equity of 8.77% over last 3 years." / "Dividend payout has been low at 10.2% of profits over last 3 years." Then Peers, Quarterly Results (13 quarters), Profit & Loss (10y + TTM) followed by **Compounded Sales Growth / Compounded Profit Growth / Stock Price CAGR / Return on Equity** each as 10Y/5Y/3Y/1Y mini-tables, Balance Sheet, Cash Flow, Ratios, Shareholding Pattern (quarterly and yearly, Promoters/FIIs/DIIs/Public/No. of Shareholders), Documents (annual reports, concalls). Free; premium for custom ratios and "Insights". Light theme, plain tables, no ads.
*Borrow:* almost everything structural. Pros/Cons from thresholds on EPS growth, dividend payout, PE vs history; the four-period compounded-growth mini-tables (NTX can do Profit, EPS, Book Value and Price CAGR across fiscal years); shareholding as a time series; "Price / PE" chart toggle.

**Tickertape** — [RELIANCE](https://www.tickertape.in/stocks/reliance-industries-RELI).
Leads with price and a six-row **Scorecard**, each row a label, a verdict and one sentence: "Performance: Low — Hasn't fared well - amongst the low performers", "Valuation: High — Seems to be overvalued vs the market average", "Growth: Low — Lagging behind the market in financials growth", "Profitability: High — Showing good signs of profitability & efficiency", "Entry point: Good — The stock is underpriced and is not in the overbought zone", "Red flags: Low — No red flag found". Then tags ("Largecap … stock is ranked 1", "Low Risk — Stock is 1.77x as volatile as Nifty"), key metrics each beside the sector average, holdings with "Promoter holding … roughly constant over six months, with insignificant pledging", a *Smallcases* list, dividend "₹4.59 per ₹1,000" phrasing, bonus/rights history under Events. Score number itself is Pro.
*Borrow:* the verdict + one-sentence row; expressing yield as "Rs X per Rs 1,000 invested"; a plain-language ownership-change sentence.

**Trendlyne** — [DVM help](https://help.trendlyne.com/support/solutions/articles/84000347982-what-is-the-trendlyne-dvm-score-and-how-does-it-work-), [stock-report structure](https://help.trendlyne.com/support/solutions/articles/84000396565-how-is-the-stock-report-structured-), public [SWOT widgets](https://trendlyne.com/web-widget/swot-widget/INFY) [docs/widgets; page blocked].
DVM = Durability, Valuation, Momentum, each 0–100 and tagged G/M/B, giving a three-letter archetype (e.g. "BGB"). SWOT lists are rule names, e.g. strength "Companies with Low Debt", weakness "Bearish Stocks - Stocks with Medium to Low Trendlyne Momentum Score". Report order: DVM summary, Checklist score, Durability, Valuation + "P/E Buy/Sell Zone", Momentum, Peers, Forecaster, Shareholding, About. Freemium.
*Borrow:* the P/E Buy/Sell Zone — current PE placed on its own historical PE range. NTX has pe_ratio per quarter and daily price + EPS, so a historical PE band is cheap.

**Moneycontrol** — [RELIANCE](https://www.moneycontrol.com/india/stockpricequote/refineries/relianceindustries/RI) (fetched).
Buried under ads: "MC SWOT" with counts ("Strengths (14): Good quarterly growth in the recent results", "Threats (1): Below 200 DMA", "Opportunities (5): Companies with Price to Earnings (PE) less than average 3 Year PE, 5 Year PE and 10 Year PE"), "MC Essentials 53% Pass", "MC Technicals: VERY BULLISH", "Stock Score 81/100 — Average Financial Strength, Average Growth Trend Stock Priced at High Valuations", and one ownership sentence "Promoter holding remains unchanged at 49.62% in Sep 2026 qtr." Heavy ads, Pro upsell.
*Borrow:* only the one-line ownership sentence; the rest is noise stacked on noise.

**Zerodha Markets** — [NSE/RELIANCE](https://zerodha.com/markets/stocks/NSE/RELIANCE).
Snapshot, key metrics incl. *P/E next to sector P/E*, then tabs Share Price, Financials, **Revenue mix**, Shareholdings, Peers, **Forensics** (several "Coming soon"). Card-and-tab, sign-up box repeated. Data powered by Tijori.
*Borrow:* the idea of a "Forensics" section name — for NEPSE this could be "Capital history" (bonus-driven dilution).

### Japan / Korea / China

**Kabutan (株探)** — [7203 top](https://kabutan.jp/stock/?code=7203), [決算 page](https://kabutan.jp/stock/finance?code=7203) (both fetched).
First screen: price, PER/PBR/利回り/信用倍率, market cap, "比較される銘柄" (compared names: ホンダ, 日産自, デンソー), next results date. Then **ヒストリカルバリュエーション** — today's PER/PBR/PSR beside the 3-year average. Then 業績推移: fiscal-year rows of 売上高/経常益/最終益/1株益/1株配 with the *company's own forecast row* marked 予 and a 前期比(%) row; a **修正履歴** (revision history) table with ↑/↓/→ arrows ("「↑」：上方修正、「↓」：下方修正"); tabs 成長性/収益性/過去最高 (record highs); 進捗率 (progress rate against full-year guidance). Headlines are templated: "三菱商事【8058】、今期最終を9％上方修正・4期ぶり最高益更新へ". Dense, light, premium for more periods.
*Borrow:* "vs own 3-year average" valuation; the record-high tab ("highest EPS since FY 20xx"); templated earnings headlines ("NABIL Q1 2083/84: profit +12% YoY, EPS annualised 33.9, highest Q1 in 5 years"). NTX's quarterly profit_amount supports all three.

**Minkabu** — [user forecast listings](https://s.minkabu.jp/pick?page=24) [vendor pages; stock page blocked].
Signature: a 株価診断 verdict (割安/割高) and 目標株価 (target price), with users posting predictions tagged "株価診断結果「割高」に賛成"; a crowd 買い:売り ratio (e.g. 86.6% : 13.4%). Opinion-first.
*Do not borrow:* crowd targets conflict with "no tips".

**Naver Finance**, **Xueqiu**, **moomoo** — all blocked for non-browser clients; no vendor documentation found describing the page. Known publicly as: Naver = compact 투자정보 box + 동일업종 비교 table + annual/quarterly 기업실적분석; Xueqiu = discussion feed first, financials behind tabs; moomoo = quote + capital-flow + institutional holdings tabs ([financial sub-pages](https://moomoo.com/stock/FUTU-US/financial/key-indicators)). **[unverified]** — treat as background only.

### Brazil / other emerging markets

**Fundamentus** — [PETR4](https://www.fundamentus.com.br/detalhes.php?papel=PETR4).
One screen of label:value pairs: Oscilações (day, month, 30d, 12m, then *each calendar year back to 2021*), Indicadores fundamentalistas in two columns (P/L, LPA, P/VP, VPA, EV/EBITDA, margins, ROIC, ROE, DY, Div.Br/Patrim, Cresc. Rec 5a), balance-sheet and 12m/3m income blocks. No chart, no prose, no ads, utilitarian. Free.
*Borrow:* the per-calendar-year return strip — for NEPSE show per-fiscal-year total return including bonus shares.

**Investidor10** — [PETR4](https://investidor10.com.br/acoes/petr4/).
Five big cards (Cotação, Variação 12M, P/L, P/VP, DY), a *real-return* table ("valores reais descontados da inflação"), a Graham fair-price card with upside % (locked), then Indicadores Fundamentalistas with "Comparar indicadores". Freemium, promotional.
*Borrow:* inflation-adjusted returns (Nepal CPI is available), and the Graham number as a transparent, formula-only "fair value" — NTX has EPS and book value.

**Status Invest** — [indexed IVZ page](https://statusinvest.com.br/acoes/eua/ivz) [blocked live]. Indexed text shows "Indicadores de Valuation: D.Y, P/L, PEG Ratio, P/VP, EV/EBITDA…" and a "Status Alpha IA" block. **[unverified]** beyond that.

**FireAnt (Vietnam)** — [VNM](https://fireant.vn/ma-chung-khoan/VNM).
Price, then a *generated company blurb* ("five-year average growth of about 16% for revenue and 12% for profit, both above the industry average"), a quick-info panel where **every multiple is shown against its industry figure** (P/E 10.5x vs 13.4x, P/B 3.48x vs 2.52x), upcoming cash dividend with ex-date, "Điểm nhấn tài chính" (financial highlights). Sections carry the label "Tự động từ số liệu. Không phải khuyến nghị đầu tư" (automatically generated from data; not investment advice). Free with membership tier.
*Borrow:* the explicit "generated from data, not advice" label and the paired "X vs industry Y" layout for every ratio. Closest spirit to NTX's principles of anything surveyed.

**Pakistan — Sarmaaya** — [OGDC](https://sarmaaya.pk/stocks/OGDC). Quote, then "Intrinsic Value / Margin of Safety / Sarmaaya Score" (Growth, Stability, Valuation, Cashflow, Inventory, each /20; premium), a "Snowflake Analysis" heading, tabs Fundamentals/Peers/Financials/Ratios/Dividends/Ownership/Insider. A Simply Wall St clone for a frontier market. PSX's own [dps.psx.com.pk](https://dps.psx.com.pk/company/OGDC) refused the fetch.

**Bangladesh — AmarStock** — [GP](https://www.amarstock.com/stock/GP). Metrics grid (LTP, P/E, NAV, EPS quarterly/annual, paid-up capital, category, rating), tabs Technical/Cash Flow/P&L/Balance Sheet/Dividend History, "Director Buy & Sell" table, "Corporate Performance at a glance". Values largely blank in the capture. stockbangladesh.com did not resolve; DSE's company page returned 410.

**Sri Lanka** — CSE company profile is a JS app (404 on direct fetch); only broker portals (CAL "AnalytiCAL") were found. Nothing distinctive verified.

### Nepal incumbents (the baseline to differ from)

**ShareSansar** — [NABIL](https://www.sharesansar.com/company/nabil). Two ad banners above the menu, then price/OHLC/52w/averages, a panel with latest dividend ("5.00% bonus, 10.8% cash, FY 2082/2083"), pivot levels S3–R3 and MA bullish/bearish flags, then 14 tabs. Chart and fundamentals behind "SS Pro". No narrative, no scores.

**MeroLagani** — [NABIL](https://merolagani.com/CompanyDetail.aspx?symbol=NABIL). Notification prompt, banner ad, then a long key-metrics list (EPS, P/E, book value, PBV, market cap, 30d avg volume), cash-dividend-by-fiscal-year panel back to FY 067-068, tabs with fiscal-year filter dropdowns. Raw figures only, crowded.

**ShareHub** — [NABIL](https://www.sharehubnepal.com/company/nabil). Price, snapshot, 13 tabs, Ownership Structure (Promoter 60% / Public 40%), "Last Dividend 15.8% (10.8% cash, 5% bonus) FY 2082/2083", three **range bars** (today / 52-week / all-time, with % distance to each end), an About paragraph with "#1 of 19 commercial banks" by market cap, FAQ ("price sits about 68% of the way through its 52-week range"). No ads, no scores. The most considered of the incumbents.

**NepaliPaisa** — [NABIL](https://nepalipaisa.com/company/NABIL). Index ticker, snapshot incl. beta, 180-day VWAP, EPS labelled "Q1, FY 83/84", P/E, BVPS, P/B; tabs. Very dense with repeated menus; app upsell.

**NEPSEAnalysis (nepsetrading.com)** — [NABIL tearsheet](https://nepsetrading.com/stockanalysis/tearsheet/NABIL). The one incumbent already doing relative context: "Key indicators vs Commercial Banks" with min / avg / max for EPS, P/E, P/B, net worth/share, ROE; last-8-quarters table marked "better/worse" vs same quarter last year; Graham number 399.59; a "MASTER score" described as "Mechanical 0–100 technical ranking"; RSI "Neutral", MACD/EMA "Bullish". Ad banner first.

**NepseAlpha, Chukul, nepalstock.com** — JS apps that returned no content. Chukul's own title is "NEPSE Live Market, Stock Analysis & Green/Red Signals"; third-party listings describe NepseAlpha as a signals + screener + premium charts product. **[unverified]**

What incumbents lead with, collectively: today's OHLC and 52-week range, moving-average bullish/bearish flags, pivot levels, latest dividend %, and a tab wall. None shows a fiscal-year time series on the first screen; none writes a sentence about the company; only NEPSEAnalysis compares to the sector.

---

## 2. Comparison table

| Product | First screen leads with | Signature feature | Auto text | Score / rank | Relative context | Density / theme | Price |
|---|---|---|---|---|---|---|---|
| Simply Wall St | Snowflake + fair value | 30 public yes/no checks | Risk/reward sentences | 5 axes 0–6 | vs market & industry | Visual, light | Freemium |
| Stockopedia | StockRank + style label | QVM archetypes ("Super Stock") | Style phrase | 0–100 ×3 | Percentile vs market | Dense report | Paid |
| Morningstar | Price vs Fair Value sentence | Moat, Uncertainty, Bulls/Bears | Analyst prose | 1–5 stars | vs fair value | Airy, light | Paid |
| StockAnalysis | Stats grid + prose | Two templated paragraphs | Yes | None | Analyst consensus | Medium, light/dark | Mostly free |
| Finviz | 100-cell grid | Everything on one grid | No | None | None | Max density | Freemium |
| TradingView | Chart, "Key facts today" | Financial cards w/ sentence titles | AI news, FAQ | Gauges | Analyst/technical | Very long | Freemium |
| Koyfin | Snapshot | Percentile Rank vs cohort | "higher than 71% of…" | 0–100 percentiles | vs sector | Terminal, dark | Paid |
| Macrotrends | One sentence + 10y chart | One metric per page | Yes | None | vs own history | Simple | Free, ads |
| Screener.in | 9 ratios + Pros/Cons | Compounded growth 10/5/3/1Y | Pros/Cons | None | Peers table, own history | Dense tables, light | Free |
| Tickertape | 6-row Scorecard | Verdict + sentence rows | Yes | Hidden (Pro) | Each metric vs sector avg | Medium | Freemium |
| Trendlyne | DVM G/M/B | P/E Buy/Sell Zone, SWOT | Rule names | 0–100 ×3 | vs own PE history, peers | Dense | Freemium |
| Moneycontrol | Ads, then SWOT counts | Stock Score /100 | One-liners | 0–100 | Peers | Chaotic | Freemium |
| Zerodha Markets | Snapshot, P/E vs sector P/E | Revenue mix, Forensics | Vendor briefs | None | Sector P/E | Cards/tabs | Free |
| Kabutan | PER/PBR + compared names | Historical valuation vs 3y avg; revision arrows; 進捗率 | Templated headlines | None | vs own history, named peers | Dense, light | Freemium |
| Minkabu | Diagnosis 割安/割高 + target | Crowd buy/sell ratio | Verdict | Crowd | vs theoretical price | Medium | Freemium |
| Fundamentus | Label:value sheet | Per-year return strip | No | None | None | Spartan | Free |
| Investidor10 | 5 big cards | Real (inflation-adj.) returns, Graham price | Little | None | Compare tool | Cards | Freemium |
| FireAnt | Price + generated blurb | Every ratio "vs industry" | Yes, labelled "auto, not advice" | None | vs industry | Clean | Freemium |
| Sarmaaya (PK) | Quote + Score/5 | Snowflake clone | Minimal | 5×/20 | Peers | Medium | Freemium |
| ShareSansar (NP) | Ads, OHLC, pivots, MA flags | SS Pro paywall | No | MA bullish/bearish | None | Cluttered | Freemium |
| MeroLagani (NP) | Ad, metrics list | Cash dividend by FY list | No | None | None | Cluttered | Free |
| ShareHub (NP) | Price, range bars | 52w/all-time range bars, "#1 of 19" | FAQ | None | Rank in sector | Medium | Free |
| NEPSEAnalysis (NP) | Ad, price, MASTER score | min/avg/max vs sector | Bullish/Neutral labels | 0–100 technical | vs sector | Medium | Free |

---

## 3. Recurring mechanisms that create "story"

1. **Machine-generated pros/cons or SWOT from thresholds.** Screener.in ("Company has a low return on equity of 8.77% over last 3 years"), Trendlyne SWOT, Moneycontrol, Simply Wall St risks/rewards, Tickertape "Red flags". *NTX data:* yes — EPS growth over N fiscal years, dividend payout (cash_dividend × paid_up / profit), PE vs own history, promoter change, bonus frequency.

2. **Checklist scoring with a published rulebook.** Simply Wall St (30 checks, open model, CC BY-NC-SA), Moneycontrol "MC Essentials 53% Pass", Sarmaaya. *NTX:* yes for Past, Value, Dividend; not Health/Future. Publishing the rules in the repo matches "built in the open".

3. **Verdict word + one sentence per dimension.** Tickertape scorecard, Morningstar "trading within a range we consider fairly valued", Trendlyne G/M/B. *NTX:* yes, as long as the words describe data (High/Low/Stable) and not actions (Buy/Sell).

4. **Named archetype from 2–3 ranks.** Stockopedia styles, Trendlyne "BGB". *NTX:* partially — Value + Momentum + Earnings-stability ranks within sector.

5. **Fair value vs price as one number.** Simply Wall St, Morningstar, Minkabu, Investidor10 Graham, NEPSEAnalysis Graham number. *NTX:* only formula-based (Graham = √(22.5 × EPS × BVPS)) — transparent, no analyst judgement. Label it as a formula, not a target.

6. **Relative context on every ratio: vs sector and vs own history.** FireAnt (P/E 10.5x vs 13.4x), Zerodha (P/E beside sector P/E), Kabutan (PER vs 3-year average), Trendlyne PE zone, Koyfin percentile, NEPSEAnalysis min/avg/max. *NTX:* yes — sector stats exist; quarterly pe_ratio gives own history.

7. **Compounded growth over several horizons.** Screener.in's 10Y/5Y/3Y/1Y mini-tables for Sales, Profit, Price, ROE. *NTX:* yes for Profit, EPS, Book Value, Price (bonus-adjusted).

8. **Guidance/revision and "record" annotations.** Kabutan's ↑/↓ revision arrows, 進捗率, 過去最高 (record high) tab, templated headlines "…4期ぶり最高益更新へ". *NTX:* no guidance in NEPSE, but "highest quarterly profit since Q3 2078/79" and "EPS up N quarters in a row" are computable.

9. **Per-period return strip.** Fundamentus per-calendar-year oscillations; Investidor10 inflation-adjusted. *NTX:* per-fiscal-year total return including bonus shares and cash — unique angle because NEPSE returns are heavily bonus-driven.

10. **Ownership change as a sentence.** Tickertape ("Promoter holding … roughly constant over six months"), Moneycontrol ("Promoter holding remains unchanged at 49.62% in Sep 2026 qtr."). *NTX:* yes once ownership is stored as a time series; plus mutual-fund holdings ("held by 7 of 23 mutual funds; NIBLSF added 40,000 units since Kartik").

11. **Provenance.** Quartr "traceable to its source"; Screener.in Documents. *NTX:* each quarterly row linking to the filed report (where the worker stores a URL) is cheap and fits the trust story.

12. **Explicit "generated from data, not advice" label.** FireAnt, Screener.in, Moneycontrol disclaimers. *NTX:* adopt verbatim in spirit.

---

## 4. Ideas for NTX, ranked by distinctiveness × feasibility

1. **Lead with a sentence, not a quote.** First screen is a short generated paragraph in the Screener.in/FireAnt register: "NABIL earned Rs 2.56 Arba in Q1 2083/84, up 12% on Q1 2082/83 and its best first quarter in five years. At Rs 550 it trades at 16× annualised EPS and 2.3× book, against a commercial-bank median of 14× and 1.6×. Promoters hold 60%; 7 mutual funds hold 1.8% of public shares. Last year's payout was 15.8% (10.8% cash, 5% bonus)." Every clause maps to an existing column. Price, change and OHLC become secondary, small. This alone breaks the broker template, where the quote is the hero.

2. **"Vs sector" and "vs own history" twin markers on every ratio.** Each of PE, PB, EPS, dividend yield shows a small dot on two horizontal strips: the sector distribution (min / median / this company / max, as NEPSEAnalysis does but visual) and the company's own 3-year range (Kabutan/Trendlyne). With 112 hydropower and 19 commercial banks the sector strip carries real information. Pure SQL over tables already present.

3. **Bonus-adjusted fiscal-year story strip.** One row per fiscal year (2078/79 … 2082/83): EPS, profit, book value, bonus %, cash %, right %, and *total shareholder return including bonus shares*. Nobody on NEPSE shows that return; incumbents list cash % only (MeroLagani) or last dividend only (ShareSansar, ShareHub). Add a derived "shares outstanding history" line so dilution from bonus issues is visible — the NEPSE equivalent of Zerodha's "Forensics".

4. **Open checklist (Screener.in Pros/Cons + Simply Wall St transparency).** Six to ten boolean checks with the rule text shown on hover and committed in the repo, e.g. "EPS grew in 3 of last 4 fiscal years", "Cash payout below 30% of profit", "PE below own 3-year median", "Promoter holding unchanged over 4 quarters", "Paid bonus in each of last 3 fiscal years". Rendered as two lists, "Holds up" / "Watch", never as a score or a buy/sell word. Fits "numbers first, opinions never" because each line is a fact with a threshold the user can read.

5. **Compounded growth mini-tables** (Screener.in): Profit, EPS, Book Value, Price (bonus-adjusted) over 5Y / 3Y / 1Y, fiscal-year aligned. Trivial from `fundamentals` + prices; instantly recognisable as "not a broker site".

6. **Mutual-fund ownership block as narrative.** "Held by 7 of 23 funds (Rs 41 Cr, 1.8% of float). Biggest: NIBLSF 0.6%. Net change last report: +52,000 units." Fund-holdings data is something only NTX has assembled; Tickertape's "Top mutual funds" and Trendlyne's shareholding section show the shape. On the fund page, invert it: "top 10 holdings as % of net assets" and sector exposure.

7. **Record and streak annotations on the quarterly chart** (Kabutan). Small labels on the profit/EPS bars: "highest Q2 since 2077/78", "4th consecutive YoY increase", "first loss in 12 quarters". Computed from `profit_amount` by fiscal_year/quarter.

8. **Home page as a sector-first story, not a ticker tape.** Lead with the 12 NEPSE sectors as rows: median PE, median PB, share of companies profitable this quarter, aggregate profit YoY, number of companies — with hydropower's 112 companies explicitly as a distribution rather than an average. Then "Today's results" (companies that filed a quarterly report, with the templated headline from idea 7) instead of top gainers/losers. Gainers/losers stay, but small and last.

A note on the Graham number: both Investidor10 and NEPSEAnalysis show it. It is feasible (EPS × book value) and transparent, but it is a "fair value" and will read as a tip. If used, show it in the vs-own-history strip as a reference line labelled "Graham formula", not as a headline.

---

## 5. What not to copy

- **Quote hero + buy widget + sign-up box** (Public, Robinhood, Zerodha, ShareSansar SS Pro). This is the broker template; the price is the least informative number on the page.
- **The tab wall** (ShareSansar 14 tabs, ShareHub 13, MeroLagani 10, Morningstar 13). Tabs hide the story; one scroll should carry it. Sub-pages only for depth (full statements, floorsheet).
- **Pivot levels, MA bullish/bearish flags, RSI verdicts, "Green/Red signals"** (ShareSansar, NEPSEAnalysis, Chukul, Moneycontrol "VERY BULLISH"). Technical verdicts are tips with extra steps and are what every NEPSE site already shows.
- **Crowd targets and buy/sell ratios** (Minkabu, Tickertape analyst gauge). Conflicts with "no tips".
- **100-cell stat grids** (Finviz, MeroLagani's metric list). Density should come from time series and comparison, not from label:value pairs.
- **Score numbers without rules** (Moneycontrol 81/100, NEPSEAnalysis MASTER, Sarmaaya /20). If a number is shown, its rule must be a sentence on the page, as Simply Wall St's checks are.
- **Ad banners above the company name and premium locks on fundamentals** (ShareSansar, Moneycontrol, NepaliPaisa). NTX's positioning is the absence of these.
- **Repeated navigation, FAQs that restate the stats, and auto-FAQ SEO blocks** (TradingView, Public, ShareHub, NepaliPaisa). They lengthen without informing.

---

## Sources checked but not usable

Simply Wall St company page (403), Stockopedia share page and academy (403), Trendlyne equity page (405), Macrotrends live page (403), Status Invest (403), Minkabu stock page (405), Naver Finance (blocked), Xueqiu (WAF challenge), moomoo (JS shell), Koyfin app (JS shell), NepseAlpha (404 on known paths), Chukul and nepalstock.com (JS shells), Fiscal.ai (429), CSE Sri Lanka (404), DSE Bangladesh (410), stockbangladesh.com (DNS). Claims about these are limited to vendor documentation and are flagged above.
