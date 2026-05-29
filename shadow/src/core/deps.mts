interface PackageJson {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
}

/**
 * Return dependency names that aren't in the allowlist — used to enforce "yaml is the only runtime
 * dependency" (devDeps limited to the isolated type-check tooling). Pure; the bin reads package.json.
 */
export function disallowedDeps(
  pkg: PackageJson,
  allow: { deps: string[]; devDeps: string[] },
): string[] {
  // An allow entry ending in "/*" is a prefix (e.g. "@actions/*" allows "@actions/core").
  const allowed = (name: string, ok: string[]): boolean =>
    ok.some((entry) =>
      entry.endsWith("/*") ? name.startsWith(entry.slice(0, -1)) : name === entry,
    );
  const extra = (have: Record<string, string> | undefined, ok: string[]): string[] =>
    Object.keys(have ?? {}).filter((name) => !allowed(name, ok));
  return [...extra(pkg.dependencies, allow.deps), ...extra(pkg.devDependencies, allow.devDeps)];
}
