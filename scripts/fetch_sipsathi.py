"""
Build apps/web/src/lib/data/nav.json (the web app's Fund[] shape, see
apps/web/src/lib/types/fund.ts) from sip-sathi.vercel.app.

The site embeds its data as a gzipped JS bundle (d/N.bin chunks) whose
`const DATA=__decodeNav({...})` call holds every open-end fund's NAV,
AUM, asset allocation and (for some funds) itemised stock holdings.

Usage:
    python3 scripts/fetch_sipsathi.py "2083 Ashwin"

Holding names come from the ntx SQLite database (NTX_DB env var or the
default path); symbols not found there (mutual fund units, promoter
shares) are kept as symbols.
"""

import datetime
import gzip
import json
import os
import re
import sqlite3
import sys
import urllib.request
from pathlib import Path

SITE = "https://sip-sathi.vercel.app/"
OUT_DIR = Path(__file__).parent.parent / "apps/web/src/lib/data"
DEFAULT_DB = Path.home() / ".local/share/ntx/market.db"

SECTOR_KEYS = {
    "BANKING": "commercial_banks",
    "DEVBANK": "development_banks",
    "FINANCE": "finance_companies",
    "LIFEINSU": "life_insurance",
    "NONLIFEINSU": "non_life_insurance",
    "HYDROPOWER": "hydropower",
    "HOTELS": "hotels",
    "MANUFACTURE": "manufacturing",
    "MICROFINANCE": "microfinance",
    "MUTUAL": "mutual_funds",
    "INVESTMENT": "investment",
    "TRADING": "trading",
}


def fetch(path: str) -> bytes:
    with urllib.request.urlopen(SITE + path, timeout=60) as r:
        return r.read()


def load_chunks(source: str, prefix: str) -> str:
    """Fetch and gunzip the `<prefix>/N.bin` chunk list referenced in `source`."""
    chunks = re.findall(rf'"({prefix}/\d+\.bin)"', source)
    return gzip.decompress(b"".join(fetch(c) for c in chunks)).decode()


def extract_data(js: str) -> dict:
    start = js.index("=__decodeNav(") + len("=__decodeNav(")
    data, _ = json.JSONDecoder().raw_decode(js[start:])
    return data


def load_names(db_path: Path) -> dict[str, str]:
    if not db_path.exists():
        return {}
    rows = sqlite3.connect(db_path).execute("select symbol, name from companies")
    return {symbol: name.strip() for symbol, name in rows}


def holding_name(symbol: str, names: dict[str, str]) -> str:
    if symbol in names:
        return names[symbol]
    # Promoter shares trade under the parent's symbol plus "PO"
    if symbol.endswith("PO") and symbol[:-2] in names:
        return f"{names[symbol[:-2]]} (Promoter)"
    return symbol


def build_funds(data: dict, names: dict[str, str], nepali_month: str) -> list[dict]:
    as_of = datetime.date.fromisoformat(data["as_of"])
    english_month = as_of.strftime("%B %Y")
    stock_sector = data["stock_sector"]
    names = {**{f["sym"]: f["name"] for f in data["funds"]}, **names}
    funds = []
    for raw in data["funds"]:
        holdings: dict[str, list[dict]] = {}
        itemised = 0
        for symbol, value in raw["holdings"] or []:
            key = SECTOR_KEYS.get(stock_sector.get(symbol, "OTHERS"), "others")
            holdings.setdefault(key, []).append({"name": holding_name(symbol, names), "value": value})
            itemised += value
        alloc = raw["alloc"]
        # Only some funds publish every holding; the rest list a top 10 or nothing,
        # so the unlisted part of their share portfolio is kept as one row.
        if raw["holdings_full"]:
            disclosure = "full"
        else:
            disclosure = "top10" if itemised else "none"
            holdings.setdefault("others", []).append(
                {"name": "Other listed shares", "value": round(alloc["cm"] - itemised)}
            )
        if alloc["cash"]:
            holdings["fixed_deposits"] = [{"name": "Cash & Fixed Deposits", "value": alloc["cash"]}]
        if alloc["fi"]:
            holdings["corporate_debentures"] = [{"name": "Fixed Income Instruments", "value": alloc["fi"]}]
        funds.append(
            {
                "symbol": raw["sym"],
                "fund_name": raw["name"],
                "fund_manager": raw["manager"],
                "report_date_nepali": nepali_month,
                "report_date_english": english_month,
                "nav_per_unit": raw["nav"],
                # NAV is net assets per unit by definition, so this is exact up to rounding
                "total_units": round(raw["aum"] / raw["nav"]),
                "net_assets": raw["aum"],
                "holdings_disclosure": disclosure,
                "holdings": holdings,
            }
        )
    return funds


def main(nepali_month: str) -> None:
    # The bootstrap HTML lists the page chunks; the page lists the data chunks.
    page = load_chunks(fetch("").decode(), "i")
    data = extract_data(load_chunks(page, "d"))
    names = load_names(Path(os.environ.get("NTX_DB", DEFAULT_DB)))
    funds = build_funds(data, names, nepali_month)
    assert len(funds) >= 10, f"only {len(funds)} funds parsed"
    for f in funds:
        assert f["nav_per_unit"] > 0 and f["net_assets"] > 0, f["symbol"]
    # Fixed filename so a monthly refresh never touches the importers; the month lives in report_date_*.
    out = OUT_DIR / "nav.json"
    out.write_text(json.dumps(funds, indent=2, ensure_ascii=False) + "\n")
    unnamed = sorted({sym for f in data["funds"] for sym, _ in f["holdings"] or [] if holding_name(sym, names) == sym})
    print(f"wrote {out} ({len(funds)} funds, as of {data['as_of']}); symbols without names: {unnamed}")


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit('usage: fetch_sipsathi.py "<BS year> <month>"  e.g. "2083 Ashwin"')
    main(sys.argv[1])
