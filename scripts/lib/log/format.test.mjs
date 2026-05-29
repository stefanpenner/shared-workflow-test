import { test } from "node:test";
import assert from "node:assert/strict";
import { heading, kv, section, group } from "./format.mjs";

test("heading prefixes the title", () => {
  assert.equal(heading("Setup"), "▸ Setup");
});

test("kv aligns values to the widest key", () => {
  assert.equal(kv({ a: "1", abc: "2" }), "  a    1\n  abc  2");
});

test("kv handles an empty map", () => {
  assert.equal(kv({}), "");
});

test("section combines a heading with aligned rows", () => {
  assert.equal(
    section("Lint", { paths: ".", config: ".eslintrc" }),
    "▸ Lint\n  paths   .\n  config  .eslintrc",
  );
});

test("group wraps a body in a collapsible GHA block", () => {
  assert.equal(group("Env", "line"), "::group::Env\nline\n::endgroup::");
});
