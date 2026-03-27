import { apiDel, apiGet, apiPatch, apiPost } from "./_client";
import type { ApiResponse } from "./_base";

type ApiEnvelope<T> = ApiResponse<T> & {
  request_id?: string;
  timestamp?: string;
  [key: string]: any;
};

type RawRecord = Record<string, any>;

const unwrap = async <T>(promise: Promise<ApiEnvelope<T> | T>): Promise<T> => {
  const resp = await promise;
  if (resp && typeof resp === "object" && "data" in (resp as Record<string, any>)) {
    return (resp as ApiEnvelope<T>).data;
  }
  return resp as T;
};

const pick = (raw: RawRecord, camel: string, snake: string) => raw[camel] ?? raw[snake];

const asArray = <T>(value: unknown): T[] => (Array.isArray(value) ? (value as T[]) : []);

export type LogisticsCarrier = {
  id: string;
  name: string;
  code: string;
  type: string;
  status: string;
  contactName: string;
  contactPhone: string;
  capabilities: Record<string, any>;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsTemplate = {
  id: string;
  name: string;
  currency: string;
  status: string;
  version: number;
  channels: any[];
  rules: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsWaybill = {
  id: string;
  orderId: string;
  carrierId: string;
  serviceCode: string;
  waybillNo: string;
  packageNo: number;
  packageKey: string;
  shipmentItems: string[];
  orderItemCount: number;
  orderFulfillmentStatus: string;
  status: string;
  feeAmount: number;
  actualFeeAmount: number;
  feeDiffAmount: number;
  billingStatus: string;
  settledAt: string;
  labelUrl: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsTracking = {
  id: string;
  waybillId: string;
  waybillNo: string;
  eventId: string;
  status: string;
  source: string;
  description: string;
  occurredAt: string;
  payload: Record<string, any>;
  createdAt: string;
};

export type LogisticsWaybillDetail = {
  waybill: LogisticsWaybill;
  tracking: LogisticsTracking[];
};

export type LogisticsWaybillETA = {
  waybillId: string;
  waybillNo: string;
  carrierId: string;
  serviceCode: string;
  timezone: string;
  pickupDeadlineAt: string;
  deliveryDeadlineAt: string;
  promisedAt: string;
  estimatedAt: string;
  delayed: boolean;
  source: string;
  lastComputedAt: string;
};

export type LogisticsRoutingRule = {
  id: string;
  name: string;
  warehouseId: string;
  destinationZone: string;
  carrierId: string;
  serviceCode: string;
  priority: number;
  minWeight: number;
  maxWeight: number;
  minOrderAmount: number;
  maxOrderAmount: number;
  fallback: boolean;
  enabled: boolean;
  ruleConfig: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsRoutingCandidate = {
  carrierId: string;
  carrierName: string;
  serviceCode: string;
  priority: number;
  score: number;
  matchedRule: string;
};

export type LogisticsRoutingPreview = {
  warehouseId: string;
  destinationZone: string;
  carrierId: string;
  carrierName: string;
  serviceCode: string;
  matchedRuleId: string;
  matchedRuleName: string;
  strategy: string;
  fallback: boolean;
  reason: string;
  candidates: LogisticsRoutingCandidate[];
};

export type LogisticsRedeliveryTask = {
  id: string;
  waybillId: string;
  waybillNo: string;
  requestKey: string;
  status: string;
  attemptNo: number;
  addressSnapshot: Record<string, any>;
  lastReason: string;
  operatorId: string;
  closedAt: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsRiskRule = {
  id: string;
  name: string;
  matchField: string;
  matchMode: string;
  pattern: string;
  decision: string;
  riskLevel: string;
  priority: number;
  enabled: boolean;
  description: string;
  ruleConfig: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsBlacklistEntry = {
  id: string;
  entryType: string;
  recipientName: string;
  recipientPhone: string;
  addressLine: string;
  reason: string;
  status: string;
  expiresAt: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsRiskHit = {
  id: string;
  waybillId: string;
  waybillNo: string;
  ruleId: string;
  blacklistId: string;
  source: string;
  decision: string;
  riskLevel: string;
  fingerprint: string;
  description: string;
  status: string;
  releasedBy: string;
  releaseReason: string;
  releasedAt: string;
  payload: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsRiskEvaluateResult = {
  blocked: boolean;
  decision: string;
  score: number;
  releasedBy: string;
  fingerprint: string;
  matchedRules: LogisticsRiskRule[];
  matchedBlacklist: LogisticsBlacklistEntry[];
  hits: LogisticsRiskHit[];
};

export type LogisticsBillingCarrierSummary = {
  carrierId: string;
  carrierName: string;
  waybillCount: number;
  estimatedFee: number;
  actualFee: number;
  diffFee: number;
  abnormalCount: number;
};

export type LogisticsBillingSnapshot = {
  summary: LogisticsBillingCarrierSummary[];
  items: LogisticsWaybill[];
};

export type LogisticsBillingCase = {
  id: string;
  waybillId: string;
  carrierId: string;
  caseNo: string;
  status: string;
  diffAmount: number;
  reason: string;
  resolution: string;
  metadata: Record<string, any>;
  closedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsLabelPrintTask = {
  id: string;
  requestKey: string;
  waybillId: string;
  waybillNo: string;
  status: string;
  attemptCount: number;
  maxAttempts: number;
  lastError: string;
  retryQueuedAt: string;
  printedAt: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsLabelPrintResult = {
  task: LogisticsLabelPrintTask | null;
  idempotencyStatus: string;
  success: boolean;
  message: string;
};

export type LogisticsLabelPrintBatchResult = {
  results: LogisticsLabelPrintResult[];
  success: number;
  failed: number;
};

export type LogisticsSLASummaryItem = {
  carrierId: string;
  carrierName: string;
  waybillCount: number;
  pickupOnTimeCount: number;
  pickupOnTimeRate: number;
  signOnTimeCount: number;
  signOnTimeRate: number;
  exceptionCount: number;
  exceptionRate: number;
  pickupSLAHours: number;
  deliverySLAHours: number;
};

export type LogisticsSLASnapshot = {
  summary: LogisticsSLASummaryItem[];
  total: LogisticsSLASummaryItem;
};

export type LogisticsRateQuoteResult = {
  templateId: string;
  template: string;
  currency: string;
  billingType: string;
  matchedZone: {
    region: string;
    firstMetric: number;
    firstFee: number;
    additionalStep: number;
    additionalFee: number;
    freeThreshold: number;
  };
  feeAmount: number;
  breakdown: Record<string, any>;
};

export type LogisticsNotificationTemplate = {
  id: string;
  name: string;
  event: string;
  channel: string;
  title: string;
  body: string;
  enabled: boolean;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsNotificationRecord = {
  id: string;
  templateId: string;
  waybillId: string;
  event: string;
  channel: string;
  status: string;
  attemptCount: number;
  maxAttempts: number;
  idempotencyKey: string;
  lastError: string;
  renderedTitle: string;
  renderedBody: string;
  payload: Record<string, any>;
  sentAt: string;
  createdAt: string;
  updatedAt: string;
};

const normalizeCarrier = (raw: RawRecord): LogisticsCarrier => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  code: String(pick(raw, "code", "code") || ""),
  type: String(pick(raw, "type", "type") || ""),
  status: String(pick(raw, "status", "status") || ""),
  contactName: String(pick(raw, "contactName", "contact_name") || ""),
  contactPhone: String(pick(raw, "contactPhone", "contact_phone") || ""),
  capabilities: (pick(raw, "capabilities", "capabilities") || {}) as Record<string, any>,
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeTemplate = (raw: RawRecord): LogisticsTemplate => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  currency: String(pick(raw, "currency", "currency") || "CNY"),
  status: String(pick(raw, "status", "status") || "draft"),
  version: Number(pick(raw, "version", "version") || 1),
  channels: asArray<any>(pick(raw, "channels", "channels") || []),
  rules: (pick(raw, "rules", "rules") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeWaybill = (raw: RawRecord): LogisticsWaybill => ({
  id: String(pick(raw, "id", "id") || ""),
  orderId: String(pick(raw, "orderId", "order_id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  serviceCode: String(pick(raw, "serviceCode", "service_code") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  packageNo: Number(pick(raw, "packageNo", "package_no") || 1),
  packageKey: String(pick(raw, "packageKey", "package_key") || ""),
  shipmentItems: asArray<string>(pick(raw, "shipmentItems", "shipment_items") || []),
  orderItemCount: Number(pick(raw, "orderItemCount", "order_item_count") || 0),
  orderFulfillmentStatus: String(
    pick(raw, "orderFulfillmentStatus", "order_fulfillment_status") || "partial_shipped",
  ),
  status: String(pick(raw, "status", "status") || "created"),
  feeAmount: Number(pick(raw, "feeAmount", "fee_amount") || 0),
  actualFeeAmount: Number(pick(raw, "actualFeeAmount", "actual_fee_amount") || 0),
  feeDiffAmount: Number(pick(raw, "feeDiffAmount", "fee_diff_amount") || 0),
  billingStatus: String(pick(raw, "billingStatus", "billing_status") || "pending"),
  settledAt: String(pick(raw, "settledAt", "settled_at") || ""),
  labelUrl: String(pick(raw, "labelUrl", "label_url") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeBillingSummary = (raw: RawRecord): LogisticsBillingCarrierSummary => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  waybillCount: Number(pick(raw, "waybillCount", "waybill_count") || 0),
  estimatedFee: Number(pick(raw, "estimatedFee", "estimated_fee") || 0),
  actualFee: Number(pick(raw, "actualFee", "actual_fee") || 0),
  diffFee: Number(pick(raw, "diffFee", "diff_fee") || 0),
  abnormalCount: Number(pick(raw, "abnormalCount", "abnormal_count") || 0),
});

const normalizeTracking = (raw: RawRecord): LogisticsTracking => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  eventId: String(pick(raw, "eventId", "event_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  source: String(pick(raw, "source", "source") || ""),
  description: String(pick(raw, "description", "description") || ""),
  occurredAt: String(pick(raw, "occurredAt", "occurred_at") || ""),
  payload: (pick(raw, "payload", "payload") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeWaybillETA = (raw: RawRecord): LogisticsWaybillETA => ({
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  serviceCode: String(pick(raw, "serviceCode", "service_code") || ""),
  timezone: String(pick(raw, "timezone", "timezone") || "UTC"),
  pickupDeadlineAt: String(pick(raw, "pickupDeadlineAt", "pickup_deadline_at") || ""),
  deliveryDeadlineAt: String(pick(raw, "deliveryDeadlineAt", "delivery_deadline_at") || ""),
  promisedAt: String(pick(raw, "promisedAt", "promised_at") || ""),
  estimatedAt: String(pick(raw, "estimatedAt", "estimated_at") || ""),
  delayed: Boolean(pick(raw, "delayed", "delayed")),
  source: String(pick(raw, "source", "source") || "calculated"),
  lastComputedAt: String(pick(raw, "lastComputedAt", "last_computed_at") || ""),
});

const normalizeRoutingRule = (raw: RawRecord): LogisticsRoutingRule => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  serviceCode: String(pick(raw, "serviceCode", "service_code") || "std"),
  priority: Number(pick(raw, "priority", "priority") || 100),
  minWeight: Number(pick(raw, "minWeight", "min_weight") || 0),
  maxWeight: Number(pick(raw, "maxWeight", "max_weight") || 0),
  minOrderAmount: Number(pick(raw, "minOrderAmount", "min_order_amount") || 0),
  maxOrderAmount: Number(pick(raw, "maxOrderAmount", "max_order_amount") || 0),
  fallback: Boolean(pick(raw, "fallback", "fallback")),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  ruleConfig: (pick(raw, "ruleConfig", "rule_config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRoutingCandidate = (raw: RawRecord): LogisticsRoutingCandidate => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  serviceCode: String(pick(raw, "serviceCode", "service_code") || "std"),
  priority: Number(pick(raw, "priority", "priority") || 0),
  score: Number(pick(raw, "score", "score") || 0),
  matchedRule: String(pick(raw, "matchedRule", "matched_rule") || ""),
});

const normalizeRoutingPreview = (raw: RawRecord): LogisticsRoutingPreview => ({
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  serviceCode: String(pick(raw, "serviceCode", "service_code") || "std"),
  matchedRuleId: String(pick(raw, "matchedRuleId", "matched_rule_id") || ""),
  matchedRuleName: String(pick(raw, "matchedRuleName", "matched_rule_name") || ""),
  strategy: String(pick(raw, "strategy", "strategy") || ""),
  fallback: Boolean(pick(raw, "fallback", "fallback")),
  reason: String(pick(raw, "reason", "reason") || ""),
  candidates: asArray<RawRecord>(pick(raw, "candidates", "candidates")).map(normalizeRoutingCandidate),
});

const normalizeRedeliveryTask = (raw: RawRecord): LogisticsRedeliveryTask => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  status: String(pick(raw, "status", "status") || "initiated"),
  attemptNo: Number(pick(raw, "attemptNo", "attempt_no") || 1),
  addressSnapshot: (pick(raw, "addressSnapshot", "address_snapshot") || {}) as Record<string, any>,
  lastReason: String(pick(raw, "lastReason", "last_reason") || ""),
  operatorId: String(pick(raw, "operatorId", "operator_id") || ""),
  closedAt: String(pick(raw, "closedAt", "closed_at") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRiskRule = (raw: RawRecord): LogisticsRiskRule => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  matchField: String(pick(raw, "matchField", "match_field") || "address"),
  matchMode: String(pick(raw, "matchMode", "match_mode") || "contains"),
  pattern: String(pick(raw, "pattern", "pattern") || ""),
  decision: String(pick(raw, "decision", "decision") || "review"),
  riskLevel: String(pick(raw, "riskLevel", "risk_level") || "medium"),
  priority: Number(pick(raw, "priority", "priority") || 100),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  description: String(pick(raw, "description", "description") || ""),
  ruleConfig: (pick(raw, "ruleConfig", "rule_config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeBlacklistEntry = (raw: RawRecord): LogisticsBlacklistEntry => ({
  id: String(pick(raw, "id", "id") || ""),
  entryType: String(pick(raw, "entryType", "entry_type") || "recipient"),
  recipientName: String(pick(raw, "recipientName", "recipient_name") || ""),
  recipientPhone: String(pick(raw, "recipientPhone", "recipient_phone") || ""),
  addressLine: String(pick(raw, "addressLine", "address_line") || ""),
  reason: String(pick(raw, "reason", "reason") || ""),
  status: String(pick(raw, "status", "status") || "active"),
  expiresAt: String(pick(raw, "expiresAt", "expires_at") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRiskHit = (raw: RawRecord): LogisticsRiskHit => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  ruleId: String(pick(raw, "ruleId", "rule_id") || ""),
  blacklistId: String(pick(raw, "blacklistId", "blacklist_id") || ""),
  source: String(pick(raw, "source", "source") || "rule"),
  decision: String(pick(raw, "decision", "decision") || "review"),
  riskLevel: String(pick(raw, "riskLevel", "risk_level") || "medium"),
  fingerprint: String(pick(raw, "fingerprint", "fingerprint") || ""),
  description: String(pick(raw, "description", "description") || ""),
  status: String(pick(raw, "status", "status") || "open"),
  releasedBy: String(pick(raw, "releasedBy", "released_by") || ""),
  releaseReason: String(pick(raw, "releaseReason", "release_reason") || ""),
  releasedAt: String(pick(raw, "releasedAt", "released_at") || ""),
  payload: (pick(raw, "payload", "payload") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRiskEvaluateResult = (raw: RawRecord): LogisticsRiskEvaluateResult => ({
  blocked: Boolean(pick(raw, "blocked", "blocked")),
  decision: String(pick(raw, "decision", "decision") || "allow"),
  score: Number(pick(raw, "score", "score") || 0),
  releasedBy: String(pick(raw, "releasedBy", "released_by") || ""),
  fingerprint: String(pick(raw, "fingerprint", "fingerprint") || ""),
  matchedRules: asArray<RawRecord>(pick(raw, "matchedRules", "matched_rules")).map(normalizeRiskRule),
  matchedBlacklist: asArray<RawRecord>(pick(raw, "matchedBlacklist", "matched_blacklist")).map(normalizeBlacklistEntry),
  hits: asArray<RawRecord>(pick(raw, "hits", "hits")).map(normalizeRiskHit),
});

const normalizeLabelPrintTask = (raw: RawRecord): LogisticsLabelPrintTask => ({
  id: String(pick(raw, "id", "id") || ""),
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  status: String(pick(raw, "status", "status") || "pending"),
  attemptCount: Number(pick(raw, "attemptCount", "attempt_count") || 0),
  maxAttempts: Number(pick(raw, "maxAttempts", "max_attempts") || 3),
  lastError: String(pick(raw, "lastError", "last_error") || ""),
  retryQueuedAt: String(pick(raw, "retryQueuedAt", "retry_queued_at") || ""),
  printedAt: String(pick(raw, "printedAt", "printed_at") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeLabelPrintResult = (raw: RawRecord): LogisticsLabelPrintResult => ({
  task: raw?.task ? normalizeLabelPrintTask(raw.task as RawRecord) : null,
  idempotencyStatus: String(pick(raw, "idempotencyStatus", "idempotency_status") || "created"),
  success: Boolean(pick(raw, "success", "success")),
  message: String(pick(raw, "message", "message") || ""),
});

const normalizeSLASummary = (raw: RawRecord): LogisticsSLASummaryItem => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  waybillCount: Number(pick(raw, "waybillCount", "waybill_count") || 0),
  pickupOnTimeCount: Number(pick(raw, "pickupOnTimeCount", "pickup_on_time_count") || 0),
  pickupOnTimeRate: Number(pick(raw, "pickupOnTimeRate", "pickup_on_time_rate") || 0),
  signOnTimeCount: Number(pick(raw, "signOnTimeCount", "sign_on_time_count") || 0),
  signOnTimeRate: Number(pick(raw, "signOnTimeRate", "sign_on_time_rate") || 0),
  exceptionCount: Number(pick(raw, "exceptionCount", "exception_count") || 0),
  exceptionRate: Number(pick(raw, "exceptionRate", "exception_rate") || 0),
  pickupSLAHours: Number(pick(raw, "pickupSLAHours", "pickup_sla_hours") || 24),
  deliverySLAHours: Number(pick(raw, "deliverySLAHours", "delivery_sla_hours") || 72),
});

const normalizeNotificationTemplate = (raw: RawRecord): LogisticsNotificationTemplate => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  event: String(pick(raw, "event", "event") || ""),
  channel: String(pick(raw, "channel", "channel") || "sms"),
  title: String(pick(raw, "title", "title") || ""),
  body: String(pick(raw, "body", "body") || ""),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeNotificationRecord = (raw: RawRecord): LogisticsNotificationRecord => ({
  id: String(pick(raw, "id", "id") || ""),
  templateId: String(pick(raw, "templateId", "template_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  event: String(pick(raw, "event", "event") || ""),
  channel: String(pick(raw, "channel", "channel") || ""),
  status: String(pick(raw, "status", "status") || "pending"),
  attemptCount: Number(pick(raw, "attemptCount", "attempt_count") || 0),
  maxAttempts: Number(pick(raw, "maxAttempts", "max_attempts") || 3),
  idempotencyKey: String(pick(raw, "idempotencyKey", "idempotency_key") || ""),
  lastError: String(pick(raw, "lastError", "last_error") || ""),
  renderedTitle: String(pick(raw, "renderedTitle", "rendered_title") || ""),
  renderedBody: String(pick(raw, "renderedBody", "rendered_body") || ""),
  payload: (pick(raw, "payload", "payload") || {}) as Record<string, any>,
  sentAt: String(pick(raw, "sentAt", "sent_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeBillingCase = (raw: RawRecord): LogisticsBillingCase => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  caseNo: String(pick(raw, "caseNo", "case_no") || ""),
  status: String(pick(raw, "status", "status") || "open"),
  diffAmount: Number(pick(raw, "diffAmount", "diff_amount") || 0),
  reason: String(pick(raw, "reason", "reason") || ""),
  resolution: String(pick(raw, "resolution", "resolution") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  closedAt: String(pick(raw, "closedAt", "closed_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRateQuote = (raw: RawRecord): LogisticsRateQuoteResult => {
  const zone = (pick(raw, "matchedZone", "matched_zone") || {}) as RawRecord;
  return {
    templateId: String(pick(raw, "templateId", "template_id") || ""),
    template: String(pick(raw, "template", "template") || ""),
    currency: String(pick(raw, "currency", "currency") || "CNY"),
    billingType: String(pick(raw, "billingType", "billing_type") || "weight"),
    matchedZone: {
      region: String(pick(zone, "region", "region") || ""),
      firstMetric: Number(pick(zone, "firstMetric", "first_metric") || 0),
      firstFee: Number(pick(zone, "firstFee", "first_fee") || 0),
      additionalStep: Number(pick(zone, "additionalStep", "additional_step") || 0),
      additionalFee: Number(pick(zone, "additionalFee", "additional_fee") || 0),
      freeThreshold: Number(pick(zone, "freeThreshold", "free_threshold") || 0),
    },
    feeAmount: Number(pick(raw, "feeAmount", "fee_amount") || 0),
    breakdown: (pick(raw, "breakdown", "breakdown") || {}) as Record<string, any>,
  };
};

export function useLogisticsApi() {
  const basePath = "/admin/logistics";

  return {
    listCarriers: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/carriers`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCarrier);
    },

    upsertCarrier: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/carriers`, payload, init));
      return normalizeCarrier(raw || {});
    },

    disableCarrier: async (id: string, init?: any) => {
      const raw = await unwrap(apiPatch<ApiEnvelope<RawRecord>>(`${basePath}/carriers/${id}`, {}, init));
      return normalizeCarrier(raw || {});
    },

    testCarrier: (id: string, init?: any) =>
      unwrap(apiPost<ApiEnvelope<{ carrier_id: string; reachable: boolean; message: string }>>(`${basePath}/carriers/${id}/test`, {}, init)),

    listTemplates: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/templates`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeTemplate);
    },

    upsertTemplate: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/templates`, payload, init));
      return normalizeTemplate(raw || {});
    },

    publishTemplate: async (id: string, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/templates/${id}/publish`, {}, init));
      return normalizeTemplate(raw || {});
    },

    quoteTemplate: async (
      id: string,
      payload: { region: string; weight?: number; piece_count?: number; volume?: number; order_amount?: number },
      init?: any,
    ): Promise<LogisticsRateQuoteResult> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/templates/${id}/quote`, payload, init));
      return normalizeRateQuote(raw || {});
    },

    listWaybills: async (init?: any) => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/waybills`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeWaybill);
    },

    listRoutingRules: async (init?: any): Promise<LogisticsRoutingRule[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/routing/rules`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeRoutingRule);
    },

    upsertRoutingRule: async (payload: Record<string, any>, init?: any): Promise<LogisticsRoutingRule> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/routing/rules`, payload, init));
      return normalizeRoutingRule((raw || {}) as RawRecord);
    },

    deleteRoutingRule: async (id: string, init?: any): Promise<{ id: string; deleted: boolean }> => {
      const raw = await unwrap(apiDel<ApiEnvelope<RawRecord>>(`${basePath}/routing/rules/${id}`, undefined, init));
      return {
        id: String(pick((raw || {}) as RawRecord, "id", "id") || id),
        deleted: Boolean(pick((raw || {}) as RawRecord, "deleted", "deleted")),
      };
    },

    previewRouting: async (
      payload: {
        warehouse_id?: string;
        destination_zone?: string;
        weight?: number;
        order_amount?: number;
        preferred_carrier_id?: string;
        service_code?: string;
      },
      init?: any,
    ): Promise<LogisticsRoutingPreview> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/routing/preview`, payload, init));
      return normalizeRoutingPreview((raw || {}) as RawRecord);
    },

    listRedeliveryTasks: async (
      query?: { waybill_id?: string; status?: string },
      init?: any,
    ): Promise<LogisticsRedeliveryTask[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/redelivery/tasks`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeRedeliveryTask);
    },

    initiateRedeliveryTask: async (
      payload: Record<string, any>,
      init?: any,
    ): Promise<{ task: LogisticsRedeliveryTask; idempotencyStatus: string }> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ task: RawRecord; idempotency_status: string }>>(
          `${basePath}/redelivery/tasks/initiate`,
          payload,
          init,
        ),
      );
      return {
        task: normalizeRedeliveryTask((raw?.task || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    updateRedeliveryAddress: async (id: string, payload: Record<string, any>, init?: any): Promise<LogisticsRedeliveryTask> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/redelivery/tasks/${id}/address`, payload, init),
      );
      return normalizeRedeliveryTask((raw || {}) as RawRecord);
    },

    redispatchRedeliveryTask: async (
      id: string,
      payload: Record<string, any>,
      init?: any,
    ): Promise<{ task: LogisticsRedeliveryTask; idempotencyStatus: string }> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ task: RawRecord; idempotency_status: string }>>(
          `${basePath}/redelivery/tasks/${id}/redispatch`,
          payload,
          init,
        ),
      );
      return {
        task: normalizeRedeliveryTask((raw?.task || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    closeRedeliveryTask: async (id: string, payload: Record<string, any>, init?: any): Promise<LogisticsRedeliveryTask> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/redelivery/tasks/${id}/close`, payload, init),
      );
      return normalizeRedeliveryTask((raw || {}) as RawRecord);
    },

    listRiskRules: async (init?: any): Promise<LogisticsRiskRule[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/risk/rules`, undefined, init));
      return asArray<RawRecord>(raw?.items).map(normalizeRiskRule);
    },

    upsertRiskRule: async (payload: Record<string, any>, init?: any): Promise<LogisticsRiskRule> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/risk/rules`, payload, init));
      return normalizeRiskRule((raw || {}) as RawRecord);
    },

    listRiskBlacklist: async (query?: { status?: string }, init?: any): Promise<LogisticsBlacklistEntry[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/risk/blacklist`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeBlacklistEntry);
    },

    upsertRiskBlacklist: async (payload: Record<string, any>, init?: any): Promise<LogisticsBlacklistEntry> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/risk/blacklist`, payload, init));
      return normalizeBlacklistEntry((raw || {}) as RawRecord);
    },

    listRiskHits: async (
      query?: { waybill_id?: string; status?: string },
      init?: any,
    ): Promise<LogisticsRiskHit[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/risk/hits`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeRiskHit);
    },

    evaluateRisk: async (payload: Record<string, any>, init?: any): Promise<LogisticsRiskEvaluateResult> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/risk/evaluate`, payload, init));
      return normalizeRiskEvaluateResult((raw || {}) as RawRecord);
    },

    releaseRiskHit: async (id: string, payload: Record<string, any>, init?: any): Promise<LogisticsRiskHit> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/risk/hits/${id}/release`, payload, init));
      return normalizeRiskHit((raw || {}) as RawRecord);
    },

    listWaybillETA: async (
      query?: {
        waybill_ids?: string[] | string;
        destination_zone?: string;
        timezone?: string;
        force_recompute?: boolean;
      },
      init?: any,
    ): Promise<LogisticsWaybillETA[]> => {
      const waybillIDs = query?.waybill_ids;
      const normalizedQuery = {
        ...query,
        waybill_ids: Array.isArray(waybillIDs) ? waybillIDs.join(",") : waybillIDs,
      };
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/eta`, normalizedQuery, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeWaybillETA);
    },

    getWaybillETA: async (
      id: string,
      query?: { destination_zone?: string; timezone?: string; force_recompute?: boolean },
      init?: any,
    ): Promise<LogisticsWaybillETA> => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord>>(`${basePath}/eta/${id}`, query, init));
      return normalizeWaybillETA((raw || {}) as RawRecord);
    },

    createWaybill: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(apiPost<ApiEnvelope<{ waybill: RawRecord; idempotency_status: string }>>(`${basePath}/waybills`, payload, init));
      return {
        waybill: normalizeWaybill((raw?.waybill || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    getWaybillDetail: async (id: string, init?: any): Promise<LogisticsWaybillDetail> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ waybill: RawRecord; tracking: RawRecord[] }>>(`${basePath}/waybills/${id}`, undefined, init));
      return {
        waybill: normalizeWaybill((raw?.waybill || {}) as RawRecord),
        tracking: asArray<RawRecord>(raw?.tracking).map(normalizeTracking),
      };
    },

    updateWaybillCost: async (id: string, payload: { actual_fee_amount: number }, init?: any) => {
      const raw = await unwrap(apiPatch<ApiEnvelope<RawRecord>>(`${basePath}/waybills/${id}/cost`, payload, init));
      return normalizeWaybill(raw || {});
    },

    getBillingSummary: async (
      query?: { carrier_id?: string; from?: string; to?: string },
      init?: any,
    ): Promise<LogisticsBillingSnapshot> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ summary: RawRecord[]; items: RawRecord[] }>>(`${basePath}/billing/summary`, query, init),
      );
      return {
        summary: asArray<RawRecord>(raw?.summary).map(normalizeBillingSummary),
        items: asArray<RawRecord>(raw?.items).map(normalizeWaybill),
      };
    },

    exportBilling: async (
      query?: { carrier_id?: string; from?: string; to?: string; format?: "csv" | "json" | string },
      init?: any,
    ) => {
      return unwrap(apiGet<ApiEnvelope<Record<string, any>>>(`${basePath}/billing/export`, query, init));
    },

    listBillingCases: async (
      query?: { carrier_id?: string; status?: string },
      init?: any,
    ): Promise<LogisticsBillingCase[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/billing/cases`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeBillingCase);
    },

    createBillingCase: async (
      payload: { waybill_id: string; reason?: string; metadata?: Record<string, any> },
      init?: any,
    ): Promise<{ case: LogisticsBillingCase; idempotencyStatus: string }> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ case: RawRecord; idempotency_status: string }>>(
          `${basePath}/billing/cases`,
          payload,
          init,
        ),
      );
      return {
        case: normalizeBillingCase((raw?.case || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    transitionBillingCase: async (
      id: string,
      payload: { action: "confirm" | "appeal" | "writeoff"; note?: string; operator_id?: string },
      init?: any,
    ): Promise<LogisticsBillingCase> => {
      const raw = await unwrap(
        apiPatch<ApiEnvelope<RawRecord>>(`${basePath}/billing/cases/${id}/transition`, payload, init),
      );
      return normalizeBillingCase(raw || {});
    },

    listNotificationTemplates: async (init?: any): Promise<LogisticsNotificationTemplate[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/notifications/templates`, undefined, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeNotificationTemplate);
    },

    upsertNotificationTemplate: async (payload: Record<string, any>, init?: any) => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/notifications/templates`, payload, init),
      );
      return normalizeNotificationTemplate(raw || {});
    },

    listNotificationRecords: async (
      query?: { status?: string },
      init?: any,
    ): Promise<LogisticsNotificationRecord[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/notifications/records`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeNotificationRecord);
    },

    sendNotification: async (
      payload: Record<string, any>,
      init?: any,
    ): Promise<{ record: LogisticsNotificationRecord; idempotencyStatus: string }> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ record: RawRecord; idempotency_status: string }>>(
          `${basePath}/notifications/send`,
          payload,
          init,
        ),
      );
      return {
        record: normalizeNotificationRecord((raw?.record || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    retryNotification: async (id: string, init?: any): Promise<LogisticsNotificationRecord> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/notifications/records/${id}/retry`, {}, init),
      );
      return normalizeNotificationRecord(raw || {});
    },

    listLabelPrintTasks: async (query?: { status?: string }, init?: any) => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/labels/prints`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeLabelPrintTask);
    },

    batchPrintLabels: async (payload: Record<string, any>, init?: any): Promise<LogisticsLabelPrintBatchResult> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ results: RawRecord[]; success: number; failed: number }>>(`${basePath}/labels/prints`, payload, init),
      );
      return {
        results: asArray<RawRecord>(raw?.results).map(normalizeLabelPrintResult),
        success: Number(raw?.success || 0),
        failed: Number(raw?.failed || 0),
      };
    },

    retryLabelPrint: async (payload: Record<string, any>, init?: any): Promise<LogisticsLabelPrintBatchResult> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ results: RawRecord[]; success: number; failed: number }>>(`${basePath}/labels/prints/retry`, payload, init),
      );
      return {
        results: asArray<RawRecord>(raw?.results).map(normalizeLabelPrintResult),
        success: Number(raw?.success || 0),
        failed: Number(raw?.failed || 0),
      };
    },

    getSLADashboard: async (
      query?: { carrier_id?: string; from?: string; to?: string; pickup_sla_hours?: number; delivery_sla_hours?: number },
      init?: any,
    ): Promise<LogisticsSLASnapshot> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ summary: RawRecord[]; total: RawRecord }>>(`${basePath}/sla/dashboard`, query, init),
      );
      return {
        summary: asArray<RawRecord>(raw?.summary).map(normalizeSLASummary),
        total: normalizeSLASummary((raw?.total || {}) as RawRecord),
      };
    },
  };
}
