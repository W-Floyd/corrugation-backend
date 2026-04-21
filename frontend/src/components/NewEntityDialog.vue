<script setup lang="ts" name="NewEntityDialog">
import { ref, onMounted, watch } from 'vue';
import { useEntitiesStore } from '@/stores/entities';
import { useCameraStore } from '@/stores/camera';
import { useToastsStore } from '@/stores/toasts';
import { api } from '@/api';
import type { Entity, EntityCreate } from '@/api/types';

const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const toastsStore = useToastsStore();

const props = withDefaults(defineProps<{
  visible?: boolean;
  location?: number;
}>(), {
  visible: false,
  location: 0,
});

const emit = defineEmits<{
  created: [entityId: number];
  'update:visible': [value: boolean];
}>();

const dialogVisible = ref(false);
const entity = ref<EntityCreate>({
  name: null,
  description: null,
  artifacts: null,
  location: 0,
  metadata: {
    quantity: null,
    owners: null,
    tags: null,
    isLabeled: false,
        lastModified: null,
        lastModifiedBy: null,
  },
});
const files = ref<File[]>([]);
const freeId = ref<number>(0);
const availableId = ref<number>(0);

watch(
  () => props.visible,
  async (visible) => {
    dialogVisible.value = visible;
    if (visible) {
      await resetDialog();
      await fetchIds();
    }
  },
  { immediate: true }
);

watch(
  () => props.location,
  (location) => {
    entity.value.location = location ?? 0;
  }
);

watch(
  () => entity.value.metadata.isLabeled,
  () => fetchIds()
);

const resetDialog = async (): Promise<void> => {
  entity.value = {
    name: null,
    description: null,
    artifacts: null,
    location: props.location ?? 0,
    metadata: {
      quantity: null,
      owners: null,
      tags: null,
      isLabeled: false,
      lastModified: null,
      lastModifiedBy: null,
    },
  };
  files.value = [];
  freeId.value = 0;
  availableId.value = 0;
};

const fetchIds = async (): Promise<void> => {
  if (entity.value.metadata.isLabeled) {
    availableId.value = await api.firstAvailableId();
  } else {
    freeId.value = await api.firstFreeId();
  }
};

const handleSubmit = async (): Promise<void> => {
  if (!entity.value.name || !entity.value.name.trim()) {
    toastsStore.add('Name is required');
    return;
  }

  try {
    const location = props.location || 0;
    entity.value.location = location;
    const entityId = await api.createEntity(entity.value);
    await entitiesStore.reload();
    emit('created', entityId);
    emit('update:visible', false);
    dialogVisible.value = false;
    toastsStore.add('Entity created');
  } catch (error) {
    console.error('Failed to create entity:', error);
    toastsStore.add('Failed to create entity');
  }
};

const handleDialogClose = (): void => {
  dialogVisible.value = false;
  emit('update:visible', false);
};

const handleCameraOpen = async (): Promise<void> => {
  await new Promise<void>((resolve) => {
    cameraStore.open((fileList: File[]) => {
      files.value = fileList;
      resolve();
    });
  });
};

onMounted(() => {
  fetchIds();
});
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
          <div class="grid grid-cols-2 w-fit space-x-2 items-baseline mb-4">
            <h1 class="pb-4 text-3xl font-medium">Create New Entity</h1>
            <h2
              class="pb-4 text-2xl font-medium text-gray-500 dark:text-white/50"
            >
              ({{
                entity.metadata.isLabeled ? availableId : freeId
              }})
            </h2>
          </div>

          <!-- Form -->
          <div class="grid grid-cols-[8rem_1fr] gap-x-4 gap-y-3 items-center">
            <label for="name">Name</label>
            <input
              id="name"
              type="text"
              v-model="entity.name"
              class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1"
              autofocus
            />

            <label for="islabeled">Is Labeled</label>
            <input
              id="islabeled"
              type="checkbox"
              v-model="entity.metadata.isLabeled"
              class="w-4 h-4 justify-self-start"
            />

            <label for="description">Description</label>
            <textarea
              id="description"
              v-model="entity.description"
              class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1"
              rows="3"
            ></textarea>

            <label for="quantity">Quantity</label>
            <input
              id="quantity"
              type="number"
              min="0"
              v-model.number="entity.metadata.quantity"
              class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1"
            />

            <label for="file">Image</label>
            <div class="flex flex-wrap items-center gap-2">
              <input
                id="file"
                type="file"
                class="bg-white rounded-sm ring-1 dark:bg-gray-900 dark:hover:bg-gray-700"
                accept="image/*"
                multiple
                @change="
                  (e) => {
                    files = Array.from((e.target as HTMLInputElement).files || []);
                  }
                "
              />
              <button
                type="button"
                @click="handleCameraOpen"
                class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600"
              >
                Camera
              </button>
              <span v-if="files.length > 0" class="text-sm text-gray-500">
                {{ files.length }} file(s) selected
              </span>
            </div>
          </div>

          <!-- Buttons -->
          <div class="flex mt-8 space-x-2">
            <button
              type="button"
              @click="handleSubmit"
              class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600"
            >
              Submit
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
