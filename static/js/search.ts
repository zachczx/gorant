window.addEventListener('load', highlightSearchResults);
window.addEventListener('htmx:afterSwap', highlightSearchResults);

function highlightSearchResults() {
	const searchResultsEl = document.getElementById('search-results-list') as HTMLDivElement;
	const query = searchResultsEl?.dataset.query;
	const results = document.getElementsByClassName('search-result-content');
	for (const x of results) {
		const inner = x.innerHTML;
		if (inner && query) {
			if (inner.toLowerCase().includes(query.toLowerCase())) {
				const replaced = highlight(inner, query);
				x.innerHTML = replaced;
			}
		}
	}
}
function highlight(mainText: string, query: string) {
	const re = new RegExp(query, 'gi'); // global, insensitive
	const newText = mainText.replace(re, `<b>$&</b>`);
	return newText;
}

window.addEventListener('htmx:beforeRequest', () => {
	console.log('event triggered htmx:beforeRequest');
	checkValue();
});

window.addEventListener('htmx:beforeProcessNode', () => {
	console.log('event triggered htmx:beforeProcessNode');
	checkValue();
});
checkValue();
function checkValue() {
	const sortValue = document.getElementsByClassName('search-sort') as HTMLCollectionOf<HTMLInputElement>;
	console.log(sortValue);
	for (const v of sortValue) {
		console.log(v.checked);
	}
}
