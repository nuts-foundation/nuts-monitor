const js = require('@eslint/js');
const pluginVue = require('eslint-plugin-vue');
const globals = require('globals');

module.exports = [
  js.configs.recommended,
  ...pluginVue.configs['flat/essential'],
  {
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
    },
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
];
