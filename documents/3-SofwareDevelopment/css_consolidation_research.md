# SCSS Consolidation Research & Strategy

**Date**: November 28, 2025  
**Current Branch**: feature/MVP-WI-012_workbench_chat_panel  
**Status**: 🎉 **SCSS MIGRATION 100% COMPLETE** - All 11 custom CSS files migrated to modular SCSS!

---

## 🎯 Current Architecture (SCSS-First)

### SCSS Source Structure
```
static/scss/
├── _variables.scss                # Shared variables (colors, spacing, fonts, transitions)
├── _mixins.scss                  # Reusable mixins (flex, panels, scrollbars, hover states)
│
├── agency-designer.scss          # Main entry point for agency designer
├── workflow-designer.scss        # Main entry point for workflow designer ✅
├── vscode-designer-shared.scss   # Shared VS Code layout ✅
├── styles.scss                   # Dashboard custom styles ✅
├── themes.scss                   # Theme system with SCSS maps ✅
├── agencies.scss                 # Agency homepage and cards ✅
├── common-layout.scss            # Navbar and status bar ✅
├── common-animations.scss        # Consolidated animations ✅
├── ai-policy-wizard.scss         # AI policy wizard ✅
├── vscode-status-bar.scss        # Status bar component ✅
├── raci-matrix.scss              # RACI matrix editor ✅
│
├── agency-designer/              # Modular SCSS files (6 modules)
│   ├── _layout.scss              # Designer panels, view switching, sidebars
│   ├── _overview.scss            # Overview sections, navigation, cards, tips
│   ├── _details.scss             # Agent details, properties, relationships
│   ├── _chat.scss                # Chat panel, messages, typing indicators
│   ├── _context.scss             # Context selection system, menus, badges
│   └── _forms.scss               # Form overrides, publish toolbar, dialogs
│
├── workflow-designer/            # Modular SCSS files (6 modules) ✅
│   ├── _layout.scss              # Container, work items panel
│   ├── _toolbar.scss             # Top toolbar controls
│   ├── _canvas.scss              # Canvas, grid, markers, flow connectors
│   ├── _steps.scss               # Workflow steps, parallel execution
│   ├── _dropzones.scss           # Drop zone styling
│   └── _animations.scss          # Slide animations
│
├── vscode-designer-shared/       # Modular SCSS files (6 modules) ✅
│   ├── _layout.scss              # Container, navbar overrides, grid layout
│   ├── _sidebar.scss             # Sidebar panel structure
│   ├── _details.scss             # Details panel
│   ├── _chat.scss                # Chat panel, messages, typing, context
│   ├── _scrollbar.scss           # Custom scrollbar styling
│   └── _responsive.scss          # Mobile/tablet responsive design
│
└── styles/                       # Modular SCSS files (9 modules) ✅
    ├── _utilities.scss           # Icon sizes, truncate, blur-load, focus
    ├── _htmx.scss                # HTMX loading, progress bar, states
    ├── _scrollbar.scss           # Custom scrollbar with dark mode
    ├── _animations.scss          # Spin, shimmer, pulse, slide animations
    ├── _components.scss          # Status badge, card hover, buttons, toasts
    ├── _logs.scss                # Log viewer with error/warn/info levels
    ├── _charts.scss              # Chart container
    ├── _status.scss              # Agent status color classes
    └── _responsive.scss          # Mobile/tablet responsive, print styles
```

### Build Process
- **Source**: `static/scss/*.scss` (tracked in git)
- **Compilation**: `make css` (or `make build`, `make run` - auto-runs)
- **Output**: `static/css/*.css` (auto-generated, gitignored)
- **Compiler**: Sass via npm (`npx sass`)

### CSS Output (Auto-Generated)
```
static/css/
├── agency-designer.css           # 🤖 Compiled from SCSS (18K minified)
├── workflow-designer.css         # 🤖 Compiled from SCSS (12K minified) ✅
├── vscode-designer-shared.css    # 🤖 Compiled from SCSS (11K minified) ✅
├── styles.css                    # 🤖 Compiled from SCSS (8K minified) ✅
├── themes.css                    # 🤖 Compiled from SCSS (9K minified) ✅
├── agencies.css                  # 🤖 Compiled from SCSS (5K minified) ✅
├── common-layout.css             # 🤖 Compiled from SCSS (3K minified) ✅
├── common-animations.css         # 🤖 Compiled from SCSS (4K minified) ✅
├── ai-policy-wizard.css          # 🤖 Compiled from SCSS (4K minified) ✅
├── vscode-status-bar.css         # 🤖 Compiled from SCSS (2K minified) ✅
├── raci-matrix.css               # 🤖 Compiled from SCSS (1K minified) ✅
│
└── [3rd-party CSS files]         # bulma.min.css, mapbox-gl.css, etc. (unchanged)
    ├── bulma.min.css
    ├── mapbox-gl.css
    └── maplibre-gl.css
```

**Note**: All compiled CSS files are gitignored (except 3rd-party). SCSS source is the single source of truth.

---

## ✅ Completed Work (November 28, 2025)

### 1. Fixed Navbar Visibility Issue
- **Problem**: Navbar not visible on agency designer pages
- **Root Cause**: `@NavbarWithAgency()` was rendered inside `<main class="section">` container
- **Solution**: Moved navbar outside main section to body top level in `layout_with_agency.templ`
- **Files Changed**: `internal/web/components/layout_with_agency.templ`
- **Status**: ✅ Fixed and committed

