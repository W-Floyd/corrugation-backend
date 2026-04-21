import "./assets/main.css";

import { createApp } from "vue";
import { createPinia } from "pinia";

import App from "./App.vue";
import router from "./router";

const pinia = createPinia();
const app = createApp(App);
app.use(pinia);
app.use(router);

// Connect CLIP store with entities store for visual search
app.config.globalProperties.$store = {
  entities: undefined,
  clip: undefined,
};

app.mount("#app");
