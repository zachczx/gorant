(() => {
	const passwordInput = document.getElementById('password-input') as HTMLInputElement;
	const revealPasswordButton = document.getElementById('show-password-button') as HTMLButtonElement;
	const iconShow = document.getElementById('icon-show') as HTMLButtonElement;
	const iconHide = document.getElementById('icon-hide') as HTMLButtonElement;

	revealPasswordButton.addEventListener('click', () => {
		if (revealPasswordButton.dataset.status === 'hidden') {
			passwordInput.type = 'text';
			revealPasswordButton.dataset.status = 'show';
			iconShow.classList.add('hidden');
			iconHide.classList.remove('hidden');
		} else {
			passwordInput.type = 'password';
			revealPasswordButton.dataset.status = 'hidden';
			iconHide.classList.add('hidden');
			iconShow.classList.remove('hidden');
		}
	});
})();
