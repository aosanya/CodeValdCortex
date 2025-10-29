#!/bin/bash
# Download frontend assets for self-hosting
# CodeValdCortex must work in air-gapped environments

set -e

echo "================================================"
echo "Downloading Frontend Assets for CodeValdCortex"
echo "================================================"
echo ""

# Create static directories
echo "Creating static directories..."
mkdir -p static/{css,js,img}

# Download HTMX
echo "Downloading HTMX v1.9.10..."
curl -L https://unpkg.com/htmx.org@1.9.10/dist/htmx.min.js -o static/js/htmx.min.js
echo "✓ HTMX downloaded"

# Download Alpine.js
echo "Downloading Alpine.js v3.13.3..."
curl -L https://unpkg.com/alpinejs@3.13.3/dist/cdn.min.js -o static/js/alpine.min.js
echo "✓ Alpine.js downloaded"

# Download Chart.js
echo "Downloading Chart.js v4.4.1..."
curl -L https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.min.js -o static/js/chart.umd.min.js
curl -L https://cdn.jsdelivr.net/npm/chart.js@4.4.1/dist/chart.umd.js.map -o static/js/chart.umd.js.map
echo "✓ Chart.js and source map downloaded"

echo ""
echo "================================================"
echo "✓ All assets downloaded successfully"
echo "================================================"
echo ""
echo "Downloaded files:"
ls -lh static/js/

echo ""
echo "Next steps:"
echo "  1. Run: make assets-verify   (to verify all assets)"