### 2. Created Shared Layout Components (DRY Principle)
- **Created**: `internal/web/components/head_includes.templ`
  - `HeadIncludes()` - Centralizes all CSS/JS imports (Bulma, FontAwesome, Alpine, HTMX, Chart.js, etc.)
  - `NavbarBurgerScript()` - Shared navbar toggle script
- **Updated**: Both `layout.templ` and `layout_with_agency.templ` now use shared components
- **Benefit**: Single source of truth for CSS/JS includes, easier maintenance
- **Status**: ✅ Implemented and committed

### 3. Fixed Code Quality Issues
- **Linting**: Fixed QF1003 warning in `chat_panel.templ` by converting if-else chain to switch statement
- **Missing File**: Created `static/js/workbench-chat.js` for workbench chat functionality
- **Status**: ✅ Fixed and committed

### 4. Git Commit
```
Commit: 15c5ffd - feat(ui): refactor layout components and add workbench chat
Files: 16 changed, 1315 insertions(+), 401 deletions(-)

Commit: ab6d07d - docs: update CSS consolidation research with completed work  
Files: 1 changed, 70 insertions(+), 3 deletions(-)

Commit: 847c77b - refactor(css): apply CSS consolidation recommendations
Files: 6 changed, 222 insertions(+), 6 deletions(-)

Commit: 08c8836 - refactor(css): migrate agency-designer to modular SCSS architecture
Files: 15 changed, 2223 insertions(+), 1718 deletions(-)

Commit: 097010f - refactor(css): migrate workflow-designer to modular SCSS architecture
Files: 9 changed, 792 insertions(+), 466 deletions(-)

Commit: 59e2f23 - refactor(css): migrate agencies, common-layout, and common-animations to SCSS
Files: 6 changed, 547 insertions(+), 570 deletions(-)

Commit: cc99e74 - docs(css): update SCSS migration status and remaining work
Files: 1 changed, 99 insertions(+), 38 deletions(-)

Commit: c2a8b7d - refactor(css): migrate vscode-designer-shared to modular SCSS architecture
Files: 8 changed, 483 insertions(+), 486 deletions(-)

Commit: 358f26b - refactor(css): migrate styles.css to modular SCSS architecture
Files: 11 changed, 379 insertions(+), 360 deletions(-)

Commit: 7487a52 - refactor(css): migrate themes.css to SCSS with map-based theme system
Files: 2 changed, 380 insertions(+), 362 deletions(-)

Commit: 48e9f89 - refactor(css): migrate remaining small CSS files to SCSS
Files: 6 changed, 339 insertions(+), 294 deletions(-)
```

**Total Files Changed**: 65+ files  
**Total Lines Changed**: ~3,500+ insertions, ~3,300+ deletions  
**SCSS Modules Created**: 40+ modular SCSS files

### 5. Applied CSS Consolidation Recommendations
- **Priority 0 (Critical)**: ✅ Removed navbar negative margin bug
  - Removed CSS `.navbar-brand { margin-left: -5rem; }` from `common-layout.css`
  - Removed inline `style="margin-left: -5rem;"` from `navbar_with_agency.templ`
  - Navbar now displays correctly and doesn't overflow off-screen
  
- **Priority 2 (Consolidation)**: ✅ Created `common-animations.css`
  - Consolidated all animations from `styles.css`, `agencies.css`, `agency-designer.css`, `workflow-designer.css`
  - Standardized animations: spin, fadeIn, fadeInUp, slideIn/Out, pulse, shimmer, typing
  - Added utility classes for easy animation application
  - Included in `HeadIncludes` component for global availability
  
- **Status**: ✅ Implemented and committed

### 6. SCSS Migration - 100% COMPLETE! 🎉 (November 28, 2025)

**All 11 Custom CSS Files Migrated to SCSS**:

✅ **agency-designer.css → SCSS** (Commit: 08c8836)
  - Split into 6 modules: _layout, _overview, _details, _chat, _context, _forms
  - All modules under 280 lines each
  - Shared _variables.scss and _mixins.scss created
  - Applied SCSS nesting and design tokens

✅ **workflow-designer.css → SCSS** (Commit: 097010f)
  - Split into 6 modules: _layout, _toolbar, _canvas, _steps, _dropzones, _animations
  - All modules under 150 lines each
  - Uses shared variables for colors, spacing, transitions
  - Applied SCSS nesting for cleaner code
  
✅ **agencies.css → SCSS** (Commit: 59e2f23)
  - Single file SCSS (~240 lines)
  - Staggered animation using SCSS @for loop
  - Converted hardcoded colors/spacing to design tokens
  - Card patterns and filters with proper nesting
  
✅ **common-layout.css → SCSS** (Commit: 59e2f23)
  - Navbar and status bar shared styles (~120 lines)
  - Global font sizing and spacing
  - Theme switcher dropdown styles
  
✅ **common-animations.css → SCSS** (Commit: 59e2f23)
  - Consolidated animations library (~250 lines)
  - All animations from multiple CSS files centralized
  - Utility classes for easy application
  
✅ **vscode-designer-shared.css → SCSS** (Commit: c2a8b7d)
  - Split into 6 modules: _layout, _sidebar, _details, _chat, _scrollbar, _responsive
  - CSS custom properties for VS Code theme variables
  - All modules under 240 lines each
  - Shared layout used by agency-designer and workbench

✅ **styles.css → SCSS** (Commit: 358f26b)
  - Split into 9 modules: _utilities, _htmx, _scrollbar, _animations, _components, _logs, _charts, _status, _responsive
  - All modules under 80 lines each
  - Dashboard custom styles properly organized by domain
  - Clean separation of concerns

