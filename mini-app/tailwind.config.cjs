/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
  // 微信小程序 WXSS 对 `*`、`::before/after`、`:is()` 等兼容性较差，关闭 preflight。
  corePlugins: {
    preflight: false,
  },
  theme: {
    extend: {
      colors: {
        primary: "#4F8A7E",
        "primary-hover": "#3D6F64",
        "primary-10": "rgba(79,138,126,0.10)",
        "secondary-accent": "#E6F0EE",
        "background-light": "#F8FAF9",
        "background-dark": "#1A2624",
        "text-dark": "#1C2B28",
        // wxss 兼容：避免 `bg-white/95` 这类会生成反斜杠选择器的写法
        surface: {
          90: "rgba(255,255,255,0.90)",
          95: "rgba(255,255,255,0.95)"
        },
        line: {
          5: "rgba(0,0,0,0.05)"
        },
        muted: "rgba(28,43,40,0.70)"
      },
      fontFamily: {
        display: ["ui-sans-serif", "system-ui", "sans-serif"],
      },
      borderRadius: {
        DEFAULT: "0.25rem",
        lg: "0.5rem",
        xl: "0.75rem",
        full: "9999px",
      },
      boxShadow: {
        "primary-soft": "0 16px 40px rgba(79, 138, 126, 0.14)",
        card: "0 18px 50px rgba(79, 138, 126, 0.10)",
        cta: "0 14px 28px rgba(79, 138, 126, 0.22)",
      },
      height: {
        // wxss 兼容：避免 `h-[180px]` 这类会生成反斜杠选择器的写法
        180: "180px",
      },
      fontSize: {
        // wxss 兼容：避免 `text-[10px]` 这类会生成反斜杠选择器的写法
        10: ["10px", "12px"],
      },
    },
  },
  // 小程序端优先保证 WXSS 可编译；如后续确有需要再按端区分引入插件
  plugins: [],
};
