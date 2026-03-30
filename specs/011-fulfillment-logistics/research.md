# Research: 履约与物流全模块闭环

## Decision 1: 采用 M1/M2/M3 分阶段交付
- Decision: 按 M1（正向履约）→ M2（任务异常）→ M3（三方适配与逆向增强）交付。
- Rationale: 降低一次性上线风险，先拿到可运行主链路并逐步扩展。
- Alternatives considered:
  - 一次性全量上线：范围过大，回归风险高。
  - 仅做物流核心：无法覆盖仓内与售后闭环。

## Decision 2: 运单创建入口“订单触发为主，手工补单为辅”
- Decision: 订单履约触发是主入口，人工仅用于异常补单。
- Rationale: 保证主流程自动化与一致性，同时保留运营兜底能力。
- Alternatives considered:
  - 仅自动触发：异常场景处理弹性不足。
  - 仅人工创建：易漏单，效率与一致性差。

## Decision 3: 轨迹状态源“承运商回调优先”
- Decision: 承运商回调为主状态源，人工仅补录与修正。
- Rationale: 外部状态源更接近真实物流状态，能降低人为覆盖导致的数据漂移。
- Alternatives considered:
  - 人工优先：容易破坏状态一致性。
  - 同权覆盖：冲突场景难以判定。

## Decision 4: 幂等键统一为“事件ID + 运单号”
- Decision: 所有回调/触发幂等判定使用联合键。
- Rationale: 同时规避“同事件重放”和“同运单多次触发”两类重复风险。
- Alternatives considered:
  - 仅事件ID：无法覆盖无事件ID或事件重复生成场景。
  - 仅运单号：无法区分同运单不同事件。

## Decision 5: 异常 24h 自动升级
- Decision: 超过 24 小时未处理的异常自动升级并记录责任链。
- Rationale: 与 spec KPI 对齐，规则简单明确，便于运营执行。
- Alternatives considered:
  - 2h/8h 升级：对当前组织负担较重。
  - 纯人工跟进：不可控且难审计。

## Decision 6: 第三方适配采用统一能力边界
- Decision: 使用统一动作边界（创建运单、查询/推送轨迹、连通性测试）承载不同承运商适配。
- Rationale: 降低业务层耦合，便于扩展多个服务商。
- Alternatives considered:
  - 每家承运商单独业务流程：维护成本高、测试复杂。
