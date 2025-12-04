// View JSON modal controller for nodes
import { Modal } from '../modal.js';

export class ViewJsonModalController {
    constructor(modalManager, graphState, onOpenNode) {
        this.graphState = graphState;
        this.onOpenNode = onOpenNode;
        this.titleElement = document.getElementById('view-json-title');
        this.preElement = document.getElementById('view-json-pre');
        this.closeBtn = document.getElementById('view-json-close-btn');

        this.modal = new Modal('view-json-modal');
        modalManager.register(this.modal);

        this.closeBtn.addEventListener('click', () => this.close());
        this.preElement.addEventListener('click', (e) => this.handleLinkClick(e));
    }

    open(node) {
        if (!node) return;

        this.titleElement.textContent = `${node.name || 'Node'} JSON`;
        this.preElement.innerHTML = this.renderJson(node);
        this.modal.open();
    }

    renderJson(node) {
        const jsonString = JSON.stringify(node, null, 2);
        // Linkify node_id values (case-insensitive)
        return jsonString.replace(/"node_id":\s*"([A-Fa-f0-9-]+)"/g, (_match, id) => {
            return `"node_id": "<a href=\"#\" data-node-id=\"${id}\">${id}</a>"`;
        });
    }

    handleLinkClick(e) {
        const link = e.target.closest('[data-node-id]');
        if (!link) return;
        e.preventDefault();
        const nodeId = link.getAttribute('data-node-id');
        const node = this.graphState.getNode(nodeId);
        if (node) {
            this.open(node);
        } else if (this.onOpenNode) {
            this.onOpenNode(nodeId);
        }
    }

    close() {
        this.modal.close();
    }
}
