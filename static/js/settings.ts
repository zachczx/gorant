const settingsFormDisplayName = document.getElementById('settings-form-display-name') as HTMLInputElement;
const settingsFormDisplayNameErrorClasses: string[] = ['border-error', 'focus:border-error', 'focus:outline-error'];
const settingsFormDisplayNameError = document.getElementById('settings-form-error-display-name') as HTMLSpanElement;
const regex = /^[0-9A-Za-z \-_+()[\]|@.]+$/;
settingsFormDisplayName.addEventListener('keyup', () => {
	console.log(regex.test(settingsFormDisplayName.value));
	if (!regex.test(settingsFormDisplayName.value)) {
		settingsFormDisplayNameError.classList.remove('hidden');
		settingsFormDisplayName.classList.add(...settingsFormDisplayNameErrorClasses);
		settingsFormDisplayNameError.innerText = 'Special characters not allowed! Use only A-Z, a-z, 0-9, -, _, (), +';
	} else {
		if (!settingsFormDisplayNameError.classList.contains('hidden')) {
			settingsFormDisplayNameError.classList.add('hidden');
		}
		settingsFormDisplayName.classList.remove(...settingsFormDisplayNameErrorClasses);
	}
});
