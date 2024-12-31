import { keyboardShortcut } from './common';

// const instantSSEEl = document.getElementById('instant-sse');

// window.addEventListener('htmx:sseBeforeMessage', (evt) => {
// 	console.log(evt);

// 	if (instantSSEEl.innerHTML) {
// 		evt.preventDefault();
// 		let newEl = document.createElement('li');
// 		newEl.innerHTML = evt.detail.data;
// 		newEl.classList.add('text-error');
// 		instantSSEEl.appendChild(newEl);
// 	}
// });
const liveCommentForm = document.getElementById('live-comment-form') as HTMLFormElement;
const contentInput = document.getElementById('content-input') as HTMLInputElement;
const postButton = document.getElementById('post-button') as HTMLButtonElement;
// const nearInstantDelayMs: number = 100;

(() => {
	window.addEventListener('load', () => {
		if (contentInput && postButton) {
			contentInput.focus();
			keyboardShortcut(contentInput, postButton, 'md', liveCommentForm);
		}
		window.addEventListener('htmx:sseMessage', () => {
			const sse = document.getElementById('instant-sse') as HTMLUListElement;
			const spinner = document.getElementById('spinner') as HTMLDivElement;

			// setTimeout(() => {
			// 	sse.classList.remove('hidden');
			// 	spinner.classList.add('hidden');
			// 	sse.classList.add('grid');
			// }, nearInstantDelayMs);

			setTimeout(() => {
				sse.classList.remove('hidden');
				spinner.classList.add('hidden');
				sse.classList.add('grid');
			}, 100);
		});
	});
})();

(() => {
	window.addEventListener('htmx:afterRequest', ((evt: HtmxAfterRequest) => {
		const reqStatus = evt.detail.successful;
		window.addEventListener('htmx:sseMessage', () => {
			if (reqStatus && contentInput && evt.detail.elt === liveCommentForm) {
				contentInput.value = '';
				contentInput.focus();
			}
		});
	}) as EventListener);
})();
