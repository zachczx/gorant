import { keyboardShortcut } from './common';

// For UI supporting Tag's tag-style input

//Window is used because document.addEventListener() is unreliable

export default function tags() {
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
}

function checkDomForTagsEls() {
	if (
		document.getElementById('tags-input') &&
		document.getElementById('tags-list') &&
		document.getElementById('tags-data') &&
		document.getElementById('tags-save-button') &&
		document.getElementById('tags-container')
	) {
		return true;
	}
	return false;
}

function tagsUi() {
	console.log('Loading tagsUI');
	const tagsInput = document.getElementById('tags-input');
	const tagsList = document.getElementById('tags-list');
	const tagsSaveButton = document.getElementById('tags-save-button');
	const tagsForm = document.getElementById('tags-container');
	const tagsData = document.getElementById('tags-data');
	let margin = 0;

	tagsInput.focus();

	// Classes to add for each tag
	const classes = [
		'user-tag',
		'btn',
		'bg-secondary',
		'text-secondary-content',
		'border-0',
		'hover:bg-secondary',
		'btn-sm',
		'text-md',
		'me-2',
		'my-1',
	];

	// Fetch tags from hidden form value if there already are tags there
	if (tagsData.value.length > 0) {
		const tags = tagsData.value.split(',');
		for (let i = 0; i < tags.length; i++) {
			tags[i] = tags[i].trim().toLowerCase();
			if (tags[i].length > 0) {
				const el = document.createElement('li');
				el.id = `user-tag-${tags[i]}`;
				el.classList.add(...classes);
				el.innerHTML = tags[i];
				tagsList.appendChild(el);
			}
		}
	}

	tagsInput.addEventListener('keydown', (evt) => {
		if (
			(evt.key === 'Enter' && !evt.ctrlKey) ||
			evt.key === ',' ||
			evt.key === ' ' ||
			evt.key === ';' ||
			evt.key === '.' ||
			(evt.key === 'Tab' && tagsInput.value.length > 0)
		) {
			evt.preventDefault();
			const tags = tagsInput.value.split(',');
			for (let i = 0; i < tags.length; i++) {
				tags[i] = tags[i].trim().toLowerCase();
				if (tags[i].length > 0) {
					const el = document.createElement('li');
					el.id = `user-tag-${tags[i]}`;
					el.classList.add(...classes);
					el.innerHTML = tags[i];
					tagsList.appendChild(el);
					if (tagsData.value.length === 0) {
						tagsData.value = tags[i];
					} else {
						tagsData.value = `${tagsData.value},${tags[i]}`;
					}
				}
			}
			tagsInput.value = '';

			if (tagsList.childElementCount > 0) {
				margin = `me-${String(tagsList.childElementCount * 2)}`;
				tagsList.classList.add(margin);
			}
		}
	});

	tagsList.addEventListener('click', (evt) => {
		if (evt.target.id.includes('user-tag-')) {
			evt.preventDefault();
			console.log('Clicked a tag');
			tagsList.removeChild(evt.target);
			if (!tagsData.value.includes(',')) {
				tagsData.value = tagsData.value.replace(evt.target.innerText, '');
			} else if (tagsData.value.includes(',') && tagsData.value.includes(`${evt.target.innerText},`)) {
				tagsData.value = tagsData.value.replace(`${evt.target.innerText},`, '');
			} else if (tagsData.value.includes(',') && tagsData.value.includes(`,${evt.target.innerText}`)) {
				tagsData.value = tagsData.value.replace(`,${evt.target.innerText}`, '');
			}
		}
	});

	// This helps when user enters into input field, but doesn't press any of the triggers to add the value to the hidden field.
	// This does a last check to add all remaining input in the field before posting it
	//
	// Note: Seems like just the click event precedes the request, so this doesn't require evt.preventDefault()
	// Note: However, I added a delay of 100ms to hx-trigger just to be safe
	window.addEventListener('htmx:configRequest', (evt) => {
		// htmx:configRequest triggers after htmx collected params - https://htmx.org/events/#htmx:configRequest
		// htmx:beforeRequest does not change params.
		// Alternative to listening to configRequest is to add eventListeners directly to the button or listen for keypresses ctrl+enter, which is tedious
		if (evt.detail.elt === tagsForm) {
			if (tagsInput.value.length > 0) {
				if (tagsData.value.length > 0) {
					tagsData.value = `${tagsData.value},${tagsInput.value}`;
					evt.detail.parameters['tags-data'] = tagsData.value;
					console.log('Existing tagsData found, ', tagsData.value);
				} else {
					tagsData.value = tagsInput.value;
					evt.detail.parameters['tags-data'] = tagsData.value;
					console.log('No existing tagsData, ', tagsData.value);
				}
			}
		}
		console.log('tags data', evt.detail.parameters['tags-data']);
	});

	//Keyboard shortcuts for tags UI
	keyboardShortcut(tagsInput, tagsSaveButton, tagsForm);
}
