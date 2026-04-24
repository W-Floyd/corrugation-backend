export interface Metadata {
  quantity: number | null;
  owners: string[] | null;
  tags: string[] | null;
  lastModified: string | null;
  lastModifiedBy: string | null;
  islabeled: boolean | null;
}

export interface Entity {
  id: number;
  name: string | null;
  description: string | null;
  artifacts: number[] | null;
  location: number;
  metadata: Metadata;
}

export interface Artifact {
  artifactid: number;
  path: string;
  image: boolean;
}

export type EntityCreate = Omit<Entity, "id">;

export interface FullState {
  entities: Record<number, Entity>;
  artifacts: Record<number, Artifact>;
  storeversion: number;
}

export interface BackendArtifactRef {
  ID: number;
}

export interface BackendTag {
  Title: string;
  Color?: string;
}

export interface BackendRecord {
  ID: number;
  CreatedAt?: string;
  UpdatedAt?: string;
  Label?: string;
  Title?: string;
  Description?: string;
  Quantity?: number;
  Tags?: BackendTag[];
  Artifacts?: BackendArtifactRef[];
  ParentID?: number;
  LastModifiedBy?: string;
}

export function recordToEntity(r: BackendRecord): Entity {
  return {
    id: r.ID,
    name: r.Label ?? r.Title ?? null,
    description: r.Description ?? null,
    artifacts: r.Artifacts?.map((a) => a.ID) ?? null,
    location: r.ParentID ?? 0,
    metadata: {
      quantity: r.Quantity ?? null,
      owners: null,
      tags: r.Tags?.map((t) => t.Title) ?? null,
      islabeled: r.Label != null,
      lastModified: r.UpdatedAt ?? null,
      lastModifiedBy: r.LastModifiedBy ?? null,
    },
  };
}
