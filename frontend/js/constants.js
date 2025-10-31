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
        width: 160,
        height: 120,
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

// Debounce delays (in milliseconds)
export const DEBOUNCE_DELAYS = {
    layoutSave: 500,
    viewportSave: 500
};

// Zoom configuration
export const ZOOM_CONFIG = {
    factor: {
        in: 1.1,   // Zoom in multiplier
        out: 0.9   // Zoom out multiplier
    },
    limits: {
        min: 0.1,  // Minimum zoom level
        max: 5     // Maximum zoom level
    }
};

// Layout configuration for auto-layout
export const LAYOUT_CONFIG = {
    gridColumns: 3,
    gridSpacing: 100,
    initialOffset: 100
};

// Connection delete button configuration
export const CONNECTION_DELETE_BUTTON = {
    radius: 12,
    fontSize: 20,
    hoverColor: '#e74c3c'
};

// Node type configurations are now loaded dynamically from the backend API
// See node-type-schemas.js and node-type-config-store.js

// DOM data attributes
export const DATA_ATTRIBUTES = {
    nodeId: 'data-node-id',
    inputName: 'data-input-name',
    outputName: 'data-output-name',
    imageId: 'data-image-id'
};

// CSS class names
export const CSS_CLASSES = {
    node: 'node',
    nodeRect: 'node-rect',
    nodeTitleBar: 'node-title-bar',
    nodeTitle: 'node-title',
    nodeThumbnail: 'node-thumbnail',
    portCell: 'port-cell',
    portCellInput: 'port-cell-input',
    portCellOutput: 'port-cell-output',
    port: 'port',
    portLabel: 'port-label',
    connection: 'connection',
    connectionGroup: 'connection-group',
    connectionPath: 'connection-path',
    tempConnection: 'temp-connection',
    deleteButton: 'connection-delete-button'
};
