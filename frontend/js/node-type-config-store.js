// Central store for node type configurations loaded from the backend
// This allows synchronous access to configs after they've been loaded

let configs = null;

export function setNodeTypeConfigs(loadedConfigs) {
    configs = loadedConfigs;
}

export function getNodeTypeConfigs() {
    return configs;
}

export function getNodeTypeConfig(nodeType) {
    return configs?.[nodeType] || null;
}
