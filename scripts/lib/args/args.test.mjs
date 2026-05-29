import { test } from "node:test";
import assert from "node:assert/strict";
import { requireArgs } from "./args.mjs";

test("reads named flags in both --flag value and --flag=value forms", () => {
  assert.deepEqual(requireArgs(["paths", "config"], ["--paths", "src", "--config=.eslintrc"]), {
    paths: "src",
    config: ".eslintrc",
  });
});

test("returns exactly the requested names", () => {
  assert.deepEqual(requireArgs(["suite"], ["--suite=unit"]), { suite: "unit" });
});

test("rejects unknown flags (strict parsing)", () => {
  assert.throws(() => requireArgs(["suite"], ["--suite=unit", "--bogus=x"]), /bogus/);
});

test("throws, naming the flag, when a required flag is absent", () => {
  assert.throws(() => requireArgs(["project-name"], []), /--project-name/);
});

test("throws, naming the flag, when a flag is present but empty", () => {
  assert.throws(() => requireArgs(["project-name"], ["--project-name="]), /--project-name/);
});

test("defaults argv to process.argv.slice(2) when not given", () => {
  const saved = process.argv;
  try {
    process.argv = ["node", "cli.mjs", "--suite=integration"];
    assert.deepEqual(requireArgs(["suite"]), { suite: "integration" });
  } finally {
    process.argv = saved;
  }
});
