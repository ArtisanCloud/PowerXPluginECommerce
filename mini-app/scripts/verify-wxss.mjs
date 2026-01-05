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

