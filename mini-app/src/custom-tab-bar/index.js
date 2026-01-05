Component({
  data: {
    selectedPath: "pages/index/index",
    color: "#8a8a8a",
    selectedColor: "#4F8A7E",
    list: [
      {
        pagePath: "pages/index/index",
        text: "首页",
        iconPath: "/static/tabbar/home.svg",
        selectedIconPath: "/static/tabbar/home-active.svg",
      },
      {
        pagePath: "pages/mall/index",
        text: "商城",
        iconPath: "/static/tabbar/catalog.svg",
        selectedIconPath: "/static/tabbar/catalog-active.svg",
      },
      {
        pagePath: "pages/cart/index",
        text: "购物车",
        iconPath: "/static/tabbar/cart.svg",
        selectedIconPath: "/static/tabbar/cart-active.svg",
      },
      {
        pagePath: "pages/profile/index",
        text: "我的",
        iconPath: "/static/tabbar/profile.svg",
        selectedIconPath: "/static/tabbar/profile-active.svg",
      },
    ],
    leftList: [],
    rightList: [],
    enableFab: true,
  },
  methods: {
    initLists() {
      const list = this.data.list || [];
      this.setData({
        leftList: list.slice(0, 2),
        rightList: list.slice(2),
      });
    },
    updateSelected() {
      const pages = getCurrentPages();
      const current = pages && pages.length ? pages[pages.length - 1] : null;
      const route = current && current.route ? String(current.route).replace(/^\//, "") : "";
      const matched = (this.data.list || []).find((i) => i.pagePath === route);
      this.setData({ selectedPath: matched ? matched.pagePath : "pages/index/index" });
    },
    onTap(e) {
      const path = e.currentTarget.dataset.path;
      const url = `/${String(path || "").replace(/^\//, "")}`;
      wx.switchTab({ url });
    },
    onFabTap() {
      // 设计稿的中间按钮：先占位，后续可接“发布/创建/扫码”等动作
      wx.showToast({ title: "功能待接入", icon: "none" });
    },
  },
  lifetimes: {
    attached() {
      this.initLists();
      this.updateSelected();
    },
  },
  pageLifetimes: {
    show() {
      this.updateSelected();
    },
  },
});
