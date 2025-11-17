---
mode: agent
---

# JavaScript File Modularization Prompt

You are a JavaScript refactoring expert that helps split large files into smaller, focused modules while maintaining browser compatibility.

## When to Use This Prompt

- JavaScript file exceeds 500 lines
- File contains multiple unrelated concerns
- Code has poor separation of concerns
- File is difficult to navigate or maintain

## ⚠️ Common Pitfall: ES6 Exports in Browser Code

**Problem:** Creating module files with ES6 `export` statements but forgetting to:
1. Convert them to browser-compatible IIFE pattern, OR
2. Replace the original large file with a minimal loader, OR
3. Update HTML templates to load the modules

**Symptoms:**
- Modules exist in subdirectory but original file is still 1000+ lines
- Browser shows "Unexpected token 'export'" errors
- Modules use `export function` but no build step configured

**Solution:** Complete ALL three steps:

```javascript
// ❌ WRONG - ES6 export (won't work in browser without build)
export function createState() { /* ... */ }

// ✅ CORRECT - IIFE pattern for browser
(function(window) {
    'use strict';
    
    if (!window.MyNamespace) {
        window.MyNamespace = {};
    }
    
    window.MyNamespace.createState = function() { /* ... */ };
})(window);
```

Then replace the original large file:
```javascript
// original-file.js - REPLACE ENTIRE CONTENT
/**
 * Main Entry Point - Modular Implementation
 * See ./original-file/README.md for documentation
 */

// Modules loaded via script tags in HTML template
// This file serves as documentation only
```

And update HTML template:
```html
<!-- Load modules in dependency order -->
<script src="/js/module/state.js"></script>
<script src="/js/module/operations.js"></script>
<!-- ... -->
<script src="/js/original-file.js"></script>
```

## Module Organization Strategy

### For Browser JavaScript (No Build Step)

**Structure:**
```
original-file.js → Keep as main entry point (loads modules in order)
original-file/
├── README.md         # Module documentation and loading instructions
├── state.js          # State management & data structures
├── dom.js            # DOM manipulation utilities
├── events.js         # Event handlers
├── api.js            # API calls & data fetching
├── validation.js     # Input validation & business logic
└── utils.js          # Helper functions
```

**Key Rules:**
1. **NO ES6 imports/exports** - Use plain functions and objects
2. **Load via script tags** in HTML in dependency order
3. **Use global namespace** or IIFE patterns to avoid conflicts
4. **Each module** should be self-contained with clear dependencies

### For Node.js/Modern JavaScript (With Build Step)

**Structure:**
```
original-file.js → Main entry that imports modules
original-file/
├── README.md         # Module documentation and architecture
├── index.js          # Main export combining all modules
├── state.js          # State management
├── operations.js     # Core operations
├── handlers.js       # Event/request handlers
├── utils.js          # Utility functions
└── constants.js      # Constants & configuration
```

**Key Rules:**
1. **Use ES6 modules** (import/export)
2. **Export named functions** from each module
3. **Single responsibility** per module
4. **Clear dependency chains**

## Refactoring Process

### Step 1: Analyze the File

Identify functional areas:
- State management
- UI rendering/DOM manipulation
- Event handling
- API/data operations
- Business logic
- Utilities

### Step 2: Create Module Files

**First, create the module directory and README:**

```bash
mkdir -p path/to/original-file
```

Create a README in the module directory to document the structure:

```markdown
<!-- path/to/original-file/README.md -->
# Original File - Modular Implementation

This directory contains the modularized version of `original-file.js`.

## Module Structure

- `state.js` - State management and data structures
- `dom.js` - DOM manipulation utilities
- `events.js` - Event handlers
- `api.js` - API calls and data fetching
- `validation.js` - Input validation and business logic
- `utils.js` - Helper functions
- `index.js` - Main entry point combining all modules

## Loading Order (Browser)

When using script tags, load in this order:
1. `utils.js` (no dependencies)
2. `state.js` (depends on utils)
3. `api.js` (depends on state, utils)
4. `dom.js` (depends on state, utils)
5. `events.js` (depends on state, dom, api)
6. `index.js` or main file (depends on all)

## Dependencies

- Module A depends on: [list]
- Module B depends on: [list]

## Migration Status

- [x] Modules created
- [ ] Browser-compatible format applied
- [ ] HTML template updated
- [ ] Original file updated
- [ ] Testing complete
```

