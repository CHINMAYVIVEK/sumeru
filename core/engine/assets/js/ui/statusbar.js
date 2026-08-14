/**
 * Clickable form statusbar (e.g. CRM stage_id) — persists via POST /web/kanban/move.
 */
import { moveKanbanRecord } from "../lib/kanban-move.js";

export function initStatusbar() {
  document.querySelectorAll("[data-sum-statusbar]").forEach((bar) => {
    const model = bar.dataset.model || "";
    const recordId = parseInt(bar.dataset.recordId || "0", 10);
    const field = bar.dataset.field || "";
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
          await moveKanbanRecord({
            model,
            recordId,
            field,
            value: stageId,
            csrfEl: bar,
          });
        } catch (_) {
          window.location.reload();
        }
      });
    });
  });
}
