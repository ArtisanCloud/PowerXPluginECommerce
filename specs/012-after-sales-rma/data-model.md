# Data Model: 售后 RMA 与退货门户

## 1. AfterSaleCase（售后单）
- Purpose: 承载客户发起的退款/退货退款/换货申请主记录。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - case_no (string, unique)
  - order_id (uuid/string)
  - order_item_id (uuid/string)
  - customer_id (string)
  - case_type (enum: refund_only/return_refund/exchange)
  - status (enum: pending/accepted/reviewing/approved/rejected/completed/closed)
  - reason_code (string)
  - reason_detail (string)
  - requested_qty (int)
  - requested_amount_minor (int)
  - currency (string)
  - source_channel (string)
  - created_at / updated_at (timestamp)
  - closed_at (timestamp, nullable)
- Rules:
  - (tenant_uuid, case_no) 唯一。
  - 进行中状态（pending/accepted/reviewing/approved）同一 `(tenant_uuid, order_item_id)` 仅允许一条。
  - 终态（rejected/completed/closed）后可再次申请，但生成新 case_no。

## 2. AfterSaleTimeline（售后轨迹）
- Purpose: 记录售后状态变更与动作轨迹，支撑审计与用户进度展示。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - case_id (uuid)
  - action (enum: create/accept/review/approve/reject/complete/close/comment)
  - from_status (string)
  - to_status (string)
  - operator_type (enum: customer/operator/system)
  - operator_id (string)
  - note (string)
  - created_at (timestamp)
- Rules:
  - 任何状态变化必须落轨迹。

## 3. AfterSaleEvidence（售后凭证）
- Purpose: 记录客户或运营提交的凭证信息。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - case_id (uuid)
  - uploader_type (enum: customer/operator)
  - uploader_id (string)
  - evidence_type (enum: image/video/text/other)
  - content_ref (string)
  - description (string)
  - created_at (timestamp)
- Rules:
  - 终态后禁止修改历史凭证，仅允许新增补充凭证（可选策略）。

## 4. AfterSaleDecision（审核决策）
- Purpose: 存储运营审核结论与决策信息。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - case_id (uuid)
  - decision (enum: approved/rejected)
  - reject_reason_code (string, nullable)
  - decision_note (string)
  - decided_by (string)
  - decided_at (timestamp)
- Rules:
  - 每次审核动作生成一条记录。

## 5. ReturnLogisticsLink（逆向物流关联）
- Purpose: 建立售后单与逆向物流过程的关联。
- Key Fields:
  - id (uuid)
  - tenant_uuid (uuid)
  - case_id (uuid)
  - reverse_waybill_id (uuid/string)
  - reverse_waybill_no (string)
  - carrier_code (string)
  - receive_status (enum: pending/received/rejected)
  - received_at (timestamp, nullable)
  - created_at / updated_at (timestamp)
- Rules:
  - 仅退货退款/换货类型允许关联逆向物流。

## Relationships
- AfterSaleCase 1:N AfterSaleTimeline
- AfterSaleCase 1:N AfterSaleEvidence
- AfterSaleCase 1:N AfterSaleDecision
- AfterSaleCase 0..1:N ReturnLogisticsLink

## State Transitions
- pending -> accepted
- accepted -> reviewing
- reviewing -> approved | rejected
- approved -> completed | closed
- rejected -> closed
- completed -> closed (可选，用于归档)

## Validation Mapping
- FR-002/FR-002A: 创建时校验订单状态、支付状态、售后窗口。
- FR-006: 状态流转必须符合状态机。
- FR-008/FR-008A: 同明细进行中防重，终态后可新建。
- FR-012: 所有查询/修改都按 tenant_uuid 与主体权限约束。
- FR-016: 换货首版不触发自动补发与库存自动占用。
