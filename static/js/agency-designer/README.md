# Agency Designer JavaScript Modules

This directory contains the modularized JavaScript code for the Agency Designer interface. The original monolithic `agency-designer.js` file has been split into focused modules for better maintainability.

## File Structure

```
/static/js/agency-designer/
├── index.js           # Entry point and module loader
├── main.js            # Main initialization and coordination
├── utils.js           # Common utility functions
├── chat.js            # Chat interface and messaging
├── views.js           # View switching (overview, agent-types, layout)
├── agents.js          # Agent type selection and management
├── overview.js        # Overview section navigation
├── introduction.js    # Introduction editor functionality
├── problems.js        # Problem definition management
├── units.js           # Units of Work management
└── htmx.js            # HTMX event handling
```

## Module Responsibilities

### `index.js`
- Entry point for the modular system
- Handles ES6 module loading with fallback to original file
- Browser compatibility layer

### `main.js` 
- Coordinates all module initialization
- Exports functions to global scope for onclick handlers
- Main DOM ready event handler

### `utils.js`
- `getCurrentAgencyId()` - Extract agency ID from URL or attributes
- `showNotification()` - Display toast notifications
- Other shared utility functions

### `chat.js`
- `initializeChatScroll()` - Auto-scroll setup for chat
- `scrollToBottom()` - Scroll chat to latest message
- Chat interface management

### `views.js`
- `initializeViewSwitcher()` - Tab navigation setup
- `switchView()` - Handle view transitions (overview/agent-types/layout)
- View state management

### `agents.js`
- `selectAgentType()` - Handle agent selection in sidebar
- `initializeAgentSelection()` - Auto-select first agent
- Agent type interface management

### `overview.js`
- `initializeOverview()` - Overview section setup
- `selectOverviewSection()` - Handle section switching (intro/problems/units)
- Section navigation coordination

### `introduction.js`
- `loadIntroductionEditor()` - Load existing introduction text
- `saveOverviewIntroduction()` - Save introduction changes
- `undoOverviewIntroduction()` - Revert introduction changes
- Introduction editor state management

### `problems.js`
- `loadProblems()` - Fetch and display problems list
- `showProblemEditor()` - Show add/edit problem form
- `saveProblemFromEditor()` - Save problem changes
- `cancelProblemEdit()` - Cancel problem editing
- `deleteProblem()` - Delete problem with confirmation
- Problem CRUD operations

### `units.js`
- `loadUnits()` - Fetch and display units of work list
- `showUnitEditor()` - Show add/edit unit form
- `saveUnitFromEditor()` - Save unit changes
- `cancelUnitEdit()` - Cancel unit editing
- `deleteUnit()` - Delete unit with confirmation
- Units of Work CRUD operations

### `htmx.js`
- `initializeHTMXEvents()` - Set up HTMX event listeners
- Typing indicator management
- Error handling for HTMX requests
- Form submission and response handling

## Usage

The modular system is designed to be a drop-in replacement for the original monolithic file. Simply include the entry point:

```html
<script src="/static/js/agency-designer/index.js" type="module"></script>
```

For browsers that don't support ES6 modules, the system will automatically fall back to the original `agency-designer-original.js` file.

## Benefits

1. **Maintainability** - Each module has a single responsibility
2. **Reusability** - Modules can be imported individually
3. **Testing** - Each module can be unit tested separately  
4. **Development** - Easier to find and modify specific functionality
5. **Performance** - Can implement lazy loading if needed
6. **Collaboration** - Multiple developers can work on different modules

## Global Functions

For compatibility with onclick handlers in templates, all interactive functions are exported to the global `window` object:

- `window.selectAgentType`
- `window.selectOverviewSection`
- `window.saveOverviewIntroduction`
- `window.undoOverviewIntroduction`
- `window.showProblemEditor`
- `window.saveProblemFromEditor`
- `window.cancelProblemEdit`
- `window.deleteProblem`
- `window.showUnitEditor`
- `window.saveUnitFromEditor`
- `window.cancelUnitEdit`
- `window.deleteUnit`

## Migration Notes

- Original file backed up as `agency-designer-original.js`
- All functionality preserved with identical APIs
- No changes required to HTML templates
- Automatic fallback ensures compatibility