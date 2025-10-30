# Status Bar Component

A reusable VS Code-style status bar component that can be used across all pages in the application.

## Component Location

- **Template**: `/internal/web/components/status_bar.templ`
- **CSS**: `/static/css/vscode-status-bar.css`

## Usage

### Basic Usage with Agency and Conversation

```go
@components.StatusBar(currentAgency, conversation)
```

This will display:
- Agency name with network icon
- Conversation phase (if conversation exists)
- Action button for validation phase (if applicable)

### Simple Usage with Custom Items

```go
@components.StatusBarSimple(currentAgency, "Dashboard", "Ready")
```

This allows you to add custom status items without conversation info.

## Page Setup

To use the status bar on a page, follow this pattern:

### 1. HTML Structure

```go
<div class="page-with-status-bar">
    <div class="page-main-content">
        <!-- Your page content here -->
    </div>
    @components.StatusBar(currentAgency, conversation)
</div>
```

### 2. Include CSS

Add this to your page template:

```html
<link rel="stylesheet" href="/static/css/vscode-status-bar.css"/>
```

Or import in your page-specific CSS:

```css
@import url('./vscode-status-bar.css');
```

## CSS Classes

### Container Classes

- `.page-with-status-bar` - Grid container for page with status bar
- `.page-main-content` - Main content area that fills available space
- `.vscode-status-bar` - Status bar itself (grid-area: status)

### Status Bar Elements

- `.status-bar-left` - Left side container for status items
- `.status-bar-right` - Right side container for actions
- `.status-item` - Individual status item
- `.status-text` - Text content (truncates with ellipsis)
- `.status-separator` - Pipe separator between items
- `.status-action-btn` - Action button styling

## Features

- **Fixed Height**: 22px (like VS Code)
- **Grid Layout**: Uses CSS Grid for proper height constraints
- **Responsive**: Text truncates with ellipsis if too long
- **Blue Background**: Matches VS Code accent color (#007acc)
- **Icon Support**: Font Awesome icons integrated

## Example Implementation

See `/internal/web/pages/agency_designer/agency_designer.templ` for a complete example.

The agency designer page structure:

```go
<div class="vscode-designer-container">
    <div class="columns is-gapless vscode-main-content">
        <!-- Page content -->
    </div>
    @components.StatusBar(currentAgency, conversation)
</div>
```

## Customization

The status bar CSS uses standard colors that can be overridden:
- Background: `#007acc` (VS Code blue)
- Text: `white`
- Border: `#e5e5e5`

## Notes

- The status bar always stays at the bottom (22px height)
- Content above scrolls independently
- Text truncates automatically if agency name is too long
- Action buttons adapt to conversation phase automatically
