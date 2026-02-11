"""
Fix Poush 2082 NAV JSON based on image audits of all fund NAV reports.

The original nav_2026_01.json was auto-generated from Mangsir data using multipliers.
This script applies corrections based on visual reading of actual Poush NAV report images.

Key findings:
- Every fund has significantly more holdings than the Mangsir-based JSON
- New sectors missing from JSON: mutual_funds, investment, trading, finance_companies
- Many small hydropower, microfinance, and insurance companies missing
- Some company names changed (mergers)
- NADDF was severely incomplete (had only bonds + banks, missing all equity sectors)
"""

import json
import copy
from pathlib import Path

DATA_DIR = Path(__file__).parent / "data"
WEB_DIR = Path(__file__).parent.parent / "apps" / "web" / "src" / "lib" / "data"


def h(name, units, value):
    """Helper to create a holding entry."""
    return {"name": name, "units": units, "value": value}


def load_nav():
    with open(DATA_DIR / "nav_2026_01.json") as f:
        return json.load(f)


def save_nav(data):
    for path in [DATA_DIR / "nav_2026_01.json", WEB_DIR / "nav_2026_01.json"]:
        with open(path, "w") as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        print(f"Written to {path}")


# ---------------------------------------------------------------------------
# Per-fund corrections from image readings
# Each function takes the fund dict and modifies holdings in-place
# ---------------------------------------------------------------------------

def fix_niblsf(fund):
    """NIBLSF - NIBL Sahabhagita Fund (image: scripts/images/NIBLSF.jpg)
    Image shows 18 mutual fund holdings under सामूहिक लगानी कोष समूह.
    """
    holdings = fund["holdings"]

    holdings["mutual_funds"] = [
        h("NIC Asia Flexi Cap Fund", 469000, 4370540),
        h("Sanima Growth Fund", 600000, 5636000),
        h("Sunrise Focused Equity Fund", 1243600, 11926124),
        h("Prabhu Smart Fund", 200000, 2196000),
        h("RBB Mutual Fund - 2", 136000, 1052261),
        h("Siddhartha Investment Growth Scheme 1", 231000, 2332330),
        h("Citizens Super 30 Mutual Fund", 732000, 6928660),
        h("Laxmi Value Fund 2", 435000, 4493020),
        h("Himalayan 50-20", 600000, 7080000),
        h("NIC Asia Growth Fund-2", 2400000, 22040000),
        h("Kumari Sabal Yojana", 2640000, 15228000),
        h("Muktinath Mutual Fund 1", 240000, 2300000),
        h("Garima Samriddhi Yojana", 2000000, 9060000),
        h("NMB Hybrid Fund L-2", 600000, 4524000),
        h("MBL Equity Fund", 300000, 2654000),
        h("Reliable Samriddhi Yojana", 300000, 2930000),
        h("Global IME Samunnat Yojana 2", 1400000, 13950000),
        h("HLI Large Cap Fund", 1000000, 9000000),
    ]


def fix_nmbsbf(fund):
    """NMBSBF - NMB Saral Bachat Fund - E (image: scripts/images/NMBSBF.jpg)
    Image shows 42 listed holdings + debentures + public issue shares + mutual funds.
    The Mangsir template had wrong companies in several sectors.
    """
    holdings = fund["holdings"]

    # Non-life insurance: Image shows Neco, Sagarmatha Lumbini, Siddharth Premier
    holdings["non_life_insurance"] = [
        h("Neco Insurance Limited", 264500, 163461000),
        h("Sagarmatha Lumbini Insurance Company Limited", 347996, 209651456),
        h("Siddharth Premier Insurance Limited", 266658, 199193966),
    ]

    # Life insurance: Image shows Himalayan Life, Nepal Life, National Life
    holdings["life_insurance"] = [
        h("Himalayan Life Insurance Company Limited", 123156, 45232274),
        h("Nepal Life Insurance Company Limited", 363030, 279175453),
        h("National Life Insurance Company Limited", 160393, 262894063),
    ]

    # Hydropower: Image shows completely different companies from JSON
    holdings["hydropower"] = [
        h("Vikash Hydropower Company Limited", 3462, 2303064),
        h("Him Star Urja Company Limited", 9926, 9092274),
        h("Mabilung Energy Limited", 7869, 9369524),
        h("Darmakhola Hydro Energy Limited", 3329, 2265496),
        h("Bujjal Hydro Limited", 2604, 1649269),
    ]

    # Hotels: Image shows Bandipur Kewalkar only
    holdings["hotels"] = [
        h("Bandipur Kewalkar & Tourism Limited", 6634, 5514520),
    ]

    # Manufacturing: Image shows Shivam, Unilever, Shrinagar Agritech, SY Panel
    holdings["manufacturing"] = [
        h("Shivam Cements Limited", 396008, 284774920),
        h("Unilever Nepal Limited", 9254, 49063400),
        h("Shrinagar Agritech Industries Limited", 3862, 4270200),
        h("SY Panel Nepal Limited", 6345, 6609675),
    ]

    # Others: Image shows Jhapa Energy
    holdings["others"] = [
        h("Jhapa Energy Limited", 794, 9096930),
    ]

    # Microfinance: Image shows 5 companies in public issue section
    holdings["microfinance"] = [
        h("Chhimek Laghubitta Bittiya Sanstha Limited", 276612, 273554050),
        h("Nirdhan Utthan Laghubitta Bittiya Sanstha Limited", 6960, 8492947),
        h("National Laghubitta Bittiya Sanstha Limited", 152030, 129970436),
        h("Sana Kisan Bikas Laghubitta Bittiya Sanstha Limited", 929993, 7690000),
        h("Swastik Laghubitta Bittiya Sanstha Limited", 5940, 597696),
    ]

    # Development banks: Image shows Garima, Muktinath, Kamana Sewa
    holdings["development_banks"] = [
        h("Garima Bikas Bank Limited", 645434, 8095230),
        h("Muktinath Bikas Bank Limited", 609649, 5417412),
        h("Kamana Sewa Bikas Bank Limited", 33300, 8231355),
    ]

    # NEW: Mutual funds (12 holdings from image)
    holdings["mutual_funds"] = [
        h("NIBL Growth Fund", 300000, 2903000),
        h("Citizens Super 30 Mutual Fund", 299900, 2054652),
        h("NIC Asia Growth Fund-2", 500000, 4550000),
        h("Kumari Sabal Yojana", 500000, 4080000),
        h("NIBL Stable Fund", 450000, 3906000),
        h("Muktinath Mutual Fund 1", 300000, 2760000),
        h("Garima Samriddhi Yojana", 40000, 453000),
        h("MBL Equity Fund", 500000, 4625000),
        h("Reliable Samriddhi Yojana", 700000, 6650000),
        h("Global IME Samunnat Yojana-2", 250000, 2275000),
        h("Himalayan Large Cap Fund", 2000000, 16000000),
        h("RBB Focus 40", 600000, 4926000),
    ]


