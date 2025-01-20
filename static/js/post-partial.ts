/**
 * Comment Delete Animation
 */
(function commentDeleteButtonListeners() {
	const deleteButtons = document.getElementsByClassName('comment-delete-button') as HTMLCollectionOf<HTMLElement>;
	const postUpvoteClasses = ['bg-red-400/40', 'transition-all', 'opacity-40', 'duration-1000', 'ease-out'];
	const postBodyClasses = ['bg-red-400/20', 'transition-all', 'opacity-40', 'duration-1000', 'ease-out'];
	for (const button of deleteButtons) {
		button.addEventListener('click', () => {
			const parentCommentId = 'comment-' + button.dataset.parentCommentId;
			const parentComment = document.getElementById(parentCommentId);
			if (parentComment?.classList.contains('animate-highlight-border')) {
				parentComment.classList.remove('animate-highlight-border');
			}

			const postUpvote = document.getElementById('comment-upvote-' + button.dataset.parentCommentId) as HTMLDivElement;
			const postBody = document.getElementById('comment-body-' + button.dataset.parentCommentId) as HTMLDivElement;
			const deleteLoader = document.getElementById(
				'comment-delete-loader-' + button.dataset.parentCommentId,
			) as HTMLDivElement;

			deleteLoader?.classList.remove('hidden');
			deleteLoader?.classList.add('flex');

			postUpvote?.classList.remove('bg-primary/30');
			postUpvote?.classList.add(...postUpvoteClasses);
			postBody?.classList.add(...postBodyClasses);
			// Needs to remove the animation and force a recalculation.
			// Otherwise css animation won't start on the el which had animate-highlight-border applied to it.
			// Read more: https://stackoverflow.com/questions/22093141/adding-class-via-js-wont-trigger-css-animation
			// void parentComment.offsetWidth
			// parentComment.classList.add(...classes)
		});
	}
})();

/**
 * The alternative to triggering the entire function is to put this in post.ts and listen like the one below.
 *
 * If so, I can't listen to afterSettle or afterSwap, else it'd trigger twice.
 * Even listening on afterRequest triggers twice, but only on the same loop and same values,
 * so it's not like flicking a switch on/off back and forth.
 */
/* (function commentReplyButton() {
	window.addEventListener('load', replyButtonListener);


	window.addEventListener('htmx:afterRequest', replyButtonListener);
})(); */

/**
 * This function sets listeners for comment reply buttons to show input field for reply.
 * Using a closure since this gets swapped into the DOM after every request.
 */
(function CommentReplyButtonListeners() {
	const replyButtons = document.getElementsByClassName('reply-button') as HTMLCollectionOf<HTMLDivElement>;
	for (const button of replyButtons) {
		const commentReplyFormId = 'comment-' + button.dataset.commentId + '-reply-form';
		const commentReplyForm = document.getElementById(commentReplyFormId) as HTMLFormElement;
		const commentReplyInput = document.getElementById(
			'comment-' + button.dataset.commentId + '-reply-input',
		) as HTMLLabelElement;
		if (button) {
			button.addEventListener('click', () => {
				if (commentReplyForm.classList.contains('hidden')) {
					commentReplyForm.classList.remove('hidden');
					commentReplyInput.classList.remove('hidden');
				} else {
					commentReplyForm.classList.add('hidden');
					commentReplyInput.classList.add('hidden');
				}
			});
		}
	}
})();

(function replyDeleteButtonListeners() {
	const deleteButtons = document.getElementsByClassName('reply-delete-button') as HTMLCollectionOf<HTMLElement>;
	for (const button of deleteButtons) {
		button.addEventListener('click', () => {
			/**
			 * DaisyUI keeps the dropdown open until it loses focus, so clicking it keeps it in focus.
			 * See: https://www.reddit.com/r/tailwindcss/comments/rm0rpu/tailwind_and_daisyui_how_to_fix_the_issue_with/
			 *
			 * document.activeElement is only guaranteed to exist on HTMLElements, not all Elements, so svg would cause an exception.
			 */
			if (document.activeElement instanceof HTMLElement) {
				document.activeElement?.blur();
			}
			const parentReplyId = 'reply-' + button.dataset.parentReplyId;
			const reply = document.getElementById(parentReplyId) as HTMLDivElement;
			if (reply?.classList.contains('animate-highlight-border')) {
				reply.classList.remove('animate-highlight-border');
			}
			const deleteLoader = document.getElementById(
				'reply-delete-loader-' + button.dataset.parentReplyId,
			) as HTMLDivElement;
			deleteLoader?.classList.remove('hidden');
			deleteLoader?.classList.add('flex');
		});
	}
})();
