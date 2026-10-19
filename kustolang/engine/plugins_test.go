package kustolang

import (
	"testing"
)

// Covers the PlugIns.cs change at upstream commit 8aa8e708fe30351395a1b05d9af452a8c7657023.
func TestModelRequestPlugInsAcceptExactlyTwoStringArguments(t *testing.T) {
	for _, plugin := range []*FunctionSymbol{PlugIns.DaxRequest, PlugIns.GqlRequest} {
		t.Run(plugin.SymbolName(), func(t *testing.T) {
			name := plugin.SymbolName()
			if plugin.IsHidden() {
				t.Fatal("model request plug-in is hidden")
			}
			for _, test := range []struct {
				arguments string
				wantCode  string
			}{
				{`'connection', 'query'`, ""},
				{`'connection'`, "KS119"},
				{`'connection', 'query', dynamic({})`, "KS119"},
				{`'connection', 'query', dynamic({}), dynamic({})`, "KS119"},
				{`1, 'query'`, "KS107"},
				{`'connection', 1`, "KS107"},
			} {
				code := ParseAndAnalyze("print Input=1 | evaluate " + name + "(" + test.arguments + ")")
				diagnostics := code.GetDiagnostics()
				if test.wantCode == "" {
					if len(diagnostics) != 0 {
						t.Fatalf("arguments %s: unexpected diagnostics %#v", test.arguments, diagnostics)
					}
					table, ok := TableSymbolForType(code.ResultType)
					if !ok || !table.IsOpen() || len(table.Columns()) != 0 {
						t.Fatalf("result = %#v, want empty open runtime schema", code.ResultType)
					}
				} else if len(diagnostics) != 1 || diagnostics[0].Code != test.wantCode {
					t.Fatalf("arguments %s: diagnostics %#v, want %s", test.arguments, diagnostics, test.wantCode)
				}
			}
		})
	}
}

func TestModelRequestPlugInsAppearInEvaluateCompletion(t *testing.T) {
	text := "print Input=1 | evaluate "
	items := NewKustoCodeService(text).GetCompletionItems(len(text), CompletionOptionsDefault, CancellationToken{}).Items
	requireCompletionKind(t, items, "dax_request(connection_string, dax_query)", CompletionKindBuiltInFunction)
	requireCompletionKind(t, items, "gql_request(connection_string, gql_query)", CompletionKindBuiltInFunction)
}

func TestPlugInsAreAvailableThroughGlobalState(t *testing.T) {
	globals := NewGlobalState(NewDatabaseSymbol("db"))
	if len(globals.PlugIns) != len(PlugIns.KnownCurrent) {
		t.Fatalf("global plugin count = %d", len(globals.PlugIns))
	}
	if globals.GetPlugIn("bag_unpack") != PlugIns.BagUnpack {
		t.Fatal("bag_unpack was not found through GlobalState")
	}
	if globals.GetFunction("bag_unpack") != nil {
		t.Fatal("GlobalState.GetFunction should not search the plug-in catalog")
	}
	if !globals.IsBuiltInFunction(PlugIns.BagUnpack) {
		t.Fatal("plugin should be treated as a built-in function")
	}
	if globals.GetPlugIn("schema_merge") != nil {
		t.Fatal("schema_merge should not be part of upstream PlugIns.All lookup")
	}
	withoutPlugins := globals.WithPlugIns(nil)
	if withoutPlugins.GetPlugIn("bag_unpack") != nil {
		t.Fatal("WithPlugIns(nil) should replace default plugins")
	}
}

