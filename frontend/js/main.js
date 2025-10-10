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
const graphList = document.getElementById('graph-list');
const graphNameElement = document.getElementById('graph-name');
const createGraphBtn = document.getElementById('create-graph-btn');
const refreshBtn = document.getElementById('refresh-btn');
const addNodeButtons = document.querySelectorAll('[data-node-type]');
const modal = document.getElementById('create-graph-modal');
const graphNameInput = document.getElementById('graph-name-input');
const modalCreateBtn = document.getElementById('modal-create-btn');
const modalCancelBtn = document.getElementById('modal-cancel-btn');

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
    graphList.innerHTML = '';

    if (graphs.length === 0) {
        graphList.innerHTML = '<p style="color: #95a5a6; font-size: 12px; padding: 10px;">No graphs yet</p>';
        return;
    }

    graphs.forEach(graph => {
        const item = document.createElement('div');
        item.className = 'graph-item';
        item.textContent = graph.name;
        item.addEventListener('click', () => selectGraph(graph.id));

        if (graph.id === graphState.getCurrentGraphId()) {
            item.classList.add('active');
        }

        graphList.appendChild(item);
    });
}

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

// Modal functions
function openModal() {
    modal.classList.add('active');
    graphNameInput.value = '';
    graphNameInput.focus();
}

function closeModal() {
    modal.classList.remove('active');
}

// Create new graph
createGraphBtn.addEventListener('click', () => {
    openModal();
});

modalCancelBtn.addEventListener('click', () => {
    closeModal();
});

modalCreateBtn.addEventListener('click', async () => {
    const name = graphNameInput.value.trim();
    if (!name) return;

    try {
        const graph_id = await api.createImageGraph(name);
        closeModal();
        await loadGraphList();
        await selectGraph(graph_id);
    } catch (error) {
        console.error('Failed to create graph:', error);
        alert(`Failed to create graph: ${error.message}`);
    }
});

// Allow Enter key to submit modal
graphNameInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        modalCreateBtn.click();
    }
});

// Close modal on background click
modal.addEventListener('click', (e) => {
    if (e.target === modal) {
        closeModal();
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

// Add node buttons
addNodeButtons.forEach(button => {
    button.addEventListener('click', async () => {
        const graphId = graphState.getCurrentGraphId();
        if (!graphId) {
            alert('Please select a graph first');
            return;
        }

        const nodeType = button.getAttribute('data-node-type');

        try {
            await api.addNode(graphId, nodeType);
            // Refresh graph to show new node
            const graph = await api.getImageGraph(graphId);
            graphState.setCurrentGraph(graph);
        } catch (error) {
            console.error('Failed to add node:', error);
            alert(`Failed to add node: ${error.message}`);
        }
    });
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
