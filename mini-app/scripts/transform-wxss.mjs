import fs from "node:fs";
import path from "node:path";

const target = process.argv[2];
if (!target) {
  console.error("Usage: node scripts/transform-wxss.mjs <wxss-file-or-dir>");
  process.exit(2);
}

function isDir(p) {
  try {
    return fs.statSync(p).isDirectory();
  } catch {
    return false;
  }
}

function listWxssFiles(root) {
  const results = [];
  const stack = [root];
  while (stack.length) {
    const current = stack.pop();
    if (!current) continue;
    const st = fs.statSync(current);
    if (st.isDirectory()) {
      const items = fs.readdirSync(current).map((f) => path.join(current, f));
      stack.push(...items);
      continue;
    }
    if (st.isFile() && current.endsWith(".wxss")) results.push(current);
  }
  return results;
}

function transformWxss(content) {
  let s = content;

  // 0) 选择器兼容：WXSS 对 :not() / 通配选择器 * / 一般兄弟选择器 ~ 支持不完整
  //    Tailwind 的 space/divide 规则常见形态：
  //      .space-y-5 > :not([hidden]) ~ :not([hidden]) { ... }
  //    这里改写成更保守的 `view + view`（同类兄弟，跳过第一个元素），确保可编译运行。
  //    代价：仅对 view 子节点生效（uni-app 常见布局默认就是 view）。
  s = s.replace(
    /(\.space-[xy]-[^>{\s]+)\s*>\s*:not\(\[hidden\]\)\s*~\s*:not\(\[hidden\]\)\s*\{/g,
    "$1 > view + view {",
  );
  s = s.replace(
    /(\.divide-[xy]-[^>{\s]+)\s*>\s*:not\(\[hidden\]\)\s*~\s*:not\(\[hidden\]\)\s*\{/g,
    "$1 > view + view {",
  );

  // 0.1) 兜底：如果还有残余 :not([hidden])，直接移除 :not(...) 本身，避免冒号导致编译失败
  s = s.replace(/:not\(\[hidden\]\)/g, "");
  // 0.2) 兜底：避免出现 `> * ~ *` 这类 WXSS 不接受的组合
  s = s.replace(/\s*>\s*\*\s*~\s*\*\s*\{/g, " > view + view {");

  // 1) shadow：把 Tailwind 的 var 组合换成静态 box-shadow
  s = s.replace(
    /--tw-shadow:\s*([^;]+);--tw-shadow-colored:[^;]+;box-shadow:var\(--tw-ring-offset-shadow,[^)]*\),var\(--tw-ring-shadow,[^)]*\),var\(--tw-shadow\)/g,
    "box-shadow:$1",
  );

  // 2) transform/filter：WXSS 不支持 var()，降级为 none
  s = s.replace(/\.transform\{[^}]*?\}/g, ".transform{transform:none}");
  s = s.replace(/\.filter\{[^}]*?\}/g, ".filter{filter:none}");

  // 3) 颜色：把 rgb(R G B / A) 转成 rgba(R,G,B,A)，避免空格分隔与 var()
  s = s.replace(
    /rgb\(\s*(\d+)\s+(\d+)\s+(\d+)\s*\/\s*([0-9.]+)\s*\)/g,
    "rgba($1,$2,$3,$4)",
  );
  s = s.replace(
    /rgb\(\s*(\d+)\s+(\d+)\s+(\d+)\s*\/\s*var\(--tw-[^,]+,\s*([^)]+)\)\s*\)/g,
    "rgba($1,$2,$3,$4)",
  );
  s = s.replace(
    /rgb\(\s*(\d+)\s+(\d+)\s+(\d+)\s*\/\s*var\(--tw-[^)]+\)\s*\)/g,
    "rgba($1,$2,$3,1)",
  );
  s = s.replace(/rgb\(\s*(\d+)\s+(\d+)\s+(\d+)\s*\)/g, "rgb($1,$2,$3)");

  // 4) 移除 --tw-* 自定义属性声明
  s = s.replace(/--tw-[a-zA-Z0-9-]+\s*:\s*[^;{}]+;?/g, "");

  // 5) 清理残余 var(--tw-*) 引用
  s = s.replace(/var\(--tw-[^,)\s]+,\s*([^)]+)\)/g, "$1");
  s = s.replace(/var\(--tw-[^)]+\)/g, "0");

  // 6) 清理多余分号
  s = s.replace(/;+/g, ";");
  s = s.replace(/\{;/g, "{");
  s = s.replace(/;\}/g, "}");

  return s;
}

function verifyNoUnsupported(content, file) {
  const needles = ["--tw-", "var(--tw-"];
  const hit = needles.find((n) => content.includes(n));
  if (hit) {
    console.error(`WXSS still contains unsupported Tailwind var: ${hit} in ${file}`);
    process.exit(1);
  }
}

const files = isDir(target) ? listWxssFiles(target) : [target];
files.forEach((file) => {
  let content = "";
  try {
    content = fs.readFileSync(file, "utf8");
  } catch (err) {
    console.error(`Failed to read ${file}:`, err?.message || err);
    process.exit(2);
  }
  const out = transformWxss(content);
  verifyNoUnsupported(out, file);
  fs.writeFileSync(file, out, "utf8");
});
