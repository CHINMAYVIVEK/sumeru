/** Build /web/rel/search URLs and normalize JSON results. */
import { getJSON } from "./fetch-json.js";

export function relSearchURL(model, { query = "", limit = 20, filterField = "", filterId = 0 } = {}) {
  const params = new URLSearchParams({
    model,
    q: query,
    limit: String(limit),
  });
  if (filterField) {
    params.set("filter_field", filterField);
    params.set("filter_id", String(filterId || 0));
  }
  return `/web/rel/search?${params.toString()}`;
}

export async function fetchRelSearchRows(model, options = {}) {
  try {
    const data = await getJSON(relSearchURL(model, options));
    return Array.isArray(data.results) ? data.results : [];
  } catch (_) {
    return [];
  }
}
