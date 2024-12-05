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
		document.getElementById('tags-data')
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
		if (evt.key === 'Enter' || evt.key === ',' || evt.key === ' ' || evt.key === ';' || evt.key === '.') {
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
}
