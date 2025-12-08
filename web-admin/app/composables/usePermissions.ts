import { computed } from "vue";
import { useAuth } from "./useAuth";

const normalizePermissions = (input: unknown): string[] => {
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

export const usePermissions = () => {
  const auth = useAuth();
  const permissionSet = computed(() => {
    const current = new Set<string>();
    const user = auth.user.value as Record<string, any> | null;
    if (!user) return current;
    const sources = [
      user.permissions,
      user.permissionCodes,
      user.permission_codes,
      user.permission_ids,
    ];
    sources.forEach((src) => {
      normalizePermissions(src).forEach((code) => current.add(code));
    });
    return current;
  });

  const hasPermission = (code?: string) => {
    if (!code) return true;
    const set = permissionSet.value;
    if (set.size === 0) {
      // 缺省情况下放行，避免本地开发阻塞
      return true;
    }
    return (
      set.has(code) ||
      set.has("*") ||
      set.has("all") ||
      set.has("admin")
    );
  };

  return {
    permissions: permissionSet,
    hasPermission,
  };
};
