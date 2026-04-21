<script setup lang="ts">
import { ref, watchEffect, nextTick } from 'vue';
import { useCameraStore } from '@/stores/camera';

const cameraStore = useCameraStore();
const videoEl = ref<HTMLVideoElement | null>(null);

watchEffect(async () => {
  if (cameraStore.opened && cameraStore.stream) {
    await nextTick();
    if (videoEl.value) {
      videoEl.value.srcObject = cameraStore.stream;
    }
  }
});
</script>

<template>
  <Teleport to="body">
    <div
      v-show="cameraStore.opened"
      v-if="cameraStore.opened"
      class="fixed inset-0 z-50 bg-black"
    >
      <!-- Live viewfinder -->
      <video
        v-show="!cameraStore.previewUrl"
        ref="videoEl"
        id="cameraVideo"
        autoplay
        playsinline
        class="absolute inset-0 w-full h-full object-contain"
      ></video>

      <!-- Preview after capture -->
      <img
        v-show="cameraStore.previewUrl"
        :src="cameraStore.previewUrl ?? undefined"
        class="absolute inset-0 w-full h-full object-contain"
      />

      <canvas id="cameraCanvas" class="hidden"></canvas>

      <!-- Shooting controls -->
      <div
        v-show="!cameraStore.previewUrl"
        class="absolute bottom-0 left-0 w-full flex flex-row items-center justify-center gap-4"
        style="padding-bottom: max(2rem, env(safe-area-inset-bottom))"
      >
        <button
          type="button"
          @click="cameraStore.capture()"
          class="h-16 w-16 bg-white rounded-full shadow-lg border-4 border-gray-300 hover:bg-gray-100 active:scale-95"
          title="Capture photo"
        ></button>
        <button
          type="button"
          @click="cameraStore.close()"
          class="h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg"
          style="transition: transform 0.3s ease;"
        >
          Cancel
        </button>
      </div>

      <!-- Preview controls -->
      <div
        v-show="cameraStore.previewUrl"
        class="absolute bottom-0 left-0 w-full flex flex-row items-center justify-center gap-4"
        style="padding-bottom: max(2rem, env(safe-area-inset-bottom))"
      >
        <button
          type="button"
          @click="cameraStore.confirm()"
          class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg"
        >
          Use
        </button>
        <button
          type="button"
          @click="cameraStore.rotate()"
          class="h-10 px-4 py-2 text-white bg-yellow-500 rounded-full shadow hover:bg-yellow-600 active:shadow-lg"
        >
          Rotate
        </button>
        <button
          type="button"
          @click="cameraStore.retake()"
          class="h-10 px-4 py-2 text-white bg-gray-500 rounded-full shadow hover:bg-gray-600 active:shadow-lg"
        >
          Retake
        </button>
        <button
          type="button"
          @click="cameraStore.close()"
          class="h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg"
        >
          Cancel
        </button>
      </div>
    </div>
  </Teleport>
</template>
