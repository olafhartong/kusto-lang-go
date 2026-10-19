package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicEditorResultConstructorsSnapshotCollections(t *testing.T) {
	outlineRanges := []kustolang.OutlineRange{kustolang.NewOutlineRange(1, 4, "...")}
	outline := kustolang.NewOutlineInfo(outlineRanges)
	outlineRanges[0] = kustolang.NewOutlineRange(9, 1, "changed")
	if !reflect.DeepEqual(outline.Ranges, []kustolang.OutlineRange{kustolang.NewOutlineRange(1, 4, "...")}) {
		t.Fatalf("outline = %#v", outline)
	}

	classifiedRanges := []kustolang.ClassifiedRange{kustolang.NewClassifiedRange(kustolang.ClassificationKindKeyword, 2, 5)}
	classification := kustolang.NewClassificationInfo(classifiedRanges)
	classifiedRanges[0] = kustolang.NewClassifiedRange(kustolang.ClassificationKindComment, 0, 1)
	if !reflect.DeepEqual(classification.Classifications, []kustolang.ClassifiedRange{kustolang.NewClassifiedRange(kustolang.ClassificationKindKeyword, 2, 5)}) {
		t.Fatalf("classification = %#v", classification)
	}

	parts := []kustolang.ClassifiedText{kustolang.NewClassifiedTextKind(kustolang.ClassificationKindColumn, "Column")}
	item := kustolang.NewQuickInfoItem(kustolang.QuickInfoKindTable, parts...)
	parts[0] = kustolang.NewClassifiedText("changed")
	items := []kustolang.QuickInfoItem{item}
	quickInfo := kustolang.NewQuickInfoItems(items)
	items[0] = kustolang.NewQuickInfoItemText(kustolang.QuickInfoKindText, "changed")
	if quickInfo.Text() != "(table) Column" {
		t.Fatalf("quick info text = %q", quickInfo.Text())
	}

	elements := []kustolang.RelatedElement{kustolang.NewRelatedElement(3, 4, kustolang.RelatedElementKindReference)}
	related := kustolang.NewRelatedInfo(elements, 0)
	elements[0] = kustolang.NewRelatedElement(0, 1, kustolang.RelatedElementKindOther)
	if len(related.Elements) != 1 || related.Elements[0].Start() != 3 || related.Elements[0].CursorLeft != 3 || related.Elements[0].CursorRight != 7 {
		t.Fatalf("related info = %#v", related)
	}

	diagnostic := kustolang.NewDiagnostic("KS900", "test")
	diagnostics := []kustolang.Diagnostic{diagnostic}
	analyzer := kustolang.NewAnalyzerInfo("sample", diagnostics)
	diagnostics[0] = kustolang.NewDiagnostic("KS901", "changed")
	if analyzer.Name != "sample" || !reflect.DeepEqual(analyzer.Diagnostics, []kustolang.Diagnostic{diagnostic}) {
		t.Fatalf("analyzer info = %#v", analyzer)
	}
}

func TestPublicQuickInfoDefendsNestedParts(t *testing.T) {
	parts := []kustolang.ClassifiedText{kustolang.NewClassifiedText("Events")}
	items := []kustolang.QuickInfoItem{{Kind: kustolang.QuickInfoKindTable, Parts: parts}}
	info := kustolang.NewQuickInfoItems(items)

	parts[0] = kustolang.NewClassifiedText("changed source")
	items[0].Parts[0] = kustolang.NewClassifiedText("changed item")
	if got := info.Text(); got != "(table) Events" {
		t.Fatalf("quick info retained caller storage: %q", got)
	}

	copyItems := info.ItemsList()
	copyItems[0].Parts[0] = kustolang.NewClassifiedText("changed copy")
	if got := info.Text(); got != "(table) Events" {
		t.Fatalf("ItemsList exposed nested storage: %q", got)
	}
	copyParts := info.Items[0].PartsList()
	copyParts[0] = kustolang.NewClassifiedText("changed parts")
	if got := info.Items[0].Text(); got != "Events" {
		t.Fatalf("PartsList exposed storage: %q", got)
	}
}

func TestPublicEditorResultDefaultsAndReferences(t *testing.T) {
	emptyCollections := []struct {
		name   string
		value  any
		length int
	}{
		{name: "outline", value: kustolang.OutlineInfoEmpty.Ranges, length: len(kustolang.OutlineInfoEmpty.Ranges)},
		{name: "classification", value: kustolang.ClassificationInfoEmpty.Classifications, length: len(kustolang.ClassificationInfoEmpty.Classifications)},
		{name: "quick info", value: kustolang.QuickInfoEmpty.Items, length: len(kustolang.QuickInfoEmpty.Items)},
		{name: "related", value: kustolang.RelatedInfoEmpty.Elements, length: len(kustolang.RelatedInfoEmpty.Elements)},
	}
	for _, tt := range emptyCollections {
		if reflect.ValueOf(tt.value).IsNil() || tt.length != 0 {
			t.Fatalf("%s empty value = %#v", tt.name, tt.value)
		}
	}
	if kustolang.TextRangeEmpty != kustolang.NewTextRange(0, 0) {
		t.Fatalf("text range empty = %#v", kustolang.TextRangeEmpty)
	}
	if got := kustolang.NewOutlineRangeOptional(2, 3, nil); got.CollapsedText != "" {
		t.Fatalf("optional outline range = %#v", got)
	}
	if got := kustolang.NewQuickInfoText("plain"); got.Text() != "plain" || len(got.Items) != 1 {
		t.Fatalf("text quick info = %#v text=%q", got, got.Text())
	}

	formatted := kustolang.NewFormattedText("formatted", 7)
	if formatted.Text != "formatted" || formatted.Position != 7 {
		t.Fatalf("formatted text = %#v", formatted)
	}
	rng := kustolang.TextRangeFromBounds(4, 9)
	syntax := kustolang.NewSyntaxReferenceRange(rng)
	cluster := kustolang.NewClusterReference("help", 4, 5)
	database := kustolang.NewDatabaseReference("Samples", "help", 4, 5)
	if syntax.Start() != 4 || syntax.End() != 9 || cluster.Cluster != "help" || database.Database != "Samples" || database.Cluster != "help" {
		t.Fatalf("references = syntax:%#v cluster:%#v database:%#v", syntax, cluster, database)
	}

	index := 2
	parameter := kustolang.NewClientParameter("p", &index, rng)
	index = 9
	if parameter.Index == nil || *parameter.Index != 2 || parameter.String() != "p[2]" || parameter.Start() != 4 || parameter.Length() != 5 {
		t.Fatalf("indexed parameter = %#v", parameter)
	}
	plain := kustolang.NewClientParameterReference("q", 1, 2)
	if plain.Index != nil || plain.String() != "q" || plain.Start() != 1 || plain.Length() != 2 {
		t.Fatalf("plain parameter = %#v", plain)
	}
}
