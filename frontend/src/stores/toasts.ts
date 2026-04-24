import { defineStore } from "pinia";
import { ref } from "vue";

interface Toast {
  id: number;
  message: string;
}

export const useToastsStore = defineStore("toasts", () => {
  const items = ref<Toast[]>([]);

  function add(message: string): void {
    const id = Date.now();
    items.value = [...items.value, { id, message }];
    setTimeout(() => remove(id), 5000);
  }

  function remove(id: number): void {
    items.value = items.value.filter((t) => t.id !== id);
  }

  return {
    items,
    add,
    remove,
  };
});