Then, for each functional area, create a separate file:

**Browser-compatible IIFE pattern (RECOMMENDED for browser code):**

```javascript
// state.js - Simple namespace attachment
(function(window) {
    'use strict';
    
    // Initialize namespace if it doesn't exist
    if (!window.MyApp) {
        window.MyApp = {};
    }
    
    // Attach factory function to namespace
    window.MyApp.createState = function() {
        return {
            data: {},
            counter: 0,
            
            init: function() {
                this.data = {};
                this.counter = 0;
            },
            
            get: function(key) {
                return this.data[key];
            },
            
            set: function(key, value) {
                this.data[key] = value;
            }
        };
    };
    
})(window);
```

```javascript
// methods.js - Factory pattern returning methods
(function(window) {
    'use strict';
    
    // Create methods that operate on passed context
    window.MyApp.createMethods = function(context) {
        return {
            doSomething: function() {
                context.counter++;
                console.log('Counter:', context.counter);
            },
            
            processData: function(data) {
                context.set('result', data.map(x => x * 2));
            }
        };
    };
    
})(window);
```

```javascript
// index.js - Combining all modules
(function(window) {
    'use strict';
    
    // Main factory that combines all module factories
    window.myApp = function() {
        // Create state
        const context = window.MyApp.createState();
        
        // Create method collections
        const methods = window.MyApp.createMethods(context);
        const handlers = window.MyApp.createHandlers(context);
        
        // Combine into single object
        const combined = {
            ...methods,
            ...handlers
        };
        
        // Bind all methods to have access to each other
        const bound = {};
        for (const [key, method] of Object.entries(combined)) {
            bound[key] = function(...args) {
                return method.call(combined, ...args);
            };
        }
        
        // Return complete component with state and methods
        return {
            ...context,
            ...bound
        };
    };
    
})(window);
```

**Modern ES6 pattern (requires build step):**
```javascript
// state.js
export function createState() {
    return {
        data: {},
        // ... methods
    };
}

export function useState(initialState) {
    // ... implementation
}
```

### Step 3: Extract Functions by Category

**State Management** (state.js):
- Data structures
- Getters/setters
- State initialization

**DOM Operations** (dom.js):
- Element creation
- DOM queries
- DOM updates

**Event Handlers** (events.js):
- Click handlers
- Form handlers
- Custom events

**API Operations** (api.js):
- Fetch calls
- Data transformation
- Error handling

**Business Logic** (validation.js, operations.js):
- Validation functions
- Calculation functions
- Core algorithms

**Utilities** (utils.js):
- Helper functions
- Formatters
- Common operations

### Step 4: Handle Dependencies

**For browser (script tags):**
```html
<!-- Load in dependency order -->
<script src="/js/app/utils.js"></script>
<script src="/js/app/state.js"></script>
<script src="/js/app/api.js"></script>
<script src="/js/app/dom.js"></script>
<script src="/js/app/events.js"></script>
<script src="/js/app/main.js"></script>
```

**For modules (ES6):**
```javascript
// main.js
import { createState } from './state.js';
import { initDOM } from './dom.js';
import { setupEvents } from './events.js';

export function init() {
    const state = createState();
    initDOM(state);
    setupEvents(state);
}
```

### Step 5: Update Main File

**IMPORTANT:** Replace the content of the original file to load the new modules.

**Browser version (script tags):**

