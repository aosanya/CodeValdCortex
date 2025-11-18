/**
 * Workflow Designer - Utils Module
 * Utility functions for geometry calculations and line intersection detection
 * No dependencies
 */

(function (window) {
    'use strict';

    // Initialize namespace if it doesn't exist
    if (!window.WorkflowDesigner) {
        window.WorkflowDesigner = {};
    }

    /**
     * Create utility functions
     * These are pure functions that don't require context
     */
    window.WorkflowDesigner.createUtils = function () {
        return {
            /**
             * Calculate distance from a point to a line segment
             * Returns Infinity if the point is not between the line segment endpoints
             * @param {number} px - Point X coordinate
             * @param {number} py - Point Y coordinate
             * @param {number} x1 - Line start X
             * @param {number} y1 - Line start Y
             * @param {number} x2 - Line end X
             * @param {number} y2 - Line end Y
             * @returns {number} Distance or Infinity
             */
            pointToLineDistance: function (px, py, x1, y1, x2, y2) {
                const A = px - x1;
                const B = py - y1;
                const C = x2 - x1;
                const D = y2 - y1;

                const dot = A * C + B * D;
                const lenSq = C * C + D * D;
                let param = -1;

                if (lenSq !== 0) {
                    param = dot / lenSq;
                }

                // If the point is outside the line segment bounds, return a large distance
                // This ensures edge splitting only happens when dropping ON the line segment
                if (param < 0 || param > 1) {
                    return Infinity; // Point is not between source and target
                }

                // Point is on the line segment, calculate perpendicular distance
                const xx = x1 + param * C;
                const yy = y1 + param * D;

                const dx = px - xx;
                const dy = py - yy;

                return Math.sqrt(dx * dx + dy * dy);
            },

            /**
             * Check if a line segment intersects with a rectangular box
             * @param {number} x1 - Line start X
             * @param {number} y1 - Line start Y
             * @param {number} x2 - Line end X
             * @param {number} y2 - Line end Y
             * @param {number} boxLeft - Box left edge
             * @param {number} boxTop - Box top edge
             * @param {number} boxRight - Box right edge
             * @param {number} boxBottom - Box bottom edge
             * @returns {boolean} True if line intersects box
             */
            lineIntersectsBox: function (x1, y1, x2, y2, boxLeft, boxTop, boxRight, boxBottom) {
                // Check if either endpoint is inside the box
                const p1Inside = (x1 >= boxLeft && x1 <= boxRight && y1 >= boxTop && y1 <= boxBottom);
                const p2Inside = (x2 >= boxLeft && x2 <= boxRight && y2 >= boxTop && y2 <= boxBottom);

                if (p1Inside || p2Inside) {
                    return true;
                }

                // Check if line intersects any of the four box edges
                const boxEdges = [
                    [boxLeft, boxTop, boxRight, boxTop],       // Top edge
                    [boxRight, boxTop, boxRight, boxBottom],   // Right edge
                    [boxRight, boxBottom, boxLeft, boxBottom], // Bottom edge
                    [boxLeft, boxBottom, boxLeft, boxTop]      // Left edge
                ];

                for (let i = 0; i < boxEdges.length; i++) {
                    const [bx1, by1, bx2, by2] = boxEdges[i];
                    if (this.lineSegmentsIntersect(x1, y1, x2, y2, bx1, by1, bx2, by2)) {
                        return true;
                    }
                }

                return false;
            },

            /**
             * Check if two line segments intersect
             * @param {number} x1 - First line start X
             * @param {number} y1 - First line start Y
             * @param {number} x2 - First line end X
             * @param {number} y2 - First line end Y
             * @param {number} x3 - Second line start X
             * @param {number} y3 - Second line start Y
             * @param {number} x4 - Second line end X
             * @param {number} y4 - Second line end Y
             * @returns {boolean} True if segments intersect
             */
            lineSegmentsIntersect: function (x1, y1, x2, y2, x3, y3, x4, y4) {
                const denom = ((y4 - y3) * (x2 - x1)) - ((x4 - x3) * (y2 - y1));
                if (denom === 0) return false; // Parallel lines

                const ua = (((x4 - x3) * (y1 - y3)) - ((y4 - y3) * (x1 - x3))) / denom;
                const ub = (((x2 - x1) * (y1 - y3)) - ((y2 - y1) * (x1 - x3))) / denom;

                return (ua >= 0 && ua <= 1 && ub >= 0 && ub <= 1);
            }
        };
    };

})(window);
