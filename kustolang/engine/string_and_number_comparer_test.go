package kustolang

import "testing"

func TestStringAndNumberCompareOrdersNumericSegmentsNumerically(t *testing.T) {
	if got := StringAndNumberCompareOrdinal("item2", "item10"); got >= 0 {
		t.Fatalf("item2 compare item10 = %d, want negative", got)
	}
	if got := StringAndNumberCompareOrdinal("item10", "item2"); got <= 0 {
		t.Fatalf("item10 compare item2 = %d, want positive", got)
	}
	if got := StringAndNumberCompareOrdinal("item10", "item10"); got != 0 {
		t.Fatalf("item10 compare item10 = %d, want zero", got)
	}
}

func TestStringAndNumberCompareTextSegmentsAndCase(t *testing.T) {
	if got := StringAndNumberCompareOrdinalIgnoreCase("Alpha2", "alpha2"); got != 0 {
		t.Fatalf("ignore-case compare = %d, want zero", got)
	}
	if got := StringAndNumberCompareOrdinal("Alpha2", "alpha2"); got >= 0 {
		t.Fatalf("case-sensitive compare = %d, want uppercase before lowercase", got)
	}
	if got := StringAndNumberCompareOrdinal("alpha", "alpha1"); got >= 0 {
		t.Fatalf("prefix compare = %d, want shorter string first", got)
	}
}

func TestStringAndNumberCompareDigitVsTextAndLeadingZeros(t *testing.T) {
	if got := StringAndNumberCompareOrdinal("a1", "aa"); got >= 0 {
		t.Fatalf("digit/text compare = %d, want digit before letter", got)
	}
	if got := StringAndNumberCompareOrdinal("a02", "a2"); got <= 0 {
		t.Fatalf("leading-zero compare = %d, want longer numeric segment after shorter segment like upstream", got)
	}
	if got := StringAndNumberCompareOrdinal("a002", "a02"); got <= 0 {
		t.Fatalf("multi-leading-zero compare = %d, want longer numeric segment after shorter segment", got)
	}
}

func TestStringAndNumberCompareUsesUTF16CodeUnits(t *testing.T) {
	// Supplementary-plane decimal digits are surrogate pairs. The pinned C#
	// char-based implementation treats the surrogates as text, not one digit.
	if got := StringAndNumberCompareOrdinal("𝟡2", "𝟙10"); got <= 0 {
		t.Fatalf("supplementary digit comparison = %d, want positive", got)
	}
	// UTF-16 ordinal order differs from Unicode scalar order in this case.
	if got := StringAndNumberCompareOrdinal("𐀀", "\ue000"); got >= 0 {
		t.Fatalf("UTF-16 ordinal comparison = %d, want negative", got)
	}
}
