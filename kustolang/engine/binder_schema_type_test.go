package kustolang

import "testing"

type unsupportedSchemaTypeExpression struct {
	*SyntaxNode
}

func (*unsupportedSchemaTypeExpression) isExpression()     {}
func (*unsupportedSchemaTypeExpression) isTypeExpression() {}

func TestNativeSchemaTypeFlatNestedEmptyStarAndIdentity(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	flat := parseQueryGrammarProduction(t, grammar.SchemaType, "(A:string, B:long, A:bool)")
	binder := NewBinder(NewGlobalState(nil))
	binder.TryBind(flat)

	flatTable := schemaTypeTable(t, flat)
	assertOrderedColumns(t, flatTable, "A", "B", "A")
	flatDeclarations := schemaTypeDeclarations(flat)
	for i, declaration := range flatDeclarations {
		if declaration.Name().ReferencedSymbol() != flatTable.Columns()[i] {
			t.Fatalf("declaration %d symbol = %#v, want column identity %#v", i, declaration.Name().ReferencedSymbol(), flatTable.Columns()[i])
		}
	}
	if flatTable.Columns()[0] == flatTable.Columns()[2] {
		t.Fatal("duplicate declarations did not receive fresh columns")
	}
	assertSchemaVoidResult(t, flat)

	binder.TryBind(flat)
	if got := schemaTypeTable(t, flat); got != flatTable {
		t.Fatal("repeated binding replaced the schema table")
	}
	for i, declaration := range flatDeclarations {
		if declaration.Name().ReferencedSymbol() != flatTable.Columns()[i] {
			t.Fatalf("repeated binding replaced declaration %d identity", i)
		}
	}

	nested := parseQueryGrammarProduction(t, grammar.SchemaType, "(Nested:(X:long, Y:dynamic), Tail:string)")
	binder.TryBind(nested)
	outerTable := schemaTypeTable(t, nested)
	nestedDeclaration := schemaTypeDeclarations(nested)[0]
	nestedSyntax := nestedDeclaration.Type().(*SchemaTypeExpression)
	nestedTable := schemaTypeTable(t, nestedSyntax)
	assertOrderedColumns(t, nestedTable, "X", "Y")
	if outerTable.Columns()[0].Type != &nestedTable.TypeSymbol {
		t.Fatal("outer nested column does not use the nested schema table identity")
	}
	if nestedDeclaration.Name().ReferencedSymbol() != outerTable.Columns()[0] {
		t.Fatal("outer nested declaration does not reference its fresh column")
	}
	for i, declaration := range schemaTypeDeclarations(nestedSyntax) {
		if declaration.Name().ReferencedSymbol() != nestedTable.Columns()[i] {
			t.Fatalf("nested declaration %d does not reference its table column", i)
		}
	}
	assertSchemaVoidResult(t, nestedSyntax)

	for _, text := range []string{"()", "(*)"} {
		schema := parseQueryGrammarProduction(t, grammar.SchemaType, text)
		binder.TryBind(schema)
		if columns := schemaTypeTable(t, schema).Columns(); len(columns) != 0 {
			t.Fatalf("schema %q columns = %#v, want empty", text, columns)
		}
		assertDiagnosticCodes(t, schema)
	}
}

func TestNativeSchemaTypeDiagnosticsRecoveryAndUnsupportedTypes(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	parented := parseQueryGrammarProduction(t, grammar.SchemaType, "(Nested:())")
	NewBinder(NewGlobalState(nil)).TryBind(parented)
	nested := schemaTypeDeclarations(parented)[0].Type().(*SchemaTypeExpression)
	assertDiagnosticCodes(t, nested, "KS220")

	invalidElement := parseQueryGrammarProduction(t, grammar.SchemaType, "(A:string)")
	literal := MustNewLiteralExpressionTyped(SyntaxKindLongLiteralExpression, SyntaxTokenFromLexical(Lex("1")[0]))
	invalidElement.SyntaxNode.children[1] = NewTypedSyntaxList(
		"SeparatedElement<Expression>", NewSeparatedElement(literal, nil))
	NewBinder(NewGlobalState(nil)).TryBind(invalidElement)
	assertDiagnosticCodes(t, invalidElement, "KS170")
	if len(schemaTypeTable(t, invalidElement).Columns()) != 0 || !literal.IsBound() {
		t.Fatal("invalid schema element did not recover as a bound, omitted declaration")
	}

	recovered := parseQueryGrammarProduction(t, grammar.SchemaType, "(Bad:bogus, Missing:, Good:long")
	NewBinder(NewGlobalState(nil)).TryBind(recovered)
	table := schemaTypeTable(t, recovered)
	assertOrderedColumns(t, table, "Bad", "Missing", "Good")
	if table.Columns()[0].Type != &Error.TypeSymbol || table.Columns()[1].Type != &Error.TypeSymbol ||
		table.Columns()[2].Type != ScalarTypes.Long.Type() {
		t.Fatalf("recovery types = %q, %q, %q", table.Columns()[0].Type.SymbolName(), table.Columns()[1].Type.SymbolName(), table.Columns()[2].Type.SymbolName())
	}
	ks160 := 0
	for _, diagnostic := range GetContainedDiagnosticsPreserveDuplicates(recovered, DiagnosticsIncludeDefault) {
		if diagnostic.Code == "KS160" {
			ks160++
		}
	}
	if ks160 != 1 {
		t.Fatalf("KS160 count = %d, want one parser-owned diagnostic", ks160)
	}

	unsupported := parseQueryGrammarProduction(t, grammar.SchemaType, "(Value:string)")
	declaration := schemaTypeDeclarations(unsupported)[0]
	declaration.SyntaxNode.children[2] = &unsupportedSchemaTypeExpression{SyntaxNode: NewSyntaxNode(SyntaxKindNone, nil)}
	NewBinder(NewGlobalState(nil)).TryBind(unsupported)
	assertDiagnosticCodes(t, unsupported, "KS172")
	unsupportedTable := schemaTypeTable(t, unsupported)
	if unsupportedTable.Columns()[0].Type != &Error.TypeSymbol || declaration.Name().ReferencedSymbol() != unsupportedTable.Columns()[0] {
		t.Fatal("unsupported type did not recover to an identity-bound error column")
	}
}

