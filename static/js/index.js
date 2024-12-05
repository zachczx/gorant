import tags from './tags';

// Input cancel button

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

const regex = /^[A-Za-z0-9 _!.$\/\\|()[\]=`{}<>?@#%^&*—,:;'"+\-"]+$/;

function BlockSpecialChars() {
	const inputEl = document.getElementById('post-title');
	const postButton = document.getElementById('post-button');
	const postFormMessage = document.getElementById('post-form-message');

	if (inputEl && postButton && postFormMessage) {
		inputEl.addEventListener('keyup', () => {
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

tags();
