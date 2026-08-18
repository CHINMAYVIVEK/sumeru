import type { HttpService } from "../services/http.js";
import type { SwcBootstrap } from "../types/bootstrap.js";

/** Company switcher in the shell top bar (SPA POST without full reload). */
export function initCompanySwitcher(boot: SwcBootstrap, http: HttpService): void {
  if (!boot.showCompanySwitcher || boot.companies.length <= 1) return;
  const host = document.getElementById("sum-company-switcher");
  if (!host) return;

  const select = document.createElement("select");
  select.className = "sum-company-switcher-select";
  select.setAttribute("aria-label", "Company");

  for (const company of boot.companies) {
    const opt = document.createElement("option");
    opt.value = String(company.id);
    opt.textContent = company.name;
    if (company.id === boot.activeCompanyId) opt.selected = true;
    select.appendChild(opt);
  }

  select.addEventListener("change", () => {
    void http
      .postForm("/web/company/switch", {
        company_id: select.value,
        next: window.location.pathname + window.location.search,
      })
      .then(() => {
        document.dispatchEvent(new CustomEvent("swc:company-changed"));
        window.dispatchEvent(new PopStateEvent("popstate"));
      });
  });

  host.replaceChildren(select);
}
