<script setup lang="ts" name="EntityCard">
import { ref, computed } from 'vue';
import { useEntitiesStore } from '@/stores/entities';
import { useCameraStore } from '@/stores/camera';
import { useClipStore } from '@/stores/clip';
import { useToastsStore } from '@/stores/toasts';
import { api } from '@/api';
import type { Entity } from '@/api/types';
import TrashCanIcon from 'vue-material-design-icons/TrashCan.vue';
import FolderMoveIcon from 'vue-material-design-icons/FolderMove.vue';
import PencilIcon from 'vue-material-design-icons/Pencil.vue';
import CameraIcon from 'vue-material-design-icons/Camera.vue';
import PlusIcon from 'vue-material-design-icons/Plus.vue';
import CheckIcon from 'vue-material-design-icons/Check.vue';
import CloseIcon from 'vue-material-design-icons/Close.vue';

const props = defineProps<{
  entity: Entity;
}>();

const emit = defineEmits<{
  entityUpdated: [entity: Entity];
  createChild: [locationId: number];
}>();

const entitiesStore = useEntitiesStore();
const cameraStore = useCameraStore();
const clipStore = useClipStore();
const toastsStore = useToastsStore();

const editMode = ref(false);
const localEntity = ref<Entity>({
  ...props.entity,
  metadata: { ...props.entity.metadata },
});
const pendingDeletions = ref<Set<number>>(new Set());

const handleUpdate = async (): Promise<void> => {
  try {
    await Promise.all([...pendingDeletions.value].map(id => api.deleteArtifact(id)));
    const artifacts = (localEntity.value.artifacts ?? []).filter(id => !pendingDeletions.value.has(id));
    await api.patchEntity(props.entity.id, { ...localEntity.value, artifacts });
    pendingDeletions.value = new Set();
    await entitiesStore.reload();
    editMode.value = false;
    emit('entityUpdated', localEntity.value);
    toastsStore.add('Entity updated');
  } catch (error) {
    console.error('Failed to update entity:', error);
    toastsStore.add('Failed to update entity');
  }
};

const handleDelete = async (): Promise<void> => {
  try {
    await api.deleteEntity(props.entity.id);
    await entitiesStore.reload();
    toastsStore.add('Entity deleted');
  } catch (error) {
    console.error('Failed to delete entity:', error);
    toastsStore.add('Failed to delete entity');
  }
};

const handleMove = async (): Promise<void> => {
  const targetId = prompt('Enter target entity ID:');
  if (targetId === null) return;
  const target = parseInt(targetId, 10);
  if (isNaN(target)) {
    toastsStore.add('Invalid entity ID');
    return;
  }
  try {
    await api.moveEntity(props.entity.id, target);
    await entitiesStore.reload();
    toastsStore.add('Entity moved');
  } catch (error) {
    console.error('Failed to move entity:', error);
    toastsStore.add('Failed to move entity');
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
    emit('entityUpdated', props.entity);
    toastsStore.add('Artifact captured and added');
  } catch (error) {
    console.error('Failed to capture artifact:', error);
    toastsStore.add('Failed to capture artifact');
  }
};

