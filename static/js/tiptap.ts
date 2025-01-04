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
import Link from '@tiptap/extension-link';
import Strike from '@tiptap/extension-strike';
import TextAlign from '@tiptap/extension-text-align';
import CharacterCount from '@tiptap/extension-character-count';

const editor = new Editor({
	element: document.querySelector('.element') as HTMLDivElement,
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
			element: document.querySelector('.tiptap-editor-menu') as HTMLDivElement,
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
		TextAlign.configure({
			alignments: ['left', 'center', 'right'],
			types: ['heading', 'paragraph'],
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
	if (evt.target === document.getElementById('input-button-strike')) {
		editor.chain().focus().toggleStrike().run();
	}
	if (evt.target === document.getElementById('input-button-left')) {
		editor.chain().focus().setTextAlign('left').run();
	}
	if (evt.target === document.getElementById('input-button-center')) {
		editor.chain().focus().setTextAlign('center').run();
	}
	if (evt.target === document.getElementById('input-button-right')) {
		editor.chain().focus().setTextAlign('right').run();
	}
});

// editor.on('selectionUpdate', ({ editor }) => {});

const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLTextAreaElement; //as HTMLTextAreaElement;

editor.on('create', () => {
	showChars();
});
editor.on('update', ({ editor }) => {
	commentFormMessageInput.value = editor.getHTML();
	showChars();
});

/**
 * Post Form calculation feature for remaining chars.
 */
function showChars() {
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
