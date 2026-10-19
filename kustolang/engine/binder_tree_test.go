package kustolang

import (
	"strings"
	"testing"
)

func TestNativeTreeBinderConsumeVoidParametersAndIdempotence(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery("T | consume decodeblocks=false")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	consume := firstBoundNode[*ConsumeOperator](t, query)
	info, ok := consume.GetSemanticInfo()
	if !ok || info.ReferencedSymbol() != Void || info.ResultType != &Void.TypeSymbol {
		t.Fatalf("consume semantic info = %#v, want Void", info)
	}
	assertDiagnosticCodes(t, consume)
	parameter := consume.Parameters().At(0)
	if parameter.Name().ReferencedSymbol() != QueryOperatorParameters.DecodeBlocks {
		t.Fatalf("decodeblocks symbol = %#v", parameter.Name().ReferencedSymbol())
	}
	if expression := parameter.Expression(); simpleBinderSyntaxNode(expression) == nil || !simpleBinderSyntaxNode(expression).IsBound() ||
		!sameType(ExpressionResultType(expression), ScalarTypes.Bool.Type()) {
		t.Fatalf("decodeblocks expression was not bound as bool: %#v", expression)
	}
	pipe := firstBoundNode[*PipeExpression](t, query)
	pipeInfo, ok := pipe.GetSemanticInfo()
	if !ok || pipeInfo.ReferencedSymbol() != Void || pipeInfo.ResultType != &Void.TypeSymbol {
		t.Fatalf("consume pipe semantic info = %#v, want Void", pipeInfo)
	}

	binder.TryBind(query)
	rebound, ok := consume.GetSemanticInfo()
	if !ok || rebound.ReferencedSymbol() != Void || rebound.ResultType != &Void.TypeSymbol || len(rebound.Diagnostics) != 0 {
		t.Fatalf("consume rebinding changed semantics: %#v", rebound)
	}
}

func TestNativeTreeBinderConsumeParameterDiagnosticsAndRowScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, test := range []struct {
		name string
		text string
		want []string
	}{
		{name: "invalid literal", text: "T | consume decodeblocks=1", want: []string{"KS223"}},
		{name: "nonliteral", text: "T | consume decodeblocks=A", want: []string{"KS223"}},
		{name: "duplicate", text: "T | consume decodeblocks=true decodeblocks=false", want: []string{"KS180"}},
		{name: "unsupported", text: "T | consume hint.materialized=true", want: []string{"KS179"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			consume := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
				strings.TrimPrefix(test.text, "T | ")).(*ConsumeOperator)
			tree := newTreeBinder(NewBinder(globals))
			tree.inPipe = true
			tree.binder.rowScope = table
			tree.visitConsume(consume)
			if !consume.IsBound() {
				t.Fatalf("consume was not bound: parameters=%d name=%q supported=%t", consume.Parameters().Count(),
					NameDeclarationSimpleName(consume.Parameters().At(0).Name()), supportsConsumeOperator(consume))
			}
			assertDiagnosticCodes(t, consume, test.want...)
			info, ok := consume.GetSemanticInfo()
			if !ok || info.ReferencedSymbol() != Void {
				t.Fatalf("invalid consume result = %#v, want Void", info)
			}
		})
	}

	consume := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"consume hint.shufflekey=A").(*ConsumeOperator)
	tree := newTreeBinder(NewBinder(globals))
	tree.inPipe = true
	tree.binder.rowScope = table
	tree.visitConsume(consume)
	if !consume.IsBound() {
		t.Fatalf("row-scoped consume was not bound: parameters=%d name=%q supported=%t", consume.Parameters().Count(),
			NameDeclarationSimpleName(consume.Parameters().At(0).Name()), supportsConsumeOperator(consume))
	}
	assertDiagnosticCodes(t, consume, "KS179")
	reference, ok := consume.Parameters().At(0).Expression().(*NameReference)
	if !ok || reference.ReferencedSymbol() != a {
		t.Fatalf("consume parameter row-scope symbol = %#v, want A", consume.Parameters().At(0).Expression())
	}
}

func TestNativeTreeBinderConsumeStandaloneRecoveryUnresolvedAndDownstream(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "consume decodeblocks=false").(*ConsumeOperator)
	NewBinder(globals).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")
	if info, ok := standalone.GetSemanticInfo(); !ok || info.ReferencedSymbol() != Void {
		t.Fatalf("standalone consume result = %#v, want Void", info)
	}

	for _, text := range []string{"consume decodeblocks=", "consume decodeblocks true"} {
		recovered := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, text).(*ConsumeOperator)
		NewBinder(globals).TryBind(recovered)
		if !recovered.IsBound() {
			t.Fatalf("recovered consume %q was not bound", text)
		}
		assertDiagnosticCodes(t, recovered, "KS174")
	}

	unresolved := ParseQuery("Missing | consume decodeblocks=false")
	NewBinder(globals).TryBind(unresolved)
	consume := boundNodes[*ConsumeOperator](unresolved)[0]
	if !consume.IsBound() {
		t.Fatal("consume with an unknown tabular input was not bound")
	}
	for _, pipe := range boundNodes[*PipeExpression](unresolved) {
		if !pipe.IsBound() {
			t.Fatal("consume with an unknown tabular input left its pipe unbound")
		}
	}
	assertSemanticDiagnostic(t, firstNamedReference(t, unresolved, "Missing"), "KS204")

	downstream := ParseQuery("T | consume | count")
	NewBinder(globals).TryBind(downstream)
	downstreamConsume := firstBoundNode[*ConsumeOperator](t, downstream)
	if info, ok := downstreamConsume.GetSemanticInfo(); !ok || info.ReferencedSymbol() != Void {
		t.Fatalf("downstream consume result = %#v, want Void", info)
	}
	assertDiagnosticCodes(t, firstBoundNode[*CountOperator](t, downstream))
}

func TestNativeTreeBinderConsumeRollbackAndStateRestoration(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())).WithIsOpen(true)
	consume := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		"consume hint.shufflekey=Future decodeblocks=true").(*ConsumeOperator)
	badLiteral := consume.Parameters().At(1).Expression().(*LiteralExpression)
	badLiteral.SyntaxNode.kind = SyntaxKindAddExpression

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Bool.Type()))
	binder.rowScope = table
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	existing := binder.getOpenTableInferredColumn("Existing", table)
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.scanAliases = map[string]*VariableSymbol{"prior": NewVariableSymbol("prior", ScalarTypes.Long, false, nil, nil)}
	tree.nodeBinder.inPipe = true
	tree.nodeBinder.schemaCannotBeEmpty = true
	tree.nodeBinder.defaultColumnNameSuffix = 37
	tree.nodeBinder.pipedAsRowScope = outer
	tree.nodeBinder.joinRightRowScope = outer
	tree.nodeBinder.lookupRightRowScope = outer
	tree.visitConsume(consume)

	if expression := simpleBinderSyntaxNode(consume.Parameters().At(0).Expression()); consume.IsBound() ||
		(expression != nil && expression.IsBound()) || badLiteral.IsBound() {
		t.Fatal("failed consume committed operator or parameter semantics")
	}
	if binder.rowScope != table || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.scanAliases["prior"] == nil || !tree.nodeBinder.inPipe ||
		!tree.nodeBinder.schemaCannotBeEmpty || tree.nodeBinder.defaultColumnNameSuffix != 37 ||
		tree.nodeBinder.pipedAsRowScope != outer || tree.nodeBinder.joinRightRowScope != outer ||
		tree.nodeBinder.lookupRightRowScope != outer {
		t.Fatal("consume rollback leaked binder or tree state")
	}
	columns := binder.openTableInferredColumns[table]
	order := binder.openTableInferredOrder[table]
	if len(binder.openTableInferredColumns) != 1 || len(columns) != 1 || columns["Existing"] != existing ||
		len(binder.openTableInferredOrder) != 1 || len(order) != 1 || order[0] != existing {
		t.Fatal("consume rollback leaked open-table inference")
	}

	valid := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "consume decodeblocks=true").(*ConsumeOperator)
	tree.visitConsume(valid)
	if !valid.IsBound() {
		t.Fatal("valid consume was not bound after rollback")
	}
	if binder.rowScope != table || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.scanAliases["prior"] == nil || !tree.nodeBinder.inPipe ||
		!tree.nodeBinder.schemaCannotBeEmpty || tree.nodeBinder.defaultColumnNameSuffix != 37 ||
		tree.nodeBinder.pipedAsRowScope != outer || tree.nodeBinder.joinRightRowScope != outer ||
		tree.nodeBinder.lookupRightRowScope != outer {
		t.Fatal("successful consume leaked binder or tree state")
	}
}

func TestNativeTreeBinderExecuteAndCacheShapeStateAndIdempotence(t *testing.T) {
	declared := NewColumnSymbol("Declared", ScalarTypes.String.Type())
	table := NewTableSymbol("T", declared).WithIsOpen(true).WithIsSerialized(true).WithIsSorted(true)
	query := ParseQuery("T | __executeAndCache | project Declared, Future")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	future := binder.getOpenTableInferredColumn("Future", table)
	binder.TryBind(query)

	operator := firstBoundNode[*ExecuteAndCacheOperator](t, query)
	result := semanticTable(t, operator)
	if result == table {
		t.Fatal("execute-and-cache reused the input table instead of producing a fresh table")
	}
	assertOrderedColumns(t, result, "Declared", "Future")
	if result.GetColumn("Declared") != declared || result.GetColumn("Future") != future {
		t.Fatal("execute-and-cache did not preserve input column identity")
	}
	if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatalf("execute-and-cache state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}
	assertDiagnosticCodes(t, operator)
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Declared", "Future")

	binder.TryBind(query)
	if rebound := semanticTable(t, operator); rebound != result {
		t.Fatal("rebinding replaced the execute-and-cache result table")
	}
}

func TestNativeTreeBinderExecuteAndCacheStandaloneMissingAndErrorInputs(t *testing.T) {
	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "__executeAndCache").(*ExecuteAndCacheOperator)
	NewBinder(nil).TryBind(standalone)
	assertOrderedColumns(t, semanticTable(t, standalone))
	assertDiagnosticCodes(t, standalone, "KS174")

	for _, test := range []struct {
		name   string
		mutate func(*QueryBlock)
	}{
		{name: "error"},
		{name: "missing", mutate: func(query *QueryBlock) {
			reference := boundNodes[*NameReference](query)[0]
			reference.Name().(*TokenName).SyntaxNode.children[0] = NewMissingToken(SyntaxKindIdentifierToken, DiagnosticMissingName())
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery("MissingTable | __executeAndCache")
			if test.mutate != nil {
				test.mutate(query)
			}
			NewBinder(nil).TryBind(query)
			operator := firstBoundNode[*ExecuteAndCacheOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, operator))
			for _, diagnostic := range operator.SemanticDiagnostics() {
				if diagnostic.Code == "KS174" {
					t.Fatalf("piped execute-and-cache reported false KS174: %#v", operator.SemanticDiagnostics())
				}
			}
		})
	}
}

func TestNativeTreeBinderExecuteAndCacheCommaTailPreserved(t *testing.T) {
	query := ParseQuery("T | __executeAndCache, Extra")
	operator := boundNodes[*ExecuteAndCacheOperator](query)[0]
	if operator.ChildCount() != 1 || operator.ExecuteAndCacheKeyword() == nil {
		t.Fatalf("execute-and-cache generated shape = children:%d keyword:%#v", operator.ChildCount(), operator.ExecuteAndCacheKeyword())
	}
	before := operator.ContainedDiagnostics()
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))))).TryBind(query)
	if !operator.IsBound() {
		t.Fatal("execute-and-cache with comma tail was not bound")
	}
	if got := operator.ContainedDiagnostics(); len(got) != len(before) {
		t.Fatalf("comma-tail diagnostics changed during binding: before=%#v after=%#v", before, got)
	}
}

func TestNativeTreeBinderExecuteAndCacheStructuralCorruptionIsTransactional(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type())).WithIsOpen(true)
	for _, mutate := range []func(*ExecuteAndCacheOperator){
		func(node *ExecuteAndCacheOperator) { node.SyntaxNode.children[0] = nil },
		func(node *ExecuteAndCacheOperator) {
			node.SyntaxNode.children[0] = SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'wrong'"))
		},
		func(node *ExecuteAndCacheOperator) {
			node.SyntaxNode.children = append(node.SyntaxNode.children,
				SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCommaToken, "", ",")))
		},
	} {
		query := ParseQuery("T | __executeAndCache | project Future")
		operator := boundNodes[*ExecuteAndCacheOperator](query)[0]
		mutate(operator)

		binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
		outer := NewTableSymbol("Outer", NewColumnSymbol("OuterColumn", ScalarTypes.Bool.Type()))
		binder.rowScope = outer
		binder.scopeKind = ScopeKindAggregate
		local := binder.LocalScope
		existingInferred := binder.getOpenTableInferredColumn("Existing", table)
		tree := newTreeBinder(binder)
		tree.inPipe = true
		tree.inWildcard = true
		tree.scanAliases = map[string]*VariableSymbol{"prior": NewVariableSymbol("prior", ScalarTypes.Long, false, nil, nil)}
		tree.nodeBinder.inPipe = true
		tree.nodeBinder.schemaCannotBeEmpty = true
		tree.nodeBinder.defaultColumnNameSuffix = 37
		tree.nodeBinder.pipedAsRowScope = outer
		tree.nodeBinder.joinRightRowScope = outer
		tree.nodeBinder.lookupRightRowScope = outer
		tree.bind(query)

		if operator.IsBound() {
			t.Fatal("structurally corrupt execute-and-cache was bound")
		}
		for _, pipe := range boundNodes[*PipeExpression](query) {
			if pipe.IsBound() {
				t.Fatal("structurally corrupt execute-and-cache committed a containing pipe")
			}
		}
		for _, node := range boundNodes[*SyntaxNode](query) {
			if node.IsBound() {
				t.Fatalf("structurally corrupt execute-and-cache committed child semantics: %v", node.Kind())
			}
		}
		if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
			!tree.inPipe || !tree.inWildcard || tree.scanAliases["prior"] == nil || !tree.nodeBinder.inPipe ||
			!tree.nodeBinder.schemaCannotBeEmpty || tree.nodeBinder.defaultColumnNameSuffix != 37 ||
			tree.nodeBinder.pipedAsRowScope != outer || tree.nodeBinder.joinRightRowScope != outer ||
			tree.nodeBinder.lookupRightRowScope != outer {
			t.Fatal("structural preflight changed binder or tree state")
		}
		columns := binder.openTableInferredColumns[table]
		order := binder.openTableInferredOrder[table]
		if len(binder.openTableInferredColumns) != 1 || len(columns) != 1 || columns["Existing"] != existingInferred ||
			len(binder.openTableInferredOrder) != 1 || len(order) != 1 || order[0] != existingInferred {
			t.Fatal("structural preflight changed open-table inference state")
		}
	}
}

func TestNativeTreeBinderJoinExactBracketedRHSCommonColumnsAndDownstream(t *testing.T) {
	leftKey := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	leftValue := NewColumnSymbol("Value", ScalarTypes.String.Type())
	rightKey := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	rightValue := NewColumnSymbol("Value", ScalarTypes.Bool.Type())
	left := NewTableSymbol("Left", leftKey, leftValue).WithIsSorted(true)
	right := NewTableSymbol("Right Table", rightKey, rightValue, NewColumnSymbol("RightOnly", ScalarTypes.Real.Type())).WithIsOpen(true).WithIsSorted(true)
	query := ParseQuery("Left | join kind=inner ['Right Table'] on ['Key'] | project Key, Key1, Value, Value1, RightOnly, Future")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	future := binder.getOpenTableInferredColumn("Future", right)
	binder.TryBind(query)

	join := firstBoundNode[*JoinOperator](t, query)
	result := semanticTable(t, join)
	assertOrderedColumns(t, result, "Key", "Value", "Key1", "Value1", "RightOnly", "Future")
	if !result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("join state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	if result.GetColumn("Key") != leftKey || result.GetColumn("Value") != leftValue {
		t.Fatal("left join columns did not preserve identity")
	}
	if originals := result.GetColumn("Key1").OriginalColumns; len(originals) != 1 || originals[0] != rightKey {
		t.Fatalf("unique right Key lineage = %#v", originals)
	}
	if result.GetColumn("Future") != future {
		t.Fatal("open right inferred column did not preserve identity")
	}
	condition := join.ConditionClause().(*JoinOnClause).Expressions().At(0).Element().(*NameReference)
	common, ok := condition.ReferencedSymbol().(*ColumnSymbol)
	if !ok || len(common.OriginalColumns) != 2 || common.OriginalColumns[0] != leftKey || common.OriginalColumns[1] != rightKey {
		t.Fatalf("common join column = %#v", condition.ReferencedSymbol())
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Key", "Key1", "Value", "Value1", "RightOnly", "Future")
	assertDiagnosticCodes(t, join)
}

func TestNativeTreeBinderJoinParenthesizedPipelineAndParameters(t *testing.T) {
	key := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	left := NewTableSymbol("Left", key, NewColumnSymbol("LeftOnly", ScalarTypes.String.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()), NewColumnSymbol("RightOnly", ScalarTypes.Bool.Type()))
	query := ParseQuery("Left | join hint.shufflekey=Key hint.strategy=broadcast kind=leftouter (Right | where Key > 0 | project Key, RightOnly) on (Key) | count")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	binder.TryBind(query)

	join := firstBoundNode[*JoinOperator](t, query)
	if join.Parameters().At(0).Expression().(*NameReference).ReferencedSymbol() != key {
		t.Fatal("hint.shufflekey did not bind in the incoming left scope")
	}
	parenthesized := join.Expression().(*ParenthesizedExpression)
	if !parenthesized.IsBound() || unionExpressionTable(parenthesized) == nil {
		t.Fatal("supported parenthesized join RHS pipeline was not bound")
	}
	assertOrderedColumns(t, semanticTable(t, join), "Key", "LeftOnly", "Key1", "RightOnly")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")
	assertDiagnosticCodes(t, join)
}

func TestNativeTreeBinderJoinKindSelectsOutputSides(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()), NewColumnSymbol("L", ScalarTypes.String.Type())).WithIsOpen(true)
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()), NewColumnSymbol("R", ScalarTypes.Bool.Type())).WithIsOpen(true)
	for _, test := range []struct {
		kind    string
		columns []string
	}{
		{"innerunique", []string{"Key", "L", "Key1", "R"}},
		{"leftanti", []string{"Key", "L"}},
		{"leftantisemi", []string{"Key", "L"}},
		{"anti", []string{"Key", "L"}},
		{"leftsemi", []string{"Key", "L"}},
		{"rightanti", []string{"Key", "R"}},
		{"rightantisemi", []string{"Key", "R"}},
		{"rightsemi", []string{"Key", "R"}},
	} {
		t.Run(test.kind, func(t *testing.T) {
			query := ParseQuery("Left | join kind=" + test.kind + " Right on Key")
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(query)
			join := firstBoundNode[*JoinOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, join), test.columns...)
			if !semanticTable(t, join).IsOpen() {
				t.Fatal("join did not retain the included open side")
			}
			assertDiagnosticCodes(t, join)
		})
	}
}

func TestNativeTreeBinderJoinDiagnosticsAndMalformedRecovery(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.String.Type()))
	query := ParseQuery("Left | join kind=sideways kind=inner Right on Key, Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(query)
	join := firstBoundNode[*JoinOperator](t, query)
	assertDiagnosticCodes(t, join, "KS005", "KS180", "KS227", "KS145")

	missingClause := ParseQuery("Left | join Right")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(missingClause)
	missingJoin := firstBoundNode[*JoinOperator](t, missingClause)
	assertDiagnosticCodes(t, missingJoin, "KS006")
	if !missingJoin.IsBound() {
		t.Fatal("join with parser-recovered missing condition was not bound")
	}

	missingCondition := ParseQuery("Left | join Right on")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(missingCondition)
	recovered := firstBoundNode[*JoinOperator](t, missingCondition)
	if !recovered.IsBound() || !diagnosticsContainMessage(recovered.ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing join condition recovery = bound:%t diagnostics:%#v", recovered.IsBound(), recovered.ContainedDiagnostics())
	}

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "join Right").(*JoinOperator)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", right))).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174", "KS006")
}

func TestNativeTreeBinderJoinRestoresBinderStateAfterSuccess(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	query := ParseQuery("Left | join Right on Key")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("OuterColumn", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 17
	tree.nodeBinder.joinRightRowScope = outer
	tree.bind(query)

	if !firstBoundNode[*JoinOperator](t, query).IsBound() {
		t.Fatal("supported join was not bound")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.defaultColumnNameSuffix != 17 || tree.nodeBinder.joinRightRowScope != outer {
		t.Fatal("successful join did not restore binder and tree state")
	}
}

func TestNativeTreeBinderLookupSchemaProvenanceOpenAndDownstream(t *testing.T) {
	leftKey := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	leftValue := NewColumnSymbol("Value", ScalarTypes.String.Type())
	rightKey := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	rightValue := NewColumnSymbol("Value", ScalarTypes.Bool.Type())
	left := NewTableSymbol("Left", leftKey, leftValue).WithIsSorted(true)
	right := NewTableSymbol("Right Table", rightKey, rightValue, NewColumnSymbol("RightOnly", ScalarTypes.Real.Type())).WithIsOpen(true).WithIsSorted(true)
	query := ParseQuery("Left | lookup ['Right Table'] on ['Key'] | project Key, Value, Value1, RightOnly, Future")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	future := binder.getOpenTableInferredColumn("Future", right)
	binder.TryBind(query)

	lookup := firstBoundNode[*LookupOperator](t, query)
	if !lookup.IsBound() {
		t.Fatal("supported lookup was not bound")
	}
	result := semanticTable(t, lookup)
	assertOrderedColumns(t, result, "Key", "Value", "Value1", "RightOnly", "Future")
	if !result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("lookup state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	key := result.GetColumn("Key")
	if key == leftKey || len(key.OriginalColumns) != 2 || key.OriginalColumns[0] != leftKey || key.OriginalColumns[1] != rightKey {
		t.Fatalf("lookup key provenance = %#v", key)
	}
	if result.GetColumn("Value") != leftValue || result.GetColumn("Future") != future {
		t.Fatal("lookup did not preserve non-key column identity")
	}
	condition := lookup.LookupClause().(*JoinOnClause).Expressions().At(0).Element().(*NameReference)
	conditionKey, ok := condition.ReferencedSymbol().(*ColumnSymbol)
	if !ok || len(conditionKey.OriginalColumns) != 2 || conditionKey.OriginalColumns[0] != leftKey || conditionKey.OriginalColumns[1] != rightKey {
		t.Fatalf("lookup key reference provenance = %#v", condition.ReferencedSymbol())
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Key", "Value", "Value1", "RightOnly", "Future")
	assertDiagnosticCodes(t, lookup)
}

func TestNativeTreeBinderLookupParenthesizedRHSParametersAndDiagnostics(t *testing.T) {
	leftKey := NewColumnSymbol("Key", ScalarTypes.Long.Type())
	left := NewTableSymbol("Left", leftKey, NewColumnSymbol("LeftOnly", ScalarTypes.String.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.String.Type()), NewColumnSymbol("RightOnly", ScalarTypes.Bool.Type()))
	query := ParseQuery("Left | lookup kind=inner (Right | project Key, RightOnly) on (Key), Missing | count")
	lookup := boundNodes[*LookupOperator](query)[0]
	parameterReference := boundNodes[*NameReference](ParseQuery("Key"))[0]
	lookup.Parameters().At(0).SyntaxNode.children[2] = parameterReference
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(query)

	lookup = firstBoundNode[*LookupOperator](t, query)
	if parameterReference.ReferencedSymbol() != leftKey {
		t.Fatal("lookup parameter expression did not bind in the incoming left scope")
	}
	if !lookup.Expression().(*ParenthesizedExpression).IsBound() {
		t.Fatal("supported parenthesized lookup RHS was not bound")
	}
	assertDiagnosticCodes(t, lookup, "KS005", "KS227", "KS145")
	assertOrderedColumns(t, semanticTable(t, lookup), "Key", "LeftOnly", "Key1", "RightOnly")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "lookup Right on Key").(*LookupOperator)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", right))).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174", "KS145")
}

func TestNativeTreeBinderLookupParserRecovery(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()), NewColumnSymbol("R", ScalarTypes.Bool.Type()))
	for _, test := range []struct {
		text    string
		message string
	}{
		{"Left | lookup Right", "Missing join on condition clause"},
		{"Left | lookup Right on", "Missing expression"},
	} {
		t.Run(test.text, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(query)
			lookup := firstBoundNode[*LookupOperator](t, query)
			if !lookup.IsBound() || !diagnosticsContainMessage(lookup.ContainedDiagnostics(), "KS006", test.message) {
				t.Fatalf("lookup recovery = bound:%t diagnostics:%#v", lookup.IsBound(), lookup.ContainedDiagnostics())
			}
		})
	}
}

func TestNativeTreeBinderLookupUnsupportedFormsAreTransactional(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type())).WithIsOpen(true)
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type())).WithIsOpen(true)
	for _, text := range []string{
		"Left | lookup Right* on Key | count",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
			outer := NewTableSymbol("Outer", NewColumnSymbol("OuterColumn", ScalarTypes.Bool.Type()))
			binder.rowScope = outer
			binder.scopeKind = ScopeKindAggregate
			local := binder.LocalScope
			tree := newTreeBinder(binder)
			tree.inPipe = true
			tree.inWildcard = true
			tree.nodeBinder.lookupRightRowScope = outer
			tree.bind(query)

			lookup := boundNodes[*LookupOperator](query)[0]
			if lookup.IsBound() || binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate ||
				binder.LocalScope != local || !tree.inPipe || !tree.inWildcard || tree.nodeBinder.lookupRightRowScope != outer {
				t.Fatal("unsupported lookup committed semantics or changed binder state")
			}
			if binder.openTableInferredColumns != nil || binder.openTableInferredOrder != nil {
				t.Fatal("unsupported lookup leaked open-table inference")
			}
			for _, node := range boundNodes[*SyntaxNode](query) {
				if node.IsBound() {
					t.Fatalf("unsupported lookup committed node semantics: %v", node.Kind())
				}
			}
		})
	}
}

