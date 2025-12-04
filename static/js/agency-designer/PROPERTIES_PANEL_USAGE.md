# Properties Panel - Generic Usage Guide

## Overview

The Properties Panel has been refactored to be a generic, reusable component that can display and edit properties for any type of data object. It supports custom save handlers, field types, and actions.

## Basic Usage

```javascript
window.PropertiesPanel.showProperties({
    title: 'My Item Properties',
    icon: 'cog',  // FontAwesome icon (without 'fa-' prefix)
    iconColor: 'primary',  // Bulma color class
    data: myDataObject,  // The object being edited
    
    fields: [
        {
            key: 'name',
            label: 'Name',
            type: 'text',
            placeholder: 'Enter name...'
        },
        {
            key: 'description',
            label: 'Description',
            type: 'textarea',
            rows: 4
        }
    ],
    
    buttons: [
        {
            action: 'save',
            label: 'Save',
            icon: 'save',
            class: 'is-primary'
        },
        {
            action: 'close',
            label: 'Close',
            icon: 'times'
        }
    ],
    
    // Callback when field changes
    onUpdate: (field, value) => {
        console.log(`Field ${field} changed to:`, value);
        // Update your data model here
    },
    
    // Callback when save button clicked
    onSave: async () => {
        console.log('Saving data...');
        // Implement save logic here
        await saveToBackend(myDataObject);
        window.showNotification('Saved successfully', 'success');
    },
    
    // Callback when delete button clicked
    onDelete: () => {
        console.log('Deleting item...');
        // Implement delete logic
    }
});
```

## Field Types

### Text Field
```javascript
{
    key: 'name',
    label: 'Name',
    type: 'text',
    placeholder: 'Enter name...',
    help: 'Optional help text'
}
```

### Static/Read-only Field
```javascript
{
    key: 'id',
    label: 'ID',
    type: 'static'  // Displays as read-only
}
```

### Textarea
```javascript
{
    key: 'description',
    label: 'Description',
    type: 'textarea',
    rows: 6,
    placeholder: 'Enter description...'
}
```

### Select/Dropdown
```javascript
{
    key: 'status',
    label: 'Status',
    type: 'select',
    options: [
        { value: 'active', label: 'Active' },
        { value: 'inactive', label: 'Inactive' }
    ]
}
```

### Tags
```javascript
{
    key: 'tags',
    label: 'Tags',
    type: 'tags'
}
```

### Badge (Display only)
```javascript
{
    key: 'count',
    label: 'Item Count',
    type: 'badge',
    icon: 'folder',
    color: 'primary',
    format: (value) => `${value} items`,
    help: 'Total number of items'
}
```

### Custom Field
```javascript
{
    key: 'custom',
    label: 'Custom Field',
    type: 'custom',
    render: (value, data) => {
        return `<div class="custom-content">${value}</div>`;
    }
}
```

## Button Actions

### Built-in Actions
- `save` - Triggers `onSave` callback
- `delete` - Triggers `onDelete` callback (with confirmation)
- `close` - Clears the properties panel
- `chat` - Switches to chat tab

### Custom Actions
```javascript
buttons: [
    {
        action: 'custom-action',
        label: 'Custom Button',
        icon: 'star',
        class: 'is-info'
    }
],

onAction: (action) => {
    if (action === 'custom-action') {
        // Handle custom action
    }
}
```

## Real Example: Deliverable Node

```javascript
window.PropertiesPanel.showProperties({
    title: `Deliverable: ${node.name}`,
    icon: node.type === 'folder' ? 'folder' : 'file-lines',
    iconColor: node.type === 'folder' ? 'info' : 'primary',
    data: node,
    
    fields: [
        {
            key: 'name',
            label: 'Name',
            type: 'text',
            placeholder: 'Node name'
        },
        {
            key: 'description',
            label: 'Description',
            type: 'text'
        },
        {
            key: 'prompt_instructions',
            label: 'AI Prompt Instructions',
            type: 'textarea',
            rows: 6,
            help: 'Instructions for AI agents'
        },
        {
            key: 'path',
            label: 'Path',
            type: 'static',
            help: 'Full path in tree'
        }
    ],
    
    buttons: [
        {
            action: 'save',
            label: 'Save',
            icon: 'save',
            class: 'is-primary'
        },
        {
            action: 'delete',
            label: 'Delete',
            icon: 'trash',
            class: 'is-danger is-light'
        },
        {
            action: 'close',
            label: 'Close',
            icon: 'times'
        }
    ],
    
    onUpdate: (field, value) => {
        // Update tree node in Alpine.js
        alpineData.updateNodeField(node.id, field, value);
    },
    
    onSave: async () => {
        // Save to backend
        await alpineData.onSave();
        window.showNotification('Saved!', 'success');
    },
    
    onDelete: () => {
        alpineData.deleteNode(node.id);
        window.PropertiesPanel.clear();
    }
});
```

## Configuration Options

| Option | Type | Description |
|--------|------|-------------|
| `title` | string | Panel title |
| `icon` | string | FontAwesome icon (without 'fa-') |
| `iconColor` | string | Bulma color class |
| `data` | object | Data object being edited |
| `fields` | array | Field definitions |
| `buttons` | array | Button definitions |
| `autoSwitchTab` | boolean | Auto-switch to properties tab (default: true) |
| `onUpdate` | function | Called when field changes: `(field, value)` |
| `onSave` | function | Called when save button clicked |
| `onDelete` | function | Called when delete button clicked |
| `onRemoveTag` | function | Called when tag removed: `(field, tag)` |
| `onAction` | function | Called for custom actions: `(action)` |

## Legacy Methods

The following methods are maintained for backward compatibility:

- `showWorkItemProperties(workItem)` - Shows work item properties
- `showDeliverableNodeProperties(node)` - Shows deliverable node properties
- `updateDeliverableNodeField(nodeId, field, value)` - Updates deliverable field
- `selectDeliverableNode(nodeId)` - Selects and shows deliverable
- `deleteDeliverableNode(nodeId)` - Deletes deliverable node
- `saveDeliverableNode(nodeId)` - Saves deliverable node
- `countDeliverables(deliverables)` - Counts deliverables recursively
- `updateWorkItemField(field, value)` - Updates work item field
- `removeTag(tag)` - Removes tag from work item
- `switchToChat()` - Switches to chat tab
- `clear()` - Clears the panel

## Integration with Alpine.js

When integrating with Alpine.js components, pass the save function via the component's initialization:

```javascript
// Initialize deliverable tree with save callback
window.initDeliverableTreeBuilder(
    agencyId, 
    workItemCode, 
    existingDeliverables,
    async () => {
        // Your save logic here
        await window.saveWorkItemFromEditor();
    }
);

// The callback will be accessible in Alpine.js component as onSave property
// Properties panel will call alpineData.onSave() when save button is clicked
```

Example with custom save function:
```javascript
window.initDeliverableTreeBuilder(
    'UC-EXAMPLE-001',
    'WI-001',
    [],
    async () => {
        console.log('Saving deliverables...');
        const data = window.getDeliverablesStructuredData();
        await fetch('/api/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ deliverables: data })
        });
    }
);
```
