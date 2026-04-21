<script setup lang="ts" name="SearchBar">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useEntitiesStore } from '@/stores/entities';
import { useClipStore } from '@/stores/clip';
import { useToastsStore } from '@/stores/toasts';

const entitiesStore = useEntitiesStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

const debounceTimer = ref<ReturnType<typeof setTimeout> | null>(null);

const handleSearchInput = (): void => {
  if (debounceTimer.value) {
    clearTimeout(debounceTimer.value);
  }

  debounceTimer.value = setTimeout(() => {
    entitiesStore.debouncesearch();
    if (clipStore.enabled) {
      clipStore.search(
        entitiesStore.searchtext,
        entitiesStore.fullstate,
        entitiesStore.currentEntity
      );
    }
  }, 500);
};

const toggleClip = (): void => {
  clipStore.enabled = !clipStore.enabled;
  if (!clipStore.enabled) {
    clipStore.results = [];
    clipStore.scores = {};
    clipStore.searching = false;
  }
};

const toggleFilterWorld = (): void => {
  entitiesStore.filterworld = !entitiesStore.filterworld;
};

const toggleSearchDescription = (): void => {
  entitiesStore.searchdescription = !entitiesStore.searchdescription;
};

const resetSearch = (): void => {
  entitiesStore.searchtext = '';
  entitiesStore.searchtextpredebounce = '';
  clipStore.results = [];
  clipStore.scores = {};
  clipStore.textMatchIds = new Set();
  clipStore.searching = false;
};

onBeforeUnmount(() => {
  if (debounceTimer.value) {
    clearTimeout(debounceTimer.value);
  }
});
</script>

<template>
  <div class="flex flex-row items-center gap-2 mb-4">
    <!-- Search icon -->
    <div class="text-gray-500 dark:text-gray-400">
      <svg
        viewBox="0 0 24 24"
        fill="currentColor"
        class="w-6 h-6"
        xmlns="http://www.w3.org/2000/svg"
      >
        <path
          fill-rule="evenodd"
          d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25zM6.262 6.072a8.25 8.25 0 1 0 10.562-.766 4.5 4.5 0 0 1-1.318 1.357l-.302.604a.809.809 0 0 1-1.086 1.085l-.302-.165a1.125 1.125 0 0 0-1.298.21l-.131.131a1.125 1.125 0 0 0 0 1.591l.296.296c.256.257.622.374.98.314l1.17-.195c.323-.054.654.036.905.245l1.33 1.108c.32.267.46.694.358 1.1a8.7 8.7 0 0 1-2.288 4.04l-.723.724a1.125 1.125 0 0 1-1.298.21l-.153-.076a1.125 1.125 0 0 1-.622-1.006v-1.089c0-.298-.119-.585-.33-.796l-1.347-1.347a1.125 1.125 0 0 1-.21-1.298l.494-.988a1.125 1.125 0 0 1 1.085-1.085l.33-.165z"
          clip-rule="evenodd"
        />
      </svg>
    </div>

    <!-- Filter world checkbox -->
    <div class="flex items-center">
      <label class="flex items-center cursor-pointer" title="Only search in current entity">
        <input
          type="checkbox"
          v-model="entitiesStore.filterworld"
          class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          @change="toggleFilterWorld"
        />
        <span class="ml-1 text-sm text-gray-600 dark:text-gray-400">World</span>
      </label>
    </div>

    <!-- Search description checkbox -->
    <div class="flex items-center">
      <label class="flex items-center cursor-pointer" title="Include description in search">
        <input
          type="checkbox"
          v-model="entitiesStore.searchdescription"
          class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          @change="toggleSearchDescription"
        />
        <span class="ml-1 text-sm text-gray-600 dark:text-gray-400">Desc</span>
      </label>
    </div>

    <!-- CLIP enable checkbox -->
    <div class="flex items-center" title="Visual search using CLIP">
      <label class="flex items-center cursor-pointer">
        <input
          type="checkbox"
          v-model="clipStore.enabled"
          class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          @change="toggleClip"
        />
        <span class="ml-1 text-sm text-gray-600 dark:text-gray-400">Visual</span>
      </label>
    </div>

    <!-- Search input -->
    <input
      v-model="entitiesStore.searchtextpredebounce"
      @input="handleSearchInput"
      @keydown.esc="resetSearch"
      placeholder="Search for an entity..."
      type="search"
      class="w-full px-4 py-2 rounded-full bg-white ring-1 ring-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:ring-gray-600 dark:text-white"
    />

    <!-- CLIP loading indicator -->
    <span
      v-if="
        clipStore.enabled &&
        (clipStore.searching ||
          clipStore.modelLoading ||
          clipStore.encoded < clipStore.total)
      "
      class="flex items-center gap-1 shrink-0 text-sm text-gray-500 dark:text-gray-400"
    >
      <svg
        v-if="clipStore.searching || clipStore.modelLoading || clipStore.encoded < clipStore.total"
        class="animate-spin w-4 h-4 text-blue-400"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12h0c0 6.627 5.373 12 12 12h0a7.962 7.962 0 01-2.91-5.291h0a7.962 7.962 0 01-5.291-2.91h0z"
        />
      </svg>
      <span
        v-if="clipStore.modelLoading"
        class="text-xs text-gray-400 whitespace-nowrap"
      >
        Loading CLIP...
      </span>
      <span
        v-if="
          !clipStore.modelLoading &&
          clipStore.encoded < clipStore.total &&
          clipStore.searching
        "
        class="text-xs text-gray-400 whitespace-nowrap"
      >
        {{ clipStore.encoded }}/{{ clipStore.total }}
      </span>
    </span>

    <!-- Clear button -->
    <button
      v-if="
        entitiesStore.searchtext ||
        clipStore.enabled ||
        entitiesStore.filterworld ||
        entitiesStore.searchdescription
      "
      @click="resetSearch"
      type="button"
      class="h-8 w-8 flex items-center justify-center rounded-full bg-gray-100 hover:bg-gray-200 text-gray-600 dark:bg-gray-700 dark:hover:bg-gray-600 dark:text-gray-300"
      title="Clear search"
    >
      &times;
    </button>
  </div>
</template>

<style scoped>
input[type='search']::-webkit-search-decoration,
input[type='search']::-webkit-search-cancel-button,
input[type='search']::-webkit-search-results-decoration {
  display: none;
}
</style>
