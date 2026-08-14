/** Parse JSON embedded in a hidden DOM element (e.g. shell bootstrap data). */
export function readEmbeddedJSON(elementId, { expectStringArray = false, fallback = null } = {}) {
  const element = document.getElementById(elementId);
  if (!element?.textContent) {
    return fallback ?? (expectStringArray ? [] : null);
  }
  try {
    const parsed = JSON.parse(element.textContent);
    if (expectStringArray) {
      return Array.isArray(parsed) ? parsed.map(String) : [];
    }
    return parsed;
  } catch (_) {
    return fallback ?? (expectStringArray ? [] : null);
  }
}
