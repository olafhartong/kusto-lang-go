package kustolang

import "strings"

// This file ports the InlineExternalTableExpression and DateTimePattern slice
// from Binder_TreeBinder and Binder_NodeBinder at upstream cd24dcf.

func (t *treeBinder) visitInlineExternalTable(node *InlineExternalTableExpression) {
	if node == nil || node.IsBound() || !supportsInlineExternalTable(node) {
		return
	}
	tx := newUnionBindingTransaction(node)
	committed := false
	oldScope, oldRow, oldKind := t.binder.LocalScope, t.binder.rowScope, t.binder.scopeKind
	defer func() {
		t.binder.LocalScope, t.binder.rowScope, t.binder.scopeKind = oldScope, oldRow, oldKind
		if !committed {
			tx.rollback()
		}
	}()

	for _, p := range node.Parameters().Elements() {
		t.visitElement(p.Expression())
	}
	for _, c := range node.Schema().Columns().Elements() {
		t.visitElement(c.Element().Type())
	}

	// Partition expressions can reference physical schema columns only.
	t.binder.LocalScope = NewLocalScope(oldScope)
	for _, c := range node.Schema().Columns().Elements() {
		d := c.Element()
		typ := typeFromBoundExpression(d.Type())
		col := NewColumnSymbolWithSource(NameDeclarationSimpleName(d.Name()), typ, d.Name())
		d.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(col))
		t.binder.LocalScope.AddSymbol(col)
	}
	if p := node.PartitionClause(); p != nil {
		for _, item := range p.PartitionColumns().Elements() {
			d := item.Element()
			t.visitElement(d.Type())
			if d.Expr() != nil {
				t.visitElement(d.Expr())
			}
		}
	}

	// Pathformat sees partition declarations, not the physical schema.
	t.binder.LocalScope = NewLocalScope(oldScope)
	if p := node.PartitionClause(); p != nil {
		for _, item := range p.PartitionColumns().Elements() {
			d := item.Element()
			col := NewColumnSymbolWithSource(NameDeclarationSimpleName(d.Name()), typeFromBoundExpression(d.Type()), d.Name())
			d.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(col))
			t.binder.LocalScope.AddSymbol(col)
		}
	}
	if path := node.PathFormat(); path != nil {
		if path.OptionalSeparatorLiteral() != nil {
			t.visitElement(path.OptionalSeparatorLiteral())
		}
		for _, item := range path.PathExpressions().Elements() {
			r := item
			t.visitElement(r.PartitionColumnExpression())
			if r.SeparatorLiteral() != nil {
				t.visitElement(r.SeparatorLiteral())
			}
		}
	}
	t.bindNode(node.SyntaxNode)
	committed = node.IsBound()
}

func (t *treeBinder) visitDateTimePattern(node *DateTimePattern) {
	if node == nil || node.IsBound() {
		return
	}
	t.visitElement(node.StringLiteral())
	t.visitElement(node.PartitionColumn())
	t.bindNode(node.SyntaxNode)
}

func supportsInlineExternalTable(n *InlineExternalTableExpression) bool {
	if !(n != nil && n.SyntaxNode != nil && n.ChildCount() == 9 && n.InlineExternalTableKeyword() != nil &&
		n.Parameters() != nil && n.Schema() != nil && supportsDataTableRowSchema(n.Schema()) &&
		n.KindParameter() != nil && n.DataFormatParameter() != nil && n.ConnectionStrings() != nil &&
		n.ConnectionStrings().ConnectionStrings() != nil) {
		return false
	}
	list := n.ConnectionStrings().ConnectionStrings().SyntaxList
	for i := 0; i < list.Count(); i++ {
		separated, ok := list.At(i).(*SeparatedElement)
		if !ok || separated.Element() == nil {
			return false
		}
	}
	return true
}

func typeFromBoundExpression(e TypeExpression) *TypeSymbol {
	if syntax := simpleBinderSyntaxNode(e); syntax != nil {
		if typ, ok := syntax.ReferencedSymbol().(*TypeSymbol); ok {
			return typ
		}
		if syntax.ReferencedSymbol() == Error {
			return &Error.TypeSymbol
		}
	}
	if symbol := declaredSymbolFromTypeExpression(e); symbol != nil {
		if typ, ok := symbol.(*TypeSymbol); ok {
			return typ
		}
		if symbol == Error {
			return &Error.TypeSymbol
		}
	}
	return ScalarTypes.Unknown.Type()
}

func (n *simpleNodeBinder) bindDateTimePattern(node *DateTimePattern) *SemanticInfo {
	d := []Diagnostic{}
	n.checkInlineExactType(node.PartitionColumn(), ScalarTypes.DateTime.Type(), &d)
	return semanticInfoPtr(NewSemanticInfoForType(ScalarTypes.DateTime.Type(), d...))
}

