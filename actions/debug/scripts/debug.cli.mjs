import { execFileSync } from "node:child_process";
import { treeReport, gitReport } from "./debug.mjs";

// Capture (don't inherit) stderr so best-effort probes that fail stay quiet,
// matching the original action's `2>/dev/null`. On failure execFileSync throws and
// the pure formatters substitute their fallback text.
const exec = (file, args) => execFileSync(file, args, { encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] });

console.log(treeReport(exec, process.env));
console.log(gitReport(exec));
