// User interaction handlers for drag and drop

export class InteractionHandler {
    constructor(svgElement, renderer, graphState, api) {
        this.svg = svgElement;
        this.renderer = renderer;
        this.graphState = graphState;
        this.api = api;

        this.draggedNode = null;
        this.dragOffset = { x: 0, y: 0 };
        this.connectionDrag = null;
        this.canvasDrag = null;

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.svg.addEventListener('mousedown', this.handleMouseDown.bind(this));
        this.svg.addEventListener('mousemove', this.handleMouseMove.bind(this));
        this.svg.addEventListener('mouseup', this.handleMouseUp.bind(this));
    }

    handleMouseDown(e) {
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
        }
    }

    endNodeDrag() {
        if (!this.draggedNode) return;

        const nodeElement = this.svg.querySelector(`[data-node-id="${this.draggedNode}"]`);
        if (nodeElement) {
            nodeElement.style.cursor = 'move';
        }

        const finalPosition = this.renderer.getNodePosition(this.draggedNode);
        console.log('Node drag ended:', {
            nodeId: this.draggedNode,
            position: finalPosition
        });

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

        this.canvasDrag = null;
        this.svg.style.cursor = 'grab';
    }

    startConnectionDrag(portElement, e) {
        const nodeElement = portElement.closest('.node');
        const nodeId = nodeElement.getAttribute('data-node-id');
        const nodePos = this.renderer.getNodePosition(nodeId);

        const portParent = portElement.parentElement;
        const isOutput = portParent.classList.contains('output-port');

        const portName = isOutput
            ? portParent.getAttribute('data-output-name')
            : portParent.getAttribute('data-input-name');

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

            const portParent = port.parentElement;
            const isTargetOutput = portParent.classList.contains('output-port');

            const targetPortName = isTargetOutput
                ? portParent.getAttribute('data-output-name')
                : portParent.getAttribute('data-input-name');

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
}
