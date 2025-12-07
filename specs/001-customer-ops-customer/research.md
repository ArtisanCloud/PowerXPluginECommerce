# Research Notes

## Decision 1: Table 与筛选实现方式
- **Decision**: 继续使用 Nuxt UI `UTable` + Pinia store，同步 Query 参数并通过后端分页 API 拉取数据。
- **Rationale**: 与现有 web-admin 列表保持一致，支持自动列定义与服务端分页，无需引入新组件库。
- **Alternatives considered**:
  - 自研虚拟滚动：对当前 50k 量级分页意义不大，维护成本高。
  - 使用第三方 DataGrid（AG Grid 等）：会引入额外许可证与体积。

## Decision 2: 批量任务与审计
- **Decision**: 所有批量操作（导入、导出、批量标签/负责人/提醒）调用既有 `/api/customers/bulk-actions`、`/api/customers/export` API，并由任务中心返回任务 ID；前端仅展示状态并跳转审计日志。
- **Rationale**: 复用后端任务与审计能力，满足宪章的可观测要求，减少重复实现。
- **Alternatives considered**:
  - 在前端自行轮询实际结果：仍需任务中心，因此改为直接使用既有任务 API。
  - 单次请求内同步完成批量操作：风险大，无法追踪大批量失败记录。

## Decision 3: 会员保级提醒
- **Decision**: 会员视角的保级提醒调用专属 `POST /api/customers/bulk-remind`，并将渠道（短信/邮件/站内）作为 payload 字段；任务中心记录结果。
- **Rationale**: 与会员积分模块共用提醒 API，避免重复设计，方便后续旅程联动。
- **Alternatives considered**:
  - 直接发放优惠券：与权益策略不同，不满足“提醒”场景。
  - 手动导出名单后在线下处理：缺乏可追踪性与 SLA。
