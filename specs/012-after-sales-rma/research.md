# Research: 售后 RMA 与退货门户

## Decision 1: 采用“分层时间窗口”而非统一窗口
- Decision: 仅退款（付款后 24 小时且未发货）、退货退款（签收后 7 天内）、换货（签收后 15 天内）。
- Rationale: 同时兼顾风控与用户预期，规则与主流电商习惯一致。
- Alternatives considered:
  - 统一窗口：规则简单但无法覆盖不同售后类型风险差异。
  - 完全可配置窗口：灵活但首版实现与运维成本过高。

## Decision 2: 重复申请“进行中禁止，终态可新开”
- Decision: 同一订单明细在存在进行中售后单时禁止重复申请；终态后允许新建售后单并要求新原因。
- Rationale: 防止并发冲突，同时保留合理二次申诉能力。
- Alternatives considered:
  - 永久只允许一次：风控强但用户体验差。
  - 任意重复：后台合单成本高、数据容易失真。

## Decision 3: 售后状态机采用受控流转
- Decision: 定义 `pending -> accepted -> reviewing -> approved/rejected -> completed/closed` 受控链路。
- Rationale: 保证审核动作可审计、可回放，避免状态跳跃。
- Alternatives considered:
  - 自由流转：灵活但难治理，容易出现非法状态。
  - 仅两态（处理中/完成）：信息粒度不足，无法支撑运营协作。

## Decision 4: 换货首版仅做“申请+审核+回填”
- Decision: 首版不触发自动补发、不做库存自动占用，仅处理申请、审核与状态回填。
- Rationale: 在不扩大范围前提下快速建立可用业务入口。
- Alternatives considered:
  - 完整自动补发链路：价值高但牵涉库存、履约、物流多域复杂耦合。
  - 首版不支持换货：范围更小，但与业务目标不一致。

## Decision 5: 联动最小集合优先“状态一致 + 防重”
- Decision: 仅落地订单售后标记同步、退款防重、逆向物流关联三项最小联动。
- Rationale: 先确保跨域一致性，避免一次性引入财务结算复杂度。
- Alternatives considered:
  - 全量财务联动：实现周期与回归成本过高。
  - 无联动：会造成订单侧与售后侧状态漂移。

## Decision 6: 权限模型采用“客户视角 + 运营视角”双轨
- Decision: 客户仅可查看本人售后单；运营按后台 RBAC 处理租户内售后单。
- Rationale: 满足最小权限原则并与现有 RBAC 架构对齐。
- Alternatives considered:
  - 统一开放查询：越权风险高。
  - 仅后台可见：用户端体验不足。
