/* *
 * Post Form calculation feature for remaining chars.
 */
function calculateCharsRemaining() {
	const commentFormMessageInputEl = document.getElementById('comment-form-message-input') as HTMLInputElement;
	const commentFormCharsRemainingEl = document.getElementById('form-message-chars-remaining') as HTMLSpanElement;

	if (commentFormMessageInputEl && commentFormCharsRemainingEl) {
		const total = 2000;
		commentFormCharsRemainingEl.innerHTML = String(total);

		if (commentFormMessageInputEl.value) {
			commentFormCharsRemainingEl.innerHTML = String(calculate(commentFormMessageInputEl.value, total));
		}

		commentFormMessageInputEl.addEventListener('keyup', () => {
			commentFormCharsRemainingEl.innerHTML = String(calculate(commentFormMessageInputEl.value, total));
		});
	}
}

calculateCharsRemaining();

window.addEventListener('htmx:afterSwap', () => {
	calculateCharsRemaining();
});

function calculate(value: string, total: number) {
	const count = value.trim().length;
	const remaining = total - count;
	return remaining;
}
