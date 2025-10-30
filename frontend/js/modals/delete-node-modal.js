// Delete node modal controller
import { Modal } from '../modal.js';

export class DeleteNodeModalController {
    constructor(api, graphState, graphManager, toastManager, modalManager, interactions) {
        this.api = api;
        this.graphState = graphState;
        this.graphManager = graphManager;
        this.toastManager = toastManager;

        // DOM elements
        this.nodeNameElement = document.getElementById('delete-node-name');
        this.confirmBtn = document.getElementById('delete-node-confirm-btn');
        this.cancelBtn = document.getElementById('delete-node-cancel-btn');

        // State
        this.currentNodeId = null;

        // Create and register modal
        this.modal = new Modal('delete-node-modal', {
            onOpen: () => interactions.cancelAllDrags(),
            onClose: () => {
                this.currentNodeId = null;
            }
        });
        modalManager.register(this.modal);

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.cancelBtn.addEventListener('click', () => this.close());
        this.confirmBtn.addEventListener('click', () => this.handleDelete());
    }

    open(nodeId) {
        const node = this.graphState.getNode(nodeId);
        if (!node) return;

        this.currentNodeId = nodeId;
        this.nodeNameElement.textContent = node.name;
        this.modal.open();
    }

    close() {
        this.modal.close();
    }

    async handleDelete() {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId || !this.currentNodeId) return;

        try {
            const node = this.graphState.getNode(this.currentNodeId);
            const nodeName = node ? node.name : 'Node';

            await this.api.deleteNode(graphId, this.currentNodeId);
            this.close();
            // Refresh graph to show node removed
            await this.graphManager.reloadCurrentGraph();
            this.toastManager.success(`"${nodeName}" deleted successfully`);
        } catch (error) {
            console.error('Failed to delete node:', error);
            this.toastManager.error(`Failed to delete node: ${error.message}`);
        }
    }
}
