<script setup lang="ts" name="EntityCard">
import { ref, computed, watch, nextTick, onUnmounted } from "vue";
import { useEntitiesStore } from "@/stores/entities";
import { useCameraStore } from "@/stores/camera";
import { useClipStore } from "@/stores/clip";
import { useToastsStore } from "@/stores/toasts";
import { api } from "@/api";
import type { Entity } from "@/api/types";
import KbdHint from "@/components/KbdHint.vue";
import ArtifactImage from "@/components/ArtifactImage.vue";
import TrashCanIcon from "vue-material-design-icons/TrashCan.vue";
import FolderMoveIcon from "vue-material-design-icons/FolderMove.vue";
import PencilIcon from "vue-material-design-icons/Pencil.vue";
import CameraIcon from "vue-material-design-icons/Camera.vue";
import CameraPlusIcon from "vue-material-design-icons/CameraPlus.vue";
import PlusIcon from "vue-material-design-icons/Plus.vue";
import CheckIcon from "vue-material-design-icons/Check.vue";
import CloseIcon from "vue-material-design-icons/Close.vue";
import ArrowUpIcon from "vue-material-design-icons/ArrowUp.vue";

const props = defineProps<{
    entity: Entity;
    isSelected?: boolean;
    showShortcuts?: boolean;
    startEdit?: boolean;
    confirmDelete?: boolean;
    confirmMove?: boolean;
}>();

const emit = defineEmits<{
    entityUpdated: [entity: Entity];
    createChild: [locationId: number];
    requestMove: [entityId: number];
    select: [];
    editStarted: [];
    editEnded: [];
    deleteConfirmed: [];
    deleteCancelled: [];
    requestDelete: [];
    moveConfirmed: [newLocation: number];
    moveCancelled: [];
}>();

const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

const cardEl = ref<HTMLElement | null>(null);
const nameInputEl = ref<HTMLInputElement | null>(null);
const editMode = ref(false);
const localEntity = ref<Entity>({
    ...props.entity,
    metadata: { ...props.entity.metadata },
});
const pendingDeletions = ref<Set<number>>(new Set());

const moveTargetLocation = ref<number>(0);
const moveSearchInputRef = ref<HTMLInputElement | null>(null);

const isDescendantOf = (entityId: number, ancestorId: number): boolean => {
    let current = entityId;
    while (current !== 0) {
        if (current === ancestorId) return true;
        const parent = entitiesStore.fullstate.entities[current];
        if (!parent) break;
        current = parent.location;
    }
    return false;
};

const moveUp = (): void => {
    if (entitiesStore.currentEntity !== 0) {
        const currentEntity =
            entitiesStore.fullstate.entities[entitiesStore.currentEntity];
        if (currentEntity?.location !== undefined) {
            emit("moveConfirmed", currentEntity.location);
        }
    }
};

const filteredMoveEntities = computed(() => {
    const term = entitiesStore.moveSearchtext.toLowerCase().trim();
    const world: Entity = {
        id: 0,
        name: "World",
        description: "",
        artifacts: [],
        location: 0,
        metadata: {
            quantity: null,
            owners: null,
            tags: null,
            lastModified: null,
            lastModifiedBy: null,
            islabeled: false,
        },
    };
    const candidates: Entity[] = [
        ...Object.values(entitiesStore.fullstate.entities).filter(
            (e) =>
                e.id !== props.entity.id &&
                !isDescendantOf(e.id, props.entity.id),
        ),
        world,
    ];
    if (!term) return candidates;
    return candidates.filter(
        (e) =>
            e.name?.toLowerCase().includes(term) ||
            e.description?.toLowerCase().includes(term) ||
            e.id.toString().includes(term),
    );
});

const currentLocationName = computed(() => {
    if (entitiesStore.currentEntity === 0) return "World";
    const e = entitiesStore.fullstate.entities[entitiesStore.currentEntity];
    return e?.name || entitiesStore.currentEntity.toString();
});

const isAtCurrentLocation = computed((): boolean => {
    if (props.entity.location === undefined) return false;
    return props.entity.location === entitiesStore.currentEntity;
});

