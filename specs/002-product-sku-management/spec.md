# Feature Specification: SKU 与变体管理能力增强

**Feature Branch**: `002-product-sku-management`  
**Created**: 2025-12-18  
**Status**: Draft  
**Input**: User description: "根据docs/plan/product/sku.md的prd，生成相关的spec文档"

## Clarifications

### Session 2025-12-18

- Q: 批量价格/库存调整是否需要审批流程，若需要应如何触发？ → A: 支持按操作类型或金额阈值配置审批策略

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 运营批量生成 SKU (Priority: P1)

商品运营在编辑 SPU 时可打开 SKU 生成器，系统按所选规格组合自动列出全部潜在 SKU，允许运营勾选需要的组合、设置默认价格/条码/库存，并一键写入 SKU 列表，再继续逐条完善。

**Why this priority**: 没有可售 SKU，后续价格、库存与渠道动作无法开始，因此 SKU 生成是整个变体管理的入口能力。

**Independent Test**: 仅交付生成器也能独立验证——给定含多个规格的 SPU，运营可独立完成 SKU 批量创建并保存结果。

**Acceptance Scenarios**:

1. **Given** SPU 已配置两个及以上规格，**When** 运营进入 SKU 生成器并勾选组合后点击生成，**Then** 系统按所选组合创建 SKU 并继承默认值。
2. **Given** 部分组合被取消选择，**When** 运营点击生成，**Then** 仅勾选的组合会被写入 SKU 列表且不会生成重复。
3. **Given** 运营修改默认重量或条码模板，**When** 再次生成新的 SKU，**Then** 新增条目继承更新后的默认值，而已生成的 SKU 保持原值。
4. **Given** 运营在“关联 SKU”弹窗中手动录入两条无规格的 SKU，**When** 点击保存关联，**Then** 两条记录都写入 `product_skus`（含 `product_sku_attributes`），并可立即通过 `GET /api/v1/admin/product/skus?spuId=...` 返回且在 SPU 页面摘要中展示相同结果。

---

### User Story 2 - 批量调整价格与库存 (Priority: P2)

运营或定价经理在 SKU 列表、矩阵或批量操作弹窗中选择若干 SKU，可设置固定值、百分比或导入模板的方式批量修改价格、库存或起订量，并在提交前看到影响预览、提交后生成后台任务并记录结果。

**Why this priority**: 批量调整是日常高频操作，影响营收与库存准确度，重要性仅次于生成 SKU。

**Independent Test**: 可单独发布批量调整能力，测试通过列表选择、设置规则、提交任务、查看结果即可验证价值。

**Acceptance Scenarios**:

1. **Given** 运营选中 50 个 SKU，**When** 选择“批量调整价格”并输入 +5% 规则，**Then** 预览中显示每条 SKU 新旧价格差异，提交后生成任务记录并更新成功条目。
2. **Given** 库存人员上传包含库存数量的模板，**When** 系统解析模板并发现缺失必填字段，**Then** 阻止执行并返回错误详情下载。
3. **Given** 夜间执行批量任务失败，**When** 用户在审计记录查看该任务，**Then** 可看到失败原因、影响条目并支持重新触发。

---

### User Story 3 - 渠道映射与库存可视 (Priority: P3)

渠道运营在 SKU 详情中选择渠道并维护渠道 SKU ID、上架状态、同步策略，同时可查看实时库存、锁定量与预警。发布动作会调用渠道映射流程，若远端报错需在详情及任务日志展示。

**Why this priority**: 虽次于生成与批量操作，但渠道映射直接影响各渠道能否销售，需要对外稳定性与透明度。

**Independent Test**: 在不改动生成/批量逻辑的情况下，只要 SKU 详情能够配置渠道并触发一次发布即可独立验收。

**Acceptance Scenarios**:

1. **Given** SKU 已存在，**When** 渠道运营配置渠道 SKU ID 与上架时间后点击“发布”，**Then** 系统记录映射关系并将状态同步到渠道列表。
2. **Given** 渠道返回错误或库存同步失败，**When** 用户查看 SKU 详情，**Then** 能见到失败信息、时间戳以及重试入口。

---

### User Story 4 - mini-app 可售性聚合（价格/库存/渠道可见性）(Priority: P1)

为实现“能上线开始做购买”的最短闭环，mini-app 不应在列表/详情/规格选择/下单按钮等位置重复拼装可售性判断；后端需要提供统一的聚合结果：`sellable + reasons[] + price + availableQty`（并按 `channel` 维度评估），前端仅依赖该结果决定“隐藏/置灰/禁用下单”。

**Why this priority**: 没有统一的可售性口径，前端会出现重复判断与口径不一致，导致“看得见但下不了单/提示不一致”等线上问题；该能力是从“可展示”走向“可购买”的硬门槛之一。

**Independent Test**: 构造一个 SPU 下 2 个 SKU：一个有价有库存、一个缺货/无价；调用 `sellability` 接口应返回每个 SKU 的可售标记与原因码，mini-app 按此禁用下单并可解释提示。

