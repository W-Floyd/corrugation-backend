<script setup lang="ts" name="BreadcrumbNav">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useEntitiesStore } from '@/stores/entities';
import type { Entity } from '@/api/types';

const router = useRouter();
const entitiesStore = useEntitiesStore();

const locationTree = computed(() => {
  const treeIds = entitiesStore.locationtree;
  return treeIds
    .map((id: number) => {
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
    <ol class="flex flex-wrap list-reset">
      <template v-for="(n, index) in locationTree" :key="n.id">
        <li>
          <a
            @click="navigateTo(n.id)"
            class="text-blue-600 no-underline cursor-pointer dark:text-sky-400 dark:hover:text-sky-300 hover:text-blue-700 hover:underline"
            :title="`Go to entity ${n.id}`"
          >
            {{ n.name }}
          </a>
        </li>

        <li v-if="index < locationTree.length - 1">
          <span class="mx-2 text-gray-500">/</span>
        </li>
      </template>

      <li>
        <button
          @click="entitiesStore.setCurrentEntity(entitiesStore.currentEntity)"
          class="text-blue-600 dark:text-sky-400 dark:hover:text-sky-300 hover:text-blue-700 cursor-pointer ml-2"
          title="Create new entity"
        >
          +
        </button>
      </li>
    </ol>
  </nav>
</template>
