export type CustomerType = "individual" | "enterprise";
export type CustomerStatus = "active" | "inactive" | "blocked";
export type BulkActionType =
  | "add-tags"
  | "remove-tags"
  | "assign-owner"
  | "bulk-remind"
  | "bulk-disable"
  | "bulk-enable";

export interface Customer {
  id: string;
  name: string;
  type: CustomerType;
  email?: string;
  phone?: string;
  country?: string;
  region?: string;
  membershipTier?: string;
  membershipTierLabel?: string;
  growthValue?: number;
  points?: number;
  lastOrderAmount?: number;
  lastOrderAt?: string;
  status?: CustomerStatus;
  riskLevel?: string;
  source?: string;
  accountManager?: string;
  tags?: string[];
  createdAt?: string;
  updatedAt?: string;
  maskedFields?: string[];
  membershipSnapshot?: MembershipSnapshot;
  metadata?: Record<string, any>;
}

export interface SavedView {
  id: string;
  name: string;
  filters: CustomerListFilters;
  columns?: string[];
  owner: string;
  shared?: boolean;
  createdAt: string;
}

export interface CustomerListMeta {
  total: number;
  savedViewId?: string | null;
}

export interface CustomerListResponse {
  data: Customer[];
  meta: CustomerListMeta;
}

export interface MembershipSnapshot {
  customerId: string;
  tier: string;
  growthValue: number;
  points: number;
  retentionStatus: "safe" | "warning" | "downgrade";
  benefits?: Array<{ name: string; used: boolean }>;
  lastBenefitUsedAt?: string | null;
}

export interface MembershipInsight {
  customer: Customer;
  snapshot: MembershipSnapshot;
}

export interface MembershipListResponse {
  data: MembershipInsight[];
  meta: {
    total: number;
  };
  stats?: MembershipStats;
}

export interface MembershipStats {
  total: number;
  active: number;
  warning: number;
  downgrade: number;
  averageGrowthValue: number;
}

export interface MembershipSegments {
  safe: number;
  warning: number;
  downgrade: number;
}

export interface CustomerListFilters {
  keyword?: string;
  tier?: string;
  source?: string;
  type?: CustomerType;
  tags?: string[];
  region?: string;
  riskLevel?: string;
  createdFrom?: string;
  createdTo?: string;
  growthRange?: [number, number];
  pointsRange?: [number, number];
  benefitStatus?: string;
  retentionStatus?: string;
  page?: number;
  pageSize?: number;
  sort?: string;
}

export interface MembershipFilters {
  tier?: string;
  growthRange?: [number, number];
  pointsRange?: [number, number];
  benefitStatus?: string;
  retentionStatus?: string;
  page?: number;
  pageSize?: number;
}

export interface MembershipReminderState {
  submitting: boolean;
  error: string | null;
  lastTaskId: string | null;
  channel: string;
  templateId: string;
  total: number;
  success: number;
}

export interface BulkActionPayload {
  action: BulkActionType;
  ids: string[];
  payload?: Record<string, any>;
}

export interface BulkReminderPayload {
  ids: string[];
  channel: string;
  templateId: string;
  metadata?: Record<string, any>;
}

export interface CustomerExportPayload {
  filters?: CustomerListFilters;
  fields?: string[];
}

export interface JobStatus {
  taskId: string;
  status: "queued" | "running" | "success" | "failed";
  message?: string;
  downloadUrl?: string;
  completedAt?: string;
  createdAt?: string;
  metadata?: Record<string, any>;
}

export interface BulkTask extends JobStatus {
  type: BulkActionType | "import" | "export" | "reminder";
  scope?: {
    ids?: string[];
    filters?: CustomerListFilters;
  };
}
