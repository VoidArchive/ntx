import cv2
import pytesseract
import argparse
import json
import requests
import io
import numpy as np
from PIL import Image
import os
import sys

# Ensure UTF-8 output for Nepali characters
sys.stdout.reconfigure(encoding='utf-8')

def get_image(source):
    """Load image from URL or local path."""
    if source.startswith('http'):
        response = requests.get(source)
        response.raise_for_status()
        image_array = np.asarray(bytearray(response.content), dtype=np.uint8)
        img = cv2.imdecode(image_array, cv2.IMREAD_COLOR)
        return img
    elif os.path.exists(source):
        return cv2.imread(source)
    else:
        raise FileNotFoundError(f"Source not found: {source}")

def preprocess_image(image):
    """Convert to grayscale, upscale, and apply thresholding."""
    print(f"Original Image Shape: {image.shape}", file=sys.stderr)
    
    # Upscale image by 2x to help Tesseract with small fonts
    image = cv2.resize(image, None, fx=2, fy=2, interpolation=cv2.INTER_CUBIC)
    
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    # Otsu's thresholding after Gaussian filtering
    blur = cv2.GaussianBlur(gray, (5, 5), 0)
    _, thresh = cv2.threshold(blur, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)

    return thresh

def extract_text(image):
    """Run Tesseract OCR on the image."""
    # psm 6: Assume a single uniform block of text.
    # lang='nep': Nepali language
    # psm 11: Sparse text. Find as much text as possible in no particular order.
    # psm 12: Sparse text with OSD.
    custom_config = r'--oem 3 --psm 11 -l nep'
    text = pytesseract.image_to_string(image, config=custom_config)
    return text

def parse_data(text):
    """
    Parse the extracted text into structured data.
    This function will need to be adapted based on the specific layout of the table.
    For now, it returns the raw text split by lines.
    """
    lines = [line.strip() for line in text.split('\n') if line.strip()]
    
    keywords = ["एनआइबिएल", "सहभागिता", "मूल्य", "मंसिर", "२०८२", "NAV"]
    found_keywords = [kw for kw in keywords if kw in text]

    return {
        "found_keywords": found_keywords,
        "raw_lines": lines
    }

def main():
    parser = argparse.ArgumentParser(description='Extract NAV data from image.')
    parser.add_argument('--image', required=True, help='Path or URL to the image')
    parser.add_argument('--debug', action='store_true', help='Save preprocessed image')
    args = parser.parse_args()

    try:
        img = get_image(args.image)
        processed_img = preprocess_image(img)
        
        if args.debug:
            cv2.imwrite('debug_processed.png', processed_img)
            
        text = extract_text(processed_img)
        data = parse_data(text)
        
        print(json.dumps(data, indent=2, ensure_ascii=False))
        
    except pytesseract.TesseractNotFoundError:
        print(json.dumps({
            "error": "Tesseract not found. Please install it.",
            "command": "sudo pacman -S tesseract tesseract-data-nep" # Arch Linux specific
        }), file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(json.dumps({"error": str(e)}), file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
