// Application-wide constants

// Node rendering constants
export const NODE_DESIGN = {
    width: 200,
    height: 180,
    titleBarHeight: 30,
    tablePadding: 8,
    rowHeight: 24,
    headerHeight: 20,
    ports: {
        radius: 6
    },
    thumbnail: {
        width: 120,
        height: 90,
        y: 48
    }
};

// API path templates
export const API_PATHS = {
    base: '/api',
    imagegraphs: '/api/imagegraphs',
    images: (imageId) => `/api/images/${imageId}`,
    graphWebSocket: (graphId) => `/api/imagegraphs/${graphId}/ws`
};

// WebSocket configuration
export const WS_CONFIG = {
    reconnectDelay: 3000 // 3 seconds
};

// Sidebar configuration
export const SIDEBAR_CONFIG = {
    minWidth: 200,
    maxWidth: 600,
    storageKey: 'artwork-sidebar-width'
};
