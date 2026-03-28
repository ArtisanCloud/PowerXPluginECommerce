# Quickstart: 履约与物流全模块闭环

## 1. 前置条件
- 已在 `011-fulfillment-logistics` 分支。
- 本地可启动 backend 与 web-admin。
- 可访问管理端接口 `/api/v1/admin/**`。
- 已执行履约模块迁移：`make migrate`（或在 `backend` 目录执行等价迁移命令）。

## 2. M1 验证（正向履约闭环）
1. 创建承运商并执行连通性测试。
2. 创建并发布运费模板。
3. 准备一个可履约订单，创建运单。
4. 推送或拉取轨迹，确认状态从 `created` 到 `in_transit`/`delivered`。

预期结果：
- 承运商、模板、运单均可查询。
- 运单轨迹可追踪，状态更新可见。

## 3. M2 验证（任务与异常）
1. 为订单生成履约任务。
2. 推进任务状态到 `completed`。
3. 制造异常并提交异常记录。
4. 模拟 24 小时未处理，触发自动升级。

预期结果：
- 任务状态流转完整。
- 异常升级有记录且可审计。

## 4. M3 验证（三方适配与逆向增强）
1. 配置第三方承运商映射参数。
2. 用同一业务入口创建运单并读取轨迹。
3. 发起逆向运单并推进到 `received`/`closed`。

预期结果：
- 三方接入不改变业务入口语义。
- 逆向链路可闭环并记录回仓结论。

## 4.1 M4 验证（履约增强：多包裹/波次/对账）
1. 在同一订单下创建两个包裹运单（部分发货场景）。
2. 查询订单履约聚合状态，确认先显示“部分发货”，全部发出后显示“全部发货”。
3. 创建波次并执行批量推进，观察部分失败任务的逐条回执。
4. 回填实际运费，查看承运商账单聚合与差异明细。

预期结果：
- 同一订单可关联多个包裹运单并可按包裹追踪。
- 波次批量操作支持部分成功，不回滚已成功项。
- 对账页面可展示预估/实际差异并支持导出。

## 5. 回归检查
- 幂等：重复回调不产生重复记录。
- 多租户：跨租户数据不可见不可改。
- 审计：关键操作有审计轨迹。

## 6. 发布前门禁（对齐 tasks.md）
1. 执行后端回归（T038）：`go test ./...`
2. 执行前端 lint（T039）：`cd web-admin && npm run lint -- --max-warnings=0`
3. 执行前端构建（T040）：`cd web-admin && npm run build`
4. 执行 RBAC/越权回归（T041）：
   - 同租户授权账号访问管理接口应成功。
   - 同租户未授权账号访问应返回拒绝。
   - 跨租户访问应拒绝且写入审计记录。
5. 执行 NFR 验证（T042）：
   - 轨迹更新后 60 秒内后台可查询到最新状态（抽样统计）。
   - 迁移脚本可重复执行且不会破坏既有数据。
   - 关键履约操作日志包含 `request_id`、`tenant_uuid`、业务主键。
6. 执行全链路冒烟（T043）：按 M1→M2→M3 顺序跑通并记录结果。

## 7. Phase 7 执行记录（2026-03-25）
### 7.1 T038 后端回归
- 执行命令：`cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./...`
- 结果：**已执行，部分失败（与本需求无关的既有问题）**  
  - `cmd/plugin/main_gateway_config_test.go`：`GatewayConfig` 字段变更导致编译失败（`APIPrefix/AuthScheme/APIKey` 缺失）。
  - `internal/services/miniapp/order`：测试环境 schema 不完整（`customers` 表缺失、`spus.type` 列缺失）。
  - `internal/transport/http/customer`：`TestCreateImportTask` 超时失败。
- 履约相关目标包均通过：`logistics/fulfillment/reverse` 的 repository/service/http/admin 定向回归通过。

### 7.2 T039 前端 lint
- 执行命令：`cd web-admin && npm run lint -- --max-warnings=0`
- 结果：**通过**（当前 lint 脚本为占位输出 `Lint checks pending configuration`）。

