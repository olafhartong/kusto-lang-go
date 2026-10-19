package kustolang

import "testing"

func TestOffsetCodeServiceGetCodeActionsClampsSelectionToUTF16TextLength(t *testing.T) {
	const offset = 7

	for _, test := range []struct {
		name string
		text string
	}{
		{name: "ASCII", text: "abcd"},
		{name: "astral", text: "a😀b"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var gotPosition, gotSelectionStart, gotSelectionLength int
			service := NewCustomCodeService(test.text, CustomCodeServiceCallbacks{
				GetCodeActions: func(position, selectionStart, selectionLength int, _ CodeActionOptions, _ bool, _ string, _ CancellationToken) CodeActionInfo {
					gotPosition = position
					gotSelectionStart = selectionStart
					gotSelectionLength = selectionLength
					return CodeActionInfo{}
				},
			})

			textLength := utf16TextWidth(test.text)
			NewOffsetCodeService(service, offset).GetCodeActions(
				offset+textLength,
				offset+2,
				5,
				CodeActionOptionsDefault,
				true,
				"recorder",
				CancellationToken{},
			)

			if gotPosition != textLength {
				t.Fatalf("translated position = %d, want %d", gotPosition, textLength)
			}
			if gotSelectionStart != 2 || gotSelectionLength != 2 {
				t.Fatalf("translated selection = (%d, %d), want (2, 2)", gotSelectionStart, gotSelectionLength)
			}
			if gotPosition > textLength || gotSelectionStart+gotSelectionLength > textLength {
				t.Fatalf("translated coordinates exceed UTF-16 text length %d: position=%d selection=(%d, %d)", textLength, gotPosition, gotSelectionStart, gotSelectionLength)
			}
		})
	}
}
