module.exports = {
  extends: ['standard', 'plugin:n/recommended', 'prettier'],
  rules: {
    'no-var': 'error',
    'prefer-const': 'error',
    'n/no-process-exit': 'off',
    'camelcase': 'off',
  },
}