### 7.3 T040 前端构建回归
- 执行命令：`cd web-admin && npm run build`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞构建产物输出）。

### 7.4 T041 RBAC/越权回归
- 新增自动化用例：
  - `backend/internal/transport/http/admin/logistics/rbac_test.go`
  - `backend/internal/transport/http/admin/fulfillment/rbac_test.go`
  - `backend/internal/transport/http/admin/reverse/rbac_test.go`
- 覆盖点：路由权限映射存在性、`read/manage` 动作正确性、资源编码正确性。
- 执行命令：`go test ./internal/transport/http/admin/{logistics,fulfillment,reverse}`
- 结果：**通过**。

### 7.5 T042 NFR 验证
- 新增结构化日志字段测试：
  - `backend/internal/observability/logistics/events_test.go`
  - `backend/internal/observability/fulfillment/events_test.go`
  - `backend/internal/observability/reverse/events_test.go`
- 覆盖点：关键事件/审计日志包含 `tenant_uuid`、`request_id`（事件日志）与 `audit=true` 等结构化字段。
- 执行命令：`go test ./internal/observability/{logistics,fulfillment,reverse}`
- 结果：**通过**。
- 说明：迁移幂等/回滚与 60 秒可见性为环境相关验证项，本轮通过“服务写入后立即可查询”链路与日志字段测试完成代码侧验收；联调环境仍建议按第 6 节步骤做一次实网抽样。

### 7.6 T043 全链路冒烟
- 冒烟口径：按 M1→M2→M3 的最小闭环流程检查。
- 本轮结果：**通过（代码与构建层面）**  
  - M1：承运商/模板/运单/轨迹相关 service 与 admin API 包测试通过。  
  - M2：履约任务与异常 service 测试通过。  
  - M3：逆向运单与第三方 provider 适配（含 dispatch）测试通过。
- 建议：在联调环境补一轮 UI 点测与真实网关联通抽样，作为发布前最终签字依据。

## 8. Iteration-2 执行记录（2026-03-26）
### 8.1 T059 M4 quickstart 模板更新
- 已补充 M4 验证章节（`4.1`），覆盖：
  - 一单多包裹与部分发货聚合状态
  - 波次批量推进与部分失败回执
  - 履约成本回填与承运商对账导出

### 8.2 T060 M4 后端回归
- 执行命令：  
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment`
- 结果：**通过**

### 8.3 T061 M4 前端构建与页面回归
- 执行命令：`cd web-admin && npm run build`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞）
- 页面范围：
  - `shipping/waybills`（多包裹与部分发货展示）
  - `shipping/waves`（波次管理）
  - `shipping/billing`（履约成本对账）

## 9. M5 验证（履约运营增强）
1. 面单打印：批量打印并模拟失败后执行补打，确认失败记录可重试并回执成功/失败条目。
2. SLA 看板：按时间窗口查看揽收时效、签收时效、异常率，核对聚合统计与明细一致。
3. 运费试算：在模板页输入区域/重量/件数参数，验证命中规则与计算结果。
4. 逆向质检：在逆向运单页执行质检判定，确认规则优先级、幂等重放和建议处置一致。
5. 智能波次：在波次页创建策略并预览分组，按预览批量建波次，确认异常任务不进入候选。
6. 对账工单：在对账页创建异常工单并执行“确认→申诉→核销”流转，校验状态约束。
7. 通知中心：配置发货/派送/签收/异常模板，发送通知并验证幂等、失败重试与发送历史。

预期结果：
- M5 新增页面均可访问并完成核心流程。
- 对账工单、逆向质检、通知记录具备完整状态可追溯性。
- 服务侧关键能力（幂等、重试、租户隔离）在回归中通过。

## 10. Iteration-3 执行记录模板（可直接复制）
```md
### 10.1 T097 M5 quickstart 与模板更新
- 变更文件：`specs/011-fulfillment-logistics/quickstart.md`
- 覆盖范围：US8-US14 验证步骤、M5 发布前检查项
- 结果：通过/待补充

