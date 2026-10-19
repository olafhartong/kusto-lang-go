package kustolang

import "testing"

func TestFormattedTextStoresTextAndPosition(t *testing.T) {
	formatted := NewFormattedText("T | count", 3)
	if formatted.Text != "T | count" || formatted.Position != 3 {
		t.Fatalf("formatted text = %#v", formatted)
	}
}

func TestSyntaxClusterAndDatabaseReferences(t *testing.T) {
	ref := NewSyntaxReference(4, 6)
	if ref.Range != NewTextRange(4, 6) || ref.Start() != 4 || ref.Length() != 6 || ref.End() != 10 {
		t.Fatalf("syntax reference = %#v start=%d length=%d end=%d", ref, ref.Start(), ref.Length(), ref.End())
	}

	rangeRef := NewSyntaxReferenceRange(NewTextRange(2, 3))
	if rangeRef.Start() != 2 || rangeRef.End() != 5 {
		t.Fatalf("range reference = %#v", rangeRef)
	}

	cluster := NewClusterReference("help", 1, 4)
	if cluster.Cluster != "help" || cluster.Start() != 1 || cluster.End() != 5 {
		t.Fatalf("cluster reference = %#v", cluster)
	}

	database := NewDatabaseReference("Samples", "help", 7, 8)
	if database.Database != "Samples" || database.Cluster != "help" || database.Start() != 7 || database.End() != 15 {
		t.Fatalf("database reference = %#v", database)
	}
}

func TestOutlineRangeAndInfoMirrorUpstreamDefaults(t *testing.T) {
	collapsed := NewOutlineRange(10, 20, "...")
	if collapsed.Start != 10 || collapsed.Length != 20 || collapsed.CollapsedText != "..." {
		t.Fatalf("outline range = %#v", collapsed)
	}

	nilText := NewOutlineRangeOptional(1, 2, nil)
	if nilText.CollapsedText != "" {
		t.Fatalf("nil collapsed text = %q, want empty", nilText.CollapsedText)
	}

	ranges := []OutlineRange{collapsed}
	info := NewOutlineInfo(ranges)
	ranges[0].CollapsedText = "changed"
	if len(info.Ranges) != 1 || info.Ranges[0].CollapsedText != "..." {
		t.Fatalf("outline info should copy ranges, got %#v", info.Ranges)
	}
	if OutlineInfoEmpty.Ranges == nil || len(OutlineInfoEmpty.Ranges) != 0 {
		t.Fatalf("empty outline ranges = %#v, want non-nil empty", OutlineInfoEmpty.Ranges)
	}
}

func TestClassificationKindNumericOrderMatchesUpstream(t *testing.T) {
	cases := []struct {
		kind  ClassificationKind
		value int
	}{
		{ClassificationKindPlainText, 0},
		{ClassificationKindComment, 1},
		{ClassificationKindPunctuation, 2},
		{ClassificationKindDirective, 3},
		{ClassificationKindLiteral, 4},
		{ClassificationKindStringLiteral, 5},
		{ClassificationKindType, 6},
		{ClassificationKindColumn, 7},
		{ClassificationKindTable, 8},
		{ClassificationKindDatabase, 9},
		{ClassificationKindFunction, 10},
		{ClassificationKindParameter, 11},
		{ClassificationKindVariable, 12},
		{ClassificationKindIdentifier, 13},
		{ClassificationKindClientParameter, 14},
		{ClassificationKindQueryParameter, 15},
		{ClassificationKindScalarOperator, 16},
		{ClassificationKindMathOperator, 17},
		{ClassificationKindQueryOperator, 18},
		{ClassificationKindCommand, 19},
		{ClassificationKindKeyword, 20},
		{ClassificationKindMaterializedView, 21},
		{ClassificationKindSchemaMember, 22},
		{ClassificationKindSignatureParameter, 23},
		{ClassificationKindOption, 24},
	}
	for _, test := range cases {
		if got := int(test.kind); got != test.value {
			t.Fatalf("classification kind %v numeric value = %d, want %d", test.kind, got, test.value)
		}
	}
}

func TestClassifiedRangeAndClassificationInfo(t *testing.T) {
	rng := NewClassifiedRange(ClassificationKindTable, 3, 5)
	if rng.Kind != ClassificationKindTable || rng.Start != 3 || rng.Length != 5 || rng.End() != 8 {
		t.Fatalf("classified range = %#v end=%d", rng, rng.End())
	}

	classifications := []ClassifiedRange{rng}
	info := NewClassificationInfo(classifications)
	classifications[0].Kind = ClassificationKindColumn
	if len(info.Classifications) != 1 || info.Classifications[0].Kind != ClassificationKindTable {
		t.Fatalf("classification info should copy classifications, got %#v", info.Classifications)
	}
	if ClassificationInfoEmpty.Classifications == nil || len(ClassificationInfoEmpty.Classifications) != 0 {
		t.Fatalf("empty classifications = %#v, want non-nil empty", ClassificationInfoEmpty.Classifications)
	}
}
