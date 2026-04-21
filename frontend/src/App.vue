<script setup lang="ts">
import { onMounted, watch, ref } from 'vue';
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
const newEntityVisible = ref(false);

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
    const id = parseInt(newId as string, 10);
    if (!isNaN(id)) {
      await entitiesStore.setCurrentEntity(id);
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
      <div class="w-full pt-4 px-4">
        <BreadcrumbNav @open-new-entity="newEntityVisible = true" />

        <!-- Search bar -->
        <SearchBar />
      </div>

      <!-- Empty state or entity list -->
      <div class="w-full px-4 mt-4">
        <div v-if="!entitiesStore.hasChildren(entitiesStore.currentEntity)">
          <p class="text-2xl text-gray-500/50">Empty</p>
        </div>

        <!-- Entity grid -->
        <div class="flex flex-wrap justify-center gap-4">
          <QuickCaptureCard />
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
    <NewEntityDialog :visible="newEntityVisible" :location="entitiesStore.currentEntity" @update:visible="newEntityVisible = $event" />
    <MoveEntityDialog />

    <!-- Toast notifications -->
    <ToastContainer />
  </div>
</template>

<style scoped>
/* Add custom styles here if needed */
</style>
