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

// Disable button if special characters are detected

window.addEventListener('load', BlockSpecialChars);

window.addEventListener('htmx:afterSwap', BlockSpecialChars);

const regex = /^[A-Za-z0-9 _!.$\/\\|()[\]=`{}<>?@#%^&*—:;'"+\-"]+$/;

function BlockSpecialChars() {
	const inputEl = document.getElementById('post-title');
	const postButton = document.getElementById('post-button');
	const postFormMessage = document.getElementById('post-form-message');

	if (inputEl && postButton && postFormMessage) {
		inputEl.addEventListener('keyup', () => {
			console.log('triggered');
			if (!regex.test(inputEl.value) && inputEl.value.length > 0) {
				postButton.disabled = 'true';
				postFormMessage.classList.remove('hidden');
				inputEl.classList.remove('input-accent');
				inputEl.classList.add('input-error');
				postFormMessage.innerText = 'Invalid characters found. Please use A-Z, a-z, 0-9, and standard symbols.';
			} else {
				postFormMessage.classList.add('hidden');
				inputEl.classList.remove('input-error');
				inputEl.classList.add('input-accent');
				if (postButton.disabled) {
					postButton.removeAttribute('disabled');
				}
			}
		});
	}
}

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
	let margin = 0;

	const classes = [
		'user-tag',
		'btn',
		'bg-primary/50',
		'text-neutral/70',
		'border-0',
		'hover:bg-primary/60',
		'btn-sm',
		'text-md',
		'me-2',
		'my-2',
	];

	tagsInput.addEventListener('keyup', (evt) => {
		if (evt.key === 'Enter' || evt.key === ',' || evt.key === ' ' || evt.key === ';' || evt.key === '.') {
			evt.preventDefault();
			let tags = tagsInput.value.split(',');
			for (let i = 0; i < tags.length; i++) {
				tags[i] = tags[i].trim().toLowerCase();
				if (tags[i].length > 0) {
					let el = document.createElement('li');
					el.id = 'user-tag-' + tags[i];
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

			if (tagsList.childElementCount > 0) {
				margin = 'me-' + String(tagsList.childElementCount * 2);
				tagsList.classList.add(margin);
			}
		}
	});

	tagsList.addEventListener('click', (evt) => {
		if (evt.target.id.includes('user-tag-')) {
			console.log('Clicked a tag');
			tagsList.removeChild(evt.target);
			if (!tagsData.value.includes(',')) {
				tagsData.value = tagsData.value.replace(evt.target.innerText, '');
			} else if (tagsData.value.includes(',') && tagsData.value.includes(evt.target.innerText + ',')) {
				tagsData.value = tagsData.value.replace(evt.target.innerText + ',', '');
			} else if (tagsData.value.includes(',') && tagsData.value.includes(',' + evt.target.innerText)) {
				tagsData.value = tagsData.value.replace(',' + evt.target.innerText, '');
			}
		}
	});
}
