# NTX sub-project 2: hydro map

Date: 2026-11-05. Status: approved after a grilling session; published as a spec issue. Tickets will be sub-issues of it. Parent record: `docs/superpowers/specs/2026-11-04-ntx-content-redesign-design.md`. Vocabulary: `GLOSSARY.md` (Hydro section). Decisions: `docs/adr/0002`, `docs/adr/0004`.

## Why

Hydropower is 112 of NEPSE's listed companies and the sector where the owner's one-hydro-per-basin rule came from. No NEPSE site shows where a company's plants are, what river they sit on, or what the weather is doing to that river. The exchange's data cannot answer "which listed hydros share the Trishuli" or "is it raining on my basin"; a curated plant table and daily readings from a public weather model can.

## Scope

v1 is public and holdings-free. It answers two questions: what sits on which river, and what the rain and discharge at each plant were today. The holdings overlay (held basins, the basin check on the company page, the one-hydro-per-basin rule) is sub-project 3 and builds on the same map and basin state.

## 1. Data

### Plant table

`data/hydro_plants.csv`, the source of truth, hand-edited. Columns: `symbol, plant, river, basin, district, mw, lat, lon, status, source_url`.

- One row per plant; a company may have several. Only plants of listed companies: a DoED row whose promoter matches no NEPSE symbol is dropped.
- `status` is `operating` or `under_construction`. Planned and licence-only plants are not rows. A plant changes status by a hand edit when a filing says so.
- `basin` is one of the nine in the glossary: Koshi, Gandaki, Karnali, Mahakali, Bagmati, Kamala, West Rapti, Babai, Kankai. Any other value fails the CSV load.
- `lat`, `lon` are decimal degrees, one point per plant. The first draft takes them from the ICIMOD point layer derived from the DoED table; `source_url` points at the DoED licence table. The review step nudges a point onto its river where the GloFAS 5 km cell would land on the wrong one (the Open-Meteo docs warn of this).
- A row without `source_url` is not committed (ADR 0002 applies to coordinates and MW).
- The first draft is produced with research help from the DoED operating and construction-licence lists; every row is reviewed by hand before commit.

The worker upserts the CSV into `hydro_plants(id, symbol, plant, river, basin, district, mw, lat, lon, status, source_url)` on every start, keyed on `(symbol, plant)`. No admin UI; a CSV edit plus a deploy is the workflow.

### Readings

`hydro_readings(plant_id, day, rain_mm, discharge_m3s, rain_source, discharge_source)` with `UNIQUE(plant_id, day)`.