def fix_ssis(fund):
    """SSIS - Siddhartha Systematic Investment Scheme (image: scripts/images/SSIS.jpg)
    Very comprehensive portfolio with many sectors.
    """
    holdings = fund["holdings"]

    # Hydropower: Image shows many small hydros
    holdings["hydropower"] = [
        h("Api Power Company Limited", 46000, 13566400),
        h("Arun Kabeli Power Limited", 20000, 4920000),
        h("Butwal Power Company Limited", 22680, 9294619),
        h("Sanima Middle Tamor Hydropower Limited", 39645, 14923440),
        h("Sanima Mai Hydropower Limited", 26460, 9654460),
        h("Mountain Energy Nepal Limited", 49432, 28406640),
        h("Sahas Urja Limited", 31522, 11396994),
        h("Upper Solu Hydro Electric Company Limited", 22768, 10345806),
        h("Bujjal Hydro Limited", 2442, 928204),
        h("Darmakhola Hydro Energy Limited", 9393, 540624),
        h("Him Star Urja Company Limited", 646, 463452),
        h("Mabilung Energy Limited", 700, 505400),
        h("Mandu Hydropower Limited", 19993, 16934964),
        h("Dadi Group Power Limited", 34040, 13513000),
        h("Sagarmatha Jalvidyut Company Limited", 39500, 19493000),
        h("Kalika Power Company Limited", 17940, 5972990),
        h("Sahas Urja Limited", 15966, 6309451),
    ]

    # Manufacturing: Image shows 8 companies
    holdings["manufacturing"] = [
        h("Shivam Cements Limited", 234950, 63609240),
        h("Sarbottam Cement Limited", 63646, 93200452),
        h("Unilever Nepal Limited", 760, 36036000),
        h("Himalayan Distillery Limited", 26000, 20402000),
        h("SY Panel Nepal Limited", 3202, 4530334),
        h("Shrinagar Agritech Industries Limited", 2952, 2369000),
        h("Sagar Distillery Limited", 1086, 2242380),
        h("Shivam Cements Limited", 8640, 3919104),
    ]
    # Remove duplicate Shivam - keep the larger one from image
    holdings["manufacturing"] = [
        h("Shivam Cements Limited", 234950, 63609240),
        h("Sarbottam Cement Limited", 63646, 93200452),
        h("Unilever Nepal Limited", 760, 36036000),
        h("Himalayan Distillery Limited", 26000, 20402000),
        h("SY Panel Nepal Limited", 3202, 4530334),
        h("Shrinagar Agritech Industries Limited", 2952, 2369000),
        h("Sagar Distillery Limited", 1086, 2242380),
    ]

    # Hotels: Image shows 3 companies
    holdings["hotels"] = [
        h("Soaltee Hotel Limited", 30922, 25399296),
        h("Chandragiri Hills Limited", 8000, 3404000),
        h("Bandipur Kewalkar & Tourism Limited", 3676, 3052060),
    ]

    # Non-life insurance: More companies than JSON
    holdings["non_life_insurance"] = [
        h("Neco Insurance Limited", 13140, 10162001),
        h("Sagarmatha Lumbini Insurance Company Limited", 7560, 5005022),
        h("Siddharth Premier Insurance Limited", 10860, 5453795),
        h("Shikhar Insurance Company Limited", 13140, 10162001),
    ]

    # Others: More companies
    holdings["others"] = [
        h("Nepal Telecom", 17640, 16299360),
        h("Citizen Investment Trust", 15300, 22009050),
        h("CEIDB Holdings Limited", 22202, 46666304),
        h("Jalvidyut Lagani Tatha Vikas Company Limited", 95948, 15275921),
        h("Trade Tower Limited", 2756, 2303462),
        h("Jhapa Energy Limited", 399, 563222),
    ]

    # NEW: Mutual funds
    holdings["mutual_funds"] = [
        h("NMB Hybrid Fund L-2", 2000000, 9240000),
        h("Garima Samriddhi Yojana", 2000000, 9060000),
        h("RBB Focus 40", 750000, 9500000),
        h("MBL Equity Fund", 499900, 4628034),
        h("NIC Asia Growth Fund 2", 500000, 4550000),
        h("Kumari Sabal Yojana", 279450, 2268942),
        h("NIBL Stable Fund", 80000, 280200),
    ]


