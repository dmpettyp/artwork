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
const graphNameElement = document.getElementById('graph-name');
const createGraphBtn = document.getElementById('create-graph-btn');
const refreshBtn = document.getElementById('refresh-btn');
const addNodeBtn = document.getElementById('add-node-btn');

// Create graph modal
const createGraphModal = document.getElementById('create-graph-modal');
const graphNameInput = document.getElementById('graph-name-input');
const modalCreateBtn = document.getElementById('modal-create-btn');
const modalCancelBtn = document.getElementById('modal-cancel-btn');

// Add node modal
const addNodeModal = document.getElementById('add-node-modal');
const nodeTypeSelect = document.getElementById('node-type-select');
const nodeNameInput = document.getElementById('node-name-input');
const nodeConfigFields = document.getElementById('node-config-fields');
const addNodeCreateBtn = document.getElementById('add-node-create-btn');
const addNodeCancelBtn = document.getElementById('add-node-cancel-btn');

// Subscribe to graph state changes
graphState.subscribe((graph) => {
    if (graph) {
        graphNameElement.textContent = graph.name;
        renderer.render(graph);
    } else {
        graphNameElement.textContent = 'No graph selected';
        renderer.clear();
    }
});

// Load and display graph list
async function loadGraphList() {
    try {
        const graphs = await api.listImageGraphs();
        renderGraphList(graphs);
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
        console.log(graph);
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
    addNodeModal.classList.add('active');
    nodeTypeSelect.value = '';
    nodeNameInput.value = '';
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

createGraphModal.addEventListener('click', (e) => {
    if (e.target === createGraphModal) {
        closeCreateGraphModal();
    }
});

// Add node handlers
addNodeBtn.addEventListener('click', () => {
    openAddNodeModal();
});

nodeTypeSelect.addEventListener('change', (e) => {
    const nodeType = e.target.value;
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

    const config = getNodeConfig();

    try {
        // The API expects config as a JSON string
        await api.addNode(graphId, nodeType, nodeName, JSON.stringify(config));
        closeAddNodeModal();
        // Refresh graph to show new node
        const graph = await api.getImageGraph(graphId);
        graphState.setCurrentGraph(graph);
    } catch (error) {
        console.error('Failed to add node:', error);
        alert(`Failed to add node: ${error.message}`);
    }
});

addNodeModal.addEventListener('click', (e) => {
    if (e.target === addNodeModal) {
        closeAddNodeModal();
    }
});

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

// Auto-refresh every 2 seconds if a graph is selected
setInterval(async () => {
    const graphId = graphState.getCurrentGraphId();
    if (graphId) {
        try {
            const graph = await api.getImageGraph(graphId);
            graphState.setCurrentGraph(graph);
        } catch (error) {
            console.error('Auto-refresh failed:', error);
        }
    }
}, 2000);

// Load initial data
loadGraphList();
