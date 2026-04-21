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

export type EntityCreate = Omit<Entity, 'id'>;

export interface FullState {
  entities: Record<number, Entity>;
  artifacts: Record<number, Artifact>;
  storeversion: number;
}
