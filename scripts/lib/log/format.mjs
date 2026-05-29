// Pure formatting helpers for action CLI output. Every function returns a string, so
// the *.cli.mjs entries just console.log the result and *.test.mjs can assert it.
// GitHub Actions renders ::group:: / ::endgroup:: as a collapsible log section.

export function heading(title) {
  return `▸ ${title}`; // ▸ title
}

// Align a { key: value } map into "  key   value" rows.
export function kv(pairs) {
  const keys = Object.keys(pairs);
  const width = keys.reduce((max, key) => Math.max(max, key.length), 0);
  return keys.map((key) => `  ${key.padEnd(width)}  ${pairs[key]}`).join("\n");
}

// A titled block: heading line followed by aligned key/value rows.
export function section(title, pairs) {
  return `${heading(title)}\n${kv(pairs)}`;
}

// A collapsible GitHub Actions log group wrapping an arbitrary body.
export function group(title, body) {
  return `::group::${title}\n${body}\n::endgroup::`;
}
