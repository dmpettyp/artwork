// Node Type Schema Management
// Fetches node type schemas from the backend API and provides them in the same format
// as the old NODE_TYPE_CONFIGS constant

import { API_PATHS } from './constants.js';

// Display names for node types (could also come from backend in future)
const DISPLAY_NAMES = {
    'input': 'Input',
    'blur': 'Blur',
    'resize': 'Resize',
    'resize_match': 'Resize Match',
    'output': 'Output'
};

/**
 * Fetches node type schemas from the backend and converts to the format
 * expected by existing frontend code
 * @returns {Promise<Object>} Node type configs in the same format as NODE_TYPE_CONFIGS
 */
export async function loadNodeTypeSchemas() {
    try {
        const response = await fetch(`${API_PATHS.base}/node-types`);
        if (!response.ok) {
            throw new Error(`Failed to fetch node type schemas: ${response.statusText}`);
        }
        const data = await response.json();

        // Convert backend schema format to frontend config format
        const configs = {};
        for (const [nodeType, schema] of Object.entries(data.node_types)) {
            configs[nodeType] = {
                name: DISPLAY_NAMES[nodeType] || nodeType,
                nameRequired: schema.name_required,
                fields: schema.fields
            };
        }

        return configs;
    } catch (error) {
        console.error('Error loading node type schemas:', error);
        throw error;
    }
}
