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
- Functions are difficult to test in isolation

## Core Refactoring Principles

### File Size Limits
- **Target maximum**: 500 lines per module (STRICT)
- **Ideal range**: 150-400 lines per module
- **Main entry file**: 50-100 lines maximum (just coordination)
- If a module exceeds 500 lines, split it further

### Functional Programming for Testability
- **Prefer pure functions**: No side effects, deterministic output
- **Extract side effects**: Separate pure logic from I/O, DOM, API calls
- **Dependency injection**: Pass dependencies as parameters, not globals
- **Single responsibility**: Each function does ONE thing
- **Avoid mutations**: Use immutable patterns where possible

**Example - BAD (hard to test):**
```javascript
// Mixes DOM, state mutation, and business logic
function updateUserStatus(userId) {
    const user = window.appState.users.find(u => u.id === userId);
    user.status = 'active';
    document.getElementById('status-' + userId).textContent = 'Active';
    fetch('/api/users/' + userId, { method: 'PUT', body: JSON.stringify(user) });
}
```

**Example - GOOD (testable):**
```javascript
// Pure function - easy to test
function calculateNewStatus(user, action) {
    return { ...user, status: action === 'activate' ? 'active' : 'inactive' };
}

// Side effect - separate concern
function updateUserInDOM(userId, status) {
    const el = document.getElementById('status-' + userId);
    if (el) el.textContent = status;
}

// Composition - orchestrates pure functions and side effects
async function handleUserStatusChange(userId, action, state, api) {
    const user = state.users.find(u => u.id === userId);
    const updated = calculateNewStatus(user, action); // Pure, testable
    await api.updateUser(userId, updated); // Injected dependency
    updateUserInDOM(userId, updated.status); // Isolated side effect
    return updated;
}
```

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
5. **Maximum 500 lines per module** - Split further if needed
6. **Prioritize pure functions** - Separate logic from side effects
7. **Enable unit testing** - Functions should be testable in isolation

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

**Pure functions module (preferred for testability):**

```javascript
// utils.js - Pure utility functions (MOST TESTABLE)
(function(window) {
    'use strict';
    
    if (!window.MyApp) {
        window.MyApp = {};
    }
    
    // Pure functions - no dependencies, no state, easy to test
    window.MyApp.utils = {
        // Pure: always returns same output for same input
        add: function(a, b) {
            return a + b;
        },
        
        // Pure: no side effects, deterministic
        formatCurrency: function(amount) {
            return '$' + amount.toFixed(2);
        },
        
        // Pure: transforms data without mutation
        filterActive: function(items) {
            return items.filter(item => item.status === 'active');
        }
    };
    
})(window);
```

**State module (immutable patterns):**

```javascript
// state.js - State management with immutability
(function(window) {
    'use strict';
    
    if (!window.MyApp) {
        window.MyApp = {};
    }
    
    // Factory returns state object with pure methods
    window.MyApp.createState = function(initialData) {
        let data = initialData || {};
        
        return {
            // Pure getter - no side effects
            get: function(key) {
                return data[key];
            },
            
            // Returns NEW state, doesn't mutate
            set: function(key, value) {
                const newData = { ...data, [key]: value };
                data = newData;
                return newData;
            },
            
            // Pure: returns copy, doesn't expose internal state
            getAll: function() {
                return { ...data };
            }
        };
    };
    
})(window);
```

**Methods module (dependency injection for testability):**