✅ **themes.css → SCSS** (Commit: 7487a52)
  - SCSS map-based theme system (~360 lines)
  - 7 themes defined in central $themes map
  - Automatic CSS custom property generation
  - Single source of truth for all theme colors
  - Easy to add/modify themes

✅ **ai-policy-wizard.css → SCSS** (Commit: 48e9f89)
  - Single file SCSS (~190 lines)
  - Steps progress indicator with nested states
  - Form validation and button hover effects
  - Responsive design for mobile

✅ **vscode-status-bar.css → SCSS** (Commit: 48e9f89)
  - Single file SCSS (~75 lines)
  - Status bar notifications with color variants
  - Page layout grid helper
  - Status action buttons

✅ **raci-matrix.css → SCSS** (Commit: 48e9f89)
  - Single file SCSS (~55 lines)
  - RACI matrix table styling
  - RACI selector buttons with active states
  - Modal card sizing
  - All animations consolidated into single file
  - Spin, fade, slide, pulse, typing, shimmer, context menu effects
  - Utility classes for easy application

**Migration Statistics**:
- **Files migrated**: 5 CSS files → 5 SCSS entry points + 12 module files
- **Total SCSS**: ~1,835 lines (modular, organized)
- **Total compiled CSS**: ~42K minified
- **Percentage complete**: ~65% of custom CSS (by line count)

**Remaining Files** (deferred for separate tasks):
- `vscode-designer-shared.css` (485 lines) - Complex shared layout, needs careful migration
- `styles.css` (300+ lines) - Dashboard styles, needs analysis and splitting
- `themes.css` (361 lines) - Complex theme system, requires SCSS maps approach

---

## � Remaining Work (Prioritized)

### Priority 0: Critical Issues Still Open
- [x] **Remove negative margin from navbar-brand** ✅ COMPLETED (Commit: 847c77b)
  - ~~Remove inline `style="margin-left: -5rem;"` from `navbar_with_agency.templ`~~
  - ~~Remove CSS `.navbar-brand { margin-left: -5rem; }` from `common-layout.css`~~
  - ~~Test responsive behavior on mobile~~
  
### Priority 1: File Size Violations
- [x] **Extract shared VS Code designer styles** ✅ COMPLETED (Commit: 666057d)
  - Created `vscode-designer-shared.css` (485 lines) with base layout, panels, chat
  - Reduced `agency-designer.css` from 1,812 → 1,706 lines
  - Shared between agency-designer AND workbench pages
  - Added to `head_includes.templ` for global availability

- [x] **Migrate agency-designer to SCSS** ✅ COMPLETED (Commit: 08c8836)
  - **Architecture**: Created modular SCSS structure in `static/scss/`
  - **Modules**: Split into 6 domain-focused files (all under 300 lines)
  - **Shared System**: _variables.scss (60 lines), _mixins.scss (80 lines)
  - **Build Integration**: make css compiles SCSS to minified CSS
  - **Benefits**: All modules under 400 lines, design system, reusable mixins

- [x] **Migrate workflow-designer.css to SCSS** ✅ COMPLETED (Commit: 097010f)
  - Split into 6 modules: _layout, _toolbar, _canvas, _steps, _dropzones, _animations
  - All modules under 150 lines each
  - Uses shared _variables.scss and _mixins.scss
  - Follows agency-designer pattern
  
- [x] **Migrate agencies.css to SCSS** ✅ COMPLETED (Commit: 59e2f23)
  - Single file SCSS (~200 lines)
  - Staggered animation using SCSS @for loop
  - Converted to design tokens
  
- [x] **Migrate common-layout.css to SCSS** ✅ COMPLETED (Commit: 59e2f23)
  - Navbar and status bar styles (~110 lines)
  - Uses shared variables
  
- [x] **Migrate common-animations.css to SCSS** ✅ COMPLETED (Commit: 59e2f23)
  - All animations consolidated (~200 lines)
  - Utility classes for easy application

- [ ] **Migrate remaining CSS files to SCSS** 🔄 NEXT PRIORITY
  - `vscode-designer-shared.css` (485 lines) - Deferred (complex shared layout)
  - `styles.css` (300+ lines) - Deferred (needs analysis)
  - `themes.css` (361 lines) - Deferred (needs SCSS maps approach)
  
### Priority 2: Consolidation & Migration
- [x] **Create `common-animations.css`** ✅ COMPLETED (Commit: 847c77b)
  - ~~Move all `@keyframes` from `styles.css`, `agencies.css`, `agency-designer.css`, `workflow-designer.css`~~
  - ~~Standardize animation names (remove duplicates: spin, fade, slide, pulse)~~
  - Note: Original definitions kept in source files for backward compatibility, can be removed in future cleanup

- [ ] **Migrate `workflow-designer.css` to SCSS** 🔄 HIGH PRIORITY
  - File size: 442 lines - good candidate for modularization
  - Create `static/scss/workflow-designer.scss` with modules:
    - `_layout.scss` - Canvas, panels, drag-drop zones
    - `_nodes.scss` - Work items, step items, connections
    - `_toolbar.scss` - Tools, buttons, controls
    - `_sidebar.scss` - Property panels, forms
  - Use shared `_variables.scss` and `_mixins.scss`
  - Follow agency-designer pattern

- [ ] **Migrate `styles.css` to SCSS** 🔄 MEDIUM PRIORITY
  - File size: 1,044 lines (previously - needs current check)
  - Split by domain: dashboard, charts, tables, forms, utilities
  - Extract shared patterns to mixins
  - Consolidate with other dashboard styles

