import type { Entity, Artifact, FullState, Metadata, BackendRecord } from "./types";
import { recordToEntity } from "./types";
import router from "../router";
import { useAuthStore } from "../stores/auth";
import { useToastsStore } from "../stores/toasts";

export interface EntityCreate {
  id?: number;
  name: string | null;
  description: string | null;
  artifacts: number[] | null;
  location: number;
  metadata: {
    quantity: number | null;
    owners: string[] | null;
    tags: string[] | null;
    islabeled: boolean | null;
  };
}

export interface EntityUpdate {
  name?: string | null;
  description?: string | null;
  artifacts?: number[] | null;
  location?: number;
  metadata?: {
    quantity?: number | null;
    owners?: string[] | null;
    tags?: string[] | null;
    islabeled?: boolean | null;
  };
}

export async function apiFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const token = localStorage.getItem("auth_token");
  const headers = new Headers(options.headers as HeadersInit);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(url, { ...options, headers });

  if (response.status === 401) {
    const currentRoute = router.currentRoute.value.name;
    DEBUG && console.warn("[apiFetch] 401 on", url, "current route:", currentRoute, new Error().stack?.split("\n")[2]?.trim());
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

export async function withErrorToast<T>(fn: () => Promise<T>): Promise<T | undefined> {
  try {
    return await fn();
  } catch (e) {
    useToastsStore().add(e instanceof Error ? e.message : String(e));
    return undefined;
  }
}

export const api = {
  async getFullState(): Promise<FullState> {
    const response = await apiFetch("/api/store");
    return response.json();
  },

  async createEntity(entity: Partial<Entity>): Promise<number> {
    const response = await apiFetch("/api/entity", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(entity),
    });
    const id = await response.json();
    return parseInt(id, 10);
  },

  async updateEntity(id: number, entity: Entity): Promise<void> {
    await apiFetch(`/api/entity/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(entity),
    });
  },

  async patchEntity(id: number, patch: Partial<Entity>): Promise<Entity> {
    const response = await apiFetch(`/api/entity/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(patch),
    });
    return response.json();
  },

  async deleteEntity(id: number): Promise<void> {
    await apiFetch(`/api/entity/${id}`, { method: "DELETE" });
  },

  async moveEntity(id: number, location: number): Promise<void> {
    await apiFetch(`/api/entity/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ location }),
    });
  },

  async uploadArtifact(file: File): Promise<number> {
    const formData = new FormData();
    formData.append("file", file);
    const response = await apiFetch("/api/artifact", { method: "POST", body: formData });
    const id = await response.json();
    return parseInt(id, 10);
  },

  async deleteArtifact(id: number): Promise<void> {
    await apiFetch(`/api/artifact/${id}`, { method: "DELETE" });
  },

  async searchEntities(query: string, parentId?: number, searchDescription = true, searchLabel = true): Promise<{ entity: Entity; imageScore?: number; textScore?: number }[]> {
    const params = new URLSearchParams({ search: query });
    if (parentId != null) {
      params.set("id", String(parentId));
      params.set("childrenDepth", "-1");
    }
    if (searchDescription) params.set("searchDescription", "true");
    if (searchLabel) params.set("searchLabel", "true");
    const response = await apiFetch(`/api/v2/records?${params}`);
    const records: BackendRecord[] = await response.json();
    return records.map((r) => ({
      entity: recordToEntity(r),
      imageScore: r.SearchConfidenceImage,
      textScore: r.SearchConfidenceText,
    }));
  },

  async firstFreeId(): Promise<number> {
    const response = await apiFetch("/api/entity/find/firstfreeid");
    return response.json();
  },

  async firstAvailableId(): Promise<number> {
    const response = await apiFetch("/api/entity/find/firstid");
    return response.json();
  },

  async quickCapture(_location: number): Promise<number | null> {
    return 0;
  },
};

export default api;
