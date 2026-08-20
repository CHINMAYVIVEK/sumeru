import type { SwcToastMessage } from "../types/bootstrap.js";

const MAX_TOASTS = 5;
const EXIT_MS = 250;

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

  success(title: string, body: string, details?: string): void {
    this.show({ kind: "success", title, body, details });
  }

  error(title: string, body: string, details?: string): void {
    this.show({ kind: "error", title, body, details });
  }

  warning(title: string, body: string, details?: string): void {
    this.show({ kind: "warning", title, body, details });
  }

  show(msg: SwcToastMessage, timeoutMs = 6000): void {
    this.capStack();
    const toast = this.buildToast(msg);
    this.stack.appendChild(toast);
    this.armTimer(toast, timeoutMs);
  }

  bootstrap(messages: SwcToastMessage[] | undefined): void {
    for (const m of messages ?? []) {
      this.show(m);
    }
  }

  private liveToasts(): HTMLElement[] {
    return [...this.stack.children].filter(
      (el) => el instanceof HTMLElement && !el.classList.contains("sum-toast-out"),
    ) as HTMLElement[];
  }

  private capStack(): void {
    const live = this.liveToasts();
    while (live.length >= MAX_TOASTS) {
      const oldest = live.shift();
      if (oldest) this.dismiss(oldest);
    }
  }

  private buildToast(msg: SwcToastMessage): HTMLElement {
    const toast = document.createElement("div");
    toast.className = `sum-toast sum-toast--${msg.kind || "info"}`;
    toast.setAttribute("role", "status");

    const title = document.createElement("span");
    title.className = "sum-toast-title";
    title.textContent = msg.title;

    const body = document.createElement("p");
    body.className = "sum-toast-body";
    body.textContent = msg.body;

    toast.append(title, body);

    if (msg.details) {
      const details = document.createElement("pre");
      details.className = "sum-toast-details";
      details.textContent = msg.details;
      toast.append(details);
    }

    const close = document.createElement("button");
    close.type = "button";
    close.className = "sum-toast-close";
    close.setAttribute("aria-label", "Close");
    close.textContent = "×";
    close.addEventListener("click", () => this.dismiss(toast));
    toast.append(close);
    return toast;
  }

  private armTimer(toast: HTMLElement, timeoutMs: number): void {
    let remaining = timeoutMs;
    let started = Date.now();
    let timer = window.setTimeout(() => this.dismiss(toast), remaining);
    toast.addEventListener("mouseenter", () => {
      window.clearTimeout(timer);
      remaining -= Date.now() - started;
    });
    toast.addEventListener("mouseleave", () => {
      started = Date.now();
      timer = window.setTimeout(() => this.dismiss(toast), Math.max(0, remaining));
    });
  }

  dismiss(toast: HTMLElement): void {
    if (!toast.isConnected || toast.classList.contains("sum-toast-out")) return;
    toast.classList.add("sum-toast-out");
    let removed = false;
    const remove = () => {
      if (removed) return;
      removed = true;
      toast.remove();
    };
    toast.addEventListener("animationend", remove, { once: true });
    window.setTimeout(remove, EXIT_MS);
  }
}
