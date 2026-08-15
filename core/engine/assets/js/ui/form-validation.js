/** Client-side required-field validation before workspace form POST. */
function markFieldError(widget, label) {
  widget.classList.add("sum-field-widget--error");
  let err = widget.querySelector(".sum-field-error");
  if (!err) {
    err = document.createElement("p");
    err.className = "sum-field-error";
    widget.appendChild(err);
  }
  err.textContent = `${label} is required.`;
}

function clearFieldError(widget) {
  widget.classList.remove("sum-field-widget--error");
  const err = widget.querySelector(".sum-field-error");
  if (err) err.remove();
}

function fieldWidgetForInput(input) {
  return input.closest(".sum-field-widget");
}

function isEmptyValue(input) {
  if (input.type === "checkbox" || input.type === "radio") {
    return !input.checked;
  }
  return String(input.value || "").trim() === "";
}

function validateRequiredFields(form) {
  let firstBad = null;
  form.querySelectorAll("[data-required='1']").forEach((input) => {
    const widget = fieldWidgetForInput(input);
    if (!widget) return;
    if (isEmptyValue(input)) {
      const label = input.getAttribute("data-field-label") || input.name || "Field";
      markFieldError(widget, label);
      if (!firstBad) firstBad = input;
    } else {
      clearFieldError(widget);
    }
  });
  if (firstBad) {
    firstBad.focus();
    return false;
  }
  return true;
}

function highlightFieldsFromQuery() {
  const params = new URLSearchParams(window.location.search);
  const raw = params.get("field_errors");
  if (!raw) return;
  raw.split(",").map((s) => s.trim()).filter(Boolean).forEach((name) => {
    const input = document.querySelector(`[name="${CSS.escape(name)}"]`);
    if (!input) return;
    const widget = fieldWidgetForInput(input);
    if (!widget) return;
    const label = input.getAttribute("data-field-label") || name;
    markFieldError(widget, label);
  });
}

export function initFormValidation() {
  const form = document.getElementById("sum-workspace-record-form");
  if (form) {
    form.addEventListener("submit", (e) => {
      form.querySelectorAll(".sum-field-widget--error").forEach(clearFieldError);
      if (!validateRequiredFields(form)) {
        e.preventDefault();
      }
    });
    form.querySelectorAll("[data-required='1']").forEach((input) => {
      input.addEventListener("input", () => {
        const widget = fieldWidgetForInput(input);
        if (widget && !isEmptyValue(input)) clearFieldError(widget);
      });
    });
  }
  highlightFieldsFromQuery();
}