**Acceptance Scenarios**:

1. **Given** SKU 未配置对外价，**When** 评估可售性，**Then** `sellable=false` 且 `reasons` 包含 `NO_PUBLIC_PRICE`。
2. **Given** SKU 可用库存为 0，**When** 评估可售性，**Then** `sellable=false` 且 `reasons` 包含 `OUT_OF_STOCK`，并返回 `availableQty=0`。
3. **Given** SPU 已发布但渠道未启用或不在可售窗口，**When** 以该 `channel` 评估可售性，**Then** `sellable=false` 且 `reasons` 包含 `CHANNEL_DISABLED` 或 `NOT_IN_AVAILABILITY_WINDOW`。
4. **Given** 可售性接口返回 `sellable=true` 的 SKU 集合，**When** mini-app 渲染列表/详情并点击下单，**Then** 仅对可售 SKU 启用下单按钮，不可售 SKU 提示原因且不触发下单请求。

---

### Edge Cases

- SPU 仅有一个规格或无规格：生成器需允许创建单一 SKU 并跳过矩阵视图。
- 已存在的 SKU 再次运行生成器：应提示冲突并允许仅补充缺失组合。
- 批量导入包含重复条码/编码：系统需逐条标记冲突并拒绝写入冲突项。
- 渠道映射时目标渠道关闭或凭证失效：阻止发布并输出可执行的修复提示。
- 库存同步延迟超过 SLA：在 SKU 详情展示“库存更新时间”并标橙提醒。
- 可售性聚合原因码需稳定：新增原因码必须向后兼容（前端默认兜底为“暂不可售”）。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统必须提供 SKU 生成器，可基于 SPU 规格组合动态列出所有潜在 SKU，并允许用户勾选、筛选和批量设置默认字段。
- **FR-001A**: 系统必须提供“销售规格维度/规格取值”的一等数据管理能力：品类维度维护可复用销售规格模板；SPU 维度维护实际启用的 `SpecGroup(code/name/required/sort/status)` 与其 `SpecOption(code/name/meta/sort/status)`，作为 SKU 组合与前端选择的唯一交易来源。
- **FR-002**: 生成器需支持设置条码、成本、起订量、重量、尺寸、单位等默认值，并在创建时写入每个 SKU，可对已有 SKU 选择是否同步。
- **FR-003**: SKU 列表与矩阵视图必须提供批量选择、复制、快速编辑、差异化字段配置，并可针对单个 SKU 编辑媒体、物流及供应信息。
- **FR-004**: 列表页需提供 SPU、类目、规格、渠道、库存区间、标签等过滤；批量操作需限定权限并支持上/下架、价格调整、库存调整、导出、指派仓库。
- **FR-005**: 批量调整提交前必须提供预览（展示受影响条目及新旧值），提交后创建可追踪任务，任务支持状态、进度、错误反馈与重新执行。
- **FR-006**: 系统需与库存模块联动，实时展示可售、锁定、在途库存，并根据阈值触发预警；需与库存同步接口保持 ≤5 分钟延迟。
- **FR-007**: 渠道映射界面必须管理渠道 SKU ID、上架状态、生效时间及同步方式，支持一键推送与错误回显，并能按渠道筛选映射状态。
- **FR-008**: 条码/编码管理需支持自动生成（EAN/UPC 模式）、手动导入、唯一性校验，以及批量打印条码标签功能。
- **FR-009**: 导入/导出需提供标准模板，允许选择字段范围，导入过程校验并生成可下载的错误报告；导出需支持筛选条件及任务异步执行。
- **FR-010**: 系统需支持可选序列号/批次信息的录入与查询，所有关键操作需写入审计日志，并受权限组（读取、管理、批量、渠道）控制。
- **FR-011**: 批量价格或库存调整需支持可配置的审批策略，可按操作类型、影响范围或金额/数量阈值自动判断是否进入审批流，并记录审批人及结论。
- **FR-012**: Web Admin 中的 SKU 列表、矩阵视图以及 SPU 编辑页的“关联 SKU”区域必须读取真实 API 数据（`GET /api/v1/admin/product/skus` 等）并在保存后刷新，禁止继续使用 mock 数据或空白占位。
- **FR-013**: “关联 SKU”弹窗/生成器产生的数据必须实时落地 `product_skus` 及关联表，摘要/列表均以该表为唯一数据源；版本 `payload` 仅用于审批、回滚与审计，不可再驱动前端展示或成为唯一存储。
- **FR-014**: SKU 必须持久化 `spec_signature`（如 `color=red|size=m`），并在数据库层面做 `(tenant_uuid, spu_id, spec_signature)` 唯一约束，防止同一规格组合重复创建多个 SKU。
- **FR-015**: 系统必须提供统一“可售性聚合”能力，按 `channel` 维度评估 SKU 的 `sellable + reasons[] + price + availableQty`，并输出稳定的原因码集合，供 mini-app 与后续下单链路复用。
- **FR-016**: mini-app 必须以可售性聚合结果作为唯一门槛决定“隐藏/置灰/禁用下单”，禁止在前端散落价格/库存/渠道窗口等硬门槛判断逻辑。

