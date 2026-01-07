import { spawn } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

const root = path.join(process.cwd(), "dist/dev/mp-weixin");
fs.mkdirSync(root, { recursive: true });
const appWxss = path.join(root, "app.wxss");

function runNode(script, args = []) {
  return new Promise((resolve, reject) => {
    const child = spawn(process.execPath, [script, ...args], { stdio: "inherit" });
    child.on("exit", (code) => {
      if (code === 0) resolve();
      else reject(new Error(`${script} exited with code ${code}`));
    });
  });
}

function transformOne(filePath) {
  return new Promise((resolve) => {
    const child = spawn(process.execPath, ["./scripts/transform-wxss.mjs", filePath], {
      stdio: "inherit",
    });
    child.on("exit", () => resolve());
  });
}

function transformDir(dirPath) {
  return transformOne(dirPath);
}

function fileContainsNotSelector(filePath) {
  try {
    const s = fs.readFileSync(filePath, "utf8");
    return s.includes(":not(") || s.includes("> *") || s.includes("~ *");
  } catch {
    return false;
  }
}

async function bootstrap() {
  await runNode("./scripts/clean-dev-mp-weixin.mjs");

  // 启动 uni 的 dev 编译（常驻进程）
  const uni = spawn("uni", ["-p", "mp-weixin"], { stdio: "inherit" });

  const pending = new Set();
  let timer = null;

  async function flush() {
    timer = null;
    const files = Array.from(pending);
    pending.clear();
    for (const f of files) await transformOne(f);
  }

  function schedule(file) {
    pending.add(file);
    if (timer) return;
    timer = setTimeout(flush, 80);
  }

  function handleChange(filename) {
    if (!filename) return;
    if (!filename.endsWith(".wxss")) return;
    const full = path.join(root, filename);
    schedule(full);
  }

  async function waitForDir(dir, timeoutMs = 15000) {
    const startedAt = Date.now();
    while (Date.now() - startedAt < timeoutMs) {
      try {
        if (fs.statSync(dir).isDirectory()) return;
      } catch {}
      await new Promise((r) => setTimeout(r, 100));
    }
    throw new Error(`Output dir not created in time: ${dir}`);
  }

  async function waitForFile(filePath, timeoutMs = 15000) {
    const startedAt = Date.now();
    while (Date.now() - startedAt < timeoutMs) {
      try {
        if (fs.statSync(filePath).isFile()) return;
      } catch {}
      await new Promise((r) => setTimeout(r, 100));
    }
    throw new Error(`Output file not created in time: ${filePath}`);
  }

  // 初次启动：尽量把现有 wxss 都过一遍（避免微信开发者工具首次导入就报错）
  if (fs.existsSync(root)) {
    const stack = [root];
    while (stack.length) {
      const cur = stack.pop();
      if (!cur) continue;
      const st = fs.statSync(cur);
      if (st.isDirectory()) {
        for (const name of fs.readdirSync(cur)) stack.push(path.join(cur, name));
        continue;
      }
      if (st.isFile() && cur.endsWith(".wxss")) schedule(cur);
    }
  }

  // clean 后目录可能被删，且 uni 创建输出目录是异步的；这里等待目录可用再 watch
  await waitForDir(root);

  // 等待 app.wxss 产出后，先对整个输出目录做一次全量转换，避免首次导入微信开发者工具就报 :not() 语法错
  await waitForFile(appWxss);
  await transformDir(root);

  try {
    fs.watch(root, { recursive: true }, (_event, filename) => handleChange(filename));
  } catch {
    // Node 在部分平台/文件系统不支持 recursive；降级为只 watch 根目录
    fs.watch(root, (_event, filename) => handleChange(filename));
  }

  // 兜底：uni 可能在 watch 触发后立刻又重写 app.wxss，微信工具读到“中间态”就会报错
  // 这里做轻量轮询，只在检测到 :not() 时才触发一次 transform，降低读写竞态概率
  setInterval(async () => {
    if (!fileContainsNotSelector(appWxss)) return;
    await transformOne(appWxss);
  }, 200);

  uni.on("exit", (code) => process.exit(code ?? 0));
}

bootstrap().catch((err) => {
  console.error(err?.message || err);
  process.exit(1);
});
