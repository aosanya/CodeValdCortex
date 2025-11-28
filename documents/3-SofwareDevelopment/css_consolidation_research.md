# CSS Consolidation Research & Strategy

**Date**: November 28, 2025  
**Current Branch**: feature/MVP-WI-012_workbench_chat_panel  
**Issue**: CSS organization is messy, common patterns are duplicated, navbar visibility issues

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
Commit: 15c5ffd
Message: feat(ui): refactor layout components and add workbench chat
Files: 16 changed, 1315 insertions(+), 401 deletions(-)
```

---

## � Remaining Work (Prioritized)

### Priority 0: Critical Issues Still Open
- [ ] **Remove negative margin from navbar-brand** (See Q4 investigation)
  - Remove inline `style="margin-left: -5rem;"` from `navbar_with_agency.templ`
  - Remove CSS `.navbar-brand { margin-left: -5rem; }` from `common-layout.css`
  - Test responsive behavior on mobile
  
### Priority 1: File Size Violations
- [ ] **Split `agency-designer.css` (1,792 lines → 4 files)**
  - Extract to: `agency-designer-layout.css`, `-chat.css`, `-context.css`, `-forms.css`
  - Update templ files to include split files
  
### Priority 2: Consolidation
- [ ] **Create `common-animations.css`**
  - Move all `@keyframes` from `styles.css`, `agencies.css`, `agency-designer.css`, `workflow-designer.css`
  - Standardize animation names (remove duplicates: spin, fade, slide, pulse)
  
- [ ] **Extract common component styles**
  - Create `common-components.css` for buttons, cards, modals
  - Remove duplicates across files
  
### Priority 3: Theme System Review
- [ ] **Reduce `!important` usage in `themes.css`**
- [ ] **Evaluate if all 6 themes are needed**
- [ ] **Consider extracting theme variables** to separate file

### Priority 4: Optimization (Future)
- [ ] Set up CSS linting (Stylelint)
- [ ] Implement CSS purging for production builds
- [ ] Add bundle size monitoring

---

## �🔍 Current State Analysis

### CSS Files Inventory (by size)

| File | Lines | Purpose | Status |
|------|-------|---------|--------|
| `agency-designer.css` | 1,792 | VS Code-style designer layout | ⚠️ **MASSIVE** |
| `workflow-designer.css` | 442 | Workflow drag-drop designer | 🔴 Medium complexity |
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
- **Related Files**: All CSS files in `static/css/`