func (n *simpleNodeBinder) bindInlineExternalTable(node *InlineExternalTableExpression) *SemanticInfo {
	d := queryOperatorParameterDiagnostics(n.binder, node.Parameters(), QueryOperatorParameters.InlineExternalTableProperties)
	columns := n.createExternalDataColumns(node.Schema(), &d)
	omitted := node.Schema().OpenParen().Width() == 0 && node.Schema().CloseParen().Width() == 0 && node.Schema().Columns().Count() == 0 && len(node.Schema().ContainedDiagnostics()) == 0
	if omitted {
		columns = nil
	}
	delta := strings.EqualFold(node.KindParameter().Value().ValueText(), "delta")
	if delta {
		if node.PartitionClause() != nil {
			d = append(d, n.binder.diagnosticAtSyntaxElement(DiagnosticInlineExternalTableDeltaPartitionByNotSupported(), node.PartitionClause()))
		}
		if node.PathFormat() != nil {
			d = append(d, n.binder.diagnosticAtSyntaxElement(DiagnosticInlineExternalTableDeltaPathFormatNotSupported(), node.PathFormat()))
		}
		if node.DataFormatParameter().DataFormatKeyword().Width() != 0 {
			d = append(d, n.binder.diagnosticAtSyntaxElement(DiagnosticInlineExternalTableDeltaDataFormatNotSupported(), node.DataFormatParameter()))
		}
		if node.ConnectionStrings().ConnectionStrings().Count() > 1 {
			d = append(d, n.binder.diagnosticAtSyntaxElement(DiagnosticInlineExternalTableDeltaRequiresSingleRootURI(), node.ConnectionStrings()))
		}
	} else {
		n.validateInlineStorage(node, &columns, &d)
	}
	// The official tree binder does not visit connection-string children. Keep
	// them unbound here; CheckIsExactType consequently adds no semantic error.
	if !delta && !containsFold(KustoFacts.InlineExternalTableDataFormats, node.DataFormatParameter().Value().ValueText()) {
		d = append(d, n.binder.diagnosticAtSyntaxElement(DiagnosticWrongDataStreamType(node.DataFormatParameter().Value().ValueText()), node.DataFormatParameter()))
	}
	table := NewTableSymbol("", columns...)
	if omitted {
		table = table.WithIsOpen(true)
	}
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, d...))
}

func (n *simpleNodeBinder) validateInlineStorage(node *InlineExternalTableExpression, columns *[]*ColumnSymbol, d *[]Diagnostic) {
	names, direct := map[string]bool{}, map[string]bool{}
	patterns := []*DateTimePattern{}
	if p := node.PartitionClause(); p != nil {
		for _, item := range p.PartitionColumns().Elements() {
			pc := item.Element()
			name := NameDeclarationSimpleName(pc.Name())
			if names[name] {
				*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticDuplicateColumnDeclaration(name), pc))
				break
			}
			names[name] = true
			primitive, ok := pc.Type().(*PrimitiveTypeExpression)
			if !ok {
				*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticWrongPartitionColumnType(), pc))
				break
			}
			typ := typeFromBoundExpression(primitive)
			if pc.Expr() == nil {
				for _, c := range *columns {
					if c.SymbolName() == name {
						*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticDuplicateColumnDeclaration(name), pc))
						return
					}
				}
				if sameType(typ, ScalarTypes.Long.Type()) {
					*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticWrongVirtualPartitionColumnType(), pc))
					break
				}
				*columns = append(*columns, NewColumnSymbolWithSource(name, typ, pc.Name()))
			} else {
				n.checkInlineExactType(pc.Expr(), typ, d)
				if f, ok := pc.Expr().(*FunctionCallExpression); ok && !containsFold(KustoFacts.InlineExternalTablePartitionColumnFunctions, NameReferenceSimpleName(f.Name())) {
					*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticWrongPartitionColumnFunction(), pc))
				}
			}
		}
	}
	if path := node.PathFormat(); path != nil {
		for _, item := range path.PathExpressions().Elements() {
			switch x := item.PartitionColumnExpression().(type) {
			case *NameReference:
				direct[NameReferenceSimpleName(x)] = true
			case *DateTimePattern:
				for _, old := range patterns {
					if NameReferenceSimpleName(old.PartitionColumn()) == NameReferenceSimpleName(x.PartitionColumn()) && old.StringLiteral().Token().ValueText() != x.StringLiteral().Token().ValueText() {
						*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticPartitionColumnCanNotBeUsedBothDirectlyAndPattern(NameReferenceSimpleName(x.PartitionColumn())), path))
						return
					}
				}
				patterns = append(patterns, x)
			}
		}
		for name := range names {
			used := direct[name]
			for _, p := range patterns {
				used = used || NameReferenceSimpleName(p.PartitionColumn()) == name
			}
			if !used {
				*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticPartitionColumnNotUsedInPathFormat(name), path))
			}
		}
		for _, p := range patterns {
			name := NameReferenceSimpleName(p.PartitionColumn())
			if direct[name] {
				*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticPartitionColumnCanNotBeUsedBothDirectlyAndPattern(name), path))
			}
		}
	}
}

func (n *simpleNodeBinder) checkInlineExactType(e Expression, want *TypeSymbol, d *[]Diagnostic) {
	got := ExpressionResultType(e)
	if got != nil && !isErrorTypeSymbol(got) && !sameType(got, ScalarTypes.Unknown.Type()) && !sameType(got, want) {
		*d = append(*d, n.binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustHaveType(want.SymbolName()), e))
	}
}

func containsFold(values []string, value string) bool {
	for _, v := range values {
		if strings.EqualFold(v, value) {
			return true
		}
	}
	return false
}
