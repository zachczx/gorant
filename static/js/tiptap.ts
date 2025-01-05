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
import TextAlign, { TextAlignOptions } from '@tiptap/extension-text-align';
import CharacterCount from '@tiptap/extension-character-count';

window.addEventListener('load', () => {
	initTiptap();
});
window.addEventListener('htmx:afterSwap', () => {
	initTiptap();
});

function initTiptap() {
	const editor = new Editor({
		element: document.querySelector('#element') as HTMLDivElement,
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
				updateDelay: 0,
				// shouldShow: () => {
				// 	return true;
				// },
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
			case document.getElementById('input-button-h1'):
				editor.chain().focus().toggleHeading({ level: 1 }).run();
				setBubbleMenuButtonColor(editor, 'heading', document.getElementById('input-button-h1') as HTMLButtonElement);
				break;
			case document.getElementById('input-button-left'):
				editor.chain().focus().setTextAlign('left').run();
				setBubbleMenuButtonColor(
					editor,
					{ textAlign: 'left' },
					document.getElementById('input-button-left') as HTMLButtonElement,
				);
				break;
			case document.getElementById('input-button-center'):
				editor.chain().focus().setTextAlign('center').run();
				setBubbleMenuButtonColor(
					editor,
					{ textAlign: 'center' },
					document.getElementById('input-button-center') as HTMLButtonElement,
				);
				break;
			case document.getElementById('input-button-right'):
				editor.chain().focus().setTextAlign('right').run();
				setBubbleMenuButtonColor(
					editor,
					{ textAlign: 'right' },
					document.getElementById('input-button-right') as HTMLButtonElement,
				);
				break;
		}
	});

	const commentFormMessageInput = document.getElementById('comment-form-message-input') as HTMLTextAreaElement; //as HTMLTextAreaElement;

	editor.on('create', ({ editor }) => {
		showChars(editor);
		populateExistingContentForEdit(editor);
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
	window.addEventListener('htmx:afterSwap', () => {
		// Destroy the instance after the swap, else there'll be 2 Tiptap editors.
		editor.destroy();
	});
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
	setBubbleMenuButtonColor(editor, 'heading', document.getElementById('input-button-h1') as HTMLButtonElement);
	setBubbleMenuButtonColor(
		editor,
		{ textAlign: 'left' },
		document.getElementById('input-button-left') as HTMLButtonElement,
	);
	setBubbleMenuButtonColor(
		editor,
		{ textAlign: 'center' },
		document.getElementById('input-button-center') as HTMLButtonElement,
	);
	setBubbleMenuButtonColor(
		editor,
		{ textAlign: 'right' },
		document.getElementById('input-button-right') as HTMLButtonElement,
	);
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
