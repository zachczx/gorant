/* *
 * Post Form calculation feature for remaining chars.
 */

window.addEventListener('load', calculateFormMessageChars);
window.addEventListener('htmx:afterRequest', calculateFormMessageChars);

function calculateFormMessageChars() {
	const commentFormMessageInputEl = document.getElementById('comment-form-message-input') as HTMLInputElement;
	const commentFormCharsRemainingEl = document.getElementById('form-message-chars') as HTMLSpanElement;
	if (commentFormMessageInputEl && commentFormCharsRemainingEl) {
		if (!commentFormMessageInputEl.value || commentFormMessageInputEl.value.length < 1) {
			commentFormCharsRemainingEl.innerText = '0';
		} else {
			commentFormCharsRemainingEl.innerText = String(commentFormMessageInputEl.value.trim().length);
		}

		// This needs to be keyup, for some reason keydown doesn't start counting for the first character,
		// so it ends up inaccurate by at least 1.
		commentFormMessageInputEl.addEventListener('keyup', () => {
			commentFormCharsRemainingEl.innerText = String(commentFormMessageInputEl.value.trim().length);
		});
	}
}

(function initCommentAttachButtonListener() {
	window.addEventListener('load', commentAttachButtonListener);
	window.addEventListener('htmx:afterRequest', commentAttachButtonListener);
})();

function commentAttachButtonListener() {
	const commentFormAttachmentButton = document.getElementById('comment-form-attachment-button') as HTMLButtonElement;
	const commentFormAttachmentAccordion = document.getElementById('comment-form-attachment-accordion') as HTMLDivElement;
	if (commentFormAttachmentButton) {
		commentFormAttachmentButton.addEventListener('click', () => {
			if (commentFormAttachmentAccordion.classList.contains('hidden')) {
				commentFormAttachmentAccordion.classList.remove('hidden');
			} else {
				commentFormAttachmentAccordion.classList.add('hidden');
			}
		});
	}
}
