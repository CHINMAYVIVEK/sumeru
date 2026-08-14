/**
 * Home dashboard (/web/home): pin to top bar and pinned grid grouping.
 */
import { getPinnedApps, togglePinnedApp, applyTopNavFilter } from "../shell/pinned-apps.js";
import { showToast } from "../lib/toast.js";

function updatePinButton(btn, displayName, pinned) {
  const name = displayName || btn.getAttribute("data-module") || "App";
  btn.classList.toggle("is-pinned", pinned);
  btn.setAttribute("aria-pressed", pinned ? "true" : "false");
  const label = pinned ? `Unpin ${name} from top bar` : `Pin ${name} to top bar`;
  btn.setAttribute("aria-label", label);
  btn.setAttribute("title", pinned ? "Pinned to top bar — click to unpin" : "Pin to top bar");
}

function syncAllPinButtons() {
  const pins = getPinnedApps();
  document.querySelectorAll(".sum-home-hub-app-pin").forEach((btn) => {
    const mod = btn.getAttribute("data-module");
    updatePinButton(btn, btn.getAttribute("data-display-name"), pins.includes(mod));
  });
}

function tileDisplayName(tile) {
  const nameEl = tile.querySelector(".sum-home-hub-app-name");
  return (nameEl?.textContent || tile.getAttribute("data-module") || "").trim().toLowerCase();
}

function sortTilesAZ(tiles) {
  return [...tiles].sort((a, b) => {
    const displayA = tileDisplayName(a);
    const displayB = tileDisplayName(b);
    if (displayA !== displayB) return displayA.localeCompare(displayB);
    return (a.getAttribute("data-module") || "").localeCompare(b.getAttribute("data-module") || "");
  });
}

function organizePinnedGrid() {
  const pinnedSection = document.getElementById("sum-home-pinned-section");
  const pinnedContainer = document.getElementById("sum-home-pinned-apps");
  const allContainer = document.getElementById("sum-home-all-apps");
  if (!pinnedSection || !pinnedContainer || !allContainer) return;

  const pins = getPinnedApps();
  const allTiles = [
    ...pinnedContainer.querySelectorAll(".sum-home-hub-app"),
    ...allContainer.querySelectorAll(".sum-home-hub-app"),
  ];
  const tilesByModule = {};
  allTiles.forEach((tile) => {
    tilesByModule[tile.getAttribute("data-module")] = tile;
  });

  const pinnedTiles = pins.map((mod) => tilesByModule[mod]).filter(Boolean);
  sortTilesAZ(pinnedTiles).forEach((tile) => {
    pinnedContainer.appendChild(tile);
  });

  sortTilesAZ(allTiles.filter((tile) => !pins.includes(tile.getAttribute("data-module")))).forEach((tile) => {
    allContainer.appendChild(tile);
  });

  pinnedSection.hidden = pinnedContainer.children.length === 0;
}

export function initHomeDashboard() {
  document.addEventListener(
    "click",
    (e) => {
      const btn = e.target.closest(".sum-home-hub-app-pin");
      if (!btn) return;
      e.preventDefault();
      e.stopPropagation();
      const mod = btn.getAttribute("data-module");
      const displayName = btn.getAttribute("data-display-name") || mod;
      const pins = togglePinnedApp(mod);
      const pinned = pins.includes(mod);
      updatePinButton(btn, displayName, pinned);
      syncAllPinButtons();
      organizePinnedGrid();
      applyTopNavFilter();
      showToast(
        "sum-home-toast",
        pinned ? `${displayName} pinned to top bar` : `${displayName} unpinned from top bar`
      );
    },
    true
  );

  syncAllPinButtons();
  organizePinnedGrid();
}

initHomeDashboard();