func TestNativeTreeBinderLookupRejectsInjectedMalformedStructuresAndRestoresSuccessState(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("Key", ScalarTypes.Long.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", left, right))

	for _, mutate := range []func(*LookupOperator){
		func(node *LookupOperator) { node.SyntaxNode.children[1] = nil },
		func(node *LookupOperator) { node.SyntaxNode.children[2] = nil },
		func(node *LookupOperator) { node.SyntaxNode.children[3] = nil },
		func(node *LookupOperator) {
			join := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "join Right where Key == Key").(*JoinOperator)
			node.SyntaxNode.children[3] = join.ConditionClause()
		},
		func(node *LookupOperator) {
			clause := node.LookupClause().(*JoinOnClause)
			clause.SyntaxNode.children[0] = NewMissingToken(SyntaxKindOnKeyword, DiagnosticIncompleteFragment())
		},
	} {
		query := ParseQuery("Left | lookup Right on Key | count")
		lookup := boundNodes[*LookupOperator](query)[0]
		mutate(lookup)
		newTreeBinder(NewBinder(globals)).bind(query)
		if lookup.IsBound() {
			t.Fatal("injected malformed lookup was bound")
		}
		for _, pipe := range boundNodes[*PipeExpression](query) {
			if pipe.IsBound() {
				t.Fatal("injected malformed lookup committed a containing pipe")
			}
		}
	}

	query := ParseQuery("Left | lookup Right on Key")
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("OuterColumn", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 23
	tree.nodeBinder.lookupRightRowScope = outer
	tree.bind(query)
	if !firstBoundNode[*LookupOperator](t, query).IsBound() || binder.rowScope != outer ||
		binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local || !tree.inPipe || !tree.inWildcard ||
		tree.nodeBinder.defaultColumnNameSuffix != 23 || tree.nodeBinder.lookupRightRowScope != outer {
		t.Fatal("successful lookup did not restore binder and tree state")
	}
}

func TestNativeTreeBinderUnionRootUnifiesSchemasAndPrependsSource(t *testing.T) {
	leftA := NewColumnSymbol("A", ScalarTypes.String.Type())
	leftN := NewColumnSymbol("N", ScalarTypes.Long.Type())
	rightA := NewColumnSymbol("A", ScalarTypes.Long.Type())
	rightN := NewColumnSymbol("N", ScalarTypes.Long.Type())
	left := NewTableSymbol("Left", leftA, leftN).WithIsSorted(true)
	right := NewTableSymbol("Right Table", rightA, rightN).WithIsOpen(true).WithIsSorted(true)
	query := ParseQuery(`union withsource=Origin Left, ['Right Table']`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right)))
	binder.getOpenTableInferredColumn("Future", right)
	binder.TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	result := semanticTable(t, union)
	assertOrderedColumns(t, result, "Origin", "A_string", "A_long", "N", "Future")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"Origin":   ScalarTypes.String.Type(),
		"A_string": ScalarTypes.String.Type(),
		"A_long":   ScalarTypes.Long.Type(),
		"N":        ScalarTypes.Long.Type(),
		"Future":   ScalarTypes.Unknown.Type(),
	})
	if !result.IsOpen() || result.IsSorted() {
		t.Fatalf("union state = open:%t sorted:%t", result.IsOpen(), result.IsSorted())
	}
	if sources := result.GetColumn("N").OriginalColumns; len(sources) != 2 || sources[0] != leftN || sources[1] != rightN {
		t.Fatalf("unified N sources = %#v", sources)
	}
	if union.Expressions().At(0).Element().(*NameReference).ReferencedSymbol() != left ||
		union.Expressions().At(1).Element().(*NameReference).ReferencedSymbol() != right {
		t.Fatal("union operands did not resolve to catalog tables")
	}
	assertDiagnosticCodes(t, union)
}

func TestNativeTreeBinderUnionPipedInputNamespaceAndDownstream(t *testing.T) {
	input := NewTableSymbol("Input", NewColumnSymbol("Remote", ScalarTypes.String.Type()), NewColumnSymbol("InputOnly", ScalarTypes.Bool.Type())).WithIsSorted(true)
	remote := NewTableSymbol("Remote", NewColumnSymbol("RemoteOnly", ScalarTypes.Long.Type())).WithIsSorted(true)
	query := ParseQuery("Input | union Remote | project InputOnly, RemoteOnly")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", input, remote)))
	binder.LocalScope.AddSymbol(NewVariableSymbol("Remote", ScalarTypes.String.Type(), false, nil, nil))
	binder.TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	operand := union.Expressions().At(0).Element().(*NameReference)
	if operand.ReferencedSymbol() != remote {
		t.Fatalf("union operand resolved to %#v, want table namespace Remote", operand.ReferencedSymbol())
	}
	unionResult := semanticTable(t, union)
	assertOrderedColumns(t, unionResult, "Remote", "InputOnly", "RemoteOnly")
	if unionResult.IsSorted() {
		t.Fatal("piped union retained sorted state")
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, project), "InputOnly", "RemoteOnly")
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if !pipe.IsBound() {
			t.Fatal("supported union pipeline was not bound downstream")
		}
	}
}

func TestNativeTreeBinderUnionParenthesizedPipelines(t *testing.T) {
	left := NewTableSymbol("Left", NewColumnSymbol("A", ScalarTypes.Long.Type()), NewColumnSymbol("Drop", ScalarTypes.String.Type()))
	right := NewTableSymbol("Right", NewColumnSymbol("A", ScalarTypes.Long.Type()), NewColumnSymbol("B", ScalarTypes.Bool.Type()))
	query := ParseQuery("union (Left | where A > 0 | project A), (Right | take 2 | extend C = B)")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", left, right))).TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, union), "A", "B", "C")
	for _, separated := range union.Expressions().Elements() {
		operand := separated.Element().(*ParenthesizedExpression)
		if !operand.IsBound() || unionExpressionTable(operand) == nil {
			t.Fatalf("parenthesized operand was not bound: %#v", operand)
		}
	}
	assertDiagnosticCodes(t, union)
}

func TestNativeTreeBinderUnionParametersAndParserOwnedRecovery(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("union kind=sideways kind=outer withsource=Src withsource=Again isfuzzy=false , T,")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	union := firstBoundNode[*UnionOperator](t, query)
	assertDiagnosticCodes(t, union, "KS005", "KS180", "KS180")
	assertOrderedColumns(t, semanticTable(t, union), "Src", "A")
	if diagnostics := union.ContainedDiagnostics(); !diagnosticsContainMessage(diagnostics, "KS006", "Missing expression") {
		t.Fatalf("parser-owned missing operand diagnostics = %#v", diagnostics)
	}
	if !union.IsBound() {
		t.Fatal("recoverable missing union operands prevented binding")
	}

	missingComma := ParseQuery("union T T")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(missingComma)
	missingCommaUnion := firstBoundNode[*UnionOperator](t, missingComma)
	if !missingCommaUnion.IsBound() || !diagnosticsContainMessage(missingCommaUnion.ContainedDiagnostics(), "KS005", "Expected: ,") {
		t.Fatalf("missing comma recovery = bound:%t diagnostics:%#v", missingCommaUnion.IsBound(), missingCommaUnion.ContainedDiagnostics())
	}
}

func TestNativeTreeBinderUnionBracketedPipeRecoveryRestoresState(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	for _, text := range []string{
		"union ['T' | count",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
			outer := NewTableSymbol("Outer", NewColumnSymbol("OuterColumn", ScalarTypes.Bool.Type()))
			binder.rowScope = outer
			binder.scopeKind = ScopeKindAggregate
			local := binder.LocalScope
			tree := newTreeBinder(binder)
			tree.inPipe = true
			tree.inWildcard = true
			tree.scanAliases = map[string]*VariableSymbol{"old": NewVariableSymbol("old", table, false, nil, nil)}
			oldAliases := tree.scanAliases
			tree.bind(query)

			union := boundNodes[*UnionOperator](query)[0]
			if !union.IsBound() {
				t.Fatal("recoverable union was not bound")
			}
			if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
				!tree.inPipe || !tree.inWildcard || tree.scanAliases["old"] != oldAliases["old"] {
				t.Fatal("recovered union changed binder or tree state")
			}
			for _, pipe := range boundNodes[*PipeExpression](query) {
				if !pipe.IsBound() {
					t.Fatal("pipe within recovered union was not bound")
				}
			}
			for _, count := range boundNodes[*CountOperator](query) {
				if !count.IsBound() {
					t.Fatal("operator within recovered union was not bound")
				}
			}
			for _, reference := range boundNodes[*NameReference](query) {
				if !reference.IsBound() {
					t.Fatal("recovered union operand was not bound")
				}
			}
		})
	}
}

func TestNativeTreeBinderBindsTakeAndCountPipeline(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	).WithIsSorted(true)
	query := ParseQuery("T | take 5 | count as Total")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	if !binder.TryBind(query) {
		t.Fatal("TryBind returned false")
	}
	source := firstNamedReference(t, query, "T")
	if sourceTable := semanticTable(t, source); sourceTable != table {
		t.Fatalf("source resolved to %#v, want catalog table", sourceTable)
	}

	take := firstBoundNode[*TakeOperator](t, query)
	if typ := ExpressionResultType(take.Expression()); !sameType(typ, ScalarTypes.Long.Type()) {
		t.Fatalf("take literal type = %#v, want long", typ)
	}
	takeTable := semanticTable(t, take)
	assertColumnTypes(t, takeTable, map[string]*TypeSymbol{
		"A": ScalarTypes.String.Type(),
		"N": ScalarTypes.Long.Type(),
	})
	if takeTable.IsSorted() {
		t.Fatal("take result retained sorted state")
	}

	count := firstBoundNode[*CountOperator](t, query)
	countTable := semanticTable(t, count)
	assertColumnTypes(t, countTable, map[string]*TypeSymbol{"Total": ScalarTypes.Long.Type()})

	pipes := boundNodes[*PipeExpression](query)
	if len(pipes) != 2 {
		t.Fatalf("bound pipes = %d, want 2", len(pipes))
	}
	for _, pipe := range pipes {
		if _, ok := pipe.GetSemanticInfo(); !ok || semanticTable(t, pipe) == nil {
			t.Fatalf("pipe semantic info = %#v", pipe)
		}
	}
	if binder.rowScope != nil || binder.scopeKind != ScopeKindNormal {
		t.Fatalf("binder state leaked after pipe: row=%#v scope=%v", binder.rowScope, binder.scopeKind)
	}
}

func TestNativeTreeBinderSamplePropagatesRowsAndClearsSorted(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type())).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | sample 2")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	sample := firstBoundNode[*SampleOperator](t, query)
	result := semanticTable(t, sample)
	assertColumnTypes(t, result, map[string]*TypeSymbol{"N": ScalarTypes.Long.Type()})
	if result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("sample table state = sorted:%t serialized:%t", result.IsSorted(), result.IsSerialized())
	}
}

func TestNativeTreeBinderRangeTypesDeclarationAndClosedResult(t *testing.T) {
	query := ParseQuery("range Value from 1 to 10 step 1")
	NewBinder(NewGlobalState(nil)).TryBind(query)
	rangeOperator := firstBoundNode[*RangeOperator](t, query)
	assertDiagnosticCodes(t, rangeOperator)
	result := semanticTable(t, rangeOperator)
	assertOrderedColumns(t, result, "Value")
	assertColumnTypes(t, result, map[string]*TypeSymbol{"Value": ScalarTypes.Long.Type()})
	if result.SymbolName() != "" || result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("range result state = name:%q open:%t sorted:%t serialized:%t", result.SymbolName(), result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	if rangeOperator.Name().ReferencedSymbol() != result.GetColumn("Value") {
		t.Fatalf("range declaration symbol = %#v, want output column %#v", rangeOperator.Name().ReferencedSymbol(), result.GetColumn("Value"))
	}

	for _, test := range []struct {
		text string
		typ  *TypeSymbol
	}{
		{"range T from datetime(2024-01-01) to datetime(2024-01-02) step 1h", ScalarTypes.DateTime.Type()},
		{"range Span from 1h to 2h step 1m", ScalarTypes.TimeSpan.Type()},
		{"range R from 1.0 to 2.0 step 0.5", ScalarTypes.Real.Type()},
	} {
		t.Run(test.text, func(t *testing.T) {
			parsed := ParseQuery(test.text)
			NewBinder(NewGlobalState(nil)).TryBind(parsed)
			operator := firstBoundNode[*RangeOperator](t, parsed)
			if got := semanticTable(t, operator).Columns()[0].Type; !sameType(got, test.typ) {
				t.Fatalf("range type = %#v, want %#v", got, test.typ)
			}
		})
	}
}

func TestNativeTreeBinderRangeIntSpecialAndAddSemantics(t *testing.T) {
	for _, test := range []struct {
		name     string
		fromType *TypeSymbol
		toType   *TypeSymbol
		stepType *TypeSymbol
		want     *TypeSymbol
	}{
		{"int", ScalarTypes.Int.Type(), ScalarTypes.Int.Type(), ScalarTypes.Int.Type(), ScalarTypes.Int.Type()},
		{"promoted", ScalarTypes.Int.Type(), ScalarTypes.Long.Type(), ScalarTypes.Real.Type(), ScalarTypes.Real.Type()},
		{"datetime-add", ScalarTypes.DateTime.Type(), ScalarTypes.DateTime.Type(), ScalarTypes.TimeSpan.Type(), ScalarTypes.DateTime.Type()},
		{"timespan-add", ScalarTypes.TimeSpan.Type(), ScalarTypes.TimeSpan.Type(), ScalarTypes.TimeSpan.Type(), ScalarTypes.TimeSpan.Type()},
	} {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery("range X from FromValue to ToValue step StepValue")
			binder := NewBinder(NewGlobalState(nil))
			binder.LocalScope.AddSymbol(NewVariableSymbol("FromValue", test.fromType, false, nil, nil))
			binder.LocalScope.AddSymbol(NewVariableSymbol("ToValue", test.toType, false, nil, nil))
			binder.LocalScope.AddSymbol(NewVariableSymbol("StepValue", test.stepType, false, nil, nil))
			binder.TryBind(query)
			operator := firstBoundNode[*RangeOperator](t, query)
			if got := semanticTable(t, operator).GetColumn("X").Type; !sameType(got, test.want) {
				t.Fatalf("range type = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNativeTreeBinderRangeDiagnosticsOrderAndLocations(t *testing.T) {
	text := "T | range X from 'a' to true step (0)"
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type())).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery(text)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	operator := firstBoundNode[*RangeOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS175", "KS138", "KS138", "KS140")
	assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
		diagnosticLocation{"KS175", strings.Index(text, "range"), len("range")},
		diagnosticLocation{"KS138", strings.Index(text, "'a'"), len("'a'")},
		diagnosticLocation{"KS138", strings.Index(text, "true"), len("true")},
		diagnosticLocation{"KS140", strings.Index(text, "(0)"), len("(0)")},
	)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "X")
	if !sameType(result.GetColumn("X").Type, ScalarTypes.Unknown.Type()) || result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("invalid piped range result = %#v", result)
	}
}

func TestNativeTreeBinderRangeMissingRecoveryAndFailedLeftPipe(t *testing.T) {
	query := ParseQuery("MissingInput | range X from MissingFrom to 10 step 1")
	NewBinder(NewGlobalState(nil)).TryBind(query)
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "MissingInput"), "KS204")
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "MissingFrom"), "KS142")
	operator := firstBoundNode[*RangeOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS175")
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "X")
	if !sameType(result.GetColumn("X").Type, ScalarTypes.Unknown.Type()) {
		t.Fatalf("recovered range type = %#v, want unknown", result.GetColumn("X").Type)
	}
	if operator.Name().ReferencedSymbol() != result.GetColumn("X") {
		t.Fatal("recovered range declaration was not attached")
	}
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if !pipe.IsBound() {
			t.Fatal("failed-left range pipe was not bound")
		}
	}

	malformedRange := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "range n from to 5 step").(*RangeOperator)
	NewBinder(NewGlobalState(nil)).TryBind(malformedRange)
	if !malformedRange.IsBound() || len(malformedRange.SemanticDiagnostics()) != 0 {
		t.Fatalf("malformed range recovery = bound:%t diagnostics:%#v", malformedRange.IsBound(), malformedRange.SemanticDiagnostics())
	}
}

func TestNativeTreeBinderRangeSupportedBinaryDiagnostics(t *testing.T) {
	query := ParseQuery("range X from (1 < 2) to (2 != 3) step (true and false)")
	NewBinder(NewGlobalState(nil)).TryBind(query)
	operator := firstBoundNode[*RangeOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS138", "KS138", "KS138", "KS106")
	assertOrderedColumns(t, semanticTable(t, operator), "X")
	if !sameType(semanticTable(t, operator).GetColumn("X").Type, ScalarTypes.Unknown.Type()) {
		t.Fatal("non-summable binary range did not produce unknown output")
	}
}

func TestNativeTreeBinderSampleDistinctPreservesColumnIdentityStateAndFeedsDownstream(t *testing.T) {
	selected := NewColumnSymbol("User Name", ScalarTypes.String.Type())
	ignored := NewColumnSymbol("Ignored", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", selected, ignored).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | sample-distinct 5 of ['User Name'] | project ['User Name']")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*SampleDistinctOperator](t, query)
	if operator.Parameters() == nil || operator.Parameters().Count() != 0 {
		t.Fatalf("sample-distinct parameters = %#v, want empty list", operator.Parameters())
	}
	result := semanticTable(t, operator)
	if result == table || result.SymbolName() != "" {
		t.Fatalf("sample-distinct result = %#v, want fresh unnamed table", result)
	}
	assertOrderedColumns(t, result, "User Name")
	if result.GetColumn("User Name") != selected {
		t.Fatal("sample-distinct did not preserve the direct column symbol")
	}
	if !result.IsOpen() || !result.IsSerialized() || result.IsSorted() {
		t.Fatalf("sample-distinct state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}

	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	if projected.GetColumn("User Name") != selected {
		t.Fatal("downstream project did not resolve the sample-distinct output column")
	}
}

func TestNativeTreeBinderSampleDistinctNonColumnUsesColumn1(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type())).WithIsOpen(true).WithIsSerialized(true)
	for _, ofExpression := range []string{"1", "(1)"} {
		t.Run(ofExpression, func(t *testing.T) {
			query := ParseQuery("T | sample-distinct (5) of " + ofExpression + " | project Column1")
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

			operator := firstBoundNode[*SampleDistinctOperator](t, query)
			assertDiagnosticCodes(t, operator, "KS109")
			result := semanticTable(t, operator)
			assertColumnTypes(t, result, map[string]*TypeSymbol{"Column1": ScalarTypes.Long.Type()})
			if !result.IsOpen() || !result.IsSerialized() || result.IsSorted() {
				t.Fatalf("sample-distinct state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
			}
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Column1")
		})
	}
}

func TestNativeTreeBinderSampleDistinctUnknownInputAndUnresolvedChildren(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	source := ParseQuery("Missing | sample-distinct 5 of N")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(source)
	assertSemanticDiagnostic(t, firstNamedReference(t, source, "Missing"), "KS204")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*SampleDistinctOperator](t, source)), "N")

	count := ParseQuery("T | sample-distinct Missing of N")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(count)
	assertSemanticDiagnostic(t, firstNamedReference(t, count, "Missing"), "KS142")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*SampleDistinctOperator](t, count)), "N")

	of := ParseQuery("T | sample-distinct 5 of Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(of)
	assertSemanticDiagnostic(t, firstNamedReference(t, of, "Missing"), "KS142")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*SampleDistinctOperator](t, of)), "Missing")
}

func TestNativeTreeBinderSampleDistinctScalarExpressionsAndRecovery(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		name, query, column string
		typ                 *TypeSymbol
		diagnostics         []diagnosticLocation
	}{
		{name: "count function", query: "T | sample-distinct tolong(5) of A", column: "A", typ: ScalarTypes.String.Type()},
		{name: "of function", query: "T | sample-distinct 5 of strlen(A)", column: "strlen_A", typ: ScalarTypes.Long.Type(), diagnostics: []diagnosticLocation{{"KS109", 25, 9}}},
		{name: "of arithmetic", query: "T | sample-distinct 5 of N + 1", column: "Column1", typ: ScalarTypes.Long.Type(), diagnostics: []diagnosticLocation{{"KS109", 25, 5}}},
		{name: "of alias", query: "T | sample-distinct 5 of X=A", column: "X", typ: ScalarTypes.String.Type(), diagnostics: []diagnosticLocation{{"KS109", 25, 3}}},
		{name: "wrong count type", query: "T | sample-distinct tolower(A) of A", column: "A", typ: ScalarTypes.String.Type(), diagnostics: []diagnosticLocation{{"KS134", 20, 10}}},
		{name: "parenthesized column", query: "T | sample-distinct 5 of (A)", column: "A", typ: ScalarTypes.String.Type(), diagnostics: []diagnosticLocation{{"KS109", 25, 3}}},
		{name: "no-op conversion", query: "T | sample-distinct 5 of tostring(A)", column: "A", typ: ScalarTypes.String.Type(), diagnostics: []diagnosticLocation{{"KS109", 25, 11}}},
		{name: "malformed call", query: "T | sample-distinct 5 of strlen(A", column: "strlen_A", typ: ScalarTypes.Long.Type(), diagnostics: []diagnosticLocation{{"KS109", 25, 8}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			operator := firstBoundNode[*SampleDistinctOperator](t, query)
			assertDiagnosticLocations(t, GetContainedDiagnostics(query, DiagnosticsIncludeSemantic), test.diagnostics...)
			assertColumnTypes(t, semanticTable(t, operator), map[string]*TypeSymbol{test.column: test.typ})
			assertOrderedColumns(t, semanticTable(t, operator), test.column)
			if test.name == "malformed call" {
				call := firstBoundNode[*FunctionCallExpression](t, query)
				if call.ArgumentList().CloseParen().Width() != 0 {
					t.Fatalf("malformed function close width = %d", call.ArgumentList().CloseParen().Width())
				}
			}
		})
	}
}

func TestNativeTreeBinderSampleDistinctRejectsNonemptyParameters(t *testing.T) {
	valid := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "sample-distinct 5 of A").(*SampleDistinctOperator)
	distinct := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "distinct hint.shufflekey=A A").(*DistinctOperator)
	operator := MustNewSampleDistinctOperatorTyped(
		valid.SampleDistinctKeyword().Clone(true).(*SyntaxToken),
		AsSyntaxListOf[*NamedParameter](distinct.Parameters().SyntaxList.Clone(true).(*SyntaxList)),
		cloneSyntaxElementPreservingGeneratedWrapper(valid.Expression(), true).(Expression),
		valid.OfKeyword().Clone(true).(*SyntaxToken),
		cloneSyntaxElementPreservingGeneratedWrapper(valid.OfExpression(), true).(Expression),
	)
	NewBinder(NewGlobalState(nil)).TryBind(operator)
	if !operator.IsBound() {
		t.Fatal("sample-distinct with nonempty parameters was not bound")
	}
	assertSemanticDiagnostic(t, operator, "KS179")
}

func TestNativeTreeBinderSampleDistinctRestoresBinderAndLocalScope(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", column)
	query := ParseQuery("T | sample-distinct 5 of A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	oldRowScope := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	binder.rowScope = oldRowScope
	binder.scopeKind = ScopeKindPlugIn
	marker := NewVariableSymbol("Existing", ScalarTypes.Bool.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(marker)
	binder.TryBind(query)

	if binder.rowScope != oldRowScope || binder.scopeKind != ScopeKindPlugIn {
		t.Fatalf("binder state leaked after sample-distinct: row=%#v scope=%v", binder.rowScope, binder.scopeKind)
	}
	if localVariable(t, binder.LocalScope, "Existing") != marker {
		t.Fatal("sample-distinct changed the existing local scope")
	}
}

func TestNativeTreeBinderTopHittersCountIdentityStateAndDownstreamScope(t *testing.T) {
	user := NewColumnSymbol("User", ScalarTypes.String.Type())
	score := NewColumnSymbol("Score", ScalarTypes.Real.Type())
	table := NewTableSymbol("T", user, score).WithIsOpen(true).WithIsSerialized(true)
	query := ParseQuery("T | top-hitters 5 of User | project approximate_count_User, User")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*TopHittersOperator](t, query)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "User", "approximate_count_User")
	if result.GetColumn("User") != user {
		t.Fatal("top-hitters replaced the direct of-column identity")
	}
	if !sameType(result.GetColumn("approximate_count_User").Type, ScalarTypes.Long.Type()) {
		t.Fatalf("approximate count type = %#v, want long", result.GetColumn("approximate_count_User").Type)
	}
	if result.SymbolName() != "" || !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatalf("top-hitters state = name:%q open:%t serialized:%t sorted:%t", result.SymbolName(), result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}
	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, projected, "approximate_count_User", "User")
	if projected.GetColumn("User") != user {
		t.Fatal("downstream project did not receive the top-hitters scope")
	}
}

func TestNativeTreeBinderTopHittersAliasesByTypeAndCollisions(t *testing.T) {
	key := NewColumnSymbol("Key", ScalarTypes.String.Type())
	weight := NewColumnSymbol("Weight", ScalarTypes.Decimal.Type())
	table := NewTableSymbol("T", key, weight)
	query := ParseQuery("T | top-hitters Limit=5 of approximate_sum_Total=Key by Total=Weight")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*TopHittersOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS109")
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "approximate_sum_Total", "approximate_sum_Total1")
	if !sameType(result.GetColumn("approximate_sum_Total1").Type, ScalarTypes.Decimal.Type()) {
		t.Fatalf("approximate sum type = %#v, want decimal", result.GetColumn("approximate_sum_Total1").Type)
	}
	ofNamed := operator.OfExpression().(*SimpleNamedExpression)
	if !ofNamed.Name().IsBound() || ofNamed.Name().ReferencedSymbol() != result.GetColumn("approximate_sum_Total") {
		t.Fatal("of alias declaration did not bind to its projected column")
	}
	if operator.ByClause().IsBound() {
		t.Fatal("TopHittersByClause itself was bound")
	}
}