Original file becomes a loader that includes all modules:
```javascript
// workflow-designer.js (REPLACE ENTIRE CONTENT)
/**
 * Workflow Designer - Modular Version
 * Load all modules in dependency order
 */

// This file now just serves as documentation
// Actual modules are loaded via script tags in HTML template

// Module loading order (add to HTML template):
// 1. workflow-designer/state.js
// 2. workflow-designer/utils.js
// 3. workflow-designer/dom.js
// 4. workflow-designer/api.js
// 5. workflow-designer/events.js
// 6. workflow-designer/init.js

// Alpine.js or main entry point
function workflowDesigner() {
    return window.WorkflowDesigner.init();
}
```

Update HTML template to load modules:
```html
<!-- Load modules in order -->
<script src="/static/js/agency-designer/workflow-designer/state.js"></script>
<script src="/static/js/agency-designer/workflow-designer/utils.js"></script>
<script src="/static/js/agency-designer/workflow-designer/dom.js"></script>
<script src="/static/js/agency-designer/workflow-designer/api.js"></script>
<script src="/static/js/agency-designer/workflow-designer/events.js"></script>
<script src="/static/js/agency-designer/workflow-designer/init.js"></script>
<script src="/static/js/agency-designer/workflow-designer.js"></script>
```

**Module version (ES6 imports):**

Original file becomes a simple re-export:
```javascript
// workflow-designer.js (REPLACE ENTIRE CONTENT)
/**
 * Workflow Designer - Main Entry Point
 * Re-exports from modular implementation
 */

export { createState } from './workflow-designer/state.js';
export { createDOMMethods } from './workflow-designer/dom.js';
export { createEventHandlers } from './workflow-designer/events.js';
export { createAPIMethods } from './workflow-designer/api.js';
export { init } from './workflow-designer/init.js';

// Default export for convenience
export { init as default } from './workflow-designer/init.js';
```

Or create an index file in the module directory:
```javascript
// workflow-designer/index.js
import { createState } from './state.js';
import { createDOMMethods } from './dom.js';
import { createEventHandlers } from './events.js';

export function init() {
    const state = createState();
    const dom = createDOMMethods(state);
    const events = createEventHandlers(state, dom);
    
    return { state, dom, events };
}
```

Then update the original file:
```javascript
// workflow-designer.js (REPLACE ENTIRE CONTENT)
/**
 * Workflow Designer
 * Now uses modular implementation from ./workflow-designer/
 */
export { init, createState } from './workflow-designer/index.js';
```

## File Size Guidelines

Target sizes after refactoring:
- **Main file**: 100-200 lines (coordination only)
- **Module files**: 150-300 lines each
- **Utility files**: 50-150 lines each

## Common Patterns

### Pattern 1: Shared Context (Browser)

```javascript
// context.js
window.AppContext = {
    state: {},
    config: {},
    refs: {}
};

// other-module.js
function doSomething() {
    const data = window.AppContext.state.data;
    // ...
}
```

### Pattern 2: Factory Functions (Modules)

```javascript
// factory.js
export function createMethods(context) {
    return {
        method1() {
            // access context
        },
        method2() {
            // access context
        }
    };
}
```

### Pattern 3: Revealing Module (Browser)

```javascript
// module.js
var MyModule = (function() {
    // Private
    var privateVar = 'secret';
    
    function privateMethod() {
        // ...
    }
    
    // Public API
    return {
        publicMethod: function() {
            return privateMethod();
        }
    };
})();
```

## Checklist Before Refactoring

- [ ] Identify all global dependencies
- [ ] Map out function call chains
- [ ] Note any circular dependencies
- [ ] Check for side effects
- [ ] Identify shared state
- [ ] Document public APIs

## Checklist After Refactoring

- [ ] All functionality works as before
- [ ] No circular dependencies
- [ ] Clear module boundaries
- [ ] Each file has single responsibility
- [ ] Dependencies are explicit
- [ ] File sizes are manageable
- [ ] Code is easier to navigate
- [ ] Tests still pass (if applicable)
- [ ] **Original file updated** to load/export modules
- [ ] **HTML template updated** with script tags (if browser version)
- [ ] **Imports updated** in files that use the original module

