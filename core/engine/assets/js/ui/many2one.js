/**
 * Many2one typeahead against /web/rel/search.
 */
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
      rows.forEach((r) => {
        const li = document.createElement("li");
        li.className = "sum-m2o-option";
        li.textContent = r.name || String(r.id);
        li.addEventListener("mousedown", (ev) => {
          ev.preventDefault();
          idInput.value = String(r.id);
          search.value = r.name || String(r.id);
          hide();
        });
        list.appendChild(li);
      });
      list.hidden = false;
    };
    const fetchRows = async (q) => {
      const url = `/web/rel/search?model=${encodeURIComponent(comodel)}&q=${encodeURIComponent(q || "")}&limit=20`;
      const res = await fetch(url, { credentials: "same-origin" });
      if (!res.ok) return [];
      const data = await res.json();
      return Array.isArray(data.results) ? data.results : [];
    };
    search.addEventListener("input", () => {
      idInput.value = "";
      clearTimeout(timer);
      timer = window.setTimeout(async () => {
        show(await fetchRows(search.value.trim()));
      }, 220);
    });
    search.addEventListener("focus", async () => {
      if (search.value.trim()) show(await fetchRows(search.value.trim()));
    });
    search.addEventListener("blur", () => {
      window.setTimeout(hide, 150);
    });
  });
}
