// Thin entry the Test action invokes: parse named args (params), call the pure module.
import { requireArgs } from "../../../scripts/lib/args/args.mjs";
import { report } from "./test.mjs";

const { suite, coverage } = requireArgs(["suite", "coverage"]);
console.log(report(suite, coverage));
