package kustolang

import "testing"

func TestQueryGrammarPublicInlineExternalTableStorageAST(t *testing.T) {
	inline := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Account:string, Score:long) kind=storage dataformat=csv ("https://one", strcat("https://", Host), BaseUri + "/tail",) with (format="csv", compressed=true)`)

	if inline.Kind() != SyntaxKindInlineExternalTableExpression ||
		inline.InlineExternalTableKeyword().Kind() != SyntaxKindInlineExternalTableKeyword ||
		inline.InlineExternalTableKeyword().Width() == 0 {
		t.Fatalf("inline external table kind/keyword = %v/%#v", inline.Kind(), inline.InlineExternalTableKeyword())
	}
	if inline.Parameters().Count() != 0 {
		t.Fatalf("inline external table parameters = %d, want 0", inline.Parameters().Count())
	}
	if inline.Schema().Columns().Count() != 2 {
		t.Fatalf("inline external table schema columns = %d, want 2", inline.Schema().Columns().Count())
	}
	assertQueryGrammarRowSchemaColumn(t, inline.Schema().Columns().At(0).Element(), "Account", SyntaxKindStringKeyword)
	assertQueryGrammarRowSchemaColumn(t, inline.Schema().Columns().At(1).Element(), "Score", SyntaxKindLongKeyword)
	assertQueryGrammarInlineClauseValue(t, inline.KindParameter().KindKeyword(), inline.KindParameter().EqualToken(), inline.KindParameter().Value(), SyntaxKindKindKeyword, "storage")
	assertQueryGrammarInlineClauseValue(t, inline.DataFormatParameter().DataFormatKeyword(), inline.DataFormatParameter().EqualToken(), inline.DataFormatParameter().Value(), SyntaxKindDataFormatKeyword, "csv")

	wantConnectionKinds := []SyntaxKind{SyntaxKindStringLiteralExpression, SyntaxKindFunctionCallExpression, SyntaxKindAddExpression}
	connections := inline.ConnectionStrings()
	if connections.OpenParen().Width() == 0 || connections.CloseParen().Width() == 0 || connections.ConnectionStrings().Count() != len(wantConnectionKinds) {
		t.Fatalf("inline external table connections = %#v", connections)
	}
	for i, want := range wantConnectionKinds {
		if got := connections.ConnectionStrings().At(i).Element().Kind(); got != want {
			t.Fatalf("connection %d kind = %v, want %v", i, got, want)
		}
	}
	if separator := connections.ConnectionStrings().At(2).Separator(); separator == nil || separator.Kind() != SyntaxKindCommaToken {
		t.Fatalf("connection trailing comma = %#v", separator)
	}

	with := inline.WithClause()
	if with == nil || with.Properties().Count() != 2 ||
		NameDeclarationSimpleName(with.Properties().At(0).Element().Name()) != "format" ||
		with.Properties().At(0).Element().Expression().Kind() != SyntaxKindStringLiteralExpression ||
		NameDeclarationSimpleName(with.Properties().At(1).Element().Name()) != "compressed" ||
		with.Properties().At(1).Element().Expression().Kind() != SyntaxKindBooleanLiteralExpression {
		t.Fatalf("inline external table with clause = %#v", with)
	}
	if diagnostics := inline.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid inline external table diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicInlineExternalTableSchemaRecoveryAndSchemaLessKinds(t *testing.T) {
	emptyText := `inline_external_table() kind=storage dataformat=csv ("https://one")`
	emptyResult := QueryGrammarFrom(nil).Expression.ParseValue(NewArraySource(Lex(emptyText)), 0)
	if emptyResult.Length != 3 {
		t.Fatalf("empty inline schema prefix parse = %#v, want three-token function call", emptyResult)
	}
	emptySchema := emptyResult.Value
	if _, ok := emptySchema.(*FunctionCallExpression); !ok {
		t.Fatalf("empty inline schema expression = %T %#v, want function call", emptySchema, emptySchema)
	}

	delta := parseQueryGrammarInlineExternalTable(t, `inline_external_table kind=delta ("https://delta")`)
	if delta.Schema().OpenParen().Width() != 0 || delta.Schema().CloseParen().Width() != 0 || delta.Schema().Columns().Count() != 0 ||
		len(delta.Schema().ContainedDiagnostics()) != 0 {
		t.Fatalf("delta omitted schema = %#v diagnostics:%#v", delta.Schema(), delta.Schema().ContainedDiagnostics())
	}
	if delta.DataFormatParameter().DataFormatKeyword().Width() != 0 || delta.DataFormatParameter().EqualToken().Width() != 0 ||
		delta.DataFormatParameter().Value().Width() != 0 || len(delta.DataFormatParameter().ContainedDiagnostics()) != 0 {
		t.Fatalf("delta omitted dataformat = %#v diagnostics:%#v", delta.DataFormatParameter(), delta.DataFormatParameter().ContainedDiagnostics())
	}
	if diagnostics := delta.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("schema-less delta diagnostics = %#v", diagnostics)
	}

	storage := parseQueryGrammarInlineExternalTable(t, `inline_external_table kind=storage ("https://storage")`)
	if storage.Schema().Width() != 0 || !diagnosticsContainMessage(storage.Schema().ContainedDiagnostics(), "KS006", "Missing schema declaration") {
		t.Fatalf("schema-less storage schema = %#v diagnostics:%#v", storage.Schema(), storage.Schema().ContainedDiagnostics())
	}
	if storage.DataFormatParameter().Width() != 0 ||
		!diagnosticsContainMessage(storage.DataFormatParameter().ContainedDiagnostics(), "KS253", "DataFormat definition is expected.") {
		t.Fatalf("schema-less storage dataformat = %#v diagnostics:%#v", storage.DataFormatParameter(), storage.DataFormatParameter().ContainedDiagnostics())
	}

	explicit := parseQueryGrammarInlineExternalTable(t, `inline_external_table kind=delta dataformat=parquet ("https://delta")`)
	assertQueryGrammarInlineClauseValue(t, explicit.DataFormatParameter().DataFormatKeyword(), explicit.DataFormatParameter().EqualToken(), explicit.DataFormatParameter().Value(), SyntaxKindDataFormatKeyword, "parquet")
	if diagnostics := explicit.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("explicit delta dataformat diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicInlineExternalTableDeclaredOnlyParameters(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	text := `inline_external_table hint.materialized=true (Account:string) kind=storage dataformat=csv ("https://one")`
	tokens := Lex(text)
	result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
	reference, ok := result.Value.(*NameReference)
	if result.Length != 1 || !ok || NameReferenceSimpleName(reference) != "inline_external_table" || result.Length == len(tokens)-1 {
		t.Fatalf("undeclared inline parameter dispatch = %#v, want bare name with unconsumed parameter tail", result)
	}
}

func TestQueryGrammarPublicInlineExternalTableKindRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	missing := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Account:string) dataformat=csv ("https://one")`)
	if missing.KindParameter().Width() != 0 ||
		!diagnosticsContainMessage(missing.KindParameter().ContainedDiagnostics(), "KS254", "External Table Kind definition is expected.") {
		t.Fatalf("missing inline kind = %#v diagnostics:%#v", missing.KindParameter(), missing.KindParameter().ContainedDiagnostics())
	}

	text := `inline_external_table(Account:string) kind=bogus dataformat=csv ("https://one")`
	tokens := Lex(text)
	result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
	inline, ok := result.Value.(*InlineExternalTableExpression)
	if result.Length < 0 || result.Length == len(tokens)-1 || !ok {
		t.Fatalf("invalid inline kind parse = %#v, want recovered prefix with unconsumed tail", result)
	}
	if inline.KindParameter().Value().Width() != 0 ||
		!diagnosticsContainMessage(inline.KindParameter().Value().Diagnostics(), "KS005", "Expected one of: storage, delta") {
		t.Fatalf("invalid inline kind value = %#v diagnostics:%#v", inline.KindParameter().Value(), inline.KindParameter().ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInlineExternalTableConnectionRecovery(t *testing.T) {
	missing := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Account:string) kind=storage dataformat=csv`)
	connections := missing.ConnectionStrings()
	if connections.Width() != 0 || connections.ConnectionStrings().Count() != 1 ||
		!diagnosticsContainMessage(connections.OpenParen().Diagnostics(), "KS005", "Expected: (") ||
		!diagnosticsContainMessage(connections.ConnectionStrings().At(0).Element().ContainedDiagnostics(), "KS006", "Missing expression") ||
		!diagnosticsContainMessage(connections.CloseParen().Diagnostics(), "KS005", "Expected: )") ||
		diagnosticsContain(connections.ContainedDiagnostics(), "KS256") {
		t.Fatalf("missing inline connections = %#v diagnostics:%#v", connections, connections.ContainedDiagnostics())
	}

	empty := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Account:string) kind=storage dataformat=csv ()`)
	connections = empty.ConnectionStrings()
	if connections.OpenParen().Width() == 0 || connections.CloseParen().Width() == 0 || connections.ConnectionStrings().Count() != 1 {
		t.Fatalf("present empty inline connections = %#v", connections)
	}
	value := connections.ConnectionStrings().At(0).Element()
	if value.Width() != 0 || !diagnosticsContainMessage(value.ContainedDiagnostics(), "KS006", "Missing expression") ||
		diagnosticsContain(connections.ContainedDiagnostics(), "KS256") {
		t.Fatalf("present empty inline connection recovery = %#v diagnostics:%#v", value, connections.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInlineExternalTablePartitionColumns(t *testing.T) {
	inline := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage partition by (When:datetime=startofday(Timestamp), Bucket:long=hash(Source), Tenant:string) dataformat=csv ("https://one")`)
	partition := inline.PartitionClause()
	if partition == nil || partition.PartitionKeyword().Kind() != SyntaxKindPartitionKeyword ||
		partition.ByKeyword().Kind() != SyntaxKindByKeyword || partition.PartitionColumns().Count() != 3 {
		t.Fatalf("inline partition clause = %#v", partition)
	}
	wantTypes := []SyntaxKind{SyntaxKindDateTimeKeyword, SyntaxKindLongKeyword, SyntaxKindStringKeyword}
	for i, want := range wantTypes {
		column := partition.PartitionColumns().At(i).Element()
		primitive, ok := column.Type().(*PrimitiveTypeExpression)
		if !ok || primitive.Type().Kind() != want {
			t.Fatalf("partition column %d type = %#v, want %v", i, column.Type(), want)
		}
	}
	if partition.PartitionColumns().At(0).Element().Equal() == nil || partition.PartitionColumns().At(0).Element().Expr().Kind() != SyntaxKindFunctionCallExpression ||
		partition.PartitionColumns().At(1).Element().Equal() == nil || partition.PartitionColumns().At(1).Element().Expr().Kind() != SyntaxKindFunctionCallExpression ||
		partition.PartitionColumns().At(2).Element().Equal() != nil || partition.PartitionColumns().At(2).Element().Expr() != nil {
		t.Fatalf("partition equals/expressions = %#v", partition.PartitionColumns())
	}
	if diagnostics := inline.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid inline partition diagnostics = %#v", diagnostics)
	}

	invalid := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage partition by (Tenant:bool) dataformat=csv ("https://one")`)
	invalidType := invalid.PartitionClause().PartitionColumns().At(0).Element().Type()
	if invalidType.Kind() != SyntaxKindPrimitiveTypeExpression || !diagnosticsContain(invalidType.ContainedDiagnostics(), "KS160") {
		t.Fatalf("invalid partition type = %#v diagnostics:%#v", invalidType, invalidType.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInlineExternalTablePartitionExpressionAndCommaRecovery(t *testing.T) {
	independent := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage partition by (Tenant:string Source) dataformat=csv ("https://one")`)
	column := independent.PartitionClause().PartitionColumns().At(0).Element()
	if column.Equal() != nil || column.Expr() == nil || column.Expr().Kind() != SyntaxKindNameReference {
		t.Fatalf("partition expression independent of equals = %#v", column)
	}

	tailText := `inline_external_table(Source:string) kind=storage partition by (,Tenant:string,,Bucket:long,) dataformat=csv ("https://one")`
	tailTokens := Lex(tailText)
	tailResult := QueryGrammarFrom(nil).Expression.ParseValue(NewArraySource(tailTokens), 0)
	tailInline, ok := tailResult.Value.(*InlineExternalTableExpression)
	if tailResult.Length < 0 || tailResult.Length == len(tailTokens)-1 || !ok {
		t.Fatalf("partition double-comma tail = %#v, want recovered prefix", tailResult)
	}
	partition := tailInline.PartitionClause()
	if partition.LeadingComma() == nil || partition.PartitionColumns().Count() != 1 || partition.CloseParen().Width() != 0 {
		t.Fatalf("partition double-comma prefix recovery = %#v", partition)
	}
	if trailing := partition.PartitionColumns().At(0).Separator(); trailing == nil || trailing.Kind() != SyntaxKindCommaToken {
		t.Fatalf("partition trailing comma = %#v", trailing)
	}

	recovered := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage partition by (,,) dataformat=csv ("https://one")`)
	partition = recovered.PartitionClause()
	if partition.LeadingComma() == nil || partition.PartitionColumns().Count() != 1 || partition.PartitionColumns().At(0).Separator() == nil {
		t.Fatalf("partition missing primary recovery = %#v", partition)
	}
	missing := partition.PartitionColumns().At(0).Element()
	if missing.Width() != 0 || !diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS252", "Partition column definition is expected.") ||
		missing.Name() == nil || missing.Colon() == nil || missing.Type() == nil || missing.Equal() == nil || missing.Expr() == nil {
		t.Fatalf("missing partition column factory = %#v diagnostics:%#v", missing, missing.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInlineExternalTablePathFormatAST(t *testing.T) {
	inline := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage pathformat=("/" Tenant "/" datetime_pattern("yyyy/MM/dd", Day) "-" Region) dataformat=csv ("https://one")`)
	path := inline.PathFormat()
	if path == nil || path.PathFormatKeyword().Kind() != SyntaxKindPathFormatKeyword || path.EqualToken().Width() == 0 ||
		path.OpenParen().Width() == 0 || path.CloseParen().Width() == 0 || path.OptionalSeparatorLiteral().Token().ValueText() != "/" ||
		path.PathExpressions().Count() != 3 {
		t.Fatalf("inline pathformat = %#v", path)
	}
	first := path.PathExpressions().At(0)
	if first.PartitionColumnExpression().Kind() != SyntaxKindNameReference ||
		NameReferenceSimpleName(first.PartitionColumnExpression().(*NameReference)) != "Tenant" ||
		first.SeparatorLiteral() == nil || first.SeparatorLiteral().Token().ValueText() != "/" {
		t.Fatalf("first pathformat element = %#v", first)
	}
	pattern, ok := path.PathExpressions().At(1).PartitionColumnExpression().(*DateTimePattern)
	if !ok || pattern.StringLiteral().Token().ValueText() != "yyyy/MM/dd" ||
		NameReferenceSimpleName(pattern.PartitionColumn()) != "Day" ||
		path.PathExpressions().At(1).SeparatorLiteral().Token().ValueText() != "-" {
		t.Fatalf("datetime pathformat element = %#v", path.PathExpressions().At(1))
	}
	last := path.PathExpressions().At(2)
	if NameReferenceSimpleName(last.PartitionColumnExpression().(*NameReference)) != "Region" || last.SeparatorLiteral() != nil {
		t.Fatalf("last pathformat element = %#v", last)
	}
	if diagnostics := inline.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("valid inline pathformat diagnostics = %#v", diagnostics)
	}
}

func TestQueryGrammarPublicInlineExternalTablePathFormatRecovery(t *testing.T) {
	empty := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage pathformat=() dataformat=csv ("https://one")`)
	path := empty.PathFormat()
	if path.PathExpressions().Count() != 1 {
		t.Fatalf("empty pathformat list = %#v", path.PathExpressions())
	}
	missing := path.PathExpressions().At(0)
	if missing.Width() != 0 || missing.PartitionColumnExpression() == nil || missing.SeparatorLiteral() == nil ||
		!diagnosticsContainMessage(missing.ContainedDiagnostics(), "KS255", "External Table Path Format is empty.") {
		t.Fatalf("missing pathformat token factory = %#v diagnostics:%#v", missing, missing.ContainedDiagnostics())
	}

	missingClauseTokens := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage pathformat Tenant dataformat=csv ("https://one")`).PathFormat()
	if missingClauseTokens.EqualToken().Width() != 0 || missingClauseTokens.OpenParen().Width() != 0 || missingClauseTokens.CloseParen().Width() != 0 ||
		missingClauseTokens.PathExpressions().Count() != 1 ||
		!diagnosticsContainMessage(missingClauseTokens.EqualToken().Diagnostics(), "KS005", "Expected: =") ||
		!diagnosticsContainMessage(missingClauseTokens.OpenParen().Diagnostics(), "KS005", "Expected: (") ||
		!diagnosticsContainMessage(missingClauseTokens.CloseParen().Diagnostics(), "KS005", "Expected: )") {
		t.Fatalf("missing pathformat clause tokens = %#v diagnostics:%#v", missingClauseTokens, missingClauseTokens.ContainedDiagnostics())
	}

	dateTime := parseQueryGrammarInlineExternalTable(t,
		`inline_external_table(Source:string) kind=storage pathformat=(datetime_pattern()) dataformat=csv ("https://one")`).PathFormat().PathExpressions().At(0).PartitionColumnExpression().(*DateTimePattern)
	if dateTime.StringLiteral().Width() != 0 || dateTime.Comma().Width() != 0 || dateTime.PartitionColumn().Width() != 0 ||
		!diagnosticsContainMessage(dateTime.StringLiteral().ContainedDiagnostics(), "KS006", "Missing string") ||
		!diagnosticsContainMessage(dateTime.Comma().Diagnostics(), "KS005", "Expected: ,") ||
		!diagnosticsContainMessage(dateTime.PartitionColumn().ContainedDiagnostics(), "KS006", "Missing name") {
		t.Fatalf("datetime_pattern missing tokens = %#v diagnostics:%#v", dateTime, dateTime.ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInlineExternalTableRootPipelineAndOperatorPlacement(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	root := parseQueryGrammarProduction(t, grammar.Expression,
		`inline_external_table(Account:string) kind=storage dataformat=csv ("https://one")`)
	if _, ok := root.(*InlineExternalTableExpression); !ok {
		t.Fatalf("inline root = %T, want *InlineExternalTableExpression", root)
	}

	pipeline := parseQueryGrammarProduction(t, grammar.Expression,
		`inline_external_table(Account:string) kind=storage dataformat=csv ("https://one") | where Account != "" | project Account`)
	projectPipe, _ := queryGrammarPipeStage(t, pipeline, SyntaxKindProjectOperator)
	wherePipe, _ := queryGrammarPipeStage(t, projectPipe.Expression(), SyntaxKindFilterOperator)
	if _, ok := wherePipe.Expression().(*InlineExternalTableExpression); !ok {
		t.Fatalf("inline pipeline source = %T, want *InlineExternalTableExpression", wherePipe.Expression())
	}

	text := `inline_external_table(Account:string) kind=storage dataformat=csv ("https://one")`
	assertQueryGrammarProductionRejects(t, grammar.QueryOperator, text)
	assertQueryGrammarProductionRejects(t, grammar.FollowingPipeElementExpression, text)

	followingText := `datatable(Account:string)["old"] | ` + text
	tokens := Lex(followingText)
	result := grammar.Expression.ParseValue(NewArraySource(tokens), 0)
	pipe, ok := result.Value.(*PipeExpression)
	if result.Length < 0 || result.Length == len(tokens)-1 || !ok {
		t.Fatalf("inline following-operator parse = %#v, want recovered pipe with unconsumed inline tail", result)
	}
	bad, ok := pipe.Operator().(*BadQueryOperator)
	if !ok || bad.Keyword().Width() != 0 || !diagnosticsContain(bad.ContainedDiagnostics(), "KS176") {
		t.Fatalf("inline following operator recovery = %#v diagnostics:%#v", pipe.Operator(), pipe.Operator().ContainedDiagnostics())
	}
}

func TestQueryGrammarPublicInlineExternalTableBareKeywordAndFunctionCallDispatch(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	bare := parseQueryGrammarProduction(t, grammar.Expression, `inline_external_table`)
	reference, ok := bare.(*NameReference)
	if !ok || NameReferenceSimpleName(reference) != "inline_external_table" {
		t.Fatalf("bare inline_external_table = %#v, want name reference", bare)
	}

	call := parseQueryGrammarProduction(t, grammar.Expression, `inline_external_table()`)
	if _, ok := call.(*FunctionCallExpression); !ok {
		t.Fatalf("inline_external_table empty schema dispatch = %#v, want function call", call)
	}
}

func parseQueryGrammarInlineExternalTable(t *testing.T, text string) *InlineExternalTableExpression {
	t.Helper()
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression, text)
	inline, ok := expression.(*InlineExternalTableExpression)
	if !ok {
		t.Fatalf("inline external table expression = %T, want *InlineExternalTableExpression", expression)
	}
	return inline
}

func assertQueryGrammarInlineClauseValue(t *testing.T, keyword, equal, value *SyntaxToken, wantKeyword SyntaxKind, wantValue string) {
	t.Helper()
	if keyword.Kind() != wantKeyword || keyword.Width() == 0 || equal.Kind() != SyntaxKindEqualToken || equal.Width() == 0 ||
		value.Width() == 0 || value.ValueText() != wantValue {
		t.Fatalf("inline clause = keyword:%#v equal:%#v value:%#v, want %v=%s", keyword, equal, value, wantKeyword, wantValue)
	}
}
