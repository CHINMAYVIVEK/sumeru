/** Persisted shell preferences (localStorage). */

export const KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
export const KEY_ACTIVITY_W = "sum.shell.activityWidthPx";
export const KEY_ACTIVITY_HIDDEN = "sum.shell.activityHidden";

export function readBool(key) {
  try {
    return localStorage.getItem(key) === "1";
  } catch {
    return false;
  }
}

export function writeBool(key, v) {
  try {
    localStorage.setItem(key, v ? "1" : "0");
  } catch (_) {}
}

export function readActivityWidth() {
  try {
    const n = parseInt(localStorage.getItem(KEY_ACTIVITY_W), 10);
    if (n >= 200 && n <= 520) return n;
  } catch (_) {}
  return 300;
}

export function writeActivityWidth(n) {
  try {
    localStorage.setItem(KEY_ACTIVITY_W, String(Math.round(n)));
  } catch (_) {}
}

export function readJSON(key, fallback) {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    const v = JSON.parse(raw);
    return Array.isArray(v) ? v : fallback;
  } catch (_) {
    return fallback;
  }
}

export function writeJSON(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch (_) {}
}
