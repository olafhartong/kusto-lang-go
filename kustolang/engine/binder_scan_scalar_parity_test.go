package kustolang

import (
	"strings"
	"testing"
)

func TestNativeScanGeneralScalarClausesAssignmentsAndDownstream(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Timestamp", ScalarTypes.DateTime.Type()),
		NewColumnSymbol("User", ScalarTypes.String.Type()),
	)
	query := ParseQuery(`T
| scan order by bin(Timestamp, 1h), Timestamp asc
    partition by tostring(User)
    declare (State:string)
    with (step s: isnotempty(User) => State=tostring(User);)
| project State`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	scan := firstBoundNode[*ScanOperator](t, query)
	if !scan.IsBound() {
		t.Fatal("scan with general scalar clauses was not bound")
	}
	assertDiagnosticCodes(t, scan)
	assertOrderedColumns(t, semanticTable(t, scan), "Timestamp", "User", "State", "match_id")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "State")

	ordered, ok := scan.OrderByClause().Expressions().At(1).Element().(*OrderedExpression)
	if !ok || !ordered.IsBound() || !sameType(ExpressionResultType(ordered), ScalarTypes.DateTime.Type()) {
		t.Fatalf("ordered expression = %T bound:%t type:%s", ordered, ordered != nil && ordered.IsBound(), typeName(ExpressionResultType(ordered)))
	}
	for _, name := range []string{"bin", "tostring", "isnotempty"} {
		call := namedFunctionCall(t, scan, name)
		if !call.IsBound() || call.ReferencedSignature() == nil {
			t.Fatalf("function %q = bound:%t signature:%#v", name, call.IsBound(), call.ReferencedSignature())
		}
	}
	assignment := scan.Steps().At(0).ComputationClause().Assignments().At(0).Element()
	if _, ok := assignment.Name().ReferencedSymbol().(*ColumnSymbol); !ok {
		t.Fatalf("assignment target = %#v, want declared column", assignment.Name().ReferencedSymbol())
	}
}

func TestNativeScanGeneralScalarDynamicElementPathAndUnary(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("User", ScalarTypes.String.Type()),
		NewColumnSymbol("Bag", ScalarTypes.DynamicBag.Type()),
	)
	query := ParseQuery(`T
| scan partition by Bag.users[0], Bag["users"][0]
    with (step s: -tolong("1") < 0 and isnotempty(tostring(dynamic({"users":[User]})));)`)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	scan := firstBoundNode[*ScanOperator](t, query)
	if !scan.IsBound() {
		t.Fatal("scan with dynamic, element, path, and unary expressions was not bound")
	}
	assertDiagnosticCodes(t, scan)
	for _, element := range []SyntaxElement{
		firstBoundNode[*DynamicExpression](t, scan),
		firstBoundNode[*PathExpression](t, scan),
		firstBoundNode[*ElementExpression](t, scan),
		firstBoundNode[*PrefixUnaryExpression](t, scan),
	} {
		if node := simpleBinderSyntaxNode(element); node == nil || !node.IsBound() {
			t.Fatalf("general scan scalar %T was not bound", element)
		}
	}
}

func TestNativeScanMissingNamesReportExactChildDiagnosticsAndKeepShape(t *testing.T) {
	text := `T | scan with (
    step s: isnotempty(Missing)
      => User=tostring(Missing);
  )
| project User`
	table := NewTableSymbol("T", NewColumnSymbol("User", ScalarTypes.String.Type()))
	query := ParseQuery(text)
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)

	scan := firstBoundNode[*ScanOperator](t, query)
	if !scan.IsBound() {
		t.Fatal("scan containing diagnosed child references was not bound")
	}
	first := strings.Index(text, "Missing")
	second := strings.LastIndex(text, "Missing")
	assertDiagnosticLocations(t, GetContainedDiagnostics(scan, DiagnosticsIncludeSemantic),
		diagnosticLocation{"KS142", first, len("Missing")},
		diagnosticLocation{"KS142", second, len("Missing")},
	)
	assertOrderedColumns(t, semanticTable(t, scan), "User", "match_id")
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, query)), "User")
}

func TestNativeScanRecoveredScalarsBindAndInjectedLaterCallRollsBack(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("User", ScalarTypes.String.Type())).WithIsOpen(true)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	for _, text := range []string{
		`T | scan with (step s: isnotempty(User; )`,
		`T | scan with (step s: isnotempty(User) and ;)`,
	} {
		query := ParseQuery(text)
		NewBinder(globals).TryBind(query)
		if scan := firstBoundNode[*ScanOperator](t, query); !scan.IsBound() {
			t.Fatalf("parser-recovered scan was not bound: %q", text)
		}
	}

	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator,
		`scan order by tostring(Inferred) declare (State:string) with (step first: isnotempty(User) => State=tostring(User); step later: isnotempty(User);)`).(*ScanOperator)
	calls := boundNodes[*FunctionCallExpression](operator)
	calls[len(calls)-1].ArgumentList().OpenParen().text = ""
	binder := NewBinder(globals)
	outer := NewTableSymbol("Outer", NewColumnSymbol("Marker", ScalarTypes.Bool.Type())).WithIsOpen(true)
	binder.rowScope = outer
	binder.scopeKind = ScopeKindAggregate
	local := binder.LocalScope
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.defaultColumnNameSuffix = 31
	tree.bind(operator)

	if operator.IsBound() {
		t.Fatal("scan with an injected malformed later function call was bound")
	}
	WalkSyntax(operator, func(element SyntaxElement) {
		if node := simpleBinderSyntaxNode(element); node != nil && node.IsBound() {
			t.Fatalf("failed scan retained semantic info on %T", WrapGeneratedSyntaxElement(element))
		}
	})
	if binder.rowScope != outer || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != local ||
		!tree.inPipe || !tree.inWildcard || tree.nodeBinder.defaultColumnNameSuffix != 31 {
		t.Fatal("failed scan changed binder or tree state")
	}
	if binder.openTableInferredColumns != nil || binder.openTableInferredOrder != nil {
		t.Fatal("failed scan leaked inferred open-table columns")
	}
}
