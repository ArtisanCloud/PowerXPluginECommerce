# Quickstart: SKU 与变体管理能力增强

## 前置
- 安装 Go 1.24、Node 20，运行 `npm install` 于 `web-admin`。  
- 准备 Postgres（schema `powerx_plugin_base`）+ Redis，导入基础 SPU/规格数据。  
- 复制 `backend/etc/config.yaml.example` 为局部配置，确保 `POWERX_DEV_MODE=true` 仅限本地。

## 启动
1. **迁移与种子**：`make migrate seed`，确认创建 `product_skus` 及相关表。  
2. **后端服务**：`make dev`（或 `go run ./backend/cmd/plugin -c backend/etc/config.yaml`）启动 HTTP `:8086/v1`。  
3. **前端 Admin**：在 `web-admin` 目录运行 `npm run dev`，浏览器访问 `http://localhost:3000`（宿主反代 `/product/skus`）。

## 功能验证
1. 在 SPU 编辑页点击“SKU 生成器”，验证笛卡尔组合勾选、默认值设定。  
2. 在 SKU 列表勾选多条记录，发起“批量调整价格”并观察审批策略是否按阈值触发。  
3. 打开 SKU 详情的“渠道映射”页签，配置渠道 SKU ID 并推送，查看任务日志。  
4. 在“库存”页签检查实时库存数值、锁定量与预警提示时间戳。  
5. 执行导入/导出：通过模板导入 1k 行 SKU 库存，确认 1 分钟内返回校验结果。

## 测试
- 后端：`make test`，重点覆盖服务层 SKU 生成、批量任务、审批策略、渠道发布、库存同步。  
- 前端：`make test-admin` 或 `npm run test`，补齐批量操作弹窗、生成器交互与渠道映射组件单测。  
- 合同：运行 `make integration-smoke` 验证 `/v1/products/skus/**` 接口，必要时添加 Playwright 场景来回归生成器与渠道流程。
