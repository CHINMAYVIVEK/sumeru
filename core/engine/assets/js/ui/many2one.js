/**
 * Many2one typeahead against /web/rel/search.
 */
import { fetchRelSearchRows } from "../lib/rel-search.js";

export function initMany2One() {
  document.querySelectorAll("[data-sum-m2o]").forEach((root) => {
    const comodel = root.getAttribute("data-comodel") || "";
    const idInput = root.querySelector("[data-sum-m2o-id]");
    const search = root.querySelector("[data-sum-m2o-search]");
    const list = root.querySelector("[data-sum-m2o-results]");
    if (!comodel || !idInput || !search || !list) return;

    let timer = 0;
    const hide = () => {
      list.hidden = true;
      list.innerHTML = "";
    };
    const show = (rows) => {
      list.innerHTML = "";
      if (!rows.length) {
        hide();
        return;
      }
      rows.forEach((row) => {
        const li = document.createElement("li");
        li.className = "sum-m2o-option";
        li.textContent = row.name || String(row.id);
        li.addEventListener("mousedown", (ev) => {
          ev.preventDefault();
          idInput.value = String(row.id);
          search.value = row.name || String(row.id);
          hide();
        });
        list.appendChild(li);
      });
      list.hidden = false;
    };
    search.addEventListener("input", () => {
      idInput.value = "";
      clearTimeout(timer);
      timer = window.setTimeout(async () => {
        show(await fetchRelSearchRows(comodel, { query: search.value.trim(), limit: 20 }));
      }, 220);
    });
    search.addEventListener("focus", async () => {
      if (search.value.trim()) show(await fetchRelSearchRows(comodel, { query: search.value.trim(), limit: 20 }));
    });
    search.addEventListener("blur", () => {
      window.setTimeout(hide, 150);
    });
  });
}
