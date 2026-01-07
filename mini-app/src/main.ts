import { createSSRApp } from "vue";
import App from "./App.vue";
import { createAppI18n } from "./i18n";
export function createApp() {
  const app = createSSRApp(App);
  app.use(createAppI18n());
  return {
    app,
  };
}
