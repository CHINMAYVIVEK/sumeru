import { describe, it, expect } from "vitest";
import { RecordStore, SwcRecord } from "../../src/model/record.ts";

describe("RecordStore", () => {
  it("tracks dirty fields", () => {
    const rec = new SwcRecord("test.model", 1, { name: "A" });
    rec.set("name", "B");
    expect(rec.isDirty()).toBe(true);
    expect(rec.dirtyValues()).toEqual({ name: "B" });
  });
});

describe("RecordStore validate", () => {
  it("throws on missing required field", () => {
    const store = new RecordStore({} as never);
    const rec = store.fromPayload("m", 0, {});
    expect(() => store.validate(rec, ["name"])).toThrow();
  });
});
