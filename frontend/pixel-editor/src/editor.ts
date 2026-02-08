import {
  clonePixels,
  createEmptyPixels,
  inBounds,
  setPixelRGBA,
} from "./pixel-buffer.js";
import type {
  ActiveTool,
  CreateEditorOptions,
  Editor,
  EditorConfig,
  EditorHooks,
  EditorInitialState,
  EditorState,
  EraserSelection,
  ImageSnapshot,
  InitialImage,
  PaintbrushSelection,
  PersistPayload,
  PersistReason,
  RGBA,
  Tool,
} from "./types.js";

const TRANSPARENT: RGBA = [0, 0, 0, 0];
const DEFAULT_COLOR: RGBA = [0, 0, 0, 255];

function normalizeTool(tool: unknown): Tool {
  if (tool === "paintbrush" || tool === "eraser") {
    return tool;
  }
  if (tool === "pencil") {
    return "paintbrush";
  }
  throw new Error(`unsupported tool: ${String(tool)}`);
}

function normalizeConfig(config: EditorConfig): EditorConfig {
  return {
    width: config.width,
    height: config.height,
    zoom: config.zoom,
  };
}

function validateConfig(config: EditorConfig): void {
  if (!Number.isInteger(config.width) || config.width <= 0) {
    throw new Error("config.width must be a positive integer");
  }
  if (!Number.isInteger(config.height) || config.height <= 0) {
    throw new Error("config.height must be a positive integer");
  }
  if (!Number.isFinite(config.zoom) || config.zoom <= 0) {
    throw new Error("config.zoom must be a positive number");
  }
}

function validateBrushSize(size: number, fieldName: string): void {
  if (!Number.isInteger(size) || size <= 0) {
    throw new Error(`${fieldName} must be a positive integer`);
  }
}

function cloneState(state: EditorState): EditorState {
  return {
    activeTool: state.activeTool,
    color: [...state.color] as RGBA,
    paintbrush: {
      size: state.paintbrush.size,
    },
    eraser: {
      size: state.eraser.size,
    },
  };
}

function normalizeState(initial: EditorInitialState): EditorState {
  const legacy = initial.config as EditorConfig & {
    tool?: unknown;
    color?: unknown;
    brushSize?: unknown;
  };
  const state = initial.state;

  const paintbrushSize = state?.paintbrush?.size ?? (typeof legacy.brushSize === "number" ? legacy.brushSize : 1);
  const eraserSize = state?.eraser?.size ?? (typeof legacy.brushSize === "number" ? legacy.brushSize : 1);
  validateBrushSize(paintbrushSize, "state.paintbrush.size");
  validateBrushSize(eraserSize, "state.eraser.size");

  const color = (state?.color ?? legacy.color ?? DEFAULT_COLOR) as RGBA;
  const activeTool = normalizeTool(state?.activeTool ?? legacy.tool ?? "paintbrush");

  return {
    activeTool,
    color: [...color] as RGBA,
    paintbrush: {
      size: paintbrushSize,
    },
    eraser: {
      size: eraserSize,
    },
  };
}

function cloneConfig(config: EditorConfig): EditorConfig {
  return {
    ...config,
  };
}

function normalizeImage(image: InitialImage | undefined, config: EditorConfig): ImageSnapshot {
  if (!image || image.kind === "empty") {
    return {
      width: config.width,
      height: config.height,
      pixels: createEmptyPixels(config.width, config.height),
    };
  }

  if (image.kind === "imageData") {
    return {
      width: image.data.width,
      height: image.data.height,
      pixels: new Uint8ClampedArray(image.data.data),
    };
  }

  const expectedLength = image.width * image.height * 4;
  if (image.pixels.length !== expectedLength) {
    throw new Error("initial rgba pixels length does not match width*height*4");
  }

  return {
    width: image.width,
    height: image.height,
    pixels: new Uint8ClampedArray(image.pixels),
  };
}

function snapshot(config: EditorConfig, state: EditorState, image: ImageSnapshot): PersistPayload {
  return {
    config: cloneConfig(config),
    state: cloneState(state),
    image: {
      width: image.width,
      height: image.height,
      pixels: clonePixels(image.pixels),
    },
    reason: "manual",
  };
}

function brushRadius(size: number): number {
  return Math.max(0, Math.floor((size - 1) / 2));
}

function stampCircle(
  width: number,
  height: number,
  pixels: Uint8ClampedArray,
  cx: number,
  cy: number,
  size: number,
  rgba: RGBA,
): boolean {
  const radius = brushRadius(size);
  let changed = false;

  for (let dy = -radius; dy <= radius; dy += 1) {
    for (let dx = -radius; dx <= radius; dx += 1) {
      if (dx * dx + dy * dy > radius * radius) {
        continue;
      }
      const x = cx + dx;
      const y = cy + dy;
      if (!inBounds(width, height, x, y)) {
        continue;
      }
      setPixelRGBA(width, height, pixels, x, y, rgba);
      changed = true;
    }
  }

  return changed;
}

