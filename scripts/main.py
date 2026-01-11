"""
Mutual Fund NAV Extractor
Fetches open-end mutual fund NAV images from ShareSansar and extracts data using OCR.
"""

import json
import os
import re
import time
from pathlib import Path

import requests
from bs4 import BeautifulSoup

# Output directories
IMAGES_DIR = Path(__file__).parent / "images"
DATA_DIR = Path(__file__).parent / "data"


def fetch_open_end_funds() -> list[dict]:
    """Fetch list of open-end mutual funds from ShareSansar API."""
    url = "https://www.sharesansar.com/mutual-fund-navs"
    params = {
        "draw": 1,
        "start": 0,
        "length": 50,
        "type": 2,  # Open End funds
    }

    headers = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
        "X-Requested-With": "XMLHttpRequest",
    }

    resp = requests.get(url, params=params, headers=headers)
    resp.raise_for_status()

    data = resp.json()
    funds = []

    for item in data.get("data", []):
        # Extract symbol from HTML link
        symbol_match = re.search(r'>([A-Z0-9]+)</a>', item.get("symbol", ""))
        symbol = symbol_match.group(1) if symbol_match else item.get("symbol", "")

        # Extract name from HTML link
        name_match = re.search(r'>([^<]+)</a>', item.get("companyname", ""))
        name = name_match.group(1) if name_match else item.get("companyname", "")

        funds.append({
            "symbol": symbol,
            "name": name,
            "fund_size": item.get("fund_size", ""),
            "daily_nav": item.get("daily_nav_price", ""),
            "daily_date": item.get("daily_date", ""),
            "weekly_nav": item.get("weekly_nav_price", ""),
            "weekly_date": item.get("weekly_date", ""),
            "monthly_nav": item.get("monthly_nav_price", ""),
            "monthly_date": item.get("monthly_date", ""),
        })

    return funds


def fetch_fund_announcements(symbol: str, fund_name: str) -> list[dict]:
    """Fetch announcements using session with CSRF token."""
    session = requests.Session()

    headers = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
    }

    # First visit company page to get CSRF token, company ID, and sector
    company_url = f"https://www.sharesansar.com/company/{symbol.lower()}"
    resp = session.get(company_url, headers=headers)

    soup = BeautifulSoup(resp.text, "html.parser")

    # Get CSRF token from meta tag (that's what the JS uses)
    meta_token = soup.find("meta", {"name": "_token"})
    csrf_token = meta_token.get("content") if meta_token else ""

    # Get company ID, symbol, and sector from hidden elements
    companyid_elem = soup.find(id="companyid")
    symbol_elem = soup.find(id="symbol")
    sector_elem = soup.find(id="sector")

    company_id = companyid_elem.get_text(strip=True) if companyid_elem else ""
    symbol_val = symbol_elem.get_text(strip=True) if symbol_elem else symbol
    sector_val = sector_elem.get_text(strip=True) if sector_elem else ""

    # Make the AJAX request with proper headers and data
    ajax_url = "https://www.sharesansar.com/company-announcements"

    ajax_headers = {
        **headers,
        "X-Requested-With": "XMLHttpRequest",
        "X-CSRF-Token": csrf_token,
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
        "Referer": company_url,
    }

    data = {
        "draw": "1",
        "start": "0",
        "length": "10",
        "company": company_id,
        "symbol": symbol_val,
        "sector": sector_val,
    }

    resp = session.post(ajax_url, data=data, headers=ajax_headers)
    resp.raise_for_status()

    result = resp.json()
    announcements = []

    for item in result.get("data", []):
        title_html = item.get("title", "")
        url_match = re.search(r'href="([^"]+)"', title_html)
        title_match = re.search(r'>([^<]+)</a>', title_html)

        if url_match:
            announcements.append({
                "date": item.get("published_date", ""),
                "title": title_match.group(1) if title_match else "",
                "url": url_match.group(1),
            })

    return announcements


def find_latest_nav_announcement(announcements: list[dict]) -> dict | None:
    """Find the most recent NAV announcement."""
    for ann in announcements:
        title = ann.get("title", "").lower()
        if "nav" in title or "net assets value" in title:
            return ann
    return None


def extract_image_url(announcement_url: str) -> str | None:
    """Extract the NAV image URL from an announcement page."""
    headers = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
    }

    resp = requests.get(announcement_url, headers=headers)
    resp.raise_for_status()

    soup = BeautifulSoup(resp.text, "html.parser")

    # Look for announcement images
    for img in soup.find_all("img"):
        src = img.get("src", "")
        if "announcement" in src and src.endswith((".jpg", ".jpeg", ".png")):
            return src

    return None


