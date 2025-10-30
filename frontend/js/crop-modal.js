// Interactive crop modal with visual image cropping
import { API_PATHS } from './constants.js';

export class CropModal {
    constructor() {
        this.modal = document.getElementById('crop-modal');
        this.canvas = document.getElementById('crop-canvas');
        this.ctx = this.canvas.getContext('2d');

        // Input fields
        this.nameInput = document.getElementById('crop-node-name');
        this.leftInput = document.getElementById('crop-left');
        this.rightInput = document.getElementById('crop-right');
        this.topInput = document.getElementById('crop-top');
        this.bottomInput = document.getElementById('crop-bottom');

        // Buttons
        this.cancelBtn = document.getElementById('crop-cancel-btn');
        this.saveBtn = document.getElementById('crop-save-btn');

        // State
        this.image = null;
        this.imageScale = 1;
        this.imageOffsetX = 0;
        this.imageOffsetY = 0;
        this.cropRect = { left: 0, top: 0, right: 100, bottom: 100 };
        this.isDrawing = false;
        this.isDragging = false;
        this.drawStartX = 0;
        this.drawStartY = 0;
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
        this.canvas.addEventListener('wheel', (e) => this.handleWheel(e), { passive: false });

        // Continue drawing even when mouse leaves canvas
        document.addEventListener('mousemove', (e) => {
            if (this.isDrawing) {
                this.handleMouseMove(e);
            }
        });
        document.addEventListener('mouseup', (e) => {
            if (this.isDrawing) {
                this.handleMouseUp(e);
            }
        });
    }

    async show(inputImageId, existingConfig = {}, nodeName = '') {
        // Set node name
        this.nameInput.value = nodeName;

        // Set initial crop values
        if (existingConfig.left !== undefined) {
            this.cropRect = {
                left: existingConfig.left,
                right: existingConfig.right,
                top: existingConfig.top,
                bottom: existingConfig.bottom
            };
        }

        // Load the input image
        if (inputImageId) {
            try {
                const response = await fetch(API_PATHS.images(inputImageId));
                if (!response.ok) throw new Error('Failed to load image');

                const blob = await response.blob();
                const imageUrl = URL.createObjectURL(blob);

                // Wait for image to load before showing modal
                await new Promise((resolve, reject) => {
                    this.image = new Image();
                    this.image.onload = () => {
                        URL.revokeObjectURL(imageUrl);
                        this.fitImageToCanvas();
                        this.updateFieldsFromRect();
                        this.render();
                        resolve();
                    };
                    this.image.onerror = () => {
                        URL.revokeObjectURL(imageUrl);
                        reject(new Error('Failed to load image'));
                    };
                    this.image.src = imageUrl;
                });
            } catch (error) {
                console.error('Failed to load image for crop:', error);
                alert('Failed to load input image. Make sure the crop node has an input connection.');
                return;
            }
        } else {
            alert('Cannot open crop editor: No input image available. Connect an input to this crop node first.');
            return;
        }

        this.modal.classList.add('active');
    }

    hide() {
        this.modal.classList.remove('active');
        this.image = null;
    }

    fitImageToCanvas() {
        if (!this.image) return;

        // Define max canvas dimensions
        const maxCanvasWidth = 900;
        const maxCanvasHeight = 700;

        const imageAspect = this.image.width / this.image.height;

        // Calculate canvas size to fit image aspect ratio
        let canvasWidth, canvasHeight;

        if (this.image.width > this.image.height) {
            // Landscape: constrain by width
            canvasWidth = Math.min(maxCanvasWidth, this.image.width);
            canvasHeight = canvasWidth / imageAspect;

            // If height exceeds max, recalculate
            if (canvasHeight > maxCanvasHeight) {
                canvasHeight = maxCanvasHeight;
                canvasWidth = canvasHeight * imageAspect;
            }
        } else {
            // Portrait: constrain by height
            canvasHeight = Math.min(maxCanvasHeight, this.image.height);
            canvasWidth = canvasHeight * imageAspect;

            // If width exceeds max, recalculate
            if (canvasWidth > maxCanvasWidth) {
                canvasWidth = maxCanvasWidth;
                canvasHeight = canvasWidth / imageAspect;
            }
        }

        // Set canvas dimensions
        this.canvas.width = canvasWidth;
        this.canvas.height = canvasHeight;

        // Calculate scale to fit image in canvas
        this.imageScale = canvasWidth / this.image.width;

        // Image fills entire canvas, no offset needed
        this.imageOffsetX = 0;
        this.imageOffsetY = 0;

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
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.5)';
        this.ctx.lineWidth = 5;
        this.ctx.strokeRect(rect.left, rect.top, rect.right - rect.left, rect.bottom - rect.top);
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

    handleMouseDown(e) {
        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        // Convert to image coordinates
        const imageCoords = this.canvasToImageCoords(x, y);

        // If shift is held, start dragging the existing crop region
        if (e.shiftKey) {
            this.isDragging = true;
            this.drawStartX = imageCoords.x;
            this.drawStartY = imageCoords.y;
            this.dragStartRect = { ...this.cropRect };
            this.canvas.style.cursor = 'move';
        } else {
            // Start drawing a new crop box
            this.isDrawing = true;
            this.drawStartX = imageCoords.x;
            this.drawStartY = imageCoords.y;

            // Initialize crop rect at the starting point
            this.cropRect = {
                left: Math.round(imageCoords.x),
                top: Math.round(imageCoords.y),
                right: Math.round(imageCoords.x),
                bottom: Math.round(imageCoords.y)
            };

            this.canvas.style.cursor = 'crosshair';
        }
    }

