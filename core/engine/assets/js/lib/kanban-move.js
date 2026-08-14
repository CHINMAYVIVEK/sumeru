/** Persist a kanban/statusbar field change via POST /web/kanban/move. */
import { postJSON } from "./fetch-json.js";

export async function moveKanbanRecord({ model, recordId, field, value, csrfEl }) {
  return postJSON(
    "/web/kanban/move",
    {
      model,
      record_id: recordId,
      field,
      value,
    },
    { csrfEl }
  );
}
