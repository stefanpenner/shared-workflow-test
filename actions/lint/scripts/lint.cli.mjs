import { lintSummary } from "./lint.mjs";

console.log(lintSummary(process.env.LINT_PATHS, process.env.LINT_CONFIG));