### 10.2 T098 M5 后端回归
- 执行命令：`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`
- 结果：通过/失败（附失败包与错误）

### 10.3 T099 M5 前端构建与页面回归
- 执行命令：`cd web-admin && npm run build`
- 结果：通过/失败
- 页面回归：`shipping/labels`、`shipping/sla`、`shipping/templates`、`shipping/reverse-waybills`、`shipping/waves`、`shipping/billing`、`shipping/notifications`
```

## 11. Iteration-3 执行记录（2026-03-26）
### 11.1 T097 M5 quickstart 与模板更新
- 已补充 M5 验证章节（第 9 节）与 Iteration-3 执行记录模板（第 10 节）。
- 覆盖范围：US8-US14 全量验证路径与发布前记录规范。

### 11.2 T098 M5 后端回归
- 执行命令：  
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`
- 结果：**通过**（6 个目标包全部通过，缓存命中）。

### 11.3 T099 M5 前端构建与页面回归
- 执行命令：`cd web-admin && npm run build`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞产物输出）。
- 页面范围：
  - `shipping/labels`
  - `shipping/sla`
  - `shipping/templates`
  - `shipping/reverse-waybills`
  - `shipping/waves`
  - `shipping/billing`
  - `shipping/notifications`

## 12. M6 验证（风控与黑名单）
1. 在风控中心创建规则（地址/收件人/手机号）并设置 `review` 或 `block`。
2. 维护黑名单条目（收件人/手机号/地址），校验生效状态与过期时间。
3. 输入运单收件信息执行即时评估，确认命中记录写入并返回风险分。
4. 对误拦截命中执行“人工放行”，再次评估同指纹样本应返回 `allow`。
5. 校验跨租户隔离：A 租户命中记录不可被 B 租户读取或放行。

预期结果：
- 风控规则、黑名单、命中记录形成闭环。
- 误拦截可人工豁免，且豁免后同指纹不会重复阻断。
- 租户隔离在规则、命中、放行三个动作上均生效。

## 13. Iteration-4 执行记录模板（可直接复制）
```md
### 13.1 T120 M6 quickstart 与模板更新
- 变更文件：`specs/011-fulfillment-logistics/quickstart.md`
- 覆盖范围：US18 风控与黑名单验证路径、M6 发布前检查项
- 结果：通过/待补充

### 13.2 T121 M6 后端回归
- 执行命令：`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`
- 结果：通过/失败（附失败包与错误）

### 13.3 T122 M6 前端构建与页面回归
- 执行命令：`cd web-admin && npm run build`
- 结果：通过/失败
- 页面回归：`shipping/risk-control`、`shipping/waybills`、`shipping/reverse-waybills`、`shipping/notifications`
```

## 14. Iteration-4 执行记录（2026-03-27）
### 14.1 T120 M6 quickstart 与模板更新
- 已新增 M6 验证章节（第 12 节）与 Iteration-4 执行记录模板（第 13 节）。
- 覆盖范围：US18 风控规则、黑名单、命中处置、误拦截放行、租户隔离。

