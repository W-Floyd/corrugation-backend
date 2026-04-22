<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
    <div class="w-full max-w-sm bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-6">Sign in</h1>

      <form @submit.prevent="handleSubmit" class="flex flex-col gap-4">
        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Username</label>
          <input
            v-model="username"
            type="text"
            autocomplete="username"
            required
            class="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div class="flex flex-col gap-1">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
            required
            class="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <p v-if="error" class="text-sm text-red-500">{{ error }}</p>

        <button
          type="submit"
          :disabled="loading"
          class="mt-2 rounded-lg bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-medium py-2 px-4 text-sm transition-colors"
        >
          {{ loading ? "Signing in…" : "Sign in" }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { api } from "../api";
import { useAuthStore } from "../stores/auth";
import { useEntitiesStore } from "../stores/entities";

const router = useRouter();
const authStore = useAuthStore();
const entitiesStore = useEntitiesStore();

const username = ref("");
const password = ref("");
const error = ref("");
const loading = ref(false);

async function handleSubmit() {
  error.value = "";
  loading.value = true;
  try {
    const token = await api.login(username.value, password.value);
    authStore.setToken(token);
    await entitiesStore.reload();
    router.push({ name: "entity" });
  } catch {
    error.value = "Invalid username or password.";
  } finally {
    loading.value = false;
  }
}
</script>
