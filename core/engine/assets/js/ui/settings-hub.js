/**
 * Settings hub (/web/settings): client-side filter for section cards.
 */
import { initSearchShortcut } from "./search-shortcut.js";
import { tokenizeQuery, matchesAllTokens } from "../lib/token-match.js";

export function initSettingsHub() {
  const root = document.querySelector(".sum-settings-hub");
  if (!root) return;

  const input = document.getElementById("sum-settings-hub-q");
  const wrap = document.getElementById("sum-settings-hub-sections");
  if (!input || !wrap) return;

  const sections = () => Array.from(wrap.querySelectorAll(".sum-settings-hub-section"));

  function applyFilter() {
    const tokens = tokenizeQuery(input.value);
    let anyVisible = false;
    for (const section of sections()) {
      const filterText = section.getAttribute("data-sum-settings-filter") || "";
      const match = matchesAllTokens(filterText, tokens);
      section.classList.toggle("is-filtered-out", !match);
      if (match) anyVisible = true;
    }
    let empty = wrap.querySelector(".sum-settings-hub-filter-empty");
    if (tokens.length > 0 && !anyVisible && sections().length > 0) {
      if (!empty) {
        empty = document.createElement("p");
        empty.className = "sum-settings-hub-filter-empty sum-settings-hub-empty";
        empty.setAttribute("role", "status");
        empty.textContent = "No sections match your search.";
        wrap.appendChild(empty);
      }
    } else if (empty) {
      empty.remove();
    }
  }

  input.addEventListener("input", applyFilter);
  initSearchShortcut("sum-settings-hub-q", { skipWhenLauncherOpen: false });
}

initSettingsHub();
