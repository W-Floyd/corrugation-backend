<script setup lang="ts" name="MoveEntityDialog">
import { ref, watch, computed, nextTick } from "vue";
import { useEntitiesStore } from "@/stores/entities";
import { useToastsStore } from "@/stores/toasts";
import { api } from "@/api";
import type { Entity } from "@/api/types";

const entitiesStore = useEntitiesStore();
const toastsStore = useToastsStore();

const props = withDefaults(
    defineProps<{
        visible?: boolean;
        targetEntityId?: number;
    }>(),
    {
        visible: false,
        targetEntityId: 0,
    },
);

const emit = defineEmits<{
    moved: [entityId: number, newLocation: number];
    "update:visible": [value: boolean];
}>();

const dialogVisible = ref(false);
const entity = ref<Entity | null>(null);
const searchtext = ref("");
const targetLocation = ref<number>(0);
const searchInputRef = ref<HTMLInputElement | null>(null);

const currentLocationName = computed(() => {
    if (entitiesStore.currentEntity === 0) {
        return "World";
    }
    const entity =
        entitiesStore.fullstate.entities[entitiesStore.currentEntity];
    return entity?.name || entitiesStore.currentEntity.toString();
});

watch(
    () => props.visible,
    (visible) => {
        dialogVisible.value = visible;
        if (visible && props.targetEntityId) {
            entity.value =
                entitiesStore.fullstate.entities[props.targetEntityId] || null;
            nextTick(() => {
                searchInputRef.value?.focus();
            });
        }
    },
    { immediate: true },
);

watch(
    () => props.targetEntityId,
    (id) => {
        if (id && dialogVisible.value) {
            entity.value = entitiesStore.fullstate.entities[id] || null;
        }
    },
);

const filteredEntities = () => {
    const isDescendant = (entityId: number): boolean => {
        let current = entityId;
        while (current !== 0) {
            if (current === entity.value?.id) return true;
            const parent = entitiesStore.fullstate.entities[current];
            if (!parent) break;
            current = parent.location;
        }
        return false;
    };

    const getFilteredEntities = (): Entity[] => {
        const result = Object.values(entitiesStore.fullstate.entities).filter(
            (e) => e.id !== entity.value?.id && !isDescendant(e.id),
        );
        // Add World (0) to the results if it's not being moved
        if (entity.value?.id !== 0) {
            result.push({
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
            });
        }
        return result;
    };

    if (!searchtext.value.trim()) {
        return getFilteredEntities();
    }

    const term = searchtext.value.toLowerCase();
    const results = Object.values(entitiesStore.fullstate.entities).filter(
        (e) =>
            e.id !== entity.value?.id &&
            !isDescendant(e.id) &&
            (e.name?.toLowerCase().includes(term) ||
                e.description?.toLowerCase().includes(term) ||
                e.id.toString().includes(term)),
    );

    if (term && entity.value?.id !== 0) {
        const worldMatch =
            "World".toLowerCase().includes(term) || "0".includes(term);
        if (worldMatch) {
            results.push({
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
            });
        }
    }

    return results;
};

const searchResults = computed(() => filteredEntities());

watch(searchResults, (results) => {
    if (
        searchtext.value.trim() &&
        results.length > 0 &&
        targetLocation.value !== 0
    ) {
        const hasSelected = results.some((r) => r.id === targetLocation.value);
        if (!hasSelected) {
            const first = results[0];
            if (first) {
                targetLocation.value = first.id;
            }
        }
    } else if (
        searchtext.value.trim() &&
        results.length > 0 &&
        targetLocation.value === 0
    ) {
        const first = results[0];
        if (first) {
            targetLocation.value = first.id;
        }
    }
});

const handleMove = async (): Promise<void> => {
    if (!entity.value) {
        console.log("MoveEntityDialog: entity.value is null");
        return;
    }
    console.log(
        "MoveEntityDialog: Moving entity",
        entity.value.id,
        "to location",
        targetLocation.value,
    );

    try {
        await api.moveEntity(entity.value.id, targetLocation.value);
        console.log("MoveEntityDialog: moveEntity succeeded");
        await entitiesStore.reload();
        console.log("MoveEntityDialog: reload succeeded");
        emit("moved", entity.value.id, targetLocation.value);
        emit("update:visible", false);
        dialogVisible.value = false;
        toastsStore.add("Entity moved");
    } catch (error) {
        console.error("Failed to move entity:", error);
        toastsStore.add("Failed to move entity");
    }
};

