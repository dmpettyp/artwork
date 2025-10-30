// Create graph modal controller
import { Modal } from '../modal.js';

export class CreateGraphModalController {
    constructor(api, graphManager, toastManager, modalManager, interactions) {
        this.api = api;
        this.graphManager = graphManager;
        this.toastManager = toastManager;

        // DOM elements
        this.nameInput = document.getElementById('graph-name-input');
        this.createBtn = document.getElementById('modal-create-btn');
        this.cancelBtn = document.getElementById('modal-cancel-btn');

        // Create and register modal
        this.modal = new Modal('create-graph-modal', {
            onOpen: () => {
                interactions.cancelAllDrags();
                this.onOpen();
            }
        });
        modalManager.register(this.modal);

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.cancelBtn.addEventListener('click', () => this.close());
        this.createBtn.addEventListener('click', () => this.handleCreate());
        this.nameInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.handleCreate();
            }
        });
    }

    onOpen() {
        this.nameInput.value = '';
        this.nameInput.focus();
    }

    open() {
        this.modal.open();
    }

    close() {
        this.modal.close();
    }

    async handleCreate() {
        const name = this.nameInput.value.trim();
        if (!name) return;

        try {
            const graph_id = await this.api.createImageGraph(name);
            this.close();
            await this.graphManager.loadGraphList();
            await this.graphManager.selectGraph(graph_id);
            this.toastManager.success(`Graph "${name}" created successfully`);
        } catch (error) {
            console.error('Failed to create graph:', error);
            this.toastManager.error(`Failed to create graph: ${error.message}`);
        }
    }
}
