# Vendor JavaScript Libraries

This directory contains third-party JavaScript libraries for the Network Topology Visualizer (INFRA-017).

## Libraries

- **deck.gl** (v9.x) - WebGL-powered framework for visual exploratory data analysis
  - File: `deck.gl.min.js`
  - License: MIT
  - Source: https://deck.gl

- **d3-hierarchy** (v3.x) - Hierarchical layouts for D3
  - File: `d3-hierarchy.min.js`
  - License: ISC
  - Source: https://github.com/d3/d3-hierarchy

- **maplibre-gl** (v3.x) - Open-source interactive maps (no API token required!)
  - File: `maplibre-gl.js`
  - License: BSD-3-Clause
  - Source: https://maplibre.org
  - Note: Free and open-source alternative to Mapbox

- **d3** (v7.x) - Data manipulation and utilities
  - File: `d3.min.js`
  - License: ISC
  - Source: https://d3js.org

## Usage

Include these libraries in your HTML before using the topology visualizer:

```html
<!-- CSS -->
<link rel="stylesheet" href="/static/css/maplibre-gl.css">

<!-- JavaScript Libraries -->
<script src="/static/js/vendor/maplibre-gl.js"></script>
<script src="/static/js/vendor/deck.gl.min.js"></script>
<script src="/static/js/vendor/d3.min.js"></script>
<script src="/static/js/vendor/d3-hierarchy.min.js"></script>
```

