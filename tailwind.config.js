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
				cyberpunk: {
					...require('daisyui/src/theming/themes')['cyberpunk'],
				},
			},
		],
	},
};
