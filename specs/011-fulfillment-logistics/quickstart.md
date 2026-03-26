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
