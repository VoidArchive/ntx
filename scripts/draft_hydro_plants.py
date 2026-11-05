"""Draft data/hydro_plants.csv from the DoED licence tables.

Reads the two public DoED tables (operating plants above 1 MW, and hydro
generation licences above 1 MW), keeps the rows whose promoter is a NEPSE-listed
hydro company according to PROMOTERS below, and writes one CSV row per plant.
A plant in the operating table is `operating`; a plant only in the licence table
is `under_construction`. Coordinates are the centre of the licence area's bounds.

The output is a draft: every row is reviewed by hand before it is committed
(docs/superpowers/specs/2026-11-05-ntx-hydro-map-design.md, section 1).

Usage: uv run draft_hydro_plants.py [--cache DIR] OUT.csv
"""

import argparse
import csv
import html
import pathlib
import re
import sys

import requests

OPERATING_URL = "https://www.doed.gov.np/pages/powerplantsmorethan1"
LICENCE_URL = "https://www.doed.gov.np/pages/clhydromorethan1"

# NEPSE symbol -> the promoter strings DoED uses for it (DoED spells one company
# several ways). Edited by hand after reading both tables on 2026-11-05.
PROMOTERS: dict[str, list[str]] = {
    "AHL": ["Asian Hydropower Pvt. Ltd"],
    "AHPC": [
        "Arun Valley Hydropower Development Company Ltd.",
        "Arun Valley Hydropower Development Company Pvt. Ltd.",
        "Arun Valley Hydropower Development Company Limited",
    ],
    "AKJCL": ["Ankhu Khola Jalvidut Co. Ltd"],
    "AKPL": ["Arun Kabeli Power Limited."],
    "APHL": ["Apolo Hydropower Pvt.Ltd"],
    "API": ["Api Power Co. Ltd", "Api Power Company Ltd.", "Api Power Company Pvt. Ltd"],
    "BARUN": ["Barun Hydropower Development Co. Pvt. Ltd"],
    "BEDC": ["Bhugol Energy Development Company Pvt Ltd"],
    "BENH": ["Beni Hydropower Project"],
    "BGWT": ["Bhagawati Hydropower Development Company"],
    "BHCL": ["Bikash Hydropower Company Pvt. Ltd."],
    "BHDC": [
        "Bindhabasini Hydropower Developmwnt Company Pvt.Ltd",
        "Bindhyabasini Hydropower Development Co. Pvt Ltd",
    ],
    "BHL": ["Balephi Hydropower Ltd."],
    "BHPL": ["Barahi Hydropower Pvt Ltd"],
    "BJHL": ["Bhujung Hydro Power Ltd.,"],
    "BNHC": ["Buddha Bhumi Nepal Hydropower Co. Pvt. ltd."],
    "BPCL": ["Butwal Power Company", "Butawal Power Company"],
    "BUNGAL": ["Bungal Hydro Ltd"],
    "CHCL": ["Chilime Hydropower Company Limited"],
    "CHL": ["Chhyandi Hydropower Co. P. Ltd", "Chhyandi Hysropower LTD"],
    "CKHL": ["Chirkhuwa Hydropower Company Pvt td"],
    "DHEL": ["Daram Khola Hydro Energy Limited"],
    "DHPL": ["Dibyaswari Hydropower P Ltd"],
    "DOLTI": ["Dolti Power Company P. Ltd"],
    "DORDI": ["Dordi Khola Jalvidyut Company Limited"],
    "EHPL": ["Eastern Hydropower P Ltd"],
    "GHL": ["Ghalemdi Hydro Limited"],
    "GLH": ["Green Life Energy Pvt.Ltd,Sirin Bhawan, Lalitpur-1, Kupakdole"],
    "GVL": ["Green Ventures Ltd"],
    "HDHPC": ["Himal Dolkha Hydropower Co Ltd", "Himal Dolkha Hydropower Company Pvt Ltd"],
    "HHL": ["Himalayan Hydropower Pvt.Ltd"],
    "HIMSTAR": ["Him Star Urja Company Ltd."],
    "HPPL": ["Himalaya Power Partner Pvt. Ltd"],
    "HURJA": ["Himalaya Urja Bikas Co. Pvt. Ltd.", "Himalaya Urja Bikash Company Limited"],
    "IHL": ["Ingwa Hydopower Ltd.,"],
    "JOSHI": ["Joshi Hydropower Co. P. Ltd"],
    "KAHL": ["Kalanga Hydropower P Ltd"],
    "KHPL": ["Kalinchowk Hydropower Pvt. Ltd."],
    "KKHC": ["Khani Khola Hydropower Company Ltd"],
    "KPCL": ["Daraundi Kalika Hydro"],
    "LEC": ["Liberty Energy Hydropower Pvt. Ltd."],
    "MABEL": ["Mabilung Energy.Ltd"],
    "MAKAR": ["Makar Jitumaya Hudropower Pvt.Ltd."],
    "MANDU": ["Mandu Hydropower Ltd."],
    "MBJC": ["Madhya Bhotekoshi Jalbidyut Company Ltd."],
    "MCHL": ["Menchhiyam Hydropower P Ltd. Itahari,Sunsari,9841521159"],
    "MEHL": ["Mankamana Engineering HP"],
    "MEL": ["Modi Energy Pvt . Ltd"],
    "MEN": ["Mountain Energy Nepal Limited", "Mountain Energy Nepal Ltd"],
    "MEPDL": ["Mount Everest Power Development P.Ltd."],
    "MHCL": ["Molun Hydropower Co. Pvt. Ltd"],
    "MHL": ["Mandakini Hydropower Ltd.", "Mandakini Hydropower Ltd"],
    "MHNL": ["Mountain Hydro Nepal (P.) Ltd"],
    "MKHC": ["Maya Khola Hydropower Company Ltd"],
    "MKHL": ["Mai Khola Hydropower Pvt.Ltd."],
    "MKJC": ["Mailun Khola Hydropower Company Pvt. Ltd"],
    "MMKJL": ["Mathillo Mailung Khola Jalbidhyut Ltd", "Mathillo Mailung Khola Jalvidyut Ltd."],
    "MSHL": ["Mid Solu Hydropower Ltd."],
    "NGPL": ["Nyadi Group Pvt Ltd"],
    "NHDL": ["Nepal Hydro Developer Pvt Ltd"],
    "NHPC": ["National Hydropower Company Pvt. Ltd.", "National Hydropower Company Limited"],
    "NYADI": ["Nyadi Hydropower Limited"],
    "PHCL": ["Peoples Hydropower Company Pvt.Ltd.", "Peoples Hydropower Company Pvt. Ltd"],
    "PMHPL": ["Mai Valley Hydropower P Ltd.", "Mai Valley Hydropower P.L.,"],
    "PPCL": ["Panchthar Power Company Pvt. Ltd."],
    "PPL": ["Peoples Power Ltd"],
    "RADHI": ["Radhi Bidyut Co. Ltd"],
    "RAWA": ["Rawa Energy Development Pvt Ltd"],
    "RFPL": ["River Falls Hydropower Development P. Ltd."],
    "RHGCL": ["Rapti Hydro and General Construction Pvt. Ltd."],
    "RHPL": ["Rasuwagadhi Hydropower Company Ltd."],
    "RIDI": ["Ridi Power Co. Ltd.", "Ridi Power Company Ltd."],
    "RLEL": ["Ridge Line Energy Pvt. Ltd"],
    "RURU": ["Ruru Jalbidyut Pariyojana Pvt. Ltd"],
    "SAHAS": ["Sahas Urja Limited"],
    "SANVI": ["Sanvi Energy Pvt. Ltd."],
    "SGHC": [
        "Swaita Ganga Hydropower and Construction Pvt. Ltd.",
        "Swet Ganga Hydropower & Construction Limited",
    ],
    "SGHL": ["Sani Gad Hydro pvt. Ltd"],
    "SHEL": ["Singati Hydro Energy Pvt. Ltd."],
    "SHPC": ["Sanima Mai Hydropower Limited"],
    "SIKLES": ["Silkes Hydropower Pvt.Ltd"],
    "SIPD": ["Shikhar Power Development Pvt.Ltd.", "Shikhar Power Development Pvt. Ltd."],
    "SJCL": ["Sanjen Jalvidhyut Co", "Sanjen Jalvidyut Co."],
    "SKHEL": ["Surya Kund Hydro Electric Pvt. Ltd."],
    "SKHL": ["Super Khudi Hydropower Ltd."],
    "SMH": ["Supermai Hydropower Pvt.Ltd."],
    "SMHL": ["Himal Hydro and General Construction Ltd."],
    "SMJC": ["Sagarmatha Jalbidhyut Company P.Ltd."],
    "SNORL": ["Snow River Ltd"],
    "SOHL": ["Solu Hydropower Ltd"],
    "SPC": ["Samling Power company Pvt.Ltd"],
    "SPDL": ["Synergy Power Development P Ltd"],
    "SPHL": ["Sayapatri Hydropower Pvt. Ltd."],
    "SSHL": ["Shiva Sri Hydropower Pvt. Ltd"],
    "TAMOR": ["Sanima Middle Tamor Hydropower Ltd"],
    "TPC": ["Terhathum Power Company Pvt. Ltd."],
    "TPKHL": ["Taksar Pikhuwa Khola Hydropower Ltd."],
    "TSHL": ["Three Star Hydro Power Pvt.Ltd."],
    "UHEWA": ["Upper Hewa Khola Hydropower Company Pvt Ltd"],
    "ULHC": ["Upper Lohore Khola Hydropower Company. Pvt. Ltd"],
    "UMHL": [
        "United Modi Hydropower Pvt. Ltd., 1st Floor Heritage Plaza 2; Kamaladi, Kathmandu Metropolitan - 31"
    ],
    "UMRH": ["United Idimardi and R.B. Hydropower Pvt Ltd"],
    "UNHPL": ["Union Hydropower P.Ltd", "Union Hydropower Limited", "Union Hydropower Ltd"],
    "UPCL": ["Universal Power Company P. Ltd"],
    "UPPER": ["Upper Tamakoshi Hydropower Limited", "Upper Tamakoshi Hydropower Company Limited"],
    "USHEC": ["Upper Solu Hydroelectric Company Pvt Ltd"],
    "USHL": ["Upper Syange Hydropower Pvt. Ltd."],
    "VLUCL": ["Vision Lumbini Urja Company Limited."],
    "YMHL": ["Yambaling Hydropower Ltd."],
}

