/** Snapshot shape read from window.__SWC_DEVTOOLS__ in the inspected page. */

export interface ComponentSnapshot {
  id: number;
  name: string;
}

export interface SwcDevtoolsSnapshot {
  components: ComponentSnapshot[];
}

export function parseDevtoolsSnapshot(raw: unknown): SwcDevtoolsSnapshot {
  if (typeof raw !== "string" || !raw) {
    return { components: [] };
  }
  try {
    const parsed = JSON.parse(raw) as { components?: ComponentSnapshot[] };
    return { components: parsed.components ?? [] };
  } catch {
    return { components: [] };
  }
}

/** Expression evaluated inside the inspected SWC page. */
export const DEVTOOLS_SNAPSHOT_EXPR = `window.__SWC_DEVTOOLS__
  ? JSON.stringify({
      components: window.__SWC_DEVTOOLS__.components.map((c) => ({
        id: c.id,
        name: c.name,
      })),
    })
  : ""`;
