/**
 * Bulk import mapping form: sync column mapping JSON before confirm.
 */
function parseMapping(raw) {
  if (!raw) return {};
  try {
    return JSON.parse(raw);
  } catch {
    return {};
  }
}

function syncMappingInput(root) {
  const hidden = root.querySelector('input[name="column_mapping"], textarea[name="column_mapping"]');
  if (!hidden) return;
  const mapping = {};
  root.querySelectorAll(".sum-bulk-map-row").forEach((row) => {
    const header = row.dataset.csvHeader;
    const field = row.querySelector(".sum-bulk-map-select")?.value || "";
    if (header) mapping[header] = field;
  });
  hidden.value = JSON.stringify(mapping);
}

export function initBulkImportForm() {
  const root = document.querySelector(".sum-bulk-import-form");
  if (!root) return;

  root.querySelectorAll(".sum-bulk-map-select").forEach((sel) => {
    sel.addEventListener("change", () => syncMappingInput(root));
  });

  root.closest("form")?.addEventListener("submit", () => syncMappingInput(root));
}
