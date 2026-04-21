<script setup lang="ts" name="MoveEntityDialog">
import { ref, watch } from 'vue';
import { useEntitiesStore } from '@/stores/entities';
import { useToastsStore } from '@/stores/toasts';
import { api } from '@/api';
import type { Entity } from '@/api/types';

const entitiesStore = useEntitiesStore();
const toastsStore = useToastsStore();

const props = withDefaults(defineProps<{
  visible?: boolean;
  targetEntityId?: number;
}>(), {
  visible: false,
  targetEntityId: 0,
});

const emit = defineEmits<{
  moved: [entityId: number, newLocation: number];
}>();

const dialogVisible = ref(false);
const entity = ref<Entity | null>(null);
const searchtext = ref('');
const targetLocation = ref<number>(0);

watch(
  () => props.visible,
  (visible) => {
    dialogVisible.value = visible;
    if (visible && props.targetEntityId) {
      entity.value = entitiesStore.fullstate.entities[props.targetEntityId] || null;
    }
  },
  { immediate: true }
);

watch(
  () => props.targetEntityId,
  (id) => {
    if (id && dialogVisible.value) {
      entity.value = entitiesStore.fullstate.entities[id] || null;
    }
  }
);

const filteredEntities = () => {
  if (!searchtext.value.trim()) {
    return Object.values(entitiesStore.fullstate.entities).filter(
      (e) => e.id !== entity.value?.id && e.location !== 0
    );
  }

  const term = searchtext.value.toLowerCase();
  return Object.values(entitiesStore.fullstate.entities).filter(
    (e) =>
      e.id !== entity.value?.id &&
      e.location !== 0 &&
      (e.name?.toLowerCase().includes(term) ||
        e.description?.toLowerCase().includes(term) ||
        e.id.toString().includes(term))
  );
};

const hasChildren = (entityId: number): boolean => {
  return entitiesStore.hasChildren(entityId);
};

const formatOption = (entityId: number): string => {
  const tree: string[] = [];
  let target = entityId;
  while (target !== 0) {
    const elem = entitiesStore.fullstate.entities[target];
    if (!elem) {
      tree.push(target.toString());
      break;
    }
    if (!elem.name) {
      tree.push(target.toString());
    } else {
      tree.push(elem.name);
    }
    target = elem.location;
  }
  tree.push('World');
  tree.reverse();
  return `(${entityId}) ${tree.join('/')}`;
};

const handleMove = async (): Promise<void> => {
  if (!entity.value) return;

  try {
    await api.moveEntity(entity.value.id, targetLocation.value);
    await entitiesStore.reload();
    emit('moved', entity.value.id, targetLocation.value);
    dialogVisible.value = false;
    toastsStore.add('Entity moved');
  } catch (error) {
    console.error('Failed to move entity:', error);
    toastsStore.add('Failed to move entity');
  }
};

const handleDialogClose = (): void => {
  dialogVisible.value = false;
};
</script>

<template>
  <Teleport to="body">
    <div
      v-if="dialogVisible"
      class="fixed inset-0 overflow-y-auto z-50"
      role="dialog"
      aria-modal="true"
    >
      <!-- Overlay -->
      <div
        class="fixed inset-0 bg-black bg-opacity-50 dark:bg-gray-900 dark:bg-opacity-50"
        @click="handleDialogClose"
      ></div>

      <!-- Panel -->
      <div
        class="relative flex items-center justify-center min-h-screen p-4"
        @click.stop
      >
        <div
          class="relative w-full max-w-2xl p-8 overflow-y-auto bg-white border border-gray-300 rounded-lg dark:bg-gray-800"
        >
          <!-- Title -->
          <h1 class="pb-4 text-3xl font-medium">Move Entity</h1>

          <div v-if="entity">
            <p class="text-gray-600 dark:text-gray-400 mb-4">
              Moving: {{ entity.name || entity.id }}
            </p>

            <!-- Search -->
            <div class="mb-4">
              <input
                v-model="searchtext"
                type="search"
                placeholder="Search for a location..."
                class="w-full px-4 py-2 rounded-full bg-white ring-1 dark:bg-gray-900"
              />
            </div>

            <!-- Location selector -->
            <div class="mb-4">
              <label for="location" class="block mb-2">
                Select location
              </label>
              <select
                v-model="targetLocation"
                id="location"
                class="w-full px-4 py-2 rounded-full bg-white ring-1 dark:bg-gray-900"
              >
                <option value="" disabled>Select a location</option>
                <template v-for="loc in filteredEntities()" :key="loc.id">
                  <option
                    v-if="!hasChildren(loc.id)"
                    :value="loc.id"
                  >
                    {{ formatOption(loc.id) }}
                  </option>
                </template>
              </select>
            </div>

            <!-- Info about current entity -->
            <div v-if="hasChildren(entity.id)" class="mb-4">
              <p class="text-gray-600 dark:text-gray-400 mb-2">
                This entity has children:
              </p>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="childId in entitiesStore.listChildLocations(entity.id)"
                  :key="childId"
                  class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded"
                >
                  {{ formatOption(childId) }}
                </span>
              </div>
              <p class="text-sm text-gray-500 mt-2">
                Note: Children will follow this entity to the new location.
              </p>
            </div>
          </div>

          <!-- Buttons -->
          <div class="flex mt-8 space-x-2">
            <button
              type="button"
              @click="handleMove"
              class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600"
            >
              Move Here
            </button>
            <button
              type="button"
              @click="handleDialogClose"
              class="h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
