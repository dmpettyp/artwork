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
    }

    handleMouseMove(e) {
        if (this.draggedNode) {
            this.updateNodeDrag(e);
        } else if (this.connectionDrag) {
            this.updateConnectionDrag(e);
        }
    }

    handleMouseUp(e) {
        if (this.draggedNode) {
            this.endNodeDrag();
        } else if (this.connectionDrag) {
            this.endConnectionDrag(e);
        }
    }

    startNodeDrag(nodeElement, e) {
        const nodeId = nodeElement.getAttribute('data-node-id');
        const pos = this.renderer.getNodePosition(nodeId);

        const svgRect = this.svg.getBoundingClientRect();
        const mouseX = e.clientX - svgRect.left;
        const mouseY = e.clientY - svgRect.top;

        this.draggedNode = nodeId;
        this.dragOffset = {
            x: mouseX - pos.x,
            y: mouseY - pos.y
        };

        nodeElement.style.cursor = 'grabbing';
    }

    updateNodeDrag(e) {
        if (!this.draggedNode) return;

        const svgRect = this.svg.getBoundingClientRect();
        const mouseX = e.clientX - svgRect.left;
        const mouseY = e.clientY - svgRect.top;

        const newX = mouseX - this.dragOffset.x;
        const newY = mouseY - this.dragOffset.y;

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

        this.draggedNode = null;
        this.dragOffset = { x: 0, y: 0 };
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
        const mouseX = e.clientX - svgRect.left;
        const mouseY = e.clientY - svgRect.top;

        this.renderer.renderTempConnection(
            this.connectionDrag.startX,
            this.connectionDrag.startY,
            mouseX,
            mouseY
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
