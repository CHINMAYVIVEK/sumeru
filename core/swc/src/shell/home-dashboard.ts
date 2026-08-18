import type { HttpService } from "../services/http.js";
import {
  applyTopNavFilter,
  getPinnedApps,
  togglePinnedApp,
} from "./pinned-apps.js";

function updatePinButton(btn: HTMLElement, displayName: string, pinned: boolean): void {
  const name = displayName || btn.getAttribute("data-module") || "App";
  btn.classList.toggle("is-pinned", pinned);
  btn.setAttribute("aria-pressed", pinned ? "true" : "false");
  const label = pinned ? `Unpin ${name} from top bar` : `Pin ${name} to top bar`;
  btn.setAttribute("aria-label", label);
  btn.setAttribute("title", pinned ? "Pinned to top bar — click to unpin" : "Pin to top bar");
}

function syncAllPinButtons(): void {
  const pins = getPinnedApps();
  document.querySelectorAll<HTMLElement>(".sum-home-hub-app-pin").forEach((btn) => {
    const mod = btn.getAttribute("data-module") ?? "";
    updatePinButton(btn, btn.getAttribute("data-display-name") ?? mod, pins.includes(mod));
  });
}

function tileDisplayName(tile: Element): string {
  const nameEl = tile.querySelector(".sum-home-hub-app-name");
  return (nameEl?.textContent ?? tile.getAttribute("data-module") ?? "").trim().toLowerCase();
}

function sortTilesAZ(tiles: Element[]): Element[] {
  return [...tiles].sort((a, b) => {
    const displayA = tileDisplayName(a);
    const displayB = tileDisplayName(b);
    if (displayA !== displayB) return displayA.localeCompare(displayB);
    return (a.getAttribute("data-module") ?? "").localeCompare(b.getAttribute("data-module") ?? "");
  });
}

export function organizePinnedGrid(): void {
  const pinnedSection = document.getElementById("sum-home-pinned-section");
  const pinnedContainer = document.getElementById("sum-home-pinned-apps");
  const allContainer = document.getElementById("sum-home-all-apps");
  if (!pinnedSection || !pinnedContainer || !allContainer) return;

  const pins = getPinnedApps();
  const allTiles = [
    ...pinnedContainer.querySelectorAll(".sum-home-hub-app"),
    ...allContainer.querySelectorAll(".sum-home-hub-app"),
  ];
  const tilesByModule: Record<string, Element> = {};
  allTiles.forEach((tile) => {
    const mod = tile.getAttribute("data-module");
    if (mod) tilesByModule[mod] = tile;
  });

  const pinnedTiles = pins.map((mod) => tilesByModule[mod]).filter(Boolean);
  sortTilesAZ(pinnedTiles).forEach((tile) => {
    pinnedContainer.appendChild(tile);
  });

  sortTilesAZ(allTiles.filter((tile) => !pins.includes(tile.getAttribute("data-module") ?? ""))).forEach(
    (tile) => {
      allContainer.appendChild(tile);
    },
  );

  pinnedSection.hidden = pinnedContainer.children.length === 0;
}

function showHomeToast(message: string): void {
  const toast = document.getElementById("sum-home-toast");
  if (!toast) return;
  toast.textContent = message;
  toast.hidden = false;
  window.setTimeout(() => {
    toast.hidden = true;
  }, 3200);
}

/** Pin/unpin controls on /web/home (grid layout only). */
export function initHomeDashboard(http: HttpService): void {
  if (!document.getElementById("sum-home-hub")) return;

  document.addEventListener(
    "click",
    (ev) => {
      const btn = (ev.target as Element | null)?.closest<HTMLElement>(".sum-home-hub-app-pin");
      if (!btn) return;
      ev.preventDefault();
      ev.stopPropagation();
      const mod = btn.getAttribute("data-module") ?? "";
      const displayName = btn.getAttribute("data-display-name") || mod;
      const pins = togglePinnedApp(http, mod);
      const pinned = pins.includes(mod);
      updatePinButton(btn, displayName, pinned);
      syncAllPinButtons();
      organizePinnedGrid();
      applyTopNavFilter();
      showHomeToast(
        pinned ? `${displayName} pinned to top bar` : `${displayName} unpinned from top bar`,
      );
    },
    true,
  );

  syncAllPinButtons();
  organizePinnedGrid();
}
