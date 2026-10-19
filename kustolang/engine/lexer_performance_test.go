package kustolang

import "testing"

func scanFixedTextTokenIndependent(text string, pos int) (SyntaxKind, string, bool) {
	remaining := text[pos:]
	for _, data := range syntaxTextsByLength {
		if data.Text == "" || len(data.Text) > len(remaining) {
			continue
		}
		if remaining[:len(data.Text)] != data.Text {
			continue
		}
		if data.Category == SyntaxCategoryKeyword && !keywordBoundary(text, pos, len(data.Text)) {
			continue
		}
		return data.Kind, remaining[:len(data.Text)], true
	}
	return SyntaxKindNone, "", false
}

func TestScanFixedTextTokenFirstByteIndexPreservesResults(t *testing.T) {
	inputs := []string{
		"where ResultType !in~ (0, 1)",
		"project-away Account, Computer",
		"summarize hint.shufflekey = User count() by Host",
		"datatable(x:string)[\"project\"] | extend y=x contains_cs \"p\"",
		"unknown_identifier",
		"$left.Account == $right.Account",
	}
	for _, input := range inputs {
		for pos := 0; pos <= len(input); pos++ {
			gotKind, gotText, gotOK := scanFixedTextToken(input, pos)
			wantKind, wantText, wantOK := scanFixedTextTokenIndependent(input, pos)
			if gotKind != wantKind || gotText != wantText || gotOK != wantOK {
				t.Fatalf("scanFixedTextToken(%q, %d) = (%v, %q, %v), want (%v, %q, %v)", input, pos, gotKind, gotText, gotOK, wantKind, wantText, wantOK)
			}
		}
	}
}

func BenchmarkScanFixedTextToken(b *testing.B) {
	const input = "where ResultType !in~ (0, 1) | project-away Account, Computer"
	positions := make([]int, len(input))
	for i := range positions {
		positions[i] = i
	}

	b.Run("first_byte_index", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, pos := range positions {
				scanFixedTextToken(input, pos)
			}
		}
	})
	b.Run("independent_scan", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, pos := range positions {
				scanFixedTextTokenIndependent(input, pos)
			}
		}
	})
}
