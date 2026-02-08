export type RGBA = readonly [number, number, number, number];

export type Tool = "paintbrush" | "eraser";

export type EditorConfig = {
  width: number;
  height: number;
  zoom: number;
};

export type PaintbrushState = {
  size: number;
};

export type EraserState = {
  size: number;
};

export type EditorState = {
  activeTool: Tool;
  color: RGBA;
  paintbrush: PaintbrushState;
  eraser: EraserState;
};

export type ImageSnapshot = {
  width: number;
  height: number;
  pixels: Uint8ClampedArray;
};

export type InitialImage =
  | { kind: "empty" }
  | { kind: "imageData"; data: ImageData }
  | { kind: "rgba"; width: number; height: number; pixels: Uint8ClampedArray };

export type EditorInitialState = {
  config: EditorConfig;
  state?: Partial<EditorState>;
  image?: InitialImage;
};

export type PersistReason = "manual" | "checkpoint";

export type PersistPayload = {
  config: EditorConfig;
  state: EditorState;
  image: ImageSnapshot;
  reason: PersistReason;
};

export type EditorHooks = {
  load?: () => Promise<EditorInitialState | null> | EditorInitialState | null;
  save?: (payload: PersistPayload) => Promise<void> | void;
  onError?: (error: unknown) => void;
};

export type CreateEditorOptions = {
  initial?: EditorInitialState;
  hooks?: EditorHooks;
};

export type PaintbrushSelection = {
  size?: number;
};

export type EraserSelection = {
  size?: number;
};

export type ActiveTool =
  | { kind: "paintbrush"; size: number }
  | { kind: "eraser"; size: number };

export interface Editor {
  getConfig(): EditorConfig;
  getState(): EditorState;
  setColor(color: RGBA): void;
  getColor(): RGBA;
  getImage(): ImageSnapshot;
  selectPaintbrush(selection?: PaintbrushSelection): void;
  selectEraser(selection?: EraserSelection): void;
  getActiveTool(): ActiveTool;
  dab(x: number, y: number): void;
  stroke(fromX: number, fromY: number, toX: number, toY: number): void;
  clear(color?: RGBA): void;
  save(reason?: PersistReason): Promise<void>;
  isDirty(): boolean;
}
