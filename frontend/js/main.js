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
const initialGraphId = graphManager.getGraphIdFromUrl();
const LEGO_PALETTE_COLORS = '#000000,#00385e,#004a2d,#006cb7,#009247,#00a2d9,#00af4c,#00bdd2,#3a170d,#41413d,#489ece,#4c2f92,#646765,#678297,#692e14,#6e9379,#78bee9,#7f131b,#828353,#878d8f,#947e5f,#9675b4,#99c93c,#a0a19e,#a55222,#ae7345,#b41b7d,#bca6d0,#c0e3da,#c39737,#cce197,#dd1a21,#ddc48e,#de8b5f,#e6edcf,#e85da2,#f3f3f3,#f57d20,#f6accd,#fbab18,#fcc39e,#ffcd03,#fff478';
const GREYSCALE_COLORS = '#000000,#111111,#222222,#333333,#444444,#555555,#666666,#777777,#888888,#999999,#aaaaaa,#bbbbbb,#cccccc,#dddddd,#eeeeee,#ffffff';

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
        graphManager.clearSelection();
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
        // Show node context menu, hide canvas context menu
        const nodeId = clickedNode.getAttribute('data-node-id');
        contextMenuNodeId = nodeId;

        nodeContextMenu.style.left = `${e.clientX - 5}px`;
        nodeContextMenu.style.top = `${e.clientY - 5}px`;
        nodeContextMenu.classList.add('active');
        contextMenu.classList.remove('active');
    } else if (!clickedConnection) {
        // Show canvas context menu, hide node context menu
        const svgRect = svg.getBoundingClientRect();
        const screenX = e.clientX - svgRect.left;
        const screenY = e.clientY - svgRect.top;
        contextMenuPosition = interactions.screenToCanvas(screenX, screenY);

        contextMenu.style.left = `${e.clientX - 5}px`;
        contextMenu.style.top = `${e.clientY - 5}px`;
        contextMenu.classList.add('active');
        nodeContextMenu.classList.remove('active');
    }
});

// Close context menu when clicking anywhere else
document.addEventListener('click', (e) => {
    // Check if click is outside any context menu
    if (!e.target.closest('.context-menu')) {
        contextMenu.classList.remove('active');
        nodeContextMenu.classList.remove('active');
    }
});

// Prevent clicks inside context menu from bubbling to document click handler
// This ensures the menu stays open when clicking on non-selectable areas
contextMenu.addEventListener('click', (e) => {
    const nodeTypeItem = e.target.closest('[data-node-type]');
    const presetItem = e.target.closest('[data-palette-preset]');

    if (presetItem) {
        const preset = presetItem.getAttribute('data-palette-preset');
        contextMenu.classList.remove('active');
        if (preset === 'lego') {
            createLegoPaletteNode(contextMenuPosition);
        } else if (preset === 'greyscale') {
            createGreyscalePaletteNode(contextMenuPosition);
        }
        return;
    }

    if (nodeTypeItem) {
        // Item selected - close menu and open modal
        const nodeType = nodeTypeItem.getAttribute('data-node-type');
        contextMenu.classList.remove('active');
        modals?.addNode.open(nodeType, contextMenuPosition);
        return;
    }
    // Don't close menu if clicking on parent items or empty space
    // The menu will only close when an item is selected or when clicking outside (handled by document click handler)
});

