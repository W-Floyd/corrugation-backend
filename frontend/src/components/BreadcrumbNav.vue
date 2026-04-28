<script setup lang="ts" name="BreadcrumbNav">
import { computed, ref } from "vue";
import { useEntitiesStore } from "@/stores/entities";
import { useToastsStore } from "@/stores/toasts";
import { api } from "@/api";

const entitiesStore = useEntitiesStore();
const toastsStore = useToastsStore();

const emit = defineEmits<{ openNewEntity: [] }>();

const locationTree = computed(() =>
    entitiesStore.locationtree.map((id: number) => ({
        id,
        name: entitiesStore.readname(id),
    })),
);

const navigateTo = async (entityId: number): Promise<void> => {
    await entitiesStore.setCurrentEntity(entityId);
};

const dragOverId = ref<number | null>(null);

const handleDragOver = (e: DragEvent, id: number): void => {
    if (!e.dataTransfer?.types.includes("entityid")) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = "move";
    dragOverId.value = id;
};

const handleDragLeave = (e: DragEvent): void => {
    if (!(e.currentTarget as HTMLElement).contains(e.relatedTarget as Node)) {
        dragOverId.value = null;
    }
};

const handleDrop = async (e: DragEvent, targetId: number): Promise<void> => {
    dragOverId.value = null;
    const raw = e.dataTransfer?.getData("entityId");
    if (!raw) return;
    const entityId = parseInt(raw, 10);
    if (isNaN(entityId) || entityId === targetId) return;
    try {
        await api.moveRecord(entityId, targetId);
        await entitiesStore.reload();
        toastsStore.add("Entity moved", "info");
    } catch {
        toastsStore.add("Failed to move entity");
    }
};
</script>

<template>
    <nav class="w-full">
        <ol class="flex flex-wrap items-center gap-x-1">
            <template v-for="(n, index) in locationTree" :key="n.id">
                <li>
                    <a @click="navigateTo(n.id)"
                        @dragover="handleDragOver($event, n.id)"
                        @dragleave="handleDragLeave"
                        @drop="handleDrop($event, n.id)"
                        :class="[
                            'text-blue-600 no-underline cursor-pointer dark:text-sky-400 dark:hover:text-sky-300 hover:text-blue-700 hover:underline px-1 rounded transition-colors',
                            dragOverId === n.id ? 'ring-2 ring-green-500 shadow shadow-green-200 dark:shadow-green-900 bg-green-50 dark:bg-green-900/20' : '',
                        ]"
                        :title="`Go to entity ${n.id}`">
                        {{ n.name }}
                    </a>
                </li>

                <li v-if="index < locationTree.length - 1" aria-hidden="true">
                    <span class="text-gray-400">/</span>
                </li>
            </template>
        </ol>
    </nav>
</template>
