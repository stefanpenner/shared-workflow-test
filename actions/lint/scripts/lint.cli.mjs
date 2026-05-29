import { report } from "./lint.mjs";

console.log(report(process.env.LINT_PATHS, process.env.LINT_CONFIG));
