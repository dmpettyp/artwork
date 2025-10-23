// SVG rendering for nodes and connections

import { NODE_DESIGN, API_PATHS, ZOOM_CONFIG, LAYOUT_CONFIG, CONNECTION_DELETE_BUTTON } from './constants.js';

export class Renderer {
    constructor(svgElement, nodesLayer, connectionsLayer, graphState = null) {
        this.svg = svgElement;
        this.nodesLayer = nodesLayer;
        this.connectionsLayer = connectionsLayer;
        this.graphState = graphState;
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
            const zoomFactor = delta > 0 ? ZOOM_CONFIG.factor.in : ZOOM_CONFIG.factor.out;

            // Get mouse position relative to SVG
            const rect = this.svg.getBoundingClientRect();
            const mouseX = e.clientX - rect.left;
            const mouseY = e.clientY - rect.top;

            // Calculate new zoom
            const newZoom = Math.max(ZOOM_CONFIG.limits.min, Math.min(ZOOM_CONFIG.limits.max, this.zoom * zoomFactor));

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
                const col = index % LAYOUT_CONFIG.gridColumns;
                const row = Math.floor(index / LAYOUT_CONFIG.gridColumns);
                this.nodePositions.set(node.id, {
                    x: LAYOUT_CONFIG.initialOffset + col * (NODE_DESIGN.width + LAYOUT_CONFIG.gridSpacing),
                    y: LAYOUT_CONFIG.initialOffset + row * (NODE_DESIGN.height + LAYOUT_CONFIG.gridSpacing)
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

        const inputs = node.inputs || [];
        const outputs = node.outputs || [];

        // Store default thumbnail image ID (first output's image)
        const defaultImageId = (node.outputs && node.outputs.length > 0 && node.outputs[0].image_id)
            ? node.outputs[0].image_id
            : null;
        if (defaultImageId) {
            g.setAttribute('data-default-image-id', defaultImageId);
        }

        // Layout constants
        const thumbnailY = NODE_DESIGN.titleBarHeight + 10;
        const portTableY = thumbnailY + NODE_DESIGN.thumbnail.height + 10;
        const maxRows = Math.max(inputs.length, outputs.length, 1);
        const portTableHeight = NODE_DESIGN.headerHeight + maxRows * NODE_DESIGN.rowHeight;
        const nodeHeight = portTableY + portTableHeight + NODE_DESIGN.tablePadding;

        // Node rectangle (main body)
        const rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
        rect.classList.add('node-rect');
        rect.setAttribute('width', NODE_DESIGN.width);
        rect.setAttribute('height', nodeHeight);
        g.appendChild(rect);

        // Title bar background - rounded top, square bottom
        const titleBar = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        titleBar.classList.add('node-title-bar');
        const radius = 12;
        const width = NODE_DESIGN.width;
        const height = NODE_DESIGN.titleBarHeight;
        // Path: start at bottom-left, go to top-left with curve, across top with curve, down to bottom-right, across bottom
        const d = `
            M 0,${height}
            L 0,${radius}
            Q 0,0 ${radius},0
            L ${width - radius},0
            Q ${width},0 ${width},${radius}
            L ${width},${height}
            Z
        `;
        titleBar.setAttribute('d', d);
        g.appendChild(titleBar);

        // Node title (type and name) - add placeholder first, then truncate after rendering
        const title = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        title.classList.add('node-title');
        title.setAttribute('x', NODE_DESIGN.width / 2);
        title.setAttribute('y', NODE_DESIGN.titleBarHeight / 2 + 5);
        title.setAttribute('text-anchor', 'middle');

        const fullTitle = `${node.type}: ${node.name}`;
        title.textContent = fullTitle;
        g.appendChild(title);

        // Need to wait for next frame for text to be rendered and measurable
        requestAnimationFrame(() => {
            const maxWidth = NODE_DESIGN.width - 20; // 10px padding on each side
            const textLength = title.getComputedTextLength();

            if (textLength > maxWidth) {
                // Truncate the name but always keep the type
                const typePrefix = `${node.type}: `;

                // Start by trying to fit as much of the name as possible
                for (let i = node.name.length; i >= 0; i--) {
                    const truncatedName = node.name.substring(0, i);
                    title.textContent = typePrefix + truncatedName + '...';

                    if (title.getComputedTextLength() <= maxWidth) {
                        break;
                    }
                }
            }
        });

        // Render thumbnail if first output has an image
        if (defaultImageId) {
            this.renderThumbnail(g, defaultImageId, thumbnailY);
        } else if (node.state === 'waiting') {
            // Show "Waiting For Inputs..." message when in waiting state
            this.renderWaitingMessage(g, thumbnailY);
        } else if (node.state === 'generating') {
            // Show "Generating Outputs..." message when in generating state
            this.renderGeneratingMessage(g, thumbnailY);
        }

        // Render port table
        this.renderPortTable(g, inputs, outputs, portTableY, NODE_DESIGN.tablePadding);

        this.nodesLayer.appendChild(g);
    }

    renderThumbnail(parentG, imageId, yPos = NODE_DESIGN.thumbnail.y) {
        const image = document.createElementNS('http://www.w3.org/2000/svg', 'image');
        image.classList.add('node-thumbnail');
        image.setAttribute('x', (NODE_DESIGN.width - NODE_DESIGN.thumbnail.width) / 2);
        image.setAttribute('y', yPos);
        image.setAttribute('width', NODE_DESIGN.thumbnail.width);
        image.setAttribute('height', NODE_DESIGN.thumbnail.height);
        image.setAttribute('href', API_PATHS.images(imageId));
        image.setAttribute('preserveAspectRatio', 'xMidYMid meet');
        image.style.imageRendering = 'pixelated';
        parentG.appendChild(image);
    }

    renderWaitingMessage(parentG, yPos = NODE_DESIGN.thumbnail.y) {
        // Create a centered text message in the thumbnail area
        const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        text.classList.add('node-waiting-message');
        text.setAttribute('x', NODE_DESIGN.width / 2);
        text.setAttribute('y', yPos + NODE_DESIGN.thumbnail.height / 2);
        text.setAttribute('text-anchor', 'middle');
        text.setAttribute('dominant-baseline', 'middle');
        text.textContent = 'Waiting For Inputs...';
        parentG.appendChild(text);
    }

    renderGeneratingMessage(parentG, yPos = NODE_DESIGN.thumbnail.y) {
        // Create a centered text message in the thumbnail area
        const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        text.classList.add('node-generating-message');
        text.setAttribute('x', NODE_DESIGN.width / 2);
        text.setAttribute('y', yPos + NODE_DESIGN.thumbnail.height / 2);
        text.setAttribute('text-anchor', 'middle');
        text.setAttribute('dominant-baseline', 'middle');
        text.textContent = 'Generating Outputs...';
        parentG.appendChild(text);
    }

    updateThumbnail(nodeGroup, imageId) {
        // Find existing thumbnail
        const existingThumbnail = nodeGroup.querySelector('.node-thumbnail');

        if (existingThumbnail) {
            // Update the href to show the new image
            existingThumbnail.setAttribute('href', API_PATHS.images(imageId));
        } else {
            // Create thumbnail if it doesn't exist
            this.renderThumbnail(nodeGroup, imageId);
        }
    }

    restoreDefaultThumbnail(nodeGroup) {
        const defaultImageId = nodeGroup.getAttribute('data-default-image-id');
        const existingThumbnail = nodeGroup.querySelector('.node-thumbnail');

        if (defaultImageId && existingThumbnail) {
            // Restore to default image
            existingThumbnail.setAttribute('href', API_PATHS.images(defaultImageId));
        } else if (!defaultImageId && existingThumbnail) {
            // Remove thumbnail if there's no default
            existingThumbnail.remove();
        }
    }

    renderPortTable(parentG, inputs, outputs, startY, padding) {
        const maxRows = Math.max(inputs.length, outputs.length);
        const tableWidth = NODE_DESIGN.width - padding * 2; // Account for left and right padding
        const halfWidth = tableWidth / 2;

        // Render header row
        this.renderTableHeader(parentG, startY, padding, halfWidth, NODE_DESIGN.headerHeight, inputs.length > 0, outputs.length > 0);

        // Render port rows
        for (let i = 0; i < maxRows; i++) {
            const rowY = startY + NODE_DESIGN.headerHeight + i * NODE_DESIGN.rowHeight;

            // Left cell (input)
            if (i < inputs.length) {
                const imageId = inputs[i].image_id;
                this.renderPortCell(parentG, inputs[i].name, padding, rowY, halfWidth, NODE_DESIGN.rowHeight, 'input', imageId);
            }

            // Right cell (output)
            if (i < outputs.length) {
                const imageId = outputs[i].image_id;
                this.renderPortCell(parentG, outputs[i].name, padding + halfWidth, rowY, halfWidth, NODE_DESIGN.rowHeight, 'output', imageId);
            }

            // Divider line between left and right
            if (inputs.length > 0 && outputs.length > 0) {
                const divider = document.createElementNS('http://www.w3.org/2000/svg', 'line');
                divider.classList.add('port-divider');
                divider.setAttribute('x1', padding + halfWidth);
                divider.setAttribute('y1', rowY);
                divider.setAttribute('x2', padding + halfWidth);
                divider.setAttribute('y2', rowY + NODE_DESIGN.rowHeight);
                parentG.appendChild(divider);
            }
        }
    }

    renderTableHeader(parentG, startY, padding, halfWidth, headerHeight, hasInputs, hasOutputs) {
        // Left header (Inputs)
        if (hasInputs) {
            const leftHeader = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
            leftHeader.classList.add('port-table-header-bg');
            leftHeader.setAttribute('x', padding);
            leftHeader.setAttribute('y', startY);
            leftHeader.setAttribute('width', halfWidth);
            leftHeader.setAttribute('height', headerHeight);
            parentG.appendChild(leftHeader);

            const leftLabel = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            leftLabel.classList.add('port-table-header-label');
            leftLabel.setAttribute('x', padding + halfWidth / 2);
            leftLabel.setAttribute('y', startY + headerHeight / 2 + 4);
            leftLabel.setAttribute('text-anchor', 'middle');
            leftLabel.textContent = 'Inputs';
            parentG.appendChild(leftLabel);
        }

        // Right header (Outputs)
        if (hasOutputs) {
            const rightHeader = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
            rightHeader.classList.add('port-table-header-bg');
            rightHeader.setAttribute('x', padding + halfWidth);
            rightHeader.setAttribute('y', startY);
            rightHeader.setAttribute('width', halfWidth);
            rightHeader.setAttribute('height', headerHeight);
            parentG.appendChild(rightHeader);

            const rightLabel = document.createElementNS('http://www.w3.org/2000/svg', 'text');
            rightLabel.classList.add('port-table-header-label');
            rightLabel.setAttribute('x', padding + halfWidth + halfWidth / 2);
            rightLabel.setAttribute('y', startY + headerHeight / 2 + 4);
            rightLabel.setAttribute('text-anchor', 'middle');
            rightLabel.textContent = 'Outputs';
            parentG.appendChild(rightLabel);
        }

        // Divider between headers
        if (hasInputs && hasOutputs) {
            const divider = document.createElementNS('http://www.w3.org/2000/svg', 'line');
            divider.classList.add('port-divider');
            divider.setAttribute('x1', padding + halfWidth);
            divider.setAttribute('y1', startY);
            divider.setAttribute('x2', padding + halfWidth);
            divider.setAttribute('y2', startY + headerHeight);
            parentG.appendChild(divider);
        }
    }

    renderPortCell(parentG, portName, x, y, width, height, type, imageId = null) {
        const cellGroup = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        cellGroup.classList.add('port-cell');
        cellGroup.classList.add(`port-cell-${type}`);

        if (type === 'input') {
            cellGroup.setAttribute('data-input-name', portName);
        } else {
            cellGroup.setAttribute('data-output-name', portName);
        }

        // Store image ID if present
        const hasImage = imageId !== null && imageId !== '';
        if (hasImage) {
            cellGroup.setAttribute('data-image-id', imageId);
        }

        // Cell background
        const cellBg = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
        cellBg.classList.add('port-cell-bg');
        cellBg.setAttribute('x', x);
        cellBg.setAttribute('y', y);
        cellBg.setAttribute('width', width);
        cellBg.setAttribute('height', height);
        cellGroup.appendChild(cellBg);

        // Port circle at node edge
        const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
        circle.classList.add('port');
        // Push ports to the very edge of the node (0 for inputs, NODE_DESIGN.width for outputs)
        circle.setAttribute('cx', type === 'input' ? 0 : NODE_DESIGN.width);
        circle.setAttribute('cy', y + height / 2);
        circle.setAttribute('r', NODE_DESIGN.ports.radius);
        if (hasImage) {
            circle.style.fill = '#27ae60';
        }
        cellGroup.appendChild(circle);

        // Label centered in cell
        const label = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        label.classList.add('port-label');
        label.setAttribute('x', x + width / 2);
        label.setAttribute('y', y + height / 2 + 4);
        label.setAttribute('text-anchor', 'middle');
        label.textContent = portName;
        cellGroup.appendChild(label);

        // Add hover event handlers to show image preview
        if (hasImage) {
            cellGroup.addEventListener('mouseenter', () => {
                this.updateThumbnail(parentG, imageId);
            });

            cellGroup.addEventListener('mouseleave', () => {
                this.restoreDefaultThumbnail(parentG);
            });
        }

        parentG.appendChild(cellGroup);
    }

    renderConnection(sourceNodeId, sourceOutput, targetNodeId, targetInput, hasImage) {
        const sourcePosNode = this.nodePositions.get(sourceNodeId);
        const targetPosNode = this.nodePositions.get(targetNodeId);

        if (!sourcePosNode || !targetPosNode) return;

        // Find port positions
        const sourceNode = this.nodesLayer.querySelector(`[data-node-id="${sourceNodeId}"]`);
        const targetNode = this.nodesLayer.querySelector(`[data-node-id="${targetNodeId}"]`);

        if (!sourceNode || !targetNode) return;

        const sourcePort = sourceNode.querySelector(`[data-output-name="${sourceOutput}"] .port`);
        const targetPort = targetNode.querySelector(`[data-input-name="${targetInput}"] .port`);

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
        btnCircle.setAttribute('r', CONNECTION_DELETE_BUTTON.radius);
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

    updateNodeConnections(nodeId) {
        // Find and update all connections involving this node (both as source and target)
        const graph = this.graphState.getCurrentGraph();
        if (!graph) return;

        // Remove old connections involving this node
        const connections = this.connectionsLayer.querySelectorAll(
            `[data-from-node="${nodeId}"], [data-to-node="${nodeId}"]`
        );
        connections.forEach(conn => conn.remove());

        // Find the node in the graph
        const node = graph.nodes.find(n => n.id === nodeId);
        if (!node) return;

        // Re-render connections FROM this node
        (node.outputs || []).forEach(output => {
            (output.connections || []).forEach(conn => {
                this.renderConnection(
                    node.id,
                    output.name,
                    conn.node_id,
                    conn.input_name,
                    output.image_id !== null && output.image_id !== ''
                );
            });
        });

        // Re-render connections TO this node (from other nodes)
        graph.nodes.forEach(otherNode => {
            if (otherNode.id === nodeId) return;
            (otherNode.outputs || []).forEach(output => {
                (output.connections || []).forEach(conn => {
                    if (conn.node_id === nodeId) {
                        this.renderConnection(
                            otherNode.id,
                            output.name,
                            conn.node_id,
                            conn.input_name,
                            output.image_id !== null && output.image_id !== ''
                        );
                    }
                });
            });
        });
    }

    getNodePosition(nodeId) {
        return this.nodePositions.get(nodeId);
    }

    // Export current viewport state
    exportViewport() {
        return {
            zoom: this.zoom,
            panX: this.panX,
            panY: this.panY
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
        // Clear existing positions when switching graphs
        this.nodePositions.clear();

        if (nodePositions) {
            // nodePositions is now an array of {node_id, x, y}
            for (const pos of nodePositions) {
                this.nodePositions.set(pos.node_id, { x: pos.x, y: pos.y });
            }
        }
    }
}
