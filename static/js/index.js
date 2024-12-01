// import { gsap } from 'gsap/dist/gsap';

(function showInputCancelButton() {
	document.getElementById('post-title').addEventListener('keydown', (evt) => {
		if (evt.target.value.length > 0) {
			document.getElementById('input-cancel-button').classList.remove('hidden');
		} else if (evt.target.value.length === 0) {
			document.getElementById('input-cancel-button').classList.add('hidden');
		}
	});
})();

(function clearInput() {
	document.getElementById('input-cancel-button').addEventListener('click', () => {
		const input = document.getElementById('post-title');
		input.value = '';
		document.getElementById('input-cancel-button').classList.add('hidden');
		document.getElementById('post-form-message').innerHTML = '';
		input.focus();
	});
})();
// (function createnewFocus() {
// 	const elementIds = ['navbar', 'logo', 'content', 'footer'];
// 	const classList = ['blur-sm'];
// 	const addNewDiv = document.getElementById('add-new');
// 	addNewDiv.addEventListener('click', (evt) => {
// 		console.log('triggered by', evt.target);
// 		document.getElementById('post-button').innerText = 'Create';
// 		const postIdInputEl = document.getElementById('post-id');
// 		const classes = [
// 			'active:border-4',
// 			'focus:border-2',
// 			'focus:ring-2',
// 			'focus:ring-secondary',
// 			'focus:ring-offset-4',
// 			'focus:border-primary',
// 			'active:border-primary',
// 		];
// 		postIdInputEl.classList.add(...classes);
// 		postIdInputEl.focus();

// 		// Blur effect

// 		for (let i = 0; i < elementIds.length; i++) {
// 			const el = document.getElementById(elementIds[i]);
// 			el.classList.add(...classList);
// 		}
// 	});

// 	document.addEventListener('click', (e) => {
// 		if (
// 			e.target.id !== 'post-form' &&
// 			e.target.id !== 'post-id' &&
// 			e.target.id !== 'post-button' &&
// 			!e.target.id.includes('add-new')
// 		) {
// 			document.getElementById('post-button').innerText = 'Go';
// 			for (let i = 0; i < elementIds.length; i++) {
// 				document.getElementById(elementIds[i]).classList.remove(...classList);
// 			}
// 		}
// 	});
// })();

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

// For UI supporting Tag's tag-style input

//Window is used because document.addEventListener() is unreliable
window.addEventListener('load', () => {
	if (checkDomForTagsEls()) {
		tagsUi();
	}
});

document.addEventListener('htmx:afterSwap', () => {
	if (checkDomForTagsEls()) {
		tagsUi();
	}
});

function checkDomForTagsEls() {
	if (
		document.getElementById('tags-input') &&
		document.getElementById('tags-list') &&
		document.getElementById('tags-data')
	) {
		return true;
	} else {
		return false;
	}
}

function tagsUi() {
	const tagsInput = document.getElementById('tags-input');
	const tagsList = document.getElementById('tags-list');
	let tagsData = document.getElementById('tags-data');

	const classes = [
		'btn',
		'btn-outline',
		'border-neutral/70',
		'text-neutral/70',
		'hover:bg-transparent',
		'btn-xs',
		'me-2',
	];

	tagsInput.addEventListener('keyup', (evt) => {
		if (evt.key === 'Enter' || evt.key === ',' || evt.key === ' ' || evt.key === ';' || evt.key === '.') {
			evt.preventDefault();
			let tags = tagsInput.value.split(',');
			for (let i = 0; i < tags.length; i++) {
				tags[i] = tags[i].trim();
				if (tags[i].length > 0) {
					let el = document.createElement('li');
					el.classList.add(...classes);
					el.innerHTML = tags[i];
					tagsList.appendChild(el);

					if (tagsData.value.length === 0) {
						tagsData.value = tags[i];
					} else {
						tagsData.value = tagsData.value + ',' + tags[i];
					}
				}
			}
			tagsInput.value = '';
		}
	});
}
