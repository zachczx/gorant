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
(() => {
	window.addEventListener('load', () => {
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