def fix_nfcf(fund):
    """NFCF - Nabil Flexi Cap Fund (image: scripts/images/NFCF.jpg)
    Very detailed portfolio with 100+ holdings visible.
    """
    holdings = fund["holdings"]

    # Hydropower: Image shows 30+ hydropower companies
    holdings["hydropower"] = [
        h("Aankhu Khola Jalvidyut Company Limited", 5000, 9048000),
        h("Aap Power Company Limited", 10000, 5500000),
        h("Arun Kabeli Power Limited", 200200, 6250000),
        h("Arun Valley Hydropower Development Company Limited", 224996, 18816795),
        h("Asian Hydropower Limited", 9430, 8022000),
        h("Vikash Hydropower Company Limited", 5000, 2650000),
        h("Buddha Bhumi Nepal Hydropower Company Limited", 32220, 20060000),
        h("Bujjal Hydro Limited", 2209, 1323655),
        h("Butwal Power Company Limited", 50002, 9663769),
        h("Darmakhola Hydro Energy Limited", 9685, 1939600),
        h("Green Ventures Limited", 4036, 3500000),
        h("Him Star Urja Company Limited", 850, 618000),
        h("Himalayan Power Partner Limited", 6000, 3226000),
        h("Indra Hydropower Limited", 6500, 4555000),
        h("Kalika Power Company Limited", 3800, 2200000),
        h("Mabilung Energy Limited", 1432, 1035000),
        h("Mailud Khola Jalvidyut Company Limited", 14200, 2926000),
        h("Mandakini Hydropower Limited", 20000, 7500000),
        h("Mountain Energy Nepal Limited", 27604, 6528000),
        h("National Hydro Power Company Limited", 29200, 26254000),
        h("Nepal Hydro Developers Limited", 29334, 5800000),
        h("Ridi Group Power Limited", 25005, 50293772),
        h("Peoples Hydropower Company Limited", 27302, 20506646),
        h("Sanima Mai Hydropower Limited", 27630, 10081358),
        h("Sanima Middle Tamor Hydropower Limited", 20000, 8000000),
        h("Chilime Hydropower Company Limited", 23130, 17002632),
        h("Upper Tamakoshi Hydropower Limited", 28410, 18816795),
        h("Ridi Hydropower Development Company Limited", 26605, 3666364),
        h("Peoples Hydropower Company Limited", 27302, 20506646),
        h("Sagarmatha Jalvidyut Company Limited", 15000, 7500000),
        h("Sahas Urja Limited", 14500, 5500000),
        h("Universal Power Company Limited", 23000, 8262300),
    ]
    # Deduplicate
    seen = set()
    deduped = []
    for item in holdings["hydropower"]:
        if item["name"] not in seen:
            seen.add(item["name"])
            deduped.append(item)
    holdings["hydropower"] = deduped

    # Non-life insurance: Image shows 9 companies
    holdings["non_life_insurance"] = [
        h("Himalayan Everest Insurance Limited", 11360, 5936003),
        h("IGI Prudential Insurance Limited", 65484, 8000000),
        h("Neco Insurance Limited", 9050, 628900),
        h("Nepal Insurance Company Limited", 96226, 7000000),
        h("Nepal Micro Insurance Company Limited", 1932, 800000),
        h("Prabhu Insurance Limited", 23905, 16695548),
        h("Sagarmatha Lumbini Insurance Company Limited", 6300, 5203634),
        h("Sanima GIC Insurance Limited", 45322, 7000000),
        h("United Ajod Insurance Limited", 46229, 7076554),
    ]

    # Life insurance: Image shows many companies
    holdings["life_insurance"] = [
        h("Asian Life Insurance Company Limited", 35502, 22643400),
        h("Citizen Life Insurance Company Limited", 55946, 12000000),
        h("Nepal Life Insurance Company Limited", 28290, 42835586),
        h("National Life Insurance Company Limited", 6966, 4020000),
        h("Prabhu Mahalaxmi Life Insurance Limited", 20903, 9664930),
        h("Reliable Nepal Life Insurance Limited", 6840, 3935964),
        h("Sun Nepal Life Insurance Company Limited", 4209, 2995000),
        h("Suryajyoti Life Insurance Company Limited", 23000, 20244400),
        h("Life Insurance Corporation Nepal Limited", 6930, 6580732),
        h("Surya Life Insurance Company Limited", 15330, 6821237),
        h("Himalayan Life Insurance Company Limited", 15000, 5500000),
    ]

    # Microfinance: Many more companies
    holdings["microfinance"] = [
        h("Chhimek Laghubitta Bittiya Sanstha Limited", 4800, 9165312),
        h("Nirdhan Utthan Laghubitta Bittiya Sanstha Limited", 5640, 6883056),
        h("Mithila Laghubitta Bittiya Sanstha Limited", 57, 34295),
        h("National Laghubitta Bittiya Sanstha Limited", 3000, 2500000),
        h("Sana Kisan Bikas Laghubitta Bittiya Sanstha Limited", 4320, 4478976),
        h("Swabalamban Laghubitta Bittiya Sanstha Limited", 2000, 1500000),
        h("Swadhin Laghubitta Bittiya Sanstha Limited", 1500, 1200000),
        h("Ashalay Laghubitta Bittiya Sanstha Limited", 79154, 8000000),
    ]

    # Manufacturing
    holdings["manufacturing"] = [
        h("Unilever Nepal Limited", 1020, 19278000),
        h("Shivam Cements Limited", 20052, 10000000),
        h("Sarbottam Cement Limited", 7000, 6100000),
        h("Shrinagar Agritech Industries Limited", 2952, 2369000),
        h("SY Panel Nepal Limited", 3706, 5130000),
    ]

    # Hotels: 4 companies
    holdings["hotels"] = [
        h("Bandipur Kewalkar & Tourism Limited", 5203, 4326840),
        h("Chandragiri Hills Limited", 20500, 6934500),
        h("Oriental Hotels Limited", 30000, 22220000),
        h("Taragaon Regency Hotels Limited", 20000, 7200000),
    ]

    # Others: Many more companies
    holdings["others"] = [
        h("Himalayan Reinsurance Limited", 79552, 69039440),
        h("Jhapa Energy Limited", 462, 629604),
        h("Muktinath Krishi Company Limited", 9344, 20026942),
        h("Nepal Telecom", 24000, 22904000),
        h("Nepal Punarbima Company Limited", 23000, 29043000),
        h("Trade Tower Limited", 3422, 2856054),
    ]

    # NEW: Investment
    holdings["investment"] = [
        h("CEIDB Holdings Limited", 21646, 26936930),
        h("Citizen Investment Trust", 26429, 39493046),
        h("Jalvidyut Lagani Tatha Vikas Company Limited", 202823, 25660946),
        h("Nepal Infrastructure Bank Limited", 10000, 5000000),
        h("NRN Infrastructure & Development Limited", 23494, 26655455),
    ]

    # NEW: Mutual funds
    holdings["mutual_funds"] = [
        h("HLI Large Cap Fund", 250000, 2250000),
        h("MBL Equity Fund", 300000, 2442400),
        h("NMB Hybrid Fund L-2", 250000, 2320000),
        h("Reliable Samriddhi Yojana", 500000, 4695000),
        h("Sunrise Bluechip Fund", 326362, 2954936),
    ]


