import type {
  PowerXCapabilityBridge,
  PowerXCapabilityRequest,
  PowerXCapabilityResponse,
} from "~/composables/powerxCapabilityTypes";

export function usePowerXCapability() {
  const nuxtApp = useNuxtApp();
  const bridge = nuxtApp.$powerxCapability as PowerXCapabilityBridge | undefined;

  if (!bridge) {
    throw new Error("PowerX capability bridge is not initialized.");
  }

  const invoke = (request: PowerXCapabilityRequest): Promise<PowerXCapabilityResponse> => {
    return bridge.invoke(request);
  };

  return {
    invoke,
  };
}
