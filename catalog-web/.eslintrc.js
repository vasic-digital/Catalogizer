module.exports = {
  root: true,
  env: { browser: true, es2020: true },
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended',
  ],
  ignorePatterns: ['dist', '.eslintrc.js'],
  parser: '@typescript-eslint/parser',
  settings: {
    react: {
      version: 'detect',
    },
  },
  rules: {
    'react/react-in-jsx-scope': 'off',
    // Disable prop-types - TypeScript handles type checking
    'react/prop-types': 'off',
    // Allow display-name rule to be a warning (common with arrow functions and memo)
    'react/display-name': 'warn',
    // Allow any type in certain cases (can be tightened later)
    '@typescript-eslint/no-explicit-any': 'warn',
    // Allow unused vars that start with underscore
    '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
    // Allow empty functions in certain cases (common in tests and handlers)
    '@typescript-eslint/no-empty-function': 'warn',
    // Allow inferrable types (helpful for documentation but not required)
    '@typescript-eslint/no-inferrable-types': 'off',
  },
}