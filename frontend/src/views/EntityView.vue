<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useEntitiesStore } from '@/stores/entities';
import { useCameraStore } from '@/stores/camera';
import { useClipStore } from '@/stores/clip';
import { useToastsStore } from '@/stores/toasts';
import { api } from '@/api';
import type { Entity } from '@/api/types';

const route = useRoute();
const router = useRouter();
const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

const isLoading = computed(() => entitiesStore.isLoading);

// Watch for route changes to update current entity
import { watch } from 'vue';
import { onMounted } from 'vue';

onMounted(() => {
  const hashId = parseInt(window.location.hash.slice(1), 10);
  if (!isNaN(hashId)) {
    entitiesStore.setCurrentEntity(hashId);
  } else {
    entitiesStore.setCurrentEntity(0);
  }
  entitiesStore.connectWS();
});

watch(
  () => route.params.entityId,
  async (newId) => {
    if (newId !== undefined && newId !== null) {
      await entitiesStore.setCurrentEntity(parseInt(newId as string, 10));
    }
  },
  { immediate: true }
);

// Dialog state
const newEntityVisible = ref(false);
const moveEntityVisible = ref(false);
const newEntityTemp = ref({ name: '', description: '' });
const moveEntitySearch = ref('');
const moveEntityTarget = ref<number>(0);

// Filtered entities for move dialog
const filteredEntities = computed(() => {
  if (!moveEntitySearch.value.trim()) {
    return Object.values(entitiesStore.fullstate.entities).filter(
      (e) => e.id !== 0 && e.location !== 0
    );
  }
  const term = moveEntitySearch.value.toLowerCase();
  return Object.values(entitiesStore.fullstate.entities).filter(
    (e) =>
      e.id !== 0 &&
      e.location !== 0 &&
      (e.name?.toLowerCase().includes(term) ||
        e.description?.toLowerCase().includes(term) ||
        e.id.toString().includes(term))
  );
});

const handleNewEntitySubmit = async (): Promise<void> => {
  if (!newEntityTemp.value.name || !newEntityTemp.value.name.trim()) return;
  try {
    const entity = {
      name: newEntityTemp.value.name,
      description: newEntityTemp.value.description || null,
      artifacts: null,
      location: entitiesStore.currentEntity,
      metadata: {
        quantity: null,
        owners: null,
        tags: null,
        islabeled: false,
        lastModified: null,
        lastModifiedBy: null,
      },
    };
    const entityId = await api.createEntity(entity);
    await entitiesStore.reload();
    await entitiesStore.setCurrentEntity(entityId);
    newEntityVisible.value = false;
  } catch (error) {
    console.error('Failed to create entity:', error);
  }
};

function openNewEntityDialog(_entityId?: number): void {
  newEntityVisible.value = true;
}

const handleMoveEntitySubmit = async (): Promise<void> => {
  if (!moveEntityTarget.value || moveEntityTarget.value === 0) return;
  try {
    await api.moveEntity(Number(route.params.entityId), moveEntityTarget.value);
    await entitiesStore.reload();
    moveEntityVisible.value = false;
  } catch (error) {
    console.error('Failed to move entity:', error);
  }
};
</script>

