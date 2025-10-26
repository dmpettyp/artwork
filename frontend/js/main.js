// Main application entry point

import * as api from './api.js';
import { GraphState } from './graph.js';
import { Renderer } from './renderer.js';
import { InteractionHandler } from './interactions.js';
import { Modal, ModalManager } from './modal.js';
import { ToastManager } from './toast.js';
import { NodeConfigFormBuilder } from './form-builder.js';
import { GraphManager } from './graph-manager.js';
import { API_PATHS, SIDEBAR_CONFIG } from './constants.js';
import { loadNodeTypeSchemas } from './node-type-schemas.js';
import { setNodeTypeConfigs, getNodeTypeConfigs } from './node-type-config-store.js';

// Initialize state and rendering
const graphState = new GraphState();
const svg = document.getElementById('graph-canvas');
const nodesLayer = document.getElementById('nodes-layer');
const connectionsLayer = document.getElementById('connections-layer');
const renderer = new Renderer(svg, nodesLayer, connectionsLayer, graphState);

// Initialize modal manager and toast manager early so they can be passed to InteractionHandler
const modalManager = new ModalManager();
const toastManager = new ToastManager();

// Initialize graph manager
const graphManager = new GraphManager(api, graphState, renderer, toastManager);

// Declare functions early so they can be passed to InteractionHandler
let renderOutputs;
let openEditConfigModal;

