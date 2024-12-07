export function keyboardShortcut(inputEl, buttonEl) {
	//This is keydown, so it's faster than the keyup submit hx-trigger on the form
	inputEl.addEventListener('keydown', (evt) => {
		if (evt.ctrlKey && evt.key === 'Enter') {
			console.log('Received signal, changing to spinner');
			buttonEl.innerHTML = '<span class="loading loading-spinner loading-xs"></span>';
		}
	});

	buttonEl.addEventListener('click', () => {
		console.log('Received signal, changing to spinner');
		buttonEl.innerHTML = '<span class="loading loading-spinner loading-xs"></span>';
	});

	document.addEventListener('htmx:afterSwap', (evt) => {
		if (evt.detail.elt === inputEl) {
			buttonEl.innerHTML = 'Add Comment';
		}
	});
}
