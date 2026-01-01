# Quickstart: SKU 与变体管理能力增强

## 前置
- 安装 Go 1.24、Node 20，运行 `npm install` 于 `web-admin`。  
- 准备 Postgres（schema `powerx_plugin_base`）+ Redis，导入基础 SPU/规格数据。  
- 复制 `backend/etc/config.yaml.example` 为局部配置，确保 `POWERX_DEV_MODE=true` 仅限本地。

## 启动
1. **迁移与种子**：`make migrate seed`，确认创建 `product_skus` 及相关表。  
2. **后端服务**：`make dev`（或 `go run ./backend/cmd/plugin -c backend/etc/config.yaml`）启动 HTTP `:8086/v1`。  
3. **前端 Admin**：在 `web-admin` 目录运行 `npm run dev`，浏览器访问 `http://localhost:3000`（宿主反代 `/product/skus`）。

## 核心流程验证

### US1：运营批量生成 SKU
1. 访问 `商品 > SPU > 编辑` 页，在“SKU 生成器”按钮打开 `GeneratorPanel`，勾选多规格组合并填写默认重量/条码前缀。  
2. 点击“生成预览”，确认矩阵中冲突项会高亮并可单独调整；提交后在 `product_skus` 表新增记录。  
3. 返回 `商品 > SKU` 列表或矩阵视图，确认 Pinia store 中新建结果即时刷新，且单规格 SPU 会自动跳过矩阵模式。

### US2：批量调整价格/库存
1. 在 `商品 > SKU` 列表勾选 ≥50 条记录，打开“批量操作”选择价格/库存调整，填写百分比或绝对值。  
2. 提交后观察弹窗反馈 `approval_required`、`scope_size`，并在 “审批抽屉” 里通过/驳回，审核记录会写入任务日志。  
3. 打开 `商品 > SKU > 批量任务` 页面，查看任务状态、失败明细与重试按钮；若需要导入/导出，使用模板上传 CSV 并下载错误报告。  
4. 可通过 `/_p/<plugin-id>/api/v1/products/skus/bulk-tasks` 拉取任务状态，验证后端 `product_sku_bulk_tasks` 表同步更新。

### US3：渠道映射 / 库存可视 / 条码 & 序列
1. 进入某 SKU 详情的“渠道映射”页签，新增渠道 SKU ID、同步策略并点击发布；在“任务日志”查看推送结果与错误回推。  
2. 切换到“库存”页签，确认实时库存、锁定量、在途量与 `last_synced_at` 展示；超 SLA (5 分钟) 时会高亮提示。  
3. 在“条码管理”区生成或校验条码，导出打印模板；在“序列号 / 批次”区录入单号并查阅审计日志。  
4. 后端可在 `backend/internal/observability/product_sku/logger.go` 输出结构化事件，便于排查渠道、库存、条码相关操作。

## 导入/导出与排障
1. 下载前端提供的导入模板，按字段填写价格/库存数据后上传，系统会即时返回校验/错误报告链接。  
2. 导出遵循列表筛选条件，会生成后台任务，完成后可在任务列表或通知栏点击下载。  
3. 如果批量任务长时间运行，可检查后台 `logs/` 或任务表的 `error_report_url` 字段，或在控制台通过 `make integration-smoke` 复现接口链路。

## 测试
- 后端：`make test`，重点覆盖服务层 SKU 生成、批量任务、审批策略、渠道发布、库存同步。  
- 前端：`make test-admin`（聚合 `npm run test:unit` + `npm run test:e2e`），确保批量操作弹窗、生成器交互、渠道映射以及条码/序列管理面板均有可回归用例。  
- 合同：运行 `make integration-smoke` 验证 `/v1/products/skus/**` 接口，必要时添加 Playwright 场景来回归生成器与渠道流程。

## 自动化冒烟脚本
- `web-admin/tests/e2e/product-sku.spec.ts`：驱动浏览器完成 SKU 生成 → 批量调价 → 渠道发布与条码管理，全流程命令 `npm run test:e2e` 或 `make test-admin`。  
- 发布前可结合 `reports/sku-release.md` checklist，记录每次冒烟结果与遗留问题。