### API Contract Additions

- **Admin - 规格定义（SPU 维度）**
  - `GET /api/v1/admin/product/spus/{spuId}/spec-groups`：返回规格维度及其取值（嵌套 options）。
  - `PUT /api/v1/admin/product/spus/{spuId}/spec-groups`：整体替换该 SPU 的规格维度与取值（用于管理端可视化编辑）。

- **Admin - 销售规格模板（品类维度）**
  - `GET /api/v1/admin/product/categories/{categoryId}/sale-specs`：返回该品类可复用的销售规格模板。
  - `PUT /api/v1/admin/product/categories/{categoryId}/sale-specs`：整体替换该品类的销售规格模板；SPU 需显式同步，不自动覆盖。

### 销售规格到 SKU 的业务链路

SKU 不直接从品类模板生成，而是从 SPU 实际启用规格生成。这样可以保证“品类有标准、商品可裁剪、SKU 可交易”。

业务链路：

```text
品类销售规格模板
  -> SPU 点击“从品类同步”
  -> SPU 删除不用的规格或规格值
  -> 保存为 SPU 实际启用规格
  -> SKU 生成器按 SPU 规格做组合
  -> 每个组合落成一个 SKU
```

示例：

```text
品类：毛绒玩具
  模板规格：
    尺寸：10cm、20cm、30cm
    体型：普通体、海星体、骨架体

SPU：努努娃娃
  实际启用：
    尺寸：10cm、20cm
    体型：海星体、骨架体

生成 SKU：
  10cm / 海星体
  10cm / 骨架体
  20cm / 海星体
  20cm / 骨架体
```

这能避免同一品类下出现“尺寸/大小/高度”“10cm/10厘米/10 CM”等不同写法，也能让后续筛选、导入、渠道映射和库存管理更稳定。

- **MiniApp - 一次取齐用于规格选择**
  - `GET /api/v1/mini-app/products/{spuId}/detail`：返回 `spu + spec(groups/options) + skus(specSignature + spec映射)`，前端据此做禁用态与 skuId 匹配。

- **MiniApp - 可售性聚合（购买闭环 MVP）**
  - `GET /api/v1/mini-app/products/{spuId}/sellability?channel=xxx&locale=zh-CN`：返回该 SPU 下 SKU 的 `sellable + reasons[] + price + availableQty` 聚合结果，mini-app 用于“是否可下单”的统一判断。

### Key Entities *(include if feature involves data)*

- **品类销售规格模板**: 在品类维度沉淀可复用的销售规格维度和值域，供同品类 SPU 显式同步和裁剪。
- **SPU（标准产品单元）**: 定义产品主信息及实际启用规格；驱动 SKU 生成器的规格集合。
- **SKU**: 具体可售变体，包含规格组合、编码、条码、状态、价格、库存阈值、物流与供应商信息。
- **SKU 属性/规格值**: 记录 SKU 与各规格值的关联，确保组合唯一性与展示顺序。
- **SKU 渠道映射**: 存储渠道标识、渠道 SKU ID、状态、生效时间、同步策略以及最近一次推送结果。
- **SKU 库存快照**: 保存实时库存、锁定量、在途量与最后更新时间，为预警与展示提供依据。
- **批量任务**: 描述批量价格/库存/状态调整或导入/导出任务，含提交人、规则、影响范围、执行进度与结果。
- **SKU 媒体**: 附着于 SKU 的图片或视频，标注渠道/主图属性及排序。
- **可售性评估结果（Sellability Result）**: 面向前端/mini-app 的聚合输出，统一表达是否可下单及原因码，避免多处散落判断。

### Assumptions & Dependencies

- 规格与属性模块能够返回合法且排序后的规格集合，供生成器直接使用。
- 库存、定价、渠道服务已提供统一的同步接口与回调错误信息。
- 条码打印依赖现有打印组件，仅需产出可下载的标签模板。
- 审计模块可写入结构化操作日志并在详情页复用展示组件。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 运营可在 10 分钟内完成 50 个 SKU 的生成、默认值设置与保存，且无失败。
- **SC-002**: 批量任务在高峰期的成功率 ≥ 98%，失败任务 100% 提供可重试入口与错误说明。
- **SC-003**: 渠道映射的同步准确率 ≥ 99%，且每次推送均记录可追踪日志。
- **SC-004**: 库存数据展示与真实库存源之间的时延 ≤ 5 分钟，库存预警通知在异常发生后 2 分钟内出现。
- **SC-005**: 导入校验可在 1 分钟内返回 1,000 行数据的错误报告，导出响应在 2 分钟内生成下载链接。
- **SC-006**: 条码/编码重复率控制在 <0.5%，发现重复时阻止写入并提示冲突来源。
- **SC-007**: mini-app 仅通过一次“可售性聚合”接口即可得到“可下单与否 + 可解释原因 + 价格 + 可用库存”，且不同页面的可售判断口径一致。
