import { createRouter, createWebHashHistory } from "vue-router";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/:entityId?",
      name: "entity",
      component: () => import("../views/EntityView.vue"),
    },
  ],
});

export default router;