def fix_naddf(fund):
    """NADDF - NIC Asia Dynamic Debt Fund (image: scripts/images/NADDF.jpg)
    The JSON only had government_bonds, corporate_debentures, fixed_deposits, commercial_banks.
    Image shows full equity portfolio across all sectors.
    """
    holdings = fund["holdings"]

    # Commercial banks: Image shows 8 (4 visible + 4 cut off at top)
    holdings["commercial_banks"] = [
        h("NIC Asia Bank Limited", 80000, 38000000),
        h("Himalayan Bank Limited", 50000, 20800000),
        h("Global IME Bank Limited", 60000, 22700000),
        h("Everest Bank Limited", 25000, 37600000),
        h("Nabil Bank Limited", 68402, 33345466),
        h("Nepal Bank Limited", 49433, 22993353),
        h("Prime Commercial Bank Limited", 60724, 14520646),
        h("Sanima Bank Limited", 42222, 23657777),
    ]

    # Development banks
    holdings["development_banks"] = [
        h("Garima Bikas Bank Limited", 80646, 23692548),
        h("Kamana Sewa Bikas Bank Limited", 34200, 15432800),
        h("Lumbini Bikas Bank Limited", 26000, 9766000),
        h("Muktinath Bikas Bank Limited", 54622, 29029586),
        h("Shangrila Development Bank Limited", 35000, 6029000),
        h("Shine Resunga Development Bank Limited", 67487, 26222296),
    ]

    # Microfinance
    holdings["microfinance"] = [
        h("Diprox Laghubitta Bittiya Sanstha Limited", 2376, 2000000),
        h("Nirdhan Utthan Laghubitta Bittiya Sanstha Limited", 33069, 23432054),
        h("Sana Kisan Bikas Laghubitta Bittiya Sanstha Limited", 28302, 22589082),
        h("Swastik Laghubitta Bittiya Sanstha Limited", 206, 100000),
    ]

    # Life insurance
    holdings["life_insurance"] = [
        h("Asian Life Insurance Company Limited", 226932, 45593623),
        h("National Life Insurance Company Limited", 43000, 25000000),
        h("Nepal Life Insurance Company Limited", 84608, 65000000),
        h("Suryajyoti Life Insurance Company Limited", 63346, 30000000),
    ]

    # Non-life insurance
    holdings["non_life_insurance"] = [
        h("Himalayan Everest Insurance Limited", 50992, 97260562),
        h("IGI Prudential Insurance Limited", 99660, 15000000),
        h("Neco Insurance Limited", 66045, 24648000),
        h("Nepal Insurance Company Limited", 55573, 12000000),
        h("Prabhu Insurance Limited", 45093, 10000000),
        h("Sagarmatha Lumbini Insurance Company Limited", 43646, 8000000),
        h("Sanima GIC Insurance Limited", 37268, 55999000),
        h("Shikhar Insurance Company Limited", 45000, 36000000),
        h("Siddharth Premier Insurance Limited", 30000, 15000000),
        h("United Ajod Insurance Limited", 42422, 29048950),
    ]

    # Mutual funds
    holdings["mutual_funds"] = [
        h("Kumari Equity Fund", 300000, 2867000),
        h("Kumari Sabal Yojana", 500000, 8580000),
        h("MBL Equity Fund", 300000, 2899000),
        h("NIBL Growth Fund", 2000000, 14020000),
        h("Reliable Samriddhi Yojana", 1500000, 14665000),
    ]

    # Hydropower
    holdings["hydropower"] = [
        h("Aap Power Company Limited", 59000, 17096000),
        h("Arun Kabeli Power Limited", 23000, 3926000),
        h("Vikash Hydropower Company Limited", 5093, 2709476),
        h("Bujjal Hydro Limited", 4665, 3222722),
        h("Chirkhwa Hydropower Company Limited", 4, 2000),
        h("Darmakhola Hydro Energy Limited", 5200, 3304600),
        h("Him Star Urja Company Limited", 621, 726079),
        h("Mabilung Energy Limited", 1353, 980925),
        h("Dadi Rup Power Limited", 6988, 3568236),
        h("Rasuwagadhi Hydropower Company Limited", 23002, 3520540),
        h("Ridi Power Company Limited", 26605, 3666364),
        h("Sanbhi Energy Limited", 3473, 2856225),
    ]

    # Hotels
    holdings["hotels"] = [
        h("Bandipur Kewalkar & Tourism Limited", 7758, 6462360),
        h("Chandragiri Hills Limited", 23000, 12063000),
        h("Oriental Hotels Limited", 4020, 2630000),
        h("Soaltee Hotel Limited", 36000, 16426000),
    ]

    # Manufacturing
    holdings["manufacturing"] = [
        h("Sagar Distillery Limited", 2422, 4636046),
        h("Sarbottam Cement Limited", 24500, 22693000),
        h("Shivam Cements Limited", 20590, 6522000),
        h("Shrinagar Agritech Industries Limited", 3698, 5067600),
        h("SY Panel Nepal Limited", 3492, 4635035),
    ]

    # Others
    holdings["others"] = [
        h("Nepal Punarbima Company Limited", 9000, 11799000),
        h("Himalayan Reinsurance Limited", 6992, 5876776),
        h("Jhapa Energy Limited", 1020, 2450440),
        h("Trade Tower Limited", 3285, 2827952),
    ]


