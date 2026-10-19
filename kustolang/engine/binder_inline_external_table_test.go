package kustolang

import "testing"

func TestNativeInlineExternalTableScopesShapeAndPipe(t *testing.T) {
	q := ParseQuery(`inline_external_table(Source:string, Timestamp:datetime) kind=storage partition by (Day:datetime=startofday(Timestamp), Tenant:string) pathformat=(Tenant "/" datetime_pattern("yyyy/MM/dd", Day)) dataformat=csv ("https://one") | project Source, Tenant`)
	NewBinder(nil).TryBind(q)
	in := firstBoundNode[*InlineExternalTableExpression](t, q)
	result := semanticTable(t, in)
	assertOrderedColumns(t, result, "Source", "Timestamp", "Tenant")
	assertDiagnosticCodes(t, in)
	partition := in.PartitionClause().PartitionColumns()
	resolved, ok := partition.At(0).Element().Expr().(*FunctionCallExpression).ArgumentList().Expressions().At(0).Element().(*NameReference).ReferencedSymbol().(*ColumnSymbol)
	if !ok || resolved.SymbolName() != "Timestamp" || !sameType(resolved.Type, ScalarTypes.DateTime.Type()) {
		t.Fatal("partition expression did not resolve against schema declaration")
	}
	pathTenant := in.PathFormat().PathExpressions().At(0).PartitionColumnExpression().(*NameReference)
	pattern := in.PathFormat().PathExpressions().At(1).PartitionColumnExpression().(*DateTimePattern)
	if pathTenant.ReferencedSymbol() != partition.At(1).Element().Name().ReferencedSymbol() || pattern.PartitionColumn().ReferencedSymbol() != partition.At(0).Element().Name().ReferencedSymbol() {
		t.Fatal("pathformat did not resolve against partition declarations")
	}
	if !sameType(ExpressionResultType(pattern), ScalarTypes.DateTime.Type()) {
		t.Fatal("datetime_pattern did not bind as datetime")
	}
	assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, q)), "Source", "Tenant")
}

func TestNativeInlineExternalTableStorageDiagnostics(t *testing.T) {
	tests := []struct {
		text  string
		codes []string
	}{
		{`inline_external_table(A:long, A:string, Bad:(X:long)) kind=storage dataformat=wat (1)`, []string{"KS171", "KS170", "KS267"}},
		{`inline_external_table(A:string) kind=storage partition by (P:long=tolower(A)) pathformat=(P) dataformat=csv ("u")`, []string{"KS141", "KS259"}},
		{`inline_external_table(A:string) kind=storage partition by (P:string=tolower(A)) pathformat=(P) dataformat=csv ("u")`, []string{"KS259"}},
		{`inline_external_table(A:string) kind=storage partition by (P:long=A) pathformat=(P) dataformat=csv ("u")`, []string{"KS141"}},
		{`inline_external_table(A:string) kind=storage partition by (P:string, Q:datetime) pathformat=(P) dataformat=csv ("u")`, []string{"KS264"}},
		{`inline_external_table(A:string) kind=storage partition by (P:datetime) pathformat=(P "/" datetime_pattern("d", P)) dataformat=csv ("u")`, []string{"KS265"}},
		{`inline_external_table(A:string) kind=storage partition by (P:long) pathformat=(P) dataformat=csv ("u")`, []string{"KS266"}},
		{`inline_external_table(A:string) kind=storage partition by (P:datetime) pathformat=(datetime_pattern("d", P) "/" datetime_pattern("yyyy", P)) dataformat=csv ("u")`, []string{"KS265"}},
		{`inline_external_table(A:string) kind=storage partition by (P:string=A) pathformat=(P) dataformat=csv ("u")`, nil},
	}
	badType := parsedQueryExpression(t, `inline_external_table(A:string) kind=storage partition by (P:string) pathformat=(P) dataformat=csv ("u")`).(*InlineExternalTableExpression)
	schemaType := parsedQueryExpression(t, `datatable(A:(X:long))[]`).(*DataTableExpression).Schema().Columns().At(0).Element().Type()
	badType.PartitionClause().PartitionColumns().At(0).Element().SyntaxNode.children[2] = syntaxElementFromTyped(schemaType)
	NewBinder(nil).TryBind(badType)
	assertDiagnosticCodes(t, badType, "KS258")
	for _, tc := range tests {
		in := parsedQueryExpression(t, tc.text).(*InlineExternalTableExpression)
		NewBinder(nil).TryBind(in)
		assertDiagnosticCodes(t, in, tc.codes...)
	}
}

func TestNativeInlineExternalTableConnectionExpressionsRemainUnbound(t *testing.T) {
	for _, text := range []string{
		`inline_external_table(A:string) kind=storage dataformat=csv (42)`,
		`inline_external_table(A:string) kind=storage dataformat=csv (Missing())`,
	} {
		in := parsedQueryExpression(t, text).(*InlineExternalTableExpression)
		connection := in.ConnectionStrings().ConnectionStrings().At(0).Element()
		NewBinder(nil).TryBind(in)
		assertDiagnosticCodes(t, in)
		if simpleBinderSyntaxNode(connection).IsBound() {
			t.Fatalf("connection expression %q was bound", connection.Text(IncludeTriviaAll))
		}
		assertOrderedColumns(t, semanticTable(t, in), "A")
	}
}

func TestNativeInlineExternalTableWithExpressionsRemainUnbound(t *testing.T) {
	in := parsedQueryExpression(t, `inline_external_table(A:string) kind=storage dataformat=csv ("u") with(format=123)`).(*InlineExternalTableExpression)
	property := in.WithClause().Properties().At(0).Element().Expression()
	NewBinder(nil).TryBind(in)
	assertDiagnosticCodes(t, in)
	if simpleBinderSyntaxNode(property).IsBound() {
		t.Fatalf("with expression %q was bound", property.Text(IncludeTriviaAll))
	}
	assertOrderedColumns(t, semanticTable(t, in), "A")
}

func TestNativeInlineExternalTableDeltaAndOpenShape(t *testing.T) {
	in := parsedQueryExpression(t, `inline_external_table kind=delta partition by (P:string) pathformat=(P) dataformat=csv ("one", "two")`).(*InlineExternalTableExpression)
	NewBinder(nil).TryBind(in)
	assertDiagnosticCodes(t, in, "KS270", "KS271", "KS269", "KS272")
	if table := semanticTable(t, in); !table.IsOpen() || len(table.Columns()) != 0 {
		t.Fatalf("omitted delta schema = %#v, want open empty table", table)
	}
}

func TestNativeInlineExternalTableRollbackAndIdempotence(t *testing.T) {
	in := parsedQueryExpression(t, `inline_external_table(A:string) kind=storage dataformat=csv ("u")`).(*InlineExternalTableExpression)
	original := in.ConnectionStrings().ConnectionStrings().SyntaxList.elements[0]
	in.ConnectionStrings().ConnectionStrings().SyntaxList.elements[0] = in.ConnectionStrings().ConnectionStrings().At(0).Element()
	b := NewBinder(nil)
	oldScope := b.LocalScope
	b.TryBind(in)
	if in.IsBound() || b.LocalScope != oldScope {
		t.Fatal("malformed preflight committed binding state")
	}
	in.ConnectionStrings().ConnectionStrings().SyntaxList.elements[0] = original
	b.TryBind(in)
	first := semanticTable(t, in)
	b.TryBind(in)
	if !in.IsBound() || semanticTable(t, in) != first {
		t.Fatal("valid retry or idempotent rebind failed")
	}
}
