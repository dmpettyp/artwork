// Interactive crop modal with visual image cropping
import { API_PATHS } from './constants.js';

export class CropModal {
    constructor() {
        console.log('[CropModal.constructor] Initializing...');
        this.modal = document.getElementById('crop-modal');
        console.log('[CropModal.constructor] modal element:', this.modal);
        this.canvas = document.getElementById('crop-canvas');
        console.log('[CropModal.constructor] canvas element:', this.canvas);
        this.ctx = this.canvas.getContext('2d');

        // Input fields
        this.leftInput = document.getElementById('crop-left');
        this.rightInput = document.getElementById('crop-right');
        this.topInput = document.getElementById('crop-top');
        this.bottomInput = document.getElementById('crop-bottom');

        // Buttons
        this.cancelBtn = document.getElementById('crop-cancel-btn');
        this.saveBtn = document.getElementById('crop-save-btn');
        console.log('[CropModal.constructor] All elements found');

        // State
        this.image = null;
        this.imageScale = 1;
        this.imageOffsetX = 0;
        this.imageOffsetY = 0;
        this.cropRect = { left: 0, top: 0, right: 100, bottom: 100 };
        this.dragState = null; // null, 'move', or handle identifier
        this.dragStartX = 0;
        this.dragStartY = 0;
        this.dragStartRect = null;

        this.onSave = null;

        this.setupEventListeners();
    }

    setupEventListeners() {
        // Modal controls
        this.cancelBtn.addEventListener('click', () => this.hide());
        this.saveBtn.addEventListener('click', () => this.handleSave());

        // Canvas interaction
        this.canvas.addEventListener('mousedown', (e) => this.handleMouseDown(e));
        this.canvas.addEventListener('mousemove', (e) => this.handleMouseMove(e));
        this.canvas.addEventListener('mouseup', (e) => this.handleMouseUp(e));
        this.canvas.addEventListener('mouseleave', (e) => this.handleMouseUp(e));

        // Input field changes
        this.leftInput.addEventListener('input', () => this.handleFieldChange());
        this.rightInput.addEventListener('input', () => this.handleFieldChange());
        this.topInput.addEventListener('input', () => this.handleFieldChange());
        this.bottomInput.addEventListener('input', () => this.handleFieldChange());
    }

    async show(inputImageId, existingConfig = {}) {
        console.log('[CropModal.show] Starting with inputImageId:', inputImageId, 'existingConfig:', existingConfig);

        // Set initial crop values
        if (existingConfig.left !== undefined) {
            this.cropRect = {
                left: existingConfig.left,
                right: existingConfig.right,
                top: existingConfig.top,
                bottom: existingConfig.bottom
            };
            console.log('[CropModal.show] Set cropRect from config:', this.cropRect);
        }

        // Load the input image
        if (inputImageId) {
            try {
                console.log('[CropModal.show] Fetching image:', inputImageId);
                const response = await fetch(API_PATHS.images(inputImageId));
                console.log('[CropModal.show] Fetch response:', response.ok, response.status);
                if (!response.ok) throw new Error('Failed to load image');

                const blob = await response.blob();
                const imageUrl = URL.createObjectURL(blob);
                console.log('[CropModal.show] Created blob URL:', imageUrl);

                // Wait for image to load before showing modal
                await new Promise((resolve, reject) => {
                    this.image = new Image();
                    this.image.onload = () => {
                        console.log('[CropModal.show] Image loaded successfully');
                        URL.revokeObjectURL(imageUrl);
                        this.fitImageToCanvas();
                        this.updateFieldsFromRect();
                        this.render();
                        resolve();
                    };
                    this.image.onerror = () => {
                        console.error('[CropModal.show] Image load error');
                        URL.revokeObjectURL(imageUrl);
                        reject(new Error('Failed to load image'));
                    };
                    this.image.src = imageUrl;
                    console.log('[CropModal.show] Set image.src, waiting for load...');
                });
            } catch (error) {
                console.error('Failed to load image for crop:', error);
                alert('Failed to load input image. Make sure the crop node has an input connection.');
                return;
            }
        } else {
            console.error('[CropModal.show] No inputImageId provided');
            alert('Cannot open crop editor: No input image available. Connect an input to this crop node first.');
            return;
        }

        console.log('[CropModal.show] Adding active class to modal');
        this.modal.classList.add('active');
        console.log('[CropModal.show] Modal should now be visible');
    }

    hide() {
        this.modal.classList.remove('active');
        this.image = null;
    }

