package kustolang

import "testing"

func TestProjectionBuilderAddMakesDuplicateNamesUnique(t *testing.T) {
	builder := NewProjectionBuilder()
	first := NewColumnSymbol("Value", ScalarTypes.String.Type())
	second := NewColumnSymbol("Value", ScalarTypes.Long.Type())

	addedFirst := builder.Add(first)
	addedSecond := builder.Add(second)
	projection := builder.GetProjection()

	if addedFirst != first {
		t.Fatalf("first add returned %#v, want original", addedFirst)
	}
	if len(projection) != 2 {
		t.Fatalf("projection length = %d, want 2", len(projection))
	}
	if projection[0] != first || projection[0].SymbolName() != "Value" {
		t.Fatalf("first projection column = %#v", projection[0])
	}
	if addedSecond.SymbolName() != "Value1" || projection[1] != addedSecond {
		t.Fatalf("second projection column = %#v", addedSecond)
	}
	if len(addedSecond.OriginalColumns) != 1 || addedSecond.OriginalColumns[0] != second {
		t.Fatalf("second original columns = %#v, want original second", addedSecond.OriginalColumns)
	}
}

func TestProjectionBuilderAddUsesExplicitBaseNameForDuplicate(t *testing.T) {
	builder := NewProjectionBuilder()
	builder.Add(NewColumnSymbol("Value", ScalarTypes.String.Type()))

	renamed := builder.Add(NewColumnSymbol("Value", ScalarTypes.Long.Type()), ProjectionAddOptions{
		BaseName:    "Generated",
		HasBaseName: true,
	})

	if renamed.SymbolName() != "Generated1" {
		t.Fatalf("renamed duplicate = %q, want Generated1", renamed.SymbolName())
	}
}

func TestProjectionBuilderAddReplaceUpdatesExistingColumn(t *testing.T) {
	builder := NewProjectionBuilder()
	first := NewColumnSymbol("Value", ScalarTypes.String.Type())
	second := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	builder.Add(first)

	added := builder.Add(second, ProjectionAddOptions{Replace: true})
	projection := builder.GetProjection()

	if added != second || len(projection) != 1 || projection[0] != second {
		t.Fatalf("replace result = added:%#v projection:%#v", added, projection)
	}
}

