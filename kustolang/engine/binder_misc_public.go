package kustolang

// Ported from the public surface of src/Kusto.Language/Binder/Binder_Misc.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// GetDeclaredAndInferredColumns gets the declared columns followed by columns
// inferred by this binder while processing an open table.
func (b *Binder) GetDeclaredAndInferredColumns(table *TableSymbol) []*ColumnSymbol {
	if table == nil {
		return []*ColumnSymbol{}
	}
	return b.declaredAndInferredColumns(table)
}

// TryGetDeclaredOrInferredColumn resolves a declared column or creates one
// stable unknown-typed inferred column when table is open.
func (b *Binder) TryGetDeclaredOrInferredColumn(table *TableSymbol, name string) (*ColumnSymbol, bool) {
	if b == nil || table == nil {
		return nil, false
	}
	if column, ok := table.TryGetColumn(name); ok {
		return column, true
	}
	if !table.IsOpen() {
		return nil, false
	}
	return b.getOpenTableInferredColumn(name, table), true
}

// ApplyDirective applies one directive to this binder and returns any context diagnostic.
func (b *Binder) ApplyDirective(directive *Directive) []Diagnostic {
	if b == nil || directive == nil {
		return []Diagnostic{}
	}
	if node := directive.SyntaxNode; node != nil {
		if b.syntaxNodeLocations == nil || b.syntaxTokenLocations == nil {
			b.captureSyntaxLocations(node)
		}
	}
	return append([]Diagnostic{}, b.applyDirective(directive)...)
}

// GetMacroExpandScope finds the entity-group element that owns a path at the
// supplied syntax location.
func GetMacroExpandScope(location *SyntaxNode) *EntityGroupElementSymbol {
	if location == nil {
		return nil
	}
	var path *PathExpression
	if location.Kind() == SyntaxKindPathExpression {
		path, _ = WrapGeneratedSyntaxElement(location).(*PathExpression)
	}
	if location.Kind() == SyntaxKindNameReference {
		reference, _ := WrapGeneratedSyntaxElement(location).(*NameReference)
		if reference != nil {
			if call, ok := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); ok {
				path, _ = WrapGeneratedSyntaxElement(call.Parent()).(*PathExpression)
			} else if parentPath, ok := WrapGeneratedSyntaxElement(reference.Parent()).(*PathExpression); ok &&
				simpleBinderSyntaxNode(parentPath.Selector()) == reference.SyntaxNode {
				path = parentPath
			}
		}
	}
	if path == nil || path.Expression() == nil {
		return nil
	}
	symbol := expressionReferencedSymbol(path.Expression())
	if variable, ok := symbol.(*VariableSymbol); ok {
		symbol = variable.Type
	}
	scope, _ := symbol.(*EntityGroupElementSymbol)
	return scope
}

// GetMacroExpandScopeForGroup preserves the compatibility overload that creates
// a scope symbol directly from an entity group.
func GetMacroExpandScopeForGroup(name string, entityGroup *EntityGroupSymbol) *EntityGroupElementSymbol {
	underlying := Symbol(Error)
	if entityGroup != nil {
		for _, member := range entityGroup.Members() {
			if isEntityGroupMemberType(member) {
				underlying = member
				break
			}
		}
	}
	return NewEntityGroupElementSymbol(name, entityGroup, underlying)
}
