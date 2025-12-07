# SKU & 内容映射 PRD

> 覆盖 `web-admin/app/pages/channels/sku-mapping.vue` 及上架流程中引用的映射组件。负责本地 SKU 与渠道 SKU/类目/属性的映射、内容适配、自动匹配、异常处理。

## 1. 目标
- 解决本地 SKU 与各渠道/店铺的数据标准差异，保证属性/图片/描述/合规内容的准确同步。

## 2. 信息架构
- 映射列表：本地 SKU、渠道 SKU、渠道、状态、最近同步、错误、负责人。
- 映射详情：属性映射（颜色、尺码、自定义属性）、类目映射、内容模板、图片/媒体映射、合规字段、审核状态。
- 自动匹配面板：提供建议匹配、属性比对、差异提示。

## 3. 功能
| 功能 | 描述 |
| --- | --- |
| 属性/类目映射 | 定义本地属性 → 渠道属性的映射表，支持规则（正则/模板） |
| 自动匹配 | 根据 SKU 名称、属性、条码自动推荐渠道 SKU；人工确认 |
| 内容模板 | 渠道描述模板（多语言、合规文案、标签），支持变量 |
| 媒体适配 | 选择渠道图/视频，自动生成尺寸/格式版本 |
| 合规校验 | 渠道规则校验（敏感词、类目要求、认证文件），预览与提示 |
| 批量导入导出 | 支持 CSV、API 批量映射 |
| 监控 | 映射成功率、失败原因、告警 |

## 4. 数据 & API
- 表：`channel_sku_mappings`、`channel_attribute_mappings`、`channel_content_templates`、`channel_mapping_logs`。
- API：`GET/POST /api/channel/sku-mappings`、`POST /api/channel/sku-mappings/bulk`、`POST /api/channel/sku-mappings/{id}/validate`。

## 5. 权限 & 审计
- 权限：`channel.mapping.read`、`channel.mapping.manage`、`channel.mapping.validate`。
- 审计：映射变更、模板变更、校验结果。

## 6. KPI
| 指标 | 目标 |
| --- | --- |
| 映射成功率 | ≥ 98% |
| 自动匹配命中率 | ≥ 70% |
| 映射异常响应 | < 1 小时 |

## 7. Backlog
- AI 自动映射建议。
- 渠道内容差异化（按国家/语言）管理。
- 与商品媒体/属性库联动。
