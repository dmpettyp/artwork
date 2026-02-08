import { createReadStream } from "node:fs";
import { stat } from "node:fs/promises";
import http from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, "..");
const port = Number.parseInt(process.env.PORT ?? "4173", 10);

const contentType = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".map": "application/json; charset=utf-8",
  ".png": "image/png",
};

function sendNotFound(res) {
  res.statusCode = 404;
  res.setHeader("content-type", "text/plain; charset=utf-8");
  res.end("Not found");
}

function safePath(urlPath) {
  const cleaned = path.normalize(urlPath).replace(/^(\.\.(\/|\\|$))+/, "");
  return path.join(root, cleaned);
}

const server = http.createServer(async (req, res) => {
  const reqUrl = new URL(req.url ?? "/", `http://${req.headers.host ?? "localhost"}`);
  let reqPath = reqUrl.pathname;

  if (reqPath === "/") {
    reqPath = "/demo/";
  }
  if (reqPath === "/demo") {
    reqPath = "/demo/";
  }
  if (reqPath.endsWith("/")) {
    reqPath = `${reqPath}index.html`;
  }

  const diskPath = safePath(reqPath);

  try {
    const info = await stat(diskPath);
    if (!info.isFile()) {
      sendNotFound(res);
      return;
    }

    const ext = path.extname(diskPath).toLowerCase();
    res.statusCode = 200;
    res.setHeader("content-type", contentType[ext] ?? "application/octet-stream");
    createReadStream(diskPath).pipe(res);
  } catch {
    sendNotFound(res);
  }
});

server.listen(port, () => {
  console.log(`Pixel editor demo: http://localhost:${port}/demo/`);
});

server.on("error", (error) => {
  console.error(`Failed to start demo server on port ${port}: ${String(error)}`);
  process.exitCode = 1;
});
