import type { Editor, ImageSnapshot } from "./types.js";

export type CanvasEditorAdapterOptions = {
  editor: Editor;
  canvas: HTMLCanvasElement;
  onDraw?: () => void;
};

export type CanvasEditorAdapter = {
  render(): void;
  destroy(): void;
};

function drawImageToContext(ctx: CanvasRenderingContext2D, image: ImageSnapshot): void {
  const imageData = new ImageData(new Uint8ClampedArray(image.pixels), image.width, image.height);
  ctx.putImageData(imageData, 0, 0);
}

function canvasPixelFromPointer(
  event: PointerEvent,
  canvas: HTMLCanvasElement,
  width: number,
  height: number,
): [number, number] {
  const rect = canvas.getBoundingClientRect();
  const x = Math.floor(((event.clientX - rect.left) / rect.width) * width);
  const y = Math.floor(((event.clientY - rect.top) / rect.height) * height);
  return [x, y];
}

export function attachEditorCanvas(options: CanvasEditorAdapterOptions): CanvasEditorAdapter {
  const { editor, canvas, onDraw } = options;
  const context = canvas.getContext("2d");
  if (!context) {
    throw new Error("2d canvas context is unavailable");
  }
  const ctx: CanvasRenderingContext2D = context;

  const config = editor.getConfig();
  canvas.width = config.width;
  canvas.height = config.height;
  canvas.style.width = `${config.width * config.zoom}px`;
  canvas.style.height = `${config.height * config.zoom}px`;
  ctx.imageSmoothingEnabled = false;

  let pointerDown = false;
  let lastPointer: [number, number] | null = null;

  function render(): void {
    drawImageToContext(ctx, editor.getImage());
  }

  function paint(event: PointerEvent): void {
    const [x, y] = canvasPixelFromPointer(event, canvas, config.width, config.height);

    if (lastPointer) {
      editor.stroke(lastPointer[0], lastPointer[1], x, y);
    } else {
      editor.dab(x, y);
    }

    lastPointer = [x, y];
    render();
    onDraw?.();
  }

  function onPointerDown(event: PointerEvent): void {
    pointerDown = true;
    canvas.setPointerCapture(event.pointerId);
    lastPointer = null;
    paint(event);
  }

  function onPointerMove(event: PointerEvent): void {
    if (!pointerDown) {
      return;
    }
    paint(event);
  }

  function onPointerUp(): void {
    pointerDown = false;
    lastPointer = null;
  }

  canvas.addEventListener("pointerdown", onPointerDown);
  canvas.addEventListener("pointermove", onPointerMove);
  canvas.addEventListener("pointerup", onPointerUp);
  canvas.addEventListener("pointercancel", onPointerUp);

  return {
    render,
    destroy(): void {
      canvas.removeEventListener("pointerdown", onPointerDown);
      canvas.removeEventListener("pointermove", onPointerMove);
      canvas.removeEventListener("pointerup", onPointerUp);
      canvas.removeEventListener("pointercancel", onPointerUp);
    },
  };
}
