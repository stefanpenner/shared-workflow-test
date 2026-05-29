import { test } from "node:test";
import assert from "node:assert/strict";
import { inlineErrors, isSingleInvocation } from "./check-no-inline-scripts.mjs";

test("isSingleInvocation accepts interpreter and bare-script forms", () => {
  assert.ok(isSingleInvocation("node ${{ github.action_path }}/scripts/setup.cli.mjs"));
  assert.ok(isSingleInvocation('node --test --experimental-test-coverage "actions/**/*.test.mjs"'));
  assert.ok(isSingleInvocation("bash scripts/ci/run.sh"));
  assert.ok(isSingleInvocation("scripts/ci/run.sh"));
});

test("isSingleInvocation rejects shell operators and inline eval", () => {
  assert.equal(isSingleInvocation("mkdir -p x && echo y"), false);
  assert.equal(isSingleInvocation("echo hi"), false);
  assert.equal(isSingleInvocation("cat a | grep b"), false);
  assert.equal(isSingleInvocation("echo x > y"), false);
  assert.equal(isSingleInvocation('node -e "process.exit(1)"'), false);
  assert.equal(isSingleInvocation(""), false);
});

test("inlineErrors flags block scalars and an empty run value", () => {
  assert.equal(inlineErrors("steps:\n  - run: |\n      echo hi\n      ls\n").length, 1);
  assert.equal(inlineErrors("steps:\n  - run: >\n      echo hi\n").length, 1);
  assert.equal(inlineErrors("steps:\n  - run: \n").length, 1);
});

test("inlineErrors accepts a single-quoted external invocation", () => {
  assert.equal(inlineErrors("steps:\n  - run: 'scripts/ci/run.sh'\n").length, 0);
});

test("inlineErrors flags inline one-liners with shell operators", () => {
  const errors = inlineErrors("steps:\n  - run: mkdir -p x && echo y\n");
  assert.equal(errors.length, 1);
  assert.match(errors[0].message, /inline logic/);
});

test("inlineErrors allows a single external invocation", () => {
  assert.equal(inlineErrors('steps:\n  - run: "node ${{ github.action_path }}/scripts/run.cli.mjs"\n').length, 0);
});

test("inlineErrors honours an allowlisted step name (mechanism; default allowlist is empty)", () => {
  const yaml = "steps:\n  - name: Bootstrap\n    run: mkdir -p x && echo y >> z\n";
  assert.equal(inlineErrors(yaml).length, 1); // not allowed by default
  assert.equal(inlineErrors(yaml, new Set(["Bootstrap"])).length, 0); // allowed when injected
});

test("inlineErrors does not let an unnamed run inherit an allowlisted name", () => {
  const yaml = "steps:\n  - name: Bootstrap\n    run: mkdir -p x && echo y\n  - run: rm -rf / && echo bad\n";
  assert.equal(inlineErrors(yaml, new Set(["Bootstrap"])).length, 1);
});