```javascript
// methods.js - Business logic with injected dependencies
(function(window) {
    'use strict';
    
    // Create methods that accept dependencies (easier to test)
    window.MyApp.createMethods = function(deps) {
        // deps = { state, api, logger, utils }
        
        return {
            // Pure business logic - testable without mocks
            calculateTotal: function(items) {
                return items.reduce((sum, item) => sum + item.price, 0);
            },
            
            // Orchestration with injected dependencies
            processOrder: async function(order) {
                const total = this.calculateTotal(order.items);
                const formatted = deps.utils.formatCurrency(total);
                
                deps.logger.info('Processing order', { total: formatted });
                
                const result = await deps.api.submitOrder({
                    ...order,
                    total: total
                });
                
                deps.state.set('lastOrder', result);
                return result;
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

**Critical: Separate Pure from Impure**

For each functional area, separate:
1. **Pure functions** (utils, calculations, transformations) - MOST TESTABLE
2. **Stateful operations** (state management)
3. **Side effects** (DOM, API, I/O) - NEEDS MOCKING

**Utilities Module (utils.js)** - MAX 400 lines:
- Pure helper functions (no side effects)
- Data transformations
- Calculations and algorithms
- Formatters and validators
- String/array/object utilities
- **All functions MUST be pure and independently testable**

**State Management (state.js)** - MAX 300 lines:
- Data structures
- Immutable getters/setters
- State initialization
- State selectors (pure functions that derive data)
- **Prefer immutable updates**

**DOM Operations (dom.js)** - MAX 400 lines:
- Element creation (pure functions that return config)
- DOM queries (minimal, focused)
- DOM updates (side effects isolated)
- **Separate what to render (pure) from rendering (side effect)**

**Event Handlers (events.js)** - MAX 400 lines:
- Event registration
- Event handlers (orchestrate pure functions)
- Event delegation
- **Keep handlers thin, delegate to pure business logic**

**API Operations (api.js)** - MAX 400 lines:
- API client configuration
- Request builders (pure functions)
- Response transformers (pure functions)
- Actual fetch calls (side effects)
- **Separate request building from request execution**

**Business Logic (operations.js, validation.js)** - MAX 400 lines each:
- **MUST BE PURE FUNCTIONS**
- Validation rules (pure predicates)
- Business calculations
- Data transformations
- Workflow logic
- **Zero side effects - easy to test**
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

## File Size Guidelines (STRICT LIMITS)

Target sizes after refactoring:
- **Main entry file**: 50-100 lines (coordination only)
- **Module files**: 200-400 lines each (MAXIMUM 500 lines)
- **Utility files**: 100-300 lines each (pure functions)
- **Any file > 500 lines**: MUST be split further

**If a module exceeds 500 lines:**
1. Identify sub-concerns within the module
2. Extract into separate focused modules
3. Create a sub-namespace if needed (e.g., `MyApp.nodes.create`, `MyApp.nodes.update`)

## Testability Patterns

### Pattern 1: Pure Function Module (EASIEST TO TEST)

```javascript
// geometry-utils.js - All pure functions, zero dependencies
(function(window) {
    'use strict';
    
    if (!window.MyApp) window.MyApp = {};
    
    window.MyApp.geometry = {
        // Pure: easy to test with simple assertions
        distance: function(x1, y1, x2, y2) {
            const dx = x2 - x1;
            const dy = y2 - y1;
            return Math.sqrt(dx * dx + dy * dy);
        },
        
        // Pure: deterministic, no side effects
        isInside: function(px, py, x, y, width, height) {
            return px >= x && px <= x + width && 
                   py >= y && py <= y + height;
        },
        
        // Pure: transforms input to output
        clamp: function(value, min, max) {
            return Math.max(min, Math.min(max, value));
        }
    };
    
})(window);

// Easy to test:
// assert(MyApp.geometry.distance(0, 0, 3, 4) === 5);
// assert(MyApp.geometry.isInside(5, 5, 0, 0, 10, 10) === true);
```

### Pattern 2: Dependency Injection (TESTABLE WITH MOCKS)

```javascript
// api-client.js - Injected dependencies
(function(window) {
    'use strict';
    
    if (!window.MyApp) window.MyApp = {};
    
    // Factory accepts dependencies - easy to mock for testing
    window.MyApp.createApiClient = function(deps) {
        // deps = { baseUrl, fetch, logger }
        
        return {
            // Pure: builds request object
            buildRequest: function(method, path, body) {
                return {
                    url: deps.baseUrl + path,
                    method: method,
                    headers: { 'Content-Type': 'application/json' },
                    body: body ? JSON.stringify(body) : undefined
                };
            },
            
            // Impure: uses injected fetch (can mock in tests)
            request: async function(method, path, body) {
                const req = this.buildRequest(method, path, body);
                deps.logger.info('API request', { method, path });
                
                try {
                    const response = await deps.fetch(req.url, req);
                    return await response.json();
                } catch (error) {
                    deps.logger.error('API error', error);
                    throw error;
                }
            }
        };
    };
    
})(window);

