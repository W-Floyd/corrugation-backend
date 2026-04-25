import type { Entity, BackendRecord, RecordBody } from "./types";
import { recordToEntity } from "./types";
import router from "../router";
import { useAuthStore } from "../stores/auth";
import { useToastsStore } from "../stores/toasts";

export async function apiFetch(
  url: string,
  options: RequestInit = {},
): Promise<Response> {
  const token = localStorage.getItem("auth_token");
  const headers = new Headers(options.headers as HeadersInit);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(url, { ...options, headers });

  if (response.status === 401) {
    const currentRoute = router.currentRoute.value.name;
    DEBUG &&
      console.warn(
        "[apiFetch] 401 on",
        url,
        "current route:",
        currentRoute,
        new Error().stack?.split("\n")[2]?.trim(),
      );
    if (currentRoute !== "callback") {
      useAuthStore().clearToken();
      router.push({ name: "login" });
      throw new Error("Unauthorized");
    }
    throw new Error("Unauthorized");
  }

  if (!response.ok) {
    const body = await response.text();
    const message = body || `HTTP ${response.status}`;
    useToastsStore().add(message);
    throw new Error(message);
  }

  return response;
}

export async function withErrorToast<T>(
  fn: () => Promise<T>,
): Promise<T | undefined> {
  try {
    return await fn();
  } catch (e) {
    useToastsStore().add(e instanceof Error ? e.message : String(e));
    return undefined;
  }
}

export const api = {
  // Fetch records at a location. id=0 → top-level. global=true → all.
  async getRecords(
    locationId: number,
    opts: {
      childrenDepth?: number;
      parentDepth?: number;
      global?: boolean;
      timestamps?: boolean;
    } = {},
  ): Promise<BackendRecord[]> {
    const params = new URLSearchParams();
    if (opts.global) {
      params.set("global", "true");
    } else {
      params.set("id", String(locationId));
    }
    if (opts.childrenDepth !== undefined)
      params.set("childrenDepth", String(opts.childrenDepth));
    if (opts.parentDepth !== undefined)
      params.set("parentDepth", String(opts.parentDepth));
    if (opts.timestamps) params.set("timestamps", "true");
    const response = await apiFetch(`/api/v2/records?${params}`);
    return response.json();
  },

  async createRecord(body: RecordBody): Promise<BackendRecord> {
    const response = await apiFetch("/api/v2/record", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    return response.json();
  },

  async updateRecord(id: number, body: RecordBody): Promise<BackendRecord> {
    const response = await apiFetch(`/api/v2/record/${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    return response.json();
  },

  async deleteRecord(id: number): Promise<void> {
    await apiFetch(`/api/v2/record/${id}`, { method: "DELETE" });
  },

  async moveRecord(id: number, locationId: number): Promise<void> {
    await apiFetch(`/api/v2/record/${id}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ ParentID: locationId || null }),
    });
  },

  async searchRecords(
    query: string,
    opts: {
      parentId?: number;
      searchImage?: boolean;
      searchTextEmbedded?: boolean;
      searchTextSubstring?: boolean;
    } = {},
  ): Promise<{ results: { entity: Entity; imageScore?: number; textScore?: number }[]; partial: boolean; searchId: string | null }> {
    const params = new URLSearchParams({ search: query });
    if (opts.parentId != null) {
      params.set("id", String(opts.parentId));
      params.set("childrenDepth", "-1");
    } else {
      params.set("global", "true");
    }
    if (opts.searchImage !== false) params.set("searchImage", "true");
    if (opts.searchTextEmbedded !== false)
      params.set("searchTextEmbedded", "true");
    if (opts.searchTextSubstring !== false)
      params.set("searchTextSubstring", "true");
    const response = await apiFetch(`/api/v2/records?${params}`);
    const partial = response.status === 207;
    const searchId = response.headers.get("X-Search-ID");
    const records: BackendRecord[] = await response.json();
    return {
      partial,
      searchId,
      results: records.map((r) => ({
        entity: recordToEntity(r),
        imageScore: r.SearchConfidenceImage,
        textScore: r.SearchConfidenceText,
      })),
    };
  },

  async uploadArtifact(file: File): Promise<number> {
    const formData = new FormData();
    formData.append("file", file);
    const response = await apiFetch("/api/v2/artifact", {
      method: "POST",
      body: formData,
    });
    const id = await response.json();
    return typeof id === "number" ? id : parseInt(id, 10);
  },

  async deleteArtifact(id: number): Promise<void> {
    await apiFetch(`/api/v2/artifact/${id}`, { method: "DELETE" });
  },

  // Next free record ID (for assigning a specific DB id — legacy use)
  async nextFreeId(): Promise<number> {
    const response = await apiFetch("/api/entity/find/firstfreeid");
    return response.json();
  },

  // Next available reference number not held by any labeled record
  async nextReferenceNumber(): Promise<number> {
    const response = await apiFetch("/api/v2/records/nextid?labeled=true");
    return response.json();
  },

  async getSearchEmbeddingProgress(opts: {
    id?: number;
    global?: boolean;
    childrenDepth?: number;
    searchImage?: boolean;
    searchTextEmbedded?: boolean;
  }): Promise<{ indexed: number; pending: number; total: number; ready: boolean }> {
    const params = new URLSearchParams();
    if (opts.global) params.set("global", "true");
    else if (opts.id != null) params.set("id", String(opts.id));
    if (opts.childrenDepth != null) params.set("childrenDepth", String(opts.childrenDepth));
    if (opts.searchImage) params.set("searchImage", "true");
    if (opts.searchTextEmbedded) params.set("searchTextEmbedded", "true");
    const response = await apiFetch(`/api/v2/embeddings/search-progress?${params}`);
    return response.json();
  },

  async getStoreVersion(): Promise<number> {
    const response = await apiFetch("/api/store/version");
    return response.json();
  },
};

export default api;
