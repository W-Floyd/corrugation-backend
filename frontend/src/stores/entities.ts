import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { Entity, FullState } from "@/api/types";
import { api, apiFetch } from "@/api";

interface LocationNode {
  id: number;
  name: string;
}

export const useEntitiesStore = defineStore("entities", () => {
  // State
  const fullstate = ref<FullState>({
    entities: {},
    artifacts: {},
    storeversion: -1,
  });

  const searchtextpredebounce = ref("");
  const searchtext = ref("");
  const filterToMissingImage = ref(false);
  const filterToOnlyImage = ref(false);
  const searching = ref(false);
  const filterworld = ref(false);
  const searchdescription = ref(true);
  const currentEntity = ref<number>(0);
  const locationtree = ref<number[]>([]);

  // WebSocket
  let ws: WebSocket | null = null;

  // Actions
  async function loadFullState(): Promise<void> {
    try {
      const versionResponse = await apiFetch("/api/store/version");
      const newVersion = await versionResponse.json();

      if (fullstate.value.storeversion !== newVersion) {
        fullstate.value.storeversion = newVersion;
        const response = await api.getFullState();
        fullstate.value = response;
      }
    } catch (error) {
      console.error("Failed to load full state:", error);
    }
  }

  async function reload(): Promise<void> {
    await loadFullState();
    await loadLocationTree();
  }

  function connectWS(): void {
    const protocol = location.protocol === "https:" ? "wss" : "ws";
    const url = `${protocol}://${location.host}/ws`;

    ws = new WebSocket(url);

    ws.onmessage = () => {
      reload();
    };

    ws.onclose = () => {
      setTimeout(() => connectWS(), 3000);
    };
  }

  function hasChildren(entityId: number): boolean {
    for (const key in fullstate.value.entities) {
      if (fullstate.value.entities[key]?.location === entityId) {
        return true;
      }
    }
    return false;
  }

  function listChildLocations(entityId: number): number[] {
    const childLocations: number[] = [];
    for (const key in fullstate.value.entities) {
      if (fullstate.value.entities[key]?.location === entityId) {
        childLocations.push(parseInt(key, 10));
      }
    }
    return childLocations.sort(sortEntityID);
  }

  function listChildLocationsDeep(entityId: number): number[] {
    const returnValue: number[] = [];
    for (const key in fullstate.value.entities) {
      if (fullstate.value.entities[key]?.location === entityId) {
        returnValue.push(parseInt(key, 10));
        returnValue.push(...listChildLocationsDeep(parseInt(key, 10)));
      }
    }
    return returnValue;
  }

  function readname(entityId: number): string {
    if (entityId === 0) {
      return "World";
    }
    const entity = fullstate.value.entities[entityId];
    if (!entity || !entity.name || entity.name === "") {
      return entityId.toString();
    }
    return entity.name;
  }

  async function setCurrentEntity(entityId: number): Promise<void> {
    if (isNaN(entityId)) entityId = 0;
    currentEntity.value = entityId;
    window.location.hash = entityId === 0 ? "" : entityId.toString();
    searchtext.value = "";
    await reload();
    if (entityId !== 0 && !fullstate.value.entities[entityId]) {
      await setCurrentEntity(0);
    }
  }

  function debouncesearch(): void {
    searchtext.value = searchtextpredebounce.value;
  }

  function selectImages(entityId: number): number[] {
    const entity = fullstate.value.entities[entityId];
    if (!entity || !entity.artifacts || entity.artifacts.length === 0) {
      return [];
    }

    const images: number[] = [];
    for (const artifactId of entity.artifacts) {
      const artifact = fullstate.value.artifacts[artifactId];
      if (artifact && artifact.image) {
        images.push(artifactId);
      }
    }
    return images;
  }

  function recurseLocationTree(entityId: number): void {
    locationtree.value.push(entityId);
    if (entityId !== 0) {
      const elem = fullstate.value.entities[entityId];
      if (elem) {
        recurseLocationTree(elem.location);
      }
    }
  }

  async function loadLocationTree(): Promise<void> {
    locationtree.value = [];
    recurseLocationTree(currentEntity.value);
    locationtree.value.reverse();
  }

  function load(matchId: number, searchText: string): Entity[] {
    if (searchText !== "") {
      searching.value = true;
      const children: number[] = [];

      if (filterworld.value) {
        const childIds = listChildLocationsDeep(0);
        childIds.forEach((id) => children.push(id));
      } else {
        const childIds = listChildLocationsDeep(currentEntity.value);
        childIds.forEach((id) => children.push(id));
      }

      if (searchText.includes("filter:missing-image")) {
        filterToMissingImage.value = true;
        filterToOnlyImage.value = false;
        searchText = searchText.replace("filter:missing-image", "");
      } else {
        filterToMissingImage.value = false;
      }

      if (searchText.includes("filter:only-image")) {
        filterToMissingImage.value = false;
        filterToOnlyImage.value = true;
        searchText = searchText.replace("filter:only-image", "");
      } else {
        filterToOnlyImage.value = false;
      }

      const results: Entity[] = [];
      for (const cid of children) {
        const id = cid.toString();
        const entity = fullstate.value.entities[parseInt(id, 10)];
        if (!entity) continue;

        const nameMatch = entity.name
          ?.toLowerCase()
          .includes(searchText.toLowerCase());
        const descMatch = entity.description
          ?.toLowerCase()
          .includes(searchText.toLowerCase());
        const idMatch = id === searchText.toLowerCase();

        if (nameMatch || descMatch || idMatch) {
          const hasImages = entity.artifacts && entity.artifacts.length > 0;
          const hasNoImages =
            !entity.artifacts || entity.artifacts.length === 0;

          if (
            filterToMissingImage.value &&
            !(hasNoImages || (filterToOnlyImage.value && hasImages))
          ) {
            continue;
          }

          if (
            id === searchText ||
            entity.name?.toLowerCase() === searchText.toLowerCase()
          ) {
            results.unshift(entity);
          } else {
            results.push(entity);
          }
        }
      }

      return results;
    } else {
      searching.value = false;
      const childIDs: number[] = [];
      const childEntities: Entity[] = [];

      for (const key in fullstate.value.entities) {
        if (fullstate.value.entities[key]?.location === matchId) {
          childIDs.push(parseInt(key, 10));
        }
      }

      childIDs.sort(sortEntityID).forEach((id) => {
        childEntities.push(fullstate.value.entities[id]!);
      });

      return childEntities;
    }
  }

  // Computed
  const isLoading = computed(() => fullstate.value.storeversion === -1);

  return {
    fullstate,
    searchtextpredebounce,
    searchtext,
    filterToMissingImage,
    filterToOnlyImage,
    searching,
    filterworld,
    searchdescription,
    currentEntity,
    locationtree,
    isLoading,
    loadFullState,
    reload,
    connectWS,
    hasChildren,
    listChildLocations,
    listChildLocationsDeep,
    readname,
    setCurrentEntity,
    debouncesearch,
    selectImages,
    recurseLocationTree,
    loadLocationTree,
    load,
  };
});

function sortEntityID(a: number, b: number): number {
  const { fullstate } = useEntitiesStore();
  const ea = fullstate.entities[a];
  const eb = fullstate.entities[b];

  if (!ea || !eb) return 0;

  const fa = ea.name?.toLowerCase() ?? "";
  const fb = eb.name?.toLowerCase() ?? "";

  const collator = new Intl.Collator([], { numeric: true });
  let retval = collator.compare(fa, fb);

  if (retval !== 0) {
    return retval;
  }

  const faDesc = ea.description ? ea.description.toLowerCase() : "";
  const fbDesc = eb.description ? eb.description.toLowerCase() : "";

  retval = collator.compare(faDesc, fbDesc);

  if (retval !== 0) {
    return retval;
  }

  return collator.compare(a.toString(), b.toString());
}
