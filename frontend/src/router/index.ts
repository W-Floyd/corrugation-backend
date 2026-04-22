import { createRouter, createWebHashHistory } from "vue-router";

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("../views/LoginView.vue"),
    },
    {
      path: "/:entityId?",
      name: "entity",
      component: () => import("../views/EntityView.vue"),
    },
  ],
});

router.beforeEach((to) => {
  const token = localStorage.getItem("auth_token");
  if (to.name === "login" && token) {
    return { name: "entity" };
  }
});

export default router;
