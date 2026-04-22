<script setup lang="ts">
import { onMounted, onUnmounted, watch, ref, computed, nextTick } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useEntitiesStore } from "@/stores/entities";
import { useCameraStore } from "@/stores/camera";
import { useClipStore } from "@/stores/clip";
import { useToastsStore } from "@/stores/toasts";
import EntityCard from "@/components/EntityCard.vue";
import CameraModal from "@/components/CameraModal.vue";
import NewEntityDialog from "@/components/NewEntityDialog.vue";
import CommandDialog from "@/components/CommandDialog.vue";
import SearchBar from "@/components/SearchBar.vue";
import BreadcrumbNav from "@/components/BreadcrumbNav.vue";
import QuickCaptureCard from "@/components/QuickCaptureCard.vue";
import ToastContainer from "@/components/ToastContainer.vue";
import KbdHint from "@/components/KbdHint.vue";
import LoginView from "@/views/LoginView.vue";
import PlusIcon from "vue-material-design-icons/Plus.vue";
import CameraIcon from "vue-material-design-icons/Camera.vue";
import { api } from "@/api";

const router = useRouter();
const route = useRoute();
const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

const newEntityVisible = ref(false);
const newEntityLocation = ref(0);
const confirmMoveId = ref<number | null>(null);
const commandDialogVisible = ref(false);
const selectedEntityId = ref<number | null>(null);
const showShortcuts = ref(false);
const editEntityId = ref<number | null>(null);
const cardRefs = ref<Record<number, { cardEl: HTMLElement | null }>>({});
const deleteConfirmId = ref<number | null>(null);
const searchBarRef = ref<{ focusSearch: () => void } | null>(null);
const editingCardId = ref<number | null>(null);

const visibleEntities = computed(() =>
    clipStore.merge(
        entitiesStore.load(
            entitiesStore.currentEntity,
            entitiesStore.searchtext,
        ),
        entitiesStore,
    ),
);

const anyDialogOpen = computed(
    () =>
        newEntityVisible.value ||
        confirmMoveId.value !== null ||
        commandDialogVisible.value,
);

const handleMoveConfirmed = async (
    entityId: number,
    newLocation: number,
): Promise<void> => {
    const idx = visibleEntities.value.findIndex((e) => e.id === entityId);
    const rest = visibleEntities.value.filter((e) => e.id !== entityId);
    const nextId =
        rest.length > 0 ? rest[Math.min(idx, rest.length - 1)].id : null;
    confirmMoveId.value = null;
    selectedEntityId.value = null;
    try {
        await api.moveEntity(entityId, newLocation);
        await entitiesStore.reload();
        toastsStore.add("Entity moved");
        if (newLocation === entitiesStore.currentEntity) {
            selectedEntityId.value = entityId;
        } else if (nextId !== null) {
            selectedEntityId.value = nextId;
        }
    } catch {
        toastsStore.add("Failed to move entity");
    }
};

const handleFabCapture = async (): Promise<void> => {
    const capturedFiles: File[] = [];
    await new Promise<void>((resolve) => {
        cameraStore.open((files: File[]) => {
            capturedFiles.push(...files);
            resolve();
        });
    });
    if (!capturedFiles[0]) return;
    try {
        const artifactId = await api.uploadArtifact(capturedFiles[0]);
        await api.createEntity({
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
        });
        await entitiesStore.reload();
        toastsStore.add("Entity created from photo");
    } catch {
        toastsStore.add("Failed to create entity from photo");
    }
};

const confirmDeleteEntity = async (entityId: number): Promise<void> => {
    const beforeList = visibleEntities.value.filter((e) => e.id !== entityId);
    const idx = visibleEntities.value.findIndex((e) => e.id === entityId);
    const nextId =
        beforeList.length > 0
            ? beforeList[Math.min(idx, beforeList.length - 1)].id
            : null;
    deleteConfirmId.value = null;
    selectedEntityId.value = null;
    try {
        await api.deleteEntity(entityId);
        await entitiesStore.reload();
        toastsStore.add("Entity deleted");
        if (nextId !== null) {
            selectedEntityId.value = nextId;
        }
    } catch {
        toastsStore.add("Failed to delete entity");
    }
};