- [ ] **Migrate `agencies.css` to SCSS** 🔄 MEDIUM PRIORITY
  - File size: 233 lines - straightforward migration
  - Create `static/scss/agencies.scss`
  - Extract card patterns to shared mixins (reusable across modules)

- [ ] **Convert `vscode-designer-shared.css` to SCSS partial** 🔄 MEDIUM PRIORITY
  - Rename to `static/scss/_vscode-designer-shared.scss`
  - Use variables for VS Code colors, spacing
  - Import in both agency-designer and workbench SCSS
  
- [ ] **Extract common component styles to SCSS modules**
  - Create `static/scss/_components.scss` with:
    - `@mixin button-variant` for consistent button styles
    - `@mixin card-base` for card patterns
    - `@mixin modal-base` for modal dialogs
  - Remove duplicates across files
  
### Priority 3: Theme System & SCSS Enhancement
- [ ] **Migrate `themes.css` to SCSS with CSS custom properties**
  - Create `static/scss/_themes.scss`
  - Use SCSS maps for theme definitions
  - Generate CSS custom properties from SCSS
  - Example:
    ```scss
    $themes: (
      "light": (bg: #ffffff, text: #1a1a1a, ...),
      "dark": (bg: #1e1e1e, text: #d4d4d4, ...)
    );
    ```

- [ ] **Reduce `!important` usage**
  - Refactor theme overrides to use proper CSS specificity
  - Use SCSS nesting to increase specificity naturally
  
- [ ] **Evaluate if all 6 themes are needed**
  - Analyze theme usage metrics (if available)
  - Consider consolidating similar themes

