import { defineStore } from "pinia";
import { ref, computed, watch } from "vue";
import type { Entity, BackendRecord } from "@/api/types";
import { recordToEntity } from "@/api/types";
import { api } from "@/api";
import { useToastsStore } from "@/stores/toasts";
import { useAuthStore } from "@/stores/auth";

export const useEntitiesStore = defineStore("entities", () => {
  const currentEntity = ref<number>(0);

  // All records fetched globally — kept in sync via WebSocket
  const allRecords = ref<BackendRecord[]>([]);

  const locationtree = ref<number[]>([0]);

  const searchtextpredebounce = ref("");
  const searchtext = ref("");
  const moveSearchtext = ref("");
  const selectedEntityId = ref<number | null>(null);
  const filterworld = ref(false);
  const searching = ref(false);
  const searchImage = ref(true);
  const searchTextEmbedded = ref(true);
  const searchTextSubstring = ref(true);
  const apiSearchResults = ref<Entity[]>([]);
  const apiSearchScores = ref<Record<number, { image?: number; text?: number }>>({});
  const filterToMissingImage = ref(false);
  const filterToOnlyImage = ref(false);

  const isLoading = ref(true);

  let ws: WebSocket | null = null;

  // All records indexed by ID
  const recordById = computed<Record<number, BackendRecord>>(() => {
    const m: Record<number, BackendRecord> = {};
    for (const r of allRecords.value) m[r.ID] = r;
    return m;
  });

  const entityMap = computed<Record<number, Entity>>(() => {
    const m: Record<number, Entity> = {};
    for (const r of allRecords.value) m[r.ID] = recordToEntity(r);
    return m;
  });

  const nameMap = computed<Record<number, string>>(() => {
    const m: Record<number, string> = { 0: "World" };
    for (const r of allRecords.value) {
      m[r.ID] = r.ReferenceNumber ?? r.Title ?? String(r.ID);
    }
    return m;
  });

  // Derived from allRecords — no extra fetch needed
  function buildLocationTree(entityId: number): number[] {
    if (entityId === 0) return [0];
    const tree: number[] = [];
    let cur: number | undefined = entityId;
    while (cur !== undefined && cur !== 0) {
      tree.push(cur);
      cur = recordById.value[cur]?.ParentID ?? undefined;
    }
    tree.push(0);
    tree.reverse();
    return tree;
  }

  async function reload(): Promise<void> {
    try {
      allRecords.value = await api.getRecords(0, { global: true });
      locationtree.value = buildLocationTree(currentEntity.value);
      isLoading.value = false;
    } catch (e) {
      console.error("reload failed:", e);
    }
  }

  type PartialScope = { scopeId: number | undefined; searchImage: boolean; searchTextEmbedded: boolean; searchId: string | null };
  let partialScope: PartialScope | null = null;
  let progressToastId: number | null = null;

  function clearProgressToast(): void {
    if (progressToastId !== null) {
      useToastsStore().remove(progressToastId);
      progressToastId = null;
    }
    partialScope = null;
  }

  async function fetchAndUpdateProgress(scope: PartialScope): Promise<void> {
    try {
      const progress = await api.getSearchEmbeddingProgress({
        id: scope.scopeId,
        global: scope.scopeId == null,
        childrenDepth: scope.scopeId != null ? -1 : undefined,
        searchImage: scope.searchImage,
        searchTextEmbedded: scope.searchTextEmbedded,
      });
      if (progressToastId === null) return; // toast was dismissed
      if (progress.ready) {
        useToastsStore().update(progressToastId, "Embeddings ready — re-run search for complete results", "info");
        useToastsStore().finalize(progressToastId);
        progressToastId = null;
        partialScope = null;
      } else {
        useToastsStore().update(
          progressToastId,
          `Indexing ${progress.indexed}/${progress.total} embeddings — results may be incomplete`,
        );
      }
    } catch {
      // ignore — will retry on next embedding_progress message
    }
  }

  function connectWS(): void {
    const protocol = location.protocol === "https:" ? "wss" : "ws";
    const token = localStorage.getItem("auth_token");
    const url = `${protocol}://${location.host}/ws${token ? `?token=${encodeURIComponent(token)}` : ""}`;
    DEBUG && console.log("[entities] connectWS", url);
    ws = new WebSocket(url);
    ws.onopen = () => { reload(); };
    ws.onmessage = async (e) => {
      if (typeof e.data === "string" && e.data.startsWith("embedding_progress")) {
        const msgSearchId = e.data.includes(":") ? e.data.split(":")[1] : null;
        const hasSearchId = msgSearchId !== null && msgSearchId !== "";
        if (useAuthStore().authConfig.enabled && !hasSearchId) return;
        if (partialScope !== null && (!hasSearchId || msgSearchId === partialScope.searchId)) {
          fetchAndUpdateProgress(partialScope);
        }
      } else {
        reload();
      }
    };
    ws.onclose = () => { setTimeout(() => connectWS(), 3000); };
  }

  async function setCurrentEntity(entityId: number): Promise<void> {
    if (isNaN(entityId)) entityId = 0;
    currentEntity.value = entityId;
    locationtree.value = buildLocationTree(entityId);
    import("../router").then(({ default: router }) => {
      router.push({ query: entityId === 0 ? {} : { entity: entityId } });
    });
    searchtext.value = "";
  }

  function readname(entityId: number): string {
    if (entityId === 0) return "World";
    return nameMap.value[entityId] ?? String(entityId);
  }

  function hasChildren(entityId: number): boolean {
    return allRecords.value.some((r) => (r.ParentID ?? 0) === entityId);
  }

  function listChildLocations(entityId: number): number[] {
    return allRecords.value
      .filter((r) => (r.ParentID ?? 0) === entityId)
      .map((r) => r.ID)
      .sort((a, b) => {
        const na = nameMap.value[a]?.toLowerCase() ?? "";
        const nb = nameMap.value[b]?.toLowerCase() ?? "";
        return na.localeCompare(nb, undefined, { numeric: true });
      });
  }

  function load(locationId: number, searchTextVal: string): Entity[] {
    if (searchTextVal.trim()) {
      let results = [...apiSearchResults.value];
      if (filterToMissingImage.value) {
        results = results.filter((e) => !e.artifacts || e.artifacts.length === 0);
      } else if (filterToOnlyImage.value) {
        results = results.filter((e) => e.artifacts && e.artifacts.length > 0);
      }
      return results;
    }
    return allRecords.value
      .filter((r) => (r.ParentID ?? 0) === locationId)
      .map(recordToEntity)
      .sort((a, b) => {
        const na = (a.name ?? "").toLowerCase();
        const nb = (b.name ?? "").toLowerCase();
        return na.localeCompare(nb, undefined, { numeric: true });
      });
  }

  function debouncesearch(): void {
    searchtext.value = searchtextpredebounce.value;
  }

  watch(
    [searchtext, filterworld, currentEntity, searchImage, searchTextEmbedded, searchTextSubstring],
    async ([text]) => {
      if (!text.trim()) {
        searching.value = false;
        apiSearchResults.value = [];
        apiSearchScores.value = {};
        filterToMissingImage.value = false;
        filterToOnlyImage.value = false;
        clearProgressToast();
        return;
      }

      let query = text;
      filterToMissingImage.value = query.includes("filter:missing-image");
      filterToOnlyImage.value = !filterToMissingImage.value && query.includes("filter:only-image");
      query = query.replace("filter:missing-image", "").replace("filter:only-image", "").trim();

      clearProgressToast();
      searching.value = true;
      try {
        if (query) {
          const scopeId = !filterworld.value && currentEntity.value !== 0
            ? currentEntity.value
            : undefined;
          const { results, partial, searchId } = await api.searchRecords(query, {
            parentId: scopeId,
            searchImage: searchImage.value,
            searchTextEmbedded: searchTextEmbedded.value,
            searchTextSubstring: searchTextSubstring.value,
          });
          if (partial) {
            const scope: PartialScope = { scopeId, searchImage: searchImage.value, searchTextEmbedded: searchTextEmbedded.value, searchId };
            partialScope = scope;
            progressToastId = useToastsStore().add("Indexing embeddings — results may be incomplete", "warn", true);
            fetchAndUpdateProgress(scope); // populates the count async
          } else {
            partialScope = null;
          }
          apiSearchResults.value = results.map((r) => r.entity);
          const scores: Record<number, { image?: number; text?: number }> = {};
          for (const r of results) {
            scores[r.entity.id] = { image: r.imageScore, text: r.textScore };
          }
          apiSearchScores.value = scores;
        } else {
          apiSearchResults.value = [];
        }
      } catch (e) {
        console.error("Search failed:", e);
        apiSearchResults.value = [];
      } finally {
        searching.value = false;
      }
    },
  );

  // Clear progress toast on navigation
  watch(currentEntity, () => clearProgressToast());

  return {
    currentEntity,
    allRecords,
    nameMap,
    entityMap,
    locationtree,
    searchtextpredebounce,
    searchtext,
    moveSearchtext,
    selectedEntityId,
    filterworld,
    searching,
    searchImage,
    searchTextEmbedded,
    searchTextSubstring,
    apiSearchResults,
    apiSearchScores,
    filterToMissingImage,
    filterToOnlyImage,
    isLoading,
    reload,
    connectWS,
    setCurrentEntity,
    readname,
    hasChildren,
    listChildLocations,
    load,
    debouncesearch,
  };
});