const handleMoveKeydown = (e: KeyboardEvent): void => {
    if (!props.confirmMove) return;
    const target = e.target as HTMLElement;
    if (e.key === "Escape") {
        e.preventDefault();
        e.stopImmediatePropagation();
        if (target.matches("input, select")) {
            (target as HTMLElement).blur();
        } else {
            emit("moveCancelled");
        }
    } else if (e.key === "Enter") {
        e.preventDefault();
        e.stopImmediatePropagation();
        emit("moveConfirmed", moveTargetLocation.value);
    } else if ((e.key === "h" || e.key === "H") && !target.matches("input")) {
        e.preventDefault();
        e.stopImmediatePropagation();
        emit("moveConfirmed", entitiesStore.currentEntity);
    } else if (e.key === "u" || e.key === "U") {
        e.preventDefault();
        e.stopImmediatePropagation();
        moveUp();
    }
};

watch(
    () => props.confirmMove,
    (val) => {
        if (val) {
            const results = filteredMoveEntities.value;
            const hasSearch = entitiesStore.moveSearchtext.trim() !== "";
            moveTargetLocation.value =
                hasSearch && results.length > 0
                    ? (results[0]?.id ?? 0)
                    : entitiesStore.currentEntity;
            window.addEventListener("keydown", handleMoveKeydown, true);
            nextTick(() => moveSearchInputRef.value?.focus());
        } else {
            window.removeEventListener("keydown", handleMoveKeydown, true);
        }
    },
);

watch(filteredMoveEntities, (results) => {
    if (!props.confirmMove) return;
    if (!results.some((r) => r.id === moveTargetLocation.value)) {
        moveTargetLocation.value = results[0]?.id ?? 0;
    }
});

const formatOption = (entityId: number): string => {
    const tree: string[] = [];
    let target = entityId;
    while (target !== 0) {
        const elem = entitiesStore.fullstate.entities[target];
        if (!elem) {
            tree.push(target.toString());
            break;
        }
        if (!elem.name) {
            tree.push(target.toString());
        } else {
            tree.push(elem.name);
        }
        target = elem.location;
    }
    tree.push("World");
    tree.reverse();
    return tree.join("/");
};

watch(
    () => props.isSelected,
    (val) => {
        if (val)
            nextTick(() =>
                (cardEl.value as HTMLElement)?.scrollIntoView({
                    behavior: "smooth",
                    block: "nearest",
                }),
            );
    },
);

watch(
    () => props.startEdit,
    (val) => {
        if (val && !editMode.value) {
            handleEditToggle();
            emit("editStarted");
        }
    },
);

const handleEditKeydown = (e: KeyboardEvent): void => {
    if (cameraStore.opened) return;
    const target = e.target as HTMLElement;
    if (e.key === "Escape") {
        e.preventDefault();
        e.stopImmediatePropagation();
        if (target.matches("input, textarea")) {
            (target as HTMLElement).blur();
        } else {
            handleCancel();
        }
    } else if (e.key === "Enter" && !target.matches("textarea")) {
        e.preventDefault();
        e.stopImmediatePropagation();
        handleSave();
    } else if (
        (e.key === "p" || e.key === "P") &&
        !target.matches("input, textarea")
    ) {
        e.preventDefault();
        e.stopImmediatePropagation();
        cameraStore.open((files: File[]) =>
            files.forEach((f) => handleEditArtifact(f)),
        );
    }
};

watch(editMode, (val) => {
    if (val) {
        window.addEventListener("keydown", handleEditKeydown, true);
        nextTick(() => nameInputEl.value?.focus());
    } else {
        window.removeEventListener("keydown", handleEditKeydown, true);
        emit("editEnded");
    }
});

onUnmounted(() => {
    window.removeEventListener("keydown", handleEditKeydown, true);
    window.removeEventListener("keydown", handleMoveKeydown, true);
});

