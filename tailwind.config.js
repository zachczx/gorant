/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ['./templates/*.{templ,txt}', './static/js/output/*.js'],
	theme: {
		fontFamily: {
			sans: '"Inter Variable"',
			mono: '"Fira Code Variable"',
		},
		extend: {
			animation: {
				'slide-up-down': 'slide-up-down 4.8s ease-out 0s 1 forwards',
				'highlight-border': 'highlight-border 4s linear 0s 1 forwards',
				'highlight-comment-main': 'highlight-comment-main 4s steps(1) 0s 1 forwards',
				'highlight-comment-side': 'highlight-comment-side 4s steps(1) 0s 1 forwards',
				'slide-down': 'slide-down 0.4s ease-out 0s 1 forwards',
			},
			keyframes: {
				'slide-down': {
					'0%': {
						transform: 'translateY(-0.2rem)',
						opacity: '0%',
					},
					'100%': {
						transform: 'translateY(0rem)',
						opacity: '100%',
					},
				},
				'slide-up-down': {
					'0%': {
						transform: 'translateY(0.7rem)',
						opacity: '5%',
					},
					'10%': {
						transform: 'translateY(-0.7rem)',
						opacity: '100%',
					},
					'90%': {
						transform: 'translateY(-0.7rem)',
						opacity: '100%',
					},
					'100%': {
						transform: 'translateY(0.7rem)',
						opacity: '5%',
					},
				},
				'highlight-border': {
					'5%, 95%': {
						border: '3px solid #ea580c',
					},
					'0, 100%': {
						border: '1px solid rgba(38, 41, 49, 0.3)',
					},
				},
				'highlight-comment-main': {
					'0%': {
						backgroundColor: '#feede4',
					},
					'100%': { backgroundColor: 'rgba(190, 242, 100, 0.1)' },
				},
				'highlight-comment-side': {
					'0%': {
						backgroundColor: '#f9ae86',
					},
					'100%': { backgroundColor: 'rgba(151, 220, 66, 0.3)' },
				},
			},
		},
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: [
			{
				custom: {
					primary: '#97dc42',
					'primary-content': '#081101',
					secondary: '#35c46c', //'#1c64d9', //https://mycolor.space/?hex=%2397DC42&sub=1
					'secondary-content': '#d2e1fa',
					accent: '#1b510f', //'#4de42c',
					'accent-content': '#fdfdfd', //'#021201',
					neutral: '#262931',
					'neutral-content': '#cfd0d2',
					'base-100': '#f8f6eb',
					'base-200': '#d8d6cc',
					'base-300': '#b8b7ae',
					'base-content': '#151513',
					info: '#2563EB',
					'info-content': '#d2e2ff',
					success: '#28A528', //'#16A34A',
					'success-content': 'white', //'#ffd9d4', //'#000a02',
					warning: '#D97706',
					'warning-content': '#110500',
					error: '#DC2626',
					'error-content': '#fdfdfd', //'#ffd9d4',
				},
				// cyberpunk: {
				// 	...require('daisyui/src/theming/themes')['cyberpunk'],
				// },
			},
		],
	},
};
