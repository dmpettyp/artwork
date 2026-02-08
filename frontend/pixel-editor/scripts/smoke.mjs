import { mkdir, writeFile } from "node:fs/promises";
import { deflateSync } from "node:zlib";
import { createEditor } from "../dist/index.js";

function crc32(bytes) {
  let crc = 0xffffffff;
  for (const b of bytes) {
    crc ^= b;
    for (let i = 0; i < 8; i += 1) {
      const mask = -(crc & 1);
      crc = (crc >>> 1) ^ (0xedb88320 & mask);
    }
  }
  return (crc ^ 0xffffffff) >>> 0;
}

function u32be(value) {
  const buf = Buffer.alloc(4);
  buf.writeUInt32BE(value >>> 0, 0);
  return buf;
}

function pngChunk(type, data) {
  const typeBytes = Buffer.from(type, "ascii");
  const crc = crc32(Buffer.concat([typeBytes, data]));
  return Buffer.concat([u32be(data.length), typeBytes, data, u32be(crc)]);
}

function toPng(image) {
  const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(image.width, 0);
  ihdr.writeUInt32BE(image.height, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // color type RGBA
  ihdr[10] = 0; // compression method
  ihdr[11] = 0; // filter method
  ihdr[12] = 0; // interlace

  const stride = image.width * 4;
  const raw = Buffer.alloc((stride + 1) * image.height);
  for (let y = 0; y < image.height; y += 1) {
    const rowStart = y * (stride + 1);
    raw[rowStart] = 0; // no filter
    for (let x = 0; x < image.width; x += 1) {
      const src = (y * image.width + x) * 4;
      const dst = rowStart + 1 + x * 4;
      raw[dst] = image.pixels[src] ?? 0;
      raw[dst + 1] = image.pixels[src + 1] ?? 0;
      raw[dst + 2] = image.pixels[src + 2] ?? 0;
      raw[dst + 3] = image.pixels[src + 3] ?? 0;
    }
  }

  const idat = deflateSync(raw);
  return Buffer.concat([
    signature,
    pngChunk("IHDR", ihdr),
    pngChunk("IDAT", idat),
    pngChunk("IEND", Buffer.alloc(0)),
  ]);
}

function drawCross(editor, width, height) {
  const n = Math.min(width, height);
  editor.selectPaintbrush({ size: 1 });
  editor.setColor([255, 64, 64, 255]);
  for (let i = 0; i < n; i += 1) {
    editor.dab(i, i);
  }

  editor.selectPaintbrush({ size: 1 });
  editor.setColor([64, 128, 255, 255]);
  for (let i = 0; i < n; i += 1) {
    editor.dab(width - 1 - i, i);
  }
}

async function main() {
  const width = 16;
  const height = 16;
  const outputDir = new URL("../tmp/", import.meta.url);
  const outputFile = new URL("smoke.png", outputDir);

  const editor = await createEditor({
    initial: {
      config: {
        width,
        height,
        zoom: 16,
      },
      state: {
        activeTool: "paintbrush",
        color: [0, 0, 0, 255],
        paintbrush: {
          size: 1,
        },
        eraser: {
          size: 1,
        },
      },
      image: { kind: "empty" },
    },
  });

  drawCross(editor, width, height);
  const image = editor.getImage();
  const png = toPng(image);

  await mkdir(outputDir, { recursive: true });
  await writeFile(outputFile, png);

  console.log(`Smoke test wrote ${new URL(outputFile).pathname}`);
  console.log(`Dirty before save: ${editor.isDirty()}`);
  await editor.save().catch((error) => {
    console.log(`Expected save failure without hook: ${String(error)}`);
  });
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
