/**
 * @param {HTMLInputElement} inputEl - the HTML input element
 * @param {HTMLButtonElement} buttonEl - the button to submit
 * @param {('xs'|'sm'|'md'|'lg')} loaderSize - size of the loading spinner
 * @param {HTMLFormElement|null} formEl - form element to check if htmx:afterRequest originated from, used to get successful post even when there's no swap (i.e. SSE mode)
 * @param {('input'|'textarea')} elType - controls whether enter button should change loading icon - input should but textarea shouldn't
 */
export function keyboardShortcut(inputEl, buttonEl, loaderSize = 'xs', formEl = null, elType = 'input') {
	// Save the inner HTML first.
	const buttonHTML = buttonEl.innerHTML;

	//This is keydown, so it's faster than the keyup submit hx-trigger on the form
	inputEl.addEventListener('keydown', (evt) => {
		if (elType == 'input') {
			if (!evt.ctrlKey && evt.key === 'Enter') {
				console.log('Received signal, changing to spinner');
				buttonEl.innerHTML = `<span class="loading loading-spinner loading-${loaderSize}"></span>`;
			}
		}

		if (evt.ctrlKey && evt.key === 'Enter') {
			console.log('Received signal, changing to spinner');
			buttonEl.innerHTML = `<span class="loading loading-spinner loading-${loaderSize}"></span>`;
		}
	});

	buttonEl.addEventListener('click', () => {
		console.log('Received signal, changing to spinner');
		buttonEl.innerHTML = `<span class="loading loading-spinner loading-${loaderSize}"></span>`;
	});

	window.addEventListener('htmx:afterSwap', (evt) => {
		if (evt.detail.elt === inputEl) {
			buttonEl.innerHTML = buttonHTML; //'Add Comment';
		}
	});

	window.addEventListener('htmx:validation:failed', () => {
		console.log('Changing back to text button');
		setTimeout(() => {
			buttonEl.innerHTML = buttonHTML; //'Add Comment';
		}, 1);
	});

	if (formEl) {
		window.addEventListener('htmx:afterRequest', (evt) => {
			if (evt.detail.elt === formEl) {
				setTimeout(() => {
					buttonEl.innerHTML = buttonHTML;
				}, 200);
			}
		});
	}
}
