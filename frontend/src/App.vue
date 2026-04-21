<script setup lang="ts">
import { onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useEntitiesStore } from '@/stores/entities';
import { useCameraStore } from '@/stores/camera';
import { useClipStore } from '@/stores/clip';
import { useToastsStore } from '@/stores/toasts';
import EntityCard from '@/components/EntityCard.vue';
import CameraModal from '@/components/CameraModal.vue';
import NewEntityDialog from '@/components/NewEntityDialog.vue';
import MoveEntityDialog from '@/components/MoveEntityDialog.vue';
import SearchBar from '@/components/SearchBar.vue';
import BreadcrumbNav from '@/components/BreadcrumbNav.vue';
import QuickCaptureCard from '@/components/QuickCaptureCard.vue';
import ToastContainer from '@/components/ToastContainer.vue';

const router = useRouter();
const route = useRoute();
const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

// Initialize router from hash on app mount
onMounted(() => {
  const hashId = parseInt(window.location.hash.slice(1), 10);
  if (!isNaN(hashId)) {
    entitiesStore.setCurrentEntity(hashId);
  } else {
    entitiesStore.setCurrentEntity(0);
  }
  entitiesStore.connectWS();
});

// Watch hash changes
watch(
  () => route.params.entityId,
  async (newId) => {
    if (newId !== undefined) {
      await entitiesStore.setCurrentEntity(parseInt(newId as string, 10));
    }
  }
);
</script>

<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white">
    <!-- Loading state -->
    <div v-if="entitiesStore.isLoading" class="flex items-center justify-center h-screen">
      <span class="text-2xl text-gray-500">Loading...</span>
    </div>

    <!-- Main content -->
    <div v-else>
      <!-- Header with breadcrumbs -->
      <div class="container mx-auto pt-4 px-4">
        <nav class="w-full">
          <BreadcrumbNav />
          <button
            @click="entitiesStore.setCurrentEntity(entitiesStore.currentEntity)"
            class="text-blue-600 dark:text-sky-400 hover:text-blue-700 dark:hover:text-sky-300 ml-2"
            title="Create new entity"
          >
            +
          </button>
        </nav>

        <!-- Search bar -->
        <SearchBar />
      </div>

      <!-- Empty state or entity list -->
      <div class="container mx-auto px-4 mt-4">
        <div v-if="!entitiesStore.hasChildren(entitiesStore.currentEntity)">
          <p class="text-2xl text-gray-500/50">Empty</p>
        </div>

        <!-- Quick capture card -->
        <QuickCaptureCard />

        <!-- Entity grid -->
        <div class="flex flex-wrap justify-center gap-4">
          <EntityCard
            v-for="entity in clipStore.merge(entitiesStore.load(entitiesStore.currentEntity, entitiesStore.searchtext), entitiesStore)"
            :key="entity.id"
            :entity="entity"
          />
        </div>
      </div>
    </div>

    <!-- Camera modal -->
    <CameraModal />

    <!-- Dialogs -->
    <NewEntityDialog />
    <MoveEntityDialog />

    <!-- Toast notifications -->
    <ToastContainer />
  </div>
</template>

<style scoped>
/* Add custom styles here if needed */
</style>