    fitImageToCanvas() {
        if (!this.image) return;

        const canvasWidth = this.canvas.width;
        const canvasHeight = this.canvas.height;
        const imageAspect = this.image.width / this.image.height;
        const canvasAspect = canvasWidth / canvasHeight;

        if (imageAspect > canvasAspect) {
            // Image is wider - fit to width
            this.imageScale = canvasWidth / this.image.width;
        } else {
            // Image is taller - fit to height
            this.imageScale = canvasHeight / this.image.height;
        }

        const scaledWidth = this.image.width * this.imageScale;
        const scaledHeight = this.image.height * this.imageScale;

        this.imageOffsetX = (canvasWidth - scaledWidth) / 2;
        this.imageOffsetY = (canvasHeight - scaledHeight) / 2;

        // Set default crop to full image if not set
        if (this.cropRect.right === 100 && this.cropRect.bottom === 100) {
            this.cropRect = {
                left: 0,
                top: 0,
                right: this.image.width,
                bottom: this.image.height
            };
        }
    }

    render() {
        if (!this.image) return;

        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

        // Draw image
        const scaledWidth = this.image.width * this.imageScale;
        const scaledHeight = this.image.height * this.imageScale;
        this.ctx.drawImage(
            this.image,
            this.imageOffsetX,
            this.imageOffsetY,
            scaledWidth,
            scaledHeight
        );

        // Draw crop overlay
        this.drawCropOverlay();
    }

    drawCropOverlay() {
        const rect = this.imageCoordsToCanvas(this.cropRect);

        // Semi-transparent overlay outside crop area
        this.ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';

        // Top
        this.ctx.fillRect(this.imageOffsetX, this.imageOffsetY,
            this.image.width * this.imageScale, rect.top - this.imageOffsetY);
        // Bottom
        this.ctx.fillRect(this.imageOffsetX, rect.bottom,
            this.image.width * this.imageScale,
            (this.imageOffsetY + this.image.height * this.imageScale) - rect.bottom);
        // Left
        this.ctx.fillRect(this.imageOffsetX, rect.top,
            rect.left - this.imageOffsetX, rect.bottom - rect.top);
        // Right
        this.ctx.fillRect(rect.right, rect.top,
            (this.imageOffsetX + this.image.width * this.imageScale) - rect.right,
            rect.bottom - rect.top);

        // Crop rectangle border
        this.ctx.strokeStyle = '#3498db';
        this.ctx.lineWidth = 2;
        this.ctx.strokeRect(rect.left, rect.top, rect.right - rect.left, rect.bottom - rect.top);

        // Resize handles
        this.drawHandle(rect.left, rect.top); // Top-left
        this.drawHandle(rect.right, rect.top); // Top-right
        this.drawHandle(rect.left, rect.bottom); // Bottom-left
        this.drawHandle(rect.right, rect.bottom); // Bottom-right
        this.drawHandle((rect.left + rect.right) / 2, rect.top); // Top-mid
        this.drawHandle((rect.left + rect.right) / 2, rect.bottom); // Bottom-mid
        this.drawHandle(rect.left, (rect.top + rect.bottom) / 2); // Left-mid
        this.drawHandle(rect.right, (rect.top + rect.bottom) / 2); // Right-mid
    }

    drawHandle(x, y) {
        this.ctx.fillStyle = '#3498db';
        this.ctx.fillRect(x - 4, y - 4, 8, 8);
    }

    imageCoordsToCanvas(imageCoords) {
        return {
            left: this.imageOffsetX + imageCoords.left * this.imageScale,
            top: this.imageOffsetY + imageCoords.top * this.imageScale,
            right: this.imageOffsetX + imageCoords.right * this.imageScale,
            bottom: this.imageOffsetY + imageCoords.bottom * this.imageScale
        };
    }

    canvasToImageCoords(canvasX, canvasY) {
        return {
            x: (canvasX - this.imageOffsetX) / this.imageScale,
            y: (canvasY - this.imageOffsetY) / this.imageScale
        };
    }

    getHandleAtPoint(x, y) {
        const rect = this.imageCoordsToCanvas(this.cropRect);
        const threshold = 8;

        const handles = {
            'tl': { x: rect.left, y: rect.top },
            'tr': { x: rect.right, y: rect.top },
            'bl': { x: rect.left, y: rect.bottom },
            'br': { x: rect.right, y: rect.bottom },
            't': { x: (rect.left + rect.right) / 2, y: rect.top },
            'b': { x: (rect.left + rect.right) / 2, y: rect.bottom },
            'l': { x: rect.left, y: (rect.top + rect.bottom) / 2 },
            'r': { x: rect.right, y: (rect.top + rect.bottom) / 2 }
        };

        for (const [handle, pos] of Object.entries(handles)) {
            if (Math.abs(x - pos.x) <= threshold && Math.abs(y - pos.y) <= threshold) {
                return handle;
            }
        }

        return null;
    }

