# Research Notes - Channel Master Data & Authorization

## Decision 1: Tenant-scoped Channel Storage
- **Decision**: `channel_*` 表全部含 `tenant_uuid`（UUID 字符串），Repository 在 `BeginTenantTx` 中 `SET LOCAL app.tenant_uuid` 并启用 Postgres RLS；同一店铺 ID 可在不同租户重复存在。  
- **Rationale**: 宪章要求零信任 + RLS，且渠道数据属于租户资产；以租户字段做联合唯一索引即可避免跨租户串数据。  
- **Alternatives considered**: 
  - 全局唯一 Channel + 租户映射表：增加额外 join 且难以隔离数据血缘。  
  - 单库多 schema：违反插件仅用 `powerx_plugin_base` 的约束。  

## Decision 2: Envelope Encryption for Credentials
- **Decision**: 凭证密文字段使用 envelope encryption（Go `pkg/security/encryption`），数据密钥（DEK）临时生成并使用 STS 提供的 KMS key 加密后存储；存库字段只保留 `ciphertext`, `dek_ciphertext`, `algorithm`, `last_rotated_at`。  
- **Rationale**: 符合宪章“秘钥轮换+最小权限”，同时允许按租户分区轮转；DEK 被泄露也无法直接解密。  
- **Alternatives considered**: 
  - 单层 AES-256 密文：密钥管理复杂且无法做到 per-tenant rotation。  
  - 交由宿主 Connector 统一保管：插件仍需存第三方 token，无法完全依赖宿主。  

## Decision 3: KPI & Health Data Source
- **Decision**: KPI/健康度只读取现有任务中心或数据仓暴露的 `channel_metrics` 视图（经 ETL 汇总）；Channel 服务提供读写 API，但计算任务由 ETL job 定期写入表。  
- **Rationale**: Spec 假设 KPI 已由任务中心产出，本次仅需读取；沿用视图可确保跨渠道指标一致，同时减少实时计算压力。  
- **Alternatives considered**: 
  - 在插件内实时汇总：需要消费大量任务事件且与现有数据仓重复。  
  - 通过外部 API 拉某个平台 KPI：平台差异大，难以提供统一指标。  

## Decision 4: Alerting & Notification Path
- **Decision**: 告警触发由 channel service 写入 `ChannelAlert` 后通过 Observability emitter 推送至宿主通知中心（站内 + Webhook），并在 web-admin 轮询未处理告警；将“数据中断”“凭证将过期”实现为异步 job + alert item。  
- **Rationale**: 利用既有通知通道可统一控制；web-admin 只读 alerts，减少前端状态；异步 job 便于集中处理阈值。  
- **Alternatives considered**: 
  - 前端直接订阅 SSE/WS：宿主代理未必允许长连接。  
  - 插件自行发邮件：绕过宿主审计与偏好中心。  

## Decision 5: Frontend Data Access Strategy
- **Decision**: Nuxt `useChannels` composable 通过 `$fetch` 调用 `/v1/channels` API，所有 baseURL 来自 `runtimeConfig.public.apiBaseUrl`；对 KPI/告警采用并行请求与缓存失效策略（Pinia store 缓存 5 min）。  
- **Rationale**: 遵守宪章代理要求，并行请求 + 轻缓存保证 3s SLA，Pinia 统一状态方便注入健康度标签。  
- **Alternatives considered**: 
  - 通过 server/api 中转：会复制后端 DTO，除非需要 SSR；当前 CSR + API 足够。  
  - 直接调用任务中心 API：破坏 Host Contract，且认证方式不同。  
