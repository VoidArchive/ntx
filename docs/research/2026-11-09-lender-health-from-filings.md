# Can lender health be read from the quarterly filing PDFs?

Spike for the Financial Health axis (issue #53, part 3). Question: can one parser
read NPL to total loans, capital fund to risk-weighted assets and total
distributable profit from the "major financial highlights" table of the
quarterly filings NEPSE publishes for banks, development banks, finance
companies and microfinance? Success was set at 18 of 20 filings.

## Method

Twenty filings, the latest quarterly PDF for each of twenty lenders across
the four sectors (five commercial banks, four development banks, four finance
companies, seven microfinance), downloaded from the `source_url` the story
engine already stores. `pdftotext -layout`, then one regular expression per
figure over the text, taking the numbers on the matching line. Throwaway
script; nothing stored.

## Result: 3 of 20 clean, so the bar is not met

| Outcome | Filings | Which |
|---|---|---|
| All three figures read correctly | 3 | DDBL, JBBL, NBL (NBL's NPL label was missed, its other two read) |
| Two of three (NPL and capital fund) | 5 | EBL, FOWAD, GUFL, MERO, SWBBL |
| One of three | 4 | ICFC, LLBS, MFIL, MLBL |
| Nothing usable | 8 | GBBL, HBL, NUBL, SCB (text layer, different layout); NABIL, SHINE, CFCL, CYCL (scanned images, no text layer) |

Where the table is text and in the NRB layout, the labels are stable enough to
match: "Non-Performing Loan (NPL) to Total Loan", "Capital Fund to RWA", "Total
Distributable profit or (loss) as on <date>". The two values read for Nepal Bank
(capital fund 13.64%, distributable profit Rs 2,752,845,062 at Ashoj end 2083)
match the figures in its filing as reprinted by nepsetrading, and Deprosc's NPL
of 6.70% against 8.73% a year earlier matches the prior-year figure in
ShareSansar's write-up.

## Why it fails

- **One filing in five is a scanned image.** NABIL, SHINE, CFCL and CYCL have
  no text layer at all. Reading them needs OCR, which the mutual-fund scripts
  already carry for NAV reports, but at a different accuracy cost for a 1-page
  table of percentages.
- **Label drift.** "Gross Non Performing Loan (NPL)" (GUFL), "non-performing
  assets (NPA)" in prose only (MLBL), "Retained Earning" where others say
  "Distributable profit" (SWBBL, LLBS). A per-label synonym list helps but each
  new filer can add one.
- **Layout.** Q4 filings (HBL, MLBL) are annual statements with the highlights
  table elsewhere or absent; two-column pages interleave the Nepali and English
  tables so a regular expression over a line picks up the neighbour's number
  (FOWAD's distributable profit matched "Derivative Financial Instrument").
- **Nepali fonts without Unicode mapping** come out as garbage (MFIL), so even
  the English half of a mixed line can be lost.
- **The date lives in the label** ("as on Ashoj end 2083"), so a naive "numbers
  on the line" reader captures the year before the value; fixable, but a sign
  that every figure needs its own rule.

## Recommendation

Do not build the lender Health checks on pdftotext alone. A reader that clears
the bar needs OCR for scanned filings, a synonym list per figure, column-aware
table parsing and a per-filing confidence that refuses rather than guesses. That
is its own project, and the wrong-NPL incident behind ADR 0002 says a guessed
figure is worse than none. Two cheaper paths to test before building it:

1. **NRB's own publications.** The monthly Banking and Financial Statistics
   give NPL and capital ratios by class, not by bank; check whether NRB's
   financial-soundness or bank-wise tables exist in a machine-readable form.
2. **A narrower target.** NPL and capital fund alone read in 8 of 16 text
   filings with no tuning; distributable profit is the figure that fails most.
   Two figures from the text filings, with "not recorded" for the rest, might
   be enough for one Health check for lenders, stated with its coverage.

Until one of those lands, the Health axis for lenders stays as it is (book value
growth), and the company page says nothing it cannot source.
