package kustolang

import (
	"strings"
	"testing"
)

func TestNativeScalarOperatorsUnaryArithmeticComparisonAndString(t *testing.T) {
	tests := []struct {
		text     string
		kind     SyntaxKind
		wantType *TypeSymbol
		constant bool
	}{
		{"- 1", SyntaxKindUnaryMinusExpression, ScalarTypes.Long.Type(), true},
		{"1 + 2.0", SyntaxKindAddExpression, ScalarTypes.Real.Type(), true},
		{"datetime(2020-01-01) - datetime(2019-01-01)", SyntaxKindSubtractExpression, ScalarTypes.TimeSpan.Type(), true},
		{"1 < 2", SyntaxKindLessThanExpression, ScalarTypes.Bool.Type(), true},
		{`"hello" has "ell"`, SyntaxKindHasExpression, ScalarTypes.Bool.Type(), true},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			expression := ParseExpression(test.text)
			NewBinder(nil).TryBind(expression)
			node := firstSyntaxNodeOfKind(expression, test.kind)
			if node == nil || !node.IsBound() {
				t.Fatalf("operator %v was not bound", test.kind)
			}
			info, _ := node.GetSemanticInfo()
			if !sameType(info.ResultType, test.wantType) || info.IsConstant != test.constant || len(info.Diagnostics) != 0 {
				t.Fatalf("operator info = type %s constant=%t diagnostics=%#v, want type %s constant=%t", info.ResultType.SymbolName(), info.IsConstant, info.Diagnostics, test.wantType.SymbolName(), test.constant)
			}
		})
	}
}

func TestNativeScalarOperatorsListsBetweenAndRowScope(t *testing.T) {
	columns := []*ColumnSymbol{
		NewColumnSymbol("N", ScalarTypes.Long.Type()),
		NewColumnSymbol("S", ScalarTypes.String.Type()),
	}
	table := NewTableSymbol("T", columns...)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	tests := []struct {
		text string
		kind SyntaxKind
	}{
		{"N in (1, 2, 3)", SyntaxKindInExpression},
		{`S has_any ("a", "b")`, SyntaxKindHasAnyExpression},
		{`S has_all ("a", "b")`, SyntaxKindHasAllExpression},
		{"N between (1 .. 10)", SyntaxKindBetweenExpression},
		{"N !between (1 .. 10)", SyntaxKindNotBetweenExpression},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			expression := ParseExpression(test.text)
			binder := NewBinder(globals)
			binder.rowScope = table
			binder.TryBind(expression)
			node := firstSyntaxNodeOfKind(expression, test.kind)
			if node == nil || !node.IsBound() || !sameType(ExpressionResultType(WrapGeneratedSyntaxNode(node).(Expression)), ScalarTypes.Bool.Type()) {
				t.Fatalf("operator %v was not bound as bool: found=%t supported=%t diagnostics=%#v", test.kind, node != nil, supportsScalarOperator(node), expression.ContainedDiagnostics())
			}
			if reference := firstBoundNode[*NameReference](t, node); reference.ReferencedSymbol() != columns[map[SyntaxKind]int{SyntaxKindHasAnyExpression: 1, SyntaxKindHasAllExpression: 1}[test.kind]] {
				t.Fatalf("row-scope reference = %#v", reference.ReferencedSymbol())
			}
			binder.TryBind(expression)
			if !node.IsBound() {
				t.Fatal("operator lost binding on repeated TryBind")
			}
		})
	}
}

func TestNativeScalarOperatorUndefinedDiagnosticLocationAndCascadeSuppression(t *testing.T) {
	text := `1 + "x"`
	query := ParseQuery(text)
	NewBinder(nil).TryBind(query)
	node := firstSyntaxNodeOfKind(query, SyntaxKindAddExpression)
	info, ok := node.GetSemanticInfo()
	if !ok {
		t.Fatal("invalid add was not bound")
	}
	assertDiagnosticLocations(t, info.Diagnostics,
		diagnosticLocation{"KS106", strings.Index(text, "+"), 1})

	unknown := ParseQuery("Missing + 1")
	NewBinder(nil).TryBind(unknown)
	unknownInfo, _ := firstSyntaxNodeOfKind(unknown, SyntaxKindAddExpression).GetSemanticInfo()
	if diagnosticsContain(unknownInfo.Diagnostics, "KS106") {
		t.Fatalf("unknown operand caused cascading KS106: %#v", unknownInfo.Diagnostics)
	}
}

func TestNativeScalarOperatorMalformedListRecoversAndDownstreamBinds(t *testing.T) {
	malformed := ParseExpression("1 in (2,)")
	NewBinder(nil).TryBind(malformed)
	in := firstSyntaxNodeOfKind(malformed, SyntaxKindInExpression)
	if in == nil {
		t.Fatal("missing recovered in expression")
	}
	if !in.IsBound() || !simpleBinderSyntaxNode(WrapGeneratedSyntaxNode(in).(*InExpression).Left()).IsBound() ||
		!sameType(ExpressionResultType(WrapGeneratedSyntaxNode(in).(*InExpression)), ScalarTypes.Bool.Type()) {
		t.Fatal("trailing-comma list did not retain recovered bool semantics")
	}
	assertDiagnosticLocations(t, GetContainedDiagnostics(malformed, DiagnosticsIncludeDefault), diagnosticLocation{"KS006", 8, 0})

	table := NewTableSymbol("T", NewColumnSymbol("N", ScalarTypes.Long.Type()))
	query := ParseQuery("T | where N > 1 | take 1")
	NewBinder(NewGlobalState(NewDatabaseSymbol("db", table))).TryBind(query)
	if !firstBoundNode[*FilterOperator](t, query).IsBound() || !firstBoundNode[*TakeOperator](t, query).IsBound() {
		t.Fatal("scalar operator prevented downstream query operators from binding")
	}
}

func firstSyntaxNodeOfKind(root SyntaxElement, kind SyntaxKind) *SyntaxNode {
	var found *SyntaxNode
	var walk func(SyntaxElement)
	walk = func(element SyntaxElement) {
		if element == nil || found != nil {
			return
		}
		if node := simpleBinderSyntaxNode(element); node != nil && node.Kind() == kind {
			found = node
			return
		}
		for i := 0; i < element.ChildCount(); i++ {
			walk(element.GetChild(i))
		}
	}
	walk(root)
	return found
}
