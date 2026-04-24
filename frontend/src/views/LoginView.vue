<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900">
    <div class="w-full max-w-sm bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-8 flex flex-col gap-6">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Sign in</h1>

      <button
        v-if="authStore.authConfig.enabled"
        @click="authStore.startLogin()"
        class="rounded-lg bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 text-sm transition-colors"
      >
        Sign in with Authentik
      </button>

      <p v-else class="text-sm text-gray-500 dark:text-gray-400">
        Authentication is not configured.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "../stores/auth";

const authStore = useAuthStore();
</script>
