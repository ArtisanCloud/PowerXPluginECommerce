---
id: ecommerce.template.basic
name: template
title: 电商模板基础能力
provider: com.powerx.plugins.ecommerce
version: 1.0.0
description: 管理电商插件的基础模板对象，用于验证插件能力注册、Agent 调用和安装态能力路由。
intent_examples:
  - 帮我创建一个标题为测试模板的模板，描述是用于验证插件能力，内容是这是一条测试内容
  - 查询 ID 为 123 的模板
  - 列出所有模板
response_guidance:
  capability_intro:
    - 说明这是电商插件的基础模板对象能力。
    - 能力介绍只概括创建模板。
  capability_howto:
    - create 需要用户提供标题、描述和内容。
  clarify_params:
    - 用户只说创建模板时，只追问：请提供这个模板的标题、描述和内容。
action_required_args:
  create:
    - template.title
    - template.description
    - template.content
slot_mapping:
  template.title:
    labels: ["标题", "名称", "模板标题"]
  template.description:
    labels: ["描述", "用途", "说明"]
  template.content:
    labels: ["内容", "正文", "模板内容"]
capability: ecommerce.template
action_capabilities:
  create: com.powerx.plugins.ecommerce.template.create
visibility: tenant
status: active
executor:
  type: capability
  capability: ecommerce.template
  prepare_capability: com.powerx.plugins.ecommerce.template.prepare
  action_map:
    create: com.powerx.plugins.ecommerce.template.create
  timeout_ms: 30000
  async_supported: true
  risk_level: low
input_schema: ./schema.input.json
output_schema: ./schema.output.json
---

# 电商模板基础能力

## Purpose

管理电商插件的基础模板对象。该技能用于验证安装态插件能力注册、Agent 参数收集和能力调用链路。

## When To Use

当用户希望创建电商插件模板对象时使用。用户可能会说“创建模板”，也可能直接给出标题、描述和内容。

## Instructions

- 识别用户创建模板对象的意图，并转换为结构化 `create` action。
- 创建时提取 `template.title`、`template.description`、`template.content`。
- 如果缺少必要参数，先用自然语言追问缺失信息。
- 不要要求用户必须输入 JSON。
