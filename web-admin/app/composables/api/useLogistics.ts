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

export type LogisticsWaybillSyncResult = {
  waybillId: string;
  waybillNo: string;
  provider: string;
  totalFetched: number;
  appended: number;
  replayed: number;
  currentStatus: string;
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

export type LogisticsTrackingSyncJob = {
  id: string;
  carrierId: string;
  waybillStatus: string;
  status: string;
  batchLimit: number;
  eventLimit: number;
  totalWaybills: number;
  successCount: number;
  failedCount: number;
  appendedCount: number;
  replayedCount: number;
  p95LatencyMS: number;
  lastError: string;
  startedAt: string;
  finishedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsGatewayHealthSummary = {
  windowHours: number;
  totalRequests: number;
  successRequests: number;
  failedRequests: number;
  successRate: number;
  p95LatencyMS: number;
};

export type LogisticsGatewayCarrierHealth = {
  carrierId: string;
  totalRequests: number;
  successRequests: number;
  failedRequests: number;
  successRate: number;
  p95LatencyMS: number;
};

export type LogisticsGatewayAlert = {
  level: string;
  code: string;
  message: string;
};

export type LogisticsGatewayHealthSnapshot = {
  summary: LogisticsGatewayHealthSummary;
  carriers: LogisticsGatewayCarrierHealth[];
  alerts: LogisticsGatewayAlert[];
};

export type LogisticsSLOGuardPolicy = {
  id: string;
  name: string;
  carrierId: string;
  windowHours: number;
  minSuccessRate: number;
  maxP95LatencyMS: number;
  maxFailedRequests: number;
  action: string;
  throttleRatio: number;
  enabled: boolean;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsSLOGuardTriggeredItem = {
  policyId: string;
  policyName: string;
  carrierId: string;
  action: string;
  throttleRatio: number;
  reasonCode: string;
  reasonMessage: string;
  successRate: number;
  failedCount: number;
  p95LatencyMS: number;
};

export type LogisticsSLOGuardStatusSnapshot = {
  windowHours: number;
  carrierId: string;
  totalPolicies: number;
  enabledPolicies: number;
  throttledPolicies: number;
  throttleRequired: boolean;
  gateway: LogisticsGatewayHealthSummary;
  triggered: LogisticsSLOGuardTriggeredItem[];
};

export type LogisticsGatewayCostSummary = {
  windowHours: number;
  totalRequests: number;
  successCount: number;
  failedCount: number;
  totalCost: number;
  quotaLimit: number;
  quotaUsed: number;
  quotaUsageRate: number;
};

export type LogisticsGatewayCostCarrier = {
  carrierId: string;
  provider: string;
  requestCount: number;
  successCount: number;
  failedCount: number;
  costAmount: number;
  quotaConsumed: number;
};

export type LogisticsGatewayCostAlert = {
  level: string;
  code: string;
  message: string;
  carrierId: string;
  provider: string;
  usageRate: number;
  requestRate: number;
};

export type LogisticsGatewayCostSnapshot = {
  summary: LogisticsGatewayCostSummary;
  carriers: LogisticsGatewayCostCarrier[];
  alerts: LogisticsGatewayCostAlert[];
};

export type LogisticsTrackingSyncSchedule = {
  id: string;
  name: string;
  cronExpr: string;
  carrierId: string;
  waybillStatus: string;
  enabled: boolean;
  maxConcurrency: number;
  dedupeWindowSec: number;
  batchLimit: number;
  eventLimit: number;
  lastTriggeredAt: string;
  nextTriggerAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsGatewayFailureEvent = {
  id: string;
  carrierId: string;
  waybillId: string;
  waybillNo: string;
  provider: string;
  sourceJobId: string;
  errorClass: string;
  errorCode: string;
  errorMessage: string;
  status: string;
  retryCount: number;
  nextRetryAt: string;
  circuitOpenTill: string;
  recoveredAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsExceptionOrchestrationRule = {
  id: string;
  name: string;
  triggerEvent: string;
  action: string;
  priority: number;
  enabled: boolean;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsExceptionOrchestrationRun = {
  id: string;
  ruleId: string;
  waybillId: string;
  waybillNo: string;
  trigger: string;
  result: string;
  message: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsAddressValidation = {
  id: string;
  requestKey: string;
  waybillId: string;
  waybillNo: string;
  rawAddress: string;
  normalized: string;
  reachable: boolean;
  riskLevel: string;
  suggestion: string;
  needManualReview: boolean;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsRoutingOptimizerStrategy = {
  id: string;
  name: string;
  timelinessWeight: number;
  costWeight: number;
  quotaWeight: number;
  riskWeight: number;
  fallbackStrategy: string;
  enabled: boolean;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsRoutingOptimizerCandidate = {
  carrierId: string;
  carrierName: string;
  finalScore: number;
  timeliness: number;
  cost: number;
  quota: number;
  risk: number;
  explain: string;
};

export type LogisticsRoutingOptimizerSimulation = {
  requestKey: string;
  profileID: string;
  strategy: string;
  degraded: boolean;
  reason: string;
  carrierId: string;
  carrierName: string;
  explain: string;
  candidates: LogisticsRoutingOptimizerCandidate[];
  createdAt: string;
};

export type LogisticsSettlementBatch = {
  id: string;
  batchNo: string;
  carrierId: string;
  status: string;
  waybillCount: number;
  diffCount: number;
  totalExpectedFee: number;
  totalActualFee: number;
  totalDiffAmount: number;
  suggestionSummary: Record<string, any>;
  confirmedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsSettlementDiff = {
  id: string;
  batchId: string;
  waybillId: string;
  waybillNo: string;
  carrierId: string;
  expectedFee: number;
  actualFee: number;
  diffAmount: number;
  attribution: string;
  suggestion: string;
  status: string;
  handledAction: string;
  handledNote: string;
  handledBy: string;
  handledAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsReconciliationBatch = {
  id: string;
  batchNo: string;
  carrierId: string;
  status: string;
  recordCount: number;
  matchedCount: number;
  exceptionCount: number;
  totalBillAmount: number;
  totalBankAmount: number;
  totalInvoiceAmount: number;
  summary: Record<string, any>;
  executedAt: string;
  confirmedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsReconciliationRecord = {
  id: string;
  batchId: string;
  waybillId: string;
  waybillNo: string;
  carrierId: string;
  billAmount: number;
  bankAmount: number;
  invoiceAmount: number;
  diffAmount: number;
  matchType: string;
  suggestion: string;
  status: string;
  caseId: string;
  handledAction: string;
  handledNote: string;
  handledBy: string;
  handledAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsReconciliationCase = {
  id: string;
  caseNo: string;
  batchId: string;
  recordId: string;
  carrierId: string;
  waybillNo: string;
  status: string;
  reason: string;
  suggestion: string;
  actionNote: string;
  handledBy: string;
  handledAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsControlTowerSummary = {
  windowHours: number;
  totalWaybills: number;
  inTransitCount: number;
  exceptionCount: number;
  timeoutCount: number;
  deliveredCount: number;
  onTimeRate: number;
  totalCost: number;
  alertCount: number;
};

export type LogisticsControlTowerAlert = {
  level: string;
  code: string;
  message: string;
};

export type LogisticsControlTowerOverview = {
  summary: LogisticsControlTowerSummary;
  alerts: LogisticsControlTowerAlert[];
};

export type LogisticsControlTowerDrilldownItem = {
  waybillId: string;
  waybillNo: string;
  carrierId: string;
  status: string;
  warehouseId: string;
  destinationZone: string;
  costAmount: number;
  elapsedHours: number;
  timeoutRiskLevel: string;
};

export type LogisticsControlTowerSubscription = {
  id: string;
  name: string;
  carrierId: string;
  warehouseId: string;
  destinationZone: string;
  minOnTimeRate: number;
  maxTimeoutCount: number;
  maxCostAmount: number;
  enabled: boolean;
  config: Record<string, any>;
  lastNotifiedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsKPIDashboardOverview = {
  windowHours: number;
  dimension: string;
  totalWaybills: number;
  deliveredCount: number;
  exceptionCount: number;
  timeoutCount: number;
  onTimeRate: number;
  deliverySuccessRate: number;
  avgTransitHours: number;
  totalCost: number;
  avgCost: number;
};

export type LogisticsKPIDashboardTrend = {
  dimensionKey: string;
  totalWaybills: number;
  deliveredCount: number;
  exceptionCount: number;
  onTimeRate: number;
  deliverySuccessRate: number;
  avgTransitHours: number;
  totalCost: number;
  avgCost: number;
};

export type LogisticsKPIDashboardDrilldown = {
  waybillId: string;
  waybillNo: string;
  carrierId: string;
  status: string;
  warehouseId: string;
  destinationZone: string;
  elapsedHours: number;
  costAmount: number;
  timeoutRiskLevel: string;
};

export type LogisticsCapacityPlan = {
  id: string;
  name: string;
  carrierId: string;
  warehouseId: string;
  destinationZone: string;
  dailyCapacity: number;
  reservedCapacity: number;
  usedCapacity: number;
  status: string;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsAllocationCandidate = {
  carrierId: string;
  carrierName: string;
  available: number;
  usageRate: number;
  timelinessScore: number;
  costScore: number;
  finalScore: number;
  reason: string;
};

export type LogisticsAllocationResult = {
  decisionID: string;
  requestKey: string;
  strategy: string;
  carrierId: string;
  carrierName: string;
  manualOverride: boolean;
  reason: string;
  candidates: LogisticsAllocationCandidate[];
  createdAt: string;
};

export type LogisticsCapacityForecast = {
  id: string;
  planId: string;
  carrierId: string;
  warehouseId: string;
  destinationZone: string;
  windowDays: number;
  currentDailyCapacity: number;
  predictedDailyVolume: number;
  targetCapacity: number;
  recommendedQuota: number;
  confidence: number;
  riskLevel: string;
  strategy: string;
  status: string;
  metrics: Record<string, any>;
  appliedAt: string;
  createdBy: string;
  updatedBy: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsLastmileRecoveryRule = {
  id: string;
  name: string;
  triggerEvent: string;
  action: string;
  priority: number;
  maxRetries: number;
  enabled: boolean;
  config: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsLastmileRecoveryRun = {
  id: string;
  requestKey: string;
  ruleId: string;
  waybillId: string;
  waybillNo: string;
  triggerEvent: string;
  action: string;
  status: string;
  retryCount: number;
  maxRetries: number;
  message: string;
  manualTaken: boolean;
  takenBy: string;
  takenReason: string;
  takenAt: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsCrossborderDocument = {
  id: string;
  waybillId: string;
  waybillNo: string;
  docType: string;
  docNo: string;
  countryFrom: string;
  countryTo: string;
  status: string;
  validatedAt: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsCrossborderTaxQuote = {
  id: string;
  requestKey: string;
  waybillId: string;
  waybillNo: string;
  destinationCountry: string;
  currency: string;
  declaredValue: number;
  shippingFee: number;
  insuranceFee: number;
  exemptionAmount: number;
  dutyRate: number;
  vatRate: number;
  dutyAmount: number;
  vatAmount: number;
  totalTaxAmount: number;
  normalizedStatus: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsCrossborderTrackingMap = {
  id: string;
  provider: string;
  providerStatus: string;
  normalizedStatus: string;
  description: string;
  priority: number;
  enabled: boolean;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsCustomsRulePack = {
  id: string;
  name: string;
  countryCode: string;
  status: string;
  strategy: string;
  defaultRiskLevel: string;
  description: string;
  metadata: Record<string, any>;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsCustomsRuleVersion = {
  id: string;
  packId: string;
  versionNo: number;
  status: string;
  hitStrategy: string;
  rules: Record<string, any>[];
  riskSnapshot: Record<string, any>;
  publishedAt: string;
  createdAt: string;
  updatedAt: string;
};

export type LogisticsCustomsPrecheckMatchedRule = {
  code: string;
  name: string;
  riskLevel: string;
  suggestion: string;
  reason: string;
};

export type LogisticsCustomsPrecheckResult = {
  packId: string;
  versionId: string;
  versionNo: number;
  countryCode: string;
  decision: string;
  riskLevel: string;
  suggestion: string;
  matchedRules: LogisticsCustomsPrecheckMatchedRule[];
  manualRelease: boolean;
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

const normalizeWaybillSyncResult = (raw: RawRecord): LogisticsWaybillSyncResult => ({
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  provider: String(pick(raw, "provider", "provider") || ""),
  totalFetched: Number(pick(raw, "totalFetched", "total_fetched") || 0),
  appended: Number(pick(raw, "appended", "appended") || 0),
  replayed: Number(pick(raw, "replayed", "replayed") || 0),
  currentStatus: String(pick(raw, "currentStatus", "current_status") || ""),
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

const normalizeTrackingSyncJob = (raw: RawRecord): LogisticsTrackingSyncJob => ({
  id: String(pick(raw, "id", "id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  waybillStatus: String(pick(raw, "waybillStatus", "waybill_status") || ""),
  status: String(pick(raw, "status", "status") || ""),
  batchLimit: Number(pick(raw, "batchLimit", "batch_limit") || 0),
  eventLimit: Number(pick(raw, "eventLimit", "event_limit") || 0),
  totalWaybills: Number(pick(raw, "totalWaybills", "total_waybills") || 0),
  successCount: Number(pick(raw, "successCount", "success_count") || 0),
  failedCount: Number(pick(raw, "failedCount", "failed_count") || 0),
  appendedCount: Number(pick(raw, "appendedCount", "appended_count") || 0),
  replayedCount: Number(pick(raw, "replayedCount", "replayed_count") || 0),
  p95LatencyMS: Number(pick(raw, "p95LatencyMS", "p95_latency_ms") || 0),
  lastError: String(pick(raw, "lastError", "last_error") || ""),
  startedAt: String(pick(raw, "startedAt", "started_at") || ""),
  finishedAt: String(pick(raw, "finishedAt", "finished_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeGatewayCarrierHealth = (raw: RawRecord): LogisticsGatewayCarrierHealth => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  totalRequests: Number(pick(raw, "totalRequests", "total_requests") || 0),
  successRequests: Number(pick(raw, "successRequests", "success_requests") || 0),
  failedRequests: Number(pick(raw, "failedRequests", "failed_requests") || 0),
  successRate: Number(pick(raw, "successRate", "success_rate") || 0),
  p95LatencyMS: Number(pick(raw, "p95LatencyMS", "p95_latency_ms") || 0),
});

const normalizeGatewayAlert = (raw: RawRecord): LogisticsGatewayAlert => ({
  level: String(pick(raw, "level", "level") || ""),
  code: String(pick(raw, "code", "code") || ""),
  message: String(pick(raw, "message", "message") || ""),
});

const normalizeGatewayHealthSummary = (raw: RawRecord): LogisticsGatewayHealthSummary => ({
  windowHours: Number(pick(raw, "windowHours", "window_hours") || 24),
  totalRequests: Number(pick(raw, "totalRequests", "total_requests") || 0),
  successRequests: Number(pick(raw, "successRequests", "success_requests") || 0),
  failedRequests: Number(pick(raw, "failedRequests", "failed_requests") || 0),
  successRate: Number(pick(raw, "successRate", "success_rate") || 0),
  p95LatencyMS: Number(pick(raw, "p95LatencyMS", "p95_latency_ms") || 0),
});

const normalizeSLOGuardPolicy = (raw: RawRecord): LogisticsSLOGuardPolicy => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  windowHours: Number(pick(raw, "windowHours", "window_hours") || 24),
  minSuccessRate: Number(pick(raw, "minSuccessRate", "min_success_rate") || 95),
  maxP95LatencyMS: Number(pick(raw, "maxP95LatencyMS", "max_p95_latency_ms") || 2000),
  maxFailedRequests: Number(pick(raw, "maxFailedRequests", "max_failed_requests") || 10),
  action: String(pick(raw, "action", "action") || "throttle"),
  throttleRatio: Number(pick(raw, "throttleRatio", "throttle_ratio") || 50),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeSLOGuardTriggered = (raw: RawRecord): LogisticsSLOGuardTriggeredItem => ({
  policyId: String(pick(raw, "policyId", "policy_id") || ""),
  policyName: String(pick(raw, "policyName", "policy_name") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  action: String(pick(raw, "action", "action") || "throttle"),
  throttleRatio: Number(pick(raw, "throttleRatio", "throttle_ratio") || 0),
  reasonCode: String(pick(raw, "reasonCode", "reason_code") || ""),
  reasonMessage: String(pick(raw, "reasonMessage", "reason_message") || ""),
  successRate: Number(pick(raw, "successRate", "success_rate") || 0),
  failedCount: Number(pick(raw, "failedCount", "failed_count") || 0),
  p95LatencyMS: Number(pick(raw, "p95LatencyMS", "p95_latency_ms") || 0),
});

const normalizeSLOGuardStatus = (raw: RawRecord): LogisticsSLOGuardStatusSnapshot => ({
  windowHours: Number(pick(raw, "windowHours", "window_hours") || 24),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  totalPolicies: Number(pick(raw, "totalPolicies", "total_policies") || 0),
  enabledPolicies: Number(pick(raw, "enabledPolicies", "enabled_policies") || 0),
  throttledPolicies: Number(pick(raw, "throttledPolicies", "throttled_policies") || 0),
  throttleRequired: Boolean(pick(raw, "throttleRequired", "throttle_required")),
  gateway: normalizeGatewayHealthSummary((pick(raw, "gateway", "gateway") || {}) as RawRecord),
  triggered: asArray<RawRecord>(pick(raw, "triggered", "triggered")).map(normalizeSLOGuardTriggered),
});

const normalizeGatewayCostSummary = (raw: RawRecord): LogisticsGatewayCostSummary => ({
  windowHours: Number(pick(raw, "windowHours", "window_hours") || 24),
  totalRequests: Number(pick(raw, "totalRequests", "total_requests") || 0),
  successCount: Number(pick(raw, "successCount", "success_count") || 0),
  failedCount: Number(pick(raw, "failedCount", "failed_count") || 0),
  totalCost: Number(pick(raw, "totalCost", "total_cost") || 0),
  quotaLimit: Number(pick(raw, "quotaLimit", "quota_limit") || 0),
  quotaUsed: Number(pick(raw, "quotaUsed", "quota_used") || 0),
  quotaUsageRate: Number(pick(raw, "quotaUsageRate", "quota_usage_rate") || 0),
});

const normalizeGatewayCostCarrier = (raw: RawRecord): LogisticsGatewayCostCarrier => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  provider: String(pick(raw, "provider", "provider") || ""),
  requestCount: Number(pick(raw, "requestCount", "request_count") || 0),
  successCount: Number(pick(raw, "successCount", "success_count") || 0),
  failedCount: Number(pick(raw, "failedCount", "failed_count") || 0),
  costAmount: Number(pick(raw, "costAmount", "cost_amount") || 0),
  quotaConsumed: Number(pick(raw, "quotaConsumed", "quota_consumed") || 0),
});

const normalizeGatewayCostAlert = (raw: RawRecord): LogisticsGatewayCostAlert => ({
  level: String(pick(raw, "level", "level") || ""),
  code: String(pick(raw, "code", "code") || ""),
  message: String(pick(raw, "message", "message") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  provider: String(pick(raw, "provider", "provider") || ""),
  usageRate: Number(pick(raw, "usageRate", "usage_rate") || 0),
  requestRate: Number(pick(raw, "requestRate", "request_rate") || 0),
});

const normalizeTrackingSyncSchedule = (raw: RawRecord): LogisticsTrackingSyncSchedule => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  cronExpr: String(pick(raw, "cronExpr", "cron_expr") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  waybillStatus: String(pick(raw, "waybillStatus", "waybill_status") || ""),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  maxConcurrency: Number(pick(raw, "maxConcurrency", "max_concurrency") || 1),
  dedupeWindowSec: Number(pick(raw, "dedupeWindowSec", "dedupe_window_sec") || 0),
  batchLimit: Number(pick(raw, "batchLimit", "batch_limit") || 0),
  eventLimit: Number(pick(raw, "eventLimit", "event_limit") || 0),
  lastTriggeredAt: String(pick(raw, "lastTriggeredAt", "last_triggered_at") || ""),
  nextTriggerAt: String(pick(raw, "nextTriggerAt", "next_trigger_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeGatewayFailureEvent = (raw: RawRecord): LogisticsGatewayFailureEvent => ({
  id: String(pick(raw, "id", "id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  provider: String(pick(raw, "provider", "provider") || ""),
  sourceJobId: String(pick(raw, "sourceJobId", "source_job_id") || ""),
  errorClass: String(pick(raw, "errorClass", "error_class") || ""),
  errorCode: String(pick(raw, "errorCode", "error_code") || ""),
  errorMessage: String(pick(raw, "errorMessage", "error_message") || ""),
  status: String(pick(raw, "status", "status") || ""),
  retryCount: Number(pick(raw, "retryCount", "retry_count") || 0),
  nextRetryAt: String(pick(raw, "nextRetryAt", "next_retry_at") || ""),
  circuitOpenTill: String(pick(raw, "circuitOpenTill", "circuit_open_till") || ""),
  recoveredAt: String(pick(raw, "recoveredAt", "recovered_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeExceptionOrchestrationRule = (raw: RawRecord): LogisticsExceptionOrchestrationRule => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  triggerEvent: String(pick(raw, "triggerEvent", "trigger_event") || ""),
  action: String(pick(raw, "action", "action") || ""),
  priority: Number(pick(raw, "priority", "priority") || 100),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeExceptionOrchestrationRun = (raw: RawRecord): LogisticsExceptionOrchestrationRun => ({
  id: String(pick(raw, "id", "id") || ""),
  ruleId: String(pick(raw, "ruleId", "rule_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  trigger: String(pick(raw, "trigger", "trigger") || ""),
  result: String(pick(raw, "result", "result") || ""),
  message: String(pick(raw, "message", "message") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeAddressValidation = (raw: RawRecord): LogisticsAddressValidation => ({
  id: String(pick(raw, "id", "id") || ""),
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  rawAddress: String(pick(raw, "rawAddress", "raw_address") || ""),
  normalized: String(pick(raw, "normalized", "normalized") || ""),
  reachable: Boolean(pick(raw, "reachable", "reachable")),
  riskLevel: String(pick(raw, "riskLevel", "risk_level") || "low"),
  suggestion: String(pick(raw, "suggestion", "suggestion") || ""),
  needManualReview: Boolean(pick(raw, "needManualReview", "need_manual_review")),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRoutingOptimizerStrategy = (raw: RawRecord): LogisticsRoutingOptimizerStrategy => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  timelinessWeight: Number(pick(raw, "timelinessWeight", "timeliness_weight") || 0.4),
  costWeight: Number(pick(raw, "costWeight", "cost_weight") || 0.3),
  quotaWeight: Number(pick(raw, "quotaWeight", "quota_weight") || 0.2),
  riskWeight: Number(pick(raw, "riskWeight", "risk_weight") || 0.1),
  fallbackStrategy: String(pick(raw, "fallbackStrategy", "fallback_strategy") || "highest_timeliness"),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeRoutingOptimizerCandidate = (raw: RawRecord): LogisticsRoutingOptimizerCandidate => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  finalScore: Number(pick(raw, "finalScore", "final_score") || 0),
  timeliness: Number(pick(raw, "timeliness", "timeliness") || 0),
  cost: Number(pick(raw, "cost", "cost") || 0),
  quota: Number(pick(raw, "quota", "quota") || 0),
  risk: Number(pick(raw, "risk", "risk") || 0),
  explain: String(pick(raw, "explain", "explain") || ""),
});

const normalizeRoutingOptimizerSimulation = (raw: RawRecord): LogisticsRoutingOptimizerSimulation => ({
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  profileID: String(pick(raw, "profileID", "profile_id") || ""),
  strategy: String(pick(raw, "strategy", "strategy") || ""),
  degraded: Boolean(pick(raw, "degraded", "degraded")),
  reason: String(pick(raw, "reason", "reason") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  explain: String(pick(raw, "explain", "explain") || ""),
  candidates: asArray<RawRecord>(pick(raw, "candidates", "candidates")).map(normalizeRoutingOptimizerCandidate),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeSettlementBatch = (raw: RawRecord): LogisticsSettlementBatch => ({
  id: String(pick(raw, "id", "id") || ""),
  batchNo: String(pick(raw, "batchNo", "batch_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  waybillCount: Number(pick(raw, "waybillCount", "waybill_count") || 0),
  diffCount: Number(pick(raw, "diffCount", "diff_count") || 0),
  totalExpectedFee: Number(pick(raw, "totalExpectedFee", "total_expected_fee") || 0),
  totalActualFee: Number(pick(raw, "totalActualFee", "total_actual_fee") || 0),
  totalDiffAmount: Number(pick(raw, "totalDiffAmount", "total_diff_amount") || 0),
  suggestionSummary: (pick(raw, "suggestionSummary", "suggestion_summary") || {}) as Record<string, any>,
  confirmedAt: String(pick(raw, "confirmedAt", "confirmed_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeSettlementDiff = (raw: RawRecord): LogisticsSettlementDiff => ({
  id: String(pick(raw, "id", "id") || ""),
  batchId: String(pick(raw, "batchId", "batch_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  expectedFee: Number(pick(raw, "expectedFee", "expected_fee") || 0),
  actualFee: Number(pick(raw, "actualFee", "actual_fee") || 0),
  diffAmount: Number(pick(raw, "diffAmount", "diff_amount") || 0),
  attribution: String(pick(raw, "attribution", "attribution") || ""),
  suggestion: String(pick(raw, "suggestion", "suggestion") || ""),
  status: String(pick(raw, "status", "status") || ""),
  handledAction: String(pick(raw, "handledAction", "handled_action") || ""),
  handledNote: String(pick(raw, "handledNote", "handled_note") || ""),
  handledBy: String(pick(raw, "handledBy", "handled_by") || ""),
  handledAt: String(pick(raw, "handledAt", "handled_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeReconciliationBatch = (raw: RawRecord): LogisticsReconciliationBatch => ({
  id: String(pick(raw, "id", "id") || ""),
  batchNo: String(pick(raw, "batchNo", "batch_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  recordCount: Number(pick(raw, "recordCount", "record_count") || 0),
  matchedCount: Number(pick(raw, "matchedCount", "matched_count") || 0),
  exceptionCount: Number(pick(raw, "exceptionCount", "exception_count") || 0),
  totalBillAmount: Number(pick(raw, "totalBillAmount", "total_bill_amount") || 0),
  totalBankAmount: Number(pick(raw, "totalBankAmount", "total_bank_amount") || 0),
  totalInvoiceAmount: Number(pick(raw, "totalInvoiceAmount", "total_invoice_amount") || 0),
  summary: (pick(raw, "summary", "summary") || {}) as Record<string, any>,
  executedAt: String(pick(raw, "executedAt", "executed_at") || ""),
  confirmedAt: String(pick(raw, "confirmedAt", "confirmed_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeReconciliationRecord = (raw: RawRecord): LogisticsReconciliationRecord => ({
  id: String(pick(raw, "id", "id") || ""),
  batchId: String(pick(raw, "batchId", "batch_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  billAmount: Number(pick(raw, "billAmount", "bill_amount") || 0),
  bankAmount: Number(pick(raw, "bankAmount", "bank_amount") || 0),
  invoiceAmount: Number(pick(raw, "invoiceAmount", "invoice_amount") || 0),
  diffAmount: Number(pick(raw, "diffAmount", "diff_amount") || 0),
  matchType: String(pick(raw, "matchType", "match_type") || ""),
  suggestion: String(pick(raw, "suggestion", "suggestion") || ""),
  status: String(pick(raw, "status", "status") || ""),
  caseId: String(pick(raw, "caseId", "case_id") || ""),
  handledAction: String(pick(raw, "handledAction", "handled_action") || ""),
  handledNote: String(pick(raw, "handledNote", "handled_note") || ""),
  handledBy: String(pick(raw, "handledBy", "handled_by") || ""),
  handledAt: String(pick(raw, "handledAt", "handled_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeReconciliationCase = (raw: RawRecord): LogisticsReconciliationCase => ({
  id: String(pick(raw, "id", "id") || ""),
  caseNo: String(pick(raw, "caseNo", "case_no") || ""),
  batchId: String(pick(raw, "batchId", "batch_id") || ""),
  recordId: String(pick(raw, "recordId", "record_id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  status: String(pick(raw, "status", "status") || ""),
  reason: String(pick(raw, "reason", "reason") || ""),
  suggestion: String(pick(raw, "suggestion", "suggestion") || ""),
  actionNote: String(pick(raw, "actionNote", "action_note") || ""),
  handledBy: String(pick(raw, "handledBy", "handled_by") || ""),
  handledAt: String(pick(raw, "handledAt", "handled_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeControlTowerSummary = (raw: RawRecord): LogisticsControlTowerSummary => ({
  windowHours: Number(pick(raw, "windowHours", "window_hours") || 24),
  totalWaybills: Number(pick(raw, "totalWaybills", "total_waybills") || 0),
  inTransitCount: Number(pick(raw, "inTransitCount", "in_transit_count") || 0),
  exceptionCount: Number(pick(raw, "exceptionCount", "exception_count") || 0),
  timeoutCount: Number(pick(raw, "timeoutCount", "timeout_count") || 0),
  deliveredCount: Number(pick(raw, "deliveredCount", "delivered_count") || 0),
  onTimeRate: Number(pick(raw, "onTimeRate", "on_time_rate") || 0),
  totalCost: Number(pick(raw, "totalCost", "total_cost") || 0),
  alertCount: Number(pick(raw, "alertCount", "alert_count") || 0),
});

const normalizeControlTowerAlert = (raw: RawRecord): LogisticsControlTowerAlert => ({
  level: String(pick(raw, "level", "level") || ""),
  code: String(pick(raw, "code", "code") || ""),
  message: String(pick(raw, "message", "message") || ""),
});

const normalizeControlTowerDrilldownItem = (raw: RawRecord): LogisticsControlTowerDrilldownItem => ({
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  costAmount: Number(pick(raw, "costAmount", "cost_amount") || 0),
  elapsedHours: Number(pick(raw, "elapsedHours", "elapsed_hours") || 0),
  timeoutRiskLevel: String(pick(raw, "timeoutRiskLevel", "timeout_risk_level") || ""),
});

const normalizeControlTowerSubscription = (raw: RawRecord): LogisticsControlTowerSubscription => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  minOnTimeRate: Number(pick(raw, "minOnTimeRate", "min_on_time_rate") || 95),
  maxTimeoutCount: Number(pick(raw, "maxTimeoutCount", "max_timeout_count") || 0),
  maxCostAmount: Number(pick(raw, "maxCostAmount", "max_cost_amount") || 0),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  lastNotifiedAt: String(pick(raw, "lastNotifiedAt", "last_notified_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeKPIDashboardOverview = (raw: RawRecord): LogisticsKPIDashboardOverview => ({
  windowHours: Number(pick(raw, "windowHours", "window_hours") || 24),
  dimension: String(pick(raw, "dimension", "dimension") || "carrier"),
  totalWaybills: Number(pick(raw, "totalWaybills", "total_waybills") || 0),
  deliveredCount: Number(pick(raw, "deliveredCount", "delivered_count") || 0),
  exceptionCount: Number(pick(raw, "exceptionCount", "exception_count") || 0),
  timeoutCount: Number(pick(raw, "timeoutCount", "timeout_count") || 0),
  onTimeRate: Number(pick(raw, "onTimeRate", "on_time_rate") || 0),
  deliverySuccessRate: Number(pick(raw, "deliverySuccessRate", "delivery_success_rate") || 0),
  avgTransitHours: Number(pick(raw, "avgTransitHours", "avg_transit_hours") || 0),
  totalCost: Number(pick(raw, "totalCost", "total_cost") || 0),
  avgCost: Number(pick(raw, "avgCost", "avg_cost") || 0),
});

const normalizeKPIDashboardTrend = (raw: RawRecord): LogisticsKPIDashboardTrend => ({
  dimensionKey: String(pick(raw, "dimensionKey", "dimension_key") || ""),
  totalWaybills: Number(pick(raw, "totalWaybills", "total_waybills") || 0),
  deliveredCount: Number(pick(raw, "deliveredCount", "delivered_count") || 0),
  exceptionCount: Number(pick(raw, "exceptionCount", "exception_count") || 0),
  onTimeRate: Number(pick(raw, "onTimeRate", "on_time_rate") || 0),
  deliverySuccessRate: Number(pick(raw, "deliverySuccessRate", "delivery_success_rate") || 0),
  avgTransitHours: Number(pick(raw, "avgTransitHours", "avg_transit_hours") || 0),
  totalCost: Number(pick(raw, "totalCost", "total_cost") || 0),
  avgCost: Number(pick(raw, "avgCost", "avg_cost") || 0),
});

const normalizeKPIDashboardDrilldown = (raw: RawRecord): LogisticsKPIDashboardDrilldown => ({
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  status: String(pick(raw, "status", "status") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  elapsedHours: Number(pick(raw, "elapsedHours", "elapsed_hours") || 0),
  costAmount: Number(pick(raw, "costAmount", "cost_amount") || 0),
  timeoutRiskLevel: String(pick(raw, "timeoutRiskLevel", "timeout_risk_level") || ""),
});

const normalizeCapacityPlan = (raw: RawRecord): LogisticsCapacityPlan => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  dailyCapacity: Number(pick(raw, "dailyCapacity", "daily_capacity") || 0),
  reservedCapacity: Number(pick(raw, "reservedCapacity", "reserved_capacity") || 0),
  usedCapacity: Number(pick(raw, "usedCapacity", "used_capacity") || 0),
  status: String(pick(raw, "status", "status") || "active"),
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeAllocationCandidate = (raw: RawRecord): LogisticsAllocationCandidate => ({
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  available: Number(pick(raw, "available", "available") || 0),
  usageRate: Number(pick(raw, "usageRate", "usage_rate") || 0),
  timelinessScore: Number(pick(raw, "timelinessScore", "timeliness_score") || 0),
  costScore: Number(pick(raw, "costScore", "cost_score") || 0),
  finalScore: Number(pick(raw, "finalScore", "final_score") || 0),
  reason: String(pick(raw, "reason", "reason") || ""),
});

const normalizeAllocationResult = (raw: RawRecord): LogisticsAllocationResult => ({
  decisionID: String(pick(raw, "decisionID", "decision_id") || ""),
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  strategy: String(pick(raw, "strategy", "strategy") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  carrierName: String(pick(raw, "carrierName", "carrier_name") || ""),
  manualOverride: Boolean(pick(raw, "manualOverride", "manual_override")),
  reason: String(pick(raw, "reason", "reason") || ""),
  candidates: asArray<RawRecord>(pick(raw, "candidates", "candidates")).map(normalizeAllocationCandidate),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
});

const normalizeCapacityForecast = (raw: RawRecord): LogisticsCapacityForecast => ({
  id: String(pick(raw, "id", "id") || ""),
  planId: String(pick(raw, "planId", "plan_id") || ""),
  carrierId: String(pick(raw, "carrierId", "carrier_id") || ""),
  warehouseId: String(pick(raw, "warehouseId", "warehouse_id") || ""),
  destinationZone: String(pick(raw, "destinationZone", "destination_zone") || ""),
  windowDays: Number(pick(raw, "windowDays", "window_days") || 7),
  currentDailyCapacity: Number(pick(raw, "currentDailyCapacity", "current_daily_capacity") || 0),
  predictedDailyVolume: Number(pick(raw, "predictedDailyVolume", "predicted_daily_volume") || 0),
  targetCapacity: Number(pick(raw, "targetCapacity", "target_capacity") || 0),
  recommendedQuota: Number(pick(raw, "recommendedQuota", "recommended_quota") || 0),
  confidence: Number(pick(raw, "confidence", "confidence") || 0),
  riskLevel: String(pick(raw, "riskLevel", "risk_level") || ""),
  strategy: String(pick(raw, "strategy", "strategy") || ""),
  status: String(pick(raw, "status", "status") || "suggested"),
  metrics: (pick(raw, "metrics", "metrics") || {}) as Record<string, any>,
  appliedAt: String(pick(raw, "appliedAt", "applied_at") || ""),
  createdBy: String(pick(raw, "createdBy", "created_by") || ""),
  updatedBy: String(pick(raw, "updatedBy", "updated_by") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeLastmileRecoveryRule = (raw: RawRecord): LogisticsLastmileRecoveryRule => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  triggerEvent: String(pick(raw, "triggerEvent", "trigger_event") || ""),
  action: String(pick(raw, "action", "action") || ""),
  priority: Number(pick(raw, "priority", "priority") || 100),
  maxRetries: Number(pick(raw, "maxRetries", "max_retries") || 3),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  config: (pick(raw, "config", "config") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeLastmileRecoveryRun = (raw: RawRecord): LogisticsLastmileRecoveryRun => ({
  id: String(pick(raw, "id", "id") || ""),
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  ruleId: String(pick(raw, "ruleId", "rule_id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  triggerEvent: String(pick(raw, "triggerEvent", "trigger_event") || ""),
  action: String(pick(raw, "action", "action") || ""),
  status: String(pick(raw, "status", "status") || ""),
  retryCount: Number(pick(raw, "retryCount", "retry_count") || 0),
  maxRetries: Number(pick(raw, "maxRetries", "max_retries") || 0),
  message: String(pick(raw, "message", "message") || ""),
  manualTaken: Boolean(pick(raw, "manualTaken", "manual_taken")),
  takenBy: String(pick(raw, "takenBy", "taken_by") || ""),
  takenReason: String(pick(raw, "takenReason", "taken_reason") || ""),
  takenAt: String(pick(raw, "takenAt", "taken_at") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeCrossborderDocument = (raw: RawRecord): LogisticsCrossborderDocument => ({
  id: String(pick(raw, "id", "id") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  docType: String(pick(raw, "docType", "doc_type") || ""),
  docNo: String(pick(raw, "docNo", "doc_no") || ""),
  countryFrom: String(pick(raw, "countryFrom", "country_from") || ""),
  countryTo: String(pick(raw, "countryTo", "country_to") || ""),
  status: String(pick(raw, "status", "status") || "pending"),
  validatedAt: String(pick(raw, "validatedAt", "validated_at") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeCrossborderTaxQuote = (raw: RawRecord): LogisticsCrossborderTaxQuote => ({
  id: String(pick(raw, "id", "id") || ""),
  requestKey: String(pick(raw, "requestKey", "request_key") || ""),
  waybillId: String(pick(raw, "waybillId", "waybill_id") || ""),
  waybillNo: String(pick(raw, "waybillNo", "waybill_no") || ""),
  destinationCountry: String(pick(raw, "destinationCountry", "destination_country") || ""),
  currency: String(pick(raw, "currency", "currency") || "USD"),
  declaredValue: Number(pick(raw, "declaredValue", "declared_value") || 0),
  shippingFee: Number(pick(raw, "shippingFee", "shipping_fee") || 0),
  insuranceFee: Number(pick(raw, "insuranceFee", "insurance_fee") || 0),
  exemptionAmount: Number(pick(raw, "exemptionAmount", "exemption_amount") || 0),
  dutyRate: Number(pick(raw, "dutyRate", "duty_rate") || 0),
  vatRate: Number(pick(raw, "vatRate", "vat_rate") || 0),
  dutyAmount: Number(pick(raw, "dutyAmount", "duty_amount") || 0),
  vatAmount: Number(pick(raw, "vatAmount", "vat_amount") || 0),
  totalTaxAmount: Number(pick(raw, "totalTaxAmount", "total_tax_amount") || 0),
  normalizedStatus: String(pick(raw, "normalizedStatus", "normalized_status") || "estimated"),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeCrossborderTrackingMap = (raw: RawRecord): LogisticsCrossborderTrackingMap => ({
  id: String(pick(raw, "id", "id") || ""),
  provider: String(pick(raw, "provider", "provider") || ""),
  providerStatus: String(pick(raw, "providerStatus", "provider_status") || ""),
  normalizedStatus: String(pick(raw, "normalizedStatus", "normalized_status") || "created"),
  description: String(pick(raw, "description", "description") || ""),
  priority: Number(pick(raw, "priority", "priority") || 100),
  enabled: Boolean(pick(raw, "enabled", "enabled")),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeCustomsRulePack = (raw: RawRecord): LogisticsCustomsRulePack => ({
  id: String(pick(raw, "id", "id") || ""),
  name: String(pick(raw, "name", "name") || ""),
  countryCode: String(pick(raw, "countryCode", "country_code") || ""),
  status: String(pick(raw, "status", "status") || "draft"),
  strategy: String(pick(raw, "strategy", "strategy") || "first_hit"),
  defaultRiskLevel: String(pick(raw, "defaultRiskLevel", "default_risk_level") || "low"),
  description: String(pick(raw, "description", "description") || ""),
  metadata: (pick(raw, "metadata", "metadata") || {}) as Record<string, any>,
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeCustomsRuleVersion = (raw: RawRecord): LogisticsCustomsRuleVersion => ({
  id: String(pick(raw, "id", "id") || ""),
  packId: String(pick(raw, "packId", "pack_id") || ""),
  versionNo: Number(pick(raw, "versionNo", "version_no") || 1),
  status: String(pick(raw, "status", "status") || "draft"),
  hitStrategy: String(pick(raw, "hitStrategy", "hit_strategy") || "first_hit"),
  rules: asArray<RawRecord>(pick(raw, "rules", "rules")),
  riskSnapshot: (pick(raw, "riskSnapshot", "risk_snapshot") || {}) as Record<string, any>,
  publishedAt: String(pick(raw, "publishedAt", "published_at") || ""),
  createdAt: String(pick(raw, "createdAt", "created_at") || ""),
  updatedAt: String(pick(raw, "updatedAt", "updated_at") || ""),
});

const normalizeCustomsPrecheckMatchedRule = (raw: RawRecord): LogisticsCustomsPrecheckMatchedRule => ({
  code: String(pick(raw, "code", "code") || ""),
  name: String(pick(raw, "name", "name") || ""),
  riskLevel: String(pick(raw, "riskLevel", "risk_level") || ""),
  suggestion: String(pick(raw, "suggestion", "suggestion") || ""),
  reason: String(pick(raw, "reason", "reason") || ""),
});

const normalizeCustomsPrecheckResult = (raw: RawRecord): LogisticsCustomsPrecheckResult => ({
  packId: String(pick(raw, "packId", "pack_id") || ""),
  versionId: String(pick(raw, "versionId", "version_id") || ""),
  versionNo: Number(pick(raw, "versionNo", "version_no") || 1),
  countryCode: String(pick(raw, "countryCode", "country_code") || ""),
  decision: String(pick(raw, "decision", "decision") || "pass"),
  riskLevel: String(pick(raw, "riskLevel", "risk_level") || "low"),
  suggestion: String(pick(raw, "suggestion", "suggestion") || ""),
  matchedRules: asArray<RawRecord>(pick(raw, "matchedRules", "matched_rules")).map(normalizeCustomsPrecheckMatchedRule),
  manualRelease: Boolean(pick(raw, "manualRelease", "manual_release")),
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

    syncWaybillTracking: async (
      id: string,
      query?: { limit?: number },
      init?: any,
    ): Promise<LogisticsWaybillSyncResult> => {
      const suffix = query?.limit && query.limit > 0 ? `?limit=${query.limit}` : "";
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/waybills/${id}/sync-track${suffix}`, {}, init),
      );
      return normalizeWaybillSyncResult((raw || {}) as RawRecord);
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

    listSLOGuardPolicies: async (
      query?: { carrier_id?: string; enabled?: boolean; limit?: number },
      init?: any,
    ): Promise<LogisticsSLOGuardPolicy[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/slo-guard/policies`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeSLOGuardPolicy);
    },

    upsertSLOGuardPolicy: async (payload: Record<string, any>, init?: any): Promise<LogisticsSLOGuardPolicy> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/slo-guard/policies/${id}` : `${basePath}/slo-guard/policies`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeSLOGuardPolicy((raw || {}) as RawRecord);
    },

    getSLOGuardStatus: async (
      query?: { carrier_id?: string; window_hours?: number },
      init?: any,
    ): Promise<LogisticsSLOGuardStatusSnapshot> => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord>>(`${basePath}/slo-guard/status`, query, init));
      return normalizeSLOGuardStatus((raw || {}) as RawRecord);
    },

    evaluateSLOGuard: async (
      payload?: { carrier_id?: string; window_hours?: number },
      init?: any,
    ): Promise<LogisticsSLOGuardStatusSnapshot> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/slo-guard/evaluate`, payload || {}, init));
      return normalizeSLOGuardStatus((raw || {}) as RawRecord);
    },

    releaseSLOGuardPolicy: async (
      id: string,
      payload?: { operator_id?: string; reason?: string },
      init?: any,
    ): Promise<RawRecord> => {
      return await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/slo-guard/policies/${id}/release`, payload || {}, init));
    },

    listTrackingSyncJobs: async (
      query?: { carrier_id?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsTrackingSyncJob[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/tracking-sync/jobs`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeTrackingSyncJob);
    },

    createTrackingSyncJob: async (
      payload: { carrier_id?: string; waybill_status?: string; batch_limit?: number; event_limit?: number },
      init?: any,
    ): Promise<LogisticsTrackingSyncJob> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tracking-sync/jobs`, payload, init));
      return normalizeTrackingSyncJob((raw || {}) as RawRecord);
    },

    cancelTrackingSyncJob: async (id: string, payload?: { reason?: string }, init?: any): Promise<LogisticsTrackingSyncJob> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tracking-sync/jobs/${id}/cancel`, payload || {}, init));
      return normalizeTrackingSyncJob((raw || {}) as RawRecord);
    },

    retryTrackingSyncJob: async (id: string, init?: any): Promise<LogisticsTrackingSyncJob> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tracking-sync/jobs/${id}/retry`, {}, init));
      return normalizeTrackingSyncJob((raw || {}) as RawRecord);
    },

    getGatewayHealth: async (query?: { window_hours?: number }, init?: any): Promise<LogisticsGatewayHealthSnapshot> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ summary: RawRecord; carriers: RawRecord[]; alerts: RawRecord[] }>>(`${basePath}/gateway/health`, query, init),
      );
      return {
        summary: normalizeGatewayHealthSummary((raw?.summary || {}) as RawRecord),
        carriers: asArray<RawRecord>(raw?.carriers).map(normalizeGatewayCarrierHealth),
        alerts: asArray<RawRecord>(raw?.alerts).map(normalizeGatewayAlert),
      };
    },

    getGatewayCosts: async (
      query?: { window_hours?: number; carrier_id?: string; provider?: string; unit_price?: number; quota_limit?: number },
      init?: any,
    ): Promise<LogisticsGatewayCostSnapshot> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ summary: RawRecord; carriers: RawRecord[]; alerts: RawRecord[] }>>(`${basePath}/gateway/costs`, query, init),
      );
      return {
        summary: normalizeGatewayCostSummary((raw?.summary || {}) as RawRecord),
        carriers: asArray<RawRecord>(raw?.carriers).map(normalizeGatewayCostCarrier),
        alerts: asArray<RawRecord>(raw?.alerts).map(normalizeGatewayCostAlert),
      };
    },

    getGatewayCostAlerts: async (
      query?: { window_hours?: number; carrier_id?: string; provider?: string; unit_price?: number; quota_limit?: number },
      init?: any,
    ): Promise<LogisticsGatewayCostAlert[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/gateway/cost-alerts`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeGatewayCostAlert);
    },

    listTrackingSyncSchedules: async (
      query?: { enabled?: boolean; limit?: number },
      init?: any,
    ): Promise<LogisticsTrackingSyncSchedule[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/tracking-sync/schedules`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeTrackingSyncSchedule);
    },

    upsertTrackingSyncSchedule: async (
      payload: {
        id?: string;
        name: string;
        cron_expr: string;
        carrier_id?: string;
        waybill_status?: string;
        enabled?: boolean;
        max_concurrency?: number;
        dedupe_window_sec?: number;
        batch_limit?: number;
        event_limit?: number;
      },
      init?: any,
    ): Promise<LogisticsTrackingSyncSchedule> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/tracking-sync/schedules/${id}` : `${basePath}/tracking-sync/schedules`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeTrackingSyncSchedule((raw || {}) as RawRecord);
    },

    toggleTrackingSyncSchedule: async (id: string, enabled: boolean, init?: any): Promise<LogisticsTrackingSyncSchedule> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tracking-sync/schedules/${id}/toggle`, { enabled }, init),
      );
      return normalizeTrackingSyncSchedule((raw || {}) as RawRecord);
    },

    triggerTrackingSyncSchedule: async (id: string, init?: any): Promise<LogisticsTrackingSyncJob> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tracking-sync/schedules/${id}/trigger`, {}, init));
      return normalizeTrackingSyncJob((raw || {}) as RawRecord);
    },

    runDueTrackingSyncSchedules: async (query?: { limit?: number }, init?: any): Promise<{ triggered: number }> => {
      const suffix = query?.limit && query.limit > 0 ? `?limit=${query.limit}` : "";
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/tracking-sync/schedules/run-due${suffix}`, {}, init),
      );
      return { triggered: Number(pick((raw || {}) as RawRecord, "triggered", "triggered") || 0) };
    },

    listGatewayFailures: async (
      query?: { carrier_id?: string; waybill_no?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsGatewayFailureEvent[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/gateway/failures`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeGatewayFailureEvent);
    },

    ingestGatewayFailures: async (query?: { hours?: number }, init?: any): Promise<{ created: number }> => {
      const suffix = query?.hours && query.hours > 0 ? `?hours=${query.hours}` : "";
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/gateway/failures/ingest${suffix}`, {}, init));
      return { created: Number(pick((raw || {}) as RawRecord, "created", "created") || 0) };
    },

    compensateGatewayFailure: async (id: string, init?: any): Promise<LogisticsGatewayFailureEvent> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/gateway/failures/${id}/compensate`, {}, init));
      return normalizeGatewayFailureEvent((raw || {}) as RawRecord);
    },

    listExceptionOrchestrationRules: async (query?: { enabled?: boolean }, init?: any): Promise<LogisticsExceptionOrchestrationRule[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/exceptions/orchestration/rules`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeExceptionOrchestrationRule);
    },

    upsertExceptionOrchestrationRule: async (payload: Record<string, any>, init?: any): Promise<LogisticsExceptionOrchestrationRule> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/exceptions/orchestration/rules/${id}` : `${basePath}/exceptions/orchestration/rules`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeExceptionOrchestrationRule((raw || {}) as RawRecord);
    },

    listExceptionOrchestrationRuns: async (
      query?: { waybill_no?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsExceptionOrchestrationRun[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/exceptions/orchestration/runs`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeExceptionOrchestrationRun);
    },

    executeExceptionOrchestration: async (payload: Record<string, any>, init?: any): Promise<LogisticsExceptionOrchestrationRun> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/exceptions/orchestration/execute`, payload, init));
      return normalizeExceptionOrchestrationRun((raw || {}) as RawRecord);
    },

    checkAddressValidation: async (payload: Record<string, any>, init?: any): Promise<LogisticsAddressValidation> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/address-validation/check`, payload, init));
      return normalizeAddressValidation((raw || {}) as RawRecord);
    },

    listAddressValidationRecords: async (
      query?: { waybill_no?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsAddressValidation[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/address-validation/records`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeAddressValidation);
    },

    getRoutingOptimizerStrategy: async (init?: any): Promise<LogisticsRoutingOptimizerStrategy> => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord>>(`${basePath}/routing/optimizer/strategy`, undefined, init));
      return normalizeRoutingOptimizerStrategy((raw || {}) as RawRecord);
    },

    upsertRoutingOptimizerStrategy: async (payload: Record<string, any>, init?: any): Promise<LogisticsRoutingOptimizerStrategy> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/routing/optimizer/strategy`, payload, init));
      return normalizeRoutingOptimizerStrategy((raw || {}) as RawRecord);
    },

    simulateRoutingOptimizer: async (payload: Record<string, any>, init?: any): Promise<LogisticsRoutingOptimizerSimulation> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/routing/optimizer/simulate`, payload, init));
      return normalizeRoutingOptimizerSimulation((raw || {}) as RawRecord);
    },

    listSettlementBatches: async (
      query?: { carrier_id?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsSettlementBatch[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/settlement/batches`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeSettlementBatch);
    },

    createSettlementBatch: async (
      payload: { carrier_id?: string; from?: string; to?: string },
      init?: any,
    ): Promise<LogisticsSettlementBatch> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/settlement/batches`, payload, init));
      return normalizeSettlementBatch((raw || {}) as RawRecord);
    },

    listSettlementDiffs: async (
      query?: { batch_id?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsSettlementDiff[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/settlement/diffs`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeSettlementDiff);
    },

    handleSettlementDiff: async (
      id: string,
      payload: { action: "accept" | "dispute"; operator_id?: string; note?: string },
      init?: any,
    ): Promise<LogisticsSettlementDiff> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/settlement/diffs/${id}/handle`, payload, init));
      return normalizeSettlementDiff((raw || {}) as RawRecord);
    },

    confirmSettlementBatch: async (id: string, init?: any): Promise<LogisticsSettlementBatch> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/settlement/batches/${id}/confirm`, {}, init));
      return normalizeSettlementBatch((raw || {}) as RawRecord);
    },

    listReconciliationBatches: async (
      query?: { carrier_id?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsReconciliationBatch[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/reconciliation/batches`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeReconciliationBatch);
    },

    createReconciliationBatch: async (
      payload: { carrier_id?: string; from?: string; to?: string; records?: Record<string, any>[] },
      init?: any,
    ): Promise<LogisticsReconciliationBatch> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/reconciliation/batches`, payload, init));
      return normalizeReconciliationBatch((raw || {}) as RawRecord);
    },

    listReconciliationRecords: async (
      query?: { batch_id?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsReconciliationRecord[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/reconciliation/records`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeReconciliationRecord);
    },

    listReconciliationCases: async (
      query?: { batch_id?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsReconciliationCase[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/reconciliation/cases`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeReconciliationCase);
    },

    handleReconciliationCase: async (
      id: string,
      payload: { action: "confirm" | "appeal" | "close" | "manual_review"; operator_id?: string; note?: string },
      init?: any,
    ): Promise<LogisticsReconciliationCase> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/reconciliation/cases/${id}/handle`, payload, init));
      return normalizeReconciliationCase((raw || {}) as RawRecord);
    },

    getControlTowerOverview: async (
      query?: { carrier_id?: string; warehouse_id?: string; destination_zone?: string; window_hours?: number },
      init?: any,
    ): Promise<LogisticsControlTowerOverview> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ summary: RawRecord; alerts: RawRecord[] }>>(`${basePath}/control-tower/overview`, query, init),
      );
      return {
        summary: normalizeControlTowerSummary((raw?.summary || {}) as RawRecord),
        alerts: asArray<RawRecord>(raw?.alerts).map(normalizeControlTowerAlert),
      };
    },

    getControlTowerDrilldown: async (
      query?: {
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
        status?: string;
        window_hours?: number;
        limit?: number;
      },
      init?: any,
    ): Promise<LogisticsControlTowerDrilldownItem[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/control-tower/drilldown`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeControlTowerDrilldownItem);
    },

    listControlTowerSubscriptions: async (
      query?: { enabled?: boolean },
      init?: any,
    ): Promise<LogisticsControlTowerSubscription[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/control-tower/subscriptions`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeControlTowerSubscription);
    },

    upsertControlTowerSubscription: async (payload: Record<string, any>, init?: any): Promise<LogisticsControlTowerSubscription> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/control-tower/subscriptions/${id}` : `${basePath}/control-tower/subscriptions`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeControlTowerSubscription((raw || {}) as RawRecord);
    },

    getKPIDashboardOverview: async (
      query?: {
        dimension?: "tenant" | "carrier" | "warehouse" | "destination_zone";
        window_hours?: number;
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
      },
      init?: any,
    ): Promise<LogisticsKPIDashboardOverview> => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord>>(`${basePath}/kpi-dashboard/overview`, query, init));
      return normalizeKPIDashboardOverview((raw || {}) as RawRecord);
    },

    getKPIDashboardTrends: async (
      query?: {
        dimension?: "tenant" | "carrier" | "warehouse" | "destination_zone";
        window_hours?: number;
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
      },
      init?: any,
    ): Promise<LogisticsKPIDashboardTrend[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/kpi-dashboard/trends`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeKPIDashboardTrend);
    },

    getKPIDashboardDrilldown: async (
      query?: {
        dimension?: "tenant" | "carrier" | "warehouse" | "destination_zone";
        window_hours?: number;
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
        limit?: number;
      },
      init?: any,
    ): Promise<LogisticsKPIDashboardDrilldown[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/kpi-dashboard/drilldown`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeKPIDashboardDrilldown);
    },

    exportKPIDashboard: async (
      query?: {
        dimension?: "tenant" | "carrier" | "warehouse" | "destination_zone";
        window_hours?: number;
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
      },
      init?: any,
    ): Promise<{ content: string; format: string }> => {
      const raw = await unwrap(apiGet<ApiEnvelope<RawRecord>>(`${basePath}/kpi-dashboard/export`, query, init));
      return {
        content: String(pick((raw || {}) as RawRecord, "content", "content") || ""),
        format: String(pick((raw || {}) as RawRecord, "format", "format") || "csv"),
      };
    },

    listAllocationPlans: async (
      query?: { carrier_id?: string; warehouse_id?: string; destination_zone?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsCapacityPlan[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/allocation/plans`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeCapacityPlan);
    },

    upsertAllocationPlan: async (payload: Record<string, any>, init?: any): Promise<LogisticsCapacityPlan> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/allocation/plans/${id}` : `${basePath}/allocation/plans`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeCapacityPlan((raw || {}) as RawRecord);
    },

    allocateCarrier: async (payload: Record<string, any>, init?: any): Promise<LogisticsAllocationResult> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/allocation/allocate`, payload, init));
      return normalizeAllocationResult((raw || {}) as RawRecord);
    },

    overrideAllocation: async (payload: Record<string, any>, init?: any): Promise<LogisticsAllocationResult> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/allocation/override`, payload, init));
      return normalizeAllocationResult((raw || {}) as RawRecord);
    },

    listCapacityForecasts: async (
      query?: {
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
        status?: "suggested" | "applied" | "dismissed" | string;
        limit?: number;
      },
      init?: any,
    ): Promise<LogisticsCapacityForecast[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/allocation/forecasts`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCapacityForecast);
    },

    generateCapacityForecast: async (
      payload: {
        carrier_id?: string;
        warehouse_id?: string;
        destination_zone?: string;
        window_days?: number;
        operator_id?: string;
      },
      init?: any,
    ): Promise<LogisticsCapacityForecast[]> => {
      const raw = await unwrap(apiPost<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/allocation/forecasts/generate`, payload, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCapacityForecast);
    },

    applyCapacityForecast: async (
      id: string,
      payload?: { operator_id?: string },
      init?: any,
    ): Promise<LogisticsCapacityForecast> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/allocation/forecasts/${id}/apply`, payload || {}, init));
      return normalizeCapacityForecast((raw || {}) as RawRecord);
    },

    listLastmileRecoveryRules: async (query?: { enabled?: boolean }, init?: any): Promise<LogisticsLastmileRecoveryRule[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/lastmile-recovery/rules`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeLastmileRecoveryRule);
    },

    upsertLastmileRecoveryRule: async (payload: Record<string, any>, init?: any): Promise<LogisticsLastmileRecoveryRule> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/lastmile-recovery/rules/${id}` : `${basePath}/lastmile-recovery/rules`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeLastmileRecoveryRule((raw || {}) as RawRecord);
    },

    executeLastmileRecovery: async (
      payload: Record<string, any>,
      init?: any,
    ): Promise<{ run: LogisticsLastmileRecoveryRun; idempotencyStatus: string }> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ run: RawRecord; idempotency_status: string }>>(`${basePath}/lastmile-recovery/execute`, payload, init),
      );
      return {
        run: normalizeLastmileRecoveryRun((raw?.run || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    listLastmileRecoveryRuns: async (
      query?: { waybill_no?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsLastmileRecoveryRun[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/lastmile-recovery/runs`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeLastmileRecoveryRun);
    },

    takeoverLastmileRecovery: async (id: string, payload: Record<string, any>, init?: any): Promise<LogisticsLastmileRecoveryRun> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<RawRecord>>(`${basePath}/lastmile-recovery/runs/${id}/takeover`, payload, init),
      );
      return normalizeLastmileRecoveryRun((raw || {}) as RawRecord);
    },

    listCrossborderDocuments: async (
      query?: { waybill_no?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsCrossborderDocument[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/crossborder/documents`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCrossborderDocument);
    },

    upsertCrossborderDocument: async (payload: Record<string, any>, init?: any): Promise<LogisticsCrossborderDocument> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/crossborder/documents/${id}` : `${basePath}/crossborder/documents`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeCrossborderDocument((raw || {}) as RawRecord);
    },

    quoteCrossborderTax: async (
      payload: Record<string, any>,
      init?: any,
    ): Promise<{ quote: LogisticsCrossborderTaxQuote; idempotencyStatus: string }> => {
      const raw = await unwrap(
        apiPost<ApiEnvelope<{ quote: RawRecord; idempotency_status: string }>>(`${basePath}/crossborder/tax-quote`, payload, init),
      );
      return {
        quote: normalizeCrossborderTaxQuote((raw?.quote || {}) as RawRecord),
        idempotencyStatus: String(raw?.idempotency_status || "created"),
      };
    },

    listCrossborderTrackingMaps: async (
      query?: { provider?: string; enabled?: boolean; limit?: number },
      init?: any,
    ): Promise<LogisticsCrossborderTrackingMap[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/crossborder/tracking-maps`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCrossborderTrackingMap);
    },

    upsertCrossborderTrackingMap: async (payload: Record<string, any>, init?: any): Promise<LogisticsCrossborderTrackingMap> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/crossborder/tracking-maps/${id}` : `${basePath}/crossborder/tracking-maps`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeCrossborderTrackingMap((raw || {}) as RawRecord);
    },

    normalizeCrossborderTracking: async (
      payload: { provider: string; provider_status: string },
      init?: any,
    ): Promise<{ normalizedStatus: string; source: string }> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/crossborder/tracking-maps/normalize`, payload, init));
      return {
        normalizedStatus: String(pick((raw || {}) as RawRecord, "normalizedStatus", "normalized_status") || ""),
        source: String(pick((raw || {}) as RawRecord, "source", "source") || ""),
      };
    },

    listCustomsRulePacks: async (
      query?: { country_code?: string; status?: string; limit?: number },
      init?: any,
    ): Promise<LogisticsCustomsRulePack[]> => {
      const raw = await unwrap(apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/customs/rule-packs`, query, init));
      return asArray<RawRecord>(raw?.items).map(normalizeCustomsRulePack);
    },

    upsertCustomsRulePack: async (payload: Record<string, any>, init?: any): Promise<LogisticsCustomsRulePack> => {
      const id = String(payload.id || "").trim();
      const path = id ? `${basePath}/customs/rule-packs/${id}` : `${basePath}/customs/rule-packs`;
      const req = id ? apiPatch<ApiEnvelope<RawRecord>>(path, payload, init) : apiPost<ApiEnvelope<RawRecord>>(path, payload, init);
      const raw = await unwrap(req);
      return normalizeCustomsRulePack((raw || {}) as RawRecord);
    },

    listCustomsRuleVersions: async (packId: string, query?: { limit?: number }, init?: any): Promise<LogisticsCustomsRuleVersion[]> => {
      const raw = await unwrap(
        apiGet<ApiEnvelope<{ items: RawRecord[] }>>(`${basePath}/customs/rule-packs/${packId}/versions`, query, init),
      );
      return asArray<RawRecord>(raw?.items).map(normalizeCustomsRuleVersion);
    },

    publishCustomsRuleVersion: async (
      packId: string,
      payload: Record<string, any>,
      init?: any,
    ): Promise<LogisticsCustomsRuleVersion> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/customs/rule-packs/${packId}/versions`, payload, init));
      return normalizeCustomsRuleVersion((raw || {}) as RawRecord);
    },

    customsPrecheck: async (payload: Record<string, any>, init?: any): Promise<LogisticsCustomsPrecheckResult> => {
      const raw = await unwrap(apiPost<ApiEnvelope<RawRecord>>(`${basePath}/customs/precheck`, payload, init));
      return normalizeCustomsPrecheckResult((raw || {}) as RawRecord);
    },
  };
}
