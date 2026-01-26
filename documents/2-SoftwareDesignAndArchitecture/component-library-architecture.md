# Component Library Architecture

## Overview

**CodeValdCortex** is establishing a reusable component library to promote consistency, reduce code duplication, and improve maintainability across the application. This library follows templ-first architecture principles and provides type-safe, server-rendered components.

**Location**: `/workspaces/CodeValdCortex/internal/web/components/`

**Motivation**:
- Reduce repository bloat by maximizing code reuse
- Ensure consistent UI/UX across all pages
- Enable rapid development with validated, accessible form components
- Single source of truth for common UI patterns

## Directory Structure

```
internal/web/components/
├── README.md                    # Component library documentation
├── layout.templ                 # Page layouts (LayoutWithAgency, etc.)
├── navbar.templ                 # Navigation components
├── status_bar.templ             # Status bar components
├── vscode_status_bar.templ      # VS Code-style status bars
├── forms/                       # Form components with validation (NEW)
│   ├── README.md               # Forms documentation
│   ├── text_field.templ        # Single-line text input with validation
│   ├── textarea_field.templ    # Multi-line textarea with validation
│   ├── select_field.templ      # Dropdown select with validation
│   ├── checkbox_field.templ    # Checkbox with validation
│   ├── radio_field.templ       # Radio button group with validation
│   └── validation.templ        # Shared validation message components
├── cards/                       # Card components (FUTURE)
├── buttons/                     # Button components (FUTURE)
└── tables/                      # Table components (FUTURE)
```

## Design Principles

### 1. Templ-First Architecture
- **All HTML in `.templ` files** - Never generate HTML in Go handlers or JavaScript
- **Type-safe** - Leverage Go's type system for compile-time safety
- **Server-rendered** - Components render on server, reducing client-side complexity

### 2. Bulma CSS Framework
- Use Bulma's built-in classes whenever possible
- Minimize custom CSS
- Maintain consistent styling across components

### 3. Accessibility First
- Proper ARIA labels
- Keyboard navigation support
- Screen reader compatibility
- Error messages associated with form fields

### 4. Validation Support
- Inline validation messages next to fields
- Clear error states with visual indicators
- Help text to guide users
- Real-time validation feedback (client-side)
- Server-side validation integration

### 5. Composability
- Small, focused components that do one thing well
- Composable building blocks for complex forms
- Consistent API across similar components

## Forms Component Library

### Goals

1. **Inline Validation** - Error messages appear directly below form fields
2. **Visual Feedback** - Red borders, icons, and error text for invalid fields
3. **Accessibility** - Proper labeling and ARIA attributes
4. **Consistency** - All form fields follow the same validation pattern
5. **Type Safety** - Go structs for field configuration

### Component API Design

#### TextAreaField Component

**Purpose**: Multi-line text input with integrated validation support

**Template Signature**:
```go
templ TextAreaField(config TextAreaFieldConfig)
```

**Configuration Struct**:
```go
type TextAreaFieldConfig struct {
    // Field Identification
    ID          string  // HTML id attribute (required)
    Name        string  // HTML name attribute (optional, defaults to ID)
    
    // Label and Help
    Label       string  // Field label text (required)
    HelpText    string  // Help text below field (optional)
    
    // Input Properties
    Placeholder string  // Placeholder text (optional)
    Rows        int     // Number of visible rows (default: 6)
    MaxLength   int     // Maximum character length (optional)
    Value       string  // Initial value (optional)
    
    // Validation
    Required    bool    // Is field required?
    ErrorMsg    string  // Error message to display (optional)
    
    // Styling
    Classes     string  // Additional CSS classes (optional)
    Style       string  // Inline styles (optional)
    
    // Accessibility
    AriaLabel   string  // ARIA label override (optional)
}
```

**Usage Example**:
```templ
@forms.TextAreaField(forms.TextAreaFieldConfig{
    ID:          "workflow-description-editor",
    Label:       "Description",
    HelpText:    "Explain what this workflow does and when it should be used.",
    Placeholder: "Describe the workflow purpose and what it accomplishes...",
    Rows:        6,
    Required:    true,
    Style:       "font-family: monospace; font-size: 14px;",
})
```

