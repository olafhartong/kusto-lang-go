package kustolang

import "testing"

func TestNativeQueryParametersPublishTypedVariablesToFollowingStatements(t *testing.T) {
	query := ParseQuery(`declare query_parameters(User:string="admin", Limit:long=5); print User=User, Limit=Limit`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)

	user := localVariable(t, binder.LocalScope, "User")
	limit := localVariable(t, binder.LocalScope, "Limit")
	if user.Type != ScalarTypes.String.Type() || limit.Type != ScalarTypes.Long.Type() {
		t.Fatalf("query parameter types = User:%#v Limit:%#v", user.Type, limit.Type)
	}
	if !user.IsConstant || !limit.IsConstant {
		t.Fatalf("defaulted query parameters should retain constant metadata: User:%t Limit:%t", user.IsConstant, limit.IsConstant)
	}
	if firstNamedReference(t, query, "User").ReferencedSymbol() != user {
		t.Fatal("following User reference did not resolve to the declared query parameter")
	}
	if firstNamedReference(t, query, "Limit").ReferencedSymbol() != limit {
		t.Fatal("following Limit reference did not resolve to the declared query parameter")
	}
	if got := semanticTable(t, firstBoundNode[*PrintOperator](t, query)); got.GetColumn("User").Type != ScalarTypes.String.Type() || got.GetColumn("Limit").Type != ScalarTypes.Long.Type() {
		t.Fatalf("print result = %#v", got.Columns())
	}
}

func TestNativeQueryParametersLastDuplicateDeclarationWins(t *testing.T) {
	query := ParseQuery(`declare query_parameters(User:string, User:long); print Value=User`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)

	user := localVariable(t, binder.LocalScope, "User")
	if user.Type != ScalarTypes.Long.Type() {
		t.Fatalf("duplicate query parameter type = %#v, want long", user.Type)
	}
	if firstNamedReference(t, query, "User").ReferencedSymbol() != user {
		t.Fatal("following reference did not resolve to the last duplicate declaration")
	}
}

func TestNativeQueryParametersAreIsolatedWhileDefaultsBind(t *testing.T) {
	query := ParseQuery(`declare query_parameters(A:long=1, B:long=A); print Value=B`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)

	parameters := boundNodes[*FunctionParameter](query)
	if len(parameters) != 2 || !diagnosticsContain(parameters[1].SemanticDiagnostics(), "KS132") {
		t.Fatalf("parameter diagnostics = %#v, want second parameter KS132", GetContainedDiagnostics(query, DiagnosticsIncludeSemantic))
	}
	if firstNamedReference(t, query, "B").ReferencedSymbol() != localVariable(t, binder.LocalScope, "B") {
		t.Fatal("query parameters were not published after all defaults were bound")
	}
}

func TestNativeFunctionBodyQueryParametersDoNotLeak(t *testing.T) {
	query := ParseQuery(`let F=() { declare query_parameters(Inside:string); print Value=Inside }; print x=1`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)

	if binder.LocalScope.ContainsSymbol("Inside") {
		t.Fatal("function-body query parameter leaked into the outer query scope")
	}
	inside := firstNamedReference(t, query, "Inside")
	variable, ok := inside.ReferencedSymbol().(*VariableSymbol)
	if !ok || variable.Type != ScalarTypes.String.Type() {
		t.Fatalf("function-body reference = %#v, want local string variable", inside.ReferencedSymbol())
	}
}

func TestNativeQueryParametersInvalidDefaultsAreDiagnosed(t *testing.T) {
	tests := []struct {
		query string
		code  string
	}{
		{query: `declare query_parameters(User:string=Missing); print x=1`, code: "KS132"},
		{query: `declare query_parameters(Limit:long="five"); print x=1`, code: "KS141"},
	}
	for _, test := range tests {
		t.Run(test.code+test.query, func(t *testing.T) {
			query := ParseQuery(test.query)
			binder := NewBinder(NewGlobalState(nil))
			binder.TryBind(query)
			parameter := firstBoundNode[*FunctionParameter](t, query)
			if !diagnosticsContain(parameter.SemanticDiagnostics(), test.code) {
				t.Fatalf("diagnostics = %#v, want %s", parameter.SemanticDiagnostics(), test.code)
			}
		})
	}
}

func TestNativeSetOptionBindsKnownNameAndValidatesValue(t *testing.T) {
	tests := []struct {
		query string
		code  string
	}{
		{query: `set query_take_max_records=1000; print x=1`},
		{query: `set query_take_max_records="many"; print x=1`, code: "KS141"},
		{query: `set query_take_max_records=1 + 2; print x=1`, code: "KS130"},
		{query: `set query_now=Missing; print x=1`, code: "KS142"},
		{query: `set ['query_take_max_records']=1000; print x=1`, code: "KS105"},
		{query: `set made_up_option="x"; print x=1`},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(NewGlobalState(nil)).TryBind(query)
			set := firstBoundNode[*SetOptionStatement](t, query)
			if test.code == "KS142" {
				assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), test.code)
			} else if test.code == "" && len(set.SemanticDiagnostics()) != 0 {
				t.Fatalf("diagnostics = %#v, want none", set.SemanticDiagnostics())
			} else if test.code != "" && !diagnosticsContain(set.SemanticDiagnostics(), test.code) {
				t.Fatalf("diagnostics = %#v, want %s", set.SemanticDiagnostics(), test.code)
			}
			if NameDeclarationSimpleName(set.Name()) == "query_take_max_records" && set.Name().ReferencedSymbol() != Options.TakeMaxRecords {
				t.Fatalf("option reference = %#v, want %#v", set.Name().ReferencedSymbol(), Options.TakeMaxRecords)
			}
		})
	}
}

func TestNativeSetOptionUsesCustomGlobalStateCatalog(t *testing.T) {
	custom := NewOptionSymbolWithType("custom_limit", "", ScalarTypes.Long, nil)
	query := ParseQuery(`set custom_limit="wrong"; print x=1`)
	NewBinder(NewGlobalState(nil).WithOptions([]*OptionSymbol{custom})).TryBind(query)

	set := firstBoundNode[*SetOptionStatement](t, query)
	if set.Name().ReferencedSymbol() != custom {
		t.Fatalf("option reference = %#v, want custom catalog symbol", set.Name().ReferencedSymbol())
	}
	assertSemanticDiagnostic(t, set, "KS141")
}

func TestNativeQueryStatementBindingIsIdempotent(t *testing.T) {
	query := ParseQuery(`declare query_parameters(User:string); set query_now=datetime(2024-01-01); print User`)
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(query)
	first := localVariable(t, binder.LocalScope, "User")
	binder.TryBind(query)
	if localVariable(t, binder.LocalScope, "User") != first {
		t.Fatal("rebinding replaced the published query parameter")
	}
}
