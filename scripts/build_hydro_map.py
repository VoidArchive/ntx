# /// script
# requires-python = ">=3.12"
# dependencies = ["pyshp>=2.3", "shapely>=2.0"]
# ///
"""Build apps/web/src/lib/data/hydro_map.json: Nepal's outline and rivers, projected once.

Inputs (downloaded by hand, never at build or runtime; ADR 0004):
  geoBoundaries NPL ADM0 GeoJSON  https://www.geoboundaries.org/api/current/gbOpen/NPL/ADM0/  (CC BY 4.0)
  HydroRIVERS v1.0 Asia shapefile https://data.hydrosheds.org/file/HydroRIVERS/HydroRIVERS_v10_as_shp.zip
  data/hydro_plants.csv

Reaches are clipped to the outline and kept when their Strahler order is at least
--order, or they lie within --radius-km of a plant so every plant's river is drawn.
Each reach's basin is found by walking downstream until the walk passes one of
the anchor points below (a well-known spot on each basin's main stem); a reach
whose walk meets no anchor is dropped unless a plant is near it, in which case it
takes the plant's basin. The outline and rivers are projected with one Mercator
whose parameters are stored in the file, so the page projects plant dots with the
same formula and no geo library.

Usage: uv run build_hydro_map.py --rivers HydroRIVERS_v10_as.shp --outline geoBoundaries-NPL-ADM0.geojson
"""

import argparse
import csv
import json
import math
import sys
from collections import Counter, defaultdict

import shapefile
from shapely.geometry import LineString, Point, shape
from shapely.strtree import STRtree

# One anchor per main stem, inside Nepal, named so the reader can check them on any map.
ANCHORS = {
    "Koshi": [(87.15, 26.86), (85.70, 27.63), (87.21, 27.31), (87.33, 26.93), (86.20, 27.03)],  # Chatara, Dolalghat, Tumlingtar, Mulghat, Sunkoshi at Harkapur
    "Gandaki": [(84.42, 27.70), (84.56, 27.86), (83.44, 27.95), (84.80, 27.85), (84.00, 28.10)],  # Narayanghat, Mugling, Ridi, Budhigandaki at Benighat, Seti at Damauli
    "Karnali": [(81.29, 28.64), (81.60, 28.85), (81.45, 28.95), (82.00, 28.75)],  # Chisapani, Bheri-Karnali confluence, Karnali above it, Bheri at Jajarkot
    "Mahakali": [(80.55, 29.85), (80.37, 29.63), (80.12, 29.08), (80.60, 29.60)],  # Darchula, Jhulaghat, Tanakpur, Chameliya mouth
    "Bagmati": [(85.29, 27.66), (85.47, 27.14), (85.37, 27.45)],  # Chobhar, Karmaiya, Bagmati at Pandheradobhan
    "Kamala": [(86.14, 26.95), (86.10, 26.85)],  # Kamala at Chisapani (Sindhuli), below it
    "West Rapti": [(82.78, 27.80), (82.55, 27.85), (83.45, 27.70), (82.20, 27.95)],  # Bhalubang, Lamahi, Tinau at Butwal, Rapti at Jalkundi
    "Babai": [(81.57, 28.30), (81.90, 28.35)],  # Babai in Bardiya, Babai at Chepang
    "Kankai": [(87.72, 26.63), (88.15, 26.65), (87.93, 26.85)],  # Kankai at Surunga, Mechi at Kakarbhitta, Mai at Ilam
}
ANCHOR_KM = 6.0

PROJECTION = {"lon0": 84.15, "lat0": 28.35, "width": 1000}


def merc(lat_deg: float) -> float:
    return math.log(math.tan(math.pi / 4 + math.radians(lat_deg) / 2))


def make_projection(outline_bounds):
    minx, miny, maxx, maxy = outline_bounds
    k = PROJECTION["width"] / math.radians(maxx - minx)
    height = k * (merc(maxy) - merc(miny))
    tx = -math.radians(minx - PROJECTION["lon0"]) * k
    ty = (merc(maxy) - merc(PROJECTION["lat0"])) * k
    return {"k": round(k, 3), "lon0": PROJECTION["lon0"], "lat0": PROJECTION["lat0"], "tx": round(tx, 2), "ty": round(ty, 2),
            "width": PROJECTION["width"], "height": round(height, 1)}


def project(p, lon, lat):
    x = math.radians(lon - p["lon0"]) * p["k"] + p["tx"]
    y = -(merc(lat) - merc(p["lat0"])) * p["k"] + p["ty"]
    return round(x, 1), round(y, 1)


def path(p, coords) -> str:
    pts = [project(p, lon, lat) for lon, lat in coords]
    out, last = [], None
    for i, (x, y) in enumerate(pts):
        if (x, y) == last:
            continue
        out.append(("M" if i == 0 else "L") + f"{x:g} {y:g}")
        last = (x, y)
    return "".join(out)