def fix_ksly(fund):
    """KSLY - Kumari Sunaulo Lagani Yojana (image: scripts/images/KSLY.jpg)"""
    holdings = fund["holdings"]

    # Life insurance: Image shows 4
    holdings["life_insurance"] = [
        h("Himalayan Life Insurance Company Limited", 30000, 10947000),
        h("Life Insurance Corporation Nepal Limited", 30000, 25602000),
        h("Nepal Life Insurance Company Limited", 38150, 29337350),
        h("National Life Insurance Company Limited", 44276, 25546406),
    ]

    # Non-life: Image shows 5 (incl promoter shares)
    holdings["non_life_insurance"] = [
        h("Neco Insurance Limited", 29916, 16469324),
        h("Prabhu Insurance Limited", 15000, 10445000),
        h("Sagarmatha Lumbini Insurance Company Limited", 34500, 20603400),
        h("Siddharth Premier Insurance Limited", 33293, 16709069),
    ]

    # Hotels
    holdings["hotels"] = [
        h("Chandragiri Hills Limited", 22950, 19530040),
        h("Bandipur Kewalkar & Tourism Limited", 2462, 2043860),
    ]

    # Hydropower: Image shows 13 companies
    holdings["hydropower"] = [
        h("Api Power Company Limited", 46000, 13566400),
        h("Butwal Power Company Limited", 13630, 10533900),
        h("Bujjal Hydro Limited", 968, 694698),
        h("Darmakhola Hydro Energy Limited", 9393, 540624),
        h("Him Star Urja Company Limited", 420, 369000),
        h("Mabilung Energy Limited", 700, 505400),
        h("Mandu Hydropower Limited", 19993, 16934964),
        h("Mountain Energy Nepal Limited", 29092, 29483024),
        h("Dadi Group Power Limited", 34040, 13513000),
        h("Sahas Urja Limited", 15966, 6309451),
        h("Sagarmatha Jalvidyut Company Limited", 39500, 19493000),
        h("Kalika Power Company Limited", 17940, 5972990),
        h("Nepal Hydro Jalvidyut Pariyojana Limited", 97320, 99499800),
    ]

    # Manufacturing: 6 companies
    holdings["manufacturing"] = [
        h("Bottlers Nepal (Terai) Limited", 1070, 12429000),
        h("Sagar Distillery Limited", 577, 1439565),
        h("Shrinagar Agritech Industries Limited", 1447, 1549900),
        h("Sarbottam Cement Limited", 27366, 23934364),
        h("Shivam Cements Limited", 30900, 16599500),
        h("SY Panel Nepal Limited", 2451, 3346635),
    ]

    # Investment
    holdings["investment"] = [
        h("Citizen Investment Trust", 15090, 26769440),
        h("Hydroelectricity Investment & Development Company Limited", 137630, 24697267),
        h("Nepal Infrastructure Bank Limited", 7000, 1620000),
    ]

    # Others
    holdings["others"] = [
        h("Himalayan Reinsurance Limited", 29500, 23993750),
        h("Nepal Telecom", 11362, 9640554),
        h("Jhapa Energy Limited", 265, 376630),
    ]

    # Trading
    holdings["trading"] = [
        h("Salt Trading Corporation Limited", 3590, 19641609),
    ]

    # Mutual funds
    holdings["mutual_funds"] = [
        h("Global IME Samunnat Yojana-2", 500000, 4550000),
        h("Himalayan 50-20", 152393, 1729937),
        h("MBL Equity Fund", 406476, 3957294),
        h("NIBL Stable Fund", 9300000, 11264000),
        h("NIC Asia Growth Fund 2", 600000, 5460000),
        h("NMB Hybrid Fund-2", 500000, 4620000),
        h("Prabhu Smart Fund", 567300, 6234627),
        h("Reliable Samriddhi Yojana", 250000, 2447500),
    ]


def fix_csdy(fund):
    """CSDY - Citizens Sadabahar Yojana (image: scripts/images/CSDY.jpg)"""
    holdings = fund["holdings"]

    # Hydropower: 12 companies from image
    holdings["hydropower"] = [
        h("Chilime Hydropower Company Limited", 10964, 5953060),
        h("Shiv Shri Hydropower Limited", 37695, 5924496),
        h("Mailud Khola Jalvidyut Company Limited", 3500, 1602500),
        h("Bujjal Hydro Limited", 1544, 1174996),
        h("Jhapa Energy Limited", 507, 720954),
        h("Mabilung Energy Limited", 1336, 930050),
        h("Mountain Energy Nepal Limited", 42622, 24550000),
        h("Sanima Mai Hydropower Limited", 2000, 1903200),
        h("United Modi Hydropower Limited", 16000, 10490200),
        h("Siganti Hydro Energy Limited", 49669, 11547663),
        h("Universal Power Company Limited", 4449, 1674904),
        h("Sahas Urja Limited", 51446, 26966349),
    ]

    # Life insurance: just National Life
    holdings["life_insurance"] = [
        h("National Life Insurance Company Limited", 20000, 11540000),
    ]

    # Non-life: 3 companies
    holdings["non_life_insurance"] = [
        h("Sagarmatha Lumbini Insurance Company Limited", 12660, 7566640),
        h("Shikhar Insurance Company Limited", 10246, 8529282),
        h("Siddharth Premier Insurance Limited", 12867, 9143939),
    ]

    # Hotels
    holdings["hotels"] = [
        h("Soaltee Hotel Limited", 19250, 5590500),
        h("Bandipur Kewalkar & Tourism Limited", 4706, 3904960),
        h("Chandragiri Hills Limited", 15945, 13566195),
    ]

    # Manufacturing
    holdings["manufacturing"] = [
        h("Shivam Cements Limited", 960, 990700),
        h("Sagar Distillery Limited", 1367, 2940635),
        h("Shrinagar Agritech Industries Limited", 2755, 3030500),
        h("SY Panel Nepal Limited", 4966, 5797690),
        h("Sarbottam Cement Limited", 6000, 5244000),
    ]

    # Others
    holdings["others"] = [
        h("Nepal Punarbima Company Limited", 14660, 19407680),
        h("Himalayan Reinsurance Limited", 34695, 29969196),
    ]

    # Investment
    holdings["investment"] = [
        h("NRN Infrastructure & Development Limited", 1709, 2362693),
    ]

    # Mutual funds
    holdings["mutual_funds"] = [
        h("Prabhu Smart Fund", 94000, 1033060),
        h("HLI Large Cap Fund", 2000000, 16000000),
    ]


