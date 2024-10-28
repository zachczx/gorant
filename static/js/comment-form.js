import { gsap } from 'gsap/dist/gsap';

// gsap.from('.content', { y: 100, duration: 0.7, autoAlpha: 0 });

document.addEventListener('keydown', (evt) => {
	if (evt.target.id === 'comment-form-name-input' || evt.target.id === 'comment-form-message-input') {
		if (evt.ctrlKey && evt.key === 'Enter') {
			document.getElementById('comment-submit-button').innerHTML =
				'<span class="loading loading-spinner loading-md"></span>';
			setTimeout(() => {
				document.getElementById('comment-form').requestSubmit();
			}, 1000);
		}
	}

	if (evt.target.id === 'comment-submit-button') {
		document.getElementById('comment-submit-button').innerHTML =
			'<span class="loading loading-spinner loading-md"></span>';
		setTimeout(() => {
			document.getElementById('comment-form').requestSubmit();
		}, 1000);
	}
});

document.addEventListener('click', (evt) => {
	if (evt.target.id === 'comment-form-name-input' || evt.target.id === 'comment-form-message-input') {
		if (evt.ctrlKey && evt.key === 'Enter') {
			evt.preventDefault();
			document.getElementById('comment-submit-button').innerHTML =
				'<span class="loading loading-spinner loading-md"></span>';
			setTimeout(() => {
				document.getElementById('comment-form').requestSubmit();
			}, 1000);
		}
	}

	if (evt.target.id === 'comment-submit-button') {
		evt.preventDefault();
		document.getElementById('comment-submit-button').innerHTML =
			'<span class="loading loading-spinner loading-md"></span>';
		setTimeout(() => {
			document.getElementById('comment-form').requestSubmit();
		}, 1000);
	}
});
