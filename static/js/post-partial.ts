// Comment Delete Animation

(function deleteButton() {
	const deleteButtons = document.getElementsByClassName('delete-button') as HTMLCollectionOf<HTMLElement>;
	const postUpvoteClasses = ['bg-red-400/40', 'transition-all', 'opacity-40', 'duration-1000', 'ease-out'];
	const postBodyClasses = ['bg-red-400/20', 'transition-all', 'opacity-40', 'duration-1000', 'ease-out'];
	for (const button of deleteButtons) {
		button.addEventListener('click', () => {
			const parentCommentId = 'post-' + button.dataset.parentCommentId;

			const parentComment = document.getElementById(parentCommentId);
			if (parentComment?.classList.contains('animate-highlight-border')) {
				parentComment.classList.remove('animate-highlight-border');
			}

			const postUpvote = document.getElementById('post-upvote-' + button.dataset.parentCommentId) as HTMLDivElement;
			const postBody = document.getElementById('post-body-' + button.dataset.parentCommentId) as HTMLDivElement;
			const deleteLoader = document.getElementById(
				'post-delete-loader-' + button.dataset.parentCommentId,
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