def fix_slk(fund):
    """SLK - Shubha Laxmi Kosh (image: scripts/images/SLK.jpg)"""
    holdings = fund["holdings"]

    # Microfinance: 9 companies from image
    holdings["microfinance"] = [
        h("Diprox Laghubitta Bittiya Sanstha Limited", 11000, 9068000),
        h("First Microfinance Laghubitta Bittiya Sanstha Limited", 12000, 8693200),
        h("Global IME Laghubitta Bittiya Sanstha Limited", 20950, 24364650),
        h("Nerude Mimire Laghubitta Bittiya Sanstha Limited", 14632, 8278200),
        h("Nirdhan Utthan Laghubitta Bittiya Sanstha Limited", 20946, 16639168),
        h("RSDC Laghubitta Bittiya Sanstha Limited", 11551, 7399486),
        h("Sana Kisan Bikas Laghubitta Bittiya Sanstha Limited", 3054, 2354634),
        h("Swarojgar Laghubitta Bittiya Sanstha Limited", 6855, 5609220),
        h("Swastik Laghubitta Bittiya Sanstha Limited", 36, 102846),
    ]

    # Life insurance: 7 companies
    holdings["life_insurance"] = [
        h("Asian Life Insurance Company Limited", 24172, 11450276),
        h("Best Micro Life Insurance Limited", 395, 486205),
        h("Guardian Micro Life Insurance Limited", 429, 596199),
        h("Himalayan Life Insurance Company Limited", 29484, 10784645),
        h("IME Life Insurance Company Limited", 5124, 3696643),
        h("National Life Insurance Company Limited", 6966, 4020436),
        h("Suryajyoti Life Insurance Company Limited", 25000, 11039500),
    ]

    # Non-life: 7 companies
    holdings["non_life_insurance"] = [
        h("Neco Insurance Limited", 13651, 5559999),
        h("Nepal Insurance Company Limited", 11966, 6092329),
        h("Nepal Micro Insurance Company Limited", 463, 469000),
        h("Prabhu Insurance Limited", 23662, 18497283),
        h("Sagarmatha Lumbini Insurance Company Limited", 5009, 5094000),
        h("Shikhar Insurance Company Limited", 5009, 5094000),
        h("Siddharth Premier Insurance Limited", 5626, 4036296),
    ]

    # Hydropower: 17 companies
    holdings["hydropower"] = [
        h("Api Power Company Limited", 39500, 8992950),
        h("Arun Kabeli Power Limited", 20000, 4920000),
        h("Arun Valley Hydropower Development Company Limited", 29000, 5428400),
        h("Vikash Hydropower Company Limited", 552, 293664),
        h("Bujjal Hydro Limited", 453, 268652),
        h("Darmakhola Hydro Energy Limited", 455, 294600),
        h("Green Ventures Limited", 32000, 11069600),
        h("Him Star Urja Company Limited", 145, 130355),
        h("Jhapa Energy Limited", 925, 757540),
        h("Mabilung Energy Limited", 233, 168925),
        h("Mandakini Hydropower Limited", 95000, 73426000),
        h("Mountain Energy Nepal Limited", 19294, 11993384),
        h("Dadi Group Power Limited", 26188, 10396636),
        h("Ridi Hydropower Development Company Limited", 11996, 2646319),
        h("Sahas Urja Limited", 12900, 6629490),
        h("Sanima Mai Hydropower Limited", 20000, 11030000),
        h("Universal Power Company Limited", 23000, 8262300),
    ]

    # Manufacturing: 5 companies
    holdings["manufacturing"] = [
        h("Himalayan Distillery Limited", 12360, 14078080),
        h("Sagar Distillery Limited", 336, 673660),
        h("Shivam Cements Limited", 7693, 4869195),
        h("Shrinagar Agritech Industries Limited", 617, 764700),
        h("SY Panel Nepal Limited", 1029, 1425965),
    ]

    # Trading
    holdings["trading"] = [
        h("Salt Trading Corporation Limited", 1000, 5555000),
    ]

    # Others
    holdings["others"] = [
        h("Himalayan Reinsurance Limited", 15000, 12607500),
    ]

    # Hotels
    holdings["hotels"] = [
        h("Bandipur Kewalkar & Tourism Limited", 825, 864750),
        h("Soaltee Hotel Limited", 13754, 6849492),
    ]

    # Investment
    holdings["investment"] = [
        h("Nepal Infrastructure Bank Limited", 30000, 7500000),
    ]

    # Finance companies
    holdings["finance_companies"] = [
        h("Manjushree Finance Limited", 3000, 2397000),
    ]


def fix_sff(fund):
    """SFF - Sanima Flexi Fund (image: scripts/images/SFF.jpg)"""
    holdings = fund["holdings"]

    # Microfinance: 9 companies
    holdings["microfinance"] = [
        h("First Microfinance Laghubitta Bittiya Sanstha Limited", 6000, 4506600),
        h("Kalika Laghubitta Bittiya Sanstha Limited", 6657, 6344929),
        h("Suryodaya Bomi Laghubitta Bittiya Sanstha Limited", 5000, 5400000),
        h("National Microfinance Laghubitta Bittiya Sanstha Limited", 4000, 4469000),
        h("Chhimek Laghubitta Bittiya Sanstha Limited", 4000, 4592400),
        h("Nirdhan Utthan Laghubitta Bittiya Sanstha Limited", 6353, 4993924),
        h("Himalayan Laghubitta Bittiya Sanstha Limited", 90000, 7000000),
        h("Swastik Laghubitta Bittiya Sanstha Limited", 927, 369943),
        h("Sana Kisan Bikas Laghubitta Bittiya Sanstha Limited", 90000, 7690000),
    ]

    # Non-life: 6 companies
    holdings["non_life_insurance"] = [
        h("Neco Insurance Limited", 3622, 2386796),
        h("IGI Prudential Insurance Company Limited", 7932, 3004505),
        h("Nepal Insurance Company Limited", 7068, 3597692),
        h("NLG Insurance Company Limited", 6669, 4863539),
        h("Shikhar Insurance Company Limited", 4076, 3305474),
        h("Siddharth Premier Insurance Company Limited", 7000, 3545000),
    ]

    # Life insurance: 6 companies
    holdings["life_insurance"] = [
        h("Citizen Life Insurance Company Limited", 93000, 9962000),
        h("IME Life Insurance Company Limited", 4924, 4867928),
        h("Life Insurance Corporation Nepal Limited", 5000, 6629200),
        h("Himalayan Life Insurance Company Limited", 15000, 5963540),
        h("Sun Nepal Life Insurance Company Limited", 90000, 4742000),
        h("Nepal Life Insurance Company Limited", 95000, 99534000),
    ]

    # Hydropower: 11 companies
    holdings["hydropower"] = [
        h("Peoples Hydropower Company Limited", 15000, 3990000),
        h("Sanima Middle Tamor Hydropower Limited", 12996, 5836902),
        h("Super Madi Hydropower Limited", 10000, 5946000),
        h("Api Power Company Limited", 11000, 3962300),
        h("Him Star Urja Company Limited", 596, 465662),
        h("Mountain Energy Nepal Limited", 15000, 6640000),
        h("Sahas Urja Limited", 7000, 3345000),
        h("Sanima Mai Hydropower Limited", 9000, 4963300),
        h("Darmakhola Hydro Energy Limited", 1598, 1035504),
        h("Mabilung Energy Limited", 624, 597500),
        h("Bujjal Hydro Limited", 1175, 746790),
    ]

    # Manufacturing: 4 companies
    holdings["manufacturing"] = [
        h("Himalayan Distillery Limited", 10600, 12301200),
        h("Sagar Distillery Limited", 570, 1744340),
        h("Shrinagar Agritech Industries Limited", 1755, 1930500),
        h("SY Panel Nepal Limited", 2676, 3700030),
    ]

    # Mutual funds: 5 from image
    holdings["mutual_funds"] = [
        h("Global IME Samunnat Yojana 2", 246000, 2236600),
        h("MBL Equity Fund", 400000, 4625000),
        h("Reliable Samriddhi Yojana", 250000, 2447500),
        h("HLI Large Cap Fund", 290000, 1690000),
        h("RBB Focus 40", 490000, 4900000),
    ]

    # Hotels: 3 companies
    holdings["hotels"] = [
        h("Oriental Hotels Limited", 2000, 1406000),
        h("Soaltee Hotel Limited", 10000, 4960000),
        h("Bandipur Kewalkar & Tourism Limited", 2696, 2505280),
    ]

    # Others
    holdings["others"] = [
        h("Pure Energy Limited", 900, 943200),
        h("Jhapa Energy Limited", 323, 459306),
        h("Nepal Punarbima Company Limited", 3000, 3933000),
        h("Muktinath Krishi Company Limited", 2767, 3765256),
    ]

    # Investment
    holdings["investment"] = [
        h("Citizen Investment Trust", 14000, 24636000),
    ]


