/**
 * Many2one <select widget="selection"> with country→state→city cascade.
 */
export function initMany2OneSelect() {
  const roots = Array.from(document.querySelectorAll("[data-sum-m2o-select]"));
  if (!roots.length) return;

  const byField = new Map();
  roots.forEach((root) => {
    const field = root.getAttribute("data-field") || "";
    if (field) byField.set(field, root);
  });

  const fillSelect = async (root, filterField, filterId) => {
    const comodel = root.getAttribute("data-comodel") || "";
    const select = root.querySelector("[data-sum-m2o-select-el]");
    if (!comodel || !select) return;
    const prev = select.value;
    let url = `/web/rel/search?model=${encodeURIComponent(comodel)}&limit=500&q=`;
    if (filterField) {
      url += `&filter_field=${encodeURIComponent(filterField)}&filter_id=${encodeURIComponent(String(filterId || 0))}`;
    }
    let rows = [];
    try {
      const res = await fetch(url, { credentials: "same-origin" });
      if (res.ok) {
        const data = await res.json();
        rows = Array.isArray(data.results) ? data.results : [];
      }
    } catch (_) {
      rows = [];
    }
    select.innerHTML = `<option value="">—</option>`;
    rows.forEach((r) => {
      const opt = document.createElement("option");
      opt.value = String(r.id);
      opt.textContent = r.name || String(r.id);
      if (r.phone_code) opt.setAttribute("data-phone-code", String(r.phone_code));
      if (String(r.id) === prev) opt.selected = true;
      select.appendChild(opt);
    });
    if (prev && !Array.from(select.options).some((o) => o.value === prev && o.value !== "")) {
      select.value = "";
    }
  };

  const updatePhoneCode = (root) => {
    const select = root.querySelector("[data-sum-m2o-select-el]");
    const badge = root.querySelector("[data-sum-phone-code]");
    if (!select || !badge) return;
    const opt = select.selectedOptions && select.selectedOptions[0];
    const code = opt ? opt.getAttribute("data-phone-code") || "" : "";
    badge.textContent = code ? `+${code}` : "";
  };

  roots.forEach((root) => {
    const select = root.querySelector("[data-sum-m2o-select-el]");
    if (!select) return;
    select.addEventListener("change", async () => {
      const field = root.getAttribute("data-field") || "";
      if (field === "country_id") {
        updatePhoneCode(root);
        const stateRoot = byField.get("state_id");
        const cityRoot = byField.get("city_id");
        const countryId = select.value || "0";
        if (stateRoot) {
          await fillSelect(stateRoot, "country_id", countryId);
          const stateSelect = stateRoot.querySelector("[data-sum-m2o-select-el]");
          if (stateSelect) stateSelect.value = "";
        }
        if (cityRoot) {
          await fillSelect(cityRoot, "country_id", countryId);
          const citySelect = cityRoot.querySelector("[data-sum-m2o-select-el]");
          if (citySelect) citySelect.value = "";
        }
      } else if (field === "state_id") {
        const cityRoot = byField.get("city_id");
        if (cityRoot) {
          const stateId = select.value || "0";
          if (stateId && stateId !== "0") {
            await fillSelect(cityRoot, "state_id", stateId);
          } else {
            const countryRoot = byField.get("country_id");
            const countrySelect = countryRoot && countryRoot.querySelector("[data-sum-m2o-select-el]");
            const countryId = countrySelect ? countrySelect.value || "0" : "0";
            await fillSelect(cityRoot, "country_id", countryId);
          }
          const citySelect = cityRoot.querySelector("[data-sum-m2o-select-el]");
          if (citySelect) citySelect.value = "";
        }
      }
    });
  });
}
