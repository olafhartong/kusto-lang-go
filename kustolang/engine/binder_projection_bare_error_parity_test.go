package kustolang

import "testing"

func TestNext193NativeBareErrorProjectionRecoveryControls(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("A", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	operatorNames := []string{"project", "extend", "serialize"}

	for _, operatorName := range operatorNames {
		t.Run(operatorName+"/unbound", func(t *testing.T) {
			node, _ := next193BareProjectionOperator(t, operatorName, "Missing")
			if info := next193BindProjectionOperator(globals, table, node); info != nil {
				t.Fatalf("unbound %s returned semantic info %#v", operatorName, info)
			}
		})

		t.Run(operatorName+"/tabular-error-type", func(t *testing.T) {
			node, reference := next193BareProjectionOperator(t, operatorName, "Missing")
			reference.SetSemanticInfo(NewSemanticInfoForSymbol(Error, &table.TypeSymbol))
			if info := next193BindProjectionOperator(globals, table, node); info != nil {
				t.Fatalf("tabular error-shaped %s returned semantic info %#v", operatorName, info)
			}
		})

		t.Run(operatorName+"/incomplete", func(t *testing.T) {
			node, reference := next193BareProjectionOperator(t, operatorName, "['Missing']")
			reference.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
			bracketed := reference.Name().(*BracketedName)
			bracketed.SyntaxNode.children[2] = NewMissingToken(SyntaxKindCloseBracketToken, Diagnostic{})
			if info := next193BindProjectionOperator(globals, table, node); info != nil {
				t.Fatalf("incomplete %s returned semantic info %#v", operatorName, info)
			}
		})

		t.Run(operatorName+"/injected-diagnostic", func(t *testing.T) {
			node, reference := next193BareProjectionOperator(t, operatorName, "Missing")
			reference.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
			reference.SyntaxNode.diagnostics = append(reference.SyntaxNode.diagnostics, NewDiagnostic("TEST", "injected"))
			if info := next193BindProjectionOperator(globals, table, node); info != nil {
				t.Fatalf("diagnostic-bearing %s returned semantic info %#v", operatorName, info)
			}
		})
	}

	for _, operatorName := range []string{"project", "extend"} {
		t.Run(operatorName+"/tabular-reference", func(t *testing.T) {
			node, reference := next193BareProjectionOperator(t, operatorName, "T")
			reference.SetSemanticInfo(NewSemanticInfoForSymbol(table, &table.TypeSymbol))
			info := next193BindProjectionOperator(globals, table, node)
			if info == nil {
				t.Fatalf("tabular %s was not bound", operatorName)
			}
			result, ok := info.ReferencedSymbol().(*TableSymbol)
			if !ok || result.GetColumn("T") == nil || result.GetColumn("T").Type != ScalarTypes.Unknown.Type() {
				t.Fatalf("tabular %s result = %#v, want T:unknown", operatorName, info)
			}
			if len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS108" {
				t.Fatalf("tabular %s diagnostics = %#v, want KS108", operatorName, info.Diagnostics)
			}
		})

		t.Run(operatorName+"/arbitrary-scalar", func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, operatorName+" 1")
			literal := firstBoundNode[*LiteralExpression](t, operator)
			literal.SetSemanticInfo(NewSemanticInfoForType(ScalarTypes.Long.Type()))
			info := next193BindProjectionOperator(globals, table, simpleBinderSyntaxNode(operator))
			if info == nil {
				t.Fatalf("unnamed scalar %s was not bound", operatorName)
			}
			result, ok := info.ReferencedSymbol().(*TableSymbol)
			if !ok || result.GetColumn("Column1") == nil || result.GetColumn("Column1").Type != ScalarTypes.Long.Type() {
				t.Fatalf("unnamed scalar %s result = %#v, want Column1:long", operatorName, info)
			}
		})
	}
}

func next193BareProjectionOperator(t *testing.T, operatorName, expression string) (*SyntaxNode, *NameReference) {
	t.Helper()
	operator := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).QueryOperator, operatorName+" "+expression)
	references := boundNodes[*NameReference](operator)
	if len(references) != 1 {
		t.Fatalf("%s reference count = %d, want one", operatorName, len(references))
	}
	reference := references[0]
	return simpleBinderSyntaxNode(operator), reference
}

func next193BindProjectionOperator(globals *GlobalState, rowScope *TableSymbol, node *SyntaxNode) *SemanticInfo {
	binder := NewBinder(globals)
	binder.rowScope = rowScope
	nodeBinder := newSimpleNodeBinder(binder)
	nodeBinder.inPipe = true
	return nodeBinder.bind(node)
}