def fix_psis(fund):
    """PSIS - Prabhu Systematic Investment Scheme (image: scripts/images/PSIS.jpg)"""
    holdings = fund["holdings"]

    # Hydropower: 11 companies
    holdings["hydropower"] = [
        h("Api Power Company Limited", 26166, 9629664),
        h("Bujjal Hydro Limited", 1327, 575932),
        h("Darmakhola Hydro Energy Limited", 3429, 2000000),
        h("Green Ventures Limited", 4036, 3200000),
        h("Him Star Urja Company Limited", 391, 280000),
        h("Mabilung Energy Limited", 2375, 1717250),
        h("Mandakini Hydropower Limited", 45000, 34000000),
        h("Mountain Energy Nepal Limited", 49432, 28406640),
        h("Sahas Urja Limited", 31522, 11396994),
        h("Sanima Middle Tamor Hydropower Limited", 31645, 14923440),
        h("Upper Solu Hydro Electric Company Limited", 22768, 10345806),
    ]

    # Non-life: 3 companies
    holdings["non_life_insurance"] = [
        h("Himalayan Everest Insurance Limited", 10326, 6561990),
        h("Neco Insurance Limited", 1960, 1300000),
        h("Sagarmatha Lumbini Insurance Company Limited", 5036, 4222000),
    ]

    # Microfinance: 3 companies
    holdings["microfinance"] = [
        h("Chhimek Laghubitta Bittiya Sanstha Limited", 12329, 11246063),
        h("Forward Microfinance Laghubitta Bittiya Sanstha Limited", 9633, 6255628),
        h("Swastik Laghubitta Bittiya Sanstha Limited", 500, 376062),
    ]

    # Life insurance: 2 companies
    holdings["life_insurance"] = [
        h("Nepal Life Insurance Company Limited", 21582, 16591996),
        h("National Life Insurance Company Limited", 20072, 16992920),
    ]

    # Manufacturing: 6 companies
    holdings["manufacturing"] = [
        h("Himalayan Distillery Limited", 4000, 5592000),
        h("Sagar Distillery Limited", 434, 1933000),
        h("Sarbottam Cement Limited", 45585, 24606946),
        h("Shivam Cements Limited", 8640, 3919104),
        h("Shrinagar Agritech Industries Limited", 1300, 1430000),
        h("SY Panel Nepal Limited", 1474, 2033535),
    ]

    # Hotels
    holdings["hotels"] = [
        h("Bandipur Kewalkar & Tourism Limited", 2270, 1723100),
    ]

    # Others
    holdings["others"] = [
        h("Jhapa Energy Limited", 226, 339580),
    ]

    # Investment
    holdings["investment"] = [
        h("CEIDB Holdings Limited", 2000, 4930000),
        h("NRN Infrastructure & Development Limited", 1000, 1363000),
    ]

    # Mutual funds
    holdings["mutual_funds"] = [
        h("Siddhartha Equity Fund", 300000, 1946000),
    ]


def fix_elis(fund):
    """ELIS - NIC ASIA Equity Linked Investment Scheme (image: scripts/images/ELIS.jpg)"""
    holdings = fund["holdings"]

    # Commercial banks: 6 from image
    holdings["commercial_banks"] = [
        h("Sanima Bank Limited", 69000, 32649000),
        h("Nabil Bank Limited", 23000, 34845000),
        h("Kumari Bank Limited", 48000, 16272000),
        h("Nepal Bank Limited", 16000, 4333000),
        h("Prabhu Bank Limited", 15500, 4165850),
        h("Siddhartha Bank Limited", 25500, 10957050),
    ]

    # Development banks: 8 companies
    holdings["development_banks"] = [
        h("Jyoti Bikas Bank Limited", 23000, 7629300),
        h("Muktinath Bikas Bank Limited", 61929, 23033000),
        h("Kamana Sewa Bikas Bank Limited", 34000, 25402000),
        h("Garima Bikas Bank Limited", 26620, 20096500),
        h("Shangrila Development Bank Limited", 29000, 9434500),
        h("Shine Resunga Development Bank Limited", 21988, 4578019),
        h("Lumbini Bikas Bank Limited", 32000, 15106400),
        h("Mahalaxmi Bikas Bank Limited", 20000, 7300000),
    ]

    # Hydropower: 5 companies
    holdings["hydropower"] = [
        h("Arun Kabeli Power Limited", 25000, 6250000),
        h("Dadi Rup Power Limited", 24383, 9600053),
        h("Butwal Power Company Limited", 30000, 12360000),
        h("Ridi Power Company Limited", 25000, 4520000),
        h("Bujjal Hydro Limited", 3000, 1926000),
    ]

    # Life insurance: 2 companies
    holdings["life_insurance"] = [
        h("Nepal Life Insurance Company Limited", 22356, 16591996),
        h("Suryajyoti Life Insurance Company Limited", 25000, 22039000),
    ]

    # Manufacturing: 4 companies
    holdings["manufacturing"] = [
        h("Sarbottam Cement Limited", 3140, 9548360),
        h("Shivam Cements Limited", 9500, 5100000),
        h("SY Panel Nepal Limited", 4000, 5633600),
        h("Shrinagar Agritech Industries Limited", 2913, 3204300),
    ]

    # Others
    holdings["others"] = [
        h("Nepal Punarbima Company Limited", 25000, 29665000),
        h("Himalayan Reinsurance Limited", 22000, 16492000),
    ]

    # Non-life: 2 companies
    holdings["non_life_insurance"] = [
        h("NLG Insurance Company Limited", 12932, 6588497),
        h("Prabhu Insurance Limited", 15322, 20536336),
    ]

    # Microfinance: 2 companies
    holdings["microfinance"] = [
        h("Nirdhan Utthan Laghubitta Bittiya Sanstha Limited", 11364, 6045722),
        h("Chhimek Laghubitta Bittiya Sanstha Limited", 22995, 22954202),
    ]

    # Hotels: 3 companies
    holdings["hotels"] = [
        h("Chandragiri Hills Limited", 15000, 12765000),
        h("Soaltee Hotel Limited", 7500, 3935000),
        h("Oriental Hotels Limited", 5512, 5204544),
    ]