- Rain: Open-Meteo forecast endpoint, `daily=precipitation_sum`, model requested explicitly as ECMWF IFS so the source can be named. Source text "Open-Meteo, ECMWF IFS".
- Discharge: Open-Meteo flood endpoint, `daily=river_discharge` (GloFAS v4, 5 km, modelled not gauged). Source text "Open-Meteo, GloFAS v4".
- Only today and past days are stored. Forecast days are discarded at ingest; the page never shows a future date.
- First run per plant backfills: 92 days of rain (the endpoint's maximum) and ten years of discharge. Rows older than ten years are never stored.

### Plant state and basin state

Computed in the API at read time from the plant's readings, never stored.

- **Heavy rain**: latest reading's `rain_mm >= 50`, the DHM 24-hour "heavy" category (source: DHM forecast terminology, cited on the page).
- **High discharge**: latest reading's discharge exceeds the 95th percentile of that plant's stored daily discharge over the ten-year window.
- **Normal** otherwise. Both rules can hold; High discharge is the worse and is the one shown.
- **Stale**: a label added when the latest reading is more than two days old. The state is still computed from that reading and its date is shown beside it.
- Basin state is the worst state among the basin's plants.

Rule text lives in the repo beside the rules and is printed on the page, as the checks are.

## 2. Map build

`scripts/build_hydro_map.py`, run once and when the cutoff changes, never at runtime or build:

1. Download geoBoundaries NPL ADM0 GeoJSON and the HydroRIVERS Asia shapefile.
2. Clip HydroRIVERS to the ADM0 polygon. Keep reaches with Strahler order (`ORD_STRA`) 4 or higher, plus any lower-order reach within a small radius of a plant in the CSV, so every plant's river is drawn. The cutoff is tuned once by eye.
3. Project outline and rivers with one fixed projection (scale, centre, translate) and write `apps/web/src/lib/data/hydro_map.json`: `{ projection: {...}, outline: "M…", rivers: [{ basin, order, d: "M…" }] }`. Rivers carry their basin, assigned from which of the nine basins the reach's main river drains to.
4. The web page projects plant dots with the same stored projection parameters, one formula, no geo library. d3-geo is a script dependency only.

No basin polygons: basin colouring is carried by the rivers and dots of each basin. No district boundaries.

## 3. Worker

One new job, `weather`, at 06:00 NPT every day of the week, independent of the market-status holiday skip. For every plant with coordinates, regardless of status: one rain call and one discharge call, upserting readings for today and past days. About 400 calls a day against a free limit of 10,000 a day and 600 a minute. A failed plant is logged and skipped; there is no retry beyond the next morning. The CSV upsert runs at worker start before the schedule.

Attribution obligations the worker's sources impose: Open-Meteo CC BY 4.0, geoBoundaries (visible link), HydroSHEDS. They are printed in the `/hydro` data footer.

## 4. RPC

Proto `ntx/v1/hydro.proto`, service `HydroService`:

- `ListPlants() → { plants[] }`. Per plant: id, symbol, company name, plant, river, basin, district, mw, lat, lon, status, state (NORMAL | HEAVY_RAIN | HIGH_DISCHARGE), stale, latest reading (day, rain_mm, discharge_m3s, sources), company close, PE, rating. Plus `basins[]` with name and state, and `rules[]` with the two rule texts.
- `GetPlantReadings(plant_id) → { readings[] }`: the last seven days, newest first, each with sources.
- `GetCompanyPlants(symbol) → { plants[] }`: the company's plants with state, for the company page.

PE and rating are the stored values from sub-project 1b, joined, not recomputed.

## 5. Pages

### `/hydro`

Sidebar gains **Hydro** between Companies and Mutual funds.

1. **Opening line.** "N plants of M listed companies on K rivers. Basins today: Koshi normal, Gandaki heavy rain, …" with the reading date. Prose first, as every page.
2. **Map.** Outline, rivers coloured by basin state, plant dots sized by MW (square-root scale), filled for operating, hollow for under construction, coloured by plant state. State colour is never the only carrier: each state has a glyph and a label in the legend and in rows. Hover on a dot shows the plant facts below; clicking a river or dot sets the basin filter. The map is read-only on phones.
3. **Plant list.** One row per plant: plant, symbol (link to company page), river, basin, MW, status, today's rain and discharge with the state glyph, reading date, close, PE, rating. Sort (basin, MW, state, PE) and filters (basin, status, state) in the URL as the screener does. This is the phone and keyboard surface; hover is a shortcut to a row.
4. **Readings.** Expanding a row shows the last seven days as a table: date, rain, discharge, state, source per figure. No chart in v1.
5. **Rules and data footer.** The two rule texts with their thresholds and the DHM citation; a sentence that discharge is modelled by GloFAS, not gauged; last weather run; attributions for Open-Meteo, geoBoundaries and HydroSHEDS.

### Company page, hydro companies only

A **Plants** section after Funds and before Price and volume: one row per plant with river, basin, MW, status, current state and reading date, linking to `/hydro?basin=…`. Opening sentence: "Two plants, 36 MW, on the Trishuli (Gandaki)." No mini-map in v1.

### Phone (390)

Opening line, basin states, then the list; the map renders below the list at reduced height without hover.

## 6. Out of scope

Holdings and held-basin highlighting, the basin check and the one-hydro-per-basin rule (sub-project 3). Forecasts, alerts, notifications. Basin polygons, district boundaries, a basin mini-map on the company page, charts of readings. Plants under 1 MW, planned plants, unlisted promoters. Scraping DoED at runtime.

## 7. Tests

- Go: CSV load rejects an unknown basin, an unknown status and a missing source URL; state rules on fixture readings for one plant (normal, heavy rain, high discharge, both, stale); basin state as worst-of; backfill discards forecast days and rows older than ten years; one integration test per RPC against a fixture database with three plants in two basins.
- Script: a check that every CSV plant has at least one drawn river reach within the radius.
- Web: `svelte-check`; screenshot pass of `/hydro` at 1440 and 390 with the fixture data.

## Open before data is committed

The HydroSHEDS licence page and the ICIMOD layer's reuse terms were summarised from secondary text, not read; confirm both before the first CSV and map file land.
