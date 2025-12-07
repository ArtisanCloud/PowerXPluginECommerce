# 推荐 / 分销 PRD

> 覆盖路径：`/customer/affiliates/index.vue`、`/customer/affiliates/distributors/**`、`/customer/affiliates/rules/**`、`/customer/affiliates/settlements/**`，以及与联盟、渠道奖励相关的弹窗/审批流。

## 1. 背景与目标
- 电商插件需要支持拉新、联盟、分销等增长模式，通过邀请、渠道合作扩展 GMV。
- 现有页面展示 KPI 与规则示例，但缺少完整的招募、规则、结算、风控闭环。

**目标**
1. 构建可配置的推荐/分销策略，支持多层级、多维度奖励。
2. 提供合作伙伴全生命周期管理（招募→审核→运营→结算→风控）。
3. 打通结算、支付、税务等模块，实现佣金发放、对账与合规审计。
4. 提供实时监控与风控，识别异常邀请、刷单风险。

## 2. 用户与角色
| 角色 | 诉求 |
| --- | --- |
| 增长/渠道运营 | 设计策略、招募合作人、查看业绩、发放奖励 |
| 财务结算 | 审核佣金、处理提现、对账税务 |
| 风控/法务 | 监控异常、拉黑违规账号、确保合规合同 |
| 数据分析 | 查看渠道贡献、ROI、预测预算 |
| 合作伙伴（外部） | 注册、查询业绩、申请提现（前台或 Partner Portal） |

## 3. 信息架构
1. **概览 & KPI (`index.vue`)**
   - 指标卡：邀请人数、奖励金额、本月新增、待发放奖励。
   - Tab：邀请关系、奖励发放、排行榜、风控预警。
2. **分销商管理 (`distributors/index.vue`)**
   - 招募：表单、审批、资料上传、合同关联。
   - 列表：等级、区域、渠道、状态、业绩、风控分。
   - 操作：启停、拉黑、重置邀请码、发送通知。
3. **规则配置 (`rules/index.vue`)**
   - 策略列表：奖励对象、内容、触发条件、分成比例、状态。
   - 策略编辑器：奖励对象、内容、触发条件、分层比例、封顶、合规条款。
   - 历史版本、模拟器（输入订单/邀请场景，输出奖励结果）。
4. **结算管理 (`settlements/**`)**
   - 结算单列表：周期、渠道、金额、状态、凭证。
   - 审核流：运营复核→财务审核→发放→税票管理。
   - 提现管理：合作伙伴发起提现，后台审批、打款、反馈状态。
5. **风控与黑名单**
   - 规则：异常下单、短时间大量邀请、跨地区违规等。
   - 处理：自动冻结、人工复核、恢复或封禁。

## 4. 功能明细
| 模块 | 功能 | 说明 |
| --- | --- | --- |
| 招募门户 | 申请、资料上传、合同签署 | 支持多渠道入口（二维码、邀请码、链接） |
| 审核中心 | 多级审核、自动审批策略 | 审批节点可配置，与 `/settings/orders` 共用流程组件 |
| 分销商档案 | 基本资料、渠道能力、业绩、风控记录 | 支持导入导出、批量分配运营负责人 |
| 规则引擎 | 奖励对象、内容、触发条件、比例、封顶 | 支持多层级（一级/二级）、混合奖励（现金+积分） |
| 奖励执行 | 订单/邀请事件触发奖励任务 | 与订单、支付事件打通，写入奖励流水 |
| 发放 & 提现 | 自动发放、手动发放、提现申请、税务校验 | 与 `payments/providers`、`settlements.vue` 联动 |
| 风控 | 黑名单、异常检测、限额、KYC | 调用风险服务，记录审计 |
| 报表 & 排行 | 渠道 GMV、ROI、Top 代理榜 | 支持导出、订阅，与 `reports/**` 共用组件 |

