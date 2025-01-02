import { Editor } from '@tiptap/core';
import StarterKit from '@tiptap/starter-kit';
import Document from '@tiptap/extension-document';
import Paragraph from '@tiptap/extension-paragraph';
import Text from '@tiptap/extension-text';

new Editor({
	element: document.querySelector('.element'),
	extensions: StarterKit.configure({
		heading: {
			levels: [1, 2, 3],
		},
	}),
	content: '<p>Hello World!</p>',
});
