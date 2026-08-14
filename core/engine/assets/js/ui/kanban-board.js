/**
 * Grouped kanban: drag cards between columns and persist via POST /web/kanban/move.
 */
import { moveKanbanRecord } from "../lib/kanban-move.js";

export function initKanbanBoard() {
  const board = document.querySelector(".sum-kanban-board--grouped[data-draggable='1']");
  if (!board) return;

  const model = board.dataset.model || "";
  const groupField = board.dataset.groupField || "";

  let dragCard = null;
  let sourceColumn = null;

  board.querySelectorAll(".sum-kanban-card[draggable='true']").forEach((card) => {
    card.addEventListener("dragstart", (e) => {
      dragCard = card;
      sourceColumn = card.closest(".sum-kanban-stage-column");
      window.__sumKanbanDragging = true;
      card.classList.add("sum-kanban-card--dragging");
      if (e.dataTransfer) {
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", card.dataset.recordId || "");
      }
    });
    card.addEventListener("dragend", () => {
      window.__sumKanbanDragging = false;
      card.classList.remove("sum-kanban-card--dragging");
      board.querySelectorAll(".sum-kanban-stage-column--drop-target").forEach((el) => {
        el.classList.remove("sum-kanban-stage-column--drop-target");
      });
      dragCard = null;
      sourceColumn = null;
    });
    card.addEventListener("click", (e) => {
      if (window.__sumKanbanDragging) {
        e.preventDefault();
        return;
      }
      const href = card.dataset.href;
      if (href) window.location.href = href;
    });
    card.addEventListener("keydown", (e) => {
      if (e.key === "Enter" && !window.__sumKanbanDragging) {
        const href = card.dataset.href;
        if (href) window.location.href = href;
      }
    });
  });

  board.querySelectorAll(".sum-kanban-cards").forEach((zone) => {
    zone.addEventListener("dragover", (e) => {
      e.preventDefault();
      const col = zone.closest(".sum-kanban-stage-column");
      if (col) col.classList.add("sum-kanban-stage-column--drop-target");
      if (e.dataTransfer) e.dataTransfer.dropEffect = "move";
    });
    zone.addEventListener("dragleave", (e) => {
      const col = zone.closest(".sum-kanban-stage-column");
      if (col && !col.contains(e.relatedTarget)) {
        col.classList.remove("sum-kanban-stage-column--drop-target");
      }
    });
    zone.addEventListener("drop", async (e) => {
      e.preventDefault();
      const col = zone.closest(".sum-kanban-stage-column");
      if (col) col.classList.remove("sum-kanban-stage-column--drop-target");
      if (!dragCard || !col) return;

      const recordId = parseInt(dragCard.dataset.recordId || "0", 10);
      const groupValue = parseInt(col.dataset.groupValue || "0", 10);
      if (!recordId || !model || !groupField) return;

      const prevColumn = sourceColumn;
      zone.appendChild(dragCard);
      updateColumnCount(col);
      if (prevColumn && prevColumn !== col) updateColumnCount(prevColumn);

      try {
        await moveKanbanRecord({
          model,
          recordId,
          field: groupField,
          value: groupValue,
          csrfEl: board,
        });
      } catch (_) {
        if (prevColumn) {
          const prevZone = prevColumn.querySelector(".sum-kanban-cards");
          if (prevZone) prevZone.appendChild(dragCard);
          updateColumnCount(prevColumn);
          updateColumnCount(col);
        }
        alert("Could not move record. Check permissions and try again.");
      }
    });
  });
}

function updateColumnCount(col) {
  const countEl = col.querySelector(".sum-kanban-stage-count");
  const cards = col.querySelectorAll(".sum-kanban-card");
  if (countEl) countEl.textContent = String(cards.length);
}