    isPointInRect(x, y) {
        const rect = this.imageCoordsToCanvas(this.cropRect);
        return x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom;
    }

    handleMouseDown(e) {
        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const handle = this.getHandleAtPoint(x, y);
        if (handle) {
            this.dragState = handle;
        } else if (this.isPointInRect(x, y)) {
            this.dragState = 'move';
        } else {
            return;
        }

        this.dragStartX = x;
        this.dragStartY = y;
        this.dragStartRect = { ...this.cropRect };

        this.canvas.style.cursor = 'grabbing';
    }

    handleMouseMove(e) {
        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        if (!this.dragState) {
            // Update cursor based on hover
            const handle = this.getHandleAtPoint(x, y);
            if (handle) {
                const cursors = {
                    'tl': 'nw-resize', 'tr': 'ne-resize',
                    'bl': 'sw-resize', 'br': 'se-resize',
                    't': 'n-resize', 'b': 's-resize',
                    'l': 'w-resize', 'r': 'e-resize'
                };
                this.canvas.style.cursor = cursors[handle] || 'default';
            } else if (this.isPointInRect(x, y)) {
                this.canvas.style.cursor = 'move';
            } else {
                this.canvas.style.cursor = 'default';
            }
            return;
        }

        const dx = x - this.dragStartX;
        const dy = y - this.dragStartY;
        const dxImage = dx / this.imageScale;
        const dyImage = dy / this.imageScale;

        if (this.dragState === 'move') {
            // Move entire rectangle
            const width = this.dragStartRect.right - this.dragStartRect.left;
            const height = this.dragStartRect.bottom - this.dragStartRect.top;

            let newLeft = this.dragStartRect.left + dxImage;
            let newTop = this.dragStartRect.top + dyImage;

            // Constrain to image bounds
            newLeft = Math.max(0, Math.min(newLeft, this.image.width - width));
            newTop = Math.max(0, Math.min(newTop, this.image.height - height));

            this.cropRect = {
                left: Math.round(newLeft),
                top: Math.round(newTop),
                right: Math.round(newLeft + width),
                bottom: Math.round(newTop + height)
            };
        } else {
            // Resize from handle
            let newRect = { ...this.dragStartRect };

            if (this.dragState.includes('l')) newRect.left += dxImage;
            if (this.dragState.includes('r')) newRect.right += dxImage;
            if (this.dragState.includes('t')) newRect.top += dyImage;
            if (this.dragState.includes('b')) newRect.bottom += dyImage;

            // Ensure minimum size and valid bounds
            newRect.left = Math.max(0, Math.min(newRect.left, newRect.right - 10));
            newRect.right = Math.min(this.image.width, Math.max(newRect.right, newRect.left + 10));
            newRect.top = Math.max(0, Math.min(newRect.top, newRect.bottom - 10));
            newRect.bottom = Math.min(this.image.height, Math.max(newRect.bottom, newRect.top + 10));

            this.cropRect = {
                left: Math.round(newRect.left),
                top: Math.round(newRect.top),
                right: Math.round(newRect.right),
                bottom: Math.round(newRect.bottom)
            };
        }

        this.updateFieldsFromRect();
        this.render();
    }

    handleMouseUp(e) {
        this.dragState = null;
        this.canvas.style.cursor = 'default';
    }

    updateFieldsFromRect() {
        this.leftInput.value = this.cropRect.left;
        this.rightInput.value = this.cropRect.right;
        this.topInput.value = this.cropRect.top;
        this.bottomInput.value = this.cropRect.bottom;
    }

    handleFieldChange() {
        // Update crop rect from fields
        const left = parseInt(this.leftInput.value) || 0;
        const right = parseInt(this.rightInput.value) || this.image.width;
        const top = parseInt(this.topInput.value) || 0;
        const bottom = parseInt(this.bottomInput.value) || this.image.height;

        // Validate
        if (left < right && top < bottom) {
            this.cropRect = { left, right, top, bottom };
            this.render();
        }
    }

    handleSave() {
        if (this.onSave) {
            this.onSave({
                left: this.cropRect.left,
                right: this.cropRect.right,
                top: this.cropRect.top,
                bottom: this.cropRect.bottom
            });
        }
        this.hide();
    }
}
