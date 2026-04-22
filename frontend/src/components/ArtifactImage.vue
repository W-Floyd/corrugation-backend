<template>
    <img v-if="objectUrl" :src="objectUrl" :alt="alt" v-bind="$attrs" />
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from "vue";
import { apiFetch } from "@/api";

const props = defineProps<{
    artifactId: number;
    alt?: string;
}>();

const objectUrl = ref<string | null>(null);

async function load(id: number) {
    try {
        const response = await apiFetch(`/api/artifact/${id}`);
        const blob = await response.blob();
        const url = URL.createObjectURL(blob);
        if (objectUrl.value) URL.revokeObjectURL(objectUrl.value);
        objectUrl.value = url;
    } catch {
        objectUrl.value = null;
    }
}

watch(() => props.artifactId, load, { immediate: true });

onUnmounted(() => {
    if (objectUrl.value) URL.revokeObjectURL(objectUrl.value);
});
</script>
