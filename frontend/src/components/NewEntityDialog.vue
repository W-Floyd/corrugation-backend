<script setup lang="ts" name="NewEntityDialog">
import { ref, onMounted, watch, nextTick } from "vue";
import KbdHint from "@/components/KbdHint.vue";
import { useEntitiesStore } from "@/stores/entities";
import { useCameraStore } from "@/stores/camera";
import { useToastsStore } from "@/stores/toasts";
import { api } from "@/api";

const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const toastsStore = useToastsStore();

const props = withDefaults(
    defineProps<{
        visible?: boolean;
        location?: number;
        showShortcuts?: boolean;
    }>(),
    {
        visible: false,
        location: 0,
        showShortcuts: false,
    },
);

const emit = defineEmits<{
    created: [entityId: number];
    "update:visible": [value: boolean];
}>();

const dialogVisible = ref(false);
const nameInput = ref<HTMLInputElement | null>(null);
const title = ref<string>("");
const description = ref<string>("");
const quantity = ref<number | null>(null);
const labeled = ref(false);
const referenceNumber = ref<string>("");
const files = ref<File[]>([]);
const nextRefNumber = ref<number>(0);

watch(
    () => props.visible,
    async (visible) => {
        dialogVisible.value = visible;
        if (visible) {
            resetDialog();
            nextTick(() => nameInput.value?.focus());
        }
    },
    { immediate: true },
);

watch(labeled, async (val) => {
    if (val) {
        nextRefNumber.value = await api.nextReferenceNumber();
        referenceNumber.value = String(nextRefNumber.value);
    } else {
        referenceNumber.value = "";
    }
});

const resetDialog = (): void => {
    title.value = "";
    description.value = "";
    quantity.value = null;
    labeled.value = false;
    referenceNumber.value = "";
    files.value = [];
    nextRefNumber.value = 0;
};

const handleSubmit = async (): Promise<void> => {
    try {
        let artifactIds: number[] = [];
        for (const file of files.value) {
            const id = await api.uploadArtifact(file);
            artifactIds.push(id);
        }

        const record = await api.createRecord({
            Title: title.value || null,
            ReferenceNumber: labeled.value ? referenceNumber.value || null : null,
            Labeled: labeled.value,
            Description: description.value || null,
            Quantity: quantity.value ?? undefined,
            ParentID: props.location || undefined,
            Artifacts: artifactIds.length ? artifactIds : undefined,
        });
        await entitiesStore.reload();
        emit("created", record.ID);
        emit("update:visible", false);
        dialogVisible.value = false;
        toastsStore.add("Entity created", "info");
    } catch (error) {
        console.error("Failed to create entity:", error);
        toastsStore.add("Failed to create entity");
    }
};

const handleDialogClose = (): void => {
    dialogVisible.value = false;
    emit("update:visible", false);
};

const handleCameraOpen = async (): Promise<void> => {
    await new Promise<void>((resolve) => {
        cameraStore.open((fileList: File[]) => {
            files.value = fileList;
            resolve();
        });
    });
};

onMounted(() => {
    // nothing needed on mount; ids fetched when dialog opens
});
</script>

<template>
    <Teleport to="body">
        <div v-if="dialogVisible" class="fixed inset-0 overflow-y-auto z-50" role="dialog" aria-modal="true">
            <!-- Overlay -->
            <div class="fixed inset-0 bg-black/40" @click="handleDialogClose"></div>

            <!-- Panel -->
            <div class="relative flex items-center justify-center min-h-screen p-4" @click.stop
                @keydown.esc.stop="handleDialogClose">
                <div
                    class="relative w-full max-w-2xl p-8 overflow-y-auto bg-white border border-gray-300 rounded-lg dark:bg-gray-800">
                    <!-- Title -->
                    <h1 class="pb-4 text-3xl font-medium">Create New Entity</h1>

                    <!-- Form -->
                    <div class="grid grid-cols-[8rem_1fr] gap-x-4 gap-y-3 items-center">
                        <label for="name">Name</label>
                        <input id="name" ref="nameInput" type="text" v-model="title"
                            class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1"
                            @keydown.enter.prevent="handleSubmit" />

                        <label for="labeled">Labeled</label>
                        <input id="labeled" type="checkbox" v-model="labeled" class="w-4 h-4 justify-self-start" />

                        <template v-if="labeled">
                            <label for="refnum">Reference #</label>
                            <input id="refnum" type="text" v-model="referenceNumber"
                                class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1"
                                :placeholder="String(nextRefNumber)" @keydown.enter.prevent="handleSubmit" />
                        </template>

                        <label for="description">Description</label>
                        <textarea id="description" v-model="description"
                            class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1" rows="3"></textarea>

                        <label for="quantity">Quantity</label>
                        <input id="quantity" type="number" min="0" v-model.number="quantity"
                            class="bg-white rounded-sm dark:bg-gray-900 ring-1 px-2 py-1" />

                        <label for="file">Image</label>
                        <div class="flex flex-wrap items-center gap-2">
                            <input id="file" type="file"
                                class="bg-white rounded-sm ring-1 dark:bg-gray-900 dark:hover:bg-gray-700"
                                accept="image/*" multiple @change="
                                    (e) => {
                                        files = Array.from(
                                            (e.target as HTMLInputElement)
                                                .files || [],
                                        );
                                    }
                                " />
                            <button type="button" @click="handleCameraOpen"
                                class="h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600">
                                Camera
                            </button>
                            <span v-if="files.length > 0" class="text-sm text-gray-500">
                                {{ files.length }} file(s) selected
                            </span>
                        </div>
                    </div>

                    <!-- Buttons -->
                    <div class="flex mt-8 gap-4">
                        <button type="button" @click="handleSubmit"
                            class="relative h-10 px-4 py-2 text-white bg-blue-500 rounded-full shadow hover:bg-blue-600">
                            Submit
                            <KbdHint shortcut="Enter" :show="props.showShortcuts" />
                        </button>
                        <button type="button" @click="handleDialogClose"
                            class="relative h-10 px-4 py-2 text-white bg-red-500 rounded-full shadow hover:bg-red-600">
                            Cancel
                            <KbdHint shortcut="Esc" :show="props.showShortcuts" />
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </Teleport>
</template>
