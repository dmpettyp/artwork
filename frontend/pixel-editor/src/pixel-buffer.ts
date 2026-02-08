import type { RGBA } from "./types.js";

function clampByte(value: number): number {
  if (Number.isNaN(value)) {
    return 0;
  }
  return Math.max(0, Math.min(255, Math.round(value)));
}

function pixelOffset(width: number, x: number, y: number): number {
  return (y * width + x) * 4;
}

export function createEmptyPixels(width: number, height: number): Uint8ClampedArray {
  return new Uint8ClampedArray(width * height * 4);
}

export function clonePixels(pixels: Uint8ClampedArray): Uint8ClampedArray {
  return new Uint8ClampedArray(pixels);
}

export function inBounds(width: number, height: number, x: number, y: number): boolean {
  return x >= 0 && y >= 0 && x < width && y < height;
}

export function setPixelRGBA(
  width: number,
  height: number,
  pixels: Uint8ClampedArray,
  x: number,
  y: number,
  rgba: RGBA,
): void {
  if (!inBounds(width, height, x, y)) {
    return;
  }
  const i = pixelOffset(width, x, y);
  pixels[i] = clampByte(rgba[0]);
  pixels[i + 1] = clampByte(rgba[1]);
  pixels[i + 2] = clampByte(rgba[2]);
  pixels[i + 3] = clampByte(rgba[3]);
}

export function getPixelRGBA(
  width: number,
  height: number,
  pixels: Uint8ClampedArray,
  x: number,
  y: number,
): RGBA {
  if (!inBounds(width, height, x, y)) {
    return [0, 0, 0, 0];
  }
  const i = pixelOffset(width, x, y);
  return [pixels[i] ?? 0, pixels[i + 1] ?? 0, pixels[i + 2] ?? 0, pixels[i + 3] ?? 0];
}
