import { useAuth } from "~/composables/useAuth";
import { useAuthService } from "~/composables/api/services/authService";
import { useUserStore } from "~/stores/user";

const PUBLIC_ROUTE_PREFIXES = ["/users"];

export default defineNuxtRouteMiddleware(async (to) => {
  if (!process.client) return;

  if (to.meta?.public === true) {
    return;
  }

  if (PUBLIC_ROUTE_PREFIXES.some((prefix) => to.path.startsWith(prefix))) {
    return;
  }

  const auth = useAuth();
  await auth.ensureFreshToken();

  if (!auth.token.value) {
    return navigateTo({
      path: "/users/login",
      query: { redirect: to.fullPath },
    });
  }

  const userStore = useUserStore();
  const lastContextToken = useState<string>("auth.contextToken", () => "");
  const token = auth.token.value || "";
  if (!token) return;
  if (userStore.context && lastContextToken.value === token) return;

  try {
    const { getMeContext } = useAuthService();
    const response = await getMeContext();
    if (response?.success && response.data) {
      userStore.setContext(response.data);
      userStore.setUser(response.data.user || null);
      lastContextToken.value = token;
    }
  } catch (error: any) {
    console.warn("[auth] load me/context failed", error?.message || error);
  }
});
