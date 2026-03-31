# Research: 订阅对账与续费治理

## Decision 1: 对账应收口径采用“净应收”

- Decision: 应收金额统一为 `账单金额 + 税费 - 优惠`（净应收）。
- Rationale:
  - 与澄清结论一致，避免财务与运营口径冲突。
  - 可直接映射“用户理论应付”，便于解释差异来源。
  - 对后续看板与导出一致性最好。
- Alternatives considered:
  - 账单原始金额（不含税优惠）：统计简单但偏离实际收款目标。
  - 支付侧应收：依赖支付系统口径，可能放大外部系统噪声。

## Decision 2: 差异处置采用分层 SLA

- Decision: 高/中/低风险任务分别要求 24h / 48h / 72h 闭环。
- Rationale:
  - 高风险先处理，直接降低资金风险。
  - 中低风险留出人工缓冲，提升整体人效。
  - 与运营排班和工单优先级机制天然匹配。
- Alternatives considered:
  - 固定 SLA（24h 或 48h）：规则简单但资源分配不经济。

## Decision 3: 自动续费失败采用递增 4 次重试

- Decision: 默认重试窗口为 `1h、24h、72h、7d`。
- Rationale:
  - 覆盖“瞬时失败”到“延迟恢复”两类场景。
  - 兼顾恢复率与触达成本，避免过度重试。
  - 便于后续按渠道/地区做策略 A/B 调整。
- Alternatives considered:
  - 固定 3 次重试：对长尾恢复用户覆盖不足。
  - 仅 1 次或不重试：恢复率不足，过早转人工。

## Decision 4: 差异类型采用标准五分类

- Decision: `漏收 / 重复收款 / 金额不一致 / 状态不一致 / 数据缺失`。
- Rationale:
  - 对应处置动作明确，能直接转任务模板。
  - 便于看板归因与 KPI 拆分。
  - 与 finance/reconciliation 规划方向兼容。
- Alternatives considered:
  - 自由文本分类：灵活但不可统计、不可自动化。

## Decision 5: 任务幂等策略

- Decision: 对账任务按 `(tenant_uuid, billing_cycle, run_type)` 保证幂等；差异任务按 `(tenant_uuid, delta_fingerprint, status!=closed)` 防重复。
- Rationale:
  - 解决定时任务重复触发与并发执行导致的重复任务问题。
  - 保证运营只看到一个“活跃任务”，降低误操作。
- Alternatives considered:
  - 仅前端去重：无法防止后端并发写入重复数据。

## Decision 6: 审计与观测最小集

- Decision: 记录三类事件：`reconciliation.generated`、`delta.task.created/closed`、`renewal.retry.executed`。
- Rationale:
  - 覆盖核心链路，可追溯资金与动作。
  - 与现有 observability 目录规范一致，便于统一检索。
- Alternatives considered:
  - 全量细粒度事件：可观测性强但初期噪声高、维护成本高。

## Decision 7: 接口形态

- Decision: 采用 REST 风格 admin 接口，围绕“生成批次、查询差异、处置任务、执行治理、看板汇总”五类动作组织。
- Rationale:
  - 与现有 PowerX 插件 admin 接口风格一致，前端接入成本低。
  - 便于后续拆分异步任务与批次回放。
- Alternatives considered:
  - GraphQL：灵活但与当前插件实践不一致，成本更高。

## Decision 8: 合规与导出

- Decision: 支持账期结束导出对账与治理结果，导出保留关键字段变化与处理结论。
- Rationale:
  - 满足审计留存与财务归档要求。
  - 直接支撑 SC-005（可追溯完整率）。
- Alternatives considered:
  - 仅在线查询：审计场景下不可替代离线归档。
