import { describe, expect, it } from "vitest";
import { collectFormFields } from "../../src/views/form/form-sheet.js";
import type { SwcViewArch } from "../../src/types/workspace.js";

describe("collectFormFields", () => {
  it("collects div, nested group, and notebook fields", () => {
    const arch: SwcViewArch = {
      type: "form",
      model: "core.partner",
      fields: [],
      sheet: {
        divs: [{ class: "sum_title", h1Fields: [{ name: "name" }] }],
        fields: [],
        groups: [
          {
            fields: [],
            groups: [
              { string: "Contact", fields: [{ name: "email" }] },
              { string: "Address", fields: [{ name: "street" }] },
            ],
          },
        ],
        notebook: [
          {
            pages: [{ title: "Notes", fields: [{ name: "comment" }], groups: [] }],
          },
        ],
      },
    };
    const names = collectFormFields(arch.sheet).map((f) => f.name);
    expect(names).toEqual(["name", "email", "street", "comment"]);
  });
});
