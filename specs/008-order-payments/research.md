# Research: 营销支付与小程序支付

## Decision 1: 退款范围
- Decision: 支持全额与部分退款。
- Rationale: 覆盖常见售后与客服处理场景，减少后续功能补齐成本。
- Alternatives considered: 仅全额退款；仅部分退款。

## Decision 2: 对账频率
- Decision: 支持日/周对账。
- Rationale: 满足财务常用结算节奏与差异处理需求。
- Alternatives considered: 仅日对账；仅周对账；自定义周期。

## Decision 3: 支付确认兜底
- Decision: 回调优先 + 前端短时轮询兜底。
- Rationale: 覆盖回调延迟/丢失场景，平衡体验与成本。
- Alternatives considered: 仅回调；长时间后台轮询；仅前端轮询。

## Decision 4: 风控拦截处理
- Decision: 提示联系支持 + 记录风险事件。
- Rationale: 既保护用户体验，又保证运营可追溯处理。
- Alternatives considered: 直接失败提示；自动重试；进入人工审核队列。

## Decision 5: 支付失败重试策略
- Decision: 允许 1 次立即重试 + 订单列表继续支付。
- Rationale: 降低支付流失，同时避免无限重试。
- Alternatives considered: 不提供重试；多次立即重试；自动重试。

## Decision 6: 回调幂等与状态确认
- Decision: 以支付单号为幂等键，回调仅允许从未终态进入终态；重复回调直接返回成功。
- Rationale: 避免多次回调导致状态回滚或重复履约。
- Alternatives considered: 仅依赖订单号幂等；忽略幂等要求。

## Decision 7: 后台代客下单流程对齐小程序
- Decision: 后台新建订单遵循“先选 SPU + 规格 → 再定位唯一 SKU”的流程，并支持一次订单包含多个 SKU 明细。
- Rationale: 与小程序下单逻辑一致，降低选错 SKU 与培训成本，支持组合下单场景。
- Alternatives considered: 直接搜索 SKU；仅支持单 SKU 下单。

## Decision 8: 客户选择的搜索与最近排序
- Decision: 客户下拉支持关键词搜索；无关键词时返回最近下单时间降序的 TopN，不展示分页控件。
- Rationale: 避免大规模客户下拉全量加载，提高查找效率。
- Alternatives considered: 分页列表；全量拉取。
