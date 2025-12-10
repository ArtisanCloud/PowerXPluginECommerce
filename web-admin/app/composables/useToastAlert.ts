import { nextTick } from "vue";

export type ToastColor = "primary" | "secondary" | "success" | "info" | "warning" | "error" | "neutral";
export type ToastVariant = "soft" | "solid" | "outline";

type ToastState = {
  visible: boolean;
  title: string;
  message: string;
  color: ToastColor;
  variant: ToastVariant;
  duration: number;
  icon: string | null;
};

export type ToastOptions = {
  title?: string;
  description?: string;
  message?: string;
  color?: ToastColor;
  variant?: ToastVariant;
  duration?: number;
  icon?: string | null;
};

const defaultToastState = (): ToastState => ({
  visible: false,
  title: "",
  message: "",
  color: "primary",
  variant: "soft",
  duration: 3000,
  icon: null,
});

export const useToastAlertState = () =>
  useState<ToastState>("toast-alert-state", () => defaultToastState());

export const useToastAlert = () => {
  const toast = useToastAlertState();

  const add = (options: ToastOptions) => {
    toast.value.title = options.title ?? "";
    toast.value.message = options.description ?? options.message ?? "";
    toast.value.color = options.color ?? "primary";
    toast.value.variant = options.variant ?? "soft";
    toast.value.duration = options.duration ?? 3000;
    toast.value.icon = options.icon ?? null;
    toast.value.visible = false;
    nextTick(() => {
      toast.value.visible = Boolean(toast.value.title || toast.value.message);
    });
  };

  const close = () => {
    toast.value.visible = false;
  };

  const reset = () => {
    toast.value = { ...defaultToastState() };
  };

  return {
    state: toast,
    add,
    close,
    reset,
  };
};
