import { parseArgs } from "node:util";

/**
 * Read required named string flags from argv — `requireArgs(['workflows-repo'])` reads
 * `--workflows-repo X` (or `--workflows-repo=X`). Throws, naming the flag, if any is missing or
 * empty. Secrets stay in env (never argv); this is for plain config. `argv` is injectable for tests.
 */
export function requireArgs<K extends string>(
  names: readonly K[],
  argv: string[] = process.argv.slice(2),
): Record<K, string> {
  const options: Record<string, { type: "string" }> = {};
  for (const name of names) options[name] = { type: "string" };

  const { values } = parseArgs({ args: argv, options, allowPositionals: false });

  const result = {} as Record<K, string>;
  for (const name of names) {
    const value = values[name];
    if (typeof value !== "string" || value === "") throw new Error(`missing required --${name}`);
    result[name] = value;
  }
  return result;
}
