import type {
  MembershipFilters,
  MembershipInsight,
  MembershipSegments,
  MembershipSnapshot,
  MembershipStats,
} from "~/types/customer";
import { useCustomerService } from "./api/services/customerService";
import { useCustomerMetrics } from "./useCustomerMetrics";

interface MembershipQueryResult {
  insights: MembershipInsight[];
  snapshots: MembershipSnapshot[];
  stats: MembershipStats;
  segments: MembershipSegments;
  total: number;
}

const defaultSegments = (): MembershipSegments => ({
  safe: 0,
  warning: 0,
  downgrade: 0,
});

const normalizeSnapshot = (
  snapshot: MembershipSnapshot,
  fallbackId?: string
): MembershipSnapshot => {
  return {
    ...snapshot,
    customerId: snapshot.customerId || fallbackId || "",
    retentionStatus: resolveRetentionStatus(snapshot),
  };
};

const resolveRetentionStatus = (
  snapshot: MembershipSnapshot
): MembershipSnapshot["retentionStatus"] => {
  if (snapshot.retentionStatus) {
    return snapshot.retentionStatus;
  }
  const growthValue = typeof snapshot.growthValue === "number" ? snapshot.growthValue : 0;
  if (growthValue < 30) return "downgrade";
  if (growthValue < 60) return "warning";
  return "safe";
};

const aggregateSegments = (snapshots: MembershipSnapshot[]): MembershipSegments => {
  return snapshots.reduce<MembershipSegments>((acc, current) => {
    const status = resolveRetentionStatus(current);
    acc[status] = (acc[status] ?? 0) + 1;
    return acc;
  }, defaultSegments());
};

const aggregateStats = (
  snapshots: MembershipSnapshot[],
  fallbackTotal?: number
): MembershipStats => {
  const segments = aggregateSegments(snapshots);
  let growthSum = 0;
  let growthCount = 0;
  snapshots.forEach((snapshot) => {
    if (typeof snapshot.growthValue === "number") {
      growthSum += snapshot.growthValue;
      growthCount += 1;
    }
  });
  const averageGrowthValue =
    growthCount > 0 ? Math.round((growthSum / growthCount) * 10) / 10 : 0;

  return {
    total: fallbackTotal ?? snapshots.length,
    active: segments.safe,
    warning: segments.warning,
    downgrade: segments.downgrade,
    averageGrowthValue,
  };
};

export const useMembershipInsights = () => {
  const service = useCustomerService();
  const metrics = useCustomerMetrics();

  const fetchInsights = async (
    filters?: MembershipFilters
  ): Promise<MembershipQueryResult> => {
    const stopTimer = metrics.startLatencyTimer("customer_membership_fetch", {
      source: "members_page",
    });
    try {
      const response = await service.listMembers(filters);
      const insights =
        (response.data || []).map((entry) => ({
          customer: entry.customer,
          snapshot: normalizeSnapshot(entry.snapshot, entry.customer?.id),
        })) ?? [];
      const snapshots = insights.map((entry) => entry.snapshot);
      const fallback = aggregateStats(snapshots, response.meta?.total);
      const stats = response.stats
        ? {
            total: response.stats.total ?? fallback.total,
            active: response.stats.active ?? fallback.active,
            warning: response.stats.warning ?? fallback.warning,
            downgrade: response.stats.downgrade ?? fallback.downgrade,
            averageGrowthValue:
              response.stats.averageGrowthValue ?? fallback.averageGrowthValue,
          }
        : fallback;
      const segments = aggregateSegments(snapshots);
      return {
        insights,
        snapshots,
        stats,
        segments,
        total: response.meta?.total ?? snapshots.length,
      };
    } finally {
      stopTimer();
    }
  };

  return {
    fetchInsights,
    aggregateStats,
    aggregateSegments,
  };
};
