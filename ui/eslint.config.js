// Lean, correctness-focused ESLint setup (flat config).
//
// The intent is to catch real bugs -- unused code, unsafe React hook usage,
// obviously wrong TypeScript -- without style bikeshedding. Formatting is left
// to the editor; this config deliberately enables no stylistic rules.
import js from '@eslint/js';
import globals from 'globals';
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';

export default tseslint.config(
  // Build output and generated assets are never linted.
  { ignores: ['dist', 'node_modules', 'coverage'] },

  js.configs.recommended,
  ...tseslint.configs.recommended,

  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: { ...globals.browser, ...globals.node },
    },
    plugins: {
      'react-hooks': reactHooks,
    },
    rules: {
      // Rules of Hooks are genuine correctness constraints in React.
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',

      // The HTTP boundary (axios responses, antd form values, dynamic backend
      // payloads) is intentionally typed as `any`. Banning it here would create
      // churn without catching real bugs, and the request/response contracts
      // are covered by the typed agent and its tests instead.
      '@typescript-eslint/no-explicit-any': 'off',

      // Unused code is a real smell, but allow the conventional underscore
      // opt-out for deliberately-ignored bindings. `caughtErrors: 'none'` keeps
      // the many `catch (error) { message.error(...) }` handlers -- which surface
      // a user-facing message without inspecting the error object -- from being
      // reported; that is a deliberate UX pattern, not dead code.
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrors: 'none',
          ignoreRestSiblings: true,
        },
      ],
    },
  },

  // Config files run in a Node context.
  {
    files: ['*.{js,ts}', 'vite.config.ts'],
    languageOptions: {
      globals: { ...globals.node },
    },
  },
);
