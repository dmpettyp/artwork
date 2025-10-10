// SVG rendering for nodes and connections

const NODE_WIDTH = 200;
const NODE_HEIGHT = 120;
const PORT_RADIUS = 6;
const PORT_SPACING = 30;

// Node type configurations (matches backend node_type.go)
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

export class Renderer {
    constructor(svgElement, nodesLayer, connectionsLayer) {
        this.svg = svgElement;
        this.nodesLayer = nodesLayer;
        this.connectionsLayer = connectionsLayer;
        this.nodePositions = new Map();
    }

    nodeTypeHasConfig(nodeType) {
        const config = nodeTypeConfigs[nodeType];
        return config && config.fields && Object.keys(config.fields).length > 0;
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

        // Render nodes first
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

        // Render connections after nodes (but they'll appear behind due to z-order in SVG)
        graph.nodes.forEach(node => {
            (node.outputs || []).forEach(output => {
                (output.connections || []).forEach(conn => {
                    this.renderConnection(node.id, output.name, conn.node_id, conn.input_name, output.image_id !== null && output.image_id !== '');
                });
            });
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

        // Node title (name)
        const title = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        title.classList.add('node-title');
        title.setAttribute('x', NODE_WIDTH / 2);
        title.setAttribute('y', 25);
        title.setAttribute('text-anchor', 'middle');
        title.textContent = node.name;
        g.appendChild(title);

        // Node type and state
        const type = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        type.classList.add('node-type');
        type.setAttribute('x', NODE_WIDTH / 2);
        type.setAttribute('y', 42);
        type.setAttribute('text-anchor', 'middle');
        type.textContent = `${node.type} • ${node.state}`;
        g.appendChild(type);

        // Render input ports (left side)
        const inputs = node.inputs || [];
        inputs.forEach((input, i) => {
            const portY = 60 + i * PORT_SPACING;
            this.renderInputPort(g, input.name, portY);
        });

        // Render output ports (right side)
        const outputs = node.outputs || [];
        outputs.forEach((output, i) => {
            const portY = 60 + i * PORT_SPACING;
            this.renderOutputPort(g, output.name, portY, output.image_id !== null && output.image_id !== '');
        });

        // Action buttons (hidden by default, shown on hover)
        const actionsGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        actionsGroup.classList.add('node-actions');

        // Check if node type has configurable fields
        const hasConfig = this.nodeTypeHasConfig(node.type);

        // Edit config button (only if node has configurable fields)
        if (hasConfig) {
            const editBtn = this.createActionButton(NODE_WIDTH - 50, 5, '⚙', 'edit-config');
            actionsGroup.appendChild(editBtn);
        }

        // Delete button (position depends on whether edit button is present)
        const deleteX = hasConfig ? NODE_WIDTH - 25 : NODE_WIDTH - 25;
        const deleteBtn = this.createActionButton(deleteX, 5, '×', 'delete');
        actionsGroup.appendChild(deleteBtn);

        g.appendChild(actionsGroup);

        this.nodesLayer.appendChild(g);
    }

    createActionButton(x, y, icon, action) {
        const btnGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        btnGroup.classList.add('node-action-btn');
        btnGroup.setAttribute('data-action', action);
        btnGroup.style.cursor = 'pointer';

        const btnRect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
        btnRect.setAttribute('x', x);
        btnRect.setAttribute('y', y);
        btnRect.setAttribute('width', 20);
        btnRect.setAttribute('height', 20);
        btnRect.setAttribute('rx', 3);
        btnRect.classList.add('action-btn-bg');

        const btnText = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        btnText.setAttribute('x', x + 10);
        btnText.setAttribute('y', y + 15);
        btnText.setAttribute('text-anchor', 'middle');
        btnText.classList.add('action-btn-icon');
        btnText.textContent = icon;

        btnGroup.appendChild(btnRect);
        btnGroup.appendChild(btnText);

        return btnGroup;
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
        console.log(outputName);
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
