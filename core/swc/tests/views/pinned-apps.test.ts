import { beforeEach, describe, expect, it, vi } from "vitest";
import { HttpService } from "../../src/services/http.js";
import {
  applyTopNavFilter,
  getPinnedApps,
  resetPinnedAppsState,
  setPinnedCache,
  togglePinnedApp,
} from "../../src/shell/pinned-apps.js";

function buildTopNav(modules: string[], activeMod = ""): HTMLElement {
  document.body.innerHTML = `
    <nav class="sum-top-nav">
      ${modules
        .map(
          (mod) =>
            `<a class="top-menu-item top-menu-item--module${mod === activeMod ? " active" : ""}" data-module="${mod}">${mod}</a>`,
        )
        .join("")}
    </nav>
  `;
  return document.body.querySelector(".sum-top-nav") as HTMLElement;
}

describe("pinned-apps", () => {
  beforeEach(() => {
    resetPinnedAppsState([]);
    localStorage.clear();
    document.body.innerHTML = "";
  });

  it("applyTopNavFilter hides modules not in pin list", () => {
    buildTopNav(["crm", "sale", "stock"]);
    setPinnedCache(["crm", "sale"]);
    applyTopNavFilter();
    const hidden = [...document.querySelectorAll(".top-menu-item--module.is-topbar-hidden")].map(
      (el) => el.getAttribute("data-module"),
    );
    expect(hidden).toEqual(["stock"]);
  });

  it("applyTopNavFilter hides all modules when no pins are set", () => {
    buildTopNav(["crm", "sale", "stock"]);
    setPinnedCache([]);
    applyTopNavFilter();
    const hidden = [...document.querySelectorAll(".top-menu-item--module.is-topbar-hidden")].map(
      (el) => el.getAttribute("data-module"),
    );
    expect(hidden.sort()).toEqual(["crm", "sale", "stock"]);
  });

  it("applyTopNavFilter hides active module when not pinned", () => {
    buildTopNav(["crm", "sale", "stock"], "stock");
    setPinnedCache(["crm"]);
    applyTopNavFilter();
    const stock = document.querySelector('.top-menu-item--module[data-module="stock"]');
    expect(stock?.classList.contains("is-topbar-hidden")).toBe(true);
    const crm = document.querySelector('.top-menu-item--module[data-module="crm"]');
    expect(crm?.classList.contains("is-topbar-hidden")).toBe(false);
    const sale = document.querySelector('.top-menu-item--module[data-module="sale"]');
    expect(sale?.classList.contains("is-topbar-hidden")).toBe(true);
  });

  it("togglePinnedApp reverts cache when save fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: async () => ({}),
      }),
    );
    resetPinnedAppsState([]);
    const http = new HttpService("test-csrf");
    togglePinnedApp(http, "crm");
    expect(getPinnedApps()).toEqual(["crm"]);
    await new Promise((r) => setTimeout(r, 0));
    expect(getPinnedApps()).toEqual([]);
    vi.unstubAllGlobals();
  });
});
