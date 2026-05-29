import { testSummary } from "./test.mjs";

console.log(testSummary(process.env.TEST_SUITE, process.env.COVERAGE_ENABLED));
