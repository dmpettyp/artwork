// Main application entry point

import * as api from './api.js';
import { GraphState } from './graph.js';
import { Renderer } from './renderer.js';
import { InteractionHandler } from './interactions.js';

// Initialize state and rendering
const graphState = new GraphState();
const svg = document.getElementById('graph-canvas');
const nodesLayer = document.getElementById('nodes-layer');
const connectionsLayer = document.getElementById('connections-layer');
const renderer = new Renderer(svg, nodesLayer, connectionsLayer);
const interactions = new InteractionHandler(svg, renderer, graphState, api);

// UI elements
const graphSelect = document.getElementById('graph-select');
const createGraphBtn = document.getElementById('create-graph-btn');
const refreshBtn = document.getElementById('refresh-btn');

// Context menu
const contextMenu = document.getElementById('context-menu');
let contextMenuPosition = { x: 0, y: 0 };

// Create graph modal
const createGraphModal = document.getElementById('create-graph-modal');
const graphNameInput = document.getElementById('graph-name-input');
const modalCreateBtn = document.getElementById('modal-create-btn');
const modalCancelBtn = document.getElementById('modal-cancel-btn');

// Add node modal
const addNodeModal = document.getElementById('add-node-modal');
const nodeTypeSelect = document.getElementById('node-type-select');
const nodeNameInput = document.getElementById('node-name-input');
const nodeImageUpload = document.getElementById('node-image-upload');
const nodeImageInput = document.getElementById('node-image-input');
const nodeConfigFields = document.getElementById('node-config-fields');
const addNodeCreateBtn = document.getElementById('add-node-create-btn');
const addNodeCancelBtn = document.getElementById('add-node-cancel-btn');

// Edit config modal
const editConfigModal = document.getElementById('edit-config-modal');
const editNodeNameInput = document.getElementById('edit-node-name-input');
const editConfigFields = document.getElementById('edit-config-fields');
const editConfigSaveBtn = document.getElementById('edit-config-save-btn');
const editConfigCancelBtn = document.getElementById('edit-config-cancel-btn');

// Delete node modal
const deleteNodeModal = document.getElementById('delete-node-modal');
const deleteNodeName = document.getElementById('delete-node-name');
const deleteNodeConfirmBtn = document.getElementById('delete-node-confirm-btn');
const deleteNodeCancelBtn = document.getElementById('delete-node-cancel-btn');

// Track current node being edited/deleted
let currentNodeId = null;

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
    } catch (error) {
        console.error('Failed to load graph:', error);
        alert(`Failed to load graph: ${error.message}`);
    }
}

// Node type configuration (matches backend node_type.go)
const nodeTypeConfigs = {
    input: {
        fields: {}
    },
    scale: {
        fields: {
            factor: { type: 'float', required: true }
        }
    }
};

// Create graph modal functions
function openCreateGraphModal() {
    interactions.cancelAllDrags();
    createGraphModal.classList.add('active');
    graphNameInput.value = '';
    graphNameInput.focus();
}

function closeCreateGraphModal() {
    createGraphModal.classList.remove('active');
}

// Add node modal functions
function openAddNodeModal() {
    if (!graphState.getCurrentGraphId()) {
        alert('Please select a graph first');
        return;
    }
    interactions.cancelAllDrags();
    addNodeModal.classList.add('active');
    nodeTypeSelect.value = '';
    nodeNameInput.value = '';
    nodeImageInput.value = '';
    nodeImageUpload.style.display = 'none';
    nodeConfigFields.innerHTML = '';
    nodeTypeSelect.focus();
}

function closeAddNodeModal() {
    addNodeModal.classList.remove('active');
}

function renderNodeConfigFields(nodeType) {
    nodeConfigFields.innerHTML = '';

    const config = nodeTypeConfigs[nodeType];
    if (!config || !config.fields) return;

    Object.entries(config.fields).forEach(([fieldName, fieldDef]) => {
        const label = document.createElement('label');
        label.setAttribute('for', `config-${fieldName}`);
        label.textContent = `${fieldName}${fieldDef.required ? ' *' : ''}`;

        const input = document.createElement('input');
        input.id = `config-${fieldName}`;
        input.className = 'form-input';
        input.setAttribute('data-field-name', fieldName);
        input.setAttribute('data-field-type', fieldDef.type);

        if (fieldDef.type === 'float' || fieldDef.type === 'int') {
            input.type = 'number';
            if (fieldDef.type === 'float') {
                input.step = 'any';
            }
        } else if (fieldDef.type === 'bool') {
            input.type = 'checkbox';
        } else {
            input.type = 'text';
        }

        if (fieldDef.required) {
            input.required = true;
        }

        nodeConfigFields.appendChild(label);
        nodeConfigFields.appendChild(input);
    });
}

