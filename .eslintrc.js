module.exports = {
  extends: ['standard', 'plugin:n/recommended', 'prettier'],
  parserOptions: { ecmaVersion: 2022 },
  rules: {
    'no-var': 'error',
    'prefer-const': 'error',
    'n/no-process-exit': 'off',
    'camelcase': 'off',
  },
}
