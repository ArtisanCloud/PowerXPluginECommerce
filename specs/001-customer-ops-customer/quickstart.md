# Quickstart

1. **安装依赖**
   ```bash
   npm install
   ```

2. **运行开发环境**
   ```bash
   npm run dev
   ```
   - 确保 `.env` 指向宿主反代：`NUXT_PUBLIC_API_BASE_URL="/_p/<plugin-id>/api/v1"`。
   - 登录 web-admin 后导航至 `/customer`、`/customer/members` 验证新视图。

3. **测试**
   ```bash
   npm run lint
   npm run test           # Vitest 单测
   npm run test:e2e       # Playwright 或 Cypress（视项目脚本而定）
   ```

4. **KPI / 错误提示验证**
   - **筛选耗时**：在 `/customer` 内连续三次搜索不同条件，观察 `customer_list_fetch` 事件是否在日志中 < 2s。
   - **会员保级提醒**：在 `/customer/members` 筛选“即将降级”，使用提醒抽屉发送一次 mock 任务，确认任务中心与 `customer_reminder_task` 事件记录成功/失败率。
   - **导入/导出**：使用新对话框提交导入、导出任务，确保任务卡片进入“任务中心”并写入审计；导出完成后下载链接可访问且 10 分钟内完成。
   - **错误提示**：断网或修改 API 基础地址触发失败场景，检查 FR-009 要求的 toast/重试按钮是否出现。

5. **构建**
   ```bash
   npm run build
   ```
   构建产物位于 `web-admin/.output/`，需随插件打包。
