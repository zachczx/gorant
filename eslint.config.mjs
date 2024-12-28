import globals from 'globals';
import js from '@eslint/js';
import eslintConfigPrettier from 'eslint-config-prettier';

/** @type {import('eslint').Linter.Config[]} */
export default [
	{ languageOptions: { globals: globals.browser } },
	{ ignores: ['static/js/output/', 'static/js/ext/', '**/**.config.js'] },
	js.configs.recommended,
	eslintConfigPrettier,
];
