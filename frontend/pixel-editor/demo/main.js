import { attachEditorCanvas, createEditor } from "../dist/index.js";

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

  const canvasAdapter = attachEditorCanvas({
    editor,
    canvas,
    onDraw: () => {
      setStatus(editor.isDirty() ? "Unsaved changes" : "Saved");
    },
  });

  function render() {
    canvasAdapter.render();
    setStatus(editor.isDirty() ? "Unsaved changes" : "Saved");
  }

  function syncToolUi() {
    const active = editor.getActiveTool();
    eraseBtn.textContent = `Eraser: ${active.kind === "eraser" ? "On" : "Off"}`;
    colorInput.value = rgbaToHex(editor.getColor());
  }

  syncToolUi();
  render();

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
    render();
  });

  saveBtn.addEventListener("click", async () => {
    await editor.save("manual");
    render();
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

  window.addEventListener("beforeunload", () => {
    canvasAdapter.destroy();
  });
}

main().catch((error) => {
  console.error(error);
  setStatus(`Failed to start: ${String(error)}`);
});