const interactions = new InteractionHandler(
    svg,
    renderer,
    graphState,
    api,
    (graph) => renderOutputs(graph),
    (nodeId) => openEditConfigModal(nodeId),
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

// Modal factory function to reduce duplication
function createAndRegisterModal(modalId, options = {}) {
    const modal = new Modal(modalId, {
        onOpen: () => {
            interactions.cancelAllDrags();
            if (options.onOpen) options.onOpen();
        },
        onClose: options.onClose,
        beforeClose: options.beforeClose
    });
    modalManager.register(modal);
    return modal;
}

// Create graph modal
const graphNameInput = document.getElementById('graph-name-input');
const modalCreateBtn = document.getElementById('modal-create-btn');
const modalCancelBtn = document.getElementById('modal-cancel-btn');

const createGraphModal = createAndRegisterModal('create-graph-modal', {
    onOpen: () => {
        graphNameInput.value = '';
        graphNameInput.focus();
    }
});

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

const addNodeModal = createAndRegisterModal('add-node-modal');

// Edit config modal
const editConfigModalTitle = document.getElementById('edit-config-modal-title');
const editNodeNameInput = document.getElementById('edit-node-name-input');
const editImageUpload = document.getElementById('edit-image-upload');
const editImageInput = document.getElementById('edit-image-input');
const editConfigFields = document.getElementById('edit-config-fields');
const editConfigSaveBtn = document.getElementById('edit-config-save-btn');
const editConfigCancelBtn = document.getElementById('edit-config-cancel-btn');

const editConfigModal = createAndRegisterModal('edit-config-modal', {
    onClose: () => {
        currentNodeId = null;
    }
});

// Delete node modal
const deleteNodeName = document.getElementById('delete-node-name');
const deleteNodeConfirmBtn = document.getElementById('delete-node-confirm-btn');
const deleteNodeCancelBtn = document.getElementById('delete-node-cancel-btn');

const deleteNodeModal = createAndRegisterModal('delete-node-modal', {
    onClose: () => {
        currentNodeId = null;
    }
});

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

// Form builder will be initialized after loading schemas
let formBuilder = null;

// Add node modal functions
function openAddNodeModal(nodeType) {
    if (!graphState.getCurrentGraphId()) {
        toastManager.warning('Please select a graph first');
        return;
    }

    // Store the node type
    addNodeType = nodeType;

    // Update modal title with display name from config
    const configs = getNodeTypeConfigs();
    const displayName = configs[nodeType]?.name || nodeType;
    addNodeModalTitle.textContent = `Add ${displayName} Node`;

    // Clear inputs
    nodeNameInput.value = '';
    nodeImageInput.value = '';
    nodeConfigFields.innerHTML = '';

    // Update name field based on whether it's required
    const nameRequired = configs[nodeType]?.nameRequired !== false;
    nodeNameInput.required = nameRequired;
    nodeNameInput.placeholder = nameRequired ? 'Enter node name' : 'Enter node name (optional)';

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

function renderNodeConfigFields(nodeType) {
    formBuilder.renderFields(nodeConfigFields, nodeType, 'config');
}

function getNodeConfig() {
    return formBuilder.getValues(nodeConfigFields);
}

// Create new graph handlers
createGraphBtn.addEventListener('click', () => {
    createGraphModal.open();
});

modalCancelBtn.addEventListener('click', () => {
    createGraphModal.close();
});

modalCreateBtn.addEventListener('click', async () => {
    const name = graphNameInput.value.trim();
    if (!name) return;

    try {
        const graph_id = await api.createImageGraph(name);
        createGraphModal.close();
        await graphManager.loadGraphList();
        await graphManager.selectGraph(graph_id);
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
    addNodeModal.close();
    addNodeType = null;
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

    // Check if name is required for this node type
    const configs = getNodeTypeConfigs();
    const nameRequired = configs[nodeType]?.nameRequired !== false;
    if (nameRequired && !nodeName) {
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

            // Persist the layout (node positions)
            const nodePositions = renderer.exportNodePositions();
            await api.updateLayout(graphId, nodePositions);

            // Clear the stored position
            delete addNodeModalElement.dataset.canvasX;
            delete addNodeModalElement.dataset.canvasY;
        }

        addNodeModal.close();
        addNodeType = null;
        // Refresh graph to show new node
        await graphManager.reloadCurrentGraph();
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
    editNodeNameInput.value = node.name || '';

    // Update modal title with display name from config
    const configs = getNodeTypeConfigs();
    const displayName = configs[node.type]?.name || node.type;
    editConfigModalTitle.textContent = `Edit ${displayName} Node`;

    // Update name field based on whether it's required
    const nameRequired = configs[node.type]?.nameRequired !== false;
    editNodeNameInput.required = nameRequired;
    editNodeNameInput.placeholder = nameRequired ? 'Enter node name' : 'Enter node name (optional)';

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

function renderEditConfigFields(nodeType, currentConfig) {
    formBuilder.renderFields(editConfigFields, nodeType, 'edit-config', currentConfig);
}

function getEditConfigValues() {
    return formBuilder.getValues(editConfigFields);
}

editConfigCancelBtn.addEventListener('click', () => {
    editConfigModal.close();
});

editConfigSaveBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId || !currentNodeId) return;

    const node = graphState.getNode(currentNodeId);
    const newName = editNodeNameInput.value.trim();

    // Check if name is required for this node type
    const configs = getNodeTypeConfigs();
    const nameRequired = configs[node.type]?.nameRequired !== false;
    if (nameRequired && !newName) {
        toastManager.warning('Please enter a node name');
        return;
    }

    // Validate config fields
    const validation = formBuilder.validate(editConfigFields, node.type);
    if (!validation.valid) {
        toastManager.warning(validation.errors[0]);
        return;
    }

    const config = getEditConfigValues();

    // Determine what changed (handle empty string vs null for optional names)
    const oldName = node.name || '';
    const nameChanged = newName !== oldName;
    const configChanged = JSON.stringify(config) !== JSON.stringify(node.config);
    const imageChanged = node.type === 'input' && editImageInput.files.length > 0;

    if (!nameChanged && !configChanged && !imageChanged) {
        editConfigModal.close();
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

        editConfigModal.close();
        // Refresh graph to show updates
        await graphManager.reloadCurrentGraph();
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

const viewImageModal = createAndRegisterModal('view-image-modal', {
    onClose: () => {
        viewImageImg.src = '';
        viewImageImg.onerror = null;
        currentNodeId = null;
    }
});

// Delete node modal handlers
function openDeleteNodeModal(nodeId) {
    const node = graphState.getNode(nodeId);
    if (!node) return;

    currentNodeId = nodeId;
    deleteNodeName.textContent = node.name;
    deleteNodeModal.open();
}

deleteNodeCancelBtn.addEventListener('click', () => {
    deleteNodeModal.close();
});

deleteNodeConfirmBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId || !currentNodeId) return;

    try {
        const node = graphState.getNode(currentNodeId);
        const nodeName = node ? node.name : 'Node';

        await api.deleteNode(graphId, currentNodeId);
        deleteNodeModal.close();
        // Refresh graph to show node removed
        await graphManager.reloadCurrentGraph();
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
        const imageUrl = API_PATHS.images(imageId);

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

viewImageCloseBtn.addEventListener('click', () => {
    viewImageModal.close();
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

// Clean up WebSocket on page unload
window.addEventListener('beforeunload', () => {
    graphManager.cleanup();
});

// Helper function to handle image download
async function handleImageDownload(imageId, nodeName) {
    try {
        const response = await fetch(API_PATHS.images(imageId));
        const blob = await response.blob();

        // Determine extension from content type
        const contentType = response.headers.get('content-type') || 'image/png';
        const extension = contentType.split('/')[1] || 'png';

        // Create filename: {imagegraph_name}-{output_node_name}.{extension}
        // Convert to lowercase and replace spaces with underscores
        const currentGraph = graphState.getCurrentGraph();
        const graphName = currentGraph?.name || 'graph';
        const sanitizedGraphName = graphName.toLowerCase().replace(/ /g, '_');
        const sanitizedNodeName = nodeName.toLowerCase().replace(/ /g, '_');
        const filename = `${sanitizedGraphName}-${sanitizedNodeName}.${extension}`;

        // Create download link and trigger
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);

        // Show success toast
        toastManager.success(`Downloaded ${filename}`);
    } catch (error) {
        console.error('Failed to download image:', error);
        toastManager.error('Failed to download image');
    }
}

// Helper function to create an output card for a node
function createOutputCard(node, output) {
    const hasImage = output?.image_id;

    // Create output card elements using DOM API (safe from XSS)
    const card = document.createElement('div');
    card.className = 'output-card';

    const header = document.createElement('div');
    header.className = 'output-card-header';
    header.textContent = node.name; // textContent is safe from XSS
    card.appendChild(header);

    const body = document.createElement('div');
    body.className = 'output-card-body';

    if (hasImage) {
        const img = document.createElement('img');
        img.src = API_PATHS.images(output.image_id);
        img.alt = node.name; // alt attribute is also escaped
        img.className = 'output-card-image';
        body.appendChild(img);

        // Add download button
        const downloadBtn = document.createElement('button');
        downloadBtn.className = 'output-card-download-btn';
        downloadBtn.textContent = 'Download';
        downloadBtn.onclick = () => handleImageDownload(output.image_id, node.name);
        body.appendChild(downloadBtn);
    } else {
        const placeholder = document.createElement('p');
        placeholder.className = 'output-card-placeholder';
        placeholder.textContent = 'No image yet';
        body.appendChild(placeholder);
    }

    card.appendChild(body);
    return card;
}

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
    sidebarContent.innerHTML = ''; // Clear existing content

    sortedOutputNodes.forEach(node => {
        const output = node.outputs?.find(o => o.name === 'final');
        const card = createOutputCard(node, output);
        sidebarContent.appendChild(card);
    });
}

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