func TestNativeTreeBinderTopHittersDiagnosticsOrderAndLocations(t *testing.T) {
	text := "top-hitters 'five' of 1 by 'weight'"
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, text).(*TopHittersOperator)
	NewBinder(NewGlobalState(nil)).TryBind(operator)

	assertDiagnosticCodes(t, operator, "KS174", "KS134", "KS109", "KS137")
	assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
		diagnosticLocation{"KS174", 0, len("top-hitters")},
		diagnosticLocation{"KS134", strings.Index(text, "'five'"), len("'five'")},
		diagnosticLocation{"KS109", strings.Index(text, "1"), 1},
		diagnosticLocation{"KS137", strings.Index(text, "'weight'"), len("'weight'")},
	)
}

func TestNativeTreeBinderTopHittersSupportedScalarForms(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	b := NewColumnSymbol("B", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, b)
	query := ParseQuery("T | top-hitters (A < B and B != 0) of (A) by Sum=(A >= 0 or B == 1)")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*TopHittersOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS134", "KS109", "KS137")
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A", "approximate_sum_Sum")
	if result.GetColumn("A") != a {
		t.Fatal("complete parenthesized of-expression did not preserve direct column identity")
	}
	if !sameType(result.GetColumn("approximate_sum_Sum").Type, ScalarTypes.Bool.Type()) {
		t.Fatalf("logical by result type = %#v, want bool", result.GetColumn("approximate_sum_Sum").Type)
	}
	if operator.ByClause().IsBound() {
		t.Fatal("TopHittersByClause itself was bound")
	}
}

func TestNativeTreeBinderTopHittersUnknownTypesSuppressCascades(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("T | top-hitters Limit of Value by Weight")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	for _, name := range []string{"Limit", "Value", "Weight"} {
		binder.LocalScope.AddSymbol(NewVariableSymbol(name, ScalarTypes.Unknown.Type(), false, nil, nil))
	}
	binder.TryBind(query)

	operator := firstBoundNode[*TopHittersOperator](t, query)
	assertDiagnosticCodes(t, operator)
	assertOrderedColumns(t, semanticTable(t, operator), "Value", "approximate_sum_Weight")
}

func TestNativeTreeBinderTopHittersDefaultNamesAdvanceAndSkipKnownNames(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("T | top-hitters 5 of 1 | top-hitters 5 of 2")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.LocalScope.AddSymbol(NewVariableSymbol("Column1", ScalarTypes.String.Type(), false, nil, nil))
	binder.TryBind(query)

	operators := boundNodes[*TopHittersOperator](query)
	if len(operators) != 2 {
		t.Fatalf("bound top-hitters operators = %d, want 2", len(operators))
	}
	assertOrderedColumns(t, semanticTable(t, operators[0]), "Column2", "approximate_count_")
	assertOrderedColumns(t, semanticTable(t, operators[1]), "Column3", "approximate_count_")
}

func TestNativeTreeBinderTopHittersMissingAndUnresolvedRecovery(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		text    string
		columns []string
		codes   []string
	}{
		{text: "T | top-hitters of A", columns: []string{"A", "approximate_count_A"}},
		{text: "T | top-hitters 5", columns: []string{"approximate_count_"}},
		{text: "T | top-hitters 5 of", columns: []string{"approximate_count_"}},
		{text: "T | top-hitters 5 of A by", columns: []string{"A", "approximate_sum_"}},
		{text: "T | top-hitters MissingCount of MissingOf by MissingBy", columns: []string{"MissingOf", "approximate_sum_MissingBy"}},
		{text: "T | top-hitters (Missing < 5) of A", columns: []string{"A", "approximate_count_A"}, codes: []string{"KS134"}},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(globals).TryBind(query)
			operator := firstBoundNode[*TopHittersOperator](t, query)
			assertDiagnosticCodes(t, operator, test.codes...)
			assertOrderedColumns(t, semanticTable(t, operator), test.columns...)
			if clause := operator.ByClause(); clause != nil && clause.IsBound() {
				t.Fatal("TopHittersByClause itself was bound during recovery")
			}
			firstBoundNode[*PipeExpression](t, query)
		})
	}

	unresolved := ParseQuery("T | top-hitters MissingCount of MissingOf by MissingBy")
	NewBinder(globals).TryBind(unresolved)
	for _, name := range []string{"MissingCount", "MissingOf", "MissingBy"} {
		assertSemanticDiagnostic(t, firstNamedReference(t, unresolved, name), "KS142")
	}
	result := semanticTable(t, firstBoundNode[*TopHittersOperator](t, unresolved))
	if !sameType(result.GetColumn("MissingOf").Type, ScalarTypes.Unknown.Type()) ||
		!sameType(result.GetColumn("approximate_sum_MissingBy").Type, ScalarTypes.Unknown.Type()) {
		t.Fatalf("unresolved recovery types = %#v", result.Columns())
	}
}

func TestNativeTreeBinderTopHittersCompleteScalarAndProjectionForms(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	tests := []struct {
		name, text string
		columns    map[string]*TypeSymbol
		codes      []string
	}{
		{name: "count function", text: "T | top-hitters strlen(A) of N", columns: map[string]*TypeSymbol{"N": ScalarTypes.Long.Type(), "approximate_count_N": ScalarTypes.Long.Type()}},
		{name: "count arithmetic", text: "T | top-hitters N + 1 of A", columns: map[string]*TypeSymbol{"A": ScalarTypes.String.Type(), "approximate_count_A": ScalarTypes.Long.Type()}},
		{name: "of arithmetic", text: "T | top-hitters 5 of N + 1", columns: map[string]*TypeSymbol{"Column1": ScalarTypes.Long.Type(), "approximate_count_": ScalarTypes.Long.Type()}, codes: []string{"KS109"}},
		{name: "of function", text: "T | top-hitters 5 of strlen(A)", columns: map[string]*TypeSymbol{"strlen_A": ScalarTypes.Long.Type(), "approximate_count_strlen_A": ScalarTypes.Long.Type()}, codes: []string{"KS109"}},
		{name: "of no-op conversion", text: "T | top-hitters 5 of tostring(A)", columns: map[string]*TypeSymbol{"A": ScalarTypes.String.Type(), "approximate_count_A": ScalarTypes.Long.Type()}, codes: []string{"KS109"}},
		{name: "by arithmetic", text: "T | top-hitters 5 of A by N + 1", columns: map[string]*TypeSymbol{"A": ScalarTypes.String.Type(), "approximate_sum_": ScalarTypes.Long.Type()}},
		{name: "by function", text: "T | top-hitters 5 of A by strlen(A)", columns: map[string]*TypeSymbol{"A": ScalarTypes.String.Type(), "approximate_sum_strlen_A": ScalarTypes.Long.Type()}},
		{name: "invalid by operator", text: "T | top-hitters 5 of A by N + A", columns: map[string]*TypeSymbol{"A": ScalarTypes.String.Type(), "approximate_sum_": ScalarTypes.Long.Type()}},
		{name: "malformed of call", text: "T | top-hitters 5 of strlen(A", columns: map[string]*TypeSymbol{"strlen_A": ScalarTypes.Long.Type(), "approximate_count_strlen_A": ScalarTypes.Long.Type()}, codes: []string{"KS109"}},
		{name: "malformed by call", text: "T | top-hitters 5 of A by tolong(N", columns: map[string]*TypeSymbol{"A": ScalarTypes.String.Type(), "approximate_sum_N": ScalarTypes.Long.Type()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			operator := firstBoundNode[*TopHittersOperator](t, query)
			assertDiagnosticCodes(t, operator, test.codes...)
			assertColumnTypes(t, semanticTable(t, operator), test.columns)
			if len(semanticTable(t, operator).Columns()) != len(test.columns) {
				t.Fatalf("columns = %#v, want %#v", semanticTable(t, operator).Columns(), test.columns)
			}
			if test.name == "invalid by operator" {
				assertSemanticDiagnostic(t, firstBoundNode[*BinaryExpression](t, query), "KS106")
			}
			if clause := operator.ByClause(); clause != nil && clause.IsBound() {
				t.Fatal("top-hitters bound its syntactic by clause")
			}
			if strings.Contains(test.name, "malformed") {
				call := firstBoundNode[*FunctionCallExpression](t, query)
				if call.ArgumentList().CloseParen().Width() != 0 {
					t.Fatalf("malformed function close width = %d", call.ArgumentList().CloseParen().Width())
				}
			}
		})
	}
}

func TestNativeTreeBinderTopHittersDynamicSelectorsAndTupleProjection(t *testing.T) {
	path := ParseQuery(`datatable(D:dynamic)[dynamic({"k":2})] | top-hitters 5 of D.k`)
	NewBinder(NewGlobalState(nil)).TryBind(path)
	pathOperator := firstBoundNode[*TopHittersOperator](t, path)
	assertDiagnosticCodes(t, pathOperator, "KS109")
	assertOrderedColumns(t, semanticTable(t, pathOperator), "D_k", "approximate_count_D_k")

	element := ParseQuery(`datatable(D:dynamic)[dynamic({"k":2})] | top-hitters 5 of D["k"]`)
	NewBinder(NewGlobalState(nil)).TryBind(element)
	elementOperator := firstBoundNode[*TopHittersOperator](t, element)
	assertDiagnosticCodes(t, elementOperator, "KS109")
	assertOrderedColumns(t, semanticTable(t, elementOperator), "D_k", "approximate_count_D_k")

	tuple := ParseQuery(`datatable(S:dynamic)[dynamic([1,2])] | top-hitters 5 of (Min,Max,Avg)=series_stats(S)`)
	NewBinder(NewGlobalState(nil)).TryBind(tuple)
	tupleOperator := firstBoundNode[*TopHittersOperator](t, tuple)
	assertDiagnosticCodes(t, tupleOperator, "KS109")
	assertOrderedColumns(t, semanticTable(t, tupleOperator),
		"Min", "Max", "Avg", "series_stats_S_max_idx", "series_stats_S_avg", "series_stats_S_stdev", "series_stats_S_variance", "approximate_count_")
	assertColumnTypes(t, semanticTable(t, tupleOperator), map[string]*TypeSymbol{
		"Min": ScalarTypes.Real.Type(), "Max": ScalarTypes.Long.Type(), "Avg": ScalarTypes.Real.Type(),
		"series_stats_S_max_idx": ScalarTypes.Long.Type(), "series_stats_S_avg": ScalarTypes.Real.Type(),
		"series_stats_S_stdev": ScalarTypes.Real.Type(), "series_stats_S_variance": ScalarTypes.Real.Type(),
		"approximate_count_": ScalarTypes.Long.Type(),
	})
}

func TestNativeTreeBinderTopHittersStandaloneIsSyntacticAndRestoresTreeState(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "top-hitters 5 of A").(*TopHittersOperator)
	binder := NewBinder(NewGlobalState(nil))
	row := NewTableSymbol("Ambient", a)
	binder.rowScope = row
	binder.scopeKind = ScopeKindPlugIn
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inWildcard = true
	tree.nodeBinder.inPipe = true
	tree.bind(operator)

	assertDiagnosticCodes(t, operator, "KS174")
	if binder.rowScope != row || binder.scopeKind != ScopeKindPlugIn || binder.LocalScope != local || tree.inPipe || !tree.inWildcard || !tree.nodeBinder.inPipe {
		t.Fatalf("top-hitters changed state: row:%#v scope:%v local:%p pipe:%t wildcard:%t node-pipe:%t", binder.rowScope, binder.scopeKind, binder.LocalScope, tree.inPipe, tree.inWildcard, tree.nodeBinder.inPipe)
	}
}

func TestNativeTreeBinderTopHittersRefusesUnresolvedInputAndRestoresState(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a)
	query := ParseQuery("Missing | top-hitters 5 of A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	oldRowScope := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	binder.rowScope = oldRowScope
	binder.scopeKind = ScopeKindPlugIn
	marker := NewVariableSymbol("Marker", ScalarTypes.Bool.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(marker)
	binder.TryBind(query)

	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS204")
	if !firstBoundNode[*TopHittersOperator](t, query).IsBound() || !firstNamedReference(t, query, "A").IsBound() {
		t.Fatal("top-hitters did not bind against the recovered open input")
	}
	if binder.rowScope != oldRowScope || binder.scopeKind != ScopeKindPlugIn || localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatalf("top-hitters leaked binder state: row:%#v scope:%v", binder.rowScope, binder.scopeKind)
	}
}

func TestNativeTreeBinderRejectsFirstTakeOperator(t *testing.T) {
	take := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "take 5").(*TakeOperator)
	NewBinder(NewGlobalState(nil)).TryBind(take)
	assertSemanticDiagnostic(t, take, "KS174")
}

func TestNativeTreeBinderChecksTakeIntegerExpression(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	query := ParseQuery("T | take 'five'")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	take := firstBoundNode[*TakeOperator](t, query)
	assertSemanticDiagnostic(t, take, "KS134")
}

func TestNativeTreeBinderTopPreservesRowsSortsAndFeedsDownstreamScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSerialized(true)
	query := ParseQuery("T | top 5 by N desc | project N, A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	top := firstBoundNode[*TopOperator](t, query)
	result := semanticTable(t, top)
	assertOrderedColumns(t, result, "A", "N")
	if result.GetColumn("A") != a || result.GetColumn("N") != n {
		t.Fatal("top did not preserve source column identity")
	}
	if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatalf("top state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}

	ordered, ok := top.ByExpression().(*OrderedExpression)
	if !ok {
		t.Fatalf("top by expression = %T, want *OrderedExpression", top.ByExpression())
	}
	if typ := ExpressionResultType(ordered); !sameType(typ, ScalarTypes.Long.Type()) {
		t.Fatalf("ordered expression type = %#v, want long", typ)
	}
	if ordered.ReferencedSymbol() != n {
		t.Fatalf("ordered expression symbol = %#v, want N column", ordered.ReferencedSymbol())
	}

	project := firstBoundNode[*ProjectOperator](t, query)
	projected := semanticTable(t, project)
	assertOrderedColumns(t, projected, "N", "A")
	if projected.GetColumn("N") != n || projected.GetColumn("A") != a {
		t.Fatal("downstream project did not resolve top output columns")
	}
}

func TestNativeTreeBinderTopDiagnostics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	other := NewTableSymbol("Other", NewColumnSymbol("X", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table, other))

	invalid := ParseQuery("T\n| top 'five' by Other")
	NewBinder(globals).TryBind(invalid)
	assertDiagnosticCodes(t, firstBoundNode[*TopOperator](t, invalid), "KS134", "KS108")
	assertDiagnosticLocations(t, GetContainedDiagnostics(invalid, DiagnosticsIncludeSemantic),
		diagnosticLocation{"KS134", 8, 6},
		diagnosticLocation{"KS108", 18, 5},
	)

	first := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "top 'five' by Other").(*TopOperator)
	NewBinder(globals).TryBind(first)
	assertDiagnosticCodes(t, first, "KS174", "KS134", "KS108")
	assertDiagnosticLocations(t, GetContainedDiagnostics(first, DiagnosticsIncludeSemantic),
		diagnosticLocation{"KS174", 0, 3},
		diagnosticLocation{"KS134", 4, 6},
		diagnosticLocation{"KS108", 14, 5},
	)
}

type diagnosticLocation struct {
	code   string
	start  int
	length int
}

func assertDiagnosticLocations(t *testing.T, diagnostics []Diagnostic, want ...diagnosticLocation) {
	t.Helper()
	if len(diagnostics) != len(want) {
		t.Fatalf("diagnostics = %#v, want locations %#v", diagnostics, want)
	}
	for i, location := range want {
		got := diagnostics[i]
		if got.Code != location.code || got.LocationKind != DiagnosticLocationAbsolute || got.Start != location.start || got.Length != location.length {
			t.Fatalf("diagnostic %d = %s %s %d:%d, want %s Absolute %d:%d", i, got.Code, got.LocationKind, got.Start, got.Length, location.code, location.start, location.length)
		}
	}
}

func TestNativeTreeBinderTopMissingColumnStillBinds(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	query := ParseQuery("T | top 5 by Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS142")
	top := firstBoundNode[*TopOperator](t, query)
	assertDiagnosticCodes(t, top)
	assertOrderedColumns(t, semanticTable(t, top), "N")
	firstBoundNode[*PipeExpression](t, query)
}

func TestNativeTreeBinderTopPartialChildrenStillBind(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	for _, text := range []string{
		"T | top strlen(A) by N",
		"T | top 5 by strlen(A)",
		"T | top 5 by N + 1",
		"T | top 5 by T.N",
		"T | top by N",
		"T | top 5 by",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			top := firstBoundNode[*TopOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, top), "A", "N")
			firstBoundNode[*PipeExpression](t, query)
		})
	}
}

func TestNativeTreeBinderTopScalarChildDiagnostics(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
		NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
	)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		name      string
		query     string
		want      []string
		doNotWant []string
	}{
		{name: "invalid addition suppresses integer cascade", query: "T | top A + 1 by N", want: []string{"KS106"}, doNotWant: []string{"KS134"}},
		{name: "invalid element suppresses integer cascade", query: "T | top A[0] by N", want: []string{"KS147"}, doNotWant: []string{"KS134"}},
		{name: "predicate result remains noninteger", query: "T | top A > N by N", want: []string{"KS134", "KS106"}},
		{name: "invalid count path suppresses integer cascade", query: "T | top A.k by N", want: []string{"KS142"}, doNotWant: []string{"KS134"}},
		{name: "invalid key path remains a child error", query: "T | top 5 by A.k", want: []string{"KS142"}, doNotWant: []string{"KS108"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.query)
			NewBinder(globals).TryBind(query)
			firstBoundNode[*TopOperator](t, query)
			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeSemantic)
			for _, code := range tt.want {
				if !diagnosticsContain(diagnostics, code) {
					t.Fatalf("diagnostics = %#v, want %s", diagnostics, code)
				}
			}
			for _, code := range tt.doNotWant {
				if diagnosticsContain(diagnostics, code) {
					t.Fatalf("diagnostics = %#v, do not want %s", diagnostics, code)
				}
			}
		})
	}
}

func TestNativeTreeBinderTopAcceptsDynamicKeyAndParameters(t *testing.T) {
	dynamic := NewColumnSymbol("D", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("T", dynamic)
	query := ParseQuery("T | top hint.progressive_top=true 5 by D")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	top := firstBoundNode[*TopOperator](t, query)
	assertDiagnosticCodes(t, top)
	if result := semanticTable(t, top); result.GetColumn("D") != dynamic || !result.IsSorted() {
		t.Fatalf("top result = %#v, want original dynamic column and sorted state", result)
	}
}

func TestNativeTreeBinderTopPreservesOpenTableInferredColumnIdentity(t *testing.T) {
	declared := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("Open", declared).WithIsOpen(true).WithIsSerialized(true)
	query := ParseQuery("Open | top Inferred by Inferred | project Inferred")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	references := boundNodes[*NameReference](query)
	var inferred []*NameReference
	for _, reference := range references {
		if NameReferenceSimpleName(reference) == "Inferred" {
			inferred = append(inferred, reference)
		}
	}
	if len(inferred) != 3 {
		t.Fatalf("inferred references = %d, want 3", len(inferred))
	}
	column, ok := inferred[0].ReferencedSymbol().(*ColumnSymbol)
	if !ok || column == nil {
		t.Fatalf("inferred symbol = %T, want *ColumnSymbol", inferred[0].ReferencedSymbol())
	}
	for _, reference := range inferred[1:] {
		if reference.ReferencedSymbol() != column {
			t.Fatal("open-table inferred column identity was not stable")
		}
	}
	top := firstBoundNode[*TopOperator](t, query)
	assertDiagnosticCodes(t, top)
	result := semanticTable(t, top)
	assertOrderedColumns(t, result, "A", "Inferred")
	if result.GetColumn("A") != declared || result.GetColumn("Inferred") != column || !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatal("top did not preserve declared and inferred open-table state")
	}
	project := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	if project.GetColumn("Inferred") != column {
		t.Fatal("downstream project did not preserve inferred column identity")
	}
}

func TestNativeTreeBinderSortAliasesPreserveRowsAndSort(t *testing.T) {
	for _, keyword := range []string{"sort", "order"} {
		t.Run(keyword, func(t *testing.T) {
			a := NewColumnSymbol("A", ScalarTypes.String.Type())
			n := NewColumnSymbol("N", ScalarTypes.Long.Type())
			table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSerialized(true)
			query := ParseQuery("T | " + keyword + " by N desc, A")
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

			sort := firstBoundNode[*SortOperator](t, query)
			result := semanticTable(t, sort)
			assertOrderedColumns(t, result, "A", "N")
			if result.GetColumn("A") != a || result.GetColumn("N") != n {
				t.Fatal("sort did not preserve source column identity")
			}
			if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
				t.Fatalf("sort state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
			}
		})
	}
}

func TestNativeTreeBinderSortDiagnostics(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
		NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
	)
	other := NewTableSymbol("Other", NewColumnSymbol("X", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table, other))

	firstText := "sort hint.strategy=bogus by Other"
	first := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, firstText).(*SortOperator)
	NewBinder(globals).TryBind(first)
	assertDiagnosticCodes(t, first, "KS174", "KS005", "KS108")
	assertDiagnosticLocations(t, GetContainedDiagnostics(first, DiagnosticsIncludeSemantic),
		diagnosticLocation{"KS174", strings.Index(firstText, "sort"), len("sort")},
		diagnosticLocation{"KS005", strings.Index(firstText, "bogus"), len("bogus")},
		diagnosticLocation{"KS108", strings.Index(firstText, "Other"), len("Other")},
	)

	dynamicText := "T\n| sort hint.strategy=bogus by D"
	dynamic := ParseQuery(dynamicText)
	NewBinder(globals).TryBind(dynamic)
	dynamicSort := firstBoundNode[*SortOperator](t, dynamic)
	assertDiagnosticCodes(t, dynamicSort, "KS005", "KS116")
	assertDiagnosticLocations(t, dynamicSort.SemanticDiagnostics(),
		diagnosticLocation{"KS005", strings.Index(dynamicText, "bogus"), len("bogus")},
		diagnosticLocation{"KS116", strings.LastIndex(dynamicText, "D"), len("D")},
	)

	nonScalarText := "T | sort by Other, D"
	nonScalar := ParseQuery(nonScalarText)
	NewBinder(globals).TryBind(nonScalar)
	nonScalarSort := firstBoundNode[*SortOperator](t, nonScalar)
	assertDiagnosticCodes(t, nonScalarSort, "KS108")
	assertDiagnosticLocations(t, nonScalarSort.SemanticDiagnostics(),
		diagnosticLocation{"KS108", strings.Index(nonScalarText, "Other"), len("Other")},
	)
}

func TestNativeTreeBinderSortParametersAreBoundAndCaseInsensitive(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	valid := ParseQuery("T | sort hint.strategy=SPLITBLOCK by N")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(valid)
	operator := firstBoundNode[*SortOperator](t, valid)
	assertDiagnosticCodes(t, operator)
	parameter := operator.Parameters().Elements()[0]
	if parameter.Name().ReferencedSymbol() != QueryOperatorParameters.SortParameters[0] {
		t.Fatalf("sort parameter symbol = %#v, want SortParameters declaration", parameter.Name().ReferencedSymbol())
	}

	duplicateText := "T | order hint.strategy=splitBlock hint.strategy=multipleBlocks by N"
	duplicate := ParseQuery(duplicateText)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(duplicate)
	duplicateOperator := firstBoundNode[*SortOperator](t, duplicate)
	assertDiagnosticCodes(t, duplicateOperator, "KS180")
	secondName := strings.LastIndex(duplicateText, "hint.strategy")
	assertDiagnosticLocations(t, duplicateOperator.SemanticDiagnostics(),
		diagnosticLocation{"KS180", secondName, len("hint.strategy")},
	)
}

func TestNativeTreeBinderSortMalformedAndErrorChildrenStillBind(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
		NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
	)
	for _, text := range []string{
		"T | sort by Missing",
		"T | sort by strlen(A)",
		"T | sort by N + 1",
		"T | sort by T.N",
		"T | sort by D",
		"T | sort by",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			if text == "T | sort by Missing" {
				assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS142")
			}
			sort := firstBoundNode[*SortOperator](t, query)
			result := semanticTable(t, sort)
			assertOrderedColumns(t, result, "A", "N", "D")
			if !result.IsSorted() {
				t.Fatal("sort result was not marked sorted")
			}
			for _, pipe := range boundNodes[*PipeExpression](query) {
				if !pipe.IsBound() {
					t.Fatal("pipe containing malformed/error sort child was not bound")
				}
			}
		})
	}
}

func TestNativeTreeBinderSortFeedsDownstreamProject(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n)
	query := ParseQuery("T | order by N | project N, A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	project := firstBoundNode[*ProjectOperator](t, query)
	result := semanticTable(t, project)
	assertOrderedColumns(t, result, "N", "A")
	if result.GetColumn("N") != n || result.GetColumn("A") != a {
		t.Fatal("downstream project did not resolve sort output columns")
	}
}

func TestNativeTreeBinderSortUsesSyntacticPipeContext(t *testing.T) {
	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "sort by 1").(*SortOperator)
	NewBinder(nil).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")

	piped := ParseQuery("Missing | sort by 1")
	NewBinder(nil).TryBind(piped)
	operator := firstBoundNode[*SortOperator](t, piped)
	if hasSemanticDiagnostic(operator, "KS174") {
		t.Fatalf("syntactically piped sort diagnostics = %#v, do not want KS174", operator.SemanticDiagnostics())
	}
	if !semanticTable(t, operator).IsSorted() {
		t.Fatal("syntactically piped sort result was not marked sorted")
	}
}

