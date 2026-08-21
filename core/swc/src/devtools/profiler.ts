export interface RenderEvent {
  ts: number;
  kind: "render" | "patch" | "pick";
  component: string;
  durationMs?: number;
}

const events: RenderEvent[] = [];
const MAX = 500;

export function logRenderEvent(kind: RenderEvent["kind"], component: string, durationMs?: number): void {
  events.push({ ts: Date.now(), kind, component, durationMs });
  if (events.length > MAX) events.shift();
}

export function getRenderEvents(): readonly RenderEvent[] {
  return events;
}

export function clearRenderEvents(): void {
  events.length = 0;
}
