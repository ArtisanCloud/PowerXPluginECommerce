import fs from "node:fs";
import path from "node:path";

const target = path.resolve(process.cwd(), "dist/build/mp-weixin");
try {
  fs.rmSync(target, { recursive: true, force: true });
} catch {}