func TestPlugInDCountIntersectReturnInheritsRowScopeProperties(t *testing.T) {
	signature := PlugIns.DCountIntersect.Signatures()[0]
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	rowScope := NewTableSymbol("Rows", account).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	arguments := []Expression{ParseExpression("H1"), ParseExpression("H2")}
	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments: arguments,
		RowScope:  rowScope,
		Signature: signature,
	}).(*TableSymbol)
	columns := result.Columns()
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() || len(columns) != 3 ||
		columns[0] != account || columns[1].SymbolName() != "s0" || columns[1].Type != ScalarTypes.Long.Type() ||
		columns[2].SymbolName() != "s1" || columns[2].Type != ScalarTypes.Long.Type() {
		t.Fatalf("dcount_intersect result = open:%t sorted:%t serialized:%t columns:%#v", result.IsOpen(), result.IsSorted(), result.IsSerialized(), columns)
	}
}

func TestPlugInPivotReturnExcludesImplicitAggregateColumn(t *testing.T) {
	signature := PlugIns.Pivot.Signatures()[0]
	category := NewColumnSymbol("Category", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Long.Type())
	user := NewColumnSymbol("User", ScalarTypes.String.Type())
	rowScope := NewTableSymbol("Rows", category, score, user)

	pivotArg := ParseExpression("Category")
	aggregateArg := ParseExpression("sum(Score)")
	pivotArg.(*NameReference).SetSemanticInfo(NewSemanticInfoForReferencedSymbol(category))
	firstNamedReference(t, aggregateArg, "Score").SetSemanticInfo(NewSemanticInfoForReferencedSymbol(score))
	context := CustomReturnTypeContext{
		Arguments: []Expression{pivotArg, aggregateArg},
		ArgumentParameters: []*Parameter{
			signature.Parameters[0],
			signature.Parameters[1],
		},
		ArgumentSemantics: []SemanticInfo{
			NewSemanticInfoForSymbol(category, category.Type),
			NewSemanticInfoForType(ScalarTypes.Long.Type()),
		},
		RowScope:  rowScope,
		Signature: signature,
	}

	result := signature.CustomReturnType(context).(*TableSymbol)
	columns := result.Columns()
	if !result.IsOpen() || len(columns) != 1 || columns[0] != user {
		t.Fatalf("pivot implicit passthrough columns = %#v open=%t", columns, result.IsOpen())
	}

	wrappedContext := context
	wrappedContext.Arguments = []Expression{pivotArg, ParseExpression("tostring(sum(Score))")}
	wrappedColumns := signature.CustomReturnType(wrappedContext).(*TableSymbol).Columns()
	if len(wrappedColumns) != 2 || wrappedColumns[0] != score || wrappedColumns[1] != user {
		t.Fatalf("pivot wrapped aggregate passthrough columns = %#v", wrappedColumns)
	}
}

func TestPlugInAIEmbeddingsReturnAddsOptionalErrorColumn(t *testing.T) {
	signature := PlugIns.AIEmbeddings.Signatures()[0]
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	rowScope := NewTableSymbol("Rows", account)
	textArg := ParseExpression("Text")
	connectionArg := ParseExpression(`"conn"`)
	optionsArg := ParseExpression(`dynamic({})`)
	errorsArg := ParseExpression("true")

	result := signature.CustomReturnType(CustomReturnTypeContext{
		Arguments: []Expression{textArg, connectionArg, optionsArg, errorsArg},
		ArgumentParameters: []*Parameter{
			signature.Parameters[0],
			signature.Parameters[1],
			signature.Parameters[2],
			signature.Parameters[3],
		},
		ArgumentResultNames:    []string{"Text", "", "", ""},
		ArgumentConstantValues: []any{nil, "conn", nil, true},
		RowScope:               rowScope,
		Signature:              signature,
	})

	table, ok := result.(*TableSymbol)
	if !ok {
		t.Fatalf("ai_embeddings result = %T", result)
	}
	columns := table.Columns()
	if len(columns) != 3 ||
		columns[0] != account ||
		columns[1].SymbolName() != "Text_embeddings" ||
		columns[1].Type != ScalarTypes.Dynamic.Type() ||
		columns[2].SymbolName() != "Text_embeddings_error" ||
		columns[2].Type != ScalarTypes.String.Type() {
		t.Fatalf("ai_embeddings columns = %#v", columns)
	}
}

