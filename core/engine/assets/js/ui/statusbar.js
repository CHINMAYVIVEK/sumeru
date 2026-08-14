/**
 * Clickable form statusbar (e.g. CRM stage_id) — persists via POST /web/kanban/move.
 */
export function initStatusbar() {
  document.querySelectorAll("[data-sum-statusbar]").forEach((bar) => {
    const model = bar.dataset.model || "";
    const recordId = parseInt(bar.dataset.recordId || "0", 10);
    const field = bar.dataset.field || "";
    const csrf =
      bar.dataset.csrf ||
      document.querySelector('meta[name="csrf-token"]')?.getAttribute("content") ||
      document.querySelector('input[name="csrf_token"]')?.value ||
      "";
    if (!model || !recordId || !field) return;

    bar.querySelectorAll(".sum-statusbar-stage").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const stageId = parseInt(btn.dataset.stageId || "0", 10);
        if (!stageId) return;
        bar.querySelectorAll(".sum-statusbar-stage").forEach((el) => {
          el.classList.remove("sum-statusbar-stage--current");
        });
        btn.classList.add("sum-statusbar-stage--current");
        try {
          const res = await fetch("/web/kanban/move", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "X-CSRF-Token": csrf,
            },
            credentials: "same-origin",
            body: JSON.stringify({
              model,
              record_id: recordId,
              field,
              value: stageId,
              csrf_token: csrf,
            }),
          });
          if (!res.ok) throw new Error("move failed");
        } catch (_) {
          window.location.reload();
        }
      });
    });
  });

  document.querySelectorAll(".sum-priority-field").forEach((wrap) => {
    const hidden = wrap.querySelector("[data-sum-priority-value]");
    if (!hidden) return;
    wrap.querySelectorAll(".sum-priority-star[data-priority]").forEach((star) => {
      star.addEventListener("click", () => {
        const val = parseInt(star.dataset.priority || "0", 10);
        hidden.value = String(val);
        wrap.querySelectorAll(".sum-priority-star").forEach((el) => {
          const p = parseInt(el.dataset.priority || "0", 10);
          el.classList.toggle("sum-priority-star--on", p <= val);
        });
      });
    });
  });
}
