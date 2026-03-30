import { defineStore } from "pinia";
import { useAuthService } from "~/composables/api/services/authService";

type TemplatesCapability = {
  can_create: boolean;
  can_update: boolean;
  can_delete: boolean;
};

type UserContext = {
  is_root?: boolean;
  current_tenant_uuid?: string;
  roles?: string[];
  permissions?: string[];
  capabilities?: {
    templates?: Partial<TemplatesCapability>;
  };
  [key: string]: any;
};

const normalizeCodes = (input: unknown): string[] => {
  if (!input) return [];
  if (Array.isArray(input)) {
    return input
      .filter((item) => typeof item === "string" && item.trim().length > 0)
      .map((item) => item.trim());
  }
  if (typeof input === "string") {
    return input
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean);
  }
  return [];
};

export const useUserStore = defineStore("user", () => {
  const profile = ref<any>(null);
  const context = ref<UserContext | null>(null);

  const setUser = (data: any) => {
    profile.value = data;
  };

  const setContext = (data: UserContext | null) => {
    context.value = data;
  };

  const updateCurrentTenantUUID = (tenantUUID: string) => {
    const next = String(tenantUUID || "").trim();
    if (!next) return;
    context.value = {
      ...(context.value || {}),
      current_tenant_uuid: next,
    };
  };

  const switchTenant = async (tenantUUID: string) => {
    const next = String(tenantUUID || "").trim();
    if (!next) return;
    try {
      const { switchTenant: requestSwitch } = useAuthService();
      const response = await requestSwitch({ tenant_uuid: next });
      if (response?.success && response.data) {
        setContext(response.data);
        if (response.data.user) {
          setUser(response.data.user);
        }
        return;
      }
      updateCurrentTenantUUID(next);
    } catch (error: any) {
      const status = error?.response?.status;
      if (status === 404) {
        // 兼容宿主未实现 switch-tenant：仅更新当前租户，不清空上下文。
        updateCurrentTenantUUID(next);
        return;
      }
      throw error;
    }
  };

  const templatesCapability = computed<TemplatesCapability>(() => {
    const ctx = context.value || {};
    const roles = normalizeCodes(ctx.roles);
    const permissions = normalizeCodes(ctx.permissions);
    const explicit = ctx.capabilities?.templates || {};
    const isRoot = Boolean(ctx.is_root);
    const canByPermission =
      isRoot ||
      permissions.includes("base.templates.manage") ||
      roles.some((role) =>
        ["system.admin", "superadmin", "admin", "tenant.admin", "template.admin"].includes(role.toLowerCase())
      );
    return {
      can_create: explicit.can_create ?? canByPermission,
      can_update: explicit.can_update ?? canByPermission,
      can_delete: explicit.can_delete ?? canByPermission,
    };
  });

  const clearUserState = () => {
    profile.value = null;
    context.value = null;
  };

  return {
    profile,
    context,
    templatesCapability,
    setUser,
    setContext,
    updateCurrentTenantUUID,
    switchTenant,
    clearUserState,
  };
});