func TestPlugInCustomLayouts(t *testing.T) {
	newActivity := PlugIns.NewActivityMetrics.Signatures()[0]
	newActivityParams := newActivity.ArgumentParametersForSymbols([]Symbol{
		NewColumnSymbol("Id", ScalarTypes.String.Type()),
		NewColumnSymbol("TimeGenerated", ScalarTypes.DateTime.Type()),
		ScalarTypes.DateTime.Type(),
		ScalarTypes.DateTime.Type(),
		ScalarTypes.TimeSpan.Type(),
		ScalarTypes.String.Type(),
		NewColumnSymbol("Region", ScalarTypes.String.Type()),
		NewTableSymbol("Lookback", NewColumnSymbol("Id", ScalarTypes.String.Type())),
	})
	if len(newActivityParams) != 8 ||
		newActivityParams[5].Name != "Cohort" ||
		newActivityParams[6].Name != "Dimension" ||
		newActivityParams[7].Name != "lookback" {
		t.Fatalf("new_activity_metrics custom parameters = %#v", newActivityParams)
	}

	sequence := PlugIns.SequenceDetect.Signatures()[0]
	sequenceParams := sequence.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.DateTime.Type(),
		ScalarTypes.TimeSpan.Type(),
		ScalarTypes.TimeSpan.Type(),
		ScalarTypes.Bool.Type(),
		ScalarTypes.Bool.Type(),
		ScalarTypes.String.Type(),
	})
	if len(sequenceParams) != 6 ||
		sequenceParams[3].Name != "Expr" ||
		sequenceParams[4].Name != "Expr" ||
		sequenceParams[5].Name != "Dimension" {
		t.Fatalf("sequence_detect custom parameters = %#v", sequenceParams)
	}

	geo := PlugIns.GeoPolygonLookup.Signatures()[0]
	namedReturnKey := functionCallArgument(t, `geo_polygon_lookup(T, key, lon, lat, return_lookup_key=true)`, 4)
	geoParams := geo.ArgumentParametersForExpressions(
		[]Expression{nil, nil, nil, nil, namedReturnKey},
		[]Symbol{ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Real.Type(), ScalarTypes.Real.Type(), ScalarTypes.Bool.Type()},
	)
	if len(geoParams) != 5 || geoParams[4].Name != "return_lookup_key" {
		t.Fatalf("geo_polygon_lookup custom parameters = %#v", geoParams)
	}

	polygonPositional := geo.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Real.Type(), ScalarTypes.Real.Type(),
		ScalarTypes.Real.Type(), ScalarTypes.Bool.Type(), ScalarTypes.Real.Type(), ScalarTypes.Bool.Type(),
	})
	assertParameterNames(t, polygonPositional,
		"LookupTable", "LookupPolygonKey", "SourceLongitude", "SourceLatitude",
		"radius", "return_unmatched", "lookup_area_radius", "return_lookup_key")

	line := PlugIns.GeoLineLookup.Signatures()[0]
	linePositional := line.ArgumentParametersForTypes([]*TypeSymbol{
		ScalarTypes.Dynamic.Type(), ScalarTypes.Dynamic.Type(), ScalarTypes.Real.Type(), ScalarTypes.Real.Type(),
		ScalarTypes.Real.Type(), ScalarTypes.Bool.Type(), ScalarTypes.Real.Type(), ScalarTypes.Bool.Type(),
	})
	assertParameterNames(t, linePositional,
		"LookupTable", "LookupLineKey", "SourceLongitude", "SourceLatitude",
		"radius", "return_unmatched", "lookup_area_radius", "return_lookup_key")
}

