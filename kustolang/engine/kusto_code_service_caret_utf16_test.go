package kustolang

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestKustoCodeServiceReadOnlyCaretFloorsInteriorSurrogatePosition(t *testing.T) {
	const text = `print value = "a😀b"`
	service := NewKustoCodeService(text)
	emojiByteStart := strings.Index(text, "😀")
	emojiBoundary := utf16TextLength(text[:emojiByteStart])
	interiorSurrogatePosition := emojiBoundary + 1

	wantElement := service.GetElement(emojiBoundary, CancellationToken{})
	gotElement := service.GetElement(interiorSurrogatePosition, CancellationToken{})
	if gotElement != wantElement {
		t.Fatalf("interior-surrogate element = %#v, want boundary element %#v", gotElement, wantElement)
	}
	literalByteStart := strings.Index(text, `"a😀b"`)
	wantLiteralElement := NewTextRange(utf16TextLength(text[:literalByteStart]), utf16TextLength(`"a😀b"`))
	if wantElement != wantLiteralElement {
		t.Fatalf("boundary element = %#v, want string literal %#v", wantElement, wantLiteralElement)
	}

	wantQuickInfo := service.GetQuickInfo(emojiBoundary, QuickInfoOptionsDefault, CancellationToken{})
	gotQuickInfo := service.GetQuickInfo(interiorSurrogatePosition, QuickInfoOptionsDefault, CancellationToken{})
	if !reflect.DeepEqual(gotQuickInfo, wantQuickInfo) {
		t.Fatalf("interior-surrogate quick info = %#v, want boundary quick info %#v", gotQuickInfo, wantQuickInfo)
	}
	if len(wantQuickInfo.Items) != 1 || wantQuickInfo.Items[0].Kind != QuickInfoKindLiteral || wantQuickInfo.Items[0].Text() != "string" {
		t.Fatalf("boundary quick info did not resolve the string literal: %#v", wantQuickInfo)
	}
}

func TestKustoCodeServiceReadOnlyCaretOutOfRangePositionsRemainSafe(t *testing.T) {
	const text = `print value = "a😀b"`
	service := NewKustoCodeService(text)
	textLength := utf16TextLength(text)

	for _, position := range []int{-1, textLength + 1} {
		t.Run(strconv.Itoa(position), func(t *testing.T) {
			_ = service.GetQuickInfo(position, QuickInfoOptionsDefault, CancellationToken{})
			element := service.GetElement(position, CancellationToken{})
			if element.Start < 0 || element.Length < 0 || element.Start+element.Length > textLength {
				t.Fatalf("element for out-of-range position %d = %#v, text length %d", position, element, textLength)
			}
		})
	}
}
