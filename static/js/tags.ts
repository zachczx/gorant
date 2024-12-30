import { keyboardShortcut } from './common';

/**
 * For UI supporting Tag's tag-style input
 *
 * @exports tagsUi
 */

/**
 * @typedef {object} TagsConfig - Object with the ids of the HTML Elements to be used in this func.
 * @property {HTMLFormElement} form - Form element with hx-post
 * @property {HTMLInputElement} input - Form input for user to key in tags
 * @property {HTMLUListElement} list - Form input for user to key in tags
 * @property {HTMLButtonElement} saveButton - Form button
 * @property {HTMLInputElement} data - Hidden form input that contains all the tags for form submission
 */

/**
 * Default configs for tags-related functions.
 *
 * @type {TagsConfig} TagsConfig - Object with the ids of the HTML Elements to be used in this func.
 */
const defaultTagsConfig = {
	form: 'tags-container',
	input: 'tags-input',
	list: 'tags-list',
	saveButton: 'tags-save-button',
	data: 'tags-data',
};

/**
 * Check if HTML elements taken from argument ID names are in the DOM.
 *
 * @param {TagsConfig} tagsConfig - Object with HTML ID names.
 * @returns {boolean} - True if HTML elements are found in the DOM, typically for after hx-swap of tags form completes.
 */
function checkDomForTagsEls(tagsConfig = defaultTagsConfig) {
	for (let el in tagsConfig) {
		if (!document.getElementById(tagsConfig[el])) {
			return false;
		}
	}
	return true;
}

/**
 * Functionality for tags (posting, editing and deleting) in the post page
 *
 * @param {TagsConfig} tagsConfig - Object with the ids of the HTML Elements to be used in this func.
 */
function tagsUi(tagsConfig = defaultTagsConfig) {
	const tagsElements = {};
	for (const name in tagsConfig) {
		tagsElements[name] = document.getElementById(tagsConfig[name]);
	}

	/** @type {string[]} Styling Classes to add for each tag */
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

	let margin = 0;

	tagsElements.input.focus();

	fetchTagsFromHiddenFormField(tagsElements.data, tagsElements.list, classes);

	tagsElements.input.addEventListener('keydown', (evt) => {
		if (
			(evt.key === 'Enter' && !evt.ctrlKey) ||
			evt.key === ',' ||
			evt.key === ' ' ||
			evt.key === ';' ||
			evt.key === '.' ||
			(evt.key === 'Tab' && tagsElements.input.value.length > 0)
		) {
			evt.preventDefault();
			const tags = tagsElements.input.value.split(',');
			for (let i = 0; i < tags.length; i++) {
				tags[i] = tags[i].trim().toLowerCase();
				if (tags[i].length > 0) {
					const el = document.createElement('li');
					el.id = `user-tag-${tags[i]}`;
					el.classList.add(...classes);
					el.innerHTML = tags[i];
					tagsElements.list.appendChild(el);
					if (tagsElements.data.value.length === 0) {
						tagsElements.data.value = tags[i];
					} else {
						tagsElements.data.value = `${tagsElements.data.value},${tags[i]}`;
					}
				}
			}
			tagsElements.input.value = '';

			if (tagsElements.list.childElementCount > 0) {
				margin = `me-${String(tagsElements.list.childElementCount * 2)}`;
				tagsElements.list.classList.add(margin);
			}
		}
	});

	tagsElements.list.addEventListener('click', (evt) => {
		if (evt.target.id.includes('user-tag-')) {
			evt.preventDefault();
			// console.log('Clicked a tag');
			tagsElements.list.removeChild(evt.target);
			if (!tagsElements.data.value.includes(',')) {
				tagsElements.data.value = tagsElements.data.value.replace(evt.target.innerText, '');
			} else if (
				tagsElements.data.value.includes(',') &&
				tagsElements.data.value.includes(`${evt.target.innerText},`)
			) {
				tagsElements.data.value = tagsElements.data.value.replace(`${evt.target.innerText},`, '');
			} else if (
				tagsElements.data.value.includes(',') &&
				tagsElements.data.value.includes(`,${evt.target.innerText}`)
			) {
				tagsElements.data.value = tagsElements.data.value.replace(`,${evt.target.innerText}`, '');
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
		if (evt.detail.elt === tagsElements.form) {
			if (tagsElements.input.value.length > 0) {
				if (tagsElements.data.value.length > 0) {
					tagsElements.data.value = `${tagsElements.data.value},${tagsElements.input.value}`;
					evt.detail.parameters['tags-data'] = tagsElements.data.value;
				} else {
					tagsElements.data.value = tagsElements.input.value;
					evt.detail.parameters['tags-data'] = tagsElements.data.value;
				}
			}
		}
		console.log('tags data', evt.detail.parameters['tags-data']);
	});

	//Keyboard shortcuts for tags UI
	keyboardShortcut(tagsElements.input, tagsElements.saveButton, tagsElements.form, undefined, 'textarea');
}

/**
 * Fetch tags from hidden form value if there already are tags there
 *
 * @param {HTMLInputElement} data - Hidden form input that contains all the tags for form submission
 * @param {HTMLUListElement} list - Button style tags under the input field
 * @param {string[]} classes - Styling Classes to add for each tag
 */
function fetchTagsFromHiddenFormField(data, list, classes) {
	if (data.value.length > 0) {
		const tags = data.value.split(',');
		for (let i = 0; i < tags.length; i++) {
			tags[i] = tags[i].trim().toLowerCase();
			if (tags[i].length > 0) {
				const el = document.createElement('li');
				el.id = `user-tag-${tags[i]}`;
				el.classList.add(...classes);
				el.innerHTML = tags[i];
				list.appendChild(el);
			}
		}
	}
}

/**
 * Init tags posting/editing/deleting user functionality This uses Window eventListener because
 * document.addEventListener() is unreliable
 *
 * @param {TagsConfig} tagsConfig - Object with the ids of the HTML Elements to be used in this func.
 */
export default function tags(tagsConfig = defaultTagsConfig) {
	window.addEventListener('load', () => {
		if (checkDomForTagsEls()) {
			tagsUi(tagsConfig);
		}
	});

	document.addEventListener('htmx:afterSwap', () => {
		if (checkDomForTagsEls()) {
			tagsUi(tagsConfig);
		}
	});
}
