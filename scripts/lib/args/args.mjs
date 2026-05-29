// Tested arg parsing for the action CLIs: parameters arrive as named flags, never env vars
// (env is for global state/sinks only — see CLAUDE.md rule 5). The .mjs twin of
// shadow/src/core/args.mts. Pure and argv-injectable so args.test.mjs needs no real process.
import { parseArgs } from "node:util";

// Read required named string flags from argv — `requireArgs(['paths'])` reads `--paths X` or
// `--paths=X`. Throws, naming the flag, if any is missing or empty.
export function requireArgs(names, argv = process.argv.slice(2)) {
  const options = {};
  for (const name of names) options[name] = { type: "string" };

  const { values } = parseArgs({ args: argv, options, allowPositionals: false });

  const result = {};
  for (const name of names) {
    const value = values[name];
    if (typeof value !== "string" || value === "") throw new Error(`missing required --${name}`);
    result[name] = value;
  }
  return result;
}
