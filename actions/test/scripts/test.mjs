// Pure logic for the Test action.
import { section } from "../../../scripts/lib/log/format.mjs";

export function report(suite, coverage) {
  const resolvedSuite = (suite ?? "").trim() || "unit";
  const coverageOn = (coverage ?? "").trim().toLowerCase() !== "false";
  return section("Test", {
    suite: resolvedSuite,
    coverage: coverageOn ? "enabled" : "disabled",
  });
}
