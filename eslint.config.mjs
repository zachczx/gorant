import eslintConfigPrettier from 'eslint-config-prettier';
import globals from 'globals';
import js from '@eslint/js';
import jsdoc from 'eslint-plugin-jsdoc';

/** @type {import('eslint').Linter.Config[]} */
export default [
	{ languageOptions: { globals: globals.browser } },
	{ ignores: ['static/js/output/', 'static/js/ext/', '**/**.config.js'] },
	js.configs.recommended,
	{
		files: ['**/*.js'],
		plugins: { jsdoc: jsdoc },
		rules: {
			'func-style': ['warn', 'declaration'],
			'no-unused-vars': ['error', { args: 'after-used' }],
			'jsdoc/require-description': 'warn',
			'jsdoc/check-values': 'error',
			'jsdoc/tag-lines': ['warn', 'never'],
			'jsdoc/require-jsdoc': [
				'warn',
				{
					require: {
						FunctionDeclaration: true,
						MethodDefinition: true,
						ClassDeclaration: true,
						ArrowFunctionExpression: true,
					},
				},
			],
		},
	},
	jsdoc.configs['flat/recommended'],
	eslintConfigPrettier,
];
