import { Editor } from '@tiptap/core';
import Document from '@tiptap/extension-document';
import Paragraph from '@tiptap/extension-paragraph';
import Text from '@tiptap/extension-text';
import Heading from '@tiptap/extension-heading';
import ListItem from '@tiptap/extension-list-item';
import OrderedList from '@tiptap/extension-ordered-list';
import BulletList from '@tiptap/extension-bullet-list';
import BubbleMenu from '@tiptap/extension-bubble-menu';
import Bold from '@tiptap/extension-bold';
import Italic from '@tiptap/extension-italic';
import Underline from '@tiptap/extension-underline';
import Gapcursor from '@tiptap/extension-gapcursor';
import Placeholder from '@tiptap/extension-placeholder';
import CharacterCount from '@tiptap/extension-character-count';

const editor = new Editor({
	element: document.querySelector('.element'),
	extensions: [
		Document,
		Paragraph,
		Text,
		Heading.configure({
			levels: [1],
		}),
		ListItem,
		BulletList,
		OrderedList,
		BubbleMenu.configure({
			element: document.querySelector('.tiptap-editor-menu'),
		}),
		Bold,
		Italic,
		Underline,
		Gapcursor,
		Placeholder.configure({
			placeholder: 'Write something here…',
		}),
		CharacterCount.configure({
			limit: 2000,
		}),
	],
	content: '',
});

window.addEventListener('click', (evt) => {
	if (evt.target === document.getElementById('input-button-bold')) {
		editor.chain().focus().toggleBold().run();
	}
	if (evt.target === document.getElementById('input-button-italic')) {
		editor.chain().focus().toggleItalic().run();
	}
	if (evt.target === document.getElementById('input-button-underline')) {
		editor.chain().focus().toggleUnderline().run();
	}
});

// editor.on('selectionUpdate', ({ editor }) => {});

const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLTextAreaElement; //as HTMLTextAreaElement;

editor.on('create', () => {
	calculateCharsRemaining();
});
editor.on('update', ({ editor }) => {
	commentFormMessageInput.value = editor.getHTML();
	calculateCharsRemaining();
});

/**
 * Post Form calculation feature for remaining chars.
 */
function calculateCharsRemaining() {
	const commentFormMessageInputEl = document.getElementById('comment-form-message-input') as HTMLInputElement;
	const commentFormCharsRemainingEl = document.getElementById('form-message-chars-remaining') as HTMLSpanElement;

	if (commentFormMessageInputEl && commentFormCharsRemainingEl) {
		const total = 2000;
		commentFormCharsRemainingEl.innerHTML = String(total);

		if (commentFormMessageInputEl.value) {
			commentFormCharsRemainingEl.innerHTML = String(total - editor.storage.characterCount.characters());
		}

		commentFormMessageInputEl.addEventListener('keyup', () => {
			commentFormCharsRemainingEl.innerHTML = String(total - editor.storage.characterCount.characters());
		});
	}
}
