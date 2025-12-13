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
   - `/v1/channels/{id}` 详情 & KPI  
   - `/v1/channels/{id}/credentials` 授权  
   - `/v1/channels/{id}/alerts` 告警  
   - `/v1/channels/{id}/sync` 手动刷新  
   所有请求带 `Authorization: Bearer <tenant JWT>` 与 `X-PowerX-Tenant`.
6. **Background jobs**  
   - `jobs/channel_master/credential_checker.go` 每 12h 检测凭证有效期  
   - `jobs/channel_master/metric_sync.go` 对接任务中心刷新 KPI

## Frontend (web-admin) Workflow
1. **Install & run**  
   ```bash
   cd web-admin
   npm install
   npm run dev   # 默认 http://localhost:3000/_p/<plugin-id>/admin
   ```
2. **Runtime config**  
   - `NUXT_PUBLIC_API_BASE_URL=http://localhost:8086/v1`（直连）  
   - 生产由宿主注入 `/_p/<plugin-id>/api/v1`.
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