const handleQuickCaptureOnEntity = async (entityId: number): Promise<void> => {
    const capturedFiles: File[] = [];
    await new Promise<void>((resolve) => {
        cameraStore.open((files: File[]) => {
            capturedFiles.push(...files);
            resolve();
        });
    });
    if (!capturedFiles[0]) return;
    try {
        const artifactId = await api.uploadArtifact(capturedFiles[0]);
        const entity = entitiesStore.fullstate.entities[entityId];
        const artifacts = [...(entity?.artifacts ?? []), artifactId];
        await api.patchEntity(entityId, { artifacts });
        await entitiesStore.reload();
        toastsStore.add("Artifact captured and added");
    } catch {
        toastsStore.add("Failed to capture artifact");
    }
};

const handleQuickCaptureNewChild = async (parentId: number): Promise<void> => {
    const capturedFiles: File[] = [];
    await new Promise<void>((resolve) => {
        cameraStore.open((files: File[]) => {
            capturedFiles.push(...files);
            resolve();
        });
    });
    if (!capturedFiles[0]) return;
    try {
        const artifactId = await api.uploadArtifact(capturedFiles[0]);
        await api.createEntity({
            name: null,
            description: null,
            artifacts: [artifactId],
            location: parentId,
            metadata: {
                quantity: null,
                owners: null,
                tags: null,
                islabeled: false,
                lastModified: null,
                lastModifiedBy: null,
            },
        });
        await entitiesStore.reload();
        toastsStore.add("Entity created from photo");
    } catch {
        toastsStore.add("Failed to create entity from photo");
    }
};

const navigateGrid = (direction: "up" | "down" | "left" | "right"): void => {
    const entities = visibleEntities.value;
    if (entities.length === 0) return;

    if (selectedEntityId.value === null) {
        selectedEntityId.value = entities[0].id;
        return;
    }

    const currentEl = cardRefs.value[selectedEntityId.value]?.cardEl;
    if (!currentEl) return;

    const cur = currentEl.getBoundingClientRect();
    const curCX = cur.left + cur.width / 2;
    const curCY = cur.top + cur.height / 2;

    let bestId: number | null = null;
    let bestScore = Infinity;

    for (const entity of entities) {
        if (entity.id === selectedEntityId.value) continue;
        const el = cardRefs.value[entity.id]?.cardEl;
        if (!el) continue;
        const r = el.getBoundingClientRect();
        const cx = r.left + r.width / 2;
        const cy = r.top + r.height / 2;
        const dx = cx - curCX;
        const dy = cy - curCY;

        const inDir =
            direction === "right"
                ? dx > 10
                : direction === "left"
                  ? dx < -10
                  : direction === "down"
                    ? dy > 10
                    : dy < -10;
        if (!inDir) continue;

        const primary =
            direction === "left" || direction === "right"
                ? Math.abs(dx)
                : Math.abs(dy);
        const secondary =
            direction === "left" || direction === "right"
                ? Math.abs(dy)
                : Math.abs(dx);
        const score = primary + secondary * 3;
        if (score < bestScore) {
            bestScore = score;
            bestId = entity.id;
        }
    }

    if (bestId !== null) selectedEntityId.value = bestId;
};

