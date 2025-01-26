import tags from './tags';
import { keyboardShortcut } from './common';

// Post title
const postTitleTruncated = document.getElementById('post-title-truncated') as HTMLButtonElement;
const postTitleTruncatedClasses = ['line-clamp-2', 'max-h-[6.3rem]'];
postTitleTruncated.addEventListener('click', (evt) => {
	if ((evt.target as HTMLInputElement)?.classList.contains('line-clamp-2')) {
		(evt.target as HTMLInputElement).classList.remove(...postTitleTruncatedClasses);
	} else {
		(evt.target as HTMLInputElement)?.classList.add(...postTitleTruncatedClasses);
	}
});

/**
 * Delete button listener for modal.
 */
(() => {
	const deletePostButton = document.getElementById('delete-post-button') as HTMLButtonElement;
	const deletePostModal = document.getElementById('delete_post_modal') as HTMLDialogElement;
	if (deletePostButton && deletePostModal) {
		deletePostButton.addEventListener('click', () => {
			deletePostModal.showModal();
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
		if ((evt.target as HTMLInputElement).value.length > 0) {
			filterInputClearButton.classList.remove('hidden');
		} else if ((evt.target as HTMLInputElement).value.length === 0) {
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

/**
 * Executes eventListeners for keyboard shortcuts to post (ctr+enter).
 */
function initKeyBoardShortcutForPosts() {
	const commentForm = document.getElementById('comment-form') as HTMLFormElement;
	const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLInputElement;
	const commentSubmitButton = document.getElementById('comment-submit-button') as HTMLButtonElement;
	if (commentFormMessageInput && commentSubmitButton) {
		keyboardShortcut(commentFormMessageInput, commentSubmitButton, undefined, commentForm, 'textarea');
	}
}

window.addEventListener('load', initKeyBoardShortcutForPosts);
// window.addEventListener('htmx:afterSwap', initKeyBoardShortcutForPosts);

type HtmxEvent = {
	detail?: {
		message: string;
		validity: ValidityState;
		elt: HTMLElement;
		xhr: XMLHttpRequest;
		target: HTMLElement;
	};
};
window.addEventListener('htmx:validation:failed', ((evt: HtmxEvent) => {
	const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLTextAreaElement;
	const formMessageLabel = document.getElementById('form-message-label') as HTMLDivElement;
	const commentFormErrorMessage = document.getElementById('comment-form-error-message') as HTMLDivElement;
	if (evt.detail?.elt === commentFormMessageInput) {
		if (commentFormMessageInput.value.length < 10) {
			commentFormMessageInput.classList.add('border-error');
			formMessageLabel.classList.add('text-error');
			commentFormErrorMessage.classList.remove('hidden');
			commentFormErrorMessage.innerText = 'Message must be at least 10 characters long.';
		}
	}
}) as EventListener);

window.addEventListener('load', uploadInputSelectionText);
window.addEventListener('htmx:afterSwap', uploadInputSelectionText);

/**
 * Replaces the text of the file input droparea with name of file.
 */
function uploadInputSelectionText() {
	// console.log('triggered commentformfileinput listener');
	const commentFormFileInput = document.getElementById('comment-file-input') as HTMLInputElement;
	const commentFormFileMessage = document.getElementById('comment-file-message') as HTMLDivElement;
	if (commentFormFileInput) {
		commentFormFileInput.addEventListener('change', () => {
			if (commentFormFileInput.files) {
				const fileName = commentFormFileInput.files[0].name;
				commentFormFileMessage.innerHTML = `<b>${fileName}</b>`;
			}
		});
	}
}

window.addEventListener('load', commentUploadDragAndDrop);
window.addEventListener('htmx:afterSwap', commentUploadDragAndDrop);

/**
 * Handles the drag and drop of files for all possible eventListeners (dragenter, dragover, dragleave, drop).
 */
function commentUploadDragAndDrop() {
	const commentFileInputDroparea = document.getElementById('comment-file-input-droparea');
	commentFileInputDroparea?.addEventListener('dragenter', (evt) => {
		evt.preventDefault();
		evt.stopPropagation();
		commentFileInputDroparea?.classList.add('bg-primary/30');
	});
	commentFileInputDroparea?.addEventListener('dragover', (evt) => {
		evt.preventDefault();
		evt.stopPropagation();
		if (!commentFileInputDroparea.classList.contains('bg-primary/30')) {
			commentFileInputDroparea?.classList.add('bg-primary/30');
		}
	});
	commentFileInputDroparea?.addEventListener('dragleave', (evt) => {
		evt.preventDefault();
		evt.stopPropagation();
		commentFileInputDroparea?.classList.remove('bg-primary/30');
	});
	commentFileInputDroparea?.addEventListener('drop', handleDrop);
}

/**
 * Handles the drag and drop of files, to then populate the file input for eventual upload only on form submission.
 */
function handleDrop(evt: DragEvent) {
	evt.preventDefault();
	// stopPropagation() is needed otherwise browser complains:
	// 		Uncaught TypeError: n.drop is not a function "content.js"
	evt.stopPropagation();
	// DataTransfer() needed to convert file to a FileList,
	// otherwise browser complains the input value is not a FileList.
	const dataTransfer = new DataTransfer();
	const commentFormFileInput = document.getElementById('comment-file-input') as HTMLInputElement;
	const commentFormFileMessage = document.getElementById('comment-file-message') as HTMLDivElement;
	if (evt.dataTransfer?.files[0]) {
		dataTransfer.items.add(evt.dataTransfer.files[0]);
		commentFormFileInput.files = dataTransfer.files;
		commentFormFileMessage.innerHTML = `<b>${evt.dataTransfer.files[0].name}</b>`;
	}
	const commentFileInputDroparea = document.getElementById('comment-file-input-droparea');
	if (commentFileInputDroparea?.classList.contains('bg-primary/30')) {
		commentFileInputDroparea?.classList.remove('bg-primary/30');
	}
}

/**
 * Edit post listeners
 */

(function initRepliesAttachButtonListener() {
	window.addEventListener('load', commentAttachButtonListener);
	window.addEventListener('htmx:afterRequest', commentAttachButtonListener);
})();

function commentAttachButtonListener() {
	const replyFormAttachmentButton = document.getElementsByClassName(
		'reply-form-attachment-button',
	) as HTMLCollectionOf<HTMLButtonElement>;

	for (const button of replyFormAttachmentButton) {
		const parentReplyId = button.dataset.parentReplyId;
		console.log(parentReplyId);

		const replyFormAttachmentAccordion = document.getElementById(
			'comment-' + parentReplyId + '-edit-form-attachment-accordion',
		) as HTMLDivElement;
		if (button) {
			button.addEventListener('click', () => {
				if (replyFormAttachmentAccordion.classList.contains('hidden')) {
					replyFormAttachmentAccordion.classList.remove('hidden');
				} else {
					replyFormAttachmentAccordion.classList.add('hidden');
				}
			});
		}
	}
}
