// Main application entry point

import * as api from './api.js';
import { GraphState } from './graph.js';
import { Renderer } from './renderer.js';
import { InteractionHandler } from './interactions.js';
import { ModalManager } from './modal.js';
import { ToastManager } from './toast.js';
import { NodeConfigFormBuilder } from './form-builder.js';
import { GraphManager } from './graph-manager.js';
import { SIDEBAR_CONFIG } from './constants.js';
import { loadNodeTypeSchemas } from './node-type-schemas.js';
import { setNodeTypeConfigs, getNodeTypeConfigs } from './node-type-config-store.js';
import { OutputSidebar } from './output-sidebar.js';
import {
    CreateGraphModalController,
    AddNodeModalController,
    EditNodeModalController,
    DeleteNodeModalController,
    ViewImageModalController,
    CropModal
} from './modals/index.js';

// Initialize state and rendering
const graphState = new GraphState();
const svg = document.getElementById('graph-canvas');
const nodesLayer = document.getElementById('nodes-layer');
const connectionsLayer = document.getElementById('connections-layer');
const renderer = new Renderer(svg, nodesLayer, connectionsLayer, graphState);

// Initialize managers
const modalManager = new ModalManager();
const toastManager = new ToastManager();
const graphManager = new GraphManager(api, graphState, renderer, toastManager);
const outputSidebar = new OutputSidebar(graphState, renderer, toastManager);

// Crop modal for visual crop configuration
const cropModal = new CropModal(toastManager);

// Form builder will be initialized after loading schemas
let formBuilder = null;

// Modal controllers - will be initialized after schemas are loaded
let modals = null;

// Initialize interaction handler with modal callback
const interactions = new InteractionHandler(
    svg,
    renderer,
    graphState,
    api,
    null, // No longer need renderOutputs callback - OutputSidebar subscribes to graphState directly
    (nodeId) => modals?.editNode.open(nodeId), // Use modal controller
    toastManager
);

// UI elements
const graphSelect = document.getElementById('graph-select');
const createGraphBtn = document.getElementById('create-graph-btn');
const refreshBtn = document.getElementById('refresh-btn');

// Context menu
const contextMenu = document.getElementById('context-menu');
const nodeContextMenu = document.getElementById('node-context-menu');
let contextMenuPosition = { x: 0, y: 0 };
let contextMenuNodeId = null;

// Subscribe to graph state changes
graphState.subscribe((graph) => {
    if (graph) {
        renderer.render(graph);
    } else {
        renderer.clear();
    }
});

// Set up viewport change callback for zoom persistence
renderer.setViewportChangeCallback(() => {
    interactions.debouncedSaveViewport();
});

// Handle graph selection from dropdown
graphSelect.addEventListener('change', (e) => {
    const graphId = e.target.value;
    if (graphId) {
        graphManager.selectGraph(graphId);
    } else {
        graphManager.disconnectWebSocket();
        graphState.setCurrentGraph(null);
    }
});

// Create graph button handler
createGraphBtn.addEventListener('click', () => {
    modals?.createGraph.open();
});


// Handle connection delete button clicks
svg.addEventListener('click', (e) => {
    // Check for connection delete button
    const connectionDeleteBtn = e.target.closest('.connection-delete-btn');
    if (connectionDeleteBtn) {
        const connectionGroup = connectionDeleteBtn.closest('.connection-group');
        const fromNode = connectionGroup.getAttribute('data-from-node');
        const fromOutput = connectionGroup.getAttribute('data-from-output');
        const toNode = connectionGroup.getAttribute('data-to-node');
        const toInput = connectionGroup.getAttribute('data-to-input');

        handleDisconnectConnection(fromNode, fromOutput, toNode, toInput);
        e.stopPropagation();
        return;
    }
});

// Handle connection disconnect
async function handleDisconnectConnection(fromNodeId, fromOutput, toNodeId, toInput) {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) return;

    try {
        await api.disconnectNodes(graphId, fromNodeId, fromOutput, toNodeId, toInput);
        // Refresh graph to show connection removed
        await graphManager.reloadCurrentGraph();
        toastManager.success('Connection removed');
    } catch (error) {
        console.error('Failed to disconnect nodes:', error);
        toastManager.error(`Failed to disconnect nodes: ${error.message}`);
    }
}

// Refresh current graph
refreshBtn.addEventListener('click', async () => {
    if (!graphState.getCurrentGraphId()) return;

    try {
        await graphManager.reloadCurrentGraph();
        toastManager.info('Graph refreshed');
    } catch (error) {
        console.error('Failed to refresh graph:', error);
        toastManager.error(`Failed to refresh graph: ${error.message}`);
    }
});

// Context menu handlers
svg.addEventListener('contextmenu', (e) => {
    e.preventDefault();

    // Only show context menu if a graph is selected
    if (!graphState.getCurrentGraphId()) {
        return;
    }

    // Check if right-clicking on a node
    const clickedNode = e.target.closest('.node');
    const clickedConnection = e.target.closest('.connection-group');

    if (clickedNode) {
        // Show node context menu
        const nodeId = clickedNode.getAttribute('data-node-id');
        contextMenuNodeId = nodeId;

        nodeContextMenu.style.left = `${e.clientX - 5}px`;
        nodeContextMenu.style.top = `${e.clientY - 5}px`;
        nodeContextMenu.classList.add('active');
    } else if (!clickedConnection) {
        // Show canvas context menu (not on a node or connection)
        const svgRect = svg.getBoundingClientRect();
        const screenX = e.clientX - svgRect.left;
        const screenY = e.clientY - svgRect.top;
        contextMenuPosition = interactions.screenToCanvas(screenX, screenY);

        contextMenu.style.left = `${e.clientX - 5}px`;
        contextMenu.style.top = `${e.clientY - 5}px`;
        contextMenu.classList.add('active');
    }
});