const handleUpdate = async (): Promise<void> => {
    try {
        await Promise.all(
            [...pendingDeletions.value].map((id) => api.deleteArtifact(id)),
        );
        const artifacts = (localEntity.value.artifacts ?? []).filter(
            (id) => !pendingDeletions.value.has(id),
        );
        await api.patchEntity(props.entity.id, {
            ...localEntity.value,
            artifacts,
        });
        pendingDeletions.value = new Set();
        await entitiesStore.reload();
        editMode.value = false;
        emit("entityUpdated", localEntity.value);
        toastsStore.add("Entity updated");
    } catch (error) {
        console.error("Failed to update entity:", error);
        toastsStore.add("Failed to update entity");
    }
};

const handleDelete = async (): Promise<void> => {
    try {
        await api.deleteEntity(props.entity.id);
        await entitiesStore.reload();
        toastsStore.add("Entity deleted");
    } catch (error) {
        console.error("Failed to delete entity:", error);
        toastsStore.add("Failed to delete entity");
    }
};

const handleQuickCapture = async (): Promise<void> => {
    await new Promise<void>((resolve) => {
        cameraStore.open((files: File[]) => {
            handleQuickCaptureCallback(files);
            resolve();
        });
    });
};

const handleQuickCaptureCallback = async (files: File[]): Promise<void> => {
    if (files.length === 0 || !files[0]) return;
    try {
        const artifactId = await api.uploadArtifact(files[0]);
        const artifacts = [...(props.entity.artifacts ?? []), artifactId];
        await api.patchEntity(props.entity.id, { artifacts });
        await entitiesStore.reload();
        editMode.value = false;
        emit("entityUpdated", props.entity);
        toastsStore.add("Artifact captured and added");
    } catch (error) {
        console.error("Failed to capture artifact:", error);
        toastsStore.add("Failed to capture artifact");
    }
};

