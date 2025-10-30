// Edit node modal controller
import { Modal } from '../modal.js';

export class EditNodeModalController {
    constructor(api, graphState, graphManager, formBuilder, cropModal, toastManager, modalManager, interactions, getNodeTypeConfigs) {
        this.api = api;
        this.graphState = graphState;
        this.graphManager = graphManager;
        this.formBuilder = formBuilder;
        this.cropModal = cropModal;
        this.toastManager = toastManager;
        this.getNodeTypeConfigs = getNodeTypeConfigs;

        // DOM elements
        this.titleElement = document.getElementById('edit-config-modal-title');
        this.nameInput = document.getElementById('edit-node-name-input');
        this.imageUpload = document.getElementById('edit-image-upload');
        this.imageInput = document.getElementById('edit-image-input');
        this.configFields = document.getElementById('edit-config-fields');
        this.saveBtn = document.getElementById('edit-config-save-btn');
        this.cancelBtn = document.getElementById('edit-config-cancel-btn');

        // State
        this.currentNodeId = null;

        // Create and register modal
        this.modal = new Modal('edit-config-modal', {
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
        this.saveBtn.addEventListener('click', () => this.handleSave());
    }

    // Helper function to get the input image ID for a node
    getNodeInputImageId(nodeId) {
        const node = this.graphState.getNode(nodeId);
        if (!node) return null;

        // Find the first connected input with an image
        const connectedInput = node.inputs?.find(input => input.connected && input.image_id);
        return connectedInput?.image_id || null;
    }

    async open(nodeId) {
        const node = this.graphState.getNode(nodeId);
        if (!node) return;

        // For crop nodes, show the custom crop modal instead
        if (node.type === 'crop') {
            await this.openCropEditModal(nodeId);
            return;
        }

        this.currentNodeId = nodeId;
        this.nameInput.value = node.name || '';

        // Update modal title
        const configs = this.getNodeTypeConfigs();
        const displayName = configs[node.type]?.name || node.type;
        this.titleElement.textContent = `Edit ${displayName} Node`;

        // Update name field based on whether it's required
        const nameRequired = configs[node.type]?.nameRequired !== false;
        this.nameInput.required = nameRequired;
        this.nameInput.placeholder = nameRequired ? 'Enter node name' : 'Enter node name (optional)';

        // Show/hide image upload based on node type
        if (node.type === 'input') {
            this.imageUpload.style.display = 'block';
            this.imageInput.value = ''; // Clear any previous file selection
        } else {
            this.imageUpload.style.display = 'none';
            this.imageInput.value = '';
        }

        // Render config fields based on node type
        this.formBuilder.renderFields(this.configFields, node.type, 'edit-config', node.config);

        this.modal.open();
    }

    close() {
        this.modal.close();
    }

    async openCropEditModal(nodeId) {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId) return;

        const node = this.graphState.getNode(nodeId);
        if (!node) return;

        // Get the input image from connected nodes
        const inputImageId = this.getNodeInputImageId(nodeId);

        // Show the crop modal with current node name and config
        await this.cropModal.show(inputImageId, node.config, node.name || '');

        // Set up the save callback
        this.cropModal.onSave = async (data) => {
            try {
                const { name, config } = data;

                // Determine what changed
                const nameChanged = name !== (node.name || '');
                const configChanged = JSON.stringify(config) !== JSON.stringify(node.config);

                if (nameChanged || configChanged) {
                    await this.api.updateNode(
                        graphId,
                        nodeId,
                        nameChanged ? name : null,
                        configChanged ? config : null
                    );
                    await this.graphManager.reloadCurrentGraph();
                    this.toastManager.success('Crop node updated');
                }
            } catch (error) {
                console.error('Failed to update crop config:', error);
                this.toastManager.error(`Failed to update crop: ${error.message}`);
            }
        };
    }

    async handleSave() {
        const graphId = this.graphState.getCurrentGraphId();
        if (!graphId || !this.currentNodeId) return;

        const node = this.graphState.getNode(this.currentNodeId);
        const newName = this.nameInput.value.trim();

        // Check if name is required for this node type
        const configs = this.getNodeTypeConfigs();
        const nameRequired = configs[node.type]?.nameRequired !== false;
        if (nameRequired && !newName) {
            this.toastManager.warning('Please enter a node name');
            return;
        }

        // Validate config fields
        const validation = this.formBuilder.validate(this.configFields, node.type);
        if (!validation.valid) {
            this.toastManager.warning(validation.errors[0]);
            return;
        }

        const config = this.formBuilder.getValues(this.configFields);

        // Determine what changed (handle empty string vs null for optional names)
        const oldName = node.name || '';
        const nameChanged = newName !== oldName;
        const configChanged = JSON.stringify(config) !== JSON.stringify(node.config);
        const imageChanged = node.type === 'input' && this.imageInput.files.length > 0;

        if (!nameChanged && !configChanged && !imageChanged) {
            this.close();
            return;
        }

        try {
            // Update name and/or config if changed
            if (nameChanged || configChanged) {
                await this.api.updateNode(
                    graphId,
                    this.currentNodeId,
                    nameChanged ? newName : null,
                    configChanged ? config : null
                );
            }

            // Upload new image if selected for input node
            if (imageChanged) {
                const imageFile = this.imageInput.files[0];
                await this.api.uploadNodeOutputImage(graphId, this.currentNodeId, 'original', imageFile);
            }

            this.close();
            // Refresh graph to show updates
            await this.graphManager.reloadCurrentGraph();
            this.toastManager.success('Node updated successfully');
        } catch (error) {
            console.error('Failed to update node:', error);
            this.toastManager.error(`Failed to update node: ${error.message}`);
        }
    }
}