// Close context menu when clicking anywhere else
document.addEventListener('click', (e) => {
    if (!e.target.closest('.context-menu')) {
        contextMenu.classList.remove('active');
        nodeContextMenu.classList.remove('active');
    }
});

// Close context menu when mouse leaves it
contextMenu.addEventListener('mouseleave', () => {
    contextMenu.classList.remove('active');
});

nodeContextMenu.addEventListener('mouseleave', () => {
    nodeContextMenu.classList.remove('active');
});

// Handle canvas context menu node type selection
contextMenu.addEventListener('click', (e) => {
    const nodeTypeItem = e.target.closest('[data-node-type]');
    if (nodeTypeItem) {
        const nodeType = nodeTypeItem.getAttribute('data-node-type');
        contextMenu.classList.remove('active');
        modals?.addNode.open(nodeType, contextMenuPosition);
    }
});

// Handle node context menu actions
nodeContextMenu.addEventListener('click', (e) => {
    const actionItem = e.target.closest('[data-action]');
    if (actionItem && contextMenuNodeId) {
        const action = actionItem.getAttribute('data-action');
        nodeContextMenu.classList.remove('active');

        if (action === 'view') {
            modals?.viewImage.open(contextMenuNodeId);
        } else if (action === 'edit-config') {
            modals?.editNode.open(contextMenuNodeId);
        } else if (action === 'delete') {
            modals?.deleteNode.open(contextMenuNodeId);
        }

        contextMenuNodeId = null;
    }
});

// Clean up WebSocket on page unload
window.addEventListener('beforeunload', () => {
    graphManager.cleanup();
});

// Sidebar resize functionality
const sidebar = document.getElementById('outputs-sidebar');
const resizeHandle = document.getElementById('sidebar-resize-handle');
let isResizing = false;
let startX = 0;
let startWidth = 0;

// Restore saved width from localStorage
const savedWidth = localStorage.getItem(SIDEBAR_CONFIG.storageKey);
if (savedWidth) {
    const width = parseInt(savedWidth, 10);
    if (width >= SIDEBAR_CONFIG.minWidth && width <= SIDEBAR_CONFIG.maxWidth) {
        sidebar.style.width = `${width}px`;
    }
}

resizeHandle.addEventListener('mousedown', (e) => {
    isResizing = true;
    startX = e.clientX;
    startWidth = sidebar.offsetWidth;

    // Add resizing class to prevent text selection
    document.body.classList.add('resizing-active');
    resizeHandle.classList.add('resizing');

    e.preventDefault();
});

document.addEventListener('mousemove', (e) => {
    if (!isResizing) return;

    // Calculate new width (resize from left, so subtract the delta)
    const deltaX = startX - e.clientX;
    const newWidth = startWidth + deltaX;

    // Constrain width
    const constrainedWidth = Math.max(SIDEBAR_CONFIG.minWidth, Math.min(SIDEBAR_CONFIG.maxWidth, newWidth));

    sidebar.style.width = `${constrainedWidth}px`;
});

document.addEventListener('mouseup', () => {
    if (isResizing) {
        isResizing = false;
        document.body.classList.remove('resizing-active');
        resizeHandle.classList.remove('resizing');

        // Save width to localStorage
        localStorage.setItem(SIDEBAR_CONFIG.storageKey, sidebar.offsetWidth);
    }
});

// Populate the "Add Node" context menu submenu with node types
function populateAddNodeContextMenu(schemas) {
    const submenu = document.getElementById('add-node-submenu');
    if (!submenu) {
        console.error('Add node submenu element not found');
        return;
    }

    // Clear existing items
    submenu.innerHTML = '';

    // Create menu items for each node type using the order from backend
    // The schemas object has an _orderedTypes array that preserves backend ordering
    const orderedTypes = schemas._orderedTypes || Object.keys(schemas);

    orderedTypes.forEach((nodeType) => {
        const config = schemas[nodeType];
        if (!config) return; // Skip if config doesn't exist (e.g., _orderedTypes itself)

        const item = document.createElement('div');
        item.className = 'context-menu-item';
        item.setAttribute('data-node-type', nodeType);
        item.textContent = config.name;
        submenu.appendChild(item);
    });
}

// Load initial data
async function initialize() {
    try {
        // Load node type schemas from backend
        const schemas = await loadNodeTypeSchemas();
        setNodeTypeConfigs(schemas);

        // Initialize form builder with loaded schemas
        formBuilder = new NodeConfigFormBuilder(schemas);

        // Initialize modal controllers now that formBuilder is ready
        modals = {
            createGraph: new CreateGraphModalController(
                api, graphManager, toastManager, modalManager, interactions
            ),
            addNode: new AddNodeModalController(
                api, graphState, graphManager, renderer, formBuilder,
                toastManager, modalManager, interactions, getNodeTypeConfigs
            ),
            editNode: new EditNodeModalController(
                api, graphState, graphManager, formBuilder, cropModal,
                toastManager, modalManager, interactions, getNodeTypeConfigs
            ),
            deleteNode: new DeleteNodeModalController(
                api, graphState, graphManager, toastManager, modalManager, interactions
            ),
            viewImage: new ViewImageModalController(
                graphState, modalManager, interactions
            )
        };

        // Populate context menu with node types
        populateAddNodeContextMenu(schemas);

        // Load graph list
        await graphManager.loadGraphList();
    } catch (error) {
        console.error('Failed to initialize application:', error);
        toastManager.show('Failed to load application configuration. Please refresh the page.', 'error');
    }
}

initialize();
