# Forms Component Library

## Overview

This directory contains reusable form components with built-in validation support. All components follow Bulma CSS framework conventions and provide consistent validation UX across the application.

## Components

### TextAreaField
Multi-line text input with validation support.

**File**: `textarea_field.templ`

**Usage**:
```templ
@forms.TextAreaField(forms.TextAreaFieldConfig{
    ID:          "description",
    Label:       "Description",
    HelpText:    "Enter a detailed description",
    Placeholder: "Type here...",
    Rows:        6,
    Required:    true,
})
```

### TextField
Single-line text input with validation support.

**File**: `text_field.templ`

**Usage**:
```templ
@forms.TextField(forms.TextFieldConfig{
    ID:          "name",
    Label:       "Name",
    HelpText:    "Enter a name",
    Placeholder: "e.g., My Item",
    Required:    true,
    MaxLength:   200,
})
```

## Validation Pattern

All form components include:
1. **Error message placeholder** - Hidden by default, shown via JavaScript
2. **ARIA attributes** - For accessibility (aria-required, aria-invalid, aria-describedby)
3. **Visual indicators** - Red border (is-danger class) when invalid
4. **Help text** - Always visible guidance for users

## JavaScript Integration

Use `window.formValidation` utilities (from `static/js/form-validation.js`):

```javascript
// Show error
window.formValidation.showError('field-id', 'Error message');

// Clear error
window.formValidation.clearError('field-id');

// Validate required field
const isValid = window.formValidation.validateRequired('field-id', 'Field Name');

// Validate multiple fields
const isValid = window.formValidation.validateFields([
    { id: 'name', name: 'Name', required: true },
    { id: 'description', name: 'Description', required: true }
]);
```

## Design Principles

1. **Consistent Structure** - All form fields follow the same HTML structure
2. **Accessibility First** - Proper ARIA labels and associations
3. **Progressive Enhancement** - Works without JavaScript, enhanced with it
4. **Bulma Native** - Uses Bulma's form classes and validation styles
5. **Type Safe** - Go structs for configuration

## Adding New Components

When creating a new form component:

1. Define a configuration struct in the package
2. Create the `.templ` file with the component
3. Include error message placeholder
4. Add ARIA attributes
5. Document in this README
6. Add usage examples
7. Test validation flow

## Related Files

- **Architecture**: `/documents/2-SoftwareDesignAndArchitecture/component-library-architecture.md`
- **Validation JS**: `/static/js/form-validation.js`
- **Coding Rules**: `/.github/instructions/rules.instructions.md`