const handleCreateChild = async (): Promise<void> => {
  try {
    const newId = await api.createEntity({
      name: 'New Entity',
      description: null,
      artifacts: null,
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
    entitiesStore.setCurrentEntity(newId);
    toastsStore.add('Entity created');
  } catch (error) {
    console.error('Failed to create entity:', error);
    toastsStore.add('Failed to create entity');
  }
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
  const artifacts = editMode.value ? localEntity.value.artifacts : props.entity.artifacts;
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
    emit('entityUpdated', updatedEntity);
    toastsStore.add('Artifact uploaded');
  } catch (error) {
    console.error('Failed to upload artifact:', error);
    toastsStore.add('Failed to upload artifact');
  }
};
</script>

<template>
  <figure
    class="relative h-full max-w-sm bg-white shadow-md dark:bg-gray-800 rounded-xl ring-1 ring-gray-500/25 hover:ring-gray-500/50 hover:shadow-lg flex flex-col"
  >
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
        class="relative"
        title="Visual similarity score"
      >
        <div
          class="text-xs text-gray-400 px-1 rounded bg-gray-100 dark:bg-gray-700 cursor-default"
          :title="`Visual: ${(clipStore.scores[entity.id]! * 100).toFixed(1)}%`"
        >
          {{ Math.round(clipStore.scores[entity.id]! * 100) }}%
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-4 flex-auto flex flex-col">
      <!-- Title -->
      <div v-if="!editMode">
        <div
          class="flex list-reset space-x-2 items-baseline mb-2 cursor-pointer"
          @click="entitiesStore.setCurrentEntity(entity.id)"
        >
          <div
            class="text-xl w-min font-medium text-gray-500 dark:text-gray-400"
          >
            ({{ entity.id }})
          </div>
          <div class="text-xl font-bold">
            {{
              entity.metadata.quantity !== null && entity.metadata.quantity !== 0
                ? `${entity.name || 'Entity'} (x${entity.metadata.quantity})`
                : entity.name || 'Entity'
            }}
          </div>
        </div>
      </div>

      <!-- Edit mode title -->
      <div v-else>
        <div class="flex-auto flex list-reset space-x-2 items-baseline mb-2">
          <div class="text-xl w-min font-medium text-gray-500 dark:text-gray-400">
            ({{ entity.id }})
          </div>
          <input
            type="text"
            v-model.lazy="localEntity.name"
            class="bg-white rounded-sm dark:bg-gray-900 ring-1"
            placeholder="Name"
          />
          <input
            type="number"
            min="0"
            v-model.number.lazy="localEntity.metadata.quantity"
            class="bg-white rounded-sm dark:bg-gray-900 ring-1 w-10"
            placeholder="Qty"
          />
          <label class="flex items-center gap-1 text-sm text-gray-500 dark:text-gray-400 cursor-pointer">
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
        <p class="text-gray-600 dark:text-gray-400">{{ entity.description }}</p>
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
        >
          <div
            v-for="childId in entitiesStore.listChildLocations(entity.id)"
            :key="childId"
            class="p-2 rounded cursor-pointer bg-gray-50 dark:bg-gray-800 dark:hover:bg-gray-700 hover:bg-gray-100 hover:shadow-sm ring-gray-200 dark:ring-slate-500 ring-1 hover:ring-blue-500/75 active:shadow-md"
            @click="entitiesStore.setCurrentEntity(childId)"
          >
            {{ entitiesStore.readname(childId) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Action buttons -->
    <div class="p-4 flex flex-wrap gap-2 border-t dark:border-gray-700">
      <button
        v-if="!editMode"
        @click="handleDelete"
        class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg text-white"
        title="Delete entity"
      >
        <TrashCanIcon :size="20" />
      </button>

      <button
        v-if="!editMode"
        @click="handleMove"
        class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
        title="Move entity"
      >
        <FolderMoveIcon :size="20" />
      </button>

      <button
        v-if="!editMode"
        @click="handleEditToggle"
        class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
        title="Edit entity"
      >
        <PencilIcon :size="20" />
      </button>

      <button
        v-if="!editMode"
        @click="handleQuickCapture"
        class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
        title="Quick capture"
      >
        <CameraIcon :size="20" />
      </button>

      <button
        @click="emit('createChild', entity.id)"
        class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
        title="Create child entity"
      >
        <PlusIcon :size="20" />
      </button>

      <!-- Edit mode controls -->
      <template v-if="editMode">
        <button
          @click="handleSave"
          class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
          title="Save"
        >
          <CheckIcon :size="20" />
        </button>

        <button
          @click="handleCancel"
          class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-red-500 rounded-full shadow hover:bg-red-600 active:shadow-lg text-white"
          title="Cancel"
        >
          <CloseIcon :size="20" />
        </button>
      </template>

      <!-- Quick capture in edit mode -->
      <button
        v-if="editMode"
        @click="
          cameraStore.open((files: File[]) => {
            const tempFiles = files;
            tempFiles.forEach((f) => {
              handleEditArtifact(f);
            });
          });
        "
        class="h-10 w-10 p-0 m-0 flex items-center justify-center bg-blue-500 rounded-full shadow hover:bg-blue-600 active:shadow-lg text-white"
        title="Capture artifact"
      >
        <CameraIcon :size="20" />
      </button>
    </div>

    <!-- Images -->
    <div class="flex flex-row justify-center w-full">
      <template v-if="!editMode && images.length > 0">
        <template v-for="n in images" :key="n">
          <img
            class="flex-1 object-cover w-full h-56 rounded-xl"
            :src="`/api/artifact/${n}`"
            :alt="`Artifact ${n}`"
          />
        </template>
      </template>

      <template v-else-if="editMode && images.length > 0">
        <template v-for="n in images" :key="n">
          <div class="relative flex-1">
            <img
              class="object-cover w-full h-56 rounded-xl transition-opacity"
              :class="pendingDeletions.has(n) ? 'opacity-30' : 'opacity-100'"
              :src="`/api/artifact/${n}`"
              :alt="`Artifact ${n}`"
            />
            <button
              type="button"
              @click="toggleArtifactDeletion(n)"
              class="absolute top-1 right-1 w-6 h-6 flex items-center justify-center rounded-full text-white text-sm leading-none transition-colors"
              :class="pendingDeletions.has(n) ? 'bg-gray-400 hover:bg-gray-500' : 'bg-red-500 hover:bg-red-600'"
              :title="pendingDeletions.has(n) ? 'Undo removal' : 'Remove artifact'"
            >
              <CloseIcon :size="16" />
            </button>
          </div>
        </template>
      </template>
    </div>
  </figure>
</template>