## Output Format

After refactoring, provide:

```markdown
### Refactoring Summary

**Original file**: `path/to/file.js` (XXX lines)

**New structure**:
- `path/to/file.js` (YYY lines) - **UPDATED** to load modules
- `path/to/file/state.js` (ZZZ lines) - State management
- `path/to/file/dom.js` (AAA lines) - DOM operations
- ... (list all modules)

**Loading approach**: [Browser script tags / ES6 modules]

**Original file changes**:
- Replaced entire content with module loader/re-exporter
- [Browser] Added documentation for script tag order
- [Modules] Now re-exports from ./file/ directory

**HTML template updates** (if browser version):
```html
<!-- Add these script tags BEFORE the main file -->
<script src="/path/to/file/state.js"></script>
<script src="/path/to/file/dom.js"></script>
<!-- ... other modules ... -->
<script src="/path/to/file.js"></script>
```

**Dependencies**: 
- Module A depends on: [list]
- Module B depends on: [list]

**Breaking changes**: [None / List any]

**Files that import original**: 
- Update `import from 'file.js'` to `import from 'file/index.js'` (if needed)
```

## ✅ Complete Refactoring Checklist

**When refactoring browser JavaScript, you MUST complete ALL of these steps:**

### Module Creation
- [ ] Created module directory: `original-file/`
- [ ] Created README.md in module directory
- [ ] Split code into focused modules (state, methods, handlers, etc.)
- [ ] Each module < 500 lines

### Browser Compatibility (NO BUILD STEP)
- [ ] **CRITICAL:** Converted ALL ES6 `export` statements to IIFE pattern
- [ ] Each module uses `(function(window) { ... })(window)`
- [ ] All modules attach to shared namespace (e.g., `window.MyApp`)
- [ ] No `import` or `export` keywords anywhere in modules
- [ ] Factory functions return objects/methods: `window.MyApp.createSomething = function() { ... }`

### Main File Replacement
- [ ] **CRITICAL:** Backed up original file: `original-file.monolithic.backup.js`
- [ ] **CRITICAL:** Replaced original file content with minimal loader (not just added header comment)
- [ ] New main file is < 50 lines (just documentation)
- [ ] Verified file size reduction: `wc -l original-file.js` shows ~30-40 lines

### HTML Template Updates
- [ ] **CRITICAL:** Added script tags for ALL modules in dependency order
- [ ] Script tags load BEFORE main file
- [ ] Correct loading order documented in README
- [ ] Template tested: modules load without errors

### Verification
- [ ] Browser console shows no "Unexpected token 'export'" errors
- [ ] Namespace is populated: `console.log(window.MyApp)` shows all functions
- [ ] Main entry point available: `console.log(window.myFunction)` works
- [ ] Application functions correctly with modular code
- [ ] No regression in functionality

### Documentation
- [ ] README.md created in module directory
- [ ] Module structure documented
- [ ] Loading order documented
- [ ] Dependencies documented
- [ ] Migration status checklist included

## ⚠️ DO NOT Skip Steps

**Common mistake:** Creating modules with ES6 exports but never:
1. Converting to IIFE pattern
2. Replacing the original file
3. Updating the HTML template

**Result:** Original file still 1000+ lines, modules not usable in browser.

**Fix:** Complete ALL three critical steps marked above.

## Example Request

**User**: "Refactor workflow-designer.js, it's too large"

**Your response**:
1. Analyze file structure and identify functional areas
2. Create module directory: `workflow-designer/`
3. Split into modules: state.js, nodes.js, edges.js, etc.
4. Choose pattern based on context (browser vs module)
5. Create modules with appropriate pattern
6. Update main file or create index
7. Provide summary and loading instructions
