// Node Type Schema Management
// Fetches node type schemas from the backend API and provides them in the same format
// as the old NODE_TYPE_CONFIGS constant

import { API_PATHS } from './constants.js';

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
        // Backend returns an array of {name, display_name, schema} objects in the desired order
        const configs = {};
        const orderedTypes = []; // Preserve order from backend

        for (const entry of data.node_types) {
            const nodeType = entry.name;
            configs[nodeType] = {
                name: entry.display_name,
                category: entry.category,
                nameRequired: entry.schema.name_required,
                fields: entry.schema.fields
            };
            orderedTypes.push(nodeType);
        }

        // Attach the ordering information to the configs object
        configs._orderedTypes = orderedTypes;

        return configs;
    } catch (error) {
        console.error('Error loading node type schemas:', error);
        throw error;
    }
}
