# Channel Master Runbook

本指南面向运营与实现团队，汇总渠道中心的常见操作及排障入口。

## 核心操作流

1. **入驻与审批**
   - 通过 `/v1/channels` 创建草稿，补齐联系人/负责人信息；
   - `POST /channels/{id}/submit` 进入待审批状态，由审批人打开 `/channels/approval` 页面完成审核；
   - 审批通过后状态为 `unauthorized`，待完成授权。
2. **授权与凭证**
   - 详情页的“授权凭证”抽屉会调用 `/channels/{id}/credentials`；
   - 支持 OAuth/API Key/线下凭证，线下模式需提供附件 URL；
   - 巡检失败或即将到期会创建 `channel_alerts` 记录，并在 UI 中以时间线呈现。
3. **策略与团队**
   - “策略与团队”板块绑定 `/channels/{id}/strategy` API；
   - 仅拥有 `channel.strategy.manage` 权限的角色可编辑；
   - Pricebook/库存/物流/客服 SLA、手续费、账期、付款条款等字段保存在 `channel_configs` 表，所有更新会写入审计日志。
4. **KPI、任务与同步**
   - `/channels/{id}` 返回 KPI、健康度、任务关联、备注与同步历史；
   - `POST /channels/{id}/sync` 可触发一次手动同步，执行结束后记录结果与耗时；
   - 任务/备注 API 会在 `channel_task_links`、`channel_notes` 表存档，便于排障。

## 观测与排障

- **审批耗时**：`channel_master.Metrics` 汇总 `AverageApprovalLead`，若 SLA 超标需检查审批流程或 RBAC 设置；
- **凭证覆盖率**：`CredentialCoverage` 值 < 1 表示仍有渠道未完成授权，可从 `/channels?status=unauthorized` 过滤；
- **同步成功率**：`SyncSuccessRate` 来源于手动/自动同步结果，低于 0.99 时需查看同步任务日志；
- **KPI 加载耗时**：`KPILoadHistogram` 捕捉 0~1s、1~2s、2~3s、>3s 桶，若 `Over3s` 增长，可启用 `web-admin/app/plugins/perf.client.ts` 输出日志进一步定位。

更多指标含义及接入 Prometheus 的方法请参考 `docs/observability/channel/master.md`。
