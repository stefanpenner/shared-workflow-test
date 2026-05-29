// Thin entry the Lint action invokes: parse named args (params), call the pure module.
import { requireArgs } from "../../../scripts/lib/args/args.mjs";
import { report } from "./lint.mjs";

const { paths, config } = requireArgs(["paths", "config"]);
console.log(report(paths, config));
