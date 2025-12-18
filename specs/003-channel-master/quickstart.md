# Quickstart — Channel Master Data & Authorization

## Prerequisites
- Go 1.24、Node 20、npm 10、PostgreSQL ≥ 13、Redis（可选，用于凭证巡检缓存）。
- `POWERX_PLUGIN_ID`, `POWERX_BIND_ADDR`, `POWERX_DEV_MODE`, `POWERX_DB_DSN`、STS 凭证等环境变量已配置。
- `plugin.yaml` 中的 `frontend.admin.basePath`、菜单项指向 `/channels` 入口。

## Backend Development Workflow
1. **Install deps**  
   ```bash
   go env -w GOPRIVATE=github.com/ArtisanCloud && cd backend && go mod tidy
   ```
2. **Run migrations**  
   ```bash
   make migrate   # 创建 channel_masters, channel_credentials, ... 表 + RLS
   ```
3. **Seed reference data (可选)**  
   ```bash
   make seed CHANNEL_TEMPLATES=channels/basic.yaml
   ```
4. **Start dev server**  
   ```bash
   make dev   # 绑定 :8086，自动加载配置
   ```
5. **Expose APIs**  
   - `/v1/channels` 列表 + 筛选  
   - `/v1/channels/{id}` 详情 & KPI（包含 `metrics`, `health`, `strategy`, `team`）  
   - `/v1/channels/{id}/credentials` 授权  
   - `/v1/channels/{id}/alerts` 告警  
   - `/v1/channels/{id}/sync` 手动刷新 & `/sync-history` 查询  
   所有请求带 `Authorization: Bearer <tenant JWT>` 与 `X-PowerX-Tenant`.
6. **Background jobs**  
   - `jobs/channel/master/credential_checker.go` 每 12h 检测凭证有效期并写入 `channel_alerts`  
   - `jobs/channel/master/metric_refresh.go` 读取任务中心视图写入 `channel_metrics`

## Channel Operations
1. **授权与凭证**：在渠道详情页打开“授权凭证”抽屉，可创建 OAuth/API Key/线下凭证；保存后系统自动触发巡检，巡检失败或即将到期会向 `/channels/{id}/alerts` 写入记录。  
2. **KPI 与健康度**：详情页默认拉取 KPI 面板与健康度标签，并在 `metrics_service.go` 计算闸门式得分；若关键指标缺失或巡检异常会追加 `no_metrics`、`credential_*` 标签，加载完成后前端通过 `$perf.logKpiLoad` 记录耗时。  
3. **任务与备注**：通过 `/channels/{id}/tasks` 与 `/channels/{id}/notes` API 维护任务中心关联与运营备注，所有操作写入 `channel_audit_logs`，可在审批/合规场景检索。  
4. **同步历史**：操作员可调用 `/channels/{id}/sync` 触发手动同步，执行结果记录到 `/channels/{id}/sync-history`，并在 UI 中展示最近 10 条记录；`SyncHistoryService.Complete` 会更新 `SyncSuccessRate` 指标。  
5. **策略与团队**：使用 `/channels/{id}/strategy` PATCH 更新 pricebook/库存/物流/客服策略及负责/审批人，RBAC 受 `channel.strategy.manage` 限制，所有写入带审计事件。  
6. **告警闭环**：`/channels/{id}/alerts` 支持前端确认/指派处理人，关闭后会更新 `channel_alerts` 状态并清除 UI 标签。更多操作指引见 `docs/guides/channel_master.md`，指标对照见 `docs/observability/channel/master.md`。

## Frontend (web-admin) Workflow
1. **Install & run**  
   ```bash
   cd web-admin
   npm install
   npm run dev   # 默认 http://localhost:3000/_p/<plugin-id>/admin
   ```
2. **Runtime config**  
   - `NUXT_PUBLIC_API_BASE_URL=http://localhost:8086/v1`（直连）  
   - 生产由宿主注入 `/_p/<plugin-id>/api/v1`; 详情页 KPI 面板需要 `useNuxtApp().$perf.logKpiLoad` 插件启用。
3. **Pages**  
   - `pages/channels/index.vue`: 列表、筛选、批量操作  
   - `pages/channels/[id].vue`: 详情、KPI、告警、策略  
   - `pages/channels/approval.vue`: 审批+历史  
4. **Components/Stores**  
   - `components/channels/ChannelForm.vue`, `ChannelCredentialDrawer.vue`, `ChannelHealthCard.vue`  
   - `stores/channels.ts` 管理分页、选择项、轮询状态  
   - `composables/useChannels.ts` 封装 API（list/detail/authorize/sync）

## Testing & Validation
1. Backend  
   ```bash
   make test
   make integration-smoke TEST_PATTERN=channel_master
   ```
2. Frontend  
   ```bash
   npm run lint -- --max-warnings=0
   npm run test
   ```
3. Contract checks  
   - 通过 Spectral（或 `oapi-codegen --validate`）校验 `specs/003-channel-master/contracts/channel.yaml`
4. Manual verification  
   - 创建→审批→授权→查看 KPI  
   - 伪造过期凭证验证告警  
   - 停用渠道并确认审计日志写入

## Release Checklist
- `make build && make frontend-build`，确保 `web-admin/.output/` 最新。  
- 更新 `plugin.yaml` version + 菜单。  
- `make release && make package` 上传 zip。  
- 附带 `research.md`, `data-model.md`, `quickstart.md`, `contracts/` 产物至 PR。
