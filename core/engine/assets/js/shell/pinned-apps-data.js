/** Parses and caches pinned apps embedded in the shell (#sum-pinned-apps-data). */

import { readEmbeddedJSON } from "../lib/dom-json.js";

let pinnedCache = null;
let cacheLoaded = false;

export function loadPinnedAppsFromShell() {
  if (cacheLoaded) return pinnedCache ?? [];
  cacheLoaded = true;
  pinnedCache = readEmbeddedJSON("sum-pinned-apps-data", { expectStringArray: true });
  return pinnedCache;
}

export function getPinnedCache() {
  if (!cacheLoaded) loadPinnedAppsFromShell();
  return pinnedCache ?? [];
}

export function setPinnedCache(modules) {
  pinnedCache = Array.isArray(modules) ? modules.map(String) : [];
  cacheLoaded = true;
}
