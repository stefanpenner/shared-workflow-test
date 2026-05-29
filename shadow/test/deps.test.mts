import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { disallowedDeps } from "../src/core/deps.mts";

const allow = { deps: ["yaml", "@actions/*"], devDeps: ["typescript", "@types/node"] };

describe("disallowedDeps", () => {
  it("passes when only allowed deps are present", () => {
    assert.deepEqual(
      disallowedDeps(
        { dependencies: { yaml: "^2" }, devDependencies: { typescript: "^5" } },
        allow,
      ),
      [],
    );
  });

  it("flags a disallowed runtime dependency", () => {
    assert.deepEqual(disallowedDeps({ dependencies: { yaml: "^2", lodash: "^4" } }, allow), [
      "lodash",
    ]);
  });

  it("allows @actions/* via the prefix entry", () => {
    assert.deepEqual(
      disallowedDeps({ dependencies: { "@actions/core": "^1", "@actions/github": "^6" } }, allow),
      [],
    );
  });

  it("flags a disallowed devDependency", () => {
    assert.deepEqual(disallowedDeps({ devDependencies: { eslint: "^9" } }, allow), ["eslint"]);
  });

  it("handles missing dependency maps", () => {
    assert.deepEqual(disallowedDeps({}, allow), []);
  });
});
