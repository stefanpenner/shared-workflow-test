// Thin entry the Setup action invokes: read env, do the real I/O, no logic of its own.
import { appendFileSync } from "node:fs";
import { resolveNodeVersion, greeting, renderOutputs } from "./setup.mjs";

const version = resolveNodeVersion(process.env.NODE_VERSION);
console.log(greeting(process.env.PROJECT_NAME));

const outputPath = process.env.GITHUB_OUTPUT;
if (!outputPath) throw new Error("GITHUB_OUTPUT is not set");
appendFileSync(outputPath, renderOutputs({ node_version: version }));
