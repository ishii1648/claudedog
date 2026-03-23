#!/usr/bin/env bash
set -euo pipefail

OUTDIR="${1:-.outputs/grafana-screenshots}"
mkdir -p "$OUTDIR"

BASE="http://localhost:13000"
# 2026-03-01T00:00:00Z → 1772323200000 ms
# 2026-03-17T00:00:00Z → 1773705600000 ms
FROM=1772323200000
TO=1773705600000
WIDTH=1000
HEIGHT=500
TZ="Asia/Tokyo"

declare -a PANELS=(
  "1:headline-kpi"
  "9:weekly-trend"
  "12:task-type-perm-rate"
  "2:pr-scorecard"
  "10:tool-breakdown-table"
  "11:tool-breakdown-bar"
)

for entry in "${PANELS[@]}"; do
  ID="${entry%%:*}"
  NAME="${entry##*:}"
  OUTFILE="${OUTDIR}/panel-${ID}-${NAME}.png"
  URL="${BASE}/render/d-solo/hitl-metrics/hitl-metrics?panelId=${ID}&from=${FROM}&to=${TO}&width=${WIDTH}&height=${HEIGHT}&tz=${TZ}"
  echo "Capturing panel ${ID} (${NAME})..."
  curl -sf -o "$OUTFILE" "$URL"
  echo "  → ${OUTFILE}"
done

# Capture full dashboard for README
FULL="${OUTDIR}/dashboard-full.png"
echo "Capturing full dashboard..."
curl -sf -o "$FULL" "${BASE}/render/d/hitl-metrics/hitl-metrics?from=${FROM}&to=${TO}&width=1200&height=1600&tz=${TZ}&kiosk"
echo "  → ${FULL}"

# Also export key panels for README docs
DOCDIR="docs/images"
mkdir -p "$DOCDIR"
for pair in "1:headline-kpi:dashboard-headline" "9:weekly-trend:dashboard-weekly-trend" "2:pr-scorecard:dashboard-pr-scorecard"; do
  ID="${pair%%:*}"
  rest="${pair#*:}"
  PANEL_NAME="${rest%%:*}"
  DOC_NAME="${rest#*:}"
  cp "${OUTDIR}/panel-${ID}-${PANEL_NAME}.png" "${DOCDIR}/${DOC_NAME}.png"
done
cp "$FULL" "${DOCDIR}/dashboard-full.png"

echo "Done: ${#PANELS[@]} panels captured in ${OUTDIR}"
