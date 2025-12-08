# Powerx Plugin Ecommerce Plugin Web Admin

该目录基于 Nuxt 4 与 `@artisan-cloud/plugin-framework-admin` Layer，默认启用 Starter 页面。

```bash
npm install
npm run dev
```

若需要覆盖框架默认页面，可在 `app/pages` 或 `app/components` 下新增同路径文件。

## 开发前提与运行配置

- **API Base**  
  - 宿主模式：设置 `POWERX_PROXY=1` 且 `POWERX_PLUGIN_ID=<插件 ID>`，Nuxt 会自动将 `runtimeConfig.public.apiBaseUrl` 指向 `/_p/<plugin-id>/api/v1`。  
  - Standalone / 本地联调：如需直连后端，可设置 `NUXT_PUBLIC_API_BASE=http://localhost:8078`（或其他网关地址）以及可选的 `NUXT_PUBLIC_API_PREFIX=/api/v1`。
- **STS / 身份交换**  
  - 使用 PowerX STS（`/_p/_internal/sts/exchange`）获取宿主颁发的短期凭证；本地 mock 场景可开启 `POWERX_DEV_MODE=1` 并在 `.env` 中写入测试 token。
- **任务中心入口**  
  - 客户域批量任务复用宿主任务中心，管理员可在 `/_p/<plugin-id>/admin/console/jobs`（或 web-admin 侧“运维控制台”菜单）追踪导入/导出与批量提醒状态；确保宿主实例开启 `jobs` API。
