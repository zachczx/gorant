/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ['./templates/*.{templ,txt}', './static/js/output/*.js'],
	theme: {
		fontFamily: {
			sans: '"Inter Variable"',
			mono: '"Fira Code Variable"',
		},
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: [
			{
				custom: {
					primary: '#97dc42',
					'primary-content': '#081101',
					secondary: '#1c64d9',
					'secondary-content': '#d2e1fa',
					accent: '#4de42c',
					'accent-content': '#021201',
					neutral: '#262931',
					'neutral-content': '#cfd0d2',
					'base-100': '#f8f6eb',
					'base-200': '#d8d6cc',
					'base-300': '#b8b7ae',
					'base-content': '#151513',
					info: '#2563EB',
					'info-content': '#d2e2ff',
					success: '#16A34A',
					'success-content': '#000a02',
					warning: '#D97706',
					'warning-content': '#110500',
					error: '#DC2626',
					'error-content': '#ffd9d4',
				},
				// cyberpunk: {
				// 	...require('daisyui/src/theming/themes')['cyberpunk'],
				// },
			},
		],
	},
};
