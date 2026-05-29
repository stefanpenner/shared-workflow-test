// Thin entry the Setup action invokes: parse named args (params), read env sinks, do the I/O.
import { appendFileSync } from "node:fs";
import { requireArgs } from "../../../scripts/lib/args/args.mjs";
import { resolveNodeVersion, report, renderOutputs } from "./setup.mjs";

const { "project-name": projectName, "node-version": nodeVersion } = requireArgs([
  "project-name",
  "node-version",
]);

const version = resolveNodeVersion(nodeVersion);
console.log(report(projectName, version));

// GITHUB_OUTPUT is a GHA-provided sink (global state), so it stays in env — not a parameter.
const outputPath = process.env.GITHUB_OUTPUT;
if (!outputPath) throw new Error("GITHUB_OUTPUT is not set");
appendFileSync(outputPath, renderOutputs({ node_version: version }));
