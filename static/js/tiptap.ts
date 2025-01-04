import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
// import Document from '@tiptap/extension-document';
// import Paragraph from '@tiptap/extension-paragraph';
// import Text from '@tiptap/extension-text';
import BubbleMenu from '@tiptap/extension-bubble-menu';
import Underline from '@tiptap/extension-underline';

const editor = new Editor({
	element: document.querySelector('.element'),
	extensions: [
		StarterKit.configure({
			heading: {
				levels: [1, 2],
			},
		}),
		BubbleMenu.configure({
			element: document.querySelector('.menu'),
		}),
		Underline,
	],
	content: '<h1>Hello WorldAAA!</h1><h2>Hello WorldAAA!</h2><p>Hello WorldAAA!</p>',
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

editor.on('selectionUpdate', () => {
	console.log('selection changed');
});