# Projects whose DoED promoter is not the listed company: the licence holder is the
# parent (NEA for Upper Trishuli 3B, built by its listed subsidiary TVCL).
PROJECTS: dict[str, str] = {"Upper Trishuli 3B": "TVCL"}

# Hand corrections after review, keyed by (symbol, DoED project name).
# KBSH (0.998 MW) and SPL (0.99 MW) are below the 1 MW floor and have no row.
FIXES: dict[tuple[str, str], dict[str, str]] = {
    ("MAKAR", "Upper Suri Khola HPP"): {"river": "Suri Khola"},  # DoED lists two village names as the river
    ("BEDC", "Dwari Khola SHP"): {"district": "Dailekh"},  # DoED leaves the district blank; the point is in Dailekh
    ("MHL", "Upper Sardi HEP"): {"river": "Sardi Khola"},  # DoED: "Sardi, kutmi"
    ("PMHPL", "Upper Mai Hydropower Project (Panchakanya Mai Hydropwer limi"): {"plant": "Upper Mai"},  # truncated by DoED
}

# Every listed hydro symbol, so the script can report the ones DoED's tables do not name.
LISTED = """AHL AHPC AKJCL AKPL APHL API BARUN BEDC BENH BGWT BHCL BHDC BHL BHPL BJHL BNHC BPCL
BUNGAL CHCL CHL CKHL DHEL DHPL DOLTI DORDI EHPL GHL GLH GVL HDHPC HHL HIMSTAR HPPL HURJA IHL
JOSHI KAHL KBSH KHPL KKHC KPCL LEC MABEL MAKAR MANDU MBJC MCHL MEHL MEL MEN MEPDL MHCL MHL MHNL
MKHC MKHL MKJC MMKJL MSHL NGPL NHDL NHPC NYADI PHCL PMHPL PPCL PPL RADHI RAWA RFPL RHGCL RHPL
RIDI RLEL RURU SAHAS SANVI SGHC SGHL SHEL SHPC SIKLES SIPD SJCL SKHEL SKHL SMH SMHL SMJC SNORL
SOHL SPC SPDL SPHL SPL SSHL TAMOR TPC TPKHL TSHL TVCL UHEWA ULHC UMHL UMRH UNHPL UPCL UPPER
USHEC USHL VLUCL YMHL""".split()

