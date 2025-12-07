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

4. **模拟批量操作**
   - 在 dev 环境设置 `POWERX_DEV_MODE=1` 以使用 mock STS。
   - 通过“批量指派负责人”或“批量提醒”触发任务，查看任务中心和审计面板是否记录。

5. **构建**
   ```bash
   npm run build
   ```
   构建产物位于 `web-admin/.output/`，需随插件打包。
