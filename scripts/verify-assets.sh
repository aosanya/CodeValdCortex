#!/bin/bash
# Verify all required static assets are present

echo "Verifying static assets..."
echo ""

REQUIRED_ASSETS=(
    "static/css/tailwind.min.css"
    "static/js/htmx.min.js"
    "static/js/alpine.min.js"
    "static/js/chart.min.js"
    "static/js/alpine-components.js"
)

all_present=true

for asset in "${REQUIRED_ASSETS[@]}"; do
    if [ -f "$asset" ]; then
        size=$(du -h "$asset" | cut -f1)
        echo "✓ $asset ($size)"
    else
        echo "✗ $asset (MISSING)"
        all_present=false
    fi
done

echo ""

if [ "$all_present" = true ]; then
    echo "================================================"
    echo "✓ All required assets are present"
    echo "================================================"
    exit 0
else
    echo "================================================"
    echo "✗ Some assets are missing!"
    echo "================================================"
    echo ""
    echo "To fix:"
    echo "  1. Run: ./scripts/download-assets.sh"
    echo "  2. Run: make assets-build"
    exit 1
fi