function getNodeConfig() {
    const config = {};
    const inputs = nodeConfigFields.querySelectorAll('input');

    inputs.forEach(input => {
        const fieldName = input.getAttribute('data-field-name');
        const fieldType = input.getAttribute('data-field-type');
        let value = input.value;

        if (fieldType === 'int') {
            value = parseInt(value, 10);
        } else if (fieldType === 'float') {
            value = parseFloat(value);
        } else if (fieldType === 'bool') {
            value = input.checked;
        }

        if (value !== '' && !isNaN(value)) {
            config[fieldName] = value;
        }
    });

    return config;
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
    } catch (error) {
        console.error('Failed to create graph:', error);
        alert(`Failed to create graph: ${error.message}`);
    }
});

graphNameInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        modalCreateBtn.click();
    }
});

// Close modal only if both mousedown and mouseup happen on the background
let createGraphModalMousedownTarget = null;
createGraphModal.addEventListener('mousedown', (e) => {
    createGraphModalMousedownTarget = e.target;
});
createGraphModal.addEventListener('mouseup', (e) => {
    if (createGraphModalMousedownTarget === createGraphModal && e.target === createGraphModal) {
        closeCreateGraphModal();
    }
    createGraphModalMousedownTarget = null;
});

// Add node handlers
nodeTypeSelect.addEventListener('change', (e) => {
    const nodeType = e.target.value;

    // Show/hide image upload based on node type
    if (nodeType === 'input') {
        nodeImageUpload.style.display = 'block';
    } else {
        nodeImageUpload.style.display = 'none';
        nodeImageInput.value = '';
    }

    renderNodeConfigFields(nodeType);
});

addNodeCancelBtn.addEventListener('click', () => {
    closeAddNodeModal();
});

addNodeCreateBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) return;

    const nodeType = nodeTypeSelect.value;
    const nodeName = nodeNameInput.value.trim();

    if (!nodeType) {
        alert('Please select a node type');
        return;
    }

    if (!nodeName) {
        alert('Please enter a node name');
        return;
    }

    // For input nodes, check if an image file is selected
    if (nodeType === 'input' && nodeImageInput.files.length === 0) {
        alert('Please select an image file for the input node');
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
        if (addNodeModal.dataset.canvasX && addNodeModal.dataset.canvasY) {
            const x = parseFloat(addNodeModal.dataset.canvasX);
            const y = parseFloat(addNodeModal.dataset.canvasY);
            renderer.updateNodePosition(nodeId, x, y);

            // Persist the position
            const viewport = renderer.exportViewport();
            const nodePositions = renderer.exportNodePositions();
            await api.updateUIMetadata(graphId, viewport, nodePositions);

            // Clear the stored position
            delete addNodeModal.dataset.canvasX;
            delete addNodeModal.dataset.canvasY;
        }

        closeAddNodeModal();
        // Refresh graph to show new node
        const graph = await api.getImageGraph(graphId);
        graphState.setCurrentGraph(graph);
    } catch (error) {
        console.error('Failed to add node:', error);
        alert(`Failed to add node: ${error.message}`);
    }
});

// Close modal only if both mousedown and mouseup happen on the background
let addNodeModalMousedownTarget = null;
addNodeModal.addEventListener('mousedown', (e) => {
    addNodeModalMousedownTarget = e.target;
});
addNodeModal.addEventListener('mouseup', (e) => {
    if (addNodeModalMousedownTarget === addNodeModal && e.target === addNodeModal) {
        closeAddNodeModal();
    }
    addNodeModalMousedownTarget = null;
});

// Edit config modal handlers
function openEditConfigModal(nodeId) {
    const node = graphState.getNode(nodeId);
    if (!node) return;

    interactions.cancelAllDrags();
    currentNodeId = nodeId;
    editNodeNameInput.value = node.name;

    // Render config fields based on node type
    renderEditConfigFields(node.type, node.config);

    editConfigModal.classList.add('active');
}

function closeEditConfigModal() {
    editConfigModal.classList.remove('active');
    currentNodeId = null;
}

