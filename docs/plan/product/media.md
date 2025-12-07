# 电商媒体与素材管理 PRD

> 覆盖页面：`web-admin/app/pages/product/media.vue`、商品/SPU/SKU 编辑中的媒体区块，以及后续在营销、促销模块复用的媒体选择器。**插件默认在 standalone（本地开发）场景走 Local 存储实现**，仅当运行在宿主 PowerX 平台时才调用底座 Media Storage（通过 gRPC/OpenAPI）作为唯一存储源；两种模式对上层保持一致的接口与数据结构。

## 1. 背景与目标
- 电商插件需要在商品、SKU、营销活动中快速调用媒体素材，并实现渠道适配、版本控制和合规审计。
- PowerX 底座提供标准媒体存储与处理能力，插件只需封装电商场景的 UI、业务字段及上下文，使上传结果被底座统一管理与复用。

**目标**
1. 提供电商专属媒体面板（入口 + 筛选 + 业务字段），让商品运营直接在插件内管理素材，同时保持与底座 Media 控制台的数据一致。
2. 通过 gRPC/OpenAPI 调用底座服务，写入 `media_id`、元数据与电商上下文（plugin/module/resource），确保宿主模式下 PowerX Media 控制台也能查看与维护这些素材。
3. 支持 local（standalone）与 hosted（宿主）两种存储模式，保证本地开发可运行、发布后自动切换到底座。

## 2. 使用角色
| 角色 | 诉求 |
| --- | --- |
| 商品运营/策划 | 在电商后台快速上传、挑选、标记商品媒体，绑定 SPU/SKU/渠道 |
| 媒资管理员 | 通过底座 Media 控制台统一审核、打标签、管理权限 |
| 渠道运营 | 为不同渠道指定合规尺寸/格式的素材 |
| 品控/法务 | 审核版权、合规，追踪素材来源 |

## 3. 信息架构（插件视角）
1. **电商媒体面板 (`product/media.vue`)**
   - 列表视图、卡片视图，支持按照商品、SKU、渠道、标签、语言筛选。
   - 媒体上传按钮：依据当前运行模式（local/hosted）调用对应存储实现。
   - 媒体详情侧栏：展示底座返回的元数据 + 电商自定义字段（用途、绑定对象、渠道适配状态）。
2. **Media Picker 组件**
   - 在 SPU/SKU/营销表单中嵌入，复用与媒体面板相同的筛选与上传逻辑。
   - 选择后返回 `media_id`、URL、缩略图等信息，写入业务表单。
3. **模式与配置**
   - 环境变量：`MEDIA_STORAGE_MODE=local|hosted`。
   - Hosted：通过 gRPC/OpenAPI 访问 PowerX Media Storage，自动写入 plugin/module/resource 元数据；底座 Media 控制台可见。
   - Local：用于 standalone 开发，可根据 `MEDIA_STORAGE_DRIVER=file|minio|s3|mock` 等配置选择驱动（例如本地硬盘、MinIO SDK、S3 兼容存储）；保持与 hosted 模式一致的返回结构，方便切换。

## 4. 功能清单
| 功能 | 描述 |
| --- | --- |
| 媒体列表与搜索 | 按标签、类型、渠道、绑定资源过滤；支持批量选择 |
| 上传/拖拽 | 支持图片、视频、3D、文档；宿主模式直接上传到底座，携带上下文 |
| 标签/分类 | 允许在插件内添加业务标签（如“天猫主图”），同步写入底座或保存在 `product_media_bindings` 中 |
| 渠道适配 | 设置渠道所需尺寸/比例，调用底座模板生成或记录需求；展示生成状态 |
| 多语言/区域 | 为媒体设置语言、市场信息，方便前台展示控制 |
| 绑定管理 | 展示素材被哪些 SPU/SKU/营销活动使用，可跳转至相应页面 |
| 审批状态 | 显示底座返回的审核状态（待审/通过/拒绝），阻止未审核素材发布 |
| 批量导入/导出 | 导出绑定关系，供媒资团队核查；导入可批量建立绑定 |

