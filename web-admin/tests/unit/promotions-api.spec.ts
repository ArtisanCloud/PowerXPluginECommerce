import { beforeEach, describe, expect, it, vi } from "vitest";

const apiGet = vi.fn();
const apiPost = vi.fn();
const apiPatch = vi.fn();

vi.mock("../../app/composables/api/_client", () => ({
  apiGet,
  apiPost,
  apiPatch,
}));

describe("usePromotionsApi", () => {
  beforeEach(() => {
    apiGet.mockReset();
    apiPost.mockReset();
    apiPatch.mockReset();
  });

  it("maps promotion management endpoints and unwraps envelopes", async () => {
    const { usePromotionsApi } = await import("../../app/composables/api/usePromotions");
    const api = usePromotionsApi();

    apiGet.mockResolvedValueOnce({ success: true, data: { items: [], total: 0, page: 1, page_size: 20 } });
    await expect(api.list({ status: "active", page: 1 })).resolves.toEqual({
      items: [],
      total: 0,
      page: 1,
      page_size: 20,
    });
    expect(apiGet).toHaveBeenCalledWith("/admin/promotions", { status: "active", page: 1 }, undefined);

    const payload = {
      name: "满减",
      condition_rule: { min_order_amount_minor: 1000 },
      scope_rule: { scope_type: "all" },
      action_rule: { discount_amount_minor: 100 },
      stacking_rule: { priority: 100, stackable: true, stackable_with_coupon: true },
      valid_from: "2026-06-10T00:00:00Z",
      valid_to: "2026-06-11T00:00:00Z",
    };
    apiPost.mockResolvedValueOnce({ data: { id: "promo-1" } });
    await expect(api.create(payload)).resolves.toEqual({ id: "promo-1" });
    expect(apiPost).toHaveBeenCalledWith("/admin/promotions", payload, undefined);

    apiPatch.mockResolvedValueOnce({ data: { id: "promo-1", name: "更新后" } });
    await api.update("promo-1", { name: "更新后" });
    expect(apiPatch).toHaveBeenCalledWith("/admin/promotions/promo-1", { name: "更新后" }, undefined);
  });

  it("maps status, clone, and audit log endpoints", async () => {
    const { usePromotionsApi } = await import("../../app/composables/api/usePromotions");
    const api = usePromotionsApi();

    apiPost.mockResolvedValue({ data: { id: "promo-1" } });
    await api.activate("promo-1");
    await api.pause("promo-1");
    await api.clone("promo-1");

    expect(apiPost).toHaveBeenNthCalledWith(1, "/admin/promotions/promo-1/activate", {}, undefined);
    expect(apiPost).toHaveBeenNthCalledWith(2, "/admin/promotions/promo-1/pause", {}, undefined);
    expect(apiPost).toHaveBeenNthCalledWith(3, "/admin/promotions/promo-1/clone", {}, undefined);

    apiGet.mockResolvedValueOnce({ data: { items: [], total: 0 } });
    await api.auditLogs("promo-1", { page: 2, page_size: 10 });
    expect(apiGet).toHaveBeenCalledWith(
      "/admin/promotions/promo-1/audit-logs",
      { page: 2, page_size: 10 },
      undefined,
    );
  });
});