function renderEditConfigFields(nodeType, currentConfig) {
    editConfigFields.innerHTML = '';

    const config = nodeTypeConfigs[nodeType];
    if (!config || !config.fields) {
        editConfigFields.innerHTML = '<p style="color: #7f8c8d;">This node has no configurable fields.</p>';
        return;
    }

    Object.entries(config.fields).forEach(([fieldName, fieldDef]) => {
        const label = document.createElement('label');
        label.setAttribute('for', `edit-config-${fieldName}`);
        label.textContent = `${fieldName}${fieldDef.required ? ' *' : ''}`;

        const input = document.createElement('input');
        input.id = `edit-config-${fieldName}`;
        input.className = 'form-input';
        input.setAttribute('data-field-name', fieldName);
        input.setAttribute('data-field-type', fieldDef.type);

        if (fieldDef.type === 'float' || fieldDef.type === 'int') {
            input.type = 'number';
            if (fieldDef.type === 'float') {
                input.step = 'any';
            }
        } else if (fieldDef.type === 'bool') {
            input.type = 'checkbox';
        } else {
            input.type = 'text';
        }

        if (fieldDef.required) {
            input.required = true;
        }

        // Set current value
        if (currentConfig && currentConfig[fieldName] !== undefined) {
            if (fieldDef.type === 'bool') {
                input.checked = currentConfig[fieldName];
            } else {
                input.value = currentConfig[fieldName];
            }
        }

        editConfigFields.appendChild(label);
        editConfigFields.appendChild(input);
    });
}

function getEditConfigValues() {
    const config = {};
    const inputs = editConfigFields.querySelectorAll('input');

    inputs.forEach(input => {
        const fieldName = input.getAttribute('data-field-name');
        const fieldType = input.getAttribute('data-field-type');
        let value = input.value;

        if (fieldType === 'int') {
            value = parseInt(value, 10);
        } else if (fieldType === 'float') {
            value = parseFloat(value);
        } else if (fieldType === 'bool') {
            value = input.checked;
        }

        if (value !== '' && !isNaN(value)) {
            config[fieldName] = value;
        }
    });

    return config;
}

editConfigCancelBtn.addEventListener('click', () => {
    closeEditConfigModal();
});

editConfigSaveBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId || !currentNodeId) return;

    const node = graphState.getNode(currentNodeId);
    const newName = editNodeNameInput.value.trim();
    const config = getEditConfigValues();

    // Determine what changed
    const nameChanged = newName !== node.name;
    const configChanged = JSON.stringify(config) !== JSON.stringify(node.config);

    if (!nameChanged && !configChanged) {
        closeEditConfigModal();
        return;
    }

    try {
        // Send both name and config - API will update what's provided
        await api.updateNode(
            graphId,
            currentNodeId,
            nameChanged ? newName : null,
            configChanged ? config : null
        );
        closeEditConfigModal();
        // Refresh graph to show updates
        const graph = await api.getImageGraph(graphId);
        graphState.setCurrentGraph(graph);
    } catch (error) {
        console.error('Failed to update node:', error);
        alert(`Failed to update node: ${error.message}`);
    }
});

// Close modal only if both mousedown and mouseup happen on the background
let editConfigModalMousedownTarget = null;
editConfigModal.addEventListener('mousedown', (e) => {
    editConfigModalMousedownTarget = e.target;
});
editConfigModal.addEventListener('mouseup', (e) => {
    if (editConfigModalMousedownTarget === editConfigModal && e.target === editConfigModal) {
        closeEditConfigModal();
    }
    editConfigModalMousedownTarget = null;
});

// Delete node modal handlers
function openDeleteNodeModal(nodeId) {
    const node = graphState.getNode(nodeId);
    if (!node) return;

    interactions.cancelAllDrags();
    currentNodeId = nodeId;
    deleteNodeName.textContent = node.name;

    deleteNodeModal.classList.add('active');
}

function closeDeleteNodeModal() {
    deleteNodeModal.classList.remove('active');
    currentNodeId = null;
}

deleteNodeCancelBtn.addEventListener('click', () => {
    closeDeleteNodeModal();
});

deleteNodeConfirmBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId || !currentNodeId) return;

    try {
        await api.deleteNode(graphId, currentNodeId);
        closeDeleteNodeModal();
        // Refresh graph to show node removed
        const graph = await api.getImageGraph(graphId);
        graphState.setCurrentGraph(graph);
    } catch (error) {
        console.error('Failed to delete node:', error);
        alert(`Failed to delete node: ${error.message}`);
    }
});

// Close modal only if both mousedown and mouseup happen on the background
let deleteNodeModalMousedownTarget = null;
deleteNodeModal.addEventListener('mousedown', (e) => {
    deleteNodeModalMousedownTarget = e.target;
});
deleteNodeModal.addEventListener('mouseup', (e) => {
    if (deleteNodeModalMousedownTarget === deleteNodeModal && e.target === deleteNodeModal) {
        closeDeleteNodeModal();
    }
    deleteNodeModalMousedownTarget = null;
});

