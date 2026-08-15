/** Workspace toast notifications (top-right stack). */
let toastStackEl = null;

function ensureToastStack() {
  if (toastStackEl) return toastStackEl;
  toastStackEl = document.getElementById("sum-toast-stack");
  if (!toastStackEl) {
    toastStackEl = document.createElement("div");
    toastStackEl.id = "sum-toast-stack";
    toastStackEl.className = "sum-toast-stack";
    toastStackEl.setAttribute("aria-live", "polite");
    document.body.appendChild(toastStackEl);
  }
  return toastStackEl;
}

export function showWorkspaceToast({ kind = "info", title = "", body = "", durationMs = 4000 } = {}) {
  const stack = ensureToastStack();
  const toast = document.createElement("div");
  toast.className = `sum-toast sum-toast--${kind}`;
  toast.setAttribute("role", "status");

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.className = "sum-toast-close";
  closeBtn.setAttribute("aria-label", "Dismiss");
  closeBtn.textContent = "×";
  closeBtn.addEventListener("click", () => toast.remove());

  if (title) {
    const strong = document.createElement("strong");
    strong.className = "sum-toast-title";
    strong.textContent = title;
    toast.appendChild(strong);
  }
  if (body) {
    const p = document.createElement("p");
    p.className = "sum-toast-body";
    p.textContent = body;
    toast.appendChild(p);
  }
  toast.appendChild(closeBtn);
  stack.appendChild(toast);

  if (durationMs > 0) {
    setTimeout(() => toast.remove(), durationMs);
  }
}

export function showToast(elementId, message, durationMs = 2400) {
  const el = document.getElementById(elementId);
  if (!el) return;
  el.textContent = message;
  el.hidden = false;
  el.classList.add("is-visible");
  clearTimeout(showToast._timers?.[elementId]);
  if (!showToast._timers) showToast._timers = {};
  showToast._timers[elementId] = setTimeout(() => {
    el.classList.remove("is-visible");
    el.hidden = true;
  }, durationMs);
}

export function initWorkspaceToastsFromDOM() {
  const node = document.getElementById("sum-toast-data");
  if (!node) return;
  try {
    const items = JSON.parse(node.textContent || "[]");
    items.forEach((item) => {
      showWorkspaceToast({
        kind: item.Kind || item.kind || "success",
        title: item.Title || item.title || "",
        body: item.Body || item.body || "",
      });
    });
  } catch (_) {
    /* ignore malformed bootstrap payload */
  }
  node.remove();
}
