package kustolang

import "testing"

func TestCurrentUnionAcceptsBareZeroArgumentLocalViews(t *testing.T) {
	code := ParseAndAnalyze(`let EventsA=view(){print User="a", Score=1};
let EventsB=view(){print User="b", Score=2};
union EventsA, EventsB | project User, Score`, NewGlobalState(nil))

	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("local-view union diagnostics = %#v", diagnostics)
	}
	result, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("local-view union result type = %#v", code.ResultType)
	}
	assertOrderedColumns(t, result, "User", "Score")
	if result.Columns()[0].Type != ScalarTypes.String.Type() || result.Columns()[1].Type != ScalarTypes.Long.Type() || result.IsOpen() {
		t.Fatalf("local-view union result = %#v", result)
	}
}

func TestCurrentPatternCallUsesDeclaredFunctionBody(t *testing.T) {
	code := ParseAndAnalyze(`declare pattern P = (x:string)[path:string] {
	  ("a").["one"] = { print y=strcat(x,path) };
};
P("a").one`, NewGlobalState(nil))

	if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("pattern call diagnostics = %#v", diagnostics)
	}
	result, ok := TableSymbolForType(code.ResultType)
	if !ok {
		t.Fatalf("pattern call result type = %#v", code.ResultType)
	}
	assertOrderedColumns(t, result, "y")
	if result.Columns()[0].Type != ScalarTypes.String.Type() || result.IsOpen() {
		t.Fatalf("pattern call result = %#v", result)
	}
}