// Handle node action button clicks
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

    // Check for node action buttons
    const actionBtn = e.target.closest('.node-action-btn');
    if (!actionBtn) return;

    const action = actionBtn.getAttribute('data-action');
    const nodeElement = actionBtn.closest('.node');
    const nodeId = nodeElement.getAttribute('data-node-id');

    if (action === 'delete') {
        openDeleteNodeModal(nodeId);
    } else if (action === 'edit-config') {
        openEditConfigModal(nodeId);
    }

    e.stopPropagation();
});

// Handle connection disconnect
async function handleDisconnectConnection(fromNodeId, fromOutput, toNodeId, toInput) {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) return;

    try {
        await api.disconnectNodes(graphId, fromNodeId, fromOutput, toNodeId, toInput);
        // Refresh graph to show connection removed
        const graph = await api.getImageGraph(graphId);
        graphState.setCurrentGraph(graph);
    } catch (error) {
        console.error('Failed to disconnect nodes:', error);
        alert(`Failed to disconnect nodes: ${error.message}`);
    }
}

// Refresh current graph
refreshBtn.addEventListener('click', async () => {
    const graphId = graphState.getCurrentGraphId();
    if (!graphId) return;

    try {
        const graph = await api.getImageGraph(graphId);
        graphState.setCurrentGraph(graph);
    } catch (error) {
        console.error('Failed to refresh graph:', error);
        alert(`Failed to refresh graph: ${error.message}`);
    }
});

// Context menu handlers
svg.addEventListener('contextmenu', (e) => {
    e.preventDefault();

    // Only show context menu if a graph is selected
    if (!graphState.getCurrentGraphId()) {
        return;
    }

    // Check if right-clicking on canvas (not on a node or connection)
    const clickedNode = e.target.closest('.node');
    const clickedConnection = e.target.closest('.connection-group');

    if (!clickedNode && !clickedConnection) {
        // Store the canvas position where the user right-clicked
        const svgRect = svg.getBoundingClientRect();
        const screenX = e.clientX - svgRect.left;
        const screenY = e.clientY - svgRect.top;
        contextMenuPosition = interactions.screenToCanvas(screenX, screenY);

        // Position and show the context menu with slight offset
        // so cursor is inside the menu when it appears
        contextMenu.style.left = `${e.clientX - 5}px`;
        contextMenu.style.top = `${e.clientY - 5}px`;
        contextMenu.classList.add('active');
    }
});

// Close context menu when clicking anywhere else
document.addEventListener('click', (e) => {
    if (!e.target.closest('.context-menu')) {
        contextMenu.classList.remove('active');
    }
});

// Close context menu when mouse leaves it
contextMenu.addEventListener('mouseleave', () => {
    contextMenu.classList.remove('active');
});

// Handle context menu node type selection
contextMenu.addEventListener('click', (e) => {
    const nodeTypeItem = e.target.closest('[data-node-type]');
    if (nodeTypeItem) {
        const nodeType = nodeTypeItem.getAttribute('data-node-type');
        contextMenu.classList.remove('active');
        openAddNodeModalAtPosition(nodeType, contextMenuPosition);
    }
});

// Open add node modal with pre-selected type and position
function openAddNodeModalAtPosition(nodeType, position) {
    if (!graphState.getCurrentGraphId()) {
        alert('Please select a graph first');
        return;
    }

    interactions.cancelAllDrags();

    // Store the position for use when creating the node
    addNodeModal.dataset.canvasX = position.x;
    addNodeModal.dataset.canvasY = position.y;

    addNodeModal.classList.add('active');
    nodeTypeSelect.value = nodeType;
    nodeNameInput.value = '';
    nodeImageInput.value = '';

    // Trigger the change event to show/hide appropriate fields
    nodeTypeSelect.dispatchEvent(new Event('change'));

    nodeNameInput.focus();
}

// Handle ESC key to close modals
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        // Close whichever modal is currently open
        if (createGraphModal.classList.contains('active')) {
            closeCreateGraphModal();
        } else if (addNodeModal.classList.contains('active')) {
            closeAddNodeModal();
        } else if (editConfigModal.classList.contains('active')) {
            closeEditConfigModal();
        } else if (deleteNodeModal.classList.contains('active')) {
            closeDeleteNodeModal();
        } else if (contextMenu.classList.contains('active')) {
            contextMenu.classList.remove('active');
        }
    }
});

// Load initial data
loadGraphList();
