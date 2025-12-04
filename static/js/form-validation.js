/**
 * Form Validation Utilities
 * 
 * Provides consistent client-side validation for form components.
 * Works with the forms component library in internal/web/components/forms/
 * 
 * @module form-validation
 */

(function() {
    'use strict';

    /**
     * Form validation utilities namespace
     * @namespace
     */
    window.formValidation = {
        /**
         * Show validation error for a field
         * @param {string} fieldId - ID of the form field
         * @param {string} message - Error message to display
         */
        showError: function(fieldId, message) {
            console.log('[FORM-VALIDATION] Showing error for:', fieldId, '-', message);
            
            const field = document.getElementById(fieldId);
            const errorElement = document.getElementById(`${fieldId}-error`);
            
            if (field) {
                // Add error styling
                field.classList.add('is-danger');
                field.setAttribute('aria-invalid', 'true');
            }
            
            if (errorElement) {
                // Show error message
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
                // Remove error styling
                field.classList.remove('is-danger');
                field.removeAttribute('aria-invalid');
            }
            
            if (errorElement) {
                // Hide error message
                errorElement.textContent = '';
                errorElement.classList.add('is-hidden');
            }
        },
        
        /**
         * Clear all validation errors in a form or container
         * @param {string} containerId - ID of the container element
         */
        clearAllErrors: function(containerId) {
            console.log('[FORM-VALIDATION] Clearing all errors in:', containerId);
            
            const container = containerId ? document.getElementById(containerId) : document;
            if (!container) return;
            
            // Clear all danger classes from fields
            const fields = container.querySelectorAll('.is-danger');
            fields.forEach(field => {
                field.classList.remove('is-danger');
                field.removeAttribute('aria-invalid');
            });
            
            // Hide all error messages
            const errorElements = container.querySelectorAll('[id$="-error"]');
            errorElements.forEach(el => {
                el.textContent = '';
                el.classList.add('is-hidden');
            });
        },
        
        /**
         * Validate a required field
         * @param {string} fieldId - ID of the form field
         * @param {string} fieldName - Human-readable field name for error message
         * @returns {boolean} True if valid, false otherwise
         */
        validateRequired: function(fieldId, fieldName) {
            const field = document.getElementById(fieldId);
            if (!field) {
                console.error('[FORM-VALIDATION] Field not found:', fieldId);
                return false;
            }
            
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
         * Validate minimum length
         * @param {string} fieldId - ID of the form field
         * @param {string} fieldName - Human-readable field name
         * @param {number} minLength - Minimum required length
         * @returns {boolean} True if valid, false otherwise
         */
        validateMinLength: function(fieldId, fieldName, minLength) {
            const field = document.getElementById(fieldId);
            if (!field) return false;
            
            const value = field.value.trim();
            
            if (value.length < minLength) {
                this.showError(fieldId, `${fieldName} must be at least ${minLength} characters`);
                field.focus();
                return false;
            }
            
            this.clearError(fieldId);
            return true;
        },
        
        /**
         * Validate maximum length
         * @param {string} fieldId - ID of the form field
         * @param {string} fieldName - Human-readable field name
         * @param {number} maxLength - Maximum allowed length
         * @returns {boolean} True if valid, false otherwise
         */
        validateMaxLength: function(fieldId, fieldName, maxLength) {
            const field = document.getElementById(fieldId);
            if (!field) return false;
            
            const value = field.value.trim();
            
            if (value.length > maxLength) {
                this.showError(fieldId, `${fieldName} must be no more than ${maxLength} characters`);
                field.focus();
                return false;
            }
            
            this.clearError(fieldId);
            return true;
        },
        
        /**
         * Validate multiple fields at once
         * @param {Array<{id: string, name: string, required?: boolean, minLength?: number, maxLength?: number}>} fields - Array of field configurations
         * @returns {boolean} True if all valid, false otherwise
         */
        validateFields: function(fields) {
            console.log('[FORM-VALIDATION] Validating fields:', fields.map(f => f.id));
            
            let allValid = true;
            let firstInvalidField = null;
            
            for (const field of fields) {
                let fieldValid = true;
                
                // Required validation
                if (field.required) {
                    if (!this.validateRequired(field.id, field.name)) {
                        fieldValid = false;
                    }
                }
                
                // Min length validation (only if field has a value)
                if (fieldValid && field.minLength) {
                    const fieldElement = document.getElementById(field.id);
                    if (fieldElement && fieldElement.value.trim()) {
                        if (!this.validateMinLength(field.id, field.name, field.minLength)) {
                            fieldValid = false;
                        }
                    }
                }
                
                // Max length validation (only if field has a value)
                if (fieldValid && field.maxLength) {
                    const fieldElement = document.getElementById(field.id);
                    if (fieldElement && fieldElement.value.trim()) {
                        if (!this.validateMaxLength(field.id, field.name, field.maxLength)) {
                            fieldValid = false;
                        }
                    }
                }
                
                if (!fieldValid) {
                    allValid = false;
                    if (!firstInvalidField) {
                        firstInvalidField = document.getElementById(field.id);
                    }
                }
            }
            
            // Focus first invalid field
            if (firstInvalidField) {
                firstInvalidField.focus();
            }
            
            console.log('[FORM-VALIDATION] Validation result:', allValid ? 'PASSED' : 'FAILED');
            return allValid;
        },
        
        /**
         * Add real-time validation listener to a field
         * @param {string} fieldId - ID of the form field
         * @param {Function} validator - Validation function that returns boolean
         */
        addRealtimeValidation: function(fieldId, validator) {
            const field = document.getElementById(fieldId);
            if (!field) return;
            
            field.addEventListener('blur', function() {
                validator(fieldId);
            });
            
            field.addEventListener('input', function() {
                // Clear error on input to give immediate feedback
                if (field.classList.contains('is-danger')) {
                    window.formValidation.clearError(fieldId);
                }
            });
        }
    };
    
    console.log('[FORM-VALIDATION] Form validation utilities loaded');
})();
