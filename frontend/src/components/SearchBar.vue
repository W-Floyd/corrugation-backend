<script setup lang="ts" name="SearchBar">
import { ref, onMounted, onBeforeUnmount } from "vue";
import KbdHint from "@/components/KbdHint.vue";
import MagnifyIcon from "vue-material-design-icons/Magnify.vue";
import LoadingIcon from "vue-material-design-icons/Loading.vue";
import CloseIcon from "vue-material-design-icons/Close.vue";
import { useEntitiesStore } from "@/stores/entities";
import { useClipStore } from "@/stores/clip";
import { useToastsStore } from "@/stores/toasts";

const props = defineProps<{ showShortcuts?: boolean }>();

const entitiesStore = useEntitiesStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

const debounceTimer = ref<ReturnType<typeof setTimeout> | null>(null);
const searchInputEl = ref<HTMLInputElement | null>(null);

const focusSearch = (): void => {
    searchInputEl.value?.focus();
    searchInputEl.value?.select();
};

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
                entitiesStore.currentEntity,
            );
        }
    }, 500);
};

const onClipChange = (): void => {
    if (!clipStore.enabled) {
        clipStore.results = [];
        clipStore.scores = {};
        clipStore.searching = false;
    }
};

const resetSearch = (): void => {
    entitiesStore.searchtext = "";
    entitiesStore.searchtextpredebounce = "";
    clipStore.results = [];
    clipStore.scores = {};
    clipStore.textMatchIds = new Set();
    clipStore.searching = false;
    searchInputEl.value?.blur();
};

defineExpose({ focusSearch });

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
            <MagnifyIcon :size="24" />
        </div>

        <!-- Filter world checkbox -->
        <div class="flex items-center">
            <label
                class="flex items-center cursor-pointer"
                title="Only search in current entity"
            >
                <input
                    type="checkbox"
                    v-model="entitiesStore.filterworld"
                    class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span class="ml-1 text-sm text-gray-600 dark:text-gray-400"
                    >World</span
                >
            </label>
        </div>

        <!-- Search description checkbox -->
        <div class="flex items-center">
            <label
                class="flex items-center cursor-pointer"
                title="Include description in search"
            >
                <input
                    type="checkbox"
                    v-model="entitiesStore.searchdescription"
                    class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span class="ml-1 text-sm text-gray-600 dark:text-gray-400"
                    >Desc</span
                >
            </label>
        </div>

        <!-- CLIP enable checkbox -->
        <div class="flex items-center" title="Visual search using CLIP">
            <label class="flex items-center cursor-pointer">
                <input
                    type="checkbox"
                    v-model="clipStore.enabled"
                    class="w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    @change="onClipChange"
                />
                <span class="ml-1 text-sm text-gray-600 dark:text-gray-400"
                    >Visual</span
                >
            </label>
        </div>

        <!-- Search input -->
        <div class="relative flex-1">
            <input
                ref="searchInputEl"
                v-model="entitiesStore.searchtextpredebounce"
                @input="handleSearchInput"
                @keydown.esc="resetSearch"
                placeholder="Search for an entity..."
                type="search"
                class="w-full px-4 py-2 rounded-full bg-white ring-1 ring-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:ring-gray-600 dark:text-white"
            />
            <kbd
                v-if="props.showShortcuts"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-sans bg-gray-800 text-white rounded px-1 leading-[14px] pointer-events-none shadow"
                >/</kbd
            >
        </div>

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
            <LoadingIcon
                v-if="
                    clipStore.searching ||
                    clipStore.modelLoading ||
                    clipStore.encoded < clipStore.total
                "
                :size="16"
                class="animate-spin text-blue-400"
            />
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

        <!-- Command palette shortcut hint -->
        <KbdHint shortcut="?" :show="props.showShortcuts" :inline="true" />

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
            <CloseIcon :size="16" />
        </button>
    </div>
</template>

<style scoped>
input[type="search"]::-webkit-search-decoration,
input[type="search"]::-webkit-search-cancel-button,
input[type="search"]::-webkit-search-results-decoration {
    display: none;
}
</style>
