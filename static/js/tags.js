// For UI supporting Tag's tag-style input

//Window is used because document.addEventListener() is unreliable

export default function tags() {
	window.addEventListener('load', () => {
		if (checkDomForTagsEls()) {
			tagsUi();
		}
	});

	document.addEventListener('htmx:afterSwap', () => {
		if (checkDomForTagsEls()) {
			tagsUi();
		}
	});
}

function checkDomForTagsEls() {
	console.log('Checking DOM for tagsEls');
	if (
		document.getElementById('tags-input') &&
		document.getElementById('tags-list') &&
		document.getElementById('tags-data') &&
		document.getElementById('tags-save-button')
	) {
		return true;
	} else {
		return false;
	}
}

function tagsUi() {
	console.log('Loading tagsUI');
	const tagsInput = document.getElementById('tags-input');
	const tagsList = document.getElementById('tags-list');
	const tagsSaveButton = document.getElementById('tags-save-button');
	let tagsData = document.getElementById('tags-data');
	let margin = 0;

	// Classes to add for each tag
	const classes = [
		'user-tag',
		'btn',
		'bg-primary/50',
		'text-accent',
		'border-0',
		'hover:bg-primary/60',
		'btn-sm',
		'text-md',
		'me-2',
		'my-2',
	];

	// Fetch tags from hidden form value if there already are tags there
	if (tagsData.value.length > 0) {
		console.log(tagsData.value);
		let tags = tagsData.value.split(',');
		console.log(tags);
		for (let i = 0; i < tags.length; i++) {
			tags[i] = tags[i].trim().toLowerCase();
			if (tags[i].length > 0) {
				let el = document.createElement('li');
				el.id = 'user-tag-' + tags[i];
				el.classList.add(...classes);
				el.innerHTML = tags[i];
				tagsList.appendChild(el);
			}
		}
	}

	tagsInput.addEventListener('keydown', (evt) => {
		if (
			evt.key === 'Enter' ||
			evt.key === ',' ||
			evt.key === ' ' ||
			evt.key === ';' ||
			evt.key === '.' ||
			(evt.key === 'Tab' && tagsInput.value.length > 0)
		) {
			evt.preventDefault();
			let tags = tagsInput.value.split(',');
			for (let i = 0; i < tags.length; i++) {
				tags[i] = tags[i].trim().toLowerCase();
				if (tags[i].length > 0) {
					let el = document.createElement('li');
					el.id = 'user-tag-' + tags[i];
					el.classList.add(...classes);
					el.innerHTML = tags[i];
					tagsList.appendChild(el);

					if (tagsData.value.length === 0) {
						tagsData.value = tags[i];
					} else {
						tagsData.value = tagsData.value + ',' + tags[i];
					}
				}
			}
			tagsInput.value = '';

			if (tagsList.childElementCount > 0) {
				margin = 'me-' + String(tagsList.childElementCount * 2);
				tagsList.classList.add(margin);
			}
		}
	});

	tagsList.addEventListener('click', (evt) => {
		if (evt.target.id.includes('user-tag-')) {
			evt.preventDefault();
			console.log('Clicked a tag');
			tagsList.removeChild(evt.target);
			if (!tagsData.value.includes(',')) {
				tagsData.value = tagsData.value.replace(evt.target.innerText, '');
			} else if (tagsData.value.includes(',') && tagsData.value.includes(evt.target.innerText + ',')) {
				tagsData.value = tagsData.value.replace(evt.target.innerText + ',', '');
			} else if (tagsData.value.includes(',') && tagsData.value.includes(',' + evt.target.innerText)) {
				tagsData.value = tagsData.value.replace(',' + evt.target.innerText, '');
			}
		}
	});

	// This helps when user enters into input field, but doesn't press any of the triggers to add the value to the hidden field.
	// This does a last check to add all remaining input in the field before posting it
	//
	// Note: Seems like just the click event precedes the request, so this doesn't require evt.preventDefault()
	// Note: However, I added a delay of 100ms to hx-trigger just to be safe
	tagsSaveButton.addEventListener('click', () => {
		if (tagsInput.value.length > 0) {
			tagsData.value = tagsData.value + ',' + tagsInput.value;
		}
	});

	window.addEventListener('htmx:beforeSend', (evt) => {
		console.log(evt.detail.elt);
		console.log(evt.detail.xhr);
		console.log(evt.detail.requestConfig);
	});
}