func TestProjectionBuilderDeclareRejectsDuplicateDeclaredName(t *testing.T) {
	builder := NewProjectionBuilder()
	location := MustNewNameReference([]SyntaxElement{SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Value"))})
	var diagnostics []Diagnostic

	builder.Declare(NewColumnSymbol("Value", ScalarTypes.String.Type()), &diagnostics, location)
	builder.Declare(NewColumnSymbol("Value", ScalarTypes.Long.Type()), &diagnostics, location)
	projection := builder.GetProjection()

	if len(projection) != 1 || projection[0].Type != ScalarTypes.String.Type() {
		t.Fatalf("projection = %#v, want first declared column only", projection)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS171" || diagnostics[0].Start != 0 || diagnostics[0].Length != location.Width() {
		t.Fatalf("diagnostics = %#v, want located KS171", diagnostics)
	}
}

func TestProjectionBuilderDeclareCanReplaceUndeclaredColumnOnce(t *testing.T) {
	builder := NewProjectionBuilder()
	var diagnostics []Diagnostic
	builder.Add(NewColumnSymbol("Value", ScalarTypes.String.Type()))

	replacement := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	builder.Declare(replacement, &diagnostics, nil, true)
	builder.Declare(NewColumnSymbol("Value", ScalarTypes.Real.Type()), &diagnostics, nil, true)
	projection := builder.GetProjection()

	if len(projection) != 1 || projection[0] != replacement {
		t.Fatalf("projection = %#v, want replacement column", projection)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS171" {
		t.Fatalf("diagnostics = %#v, want one duplicate declaration", diagnostics)
	}
}

func TestProjectionBuilderRenameExistingColumn(t *testing.T) {
	builder := NewProjectionBuilder()
	source := NewColumnSymbol("Old", ScalarTypes.String.Type())
	builder.Add(source)
	var diagnostics []Diagnostic

	renamed := builder.Rename("Old", "New", &diagnostics, nil)
	missing := builder.Rename("Missing", "Other", &diagnostics, nil)
	projection := builder.GetProjection()

	if renamed == nil || renamed.SymbolName() != "New" || len(renamed.OriginalColumns) != 1 || renamed.OriginalColumns[0] != source {
		t.Fatalf("renamed column = %#v", renamed)
	}
	if missing != nil {
		t.Fatalf("missing rename returned %#v, want nil", missing)
	}
	if len(projection) != 1 || projection[0] != renamed || len(diagnostics) != 0 {
		t.Fatalf("projection=%#v diagnostics=%#v", projection, diagnostics)
	}
}

func TestProjectionBuilderRenameRejectsDeclaredTargetName(t *testing.T) {
	builder := NewProjectionBuilder()
	var diagnostics []Diagnostic
	builder.Add(NewColumnSymbol("Old", ScalarTypes.String.Type()))
	builder.Declare(NewColumnSymbol("Declared", ScalarTypes.Long.Type()), &diagnostics, nil)

	renamed := builder.Rename("Old", "Declared", &diagnostics, nil)
	projection := builder.GetProjection()

	if renamed != nil {
		t.Fatalf("renamed = %#v, want nil", renamed)
	}
	if len(projection) != 2 || projection[0].SymbolName() != "Old" {
		t.Fatalf("projection = %#v, want unchanged Old column", projection)
	}
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS171" {
		t.Fatalf("diagnostics = %#v, want KS171", diagnostics)
	}
}

func TestProjectionBuilderDoNotAddAndDoNotRepeatUseColumnIdentity(t *testing.T) {
	builder := NewProjectionBuilder()
	blocked := NewColumnSymbol("Blocked", ScalarTypes.String.Type())
	builder.DoNotAdd(blocked)
	if builder.CanAdd(blocked) {
		t.Fatal("blocked column should not be addable")
	}
	if got := builder.Add(blocked); got != blocked || len(builder.GetProjection()) != 0 {
		t.Fatalf("blocked add = %#v projection=%#v", got, builder.GetProjection())
	}

	repeated := NewColumnSymbol("Repeated", ScalarTypes.Long.Type())
	builder.Add(repeated, ProjectionAddOptions{DoNotRepeat: true})
	builder.Add(repeated)
	builder.Add(NewColumnSymbol("Repeated", ScalarTypes.Long.Type()))
	projection := builder.GetProjection()

	if len(projection) != 2 || projection[0].SymbolName() != "Repeated" || projection[1].SymbolName() != "Repeated1" {
		t.Fatalf("projection = %#v, want identity-based do-not-repeat and renamed equivalent", projection)
	}
}

func TestProjectionBuilderAddRangeDoNotRepeatRetainsRenamedDuplicateIdentityEdge(t *testing.T) {
	builder := NewProjectionBuilder()
	first := NewColumnSymbol("x", ScalarTypes.String.Type())
	second := NewColumnSymbol("x", ScalarTypes.Long.Type())
	y := NewColumnSymbol("y", ScalarTypes.Bool.Type())
	columns := []*ColumnSymbol{first, second, y}

	builder.AddRange(columns, ProjectionAddRangeOptions{DoNotRepeat: true})
	builder.AddRange(columns, ProjectionAddRangeOptions{DoNotRepeat: true})
	projection := builder.GetProjection()

	if len(projection) != 4 {
		t.Fatalf("projection length = %d, want 4", len(projection))
	}
	for index, name := range []string{"x", "x1", "y", "x2"} {
		if projection[index].SymbolName() != name {
			t.Fatalf("projection[%d] name = %q, want %q", index, projection[index].SymbolName(), name)
		}
	}
	if projection[0] != first || projection[2] != y {
		t.Fatalf("unchanged projection identity = %#v", projection)
	}
	for _, index := range []int{1, 3} {
		if len(projection[index].OriginalColumns) != 1 || projection[index].OriginalColumns[0] != second {
			t.Fatalf("projection[%d] original columns = %#v, want second x", index, projection[index].OriginalColumns)
		}
	}
}

func TestProjectionBuilderClearResetsState(t *testing.T) {
	builder := NewProjectionBuilder()
	builder.Add(NewColumnSymbol("Value", ScalarTypes.String.Type()))
	builder.Clear()

	added := builder.Add(NewColumnSymbol("Value", ScalarTypes.Long.Type()))
	if projection := builder.GetProjection(); len(projection) != 1 || projection[0] != added || added.SymbolName() != "Value" {
		t.Fatalf("projection after clear = %#v added=%#v", projection, added)
	}
}
