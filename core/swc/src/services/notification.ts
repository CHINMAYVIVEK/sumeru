import type { SwcToastMessage } from "../types/bootstrap.js";

export class NotificationService {
  private readonly stack: HTMLElement;

  constructor(stackEl?: HTMLElement | null) {
    this.stack = stackEl ?? document.getElementById("sum-toast-stack") ?? this.createStack();
  }

  private createStack(): HTMLElement {
    const el = document.createElement("div");
    el.id = "sum-toast-stack";
    el.className = "sum-toast-stack";
    el.setAttribute("aria-live", "polite");
    document.body.appendChild(el);
    return el;
  }

  show(msg: SwcToastMessage, timeoutMs = 6000): void {
    const toast = document.createElement("div");
    toast.className = `sum-toast sum-toast--${msg.kind || "info"}`;
    toast.innerHTML = `<strong>${escape(msg.title)}</strong><span>${escape(msg.body)}</span>`;
    this.stack.appendChild(toast);
    window.setTimeout(() => toast.remove(), timeoutMs);
  }

  bootstrap(messages: SwcToastMessage[] | undefined): void {
    for (const m of messages ?? []) {
      this.show(m);
    }
  }
}

function escape(s: string): string {
  return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