### Priority 4: Modern SCSS Practices
- [ ] **Migrate from `@import` to `@use` syntax** 🔄 RECOMMENDED
  - Current: Using deprecated `@import`
  - Future: Use `@use` and `@forward` for better namespacing
  - Reference: [Sass @use documentation](https://sass-lang.com/documentation/at-rules/use)
  - Benefits: Namespace protection, explicit dependencies

- [ ] **Set up Stylelint for SCSS** 🔄 RECOMMENDED
  - Install: `stylelint`, `stylelint-config-standard-scss`
  - Configure `.stylelintrc.json` with SCSS rules
  - Add `make lint-css` to Makefile
  - Integrate with CI/CD

- [ ] **Add SCSS documentation** 📝
  - Create `docs/scss-architecture.md`
  - Document naming conventions
  - Provide examples of using variables/mixins
  - Developer onboarding guide

### Priority 5: Optimization (Future)
- [ ] Implement CSS purging for production builds (PurgeCSS)
- [ ] Add bundle size monitoring (bundlewatch)
- [ ] Consider CSS-in-JS alternatives for dynamic components

---

## �🔍 Current State Analysis

### CSS Files Inventory (by size)

| File | Lines | Purpose | Status |
|------|-------|---------|--------|
| **SCSS Source Files** | | | |
| `static/scss/_variables.scss` | ~60 | Shared variables (colors, spacing, fonts) | ✅ **NEW** |
| `static/scss/_mixins.scss` | ~80 | Reusable mixins (flex, panels, scrollbars) | ✅ **NEW** |
| `static/scss/agency-designer.scss` | ~15 | Main entry point (imports modules) | ✅ **NEW** |
| `static/scss/agency-designer/_layout.scss` | ~280 | Layout, panels, views, sidebars | ✅ **NEW** |
| `static/scss/agency-designer/_overview.scss` | ~220 | Overview sections, cards, tips | ✅ **NEW** |
| `static/scss/agency-designer/_details.scss` | ~200 | Agent details, properties, relationships | ✅ **NEW** |
| `static/scss/agency-designer/_chat.scss` | ~260 | Chat panel, messages, typing | ✅ **NEW** |
| `static/scss/agency-designer/_context.scss` | ~150 | Context selection, menus, badges | ✅ **NEW** |
| `static/scss/agency-designer/_forms.scss` | ~140 | Form overrides, publish toolbar | ✅ **NEW** |
| **Compiled CSS Files** | | | |
| `static/css/agency-designer.css` | 1 line | **AUTO-GENERATED** from SCSS | 🤖 Minified (18K) |
| `static/css/vscode-designer-shared.css` | 485 | Shared VS Code layout (agency + workbench) | ✅ **SHARED** |
| `static/css/workflow-designer.css` | 442 | Workflow drag-drop designer | 🔴 Medium complexity |
| `themes.css` | 361 | Theme system (6 themes) | ✅ Well-organized |
| `styles.css` | 359 | "Custom styles for dashboard" | 🔴 Mixed concerns |
| `agencies.css` | 233 | Agency homepage/cards | ✅ Focused |
| `ai-policy-wizard.css` | 172 | AI policy wizard specific | ✅ Focused |
| `common-layout.css` | 118 | Navbar & status bar | ⚠️ **Key file** |
| `vscode-status-bar.css` | 67 | Status bar component | ✅ Focused |
| `raci-matrix.css` | 52 | RACI matrix specific | ✅ Focused |
| **3rd-party** | | | |
| `bulma.min.css` | Large | Framework (minified) | ✅ Don't touch |
| `mapbox-gl.css` | Large | Map library | ✅ Don't touch |
| `maplibre-gl.css` | Large | Map library | ✅ Don't touch |

**Total custom CSS**: ~3,596 lines (excluding 3rd-party)

---

## 🚨 Critical Issues Identified

### 1. **Duplicate Animation Definitions** (PRIORITY 1)

**Problem**: Same `@keyframes` defined in multiple files

| Animation | Files | Issue |
|-----------|-------|-------|
| `spin` | `styles.css`, `agencies.css` | Duplicated - should be in shared file |
| `fadeIn` | `agency-designer.css` (commented out!) | Disabled but not removed |
| `pulse` | `styles.css` (pulse-green), `agency-designer.css` (pulse) | Different names, similar purpose |
| `slideIn/slideOut` | `workflow-designer.css`, `styles.css` (slide-in/slide-out) | Same concept, different naming |

**Impact**: Increases CSS bundle size, inconsistent animation behavior

---

### 2. **Navbar Visibility Issue** (PRIORITY 0 - BLOCKING)

**Root Cause Analysis**:

```css
/* common-layout.css - Line 22-24 */
.navbar-brand {
    margin-left: -5rem;  /* ⚠️ NEGATIVE MARGIN PUSHES NAVBAR OFF-SCREEN */
}
```

**Also in themes.css**:
```css
/* themes.css - Multiple navbar color overrides */
.navbar {
    background-color: var(--theme-navbar-bg) !important;
    border-bottom: 1px solid var(--theme-border);
    box-shadow: 0 2px 4px var(--theme-shadow);
}

.navbar-item, .navbar-link {
    color: var(--theme-navbar-text) !important;  /* May conflict with Bulma */
}
```

**Questions**:
1. **Why the -5rem margin?** Is this intentional for some layout requirement?
2. **Is the navbar hidden on all pages or specific ones?**
3. **Are themes overriding navbar visibility unintentionally?**

---

### 3. **Duplicate Component Styling** (PRIORITY 1)

**Button styles** appear in:
- `styles.css`: `.btn-loading`, HTMX button states
- `agency-designer.css`: `.button.is-light`, `.button.is-pulsing`
- `workflow-designer.css`: `.button.is-ghost`
- `themes.css`: `.button.is-primary`, `.button.is-light` (global theme overrides)

**Card styles** appear in:
- `agencies.css`: `.agency-card`
- `workflow-designer.css`: `.work-item-card .card`, `.step-item .card`
- `agency-designer.css`: `.step-card`, `.requirement-card`, etc.
- `styles.css`: `.card-hover`

**Modal styles** appear in:
- `raci-matrix.css`: `.modal-card`
- `agency-designer.css`: `#publish-dialog .modal-card`, `#tag-dialog .modal-card`
- `themes.css`: `.modal-card` (global theme overrides)

**Impact**: Maintenance nightmare - changing one thing requires hunting across multiple files

---

### 4. **Conflicting Global Font Sizes** (PRIORITY 2)

```css
/* common-layout.css */
html, body {
    font-size: 0.9rem;  /* Global smaller font */
}

.navbar-item {
    font-size: 0.95rem;  /* Navbar specific */
}

/* agency-designer.css - Lines 1675-1700 */
.label {
    font-size: 0.75rem !important;  /* Form labels */
}

.help {
    font-size: 0.65rem !important;  /* Help text */
}

.input, .textarea, .select select {
    font-size: 0.8rem !important;  /* Form inputs */
}
```

**Questions**:
1. **Why multiple different font sizes?** Is this intentional responsive design or accumulated cruft?
2. **Should we use a CSS variable system for typography scale?**

---

### 5. **Theme System Complexity** (PRIORITY 2)

**Current structure**:
- 6 different themes defined in `themes.css`
- Each theme has 15+ CSS variables
- Heavy use of `!important` to override Bulma
- Themes override navbar, buttons, cards, dropdowns, notifications

**Questions**:
1. **Are all 6 themes actually used?** Or can we consolidate?
2. **Can we reduce `!important` usage?** This makes debugging harder
3. **Should themes be split into separate files** (one per theme) for better maintainability?

---

### 6. **VS Code-Style Designer CSS is MASSIVE** (PRIORITY 1)

**`agency-designer.css` at 1,792 lines violates our 500-700 line rule!**

**Breakdown**:
- Lines 1-200: VS Code layout variables and grid system
- Lines 201-500: Sidebar, navbar, panels
- Lines 501-800: Chat panel (messages, typing indicators)
- Lines 801-1200: Overview sections, navigation
- Lines 1201-1600: Context selection system
- Lines 1601-1792: Form overrides, publish toolbar

**Action Required**: **MUST SPLIT** into:
1. `agency-designer-layout.css` - Grid, panels, main structure
2. `agency-designer-chat.css` - Chat panel specific
3. `agency-designer-context.css` - Context selection system
4. `agency-designer-forms.css` - Form element overrides

---

## 📊 CSS Loading Strategy Analysis

### Current CSS Load Order (from `layout_with_agency.templ`):

```html
1. bulma.min.css          (Bulma framework)
2. fontawesome/all.min.css (Icons)
3. common-layout.css       (Navbar + status bar)
4. styles.css              (Mixed utilities)
5. agencies.css            (Agency-specific)
6. themes.css              (Theme system - LAST!)
```

**Problem**: Theme CSS loads last, so it uses `!important` everywhere to override earlier styles.

**Better Strategy**:
```
1. bulma.min.css           (Framework base)
2. theme-variables.css     (CSS variables ONLY - no selectors)
3. common-layout.css       (Shared layout)
4. common-components.css   (Shared buttons, cards, modals)
5. common-animations.css   (Shared keyframes)
6. [page-specific].css     (Only when needed)
```

---

## 🎯 Strategic Questions for Decision Making

### Architecture & Organization

**Q1**: Should we adopt a **component-based CSS structure**?
```
static/css/
├── base/
│   ├── reset.css
│   ├── typography.css
│   └── variables.css
├── components/
│   ├── buttons.css
│   ├── cards.css
│   ├── modals.css
│   └── navbar.css
├── layouts/
│   ├── agency-designer.css
│   ├── workflow-designer.css
│   └── dashboard.css
├── themes/
│   ├── theme-variables.css
│   ├── light.css
│   └── dark.css
└── utilities/
    ├── animations.css
    └── helpers.css
```

**Q2**: Should we use **CSS custom properties more extensively**?
```css
/* Instead of hardcoded values everywhere */
:root {
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --spacing-md: 1rem;
  --spacing-lg: 2rem;
  
  --font-size-xs: 0.75rem;
  --font-size-sm: 0.85rem;
  --font-size-base: 1rem;
  
  --animation-speed-fast: 150ms;
  --animation-speed-normal: 300ms;
}
```

**Q3**: Should we create a **single animations file**?
- Consolidate all `@keyframes` into `common-animations.css`
- Remove duplicates (spin, fade, slide, pulse)
- Standardize animation naming convention

---

### Navbar Visibility Issue

**Q4**: What is the **intended purpose** of `.navbar-brand { margin-left: -5rem; }`?
- Is this for a specific layout requirement?
- Is the logo supposed to overflow the container?
- Can we achieve the same effect with flexbox/grid?

**Q5**: Are there **specific pages where the navbar should be hidden**?
- Agency designer (VS Code style)?
- Workflow designer?
- Or should it always be visible?

**Q6**: Should we have **different navbar variants**?
- Default navbar (full width)
- Compact navbar (for designer views)
- Hidden navbar (full-screen modes)

---

### Theme System

**Q7**: How many themes are **actively used**?
- Are all 6 themes (midnight-coral, slate-purple, charcoal-emerald, navy-orange, obsidian-cyan, dark) necessary?
- Can we consolidate to 2-3 themes (light, dark, custom)?

**Q8**: Should we **reduce `!important` usage**?
- Current: Heavy `!important` in themes.css to override Bulma
- Better: Use CSS specificity properly
- Or: Create Bulma SASS customization file

**Q9**: Should theme switching be **runtime** or **build-time**?
- Runtime: CSS variables (current approach)
- Build-time: Separate CSS bundles per theme

---

### File Organization

**Q10**: Should we **split large files immediately**?
- `agency-designer.css` (1,792 lines) - Split into 4 files?
- `workflow-designer.css` (442 lines) - Keep as-is or split?
- `styles.css` (359 lines) - Rename to better reflect purpose?

**Q11**: What's in `styles.css` that makes it "custom styles for dashboard"?
- HTMX loading indicators
- Animations (spin, shimmer, pulse-green)
- Status badges
- Toast notifications
- Log viewer
- Chart containers
- Print styles
- Dark mode transitions

**Q12**: Should we extract **utility classes** into a separate file?
- `.truncate-2-lines`
- `.blur-load`
- `.animate-spin`
- `.custom-scrollbar`
- `.focus-ring`

---

### Performance & Maintenance

**Q13**: Should we implement **CSS purging** (remove unused styles)?
- Tool: PurgeCSS or similar
- When: Build time
- Benefit: Smaller bundles for production

**Q14**: Should we add **CSS linting**?
- Tool: Stylelint
- Rules: Enforce naming conventions, no duplicates, max file size
- Integration: Pre-commit hooks

**Q15**: Should we track **CSS bundle size**?
- Current: ~3,596 lines custom CSS
- Target: ???
- Monitor: Automated size reports on PR

---

## 🎬 Proposed Action Plan (Awaiting Decisions)

### Phase 1: Critical Fixes (This Week)
1. **FIX NAVBAR VISIBILITY** (Answer Q4-Q6 first)
2. **Consolidate animations** into `common-animations.css` (Answer Q3)
3. **Split `agency-designer.css`** into 4 files (Answer Q10)

### Phase 2: Reorganization (Next Week)
4. **Create component-based structure** (Answer Q1)
5. **Extract theme variables** into separate file (Answer Q7-Q9)
6. **Consolidate component styles** (buttons, cards, modals)

### Phase 3: Optimization (Later)
7. **Implement CSS variables for design tokens** (Answer Q2)
8. **Add CSS linting** (Answer Q14)
9. **Set up bundle size monitoring** (Answer Q15)

---

## 📝 Decision Log

*To be filled in as questions are answered...*

| Question | Decision | Rationale | Date |
|----------|----------|-----------|------|
| Q4: Navbar margin purpose? | **BUG - Remove it** | Inline style `margin-left: -5rem` + CSS duplicate pushes navbar off-screen. Attempted to align logo to viewport edge but breaks layout. | Nov 28, 2025 |
| **Q1: Navbar vertical clipping?** | **✅ ROOT CAUSE FOUND** | `agency-designer.css` sets `html, body {overflow: hidden}` preventing scrolling. Navbar pushed above viewport gets clipped. Occurs on Agency Designer pages. | Nov 28, 2025 |
| **Q5: Navbar positioning fix?** | **✅ IMPLEMENTED** | Moved `@NavbarWithAgency()` outside `<main>` section in `layout_with_agency.templ`. Navbar now renders at body top level before main content. | Nov 28, 2025 |
| Q1: Component-based structure? | **✅ STARTED** | Created shared `HeadIncludes` component to centralize CSS/JS imports. Created `NavbarBurgerScript` component. Applied DRY principle to reduce duplication. | Nov 28, 2025 |
| Q3: Single animations file? | **⏸️ DEFERRED** | Will consolidate after layout fixes are stable. Current priority is navbar visibility and component extraction. | Nov 28, 2025 |
| **Code quality improvements** | **✅ IMPLEMENTED** | - Fixed QF1003 linting (converted if-else to switch in chat_panel.templ)<br/>- Added workbench-chat.js for workbench functionality<br/>- Refactored layout templates to use shared components | Nov 28, 2025 |

---

## 🐛 Q1 INVESTIGATION RESULTS: Navbar Visibility Issue

### ✅ ROOT CAUSE IDENTIFIED - VERTICAL CLIPPING

**Problem**: Navbar "moves up!" and disappears on Agency Designer pages

**Root Cause:**
```css
/* agency-designer.css lines 6-9 - CRITICAL BUG */
html, body {
    overflow: hidden;  /* ← PREVENTS vertical scrolling */
    height: 100%;
}
```

**What Happens:**
1. `agency-designer.css` disables ALL scrolling on the page (`overflow: hidden`)
2. When navbar is positioned/pushed above viewport top edge, it gets **clipped**
3. User cannot scroll up to see it - navbar is literally inaccessible
4. The `.vscode-designer-container` assumes 52px navbar height: `height: calc(100vh - 52px)`

**Affected Pages:**
- `/agencies/{id}/designer` - Main Agency Designer interface
- `/agencies/{id}/designer/raci` - RACI Designer page

**Contributing Factors:**
- Negative margins on navbar-brand (`margin-left: -5rem` inline + CSS duplicate)
- Container height calculation may be off if navbar != 52px
- Flexbox layout with negative margins can push elements vertically

### Solutions

**Option A: Remove `overflow: hidden` on html/body** (⚠️ May break VS Code layout)
```css
/* DELETE or comment out lines 6-9 */
/* html, body {
    overflow: hidden;
    height: 100%;
} */
```
- **Risk**: Layout designed to be non-scrollable (VS Code-style)
- **Test**: Check if designer panels still work correctly

**Option B: Apply `overflow: hidden` to container only**
```css
/* MOVE overflow constraint from html/body to container */
.vscode-designer-container {
    overflow: hidden;  /* Apply here instead of html/body */
    height: calc(100vh - 52px);
}
```
- **Safer**: Preserves layout intent, allows navbar to be visible

**Option C: Fix navbar positioning to stay within viewport**
```css
/* Ensure navbar is position: sticky or fixed at top */
.navbar {
    position: sticky;
    top: 0;
    z-index: 30;
}
```
- **Cleanest**: Navbar always visible at top edge
- **Adjust calc**: `calc(100vh - var(--navbar-height))` dynamically

### Recommended Fix: **Combination of B + C**

1. **Remove `overflow: hidden` from `html, body`**
2. **Apply it to `.vscode-designer-container` instead**
3. **Make navbar sticky** to ensure it stays at top
4. **Remove negative margins** (see Q4 investigation below)

---

## 🐛 Q4 INVESTIGATION RESULTS: Navbar Horizontal Margin Issue

### Root Cause Identified

**Problem Structure:**
```html
<nav class="navbar">
  <div class="container">                             <!-- Bulma: max-width, centered, padding -->
    <div class="navbar-brand" style="margin-left: -5rem;">  <!-- 🚨 INLINE STYLE -->
```

**Compounded by CSS:**
```css
/* common-layout.css line 22-24 */
.navbar-brand {
    margin-left: -5rem;  /* 🚨 DUPLICATE! */
}
```

### Impact Analysis

1. **Double negative margin**: Inline style + CSS = potentially -10rem (160px) shift
2. **Breaks container alignment**: Logo and burger menu pushed outside visible area
3. **Mobile hamburger invisible**: Burger menu off-screen on small viewports

### Intended Purpose (Best Guess)

Someone tried to make the navbar-brand align with the left viewport edge (ignoring container padding), but this:
- ❌ Breaks responsive design
- ❌ Makes navbar invisible on narrow screens
- ❌ Violates Bulma container pattern

### Correct Solutions

**Option A: Remove negative margin entirely** (Recommended)
- Let Bulma container handle alignment naturally
- Navbar respects responsive padding
- Logo stays visible on all screen sizes

**Option B: Use full-width navbar** (If left-edge alignment is required)
```html
<nav class="navbar">
  <!-- NO container wrapper -->
  <div class="navbar-brand">...</div>
  <div class="navbar-menu">...</div>
</nav>
```

**Option C: Use Bulma's `.is-fullwidth` modifier**
```html
<nav class="navbar">
  <div class="container is-fluid">  <!-- Full width with small padding -->
    <div class="navbar-brand">...</div>
```

### Decision: **Option A** - Remove negative margins

**Rationale**:
- Simplest fix
- Maintains responsive behavior
- Follows Bulma best practices
- No unintended side effects

---

## 🔍 Next Steps

**Before proceeding, we need decisions on:**

1. ✅ **Navbar visibility issue** (Q4-Q6) - BLOCKING
2. ✅ **File organization strategy** (Q1, Q10)
3. ✅ **Animation consolidation** (Q3)
4. ⏸️ **Theme system scope** (Q7-Q9) - Can defer
5. ⏸️ **Utility extraction** (Q12) - Can defer

**Once we have answers, I can:**
- Create refactoring tasks in MVP backlog
- Implement changes following the agreed strategy
- Set up linting/automation if approved

---

## 📚 References

- **Instruction Files**: `.github/copilot-instructions.md`, `.github/instructions/rules.instructions.md`
- **Current Branch**: `feature/MVP-WI-012_workbench_chat_panel`
- **Related Files**: All CSS/SCSS files in `static/css/` and `static/scss/`
- **SCSS Documentation**: [Sass Official Docs](https://sass-lang.com/documentation)
- **Build Integration**: See `Makefile` for `make css` target

---

## 🎓 SCSS Developer Guide

### How to Work with SCSS

#### 1. Editing SCSS Files
```bash
# Edit source files in static/scss/ (NOT static/css/)
vim static/scss/agency-designer/_layout.scss

# Compile changes
make css

# Or build/run (auto-compiles)
make run
```

#### 2. Using Variables
```scss
// Import variables at the top of your SCSS file
@import 'variables';

// Use variables in your styles
.my-component {
    background-color: $vscode-bg;
    padding: $spacing-md;
    font-size: $font-size-sm;
    transition: all $transition-normal;
}
```

#### 3. Using Mixins
```scss
// Import mixins
@import 'mixins';

// Use mixins with @include
.my-panel {
    @include vscode-panel;
    @include custom-scrollbar(8px);
    
    &:hover {
        @include hover-bg;
    }
}
```

#### 4. Adding New Modules
```bash
# Create new module file
touch static/scss/my-feature/_module.scss

# Import in main entry point
echo "@import 'my-feature/module';" >> static/scss/my-feature.scss

# Compile
make css
```

#### 5. SCSS Best Practices
- **Use nesting** for cleaner hierarchy (max 3 levels deep)
- **Extract variables** for repeated values
- **Create mixins** for repeated patterns
- **Keep modules focused** - one domain per file
- **Use descriptive names** - `$vscode-accent-hover` not `$color-blue-light`
- **Avoid `!important`** - use proper specificity instead

### Example: Creating a New SCSS Module

```scss
// static/scss/my-feature.scss
@import 'variables';
@import 'mixins';
@import 'my-feature/layout';
@import 'my-feature/components';

// static/scss/my-feature/_layout.scss
.my-feature-container {
    @include flex-column;
    background-color: $vscode-bg;
    padding: $spacing-lg;
    
    .header {
        @include panel-header;
        margin-bottom: $spacing-md;
    }
    
    .content {
        @include custom-scrollbar(6px);
        flex: 1;
    }
}
```

### Compilation Details

**Input**: `static/scss/**/*.scss`  
**Output**: `static/css/**/*.css`  
**Compiler**: Dart Sass via npm  
**Options**: 
- `--no-source-map` - No source maps (cleaner output)
- `--style=compressed` - Minified output

**Build process**:
```makefile
# Makefile target
css:
    npx sass static/scss:static/css --no-source-map --style=compressed
```

**Git configuration**:
```gitignore
# .gitignore - Don't commit compiled CSS
static/css/*.css
static/css/*.css.map
!static/css/bulma.min.css     # Keep 3rd party
!static/css/mapbox-gl.css
!static/css/maplibre-gl.css
```

### Migration Checklist (for other CSS files)

When migrating a plain CSS file to SCSS:

- [ ] Create SCSS directory: `static/scss/[feature-name]/`
- [ ] Split CSS into logical modules (<400 lines each)
- [ ] Extract colors/spacing to `_variables.scss`
- [ ] Identify repeated patterns for `_mixins.scss`
- [ ] Convert classes to use nesting
- [ ] Replace hardcoded values with variables
- [ ] Use mixins for common patterns
- [ ] Create main entry point that imports modules
- [ ] Update `head_includes.templ` if needed
- [ ] Compile: `make css`
- [ ] Test in browser (hard refresh: Ctrl+Shift+R)
- [ ] Add to `.gitignore` if output CSS is auto-generated
- [ ] Commit SCSS sources only (not compiled CSS)

---

## 🎯 SCSS Migration Roadmap

### Phase 1: Foundation (✅ COMPLETED - Nov 28, 2025)
- [x] Set up SCSS infrastructure (Sass, Makefile, .gitignore)
- [x] Create shared `_variables.scss` with design tokens
- [x] Create shared `_mixins.scss` with reusable patterns
- [x] Migrate `agency-designer.css` to modular SCSS (6 modules)
- [x] Verify build integration (`make css` in `make build`/`make run`)

### Phase 2: Core Pages (Q1 2026)
- [ ] Migrate `workflow-designer.css` (442 lines)
- [ ] Migrate `agencies.css` (233 lines)
- [ ] Migrate `vscode-designer-shared.css` (485 lines) to SCSS partial
- [ ] Expand `_variables.scss` with workflow/agency-specific tokens
- [ ] Create workflow-specific mixins

### Phase 3: Global Styles (Q1 2026)
- [ ] Migrate `styles.css` (1,044 lines) - split into multiple modules
- [ ] Migrate `themes.css` (361 lines) - use SCSS maps for theme system
- [ ] Migrate `common-layout.css` (118 lines)
- [ ] Create `_components.scss` for shared button/card/modal patterns
- [ ] Consolidate all animations into SCSS with variables

### Phase 4: Modern Practices (Q2 2026)
- [ ] Migrate from `@import` to `@use`/`@forward` syntax
- [ ] Set up Stylelint for SCSS linting
- [ ] Add CSS bundle size monitoring
- [ ] Implement CSS purging for production (PurgeCSS)
- [ ] Create comprehensive SCSS documentation

### Phase 5: Optimization (Q2 2026)
- [ ] Analyze CSS bundle sizes and optimize
- [ ] Remove unused styles (PurgeCSS)
- [ ] Set up CSS performance budgets
- [ ] Consider CSS-in-JS for dynamic components
- [ ] Automate CSS linting in CI/CD pipeline

**End Goal**: 100% of custom styles in modular SCSS by mid-2026
