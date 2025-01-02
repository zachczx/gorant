/**
 * For Keyboard shortcuts (ctr+enter) to submit form
 *
 * @param {HTMLInputElement} inputEl - The HTML input element
 * @param {HTMLButtonElement} buttonEl - The button to submit
 * @param {'xs' | 'sm' | 'md' | 'lg'} loaderSize - Size of the loading spinner
 * @param {HTMLFormElement | null} formEl - Form element to check if htmx:afterRequest originated from, used to get
 *   successful post even when there's no swap (i.e. SSE mode)
 * @param {'input' | 'textarea'} elType - Controls whether enter button should change loading icon - input should but
 *   textarea shouldn't
 */
export function keyboardShortcut(
	inputEl: HTMLInputElement,
	buttonEl: HTMLButtonElement,
	loaderSize: 'xs' | 'sm' | 'md' | 'lg' = 'xs',
	formEl: HTMLFormElement | null = null,
	elType: 'input' | 'textarea' = 'input',
) {
	// Save the inner HTML first.
	const buttonHTML = buttonEl.innerHTML;

	// This is keydown, so it's faster than the keyup submit hx-trigger on the form
	inputEl.addEventListener('keydown', (evt) => {
		if (elType == 'input') {
			if (!evt.ctrlKey && evt.key === 'Enter') {
				// console.log('Received signal, changing to spinner');
				buttonEl.innerHTML = `<span class="loading loading-spinner loading-${loaderSize}"></span>`;
			}
		}

		if (evt.ctrlKey && evt.key === 'Enter') {
			// console.log('Received signal, changing to spinner');
			buttonEl.innerHTML = `<span class="loading loading-spinner loading-${loaderSize}"></span>`;
		}
	});

	buttonEl.addEventListener('click', () => {
		// console.log('Received signal, changing to spinner');
		buttonEl.innerHTML = `<span class="loading loading-spinner loading-${loaderSize}"></span>`;
	});

	window.addEventListener('htmx:afterSwap', ((evt: HtmxAfterRequest) => {
		if (evt.detail.elt === formEl) {
			setTimeout(() => {
				buttonEl.innerHTML = buttonHTML; // Add Comment
			}, 200);
		}
	}) as EventListener);

	window.addEventListener('htmx:validation:failed', () => {
		// console.log('Changing back to text button');
		setTimeout(() => {
			buttonEl.innerHTML = buttonHTML; //'Add Comment';
		}, 1);
	});

	if (formEl) {
		window.addEventListener('htmx:afterRequest', ((evt: HtmxAfterRequest) => {
			if (evt.detail.elt === formEl) {
				setTimeout(() => {
					buttonEl.innerHTML = buttonHTML;
				}, 200);
			}
		}) as EventListener);
	}
}
