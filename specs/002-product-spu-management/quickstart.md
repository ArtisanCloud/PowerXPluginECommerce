# Quickstart

1. **准备环境**
   ```bash
   make deps          # 确保 Go/Node 依赖已安装（如已有可跳过）
   npm install --prefix web-admin
   ```

2. **运行数据库迁移与种子（需要 Postgres + powerx_plugin_base schema）**
   ```bash
   make migrate
   make seed # 如需基础类目/品牌/渠道数据
   ```

3. **启动开发服务**
   ```bash
   make dev           # 运行 backend，监听 :8086 并加载插件 API
   npm run dev --prefix web-admin
   ```
   - 确保 `web-admin/.env` 或 `nuxt.config` 中 `runtimeConfig.public.apiBaseUrl="/_p/<plugin-id>/api/v1"`（或本地独立 `http://localhost:8086/v1`）。
   - 登录 web-admin 后访问 `/product/spus`、`/product/spus/create`、`/product/spus/{id}` 验证列表、向导与详情。

4. **关键流程自测**
   - 新建 SPU：完成向导、提交审批、通过后发布，检查渠道任务与审计。
   - 版本回滚：编辑已发布 SPU，生成草稿，查看 diff 并回滚。
   - 批量导入：下载模板、上传 50+ 行数据，确认任务中心显示成功/失败报告。
     ```bash
     curl -X POST \
       -H "Authorization: Bearer <token>" \
       -F "templateId=default" \
       -F "file=@spu_import.csv" \
       http://localhost:8086/api/v1/admin/product/spus/import
     curl http://localhost:8086/api/v1/jobs/<taskId> -H "Authorization: Bearer <token>"
     ```
   - 批量导出：过滤字段后排队导出并从任务中心下载。
     ```bash
     curl -X POST \
       -H "Authorization: Bearer <token>" \
       -H "Content-Type: application/json" \
       -d '{"fields":["code","name","type","status"],"filters":{"status":"published"}}' \
       http://localhost:8086/api/v1/admin/product/spus/export
     ```
   - 渠道同步：在详情页配置多渠道上架，观察渠道回执展示。
   - 订阅计划：新增计划并测试“仅新订阅”与“新+存量订阅”两种作用范围。

5. **测试**
   ```bash
   make test                    # Go 单元 + 集成本
   make integration-smoke       # 覆盖渠道/任务中心联动
   npm run lint --prefix web-admin -- --max-warnings=0
   npm run test --prefix web-admin
   npm run test:e2e --prefix web-admin   # 如项目已配置 Playwright
   ```

6. **构建与打包**
   ```bash
   make build                   # backend 二进制输出到 backend/bin
   npm run build --prefix web-admin   # Nuxt 产物存于 web-admin/.output
   make package                 # 生成插件发布包（含 backend + web-admin/.output + plugin.yaml）
   ```

7. **验证 KPI**
   - 列表/详情加载：通过浏览器 DevTools 记录 tti，需满足 <2s/<1.5s 指标。
   - 导入/渠道任务：检查任务中心状态及 `admin_console_audit_events`，确认失败报告可下载。
   - 审批 SLA：触发审批流并模拟超时，验证提醒与重新指派逻辑。
