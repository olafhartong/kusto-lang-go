package kustolang

import "testing"

func TestKustoCodeServiceCanBeParsedUsesUTF16CodeUnits(t *testing.T) {
	const maxParseTextSize = 4
	globals := GlobalStateWithProperty(DefaultGlobalState, Properties.MaxParseTextSize, maxParseTextSize)

	tests := []struct {
		name string
		text string
		want bool
	}{
		{name: "ASCII below limit", text: "abc", want: true},
		{name: "ASCII at limit", text: "abcd", want: true},
		{name: "ASCII above limit", text: "abcde", want: false},
		{name: "astral below limit", text: "a\U0001F600", want: true},
		{name: "astral at limit", text: "a\U0001F600b", want: true},
		{name: "astral above limit", text: "a\U0001F600bc", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewKustoCodeService(tt.text, globals)
			if got := service.canBeParsed(); got != tt.want {
				t.Fatalf("canBeParsed() = %t, want %t (UTF-8 bytes=%d, UTF-16 units=%d, limit=%d)",
					got, tt.want, len(tt.text), utf16TextWidth(tt.text), maxParseTextSize)
			}
		})
	}
}
