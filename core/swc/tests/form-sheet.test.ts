import { describe, expect, it } from "vitest";
import { renderFormSheet } from "../src/views/form-sheet.js";
import { SwcRecord } from "../src/store/record.js";
import { registerDefaultWidgets } from "../src/widgets/registry.js";
import type { SwcArchSheet } from "../src/types/workspace.js";

registerDefaultWidgets();

function partnerSheet(): SwcArchSheet {
  return {
    divs: [{ class: "sum_title", h1Fields: [{ name: "name", string: "Name", placeholder: "Contact or Company Name..." }] }],
    fields: [],
    groups: [
      {
        fields: [],
        groups: [
          { string: "Contact", fields: [{ name: "email", string: "Email", type: "char" }] },
          { string: "Address", fields: [{ name: "street", string: "Street", type: "char" }] },
        ],
      },
    ],
    notebook: [
      {
        pages: [{ title: "Notes", fields: [{ name: "comment", string: "Notes", type: "text" }], groups: [] }],
      },
    ],
  };
}

describe("form-sheet", () => {
  it("renders title, groups, notebook, and split avatar layout", () => {
    const record = new SwcRecord("core.partner", 1, { name: "Acme", email: "a@x.com" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: partnerSheet(),
      record,
      readonly: true,
      hasImageField: true,
      activeNotebookPages: { 0: 0 },
      onNotebookTab: () => {},
    }).render();

    expect(el.matches(".sum-form-sheet") || el.querySelector(".sum-form-sheet")).toBeTruthy();
    expect(el.querySelector(".sum-form-split-layout")).toBeTruthy();
    expect(el.querySelector(".sum-form-avatar-initials")?.textContent).toBe("AC");
    expect(el.textContent).toContain("Acme");
    expect(el.textContent).toContain("Contact");
    expect(el.textContent).toContain("Notes");
    expect(el.textContent).toContain("a@x.com");
    expect(el.querySelector(".sum-form-group--full .sum-form-group-title")?.textContent).toBe("Contact");
    const tab = el.querySelector("button[role=tab]");
    expect(tab?.textContent?.trim()).toBe("Notes");
    expect(tab?.className).toContain("sum-notebook-tab--active");
  });

  it("renders contact row in sum_title", () => {
    const record = new SwcRecord("core.user", 1, { name: "Mitchell", email: "m@x.com", phone: "555" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [
          {
            class: "sum_title",
            h1Fields: [{ name: "name", placeholder: "Name" }],
            divs: [
              {
                class: "sum-title-contact-row",
                fields: [
                  { name: "email", string: "Email", widget: "email" },
                  { name: "phone", string: "Work phone" },
                ],
              },
            ],
          },
        ],
        fields: [],
        groups: [],
      },
      record,
      readonly: true,
      hasImageField: true,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    expect(el.querySelector(".sum-title-contact-row")).toBeTruthy();
    expect(el.textContent).toContain("m@x.com");
    expect(el.textContent).toContain("555");
  });

  it("renders separator and label on sheet", () => {
    const record = new SwcRecord("my.module", 1, { name: "Demo" });
    const el = renderFormSheet({
      env: { bootstrap: {} as never, services: {} as never },
      sheet: {
        divs: [],
        fields: [],
        groups: [],
        separators: [{ string: "Label example" }],
        labels: [{ for: "email", string: "Email is rendered above" }],
      },
      record,
      readonly: true,
      hasImageField: false,
      activeNotebookPages: {},
      onNotebookTab: () => {},
    }).render();

    expect(el.querySelector(".sum-separator--title")?.textContent).toBe("Label example");
    expect(el.querySelector(".sum-label--notebook")?.textContent).toBe("Email is rendered above");
  });
});
