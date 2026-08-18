import type { SwcViewTab } from "../types/workspace.js";

/** Keep server-rendered breadcrumb view tabs in sync after SPA navigation. */
export function syncWorkspaceViewTabs(viewTabs: SwcViewTab[]): void {
  if (viewTabs.length === 0) return;

  const byMode = new Map(viewTabs.map((tab) => [tab.mode, tab]));
  const tabs = document.querySelectorAll<HTMLAnchorElement>(
    ".sum-breadcrumb-right .sum-view-tab[data-view]",
  );

  for (const el of tabs) {
    const tab = byMode.get(el.dataset.view ?? "");
    if (!tab) {
      el.classList.remove("is-active");
      el.removeAttribute("aria-current");
      continue;
    }
    el.href = tab.href;
    el.classList.toggle("is-active", tab.active);
    if (tab.active) el.setAttribute("aria-current", "page");
    else el.removeAttribute("aria-current");
  }
}

/** Route view tab clicks through the workspace SPA instead of full page loads. */
export function initViewTabNavigation(): void {
  document.addEventListener("click", (ev) => {
    const tab = (ev.target as Element).closest<HTMLAnchorElement>(
      ".sum-breadcrumb-right .sum-view-tab[href]",
    );
    if (!tab?.href.includes("/web?")) return;

    ev.preventDefault();
    const url = new URL(tab.href, window.location.origin);
    window.history.pushState({}, "", `${url.pathname}${url.search}`);
    window.dispatchEvent(new PopStateEvent("popstate"));
  });
}
