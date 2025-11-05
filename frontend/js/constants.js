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
        y: 48,
        pixelatedThreshold: 100 // Images smaller than this use pixelated rendering
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

// Crop aspect ratio presets
export const CROP_ASPECT_RATIOS = [
    { value: 'none', label: 'None', ratio: null },
    { value: '1:1', label: '1:1 (Square)', ratio: [1, 1] },
    { value: '1:2', label: '1:2', ratio: [1, 2] },
    { value: '2:3', label: '2:3', ratio: [2, 3] },
    { value: '3:2', label: '3:2', ratio: [3, 2] },
    { value: '3:4', label: '3:4', ratio: [3, 4] },
    { value: '4:3', label: '4:3', ratio: [4, 3] },
    { value: '4:6', label: '4:6', ratio: [4, 6] },
    { value: '5:7', label: '5:7', ratio: [5, 7] },
    { value: '6:4', label: '6:4', ratio: [6, 4] },
    { value: '7:5', label: '7:5', ratio: [7, 5] },
    { value: 'custom', label: 'Custom', ratio: null }
];

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