def download_image(url: str, symbol: str) -> Path:
    """Download image and save to images directory."""
    IMAGES_DIR.mkdir(exist_ok=True)

    headers = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
    }

    resp = requests.get(url, headers=headers)
    resp.raise_for_status()

    # Get extension from URL
    ext = Path(url).suffix or ".jpg"
    filepath = IMAGES_DIR / f"{symbol}{ext}"

    filepath.write_bytes(resp.content)
    return filepath


def run_easyocr(image_path: Path) -> list:
    """Run EasyOCR on an image."""
    import easyocr

    reader = easyocr.Reader(["ne", "en"], gpu=False)
    results = reader.readtext(str(image_path))
    return results


def extract_nav_data(ocr_results: list) -> dict:
    """Extract structured NAV data from OCR results."""
    text_lines = [r[1] for r in ocr_results]
    full_text = "\n".join(text_lines)

    # Try to find NAV per unit (प्रति इकाई खुद मूल्य)
    nav_patterns = [
        r"(\d+\.\d+)\s*$",  # Number at end of line
        r"NAV[:\s]+(\d+\.\d+)",
        r"खुद मूल्य[:\s]+(\d+[\.,]\d+)",
    ]

    return {
        "raw_text": full_text,
        "line_count": len(text_lines),
        "ocr_results": [(r[1], round(r[2], 2)) for r in ocr_results[:20]],  # First 20 with confidence
    }


def main():
    print("=" * 60)
    print("ShareSansar Open-End Mutual Fund NAV Extractor")
    print("=" * 60)

    # Step 1: Fetch open-end funds
    print("\n[1/4] Fetching open-end mutual funds...")
    funds = fetch_open_end_funds()
    print(f"Found {len(funds)} open-end funds:")
    for f in funds:
        print(f"  - {f['symbol']}: {f['name']} (NAV: {f['daily_nav']})")

    # Save funds data
    DATA_DIR.mkdir(exist_ok=True)
    with open(DATA_DIR / "funds.json", "w") as fp:
        json.dump(funds, fp, indent=2)

    # Step 2: For each fund, get latest NAV announcement
    print("\n[2/4] Fetching NAV announcements...")
    nav_images = []

    for fund in funds:
        symbol = fund["symbol"]
        fund_name = fund["name"]
        print(f"  {symbol}...", end=" ")

        try:
            announcements = fetch_fund_announcements(symbol, fund_name)
            nav_ann = find_latest_nav_announcement(announcements)

            if nav_ann:
                print(f"found: {nav_ann['date']}")
                nav_images.append({
                    "symbol": symbol,
                    "name": fund["name"],
                    "announcement": nav_ann,
                })
            else:
                print("no NAV announcement found")
        except Exception as e:
            print(f"error: {e}")

        time.sleep(0.5)  # Be nice to the server

    # Step 3: Download images
    print("\n[3/4] Downloading NAV images...")
    downloaded = []

    for item in nav_images:
        symbol = item["symbol"]
        ann_url = item["announcement"]["url"]
        print(f"  {symbol}...", end=" ")

        try:
            img_url = extract_image_url(ann_url)
            if img_url:
                filepath = download_image(img_url, symbol)
                print(f"saved to {filepath.name}")
                downloaded.append({
                    "symbol": symbol,
                    "name": item["name"],
                    "image_path": str(filepath),
                    "image_url": img_url,
                    "announcement_url": ann_url,
                })
            else:
                print("no image found")
        except Exception as e:
            print(f"error: {e}")

        time.sleep(0.5)

    # Save download metadata
    with open(DATA_DIR / "images.json", "w") as fp:
        json.dump(downloaded, fp, indent=2)

    print(f"\nDownloaded {len(downloaded)} images to {IMAGES_DIR}/")

    # Step 4: Test OCR on first image
    if downloaded:
        print("\n[4/4] Testing EasyOCR on first image...")
        first = downloaded[0]
        print(f"  Processing {first['symbol']}...")

        try:
            ocr_results = run_easyocr(Path(first["image_path"]))
            nav_data = extract_nav_data(ocr_results)

            print(f"\n  OCR Results ({nav_data['line_count']} text blocks detected):")
            print("  " + "-" * 50)
            for text, confidence in nav_data["ocr_results"]:
                print(f"  [{confidence:.0%}] {text[:60]}")

            # Save OCR results
            with open(DATA_DIR / f"{first['symbol']}_ocr.json", "w") as fp:
                json.dump(nav_data, fp, indent=2, ensure_ascii=False)

        except Exception as e:
            print(f"  OCR error: {e}")
            print("  Make sure easyocr is installed: uv add easyocr")

    print("\n" + "=" * 60)
    print("Done! Check the data/ and images/ directories.")
    print("=" * 60)


if __name__ == "__main__":
    main()
