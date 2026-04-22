import type { Entity, Artifact, FullState, Metadata } from "./types";
import router from "../router";

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
    localStorage.removeItem("auth_token");
    router.push({ name: "login" });
    throw new Error("Unauthorized");
  }

  return response;
}

export const api = {
  async login(username: string, password: string): Promise<string> {
    const response = await fetch("/api/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username, password }),
    });
    if (!response.ok) {
      throw new Error("Invalid credentials");
    }
    const data = await response.json();
    return data.token;
  },

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
