// Dynamic form builder for node configuration fields

export class NodeConfigFormBuilder {
    constructor(nodeTypeConfigs) {
        this.nodeTypeConfigs = nodeTypeConfigs;
    }

    /**
     * Render form fields for a given node type into a container
     * @param {HTMLElement} container - The DOM element to render fields into
     * @param {string} nodeType - The type of node (e.g., 'input', 'blur', 'resize')
     * @param {string} idPrefix - Prefix for field IDs (e.g., 'config' or 'edit-config')
     * @param {Object} currentValues - Optional current values to populate fields with
     */
    renderFields(container, nodeType, idPrefix = 'config', currentValues = null) {
        container.innerHTML = '';

        const config = this.nodeTypeConfigs[nodeType];
        if (!config || !config.fields || config.fields.length === 0) {
            return;
        }

        // Fields is now an array, iterate directly to preserve order
        config.fields.forEach(fieldDef => {
            const { label, input } = this._createField(fieldDef.name, fieldDef, idPrefix, currentValues);
            container.appendChild(label);
            container.appendChild(input);
        });
    }

    /**
     * Create a single form field (label + input)
     * @private
     */
    _createField(fieldName, fieldDef, idPrefix, currentValues) {
        const label = document.createElement('label');
        label.setAttribute('for', `${idPrefix}-${fieldName}`);
        label.textContent = `${fieldName}${fieldDef.required ? ' *' : ''}`;

        let input;

        // Handle option type (dropdown)
        if (fieldDef.type === 'option') {
            input = document.createElement('select');
            input.id = `${idPrefix}-${fieldName}`;
            input.className = 'form-input';
            input.setAttribute('data-field-name', fieldName);
            input.setAttribute('data-field-type', fieldDef.type);

            // Add options
            if (fieldDef.options && Array.isArray(fieldDef.options)) {
                fieldDef.options.forEach(optionValue => {
                    const option = document.createElement('option');
                    option.value = optionValue;
                    option.textContent = optionValue;
                    input.appendChild(option);
                });
            }

            // Set required attribute
            if (fieldDef.required) {
                input.required = true;
            }

            // Set current value if provided
            if (currentValues && currentValues[fieldName] !== undefined) {
                input.value = currentValues[fieldName];
            }
        } else {
            // Standard input field
            input = document.createElement('input');
            input.id = `${idPrefix}-${fieldName}`;
            input.className = 'form-input';
            input.setAttribute('data-field-name', fieldName);
            input.setAttribute('data-field-type', fieldDef.type);

            // Set input type based on field type
            if (fieldDef.type === 'float' || fieldDef.type === 'int') {
                input.type = 'number';
                if (fieldDef.type === 'float') {
                    input.step = 'any';
                }
            } else if (fieldDef.type === 'bool') {
                input.type = 'checkbox';
            } else {
                input.type = 'text';
            }

            // Set required attribute
            if (fieldDef.required) {
                input.required = true;
            }

            // Set current value if provided
            if (currentValues && currentValues[fieldName] !== undefined) {
                if (fieldDef.type === 'bool') {
                    input.checked = currentValues[fieldName];
                } else {
                    input.value = currentValues[fieldName];
                }
            }
        }

        return { label, input };
    }

    /**
     * Validate form fields in a container
     * @param {HTMLElement} container - The container with form fields
     * @param {string} nodeType - The type of node being validated
     * @returns {Object} { valid: boolean, errors: string[] }
     */
    validate(container, nodeType) {
        const errors = [];
        const config = this.nodeTypeConfigs[nodeType];

        if (!config || !config.fields || config.fields.length === 0) {
            return { valid: true, errors: [] };
        }

        const inputs = container.querySelectorAll('input, select');
        const fieldMap = new Map();

        // Build map of field values
        inputs.forEach(input => {
            const fieldName = input.getAttribute('data-field-name');
            const fieldType = input.getAttribute('data-field-type');

            fieldMap.set(fieldName, {
                input,
                fieldType,
                value: fieldType === 'bool' ? input.checked : input.value
            });
        });

        // Validate each field definition (fields is now an array)
        config.fields.forEach(fieldDef => {
            const fieldName = fieldDef.name;
            const field = fieldMap.get(fieldName);

            if (!field) {
                if (fieldDef.required) {
                    errors.push(`Field "${fieldName}" is missing`);
                }
                return;
            }

            const { input, fieldType, value } = field;

            // Check required fields
            if (fieldDef.required) {
                if (fieldType === 'bool') {
                    // Checkboxes are valid even when unchecked
                } else if (value === '' || value === null || value === undefined) {
                    errors.push(`Field "${fieldName}" is required`);
                    input.classList.add('error');
                } else {
                    input.classList.remove('error');
                }
            }

            // Type validation
            if (value !== '' && fieldType === 'int') {
                const parsed = parseInt(value, 10);
                if (isNaN(parsed)) {
                    errors.push(`Field "${fieldName}" must be a valid integer`);
                    input.classList.add('error');
                } else {
                    input.classList.remove('error');
                }
            } else if (value !== '' && fieldType === 'float') {
                const parsed = parseFloat(value);
                if (isNaN(parsed)) {
                    errors.push(`Field "${fieldName}" must be a valid number`);
                    input.classList.add('error');
                } else {
                    input.classList.remove('error');
                }
            }
        });

        return {
            valid: errors.length === 0,
            errors
        };
    }

    /**
     * Extract config values from a container's form fields
     * @param {HTMLElement} container - The container with form fields
     * @returns {Object} The extracted configuration values
     */
    getValues(container) {
        const config = {};
        const inputs = container.querySelectorAll('input, select');

        inputs.forEach(input => {
            const fieldName = input.getAttribute('data-field-name');
            const fieldType = input.getAttribute('data-field-type');
            let value = input.value;

            if (fieldType === 'int') {
                value = parseInt(value, 10);
            } else if (fieldType === 'float') {
                value = parseFloat(value);
            } else if (fieldType === 'bool') {
                value = input.checked;
            } else if (fieldType === 'option') {
                // Keep as string
                value = input.value;
            } else if (fieldType === 'string') {
                // Keep as string
                value = input.value;
            }

            // Only include valid values
            if (value !== '' && !isNaN(value)) {
                config[fieldName] = value;
            } else if ((fieldType === 'option' || fieldType === 'string') && value !== '') {
                // Include option and string values even if they're strings
                config[fieldName] = value;
            }
        });

        return config;
    }

}
