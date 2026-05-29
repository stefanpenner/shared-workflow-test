// Flat ESLint config. Lints this repo's own JavaScript (.mjs) and TypeScript (.mts) and YAML.
// ESLint owns correctness + the import allowlist; Prettier owns layout (eslint-config-prettier /
// yml-prettier switch off stylistic rules so they never fight); `tsc` (shadow/typecheck.mjs) owns
// .mts types.
import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import yml from "eslint-plugin-yml";
import prettier from "eslint-config-prettier";

// Module allowlist (policy): only node: built-ins, relative paths, `yaml`, and `@actions/*` may be
// imported. The regex matches any specifier NOT on the allowlist (negative lookahead) and denies it.
const importAllowlist = {
  "no-restricted-imports": [
    "error",
    {
      patterns: [
        {
          regex: "^(?!(node:|@actions/|yaml($|/)|\\.{1,2}/))",
          message:
            "import not on the allowlist — only node:*, relative paths, yaml, and @actions/* are allowed",
        },
      ],
    },
  ],
};

const sharedRules = {
  "no-var": "error",
  "prefer-const": "error",
  eqeqeq: ["error", "always"],
  ...importAllowlist,
};

export default [
  {
    ignores: ["**/node_modules/**", "**/coverage/**", "shadow/dist/**", "shadow/mirror/**"],
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
      ...sharedRules,
      "no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    },
  },

  // TypeScript: shadow/'s .mts (run natively on Node 24). Lint + import allowlist only — `tsc` owns types.
  {
    files: ["**/*.mts"],
    languageOptions: {
      parser: tseslint.parser,
      ecmaVersion: 2024,
      sourceType: "module",
      globals: globals.node,
    },
    plugins: { "@typescript-eslint": tseslint.plugin },
    rules: {
      ...sharedRules,
      "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
    },
  },

  // Dev-tooling config is exempt from the source import allowlist — it legitimately imports the
  // dev-only linters/formatters (eslint/prettier/typescript-eslint). The allowlist governs source.
  {
    files: ["eslint.config.mjs"],
    rules: { "no-restricted-imports": "off" },
  },

  // YAML: workflows and action definitions.
  ...yml.configs["flat/standard"],
  ...yml.configs["flat/prettier"],
  {
    files: ["**/*.{yaml,yml}"],
    rules: {
      // GitHub Actions event triggers (`pull_request:`, `workflow_dispatch:`) are
      // intentionally empty mappings — that's the idiom, not a mistake.
      "yml/no-empty-mapping-value": "off",
    },
  },

  // Last word: turn off any remaining JS formatting rules so Prettier is authoritative.
  prettier,
];