const handleKeydown = (e: KeyboardEvent): void => {
    if (e.key === "Meta" || e.key === "Alt") {
        showShortcuts.value = true;
        return;
    }

    const tag = (e.target as HTMLElement)?.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;

    // Allow Escape through even with dialogs open
    if (e.key === "Escape") {
        commandDialogVisible.value = false;
        deleteConfirmId.value = null;
        confirmMoveId.value = null;
        selectedEntityId.value = null;
        return;
    }

    if (anyDialogOpen.value) return;

    switch (e.key) {
        case "/":
            e.preventDefault();
            searchBarRef.value?.focusSearch();
            break;

        case "?":
            e.preventDefault();
            commandDialogVisible.value = true;
            break;

        case "v":
        case "V":
            e.preventDefault();
            clipStore.enabled = !clipStore.enabled;
            if (!clipStore.enabled) {
                clipStore.results = [];
                clipStore.scores = {};
                clipStore.searching = false;
            }
            break;

        case "t":
        case "T":
            e.preventDefault();
            entitiesStore.searchdescription = !entitiesStore.searchdescription;
            break;

        case "w":
        case "W":
            e.preventDefault();
            entitiesStore.filterworld = !entitiesStore.filterworld;
            break;

        case "ArrowDown":
            e.preventDefault();
            navigateGrid("down");
            break;
        case "ArrowUp":
            e.preventDefault();
            navigateGrid("up");
            break;
        case "ArrowRight":
            e.preventDefault();
            navigateGrid("right");
            break;
        case "ArrowLeft":
            e.preventDefault();
            navigateGrid("left");
            break;

        case "Enter":
            if (cameraStore.opened || editingCardId.value !== null) break;
            e.preventDefault();
            if (deleteConfirmId.value !== null) {
                confirmDeleteEntity(deleteConfirmId.value);
            } else if (selectedEntityId.value !== null) {
                entitiesStore
                    .setCurrentEntity(selectedEntityId.value)
                    .then(() => {
                        nextTick(() => {
                            if (visibleEntities.value.length > 0) {
                                selectedEntityId.value =
                                    visibleEntities.value[0].id;
                            }
                        });
                    });
            }
            break;

        case "Backspace":
            e.preventDefault();
            {
                const cur = entitiesStore.currentEntity;
                if (cur === 0) break;
                const prevId = cur;
                const ent = entitiesStore.fullstate.entities[cur];
                entitiesStore.setCurrentEntity(ent?.location ?? 0).then(() => {
                    nextTick(() => {
                        selectedEntityId.value = prevId;
                    });
                });
            }
            break;

        case "Delete":
        case "d":
        case "D":
            if (!e.shiftKey && !e.metaKey && !e.ctrlKey) {
                if (deleteConfirmId.value !== null) {
                    e.preventDefault();
                    confirmDeleteEntity(deleteConfirmId.value);
                } else if (selectedEntityId.value !== null) {
                    e.preventDefault();
                    deleteConfirmId.value = selectedEntityId.value;
                }
            }
            break;

        case "e":
        case "E":
            if (
                !e.shiftKey &&
                !e.metaKey &&
                !e.ctrlKey &&
                selectedEntityId.value !== null
            ) {
                e.preventDefault();
                editEntityId.value = selectedEntityId.value;
            }
            break;

        case "p":
        case "P":
            if (
                !e.shiftKey &&
                !e.metaKey &&
                !e.ctrlKey &&
                selectedEntityId.value !== null
            ) {
                e.preventDefault();
                handleQuickCaptureOnEntity(selectedEntityId.value);
            }
            break;

        case "c":
        case "C":
            if (
                e.shiftKey &&
                !e.metaKey &&
                !e.ctrlKey &&
                selectedEntityId.value !== null
            ) {
                e.preventDefault();
                handleQuickCaptureNewChild(selectedEntityId.value);
            } else if (!e.shiftKey && !e.metaKey && !e.ctrlKey) {
                e.preventDefault();
                handleFabCapture();
            }
            break;

        case "n":
        case "N":
            if (
                e.shiftKey &&
                !e.metaKey &&
                !e.ctrlKey &&
                selectedEntityId.value !== null
            ) {
                e.preventDefault();
                newEntityLocation.value = selectedEntityId.value;
                newEntityVisible.value = true;
            } else if (!e.shiftKey && !e.metaKey && !e.ctrlKey) {
                e.preventDefault();
                newEntityLocation.value = entitiesStore.currentEntity;
                newEntityVisible.value = true;
            }
            break;

        case "m":
        case "M":
            if (
                !e.shiftKey &&
                !e.metaKey &&
                !e.ctrlKey &&
                selectedEntityId.value !== null
            ) {
                e.preventDefault();
                confirmMoveId.value = selectedEntityId.value;
            }
            break;
    }
};

const handleKeyup = (e: KeyboardEvent): void => {
    if (e.key === "Meta" || e.key === "Alt") {
        showShortcuts.value = false;
    }
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
    window.addEventListener("keydown", handleKeydown);
    window.addEventListener("keyup", handleKeyup);
});

onUnmounted(() => {
    window.removeEventListener("keydown", handleKeydown);
    window.removeEventListener("keyup", handleKeyup);
});

watch(selectedEntityId, (newId) => {
    if (deleteConfirmId.value !== null && newId !== deleteConfirmId.value) {
        deleteConfirmId.value = null;
    }
});

// Clear selection when navigating to a new entity
watch(
    () => entitiesStore.currentEntity,
    () => {
        selectedEntityId.value = null;
        deleteConfirmId.value = null;
    },
);