func assertParameterNames(t *testing.T, parameters []*Parameter, names ...string) {
	t.Helper()
	if len(parameters) != len(names) {
		t.Fatalf("parameter count = %d, want %d: %#v", len(parameters), len(names), parameters)
	}
	for i, name := range names {
		if parameters[i] == nil || parameters[i].Name != name {
			t.Fatalf("parameter %d = %#v, want %q", i, parameters[i], name)
		}
	}
}

func TestPlugInGeoLookupReturnUsesLookupKeyFlag(t *testing.T) {
	signature := PlugIns.GeoPolygonLookup.Signatures()[0]
	source := NewTableSymbol("Source", NewColumnSymbol("Longitude", ScalarTypes.Real.Type()))
	lookupKey := NewColumnSymbol("LookupKey", ScalarTypes.DynamicBag.Type())
	shape := NewColumnSymbol("Shape", ScalarTypes.DynamicBag.Type())
	lookup := NewTableSymbol("Lookup", lookupKey, shape)
	args := []Expression{
		ParseExpression("Lookup"),
		ParseExpression("LookupKey"),
		ParseExpression("Longitude"),
		ParseExpression("Latitude"),
		ParseExpression("1.0"),
		ParseExpression("false"),
		ParseExpression("1.0"),
		ParseExpression("false"),
	}

	context := CustomReturnTypeContext{
		Arguments:             args,
		ArgumentParameters:    signature.ArgumentParametersForExpressions(args, nil),
		ArgumentResultSymbols: []Symbol{lookup, lookupKey, nil, nil, nil, nil, nil, nil},
		ArgumentSemantics: []SemanticInfo{
			NewSemanticInfoForSymbol(lookup, &lookup.TypeSymbol),
			NewSemanticInfoForSymbol(lookupKey, lookupKey.Type),
		},
		ArgumentConstantValues: []any{nil, nil, nil, nil, nil, false, nil, false},
		RowScope:               source,
		Signature:              signature,
	}
	result := signature.CustomReturnType(context)
	table, ok := result.(*TableSymbol)
	if !ok {
		t.Fatalf("geo lookup result = %T", result)
	}
	columns := table.Columns()
	if len(columns) != 2 || columns[0].SymbolName() != "Longitude" || columns[1].SymbolName() != "Shape" {
		t.Fatalf("geo lookup return_lookup_key=false columns = %#v", columns)
	}

	context.ArgumentConstantValues[7] = true
	result = signature.CustomReturnType(context)
	table = result.(*TableSymbol)
	columns = table.Columns()
	if len(columns) != 3 || columns[1].SymbolName() != "LookupKey" || columns[2].SymbolName() != "Shape" {
		t.Fatalf("geo lookup return_lookup_key=true columns = %#v", columns)
	}

	namedReturnKey := functionCallArgument(t, `geo_polygon_lookup(Lookup, LookupKey, Longitude, Latitude, return_lookup_key=true)`, 4)
	namedContext := CustomReturnTypeContext{
		Arguments:             []Expression{args[0], args[1], args[2], args[3], namedReturnKey},
		ArgumentParameters:    signature.ArgumentParametersForExpressions([]Expression{args[0], args[1], args[2], args[3], namedReturnKey}, nil),
		ArgumentResultSymbols: []Symbol{lookup, lookupKey, nil, nil, nil},
		ArgumentSemantics: []SemanticInfo{
			NewSemanticInfoForSymbol(lookup, &lookup.TypeSymbol),
			NewSemanticInfoForSymbol(lookupKey, lookupKey.Type),
		},
		ArgumentConstantValues: []any{nil, nil, nil, nil, false},
		RowScope:               source,
		Signature:              signature,
	}
	result = signature.CustomReturnType(namedContext)
	table = result.(*TableSymbol)
	columns = table.Columns()
	if len(columns) != 3 || columns[1].SymbolName() != "LookupKey" || columns[2].SymbolName() != "Shape" {
		t.Fatalf("geo lookup named return_lookup_key=true columns = %#v", columns)
	}
}
