// Post Description

(function activateDescriptionForm() {
	const postDescriptionStatic = document.getElementById('post-description-static');
	if (postDescriptionStatic) {
		postDescriptionStatic.addEventListener('click', () => {
			const postDescriptionForm = document.getElementById('post-description-form');
			postDescriptionForm.classList.remove('hidden');
			const postDescriptionStatic = document.getElementById('post-description-static');
			postDescriptionStatic.classList.add('hidden');
		});
	}

	const postDescriptionCancel = document.getElementById('post-description-cancel');
	if (postDescriptionCancel) {
		postDescriptionCancel.addEventListener('click', () => {
			const postDescriptionForm = document.getElementById('post-description-form');
			postDescriptionForm.classList.add('hidden');
			const postDescriptionStatic = document.getElementById('post-description-static');
			postDescriptionStatic.classList.remove('hidden');
		});
	}

	const moreActionsEditButton = document.getElementById('more-actions-edit-button');
	if (moreActionsEditButton) {
		moreActionsEditButton.addEventListener('click', () => {
			const postDescriptionForm = document.getElementById('post-description-form');
			postDescriptionForm.classList.remove('hidden');
			const postDescriptionStatic = document.getElementById('post-description-static');
			postDescriptionStatic.classList.add('hidden');
		});
	}
})();

// Post Form

(function keyboardShortcutListeners() {
	//
	document.addEventListener('keydown', (evt) => {
		if (evt.target.id === 'comment-form-name-input' || evt.target.id === 'comment-form-message-input') {
			if (evt.ctrlKey && evt.key === 'Enter') {
				document.getElementById('comment-submit-button').innerHTML =
					'<span class="loading loading-spinner loading-md"></span>';
				setTimeout(() => {
					document.getElementById('comment-form').requestSubmit();
					document.getElementById('comment-submit-button').innerHTML = 'Add Comment';
				}, 1000);
			}
		}

		if (evt.target.id === 'comment-submit-button') {
			document.getElementById('comment-submit-button').innerHTML =
				'<span class="loading loading-spinner loading-md"></span>';
			setTimeout(() => {
				document.getElementById('comment-form').requestSubmit();
				document.getElementById('comment-submit-button').innerHTML = 'Add Comment';
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
					document.getElementById('comment-submit-button').innerHTML = 'Add Comment';
				}, 1000);
			}
		}

		if (evt.target.id === 'comment-submit-button') {
			evt.preventDefault();
			document.getElementById('comment-submit-button').innerHTML =
				'<span class="loading loading-spinner loading-md"></span>';
			setTimeout(() => {
				document.getElementById('comment-form').requestSubmit();
				document.getElementById('comment-submit-button').innerHTML = 'Add Comment';
			}, 1000);
		}
	});
})();

// Post Settings Button Click Outside

(function () {
	document.addEventListener('click', (evt) => {
		const moreActionsDropdown = document.getElementById('more-actions-dropdown');
		if (moreActionsDropdown) {
			if (evt.target.id !== 'more-actions-dropdown' && document.getElementById('more-actions-dropdown').open) {
				document.getElementById('more-actions-dropdown').open = false;
			}
		}
	});
})();

// Copy Post Link Button
(function copyPostLink() {
	const moreActionsCopyButton = document.getElementById('more-actions-copy-button');
	if (moreActionsCopyButton) {
		// const postId = moreActionsCopyButton.dataset.postId.slice(5)
		moreActionsCopyButton.addEventListener('click', () => {
			navigator.clipboard.writeText(window.location.href);
		});
	}
})();

// Filter bar functionality

(function showFilterCancelButton() {
	document.getElementById('filter-input').addEventListener('keyup', (evt) => {
		if (evt.target.value.length > 0) {
			document.getElementById('filter-cancel-button').classList.remove('hidden');
		} else if (evt.target.value.length === 0) {
			document.getElementById('filter-cancel-button').classList.add('hidden');
		}
	});
})();

(function clearFilterInput() {
	document.getElementById('filter-cancel-button').addEventListener('click', () => {
		document.getElementById('filter-input').value = '';
	});
})();
