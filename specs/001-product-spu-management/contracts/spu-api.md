# SPU Admin API Contract

所有接口通过插件宿主反代 `/_p/<plugin-id>/api/v1` 暴露，需携带租户 JWT（含 `tenant_uuid`）。响应遵循 JSON，对象字段以 camelCase 命名。

## 1. 列表 SPU
`GET /api/v1/products/spus`

### 查询参数
| 参数 | 类型 | 说明 |
| ---- | ---- | ---- |
| keyword | string | 编码/标题模糊匹配 |
| categoryId | string | 类目过滤 |
| brandId | string | 品牌 |
| status | enum | draft/reviewing/published/offboarded |
| channel | string | 指定渠道可售性 |
| type | enum | 商品类型 |
| owner | string | 负责人 |
| tags | string[] | 标签（AND） |
| updatedFrom/updatedTo | ISO8601 | 更新时间区间 |
| page/pageSize | int | 默认 1/20 |
| sort | string | 如 `-updatedAt`, `name` |

### 响应
```json
{
  "data": [
    {
      "id": "spu-123",
      "code": "SKU123",
      "name": "多语言标题",
      "type": "subscription",
      "status": "reviewing",
      "category": {"id": "cat-1", "name": "家电"},
      "brand": {"id": "brand-1", "name": "PX"},
      "tags": ["hot"],
      "channels": [{"channel": "official", "availability": "published"}],
      "updatedAt": "2025-12-10T10:00:00Z",
      "owner": "ops-01"
    }
  ],
  "meta": {"total": 234, "page": 1, "pageSize": 20}
}
```

## 2. 创建/更新 SPU 草稿
`POST /api/v1/products/spus`
`PATCH /api/v1/products/spus/{id}`

```json
{
  "code": "SPU-1001",
  "type": "subscription",
  "categoryId": "cat-1",
  "brandId": "brand-2",
  "defaultLocale": "zh-CN",
  "locales": [
    {"locale": "zh-CN", "title": "中文标题", "description": "..."},
    {"locale": "en-US", "title": "English", "description": "..."}
  ],
  "attributes": [{"name": "型号", "value": "Pro"}],
  "media": [{"type": "image", "url": "https://...", "isPrimary": true}],
  "subscriptionPlans": [
    {
      "planCode": "monthly",
      "billingCycle": "monthly",
      "price": 99.0,
      "trialDays": 14,
      "autoRenew": true,
      "effectScope": "new_only"
    }
  ],
  "channels": [
    {
      "channel": "official",
      "availability": "scheduled",
      "publishAt": "2025-12-15T00:00:00Z",
      "contentOverride": {"title": "渠道标题"}
    }
  ]
}
```
- 创建返回 `{ "id": "spu-123", "versionId": "ver-1" }`。
- PATCH 在草稿/审核退回状态下允许更新；如处于发布流程需先创建新版本。

## 3. 提交审核 / 审批
- `POST /api/v1/products/spus/{id}/submit` – 将当前草稿版本置为 `reviewing`，生成审批记录。
- `POST /api/v1/products/spus/{id}/versions/{versionId}/approve`
  ```json
  { "action": "approve", "comment": "合规" }
  ```
- `POST /api/v1/products/spus/{id}/versions/{versionId}/reject`
  ```json
  { "action": "reject", "comment": "缺少授权" }
  ```
- `POST /api/v1/products/spus/{id}/versions/{versionId}/rollback`
  ```json
  { "targetVersionId": "ver-2", "reason": "线上 bug" }
  ```

## 4. 发布 / 下架
- `POST /api/v1/products/spus/{id}/publish`
  ```json
  { "versionId": "ver-3", "channels": ["official","market-a"], "publishMode": "immediate" }
  ```
- `POST /api/v1/products/spus/{id}/withdraw`
  ```json
  { "channel": "official", "withdrawAt": "2025-12-20T00:00:00Z" }
  ```

## 5. 版本列表 / 差异
`GET /api/v1/products/spus/{id}/versions`
- 支持查询参数 `status`, `page`, `pageSize`。
`GET /api/v1/products/spus/{id}/versions/{versionId}` – 返回版本 payload 与 diff。

## 6. 渠道配置
`POST /api/v1/products/spus/{id}/channels`
```json
{
  "channel": "market-a",
  "availability": "scheduled",
  "publishAt": "2025-12-18T00:00:00Z",
  "contentOverride": {"title": "渠道标题", "media": ["s3://..."]}
}
```
- `DELETE /api/v1/products/spus/{id}/channels/{channel}` – 移除渠道或标记为 unlisted。

## 7. 订阅计划
`POST /api/v1/products/spus/{id}/subscription-plans`
```json
{
  "planCode": "annual",
  "name": "年付",
  "billingCycle": "yearly",
  "price": 999,
  "trialDays": 30,
  "autoRenew": true,
  "effectScope": "new_and_existing"
}
```
`PATCH /api/v1/products/spus/{id}/subscription-plans/{planId}` – 更新 `price`、`effectScope`（默认 `new_only`，如改为 `new_and_existing` 需附加 `approvalId`）。
`DELETE` 置 `status=archived`。

## 8. 批量导入/导出
- `POST /api/v1/products/spus/import`
  - `multipart/form-data`，字段 `file`, `templateId`。
  - 响应 `{ "taskId": "job-123" }`。
- `POST /api/v1/products/spus/export`
  ```json
  { "filters": {"status": "published"}, "fields": ["code","name","type","channels"] }
  ```
- 任务进度：`GET /api/v1/jobs/{taskId}` 返回 `status`, `successRows`, `failedRows`, `failedReportUrl?`。

## 9. 审计日志
`GET /api/v1/products/spus/{id}/audit`
- 支持过滤 `eventType`, `operator`, `dateRange`；分页返回最近操作。

## 10. 统计 KPI
`GET /api/v1/products/spus/metrics`
- 返回 `{ "inSale": 1200, "pendingReview": 35, "offboarded": 100, "channelCoverage": {"official": 800, "market-a": 500} }` 用于概览卡片。