// Handle node context menu actions
nodeContextMenu.addEventListener('click', (e) => {
    const actionItem = e.target.closest('[data-action]');
    if (actionItem && contextMenuNodeId) {
        // Item selected - close menu and perform action
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
    // Don't close menu if clicking on empty space
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

    // Group node types by category, preserving order of first appearance
    const orderedTypes = schemas._orderedTypes || Object.keys(schemas);
    const categorized = {};
    const categoryOrder = []; // Track category order based on first appearance

    orderedTypes.forEach((nodeType) => {
        const config = schemas[nodeType];
        if (!config) return; // Skip if config doesn't exist (e.g., _orderedTypes itself)

        const category = config.category || 'Other';
        if (!categorized[category]) {
            categorized[category] = [];
            categoryOrder.push(category); // Add to order list on first appearance
        }
        categorized[category].push({ nodeType, config });
    });

    // Create nested menu items for each category in the order they were received
    categoryOrder.forEach((category) => {
        const nodes = categorized[category];
        if (!nodes || nodes.length === 0) return;

        // Create category parent item with submenu
        const categoryParent = document.createElement('div');
        categoryParent.className = 'context-menu-item context-menu-parent';

        const categoryLabel = document.createElement('span');
        categoryLabel.textContent = category;
        categoryParent.appendChild(categoryLabel);

        // Add arrow indicator
        const arrow = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        arrow.setAttribute('class', 'context-menu-arrow');
        arrow.setAttribute('width', '8');
        arrow.setAttribute('height', '12');
        arrow.setAttribute('viewBox', '0 0 8 12');
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', 'M 1 1 L 6 6 L 1 11');
        path.setAttribute('fill', 'none');
        path.setAttribute('stroke', 'currentColor');
        path.setAttribute('stroke-width', '2');
        arrow.appendChild(path);
        categoryParent.appendChild(arrow);

        // Create submenu for this category
        const categorySubmenu = document.createElement('div');
        categorySubmenu.className = 'context-menu-submenu';

        // Add nodes in this category to the submenu
        nodes.forEach(({ nodeType, config }) => {
            const item = document.createElement('div');
            item.className = 'context-menu-item';
            item.setAttribute('data-node-type', nodeType);
            item.textContent = config.name;
            categorySubmenu.appendChild(item);
        });

        // Add preset palettes section for Palette category
        if (category === 'Palette') {
            const divider = document.createElement('div');
            divider.className = 'context-menu-divider';
            categorySubmenu.appendChild(divider);

            const legoItem = document.createElement('div');
            legoItem.className = 'context-menu-item';
            legoItem.setAttribute('data-palette-preset', 'lego');
            legoItem.textContent = 'Lego';
            categorySubmenu.appendChild(legoItem);

            const greyscaleItem = document.createElement('div');
            greyscaleItem.className = 'context-menu-item';
            greyscaleItem.setAttribute('data-palette-preset', 'greyscale');
            greyscaleItem.textContent = 'Greyscale';
            categorySubmenu.appendChild(greyscaleItem);
        }

        categoryParent.appendChild(categorySubmenu);
        submenu.appendChild(categoryParent);
    });
}

async function createLegoPaletteNode(position) {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) {
        toastManager.warning('Please select a graph first');
        return;
    }

    try {
        const nodeId = await api.addNode(graphId, 'palette_create', 'Lego', { colors: LEGO_PALETTE_COLORS });

        if (position) {
            renderer.updateNodePosition(nodeId, position.x, position.y);
            const nodePositions = renderer.exportNodePositions();
            await api.updateLayout(graphId, nodePositions);
        }

        await graphManager.reloadCurrentGraph();
        toastManager.success('Lego Palette node added');
    } catch (error) {
        console.error('Failed to create Lego Palette node:', error);
        toastManager.error(`Failed to add Lego Palette: ${error.message}`);
    }
}

async function createGreyscalePaletteNode(position) {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) {
        toastManager.warning('Please select a graph first');
        return;
    }

    try {
        const nodeId = await api.addNode(graphId, 'palette_create', 'Greyscale', { colors: GREYSCALE_COLORS });

        if (position) {
            renderer.updateNodePosition(nodeId, position.x, position.y);
            const nodePositions = renderer.exportNodePositions();
            await api.updateLayout(graphId, nodePositions);
        }

        await graphManager.reloadCurrentGraph();
        toastManager.success('Greyscale palette node added');
    } catch (error) {
        console.error('Failed to create Greyscale palette node:', error);
        toastManager.error(`Failed to add Greyscale palette: ${error.message}`);
    }
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
        await graphManager.loadGraphList(initialGraphId);
    } catch (error) {
        console.error('Failed to initialize application:', error);
        toastManager.show('Failed to load application configuration. Please refresh the page.', 'error');
    }
}

initialize();
