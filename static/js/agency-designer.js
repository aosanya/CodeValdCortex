// Agency Designer - Modular Version Entry Point
// This file loads the modular agency designer components

// Since browsers don't fully support ES6 modules without bundling,
// we'll create a simple loader that imports all functionality

// Import main module which coordinates everything
import('./agency-designer/main.js').then(() => {
    console.log('Agency Designer modules loaded successfully');
}).catch(error => {
    console.error('Error loading Agency Designer modules:', error);
    console.error('Make sure all module files are present in the agency-designer/ directory');
});