import { report } from "./test.mjs";

console.log(report(process.env.TEST_SUITE, process.env.COVERAGE_ENABLED));
