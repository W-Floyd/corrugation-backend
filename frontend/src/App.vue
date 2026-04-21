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
import PlusIcon from 'vue-material-design-icons/Plus.vue';
import CameraIcon from 'vue-material-design-icons/Camera.vue';
import { api } from '@/api';

const router = useRouter();
const route = useRoute();
const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();
const newEntityVisible = ref(false);
const newEntityLocation = ref(0);

const handleFabCapture = (): void => {
  cameraStore.open(async (files: File[]) => {
    if (!files[0]) return;
    try {
      const artifactId = await api.uploadArtifact(files[0]);
      await api.createEntity({
        name: null, description: null, artifacts: [artifactId],
        location: entitiesStore.currentEntity,
        metadata: { quantity: null, owners: null, tags: null, islabeled: false, lastModified: null, lastModifiedBy: null },
      });
      await entitiesStore.reload();
      toastsStore.add('Entity created from photo');
    } catch {
      toastsStore.add('Failed to create entity from photo');
    }
  });
};

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
      <div class="w-full pt-4 px-4 pb-4">
        <BreadcrumbNav @open-new-entity="newEntityLocation = entitiesStore.currentEntity; newEntityVisible = true" />

        <!-- Search bar -->
        <SearchBar />
      </div>

      <!-- Empty state or entity list -->
      <div class="w-full px-4 mt-8">
        <div v-if="!entitiesStore.hasChildren(entitiesStore.currentEntity)" class="flex items-center justify-center h-64">
          <p class="text-2xl text-gray-500/50">Empty</p>
        </div>

        <!-- Entity grid -->
        <div class="flex flex-wrap justify-center gap-4">
          <EntityCard
            v-for="entity in clipStore.merge(entitiesStore.load(entitiesStore.currentEntity, entitiesStore.searchtext), entitiesStore)"
            :key="entity.id"
            :entity="entity"
            @create-child="(id) => { newEntityLocation = id; newEntityVisible = true; }"
          />
        </div>
      </div>
    </div>

    <!-- Floating action buttons -->
    <div class="fixed bottom-6 right-6 flex flex-col gap-3">
      <button
        @click="newEntityLocation = entitiesStore.currentEntity; newEntityVisible = true"
        class="h-14 w-14 flex items-center justify-center rounded-full bg-blue-500 hover:bg-blue-600 text-white shadow-lg active:shadow-xl"
        title="Create new entity"
      >
        <PlusIcon :size="28" />
      </button>
      <button
        @click="handleFabCapture"
        class="h-14 w-14 flex items-center justify-center rounded-full bg-blue-500 hover:bg-blue-600 text-white shadow-lg active:shadow-xl"
        title="Quick capture"
      >
        <CameraIcon :size="28" />
      </button>
    </div>

    <!-- Camera modal -->
    <CameraModal />

    <!-- Dialogs -->
    <NewEntityDialog :visible="newEntityVisible" :location="newEntityLocation" @update:visible="newEntityVisible = $event" />
    <MoveEntityDialog />

    <!-- Toast notifications -->
    <ToastContainer />
  </div>
</template>

<style scoped>
/* Add custom styles here if needed */
</style>