**Rendered HTML Structure**:
```html
<div class="field">
    <label class="label" for="workflow-description-editor">
        Description
        <span class="has-text-danger">*</span> <!-- if Required -->
    </label>
    <div class="control">
        <textarea 
            class="textarea"
            id="workflow-description-editor"
            name="workflow-description-editor"
            placeholder="Describe the workflow purpose..."
            rows="6"
            aria-required="true"
            aria-describedby="workflow-description-editor-help workflow-description-editor-error"
            style="font-family: monospace; font-size: 14px;">
        </textarea>
    </div>
    <p class="help" id="workflow-description-editor-help">
        Explain what this workflow does and when it should be used.
    </p>
    <!-- Error message (hidden by default, shown via JS) -->
    <p class="help is-danger is-hidden" id="workflow-description-editor-error" role="alert">
        <!-- Error message injected here -->
    </p>
</div>
```

#### TextField Component

**Configuration Struct**:
```go
type TextFieldConfig struct {
    ID          string
    Name        string
    Label       string
    HelpText    string
    Placeholder string
    MaxLength   int
    Value       string
    Required    bool
    ErrorMsg    string
    Type        string  // "text", "email", "url", etc.
    Classes     string
    AriaLabel   string
}
```

**Usage Example**:
```templ
@forms.TextField(forms.TextFieldConfig{
    ID:          "workflow-name-editor",
    Label:       "Workflow Name",
    HelpText:    "A clear, descriptive name for this workflow.",
    Placeholder: "e.g., User Onboarding Process",
    MaxLength:   200,
    Required:    true,
})
```

### Validation Pattern

#### Client-Side Validation (JavaScript)

**File**: `static/js/form-validation.js`

```javascript
// Form validation utilities
window.formValidation = {
    /**
     * Show validation error for a field
     * @param {string} fieldId - ID of the form field
     * @param {string} message - Error message to display
     */
    showError: function(fieldId, message) {
        console.log('[FORM-VALIDATION] Showing error for:', fieldId, message);
        
        const field = document.getElementById(fieldId);
        const errorElement = document.getElementById(`${fieldId}-error`);
        
        if (field) {
            field.classList.add('is-danger');
            field.setAttribute('aria-invalid', 'true');
        }
        
        if (errorElement) {
            errorElement.textContent = message;
            errorElement.classList.remove('is-hidden');
        }
    },
    
    /**
     * Clear validation error for a field
     * @param {string} fieldId - ID of the form field
     */
    clearError: function(fieldId) {
        console.log('[FORM-VALIDATION] Clearing error for:', fieldId);
        
        const field = document.getElementById(fieldId);
        const errorElement = document.getElementById(`${fieldId}-error`);
        
        if (field) {
            field.classList.remove('is-danger');
            field.removeAttribute('aria-invalid');
        }
        
        if (errorElement) {
            errorElement.textContent = '';
            errorElement.classList.add('is-hidden');
        }
    },
    
    /**
     * Validate a required field
     * @param {string} fieldId - ID of the form field
     * @param {string} fieldName - Human-readable field name
     * @returns {boolean} True if valid, false otherwise
     */
    validateRequired: function(fieldId, fieldName) {
        const field = document.getElementById(fieldId);
        if (!field) return false;
        
        const value = field.value.trim();
        
        if (!value) {
            this.showError(fieldId, `${fieldName} is required`);
            field.focus();
            return false;
        }
        
        this.clearError(fieldId);
        return true;
    },
    
    /**
     * Validate multiple fields at once
     * @param {Array<{id: string, name: string, required: boolean}>} fields
     * @returns {boolean} True if all valid, false otherwise
     */
    validateFields: function(fields) {
        let allValid = true;
        
        for (const field of fields) {
            if (field.required) {
                if (!this.validateRequired(field.id, field.name)) {
                    allValid = false;
                    break; // Stop at first error
                }
            }
        }
        
        return allValid;
    }
};
```