function applyLine(fromX: number, fromY: number, toX: number, toY: number, visit: (x: number, y: number) => void): void {
  const dx = Math.abs(toX - fromX);
  const dy = Math.abs(toY - fromY);
  const sx = fromX < toX ? 1 : -1;
  const sy = fromY < toY ? 1 : -1;

  let x = fromX;
  let y = fromY;
  let err = dx - dy;

  for (;;) {
    visit(x, y);
    if (x === toX && y === toY) {
      break;
    }
    const e2 = 2 * err;
    if (e2 > -dy) {
      err -= dy;
      x += sx;
    }
    if (e2 < dx) {
      err += dx;
      y += sy;
    }
  }
}

class PixelEditor implements Editor {
  private config: EditorConfig;
  private state: EditorState;
  private image: ImageSnapshot;
  private hooks: EditorHooks;
  private dirty = false;

  constructor(state: EditorInitialState, hooks: EditorHooks) {
    const normalizedConfig = normalizeConfig(state.config);
    validateConfig(normalizedConfig);
    const normalizedState = normalizeState(state);

    this.config = cloneConfig(normalizedConfig);
    this.state = cloneState(normalizedState);
    this.image = normalizeImage(state.image, this.config);
    this.hooks = hooks;
  }

  getConfig(): EditorConfig {
    return cloneConfig(this.config);
  }

  getState(): EditorState {
    return cloneState(this.state);
  }

  setColor(color: RGBA): void {
    const nextColor = [...color] as RGBA;
    const isSame =
      this.state.color[0] === nextColor[0] &&
      this.state.color[1] === nextColor[1] &&
      this.state.color[2] === nextColor[2] &&
      this.state.color[3] === nextColor[3];
    if (isSame) {
      return;
    }

    this.state = {
      ...this.state,
      color: nextColor,
    };
    this.dirty = true;
  }

  getColor(): RGBA {
    return [...this.state.color] as RGBA;
  }

  selectPaintbrush(selection: PaintbrushSelection = {}): void {
    const nextSize = selection.size ?? this.state.paintbrush.size;
    validateBrushSize(nextSize, "paintbrush size");
    const changed = this.state.activeTool !== "paintbrush" || this.state.paintbrush.size !== nextSize;

    this.state = {
      ...this.state,
      activeTool: "paintbrush",
      paintbrush: {
        ...this.state.paintbrush,
        size: nextSize,
      },
    };
    if (changed) {
      this.dirty = true;
    }
  }

  selectEraser(selection: EraserSelection = {}): void {
    const nextSize = selection.size ?? this.state.eraser.size;
    validateBrushSize(nextSize, "eraser size");
    const changed = this.state.activeTool !== "eraser" || this.state.eraser.size !== nextSize;

    this.state = {
      ...this.state,
      activeTool: "eraser",
      eraser: {
        size: nextSize,
      },
    };
    if (changed) {
      this.dirty = true;
    }
  }

  getActiveTool(): ActiveTool {
    if (this.state.activeTool === "eraser") {
      return { kind: "eraser", size: this.state.eraser.size };
    }
    return {
      kind: "paintbrush",
      size: this.state.paintbrush.size,
    };
  }

  getImage(): ImageSnapshot {
    return {
      width: this.image.width,
      height: this.image.height,
      pixels: clonePixels(this.image.pixels),
    };
  }

  private applyAt(x: number, y: number): boolean {
    const tool = normalizeTool(this.state.activeTool);
    const size = tool === "eraser" ? this.state.eraser.size : this.state.paintbrush.size;
    const rgba = tool === "eraser" ? TRANSPARENT : this.state.color;
    return stampCircle(this.image.width, this.image.height, this.image.pixels, x, y, size, rgba);
  }

  dab(x: number, y: number): void {
    const changed = this.applyAt(x, y);
    if (changed) {
      this.dirty = true;
    }
  }

  stroke(fromX: number, fromY: number, toX: number, toY: number): void {
    let changed = false;
    applyLine(fromX, fromY, toX, toY, (x, y) => {
      if (this.applyAt(x, y)) {
        changed = true;
      }
    });
    if (changed) {
      this.dirty = true;
    }
  }

  clear(color: RGBA = TRANSPARENT): void {
    let changed = false;
    for (let y = 0; y < this.image.height; y += 1) {
      for (let x = 0; x < this.image.width; x += 1) {
        setPixelRGBA(this.image.width, this.image.height, this.image.pixels, x, y, color);
        changed = true;
      }
    }
    if (changed) {
      this.dirty = true;
    }
  }

  isDirty(): boolean {
    return this.dirty;
  }

  async save(reason: PersistReason = "manual"): Promise<void> {
    const save = this.hooks.save;
    if (!save) {
      throw new Error("save hook is not configured");
    }

    const payload = snapshot(this.config, this.state, this.image);
    payload.reason = reason;

    try {
      await save(payload);
      this.dirty = false;
    } catch (error) {
      this.hooks.onError?.(error);
      throw error;
    }
  }
}

export async function createEditor(options: CreateEditorOptions = {}): Promise<Editor> {
  const hooks = options.hooks ?? {};

  try {
    const loaded = await hooks.load?.();
    const initial = loaded ?? options.initial;

    if (!initial) {
      throw new Error("initial state is required when load hook returns null");
    }

    return new PixelEditor(initial, hooks);
  } catch (error) {
    hooks.onError?.(error);
    throw error;
  }
}