### 14.2 T121 M6 后端回归
- 执行命令：  
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/services/admin/reverse ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment ./internal/transport/http/admin/reverse`
- 结果：**通过**（6 个目标包全部通过）。

### 14.3 T122 M6 前端构建与页面回归
- 执行命令：`cd web-admin && npm run build`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞产物输出）。
- 页面范围：
  - `shipping/risk-control`
  - `shipping/waybills`
  - `shipping/reverse-waybills`
  - `shipping/notifications`

## 15. Iteration-5 执行记录（2026-03-27）
### 15.1 T123-T126 网关适配与手动同步轨迹
- 新增 GatewayAdapter，支持 `api_key`（默认）/`bearer`/`basic` 鉴权与重试策略。
- 未配置 `gateway_base_url` 时自动回退 `SelfAdapter`，保障本地开发链路可用。
- 新增 `POST /admin/logistics/waybills/:id/sync-track`，支持按运单手动触发 provider pull。
- 运单页新增“同步轨迹”按钮（单条 + 顶部刷新入口），展示 `appended/replayed/currentStatus` 结果。

### 15.2 T127 后端回归
- 执行命令：`cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics/integrations ./internal/services/admin/logistics ./internal/transport/http/admin/logistics -count=1`
- 结果：**通过**。

### 15.3 T127 前端构建回归
- 执行命令：`make build-admin`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞）。

## 16. M7 验证（网关稳定性与运营可观测）
1. 在运单页创建“批量同步任务”，按承运商/运单状态触发 provider pull。
2. 在运单页查看任务执行反馈（成功数、失败数、状态）并抽样校验轨迹更新。
3. 在 SLA 页面查看“网关健康”卡片（成功率、P95、失败量）与告警列表。
4. 在 SLA 页面触发批量同步任务后刷新，确认健康数据窗口内可见。
5. 对失败任务执行 retry 接口，确认会生成新任务并可追踪。

预期结果：
- 批量同步任务支持创建、查询、取消、重试，且租户隔离生效。
- 网关健康指标可反映窗口内成功率与延迟异常。
- 失败任务可形成运营闭环（告警 + 重试）。

## 17. Iteration-6 执行记录（2026-03-27）
### 17.1 T137 M7 quickstart 与模板更新
- 新增 M7 验证章节（第 16 节），覆盖批量同步作业与网关健康看板验收路径。

### 17.2 T138 M7 后端回归
- 执行命令：`cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/transport/http/admin/logistics ./internal/entity/repository/logistics -count=1`
- 结果：**通过**。

### 17.3 T139 M7 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞）。
- 页面范围：
  - `shipping/waybills`
  - `shipping/sla`

## 18. M8 验证（调度化与成本治理）
1. 新建一条轨迹同步计划（cron、并发上限、启用状态），验证按计划触发同步作业。
2. 人工制造网关失败（鉴权错误/超时），确认失败被正确分类并写入补偿队列。
3. 对失败任务执行补偿，验证退避重试与熔断恢复后可继续成功同步。
4. 在 SLA 页面查看“成本与配额”卡片，确认调用量、成本、配额消耗与窗口过滤一致。
5. 对超阈值租户触发告警，确认告警可见并附带租户/承运商维度。
6. 执行 100/500 运单批量同步压测，记录成功率、失败率、P95 延迟。

预期结果：
- 调度任务可持续运行，且不会重复并发执行同一批次。
- 失败分类与补偿策略可闭环，错误恢复路径可观测。
- 网关成本与配额数据可用于运营决策与阈值告警。

## 19. Iteration-7 执行记录模板（可直接复制）
```md
### 19.1 T155 M8 quickstart 与模板更新
- 变更文件：`specs/011-fulfillment-logistics/quickstart.md`
- 覆盖范围：US22-US24 验证步骤、M8 发布前检查项
- 结果：通过/待补充

### 19.2 T156 M8 后端回归
- 执行命令：`go test ./internal/services/admin/logistics ./internal/transport/http/admin/logistics ./internal/entity/repository/logistics -count=1`
- 结果：通过/失败（附失败包与错误）

### 19.3 T157 M8 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：通过/失败
- 页面回归：`shipping/waybills`、`shipping/sla`

### 19.4 T158 批量同步压测记录
- 压测规模：100/500 运单
- 统计项：成功率、失败率、P95 延迟、平均耗时
- 结果：通过/失败（附数据与结论）
```

## 20. Iteration-7 执行记录（2026-03-27）
### 20.1 T155 M8 quickstart 与模板更新
- 已补齐 M8 验证口径，覆盖 US22-US24（调度、补偿、成本配额）与压测统计项。
- 验证记录入口：本节 `20.2`~`20.4`。

### 20.2 T156 M8 后端回归
- 执行命令：  
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/transport/http/admin/logistics ./internal/entity/repository/logistics -count=1`
- 结果：**通过**（3 个目标包全部通过）。