func TestNativeTreeBinderSortPreservesOpenInferredColumnIdentity(t *testing.T) {
	declared := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("Open", declared).WithIsOpen(true).WithIsSerialized(true)
	query := ParseQuery("Open | sort by Inferred | order by Inferred | project Inferred")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	var references []*NameReference
	for _, reference := range boundNodes[*NameReference](query) {
		if NameReferenceSimpleName(reference) == "Inferred" {
			references = append(references, reference)
		}
	}
	if len(references) != 3 {
		t.Fatalf("inferred references = %d, want 3", len(references))
	}
	inferred, ok := references[0].ReferencedSymbol().(*ColumnSymbol)
	if !ok {
		t.Fatalf("inferred symbol = %T, want *ColumnSymbol", references[0].ReferencedSymbol())
	}
	for _, reference := range references[1:] {
		if reference.ReferencedSymbol() != inferred {
			t.Fatal("sort did not preserve stable open inferred column identity")
		}
	}
	for _, operator := range boundNodes[*SortOperator](query) {
		result := semanticTable(t, operator)
		if result.GetColumn("A") != declared || result.GetColumn("Inferred") != inferred || !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
			t.Fatal("sort did not preserve declared/inferred columns and inheritable state")
		}
	}
}

func TestNativeTreeBinderBareSerializePreservesRowsStateAndFeedsDownstreamScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSorted(true)
	query := ParseQuery("T | serialize | project N, A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	serialize := firstBoundNode[*SerializeOperator](t, query)
	result := semanticTable(t, serialize)
	assertOrderedColumns(t, result, "A", "N")
	if result.GetColumn("A") != a || result.GetColumn("N") != n {
		t.Fatal("serialize did not preserve source column identity")
	}
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("serialize state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}

	project := firstBoundNode[*ProjectOperator](t, query)
	projected := semanticTable(t, project)
	assertOrderedColumns(t, projected, "N", "A")
	if projected.GetColumn("N") != n || projected.GetColumn("A") != a {
		t.Fatal("downstream project did not resolve serialize output columns")
	}
}

func TestNativeTreeBinderBareSerializeNotFirst(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "serialize").(*SerializeOperator)
	NewBinder(NewGlobalState(nil)).TryBind(operator)

	assertSemanticDiagnostic(t, operator, "KS174")
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result)
	if !result.IsSerialized() {
		t.Fatal("standalone serialize result was not serialized")
	}
}

func TestNativeTreeBinderSerializeProjectionOrderIdentityLiteralsAndNames(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	column1 := NewColumnSymbol("Column1", ScalarTypes.Int.Type())
	table := NewTableSymbol("T", a, column1).WithIsSorted(true)
	query := ParseQuery("T | serialize A, 1, 'x', Copy=A, Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*SerializeOperator](t, query)
	if operator.Parameters() == nil || operator.Parameters().Count() != 0 {
		t.Fatalf("serialize parameters = %#v, want validated empty list", operator.Parameters())
	}
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A", "Column1", "Column2", "Column3", "Copy", "Missing")
	if result.GetColumn("A") != a || result.GetColumn("Column1") != column1 {
		t.Fatal("serialize replaced seeded input column identities")
	}
	if result.GetColumn("Copy") == a || !sameType(result.GetColumn("Copy").Type, ScalarTypes.String.Type()) {
		t.Fatal("named column projection did not create the source-faithful alias")
	}
	if !sameType(result.GetColumn("Column2").Type, ScalarTypes.Long.Type()) ||
		!sameType(result.GetColumn("Column3").Type, ScalarTypes.String.Type()) ||
		!sameType(result.GetColumn("Missing").Type, ScalarTypes.Unknown.Type()) {
		t.Fatalf("serialize projection types = %#v", result.Columns())
	}
	if !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("serialize state = sorted:%t serialized:%t", result.IsSorted(), result.IsSerialized())
	}
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS142")
	assertDiagnosticCodes(t, operator)
}

func TestNativeTreeBinderSerializeDiagnosticsOrderAndLocations(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a)
	other := NewTableSymbol("Other", NewColumnSymbol("B", ScalarTypes.Long.Type()))
	text := "serialize Other, Other, A, A"
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, text).(*SerializeOperator)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, other)))
	binder.rowScope = table
	binder.TryBind(operator)

	assertDiagnosticCodes(t, operator, "KS174", "KS108", "KS108", "KS171")
	firstOther := strings.Index(text, "Other")
	secondOther := strings.LastIndex(text, "Other")
	secondA := strings.LastIndex(text, "A")
	assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
		diagnosticLocation{"KS174", 0, len("serialize")},
		diagnosticLocation{"KS108", firstOther, len("Other")},
		diagnosticLocation{"KS108", secondOther, len("Other")},
		diagnosticLocation{"KS171", secondA, len("A")},
	)
	assertOrderedColumns(t, semanticTable(t, operator), "A", "Other", "Other1")
}

func TestNativeTreeBinderSerializeMissingReferencesBindUnknownAndFeedPipe(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("T | serialize Missing, Copy=AlsoMissing | project Copy, Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	for _, name := range []string{"Missing", "AlsoMissing"} {
		assertSemanticDiagnostic(t, firstNamedReference(t, query, name), "KS142")
	}
	operator := firstBoundNode[*SerializeOperator](t, query)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A", "Missing", "Copy")
	if !sameType(result.GetColumn("Missing").Type, ScalarTypes.Unknown.Type()) || !sameType(result.GetColumn("Copy").Type, ScalarTypes.Unknown.Type()) {
		t.Fatalf("recoverable serialize columns = %#v, want unknown types", result.Columns())
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, project), "Copy", "Missing")
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if !pipe.IsBound() {
			t.Fatal("pipe containing recoverable serialize reference was not bound")
		}
	}
}

func TestNativeTreeBinderSerializePreservesOpenInferredIdentity(t *testing.T) {
	declared := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("Open", declared).WithIsOpen(true).WithIsSorted(true)
	query := ParseQuery("Open | serialize Inferred, Copy=1 | serialize Inferred | project Inferred")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	var references []*NameReference
	for _, reference := range boundNodes[*NameReference](query) {
		if NameReferenceSimpleName(reference) == "Inferred" {
			references = append(references, reference)
		}
	}
	if len(references) != 3 {
		t.Fatalf("inferred references = %d, want 3", len(references))
	}
	inferred, ok := references[0].ReferencedSymbol().(*ColumnSymbol)
	if !ok {
		t.Fatalf("inferred symbol = %T, want *ColumnSymbol", references[0].ReferencedSymbol())
	}
	for _, reference := range references[1:] {
		if reference.ReferencedSymbol() != inferred {
			t.Fatal("serialize did not preserve inferred column identity")
		}
	}
	for _, serialize := range boundNodes[*SerializeOperator](query) {
		result := semanticTable(t, serialize)
		if result.GetColumn("A") != declared || result.GetColumn("Inferred") != inferred || !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
			t.Fatal("serialize did not preserve declared/inferred columns and inheritable state")
		}
	}
}

func TestNativeTreeBinderSerializeErrorValuedExpressionRecoversOutput(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	query := ParseQuery("T | serialize A[0]")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	serialize := firstBoundNode[*SerializeOperator](t, query)
	if !serialize.IsBound() {
		t.Fatal("error-valued serialize was not bound")
	}
	result := semanticTable(t, serialize)
	assertOrderedColumns(t, result, "A", "N", "0")
	if column := result.GetColumn("0"); column == nil || column.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("0 = %#v, want current unknown", column)
	}
}

func TestNativeTreeBinderSerializeMalformedExpressionsRecoverOutput(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	tests := []struct {
		query   string
		columns []string
	}{
		{query: "T | serialize Copy=", columns: []string{"A", "N", "Copy"}},
		{query: "T | serialize Copy=N,", columns: []string{"A", "N", "Copy"}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			query := ParseQuery(tt.query)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

			operator := firstBoundNode[*SerializeOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, operator), tt.columns...)
			if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS006", "Missing expression") {
				t.Fatalf("missing parser recovery diagnostic: %#v", query.ContainedDiagnostics())
			}
			for _, pipe := range boundNodes[*PipeExpression](query) {
				if !pipe.IsBound() {
					t.Fatal("pipe containing recoverable malformed serialize expression was not bound")
				}
			}
		})
	}
}

func TestNativeTreeBinderSerializeMissingParameterValueStillBinds(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("T | serialize hint.strategy=")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*SerializeOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS179")
	assertOrderedColumns(t, semanticTable(t, operator), "A")
	if !diagnosticsContainMessage(query.ContainedDiagnostics(), "KS005", "Expected one of: auto, broadcast, centralized, shuffle") {
		t.Fatalf("missing parameter-value parser diagnostic: %#v", query.ContainedDiagnostics())
	}
}

func TestNativeTreeBinderSerializeValidatesEmptyParameterSet(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	text := "T | serialize hint.strategy=shuffle"
	query := ParseQuery(text)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*SerializeOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS179")
	start := strings.Index(text, "hint.strategy")
	assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
		diagnosticLocation{"KS179", start, len("hint.strategy")},
	)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A")
	if result.GetColumn("A") != table.GetColumn("A") || !result.IsSerialized() {
		t.Fatal("serialize parameter diagnostic path did not preserve and serialize input")
	}
}

func TestNativeTreeBinderWhereBindsMissingColumn(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()), NewColumnSymbol("N", ScalarTypes.Long.Type()))
	query := ParseQuery("T | where Missing == 1")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	missing := firstNamedReference(t, query, "Missing")
	assertSemanticDiagnostic(t, missing, "KS142")
	filter := firstBoundNode[*FilterOperator](t, query)
	if hasSemanticDiagnostic(filter, "KS141") {
		t.Fatalf("where added a boolean diagnostic for a bound comparison: %#v", filter.SemanticDiagnostics())
	}
	assertColumnTypes(t, semanticTable(t, filter), map[string]*TypeSymbol{
		"A": ScalarTypes.String.Type(),
		"N": ScalarTypes.Long.Type(),
	})
}

func TestNativeTreeBinderFilterStandaloneDiagnosticsOrderAndLocations(t *testing.T) {
	for _, keyword := range []string{"where", "filter"} {
		t.Run(keyword, func(t *testing.T) {
			text := keyword + " hint.strategy=shuffle 1"
			operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, text).(*FilterOperator)
			NewBinder(NewGlobalState(nil)).TryBind(operator)

			assertDiagnosticCodes(t, operator, "KS174", "KS179", "KS141")
			assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
				diagnosticLocation{"KS174", 0, len(keyword)},
				diagnosticLocation{"KS179", strings.Index(text, "hint.strategy"), len("hint.strategy")},
				diagnosticLocation{"KS141", strings.LastIndex(text, "1"), 1},
			)
		})
	}
}

func TestNativeTreeBinderFilterUsesSyntacticPipeContext(t *testing.T) {
	for _, keyword := range []string{"where", "filter"} {
		t.Run(keyword, func(t *testing.T) {
			text := "Missing | " + keyword + " 1"
			query := ParseQuery(text)
			binder := NewBinder(NewGlobalState(nil))
			binder.rowScope = NewTableSymbol("Ambient")
			binder.TryBind(query)

			operator := firstBoundNode[*FilterOperator](t, query)
			assertDiagnosticCodes(t, operator, "KS141")
			assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
				diagnosticLocation{"KS141", strings.LastIndex(text, "1"), 1},
			)
			assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS204")
		})
	}
}

func TestNativeTreeBinderFilterAcceptsBooleanForms(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Enabled", ScalarTypes.Bool.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	for _, keyword := range []string{"where", "filter"} {
		for _, condition := range []string{"true", "Enabled", "(Enabled)", "N == 1", "Enabled and (N > 0)"} {
			t.Run(keyword+"/"+condition, func(t *testing.T) {
				query := ParseQuery("T | " + keyword + " " + condition)
				NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

				operator := firstBoundNode[*FilterOperator](t, query)
				assertDiagnosticCodes(t, operator)
				if semanticTable(t, operator).GetColumn("Enabled") != table.GetColumn("Enabled") {
					t.Fatal("filter replaced a source column on a valid boolean condition")
				}
			})
		}
	}
}

func TestNativeTreeBinderFilterReportsExactBooleanType(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Text", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	for _, keyword := range []string{"where", "filter"} {
		for _, condition := range []string{"Text", "N", "'text'", "1"} {
			t.Run(keyword+"/"+condition, func(t *testing.T) {
				text := "T | " + keyword + " " + condition
				query := ParseQuery(text)
				NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

				operator := firstBoundNode[*FilterOperator](t, query)
				assertDiagnosticCodes(t, operator, "KS141")
				assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
					diagnosticLocation{"KS141", strings.Index(text, condition), len(condition)},
				)
			})
		}
	}
}

func TestNativeTreeBinderFilterRecoversErrorUnknownAndMissingConditions(t *testing.T) {
	closed := NewTableSymbol("Closed", NewColumnSymbol("A", ScalarTypes.String.Type()))
	for _, keyword := range []string{"where", "filter"} {
		t.Run(keyword+"/error", func(t *testing.T) {
			query := ParseQuery("Closed | " + keyword + " Missing")
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", closed))).TryBind(query)

			operator := firstBoundNode[*FilterOperator](t, query)
			assertDiagnosticCodes(t, operator)
			assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS142")
			if !operator.IsBound() {
				t.Fatal("filter with an error condition was not bound")
			}
		})

		t.Run(keyword+"/missing", func(t *testing.T) {
			query := ParseQuery("Closed | " + keyword)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", closed))).TryBind(query)

			operator := firstBoundNode[*FilterOperator](t, query)
			assertDiagnosticCodes(t, operator)
			if !diagnosticsContainMessage(operator.Condition().ContainedDiagnostics(), "KS006", "Missing expression") {
				t.Fatalf("missing condition parser diagnostic was lost: %#v", operator.Condition().ContainedDiagnostics())
			}
			if !operator.IsBound() || semanticTable(t, operator).GetColumn("A") != closed.GetColumn("A") {
				t.Fatal("filter with a parser-recovered missing condition did not preserve its row scope")
			}
		})
	}
}

func TestNativeTreeBinderFilterMalformedPredicatesPreserveDownstreamScope(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	for _, text := range []string{
		"T | where T. | count",
		"T | where A in ( | count",
		"T | where | count",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

			filter := firstBoundNode[*FilterOperator](t, query)
			if !filter.IsBound() || semanticTable(t, filter).GetColumn("A") != table.GetColumn("A") {
				t.Fatal("malformed filter did not preserve its input row scope")
			}
			count := firstBoundNode[*CountOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, count), "Count")
			firstBoundNode[*PipeExpression](t, query)
		})
	}
}

func TestNativeTreeBinderFilterPreservesIdentityStateInferredOrderAndDownstreamScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | where Inferred | filter true | project Inferred, A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	filters := boundNodes[*FilterOperator](query)
	if len(filters) != 2 {
		t.Fatalf("filter operators = %d, want 2", len(filters))
	}
	first := semanticTable(t, filters[0])
	if first == table || first.SymbolName() != "" {
		t.Fatalf("filter result = %#v, want a fresh unnamed table", first)
	}
	assertOrderedColumns(t, first, "A", "N", "Inferred")
	inferred := first.GetColumn("Inferred")
	if first.GetColumn("A") != a || first.GetColumn("N") != n || inferred == nil ||
		!sameType(inferred.Type, ScalarTypes.Unknown.Type()) {
		t.Fatal("filter did not preserve declared and inferred column identities")
	}
	if !first.IsOpen() || !first.IsSorted() || !first.IsSerialized() {
		t.Fatalf("filter state = open:%t sorted:%t serialized:%t", first.IsOpen(), first.IsSorted(), first.IsSerialized())
	}
	second := semanticTable(t, filters[1])
	if second.GetColumn("Inferred") != inferred {
		t.Fatal("second filter replaced the inferred column identity")
	}
	project := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, project, "Inferred", "A")
	if project.GetColumn("Inferred") != inferred || project.GetColumn("A") != a {
		t.Fatal("downstream project did not use the filter row scope")
	}
	for _, filter := range filters {
		if hasSemanticDiagnostic(filter, "KS141") {
			t.Fatal("unknown inferred condition caused a boolean-type cascade")
		}
	}
}

func TestNativeTreeBinderFilterMalformedParametersRecover(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	for _, text := range []string{
		"T | where hint.strategy= true",
		"T | filter hint.strategy=",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

			operator := firstBoundNode[*FilterOperator](t, query)
			assertDiagnosticCodes(t, operator, "KS179")
			if len(query.ContainedDiagnostics()) == 0 {
				t.Fatal("malformed parameter parser diagnostics were lost")
			}
			if semanticTable(t, operator).GetColumn("A") != table.GetColumn("A") {
				t.Fatal("parameter recovery changed the output row scope")
			}
		})
	}
}

func TestNativeTreeBinderFilterCompleteScalarForms(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
		NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
	)
	for _, keyword := range []string{"where", "filter"} {
		for _, condition := range []string{
			"strlen(A) == 1",
			"N + 1",
			"T.A == 'x'",
			"A[0] == 'x'",
			"D.k == 2",
			"isnotempty(A) and strlen(A) > 0",
			"Flag=N > 0",
			"N > toscalar(print X=0)",
		} {
			t.Run(keyword+"/"+condition, func(t *testing.T) {
				query := ParseQuery("T | " + keyword + " " + condition)
				NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
				filter := firstBoundNode[*FilterOperator](t, query)
				assertOrderedColumns(t, semanticTable(t, filter), "A", "N", "D")
				firstBoundNode[*PipeExpression](t, query)
			})
		}
	}
}

func TestNativeTreeBinderFilterCompleteScalarDiagnostics(t *testing.T) {
	tests := []struct {
		condition string
		codes     []string
	}{
		{condition: "N + 1", codes: []string{"KS141"}},
		{condition: "N + A > 0", codes: []string{"KS106"}},
		{condition: "A[0] == \"x\"", codes: []string{"KS147"}},
	}
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	for _, tt := range tests {
		t.Run(tt.condition, func(t *testing.T) {
			query := ParseQuery("T | where " + tt.condition)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			firstBoundNode[*FilterOperator](t, query)
			diagnostics := GetContainedDiagnostics(query, DiagnosticsIncludeSemantic)
			if len(diagnostics) != len(tt.codes) {
				t.Fatalf("diagnostics = %#v, want codes %#v", diagnostics, tt.codes)
			}
			for i, code := range tt.codes {
				if diagnostics[i].Code != code {
					t.Fatalf("diagnostic %d = %#v, want code %s", i, diagnostics[i], code)
				}
			}
		})
	}
}

func TestNativeTreeBinderFilterRestoresBinderAndTreeState(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("Enabled", ScalarTypes.Bool.Type()))
	query := ParseQuery("T | where Enabled")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	local := binder.LocalScope
	marker := NewVariableSymbol("Marker", ScalarTypes.String.Type(), false, nil, nil)
	local.AddSymbol(marker)
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.bind(query)

	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatalf("binder state leaked after filter: row=%#v scope=%v local=%#v", binder.rowScope, binder.scopeKind, binder.LocalScope)
	}
	if !tree.inPipe || !tree.inWildcard {
		t.Fatalf("tree flags changed after filter: inPipe=%t inWildcard=%t", tree.inPipe, tree.inWildcard)
	}
}

func TestNativeTreeBinderProjectColumnsPreserveOrderIdentityAndState(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | project N, A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	projects := boundNodes[*ProjectOperator](query)
	if len(projects) != 1 {
		t.Fatalf("project operators = %d, want 1", len(projects))
	}
	result := semanticTable(t, projects[0])
	assertOrderedColumns(t, result, "N", "A")
	if result.GetColumn("N") != n || result.GetColumn("A") != a {
		t.Fatal("project did not preserve source column symbols")
	}
	if !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("project state = sorted:%t serialized:%t", result.IsSorted(), result.IsSerialized())
	}
}

func TestNativeTreeBinderProjectNamedScalarsAndDuplicateDeclarations(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	query := ParseQuery("T | project X=A, Literal=1, X=N")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	project := firstBoundNode[*ProjectOperator](t, query)
	result := semanticTable(t, project)
	assertOrderedColumns(t, result, "X", "Literal")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"X":       ScalarTypes.String.Type(),
		"Literal": ScalarTypes.Long.Type(),
	})
	assertSemanticDiagnostic(t, project, "KS171")

	declarations := boundNodes[*NameDeclaration](query)
	if len(declarations) != 3 {
		t.Fatalf("name declarations = %d, want 3", len(declarations))
	}
	for _, declaration := range declarations {
		column, ok := declaration.ReferencedSymbol().(*ColumnSymbol)
		if !ok || column.SymbolName() != NameDeclarationSimpleName(declaration) {
			t.Fatalf("declaration %q symbol = %#v", NameDeclarationSimpleName(declaration), declaration.ReferencedSymbol())
		}
	}
}

func TestNativeTreeBinderExtendReplacesAndAppendsInOrder(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	).WithIsOpen(true).WithIsSerialized(true)
	query := ParseQuery("T | extend N='replaced', Added=A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	extend := firstBoundNode[*ExtendOperator](t, query)
	result := semanticTable(t, extend)
	assertOrderedColumns(t, result, "A", "N", "Added")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"A":     ScalarTypes.String.Type(),
		"N":     ScalarTypes.String.Type(),
		"Added": ScalarTypes.String.Type(),
	})
	if !result.IsOpen() || !result.IsSerialized() {
		t.Fatalf("extend state = open:%t serialized:%t", result.IsOpen(), result.IsSerialized())
	}
}

func TestNativeTreeBinderProjectRenamePreservesShapeLineageAndState(t *testing.T) {
	a := NewColumnSymbolFull("A", ScalarTypes.String.Type(), "source description", nil, nil, nil)
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | project-rename Account=A | project Account, N")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	result := semanticTable(t, rename)
	assertOrderedColumns(t, result, "Account", "N")
	renamed := result.GetColumn("Account")
	if renamed == nil || renamed.Type != a.Type || renamed.Description != a.Description {
		t.Fatalf("renamed column = %#v, want source type and description", renamed)
	}
	if len(renamed.OriginalColumns) != 1 || renamed.OriginalColumns[0] != a {
		t.Fatalf("renamed lineage = %#v, want source A", renamed.OriginalColumns)
	}
	if result.GetColumn("N") != n {
		t.Fatal("unrenamed column identity was not preserved")
	}
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("rename state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}

	declarations := boundNodes[*NameDeclaration](query)
	if len(declarations) != 1 || declarations[0].ReferencedSymbol() != renamed {
		t.Fatalf("rename declaration symbol = %#v, want renamed column", declarations)
	}
	projects := boundNodes[*ProjectOperator](query)
	if len(projects) != 1 {
		t.Fatalf("downstream projects = %d, want 1", len(projects))
	}
	assertOrderedColumns(t, semanticTable(t, projects[0]), "Account", "N")
	if reference := firstNamedReference(t, projects[0], "Account"); reference.ReferencedSymbol() != renamed {
		t.Fatalf("downstream Account resolved to %#v, want renamed column", reference.ReferencedSymbol())
	}
}

func TestNativeTreeBinderProjectRenameDuplicateTarget(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	text := "T | project-rename X=A, X=N"
	query := ParseQuery(text)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	assertDiagnosticCodes(t, rename, "KS171")
	assertDiagnosticLocations(t, rename.SemanticDiagnostics(),
		diagnosticLocation{"KS171", strings.LastIndex(text, "X"), len("X")},
	)
	assertOrderedColumns(t, semanticTable(t, rename), "X", "N")
	declarations := boundNodes[*NameDeclaration](rename)
	if len(declarations) != 2 || declarations[0].ReferencedSymbol() != semanticTable(t, rename).GetColumn("X") || declarations[1].ReferencedSymbol() != nil {
		t.Fatalf("duplicate declarations = %#v", declarations)
	}
}

func TestNativeTreeBinderProjectRenameRecoveryContinuesAndFeedsDownstream(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsSorted(true)
	query := ParseQuery("T | project-rename X=A, Literal=1, Unknown=Missing, Path=A.foo, Y=N | project X, Y")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	assertDiagnosticCodes(t, rename, "KS109", "KS109", "KS109")
	assertOrderedColumns(t, semanticTable(t, rename), "X", "Y")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "X", "Y")
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS142")

	declarations := boundNodes[*NameDeclaration](rename)
	if len(declarations) != 5 {
		t.Fatalf("rename declarations = %d, want 5", len(declarations))
	}
	if declarations[0].ReferencedSymbol() != semanticTable(t, rename).GetColumn("X") ||
		declarations[4].ReferencedSymbol() != semanticTable(t, rename).GetColumn("Y") {
		t.Fatal("successful recovery declarations do not reference renamed columns")
	}
	for _, declaration := range declarations[1:4] {
		if declaration.ReferencedSymbol() != nil {
			t.Fatalf("failed rename declaration %q was mutated to %#v", NameDeclarationSimpleName(declaration), declaration.ReferencedSymbol())
		}
	}
}

func TestNativeTreeBinderProjectRenameDiagnosticsOrderAndLocations(t *testing.T) {
	text := "project-rename Bare, Literal=1, Missing=Unknown, (Left, Right)=A"
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, text).(*ProjectRenameOperator)
	NewBinder(NewGlobalState(nil)).TryBind(operator)

	assertDiagnosticCodes(t, operator, "KS174", "KS110", "KS109", "KS109", "KS110")
	assertDiagnosticLocations(t, operator.SemanticDiagnostics(),
		diagnosticLocation{"KS174", 0, len("project-rename")},
		diagnosticLocation{"KS110", strings.Index(text, "Bare"), len("Bare")},
		diagnosticLocation{"KS109", strings.Index(text, "1"), len("1")},
		diagnosticLocation{"KS109", strings.Index(text, "Unknown"), len("Unknown")},
		diagnosticLocation{"KS110", strings.Index(text, "(Left"), len("(Left, Right)=A")},
	)
	assertOrderedColumns(t, semanticTable(t, operator))
}

