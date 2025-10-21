// Main application entry point

import * as api from './api.js';
import { GraphState } from './graph.js';
import { Renderer } from './renderer.js';
import { InteractionHandler } from './interactions.js';
import { Modal, ModalManager } from './modal.js';
import { ToastManager } from './toast.js';
import { NodeConfigFormBuilder } from './form-builder.js';

// Initialize state and rendering
const graphState = new GraphState();
const svg = document.getElementById('graph-canvas');
const nodesLayer = document.getElementById('nodes-layer');
const connectionsLayer = document.getElementById('connections-layer');
const renderer = new Renderer(svg, nodesLayer, connectionsLayer);

// Declare functions early so they can be passed to InteractionHandler
let renderOutputs;
let openEditConfigModal;

const interactions = new InteractionHandler(
    svg,
    renderer,
    graphState,
    api,
    (graph) => renderOutputs(graph),
    (nodeId) => openEditConfigModal(nodeId)
);

// WebSocket connection management
let wsConnection = null;
let wsReconnectTimeout = null;
const WS_RECONNECT_DELAY = 3000; // 3 seconds

// UI elements
const graphSelect = document.getElementById('graph-select');
const createGraphBtn = document.getElementById('create-graph-btn');
const refreshBtn = document.getElementById('refresh-btn');

// Context menu
const contextMenu = document.getElementById('context-menu');
const nodeContextMenu = document.getElementById('node-context-menu');
let contextMenuPosition = { x: 0, y: 0 };
let contextMenuNodeId = null;

// Initialize modal manager and toast manager
const modalManager = new ModalManager();
const toastManager = new ToastManager();

// Create graph modal
const graphNameInput = document.getElementById('graph-name-input');
const modalCreateBtn = document.getElementById('modal-create-btn');
const modalCancelBtn = document.getElementById('modal-cancel-btn');

const createGraphModal = new Modal('create-graph-modal', {
    onOpen: () => {
        interactions.cancelAllDrags();
        graphNameInput.value = '';
        graphNameInput.focus();
    }
});
modalManager.register(createGraphModal);

// Add node modal
const addNodeModalElement = document.getElementById('add-node-modal');
const addNodeModalTitle = document.getElementById('add-node-modal-title');
const nodeNameInput = document.getElementById('node-name-input');
const nodeImageUpload = document.getElementById('node-image-upload');
const nodeImageInput = document.getElementById('node-image-input');
const nodeConfigFields = document.getElementById('node-config-fields');
const addNodeCreateBtn = document.getElementById('add-node-create-btn');
const addNodeCancelBtn = document.getElementById('add-node-cancel-btn');

// Track the node type for the add modal (set when opening)
let addNodeType = null;

const addNodeModal = new Modal('add-node-modal', {
    onOpen: () => {
        interactions.cancelAllDrags();
    }
});
modalManager.register(addNodeModal);

// Edit config modal
const editNodeNameInput = document.getElementById('edit-node-name-input');
const editImageUpload = document.getElementById('edit-image-upload');
const editImageInput = document.getElementById('edit-image-input');
const editConfigFields = document.getElementById('edit-config-fields');
const editConfigSaveBtn = document.getElementById('edit-config-save-btn');
const editConfigCancelBtn = document.getElementById('edit-config-cancel-btn');

const editConfigModal = new Modal('edit-config-modal', {
    onOpen: () => {
        interactions.cancelAllDrags();
    },
    onClose: () => {
        currentNodeId = null;
    }
});
modalManager.register(editConfigModal);

// Delete node modal
const deleteNodeName = document.getElementById('delete-node-name');
const deleteNodeConfirmBtn = document.getElementById('delete-node-confirm-btn');
const deleteNodeCancelBtn = document.getElementById('delete-node-cancel-btn');

const deleteNodeModal = new Modal('delete-node-modal', {
    onOpen: () => {
        interactions.cancelAllDrags();
    },
    onClose: () => {
        currentNodeId = null;
    }
});
modalManager.register(deleteNodeModal);

// Track current node being edited/deleted
let currentNodeId = null;

// Subscribe to graph state changes
graphState.subscribe((graph) => {
    if (graph) {
        renderer.render(graph);
        renderOutputs(graph);
    } else {
        renderer.clear();
        renderOutputs(null);
    }
});

// Set up viewport change callback for zoom persistence
renderer.setViewportChangeCallback(() => {
    interactions.debouncedSaveViewport();
});

