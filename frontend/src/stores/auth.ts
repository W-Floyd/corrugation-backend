import { defineStore } from "pinia";
import { ref, computed } from "vue";

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(localStorage.getItem("auth_token"));

  const isAuthenticated = computed(() => token.value !== null);

  function setToken(t: string) {
    token.value = t;
    localStorage.setItem("auth_token", t);
  }

  function clearToken() {
    token.value = null;
    localStorage.removeItem("auth_token");
  }

  return { token, isAuthenticated, setToken, clearToken };
});
