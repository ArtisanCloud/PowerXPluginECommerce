# Phase 0 Research — 定价中心 Pricebook（结论汇总）

> 本文件用于将“影响实现与验收的关键决策”显式化，并给出替代方案对比。结论来源：`/private/var/www/html/ArtisanCloud/X/PowerX/Core/Plugins/com.powerx.plugins.ecommerce/specs/005-pricing-pricebook/spec.md` 的 Clarifications 与 `docs/plan/pricing/pricebooks.md` 的 Phase 1 目标。

## Decision 1：API 前缀与分层

- **Decision**: 使用 `/api/v1` 作为统一前缀；管理端路由置于 `/api/v1/admin/pricing/**`；对外查价使用 `/api/v1/pricing/query`。
- **Rationale**: 与现有 Router/Registry 的 `apiPrefix()` 及 admin 分组一致，符合宪章 Host Contract First 与 RBAC 汇总机制。
- **Alternatives considered**:
  - `/v1/**`：与现有项目内真实路由不一致，易造成联调混乱。

## Decision 2：金额与币种的表示

- **Decision**: 金额使用“最小货币单位整数”（minor units）表达；币种使用 ISO 4217 大写；一期不做自动汇率换算。
- **Rationale**: 避免浮点误差，利于索引、对账与一致性；不引入汇率依赖可降低一期复杂度。
- **Alternatives considered**:
  - `decimal(18,6)`：可读性强但实现细节与序列化易不一致；长期仍需统一舍入与比较策略。

## Decision 3：未命中时回退策略（Clarification：C）

- **Decision**: 仅当“已命中某价目表生效版本，但该版本缺少目标 SKU 条目”时，回退到同币种 `Base Pricebook`；范围不命中/币种不匹配/无任何候选版本可用时，不回退，直接无价/错误。
- **Rationale**: 最大化避免“绕过范围/币种约束导致错价”的风险，同时保留条目缺失的安全兜底。
- **Alternatives considered**:
  - 全量回退到 Base：错价风险高。
  - 从不回退：运营侧条目缺失会导致大量无价，体验差。

## Decision 4：多候选命中时的选择规则（Clarification：A）

- **Decision**: 具体度优先（范围限制维度越多/越精确）→ 显式优先级（若配置）→ 最近发布优先。
- **Rationale**: 结果确定、可解释，符合“运营策略更具体的覆盖更泛化策略”的直觉；便于未来引入规则引擎与更多维度。
- **Alternatives considered**:
  - 最低价优先：可能误用不该给的价，合规与对账风险高。
  - 固定类型顺序：对组合范围与扩展不够灵活。

## Decision 5：无价 vs 错误语义（Clarification：C）

- **Decision**: 配置/数据缺失导致无法定价 → 返回“无价 + 原因码”；参数缺失/非法、币种非法 → 返回“错误 + 错误码”。
- **Rationale**: 便于调用方区分“需要补配置”与“请求要修正”，也利于监控与治理（无价可统计覆盖率，错误可追踪调用质量）。
- **Alternatives considered**:
  - 全部返回错误：调用方可能误判为系统故障。
  - 全部返回无价：会掩盖非法请求问题。

## Decision 6：发布新版本对旧版本的处理（Clarification：A）

- **Decision**: 发布新版本时自动终止旧版本有效期（旧版 `expires_at = 新版 effective_at` 或等效下线），保证同一价目表任一时刻最多一个版本可命中。
- **Rationale**: 简单、确定、易解释；规避“版本有效期重叠”带来的选择复杂度与边界错价。
- **Alternatives considered**:
  - 并行生效：复杂度高，一期风险大。
  - 强校验不重叠：可行但运营操作成本更高。

