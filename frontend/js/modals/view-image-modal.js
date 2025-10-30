// View image modal controller
import { Modal } from '../modal.js';
import { API_PATHS } from '../constants.js';

export class ViewImageModalController {
    constructor(graphState, modalManager, interactions) {
        this.graphState = graphState;

        // DOM elements
        this.titleElement = document.getElementById('view-image-title');
        this.imageElement = document.getElementById('view-image-img');
        this.messageElement = document.getElementById('view-image-message');
        this.closeBtn = document.getElementById('view-image-close-btn');

        // State
        this.currentNodeId = null;

        // Create and register modal
        this.modal = new Modal('view-image-modal', {
            onOpen: () => interactions.cancelAllDrags(),
            onClose: () => {
                this.imageElement.src = '';
                this.imageElement.onerror = null;
                this.currentNodeId = null;
            }
        });
        modalManager.register(this.modal);

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.closeBtn.addEventListener('click', () => this.close());
    }

    open(nodeId) {
        const node = this.graphState.getNode(nodeId);
        if (!node) return;

        this.currentNodeId = nodeId;
        this.titleElement.textContent = `${node.name} - Output`;

        // Get the first output
        const outputs = node.outputs || [];
        if (outputs.length === 0 || !outputs[0].image_id) {
            // No output image available
            this.imageElement.style.display = 'none';
            this.messageElement.textContent = 'No output image available for this node.';
            this.messageElement.style.display = 'block';
        } else {
            // Load and display the image
            const imageId = outputs[0].image_id;
            const imageUrl = API_PATHS.images(imageId);

            this.imageElement.src = imageUrl;
            this.imageElement.style.display = 'block';
            this.messageElement.style.display = 'none';

            // Handle image load error
            this.imageElement.onerror = () => {
                this.imageElement.style.display = 'none';
                this.messageElement.textContent = 'Failed to load image.';
                this.messageElement.style.display = 'block';
            };
        }

        this.modal.open();
    }

    close() {
        this.modal.close();
    }
}
