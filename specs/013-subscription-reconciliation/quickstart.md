# Quickstart: 订阅对账与续费治理

## 1. 环境准备

1. 切换分支并确认依赖：`git branch --show-current`（应为 `013-subscription-reconciliation`）
2. 初始化数据库结构：`make migrate`
3. 启动后端服务：`make run`
4. 启动管理端（可选）：`cd web-admin && npm run dev`

## 2. 创建对账批次（净应收口径）

请求：

```bash
curl -X POST "http://localhost:8086/api/v1/admin/subscription-reconciliation/batches" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "billingCycle": "2026-03-31",
    "runType": "daily"
  }'
```

预期：
- 返回 `success=true`
- `data.status=completed`
- 批次金额字段遵循：`expected = bill + tax - discount`

## 3. 查询差异并转处置任务

1. 查询差异：

```bash
curl "http://localhost:8086/api/v1/admin/subscription-reconciliation/batches/<batchId>/deltas?riskLevel=high" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

2. 差异转任务：

```bash
curl -X POST "http://localhost:8086/api/v1/admin/subscription-reconciliation/deltas/<deltaId>/tasks" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "assignee": "ops-001",
    "slaLevel": "high",
    "note": "自动生成高优先任务"
  }'
```

预期：
- 任务创建成功，`status=pending`
- `slaLevel=high` 时截止时间为创建后 24h（中/低分别 48h/72h）

## 4. 执行续费失败治理

请求：

```bash
curl -X POST "http://localhost:8086/api/v1/admin/subscription-reconciliation/governance/run" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "billingCycle": "2026-03-31",
    "dryRun": false
  }'
```

预期：
- 触发递增重试窗口：`1h -> 24h -> 72h -> 7d`
- 返回执行统计（`retryTotal/retrySucceeded/retryFailed/escalated`）
- 生成审计事件：`renewal.retry.executed`

## 5. 运营看板验证

请求：

```bash
curl "http://localhost:8086/api/v1/admin/subscription-reconciliation/dashboard?from=2026-03-01&to=2026-03-31" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

检查项：
- 差异率、恢复率、平均处理时长、待处理积压是否返回
- 指标与批次/任务明细可相互追溯

## 6. 回归验证命令

后端：

```bash
make test
```

前端（如涉及看板页面）：

```bash
make test-admin-ci
```

针对模块：

```bash
cd backend && go test ./internal/services/admin/subscription_reconciliation -run Test -count=1
```

通过标准：
- 单测全部通过
- 对账批次可重复触发但保持幂等
- 同一差异指纹不生成重复活跃任务

## 7. Phase 6 回归清单（Polish）

### 7.1 错误码与用户文案核对

1. 后端返回错误码应落在对账域集合：`RECONCILIATION_*`
2. 前端根据错误码展示用户文案（示例）：
   - `RECONCILIATION_BATCH_NOT_FOUND` -> 未找到对账批次，请刷新后重试
   - `RECONCILIATION_TASK_ALREADY_OPEN` -> 该差异已有未关闭处置任务，请直接跟进现有任务

### 7.2 性能回归（模块级）

对账编排基准：

```bash
cd backend && go test ./internal/services/admin/subscription_reconciliation -run TestReconciliationPerformance_10kBenchmark -count=1
```

看板查询基准：

```bash
cd backend && go test ./internal/services/admin/subscription_reconciliation -run TestDashboardPerformance_P95LessThan1s -count=1
```

### 7.3 示例请求（看板导出）

```bash
curl "http://localhost:8086/api/v1/admin/subscription-reconciliation/dashboard/export?from=2026-03-01&to=2026-03-31&channel=app&plan=pro&region=CN&failureReason=always_fail" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

预期：
- 返回 `text/csv`
- 结果至少包含 `deltaRate/recoveryRate/avgHandleHours/pendingTasks` 四行指标
