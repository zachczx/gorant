import { Editor } from '@tiptap/core';
import Document from '@tiptap/extension-document';
import Paragraph from '@tiptap/extension-paragraph';
import Text from '@tiptap/extension-text';
import ListItem from '@tiptap/extension-list-item';
import OrderedList from '@tiptap/extension-ordered-list';
import BulletList from '@tiptap/extension-bullet-list';
import BubbleMenu from '@tiptap/extension-bubble-menu';
import Bold from '@tiptap/extension-bold';
import Italic from '@tiptap/extension-italic';
import Underline from '@tiptap/extension-underline';
import Gapcursor from '@tiptap/extension-gapcursor';
import Placeholder from '@tiptap/extension-placeholder';
import Link from '@tiptap/extension-link';
import Strike from '@tiptap/extension-strike';
import CharacterCount from '@tiptap/extension-character-count';

window.addEventListener('load', () => {
	initTiptap();
});

/**
 * This needs to be afterRequest to init tiptap. Otherwise, there'll either be 2 element boxes if it's afterSettle
 * (first one will work, second one doesn't), or no tiptap initialized if I remove this eventlistener ().
 */

window.addEventListener('htmx:afterRequest', ((evt: HtmxAfterRequest) => {
	// Adding a delay because I added a delay for the delete handler to swap new comment list.
	if (evt.detail.requestConfig.verb === 'post') {
		initTiptap();
	}
}) as EventListener);

function initTiptap() {
	const editor = new Editor({
		element: document.getElementById('element') as HTMLDivElement,
		extensions: [
			Document,
			Paragraph,
			Text,
			ListItem,
			BulletList,
			OrderedList,
			BubbleMenu.configure({
				element: document.querySelector('.tiptap-editor-menu') as HTMLDivElement,
				updateDelay: 0,
			}),
			Bold,
			Italic,
			Underline,
			Gapcursor,
			Link.configure({
				openOnClick: false,
			}),
			Placeholder.configure({
				placeholder: 'Write something here…',
			}),
			Strike,
			CharacterCount.configure({
				limit: 2000,
			}),
		],
		content: '',
	});

	// Event listeners for the Bubble Menu bar buttons.
	window.addEventListener('click', (evt) => {
		switch (evt.target) {
			case document.getElementById('input-button-bold'):
				editor.chain().focus().toggleBold().run();
				setBubbleMenuButtonColor(editor, 'bold', document.getElementById('input-button-bold') as HTMLButtonElement);
				break;
			case document.getElementById('input-button-italic'):
				editor.chain().focus().toggleItalic().run();
				setBubbleMenuButtonColor(editor, 'italic', document.getElementById('input-button-italic') as HTMLButtonElement);
				break;
			case document.getElementById('input-button-underline'):
				editor.chain().focus().toggleUnderline().run();
				setBubbleMenuButtonColor(
					editor,
					'underline',
					document.getElementById('input-button-underline') as HTMLButtonElement,
				);
				break;
			case document.getElementById('input-button-strike'):
				editor.chain().focus().toggleStrike().run();
				setBubbleMenuButtonColor(editor, 'strike', document.getElementById('input-button-strike') as HTMLButtonElement);
				break;
		}
	});

	const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLTextAreaElement; //as HTMLTextAreaElement;

	editor.on('create', ({ editor }) => {
		showChars(editor);
		populateExistingContentForEdit(editor);

		const commentFormTiptapPlaceholder = document.getElementById('comment-form-tiptap-placeholder') as HTMLDivElement;
		const delayForBubbleMenuDomCreation = 10;
		if (commentFormTiptapPlaceholder) {
			// Either set this or set another eventlistener for htmx:afterSettle to add .hidden to the div.
			// But this will increase the redundant number of things firing off.
			setTimeout(() => {
				commentFormTiptapPlaceholder.classList.add('hidden');
			}, delayForBubbleMenuDomCreation);
		}
	});
	editor.on('update', ({ editor }) => {
		commentFormMessageInput.value = editor.getHTML();
		showChars(editor);
		setMultipleListenersBubbleMenuButtonColor(editor);
	});
	editor.on('selectionUpdate', ({ editor }) => {
		commentFormMessageInput.value = editor.getHTML();
		showChars(editor);
		setMultipleListenersBubbleMenuButtonColor(editor);
	});
	window.addEventListener('htmx:afterSwap', ((evt: HtmxAfterSwap) => {
		// Destroy the instance after the swap, else there'll be 2 Tiptap editors.
		// But there's no need to destroy it if we're deleting stuff, because there won't be text in the editor.
		if (evt.detail.requestConfig.verb === 'post') {
			editor.destroy();
		}
	}) as EventListener);
}

/**
 * Post Form calculation feature for remaining chars.
 */
function showChars(editor: Editor) {
	const commentFormMessageInputEl = document.getElementById('comment-form-message-input') as HTMLInputElement;
	const commentFormCharsEl = document.getElementById('form-message-chars') as HTMLSpanElement;

	if (commentFormMessageInputEl && commentFormCharsEl) {
		const empty = 0;
		commentFormCharsEl.innerHTML = String(empty);

		if (commentFormMessageInputEl.value) {
			commentFormCharsEl.innerHTML = String(editor.storage.characterCount.characters());
		}

		commentFormMessageInputEl.addEventListener('keyup', () => {
			commentFormCharsEl.innerHTML = String(editor.storage.characterCount.characters());
		});
	}
}

function populateExistingContentForEdit(editor: Editor) {
	const existingContentFromDb = document.getElementById('comment-form-existing-message') as HTMLInputElement;
	if (existingContentFromDb && editor) {
		if (existingContentFromDb.value.length > 0) {
			editor.commands.setContent(existingContentFromDb.value);

			//Reset so that it doesn't reset the edit field subsequently.
			existingContentFromDb.value = '';
		}
	}
}

type textAlign = {
	textAlign: 'left' | 'center' | 'right';
};

/**
 * Uses editor.isActive to check if each mark/node is fulfilled, and if so, highlight the relevant button on Bubble Menu.
 * Doing it in one function instead of copy pasting across events (update, selectionUpdate).
 * @param {Editor} editor - Tiptap's Editor instance.
 */
function setMultipleListenersBubbleMenuButtonColor(editor: Editor) {
	setBubbleMenuButtonColor(editor, 'bold', document.getElementById('input-button-bold') as HTMLButtonElement);
	setBubbleMenuButtonColor(editor, 'italic', document.getElementById('input-button-italic') as HTMLButtonElement);
	setBubbleMenuButtonColor(editor, 'underline', document.getElementById('input-button-underline') as HTMLButtonElement);
	setBubbleMenuButtonColor(editor, 'strike', document.getElementById('input-button-strike') as HTMLButtonElement);
}

/**
 * Check if mark/node is found in the selection, and if so, highlight the relevant button on Bubble Menu.
 */
function setBubbleMenuButtonColor(editor: Editor, mark: string | textAlign, el: HTMLButtonElement) {
	const className = 'bg-primary/50';
	if (editor.isActive(mark)) {
		el?.classList.add(className);
		if (el?.classList.contains(className)) {
			console.log(`Successfully set the ${mark} button to green`);
		}
	} else {
		el?.classList.remove(className);
	}
}
