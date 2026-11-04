"""Run: python3 scripts/test_fetch_sipsathi.py"""

from fetch_sipsathi import build_funds, extract_data

DATA = {
    "as_of": "2026-10-12",
    "stock_sector": {"NABIL": "BANKING", "CBBL": "MICROFINANCE", "C30MF": "MUTUAL", "KBLPO": "PROMOTSHARE", "NIBLSF": "MUTUAL"},
    "funds": [
        {
            "sym": "NIBLSF",
            "name": "NIBL Sahabhagita Fund",
            "manager": "NIMB Ace Capital",
            "nav": 10.15,
            "aum": 11628756471.12,
            "alloc": {"cash": 1934212289, "fi": 65000000, "cm": 9583802166},
            "holdings": [["NABIL", 319592955], ["CBBL", 324821790], ["C30MF", 1000], ["KBLPO", 2000]],
            "holdings_full": True,
        },
        {
            "sym": "NI31",
            "name": "Nabil Balanced Fund III",
            "manager": "Nabil Investment Banking",
            "nav": 9.5,
            "aum": 500000000,
            "alloc": {"cash": 123345437, "fi": 0, "cm": 360690789},
            "holdings": None,
            "holdings_full": False,
        },
        {
            "sym": "KSLY",
            "name": "Kumari Sunaulo Lagani Yojana",
            "manager": "Kumari Capital",
            "nav": 11.13,
            "aum": 1500000000,
            "alloc": {"cash": 161278373, "fi": 0, "cm": 1291079323},
            "holdings": [["NABIL", 100000000], ["NIBLSF", 5000000]],
            "holdings_full": False,
        },
    ],
}
NAMES = {
    "NABIL": "Nabil Bank Limited",
    "CBBL": "Chhimek Laghubitta Bittiya Sanstha Limited",
    "KBL": "Kumari Bank Limited",
}


def test_full_fund_maps_to_fund_record():
    fund = build_funds(DATA, NAMES, "2083 Ashwin")[0]
    assert fund["symbol"] == "NIBLSF"
    assert fund["fund_name"] == "NIBL Sahabhagita Fund"
    assert fund["fund_manager"] == "NIMB Ace Capital"
    assert fund["report_date_nepali"] == "2083 Ashwin"
    assert fund["report_date_english"] == "October 2026"
    assert fund["nav_per_unit"] == 10.15
    assert fund["net_assets"] == 11628756471.12
    assert fund["total_units"] == 1145690293  # 11628756471.12 / 10.15 = 1145690292.72, rounded
    assert "total_assets" not in fund and "total_liabilities" not in fund
    assert fund["holdings_disclosure"] == "full"
    h = fund["holdings"]
    assert h["commercial_banks"] == [{"name": "Nabil Bank Limited", "value": 319592955}]
    assert h["microfinance"] == [{"name": "Chhimek Laghubitta Bittiya Sanstha Limited", "value": 324821790}]
    assert h["mutual_funds"] == [{"name": "C30MF", "value": 1000}]  # no name known: symbol
    assert h["fixed_deposits"] == [{"name": "Cash & Fixed Deposits", "value": 1934212289}]
    assert h["corporate_debentures"] == [{"name": "Fixed Income Instruments", "value": 65000000}]
    # Promoter shares are named after the parent; a fully itemised fund gets no remainder row
    assert h["others"] == [{"name": "Kumari Bank Limited (Promoter)", "value": 2000}]


def test_fund_without_itemised_holdings_keeps_share_total():
    fund = build_funds(DATA, NAMES, "2083 Ashwin")[1]
    assert fund["holdings_disclosure"] == "none"
    assert fund["holdings"]["others"] == [{"name": "Other listed shares", "value": 360690789}]
    assert "corporate_debentures" not in fund["holdings"]


def test_top10_fund_keeps_remainder_and_names_other_funds():
    fund = build_funds(DATA, NAMES, "2083 Ashwin")[2]
    assert fund["holdings_disclosure"] == "top10"
    assert fund["holdings"]["mutual_funds"] == [{"name": "NIBL Sahabhagita Fund", "value": 5000000}]
    assert fund["holdings"]["others"] == [{"name": "Other listed shares", "value": 1291079323 - 105000000}]


def test_extract_data_reads_decode_nav_argument():
    js = 'function __decodeNav(r){return r}\nconst DATA=__decodeNav({"as_of":"2026-10-12","funds":[]});\nfoo();'
    assert extract_data(js) == {"as_of": "2026-10-12", "funds": []}


if __name__ == "__main__":
    for name, fn in list(globals().items()):
        if name.startswith("test_"):
            fn()
            print("ok", name)
