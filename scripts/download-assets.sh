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

# Download Bulma CSS
echo "Downloading Bulma CSS v1.0.2..."
curl -L https://cdn.jsdelivr.net/npm/bulma@1.0.2/css/bulma.min.css -o static/css/bulma.min.css
echo "✓ Bulma CSS downloaded"

# Download HTMX
echo "Downloading HTMX v1.9.10..."
curl -L https://unpkg.com/htmx.org@1.9.10/dist/htmx.min.js -o static/js/htmx.min.js
echo "✓ HTMX downloaded"

# Download Alpine.js
echo "Downloading Alpine.js v3.13.3..."
curl -L https://unpkg.com/alpinejs@3.13.3/dist/cdn.min.js -o static/js/alpine.min.js
echo "✓ Alpine.js downloaded"

# Download Alpine.js Collapse Plugin
echo "Downloading Alpine.js Collapse Plugin v3.13.3..."
curl -L https://unpkg.com/@alpinejs/collapse@3.13.3/dist/cdn.min.js -o static/js/alpine-collapse.min.js
echo "✓ Alpine.js Collapse plugin downloaded"

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