# Basin per district, the nine of GLOSSARY.md. Districts split between basins list the
# larger share; RIVER_BASIN below overrides by river name.
DISTRICT_BASIN: dict[str, str] = {
    **dict.fromkeys(
        ["Taplejung", "Panchthar", "Terhathum", "Sankhuwasabha", "Bhojpur", "Khotang", "Okhaldhunga",
         "Solukhumbu", "Ramechhap", "Dolakha", "Sindhupalchok"], "Koshi"),
    **dict.fromkeys(
        ["Kaski", "Lamjung", "Gorkha", "Manang", "Myagdi", "Baglung", "Parbat", "Syangja", "Gulmi",
         "Dhading", "Rasuwa", "Nuwakot"], "Gandaki"),
    **dict.fromkeys(["Bajhang", "Dailekh", "Surkhet", "Rukum"], "Karnali"),
    "Darchula": "Mahakali",
    **dict.fromkeys(["Lalitpur", "Makawanpur"], "Bagmati"),
    **dict.fromkeys(["Pyuthan", "Palpa"], "West Rapti"),
    "Ilam": "Kankai",
}
RIVER_BASIN: dict[str, str] = {"Ridi": "Gandaki"}  # joins the Kali Gandaki at Ridi, Palpa


def fetch(url: str, cache: pathlib.Path | None) -> str:
    if cache:
        path = cache / (url.rstrip("/").rsplit("/", 1)[-1] + ".html")
        if path.exists():
            return path.read_text(encoding="utf-8")
    text = requests.get(url, timeout=60).text
    if cache:
        cache.mkdir(parents=True, exist_ok=True)
        path.write_text(text, encoding="utf-8")
    return text