func TestNativeTreeBinderProjectRenameEmptyAndParserRecoveryPassthrough(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsSerialized(true)
	for _, text := range []string{
		"T | project-rename | project A, N",
		"T | project-rename X= | project A, N",
		"T | project-rename A, | project A, N",
		"T | project-rename 1, strlen(A), (X, Y)=A | project A, N",
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			rename := firstBoundNode[*ProjectRenameOperator](t, query)
			result := semanticTable(t, rename)
			assertOrderedColumns(t, result, "A", "N")
			if result.GetColumn("A") != a || result.GetColumn("N") != n || !result.IsSerialized() {
				t.Fatalf("recovery passthrough lost identity/state: %#v", result)
			}
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "A", "N")
		})
	}
}

func TestNativeTreeBinderProjectRenameMissingCommaStillBindsDownstream(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	query := ParseQuery("T | project-rename X=A Y=N | count")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	if !rename.IsBound() {
		t.Fatal("missing-comma project-rename was rolled back")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if !pipe.IsBound() {
			t.Fatal("missing-comma recovery pipe was not bound")
		}
	}
}

func TestNativeTreeBinderProjectRenameUsesOriginalScopeWithoutAliasChaining(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n)
	query := ParseQuery("T | project-rename X=A, Chained=X, Z=N | project X, Z")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	assertDiagnosticCodes(t, rename, "KS109")
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "X"), "KS142")
	result := semanticTable(t, rename)
	assertOrderedColumns(t, result, "X", "Z")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "X", "Z")
}

func TestNativeTreeBinderProjectRenameIncludesInferredOpenColumns(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a).WithIsOpen(true)
	query := ParseQuery("T | project-rename Known=Inferred | project A, Known")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	assertDiagnosticCodes(t, rename)
	result := semanticTable(t, rename)
	assertOrderedColumns(t, result, "A", "Known")
	known := result.GetColumn("Known")
	if known == nil || len(known.OriginalColumns) != 1 || known.OriginalColumns[0].SymbolName() != "Inferred" {
		t.Fatalf("inferred rename lineage = %#v", known)
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "A", "Known")
}

func TestNativeTreeBinderProjectRenameUnknownInputUsesOpenRecoveryScope(t *testing.T) {
	query := ParseQuery("Missing | project-rename X=A | count")
	NewBinder(NewGlobalState(nil)).TryBind(query)

	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS204")
	if diagnostics := firstNamedReference(t, query, "A").SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("inferred open-input column diagnostics = %#v", diagnostics)
	}
	rename := firstBoundNode[*ProjectRenameOperator](t, query)
	assertDiagnosticCodes(t, rename)
	assertOrderedColumns(t, semanticTable(t, rename), "X")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")
}

func TestNativeTreeBinderProjectRenameInjectedListRollsBackWithoutMutation(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", column)
	parsed := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project-rename X=A").(*ProjectRenameOperator)
	list := AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("Expression",
		parsed.Expressions().At(0).SeparatedElement.Clone(true).(*SeparatedElement),
		nil,
	))
	operator := MustNewProjectRenameOperatorTyped(parsed.ProjectRenameKeyword().Clone(true).(*SyntaxToken), list)

	binder := NewBinder(NewGlobalState(nil))
	binder.rowScope = table
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	marker := NewVariableSymbol("Marker", ScalarTypes.Bool.Type(), false, nil, nil)
	local.AddSymbol(marker)
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.bind(operator)

	if operator.IsBound() {
		t.Fatal("project-rename with injected nil list element was bound")
	}
	declaration := firstBoundNode[*NameDeclaration](t, operator)
	if declaration.ReferencedSymbol() != nil {
		t.Fatalf("rollback mutated declaration to %#v", declaration.ReferencedSymbol())
	}
	reference := firstNamedReference(t, operator, "A")
	if !reference.IsBound() || reference.ReferencedSymbol() != column {
		t.Fatalf("already visited child binding = %#v", reference.ReferencedSymbol())
	}
	if binder.rowScope != table || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local || localVariable(t, local, "Marker") != marker {
		t.Fatal("structural rollback changed binder or local scope")
	}
	if !tree.inPipe || !tree.inWildcard {
		t.Fatalf("structural rollback changed tree flags: inPipe=%t inWildcard=%t", tree.inPipe, tree.inWildcard)
	}
}

func TestNativeTreeBinderProjectRenameNotFirstDiagnostic(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.String.Type())
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project-rename X=A").(*ProjectRenameOperator)
	reference := firstNamedReference(t, operator, "A")
	reference.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	info := newSimpleNodeBinder(NewBinder(NewGlobalState(nil))).bind(operator.SyntaxNode)
	if info == nil {
		t.Fatal("supported project-rename shape was not bound")
	}
	operator.SetSemanticInfo(*info)

	assertSemanticDiagnostic(t, operator, "KS174")
	assertOrderedColumns(t, semanticTable(t, operator))
}

func TestNativeTreeBinderProjectionDiagnostics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))

	missingQuery := ParseQuery("T | project Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(missingQuery)
	missing := firstNamedReference(t, missingQuery, "Missing")
	assertSemanticDiagnostic(t, missing, "KS142")
	missingProject := firstBoundNode[*ProjectOperator](t, missingQuery)
	if !missingProject.IsBound() {
		t.Fatal("project with a bare unresolved reference was not bound")
	}
	missingResult := semanticTable(t, missingProject)
	assertOrderedColumns(t, missingResult, "Missing")
	if column := missingResult.GetColumn("Missing"); column == nil || column.Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("Missing = %#v, want current unknown", column)
	}
	for _, pipe := range boundNodes[*PipeExpression](missingQuery) {
		if !pipe.IsBound() {
			t.Fatal("pipe containing recoverable bare projection was not bound")
		}
	}

	project := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "project X=1").(*ProjectOperator)
	NewBinder(NewGlobalState(nil)).TryBind(project)
	assertSemanticDiagnostic(t, project, "KS174")
	assertOrderedColumns(t, semanticTable(t, project), "X")

	extend := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "extend X=1").(*ExtendOperator)
	NewBinder(NewGlobalState(nil)).TryBind(extend)
	assertSemanticDiagnostic(t, extend, "KS174")
}

func TestNativeTreeBinderProjectionErrorValuedFormsRecoverOutput(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A", ScalarTypes.String.Type()),
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
	)
	for _, queryText := range []string{
		"T | project X=A.foo",
		"T | extend X=A.foo",
	} {
		t.Run(queryText, func(t *testing.T) {
			query := ParseQuery(queryText)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			var result *TableSymbol
			projects := boundNodes[*ProjectOperator](query)
			if len(projects) != 0 {
				project := projects[0]
				if !project.IsBound() {
					t.Fatal("error-valued project was not bound")
				}
				result = semanticTable(t, project)
			} else {
				extend := firstBoundNode[*ExtendOperator](t, query)
				if !extend.IsBound() {
					t.Fatal("error-valued extend was not bound")
				}
				result = semanticTable(t, extend)
			}
			if column := result.GetColumn("X"); column == nil || column.Type != ScalarTypes.Unknown.Type() {
				t.Fatalf("X = %#v, want current unknown", column)
			}
		})
	}
}

func TestNativeTreeBinderProjectAwayKeepPreserveIdentityOrderStateAndDownstreamScope(t *testing.T) {
	a1 := NewColumnSymbol("AccountId", ScalarTypes.String.Type())
	a2 := NewColumnSymbol("AccountType", ScalarTypes.String.Type())
	user := NewColumnSymbol("User Name", ScalarTypes.String.Type())
	count := NewColumnSymbol("Count", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", count, a1, user, a2).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | project-keep Account*, ['User Name'], AccountId | project-away AccountType | project ['User Name'], AccountId")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	keep := firstBoundNode[*ProjectKeepOperator](t, query)
	kept := semanticTable(t, keep)
	assertOrderedColumns(t, kept, "AccountId", "User Name", "AccountType")
	if kept.GetColumn("AccountId") != a1 || kept.GetColumn("User Name") != user || kept.GetColumn("AccountType") != a2 {
		t.Fatal("project-keep replaced incoming column identities")
	}
	if !kept.IsOpen() || !kept.IsSorted() || !kept.IsSerialized() || kept.SymbolName() != "" || kept == table {
		t.Fatalf("project-keep result state/name/identity = open:%t sorted:%t serialized:%t name:%q same:%t", kept.IsOpen(), kept.IsSorted(), kept.IsSerialized(), kept.SymbolName(), kept == table)
	}

	wildcard := firstNamedReference(t, keep, "Account*")
	group, ok := wildcard.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 2 || group.Members()[0] != a1 || group.Members()[1] != a2 {
		t.Fatalf("wildcard symbol = %#v, want ordered Account columns", wildcard.ReferencedSymbol())
	}
	exact := firstNamedReference(t, keep, "AccountId")
	if exact.ReferencedSymbol() != a1 {
		t.Fatalf("exact symbol = %#v, want original AccountId", exact.ReferencedSymbol())
	}

	away := firstBoundNode[*ProjectAwayOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, away), "AccountId", "User Name")
	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, projected, "User Name", "AccountId")
}

func TestNativeTreeBinderProjectAwayKeepDuplicateAndCaseSensitiveWildcardBinding(t *testing.T) {
	first := NewColumnSymbol("Dup", ScalarTypes.String.Type())
	second := NewColumnSymbol("Dup", ScalarTypes.Long.Type())
	lower := NewColumnSymbol("dupLower", ScalarTypes.String.Type())
	table := NewTableSymbol("T", first, lower, second)

	query := ParseQuery("T | project-keep Dup, D*")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	keep := firstBoundNode[*ProjectKeepOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, keep), "Dup", "Dup")
	for _, pattern := range []string{"Dup", "D*"} {
		group, ok := firstNamedReference(t, keep, pattern).ReferencedSymbol().(*GroupSymbol)
		if !ok || len(group.Members()) != 2 || group.Members()[0] != first || group.Members()[1] != second {
			t.Fatalf("%s symbol = %#v, want duplicate-column group", pattern, group)
		}
	}

	lowerQuery := ParseQuery("T | project-keep d*")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(lowerQuery)
	lowerKeep := firstBoundNode[*ProjectKeepOperator](t, lowerQuery)
	assertOrderedColumns(t, semanticTable(t, lowerKeep), "dupLower")
	if firstNamedReference(t, lowerKeep, "d*").ReferencedSymbol().(*ColumnSymbol) != lower {
		t.Fatal("case-sensitive wildcard did not bind the lowercase column")
	}
}

func TestNativeTreeBinderProjectAwayKeepEmptyAndOpenZeroMatch(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", column).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	for _, queryText := range []string{"T | project-away", "T | project-keep"} {
		t.Run(queryText, func(t *testing.T) {
			query := ParseQuery(queryText)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			pipe := firstBoundNode[*PipeExpression](t, query)
			result := semanticTable(t, pipe)
			if queryText == "T | project-away" {
				assertOrderedColumns(t, result, "A")
				if result.GetColumn("A") != column {
					t.Fatal("empty project-away did not preserve source column identity")
				}
			} else {
				assertOrderedColumns(t, result)
			}
			if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
				t.Fatalf("empty projection state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
			}
		})
	}

	query := ParseQuery("T | project-keep Missing*")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	keep := firstBoundNode[*ProjectKeepOperator](t, query)
	reference := firstNamedReference(t, keep, "Missing*")
	group, ok := reference.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 0 || len(reference.SemanticDiagnostics()) != 0 {
		t.Fatalf("open zero-match wildcard = symbol:%#v diagnostics:%#v", reference.ReferencedSymbol(), reference.SemanticDiagnostics())
	}
	assertOrderedColumns(t, semanticTable(t, keep))

	closed := NewTableSymbol("Closed", column)
	closedQuery := ParseQuery("Closed | project-keep Missing*")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", closed))).TryBind(closedQuery)
	assertSemanticDiagnostic(t, firstNamedReference(t, closedQuery, "Missing*"), "KS142")
	closedKeep := firstBoundNode[*ProjectKeepOperator](t, closedQuery)
	assertSemanticDiagnostic(t, closedKeep, "KS109")
	assertOrderedColumns(t, semanticTable(t, closedKeep))
}

func TestNativeTreeBinderRejectsGenericWildcardNameReference(t *testing.T) {
	query := ParseQuery(`[Account*]`)
	NewBinder(NewGlobalState(nil)).TryBind(query)
	reference := firstNamedReference(t, query, "Account*")
	assertSemanticDiagnostic(t, reference, "KS268")
	if reference.ReferencedSymbol() != Error {
		t.Fatalf("generic wildcard symbol = %#v, want Error", reference.ReferencedSymbol())
	}
}

func TestNativeTreeBinderProjectAwayKeepDiagnosticsAndTransactions(t *testing.T) {
	column := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", column)

	for _, text := range []string{"project-away", "project-keep"} {
		t.Run("standalone "+text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(nil)).TryBind(query)
			if text == "project-away" {
				assertSemanticDiagnostic(t, firstBoundNode[*ProjectAwayOperator](t, query), "KS174")
			} else {
				assertSemanticDiagnostic(t, firstBoundNode[*ProjectKeepOperator](t, query), "KS174")
			}
		})
	}

	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	variable := NewVariableSymbol("Local", ScalarTypes.Long.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(variable)
	known := ParseQuery("T | project-away Local")
	binder.TryBind(known)
	knownOperator := firstBoundNode[*ProjectAwayOperator](t, known)
	assertDiagnosticCodes(t, knownOperator, "KS109")
	if firstNamedReference(t, known, "Local").ReferencedSymbol() != variable {
		t.Fatal("known non-column did not retain its referenced symbol")
	}
	assertOrderedColumns(t, semanticTable(t, knownOperator), "A")

	unresolved := ParseQuery("T | project-keep A, Missing")
	binder.TryBind(unresolved)
	assertSemanticDiagnostic(t, firstNamedReference(t, unresolved, "Missing"), "KS142")
	if !firstNamedReference(t, unresolved, "A").IsBound() {
		t.Fatal("successful sibling was not retained during unresolved-list recovery")
	}
	unresolvedKeep := firstBoundNode[*ProjectKeepOperator](t, unresolved)
	assertSemanticDiagnostic(t, unresolvedKeep, "KS109")
	assertOrderedColumns(t, semanticTable(t, unresolvedKeep), "A")

	unresolvedInput := ParseQuery("Missing | project-away A*")
	binder.TryBind(unresolvedInput)
	assertSemanticDiagnostic(t, firstNamedReference(t, unresolvedInput, "Missing"), "KS204")
	if !firstBoundNode[*ProjectAwayOperator](t, unresolvedInput).IsBound() {
		t.Fatal("project-away did not bind against the recovered open input")
	}

	malformed := ParseQuery("T | project-keep A,")
	binder.TryBind(malformed)
	operator := firstBoundNode[*ProjectKeepOperator](t, malformed)
	if !diagnosticsContain(operator.ContainedDiagnostics(), "KS006") {
		t.Fatalf("malformed-list parser diagnostics were lost: %#v", operator.ContainedDiagnostics())
	}
	assertSemanticDiagnostic(t, operator, "KS109")
	assertOrderedColumns(t, semanticTable(t, operator), "A")
}

func TestNativeTreeBinderProjectAwayKeepRejectInjectedExpressionAndRestoreState(t *testing.T) {
	literal := MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindLongLiteralToken, "", "1")))
	list := AsSeparatedSyntaxListOf[Expression](NewSyntaxList(NewSeparatedElement(literal, nil)))
	operator := MustNewProjectKeepOperatorTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectKeepKeyword, "", "project-keep")), list)
	binder := NewBinder(NewGlobalState(nil))
	row := NewTableSymbol("Ambient", NewColumnSymbol("A", ScalarTypes.String.Type()))
	binder.rowScope = row
	binder.scopeKind = ScopeKindPlugIn
	marker := NewVariableSymbol("Marker", ScalarTypes.Long.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(marker)
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.bind(operator)

	if operator.IsBound() || literal.IsBound() {
		t.Fatal("injected unsupported project-keep expression was bound")
	}
	if binder.rowScope != row || binder.scopeKind != ScopeKindPlugIn || !tree.inPipe || !tree.inWildcard {
		t.Fatalf("binder state changed: row:%#v scope:%v pipe:%t wildcard:%t", binder.rowScope, binder.scopeKind, tree.inPipe, tree.inWildcard)
	}
	if localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatal("unsupported project-keep changed LocalScope")
	}
}

func TestNativeTreeBinderProjectReorderExactWildcardIdentityStateAndDownstreamScope(t *testing.T) {
	count := NewColumnSymbol("Count", ScalarTypes.Long.Type())
	a2 := NewColumnSymbol("Account2", ScalarTypes.String.Type())
	user := NewColumnSymbol("User Name", ScalarTypes.String.Type())
	a10 := NewColumnSymbol("Account10", ScalarTypes.String.Type())
	a1 := NewColumnSymbol("Account1", ScalarTypes.String.Type())
	table := NewTableSymbol("T", count, a2, user, a10, a1).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | project-reorder ['User Name'], Account* granny-asc | project Account1, ['User Name'], Count")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	reorder := firstBoundNode[*ProjectReorderOperator](t, query)
	result := semanticTable(t, reorder)
	assertOrderedColumns(t, result, "User Name", "Account1", "Account2", "Account10", "Count")
	for _, column := range []*ColumnSymbol{count, a2, user, a10, a1} {
		if result.GetColumn(column.SymbolName()) != column {
			t.Fatalf("project-reorder replaced %q identity", column.SymbolName())
		}
	}
	if result == table || result.SymbolName() != "" || !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("project-reorder state/name/identity = open:%t sorted:%t serialized:%t name:%q same:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized(), result.SymbolName(), result == table)
	}

	wildcard := firstNamedReference(t, reorder, "Account*")
	group, ok := wildcard.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(group.Members()) != 3 || group.Members()[0] != a2 || group.Members()[1] != a10 || group.Members()[2] != a1 {
		t.Fatalf("wildcard binding = %#v, want incoming-order identity group", wildcard.ReferencedSymbol())
	}
	ordered := reorder.Expressions().At(1).Element().(*OrderedExpression)
	if !ordered.IsBound() || ordered.ReferencedSymbol() != wildcard.ReferencedSymbol() {
		t.Fatal("ordered expression did not publish its staged wildcard binding")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Account1", "User Name", "Count")
}

func TestNativeTreeBinderProjectReorderOrderingModes(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Item10", ScalarTypes.String.Type()),
		NewColumnSymbol("item2", ScalarTypes.String.Type()),
		NewColumnSymbol("Item1", ScalarTypes.String.Type()),
	)
	tests := []struct {
		order string
		want  []string
	}{
		{order: "", want: []string{"Item10", "item2", "Item1"}},
		{order: "asc", want: []string{"Item1", "Item10", "item2"}},
		{order: "desc", want: []string{"item2", "Item10", "Item1"}},
		{order: "granny-asc", want: []string{"Item1", "item2", "Item10"}},
		{order: "granny-desc", want: []string{"Item10", "item2", "Item1"}},
	}
	for _, test := range tests {
		t.Run(test.order, func(t *testing.T) {
			query := ParseQuery("T | project-reorder * " + test.order)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectReorderOperator](t, query)), test.want...)
		})
	}
}

func TestNativeTreeBinderProjectReorderDuplicatesOverlapAndAppendOrder(t *testing.T) {
	a1 := NewColumnSymbol("A1", ScalarTypes.String.Type())
	b := NewColumnSymbol("B", ScalarTypes.String.Type())
	a2 := NewColumnSymbol("A2", ScalarTypes.String.Type())
	c := NewColumnSymbol("C", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a1, b, a2, c)

	duplicate := ParseQuery("T | project-reorder A2, A2, A1")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(duplicate)
	operator := firstBoundNode[*ProjectReorderOperator](t, duplicate)
	assertSemanticDiagnostic(t, operator, "KS171")
	assertOrderedColumns(t, semanticTable(t, operator), "A2", "A1", "B", "C")

	for _, text := range []string{"T | project-reorder A1, A*", "T | project-reorder A*, A1"} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			reorder := firstBoundNode[*ProjectReorderOperator](t, query)
			if diagnosticsContain(reorder.SemanticDiagnostics(), "KS171") {
				t.Fatal("wildcard overlap produced a duplicate diagnostic")
			}
			assertOrderedColumns(t, semanticTable(t, reorder), "A1", "A2", "B", "C")
		})
	}
}

func TestNativeTreeBinderProjectReorderEmptyStandaloneOpenWildcardAndTransactions(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	empty := ParseQuery("T | project-reorder")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(empty)
	result := semanticTable(t, firstBoundNode[*ProjectReorderOperator](t, empty))
	assertOrderedColumns(t, result, "A")
	if result.GetColumn("A") != a || result == table || !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatal("empty project-reorder did not preserve input identity and inheritable state in a fresh table")
	}

	standalone := ParseQuery("project-reorder")
	NewBinder(NewGlobalState(nil)).TryBind(standalone)
	assertSemanticDiagnostic(t, firstBoundNode[*ProjectReorderOperator](t, standalone), "KS174")

	openZero := ParseQuery("T | project-reorder Missing*")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(openZero)
	zeroReference := firstNamedReference(t, openZero, "Missing*")
	zeroGroup, ok := zeroReference.ReferencedSymbol().(*GroupSymbol)
	if !ok || len(zeroGroup.Members()) != 0 || len(zeroReference.SemanticDiagnostics()) != 0 {
		t.Fatalf("open zero wildcard = symbol:%#v diagnostics:%#v", zeroReference.ReferencedSymbol(), zeroReference.SemanticDiagnostics())
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectReorderOperator](t, openZero)), "A")

	closed := NewTableSymbol("Closed", a)
	missing := ParseQuery("Closed | project-reorder A, Missing")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", closed))).TryBind(missing)
	assertSemanticDiagnostic(t, firstNamedReference(t, missing, "Missing"), "KS142")
	if !firstNamedReference(t, missing, "A").IsBound() {
		t.Fatal("successful reorder sibling was not retained during recovery")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectReorderOperator](t, missing)), "A", "Missing")

	malformed := ParseQuery("T | project-reorder A,")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(malformed)
	if !diagnosticsContain(firstBoundNode[*ProjectReorderOperator](t, malformed).ContainedDiagnostics(), "KS006") {
		t.Fatal("malformed reorder list lost its parser diagnostic")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectReorderOperator](t, malformed)), "A")

	unresolvedInput := ParseQuery("Missing | project-reorder A*")
	NewBinder(NewGlobalState(nil)).TryBind(unresolvedInput)
	assertSemanticDiagnostic(t, firstNamedReference(t, unresolvedInput, "Missing"), "KS204")
	if !firstBoundNode[*ProjectReorderOperator](t, unresolvedInput).IsBound() {
		t.Fatal("project-reorder did not bind against the recovered open input")
	}
}

func TestNativeTreeBinderProjectReorderRejectsKnownNonColumnAndInjectedFormsAndRestoresState(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.LocalScope.AddSymbol(NewVariableSymbol("Local", ScalarTypes.Long.Type(), false, nil, nil))
	known := ParseQuery("T | project-reorder A, Local")
	binder.TryBind(known)
	if firstNamedReference(t, known, "A").IsBound() || firstNamedReference(t, known, "Local").IsBound() {
		t.Fatal("known non-column reorder operand partially committed")
	}
	assertProjectReorderAndPipeUnbound(t, known)

	literal := MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindLongLiteralToken, "", "1")))
	badList := AsSeparatedSyntaxListOf[Expression](NewSyntaxList(NewSeparatedElement(literal, nil)))
	badOperator := MustNewProjectReorderOperatorTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectReorderKeyword, "", "project-reorder")), badList)
	ambient := NewTableSymbol("Ambient", a)
	binder.rowScope = ambient
	binder.scopeKind = ScopeKindPlugIn
	markerScope := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.bind(badOperator)
	if badOperator.IsBound() || literal.IsBound() {
		t.Fatal("injected reorder expression was bound")
	}
	if binder.rowScope != ambient || binder.scopeKind != ScopeKindPlugIn || binder.LocalScope != markerScope || !tree.inPipe || !tree.inWildcard {
		t.Fatalf("reorder leaked binder state: row:%#v scope:%v local:%p pipe:%t wildcard:%t", binder.rowScope, binder.scopeKind, binder.LocalScope, tree.inPipe, tree.inWildcard)
	}

	reference := MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "A"))), SymbolMatchColumn)
	nulls := MustNewOrderingNullsClauseTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindNullsKeyword, "", "nulls")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindFirstKeyword, "", "first")),
	)
	ordering := MustNewOrderingClauseTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindAscKeyword, "", "asc")), nulls)
	ordered := MustNewOrderedExpressionTyped(reference, ordering)
	nullsList := AsSeparatedSyntaxListOf[Expression](NewSyntaxList(NewSeparatedElement(ordered, nil)))
	nullsOperator := MustNewProjectReorderOperatorTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectReorderKeyword, "", "project-reorder")), nullsList)
	tree.bind(nullsOperator)
	if nullsOperator.IsBound() || ordered.IsBound() || reference.IsBound() {
		t.Fatal("reorder accepted an injected nulls clause")
	}

	invalidReference := MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "A"))), SymbolMatchColumn)
	invalidOrdering := MustNewOrderingClauseTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindFirstKeyword, "", "first")), nil)
	invalidOrdered := MustNewOrderedExpressionTyped(invalidReference, invalidOrdering)
	invalidList := AsSeparatedSyntaxListOf[Expression](NewSyntaxList(NewSeparatedElement(invalidOrdered, nil)))
	invalidOperator := MustNewProjectReorderOperatorTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindProjectReorderKeyword, "", "project-reorder")), invalidList)
	tree.bind(invalidOperator)
	if invalidOperator.IsBound() || invalidOrdered.IsBound() || invalidReference.IsBound() {
		t.Fatal("reorder accepted an injected unsupported ordering keyword")
	}
}

func assertProjectReorderAndPipeUnbound(t *testing.T, query SyntaxElement) {
	t.Helper()
	for _, operator := range boundNodes[*ProjectReorderOperator](query) {
		if operator.IsBound() {
			t.Fatal("unsupported project-reorder operator was marked bound")
		}
	}
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if pipe.IsBound() {
			t.Fatal("pipe containing unsupported project-reorder was marked bound")
		}
	}
}

