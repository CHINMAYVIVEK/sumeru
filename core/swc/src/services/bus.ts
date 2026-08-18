type BusHandler = (payload: unknown) => void;

/** Client event bus with optional WebSocket live updates from /web/swc/bus. */
export class BusService {
  private readonly handlers = new Map<string, Set<BusHandler>>();
  private ws: WebSocket | null = null;

  subscribe(channel: string, handler: BusHandler): () => void {
    if (!this.handlers.has(channel)) {
      this.handlers.set(channel, new Set());
    }
    this.handlers.get(channel)!.add(handler);
    return () => this.handlers.get(channel)?.delete(handler);
  }

  emit(channel: string, payload: unknown): void {
    for (const fn of this.handlers.get(channel) ?? []) {
      fn(payload);
    }
  }

  /** Connect to /web/swc/bus when bootstrap.busEnabled is true. */
  connect(url = "/web/swc/bus"): void {
    if (this.ws) return;
    try {
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      this.ws = new WebSocket(`${proto}//${window.location.host}${url}`);
      this.ws.addEventListener("message", (ev) => {
        try {
          const msg = JSON.parse(String(ev.data)) as { channel: string; payload: unknown };
          if (msg.channel) this.emit(msg.channel, msg.payload);
        } catch {
          /* ignore malformed */
        }
      });
      this.ws.addEventListener("close", () => {
        this.ws = null;
      });
    } catch {
      /* WebSocket unavailable — local-only bus */
    }
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}
