import { keyboardShortcut } from './common.js';

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
const liveCommentForm = document.getElementById('live-comment-form');
const contentInput = document.getElementById('content-input');
const postButton = document.getElementById('post-button');

(() => {
	window.addEventListener('load', () => {
		if (contentInput && postButton) {
			contentInput.focus();
			keyboardShortcut(contentInput, postButton, 'md', liveCommentForm);
		}
		window.addEventListener('htmx:sseMessage', () => {
			const sse = document.getElementById('instant-sse');
			const spinner = document.getElementById('spinner');
			console.log('triggered');

			setTimeout(() => {
				sse.classList.remove('hidden');
				spinner.classList.add('hidden');
				sse.classList.add('grid');
			}, 100);
		});
	});
})();

(() => {
	const contentInput = document.getElementById('content-input');
	window.addEventListener('htmx:afterRequest', (evt) => {
		const reqStatus = evt.detail.successful;
		console.log('reqStatus: ', reqStatus);
		console.log('contentInput: ', contentInput);
		console.log('elt: ', evt.detail.elt);
		console.log('form el: ', liveCommentForm);
		window.addEventListener('htmx:sseMessage', () => {
			if (reqStatus && contentInput && evt.detail.elt === liveCommentForm) {
				contentInput.value = '';
				contentInput.focus();
			}
		});
	});
})();
