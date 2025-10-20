// User interaction handlers for drag and drop

export class InteractionHandler {
    constructor(svgElement, renderer, graphState, api, renderOutputsCallback = null, openEditConfigCallback = null) {
        this.svg = svgElement;
        this.renderer = renderer;
        this.graphState = graphState;
        this.api = api;
        this.renderOutputsCallback = renderOutputsCallback;
        this.openEditConfigCallback = openEditConfigCallback;

        this.draggedNode = null;
        this.dragOffset = { x: 0, y: 0 };
        this.connectionDrag = null;
        this.canvasDrag = null;

        // Debounce timer for persisting UI state
        this.saveViewportTimeout = null;

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.svg.addEventListener('mousedown', this.handleMouseDown.bind(this));
        this.svg.addEventListener('mousemove', this.handleMouseMove.bind(this));
        this.svg.addEventListener('mouseup', this.handleMouseUp.bind(this));
        this.svg.addEventListener('dblclick', this.handleDoubleClick.bind(this));
    }

    handleMouseDown(e) {
        // Ignore right-click
        if (e.button === 2) {
            return;
        }

        // Check if clicking on a port (for connections)
        const port = e.target.closest('.port');
        if (port) {
            this.startConnectionDrag(port, e);
            return;
        }

        // Check if clicking on a node (for dragging)
        const node = e.target.closest('.node');
        if (node) {
            this.startNodeDrag(node, e);
            return;
        }

        // Check if clicking on a connection
        const connection = e.target.closest('.connection-group');
        if (connection) {
            return; // Don't start canvas drag on connections
        }

        // If clicking on background, start canvas pan
        if (e.target === this.svg || e.target.id === 'connections-layer' || e.target.id === 'nodes-layer') {
            this.startCanvasDrag(e);
            return;
        }
    }

    handleMouseMove(e) {
        if (this.draggedNode) {
            this.updateNodeDrag(e);
        } else if (this.connectionDrag) {
            this.updateConnectionDrag(e);
        } else if (this.canvasDrag) {
            this.updateCanvasDrag(e);
        }
    }

    handleMouseUp(e) {
        if (this.draggedNode) {
            this.endNodeDrag();
        } else if (this.connectionDrag) {
            this.endConnectionDrag(e);
        } else if (this.canvasDrag) {
            this.endCanvasDrag();
        }
    }

    handleDoubleClick(e) {
        // Check if double-clicking on a node
        const nodeElement = e.target.closest('.node');
        if (nodeElement && this.openEditConfigCallback) {
            const nodeId = nodeElement.getAttribute('data-node-id');
            if (nodeId) {
                this.openEditConfigCallback(nodeId);
            }
        }
    }

    screenToCanvas(screenX, screenY) {
        // Convert screen coordinates to canvas coordinates accounting for zoom and pan
        const canvasX = (screenX - this.renderer.panX) / this.renderer.zoom;
        const canvasY = (screenY - this.renderer.panY) / this.renderer.zoom;
        return { x: canvasX, y: canvasY };
    }

    startNodeDrag(nodeElement, e) {
        const nodeId = nodeElement.getAttribute('data-node-id');
        const pos = this.renderer.getNodePosition(nodeId);

        const svgRect = this.svg.getBoundingClientRect();
        const screenX = e.clientX - svgRect.left;
        const screenY = e.clientY - svgRect.top;

        const canvasPos = this.screenToCanvas(screenX, screenY);

        this.draggedNode = nodeId;
        this.dragOffset = {
            x: canvasPos.x - pos.x,
            y: canvasPos.y - pos.y
        };

        nodeElement.style.cursor = 'grabbing';
    }

    updateNodeDrag(e) {
        if (!this.draggedNode) return;

        const svgRect = this.svg.getBoundingClientRect();
        const screenX = e.clientX - svgRect.left;
        const screenY = e.clientY - svgRect.top;

        const canvasPos = this.screenToCanvas(screenX, screenY);

        const newX = canvasPos.x - this.dragOffset.x;
        const newY = canvasPos.y - this.dragOffset.y;

        this.renderer.updateNodePosition(this.draggedNode, newX, newY);

        // Re-render to update connections
        const graph = this.graphState.getCurrentGraph();
        if (graph) {
            this.renderer.render(graph);

            // Re-render output panel to update sorting
            if (this.renderOutputsCallback) {
                this.renderOutputsCallback(graph);
            }
        }
    }

    endNodeDrag() {
        if (!this.draggedNode) return;

        const nodeElement = this.svg.querySelector(`[data-node-id="${this.draggedNode}"]`);
        if (nodeElement) {
            nodeElement.style.cursor = 'move';
        }

        // Persist all UI state (viewport + all node positions) to backend
        this.debouncedSaveViewport();

        this.draggedNode = null;
        this.dragOffset = { x: 0, y: 0 };
    }

