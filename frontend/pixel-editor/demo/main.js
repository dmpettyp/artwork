import { createEditor } from "../dist/index.js";

const STORAGE_KEY = "pixel-editor-demo-state-v1";
const WIDTH = 32;
const HEIGHT = 32;
const ZOOM = 16;

const canvas = document.querySelector("#canvas");
const colorInput = document.querySelector("#color");
const eraseBtn = document.querySelector("#erase");
const clearBtn = document.querySelector("#clear");
const saveBtn = document.querySelector("#save");
const downloadBtn = document.querySelector("#download");
const statusEl = document.querySelector("#status");

const ctx = canvas.getContext("2d");
if (!ctx) {
  throw new Error("2d canvas context is unavailable");
}

canvas.width = WIDTH;
canvas.height = HEIGHT;
canvas.style.width = `${WIDTH * ZOOM}px`;
canvas.style.height = `${HEIGHT * ZOOM}px`;
ctx.imageSmoothingEnabled = false;

let lastPointer = null;

function setStatus(text) {
  statusEl.textContent = text;
}

function hexToRgba(hex) {
  const v = hex.replace("#", "");
  const r = Number.parseInt(v.slice(0, 2), 16);
  const g = Number.parseInt(v.slice(2, 4), 16);
  const b = Number.parseInt(v.slice(4, 6), 16);
  return [r, g, b, 255];
}

function byteToHex(value) {
  return value.toString(16).padStart(2, "0");
}

function rgbaToHex(rgba) {
  return `#${byteToHex(rgba[0])}${byteToHex(rgba[1])}${byteToHex(rgba[2])}`;
}

function serializePayload(payload) {
  return JSON.stringify({
    config: payload.config,
    state: payload.state,
    image: {
      width: payload.image.width,
      height: payload.image.height,
      pixels: Array.from(payload.image.pixels),
    },
    reason: payload.reason,
  });
}

function deserializeInitialState(raw) {
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw);
    if (!parsed?.config || !parsed?.image?.pixels) {
      return null;
    }

    const legacyTool = parsed.config?.tool;
    const legacyColor = parsed.config?.color;
    const legacyBrushSize = parsed.config?.brushSize;

    const state = parsed.state ?? {
      activeTool: legacyTool === "eraser" ? "eraser" : "paintbrush",
      color: Array.isArray(legacyColor) ? legacyColor : [0, 0, 0, 255],
      paintbrush: {
        size: Number.isInteger(legacyBrushSize) ? legacyBrushSize : 1,
      },
      eraser: {
        size: Number.isInteger(legacyBrushSize) ? legacyBrushSize : 1,
      },
    };

    return {
      config: parsed.config,
      state,
      image: {
        kind: "rgba",
        width: parsed.image.width,
        height: parsed.image.height,
        pixels: new Uint8ClampedArray(parsed.image.pixels),
      },
    };
  } catch {
    return null;
  }
}

function render(editor) {
  const image = editor.getImage();
  const imageData = new ImageData(new Uint8ClampedArray(image.pixels), image.width, image.height);
  ctx.putImageData(imageData, 0, 0);
  setStatus(editor.isDirty() ? "Unsaved changes" : "Saved");
}

function canvasPixelFromPointer(event) {
  const rect = canvas.getBoundingClientRect();
  const x = Math.floor(((event.clientX - rect.left) / rect.width) * WIDTH);
  const y = Math.floor(((event.clientY - rect.top) / rect.height) * HEIGHT);
  return [x, y];
}

async function main() {
  const saved = deserializeInitialState(localStorage.getItem(STORAGE_KEY));
  const editor = await createEditor({
    initial: {
      config: {
        width: WIDTH,
        height: HEIGHT,
        zoom: ZOOM,
      },
      state: {
        activeTool: "paintbrush",
        color: hexToRgba(colorInput.value),
        paintbrush: {
          size: 1,
        },
        eraser: {
          size: 1,
        },
      },
      image: { kind: "empty" },
    },
    hooks: {
      load: () => saved,
      save: (payload) => {
        localStorage.setItem(STORAGE_KEY, serializePayload(payload));
      },
      onError: (error) => {
        console.error(error);
        setStatus(`Error: ${String(error)}`);
      },
    },
  });

  function syncToolUi() {
    const active = editor.getActiveTool();
    eraseBtn.textContent = `Eraser: ${active.kind === "eraser" ? "On" : "Off"}`;
    colorInput.value = rgbaToHex(editor.getColor());
  }

  syncToolUi();
  render(editor);

  let pointerDown = false;

  function paint(event) {
    const [x, y] = canvasPixelFromPointer(event);

    if (lastPointer) {
      editor.stroke(lastPointer[0], lastPointer[1], x, y);
    } else {
      editor.dab(x, y);
    }
    lastPointer = [x, y];
    render(editor);
  }

  canvas.addEventListener("pointerdown", (event) => {
    pointerDown = true;
    canvas.setPointerCapture(event.pointerId);
    lastPointer = null;
    paint(event);
  });

  canvas.addEventListener("pointermove", (event) => {
    if (!pointerDown) {
      return;
    }
    paint(event);
  });

  canvas.addEventListener("pointerup", () => {
    pointerDown = false;
    lastPointer = null;
  });

  canvas.addEventListener("pointercancel", () => {
    pointerDown = false;
    lastPointer = null;
  });

  eraseBtn.addEventListener("click", () => {
    const active = editor.getActiveTool();
    if (active.kind === "eraser") {
      editor.selectPaintbrush();
    } else {
      editor.selectEraser();
    }
    syncToolUi();
  });

  colorInput.addEventListener("input", () => {
    editor.setColor(hexToRgba(colorInput.value));
    syncToolUi();
  });

  clearBtn.addEventListener("click", () => {
    editor.clear();
    render(editor);
  });

  saveBtn.addEventListener("click", async () => {
    await editor.save("manual");
    render(editor);
  });

  downloadBtn.addEventListener("click", () => {
    canvas.toBlob((blob) => {
      if (!blob) {
        setStatus("Download failed");
        return;
      }
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "pixel-editor-demo.png";
      a.click();
      URL.revokeObjectURL(url);
      setStatus("Downloaded PNG");
    }, "image/png");
  });
}

main().catch((error) => {
  console.error(error);
  setStatus(`Failed to start: ${String(error)}`);
});
