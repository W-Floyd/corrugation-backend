import { defineStore } from "pinia";
import { ref, onActivated } from "vue";
import type { FullState } from "@/api/types";
import { apiFetch } from "@/api";

interface EntityIdMap {
  [key: string]: string;
}

export const useClipStore = defineStore("clip", () => {
  const enabled = ref(false);
  const modelReady = ref(false);
  const modelLoading = ref(false);
  const encoded = ref(0);
  const total = ref(0);
  const results = ref<any[]>([]);
  const scores = ref<{ [key: number]: number }>({});
  const textMatchIds = ref<Set<number>>(new Set());
  const searching = ref(false);

  // Internal model references
  let _textModel: any = null;
  let _visionModel: any = null;
  let _processor: any = null;
  let _tokenizer: any = null;
  let _RawImage: any = null;
  let _embeddings = new Map<string, Float32Array>();
  let _artifactEntity = new Map<string, string>();

  // CLIP model lazy loading
  async function _loadModel(): Promise<void> {
    if (_textModel) return;

    modelLoading.value = true;
    try {
      // Dynamically load from CDN
      // eslint-disable-next-line @typescript-eslint/ban-ts-comment
      // @ts-ignore: CDN URL import intentional at runtime
      const transformersModule = await import(
        "https://cdn.jsdelivr.net/npm/@huggingface/transformers@3.0.1" as string
      );
      const {
        CLIPTextModelWithProjection,
        CLIPVisionModelWithProjection,
        AutoProcessor,
        AutoTokenizer,
        RawImage,
        env,
      } = transformersModule;

      env.allowLocalModels = false;
      _RawImage = RawImage;

      [_textModel, _visionModel, _processor, _tokenizer] = await Promise.all([
        CLIPTextModelWithProjection.from_pretrained(
          "Xenova/clip-vit-base-patch32",
        ),
        CLIPVisionModelWithProjection.from_pretrained(
          "Xenova/clip-vit-base-patch32",
        ),
        AutoProcessor.from_pretrained("Xenova/clip-vit-base-patch32"),
        AutoTokenizer.from_pretrained("Xenova/clip-vit-base-patch32"),
      ]);

      modelReady.value = true;
    } catch (error) {
      console.error("CLIP model loading error:", error);
      modelReady.value = false;
    } finally {
      modelLoading.value = false;
    }
  }

  function _cacheKey(id: string): string {
    return `clip:${id}`;
  }

  function _toBase64(arr: Float32Array): string {
    return btoa(String.fromCharCode(...new Uint8Array(arr.buffer)));
  }

  function _fromBase64(b64: string): Float32Array {
    const bin = atob(b64);
    const buf = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) {
      buf[i] = bin.charCodeAt(i);
    }
    return new Float32Array(buf.buffer);
  }

  async function _encodeOne(id: string): Promise<void> {
    const cached = localStorage.getItem(_cacheKey(id));
    if (cached) {
      _embeddings.set(id, _fromBase64(cached));
      encoded.value++;
      return;
    }

    try {
      // Load image from artifact endpoint
      const resp = await apiFetch(`/api/artifact/${id}`);
      const blob = await resp.blob();
      const blobUrl = URL.createObjectURL(blob);
      const image = await _RawImage.fromURL(blobUrl);
      URL.revokeObjectURL(blobUrl);
      const inputs = await _processor(image);
      const { image_embeds } = await _visionModel(inputs);
      const emb = _normalize(image_embeds.data);
      _embeddings.set(id, emb);
      localStorage.setItem(_cacheKey(id), _toBase64(emb));
      encoded.value++;
    } catch (error) {
      console.error(`Failed to encode artifact ${id}:`, error);
    }
  }

  function _indexArtifacts(fullstate: FullState): void {
    _artifactEntity.clear();
    for (const [eid, entity] of Object.entries(fullstate.entities)) {
      if (!entity.artifacts || !entity.artifacts.length) continue;
      for (const aid of entity.artifacts) {
        if (fullstate.artifacts[aid]?.image) {
          _artifactEntity.set(String(aid), String(eid));
        }
      }
    }
  }

  async function _encodeAll(
    fullstate: FullState,
    currentEntityId: number,
  ): Promise<void> {
    _indexArtifacts(fullstate);
    const descendants = new Set(
      _listChildLocationsDeep(fullstate, currentEntityId),
    );
    const ids = [..._artifactEntity.keys()].filter(
      (id) =>
        !_embeddings.has(id) &&
        descendants.has(parseInt(_artifactEntity.get(id) || "", 10)),
    );

    if (!ids.length) return;

    total.value += ids.length;

    const concurrency = Math.max(
      2,
      Math.min(navigator.hardwareConcurrency ?? 4, 8),
    );

    const queue = [...ids];

    await Promise.all(
      Array.from({ length: concurrency }, async () => {
        while (queue.length) {
          const id = queue.shift()!;
          try {
            await _encodeOne(id);
          } catch (error) {
            total.value--;
          }
        }
      }),
    );
  }

  async function activate(
    fullstate: FullState,
    currentEntityId: number,
  ): Promise<void> {
    await _loadModel();
    await _encodeAll(fullstate, currentEntityId);
  }

  function _normalize(data: number[]): Float32Array {
    let n = 0;
    for (const v of data) {
      n += v * v;
    }
    n = Math.sqrt(n);
    const out = new Float32Array(data.length);
    for (let i = 0; i < data.length; i++) {
      out[i] = data[i]! / n;
    }
    return out;
  }

  function _dot(a: Float32Array, b: Float32Array): number {
    let s = 0;
    for (let i = 0; i < a.length; i++) {
      s += a[i]! * b[i]!;
    }
    return s;
  }

  async function search(
    query: string,
    fullstate: FullState,
    currentEntityId: number,
  ): Promise<void> {
    if (!query.trim()) {
      results.value = [];
      scores.value = {};
      searching.value = false;
      return;
    }

    searching.value = true;

    if (!modelReady.value) {
      await activate(fullstate, currentEntityId);
    } else {
      await _encodeAll(fullstate, currentEntityId);
    }

    const inputs = _tokenizer!(query, {
      padding: true,
      truncation: true,
    });

    const { text_embeds } = await _textModel!(inputs);
    const tv = _normalize(text_embeds.data);

    const descendants = new Set(
      _listChildLocationsDeep(fullstate, currentEntityId),
    );
    const best = new Map<number, number>();

    for (const [aid, iv] of _embeddings) {
      const eid = _artifactEntity.get(aid);
      if (!eid || !descendants.has(parseInt(eid, 10))) continue;
      const score = _dot(tv, iv);
      if (
        !best.has(parseInt(eid, 10)) ||
        best.get(parseInt(eid, 10))! < score
      ) {
        best.set(parseInt(eid, 10), score);
      }
    }

    const sorted = [...best.entries()].sort((a, b) => b[1] - a[1]);
    scores.value = Object.fromEntries(sorted);
    results.value = sorted
      .map(([eid]) => fullstate.entities[eid])
      .filter(Boolean);

    searching.value = false;
  }

  function merge(textResults: any[], entitiesStore: any): any[] {
    textMatchIds.value = new Set(textResults.map((e: any) => e.id));

    if (!enabled.value || !results.value.length) {
      return textResults;
    }

    const byClip = (a: any, b: any) =>
      (scores.value[b.id] ?? 0) - (scores.value[a.id] ?? 0);

    const textOnly = textResults.filter((e: any) => scores.value[e.id] == null);
    const textAndClip = textResults
      .filter((e: any) => scores.value[e.id] != null)
      .sort(byClip);

    const clipOnly = results.value
      .filter((e: any) => !textMatchIds.value.has(e.id))
      .sort(byClip);

    return [...textOnly, ...textAndClip, ...clipOnly];
  }

  // Helper to list child locations deep (similar to entities store)
  function _listChildLocationsDeep(
    fullstate: FullState,
    entityId: number,
  ): number[] {
    const returnValue: number[] = [];
    for (const key in fullstate.entities) {
      if (fullstate.entities[key]?.location === entityId) {
        returnValue.push(parseInt(key, 10));
        returnValue.push(
          ..._listChildLocationsDeep(fullstate, parseInt(key, 10)),
        );
      }
    }
    return returnValue;
  }

  return {
    enabled,
    modelReady,
    modelLoading,
    encoded,
    total,
    results,
    scores,
    textMatchIds,
    searching,
    activate,
    search,
    merge,
    _indexArtifacts,
  };
});
