import type { RpcService } from "../services/rpc.js";
import { SwcError } from "../core/error.js";

export class SwcRecord {
  readonly model: string;
  readonly id: number;
  data: Record<string, unknown>;
  private dirty = new Set<string>();

  constructor(model: string, id: number, data: Record<string, unknown>) {
    this.model = model;
    this.id = id;
    this.data = { ...data };
  }

  get(field: string): unknown {
    return this.data[field];
  }

  set(field: string, value: unknown): void {
    this.data[field] = value;
    this.dirty.add(field);
  }

  isDirty(): boolean {
    return this.dirty.size > 0;
  }

  dirtyValues(): Record<string, unknown> {
    const out: Record<string, unknown> = {};
    for (const k of this.dirty) {
      out[k] = this.data[k];
    }
    return out;
  }

  clearDirty(): void {
    this.dirty.clear();
  }
}

export class RecordStore {
  private readonly rpc: RpcService;

  constructor(rpc: RpcService) {
    this.rpc = rpc;
  }

  fromPayload(model: string, id: number, data: Record<string, unknown>): SwcRecord {
    return new SwcRecord(model, id, data);
  }

  async save(rec: SwcRecord): Promise<number> {
    if (rec.id <= 0) {
      const newId = await this.rpc.create(rec.model, rec.data);
      rec.clearDirty();
      return newId;
    }
    if (!rec.isDirty()) return rec.id;
    await this.rpc.write(rec.model, [rec.id], rec.dirtyValues());
    rec.clearDirty();
    return rec.id;
  }

  async unlink(rec: SwcRecord): Promise<void> {
    if (rec.id <= 0) return;
    await this.rpc.unlink(rec.model, [rec.id]);
  }

  validate(rec: SwcRecord, requiredFields: string[]): void {
    for (const f of requiredFields) {
      const v = rec.get(f);
      if (v == null || v === "") {
        throw new SwcError(`Field ${f} is required`, "validation");
      }
    }
  }
}