    startCanvasDrag(e) {
        this.canvasDrag = {
            startX: e.clientX,
            startY: e.clientY,
            startPanX: this.renderer.panX,
            startPanY: this.renderer.panY
        };
        this.svg.style.cursor = 'grabbing';
    }

    updateCanvasDrag(e) {
        if (!this.canvasDrag) return;

        const dx = e.clientX - this.canvasDrag.startX;
        const dy = e.clientY - this.canvasDrag.startY;

        this.renderer.panX = this.canvasDrag.startPanX + dx;
        this.renderer.panY = this.canvasDrag.startPanY + dy;

        this.renderer.updateTransform();
    }

    endCanvasDrag() {
        if (!this.canvasDrag) return;

        // Persist all UI state to backend (debounced)
        this.debouncedSaveViewport();

        this.canvasDrag = null;
        this.svg.style.cursor = 'grab';
    }

    debouncedSaveViewport() {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId) return;

        // Clear any existing timeout to debounce
        clearTimeout(this.saveViewportTimeout);
        this.saveViewportTimeout = setTimeout(async () => {
            try {
                const viewport = this.renderer.exportViewport();
                const nodePositions = this.renderer.exportNodePositions();
                await this.api.updateUIMetadata(graphId, viewport, nodePositions);
            } catch (error) {
                console.error('Failed to save UI metadata:', error);
            }
        }, 500); // 500ms debounce
    }

    startConnectionDrag(portElement, e) {
        const nodeElement = portElement.closest('.node');
        const nodeId = nodeElement.getAttribute('data-node-id');
        const nodePos = this.renderer.getNodePosition(nodeId);

        const portCell = portElement.closest('.port-cell');
        const isOutput = portCell.classList.contains('port-cell-output');

        const portName = isOutput
            ? portCell.getAttribute('data-output-name')
            : portCell.getAttribute('data-input-name');

        const circle = portElement.closest('circle');
        const portX = nodePos.x + parseFloat(circle.getAttribute('cx'));
        const portY = nodePos.y + parseFloat(circle.getAttribute('cy'));

        this.connectionDrag = {
            nodeId,
            portName,
            isOutput,
            startX: portX,
            startY: portY
        };

        e.preventDefault();
        e.stopPropagation();
    }

    updateConnectionDrag(e) {
        if (!this.connectionDrag) return;

        const svgRect = this.svg.getBoundingClientRect();
        const screenX = e.clientX - svgRect.left;
        const screenY = e.clientY - svgRect.top;

        const canvasPos = this.screenToCanvas(screenX, screenY);

        this.renderer.renderTempConnection(
            this.connectionDrag.startX,
            this.connectionDrag.startY,
            canvasPos.x,
            canvasPos.y
        );
    }

    async endConnectionDrag(e) {
        if (!this.connectionDrag) return;

        const port = e.target.closest('.port');
        if (port) {
            const targetNodeElement = port.closest('.node');
            const targetNodeId = targetNodeElement.getAttribute('data-node-id');

            const portCell = port.closest('.port-cell');
            const isTargetOutput = portCell.classList.contains('port-cell-output');

            const targetPortName = isTargetOutput
                ? portCell.getAttribute('data-output-name')
                : portCell.getAttribute('data-input-name');

            // Valid connection: output -> input
            if (this.connectionDrag.isOutput && !isTargetOutput) {
                await this.createConnection(
                    this.connectionDrag.nodeId,
                    this.connectionDrag.portName,
                    targetNodeId,
                    targetPortName
                );
            }
            // Valid connection: input <- output
            else if (!this.connectionDrag.isOutput && isTargetOutput) {
                await this.createConnection(
                    targetNodeId,
                    targetPortName,
                    this.connectionDrag.nodeId,
                    this.connectionDrag.portName
                );
            }
        }

        this.renderer.removeTempConnection();
        this.connectionDrag = null;
    }

    async createConnection(sourceNodeId, sourceOutput, targetNodeId, targetInput) {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId) return;

        try {
            await this.api.connectNodes(graphId, sourceNodeId, sourceOutput, targetNodeId, targetInput);

            // Refresh the graph to show the new connection
            const graph = await this.api.getImageGraph(graphId);
            this.graphState.setCurrentGraph(graph);
        } catch (error) {
            console.error('Failed to create connection:', error);
            alert(`Failed to create connection: ${error.message}`);
        }
    }

    // Cancel all ongoing drag operations
    cancelAllDrags() {
        if (this.draggedNode) {
            const nodeElement = this.svg.querySelector(`[data-node-id="${this.draggedNode}"]`);
            if (nodeElement) {
                nodeElement.style.cursor = 'move';
            }
            this.draggedNode = null;
            this.dragOffset = { x: 0, y: 0 };
        }

        if (this.connectionDrag) {
            this.renderer.removeTempConnection();
            this.connectionDrag = null;
        }

        if (this.canvasDrag) {
            this.canvasDrag = null;
            this.svg.style.cursor = 'grab';
        }
    }
}
