window.addEventListener('load', highlightSearchResults);
window.addEventListener('htmx:afterSwap', highlightSearchResults);

function highlightSearchResults() {
	const searchResultsEl = document.getElementById('search-results-list') as HTMLDivElement;
	const query = searchResultsEl?.dataset.query;
	console.log(query);
	const results = document.getElementsByClassName('search-result-content');
	for (const x of results) {
		const inner = x.innerHTML;
		if (inner && query) {
			if (inner.toLowerCase().includes(query.toLowerCase())) {
				const replaced = highlight(inner, query);
				x.innerHTML = replaced;
			}
		}
		console.log(x.innerHTML);
	}
}
function highlight(mainText: string, query: string) {
	const re = new RegExp(query, 'gi'); // global, insensitive
	const newText = mainText.replace(re, `<b>$&</b>`);
	return newText;
}
