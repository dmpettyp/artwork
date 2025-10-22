// GraphManager - Handles graph loading, selection, and WebSocket connections

import { API_PATHS, WS_CONFIG } from './constants.js';

export class GraphManager {
    constructor(api, graphState, renderer, toastManager) {
        this.api = api;
        this.graphState = graphState;
        this.renderer = renderer;
        this.toastManager = toastManager;

        this.wsConnection = null;
        this.wsReconnectTimeout = null;

        // Callbacks
        this.onGraphListRendered = null;
    }

    // Set callback for when graph list is rendered
    setGraphListRenderedCallback(callback) {
        this.onGraphListRendered = callback;
    }

    // Load and display graph list
    async loadGraphList() {
        try {
            const graphs = await this.api.listImageGraphs();
            this.renderGraphList(graphs);

            // Auto-select the first graph if none is selected
            if (graphs.length > 0 && !this.graphState.getCurrentGraphId()) {
                await this.selectGraph(graphs[0].id);
            }
        } catch (error) {
            console.error('Failed to load graphs:', error);
        }
    }

    // Render graph list in dropdown
    renderGraphList(graphs) {
        const graphSelect = document.getElementById('graph-select');
        const currentGraphId = this.graphState.getCurrentGraphId();

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

        // Notify callback if set
        if (this.onGraphListRendered) {
            this.onGraphListRendered(graphs);
        }
    }

    // Select and load a graph
    async selectGraph(graphId) {
        try {
            const graph = await this.api.getImageGraph(graphId);

            // Load UI metadata and restore viewport/positions
            try {
                const uiMetadata = await this.api.getUIMetadata(graphId);
                this.renderer.restoreViewport(uiMetadata.viewport);
                this.renderer.restoreNodePositions(uiMetadata.node_positions);
            } catch (error) {
                console.log('No UI metadata found, using defaults:', error);
                // Not an error - just means no metadata was saved yet
            }

            this.graphState.setCurrentGraph(graph);
            await this.loadGraphList(); // Refresh list to update active state

            // Connect to WebSocket for real-time updates
            this.connectWebSocket(graphId);
        } catch (error) {
            console.error('Failed to load graph:', error);
            this.toastManager.error(`Failed to load graph: ${error.message}`);
        }
    }

    // Reload the currently selected graph
    async reloadCurrentGraph() {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId) return;

        const graph = await this.api.getImageGraph(graphId);
        this.graphState.setCurrentGraph(graph);
    }

    // WebSocket connection management
    connectWebSocket(graphId) {
        // Disconnect existing connection if any
        this.disconnectWebSocket();

        // Determine WebSocket URL (ws:// for http://, wss:// for https://)
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}${API_PATHS.graphWebSocket(graphId)}`;

        console.log('Connecting to WebSocket:', wsUrl);

        try {
            this.wsConnection = new WebSocket(wsUrl);

            this.wsConnection.onopen = () => {
                console.log('WebSocket connected for graph:', graphId);
                // Clear any pending reconnect attempts
                if (this.wsReconnectTimeout) {
                    clearTimeout(this.wsReconnectTimeout);
                    this.wsReconnectTimeout = null;
                }
            };

            this.wsConnection.onmessage = async (event) => {
                try {
                    const message = JSON.parse(event.data);
                    console.log('WebSocket message received:', message);

                    // Refresh the graph to get the latest state
                    await this.reloadCurrentGraph();
                } catch (error) {
                    console.error('Failed to handle WebSocket message:', error);
                }
            };

            this.wsConnection.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            this.wsConnection.onclose = (event) => {
                console.log('WebSocket closed:', event.code, event.reason);
                this.wsConnection = null;

                // Attempt to reconnect if the graph is still selected
                const currentGraphId = this.graphState.getCurrentGraphId();
                if (currentGraphId === graphId) {
                    console.log(`Will attempt to reconnect in ${WS_CONFIG.reconnectDelay}ms`);
                    this.wsReconnectTimeout = setTimeout(() => {
                        this.connectWebSocket(graphId);
                    }, WS_CONFIG.reconnectDelay);
                }
            };
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
        }
    }

    disconnectWebSocket() {
        // Clear any pending reconnect attempts
        if (this.wsReconnectTimeout) {
            clearTimeout(this.wsReconnectTimeout);
            this.wsReconnectTimeout = null;
        }

        // Close existing connection
        if (this.wsConnection) {
            console.log('Disconnecting WebSocket');
            this.wsConnection.close(1000, 'Client disconnecting');
            this.wsConnection = null;
        }
    }

    // Cleanup on page unload
    cleanup() {
        this.disconnectWebSocket();
    }
}