def km(a, b) -> float:
    return math.hypot((a[0] - b[0]) * 111 * math.cos(math.radians(a[1])), (a[1] - b[1]) * 111)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--rivers", required=True)
    ap.add_argument("--outline", required=True)
    ap.add_argument("--plants", default="../data/hydro_plants.csv")
    ap.add_argument("--out", default="../apps/web/src/lib/data/hydro_map.json")
    ap.add_argument("--order", type=int, default=4, help="minimum Strahler order drawn everywhere")
    ap.add_argument("--radius-km", type=float, default=3.0, help="lower-order reaches this close to a plant are drawn too")
    args = ap.parse_args()

    outline = shape(json.load(open(args.outline))["features"][0]["geometry"])
    plants = list(csv.DictReader(open(args.plants, encoding="utf-8")))
    plant_pts = [(float(r["lon"]), float(r["lat"])) for r in plants]
    radius_deg = args.radius_km / 111

    # Pass 1: every reach inside Nepal, by id, with its downstream link.
    sf = shapefile.Reader(args.rivers)
    fields = [f[0] for f in sf.fields[1:]]
    idx = {name: i for i, name in enumerate(fields)}
    minx, miny, maxx, maxy = outline.bounds
    reaches = {}
    for i, shp in enumerate(sf.iterShapes()):
        bx = shp.bbox
        if bx[2] < minx or bx[0] > maxx or bx[3] < miny or bx[1] > maxy:
            continue
        line = LineString(shp.points)
        if not line.intersects(outline):
            continue
        rec = sf.record(i)
        reaches[rec[idx["HYRIV_ID"]]] = (line, rec[idx["NEXT_DOWN"]], rec[idx["ORD_STRA"]])
    print(f"{len(reaches)} reaches touch Nepal", file=sys.stderr)

    # Basin by walking downstream to an anchor. Memoised per reach id.
    anchor_tree = STRtree([Point(a) for b in ANCHORS.values() for a in b])
    anchor_basin = [b for b, pts in ANCHORS.items() for _ in pts]
    basin_of = {}

    def walk(rid):
        seen = []
        while rid in reaches and rid not in basin_of and rid not in seen:
            seen.append(rid)
            line = reaches[rid][0]
            near = anchor_tree.query(line.buffer(ANCHOR_KM / 111))
            if len(near):
                basin_of[rid] = anchor_basin[int(near[0])]
                break
            rid = reaches[rid][1]
        found = basin_of.get(rid, "")
        for s in seen:
            basin_of.setdefault(s, found)

    for rid in reaches:
        walk(rid)

    # Pass 2: keep by order or by nearness to a plant, clip, project.
    plant_tree = STRtree([Point(p) for p in plant_pts])
    rivers, counts = [], Counter()
    kept_lines = []
    for rid, (line, _, order) in reaches.items():
        near = plant_tree.query(line.buffer(radius_deg))
        if order < args.order and not len(near):
            continue
        basin = basin_of.get(rid, "")
        if not basin and len(near):
            basin = plants[int(near[0])]["basin"]
        if not basin:
            continue
        clipped = line.intersection(outline).simplify(0.002)
        parts = [clipped] if clipped.geom_type == "LineString" else list(getattr(clipped, "geoms", []))
        for part in parts:
            if part.geom_type != "LineString" or part.is_empty:
                continue
            kept_lines.append(part)
            rivers.append({"basin": basin, "order": order, "coords": list(part.coords)})
            counts[basin] += 1

    # Check: every plant has a drawn reach within the radius, and that reach agrees with its basin.
    kept_tree = STRtree(kept_lines)
    kept_basin = [r["basin"] for r in rivers]
    missing, disagree = [], []
    for r, pt in zip(plants, plant_pts):
        near = kept_tree.query(Point(pt).buffer(radius_deg))
        if not len(near):
            missing.append(f"{r['symbol']} {r['plant']}")
        elif r["basin"] not in {kept_basin[int(i)] for i in near}:
            disagree.append(f"{r['symbol']} {r['plant']} ({r['basin']} vs {sorted({kept_basin[int(i)] for i in near})})")

    proj = make_projection(outline.bounds)
    out = {
        "projection": proj,
        "outline": path(proj, outline.simplify(0.003).exterior.coords),
        "rivers": [{"basin": r["basin"], "order": r["order"], "d": path(proj, r["coords"])} for r in rivers],
        "sources": {
            "outline": "geoBoundaries (CC BY 4.0), https://www.geoboundaries.org",
            "rivers": "HydroRIVERS v1.0, HydroSHEDS, https://www.hydrosheds.org",
        },
    }
    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(out, f, separators=(",", ":"))

    print(f"{len(rivers)} reaches drawn: {dict(counts)}", file=sys.stderr)
    print(f"plants without a drawn reach within {args.radius_km} km: {missing or 'none'}", file=sys.stderr)
    print(f"plants whose nearby reach has another basin: {disagree or 'none'}", file=sys.stderr)
    return 1 if missing else 0


if __name__ == "__main__":
    sys.exit(main())
