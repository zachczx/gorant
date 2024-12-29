import tags from './tags';
import { keyboardShortcut } from './common';

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

// Post title
const postTitleTruncated = document.getElementById('post-title-truncated');
const postTitleTruncatedClasses = ['line-clamp-2', 'max-h-[6.3rem]'];
postTitleTruncated.addEventListener('click', (evt) => {
	if (evt.target.classList.contains('line-clamp-2')) {
		evt.target.classList.remove(...postTitleTruncatedClasses);
	} else {
		evt.target.classList.add(...postTitleTruncatedClasses);
	}
});

// Post Settings Button Click Outside

(() => {
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

tags();

function initKeyBoardShortcutForPosts() {
	// const commentForm = document.getElementById('comment-form');
	const commentFormMessageInput = document.getElementById('comment-form-message-input');
	const commentSubmitButton = document.getElementById('comment-submit-button');
	if (commentFormMessageInput && commentSubmitButton) {
		keyboardShortcut(commentFormMessageInput, commentSubmitButton, undefined, undefined, 'textarea');
	}
}

window.addEventListener('load', initKeyBoardShortcutForPosts);
window.addEventListener('htmx:afterSwap', initKeyBoardShortcutForPosts);

window.addEventListener('htmx:validation:failed', (evt) => {
	const commentFormMessageInput = document.getElementById('comment-form-message-input');
	const formMessageLabel = document.getElementById('form-message-label');
	const commentFormErrorMessage = document.getElementById('comment-form-error-message');
	if (evt.detail.elt === commentFormMessageInput) {
		if (commentFormMessageInput.value.length < 10) {
			commentFormMessageInput.classList.add('border-error');
			formMessageLabel.classList.add('text-error');
			commentFormErrorMessage.innerText = 'Message must be at least 10 characters long.';
		}
	}
});