<template>
  <!-- Main content wrapper -->
  <div
    v-if="!isLoading"
    class="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white"
  >
    <!-- Header with breadcrumbs -->
    <div class="container mx-auto pt-4 px-4">
      <nav class="w-full">
        <ol class="flex flex-wrap list-reset">
          <template v-for="(n, index) in entitiesStore.locationtree" :key="n">
            <li>
              <a
                @click="entitiesStore.setCurrentEntity(n)"
                class="text-blue-600 no-underline cursor-pointer dark:text-sky-400 dark:hover:text-sky-300 hover:text-blue-700 hover:underline"
                :title="`Go to entity ${n}`"
              >
                {{ entitiesStore.readname(n) }}
              </a>
            </li>

            <li v-if="index < entitiesStore.locationtree.length - 1">
              <span class="mx-2 text-gray-500">/</span>
            </li>
          </template>
          <li>
            <a
              @click="openNewEntityDialog(entitiesStore.currentEntity)"
              class="text-blue-600 dark:text-sky-400 dark:hover:text-sky-300 hover:text-blue-700 cursor-pointer"
              title="Create new entity"
            >
              +
            </a>
          </li>
        </ol>
      </nav>

      <!-- Search bar -->
      <SearchBar />
    </div>

    <!-- Empty state or entity list -->
    <div class="container mx-auto px-4 mt-4">
      <!-- Empty state -->
      <div v-if="!entitiesStore.hasChildren(entitiesStore.currentEntity)">
        <p class="text-2xl text-gray-500/50">Empty</p>
      </div>

      <!-- Quick capture card -->
      <div
        v-if="!entitiesStore.hasChildren(entitiesStore.currentEntity)"
        class="container relative h-full max-w-sm min-h-40 grow flex items-center justify-center rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600 bg-transparent cursor-pointer hover:border-blue-400 dark:hover:border-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/10 transition-colors"
        @click="
          cameraStore.open(async (files: File[]) => {
            if (files.length === 0 || !files[0]) return;
            try {
              const artifactId = await api.uploadArtifact(files[0]);
              const entity = {
                name: null,
                description: null,
                artifacts: [artifactId],
                location: entitiesStore.currentEntity,
                metadata: {
                  quantity: null,
                  owners: null,
                  tags: null,
                  islabeled: false,
                  lastModified: null,
                  lastModifiedBy: null,
                },
              };
              const newEntityId = await api.createEntity(entity);
              await entitiesStore.reload();
              await entitiesStore.setCurrentEntity(newEntityId);
              toastsStore.add('Entity created from photo');
            } catch (error) {
              console.error('Failed to create entity:', error);
              toastsStore.add('Failed to create entity from photo');
            }
          })
        "
      >
        <div class="flex flex-col items-center gap-2 text-gray-400 dark:text-gray-500 hover:text-blue-400 dark:hover:text-blue-500 pointer-events-none select-none">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="currentColor"
            class="w-10 h-10"
          >
            <path
              d="M12 9a3.75 3.75 0 100 7.5A3.75 3.75 0 0 0 12 9Z"
            />
            <path
              fill-rule="evenodd"
              d="M9.344 3.071a49.52 49.52 0 0 1 5.312 0c.967.052 1.83.585 2.332 1.39l.821 1.317c.24.383.645.643 1.11.71.386.054.77.113 1.152.177 1.432.239 2.429 1.493 2.429 2.909V18a3 3 0 0 1-3 3h-15a3 3 0 0 1-3-3V9.574c0-1.416.997-2.67 2.429-2.909.382-.064.766-.123 1.151-.177a1.56 1.56 0 0 0 1.11-.71l.822-1.318a2.75 2.75 0 0 1 2.332-1.39ZM6.75 12.75a5.25 5.25 0 1 1 10.5 0 5.25 5.25 0 0 1-10.5 0Zm12-1.5a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Z"
              clip-rule="evenodd"
            />
          </svg>
          <span class="text-sm">Tap to capture</span>
        </div>
      </div>

      <!-- Entity grid -->
      <div class="flex flex-wrap justify-center gap-4">
        <EntityCard
          v-for="entity in clipStore.merge(entitiesStore.load(entitiesStore.currentEntity, entitiesStore.searchtext), entitiesStore)"
          :key="entity.id"
          :entity="entity"
        />
      </div>
    </div>

    <!-- Dialogs -->
    <div
      v-if="newEntityVisible"
      class="fixed inset-0 overflow-y-auto z-50"
      role="dialog"
      aria-modal="true"
    >
      <!-- Overlay -->
      <div
        class="fixed inset-0 bg-black bg-opacity-50 dark:bg-gray-900 dark:bg-opacity-50"
        @click="newEntityVisible = false"
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
            <h2 class="pb-4 text-2xl font-medium text-gray-500 dark:text-white/50">
              ({{ entitiesStore.currentEntity }})
            </h2>
          </div>

          <!-- Simple form for now -->
          <div class="mb-4">
            <input
              v-model="newEntityTemp.name"
              type="text"
              placeholder="Name"
              class="w-full px-4 py-2 rounded bg-white dark:bg-gray-900 ring-1"
              autofocus
            />
          </div>

          <div class="mb-4">
            <textarea
              v-model="newEntityTemp.description"
              placeholder="Description"
              class="w-full px-4 py-2 rounded bg-white dark:bg-gray-900 ring-1"
              rows="3"
            ></textarea>
          </div>

          <!-- Buttons -->
          <div class="flex mt-8 space-x-2">
            <button
              @click="handleNewEntitySubmit"
              class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600"
            >
              Submit
            </button>
            <button
              @click="newEntityVisible = false"
              class="h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Move Entity Dialog -->
    <div
      v-if="moveEntityVisible"
      class="fixed inset-0 overflow-y-auto z-50"
      role="dialog"
      aria-modal="true"
    >
      <!-- Overlay -->
      <div
        class="fixed inset-0 bg-black bg-opacity-50 dark:bg-gray-900 dark:bg-opacity-50"
        @click="moveEntityVisible = false"
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

          <div class="mb-4">
            <input
              v-model="moveEntitySearch"
              type="search"
              placeholder="Search for a location..."
              class="w-full px-4 py-2 rounded-full bg-white ring-1 dark:bg-gray-900"
            />
          </div>

          <div class="mb-4">
            <label for="location" class="block mb-2">Select location</label>
            <select
              v-model="moveEntityTarget"
              id="location"
              class="w-full px-4 py-2 rounded-full bg-white ring-1 dark:bg-gray-900"
            >
              <option value="" disabled>Select a location</option>
              <option value="0">(0) World</option>
              <template
                v-for="loc in filteredEntities"
                :key="loc.id"
              >
                <option
                  v-if="!entitiesStore.hasChildren(loc.id)"
                  :value="loc.id"
                >
                  ({{ loc.id }}) {{ loc.name || loc.id }}
                </option>
              </template>
            </select>
          </div>

          <!-- Buttons -->
          <div class="flex mt-8 space-x-2">
            <button
              @click="handleMoveEntitySubmit"
              class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600"
            >
              Move Here
            </button>
            <button
              @click="moveEntityVisible = false"
              class="h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
</style>
