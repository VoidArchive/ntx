---
status: accepted
---
# The hydro map is a static SVG over a hand-curated plant table

The `/hydro` map is country outline and rivers clipped once from geoBoundaries and HydroRIVERS, projected once by a script, and committed as SVG path strings; the page renders paths and projects plant dots with the same stored projection. No tiles, no map service, no geo library in the browser, no district polygons (geoBoundaries only has the pre-federal 75). Plants come from `data/hydro_plants.csv`, drafted once from the DoED licence list and reviewed by hand, with a source URL per row; the worker upserts it into SQLite on boot and nothing scrapes at runtime. Tiles would add a third-party dependency and a network call to a page whose thesis is a database read, and scraping DoED daily would put unreviewed coordinates on the map. The cost is that a new plant is a CSV edit and a deploy.