const moveToCurrentLocation = async (): Promise<void> => {
    if (!entity.value) {
        console.log("MoveEntityDialog: entity.value is null");
        return;
    }
    console.log(
        "MoveEntityDialog: Moving entity",
        entity.value.id,
        "to current location",
        entitiesStore.currentEntity,
    );
    try {
        await api.moveEntity(entity.value.id, entitiesStore.currentEntity);
        console.log("MoveEntityDialog: moveEntity succeeded");
        await entitiesStore.reload();
        console.log("MoveEntityDialog: reload succeeded");
        emit("moved", entity.value.id, entitiesStore.currentEntity);
        emit("update:visible", false);
        dialogVisible.value = false;
        toastsStore.add("Entity moved to current location");
    } catch (error) {
        console.error("Failed to move entity:", error);
        toastsStore.add("Failed to move entity");
    }
};

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
    return `(${entityId}) ${tree.join("/")}`;
};

const handleDialogClose = (): void => {
    dialogVisible.value = false;
    emit("update:visible", false);
};
</script>

<template>
    <Teleport to="body">
        <div
            v-if="dialogVisible"
            class="fixed inset-0 overflow-y-auto z-50"
            role="dialog"
            aria-modal="true"
        >
            <!-- Overlay -->
            <div
                class="fixed inset-0 bg-black/40"
                @click="handleDialogClose"
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

                    <div v-if="entity">
                        <p class="text-gray-600 dark:text-gray-400 mb-4">
                            Moving: {{ entity.name || entity.id }}
                        </p>

                        <!-- Search -->
                        <div class="mb-4">
                            <input
                                ref="searchInputRef"
                                v-model="searchtext"
                                type="search"
                                placeholder="Search for a location..."
                                class="w-full px-4 py-2 rounded-full bg-white ring-1 dark:bg-gray-900"
                            />
                        </div>

                        <!-- Location selector -->
                        <div class="mb-4">
                            <label for="location" class="block mb-2">
                                Select location
                            </label>
                            <select
                                v-model="targetLocation"
                                id="location"
                                class="w-full px-4 py-2 rounded-full bg-white ring-1 dark:bg-gray-900"
                            >
                                <option value="" disabled>
                                    Select a location
                                </option>
                                <template
                                    v-for="loc in filteredEntities()"
                                    :key="loc.id"
                                >
                                    <option :value="loc.id">
                                        {{ formatOption(loc.id) }}
                                    </option>
                                </template>
                            </select>
                        </div>

                        <!-- Info about current entity -->
                        <div
                            v-if="
                                entity && entitiesStore.hasChildren(entity.id)
                            "
                            class="mb-4"
                        >
                            <p class="text-gray-600 dark:text-gray-400 mb-2">
                                This entity has children:
                            </p>
                            <div class="flex flex-wrap gap-2">
                                <span
                                    v-for="childId in entitiesStore.listChildLocations(
                                        entity.id,
                                    )"
                                    :key="childId"
                                    class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded"
                                >
                                    {{ formatOption(childId) }}
                                </span>
                            </div>
                            <p class="text-sm text-gray-500 mt-2">
                                Note: Children will follow this entity to the
                                new location.
                            </p>
                        </div>
                    </div>

                    <!-- Buttons -->
                    <div class="flex mt-8 space-x-2">
                        <button
                            type="button"
                            @click="handleMove"
                            class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600"
                        >
                            Move
                        </button>
                        <button
                            type="button"
                            @click="moveToCurrentLocation"
                            class="h-10 px-4 py-2 text-white bg-purple-500 rounded-full shadow hover:bg-purple-600"
                        >
                            Move Here ({{ currentLocationName }})
                        </button>
                        <button
                            type="button"
                            @click="handleDialogClose"
                            class="h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600"
                        >
                            Cancel
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>
</template>