// Load and display graph list
async function loadGraphList() {
    try {
        const graphs = await api.listImageGraphs();
        renderGraphList(graphs);

        // Auto-select the first graph if none is selected
        if (graphs.length > 0 && !graphState.getCurrentGraphId()) {
            await selectGraph(graphs[0].id);
        }
    } catch (error) {
        console.error('Failed to load graphs:', error);
    }
}

function renderGraphList(graphs) {
    const currentGraphId = graphState.getCurrentGraphId();

    // Clear and add default option
    graphSelect.innerHTML = '<option value="">Select a graph...</option>';

    graphs.forEach(graph => {
        const option = document.createElement('option');
        option.value = graph.id;
        option.textContent = graph.name;

        if (graph.id === currentGraphId) {
            option.selected = true;
        }

        graphSelect.appendChild(option);
    });
}

// Handle graph selection from dropdown
graphSelect.addEventListener('change', (e) => {
    const graphId = e.target.value;
    if (graphId) {
        selectGraph(graphId);
    } else {
        disconnectWebSocket();
        graphState.setCurrentGraph(null);
    }
});

// Select and load a graph
async function selectGraph(graphId) {
    try {
        const graph = await api.getImageGraph(graphId);

        // Load UI metadata and restore viewport/positions
        try {
            const uiMetadata = await api.getUIMetadata(graphId);
            renderer.restoreViewport(uiMetadata.viewport);
            renderer.restoreNodePositions(uiMetadata.node_positions);
        } catch (error) {
            console.log('No UI metadata found, using defaults:', error);
            // Not an error - just means no metadata was saved yet
        }

        graphState.setCurrentGraph(graph);
        await loadGraphList(); // Refresh list to update active state

        // Connect to WebSocket for real-time updates
        connectWebSocket(graphId);
    } catch (error) {
        console.error('Failed to load graph:', error);
        toastManager.error(`Failed to load graph: ${error.message}`);
    }
}

// Reload the currently selected graph
async function reloadCurrentGraph() {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) return;

    const graph = await api.getImageGraph(graphId);
    graphState.setCurrentGraph(graph);
}

// Node type configuration (matches backend node_type.go)
const nodeTypeConfigs = {
    input: {
        fields: {}
    },
    blur: {
        fields: {
            radius: { type: 'int', required: true }
        }
    },
    resize: {
        fields: {
            width: { type: 'int', required: false },
            height: { type: 'int', required: false }
        }
    },
    output: {
        fields: {}
    }
};

// Initialize form builder
const formBuilder = new NodeConfigFormBuilder(nodeTypeConfigs);

// Create graph modal functions
function openCreateGraphModal() {
    createGraphModal.open();
}

function closeCreateGraphModal() {
    createGraphModal.close();
}

// Add node modal functions
function openAddNodeModal(nodeType) {
    if (!graphState.getCurrentGraphId()) {
        toastManager.warning('Please select a graph first');
        return;
    }

    // Store the node type
    addNodeType = nodeType;

    // Update modal title with capitalized node type
    const capitalizedType = nodeType.charAt(0).toUpperCase() + nodeType.slice(1);
    addNodeModalTitle.textContent = `Add ${capitalizedType} Node`;

    // Clear inputs
    nodeNameInput.value = '';
    nodeImageInput.value = '';
    nodeConfigFields.innerHTML = '';

    // Show/hide image upload based on node type
    if (nodeType === 'input') {
        nodeImageUpload.style.display = 'block';
    } else {
        nodeImageUpload.style.display = 'none';
    }

    // Render config fields for the node type
    renderNodeConfigFields(nodeType);

    addNodeModal.open();
    nodeNameInput.focus();
}

function closeAddNodeModal() {
    addNodeModal.close();
    addNodeType = null;
}

function renderNodeConfigFields(nodeType) {
    formBuilder.renderFields(nodeConfigFields, nodeType, 'config');
}

function getNodeConfig() {
    return formBuilder.getValues(nodeConfigFields);
}

// Create new graph handlers
createGraphBtn.addEventListener('click', () => {
    openCreateGraphModal();
});

modalCancelBtn.addEventListener('click', () => {
    closeCreateGraphModal();
});

modalCreateBtn.addEventListener('click', async () => {
    const name = graphNameInput.value.trim();
    if (!name) return;

    try {
        const graph_id = await api.createImageGraph(name);
        closeCreateGraphModal();
        await loadGraphList();
        await selectGraph(graph_id);
        toastManager.success(`Graph "${name}" created successfully`);
    } catch (error) {
        console.error('Failed to create graph:', error);
        toastManager.error(`Failed to create graph: ${error.message}`);
    }
});

graphNameInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        modalCreateBtn.click();
    }
});

// Add node handlers
addNodeCancelBtn.addEventListener('click', () => {
    closeAddNodeModal();
});

addNodeCreateBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) return;

    const nodeType = addNodeType;
    const nodeName = nodeNameInput.value.trim();

    if (!nodeType) {
        toastManager.warning('Node type not set');
        return;
    }

    if (!nodeName) {
        toastManager.warning('Please enter a node name');
        return;
    }

    // For input nodes, check if an image file is selected
    if (nodeType === 'input' && nodeImageInput.files.length === 0) {
        toastManager.warning('Please select an image file for the input node');
        return;
    }

    // Validate config fields
    const validation = formBuilder.validate(nodeConfigFields, nodeType);
    if (!validation.valid) {
        toastManager.warning(validation.errors[0]);
        return;
    }

    const config = getNodeConfig();

    try {
        // Add the node first to get the node ID
        const nodeId = await api.addNode(graphId, nodeType, nodeName, config);

        // If this is an input node with an image, upload it to the "original" output
        if (nodeType === 'input' && nodeImageInput.files.length > 0) {
            const imageFile = nodeImageInput.files[0];
            await api.uploadNodeOutputImage(graphId, nodeId, 'original', imageFile);
        }

        // If position was set from context menu, update node position
        if (addNodeModalElement.dataset.canvasX && addNodeModalElement.dataset.canvasY) {
            const x = parseFloat(addNodeModalElement.dataset.canvasX);
            const y = parseFloat(addNodeModalElement.dataset.canvasY);
            renderer.updateNodePosition(nodeId, x, y);

            // Persist the position
            const viewport = renderer.exportViewport();
            const nodePositions = renderer.exportNodePositions();
            await api.updateUIMetadata(graphId, viewport, nodePositions);

            // Clear the stored position
            delete addNodeModalElement.dataset.canvasX;
            delete addNodeModalElement.dataset.canvasY;
        }

        closeAddNodeModal();
        // Refresh graph to show new node
        await reloadCurrentGraph();
        toastManager.success(`Node "${nodeName}" added successfully`);
    } catch (error) {
        console.error('Failed to add node:', error);
        toastManager.error(`Failed to add node: ${error.message}`);
    }
});

// Edit config modal handlers
openEditConfigModal = function(nodeId) {
    const node = graphState.getNode(nodeId);
    if (!node) return;

    currentNodeId = nodeId;
    editNodeNameInput.value = node.name;

    // Show/hide image upload based on node type
    if (node.type === 'input') {
        editImageUpload.style.display = 'block';
        editImageInput.value = ''; // Clear any previous file selection
    } else {
        editImageUpload.style.display = 'none';
        editImageInput.value = '';
    }

    // Render config fields based on node type
    renderEditConfigFields(node.type, node.config);

    editConfigModal.open();
}

function closeEditConfigModal() {
    editConfigModal.close();
}

function renderEditConfigFields(nodeType, currentConfig) {
    formBuilder.renderFields(editConfigFields, nodeType, 'edit-config', currentConfig);
}

function getEditConfigValues() {
    return formBuilder.getValues(editConfigFields);
}

editConfigCancelBtn.addEventListener('click', () => {
    closeEditConfigModal();
});

editConfigSaveBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId || !currentNodeId) return;

    const node = graphState.getNode(currentNodeId);
    const newName = editNodeNameInput.value.trim();

    // Validate config fields
    const validation = formBuilder.validate(editConfigFields, node.type);
    if (!validation.valid) {
        toastManager.warning(validation.errors[0]);
        return;
    }

    const config = getEditConfigValues();

    // Determine what changed
    const nameChanged = newName !== node.name;
    const configChanged = JSON.stringify(config) !== JSON.stringify(node.config);
    const imageChanged = node.type === 'input' && editImageInput.files.length > 0;

    if (!nameChanged && !configChanged && !imageChanged) {
        closeEditConfigModal();
        return;
    }

    try {
        // Update name and/or config if changed
        if (nameChanged || configChanged) {
            await api.updateNode(
                graphId,
                currentNodeId,
                nameChanged ? newName : null,
                configChanged ? config : null
            );
        }

        // Upload new image if selected for input node
        if (imageChanged) {
            const imageFile = editImageInput.files[0];
            await api.uploadNodeOutputImage(graphId, currentNodeId, 'original', imageFile);
        }

        closeEditConfigModal();
        // Refresh graph to show updates
        await reloadCurrentGraph();
        toastManager.success('Node updated successfully');
    } catch (error) {
        console.error('Failed to update node:', error);
        toastManager.error(`Failed to update node: ${error.message}`);
    }
});

