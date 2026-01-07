declare function getCurrentPages(): any[];

export function syncTabBarSelected(pagePath: string) {
  const normalized = String(pagePath || "").replace(/^\//, "");
  if (!normalized) return;
  try {
    // #ifdef MP-WEIXIN
    const pages = getCurrentPages();
    const current = pages && pages.length ? pages[pages.length - 1] : null;
    const tabbar = current && typeof current.getTabBar === "function" ? current.getTabBar() : null;
    if (tabbar && typeof tabbar.setData === "function") {
      tabbar.setData({ selectedPath: normalized });
    }
    // #endif
  } catch {
    // ignore
  }
}

