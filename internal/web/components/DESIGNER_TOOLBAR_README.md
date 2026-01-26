# Designer Toolbar Component

## Overview

The `DesignerToolbar` and `AgencyDesignerHeader` components provide consistent, reusable toolbar headers for designer pages across the application.

## Location

`/workspaces/CodeValdCortex/internal/web/components/designer_toolbar.templ`

## Components

### 1. DesignerToolbar

A general-purpose toolbar for designer pages with customizable title, back button, and action buttons.

**Usage Example:**
```templ
@components.DesignerToolbar(
    "My Workflow",           // title
    "",                      // titleIcon (optional, e.g., "sitemap")
    "/agencies/123/designer#workflows",  // backURL
    "Back to Workflows",     // backLabel
    []components.DesignerToolbarButton{
        {
            Label: "Save",
            Icon: "save",
            IsPrimary: true,
            AlpineClick: "saveWorkflow()",
            AlpineDisabled: "saving",
            AlpineText: "saving ? 'Saving...' : 'Save'",
        },
        {
            Label: "Export",
            Icon: "download",
            IsLight: true,
            AlpineClick: "exportWorkflow()",
        },
    },
)
```

### 2. AgencyDesignerHeader

A specialized header for the agency designer with dark background styling.

**Usage Example:**
```templ
@components.AgencyDesignerHeader(
    []components.DesignerToolbarButton{
        {
            Label: "Validate",
            Icon: "check-circle",
            IsInfo: true,
            ID: "validate-btn",
            Title: "Validate agency specification",
        },
        {
            Label: "Publish",
            Icon: "upload",
            IsPrimary: true,
            ID: "publish-btn",
            Title: "Publish agency as new version",
        },
    },
    "",  // statusTag (optional)
)
```

## DesignerToolbarButton Fields

| Field | Type | Description |
|-------|------|-------------|
| `Label` | string | Button text label |
| `Icon` | string | FontAwesome icon name (without `fa-` prefix) |
| `IsPrimary` | bool | Apply primary button styling (blue) |
| `IsInfo` | bool | Apply info button styling (cyan) |
| `IsLight` | bool | Apply light button styling (light gray) |
| `AlpineClick` | string | Alpine.js @click binding (e.g., `"saveWorkflow()"`) |
| `AlpineDisabled` | string | Alpine.js :disabled binding (e.g., `"saving"`) |
| `AlpineText` | string | Alpine.js x-text binding for dynamic text (e.g., `"saving ? 'Saving...' : 'Save'"`) |
| `ID` | string | HTML id attribute |
| `Title` | string | HTML title attribute (tooltip) |

## CSS Classes

### DesignerToolbar
- `.designer-toolbar` - Main container
- Uses Bulma's `.level` system for responsive layout
- Supports `.is-mobile` for mobile-friendly layouts

### AgencyDesignerHeader  
- `.agency-designer-header` - Main container with dark background
- Uses Bulma's `.level` system
- Title uses `.has-text-white` for white text on dark background

## Alpine.js Integration

The component supports Alpine.js reactive bindings:

1. **@click**: Use `AlpineClick` field for click handlers
2. **:disabled**: Use `AlpineDisabled` field for conditional disabling
3. **x-text**: Use `AlpineText` field for dynamic button text

**Example:**
```go
components.DesignerToolbarButton{
    Label: "Save",           // Static fallback text
    AlpineClick: "saveWorkflow()",
    AlpineDisabled: "saving",
    AlpineText: "saving ? 'Saving...' : 'Save'",  // Dynamic text
}
```

## Benefits

1. **Consistency**: Ensures all designer pages have the same look and feel
2. **DRY Principle**: Single source of truth for toolbar markup
3. **Maintainability**: Changes to toolbar styling/structure only need to be made once
4. **Type Safety**: Templ provides compile-time checking of component usage
5. **Flexibility**: Support for both static and Alpine.js reactive buttons

## Current Usage

- ✅ Workflow Designer (`workflow_designer.templ`)
- 🔄 Agency Designer (can be migrated to use `AgencyDesignerHeader`)

## Future Enhancements

- Add support for onclick handlers (non-Alpine.js)
- Add support for dropdown menus in toolbar
- Add support for custom left-side content beyond back button
- Add support for badges/notifications in toolbar
