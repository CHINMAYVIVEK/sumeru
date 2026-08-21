import { describe, expect, it, vi } from "vitest";
import { initViewTabNavigation, syncWorkspaceViewTabs } from "../../src/shell/view-tab-sync.js";
import type { SwcViewTab } from "../../src/types/workspace.js";

function mountViewTabs(tabs: SwcViewTab[]): void {
  document.body.innerHTML = `
    <div class="sum-breadcrumb-right">
      <div class="sum-view-toolbar sum-view-toolbar--inline">
        ${tabs
          .map(
            (tab) =>
              `<a href="${tab.href}" class="sum-view-tab${tab.active ? " is-active" : ""}" data-view="${tab.mode}">${tab.label}</a>`,
          )
          .join("")}
      </div>
    </div>
  `;
}

describe("view-tab-sync", () => {
  it("updates active tab and href from workspace payload", () => {
    mountViewTabs([
      { mode: "list", label: "List", href: "/web?action=1&view_type=list", active: true },
      { mode: "form", label: "Form", href: "/web?action=1&view_type=form", active: false },
    ]);

    syncWorkspaceViewTabs([
      { mode: "list", label: "List", href: "/web?action=1&view_type=list", active: false },
      { mode: "form", label: "Form", href: "/web?action=1&view_type=form&id=7", active: true },
    ]);

    const listTab = document.querySelector('[data-view="list"]');
    const formTab = document.querySelector('[data-view="form"]');
    expect(listTab?.classList.contains("is-active")).toBe(false);
    expect(formTab?.classList.contains("is-active")).toBe(true);
    expect(formTab?.getAttribute("href")).toContain("id=7");
    expect(formTab?.getAttribute("aria-current")).toBe("page");
  });

  it("routes view tab clicks through popstate", () => {
    mountViewTabs([
      { mode: "list", label: "List", href: "/web?action=1&view_type=list", active: true },
      { mode: "form", label: "Form", href: "/web?action=1&view_type=form", active: false },
    ]);

    const onPop = vi.fn();
    window.addEventListener("popstate", onPop);
    initViewTabNavigation();

    const formTab = document.querySelector('[data-view="form"]') as HTMLAnchorElement;
    formTab.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));

    expect(window.location.pathname + window.location.search).toBe("/web?action=1&view_type=form");
    expect(onPop).toHaveBeenCalledTimes(1);

    window.removeEventListener("popstate", onPop);
  });
});