func TestNativeTreeBinderProjectByNamesLiteralPatternsIdentityStateAndDownstreamScope(t *testing.T) {
	a1 := NewColumnSymbol("A1", ScalarTypes.String.Type())
	b := NewColumnSymbol("B", ScalarTypes.Long.Type())
	a2 := NewColumnSymbol("A2", ScalarTypes.Bool.Type())
	table := NewTableSymbol("T", a1, b, a2).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery(`T | project-by-names "A*", "B", "A1" | project B, A2, A1`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*ProjectByNamesOperator](t, query)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A1", "A2", "B")
	if result == table || result.SymbolName() != "" || result.GetColumn("A1") != a1 || result.GetColumn("A2") != a2 || result.GetColumn("B") != b {
		t.Fatalf("project-by-names did not retain fresh-table identity contract: %#v", result)
	}
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("project-by-names state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, projected, "B", "A2", "A1")
	if projected.GetColumn("B") != b || projected.GetColumn("A2") != a2 || projected.GetColumn("A1") != a1 {
		t.Fatal("downstream project did not use project-by-names output scope")
	}
}

func TestNativeTreeBinderProjectByNamesUniquifiesDuplicateColumnsWithUpstreamRepeatIdentity(t *testing.T) {
	first := NewColumnSymbol("x", ScalarTypes.String.Type())
	second := NewColumnSymbol("x", ScalarTypes.Long.Type())
	y := NewColumnSymbol("y", ScalarTypes.Bool.Type())
	table := NewTableSymbol("T", first, second, y)
	query := ParseQuery(`T | project-by-names "*", "*"`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*ProjectByNamesOperator](t, query)
	assertDiagnosticCodes(t, operator)
	columns := semanticTable(t, operator).Columns()
	assertOrderedColumns(t, semanticTable(t, operator), "x", "x1", "y", "x2")
	if columns[0] != first || columns[2] != y {
		t.Fatalf("unchanged output identity = %#v", columns)
	}
	for _, index := range []int{1, 3} {
		if len(columns[index].OriginalColumns) != 1 || columns[index].OriginalColumns[0] != second {
			t.Fatalf("output[%d] original columns = %#v, want second x", index, columns[index].OriginalColumns)
		}
	}
}

func TestNativeTreeBinderProjectByNamesEmptyCaseSensitiveAndNonStringSelectors(t *testing.T) {
	upper := NewColumnSymbol("Account", ScalarTypes.String.Type())
	lower := NewColumnSymbol("accountId", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", upper, lower).WithIsSerialized(true)

	empty := ParseQuery(`T | project-by-names`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(empty)
	emptyResult := semanticTable(t, firstBoundNode[*ProjectByNamesOperator](t, empty))
	assertOrderedColumns(t, emptyResult)
	if !emptyResult.IsSerialized() {
		t.Fatal("empty project-by-names did not preserve inheritable state")
	}

	query := ParseQuery(`T | project-by-names "account*", 1, true, "Missing*"`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	operator := firstBoundNode[*ProjectByNamesOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, operator), "accountId")
	if len(operator.SemanticDiagnostics()) != 0 {
		t.Fatalf("selector misses/non-string literals produced diagnostics: %#v", operator.SemanticDiagnostics())
	}

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `project-by-names "Account"`).(*ProjectByNamesOperator)
	NewBinder(NewGlobalState(nil)).TryBind(standalone)
	assertSemanticDiagnostic(t, standalone, "KS174")
	standaloneResult := semanticTable(t, standalone)
	assertOrderedColumns(t, standaloneResult, "Column1")
	if column := standaloneResult.GetColumn("Column1"); column == nil || column.Type != ScalarTypes.String.Type() {
		t.Fatalf("standalone Column1 = %#v, want string", column)
	}
}

func TestNativeTreeBinderProjectByNamesResolvedNamesAndUnresolvedRollback(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	b1 := NewColumnSymbol("B1", ScalarTypes.Long.Type())
	b2 := NewColumnSymbol("B2", ScalarTypes.Bool.Type())
	table := NewTableSymbol("T", a, b1, b2)
	query := ParseQuery(`T | project-by-names Pick, A`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	pick := NewVariableSymbol("Pick", ScalarTypes.String.Type(), true, "B*", nil)
	binder.LocalScope.AddSymbol(pick)
	binder.TryBind(query)

	operator := firstBoundNode[*ProjectByNamesOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, operator), "B1", "B2")
	if firstNamedReference(t, query, "Pick").ReferencedSymbol() != pick || firstNamedReference(t, query, "A").ReferencedSymbol() != a {
		t.Fatal("project-by-names selector names did not bind in incoming scope")
	}

	malformed := ParseQuery(`T | project-by-names "A", Missing`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(malformed)
	assertSemanticDiagnostic(t, firstNamedReference(t, malformed, "Missing"), "KS142")
	literal := firstBoundNode[*LiteralExpression](t, malformed)
	if literal.IsBound() {
		t.Fatal("successful selector sibling was committed before unresolved rollback")
	}
	assertProjectByNamesAndPipeUnbound(t, malformed)
}

func TestNativeTreeBinderProjectByNamesDynamicAndColumnNamesOfSelectors(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	b := NewColumnSymbol("B", ScalarTypes.Long.Type())
	c := NewColumnSymbol("C", ScalarTypes.Bool.Type())
	table := NewTableSymbol("T", a, b, c)
	template := NewTableSymbol("Template", NewColumnSymbol("C", ScalarTypes.Bool.Type()), NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table, template))

	query := ParseQuery(`let Pick=dynamic(["B", 1, "A"]); T | project-by-names Pick, column_names_of(Template), pack_array("B")`)
	NewBinder(globals).TryBind(query)
	operator := firstBoundNode[*ProjectByNamesOperator](t, query)
	assertDiagnosticCodes(t, operator)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "B", "A", "C")
	if result.GetColumn("A") != a || result.GetColumn("B") != b || result.GetColumn("C") != c {
		t.Fatal("project-by-names selectors did not preserve input column identity")
	}
	if firstNamedReference(t, query, "Template").ReferencedSymbol() != template {
		t.Fatal("column_names_of argument did not bind to its source table")
	}
}

func TestNativeTreeBinderProjectByNamesTabularArgumentIgnoresRowColumnShadow(t *testing.T) {
	templateColumn := NewColumnSymbol("Template", ScalarTypes.String.Type())
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", templateColumn, a)
	template := NewTableSymbol("Template", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`T | project-by-names column_names_of(Template)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, template))).TryBind(query)

	operator := firstBoundNode[*ProjectByNamesOperator](t, query)
	assertDiagnosticCodes(t, operator)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A")
	if result.GetColumn("A") != a || firstNamedReference(t, query, "Template").ReferencedSymbol() != template {
		t.Fatal("tabular argument resolved against the incoming row instead of the database table")
	}
}

func TestNativeTreeBinderProjectByNamesSelectorBoundaries(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("A1", ScalarTypes.String.Type()),
		NewColumnSymbol("A2", ScalarTypes.Long.Type()),
		NewColumnSymbol("B", ScalarTypes.Bool.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	alias := ParseAndAnalyze(`let First=dynamic("A*"); let Second=First; T | project-by-names Second`, globals)
	if diagnostics := alias.GetDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("dynamic alias diagnostics = %#v", diagnostics)
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectByNamesOperator](t, alias.Syntax())), "A1", "A2")

	for _, text := range []string{
		`T | project-by-names pack_array("A1", "B")`,
		`let column_names_of=(x:string){dynamic(["A1"])}; T | project-by-names column_names_of("x")`,
	} {
		code := ParseAndAnalyze(text, globals)
		if diagnostics := code.GetDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", text, diagnostics)
		}
		assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectByNamesOperator](t, code.Syntax())))
	}

	scalar := ParseAndAnalyze(`T | project-by-names column_names_of(A1)`, globals)
	if !diagnosticsContain(scalar.GetDiagnostics(), "KS111") {
		t.Fatalf("scalar column_names_of diagnostics = %#v, want KS111", scalar.GetDiagnostics())
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectByNamesOperator](t, scalar.Syntax())))
}

func TestNativeTreeBinderProjectByNamesPreflightsNamesBeforeFunctionPayloads(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	template := NewTableSymbol("Template", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`T | project-by-names column_names_of(Template), Missing`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, template))).TryBind(query)

	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS142")
	call := firstBoundNode[*FunctionCallExpression](t, query)
	if call.IsBound() || call.Name().IsBound() || simpleBinderSyntaxNode(functionCallArguments(call)[0]).IsBound() {
		t.Fatal("function selector was bound before a later unresolved name failed preflight")
	}
	if _, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
		t.Fatal("failed project-by-names left a function result payload")
	}
	assertProjectByNamesAndPipeUnbound(t, query)
}

func TestNativeTreeBinderProjectByNamesMalformedAndUnsupportedAreTransactional(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	leadingComma := ParseQuery(`T | project-by-names,`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(leadingComma)
	leadingOperator := firstBoundNode[*ProjectByNamesOperator](t, leadingComma)
	assertOrderedColumns(t, semanticTable(t, leadingOperator))

	malformed := ParseQuery(`T | project-by-names "A",`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(malformed)
	operator := firstBoundNode[*ProjectByNamesOperator](t, malformed)
	assertOrderedColumns(t, semanticTable(t, operator), "A")
	if !diagnosticsContain(GetContainedDiagnostics(malformed, DiagnosticsIncludeDefault), "KS006") {
		t.Fatalf("trailing-comma diagnostics = %#v, want KS006", GetContainedDiagnostics(malformed, DiagnosticsIncludeDefault))
	}

	missingSource := ParseQuery(`Missing | project-by-names "A"`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(missingSource)
	assertSemanticDiagnostic(t, firstNamedReference(t, missingSource, "Missing"), "KS204")
	if !firstBoundNode[*ProjectByNamesOperator](t, missingSource).IsBound() {
		t.Fatal("project-by-names did not bind against the recovered open input")
	}
}

func assertProjectByNamesAndPipeUnbound(t *testing.T, query SyntaxElement) {
	t.Helper()
	for _, operator := range boundNodes[*ProjectByNamesOperator](query) {
		if operator.IsBound() {
			t.Fatal("unsupported project-by-names operator was marked bound")
		}
	}
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if pipe.IsBound() {
			t.Fatal("pipe containing unsupported project-by-names was marked bound")
		}
	}
}

func TestNativeTreeBinderDistinctProjectionIdentityStateAndDownstreamScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	b := NewColumnSymbol("B", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, b).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery(`T | distinct *, A | project A1, B, A`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	distinct := firstBoundNode[*DistinctOperator](t, query)
	result := semanticTable(t, distinct)
	assertOrderedColumns(t, result, "A", "B", "A1")
	if result.GetColumn("A") != a || result.GetColumn("B") != b || result.GetColumn("A1") == a {
		t.Fatalf("distinct column identity = %#v", result.Columns())
	}
	if !result.IsOpen() || result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("distinct state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, projected, "A1", "B", "A")
}

func TestNativeTreeBinderDistinctAliasesHintsAndParameters(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a)
	valid := ParseQuery(`T | distinct hint.shufflekey=A hint.strategy=shuffle hint.num_partitions=2 Renamed=A, Fixed=1`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(valid)
	distinct := firstBoundNode[*DistinctOperator](t, valid)
	if len(distinct.SemanticDiagnostics()) != 0 {
		t.Fatalf("valid distinct diagnostics = %#v", distinct.SemanticDiagnostics())
	}
	assertColumnTypes(t, semanticTable(t, distinct), map[string]*TypeSymbol{"Renamed": ScalarTypes.String.Type(), "Fixed": ScalarTypes.Long.Type()})
	if firstNamedReference(t, valid, "A").ReferencedSymbol() != a {
		t.Fatal("distinct hint shuffle key did not bind in row scope")
	}

	tests := []struct {
		name, query, code string
	}{
		{"unknown", `T | distinct hint.remote=auto A`, "KS179"},
		{"duplicate", `T | distinct hint.num_partitions=2 hint.num_partitions=3 A`, "KS180"},
		{"value", `T | distinct hint.strategy=sideways A`, "KS005"},
		{"integer", `T | distinct hint.num_partitions=A A`, "KS221"},
		{"column", `T | distinct hint.shufflekey=1 A`, "KS109"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			operator := firstBoundNode[*DistinctOperator](t, query)
			assertSemanticDiagnostic(t, operator, test.code)
			assertOrderedColumns(t, semanticTable(t, operator), "A")
		})
	}
}

func TestNativeTreeBinderDistinctNoOpConversionAndInferredStar(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a).WithIsOpen(true)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	inferred := binder.getOpenTableInferredColumn("Inferred", table)

	duplicate := ParseQuery(`T | distinct A, tostring(A)`)
	binder.TryBind(duplicate)
	duplicateOperator := firstBoundNode[*DistinctOperator](t, duplicate)
	assertSemanticDiagnostic(t, duplicateOperator, "KS171")
	assertOrderedColumns(t, semanticTable(t, duplicateOperator), "A")

	star := ParseQuery(`T | distinct *`)
	binder.TryBind(star)
	starOperator := firstBoundNode[*DistinctOperator](t, star)
	assertOrderedColumns(t, semanticTable(t, starOperator), "A", "Inferred")
	if semanticTable(t, starOperator).GetColumn("A") != a || semanticTable(t, starOperator).GetColumn("Inferred") != inferred {
		t.Fatalf("distinct did not preserve declared/inferred column identity: %#v", semanticTable(t, starOperator).Columns())
	}
}

func TestNativeTreeBinderDistinctStandaloneMalformedAndUnsupported(t *testing.T) {
	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `distinct A`).(*DistinctOperator)
	NewBinder(NewGlobalState(nil)).TryBind(standalone)
	assertSemanticDiagnostic(t, standalone, "KS174")
	assertSemanticDiagnostic(t, firstNamedReference(t, standalone, "A"), "KS142")

	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	recovered := ParseQuery(`T | distinct A,`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(recovered)
	operator := firstBoundNode[*DistinctOperator](t, recovered)
	if !diagnosticsContain(GetContainedDiagnostics(recovered, DiagnosticsIncludeDefault), "KS006") {
		t.Fatalf("recovered distinct diagnostics = %#v", GetContainedDiagnostics(recovered, DiagnosticsIncludeDefault))
	}
	assertOrderedColumns(t, semanticTable(t, operator), "A")

	unsupported := ParseQuery(`Missing | distinct A`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(unsupported)
	assertSemanticDiagnostic(t, firstNamedReference(t, unsupported, "Missing"), "KS204")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*DistinctOperator](t, unsupported)), "A")
}

func TestNativeTreeBinderDistinctRestoresBinderState(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`T | distinct A`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindPlugIn
	marker := NewVariableSymbol("Marker", ScalarTypes.Bool.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(marker)
	binder.TryBind(query)
	if binder.rowScope != outer || binder.scopeKind != ScopeKindPlugIn || localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatal("distinct leaked binder or local-scope state")
	}
}

func TestNativeTreeBinderAssertSchemaPreservesIdentityStateAndDownstreamScope(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	b := NewColumnSymbol("B", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, b).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery(`T | assert-schema (Wrong:bool, A:long) | project B, A`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*AssertSchemaOperator](t, query)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A", "B")
	if result == table || result.SymbolName() != "" || result.GetColumn("A") != a || result.GetColumn("B") != b {
		t.Fatalf("assert-schema result identity = %#v", result)
	}
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() || len(operator.SemanticDiagnostics()) != 0 {
		t.Fatalf("assert-schema state/diagnostics = open:%t sorted:%t serialized:%t diagnostics:%#v", result.IsOpen(), result.IsSorted(), result.IsSerialized(), operator.SemanticDiagnostics())
	}
	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, projected, "B", "A")
	if projected.GetColumn("B") != b || projected.GetColumn("A") != a {
		t.Fatal("downstream project did not use assert-schema row scope")
	}
}

func TestNativeTreeBinderAssertSchemaRootAndMalformedRecovery(t *testing.T) {
	root := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `assert-schema (A:string)`).(*AssertSchemaOperator)
	NewBinder(NewGlobalState(nil)).TryBind(root)
	assertOrderedColumns(t, semanticTable(t, root))
	if len(root.SemanticDiagnostics()) != 0 {
		t.Fatalf("root assert-schema diagnostics = %#v", root.SemanticDiagnostics())
	}

	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type())).WithIsSorted(true)
	for _, text := range []string{
		`T | assert-schema`,
		`T | assert-schema (A:)`,
		`T | assert-schema (A:string), Extra`,
		`T | assert-schema (A:string, A:long)`,
	} {
		t.Run(text, func(t *testing.T) {
			query := ParseQuery(text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
			operator := firstBoundNode[*AssertSchemaOperator](t, query)
			result := semanticTable(t, operator)
			assertOrderedColumns(t, result, "A")
			if result.GetColumn("A") != table.GetColumn("A") || !result.IsSorted() || len(operator.SemanticDiagnostics()) != 0 {
				t.Fatalf("malformed assert-schema result = %#v diagnostics:%#v", result, operator.SemanticDiagnostics())
			}
			if text != `T | assert-schema (A:string, A:long)` && len(query.ContainedDiagnostics()) == 0 {
				t.Fatalf("malformed assert-schema lost parser diagnostics: %q", text)
			}
		})
	}
}

func TestNativeTreeBinderAssertSchemaRestoresBinderState(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`T | assert-schema (A:string)`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	marker := NewVariableSymbol("Marker", ScalarTypes.Bool.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(marker)
	binder.TryBind(query)
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatal("assert-schema leaked binder state")
	}
}

func TestNativeTreeBinderGetSchemaDefaultCslIdentityAndDownstreamScope(t *testing.T) {
	inputColumn := NewColumnSymbol("Input", ScalarTypes.Bool.Type())
	input := NewTableSymbol("T", inputColumn).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", input))

	defaultQuery := ParseQuery(`T | getschema | project ColumnType, ColumnOrdinal`)
	NewBinder(globals).TryBind(defaultQuery)
	defaultOperator := firstBoundNode[*GetSchemaOperator](t, defaultQuery)
	defaultResult := semanticTable(t, defaultOperator)
	assertOrderedColumns(t, defaultResult, "ColumnName", "ColumnOrdinal", "DataType", "ColumnType")
	assertColumnTypes(t, defaultResult, map[string]*TypeSymbol{
		"ColumnName":    ScalarTypes.String.Type(),
		"ColumnOrdinal": ScalarTypes.Long.Type(),
		"DataType":      ScalarTypes.String.Type(),
		"ColumnType":    ScalarTypes.String.Type(),
	})
	if defaultResult != getSchemaResult || defaultResult.GetColumn("Input") != nil ||
		defaultResult.IsOpen() || defaultResult.IsSorted() || defaultResult.IsSerialized() {
		t.Fatalf("default getschema retained input identity/properties: %#v", defaultResult)
	}
	defaultProject := semanticTable(t, firstBoundNode[*ProjectOperator](t, defaultQuery))
	assertOrderedColumns(t, defaultProject, "ColumnType", "ColumnOrdinal")
	if defaultProject.GetColumn("ColumnType") != defaultResult.GetColumn("ColumnType") ||
		defaultProject.GetColumn("ColumnOrdinal") != defaultResult.GetColumn("ColumnOrdinal") {
		t.Fatal("downstream project did not resolve against default getschema output")
	}

	secondDefault := ParseQuery(`T | getschema`)
	NewBinder(globals).TryBind(secondDefault)
	if got := semanticTable(t, firstBoundNode[*GetSchemaOperator](t, secondDefault)); got != defaultResult {
		t.Fatalf("default schema identity changed: got %#v want %#v", got, defaultResult)
	}

	cslQuery := ParseQuery(`T | getschema kind=csl | project Schema`)
	NewBinder(globals).TryBind(cslQuery)
	cslOperator := firstBoundNode[*GetSchemaOperator](t, cslQuery)
	cslResult := semanticTable(t, cslOperator)
	assertOrderedColumns(t, cslResult, "Schema")
	assertColumnTypes(t, cslResult, map[string]*TypeSymbol{"Schema": ScalarTypes.String.Type()})
	if cslResult != getSchemaCslResult || cslResult.IsOpen() || cslResult.IsSorted() || cslResult.IsSerialized() {
		t.Fatalf("csl getschema result = %#v", cslResult)
	}
	if expression := cslOperator.KindParameter().Expression(); !simpleBinderSyntaxNode(expression).IsBound() {
		t.Fatal("getschema kind child was not bound before its operator")
	}
	cslProject := semanticTable(t, firstBoundNode[*ProjectOperator](t, cslQuery))
	if cslProject.GetColumn("Schema") != cslResult.GetColumn("Schema") {
		t.Fatal("downstream project did not resolve against csl getschema output")
	}
}

func TestNativeTreeBinderGetSchemaRecoveryAndMissingInputDiagnostics(t *testing.T) {
	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `getschema`).(*GetSchemaOperator)
	NewBinder(NewGlobalState(nil)).TryBind(standalone)
	assertOrderedColumns(t, semanticTable(t, standalone), "ColumnName", "ColumnOrdinal", "DataType", "ColumnType")
	assertDiagnosticCodes(t, standalone)

	input := NewTableSymbol("T", NewColumnSymbol("Input", ScalarTypes.Bool.Type()))
	for _, test := range []struct {
		text             string
		wantColumns      []string
		wantParserErrors bool
		wantSemantic     []string
	}{
		{text: `getschema kind=csl`, wantColumns: []string{"Schema"}},
		{text: `getschema kind=`, wantColumns: []string{"ColumnName", "ColumnOrdinal", "DataType", "ColumnType"}, wantParserErrors: true},
		{text: `getschema kind=1`, wantColumns: []string{"ColumnName", "ColumnOrdinal", "DataType", "ColumnType"}, wantSemantic: []string{"KS141"}},
		{text: `getschema kind=bogus`, wantColumns: []string{"ColumnName", "ColumnOrdinal", "DataType", "ColumnType"}},
	} {
		t.Run(test.text, func(t *testing.T) {
			query := ParseQuery(`T | ` + test.text)
			NewBinder(NewGlobalState(NewDatabaseSymbol("db", input))).TryBind(query)
			operator := firstBoundNode[*GetSchemaOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, operator), test.wantColumns...)
			assertDiagnosticCodes(t, operator, test.wantSemantic...)
			if test.wantParserErrors && len(query.ContainedDiagnostics()) == 0 {
				t.Fatal("getschema recovery lost parser diagnostics")
			}
		})
	}

	failedInput := ParseQuery(`Missing | getschema | project ColumnName`)
	NewBinder(NewGlobalState(nil)).TryBind(failedInput)
	assertSemanticDiagnostic(t, firstNamedReference(t, failedInput, "Missing"), "KS204")
	getSchema := firstBoundNode[*GetSchemaOperator](t, failedInput)
	assertOrderedColumns(t, semanticTable(t, getSchema), "ColumnName", "ColumnOrdinal", "DataType", "ColumnType")
	project := firstBoundNode[*ProjectOperator](t, failedInput)
	assertOrderedColumns(t, semanticTable(t, project), "ColumnName")
	projectedReference := project.Expressions().Elements()[0].Element().(*NameReference)
	if projectedReference.ReferencedSymbol() != semanticTable(t, getSchema).GetColumn("ColumnName") {
		t.Fatal("failed left input prevented getschema output from feeding downstream")
	}
}

func TestNativeTreeBinderGetSchemaRejectsUnsupportedInjectedFormsTransactionallyAndRestoresState(t *testing.T) {
	reference := MustNewNameReferenceTyped(
		MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "csl"))),
		SymbolMatchDefault,
	)
	parameter := MustNewNamedParameterTyped(
		MustNewNameDeclarationTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "kind")))),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindEqualToken, "", "=")),
		reference,
		CompletionHintScalar,
	)
	unsupported := MustNewGetSchemaOperatorTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindGetSchemaKeyword, "", "getschema")),
		parameter,
	)

	binder := NewBinder(NewGlobalState(nil))
	ambient := NewTableSymbol("Ambient", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	binder.rowScope = ambient
	binder.scopeKind = ScopeKindPlugIn
	markerScope := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.inPipe = false
	tree.nodeBinder.pipedAsRowScope = ambient
	tree.nodeBinder.defaultColumnNameSuffix = 37
	tree.bind(unsupported)
	if unsupported.IsBound() || parameter.IsBound() || reference.IsBound() {
		t.Fatal("unsupported injected getschema partially bound")
	}
	if binder.rowScope != ambient || binder.scopeKind != ScopeKindPlugIn || binder.LocalScope != markerScope ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.inPipe || tree.nodeBinder.pipedAsRowScope != ambient ||
		tree.nodeBinder.defaultColumnNameSuffix != 37 {
		t.Fatal("unsupported getschema changed binder state")
	}

	supported := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, `getschema kind=csl`).(*GetSchemaOperator)
	tree.bind(supported)
	if !supported.IsBound() || semanticTable(t, supported) != getSchemaCslResult {
		t.Fatal("supported injected getschema did not bind")
	}
	if binder.rowScope != ambient || binder.scopeKind != ScopeKindPlugIn || binder.LocalScope != markerScope ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.inPipe || tree.nodeBinder.pipedAsRowScope != ambient ||
		tree.nodeBinder.defaultColumnNameSuffix != 37 {
		t.Fatal("supported getschema leaked binder state")
	}
}

func TestNativeTreeBinderSummarizeCoreProjectionStateAndDownstreamScope(t *testing.T) {
	group := NewColumnSymbol("Group", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Int.Type())
	table := NewTableSymbol("T", group, value).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | summarize Total=sum(Value), Fixed=1 by Group, Inferred | project Group, Inferred, Total, Fixed")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	summarize := firstBoundNode[*SummarizeOperator](t, query)
	result := semanticTable(t, summarize)
	assertOrderedColumns(t, result, "Group", "Inferred", "Total", "Fixed")
	if result.GetColumn("Group") != group || !result.IsOpen() || !result.IsSerialized() || result.IsSorted() {
		t.Fatalf("summarize state/identity = open:%t serialized:%t sorted:%t group:%#v", result.IsOpen(), result.IsSerialized(), result.IsSorted(), result.GetColumn("Group"))
	}
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"Group": ScalarTypes.String.Type(), "Inferred": ScalarTypes.Unknown.Type(),
		"Total": ScalarTypes.Long.Type(), "Fixed": ScalarTypes.Long.Type(),
	})
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Group", "Inferred", "Total", "Fixed")
	if binder.rowScope != nil || binder.scopeKind != ScopeKindNormal {
		t.Fatalf("summarize leaked binder state: row=%#v scope=%v", binder.rowScope, binder.scopeKind)
	}
}

func TestNativeTreeBinderSummarizeCustomPercentileTuples(t *testing.T) {
	group := NewColumnSymbol("Group", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Int.Type())
	table := NewTableSymbol("T", group, value)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	tests := []struct {
		name    string
		query   string
		columns []string
	}{
		{
			name:    "direct tuple",
			query:   "T | summarize percentiles(Value, 50, 95) by Group",
			columns: []string{"Group", "percentile_Value_50", "percentile_Value_95"},
		},
		{
			name:    "single assignment renames first tuple member",
			query:   "T | summarize First=percentiles(Value, 50, 95) by Group",
			columns: []string{"Group", "First", "percentile_Value_95"},
		},
		{
			name:    "compound assignment renames tuple members",
			query:   "T | summarize (P50, P95)=percentiles(Value, 50, 95) by Group",
			columns: []string{"Group", "P50", "P95"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			operator := firstBoundNode[*SummarizeOperator](t, query)
			assertDiagnosticCodes(t, operator)
			result := semanticTable(t, operator)
			assertOrderedColumns(t, result, test.columns...)
			for _, name := range test.columns[1:] {
				if !sameType(result.GetColumn(name).Type, ScalarTypes.Long.Type()) {
					t.Fatalf("column %q type = %#v, want long", name, result.GetColumn(name).Type)
				}
			}
			call := namedFunctionCall(t, query, "percentiles")
			payload, ok := call.SyntaxNode.Payload(functionCallResultSymbolPayload)
			if _, tuple := payload.(*TupleSymbol); !ok || !tuple {
				t.Fatalf("percentiles payload = %#v, want tuple", payload)
			}
		})
	}
}

func TestNativeTreeBinderSummarizeCustomPercentileConstantAlias(t *testing.T) {
	value := NewColumnSymbol("Value", ScalarTypes.Int.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", value)))
	query := ParseQuery(`let P = 95; T | summarize percentiles(Value, P)`)
	NewBinder(globals).TryBind(query)
	operator := firstBoundNode[*SummarizeOperator](t, query)
	assertDiagnosticCodes(t, operator)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "percentile_Value_95")
	if column := result.GetColumn("percentile_Value_95"); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("aliased percentile result = %#v, want long", column)
	}
}

func TestNativeTreeBinderSummarizeCustomPercentileNestedFunctionName(t *testing.T) {
	duration := NewColumnSymbol("DurationText", ScalarTypes.String.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", duration)))
	query := ParseQuery(`T | summarize percentiles(tolong(DurationText), 50)`)
	NewBinder(globals).TryBind(query)
	operator := firstBoundNode[*SummarizeOperator](t, query)
	assertDiagnosticCodes(t, operator)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "percentile_DurationText_50")
	if column := result.GetColumn("percentile_DurationText_50"); column == nil || !sameType(column.Type, ScalarTypes.Long.Type()) {
		t.Fatalf("nested percentile result = %#v, want long", column)
	}

	wrongType := ParseQuery(`T | summarize percentiles(tolong(DurationText), tostring(DurationText))`)
	NewBinder(globals).TryBind(wrongType)
	assertSemanticDiagnostic(t, namedFunctionCall(t, wrongType, "percentiles"), "KS137")
}

func TestNativeTreeBinderSummarizeTransactionalScalarOperators(t *testing.T) {
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	account := NewColumnSymbol("Account", ScalarTypes.String.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", value, account)))

	query := ParseQuery(`T | summarize Total=sum(Value + 1), Negative=sum(-Value) by Bucket=Value % 10`)
	NewBinder(globals).TryBind(query)
	operator := firstBoundNode[*SummarizeOperator](t, query)
	assertDiagnosticCodes(t, operator)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "Bucket", "Total", "Negative")
	assertColumnTypes(t, result, map[string]*TypeSymbol{"Bucket": ScalarTypes.Long.Type(), "Total": ScalarTypes.Long.Type(), "Negative": ScalarTypes.Long.Type()})
	for _, binary := range boundNodes[*BinaryExpression](query) {
		if !binary.IsBound() {
			t.Fatal("staged summarize binary expression was not committed")
		}
	}
	if unary := firstBoundNode[*PrefixUnaryExpression](t, query); !unary.IsBound() {
		t.Fatal("staged summarize unary expression was not committed")
	}

	missing := ParseQuery(`T | summarize Total=sum(Value +)`)
	NewBinder(globals).TryBind(missing)
	if !diagnosticsContain(GetContainedDiagnostics(missing, DiagnosticsIncludeSyntactic), "KS006") {
		t.Fatalf("missing operand diagnostics = %#v", GetContainedDiagnostics(missing, DiagnosticsIncludeSyntactic))
	}

	wrongType := ParseQuery(`T | summarize Total=sum(Account + 1)`)
	NewBinder(globals).TryBind(wrongType)
	wrongTypeDiagnostics := GetContainedDiagnostics(wrongType, DiagnosticsIncludeSemantic)
	if !diagnosticsContain(wrongTypeDiagnostics, "KS106") || diagnosticsContain(wrongTypeDiagnostics, "KS138") {
		t.Fatalf("wrong-type arithmetic diagnostics = %#v, want KS106 without KS138 cascade", wrongTypeDiagnostics)
	}
	assertColumnTypes(t, semanticTable(t, firstBoundNode[*SummarizeOperator](t, wrongType)), map[string]*TypeSymbol{"Total": ScalarTypes.Long.Type()})

	transactional := ParseQuery(`T | summarize Total=sum(Value + 1), missing_aggregate(Value)`)
	NewBinder(globals).TryBind(transactional)
	for _, binary := range boundNodes[*BinaryExpression](transactional) {
		if !binary.IsBound() {
			t.Fatal("recoverable semantic failure discarded a staged binary expression")
		}
	}
}

func TestNativeTreeBinderSummarizeTransactionalPathAndElementExpressions(t *testing.T) {
	user := NewColumnSymbol("user", ScalarTypes.String.Type())
	payload := NewColumnSymbol("Payload", NewDynamicBagSymbol(user).Type())
	items := NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type())
	globals := NewGlobalState(NewDatabaseSymbol("db", NewTableSymbol("T", payload, items)))

	query := ParseQuery(`T | summarize User=take_any(Payload.user), Bracketed=take_any(Payload["user"]), First=take_any(Items[0])`)
	NewBinder(globals).TryBind(query)
	operator := firstBoundNode[*SummarizeOperator](t, query)
	assertDiagnosticCodes(t, operator)
	assertColumnTypes(t, semanticTable(t, operator), map[string]*TypeSymbol{
		"User":      ScalarTypes.DynamicString.Type(),
		"Bracketed": ScalarTypes.DynamicString.Type(),
		"First":     ScalarTypes.DynamicLong.Type(),
	})
	for _, path := range boundNodes[*PathExpression](query) {
		if !path.IsBound() || path.Selector() == nil || !expressionIsBound(path.Selector()) {
			t.Fatal("staged summarize path expression was not committed")
		}
	}
	for _, element := range boundNodes[*ElementExpression](query) {
		if !element.IsBound() || element.Selector() == nil || !expressionIsBound(element.Selector()) {
			t.Fatal("staged summarize element expression was not committed")
		}
	}

	wrongIndex := ParseQuery(`T | summarize First=take_any(Items["bad"])`)
	NewBinder(globals).TryBind(wrongIndex)
	wrongDiagnostics := GetContainedDiagnostics(wrongIndex, DiagnosticsIncludeSemantic)
	if !diagnosticsContain(wrongDiagnostics, "KS141") || diagnosticsContain(wrongDiagnostics, "KS127") {
		t.Fatalf("wrong-index diagnostics = %#v, want KS141 without aggregate overload cascade", wrongDiagnostics)
	}

	transactional := ParseQuery(`T | summarize User=take_any(Payload.user), missing_aggregate(Items[0])`)
	NewBinder(globals).TryBind(transactional)
	for _, path := range boundNodes[*PathExpression](transactional) {
		if !path.IsBound() || path.Selector() == nil || !expressionIsBound(path.Selector()) {
			t.Fatal("recoverable semantic failure discarded a staged path expression")
		}
	}
	for _, element := range boundNodes[*ElementExpression](transactional) {
		if !element.IsBound() || element.Selector() == nil || !expressionIsBound(element.Selector()) {
			t.Fatal("recoverable semantic failure discarded a staged element expression")
		}
	}
}

func containsBoundReferenceSymbol(references []*NameReference, symbol Symbol) bool {
	for _, reference := range references {
		if reference != nil && reference.ReferencedSymbol() == symbol {
			return true
		}
	}
	return false
}

func TestNativeTreeBinderSummarizeCustomStarAggregates(t *testing.T) {
	group := NewColumnSymbol("Group", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	label := NewColumnSymbol("Label", ScalarTypes.String.Type())
	table := NewTableSymbol("T", group, value, label)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	tests := []struct {
		name    string
		query   string
		columns []string
	}{
		{name: "take any", query: "T | summarize take_any(*) by Group", columns: []string{"Group", "Value", "Label"}},
		{name: "legacy any prefix", query: "T | summarize any(*) by Group", columns: []string{"Group", "any_Value", "any_Label"}},
		{name: "arg max", query: "T | summarize arg_max(Value, *) by Group", columns: []string{"Group", "Value", "Label"}},
		{name: "arg max aliased by column", query: "T | summarize arg_max(Value, *) by Key=Group", columns: []string{"Key", "Value", "Label"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := ParseQuery(test.query)
			NewBinder(globals).TryBind(query)
			operator := firstBoundNode[*SummarizeOperator](t, query)
			assertDiagnosticCodes(t, operator)
			result := semanticTable(t, operator)
			assertOrderedColumns(t, result, test.columns...)
			groupResult := result.GetColumn(test.columns[0])
			if test.columns[0] == "Group" {
				if groupResult != group {
					t.Fatal("summarize by column lost source identity")
				}
			} else if groupResult == nil || groupResult.Type != group.Type || groupResult.Source == nil {
				t.Fatalf("aliased summarize by column = %#v, want source-backed Group type", groupResult)
			}
			for _, name := range test.columns[1:] {
				if name == "Value" && result.GetColumn(name) != value {
					t.Fatal("custom aggregate lost direct Value column identity")
				}
				if name == "Label" && result.GetColumn(name) != label {
					t.Fatal("custom aggregate lost direct Label column identity")
				}
			}
			if len(boundNodes[*StarExpression](query)) == 0 || !boundNodes[*StarExpression](query)[0].IsBound() {
				t.Fatal("custom aggregate star argument was not bound")
			}
		})
	}
}

func TestNativeTreeBinderSummarizeCustomAggregateRecoveryAndTransaction(t *testing.T) {
	group := NewColumnSymbol("Group", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Int.Type())
	label := NewColumnSymbol("Label", ScalarTypes.String.Type())
	table := NewTableSymbol("T", group, value, label)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	missingComma := ParseQuery("T | summarize percentiles(Value, 50 95) by Group")
	NewBinder(globals).TryBind(missingComma)
	if !diagnosticsContainMessage(GetContainedDiagnostics(missingComma, DiagnosticsIncludeSyntactic), "KS005", "Expected: ,") {
		t.Fatalf("missing-comma syntax diagnostics = %#v", GetContainedDiagnostics(missingComma, DiagnosticsIncludeSyntactic))
	}
	missingCommaResult := semanticTable(t, firstBoundNode[*SummarizeOperator](t, missingComma))
	assertOrderedColumns(t, missingCommaResult, "Group", "percentile_Value_50", "percentile_Value_95")

	wrongType := ParseQuery("T | summarize percentiles(Value, Label) by Group")
	NewBinder(globals).TryBind(wrongType)
	wrongTypeOperator := firstBoundNode[*SummarizeOperator](t, wrongType)
	assertSemanticDiagnostic(t, namedFunctionCall(t, wrongType, "percentiles"), "KS137")
	if len(semanticTable(t, wrongTypeOperator).Columns()) == 0 {
		t.Fatal("custom aggregate recovery lost the summarize result")
	}

	transactional := ParseQuery("T | summarize percentiles(Value, 50, 95), missing_aggregate(Value) by Group")
	NewBinder(globals).TryBind(transactional)
	percentiles := namedFunctionCall(t, transactional, "percentiles")
	if !percentiles.IsBound() || !percentiles.Name().IsBound() {
		t.Fatal("recoverable later aggregate error discarded earlier custom aggregate semantics")
	}
	if payload, ok := percentiles.SyntaxNode.Payload(functionCallResultSymbolPayload); !ok || payload == nil {
		t.Fatalf("recoverable summarize lost custom result payload %#v", payload)
	}
	for _, operator := range boundNodes[*SummarizeOperator](transactional) {
		if !operator.IsBound() {
			t.Fatal("recoverable summarize operator was not bound")
		}
	}
}

func TestNativeTreeBinderSummarizeEmptyByOnlyAndAlias(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)

	emptyQuery := ParseQuery("T | summarize")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(emptyQuery)
	empty := semanticTable(t, firstBoundNode[*SummarizeOperator](t, emptyQuery))
	assertOrderedColumns(t, empty)
	if !empty.IsOpen() || !empty.IsSerialized() || empty.IsSorted() {
		t.Fatalf("empty summarize state = open:%t serialized:%t sorted:%t", empty.IsOpen(), empty.IsSerialized(), empty.IsSorted())
	}

	byQuery := ParseQuery("T | summarize by Renamed=A | project Renamed")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(byQuery)
	summarize := firstBoundNode[*SummarizeOperator](t, byQuery)
	assertOrderedColumns(t, semanticTable(t, summarize), "Renamed")
	renamed := semanticTable(t, summarize).GetColumn("Renamed")
	if renamed == nil || renamed.Type != a.Type || renamed.Source == nil || len(renamed.OriginalColumns) != 0 {
		t.Fatalf("summarize alias = %#v, want source-backed A projection", renamed)
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, byQuery)), "Renamed")
}

func TestNativeTreeBinderSummarizeParameterValidation(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	valid := ParseQuery("T | summarize hint.shufflekey=A hint.strategy=shuffle hint.num_partitions=2 Fixed=1 by A")
	NewBinder(globals).TryBind(valid)
	assertDiagnosticCodes(t, firstBoundNode[*SummarizeOperator](t, valid))

	tests := []struct {
		name string
		text string
		code string
	}{
		{"unknown", "T | summarize hint.remote=auto Fixed=1", "KS179"},
		{"duplicate", "T | summarize hint.num_partitions=2 hint.num_partitions=3 Fixed=1", "KS180"},
		{"value", "T | summarize hint.strategy=native Fixed=1", "KS005"},
		{"integer", "T | summarize hint.num_partitions=A Fixed=1", "KS221"},
		{"column", "T | summarize hint.shufflekey=1 Fixed=1", "KS109"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := ParseQuery(tt.text)
			NewBinder(globals).TryBind(query)
			assertSemanticDiagnostic(t, firstBoundNode[*SummarizeOperator](t, query), tt.code)
		})
	}
}

func TestNativeTreeBinderSummarizeRecoveryDiagnostics(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	query := ParseQuery("T | summarize X=count(), X=1, A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	operator := firstBoundNode[*SummarizeOperator](t, query)
	assertSemanticDiagnostic(t, operator, "KS171")
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "A"), "KS236")
	assertOrderedColumns(t, semanticTable(t, operator), "X", "A")

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "summarize count()").(*SummarizeOperator)
	NewBinder(NewGlobalState(nil)).TryBind(standalone)
	assertSemanticDiagnostic(t, standalone, "KS174")
}

func TestNativeTreeBinderSummarizeUnsupportedFunctionIsTransactional(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.Long.Type()))
	unsupported := NewComputedFunctionSymbol("UnsupportedRows", "T", TabularityTabular)
	query := ParseQuery("T | summarize UnsupportedRows()")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, unsupported))).TryBind(query)
	for _, summarize := range boundNodes[*SummarizeOperator](query) {
		if summarize.IsBound() {
			t.Fatal("unsupported summarize operator was marked bound")
		}
	}
	for _, call := range boundNodes[*FunctionCallExpression](query) {
		if call.IsBound() || call.Name().IsBound() {
			t.Fatal("unsupported summarize call was partially bound")
		}
	}
	for _, pipe := range boundNodes[*PipeExpression](query) {
		if pipe.IsBound() {
			t.Fatal("pipe containing unsupported summarize was marked bound")
		}
	}
}

func TestNativeTreeBinderTopNestedMultiClauseProjectionIdentityAndAliases(t *testing.T) {
	category := NewColumnSymbol("Category", ScalarTypes.String.Type())
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", category, value).WithIsOpen(true).WithIsSorted(false).WithIsSerialized(true)
	query := ParseQuery("T | top-nested 3 of Category with others='Other' by Total=count() desc, top-nested 2 of Renamed=Category by Sum=sum(Value) asc | project Category, Total, Renamed, Sum")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	operator := firstBoundNode[*TopNestedOperator](t, query)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "Category", "Total", "Renamed", "Sum")
	if result.GetColumn("Category") != category {
		t.Fatalf("first top-nested of column lost source identity: %#v", result.GetColumn("Category"))
	}
	renamed := result.GetColumn("Renamed")
	if renamed == nil || renamed.Type != category.Type || len(renamed.OriginalColumns) != 1 || renamed.OriginalColumns[0] != category || renamed.Source == nil {
		t.Fatalf("aliased top-nested of column = %#v", renamed)
	}
	if !result.IsOpen() || !result.IsSerialized() || !result.IsSorted() {
		t.Fatalf("top-nested state = open:%t serialized:%t sorted:%t", result.IsOpen(), result.IsSerialized(), result.IsSorted())
	}
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"Category": ScalarTypes.String.Type(), "Total": ScalarTypes.Long.Type(),
		"Renamed": ScalarTypes.String.Type(), "Sum": ScalarTypes.Long.Type(),
	})
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Category", "Total", "Renamed", "Sum")
	for _, name := range []string{"Renamed", "Total", "Sum"} {
		declaration := firstNameDeclaration(t, query, name)
		if declaration.ReferencedSymbol() != result.GetColumn(name) {
			t.Fatalf("top-nested alias %q symbol = %#v, want result column %#v", name, declaration.ReferencedSymbol(), result.GetColumn(name))
		}
	}
}

func TestNativeTreeBinderTopNestedUniqueNamesAcrossClauses(t *testing.T) {
	category := NewColumnSymbol("Category", ScalarTypes.String.Type())
	table := NewTableSymbol("T", category)
	query := ParseQuery("T | top-nested 3 of Category by count(), top-nested 2 of Category by count(), top-nested 1 of aggregated_Category=Category by count()")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	result := semanticTable(t, firstBoundNode[*TopNestedOperator](t, query))
	assertOrderedColumns(t, result, "Category", "aggregated_Category", "Category1", "aggregated_Category1", "aggregated_Category2", "aggregated_aggregated_Category2")
	if result.GetColumn("Category") != category {
		t.Fatal("unique-name allocation replaced the first source column")
	}
	second := result.GetColumn("Category1")
	if second == nil || len(second.OriginalColumns) != 1 || second.OriginalColumns[0] != category {
		t.Fatalf("uniquified source column = %#v", second)
	}
}

func TestNativeTreeBinderTopNestedDiagnosticsAndChildScope(t *testing.T) {
	value := NewColumnSymbol("Value", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", value)
	query := ParseQuery("T | top-nested 'three' of T with others=T by Value")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*TopNestedOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS134", "KS108", "KS108")
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Value"), "KS236")
	assertOrderedColumns(t, semanticTable(t, operator), "T", "aggregated_T")

	unknown := ParseQuery("T | top-nested 2 of Missing by count(MissingArg)")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(unknown)
	assertSemanticDiagnostic(t, firstNamedReference(t, unknown, "Missing"), "KS142")
	assertSemanticDiagnostic(t, firstNamedReference(t, unknown, "MissingArg"), "KS142")
}

func TestNativeTreeBinderTopNestedRecoveryStandaloneAndDownstream(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "top-nested 2 of A by count()").(*TopNestedOperator)
	NewBinder(globals).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")
	assertOrderedColumns(t, semanticTable(t, standalone), "A", "aggregated_A")

	recovered := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "top-nested of by").(*TopNestedOperator)
	NewBinder(globals).TryBind(recovered)
	if !recovered.IsBound() {
		t.Fatal("structurally representable missing top-nested expressions were not bound")
	}
	if got := len(semanticTable(t, recovered).Columns()); got != 2 {
		t.Fatalf("recovered top-nested columns = %d, want 2", got)
	}

	closed := ParseQuery("T | top-nested 1 of A by count() | project A, aggregated_A")
	NewBinder(globals).TryBind(closed)
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, closed)), "A", "aggregated_A")

	openTable := NewTableSymbol("Open").WithIsOpen(true)
	open := ParseQuery("Open | top-nested 1 of Inferred by count() | project Inferred, aggregated_Inferred")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", openTable))).TryBind(open)
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, open)), "Inferred", "aggregated_Inferred")
}

func TestNativeTreeBinderTopNestedPreflightRollbackAndStateRestoration(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a)
	unsupportedRows := NewComputedFunctionSymbol("UnsupportedRows", "T", TabularityTabular)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, unsupportedRows))
	unsupported := ParseQuery("T | top-nested 3 of A by count(), top-nested 2 of A by UnsupportedRows()")
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Bool.Type()))
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	marker := NewVariableSymbol("Marker", ScalarTypes.Bool.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(marker)
	binder.TryBind(unsupported)

	for _, operator := range boundNodes[*TopNestedOperator](unsupported) {
		if operator.IsBound() {
			t.Fatal("unsupported top-nested operator was marked bound")
		}
	}
	for _, call := range boundNodes[*FunctionCallExpression](unsupported) {
		if call.IsBound() || call.Name().IsBound() {
			t.Fatal("top-nested preflight committed a child before rejecting a later clause")
		}
	}
	for _, pipe := range boundNodes[*PipeExpression](unsupported) {
		if pipe.IsBound() {
			t.Fatal("pipe containing unsupported top-nested was marked bound")
		}
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatal("top-nested rollback leaked binder state")
	}

	missingLeft := ParseQuery("Missing | top-nested 1 of A by count()")
	NewBinder(globals).TryBind(missingLeft)
	assertSemanticDiagnostic(t, firstNamedReference(t, missingLeft, "Missing"), "KS204")
	if !firstBoundNode[*TopNestedOperator](t, missingLeft).IsBound() {
		t.Fatal("top-nested did not bind against the recovered open input")
	}

	valid := ParseQuery("T | top-nested 1 of A by count()")
	binder.TryBind(valid)
	if !firstBoundNode[*TopNestedOperator](t, valid).IsBound() {
		t.Fatal("valid top-nested was not bound")
	}
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || localVariable(t, binder.LocalScope, "Marker") != marker {
		t.Fatal("valid top-nested leaked binder state")
	}
}

func firstNameDeclaration(t *testing.T, root SyntaxElement, name string) *NameDeclaration {
	t.Helper()
	for _, expression := range boundNodes[*SimpleNamedExpression](root) {
		if NameDeclarationSimpleName(expression.Name()) == name {
			return expression.Name()
		}
	}
	t.Fatalf("name declaration %q not found", name)
	return nil
}

func TestNativeTreeBinderAsPreservesRowsStateAndFeedsDownstream(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | as Snapshot | project N, A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	asOperator := firstBoundNode[*AsOperator](t, query)
	result := semanticTable(t, asOperator)
	if result == table || result.SymbolName() != "" {
		t.Fatalf("as result = %#v, want a new unnamed table", result)
	}
	assertOrderedColumns(t, result, "A", "N")
	if result.GetColumn("A") != a || result.GetColumn("N") != n {
		t.Fatal("as result did not preserve source column identities")
	}
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("as state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}

	variable := localVariable(t, binder.LocalScope, "Snapshot")
	if variable.Type != table {
		t.Fatalf("alias type = %#v, want original incoming table", variable.Type)
	}
	if asOperator.Name().ReferencedSymbol() != variable {
		t.Fatalf("alias declaration symbol = %#v, want published variable", asOperator.Name().ReferencedSymbol())
	}

	project := firstBoundNode[*ProjectOperator](t, query)
	projected := semanticTable(t, project)
	assertOrderedColumns(t, projected, "N", "A")
	if projected.GetColumn("N") != n || projected.GetColumn("A") != a {
		t.Fatal("downstream project did not use the as result row scope")
	}
}

func TestNativeTreeBinderParameterizedAsPublishesAliasWithoutSemanticDiagnostics(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	n := NewColumnSymbol("N", ScalarTypes.Long.Type())
	table := NewTableSymbol("T", a, n).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | as hint.materialized=true Alias | project N, A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	asOperator := firstBoundNode[*AsOperator](t, query)
	if diagnostics := asOperator.SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("parameterized as semantic diagnostics = %#v, want none", diagnostics)
	}
	parameter := asOperator.Parameters().At(0)
	if parameter.Name().ReferencedSymbol() != QueryOperatorParameters.HintDotMaterialized {
		t.Fatalf("parameter symbol = %#v, want hint.materialized", parameter.Name().ReferencedSymbol())
	}

	result := semanticTable(t, asOperator)
	assertOrderedColumns(t, result, "A", "N")
	if result.GetColumn("A") != a || result.GetColumn("N") != n ||
		!result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("parameterized as did not preserve incoming rows and state: %#v", result)
	}
	variable := localVariable(t, binder.LocalScope, "Alias")
	if variable.Type != table || asOperator.Name().ReferencedSymbol() != variable {
		t.Fatalf("parameterized as alias = %#v, want variable over incoming table", variable)
	}
	projected := semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	assertOrderedColumns(t, projected, "N", "A")
	if projected.GetColumn("N") != n || projected.GetColumn("A") != a {
		t.Fatal("downstream project did not use the parameterized as row scope")
	}

	invalidQuery := ParseQuery("T | as hint.materialized=1 InvalidHint")
	invalidBinder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	invalidBinder.TryBind(invalidQuery)
	invalidAs := firstBoundNode[*AsOperator](t, invalidQuery)
	if diagnostics := invalidAs.SemanticDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("invalid parameter diagnostics leaked into as result: %#v", diagnostics)
	}
	if variable := localVariable(t, invalidBinder.LocalScope, "InvalidHint"); variable.Type != table {
		t.Fatalf("invalid-hint alias type = %#v, want incoming table", variable.Type)
	}
}

func TestNativeTreeBinderAsAliasResolvesInLaterStatement(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a)
	query := ParseQuery("T | as Snapshot; Snapshot | project A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	variable := localVariable(t, binder.LocalScope, "Snapshot")
	reference := firstNamedReference(t, query, "Snapshot")
	if reference.ReferencedSymbol() != variable {
		t.Fatalf("later alias reference = %#v, want published variable", reference.ReferencedSymbol())
	}
	project := firstBoundNode[*ProjectOperator](t, query)
	if projected := semanticTable(t, project); projected.GetColumn("A") != a {
		t.Fatal("later alias pipeline did not retain incoming column identity")
	}
}

func TestNativeTreeBinderAsBracketedAlias(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("T | as ['Scoped Rows']; ['Scoped Rows'] | project A")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	variable := localVariable(t, binder.LocalScope, "Scoped Rows")
	reference := firstNamedReference(t, query, "Scoped Rows")
	if reference.ReferencedSymbol() != variable {
		t.Fatalf("bracketed alias reference = %#v, want published variable", reference.ReferencedSymbol())
	}
}

func TestNativeTreeBinderAsMissingNamePreservesIncomingTable(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	for _, queryText := range []string{
		"T | as",
		"T | as hint.materialized=true",
	} {
		t.Run(queryText, func(t *testing.T) {
			binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
			marker := NewVariableSymbol("Existing", ScalarTypes.Long.Type(), false, nil, nil)
			binder.LocalScope.AddSymbol(marker)
			query := ParseQuery(queryText)
			binder.TryBind(query)

			asOperator := firstBoundNode[*AsOperator](t, query)
			if !asOperator.IsBound() {
				t.Fatal("missing-name as operator was not bound")
			}
			assertOrderedColumns(t, semanticTable(t, asOperator), "A")
			if binder.LocalScope.ContainsSymbol("Alias") {
				t.Fatal("missing-name as operator published an alias")
			}
			if localVariable(t, binder.LocalScope, "Existing") != marker {
				t.Fatal("missing-name as operator changed the existing local scope")
			}
			if !diagnosticsContain(asOperator.Name().ContainedDiagnostics(), "KS006") {
				t.Fatalf("missing alias parser diagnostics were lost: %#v", asOperator.Name().ContainedDiagnostics())
			}
		})
	}
}

func TestNativeTreeBinderAsUnknownTabularSourcePublishesAlias(t *testing.T) {
	query := ParseQuery("Missing | as Alias")
	binder := NewBinder(nil)
	binder.TryBind(query)
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS204")
	if !firstBoundNode[*AsOperator](t, query).IsBound() || !binder.LocalScope.ContainsSymbol("Alias") {
		t.Fatal("as did not bind or publish its alias for a recovered open input")
	}
}

func TestNativeTreeBinderAsStandaloneBindsWithoutPublishingAlias(t *testing.T) {
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "as Alias").(*AsOperator)
	binder := NewBinder(NewGlobalState(nil))
	binder.rowScope = NewTableSymbol("Ambient", NewColumnSymbol("A", ScalarTypes.String.Type()))
	binder.scopeKind = ScopeKindAggregate
	binder.TryBind(operator)

	if !operator.IsBound() {
		t.Fatal("standalone as operator was not bound")
	}
	assertOrderedColumns(t, semanticTable(t, operator), "A")
	if binder.LocalScope.ContainsSymbol("Alias") {
		t.Fatal("standalone as operator published its alias")
	}
}

func TestNativeTreeBinderRecoveredStandaloneAsPublishesEmptyTerminalTable(t *testing.T) {
	code := ParseAndAnalyze(`datatable(A:long)[1] | as as Rows`)
	assertKustoCodeDiagnosticCodes(t, code, "KS006", "KS005")
	result := terminalResultTable(code)
	if result == nil || result.IsOpen() {
		t.Fatalf("terminal result = %#v, want closed table", code.ResultType)
	}
	assertOrderedColumns(t, result)
}

func TestNativeTreeBinderAsRestoresBinderState(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery("T | as Alias")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	oldRowScope := NewTableSymbol("Outer", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	binder.rowScope = oldRowScope
	binder.scopeKind = ScopeKindPlugIn
	binder.TryBind(query)

	if binder.rowScope != oldRowScope || binder.scopeKind != ScopeKindPlugIn {
		t.Fatalf("binder state leaked after as: row=%#v scope=%v", binder.rowScope, binder.scopeKind)
	}
	variable := localVariable(t, binder.LocalScope, "Alias")
	if variable.Type != table {
		t.Fatalf("alias captured %#v, want pipe incoming table", variable.Type)
	}
}

func TestNativeTreeBinderScanTraversalScopeAliasesAndOutput(t *testing.T) {
	a := NewColumnSymbol("A", ScalarTypes.String.Type())
	table := NewTableSymbol("T", a).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery(`T | scan with_match_id=scan_id order by OrderUnknown partition by PartitionUnknown
		declare (State:long = DefaultUnknown)
		with (
			step s1: s2.A == A => State=State + 1;
			step s2: s1.State > 0 and StepUnknown == A;
		)`)
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	operator := firstBoundNode[*ScanOperator](t, query)
	result := semanticTable(t, operator)
	assertOrderedColumns(t, result, "A", "OrderUnknown", "PartitionUnknown", "StepUnknown", "State", "scan_id")
	if result.GetColumn("A") != a {
		t.Fatal("scan did not preserve its input column identity")
	}
	if !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("scan state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}

	declaration := operator.DeclareClause().Declarations().At(0).Element()
	declared, ok := declaration.NameAndType().Name().ReferencedSymbol().(*ColumnSymbol)
	if !ok || declared.SymbolName() != "State" || declared.Type != ScalarTypes.Long.Type() {
		t.Fatalf("declared state symbol = %#v", declaration.NameAndType().Name().ReferencedSymbol())
	}
	if simpleBinderSyntaxNode(declaration.DefaultValue().Value()).IsBound() {
		t.Fatal("scan visited a declaration default value")
	}

	steps := operator.Steps().Elements()
	if len(steps) != 2 {
		t.Fatalf("scan steps = %d, want 2", len(steps))
	}
	firstAlias, ok := steps[0].Name().ReferencedSymbol().(*VariableSymbol)
	if !ok {
		t.Fatalf("first step alias = %#v", steps[0].Name().ReferencedSymbol())
	}
	secondAlias, ok := steps[1].Name().ReferencedSymbol().(*VariableSymbol)
	if !ok {
		t.Fatalf("second step alias = %#v", steps[1].Name().ReferencedSymbol())
	}
	for _, alias := range []*VariableSymbol{firstAlias, secondAlias} {
		tuple, ok := alias.Type.(*TupleSymbol)
		if !ok || tuple.RelatedTable != table {
			t.Fatalf("step alias tuple = %#v", alias.Type)
		}
		columns := tuple.Columns()
		want := []string{"A", "OrderUnknown", "PartitionUnknown", "State"}
		if len(columns) != len(want) {
			t.Fatalf("step alias tuple columns = %#v, want %v", columns, want)
		}
		for i, name := range want {
			if columns[i] == nil || columns[i].SymbolName() != name {
				t.Fatalf("step alias tuple column %d = %#v, want %q", i, columns[i], name)
			}
		}
	}
	for _, path := range boundNodes[*PathExpression](operator) {
		root, ok := path.Expression().(*NameReference)
		if !ok {
			continue
		}
		switch NameReferenceSimpleName(root) {
		case "s1":
			if root.ReferencedSymbol() != firstAlias {
				t.Fatalf("first path root = %#v", root.ReferencedSymbol())
			}
		case "s2":
			if root.ReferencedSymbol() != secondAlias {
				t.Fatalf("forward second path root = %#v", root.ReferencedSymbol())
			}
		default:
			continue
		}
		if _, ok := path.Selector().(*NameReference).ReferencedSymbol().(*ColumnSymbol); !ok {
			t.Fatalf("path selector %v = %#v", path.Selector(), path.Selector().(*NameReference).ReferencedSymbol())
		}
	}
	if binder.LocalScope.ContainsSymbol("State") || binder.LocalScope.ContainsSymbol("s1") || binder.LocalScope.ContainsSymbol("s2") {
		t.Fatal("scan child declarations leaked into the outer local scope")
	}
}

func TestNativeTreeBinderScanSourceClausesPrecedeDeclarations(t *testing.T) {
	table := NewTableSymbol("T").WithIsOpen(true)
	query := ParseQuery(`T | scan order by Later partition by PartitionLater declare (Later:long) with (step s: Later > 0;)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*ScanOperator](t, query)
	assertOrderedColumns(t, semanticTable(t, operator), "Later", "PartitionLater", "Later", "match_id")
	orderReference := operator.OrderByClause().Expressions().At(0).Element().(*NameReference)
	partitionReference := operator.PartitionByClause().Expressions().At(0).Element().(*NameReference)
	conditionReference := firstNamedReference(t, operator.Steps().At(0).Condition(), "Later")
	if orderReference.ReferencedSymbol() == conditionReference.ReferencedSymbol() {
		t.Fatal("order-by incorrectly saw the later scan declaration")
	}
	if _, ok := orderReference.ReferencedSymbol().(*ColumnSymbol); !ok {
		t.Fatalf("order reference = %#v", orderReference.ReferencedSymbol())
	}
	if _, ok := partitionReference.ReferencedSymbol().(*ColumnSymbol); !ok {
		t.Fatalf("partition reference = %#v", partitionReference.ReferencedSymbol())
	}
	if conditionReference.ReferencedSymbol() != operator.DeclareClause().Declarations().At(0).Element().NameAndType().Name().ReferencedSymbol() {
		t.Fatal("step condition did not see the scan declaration child scope")
	}
}

func TestNativeTreeBinderScanEmptyStandaloneAndFailedInput(t *testing.T) {
	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "scan with ()").(*ScanOperator)
	NewBinder(NewGlobalState(nil)).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")
	assertOrderedColumns(t, semanticTable(t, standalone), "match_id")

	query := ParseQuery("Missing | scan declare (Declared:long) with ()")
	NewBinder(NewGlobalState(nil)).TryBind(query)
	assertSemanticDiagnostic(t, firstNamedReference(t, query, "Missing"), "KS204")
	operator := firstBoundNode[*ScanOperator](t, query)
	if hasSemanticDiagnostic(operator, "KS174") {
		t.Fatalf("piped scan after failed input has false KS174: %#v", operator.SemanticDiagnostics())
	}
	result := semanticTable(t, operator)
	if result.GetColumn("Declared") == nil || result.GetColumn("match_id") == nil {
		t.Fatalf("failed-input scan columns = %#v", result.Columns())
	}
}

