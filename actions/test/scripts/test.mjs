// Pure logic for the Test action.
export function testSummary(suite, coverage) {
  const resolvedSuite = (suite ?? "").trim() || "unit";
  const resolvedCoverage = (coverage ?? "").trim() || "true";
  return `Running tests...\nSuite: ${resolvedSuite}\nCoverage: ${resolvedCoverage}`;
}
