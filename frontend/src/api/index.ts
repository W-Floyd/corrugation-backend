import type { Entity, Artifact, FullState, Metadata } from "./types";

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

export const api = {
  /**
   * Fetch the complete store state from the backend
   */
  async getFullState(): Promise<FullState> {
    const response = await fetch("/api/store");
    return response.json();
  },

  /**
   * Create a new entity and return its ID
   */
  async createEntity(entity: Partial<Entity>): Promise<number> {
    const response = await fetch("/api/entity", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(entity),
    });
    const id = await response.json();
    return parseInt(id, 10);
  },

  /**
   * Replace an existing entity
   */
  async updateEntity(id: number, entity: Entity): Promise<void> {
    await fetch(`/api/entity/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(entity),
    });
  },

  /**
   * Patch an existing entity
   */
  async patchEntity(id: number, patch: Partial<Entity>): Promise<Entity> {
    const response = await fetch(`/api/entity/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(patch),
    });
    return response.json();
  },

  /**
   * Delete an entity
   */
  async deleteEntity(id: number): Promise<void> {
    await fetch(`/api/entity/${id}`, {
      method: "DELETE",
    });
  },

  /**
   * Move an entity to a new location
   */
  async moveEntity(id: number, location: number): Promise<void> {
    await fetch(`/api/entity/${id}`, {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ location }),
    });
  },

  /**
   * Upload an artifact and return its ID
   */
  async uploadArtifact(file: File): Promise<number> {
    const formData = new FormData();
    formData.append("file", file);

    const response = await fetch("/api/artifact", {
      method: "POST",
      body: formData,
    });
    const id = await response.json();
    return parseInt(id, 10);
  },

  /**
   * Delete an artifact
   */
  async deleteArtifact(id: number): Promise<void> {
    await fetch(`/api/artifact/${id}`, {
      method: "DELETE",
    });
  },

  /**
   * Get the first free entity ID (gap in sequence)
   */
  async firstFreeId(): Promise<number> {
    const response = await fetch("/api/entity/find/firstfreeid");
    return response.json();
  },

  /**
   * Get the next available ID (first unlabeled or first free)
   */
  async firstAvailableId(): Promise<number> {
    const response = await fetch("/api/entity/find/firstid");
    return response.json();
  },

  /**
   * Quick capture: upload artifact and create entity in one flow
   */
  async quickCapture(location: number): Promise<number | null> {
    // This is called from the camera store's open callback
    // The camera store will have the pendingFile ready
    // Returns the created entity ID
    return 0;
  },
};

export default api;
