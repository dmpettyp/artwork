// SVG rendering for nodes and connections

const NODE_WIDTH = 200;
const NODE_HEIGHT = 120;
const PORT_RADIUS = 6;
const PORT_SPACING = 30;

export class Renderer {
    constructor(svgElement, nodesLayer, connectionsLayer) {
        this.svg = svgElement;
        this.nodesLayer = nodesLayer;
        this.connectionsLayer = connectionsLayer;
        this.nodePositions = new Map();
    }

    clear() {
        this.nodesLayer.innerHTML = '';
        this.connectionsLayer.innerHTML = '';
    }

    render(graph) {
        if (!graph) {
            this.clear();
            return;
        }

        this.clear();

        // Render connections first (so they appear behind nodes)
        graph.nodes.forEach(node => {
            Object.entries(node.outputs || {}).forEach(([outputName, output]) => {
                (output.connected_to || []).forEach(conn => {
                    this.renderConnection(node.id, outputName, conn.node_id, conn.input_name, output.image !== null);
                });
            });
        });

        // Render nodes
        graph.nodes.forEach((node, index) => {
            // Simple grid layout if no position stored
            if (!this.nodePositions.has(node.id)) {
                const col = index % 3;
                const row = Math.floor(index / 3);
                this.nodePositions.set(node.id, {
                    x: 100 + col * (NODE_WIDTH + 100),
                    y: 100 + row * (NODE_HEIGHT + 100)
                });
            }

            const pos = this.nodePositions.get(node.id);
            this.renderNode(node, pos.x, pos.y);
        });
    }

    renderNode(node, x, y) {
        const g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        g.classList.add('node');
        g.classList.add(`state-${node.state}`);
        g.setAttribute('data-node-id', node.id);
        g.setAttribute('transform', `translate(${x},${y})`);

        // Node rectangle
        const rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
        rect.classList.add('node-rect');
        rect.setAttribute('width', NODE_WIDTH);
        rect.setAttribute('height', NODE_HEIGHT);
        g.appendChild(rect);

        // Node title
        const title = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        title.classList.add('node-title');
        title.setAttribute('x', NODE_WIDTH / 2);
        title.setAttribute('y', 25);
        title.setAttribute('text-anchor', 'middle');
        title.textContent = node.type;
        g.appendChild(title);

        // Node type/state
        const type = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        type.classList.add('node-type');
        type.setAttribute('x', NODE_WIDTH / 2);
        type.setAttribute('y', 42);
        type.setAttribute('text-anchor', 'middle');
        type.textContent = `${node.state} • ${node.id.substring(0, 8)}`;
        g.appendChild(type);

        // Render input ports (left side)
        const inputs = Object.keys(node.inputs || {});
        inputs.forEach((inputName, i) => {
            const portY = 60 + i * PORT_SPACING;
            this.renderInputPort(g, inputName, portY);
        });

        // Render output ports (right side)
        const outputs = Object.keys(node.outputs || {});
        outputs.forEach((outputName, i) => {
            const portY = 60 + i * PORT_SPACING;
            const output = node.outputs[outputName];
            this.renderOutputPort(g, outputName, portY, output.image !== null);
        });

        this.nodesLayer.appendChild(g);
    }

    renderInputPort(parentG, inputName, y) {
        const g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        g.classList.add('input-port');
        g.setAttribute('data-input-name', inputName);

        const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
        circle.classList.add('port');
        circle.setAttribute('cx', 0);
        circle.setAttribute('cy', y);
        circle.setAttribute('r', PORT_RADIUS);
        g.appendChild(circle);

        const label = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        label.classList.add('port-label');
        label.setAttribute('x', 12);
        label.setAttribute('y', y + 4);
        label.textContent = inputName;
        g.appendChild(label);

        parentG.appendChild(g);
    }

    renderOutputPort(parentG, outputName, y, hasImage) {
        const g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        g.classList.add('output-port');
        g.setAttribute('data-output-name', outputName);

        const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
        circle.classList.add('port');
        circle.setAttribute('cx', NODE_WIDTH);
        circle.setAttribute('cy', y);
        circle.setAttribute('r', PORT_RADIUS);
        if (hasImage) {
            circle.style.fill = '#27ae60';
        }
        g.appendChild(circle);

        const label = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        label.classList.add('port-label');
        label.setAttribute('x', NODE_WIDTH - 12);
        label.setAttribute('y', y + 4);
        label.setAttribute('text-anchor', 'end');
        label.textContent = outputName;
        g.appendChild(label);

        parentG.appendChild(g);
    }

    renderConnection(sourceNodeId, sourceOutput, targetNodeId, targetInput, hasImage) {
        const sourcePosNode = this.nodePositions.get(sourceNodeId);
        const targetPosNode = this.nodePositions.get(targetNodeId);

        if (!sourcePosNode || !targetPosNode) return;

        // Find port positions
        const sourceNode = this.nodesLayer.querySelector(`[data-node-id="${sourceNodeId}"]`);
        const targetNode = this.nodesLayer.querySelector(`[data-node-id="${targetNodeId}"]`);

        if (!sourceNode || !targetNode) return;

        const sourcePort = sourceNode.querySelector(`[data-output-name="${sourceOutput}"] circle`);
        const targetPort = targetNode.querySelector(`[data-input-name="${targetInput}"] circle`);

        if (!sourcePort || !targetPort) return;

        const x1 = sourcePosNode.x + parseFloat(sourcePort.getAttribute('cx'));
        const y1 = sourcePosNode.y + parseFloat(sourcePort.getAttribute('cy'));
        const x2 = targetPosNode.x + parseFloat(targetPort.getAttribute('cx'));
        const y2 = targetPosNode.y + parseFloat(targetPort.getAttribute('cy'));

        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.classList.add('connection');
        if (hasImage) {
            path.classList.add('has-image');
        }

        // Bezier curve for nicer connections
        const dx = x2 - x1;
        const curve = Math.abs(dx) / 2;
        const d = `M ${x1} ${y1} C ${x1 + curve} ${y1}, ${x2 - curve} ${y2}, ${x2} ${y2}`;
        path.setAttribute('d', d);

        this.connectionsLayer.appendChild(path);
    }

    renderTempConnection(x1, y1, x2, y2) {
        // Remove existing temp connection
        const existing = this.connectionsLayer.querySelector('.connection-temp');
        if (existing) {
            existing.remove();
        }

        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.classList.add('connection-temp');

        const dx = x2 - x1;
        const curve = Math.abs(dx) / 2;
        const d = `M ${x1} ${y1} C ${x1 + curve} ${y1}, ${x2 - curve} ${y2}, ${x2} ${y2}`;
        path.setAttribute('d', d);

        this.connectionsLayer.appendChild(path);
    }

    removeTempConnection() {
        const temp = this.connectionsLayer.querySelector('.connection-temp');
        if (temp) {
            temp.remove();
        }
    }

    updateNodePosition(nodeId, x, y) {
        this.nodePositions.set(nodeId, { x, y });
        const nodeElement = this.nodesLayer.querySelector(`[data-node-id="${nodeId}"]`);
        if (nodeElement) {
            nodeElement.setAttribute('transform', `translate(${x},${y})`);
        }
    }

    getNodePosition(nodeId) {
        return this.nodePositions.get(nodeId);
    }
}