// View image modal
const viewImageTitle = document.getElementById('view-image-title');
const viewImageImg = document.getElementById('view-image-img');
const viewImageMessage = document.getElementById('view-image-message');
const viewImageCloseBtn = document.getElementById('view-image-close-btn');

const viewImageModal = new Modal('view-image-modal', {
    onOpen: () => {
        interactions.cancelAllDrags();
    },
    onClose: () => {
        viewImageImg.src = '';
        viewImageImg.onerror = null;
        currentNodeId = null;
    }
});
modalManager.register(viewImageModal);

// Delete node modal handlers
function openDeleteNodeModal(nodeId) {
    const node = graphState.getNode(nodeId);
    if (!node) return;

    currentNodeId = nodeId;
    deleteNodeName.textContent = node.name;
    deleteNodeModal.open();
}

function closeDeleteNodeModal() {
    deleteNodeModal.close();
}

deleteNodeCancelBtn.addEventListener('click', () => {
    closeDeleteNodeModal();
});

deleteNodeConfirmBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId || !currentNodeId) return;

    try {
        const node = graphState.getNode(currentNodeId);
        const nodeName = node ? node.name : 'Node';

        await api.deleteNode(graphId, currentNodeId);
        closeDeleteNodeModal();
        // Refresh graph to show node removed
        await reloadCurrentGraph();
        toastManager.success(`"${nodeName}" deleted successfully`);
    } catch (error) {
        console.error('Failed to delete node:', error);
        toastManager.error(`Failed to delete node: ${error.message}`);
    }
});

function openViewImageModal(nodeId) {
    const node = graphState.getNode(nodeId);
    if (!node) return;

    currentNodeId = nodeId;
    viewImageTitle.textContent = `${node.name} - Output`;

    // Get the first output
    const outputs = node.outputs || [];
    if (outputs.length === 0 || !outputs[0].image_id) {
        // No output image available
        viewImageImg.style.display = 'none';
        viewImageMessage.textContent = 'No output image available for this node.';
        viewImageMessage.style.display = 'block';
    } else {
        // Load and display the image
        const imageId = outputs[0].image_id;
        const imageUrl = `/api/images/${imageId}`;

        viewImageImg.src = imageUrl;
        viewImageImg.style.display = 'block';
        viewImageMessage.style.display = 'none';

        // Handle image load error
        viewImageImg.onerror = () => {
            viewImageImg.style.display = 'none';
            viewImageMessage.textContent = 'Failed to load image.';
            viewImageMessage.style.display = 'block';
        };
    }

    viewImageModal.open();
}

function closeViewImageModal() {
    viewImageModal.close();
}

viewImageCloseBtn.addEventListener('click', () => {
    closeViewImageModal();
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
        await reloadCurrentGraph();
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
        await reloadCurrentGraph();
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
        openAddNodeModalAtPosition(nodeType, contextMenuPosition);
    }
});

// Handle node context menu actions
nodeContextMenu.addEventListener('click', (e) => {
    const actionItem = e.target.closest('[data-action]');
    if (actionItem && contextMenuNodeId) {
        const action = actionItem.getAttribute('data-action');
        nodeContextMenu.classList.remove('active');

        if (action === 'view') {
            openViewImageModal(contextMenuNodeId);
        } else if (action === 'edit-config') {
            openEditConfigModal(contextMenuNodeId);
        } else if (action === 'delete') {
            openDeleteNodeModal(contextMenuNodeId);
        }

        contextMenuNodeId = null;
    }
});

// Open add node modal with pre-selected type and position
function openAddNodeModalAtPosition(nodeType, position) {
    if (!graphState.getCurrentGraphId()) {
        toastManager.warning('Please select a graph first');
        return;
    }

    // Store the position for use when creating the node
    addNodeModalElement.dataset.canvasX = position.x;
    addNodeModalElement.dataset.canvasY = position.y;

    // Open the modal with the specified node type
    openAddNodeModal(nodeType);
}

