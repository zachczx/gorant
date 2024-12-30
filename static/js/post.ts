import tags from './tags';
import { keyboardShortcut } from './common';

/**
 * Post Description on Post page.
 * To show and hide edit post description form.
 */
(function activateDescriptionForm() {
	const postDescriptionStatic = document.getElementById('post-description-static') as HTMLButtonElement;
	const postDescriptionForm = document.getElementById('post-description-form') as HTMLFormElement;
	if (postDescriptionStatic) {
		postDescriptionStatic.addEventListener('click', () => {
			postDescriptionForm?.classList.remove('hidden');
			postDescriptionStatic?.classList.add('hidden');
		});
	}

	const postDescriptionCancel = document.getElementById('post-description-cancel') as HTMLButtonElement;
	if (postDescriptionCancel) {
		postDescriptionCancel.addEventListener('click', () => {
			postDescriptionForm.classList.add('hidden');
			postDescriptionStatic.classList.remove('hidden');
		});
	}

	const moreActionsEditButton = document.getElementById('more-actions-edit-button') as
		| HTMLButtonElement
		| HTMLLabelElement;
	if (moreActionsEditButton) {
		moreActionsEditButton.addEventListener('click', () => {
			postDescriptionForm?.classList.remove('hidden');
			postDescriptionStatic?.classList.add('hidden');
		});
	}
})();

// Post title
const postTitleTruncated = document.getElementById('post-title-truncated') as HTMLButtonElement;
const postTitleTruncatedClasses = ['line-clamp-2', 'max-h-[6.3rem]'];
postTitleTruncated.addEventListener('click', (evt) => {
	if ((<HTMLButtonElement>evt.target)?.classList.contains('line-clamp-2')) {
		(<HTMLButtonElement>evt.target).classList.remove(...postTitleTruncatedClasses);
	} else {
		(<HTMLButtonElement>evt.target)?.classList.add(...postTitleTruncatedClasses);
	}
});

/**
 * Utility to close post settings dropdown menu when clicking outside.
 */
(() => {
	document.addEventListener('click', (evt) => {
		const moreActionsDropdown = document.getElementById('more-actions-dropdown') as HTMLDetailsElement;
		if (moreActionsDropdown) {
			if (evt.target !== moreActionsDropdown && moreActionsDropdown.open) {
				moreActionsDropdown.open = false;
			}
		}
	});
})();

/**
 * Button to copy link to post.
 */
(function copyPostLink() {
	const moreActionsCopyButton = document.getElementById('more-actions-copy-button');
	if (moreActionsCopyButton) {
		// const postId = moreActionsCopyButton.dataset.postId.slice(5)
		moreActionsCopyButton.addEventListener('click', () => {
			navigator.clipboard.writeText(window.location.href);
		});
	}
})();

/**
 * Show/hide clear button when user types something in the filter bar input field.
 */
(function showFilterClearButton() {
	const filterInput = document.getElementById('filter-input') as HTMLInputElement;
	const filterInputClearButton = document.getElementById('filter-clear-button') as HTMLButtonElement;
	filterInput.addEventListener('keyup', (evt) => {
		if ((<HTMLInputElement>evt.target).value.length > 0) {
			filterInputClearButton.classList.remove('hidden');
		} else if ((<HTMLInputElement>evt.target).value.length === 0) {
			filterInputClearButton.classList.add('hidden');
		}
	});
})();

/**
 * Clears filter bar input field text upon click.
 */
(function clearFilterInput() {
	const filterInput = document.getElementById('filter-input') as HTMLInputElement;
	const filterInputClearButton = document.getElementById('filter-clear-button') as HTMLButtonElement;
	filterInputClearButton.addEventListener('click', () => {
		filterInput.value = '';
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
	const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLTextAreaElement;
	const formMessageLabel = document.getElementById('form-message-label') as HTMLDivElement;
	const commentFormErrorMessage = document.getElementById('comment-form-error-message') as HTMLDivElement;
	if (evt.detail.elt === commentFormMessageInput) {
		if (commentFormMessageInput.value.length < 10) {
			commentFormMessageInput.classList.add('border-error');
			formMessageLabel.classList.add('text-error');
			commentFormErrorMessage.innerText = 'Message must be at least 10 characters long.';
		}
	}
});