## 5. 核心流程
1. **招募 & 审核**
   1. 合作伙伴通过邀请链接提交申请，上传资质。
   2. 系统自动预审（KYC、黑名单），随后进入运营/法务审批。
   3. 审批通过后生成分销商档案、分配邀请码/链接、签署合同。
2. **奖励计算**
   1. 订单或邀请事件流入规则引擎 → 匹配策略 → 计算奖励。
   2. 检查封顶、黑名单、冲突规则 → 生成奖励流水。
   3. 若需审批（高额）则进入待审批队列，审批通过后进入待发放。
3. **结算 / 提现**
   1. 周期性生成结算单（按月/按周/自定义），汇总奖励流水。
   2. 运营复核 → 财务审核（校验税票、合同） → 调用支付渠道发放。
   3. 合作伙伴在 Portal 查看结算状态，必要时报税。
4. **风控与黑名单**
   1. 风控任务监测邀请/订单异常 → 触发警报。
   2. 对违规账号执行冻结、限制提现、拉黑，通知合作伙伴。
   3. 记录处理结果及审计，支持申诉流程。

## 6. 数据 & API
- **表**：`affiliate_partners`、`affiliate_contracts`、`affiliate_rules`、`affiliate_rewards`、`affiliate_settlements`、`affiliate_withdrawals`、`affiliate_risk_events`、`affiliate_audit_logs`。
- **API（示例）**：
  - `POST /api/affiliates/apply`（前台），`GET/POST /api/admin/affiliates`（后台 CRUD）。
  - `POST /api/admin/affiliates/{id}/status`（启停、拉黑）。
  - `GET /api/admin/affiliates/rules`、`POST /api/admin/affiliates/rules`、`POST /api/admin/affiliates/rules/{id}/simulate`。
  - `GET /api/admin/affiliates/rewards`、`POST /api/admin/affiliates/rewards/bulk-approve`。
  - `GET /api/admin/affiliates/settlements`、`POST /api/admin/affiliates/settlements/{id}/approve`、`POST /api/admin/affiliates/settlements/{id}/payout`。
  - `POST /api/admin/affiliates/risk/flag`、`POST /api/admin/affiliates/risk/release`。

## 7. 权限、审批、审计
- 权限建议：
  - `affiliate.read`：查看列表、关系、奖励。
  - `affiliate.manage`：招募、编辑档案、启停。
  - `affiliate.rules.manage`：配置策略。
  - `affiliate.settlement.manage`：审批、发放。
  - `affiliate.risk.manage`：风控操作。
- 审批流：新合作伙伴、规则变更、批量奖励、结算单、提现单都需审批；可与 `operations` 模块共享审批组件。
- 审计：所有策略调整、奖励发放、风控动作写入 `admin_console_audit_events`，带上请求参数、操作者、审批链。

## 8. KPI / 量化目标
| 指标 | 目标 |
| --- | --- |
| 招募转化率 | ≥ 20% 申请通过率 |
| 渠道贡献 GMV | 每季度增长 ≥ 15% |
| 奖励发放准确率 | ≥ 99.5% |
| 结算周期 | 周期结束后 5 个工作日内发放 |
| 风控响应时间 | 异常 30 分钟内处置 |

## 9. 风险与依赖
- 依赖支付、结算、税务模块处理打款、发票；需同步 `settlements.vue`、`payments/providers.vue`。
- 与订单/支付事件耦合，需保证事件幂等，防止重复奖励。
- 风控需要实时数据与算法支持，可与现有安全服务对接。
- 合同/审批流程需与 PowerX 底座统一，避免重复实现。

## 10. Backlog / 下一步
- 合作伙伴 Portal：提供自助申请、查看业绩、提现。
- 多层级分销：支持二级/三级分佣、团队奖励。
- 营销联动：结合营销触达（短信/邮件）推送招募计划、激活提醒。
- 法规合规：内置税收规则、自动生成代扣代缴报表。
- AI 风控：基于行为模式智能识别异常邀请/刷单。
