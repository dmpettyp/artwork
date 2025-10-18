// Modal management utilities

export class Modal {
    constructor(modalId, options = {}) {
        this.modal = document.getElementById(modalId);
        if (!this.modal) {
            console.error(`Modal with id "${modalId}" not found`);
            return;
        }

        this.contentElement = this.modal.querySelector('.modal-content');

        // Optional callbacks
        this.onOpen = options.onOpen || (() => {});
        this.onClose = options.onClose || (() => {});
        this.beforeClose = options.beforeClose || (() => true); // Return false to prevent close

        this._setupEventListeners();
    }

    _setupEventListeners() {
        // Close on background click (mousedown + mouseup both on background)
        let mousedownTarget = null;
        this.modal.addEventListener('mousedown', (e) => {
            mousedownTarget = e.target;
        });
        this.modal.addEventListener('mouseup', (e) => {
            if (mousedownTarget === this.modal && e.target === this.modal) {
                this.close();
            }
            mousedownTarget = null;
        });
    }

    open() {
        this.modal.classList.add('active');
        this.onOpen();
    }

    close() {
        if (!this.beforeClose()) {
            return; // Prevent closing if beforeClose returns false
        }
        this.modal.classList.remove('active');
        this.onClose();
    }

    isOpen() {
        return this.modal.classList.contains('active');
    }
}

// Global modal manager to handle ESC key and modal ordering
export class ModalManager {
    constructor() {
        this.modals = [];
        this._setupEscapeHandler();
    }

    register(modal) {
        this.modals.push(modal);
    }

    _setupEscapeHandler() {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                // Close the most recently opened modal
                const openModal = this.modals.find(modal => modal.isOpen());
                if (openModal) {
                    openModal.close();
                }
            }
        });
    }

    closeAll() {
        this.modals.forEach(modal => {
            if (modal.isOpen()) {
                modal.close();
            }
        });
    }
}
