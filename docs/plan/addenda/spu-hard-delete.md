# SPU 硬删除运维指引

当 SPU 已软删除（`deleted_at` 设置）并满足以下条件时，才允许执行硬删除：

1. 软删除时间 ≥ 30 天，确保审计窗口结束。
2. 无任何订单、套餐、营销活动等外部引用（需在对应系统中检查）。
3. 渠道任务、审批、导入导出任务均已完成。

## 操作步骤

1. **备份**：在数据库层面导出 `product_spus`、`product_spu_versions`、`product_spu_channels`、`product_spu_subscription_plans`、`product_spu_audit_logs`、`product_spu_import_tasks`、`product_spu_export_tasks` 等表中目标 SPU 相关数据。
2. **执行脚本**：运行以下命令，传入 `--spu-id` 与 `--tenant-id`：
   ```sh
   go run ./backend/cmd/tools/spu_hard_delete.go \
     --tenant-id=TENANT-UUID \
     --spu-id=SPU-UUID \
     --reason="合规要求"
   ```
   工具会：
   - 校验 soft delete 时间与引用情况；
   - 级联删除版本、渠道、订阅计划、审计、导入导出任务；
   - 将执行记录写入 `spu_hard_delete_audit.log`。
3. **复核**：确认相关渠道看板、任务中心无残留数据；更新运维记录并通知业务负责人。

> 若数据库脚本执行失败，必须立即恢复备份并重新评估，禁止部分删除。
