import tags from './tags';

/**
 * @type {NewPostConfig} - Object with ID names of HTML Elements to create new post on main page.
 * @property input - Post title.
 * @property button - Post button.
 * @property message - Div that will get swapped and contains form validation error message.
 * @property clear - Clear text button.
 */
type NewPostConfig = {
	input: string;
	button: string;
	message: string;
	clear: string;
};

/**
 * @type {NewPostElements} - Object with ID names of HTML Elements to create new post on main page.
 * @property input - Post title.
 * @property button - Post button.
 * @property message - Div that will get swapped and contains form validation error message.
 * @property clear - Clear text button.
 */
type NewPostElements = {
	input: HTMLInputElement;
	button: HTMLButtonElement;
	message: HTMLDivElement;
	clear: HTMLButtonElement;
};

const defaultNewPostConfig: NewPostConfig = {
	input: 'post-title',
	button: 'post-button',
	message: 'post-form-message',
	clear: 'input-clear-button',
};

(function initClearButtonListeners(newPostConfig = defaultNewPostConfig) {
	const newPostElements: NewPostElements = {
		input: document.getElementById(newPostConfig.input) as HTMLInputElement,
		button: document.getElementById(newPostConfig.button) as HTMLButtonElement,
		message: document.getElementById(newPostConfig.message) as HTMLDivElement,
		clear: document.getElementById(newPostConfig.clear) as HTMLButtonElement,
	};
	if (newPostElements.input) {
		showInputCancelButton();
	}

	if (newPostElements.clear) {
		clearInput();
	}
})();

/**
 * Show/hide clear button.
 * @param {NewPostConfig} newPostConfig - HTML IDs.
 */
function showInputCancelButton(newPostConfig: NewPostConfig = defaultNewPostConfig) {
	const input = document.getElementById(newPostConfig.input) as HTMLInputElement;
	const clear = document.getElementById(newPostConfig.clear) as HTMLButtonElement;

	// Keyup is better than keydown. For the latter, user needs to do 1 more backspace on empty field to remove clear button.

	if (input && clear) {
		input.addEventListener('keyup', (evt) => {
			if ((<HTMLInputElement>evt.target)?.value.length > 0) {
				clear?.classList.remove('hidden');
			} else if ((<HTMLInputElement>evt.target)?.value.length === 0) {
				clear?.classList.add('hidden');
			}
		});
	}
}

/**
 * Button to clear input.
 * @param {NewPostConfig} newPostConfig - HTML IDs.
 */
function clearInput(newPostConfig = defaultNewPostConfig) {
	const input = document.getElementById(newPostConfig.input) as HTMLInputElement;
	const clear = document.getElementById(newPostConfig.clear) as HTMLButtonElement;
	const message = document.getElementById(newPostConfig.message) as HTMLDivElement;
	clear?.addEventListener('click', () => {
		input.value = '';
		clear?.classList.add('hidden');
		message.innerHTML = '';
		input?.focus();
	});
}

// Disable button if special characters are detected

window.addEventListener('load', () => {
	BlockSpecialChars();
});

window.addEventListener('htmx:afterSwap', () => {
	BlockSpecialChars();
});

const regex = /^[A-Za-z0-9 _!.$/\\|()[\]=`{}<>?@#%^&*—,:;'"+\-"]+$/;

/**
 * Client-side validation for new post input text and prevents submission.
 * @param {NewPostConfig} newPostConfig - HTML IDs.
 */
function BlockSpecialChars(newPostConfig: NewPostConfig = defaultNewPostConfig) {
	const newPostElements: NewPostElements = {
		input: document.getElementById(newPostConfig.input) as HTMLInputElement,
		button: document.getElementById(newPostConfig.button) as HTMLButtonElement,
		message: document.getElementById(newPostConfig.message) as HTMLDivElement,
		clear: document.getElementById(newPostConfig.clear) as HTMLButtonElement,
	};
	if (newPostElements.input && newPostElements.button && newPostElements.message) {
		newPostElements.input.addEventListener('keyup', () => {
			if (!regex.test(newPostElements.input.value) && newPostElements.input.value.length > 0) {
				newPostElements.button.disabled = true;
				newPostElements.message.classList.remove('hidden');
				newPostElements.input.classList.remove('input-accent');
				newPostElements.input.classList.add('input-error');
				newPostElements.message.innerText = 'Invalid characters found. Please use A-Z, a-z, 0-9, and standard symbols.';
			} else {
				newPostElements.message.classList.add('hidden');
				newPostElements.input.classList.remove('input-error');
				newPostElements.input.classList.add('input-accent');
				if (newPostElements.button.disabled) {
					newPostElements.button.removeAttribute('disabled');
				}
			}
		});
	}
}

/** Init tags for new post tags input in drawer. */
tags();
