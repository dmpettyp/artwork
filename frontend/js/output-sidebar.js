// Output sidebar management for displaying output node images
import { API_PATHS } from './constants.js';

export class OutputSidebar {
    constructor(graphState, renderer, toastManager) {
        this.container = document.querySelector('.sidebar-content');
        this.graphState = graphState;
        this.renderer = renderer;
        this.toastManager = toastManager;

        // Subscribe to graph changes
        this.graphState.subscribe(graph => this.render(graph));
    }

    render(graph) {
        if (!graph) {
            this.container.innerHTML = '<p style="color: #7f8c8d; text-align: center; margin-top: 20px;">No graph selected</p>';
            return;
        }

        // Filter output nodes
        const outputNodes = graph.nodes.filter(node => node.type === 'output');

        if (outputNodes.length === 0) {
            this.container.innerHTML = '<p style="color: #7f8c8d; text-align: center; margin-top: 20px;">No output nodes</p>';
            return;
        }

        // Sort output nodes by y position (top to bottom)
        const sortedOutputNodes = outputNodes.sort((a, b) => {
            const posA = this.renderer.getNodePosition(a.id);
            const posB = this.renderer.getNodePosition(b.id);

            // If positions aren't available, maintain original order
            if (!posA || !posB) return 0;
            return posA.y - posB.y;
        });

        // Render each output node
        this.container.innerHTML = ''; // Clear existing content

        sortedOutputNodes.forEach(node => {
            const output = node.outputs?.find(o => o.name === 'final');
            const card = this.createOutputCard(node, output);
            this.container.appendChild(card);
        });
    }

    createOutputCard(node, output) {
        const hasImage = output?.image_id;

        // Create output card elements using DOM API (safe from XSS)
        const card = document.createElement('div');
        card.className = 'output-card';

        const header = document.createElement('div');
        header.className = 'output-card-header';
        header.textContent = node.name; // textContent is safe from XSS
        card.appendChild(header);

        const body = document.createElement('div');
        body.className = 'output-card-body';

        if (hasImage) {
            const img = document.createElement('img');
            img.src = API_PATHS.images(output.image_id);
            img.alt = node.name; // alt attribute is also escaped
            img.className = 'output-card-image';
            body.appendChild(img);

            // Add download button
            const downloadBtn = document.createElement('button');
            downloadBtn.className = 'output-card-download-btn';
            downloadBtn.textContent = 'Download';
            downloadBtn.onclick = () => this.handleImageDownload(output.image_id, node.name);
            body.appendChild(downloadBtn);
        } else {
            const placeholder = document.createElement('p');
            placeholder.className = 'output-card-placeholder';
            placeholder.textContent = 'No image yet';
            body.appendChild(placeholder);
        }

        card.appendChild(body);
        return card;
    }

    async handleImageDownload(imageId, nodeName) {
        try {
            const response = await fetch(API_PATHS.images(imageId));
            const blob = await response.blob();

            // Determine extension from content type
            const contentType = response.headers.get('content-type') || 'image/png';
            const extension = contentType.split('/')[1] || 'png';

            // Create filename: {imagegraph_name}-{output_node_name}.{extension}
            // Convert to lowercase and replace spaces with underscores
            const currentGraph = this.graphState.getCurrentGraph();
            const graphName = currentGraph?.name || 'graph';
            const sanitizedGraphName = graphName.toLowerCase().replace(/ /g, '_');
            const sanitizedNodeName = nodeName.toLowerCase().replace(/ /g, '_');
            const filename = `${sanitizedGraphName}-${sanitizedNodeName}.${extension}`;

            // Create download link and trigger
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);

            // Show success toast
            this.toastManager.success(`Downloaded ${filename}`);
        } catch (error) {
            console.error('Failed to download image:', error);
            this.toastManager.error('Failed to download image');
        }
    }
}
