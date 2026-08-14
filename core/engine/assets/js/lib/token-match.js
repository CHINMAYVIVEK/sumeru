/** Multi-word token search: every token must appear in the haystack. */

export function tokenizeQuery(query) {
  return String(query || "")
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean);
}

export function matchesAllTokens(haystack, tokens) {
  if (tokens.length === 0) return true;
  const searchText = String(haystack || "").toLowerCase();
  return tokens.every((token) => searchText.includes(token));
}
