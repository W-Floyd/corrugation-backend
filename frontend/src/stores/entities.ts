import { defineStore } from "pinia";
import { ref, computed, watch } from "vue";
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
  const moveSearchtext = ref("");
  const selectedEntityId = ref<number | null>(null);
  const filterToMissingImage = ref(false);
  const filterToOnlyImage = ref(false);
  const searching = ref(false);
  const filterworld = ref(false);
  const apiSearchResults = ref<Entity[]>([]);
  const searchdescription = ref(true);
  const currentEntity = ref<number>(0);
  const locationtree = ref<number[]>([]);

  // WebSocket
  let ws: WebSocket | null = null;

  // Actions
  async function loadFullState(): Promise<void> {
    DEBUG && console.log("[entities] loadFullState — token in storage:", !!localStorage.getItem("auth_token"));
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
    DEBUG && console.log("[entities] connectWS", url, new Error().stack?.split("\n")[2]?.trim());

    ws = new WebSocket(url);

    ws.onopen = () => {
      DEBUG && console.log("[entities] WS connected");
      reload();
    };
    ws.onmessage = () => {
      DEBUG && console.log("[entities] WS message → reload");
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
    import("../router").then(({ default: router }) => {
      router.push({ query: entityId === 0 ? {} : { entity: entityId } });
    });
    searchtext.value = "";
    await reload();
    if (entityId !== 0 && !fullstate.value.entities[entityId]) {
      await setCurrentEntity(0);
    }
  }

  function debouncesearch(): void {
    searchtext.value = searchtextpredebounce.value;
  }

  watch([searchtext, filterworld, currentEntity], async ([text]) => {
    if (!text.trim()) {
      searching.value = false;
      apiSearchResults.value = [];
      filterToMissingImage.value = false;
      filterToOnlyImage.value = false;
      return;
    }

    let query = text;
    filterToMissingImage.value = query.includes("filter:missing-image");
    filterToOnlyImage.value = !filterToMissingImage.value && query.includes("filter:only-image");
    query = query.replace("filter:missing-image", "").replace("filter:only-image", "").trim();

    searching.value = true;
    try {
      if (query) {
        const scopeId = !filterworld.value && currentEntity.value !== 0
          ? currentEntity.value
          : undefined;
        apiSearchResults.value = await api.searchEntities(query, scopeId);
      } else {
        const childIds = filterworld.value
          ? listChildLocationsDeep(0)
          : listChildLocationsDeep(currentEntity.value);
        apiSearchResults.value = childIds
          .map((id) => fullstate.value.entities[id])
          .filter((e): e is Entity => e != null);
      }
    } catch (e) {
      console.error("Search failed:", e);
      apiSearchResults.value = [];
    } finally {
      searching.value = false;
    }
  });

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
      let results = [...apiSearchResults.value];
      if (filterToMissingImage.value) {
        results = results.filter((e) => !e.artifacts || e.artifacts.length === 0);
      } else if (filterToOnlyImage.value) {
        results = results.filter((e) => e.artifacts && e.artifacts.length > 0);
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
    moveSearchtext,
    selectedEntityId,
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
