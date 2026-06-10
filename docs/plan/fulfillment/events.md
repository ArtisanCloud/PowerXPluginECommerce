# 履约事件驱动与队列适配（PowerXPlugin 对齐版）

> 目标：与 PowerXPlugin framework 口径保持一致，在 **standalone** 与 **host** 两种模式下统一 Task/Event/WS Bus 语义，禁止页面级轮询兜底。

## 0. 对齐口径（PowerXPlugin 最新）
- 租户来源统一：宿主模式由 PowerX STS 鉴权上下文解析；local+proxy 开放接口模式由 API key 鉴权上下文解析，业务请求不显式传 `tenant_uuid`。
- 模式决策统一：`IAMMode` + `POWERX_PROXY` 必须可解释、可观测（2x2）。
- 当 `IAMMode=local && POWERX_PROXY=1`：
  - 插件 IAM 走本地；
  - WS/能力出站走宿主；
  - 出站鉴权使用 PowerX API key，不透传插件入站本地 token。
- 启动日志必须打印模式决策矩阵，不可仅输出单字段 `mode=local`。

## 0.1 WS Bus 标准链路
- 插件 WS 入口（客户端）：`GET /api/ws`（standalone）；host 透传口径由宿主统一。
- Topic 标准顺序：先 `create topic`，再 `grant`，最后 `publish/subscribe`。
- 插件内部调试入口（topic 创建）：
  - `POST /api/v1/admin/runtime/internal/event-fabric/topics`
- 插件内部调试入口：
  - `POST /api/v1/admin/runtime/internal/ws-bus/grant`
  - `POST /api/v1/admin/runtime/internal/ws-bus/publish`
- framework HostClient 统一 internal 路由：
  - `POST /api/v1/internal/event-fabric/topics`
  - `POST /api/v1/internal/ws-bus/grant`
  - `POST /api/v1/internal/ws-bus/publish`

## 0.2 前端消费约束（强制）
- 页面状态更新必须由 WS Bus topic 驱动，不允许 `setInterval` 轮询任务/进度。
- 默认订阅 topic 至少包括：
  - `task.progress`
  - `powerx.task.progress.v1`
  - `worker.task.updated`
- 业务可追加域 topic（例如 `org_sync.progress`），但不得破坏统一单通道策略。

## 1. 运行模式
### 1.1 Standalone 模式
- 事件总线：插件统一 WS Bus（`/api/ws`） + taskbus bridge
- 队列：framework task 接口（若宿主不可用，允许短期本地兼容）
- 优点：部署简单、无外部依赖

### 1.2 Host 模式
- 事件总线：PowerX Framework 统一事件接口
- 队列：PowerX 框架调度（可对接 Kafka / Redis / MQ）
- 优点：可扩展、跨服务

## 2. 统一事件接口设计
```go
// 伪代码
interface EventBus {
  Publish(topic string, payload any) error
  Subscribe(topic string, handler func(event any)) error
}
```

## 3. 队列任务接口设计
```go
interface JobQueue {
  Enqueue(name string, payload any) (jobID string, err error)
  Register(name string, handler func(job any) error)
}
```

## 4. 事件类型（履约域）
- `order.created`
- `order.paid`
- `inventory.locked`
- `inventory.released`
- `inventory.deducted`
- `fulfillment.task.created`
- `fulfillment.task.completed`
- `waybill.created`
- `waybill.tracking.updated`
- `reverse.waybill.created`

## 5. 模式切换策略
- 配置项：`IAMMode`、`POWERX_PROXY`、`PX_GATEWAY_BASE_URL`、`PX_GATEWAY_API_KEY`、`POWERX_STS_CLIENT_ID`、`POWERX_STS_CLIENT_SECRET`
- `standalone`：允许本地运行，但事件与任务接口保持 framework 同构。
- `host`：通过 PowerX Framework 注入统一实现。
- 最小环境变量集：
  - local+proxy 必须：`POWERX_PROXY=1`、`IAMMode=local`、`PX_GATEWAY_BASE_URL`、`PX_GATEWAY_API_KEY`
  - host 必须：`POWERX_PROXY=1`、`IAMMode=delegated`、`PX_GATEWAY_BASE_URL`、`POWERX_STS_CLIENT_ID`、`POWERX_STS_CLIENT_SECRET`
  - 可选：`POWERX_INTERNAL_ROUTES`、`POWERX_AUTH_OPTIONAL`
  - 已废弃为决策输入：`PX_TENANT_UUID`

## 6. 幂等与一致性
- 每个事件携带 `event_id` 与 `source_id`
- 事件处理器先查 `event_id` 是否已处理（幂等）
- 失败重试：由队列层统一控制

## 7. 示例流程
### 7.1 订单支付成功
1) Payment 服务发出 `order.paid`
2) Inventory 服务消费事件 → 扣减库存
3) Fulfillment 服务消费事件 → 生成履约任务

## 8. 未来扩展
- 接入 Kafka/RabbitMQ
- 事件审计与回溯（event store）
