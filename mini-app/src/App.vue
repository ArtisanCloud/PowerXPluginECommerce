<script setup lang="ts">
import { onLaunch, onShow, onHide, onError, onUnhandledRejection } from "@dcloudio/uni-app";
onLaunch(() => {
  console.log("App Launch");
});
onShow(() => {
  console.log("App Show");
});
onHide(() => {
  console.log("App Hide");
});

const isWechatGuestModeError = (msg: string) => {
  const text = String(msg || "");
  return (
    text.includes("webapi_getwxaasyncsecinfo") ||
    text.includes("operateWXData") ||
    text.includes("appServiceSDKScriptError")
  );
};

onError((err) => {
  const msg = String(err || "");
  if (isWechatGuestModeError(msg)) {
    console.warn("[miniapp] 微信开发者工具游客模式限制：", msg);
    return;
  }
  console.error(err);
});

onUnhandledRejection((res: any) => {
  const reason = res?.reason ?? res;
  const msg = String(reason || "");
  if (isWechatGuestModeError(msg)) {
    console.warn("[miniapp] 微信开发者工具游客模式限制：", msg);
    return;
  }
  console.error(reason);
});
</script>
<style>
@import "@/styles/tailwind.css";
</style>
