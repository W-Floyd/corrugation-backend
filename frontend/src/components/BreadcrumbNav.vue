<script setup lang="ts" name="BreadcrumbNav">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useEntitiesStore } from '@/stores/entities';
import type { Entity } from '@/api/types';

const router = useRouter();
const entitiesStore = useEntitiesStore();

const emit = defineEmits<{ openNewEntity: [] }>();

const locationTree = computed(() => {
  const treeIds = entitiesStore.locationtree;
  return treeIds
    .map((id: number) => {
      if (id === 0) return { id: 0, name: 'World' };
      const entity = entitiesStore.fullstate.entities[id];
      if (!entity) return null;
      return { id, name: entity.name || id.toString() };
    })
    .filter((item: { id: number; name: string } | null): item is { id: number; name: string } => item !== null);
});

const navigateTo = async (entityId: number): Promise<void> => {
  await entitiesStore.setCurrentEntity(entityId);
};
</script>

<template>
  <nav class="w-full">
    <ol class="flex flex-wrap items-center gap-x-1">
      <template v-for="(n, index) in locationTree" :key="n.id">
        <li>
          <a
            @click="navigateTo(n.id)"
            class="text-blue-600 no-underline cursor-pointer dark:text-sky-400 dark:hover:text-sky-300 hover:text-blue-700 hover:underline px-1"
            :title="`Go to entity ${n.id}`"
          >
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
