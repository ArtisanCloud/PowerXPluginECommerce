# PowerX 插件对齐作战手册（鉴权 / RBAC / 安装 / CI）

适用对象：所有基于 PowerXPlugin skeleton 的插件项目。  
目标：让“功能迭代后即可安装可用、CI 稳定通过、线上链路不跑偏”。

---

## 1. 今日已验证的关键规范（必须统一）

### 1.1 身份鉴权路径规范

- 身份接口统一走：`/api/v1/admin/{identity}/auth/*`
- 当前默认 identity：`user`，即：`/api/v1/admin/user/auth/*`
- 不再使用旧路径：`/admin/auth/*`

### 1.2 插件网关路径边界

- 插件业务 API 走：`/_p/{plugin_id}/api/v1/*`
- 身份认证接口必须走宿主主路由：`/api/v1/admin/{identity}/auth/*`
- 不要把身份接口放到 `/_p/...` 链路里做最终鉴权。

### 1.3 RBAC 资源命名与路由推导一致

- `GET/POST /templates` 这类路由，网关按资源 `template` 推导。
- capability 与 catalog 中的 `rbac.resource` 必须与之一致（`template`），避免 `403 no permission rule`。

---

## 2. plugin.yaml / catalog 必备配置

### 2.1 `plugin.yaml` 必须包含 `migrations`

示例（已在 skeleton 对齐）：

```yaml
migrations:
  driver: go
  entry: backend/bin/migrate
  args: ["setup"]
  workdir: ./backend
  once: true
  timeout: 60s
```

目的：保证安装后可自动执行迁移，避免业务接口因表/字段缺失报 `500`。

### 2.2 单一事实源：`contracts/capabilities -> plugin.d -> runtime`

- `contracts/capabilities/*.yaml` 是能力单一事实源（SoT）。
- `plugin.d/{capabilities,exposure,rbac}.yaml` 是运行时清单产物，供网关放行、路由暴露、RBAC 映射读取，不应手改。
- `plugin.yaml` 顶层 `capabilities` 不应与 `catalogs.capabilities` 并存，避免双源漂移。

---

## 3. 安装与启用：两阶段判定（必须区分）

- `install/local` 返回 200 仅代表“包可落盘”。
- `switch_version(enable=true)` 成功才代表“版本可启用”（受启动/健康检查/运行时约束影响）。
- 排障顺序：先看 install 响应，再看 switch_version 响应，再看插件启动日志与健康探针。

---

## 4. 本地开发标准流程（推荐）

在仓库根目录执行：

```bash
make manifest-align-fix
make skeleton-reinstall VERSION=<new_version> API_BASE=http://127.0.0.1:8077/api/v1 TOKEN=$ADMIN_BEARER_TOKEN
```

说明：

- `manifest-align-fix`：自动同步 plugin.d 并校验 capability→exposure/rbac 映射。
- `skeleton-reinstall`：disable -> force install -> switch version(enable=true)。
- 每次涉及清单、权限、路由契约变更时，请递增版本号再重装。
- `local-install-run` 内置完整性预检：`plugin.d/*`、`contracts/capabilities`、`migrations`、schema 引用文件存在性。

---

## 5. 可执行规则优先级（文档服从代码）

- 最终安装门禁以 `backend/cmd/manifestcheck/main.go` 为准。
- 文档是可读说明，和可执行规则冲突时，以可执行规则为准。
- 对齐建议固定命令：

```bash
make plugin-yaml-check
make manifest-align-check
```

---

## 6. CI 建议接入（必须）

### 6.1 严格门禁

```bash
make manifest-align-check
```

失败即阻断（说明 catalog 漂移或映射不一致）。

### 6.2 devwatch 测试稳定性

`tools/cli/internal/devwatch` 相关测试应显式设置：

- `PX_RESOURCE_CPU_THRESHOLD=101`（测试内设置）

避免 CI 环境 CPU guard 误触发导致超时。

---

## 7. Manifestx 运行时约束（前置校验）

- 权限字符串必须三段式：`domain.resource.action`。
- 必须匹配 manifestx 约束正则；格式错误会在启用阶段触发异常。
- 不要把格式纠错留到线上启用时，开发阶段必须先做 `manifest-align-check` 和 `plugin-yaml-check`。

---

## 8. 故障快速判定

- `403 no permission rule`：先查 capability `rbac.resource/actions` 与 `plugin.d/exposure.yaml` 映射。
- `401 Unauthorized`：先查 token 受众/链路（宿主用户 token vs 插件 token）。
- `500 Internal Server Error`：先查插件后端日志、schema、迁移执行状态。

数据库核对（PostgreSQL）：

```sql
select to_regclass('powerx_plugin_base.template');
select column_name from information_schema.columns
 where table_schema='powerx_plugin_base' and table_name='template';
```

---

## 9. 相关命令清单

- 自动对齐（本地）：`make manifest-align-fix`
- 严格校验（CI）：`make manifest-align-check`
- 清单校验：`make plugin-yaml-check`
- 安装预检（可单独跑）：`make local-install-precheck LOCAL_INSTALL_SRC=dist/<version>`
- 重装插件：`make skeleton-reinstall VERSION=<v> API_BASE=... TOKEN=...`