**Updated Workflow Save Function**:
```javascript
window.saveWorkflowFromEditor = function () {
    console.log('[MVP-054] 🔵 saveWorkflowFromEditor CALLED');
    
    // Define fields to validate
    const fields = [
        { id: 'workflow-name-editor', name: 'Workflow Name', required: true },
        { id: 'workflow-description-editor', name: 'Description', required: true }
    ];
    
    // Validate all fields
    if (!window.formValidation.validateFields(fields)) {
        console.log('[MVP-054] ⚠️ Validation failed');
        return;
    }
    
    console.log('[MVP-054] ✅ Validation passed');
    
    // Continue with save...
    const name = document.getElementById('workflow-name-editor').value.trim();
    const description = document.getElementById('workflow-description-editor').value.trim();
    const version = document.getElementById('workflow-version-editor').value.trim();
    
    const data = { name, description, version, ... };
    
    window.saveEntity('workflows', workflowEditorState.mode, 
                     workflowEditorState.workflowId, data, 
                     'save-workflow-btn', () => {
        cancelWorkflowEdit();
        loadWorkflows();
    });
}
```

## Implementation Roadmap

### Phase 1: Forms Component Library (Current - MVP-054)
- [x] Document architecture
- [ ] Create `internal/web/components/forms/` directory
- [ ] Implement `TextAreaField` component
- [ ] Implement `TextField` component
- [ ] Create `static/js/form-validation.js`
- [ ] Update workflow editor to use new components
- [ ] Test validation flow end-to-end

### Phase 2: Additional Form Components (MVP-055)
- [ ] SelectField component
- [ ] CheckboxField component
- [ ] RadioField component
- [ ] DateField component
- [ ] NumberField component

### Phase 3: Advanced Validation (MVP-056)
- [ ] Pattern validation (regex)
- [ ] Min/max length validation
- [ ] Custom validation functions
- [ ] Async validation (API calls)
- [ ] Form-level validation summary

### Phase 4: Other Component Categories (Future)
- [ ] Cards component library
- [ ] Buttons component library
- [ ] Tables component library
- [ ] Modals component library
- [ ] Navigation component library

## Migration Strategy

### Converting Existing Forms

**Before** (Duplicate HTML in every template):
```templ
<div class="field">
    <label class="label">Description</label>
    <div class="control">
        <textarea class="textarea" id="workflow-description-editor" ...></textarea>
    </div>
    <p class="help">Explain what this workflow does...</p>
</div>
```

**After** (Reusable component):
```templ
@forms.TextAreaField(forms.TextAreaFieldConfig{
    ID:       "workflow-description-editor",
    Label:    "Description",
    HelpText: "Explain what this workflow does...",
    Required: true,
})
```

**Benefits**:
- 70% reduction in template code
- Consistent validation UX
- Single source of truth
- Easy to update all forms at once

### Target Forms for Migration

1. **Agency Designer Forms** (High Priority)
   - Workflow editor
   - Goal editor
   - Work item editor
   - Role editor

2. **Instance Management Forms** (Medium Priority)
   - Instance creation
   - Configuration forms

3. **Policy Wizard Forms** (Medium Priority)
   - Policy configuration

4. **Workbench Forms** (Low Priority)
   - Issue creation/editing

## Testing Strategy

### Component Tests
- Unit tests for validation logic
- Integration tests for form submission
- Accessibility tests (ARIA compliance)

### Visual Regression Tests
- Screenshot comparisons for form states
- Error state rendering
- Focus states

### Browser Compatibility
- Test in Chrome, Firefox, Safari, Edge
- Mobile responsive testing
- Keyboard navigation testing

## Documentation Requirements

Each component must include:
1. **Purpose** - What the component does
2. **Configuration** - All available options
3. **Usage Examples** - Common use cases
4. **Accessibility Notes** - ARIA requirements
5. **Validation Rules** - How validation works
6. **Browser Support** - Compatibility information

## Success Metrics

1. **Code Reduction**: 50%+ reduction in form-related template code
2. **Consistency**: 100% of forms use validated components
3. **Accessibility**: Pass WCAG 2.1 AA standards
4. **Developer Experience**: <5 minutes to create a new validated form
5. **Bug Reduction**: 80% reduction in form-related bugs

## Related Documentation

- [Frontend Architecture](./frontend-architecture-updated.md)
- [Coding Instructions](../../.github/copilot-instructions.md)
- [Component Reuse Rules](../../.github/instructions/rules.instructions.md)

---

**Status**: 🔄 In Progress (MVP-054)  
**Last Updated**: December 4, 2025  
**Owner**: Development Team
