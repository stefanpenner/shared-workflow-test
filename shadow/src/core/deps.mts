interface PackageJson {
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
}

/**
 * Return dependency names that aren't in the allowlist — used to enforce "yaml is the only runtime
 * dependency" (devDeps limited to the isolated type-check tooling). Pure; the bin reads package.json.
 */
export function disallowedDeps(pkg: PackageJson, allow: { deps: string[]; devDeps: string[] }): string[] {
  const extra = (have: Record<string, string> | undefined, ok: string[]): string[] =>
    Object.keys(have ?? {}).filter((name) => !ok.includes(name));
  return [...extra(pkg.dependencies, allow.deps), ...extra(pkg.devDependencies, allow.devDeps)];
}