def parse_table(page: str) -> list[list[str]]:
    rows = []
    for tr in re.findall(r"<tr[^>]*>(.*?)</tr>", page, re.S):
        cells = [html.unescape(re.sub(r"<[^>]+>", " ", c)) for c in re.findall(r"<t[dh][^>]*>(.*?)</t[dh]>", tr, re.S)]
        cells = [re.sub(r"\s+", " ", c).strip() for c in cells]
        if len(cells) >= 12 and cells[0].isdigit():
            rows.append(cells)
    return rows


def dms(text: str) -> float | None:
    """'27o 28' 28"' -> 27.4744. DoED writes the degree sign as the letter o."""
    m = re.match(r"\s*(\d+)\D+(\d+)\D+(\d+(?:\.\d+)?)", text)
    if not m:
        return None
    d, mi, s = (float(x) for x in m.groups())
    return round(d + mi / 60 + s / 3600, 4)


def centre(a: str, b: str) -> float | None:
    x, y = dms(a), dms(b)
    if x is None or y is None:
        return None
    return round((x + y) / 2, 4)


def basin_for(river: str, district: str) -> str:
    for name, basin in RIVER_BASIN.items():
        if name.lower() in river.lower():
            return basin
    for name, basin in DISTRICT_BASIN.items():
        if name.lower() in district.lower():
            return basin
    return ""


def district_of(cell: str) -> str:
    m = re.search(r"\(([^)]*)\)\s*$", cell)
    return m.group(1).strip() if m else cell.strip()


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("out")
    ap.add_argument("--cache", type=pathlib.Path)
    args = ap.parse_args()

    by_promoter = {p: sym for sym, ps in PROMOTERS.items() for p in ps}
    plants: dict[tuple[str, str], dict] = {}
    for status, url in (("operating", OPERATING_URL), ("under_construction", LICENCE_URL)):
        for row in parse_table(fetch(url, args.cache)):
            sym = PROJECTS.get(row[1]) or by_promoter.get(row[7])
            if not sym:
                continue
            key = (sym, re.sub(r"[^a-z0-9]", "", row[1].lower()))
            if key in plants:
                continue  # the operating table is read first and wins
            fix = FIXES.get((sym, row[1]), {})
            district = fix.get("district", district_of(row[13]))
            river = fix.get("river", row[3])
            plants[key] = {
                "symbol": sym,
                "plant": fix.get("plant", row[1]),
                "river": river,
                "basin": basin_for(river, district),
                "district": district,
                "mw": row[2],
                "lat": centre(row[9], row[10]),
                "lon": centre(row[11], row[12]),
                "status": status,
                "source_url": url,
            }

    rows = sorted(plants.values(), key=lambda p: (p["symbol"], p["plant"]))
    with open(args.out, "w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=["symbol", "plant", "river", "basin", "district", "mw", "lat", "lon", "status", "source_url"])
        w.writeheader()
        w.writerows(rows)

    matched = {p["symbol"] for p in rows}
    print(f"{len(rows)} plants for {len(matched)} companies", file=sys.stderr)
    print("no DoED row:", " ".join(s for s in LISTED if s not in matched), file=sys.stderr)
    print("no basin:", sum(1 for p in rows if not p["basin"]), file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
