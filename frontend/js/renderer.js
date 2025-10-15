// SVG rendering for nodes and connections

const NODE_WIDTH = 200;
const NODE_HEIGHT = 180;
const PORT_RADIUS = 6;
const PORT_SPACING = 30;
const THUMBNAIL_WIDTH = 80;
const THUMBNAIL_HEIGHT = 60;
const THUMBNAIL_Y = 48;

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
        this.zoom = 1;
        this.panX = 0;
        this.panY = 0;

        this.setupZoom();
    }

    setupZoom() {
        this.svg.addEventListener('wheel', (e) => {
            e.preventDefault();

            const delta = -e.deltaY;
            const zoomFactor = delta > 0 ? 1.1 : 0.9;

            // Get mouse position relative to SVG
            const rect = this.svg.getBoundingClientRect();
            const mouseX = e.clientX - rect.left;
            const mouseY = e.clientY - rect.top;

            // Calculate new zoom
            const newZoom = Math.max(0.1, Math.min(5, this.zoom * zoomFactor));

            // Adjust pan to zoom towards mouse position
            const scale = newZoom / this.zoom;
            this.panX = mouseX - (mouseX - this.panX) * scale;
            this.panY = mouseY - (mouseY - this.panY) * scale;

            this.zoom = newZoom;
            this.updateTransform();

            // Notify about viewport change (if handler is set)
            if (this.onViewportChange) {
                this.onViewportChange();
            }
        });
    }

    // Set a callback for viewport changes
    setViewportChangeCallback(callback) {
        this.onViewportChange = callback;
    }

    updateTransform() {
        const transform = `translate(${this.panX}, ${this.panY}) scale(${this.zoom})`;
        this.nodesLayer.setAttribute('transform', transform);
        this.connectionsLayer.setAttribute('transform', transform);
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

        // Render thumbnail if first output has an image
        if (node.outputs && node.outputs.length > 0 && node.outputs[0].image_id) {
            this.renderThumbnail(g, node.outputs[0].image_id);
        }

        // Render input ports (left side)
        const inputs = node.inputs || [];
        const portStartY = THUMBNAIL_Y + THUMBNAIL_HEIGHT + 10; // 10px padding below thumbnail
        inputs.forEach((input, i) => {
            const portY = portStartY + i * PORT_SPACING;
            this.renderInputPort(g, input.name, portY);
        });

        // Render output ports (right side)
        const outputs = node.outputs || [];
        outputs.forEach((output, i) => {
            const portY = portStartY + i * PORT_SPACING;
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

    renderThumbnail(parentG, imageId) {
        const image = document.createElementNS('http://www.w3.org/2000/svg', 'image');
        image.classList.add('node-thumbnail');
        image.setAttribute('x', (NODE_WIDTH - THUMBNAIL_WIDTH) / 2);
        image.setAttribute('y', THUMBNAIL_Y);
        image.setAttribute('width', THUMBNAIL_WIDTH);
        image.setAttribute('height', THUMBNAIL_HEIGHT);
        image.setAttribute('href', `/api/images/${imageId}`);
        image.setAttribute('preserveAspectRatio', 'xMidYMid meet');
        parentG.appendChild(image);
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

        // Create a group for the connection
        const g = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        g.classList.add('connection-group');
        g.setAttribute('data-from-node', sourceNodeId);
        g.setAttribute('data-from-output', sourceOutput);
        g.setAttribute('data-to-node', targetNodeId);
        g.setAttribute('data-to-input', targetInput);

        // Invisible wider path for better hover detection
        const hoverPath = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        hoverPath.classList.add('connection-hover');
        const dx = x2 - x1;
        const curve = Math.abs(dx) / 2;
        const d = `M ${x1} ${y1} C ${x1 + curve} ${y1}, ${x2 - curve} ${y2}, ${x2} ${y2}`;
        hoverPath.setAttribute('d', d);

        // Visible connection path
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.classList.add('connection');
        if (hasImage) {
            path.classList.add('has-image');
        }
        path.setAttribute('d', d);

        // Calculate midpoint for delete button
        const midX = (x1 + x2) / 2;
        const midY = (y1 + y2) / 2;

        // Delete button (hidden by default)
        const deleteBtn = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        deleteBtn.classList.add('connection-delete-btn');

        const btnCircle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
        btnCircle.setAttribute('cx', midX);
        btnCircle.setAttribute('cy', midY);
        btnCircle.setAttribute('r', 12);
        btnCircle.classList.add('connection-delete-bg');

        const btnText = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        btnText.setAttribute('x', midX);
        btnText.setAttribute('y', midY + 5);
        btnText.setAttribute('text-anchor', 'middle');
        btnText.classList.add('connection-delete-icon');
        btnText.textContent = '×';

        deleteBtn.appendChild(btnCircle);
        deleteBtn.appendChild(btnText);

        g.appendChild(hoverPath);
        g.appendChild(path);
        g.appendChild(deleteBtn);

        this.connectionsLayer.appendChild(g);
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

    // Export current viewport state
    exportViewport() {
        return {
            zoom: this.zoom,
            pan_x: this.panX,
            pan_y: this.panY
        };
    }

    // Export all node positions
    exportNodePositions() {
        // Return the Map directly (will be converted to array in API layer)
        return this.nodePositions;
    }

    // Restore viewport from metadata
    restoreViewport(viewport) {
        if (viewport) {
            this.zoom = viewport.zoom || 1.0;
            this.panX = viewport.pan_x || 0;
            this.panY = viewport.pan_y || 0;
            this.updateTransform();
        }
    }

    // Restore node positions from metadata
    restoreNodePositions(nodePositions) {
        if (nodePositions) {
            // nodePositions is now an array of {node_id, x, y}
            for (const pos of nodePositions) {
                this.nodePositions.set(pos.node_id, { x: pos.x, y: pos.y });
            }
        }
    }
}
