import globals from 'globals';
import pluginJs from '@eslint/js';

/** @type {import('eslint').Linter.Config[]} */
export default [
	{ languageOptions: { globals: globals.browser } },
	{ ignores: ['static/js/output/', 'static/js/ext/', '**/**.config.js'] },
	pluginJs.configs.recommended,
];
