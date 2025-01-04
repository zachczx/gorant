const navTab = document.getElementById('nav-tab') as HTMLDivElement;
const tabs = document.getElementsByClassName('tab') as HTMLCollectionOf<HTMLAnchorElement>;
const tabPages = document.getElementsByClassName('tab-page') as HTMLCollectionOf<HTMLDivElement>;

navTab.addEventListener('click', (evt: Event) => {
	console.log((evt.target as HTMLAnchorElement)?.dataset.page);
	for (const t of tabs) {
		t.classList.remove('tab-active');
	}
	(evt.target as HTMLAnchorElement)?.classList.add('tab-active');
	const page = (evt.target as HTMLAnchorElement)?.dataset.page;
	if (page) {
		for (const p of tabPages) {
			if (p.dataset.page === page) {
				p.classList.remove('hidden');
			} else {
				p.classList.add('hidden');
			}
		}
	}
});
