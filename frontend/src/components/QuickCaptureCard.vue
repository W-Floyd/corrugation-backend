<script setup lang="ts" name="QuickCaptureCard">
import { useCameraStore } from '@/stores/camera';
import { useEntitiesStore } from '@/stores/entities';
import { useToastsStore } from '@/stores/toasts';
import { api } from '@/api';
import type { Entity } from '@/api/types';

const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const toastsStore = useToastsStore();

const handleQuickCapture = async (entityId: number): Promise<void> => {
  try {
    await new Promise<void>((resolve) => {
      cameraStore.open(async (files: File[]) => {
        if (files.length === 0 || !files[0]) {
          resolve();
          return;
        }

        try {
          // Upload artifact
          const artifactId = await api.uploadArtifact(files[0]);

          // Create entity at current location
          const entity: Entity = {
            id: 0,
            name: null,
            description: null,
            artifacts: [artifactId],
            location: entityId,
            metadata: {
              quantity: null,
              owners: null,
              tags: null,
              isLabeled: false,
              lastModified: null,
              lastModifiedBy: null,
            },
          };

          await api.createEntity(entity);
          await entitiesStore.reload();
          toastsStore.add('Entity created from photo');
        } catch (error) {
          console.error('Failed to create entity:', error);
          toastsStore.add('Failed to create entity from photo');
        }

        resolve();
      });
    });
  } catch (error) {
    console.error('Camera error:', error);
    toastsStore.add('Camera error');
  }
};
</script>

<template>
  <figure
    v-if="!entitiesStore.hasChildren(entitiesStore.currentEntity)"
    class="container relative h-full max-w-sm min-h-40 grow flex items-center justify-center rounded-xl border-2 border-dashed border-gray-300 dark:border-gray-600 bg-transparent cursor-pointer hover:border-blue-400 dark:hover:border-blue-500 hover:bg-blue-50/50 dark:hover:bg-blue-900/10 transition-colors"
    @click="handleQuickCapture(entitiesStore.currentEntity)"
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
  </figure>
</template>
