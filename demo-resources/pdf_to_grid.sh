#!/usr/bin/env bash
# pdf_to_grid.sh – Combine a 12-page PDF into one 3×4 PNG sheet.
set -euo pipefail

# --- Configuration ---------------------------------------------------------
PDF="${1:-main.pdf}"        # Input PDF (default main.pdf)
OUT="${2:-main_grid.png}"   # Output image (default main_grid.png)
RES="${3:-300}"             # Resolution DPI (default 300)
TMPDIR="$(mktemp -d)"

cleanup() { rm -rf "$TMPDIR"; }
trap cleanup EXIT
# ---------------------------------------------------------------------------

echo "Rendering pages to temporary PNGs..."
# Create blank page matching PDF size
WIDTH=$(pdfinfo "$PDF" | awk '/Page size/ {print int($3)}' | head -n1)
HEIGHT=$(pdfinfo "$PDF" | awk '/Page size/ {print int($5)}' | head -n1)
PX_WIDTH=$((WIDTH * RES / 72))
PX_HEIGHT=$((HEIGHT * RES / 72))
magick -size ${PX_WIDTH}x${PX_HEIGHT} canvas:white "$TMPDIR/page-00.png"
pdftoppm -png -rx "$RES" -ry "$RES" "$PDF" "$TMPDIR/page"

echo "Building 4-column montage with as many rows as needed..."
# The pages are named page-01.png … page-12.png
magick montage "$TMPDIR"/page-*.png \
       -tile 4x -geometry +0+0 -border 1 -bordercolor black "$OUT"

echo "Done → $OUT"
