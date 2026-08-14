/** Parses `#sum-app-launcher-data` JSON embedded in the shell. */

import { readEmbeddedJSON } from "../lib/dom-json.js";

export function loadLauncherApps() {
  const apps = readEmbeddedJSON("sum-app-launcher-data", { fallback: [] });
  return Array.isArray(apps) ? apps : [];
}
