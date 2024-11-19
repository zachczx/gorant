// import { gsap } from 'gsap/dist/gsap';

(function createnewFocus() {
	const elementIds = ['navbar', 'logo', 'content', 'footer'];
	const classList = ['blur-sm'];
	const addNewDiv = document.getElementById('add-new');
	addNewDiv.addEventListener('click', (evt) => {
		console.log('triggered by', evt.target);
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
		if (
			e.target.id !== 'post-form' &&
			e.target.id !== 'post-id' &&
			e.target.id !== 'post-button' &&
			!e.target.id.includes('add-new')
		) {
			document.getElementById('post-button').innerText = 'Go';
			for (let i = 0; i < elementIds.length; i++) {
				document.getElementById(elementIds[i]).classList.remove(...classList);
			}
		}
	});
})();

// (function BlockSpecialChars() {
// 	console.log('triggered!');
// 	const pattern = /^[A-Za-z0-9_-]+$/;
// 	const inputEl = document.getElementById('post-id');
// 	const postButton = document.getElementById('post-button');
// 	const postFormMessage = document.getElementById('post-form-message');

// 	inputEl.addEventListener('keyup', () => {
// 		if (!pattern.test(inputEl.value) && inputEl.value.length > 0) {
// 			postButton.disabled = 'true';
// 			postFormMessage.classList.remove('hidden');
// 			inputEl.classList.remove('input-accent');
// 			inputEl.classList.add('input-error');
// 			postFormMessage.innerText = 'No special characters allowed! ID may contain only A-Z, a-z, 0-9, dash, underscore.';
// 		} else {
// 			postFormMessage.classList.add('hidden');
// 			inputEl.classList.remove('input-error');
// 			inputEl.classList.add('input-accent');
// 			postButton.removeAttribute('disabled');
// 		}
// 	});
// })();

// function test() {
// 	console.log('New func triggered!');
// }
