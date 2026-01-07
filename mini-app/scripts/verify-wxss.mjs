import fs from "node:fs";

const file = process.argv[2];
if (!file) {
  console.error("Usage: node scripts/verify-wxss.mjs <wxss-file>");
  process.exit(2);
}

let content = "";
try {
  content = fs.readFileSync(file, "utf8");
} catch (err) {
  console.error(`Failed to read ${file}:`, err?.message || err);
  process.exit(2);
}

// 微信 WXSS 对选择器里的反斜杠（如 .text-\\[10px\\]）会直接报错
if (content.includes("\\[" ) || content.includes("\\]") || content.includes("\\\\")) {
  console.error(`WXSS contains unsupported backslash selector escaping: ${file}`);
  // 输出前几处位置，方便定位
  const needle = "\\";
  let idx = 0;
  let count = 0;
  while ((idx = content.indexOf(needle, idx)) !== -1 && count < 5) {
    const start = Math.max(0, idx - 20);
    const end = Math.min(content.length, idx + 40);
    console.error(`...${content.slice(start, end).replace(/\n/g, "\\n")}...`);
    idx += 1;
    count += 1;
  }
  process.exit(1);
}

// 微信 WXSS 不支持 CSS 变量（Tailwind v3 默认会输出 --tw-* / var(--tw-*)）
if (content.includes("--tw-") || content.includes("var(--tw-")) {
  console.error(`WXSS contains unsupported CSS variables (Tailwind --tw-*): ${file}`);
  process.exit(1);
}

// 微信 WXSS 对 :not() 兼容性较差（不同版本会直接编译失败），这里直接禁止
if (content.includes(":not(")) {
  console.error(`WXSS contains unsupported selector :not(): ${file}`);
  process.exit(1);
}

// 微信 WXSS 对通配选择器 / 一般兄弟选择器在部分场景会编译失败（如 `> * ~ *`）
if (content.includes("> *") || content.includes("~ *")) {
  console.error(`WXSS contains potentially unsupported selectors (universal/sibling): ${file}`);
  process.exit(1);
}