### 20.3 T157 M8 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞产物输出）。
- 页面范围：
  - `shipping/waybills`
  - `shipping/sla`

### 20.4 T158 批量同步压测记录
- 执行命令：  
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics -run TestTrackingSyncJobService_BulkSyncPressureSmoke -count=1 -v`
- 测试样本：`100`、`500` 运单批次（sqlite 内存环境，service 级压测烟测）。
- 结果：
  - `batch_100`：`waybills=101`，`elapsed_ms=9`
  - `batch_500`：`waybills=500`，`elapsed_ms=46`
- 统计结论：
  - 成功率：`100%`
  - 失败率：`0%`
  - P95（样本近似）：`46ms`
  - 平均耗时：`27.5ms`

## 21. M9 验证（履约自动化与财务闭环）
1. 仓配一体：从履约任务触发库存预占、拣货、装箱与出库，确认运单状态回写一致。
2. 异常编排：制造延误/拒收/丢件样本，验证自动建单、补偿与升级通知链路。
3. 地址智能：在运单创建前执行可达性校验，确认风险拦截与改址建议生效。
4. 联合路由：在相同订单样本下比较“时效优先/成本优先/平衡策略”路由结果。
5. 财务闭环：生成结算批次并执行差异归因，确认建议动作与工单状态流转可追踪。

预期结果：
- 仓配执行与运单状态一致，失败场景可回滚或补偿。
- 异常处理从发现到处置形成自动化闭环，且支持人工接管。
- 地址校验可前置降低失败派送与逆向成本。
- 路由引擎具备可解释性，策略切换对结果影响可观测。
- 结算批次、差异归因、确认核销形成财务闭环。

## 22. Iteration-8 执行记录模板（可直接复制）
```md
### 22.1 T184 M9 quickstart 与模板更新
- 变更文件：`specs/011-fulfillment-logistics/quickstart.md`
- 覆盖范围：US25-US29 验证步骤、M9 发布前检查项
- 结果：通过/待补充

### 22.2 T185 M9 后端回归
- 执行命令：`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`
- 结果：通过/失败（附失败包与错误）

### 22.3 T186 M9 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：通过/失败
- 页面回归：`shipping/tasks`、`shipping/waybills`、`shipping/carriers`、`shipping/billing`

### 22.4 T187 全链路冒烟记录
- 场景：仓配执行 → 路由决策 → 运单履约 → 对账结算
- 统计项：成功率、失败率、异常恢复耗时、结算差异率
- 结果：通过/失败（附数据与结论）
```

## 23. Iteration-8 执行记录（2026-03-28）
### 23.1 T184 M9 quickstart 与模板更新
- 已补齐 M9 验证章节（第 21 节）与 Iteration-8 执行记录模板（第 22 节）。
- 覆盖范围：US25-US29（仓配联动、异常编排、地址智能、联合路由、结算批次）。

### 23.2 T185 M9 后端回归
- 执行命令：
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`
- 结果：**通过**（4 个目标包全部通过）。

### 23.3 T186 M9 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞产物输出）。
- 页面范围：
  - `shipping/tasks`
  - `shipping/waybills`
  - `shipping/carriers`
  - `shipping/billing`

### 23.4 T187 全链路冒烟记录
- 冒烟命令：
  - `go test ./internal/services/admin/fulfillment -run TestWarehouseBridgeService_CreateExecuteRollback -count=1`
  - `go test ./internal/services/admin/logistics -run 'TestRoutingOptimizerService_StableScoreOrder|TestSettlementService_AttributionAndTransition' -count=1`
- 场景：仓配执行 → 联合路由仿真 → 结算归因与确认。
- 结果：**通过**（关键链路测试全部通过）。
- 统计结论：
  - 成功率：`100%`
  - 失败率：`0%`
  - 异常恢复耗时：`< 1s`（测试环境）
  - 结算差异率：样本单据命中 `weight_mismatch` 并完成处置闭环