    handleMouseMove(e) {
        if (!this.isDrawing && !this.isDragging) {
            // Update cursor based on shift key
            this.canvas.style.cursor = e.shiftKey ? 'move' : 'crosshair';
            return;
        }

        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        // Convert to image coordinates
        const imageCoords = this.canvasToImageCoords(x, y);

        if (this.isDragging) {
            // Move the entire crop rectangle
            const dx = imageCoords.x - this.drawStartX;
            const dy = imageCoords.y - this.drawStartY;

            const width = this.dragStartRect.right - this.dragStartRect.left;
            const height = this.dragStartRect.bottom - this.dragStartRect.top;

            let newLeft = this.dragStartRect.left + dx;
            let newTop = this.dragStartRect.top + dy;

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
            // Drawing a new crop box
            // Clamp to image bounds
            const clampedX = Math.max(0, Math.min(this.image.width, imageCoords.x));
            const clampedY = Math.max(0, Math.min(this.image.height, imageCoords.y));

            // Update crop rectangle (handle dragging in any direction)
            const left = Math.min(this.drawStartX, clampedX);
            const right = Math.max(this.drawStartX, clampedX);
            const top = Math.min(this.drawStartY, clampedY);
            const bottom = Math.max(this.drawStartY, clampedY);

            this.cropRect = {
                left: Math.round(left),
                top: Math.round(top),
                right: Math.round(right),
                bottom: Math.round(bottom)
            };
        }

        this.updateFieldsFromRect();
        this.render();
    }

    handleMouseUp(e) {
        if (this.isDrawing) {
            this.isDrawing = false;
            this.canvas.style.cursor = e.shiftKey ? 'move' : 'crosshair';

            // Ensure minimum size (at least 10x10 pixels)
            const width = this.cropRect.right - this.cropRect.left;
            const height = this.cropRect.bottom - this.cropRect.top;

            if (width < 10 || height < 10) {
                // Reset to full image if crop is too small
                this.cropRect = {
                    left: 0,
                    top: 0,
                    right: this.image.width,
                    bottom: this.image.height
                };
                this.updateFieldsFromRect();
                this.render();
            }
        }

        if (this.isDragging) {
            this.isDragging = false;
            this.dragStartRect = null;
            this.canvas.style.cursor = e.shiftKey ? 'move' : 'crosshair';
        }
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

    handleWheel(e) {
        // Only zoom when shift is held
        if (!e.shiftKey || !this.image) return;

        e.preventDefault();

        // Calculate center of current crop rectangle
        const centerX = (this.cropRect.left + this.cropRect.right) / 2;
        const centerY = (this.cropRect.top + this.cropRect.bottom) / 2;

        // Determine zoom factor (scroll up = zoom in, scroll down = zoom out)
        const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9;

        // Calculate new dimensions
        const currentWidth = this.cropRect.right - this.cropRect.left;
        const currentHeight = this.cropRect.bottom - this.cropRect.top;

        const newWidth = currentWidth * zoomFactor;
        const newHeight = currentHeight * zoomFactor;

        // Calculate new bounds centered around the center point
        let newLeft = centerX - newWidth / 2;
        let newRight = centerX + newWidth / 2;
        let newTop = centerY - newHeight / 2;
        let newBottom = centerY + newHeight / 2;

        // Constrain to image bounds
        if (newLeft < 0) {
            newLeft = 0;
            newRight = newWidth;
        }
        if (newRight > this.image.width) {
            newRight = this.image.width;
            newLeft = this.image.width - newWidth;
        }
        if (newTop < 0) {
            newTop = 0;
            newBottom = newHeight;
        }
        if (newBottom > this.image.height) {
            newBottom = this.image.height;
            newTop = this.image.height - newHeight;
        }

        // Ensure we don't go beyond image bounds after constraining
        newLeft = Math.max(0, newLeft);
        newRight = Math.min(this.image.width, newRight);
        newTop = Math.max(0, newTop);
        newBottom = Math.min(this.image.height, newBottom);

        // Ensure minimum size (at least 10x10 pixels)
        const finalWidth = newRight - newLeft;
        const finalHeight = newBottom - newTop;

        if (finalWidth >= 10 && finalHeight >= 10) {
            this.cropRect = {
                left: Math.round(newLeft),
                top: Math.round(newTop),
                right: Math.round(newRight),
                bottom: Math.round(newBottom)
            };

            this.updateFieldsFromRect();
            this.render();
        }
    }

    handleSave() {
        if (this.onSave) {
            this.onSave({
                name: this.nameInput.value.trim(),
                config: {
                    left: this.cropRect.left,
                    right: this.cropRect.right,
                    top: this.cropRect.top,
                    bottom: this.cropRect.bottom
                }
            });
        }
        this.hide();
    }
}
