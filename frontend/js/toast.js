// Toast notification system

export class Toast {
    constructor(message, type = 'info', duration = 4000) {
        this.message = message;
        this.type = type; // 'success', 'error', 'info', 'warning'
        this.duration = duration;
        this.element = null;
        this.timeoutId = null;
    }

    create() {
        this.element = document.createElement('div');
        this.element.className = `toast toast-${this.type}`;

        const messageSpan = document.createElement('span');
        messageSpan.className = 'toast-message';
        messageSpan.textContent = this.message;

        const closeBtn = document.createElement('button');
        closeBtn.className = 'toast-close';
        closeBtn.innerHTML = '&times;';
        closeBtn.addEventListener('click', () => this.dismiss());

        this.element.appendChild(messageSpan);
        this.element.appendChild(closeBtn);

        return this.element;
    }

    show(container) {
        if (!this.element) {
            this.create();
        }

        container.appendChild(this.element);

        // Trigger animation
        requestAnimationFrame(() => {
            this.element.classList.add('toast-visible');
        });

        // Auto-dismiss after duration (unless duration is 0)
        if (this.duration > 0) {
            this.timeoutId = setTimeout(() => {
                this.dismiss();
            }, this.duration);
        }
    }

    dismiss() {
        if (this.timeoutId) {
            clearTimeout(this.timeoutId);
        }

        if (this.element) {
            this.element.classList.remove('toast-visible');
            this.element.classList.add('toast-hiding');

            // Remove from DOM after animation
            setTimeout(() => {
                if (this.element && this.element.parentNode) {
                    this.element.parentNode.removeChild(this.element);
                }
            }, 300);
        }
    }
}

export class ToastManager {
    constructor() {
        this.container = this.createContainer();
        this.toasts = [];
    }

    createContainer() {
        let container = document.getElementById('toast-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'toast-container';
            container.className = 'toast-container';
            document.body.appendChild(container);
        }
        return container;
    }

    show(message, type = 'info', duration = 4000) {
        const toast = new Toast(message, type, duration);
        this.toasts.push(toast);
        toast.show(this.container);

        // Clean up reference when dismissed
        setTimeout(() => {
            const index = this.toasts.indexOf(toast);
            if (index > -1) {
                this.toasts.splice(index, 1);
            }
        }, duration + 300);

        return toast;
    }

    success(message, duration = 4000) {
        return this.show(message, 'success', duration);
    }

    error(message, duration = 6000) {
        return this.show(message, 'error', duration);
    }

    info(message, duration = 4000) {
        return this.show(message, 'info', duration);
    }

    warning(message, duration = 5000) {
        return this.show(message, 'warning', duration);
    }

    dismissAll() {
        this.toasts.forEach(toast => toast.dismiss());
        this.toasts = [];
    }
}