def fix_ni31(fund):
    """NI31 - NI 31 (image: scripts/images/NI31.jpg)"""
    holdings = fund["holdings"]

    # Commercial banks: 5 from image
    holdings["commercial_banks"] = [
        h("Agriculture Development Bank Limited", 32836, 9650600),
        h("Everest Bank Limited", 16242, 10699520),
        h("Global IME Bank Limited", 96266, 22956000),
        h("Nepal Investment Mega Bank Limited", 22000, 4999600),
        h("Prime Commercial Bank Limited", 40240, 9610360),
    ]

    # Development banks: 3
    holdings["development_banks"] = [
        h("Garima Bikas Bank Limited", 14403, 9548693),
        h("Kamana Sewa Bikas Bank Limited", 15767, 7242451),
        h("Muktinath Bikas Bank Limited", 26550, 7209990),
    ]

    # Finance companies
    holdings["finance_companies"] = [
        h("ICFC Finance Limited", 3462, 2296920),
        h("Manjushree Finance Limited", 500, 399500),
    ]

    # Microfinance: 3
    holdings["microfinance"] = [
        h("Chhimek Laghubitta Bittiya Sanstha Limited", 9568, 6620922),
        h("National Laghubitta Bittiya Sanstha Limited", 2198, 2327752),
        h("Swarojgar Laghubitta Bittiya Sanstha Limited", 2296, 2090679),
    ]

    # Life insurance: 3
    holdings["life_insurance"] = [
        h("Citizen Life Insurance Company Limited", 1267, 600556),
        h("Life Insurance Corporation Nepal Limited", 6642, 4332963),
        h("Nepal Life Insurance Company Limited", 7775, 5906879),
    ]

    # Non-life: 2
    holdings["non_life_insurance"] = [
        h("Neco Insurance Limited", 6475, 4002550),
        h("Siddharth Premier Insurance Limited", 4636, 4049022),
    ]

    # Hydropower: 5
    holdings["hydropower"] = [
        h("Arun Valley Hydropower Development Company Limited", 32020, 6279960),
        h("Butwal Power Company Limited", 16350, 12652100),
        h("Nepal Hydro Developers Limited", 2586, 1006960),
        h("Sahas Urja Limited", 9740, 6434346),
        h("Sanima Mai Hydropower Limited", 6230, 3836565),
    ]

    # Manufacturing: 3
    holdings["manufacturing"] = [
        h("Himalayan Distillery Limited", 2996, 3493583),
        h("Sarbottam Cement Limited", 3520, 3076800),
        h("SY Panel Nepal Limited", 602, 833770),
    ]

    # Hotels: 2
    holdings["hotels"] = [
        h("Soaltee Hotel Limited", 6508, 4336548),
        h("Taragaon Regency Hotels Limited", 2994, 2036400),
    ]

    # Investment
    holdings["investment"] = [
        h("CEIDB Holdings Limited", 4500, 10372500),
        h("NRN Infrastructure & Development Limited", 3424, 4733680),
    ]

    # Others
    holdings["others"] = [
        h("Himalayan Reinsurance Limited", 11422, 9600191),
        h("Nepal Telecom", 4937, 4787230),
    ]

    # Remove fixed_deposits if present (NI31 is small, may not have)
    # Keep whatever was in the template for fixed_deposits


# ---------------------------------------------------------------------------
# Company name fixes applied to ALL funds
# ---------------------------------------------------------------------------

COMPANY_RENAMES = {
    "Sagarmatha Insurance Company Limited": "Sagarmatha Lumbini Insurance Company Limited",
}


def apply_renames(fund):
    """Apply company name changes (mergers etc) across all sectors."""
    for sector, items in fund["holdings"].items():
        if not isinstance(items, list):
            continue
        for item in items:
            if item["name"] in COMPANY_RENAMES:
                item["name"] = COMPANY_RENAMES[item["name"]]


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    data = load_nav()
    fund_map = {f["symbol"]: f for f in data}

    fixes = {
        "NIBLSF": fix_niblsf,
        "NMBSBF": fix_nmbsbf,
        "SSIS": fix_ssis,
        "NFCF": fix_nfcf,
        "NADDF": fix_naddf,
        "KSLY": fix_ksly,
        "CSDY": fix_csdy,
        "SLK": fix_slk,
        "SFF": fix_sff,
        "PSIS": fix_psis,
        "ELIS": fix_elis,
        "NI31": fix_ni31,
    }

    for symbol, fix_fn in fixes.items():
        if symbol in fund_map:
            print(f"Fixing {symbol}...")
            fix_fn(fund_map[symbol])
            apply_renames(fund_map[symbol])
        else:
            print(f"WARNING: {symbol} not found in JSON")

    # Apply renames to all funds (including NIBLSF and MSY)
    for fund in data:
        apply_renames(fund)

    save_nav(data)

    # Print summary
    print("\nSummary:")
    for fund in data:
        total_holdings = sum(
            len(v) for v in fund["holdings"].values() if isinstance(v, list)
        )
        sectors = [k for k, v in fund["holdings"].items() if isinstance(v, list) and len(v) > 0]
        print(f"  {fund['symbol']:8s} {total_holdings:3d} holdings across {len(sectors)} sectors")


if __name__ == "__main__":
    main()