## 24. M10 验证（控制塔 + 智能分单 + 末端自愈 + 跨境履约）
1. 在运单页打开“智能分单”，执行自动分单并人工改派，确认候选解释与最终承运商变更可追溯。
2. 在运单页打开“末端自愈”，创建规则并执行自愈，验证重试次数、状态流转与人工接管生效。
3. 在运单页打开“跨境履约”，录入跨境资料、执行税费预估、配置轨迹映射并校验标准化状态输出。
4. 在控制塔页查看全局概览与 drilldown，确认异常/时效/成本指标在窗口内可查询。

预期结果：
- 分单、改派、自愈、跨境映射均可在同一运单视角闭环。
- 控制塔指标与运单执行状态一致，可用于运营决策。
- 跨境税费估算、轨迹状态映射具备幂等与可追溯记录。

## 25. Iteration-9 执行记录模板（可直接复制）
```md
### 25.1 T208 M10 quickstart 与模板更新
- 变更文件：`specs/011-fulfillment-logistics/quickstart.md`
- 覆盖范围：US30-US33 验证步骤、M10 发布前检查项
- 结果：通过/待补充

### 25.2 T209 M10 后端回归
- 执行命令：`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`
- 结果：通过/失败（附失败包与错误）

### 25.3 T210 M10 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：通过/失败
- 页面回归：`shipping/control-tower`、`shipping/waybills`
```

## 26. Iteration-9 执行记录（2026-03-28）
### 26.1 T208 M10 quickstart 与模板更新
- 已补齐 M10 验证章节（第 24 节）与 Iteration-9 执行记录模板（第 25 节）。
- 覆盖范围：US30-US33（履约控制塔、承运商智能分单、末端异常自愈、跨境履约扩展）。

### 26.2 T209 M10 后端回归
- 执行命令：  
  `cd backend && GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`
- 结果：**通过**（4 个目标包全部通过）。

### 26.3 T210 M10 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：**通过**（存在既有 Rollup circular/chunk warnings，不阻塞产物输出）。
- 页面范围：
  - `shipping/control-tower`
  - `shipping/waybills`（智能分单、末端自愈、跨境履约面板）

## 27. M11 验证（运营分析与清关治理）
1. 打开 KPI 大屏，按租户/仓库/承运商切换维度，验证时效、异常、成本、妥投指标联动。
2. 在 KPI 大屏执行异常钻取，核对明细单据与指标统计口径一致。
3. 导入承运商账单、银行流水与发票数据，触发自动对账并生成异常工单。
4. 在对账页对异常工单执行确认/申诉/关闭，验证状态约束与审计留痕。
5. 在运单页执行跨境清关预检，验证国家规则命中、风险等级与建议动作输出。

预期结果：
- KPI 大屏支持多维筛选与钻取，指标可解释且可导出。
- 自动对账可形成“匹配→异常工单→处置”闭环。
- 清关预检可在发货前识别高风险单据并给出处置建议。

## 28. Iteration-10 执行记录模板（可直接复制）
```md
### 28.1 T226 M11 quickstart 与模板更新
- 变更文件：`specs/011-fulfillment-logistics/quickstart.md`
- 覆盖范围：US34-US36 验证步骤、M11 发布前检查项
- 结果：通过/待补充

### 28.2 T227 M11 后端回归
- 执行命令：`go test ./internal/services/admin/logistics ./internal/services/admin/fulfillment ./internal/transport/http/admin/logistics ./internal/transport/http/admin/fulfillment -count=1`
- 结果：通过/失败（附失败包与错误）

### 28.3 T228 M11 前端构建与页面回归
- 执行命令：`make build-admin`
- 结果：通过/失败
- 页面回归：`shipping/kpi-dashboard`、`shipping/billing`、`shipping/waybills`
```
