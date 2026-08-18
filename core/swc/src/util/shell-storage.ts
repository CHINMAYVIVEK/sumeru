/** Persisted shell preferences (legacy-compatible localStorage keys). */

export const KEY_SIDEBAR = "sum.shell.sidebarCollapsed";
export const KEY_ACTIVITY_WIDTH = "sum.shell.activityWidthPx";
export const KEY_ACTIVITY_HIDDEN = "sum.shell.activityHidden";

export function readBool(key: string): boolean {
  try {
    return localStorage.getItem(key) === "1";
  } catch {
    return false;
  }
}

export function writeBool(key: string, value: boolean): void {
  try {
    localStorage.setItem(key, value ? "1" : "0");
  } catch {
    /* quota or private mode */
  }
}

export function readActivityWidth(): number {
  try {
    const n = parseInt(localStorage.getItem(KEY_ACTIVITY_WIDTH) ?? "", 10);
    if (n >= 200 && n <= 520) return n;
  } catch {
    /* ignore */
  }
  return 300;
}

export function writeActivityWidth(px: number): void {
  try {
    localStorage.setItem(KEY_ACTIVITY_WIDTH, String(Math.round(px)));
  } catch {
    /* ignore */
  }
}

export function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    const value = JSON.parse(raw) as unknown;
    return (Array.isArray(value) ? value : fallback) as T;
  } catch {
    return fallback;
  }
}

export function writeJSON(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    /* ignore */
  }
}
