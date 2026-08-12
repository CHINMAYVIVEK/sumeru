/**
 * Multi-select from dropdown: pick options into tags; remove returns them to the list.
 * Used for core.user company_ids (Allowed Companies).
 */
export function initMultiSelect() {
  const roots = Array.from(document.querySelectorAll("[data-sum-multi-select]"));
  if (!roots.length) return;

  roots.forEach((root) => {
    const fieldName = root.getAttribute("data-name") || "";
    const tagsEl = root.querySelector("[data-sum-multi-tags]");
    const addEl = root.querySelector("[data-sum-multi-add]");
    if (!fieldName || !tagsEl || !addEl) return;

    const sortOptions = () => {
      const opts = Array.from(addEl.querySelectorAll("option")).filter((o) => o.value !== "");
      opts.sort((a, b) => (a.textContent || "").localeCompare(b.textContent || "", undefined, { sensitivity: "base" }));
      opts.forEach((o) => addEl.appendChild(o));
    };

    const addTag = (id, label) => {
      if (!id || tagsEl.querySelector(`[data-id="${CSS.escape(id)}"]`)) return;
      const tag = document.createElement("span");
      tag.className = "sum-multi-select-tag";
      tag.setAttribute("data-id", id);

      const text = document.createElement("span");
      text.className = "sum-multi-select-tag-label";
      text.textContent = label;

      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "sum-multi-select-remove";
      btn.setAttribute("data-sum-multi-remove", "");
      btn.setAttribute("aria-label", "Remove");
      btn.textContent = "×";

      const hidden = document.createElement("input");
      hidden.type = "hidden";
      hidden.name = fieldName;
      hidden.value = id;

      tag.appendChild(text);
      tag.appendChild(btn);
      tag.appendChild(hidden);
      tagsEl.appendChild(tag);
    };

    const returnOption = (id, label) => {
      if (!id || addEl.querySelector(`option[value="${CSS.escape(id)}"]`)) return;
      const opt = document.createElement("option");
      opt.value = id;
      opt.textContent = label;
      addEl.appendChild(opt);
      sortOptions();
    };

    addEl.addEventListener("change", () => {
      const opt = addEl.selectedOptions && addEl.selectedOptions[0];
      if (!opt || !opt.value) return;
      const id = opt.value;
      const label = opt.textContent || id;
      opt.remove();
      addEl.value = "";
      addTag(id, label);
    });

    root.addEventListener("click", (ev) => {
      const btn = ev.target.closest("[data-sum-multi-remove]");
      if (!btn || !root.contains(btn)) return;
      ev.preventDefault();
      const tag = btn.closest(".sum-multi-select-tag");
      if (!tag) return;
      const id = tag.getAttribute("data-id") || "";
      const labelEl = tag.querySelector(".sum-multi-select-tag-label");
      const label = labelEl ? labelEl.textContent || id : id;
      tag.remove();
      returnOption(id, label);
    });
  });
}