// Test with mocks:
// const mockFetch = async () => ({ json: async () => ({ success: true }) });
// const mockLogger = { info: () => {}, error: () => {} };
// const client = MyApp.createApiClient({ 
//     baseUrl: '/api', 
//     fetch: mockFetch, 
//     logger: mockLogger 
// });
```

### Pattern 3: Separate Logic from Effects

```javascript
// validation.js - Pure business logic
(function(window) {
    'use strict';
    
    if (!window.MyApp) window.MyApp = {};
    
    // All pure - trivial to test
    window.MyApp.validation = {
        isValidEmail: function(email) {
            return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
        },
        
        isValidPassword: function(password) {
            return password.length >= 8 && /[A-Z]/.test(password);
        },
        
        // Pure: returns validation result object
        validateUser: function(user) {
            const errors = [];
            
            if (!this.isValidEmail(user.email)) {
                errors.push({ field: 'email', message: 'Invalid email' });
            }
            
            if (!this.isValidPassword(user.password)) {
                errors.push({ field: 'password', message: 'Password too weak' });
            }
            
            return {
                isValid: errors.length === 0,
                errors: errors
            };
        }
    };
    
})(window);

// form-handler.js - Side effects separated
(function(window) {
    'use strict';
    
    window.MyApp.createFormHandler = function(deps) {
        // deps = { validation, dom, api }
        
        return {
            // Orchestrates: pure validation + side effects
            handleSubmit: async function(formData) {
                // Pure validation (easy to test separately)
                const result = deps.validation.validateUser(formData);
                
                if (!result.isValid) {
                    // Side effect: DOM update
                    deps.dom.showErrors(result.errors);
                    return;
                }
                
                // Side effect: API call
                await deps.api.createUser(formData);
                
                // Side effect: DOM update
                deps.dom.showSuccess();
            }
        };
    };
    
})(window);
```

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
- [ ] Identify pure vs impure functions
- [ ] Check for side effects
- [ ] Identify shared state
- [ ] Document public APIs
- [ ] Plan module split to stay under 500 lines each

## Checklist After Refactoring

- [ ] All functionality works as before
- [ ] No circular dependencies
- [ ] Clear module boundaries
- [ ] Each file has single responsibility
- [ ] **Every module is ≤ 500 lines** (STRICT)
- [ ] Pure functions separated from side effects
- [ ] Functions are testable in isolation
- [ ] Dependencies are explicit (injected, not global)
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
- `path/to/file/utils.js` (ZZZ lines) - Pure utility functions ✅ TESTABLE
- `path/to/file/state.js` (AAA lines) - State management
- `path/to/file/operations.js` (BBB lines) - Business logic ✅ PURE FUNCTIONS
- `path/to/file/dom.js` (CCC lines) - DOM operations (side effects)
- `path/to/file/api.js` (DDD lines) - API calls (side effects)
- ... (list all modules with line counts)

**Size validation**: ✅ All modules ≤ 500 lines

**Testability**:
- Pure function modules: utils.js, operations.js (no mocks needed)
- Stateful modules: state.js (simple test setup)
- Side effect modules: dom.js, api.js (require mocks/stubs)

**Loading approach**: [Browser script tags / ES6 modules]

**Original file changes**:
- Replaced entire content with module loader/re-exporter
- [Browser] Added documentation for script tag order
- [Modules] Now re-exports from ./file/ directory

**HTML template updates** (if browser version):
```html
<!-- Load in dependency order, pure functions first -->
<script src="/path/to/file/utils.js"></script>        <!-- Pure, no deps -->
<script src="/path/to/file/state.js"></script>        <!-- Depends on utils -->
<script src="/path/to/file/operations.js"></script>   <!-- Pure logic -->
<script src="/path/to/file/dom.js"></script>          <!-- Side effects -->
<script src="/path/to/file/api.js"></script>          <!-- Side effects -->
<script src="/path/to/file/index.js"></script>        <!-- Orchestration -->
<script src="/path/to/file.js"></script>              <!-- Main entry -->
```

**Dependencies**: 
- utils.js: none (pure functions)
- state.js: utils
- operations.js: utils (pure functions only)
- dom.js: state, utils
- api.js: state, utils
- index.js: all above

**Testability improvements**:
- XX% of code is now pure functions (easy unit tests)
- Side effects isolated to Y modules (mockable)
- Dependency injection enables test doubles

**Breaking changes**: [None / List any]
```

