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
				'slide-down-up': 'slide-down-up 4.8s ease-out 0s 1 forwards',
				'highlight-border': 'highlight-border 4s linear 0s 1 forwards',
				'highlight-comment-main': 'highlight-comment-main 4s steps(1) 0s 1 forwards',
				'highlight-comment-side': 'highlight-comment-side 4s steps(1) 0s 1 forwards',
				'slide-down': 'slide-down 0.4s ease-out 0s 1 forwards',
				'drawer-slide-down': 'drawer-slide-down 0.4s ease-out 0s 1 forwards',
				'drawer-slide-up': 'drawer-slide-up 0.4s ease-out 0s 1 forwards',
				'delete-slide-right': 'delete-slide-right 1s ease-out 0s 1 forwards',
				wiggle: 'wiggle 1.2s linear 0s infinite forwards',
			},
			keyframes: {
				'drawer-slide-down': {
					'0%': {
						display: 'none',
						opacity: '20%',
						transform: 'translateY(-0.5rem)',
					},
					'100%': {
						display: 'inline',
						opacity: '100%',
						transform: 'translateY(0rem)',
					},
				},
				'drawer-slide-up': {
					'0%': { opacity: '100%', transform: 'translateY(0rem)' },
					'100%': { opacity: '60%', transform: 'translateY(-0.5rem)' },
				},
				'delete-slide-right': {
					'100%': {
						transform: 'translateX(1rem) translate3d(0, 0, 0)',
						opacity: '5%',
					},
				},
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
				'slide-down-up': {
					'0%': {
						transform: 'translateY(-0.5rem) translateX(-50%)',
						opacity: '5%',
					},
					'10%': {
						transform: 'translateY(0.5rem) translateX(-50%)',
						opacity: '100%',
					},
					'90%': {
						transform: 'translateY(0.5rem) translateX(-50%)',
						opacity: '100%',
					},
					'100%': {
						transform: 'translateY(-0.5rem) translateX(-50%)',
						opacity: '5%',
					},
				},
				'highlight-border': {
					'5%, 95%': {
						border: '1px solid rgba(234,88,12, 0.3)',
					},
					'0, 100%': {
						border: '1px solid rgba(38, 41, 49, 0.3)',
					},
				},
				'highlight-comment-main': {
					'0%': {
						backgroundColor: '#feede4',
					},
					'100%': { backgroundColor: 'rgba(255,255,255, 0.7)' }, //rgba(190, 242, 100, 0.1)
				},
				'highlight-comment-side': {
					'0%': {
						backgroundColor: '#f9ae86',
					},
					'100%': { backgroundColor: 'rgba(151, 220, 66, 0.3)' },
				},
				wiggle: {
					'0%, 100%': {
						transform: 'rotate(0deg)',
					},
					'25%': {
						transform: 'rotate(20deg)',
					},
					'75%': {
						transform: 'rotate(-20deg)',
					},
				},
				'wiggle-alt': {
					'0%': {
						transform: 'skewX(12deg)',
					},
					'10%': {
						transform: 'skewX(-11deg)',
					},
					'20%': {
						transform: 'skewX(10deg)',
					},
					'30%': {
						transform: 'skewX(-9deg)',
					},
					'40%': {
						transform: 'skewX(8deg)',
					},
					'50%': {
						transform: 'skewX(-7deg)',
					},
					'60%': {
						transform: 'skewX(6deg)',
					},
					'70%': {
						transform: 'skewX(-5deg)',
					},
					'80%': {
						transform: 'skewX(4deg)',
					},
					'90%': {
						transform: 'skewX(3deg)',
					},
					'100%': {
						transform: 'skewX(3deg)',
					},
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
					secondary: '#2a9c56', //'#35c46c', //'#1c64d9', //https://mycolor.space/?hex=%2397DC42&sub=1
					'secondary-content': '#ffffff', //'#d2e1fa',
					accent: '#1b510f', //'#4de42c',
					'accent-content': '#fdfdfd', //'#021201',
					neutral: '#262931', //'#B8B8B3',
					'neutral-content': '#f0f0f1', //'#cfd0d2',
					'base-100': '#f8f6eb',
					'base-200': '#d8d6cc',
					'base-300': '#b8b7ae',
					'base-content': '#151513',
					info: '#2563EB',
					'info-content': '#fdfdfd', //'#d2e2ff',
					success: '#28A528', //'#16A34A',
					'success-content': 'white', //'#ffd9d4', //'#000a02',
					warning: '#D97706',
					'warning-content': '#110500',
					error: '#DC2626',
					'error-content': '#fdfdfd', //'#ffd9d4',
				},
			},
		],
	},
};
