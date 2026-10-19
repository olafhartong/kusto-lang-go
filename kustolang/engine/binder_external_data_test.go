package kustolang

import "testing"

func TestNativeTreeBinderExternalDataSchemaURIsAndDownstream(t *testing.T) {
	query := ParseQuery(`externaldata (A:long, A:string, a:bool, Bad:(X:long)) ["one", 1, true] | project a, A`)
	NewBinder(nil).TryBind(query)

	external := firstBoundNode[*ExternalDataExpression](t, query)
	result := semanticTable(t, external)
	assertOrderedColumns(t, result, "A", "a")
	assertColumnTypes(t, result, map[string]*TypeSymbol{"A": ScalarTypes.Long.Type(), "a": ScalarTypes.Bool.Type()})
	if result.IsOpen() || result.IsSorted() || result.IsSerialized() {
		t.Fatalf("externaldata result state = open:%t sorted:%t serialized:%t", result.IsOpen(), result.IsSorted(), result.IsSerialized())
	}
	assertDiagnosticCodes(t, external, "KS171", "KS170", "KS141", "KS141")
	declarations := external.Schema().Columns()
	if declarations.At(0).Element().Name().ReferencedSymbol() != result.Columns()[0] ||
		declarations.At(2).Element().Name().ReferencedSymbol() != result.Columns()[1] {
		t.Fatal("successful declarations do not reference their result columns")
	}
	if declarations.At(1).Element().Name().IsBound() || declarations.At(3).Element().Name().IsBound() {
		t.Fatal("omitted declarations unexpectedly received symbols")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "a", "A")
}

func TestNativeTreeBinderExternalDataURIExactStringAndCurrentRowScope(t *testing.T) {
	external := parsedQueryExpression(t, `externaldata (A:string) [Uri, Missing, 1]`).(*ExternalDataExpression)
	binder := NewBinder(nil)
	row := NewTableSymbol("Input", NewColumnSymbol("Uri", ScalarTypes.String.Type()))
	binder.rowScope = row
	binder.TryBind(external)

	if external.URIs().At(0).Element().(*NameReference).ReferencedSymbol() != row.Columns()[0] {
		t.Fatal("URI child did not bind in the upstream current row scope")
	}
	// Missing resolves to error and must not cascade; only the independently bad long URI reports KS141.
	assertDiagnosticCodes(t, external, "KS141")
}

func TestNativeTreeBinderExternalDataSkipsParameterContainerValidation(t *testing.T) {
	external := parsedQueryExpression(t, `externaldata hint.distribution=1 (A:string) ["one"] with (format=42, mode=native)`).(*ExternalDataExpression)
	NewBinder(nil).TryBind(external)
	if !external.IsBound() {
		t.Fatal("externaldata with parameters was not bound")
	}
	if external.Parameters().At(0).IsBound() || external.WithClause().IsBound() ||
		external.WithClause().Properties().At(0).Element().IsBound() {
		t.Fatal("parameter/property containers received parent semantic validation")
	}
	if !simpleBinderSyntaxNode(external.Parameters().At(0).Expression()).IsBound() ||
		!simpleBinderSyntaxNode(external.WithClause().Properties().At(0).Element().Expression()).IsBound() {
		t.Fatal("parameter/property child expressions were not bound before the parent")
	}
}

func TestNativeTreeBinderExternalDataFreshIdempotentAndFailedThenValid(t *testing.T) {
	first := parsedQueryExpression(t, `externaldata (A:string) ["one"]`).(*ExternalDataExpression)
	second := parsedQueryExpression(t, `externaldata (A:string) ["one"]`).(*ExternalDataExpression)
	binder := NewBinder(nil)
	binder.TryBind(first)
	firstTable := semanticTable(t, first)
	binder.TryBind(first)
	if semanticTable(t, first) != firstTable {
		t.Fatal("repeated binding replaced the committed table")
	}
	binder.TryBind(second)
	if semanticTable(t, second) == firstTable {
		t.Fatal("separate expressions shared a result table")
	}

	retry := parsedQueryExpression(t, `externaldata (A:string) ["one"]`).(*ExternalDataExpression)
	original := retry.URIs().SyntaxList.elements[0]
	retry.URIs().SyntaxList.elements[0] = retry.URIs().At(0).Element()
	binder.TryBind(retry)
	if retry.IsBound() {
		t.Fatal("structurally invalid expression was bound")
	}
	retry.URIs().SyntaxList.elements[0] = original
	binder.TryBind(retry)
	if !retry.IsBound() || semanticTable(t, retry).GetColumn("A") == nil {
		t.Fatal("valid retry after failed preflight did not bind")
	}
}

func TestNativeTreeBinderExternalDataTransactionalRollback(t *testing.T) {
	external := parsedQueryExpression(t, `externaldata (A:string) [Uri]`).(*ExternalDataExpression)
	// The outer preflight passes, but its bounded child fails its own preflight.
	nested := cloneSyntaxElementPreservingGeneratedWrapper(
		parsedQueryExpression(t, `externaldata (B:string) ["nested"]`), true,
	).(*ExternalDataExpression)
	nested.SyntaxNode.children[0] = nil
	external.URIs().SyntaxList.elements[0] = NewSeparatedElement(nested, nil)
	binder := NewBinder(nil)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Uri", ScalarTypes.String.Type())).WithIsOpen(true)
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.inPipe = true
	tree.nodeBinder.defaultColumnNameSuffix = 23
	tree.bind(external)
	if external.IsBound() || binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate ||
		binder.LocalScope != local || !tree.inPipe || !tree.inWildcard || !tree.nodeBinder.inPipe ||
		tree.nodeBinder.defaultColumnNameSuffix != 23 {
		t.Fatal("failed binding committed semantics or changed binder state")
	}
}
