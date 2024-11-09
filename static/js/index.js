// import { gsap } from 'gsap/dist/gsap';

(function createnewFocus() {
	const elementIds = ['navbar', 'logo', 'content', 'footer'];
	const classList = ['blur-md', 'brightness-75'];
	const addNewDiv = document.getElementById('add-new');
	addNewDiv.addEventListener('click', () => {
		document.getElementById('post-button').innerText = 'Create';
		const postIdInputEl = document.getElementById('post-id');
		const classes = [
			'active:border-4',
			'focus:border-2',
			'focus:ring-2',
			'focus:ring-secondary',
			'focus:ring-offset-4',
			'focus:border-primary',
			'active:border-primary',
		];
		postIdInputEl.classList.add(...classes);
		postIdInputEl.focus();

		// Blur effect

		for (let i = 0; i < elementIds.length; i++) {
			const el = document.getElementById(elementIds[i]);
			el.classList.add(...classList);
		}
	});

	document.addEventListener('click', (e) => {
		if (e.target.id !== 'post-form' && e.target.id !== 'post-button' && !e.target.id.includes('add-new')) {
			document.getElementById('post-button').innerText = 'Go';
			for (let i = 0; i < elementIds.length; i++) {
				document.getElementById(elementIds[i]).classList.remove(...classList);
			}
		}
	});
})();
