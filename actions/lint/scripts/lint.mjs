// Pure logic for the Lint action.
import { section } from "../../../scripts/lib/log/format.mjs";

export function report(paths, config) {
  return section("Lint", {
    paths: (paths ?? "").trim() || ".",
    config: (config ?? "").trim() || ".eslintrc",
  });
}