// Watch hash changes
watch(
    () => route.params.entityId,
    async (newId) => {
        const id = parseInt(newId as string, 10);
        if (!isNaN(id)) {
            await entitiesStore.setCurrentEntity(id);
        }
    },
);
</script>

<template>
    <LoginView v-if="route.name === 'login'" />

    <div
        v-else
        class="min-h-screen bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-white"
    >
        <!-- Loading state -->
        <div
            v-if="entitiesStore.isLoading"
            class="flex items-center justify-center h-screen"
        >
            <span class="text-2xl text-gray-500">Loading...</span>
        </div>

        <!-- Main content -->
        <div v-else>
            <!-- Header with breadcrumbs -->
            <div class="w-full pt-4 px-4 pb-4">
                <BreadcrumbNav
                    @open-new-entity="
                        newEntityLocation = entitiesStore.currentEntity;
                        newEntityVisible = true;
                    "
                />
                <SearchBar ref="searchBarRef" :show-shortcuts="showShortcuts" />
            </div>

            <!-- Empty state or entity list -->
            <div class="w-full px-4 mt-8">
                <div
                    v-if="visibleEntities.length === 0"
                    class="flex items-center justify-center h-64"
                >
                    <p class="text-2xl text-gray-500/50">Empty</p>
                </div>

                <!-- Entity grid -->
                <div class="flex flex-wrap justify-center gap-4">
                    <EntityCard
                        v-for="entity in visibleEntities"
                        :key="entity.id"
                        :ref="
                            (el: any) => {
                                if (el) cardRefs[entity.id] = el;
                                else delete cardRefs[entity.id];
                            }
                        "
                        :entity="entity"
                        :is-selected="selectedEntityId === entity.id"
                        :show-shortcuts="showShortcuts"
                        :start-edit="editEntityId === entity.id"
                        :confirm-delete="deleteConfirmId === entity.id"
                        :confirm-move="confirmMoveId === entity.id"
                        @select="
                            selectedEntityId = entity.id;
                            deleteConfirmId = null;
                        "
                        @create-child="
                            (id) => {
                                newEntityLocation = id;
                                newEntityVisible = true;
                            }
                        "
                        @request-move="
                            (id) => {
                                confirmMoveId = id;
                            }
                        "
                        @edit-started="
                            editEntityId = null;
                            editingCardId = entity.id;
                        "
                        @edit-ended="editingCardId = null"
                        @request-delete="
                            selectedEntityId = entity.id;
                            deleteConfirmId = entity.id;
                        "
                        @delete-confirmed="confirmDeleteEntity(entity.id)"
                        @delete-cancelled="deleteConfirmId = null"
                        @move-confirmed="
                            (newLocation) =>
                                handleMoveConfirmed(entity.id, newLocation)
                        "
                        @move-cancelled="confirmMoveId = null"
                    />
                </div>
            </div>
        </div>

        <!-- Floating action buttons -->
        <div class="fixed bottom-6 right-6 flex flex-col gap-3">
            <button
                @click="
                    newEntityLocation = entitiesStore.currentEntity;
                    newEntityVisible = true;
                "
                class="relative h-14 w-14 flex items-center justify-center rounded-full bg-blue-500 hover:bg-blue-600 text-white shadow-lg active:shadow-xl"
                title="Create new entity (N)"
            >
                <PlusIcon :size="28" />
                <KbdHint shortcut="N" :show="showShortcuts" />
            </button>
            <button
                @click="handleFabCapture"
                class="relative h-14 w-14 flex items-center justify-center rounded-full bg-blue-500 hover:bg-blue-600 text-white shadow-lg active:shadow-xl"
                title="Quick capture (C)"
            >
                <CameraIcon :size="28" />
                <KbdHint shortcut="C" :show="showShortcuts" />
            </button>
        </div>

        <!-- Camera modal -->
        <CameraModal />

        <!-- Dialogs -->
        <NewEntityDialog
            :visible="newEntityVisible"
            :location="newEntityLocation"
            :show-shortcuts="showShortcuts"
            @update:visible="newEntityVisible = $event"
            @created="
                (id) => {
                    if (newEntityLocation === entitiesStore.currentEntity)
                        selectedEntityId = id;
                }
            "
        />
        <CommandDialog
            :visible="commandDialogVisible"
            @update:visible="commandDialogVisible = $event"
        />

        <!-- Toast notifications -->
        <ToastContainer />
    </div>
</template>

<style scoped>
/* Add custom styles here if needed */
</style>
