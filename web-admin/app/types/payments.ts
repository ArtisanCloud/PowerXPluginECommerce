export type PaymentProvider = {
  id: number;
  name: string;
  type: string;
  status: string;
  feeRate: number;
  settlementCycle: string;
  currency: string;
  updatedAt: string;
};

export type PaymentTransaction = {
  id: number;
  transactionNo: string;
  orderId: string;
  orderNo: string;
  providerId: number;
  payMethod: string;
  amountTotal: number;
  amountCurrency: string;
  feeAmount: number;
  status: string;
  createdAt: string;
  completedAt?: string | null;
  failureReason?: string;
};

export type PaymentRefund = {
  id: number;
  refundNo: string;
  refundAmount: number;
  refundCurrency: string;
  status: string;
  createdAt: string;
  completedAt?: string | null;
};

export type PaymentRiskEvent = {
  id: number;
  riskType: string;
  riskScore: number;
  action: string;
  createdAt: string;
  resolvedAt?: string | null;
};

export type PaymentSplitResult = {
  id: number;
  ruleId: number;
  participant: string;
  amount: number;
  status: string;
  createdAt: string;
};

export type PaymentReconciliation = {
  id: number;
  periodType: string;
  periodStart: string;
  periodEnd: string;
  diffCount: number;
  diffTotalAmount: number;
  status: string;
  processedBy?: string;
  processedAt?: string | null;
  createdAt: string;
};

export type PaymentReconciliationItem = {
  id: number;
  reconciliationId: number;
  transactionId: number;
  diffType: string;
  diffAmount: number;
  resolution?: string;
  resolvedBy?: string;
  resolvedAt?: string | null;
  createdAt: string;
};

export type ManualPaymentReview = {
  id: number;
  orderId: string;
  orderNo: string;
  transactionId?: number | null;
  providerId: number;
  payMethod: string;
  amountMinor: number;
  currency: string;
  status: string;
  submittedBy: string;
  submittedAt: string;
  reviewedBy?: string;
  reviewedAt?: string | null;
  reviewReason?: string;
  proofNo?: string;
  note?: string;
  createdAt: string;
  updatedAt: string;
};