func TestNativeTreeBinderScanParameterAndNameDiagnosticsOnly(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	query := ParseQuery(`T | scan kind=invalid with (step s: 1 => Missing="text";)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	operator := firstBoundNode[*ScanOperator](t, query)
	assertDiagnosticCodes(t, operator, "KS005")
	assertSemanticDiagnostic(t, firstNamedReference(t, operator, "Missing"), "KS142")
	conditionNode := simpleBinderSyntaxNode(operator.Steps().At(0).Condition())
	if len(conditionNode.SemanticDiagnostics()) != 0 {
		t.Fatalf("scan invented condition diagnostics: %#v", conditionNode.SemanticDiagnostics())
	}
	assignment := operator.Steps().At(0).ComputationClause().Assignments().At(0).Element()
	assignmentNode := simpleBinderSyntaxNode(assignment.Expression())
	if len(assignmentNode.SemanticDiagnostics()) != 0 {
		t.Fatalf("scan invented assignment diagnostics: %#v", assignmentNode.SemanticDiagnostics())
	}
}

func TestNativeTreeBinderScanParserRecovery(t *testing.T) {
	recovered := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "scan with (step s: ;").(*ScanOperator)
	NewBinder(NewGlobalState(nil)).TryBind(recovered)
	if !recovered.IsBound() || semanticTable(t, recovered).GetColumn("match_id") == nil {
		t.Fatal("recovered scan did not bind its stable output")
	}
}

func TestNativeTreeBinderPrintScalarsNamesMetadataAndDownstream(t *testing.T) {
	query := ParseQuery(`print true, Local, (1), Named="value" | project print_0, print_1, print_2, Named`)
	binder := NewBinder(NewGlobalState(nil))
	local := NewVariableSymbol("Local", ScalarTypes.Long.Type(), false, nil, nil)
	binder.LocalScope.AddSymbol(local)
	binder.TryBind(query)

	print := firstBoundNode[*PrintOperator](t, query)
	result := semanticTable(t, print)
	assertOrderedColumns(t, result, "print_0", "print_1", "print_2", "Named")
	assertColumnTypes(t, result, map[string]*TypeSymbol{
		"print_0": ScalarTypes.Bool.Type(),
		"print_1": ScalarTypes.Long.Type(),
		"print_2": ScalarTypes.Long.Type(),
		"Named":   ScalarTypes.String.Type(),
	})
	if result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("print state = open:%t sorted:%t serialized:%t, want closed default state", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	if firstNamedReference(t, print, "Local").ReferencedSymbol() != local {
		t.Fatal("print local name did not resolve to the local variable")
	}
	named := print.Expressions().At(3).Element().(*SimpleNamedExpression)
	column := result.GetColumn("Named")
	if named.Name().ReferencedSymbol() != column {
		t.Fatalf("named declaration symbol = %#v, want result column %#v", named.Name().ReferencedSymbol(), column)
	}
	if column.Source != named.Expression() {
		t.Fatalf("named column source = %#v, want named value expression %#v", column.Source, named.Expression())
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "print_0", "print_1", "print_2", "Named")
}

func TestNativeTreeBinderPrintDiagnosticsAndPipedBinding(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	root := ParseQuery("print X=1, X=2, T | project X, print_2")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(root)

	print := firstBoundNode[*PrintOperator](t, root)
	assertDiagnosticCodes(t, print, "KS171", "KS108")
	result := semanticTable(t, print)
	assertOrderedColumns(t, result, "X", "print_2")
	if result.GetColumn("print_2").Type.SymbolTabularity() != TabularityTabular {
		t.Fatalf("tabular print column type = %v, want tabular", result.GetColumn("print_2").Type)
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, root)), "X", "print_2")

	piped := ParseQuery("T | print A | project A")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(piped)
	pipedPrint := firstBoundNode[*PrintOperator](t, piped)
	assertDiagnosticCodes(t, pipedPrint, "KS175")
	if semanticTable(t, pipedPrint).GetColumn("A") != table.GetColumn("A") {
		t.Fatal("piped print did not retain the directly referenced column")
	}
	if semanticTable(t, firstBoundNode[*ProjectOperator](t, piped)).GetColumn("A") != table.GetColumn("A") {
		t.Fatal("downstream project did not bind against piped print output")
	}
}

func TestNativeTreeBinderPrintParserRecoveryAndUnresolvedNames(t *testing.T) {
	for _, test := range []struct {
		text  string
		names []string
	}{
		{text: "print", names: []string{"print_0"}},
		{text: "print Value=", names: []string{"Value"}},
		{text: "print , 1", names: []string{"print_0", "print_1"}},
	} {
		t.Run(test.text, func(t *testing.T) {
			query := ParseQuery(test.text)
			NewBinder(NewGlobalState(nil)).TryBind(query)
			print := firstBoundNode[*PrintOperator](t, query)
			assertOrderedColumns(t, semanticTable(t, print), test.names...)
			for _, column := range semanticTable(t, print).Columns() {
				if column == nil || (column.SymbolName() != "print_1" && !sameType(column.Type, ScalarTypes.Unknown.Type())) {
					t.Fatalf("recovered column = %#v, want stable unknown type", column)
				}
			}
			if len(query.ContainedDiagnostics()) == 0 {
				t.Fatal("parser recovery diagnostics missing")
			}
		})
	}

	unresolved := ParseQuery("print Missing | project print_0")
	NewBinder(NewGlobalState(nil)).TryBind(unresolved)
	assertSemanticDiagnostic(t, firstNamedReference(t, unresolved, "Missing"), "KS142")
	print := firstBoundNode[*PrintOperator](t, unresolved)
	assertOrderedColumns(t, semanticTable(t, print), "print_0")
	if !sameType(semanticTable(t, print).GetColumn("print_0").Type, ScalarTypes.Unknown.Type()) {
		t.Fatal("unresolved print expression did not produce an unknown column")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, unresolved)), "print_0")

	failedInput := ParseQuery("Missing | print 1 | project print_0")
	NewBinder(NewGlobalState(nil)).TryBind(failedInput)
	assertSemanticDiagnostic(t, firstNamedReference(t, failedInput, "Missing"), "KS204")
	failedPrint := firstBoundNode[*PrintOperator](t, failedInput)
	assertDiagnosticCodes(t, failedPrint, "KS175")
	assertOrderedColumns(t, semanticTable(t, failedPrint), "print_0")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, failedInput)), "print_0")
}

func localVariable(t *testing.T, scope *LocalScope, name string) *VariableSymbol {
	t.Helper()
	var symbols []Symbol
	scope.GetSymbols(name, SymbolMatchLocal, &symbols)
	if len(symbols) != 1 {
		t.Fatalf("local symbols for %q = %#v, want one", name, symbols)
	}
	variable, ok := symbols[0].(*VariableSymbol)
	if !ok {
		t.Fatalf("local symbol for %q = %T, want *VariableSymbol", name, symbols[0])
	}
	return variable
}

func boundNodes[T SyntaxElement](root SyntaxElement) []T {
	var nodes []T
	WalkSyntax(root, func(element SyntaxElement) {
		if node, ok := WrapGeneratedSyntaxElement(element).(T); ok {
			nodes = append(nodes, node)
		}
	})
	return nodes
}

func firstBoundNode[T SyntaxElement](t *testing.T, root SyntaxElement) T {
	t.Helper()
	nodes := boundNodes[T](root)
	if len(nodes) == 0 {
		var zero T
		t.Fatalf("node %T not found", zero)
	}
	return nodes[0]
}

func firstNamedReference(t *testing.T, root SyntaxElement, name string) *NameReference {
	t.Helper()
	for _, reference := range boundNodes[*NameReference](root) {
		if NameReferenceSimpleName(reference) == name {
			return reference
		}
	}
	t.Fatalf("name reference %q not found", name)
	return nil
}

func semanticTable(t *testing.T, element interface {
	GetSemanticInfo() (SemanticInfo, bool)
}) *TableSymbol {
	t.Helper()
	info, ok := element.GetSemanticInfo()
	if !ok {
		t.Fatal("semantic info missing")
	}
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("semantic symbol = %T, want *TableSymbol", info.ReferencedSymbol())
	}
	return table
}

func assertColumnTypes(t *testing.T, table *TableSymbol, want map[string]*TypeSymbol) {
	t.Helper()
	if len(table.Columns()) != len(want) {
		t.Fatalf("columns = %#v, want %d", table.Columns(), len(want))
	}
	for name, typ := range want {
		column := table.GetColumn(name)
		if column == nil || !sameType(column.Type, typ) {
			t.Fatalf("column %q = %#v, want type %s", name, column, typ.SymbolName())
		}
	}
}

func assertOrderedColumns(t *testing.T, table *TableSymbol, names ...string) {
	t.Helper()
	columns := table.Columns()
	if len(columns) != len(names) {
		t.Fatalf("columns = %#v, want %v", columns, names)
	}
	for i, name := range names {
		if columns[i] == nil || columns[i].SymbolName() != name {
			t.Fatalf("column %d = %#v, want %q", i, columns[i], name)
		}
	}
}

func assertSemanticDiagnostic(t *testing.T, element interface{ SemanticDiagnostics() []Diagnostic }, code string) {
	t.Helper()
	if !hasSemanticDiagnostic(element, code) {
		t.Fatalf("semantic diagnostics = %#v, want %s", element.SemanticDiagnostics(), code)
	}
}

func assertDiagnosticCodes(t *testing.T, element interface{ SemanticDiagnostics() []Diagnostic }, want ...string) {
	t.Helper()
	diagnostics := element.SemanticDiagnostics()
	got := make([]string, len(diagnostics))
	for i, diagnostic := range diagnostics {
		got[i] = diagnostic.Code
	}
	if len(got) != len(want) {
		t.Fatalf("semantic diagnostic codes = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("semantic diagnostic codes = %v, want %v", got, want)
		}
	}
}

func hasSemanticDiagnostic(element interface{ SemanticDiagnostics() []Diagnostic }, code string) bool {
	for _, diagnostic := range element.SemanticDiagnostics() {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestNativeTreeBinderMvExpandColumnsAliasesTypesParametersAndDownstream(t *testing.T) {
	items := NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfLong.Type())
	bag := NewColumnSymbol("Bag", ScalarTypes.DynamicBag.Type())
	any := NewColumnSymbol("Any", ScalarTypes.Dynamic.Type())
	keep := NewColumnSymbol("Keep", ScalarTypes.String.Type())
	table := NewTableSymbol("T", items, bag, any, keep).WithIsOpen(true).WithIsSorted(true).WithIsSerialized(true)
	query := ParseQuery("T | mv-expand kind=array with_itemindex=Idx ['Items'], BagAlias = (Bag), Any to typeof(string) limit Limit | project Items, BagAlias, Any, Idx, Keep")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.LocalScope.AddSymbol(NewParameterSymbol("Limit", ScalarTypes.Long.Type(), ""))
	binder.TryBind(query)

	mvExpand := firstBoundNode[*MvExpandOperator](t, query)
	assertDiagnosticCodes(t, mvExpand)
	result := semanticTable(t, mvExpand)
	assertOrderedColumns(t, result, "Items", "Bag", "Any", "Keep", "BagAlias", "Idx")
	if result.GetColumn("Items").Type != ScalarTypes.DynamicLong.Type() ||
		result.GetColumn("BagAlias").Type != ScalarTypes.Dynamic.Type() ||
		result.GetColumn("Any").Type != ScalarTypes.String.Type() ||
		result.GetColumn("Idx").Type != ScalarTypes.Long.Type() {
		t.Fatalf("mv-expand types = Items:%v BagAlias:%v Any:%v Idx:%v", result.GetColumn("Items").Type,
			result.GetColumn("BagAlias").Type, result.GetColumn("Any").Type, result.GetColumn("Idx").Type)
	}
	if result.GetColumn("Keep") != keep || !result.IsOpen() || !result.IsSorted() || !result.IsSerialized() {
		t.Fatalf("mv-expand table state/identity = keep:%t open:%t sorted:%t serialized:%t",
			result.GetColumn("Keep") == keep, result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	if originals := result.GetColumn("Items").OriginalColumns; len(originals) != 1 || originals[0] != items {
		t.Fatalf("expanded Items lineage = %#v", originals)
	}
	alias := boundNodes[*SimpleNamedExpression](mvExpand)[0]
	if alias.Name().ReferencedSymbol() != result.GetColumn("BagAlias") {
		t.Fatalf("alias declaration = %#v", alias.Name().ReferencedSymbol())
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "Items", "BagAlias", "Any", "Idx", "Keep")
}

func TestNativeTreeBinderMvExpandBracketedInferredAndGenericDynamic(t *testing.T) {
	known := NewColumnSymbol("Known", ScalarTypes.Dynamic.Type())
	table := NewTableSymbol("Open Table", known).WithIsOpen(true)
	query := ParseQuery("['Open Table'] | mv-expand ['Future'], Known")
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	binder.TryBind(query)

	mvExpand := firstBoundNode[*MvExpandOperator](t, query)
	assertDiagnosticCodes(t, mvExpand)
	result := semanticTable(t, mvExpand)
	assertOrderedColumns(t, result, "Known", "Future")
	if result.GetColumn("Known").Type != ScalarTypes.Dynamic.Type() || result.GetColumn("Future").Type != ScalarTypes.Unknown.Type() {
		t.Fatalf("generic/inferred dynamic types = Known:%v Future:%v", result.GetColumn("Known").Type, result.GetColumn("Future").Type)
	}
	if !result.IsOpen() {
		t.Fatal("mv-expand dropped open-table state")
	}
}

func TestNativeTreeBinderMvExpandExactDiagnosticsAndRecovery(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Text", ScalarTypes.String.Type()),
		NewColumnSymbol("Items", ScalarTypes.DynamicArrayOfString.Type()),
		NewColumnSymbol("RealLimit", ScalarTypes.Real.Type()))
	query := ParseQuery("T | mv-expand Text limit RealLimit | count")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	mvExpand := firstBoundNode[*MvExpandOperator](t, query)
	assertDiagnosticCodes(t, mvExpand, "KS141", "KS134")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*CountOperator](t, query)), "Count")

	standalone := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, "mv-expand Items").(*MvExpandOperator)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(standalone)
	assertDiagnosticCodes(t, standalone, "KS174")

	recovered := ParseQuery("T | mv-expand Items, limit 5 | count")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(recovered)
	recoveredExpand := firstBoundNode[*MvExpandOperator](t, recovered)
	if !recoveredExpand.IsBound() || !firstBoundNode[*CountOperator](t, recovered).IsBound() {
		t.Fatal("parser-recovered mv-expand did not preserve downstream binding")
	}
	assertOrderedColumns(t, semanticTable(t, recoveredExpand), "Text", "Items", "RealLimit")
}
