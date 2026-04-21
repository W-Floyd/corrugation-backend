<script setup lang="ts" name="ToastContainer">
import { ref, onMounted, onUnmounted } from 'vue';
import { useToastsStore } from '@/stores/toasts';

const toastsStore = useToastsStore();

const visibleToasts = ref(new Map<number, boolean>());

onMounted(() => {
  toastsStore.items.forEach((toast) => {
    visibleToasts.value.set(toast.id, true);
  });
});

const hideToast = (toastId: number): void => {
  visibleToasts.value.set(toastId, false);
  setTimeout(() => {
    toastsStore.remove(toastId);
    visibleToasts.value.delete(toastId);
  }, 300);
};

const showToast = (toastId: number): void => {
  visibleToasts.value.set(toastId, true);
};
</script>

<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 items-end pointer-events-none">
    <TransitionGroup
      name="toast"
      tag="div"
      class="flex flex-col gap-2 items-end"
    >
      <div
        v-for="toast in toastsStore.items"
        :key="toast.id"
        v-show="visibleToasts.get(toast.id)"
        @mouseenter="visibleToasts.set(toast.id, true)"
        @mouseleave="hideToast(toast.id)"
        class="flex items-start gap-2 max-w-sm px-4 py-3 bg-red-50 dark:bg-red-900/40 text-red-700 dark:text-red-300 rounded-lg shadow-lg ring-1 ring-red-200 dark:ring-red-700 pointer-events-auto"
      >
        <span class="text-sm break-words">{{ toast.message }}</span>
        <button
          type="button"
          @click="hideToast(toast.id)"
          class="shrink-0 text-red-400 hover:text-red-600 dark:hover:text-red-200"
          title="Dismiss"
        >
          &times;
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.3s ease-out, transform 0.3s ease-out;
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(0.5rem);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(0.5rem);
}
</style>
