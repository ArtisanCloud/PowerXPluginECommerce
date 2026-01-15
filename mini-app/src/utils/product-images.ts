// 说明：在尚未接入商品头图/详图前，保持“之前”的占位图（外网素材）。
// 如果后续需要适配微信域名白名单，再统一切回本地静态占位图。
const placeholderImages = [
  "https://lh3.googleusercontent.com/aida-public/AB6AXuCg6FsSWqWzgBZu6SO_586WS1dZZSLI_4FZg3AkMAKVJBgwQ1TP64nNujeYl2F1Cy90DXslNyCZ3luU22WAp7EqKZcNRwDUcXyekZ1SJNfXz-Ng326aaor1EkIa8HQ2NKnTIRtiKSk9TyqYy3QxFGiGxXGV9cAzNv28dpANZW3OciBxNEUHYUDWQcsHHtiV9SYBMiSywXyQKu9vZXIAN50f7HHDglzRaVYMhva6Z6z6nu3pJJZXqkgqbUXCT5ioqcj1mt58pkTc-RLD",
  "https://lh3.googleusercontent.com/aida-public/AB6AXuBiNZa5UJR3RHZr_Vv8Re2D4zWPjBB4XtBV3lAj2ONg6WWhTwRbRLNQyd-xk2kpqkSOyqtRKbcNTKWCMrTocj_4IP_VvgkD1xThCVs9A9ppYiNT2HXG60wMqEbWRtKz9FGMu7VRB4GWy4A9orsa-b0163c_R6bUxiiybfMM2Lop90aSwU0d3eePRmoYR_XBOop54DRn1EBnHbB5I6ZaEM5h0XYcOEfecyhsQI8BfIWiBwzTZSvXvnMzVN1JtV90Ohi9BClg3UQH4YPI",
  "https://lh3.googleusercontent.com/aida-public/AB6AXuBYw5AQ01Azh1QEqbFKaklXCWDwMoIdv8HElP-whtGag--GXku6J3a6JO2sDY4dh3dKDgVUTjFnsD7J69s0WftgLV3ylA7zvRe9n3IbHcFtzQ1KmYxuriZZwLk-Z31iJxQUGnzPoOzoBXK-t-Hkf8AR8Ttzb0rWNPt7M32Dk93AGBLtI8XieV27ASYobj8qQgTzGW3LnEiyoZHjDr3pmrap_dFPgHoGDIZ8WC7KnIkvGeZJ9wAaHyYzXbLcRjxIvM58pdfeAr8sImA_",
  "https://lh3.googleusercontent.com/aida-public/AB6AXuC0NG4gIgXiKUK5XTBGTd0dX15JofDOGcpTEAqoNpcUMWX8CSRoeObkzmbjrOJF-dNM3EVyb7i-nHoGZ8PSek3sW2Mujt4SgWcgMh4DLUFDaC6yScZqkT-NL1mQIcjy4sepye-i2LLgcBwONBqN-AkhM9K7B06-Q8nrnYYNyquZuUQkY6N8QcoFva-J0bBvsFy3LDc_kcS9cIxOIHGFor1-nzYlJ_ukDYnRBzEjXCQ3v8sXuvRdGaZYy_-wBVtTn1pXEsA9JoTuexlE",
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
  return false;
}