## 5. 关键流程
1. **上传与保存（Hosted 模式）**
   1. 用户在媒体面板或 Picker 中点击上传 → 选择文件。
   2. 插件调用底座上传 API，传递 `tenant_id`、`plugin_id`、`module=product`、`resource` 等上下文。
   3. 上传完成后返回 `media_id` 和 URL；插件将 `media_id`、业务用途、绑定对象写入 `product_media_bindings`。
   4. 底座 Media 控制台自动可见该素材，可继续补充标签、审核。
2. **上传与保存（Local 模式）**
   1. 插件将文件写入本地 mock 存储，同时记录 `media_id`（临时生成）。
   2. 元数据结构与 hosted 模式保持一致，方便切换环境。
   3. 发布到宿主环境时，改为调用底座 API，旧数据可通过迁移脚本导入。
3. **引用媒体**
   1. 在 SPU/SKU 页面打开 Picker → 搜索或上传 → 选择素材。
   2. 写入表单并展示缩略图、合规状态。
   3. 发布商品时校验引用素材是否审核通过、渠道适配完成。
4. **渠道适配**
   1. 在媒体详情中配置渠道模板（尺寸、格式、背景）。
   2. 调用底座 `GenerateVariant` 接口生成适配版本；记录生成状态和 URL。
   3. 在渠道上架流程中读取对应版本 URL 推送到渠道。

## 6. 数据与接口
- **插件侧存储**
  - `product_media_bindings`：`media_id`、`resource_type`（spu/sku/campaign）、`resource_id`、`usage`、`channel`、`language`、`tags`、`sort_order`、`storage_mode`、`created_by`。
  - `product_media_tasks`：渠道适配生成任务、状态、错误信息。
- **PowerX Media Storage 接口（Hosted）**
  - gRPC：`Upload`, `ListAssets`, `GetAsset`, `GenerateVariant`, `SetTags`, `SetAcl`。
  - OpenAPI 等价 REST；所有请求需带上 `tenant`、`plugin`、`module`、`resource` headers。
- **插件封装层**
  - 提供 `useMediaService()` composable，根据 `MEDIA_STORAGE_MODE` 调用本地或 hosted 适配器，外部使用方法一致。

## 7. 权限与审计
- 权限项：`media.read`（查看/引用）、`media.manage`（上传/删除/标签）、`media.channel`（渠道适配/生成）、`media.approval`（标记审核状态）。
- 审计：上传、删除、替换、生成渠道版本等操作写入 `admin_console_audit_events`，并在 `product_media_bindings` 中记录 `audit_id`；宿主模式下底座也会记录。
- 访问控制：插件遵循底座返回的权限（如某素材被限制，只能查看到占位提示）。

## 8. KPI / 衡量
| 指标 | 目标 |
| --- | --- |
| 媒体复用率 | ≥ 70% 新商品复用已有素材 |
| 渠道适配成功率 | ≥ 98% |
| 上传失败率 | < 1%（提供重试机制） |
| 审核通过时长 | ≤ 24 小时（由底座驱动，但需在插件里可视化） |

## 9. 依赖与风险
- 依赖 PowerX Media Storage 可用性；需实现重试、降级、失败提示。
- Local → Hosted 切换需有迁移策略，避免 media_id 不兼容。
- 渠道适配模板若底座暂不支持，需在插件内记录需求并提供手动上传替代方案。
- 权限/合规逻辑必须以底座为准，防止绕过审批使用素材。

## 10. Backlog
- 媒体使用报表：统计素材在商品/活动中的使用次数、渠道覆盖。
- AI 辅助：自动标签、质量检测、背景抠图。
- 供应商协同：允许供应商上传素材，插件自动关联并提交审核。
- 批量替换：在渠道/活动中批量替换指定标签的素材。
