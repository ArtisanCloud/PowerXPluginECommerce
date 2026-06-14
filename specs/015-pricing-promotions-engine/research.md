# Phase 0 Research — 促销规则引擎（订单自动促销）

## 1. 促销计算顺序

- **Decision**: 采用 `pricebook 基础价 -> promotion 自动促销 -> coupon 用户券 -> payable 应付金额`。
- **Rationale**: 促销是运营配置的自动权益，应在用户选择优惠券前先影响订单金额；这与 `docs/plan/pricing/promotions.md` 和 014 优惠券机制对齐，能让券模块基于促销后的金额判断门槛与叠加限制。
- **Alternatives considered**:
  - 先券后促销：会让自动促销受用户券选择影响，客服解释和金额复算更复杂。
  - 促销与券同时求最优：一期复杂度过高，会引入组合搜索和预算/限额问题。

## 2. MVP 规则模型

- **Decision**: 一期只支持订单级 `amount_off` 与 `percent_off`，范围支持全场、指定 SKU、指定渠道，条件支持订单金额门槛和有效期。
- **Rationale**: 这是订单自动促销的最小闭环，可以覆盖常见满减/折扣场景，同时避免买赠、组合包、免邮等对库存、履约和售后产生额外耦合。
- **Alternatives considered**:
  - 直接支持买赠/组合包：需要库存锁定、赠品行、售后拆分，超出本 feature 的交易闭环目标。
  - 做通用规则 DSL：灵活但难测试、难解释，不符合 MVP 收口要求。

## 3. 多促销冲突与互斥

- **Decision**: 先按互斥组筛选，同一 `exclusion_group` 命中多条时取优惠金额最大的一条；随后按 `priority ASC` 和 `stackable` 判断是否继续叠加。
- **Rationale**: 互斥组内最大优惠对运营和客服最容易解释；优先级保留跨组计算顺序；`stackable=false` 可表达“命中后停止同层促销”的强约束。
- **Alternatives considered**:
  - 同组按优先级取一条：运营容易配置出非最优优惠，用户体验不可预期。
  - 全部命中后穷举最优组合：金额最优但复杂度高，后续有预算/限额后更难保证性能。

## 4. 促销与优惠券叠加

- **Decision**: 促销先计算；若命中促销的 `stackable_with_coupon=false`，订单保留促销并拒绝优惠券，返回 `promotion_excludes_coupon`。
- **Rationale**: 自动促销由运营活动决定，优惠券是用户资产；在促销明确排斥券时保持促销优先能保证规则确定性，并复用 014 券机制中的拒绝原因展示。
- **Alternatives considered**:
  - 保留优惠券、拒绝促销：会让运营活动效果不稳定，且自动促销无法形成确定快照。
  - 二者择优：对用户友好但一期解释成本高，需要新增“最优权益选择”策略。

## 5. 有效期边界

- **Decision**: 促销有效期以订单提交时刻为准，时间精度到秒；`submitted_at <= valid_to` 视为有效。
- **Rationale**: 含结束边界能避免用户在活动结束秒内提交订单时发生前后端认知偏差；服务端提交时刻是最终金额依据。
- **Alternatives considered**:
  - 不含结束边界：实现简单但容易产生“结束时间显示仍在该秒内却不可用”的争议。
  - 使用客户端时间：无法可信，且不同端时钟漂移会导致金额不一致。

## 6. 订单行分摊

- **Decision**: 促销优惠按订单行促销前金额占比分摊，分摊余数采用确定性规则补到金额最大的行或最后一行。
- **Rationale**: 促销前金额占比与售后退款、对账解释最直观；分摊必须确定，避免同一订单重复计算结果不同。
- **Alternatives considered**:
  - 按数量分摊：高低价 SKU 混合订单会造成优惠成本扭曲。
  - 不做分摊：售后和财务无法解释订单行优惠来源。

## 7. 数据持久化

- **Decision**: 使用 `promotion_campaigns` 保存规则 JSON，`promotion_audit_logs` 记录配置/命中审计，`order_promotion_snapshots` 固化订单当次促销结果。
- **Rationale**: JSON 规则能支持 MVP 后续扩展，独立快照保证历史订单不受促销编辑、停用、删除影响；审计日志满足客服、财务和运营追踪。
- **Alternatives considered**:
  - 将条件/动作拆成多张规范化表：查询清晰但一期开发成本高，且规则还未稳定。
  - 只写订单 price_snapshot：可减少表数量，但促销独立查询、分摊、审计能力不足。

## 8. API 与 UI 入口

- **Decision**: 管理端主入口为 `/api/v1/admin/promotions` 与 `web-admin/app/pages/pricing/promotions.vue`；交易侧优先接入订单 quote/下单，保留 `/v1/promotions/quote` 作为调试/独立试算契约。
- **Rationale**: 促销规则属于定价中心配置，不应由 `market/promotions.vue` 直接承担核心规则编辑；交易侧最终必须进入订单金额链路。
- **Alternatives considered**:
  - 仅做市场活动页：会把定价规则和营销活动概念混在一起。
  - 只提供独立促销 quote：无法保证下单最终金额与订单快照一致。
