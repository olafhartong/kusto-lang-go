package kustolang

import (
	"testing"
)

func TestNativeEvaluatePluginCustomReturnAndStarSchema(t *testing.T) {
	properties := NewColumnSymbol("Properties", ScalarTypes.DynamicBag.Type())
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", properties, account))))
	evaluate := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `evaluate bag_unpack(Properties) : (*, Extra:string)`).(*EvaluateOperator)
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.binder.rowScope = NewTableSymbol("T", properties, account)
	tree.visitEvaluate(evaluate)
	call := evaluate.FunctionCall()
	if call.ReferencedSignature() == nil || call.ReferencedSignature().Symbol != PlugIns.BagUnpack {
		t.Fatalf("evaluate call signature = %#v, want bag_unpack", call.ReferencedSignature())
	}
	payload, _ := call.SyntaxNode.Payload(functionCallResultSymbolPayload)
	pluginResult, ok := payload.(*TableSymbol)
	if !ok || !pluginResult.IsOpen() {
		t.Fatalf("custom plugin result = %T %#v, want open table", payload, payload)
	}
	info, _ := evaluate.GetSemanticInfo()
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok || table.GetColumn("Account") == nil || table.GetColumn("Extra") == nil || table.GetColumn("Properties") != nil {
		t.Fatalf("evaluate info/schema columns = %#v / %#v", info, table)
	}
	if declaration := evaluate.Schema().Schema().Columns().At(0).Element(); declaration.Name().ReferencedSymbol() != table.GetColumn("Extra") {
		t.Fatalf("explicit schema declaration was not attached to output column")
	}
}

func TestNativeEvaluateAINamedConstantPreservesExpressionWrapper(t *testing.T) {
	tests := []struct {
		name      string
		argument  string
		wantError bool
	}{
		{name: "named true", argument: "IncludeErrorMessages=true"},
		{name: "named false", argument: "IncludeErrorMessages=false"},
		{name: "positional true", argument: "true", wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(`datatable(Text:string,Keep:long)["x",1] | evaluate ai_embeddings(Text, "conn", dynamic({}), ` + tt.argument + `)`)
			NewBinder(NewGlobalState(nil)).TryBind(query)
			evaluate := firstBoundNode[*EvaluateOperator](t, query)
			table, ok := TableSymbolForType(evaluate.RawResultType())
			if !ok || table == nil {
				t.Fatalf("AI embeddings result = %#v, want table", evaluate.RawResultType())
			}
			if table.GetColumn("Text_embeddings") == nil {
				t.Fatalf("AI embeddings output columns = %#v", table.Columns())
			}
			if gotError := table.GetColumn("Text_embeddings_error") != nil; gotError != tt.wantError {
				t.Fatalf("AI embeddings error column present = %t, want %t; columns %#v", gotError, tt.wantError, table.Columns())
			}
		})
	}
}

func TestNativeEvaluateUsesPluginScopeAndRestoresState(t *testing.T) {
	binder := NewBinder(nil)
	binder.scopeKind = ScopeKindAggregate
	evaluate := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "evaluate strlen('x')").(*EvaluateOperator)
	newTreeBinder(binder).visitEvaluate(evaluate)
	if binder.scopeKind != ScopeKindAggregate {
		t.Fatalf("scope after evaluate = %v, want Aggregate", binder.scopeKind)
	}
	if diagnostics := evaluate.FunctionCall().Name().SemanticDiagnostics(); !diagnosticsContain(diagnostics, "KS124") {
		t.Fatalf("non-plugin evaluate name diagnostics = %#v, want KS124", diagnostics)
	}
}

func TestNativeEvaluateParameterNamesUsePluginScope(t *testing.T) {
	properties := NewColumnSymbol("Properties", ScalarTypes.DynamicBag.Type())
	existing := NewColumnSymbol("Existing", ScalarTypes.Long.Type())
	row := NewTableSymbol("T", properties, existing)
	evaluate := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `evaluate hint.pass_filters=true hint.pass_filters_column=Existing hint.pass_filters_column=Properties bag_unpack(Properties)`).(*EvaluateOperator)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", row)))
	binder.rowScope = row
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.visitEvaluate(evaluate)
	for i, code := range []string{"", "KS237", "KS237"} {
		expression := evaluate.Parameters().At(i).Expression()
		diagnostics := expression.(interface{ SemanticDiagnostics() []Diagnostic }).SemanticDiagnostics()
		if code == "" {
			if len(diagnostics) != 0 {
				t.Fatalf("parameter %d diagnostics = %#v", i, diagnostics)
			}
		} else if !diagnosticsContain(diagnostics, code) {
			t.Fatalf("parameter %d expression %T diagnostics = %#v, want %s", i, expression, diagnostics, code)
		}
	}
	argument := evaluate.FunctionCall().ArgumentList().Expressions().At(0).Element().(*NameReference)
	if argument.ReferencedSymbol() != properties || diagnosticsContain(argument.SemanticDiagnostics(), "KS237") {
		t.Fatalf("plugin argument = symbol %#v diagnostics %#v", argument.ReferencedSymbol(), argument.SemanticDiagnostics())
	}
}

func evaluateDiagnosticCodes(diagnostics []Diagnostic) []string {
	codes := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		codes[i] = diagnostic.Code
	}
	return codes
}

func evaluateColumnNames(columns []*ColumnSymbol) []string {
	names := make([]string, len(columns))
	for i, column := range columns {
		names[i] = column.SymbolName()
	}
	return names
}

func TestNativeEvaluateRestrictedPluginSetAndMalformedDisabledCall(t *testing.T) {
	globals := NewGlobalState(nil).WithPlugIns([]*FunctionSymbol{PlugIns.Identity})
	enabled := ParseQuery(`datatable(x:long)[1] | evaluate identity()`)
	NewBinder(globals).TryBind(enabled)
	if diagnostics := GetContainedDiagnostics(enabled, DiagnosticsIncludeDefault); len(diagnostics) != 0 {
		t.Fatalf("enabled identity diagnostics = %#v", diagnostics)
	}

	disabled := ParseQuery(`datatable(d:dynamic)[dynamic({})] | evaluate bag_unpack(d`)
	NewBinder(globals).TryBind(disabled)
	diagnostics := GetContainedDiagnostics(disabled, DiagnosticsIncludeDefault)
	if !diagnosticsContain(diagnostics, "KS125") || !diagnosticsContain(diagnostics, "KS005") || diagnosticsContain(diagnostics, "KS124") {
		t.Fatalf("malformed disabled plugin diagnostics = %#v, want KS125 and KS005 without KS124", diagnostics)
	}
	call := firstBoundNode[*FunctionCallExpression](t, disabled)
	if call.Name().SemanticDiagnostics()[0].Start != 45 || call.Name().SemanticDiagnostics()[0].Length != 10 {
		t.Fatalf("disabled plugin diagnostic location = %#v, want 45/10", call.Name().SemanticDiagnostics())
	}
}