const handleQuickCaptureNewChild = async (): Promise<void> => {
    await new Promise<void>((resolve) => {
        cameraStore.open(async (files: File[]) => {
            if (!files[0]) {
                resolve();
                return;
            }
            try {
                const artifactId = await api.uploadArtifact(files[0]);
                await api.createEntity({
                    name: null,
                    description: null,
                    artifacts: [artifactId],
                    location: props.entity.id,
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
            resolve();
        });
    });
};

const handleEditToggle = (): void => {
    if (!editMode.value) {
        localEntity.value = {
            ...props.entity,
            metadata: { ...props.entity.metadata },
        };
    }
    editMode.value = !editMode.value;
};

const handleSave = async (): Promise<void> => {
    await handleUpdate();
};

const handleCancel = (): void => {
    localEntity.value = {
        ...props.entity,
        metadata: { ...props.entity.metadata },
    };
    pendingDeletions.value = new Set();
    editMode.value = false;
};

const toggleArtifactDeletion = (artifactId: number): void => {
    const next = new Set(pendingDeletions.value);
    if (next.has(artifactId)) {
        next.delete(artifactId);
    } else {
        next.add(artifactId);
    }
    pendingDeletions.value = next;
};

const isImageArtifact = (artifactId: number): boolean => {
    return Boolean(entitiesStore.fullstate.artifacts[artifactId]?.image);
};

const images = computed(() => {
    const artifacts = editMode.value
        ? localEntity.value.artifacts
        : props.entity.artifacts;
    if (!artifacts) return [];
    return artifacts.filter(isImageArtifact);
});

const handleEditArtifact = async (file: File): Promise<void> => {
    try {
        const artifactId = await api.uploadArtifact(file);
        const updatedEntity = { ...localEntity.value };
        if (!updatedEntity.artifacts) updatedEntity.artifacts = [];
        updatedEntity.artifacts.push(artifactId);
        await api.patchEntity(props.entity.id, updatedEntity);
        await entitiesStore.reload();
        emit("entityUpdated", updatedEntity);
        toastsStore.add("Artifact uploaded");
    } catch (error) {
        console.error("Failed to upload artifact:", error);
        toastsStore.add("Failed to upload artifact");
    }
};

defineExpose({ cardEl });
</script>

<template>
    <figure
        ref="cardEl"
        class="relative h-full min-h-64 min-w-48 max-w-sm bg-white shadow-md dark:bg-gray-800 rounded-xl flex flex-col cursor-default"
        :class="
            isSelected
                ? 'ring-2 ring-blue-500 shadow-blue-200 dark:shadow-blue-900'
                : 'ring-1 ring-gray-500/25 hover:ring-gray-500/50 hover:shadow-lg'
        "
        @click="emit('select')"
    >
        <!-- Delete confirmation overlay -->
        <div
            v-if="confirmDelete"
            class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 rounded-xl bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm"
            @click.stop
        >
            <p class="text-lg font-semibold text-red-600 dark:text-red-400">
                Delete "{{ entity.name || entity.id }}"?
            </p>
            <div class="flex gap-3">
                <button
                    @click.stop="emit('deleteConfirmed')"
                    class="h-9 px-4 rounded-full bg-red-500 hover:bg-red-600 text-white text-sm shadow relative"
                >
                    Delete
                    <KbdHint
                        shortcut="Enter"
                        :show="showShortcuts && isSelected"
                    />
                </button>
                <button
                    @click.stop="emit('deleteCancelled')"
                    class="h-9 px-4 rounded-full bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-sm shadow relative"
                >
                    Cancel
                    <KbdHint
                        shortcut="Esc"
                        :show="showShortcuts && isSelected"
                    />
                </button>
            </div>
        </div>
        <!-- Move confirmation overlay -->
        <div
            v-if="confirmMove"
            class="absolute inset-0 z-10 flex flex-col gap-2 rounded-xl bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm p-4"
            @click.stop
        >
            <p class="text-sm font-semibold text-gray-700 dark:text-gray-300">
                Move "{{ entity.name || entity.id }}" to:
            </p>
            <input
                ref="moveSearchInputRef"
                v-model="entitiesStore.moveSearchtext"
                type="search"
                placeholder="Search locations..."
                class="w-full px-3 py-1.5 rounded-full bg-white ring-1 dark:bg-gray-900 text-sm"
                @click.stop
            />
            <select
                v-model="moveTargetLocation"
                class="w-full px-2 py-1 rounded-lg bg-white ring-1 dark:bg-gray-900 text-sm flex-1 min-h-0"
                size="4"
                @click.stop
            >
                <option
                    v-for="loc in filteredMoveEntities"
                    :key="loc.id"
                    :value="loc.id"
                >
                    {{ formatOption(loc.id) }}
                </option>
            </select>
            <div class="flex gap-2 flex-wrap items-center">
                <button
                    @click.stop="emit('moveConfirmed', moveTargetLocation)"
                    class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                    title="Move"
                >
                    <CheckIcon :size="20" />
                    <KbdHint
                        shortcut="Enter"
                        :show="showShortcuts && isSelected"
                    />
                </button>
                <button
                    v-if="!isAtCurrentLocation"
                    @click.stop="
                        emit('moveConfirmed', entitiesStore.currentEntity)
                    "
                    class="h-10 px-3 rounded-full bg-purple-500 hover:bg-purple-600 text-white text-sm shadow relative"
                >
                    To {{ currentLocationName }}
                    <KbdHint shortcut="H" :show="showShortcuts && isSelected" />
                </button>
                <button
                    v-if="entity.id !== 0 && entitiesStore.currentEntity !== 0"
                    @click.stop="moveUp()"
                    class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-orange-500 rounded-full shadow hover:bg-orange-600 active:shadow-lg text-white"
                    title="Move to parent"
                >
                    <ArrowUpIcon :size="20" />
                    <KbdHint shortcut="U" :show="showShortcuts && isSelected" />
                </button>
                <button
                    @click.stop="emit('moveCancelled')"
                    class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg text-white"
                    title="Cancel"
                >
                    <CloseIcon :size="20" />
                    <KbdHint
                        shortcut="Esc"
                        :show="showShortcuts && isSelected"
                    />
                </button>
            </div>
        </div>
        <!-- Match badges -->
        <div class="absolute top-2 right-2 flex gap-1 items-center">
            <div
                v-if="
                    entitiesStore.searchtext &&
                    clipStore.textMatchIds.has(entity.id) &&
                    clipStore.enabled
                "
                class="text-xs text-gray-400 font-medium px-1 rounded bg-gray-100 dark:bg-gray-700"
                title="Text match"
            >
                T
            </div>
            <div
                v-if="
                    clipStore.enabled &&
                    clipStore.scores[entity.id] !== undefined &&
                    clipStore.scores[entity.id] !== null
                "
                class="text-xs text-gray-400 px-1 rounded bg-gray-100 dark:bg-gray-700 cursor-default"
                :title="`Visual: ${(clipStore.scores[entity.id]! * 100).toFixed(1)}%`"
            >
                {{ Math.round(clipStore.scores[entity.id]! * 100) }}%
            </div>
        </div>

        <!-- Content -->
        <div class="p-4 flex-auto flex flex-col">
            <!-- Title -->
            <div v-if="!editMode">
                <div
                    class="flex list-reset space-x-3 items-baseline mb-2 cursor-pointer"
                    @click.stop="entitiesStore.setCurrentEntity(entity.id)"
                >
                    <div
                        class="text-xl w-min font-medium text-gray-500 dark:text-gray-400"
                    >
                        ({{ entity.id }})
                    </div>
                    <div class="text-xl font-bold">
                        {{
                            entitiesStore.searchtext.trim()
                                ? formatOption(entity.id)
                                : entity.metadata.quantity !== null &&
                                    entity.metadata.quantity !== 0
                                  ? `${entity.name || ""} (x${entity.metadata.quantity})`
                                  : entity.name || ""
                        }}
                    </div>
                </div>
            </div>

            <!-- Edit mode title -->
            <div v-else>
                <div
                    class="flex-auto flex list-reset space-x-2 items-baseline mb-2"
                >
                    <div
                        class="text-xl w-min font-medium text-gray-500 dark:text-gray-400"
                    >
                        ({{ entity.id }})
                    </div>
                    <input
                        ref="nameInputEl"
                        type="text"
                        v-model="localEntity.name"
                        class="bg-white rounded-sm dark:bg-gray-900 ring-1"
                        placeholder="Name"
                    />
                    <input
                        type="number"
                        min="0"
                        v-model.number="localEntity.metadata.quantity"
                        class="bg-white rounded-sm dark:bg-gray-900 ring-1 w-10"
                        placeholder="Qty"
                    />
                    <label
                        class="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400 cursor-pointer"
                    >
                        <input
                            type="checkbox"
                            v-model="localEntity.metadata.islabeled"
                            class="w-4 h-4"
                        />
                        Labeled
                    </label>
                </div>
            </div>

            <!-- Description -->
            <div v-if="!editMode && entity.description">
                <p class="text-gray-600 dark:text-gray-400">
                    {{ entity.description }}
                </p>
            </div>
            <div v-else-if="editMode">
                <textarea
                    v-model="localEntity.description"
                    class="bg-white rounded-sm dark:bg-gray-900 ring-1 w-full"
                    rows="3"
                    placeholder="Description"
                ></textarea>
            </div>

            <!-- Children -->
            <div v-if="!editMode && entitiesStore.hasChildren(entity.id)">
                <p class="mb-2 font-semibold">Contains:</p>
                <div
                    class="flex flex-wrap gap-2 overflow-hidden hover:overflow-y-auto max-h-32 shadow-md p-2 ring-1 ring-gray-500/10 hover:ring-gray-500/25 hover:shadow-lg rounded-md"
                    style="scrollbar-gutter: stable"
                >
                    <div
                        v-for="childId in entitiesStore.listChildLocations(
                            entity.id,
                        )"
                        :key="childId"
                        class="p-1 rounded cursor-pointer bg-gray-50 dark:bg-gray-800 dark:hover:bg-gray-700 hover:bg-gray-100 hover:shadow-sm ring-gray-200 dark:ring-slate-500 ring-1 hover:ring-blue-500/75 active:shadow-md"
                        @click.stop="entitiesStore.setCurrentEntity(childId)"
                    >
                        {{ entitiesStore.readname(childId) }}
                    </div>
                </div>
            </div>
        </div>

        <!-- Action buttons -->
        <div class="p-4 flex flex-wrap gap-2">
            <button
                v-if="!editMode"
                @click.stop="emit('requestDelete')"
                class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg text-white"
                title="Delete entity"
            >
                <TrashCanIcon :size="20" />
                <KbdHint shortcut="Del" :show="showShortcuts && isSelected" />
            </button>

            <button
                v-if="!editMode"
                @click.stop="emit('requestMove', entity.id)"
                class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                title="Move entity"
            >
                <FolderMoveIcon :size="20" />
                <KbdHint shortcut="M" :show="showShortcuts && isSelected" />
            </button>

            <button
                v-if="!editMode"
                @click.stop="handleEditToggle"
                class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                title="Edit entity"
            >
                <PencilIcon :size="20" />
                <KbdHint shortcut="Enter" :show="showShortcuts && isSelected" />
            </button>

            <button
                v-if="!editMode"
                @click.stop="handleQuickCapture"
                class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                title="Quick capture (add photo to this entity)"
            >
                <CameraIcon :size="20" />
                <KbdHint shortcut="P" :show="showShortcuts && isSelected" />
            </button>

            <button
                v-if="!editMode"
                @click.stop="handleQuickCaptureNewChild"
                class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                title="Quick capture new child entity"
            >
                <CameraPlusIcon :size="20" />
                <KbdHint shortcut="⇧C" :show="showShortcuts && isSelected" />
            </button>

            <button
                v-if="!editMode"
                @click.stop="emit('createChild', entity.id)"
                class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                title="New entity as child"
            >
                <PlusIcon :size="20" />
                <KbdHint shortcut="⇧N" :show="showShortcuts && isSelected" />
            </button>

            <!-- Edit mode controls -->
            <template v-if="editMode">
                <button
                    @click.stop="handleSave"
                    class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                    title="Save"
                >
                    <CheckIcon :size="20" />
                    <KbdHint shortcut="Enter" :show="showShortcuts" />
                </button>

                <button
                    @click.stop="handleCancel"
                    class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg text-white"
                    title="Cancel"
                >
                    <CloseIcon :size="20" />
                    <KbdHint shortcut="Esc" :show="showShortcuts" />
                </button>

                <button
                    @click.stop="
                        cameraStore.open((files: File[]) =>
                            files.forEach((f) => handleEditArtifact(f)),
                        )
                    "
                    class="relative h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
                    title="Capture artifact"
                >
                    <CameraIcon :size="20" />
                    <KbdHint shortcut="P" :show="showShortcuts" />
                </button>
            </template>
        </div>

        <!-- Images -->
        <div class="flex flex-row justify-center w-full">
            <template v-if="!editMode && images.length > 0">
                <template v-for="n in images" :key="n">
                    <ArtifactImage
                        class="flex-1 object-cover w-full h-56 rounded-xl"
                        :artifact-id="n"
                        :alt="`Artifact ${n}`"
                    />
                </template>
            </template>

            <template v-else-if="editMode && images.length > 0">
                <template v-for="n in images" :key="n">
                    <div class="relative flex-1">
                        <ArtifactImage
                            class="object-cover w-full h-56 rounded-xl transition-opacity"
                            :class="
                                pendingDeletions.has(n)
                                    ? 'opacity-30'
                                    : 'opacity-100'
                            "
                            :artifact-id="n"
                            :alt="`Artifact ${n}`"
                        />
                        <button
                            type="button"
                            @click.stop="toggleArtifactDeletion(n)"
                            class="absolute top-1 right-1 w-6 h-6 flex items-center justify-center rounded-full text-white text-sm leading-none transition-colors"
                            :class="
                                pendingDeletions.has(n)
                                    ? 'bg-gray-400 hover:bg-gray-500'
                                    : 'bg-red-500 hover:bg-red-600'
                            "
                            :title="
                                pendingDeletions.has(n)
                                    ? 'Undo removal'
                                    : 'Remove artifact'
                            "
                        >
                            <CloseIcon :size="16" />
                        </button>
                    </div>
                </template>
            </template>
        </div>
    </figure>
</template>
