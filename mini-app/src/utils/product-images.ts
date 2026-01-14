// 微信小程序对图片域名有白名单限制；为了保证“种子/占位图”能稳定显示，这里使用本地静态资源。
const placeholderImages = [
  "/static/logo.png",
  "/static/tabbar/home.png",
  "/static/tabbar/catalog.png",
  "/static/tabbar/cart.png",
];

export function pickPlaceholderImage(key: string) {
  const s = String(key || "");
  let sum = 0;
  for (let i = 0; i < s.length; i++) sum = (sum + s.charCodeAt(i)) % 997;
  return placeholderImages[sum % placeholderImages.length];
}

export function isLikelyPlaceholderUrl(url?: string) {
  const s = String(url || "").trim().toLowerCase();
  if (!s) return false;
  // 后端/前端常用的占位图域名（在微信环境也经常未加入白名单导致加载失败）
  if (s.includes("picsum.photos/")) return true;
  if (s.includes("googleusercontent.com/")) return true;
  return false;
}
