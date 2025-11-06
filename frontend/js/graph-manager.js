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

            // Load layout and viewport separately
            try {
                const layout = await this.api.getLayout(graphId);
                this.renderer.restoreNodePositions(layout.node_positions);
            } catch (error) {
                // Not an error - just means no layout was saved yet
            }

            try {
                const viewport = await this.api.getViewport(graphId);
                this.renderer.restoreViewport(viewport);
            } catch (error) {
                // Not an error - just means no viewport was saved yet
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

    // Reload just the layout for a graph
    async reloadLayout(graphId) {
        try {
            const layout = await this.api.getLayout(graphId);
            this.renderer.restoreNodePositions(layout.node_positions);

            // Re-render the current graph to apply the new positions
            const currentGraph = this.graphState.getCurrentGraph();
            if (currentGraph) {
                this.graphState.setCurrentGraph(currentGraph);
            }
        } catch (error) {
            console.error('Failed to reload layout:', error);
        }
    }

    // WebSocket connection management
    connectWebSocket(graphId) {
        // Disconnect existing connection if any
        this.disconnectWebSocket();

        // Determine WebSocket URL (ws:// for http://, wss:// for https://)
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}${API_PATHS.graphWebSocket(graphId)}`;

        try {
            this.wsConnection = new WebSocket(wsUrl);

            this.wsConnection.onopen = () => {
                // Clear any pending reconnect attempts
                if (this.wsReconnectTimeout) {
                    clearTimeout(this.wsReconnectTimeout);
                    this.wsReconnectTimeout = null;
                }
            };

            this.wsConnection.onmessage = async (event) => {
                try {
                    const message = JSON.parse(event.data);

                    // Handle different message types
                    if (message.type === 'layout_update') {
                        // Layout changed - fetch and apply new layout
                        await this.reloadLayout(graphId);
                    } else if (message.type === 'node_update') {
                        // Node state changed - refresh the entire graph
                        await this.reloadCurrentGraph();
                    } else {
                        // Unknown message type - refresh everything
                        await this.reloadCurrentGraph();
                    }
                } catch (error) {
                    console.error('Failed to handle WebSocket message:', error);
                }
            };

            this.wsConnection.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            this.wsConnection.onclose = (event) => {
                this.wsConnection = null;

                // Attempt to reconnect if the graph is still selected
                const currentGraphId = this.graphState.getCurrentGraphId();
                if (currentGraphId === graphId) {
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
            this.wsConnection.close(1000, 'Client disconnecting');
            this.wsConnection = null;
        }
    }

    // Cleanup on page unload
    cleanup() {
        this.disconnectWebSocket();
    }
}