func TestNativeSchemaTypeMalformedStructuresRollbackAndRestoreState(t *testing.T) {
	schema := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).SchemaType, "(A:string)")
	declaration := schemaTypeDeclarations(schema)[0]
	primitive := declaration.Type().(*PrimitiveTypeExpression)
	seedColumn := NewColumnSymbol("Seed", ScalarTypes.Real.Type())
	declaration.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(seedColumn))
	primitive.SetSemanticInfo(NewSemanticInfoForSymbol(ScalarTypes.Real.Type(), &Void.TypeSymbol))

	// An untyped injected list is not a generated SchemaTypeExpression column list.
	declaration = cloneSyntaxElementPreservingGeneratedWrapper(declaration, true).(*NameAndTypeDeclaration)
	primitive = declaration.Type().(*PrimitiveTypeExpression)
	schema.SyntaxNode.children[1] = NewSyntaxList(NewSeparatedElement(declaration, nil))
	binder := NewBinder(NewGlobalState(nil))
	rowScope := NewTableSymbol("Row", NewColumnSymbol("R", ScalarTypes.String.Type()))
	localScope := NewLocalScope(nil)
	binder.rowScope = rowScope
	binder.scopeKind = ScopeKindAggregate
	binder.LocalScope = localScope
	inferredTable := NewTableSymbol("Open").WithIsOpen(true)
	inferredColumn := binder.getOpenTableInferredColumn("Future", inferredTable)
	tree := newTreeBinder(binder)
	tree.inPipe = true
	tree.inWildcard = true
	tree.nodeBinder.inPipe = true
	tree.nodeBinder.schemaCannotBeEmpty = true
	tree.nodeBinder.defaultColumnNameSuffix = 37

	tree.bind(schema)
	if schema.IsBound() {
		t.Fatal("malformed injected schema was partially bound")
	}
	if declaration.Name().ReferencedSymbol() != seedColumn || primitive.ReferencedSymbol() != ScalarTypes.Real.Type() {
		t.Fatal("malformed schema changed pre-existing semantic payloads")
	}
	if binder.rowScope != rowScope || binder.scopeKind != ScopeKindAggregate || binder.LocalScope != localScope ||
		!tree.inPipe || !tree.inWildcard || !tree.nodeBinder.inPipe || !tree.nodeBinder.schemaCannotBeEmpty ||
		tree.nodeBinder.defaultColumnNameSuffix != 37 {
		t.Fatal("malformed schema did not preserve binder and tree state")
	}
	if binder.getOpenTableInferredColumn("Future", inferredTable) != inferredColumn ||
		len(binder.openTableInferredOrder[inferredTable]) != 1 {
		t.Fatal("malformed schema changed open-table inference state")
	}
}

func schemaTypeTable(t *testing.T, schema *SchemaTypeExpression) *TableSymbol {
	t.Helper()
	info, ok := schema.GetSemanticInfo()
	if !ok {
		t.Fatal("schema semantic info missing")
	}
	table, ok := info.ReferencedSymbol().(*TableSymbol)
	if !ok {
		t.Fatalf("schema referenced symbol = %T, want *TableSymbol", info.ReferencedSymbol())
	}
	return table
}

func schemaTypeDeclarations(schema *SchemaTypeExpression) []*NameAndTypeDeclaration {
	declarations := make([]*NameAndTypeDeclaration, 0)
	if schema == nil || schema.Columns() == nil {
		return declarations
	}
	for _, separated := range schema.Columns().Elements() {
		if separated != nil {
			if declaration, ok := separated.Element().(*NameAndTypeDeclaration); ok {
				declarations = append(declarations, declaration)
			}
		}
	}
	return declarations
}

func assertSchemaVoidResult(t *testing.T, schema *SchemaTypeExpression) {
	t.Helper()
	info, ok := schema.GetSemanticInfo()
	if !ok || info.ResultType != &Void.TypeSymbol {
		t.Fatalf("schema result type = %#v, want Void", info.ResultType)
	}
}
