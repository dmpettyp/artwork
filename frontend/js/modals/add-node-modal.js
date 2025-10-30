// Add node modal controller
import { Modal } from '../modal.js';

export class AddNodeModalController {
    constructor(api, graphState, graphManager, renderer, formBuilder, toastManager, modalManager, interactions, getNodeTypeConfigs) {
        this.api = api;
        this.graphState = graphState;
        this.graphManager = graphManager;
        this.renderer = renderer;
        this.formBuilder = formBuilder;
        this.toastManager = toastManager;
        this.getNodeTypeConfigs = getNodeTypeConfigs;

        // DOM elements
        this.modalElement = document.getElementById('add-node-modal');
        this.titleElement = document.getElementById('add-node-modal-title');
        this.nameInput = document.getElementById('node-name-input');
        this.imageUpload = document.getElementById('node-image-upload');
        this.imageInput = document.getElementById('node-image-input');
        this.configFields = document.getElementById('node-config-fields');
        this.createBtn = document.getElementById('add-node-create-btn');
        this.cancelBtn = document.getElementById('add-node-cancel-btn');

        // State
        this.currentNodeType = null;

        // Create and register modal
        this.modal = new Modal('add-node-modal', {
            onOpen: () => interactions.cancelAllDrags()
        });
        modalManager.register(this.modal);

        this.setupEventListeners();
    }

    setupEventListeners() {
        this.cancelBtn.addEventListener('click', () => this.close());
        this.createBtn.addEventListener('click', () => this.handleCreate());
    }

    open(nodeType, position = null) {
        if (!this.graphState.getCurrentGraphId()) {
            this.toastManager.warning('Please select a graph first');
            return;
        }

        // For crop nodes, show info message
        if (nodeType === 'crop') {
            this.toastManager.info('Please add the crop node first, then connect an input and edit it to configure the crop area');
        }

        // Store the node type
        this.currentNodeType = nodeType;

        // Store position if provided (from context menu)
        if (position) {
            this.modalElement.dataset.canvasX = position.x;
            this.modalElement.dataset.canvasY = position.y;
        }

        // Update modal title
        const configs = this.getNodeTypeConfigs();
        const displayName = configs[nodeType]?.name || nodeType;
        this.titleElement.textContent = `Add ${displayName} Node`;

        // Clear inputs
        this.nameInput.value = '';
        this.imageInput.value = '';
        this.configFields.innerHTML = '';

        // Update name field based on whether it's required
        const nameRequired = configs[nodeType]?.nameRequired !== false;
        this.nameInput.required = nameRequired;
        this.nameInput.placeholder = nameRequired ? 'Enter node name' : 'Enter node name (optional)';

        // Show/hide image upload based on node type
        if (nodeType === 'input') {
            this.imageUpload.style.display = 'block';
        } else {
            this.imageUpload.style.display = 'none';
        }

        // Render config fields for the node type
        this.formBuilder.renderFields(this.configFields, nodeType, 'config');

        this.modal.open();
        this.nameInput.focus();
    }

    close() {
        this.modal.close();
        this.currentNodeType = null;
        delete this.modalElement.dataset.canvasX;
        delete this.modalElement.dataset.canvasY;
    }

    async handleCreate() {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId) return;

        const nodeType = this.currentNodeType;
        const nodeName = this.nameInput.value.trim();

        if (!nodeType) {
            this.toastManager.warning('Node type not set');
            return;
        }

        // Check if name is required for this node type
        const configs = this.getNodeTypeConfigs();
        const nameRequired = configs[nodeType]?.nameRequired !== false;
        if (nameRequired && !nodeName) {
            this.toastManager.warning('Please enter a node name');
            return;
        }

        // For input nodes, check if an image file is selected
        if (nodeType === 'input' && this.imageInput.files.length === 0) {
            this.toastManager.warning('Please select an image file for the input node');
            return;
        }

        // Validate config fields
        const validation = this.formBuilder.validate(this.configFields, nodeType);
        if (!validation.valid) {
            this.toastManager.warning(validation.errors[0]);
            return;
        }

        const config = this.formBuilder.getValues(this.configFields);

        try {
            // Add the node first to get the node ID
            const nodeId = await this.api.addNode(graphId, nodeType, nodeName, config);

            // If this is an input node with an image, upload it to the "original" output
            if (nodeType === 'input' && this.imageInput.files.length > 0) {
                const imageFile = this.imageInput.files[0];
                await this.api.uploadNodeOutputImage(graphId, nodeId, 'original', imageFile);
            }

            // If position was set from context menu, update node position
            if (this.modalElement.dataset.canvasX && this.modalElement.dataset.canvasY) {
                const x = parseFloat(this.modalElement.dataset.canvasX);
                const y = parseFloat(this.modalElement.dataset.canvasY);
                this.renderer.updateNodePosition(nodeId, x, y);

                // Persist the layout (node positions)
                const nodePositions = this.renderer.exportNodePositions();
                await this.api.updateLayout(graphId, nodePositions);
            }

            this.close();
            // Refresh graph to show new node
            await this.graphManager.reloadCurrentGraph();
            this.toastManager.success(`Node "${nodeName}" added successfully`);
        } catch (error) {
            console.error('Failed to add node:', error);
            this.toastManager.error(`Failed to add node: ${error.message}`);
        }
    }
}
