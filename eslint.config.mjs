// Flat ESLint config. Lints this repo's own JavaScript (.mjs) and YAML to keep the
// reusable-workflow + action source consistent. Formatting is delegated to Prettier:
// eslint-config-prettier (JS) and yml/prettier (YAML) switch off every stylistic rule, so
// ESLint owns correctness and Prettier owns layout — they never fight.
//
// Shadow's TypeScript (.mts) is intentionally out of scope here; `tsc` covers it
// (see shadow/typecheck.mjs).
import js from "@eslint/js";
import globals from "globals";
import yml from "eslint-plugin-yml";
import prettier from "eslint-config-prettier";

export default [
  {
    ignores: [
      "**/node_modules/**",
      "**/coverage/**",
      "shadow/dist/**",
      "shadow/mirror/**",
    ],
  },

  // JavaScript: every executable script in the repo is ESM running on Node 24.
  {
    files: ["**/*.mjs"],
    languageOptions: {
      ecmaVersion: 2024,
      sourceType: "module",
      globals: globals.node,
    },
    rules: {
      ...js.configs.recommended.rules,
      "no-var": "error",
      "prefer-const": "error",
      eqeqeq: ["error", "always"],
      "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    },
  },

  // YAML: workflows and action definitions.
  ...yml.configs["flat/standard"],
  ...yml.configs["flat/prettier"],

  // Last word: turn off any remaining JS formatting rules so Prettier is authoritative.
  prettier,
];