## ✅ Complete Refactoring Checklist

**When refactoring browser JavaScript, you MUST complete ALL of these steps:**

### Module Creation & Size Limits
- [ ] Created module directory: `original-file/`
- [ ] Created README.md in module directory
- [ ] Split code into focused modules (state, methods, handlers, etc.)
- [ ] **STRICT: Each module ≤ 500 lines** (split further if needed)
- [ ] Separated pure functions from side effects
- [ ] Applied functional programming principles

### Functional Programming & Testability
- [ ] **Identified and separated pure functions** (utils, calculations)
- [ ] **Extracted side effects** (DOM, API, I/O) into separate modules
- [ ] **Applied dependency injection** (pass deps as params, not globals)
- [ ] **Made business logic pure** (no side effects in core logic)
- [ ] **Used immutable patterns** where applicable
- [ ] Each function has single responsibility
- [ ] Functions are independently testable

### Browser Compatibility (NO BUILD STEP)
- [ ] **CRITICAL:** Converted ALL ES6 `export` statements to IIFE pattern
- [ ] Each module uses `(function(window) { ... })(window)`
- [ ] All modules attach to shared namespace (e.g., `window.MyApp`)
- [ ] No `import` or `export` keywords anywhere in modules
- [ ] Factory functions return objects/methods: `window.MyApp.createSomething = function() { ... }`

### Main File Replacement
- [ ] **CRITICAL:** Backed up original file: `original-file.monolithic.backup.js`
- [ ] **CRITICAL:** Replaced original file content with minimal loader (not just added header comment)
- [ ] New main file is 50-100 lines (just coordination/documentation)
- [ ] Verified file size reduction: `wc -l original-file.js` shows dramatic decrease

### HTML Template Updates
- [ ] **CRITICAL:** Added script tags for ALL modules in dependency order
- [ ] Pure function modules loaded first (no dependencies)
- [ ] Script tags load BEFORE main file
- [ ] Correct loading order documented in README
- [ ] Template tested: modules load without errors

### Verification & Testing
- [ ] Browser console shows no "Unexpected token 'export'" errors
- [ ] Namespace is populated: `console.log(window.MyApp)` shows all functions
- [ ] Main entry point available: `console.log(window.myFunction)` works
- [ ] Application functions correctly with modular code
- [ ] No regression in functionality
- [ ] **Pure functions can be tested without mocks** (`utils`, `operations`)
- [ ] **Side effect modules use dependency injection** (testable with mocks)

### Documentation
- [ ] README.md created in module directory with:
  - [ ] Module structure and line counts
  - [ ] Loading order with dependencies
  - [ ] Testability notes (pure vs impure modules)
  - [ ] Migration status checklist included
  - [ ] Example test patterns for pure functions

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
2. **Identify pure functions** that can be extracted first (utils, calculations)
3. Create module directory: `workflow-designer/`
4. Split into modules with **STRICT 500-line limit**:
   - utils.js (pure functions - MOST TESTABLE)
   - state.js (immutable state management)
   - operations.js (pure business logic)
   - dom.js, api.js (side effects - dependency injection)
   - etc.
5. Choose pattern based on context (browser IIFE vs ES6 modules)
6. Create modules with appropriate pattern
7. **Separate pure from impure** in each module
8. Update main file to minimal loader
9. Update HTML template with script tags in dependency order
10. Provide summary with:
    - File sizes (all ≤ 500 lines)
    - Testability breakdown (% pure functions)
    - Loading order and dependencies

## 🎯 Key Success Criteria

**Your refactoring is successful ONLY if:**

1. ✅ **Every module ≤ 500 lines** (no exceptions)
2. ✅ **Pure functions separated** from side effects
3. ✅ **Utils module is 100% pure** functions (no DOM, no API, no state mutations)
4. ✅ **Business logic is testable** without mocks
5. ✅ **Dependencies are injected**, not global
6. ✅ **Original file is replaced** with minimal loader (< 100 lines)
7. ✅ **HTML template updated** with all script tags
8. ✅ **No regression** in functionality

**Remember:** The goal is not just to split files, but to create **testable, maintainable code** with clear separation of concerns and functional programming principles.

