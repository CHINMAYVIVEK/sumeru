/** Pinned/recent apps; pins persisted server-side per user (see render/pinned_apps.go). */

import { readJSON, writeJSON } from "./storage.js";
import { getPinnedCache, loadPinnedAppsFromShell, setPinnedCache } from "./pinned-apps-data.js";
import { postJSON } from "../lib/fetch-json.js";

const KEY_PINNED_LEGACY = "sumeru:pinned-apps";
const KEY_RECENT = "sumeru:recent-apps";
const MAX_RECENT = 5;
const MAX_TOPBAR_RECENT = 4;

async function persistPinnedApps(modules) {
  const data = await postJSON("/web/user/pinned-apps", { modules });
  return Array.isArray(data.modules) ? data.modules.map(String) : modules;
}

export function getPinnedApps() {
  return getPinnedCache().slice();
}

export function getRecentApps() {
  return readJSON(KEY_RECENT, []);
}

export function togglePinnedApp(moduleName) {
  const mod = String(moduleName || "").trim();
  if (!mod) return getPinnedApps();

  const previous = getPinnedCache().slice();
  let pins = previous.slice();
  if (pins.includes(mod)) {
    pins = pins.filter((m) => m !== mod);
  } else {
    pins = [mod, ...pins];
  }
  setPinnedCache(pins);

  // Optimistic update; revert cache if server save fails.
  persistPinnedApps(pins)
    .then((saved) => {
      setPinnedCache(saved);
    })
    .catch(() => {
      setPinnedCache(previous);
    });

  return pins;
}

export function pushRecentApp(moduleName) {
  const mod = String(moduleName || "").trim();
  if (!mod) return;
  let recents = getRecentApps().filter((m) => m !== mod);
  recents.unshift(mod);
  if (recents.length > MAX_RECENT) recents = recents.slice(0, MAX_RECENT);
  writeJSON(KEY_RECENT, recents);
}

export function applyTopNavFilter() {
  const nav = document.querySelector(".sum-top-nav");
  if (!nav) return;

  const moduleItems = [...nav.querySelectorAll(".top-menu-item--module")];
  if (moduleItems.length === 0) return;

  const pins = getPinnedApps();
  const recents = getRecentApps();
  const active = nav.querySelector(".top-menu-item--module.active");
  const activeMod = active ? active.getAttribute("data-module") : "";

  let visibleMods = new Set();
  if (pins.length > 0) {
    pins.forEach((m) => visibleMods.add(m));
  } else if (recents.length > 0) {
    recents.slice(0, MAX_TOPBAR_RECENT).forEach((m) => visibleMods.add(m));
  } else {
    moduleItems.forEach((el) => visibleMods.add(el.getAttribute("data-module") || ""));
  }
  if (activeMod) visibleMods.add(activeMod);

  const shouldFilter = pins.length > 0 || recents.length > 0;
  if (!shouldFilter) return;

  moduleItems.forEach((el) => {
    const mod = el.getAttribute("data-module") || "";
    el.classList.toggle("is-topbar-hidden", !visibleMods.has(mod));
  });

  const activeEl = nav.querySelector(".top-menu-item.active");
  if (activeEl && typeof activeEl.scrollIntoView === "function") {
    activeEl.scrollIntoView({ inline: "nearest", block: "nearest", behavior: "instant" });
  }
}

export function initRecentTracking() {
  document.addEventListener(
    "click",
    (e) => {
      const link = e.target.closest("[data-module]");
      if (!link) return;
      const mod = link.getAttribute("data-module");
      if (mod) pushRecentApp(mod);
    },
    true
  );
}

export function initPinnedApps() {
  loadPinnedAppsFromShell();
  // One-time migration from browser localStorage to server-side pins.
  const legacy = readJSON(KEY_PINNED_LEGACY, []);
  if (getPinnedCache().length === 0 && legacy.length > 0) {
    persistPinnedApps(legacy)
      .then((saved) => {
        setPinnedCache(saved);
        try {
          localStorage.removeItem(KEY_PINNED_LEGACY);
        } catch (_) {}
        applyTopNavFilter();
      })
      .catch(() => {});
  }
}
