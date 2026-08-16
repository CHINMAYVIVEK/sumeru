/**
 * Report toolbar: CSV/PDF export and bulk CSV upload on enabled workspace views.
 */
import { getCsrfToken } from "../lib/csrf.js";

function parseJSON(raw, fallback) {
  if (!raw) return fallback;
  try {
    return JSON.parse(raw);
  } catch {
    return fallback;
  }
}

function selectedFields(container) {
  return [...container.querySelectorAll(".sum-report-field-check:checked")].map((el) => el.value);
}

function buildQuery(params) {
  const q = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== null && String(v) !== "") q.set(k, String(v));
  });
  return q.toString();
}

function openDownload(url) {
  window.location.href = url;
}

function renderFieldList(container, fields) {
  const list = container.querySelector(".sum-report-field-list");
  if (!list) return;
  list.innerHTML = "";
  fields.forEach((f) => {
    const label = document.createElement("label");
    label.className = "sum-report-field-item";
    label.innerHTML = `<input type="checkbox" class="sum-report-field-check" value="${f.name}" checked /> ${f.label || f.name}`;
    list.appendChild(label);
  });
}

function setModalMode(container, action) {
  const pdfSize = container.querySelector(".sum-report-pdf-size");
  const importMode = container.querySelector(".sum-report-import-mode");
  if (pdfSize) pdfSize.hidden = action !== "export-pdf";
  if (importMode) importMode.hidden = action !== "bulk-upload";
}

function showModal(container, action) {
  const modal = container.querySelector(".sum-report-modal");
  if (!modal) return;
  modal.hidden = false;
  modal.setAttribute("aria-hidden", "false");
  container.dataset.pendingAction = action;
  setModalMode(container, action);
  const title = container.querySelector(".sum-report-modal-title");
  if (title) {
    const labels = {
      "export-csv": "Download CSV",
      "export-pdf": "Download PDF",
      "bulk-template": "Download import template",
      "bulk-upload": "Bulk upload CSV",
    };
    title.textContent = labels[action] || "Report options";
  }
}

function hideModal(container) {
  const modal = container.querySelector(".sum-report-modal");
  if (!modal) return;
  modal.hidden = true;
  modal.setAttribute("aria-hidden", "true");
  delete container.dataset.pendingAction;
}

function runReportAction(container) {
  const action = container.dataset.pendingAction;
  if (!action) return;

  const model = container.dataset.model || "";
  const actionId = container.dataset.actionId || "";
  const recordId = container.dataset.recordId || "";
  const fields = selectedFields(container).join(",");
  const pageSize = container.querySelector(".sum-report-page-size")?.value || "a4";
  const importMode = container.querySelector(".sum-report-import-mode-select")?.value || "create";

  if (!fields) {
    window.alert("Select at least one field.");
    return;
  }

  if (action === "export-csv") {
    openDownload(`/web/export/csv?${buildQuery({ model, fields, action: actionId, id: recordId })}`);
    hideModal(container);
    return;
  }
  if (action === "export-pdf") {
    openDownload(`/web/export/pdf?${buildQuery({ model, fields, action: actionId, id: recordId, page_size: pageSize })}`);
    hideModal(container);
    return;
  }
  if (action === "bulk-template") {
    openDownload(`/web/bulk/template?${buildQuery({ model, fields })}`);
    hideModal(container);
    return;
  }
  if (action === "bulk-upload") {
    const fileInput = container.querySelector(".sum-report-file-input");
    if (!fileInput?.files?.length) {
      fileInput?.click();
      return;
    }
    submitBulkUpload(container, fileInput.files[0], { model, fields, actionId, importMode });
  }
}

async function submitBulkUpload(container, file, opts) {
  const csrf = container.dataset.csrf || getCsrfToken(container);
  const form = new FormData();
  form.set("csrf_token", csrf);
  form.set("model", opts.model);
  form.set("fields", opts.fields);
  form.set("import_mode", opts.importMode);
  if (opts.actionId) form.set("action", opts.actionId);
  form.set("next", window.location.pathname + window.location.search);
  form.set("file", file);

  const resp = await fetch("/web/bulk/upload", { method: "POST", body: form, credentials: "same-origin" });
  if (resp.redirected) {
    window.location.href = resp.url;
    return;
  }
  if (!resp.ok) {
    const text = await resp.text();
    window.alert(text || "Upload failed");
  }
}

function initToolbar(container) {
  const fields = parseJSON(container.dataset.fields, []);
  renderFieldList(container, fields);

  const menuBtn = container.querySelector(".sum-list-btn-report");
  const menu = container.querySelector(".sum-report-menu");
  menuBtn?.addEventListener("click", (e) => {
    e.stopPropagation();
    if (menu) menu.hidden = !menu.hidden;
  });
  document.addEventListener("click", () => {
    if (menu) menu.hidden = true;
  });

  container.querySelectorAll(".sum-report-item").forEach((btn) => {
    btn.addEventListener("click", (e) => {
      e.stopPropagation();
      if (menu) menu.hidden = true;
      const action = btn.dataset.reportAction;
      if (action === "bulk-upload") {
        showModal(container, action);
      } else {
        showModal(container, action);
      }
    });
  });

  container.querySelector(".sum-report-cancel")?.addEventListener("click", () => hideModal(container));
  container.querySelector(".sum-report-run")?.addEventListener("click", () => runReportAction(container));

  const fileInput = container.querySelector(".sum-report-file-input");
  fileInput?.addEventListener("change", () => {
    if (container.dataset.pendingAction === "bulk-upload" && fileInput.files?.length) {
      runReportAction(container);
    }
  });
}

export function initReportExchange() {
  document.querySelectorAll("[data-report-exchange]").forEach((el) => initToolbar(el));
}
