#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const capabilitiesDir = path.join(root, "contracts", "capabilities");
const pluginDDir = path.join(root, "plugin.d");

function fail(message) {
  console.error(`❌ ${message}`);
  process.exit(1);
}

function info(message) {
  console.log(`ℹ️  ${message}`);
}

function ok(message) {
  console.log(`✅ ${message}`);
}

function readCapabilityFiles() {
  if (!fs.existsSync(capabilitiesDir)) {
    fail("未找到 contracts/capabilities 目录");
  }
  return fs
    .readdirSync(capabilitiesDir)
    .filter((name) => name.endsWith(".yaml") || name.endsWith(".yml"))
    .sort();
}

function parseCapability(content, fileName) {
  const idMatch = content.match(/^id:\s*([^\n]+)$/m);
  if (!idMatch) fail(`capability 缺少 id: ${fileName}`);
  const id = idMatch[1].trim();

  const versionMatch = content.match(/^version:\s*([^\n]+)$/m);
  const version = versionMatch ? versionMatch[1].trim() : "1.0.0";

  const descriptorPath = `contracts/capabilities/${fileName}`;

  const rbacBlock = content.match(/rbac:\s*([\s\S]*?)(?:\n[a-zA-Z_][^:\n]*:|\n$)/m);
  if (!rbacBlock) fail(`capability 缺少 rbac 段: ${fileName}`);

  const resourceMatch = rbacBlock[1].match(/resource:\s*([^\n]+)$/m);
  if (!resourceMatch) fail(`capability 缺少 rbac.resource: ${fileName}`);
  const resource = resourceMatch[1].trim();

  let actions = [];
  const inlineActions = rbacBlock[1].match(/actions:\s*\[([^\]]+)\]/m);
  if (inlineActions) {
    actions = inlineActions[1]
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
  } else {
    const listActions = rbacBlock[1]
      .split("\n")
      .map((line) => line.match(/^\s*-\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*$/)?.[1] || "")
      .filter(Boolean);
    actions = listActions;
  }

  if (actions.length === 0) {
    fail(`capability 缺少 rbac.actions: ${fileName}`);
  }

  return {
    id,
    version,
    descriptorPath,
    resource,
    actions,
  };
}

function ensureDir(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

function toYamlList(items, indent = 0) {
  const pad = " ".repeat(indent);
  return items.map((item) => `${pad}- ${item}`).join("\n");
}

function writeCapabilitiesYaml(capabilities) {
  const lines = ["capabilities:", "  provides:"];
  for (const cap of capabilities) {
    lines.push(`    - id: ${cap.id}`);
    lines.push(`      version: ${cap.version}`);
    lines.push(`      descriptor: ${cap.descriptorPath}`);
  }
  fs.writeFileSync(path.join(pluginDDir, "capabilities.yaml"), `${lines.join("\n")}\n`, "utf8");
}

function writeExposureYaml(capabilities) {
  const lines = ["exposure:", "  channels:"];
  for (const cap of capabilities) {
    lines.push("    - type: rest");
    lines.push(`      capability: ${cap.id}`);
    lines.push(`      rbac: ${cap.resource}:${cap.actions[0]}`);
  }
  fs.writeFileSync(path.join(pluginDDir, "exposure.yaml"), `${lines.join("\n")}\n`, "utf8");
}

function writeRBACYaml(capabilities) {
  const resourceActions = new Map();
  for (const cap of capabilities) {
    if (!resourceActions.has(cap.resource)) {
      resourceActions.set(cap.resource, new Set());
    }
    for (const action of cap.actions) {
      resourceActions.get(cap.resource).add(action);
    }
  }

  const lines = ["rbac:", "  resources:"];
  for (const resource of Array.from(resourceActions.keys()).sort()) {
    lines.push(`    - resource: ${resource}`);
    const actions = Array.from(resourceActions.get(resource)).sort();
    lines.push(`      actions: [${actions.join(", ")}]`);
  }

  fs.writeFileSync(path.join(pluginDDir, "rbac.yaml"), `${lines.join("\n")}\n`, "utf8");
}

function main() {
  const files = readCapabilityFiles();
  if (files.length === 0) {
    fail("contracts/capabilities 下没有任何 capability 描述文件");
  }

  const capabilities = files.map((fileName) => {
    const content = fs.readFileSync(path.join(capabilitiesDir, fileName), "utf8");
    return parseCapability(content, fileName);
  });

  ensureDir(pluginDDir);
  writeCapabilitiesYaml(capabilities);
  writeExposureYaml(capabilities);
  writeRBACYaml(capabilities);

  ok("plugin.d/capabilities.yaml 已同步");
  ok("plugin.d/exposure.yaml 已同步");
  ok("plugin.d/rbac.yaml 已同步");
}

main();