// WebSocket connection management functions
function connectWebSocket(graphId) {
    // Disconnect existing connection if any
    disconnectWebSocket();

    // Determine WebSocket URL (ws:// for http://, wss:// for https://)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/imagegraphs/${graphId}/ws`;

    console.log('Connecting to WebSocket:', wsUrl);

    try {
        wsConnection = new WebSocket(wsUrl);

        wsConnection.onopen = () => {
            console.log('WebSocket connected for graph:', graphId);
            // Clear any pending reconnect attempts
            if (wsReconnectTimeout) {
                clearTimeout(wsReconnectTimeout);
                wsReconnectTimeout = null;
            }
        };

        wsConnection.onmessage = async (event) => {
            try {
                const message = JSON.parse(event.data);
                console.log('WebSocket message received:', message);

                // Refresh the graph to get the latest state
                await reloadCurrentGraph();
            } catch (error) {
                console.error('Failed to handle WebSocket message:', error);
            }
        };

        wsConnection.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        wsConnection.onclose = (event) => {
            console.log('WebSocket closed:', event.code, event.reason);
            wsConnection = null;

            // Attempt to reconnect if the graph is still selected
            const currentGraphId = graphState.getCurrentGraphId();
            if (currentGraphId === graphId) {
                console.log(`Will attempt to reconnect in ${WS_RECONNECT_DELAY}ms`);
                wsReconnectTimeout = setTimeout(() => {
                    connectWebSocket(graphId);
                }, WS_RECONNECT_DELAY);
            }
        };
    } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
    }
}

function disconnectWebSocket() {
    // Clear any pending reconnect attempts
    if (wsReconnectTimeout) {
        clearTimeout(wsReconnectTimeout);
        wsReconnectTimeout = null;
    }

    // Close existing connection
    if (wsConnection) {
        console.log('Disconnecting WebSocket');
        wsConnection.close(1000, 'Client disconnecting');
        wsConnection = null;
    }
}

// Clean up WebSocket on page unload
window.addEventListener('beforeunload', () => {
    disconnectWebSocket();
});

// Render output nodes in sidebar
renderOutputs = function(graph) {
    const sidebarContent = document.querySelector('.sidebar-content');

    if (!graph) {
        sidebarContent.innerHTML = '<p style="color: #7f8c8d; text-align: center; margin-top: 20px;">No graph selected</p>';
        return;
    }

    // Filter output nodes
    const outputNodes = graph.nodes.filter(node => node.type === 'output');

    if (outputNodes.length === 0) {
        sidebarContent.innerHTML = '<p style="color: #7f8c8d; text-align: center; margin-top: 20px;">No output nodes</p>';
        return;
    }

    // Sort output nodes by y position (top to bottom)
    const sortedOutputNodes = outputNodes.sort((a, b) => {
        const posA = renderer.getNodePosition(a.id);
        const posB = renderer.getNodePosition(b.id);

        // If positions aren't available, maintain original order
        if (!posA || !posB) return 0;
        return posA.y - posB.y;
    });

    // Render each output node
    sidebarContent.innerHTML = sortedOutputNodes.map(node => {
        const output = node.outputs?.find(o => o.name === 'final');
        const hasImage = output?.image_id;
        const imageUrl = hasImage ? `/api/images/${output.image_id}` : '';

        return `
            <div class="output-card">
                <div class="output-card-header">${node.name}</div>
                <div class="output-card-body">
                    ${hasImage
                ? `<img src="${imageUrl}" alt="${node.name}" class="output-card-image" />`
                : '<p class="output-card-placeholder">No image yet</p>'}
                </div>
            </div>
        `;
    }).join('');
}

// Sidebar resize functionality
const sidebar = document.getElementById('outputs-sidebar');
const resizeHandle = document.getElementById('sidebar-resize-handle');
let isResizing = false;
let startX = 0;
let startWidth = 0;

const SIDEBAR_MIN_WIDTH = 200;
const SIDEBAR_MAX_WIDTH = 600;
const SIDEBAR_WIDTH_KEY = 'artwork-sidebar-width';

// Restore saved width from localStorage
const savedWidth = localStorage.getItem(SIDEBAR_WIDTH_KEY);
if (savedWidth) {
    const width = parseInt(savedWidth, 10);
    if (width >= SIDEBAR_MIN_WIDTH && width <= SIDEBAR_MAX_WIDTH) {
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
    const constrainedWidth = Math.max(SIDEBAR_MIN_WIDTH, Math.min(SIDEBAR_MAX_WIDTH, newWidth));

    sidebar.style.width = `${constrainedWidth}px`;
});

document.addEventListener('mouseup', () => {
    if (isResizing) {
        isResizing = false;
        document.body.classList.remove('resizing-active');
        resizeHandle.classList.remove('resizing');

        // Save width to localStorage
        localStorage.setItem(SIDEBAR_WIDTH_KEY, sidebar.offsetWidth);
    }
});

// Load initial data
loadGraphList();
