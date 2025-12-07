# Customers API Contract

All endpoints are served via `runtimeConfig.public.apiBaseUrl` (e.g. `/_p/<plugin-id>/api/v1`). Requests必须携带租户 JWT；响应遵循 JSON:API 风格。

## 1. 列表 /customers

`GET /api/customers`

### Query 参数
| 参数 | 类型 | 说明 |
|------|------|------|
| keyword | string | 姓名/邮箱/手机号模糊匹配 |
| tier | string | 会员等级 |
| source | string | 来源渠道 |
| type | enum | individual/enterprise |
| tags | string[] | 多标签 AND 过滤 |
| region | string | 地区 |
| riskLevel | string | 风险等级 |
| createdFrom / createdTo | ISO8601 | 注册日期区间 |
| page / pageSize | int | 分页（默认 1 / 20） |
| sort | string | e.g. `-createdAt`, `lastOrderAt` |

### 响应
```json
{
  "data": [Customer],
  "meta": {
    "total": 12345,
    "savedViewId": "optional"
  }
}
```

## 2. 详情 /customers/{id}
`GET /api/customers/{id}`
- 返回 Customer + 最近订单、售后、权益、审计摘要。

## 3. 创建 /customers
`POST /api/customers`
```json
{
  "name": "张三",
  "type": "individual",
  "email": "foo@example.com",
  "phone": "+86****",
  "source": "website",
  "tags": ["VIP"],
  "membershipTier": "gold",
  "accountManager": "ops-01"
}
```

## 4. 更新 /customers/{id}
`PATCH /api/customers/{id}` – 局部更新；支持标签、等级、负责人、备注。

## 5. 批量动作 /customers/bulk-actions
`POST /api/customers/bulk-actions`
```json
{
  "action": "assign-owner",
  "ids": ["cust-1","cust-2"],
  "payload": {"owner": "ops-02"}
}
```
返回 `{ "taskId": "job-123" }`，任务中心 `/api/jobs/{taskId}` 提供进度。
可选 action：`add-tags`, `remove-tags`, `assign-owner`, `bulk-remind`, `bulk-disable`, `bulk-enable`。

## 6. 导入 /customers/import
`POST /api/customers/import`
- Content-Type: `multipart/form-data`; 字段 `file`。
- 响应 `{ "taskId": "job-456" }`。

## 7. 导出 /customers/export
`POST /api/customers/export`
```json
{
  "filters": { ... 同列表参数 ... },
  "fields": ["id","name","email","membershipTier"]
}
```
- 响应 `{ "taskId": "job-789" }`；任务完成后通过通知中心提供签名下载链接。

## 8. 会员视角 /customers/members
`GET /api/customers/members`
- Query：`tier`, `growthRange`, `pointsRange`, `benefitStatus`, `retentionStatus`, `page`, `pageSize`。
- 响应数据包含 `membershipSnapshot`。

## 9. 会员操作 /customers/{id}/membership
`POST /api/customers/{id}/membership`
```json
{
  "action": "adjust-growth",
  "amount": 50,
  "reason": "Manual Compensation"
}
```
- action 支持 `adjust-growth`, `adjust-points`, `grant-benefit`。

## 10. 保级提醒 /customers/bulk-remind
`POST /api/customers/bulk-remind`
```json
{
  "ids": ["cust-1","cust-2"],
  "channel": "sms",
  "templateId": "retain-tier-1"
}
```
- 返回 `{ "taskId": "job-901" }`。

## 11. 任务中心 /jobs/{id}
- `GET /api/jobs/{id}`：返回 `{status,message,downloadUrl?}`；downloadUrl 仅在导出时出现且需短期签名。

## 12. 审计
- 前端在请求头添加 `X-Audit-Action`, `X-Audit-Resource`；后端写入 `admin_console_audit_events`。
