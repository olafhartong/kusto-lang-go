package kustolang

import "testing"

func TestDiagnosticFactsMessages(t *testing.T) {
	if d := DiagnosticSummableLiteralExpected(); d.Code != "KS224" || d.Message != "Summable literal expected." {
		t.Fatalf("unexpected KS224 diagnostic: %#v", d)
	}
	if d := DiagnosticUnknownQueryOperatorParameterName("Hint"); d.Severity != DiagnosticSeverityWarning {
		t.Fatalf("unexpected KS179 severity: %#v", d)
	}
	if d := DiagnosticFunctionExpectsArgumentCountExact("MyFunction", 1); d.Code != "KS119" || d.Message != "The function 'MyFunction' expects 1 argument." {
		t.Fatalf("unexpected KS119 diagnostic: %#v", d)
	}
	if d := DiagnosticNameDoesNotReferToAnyKnownFunction("FileProfile"); d.Code != "KS211" || d.Message != "The name 'FileProfile' does not refer to any known function." {
		t.Fatalf("unexpected KS211 diagnostic: %#v", d)
	}
	if d := DiagnosticAggregateNotAllowedInThisContext("sum"); d.Code != "KS144" || d.Message != "The aggregate function 'sum' is not allowed in this context." {
		t.Fatalf("unexpected KS144 diagnostic: %#v", d)
	}
	if d := DiagnosticInvalidNameInAggregateContext("Score"); d.Code != "KS236" || d.Message != "The name 'Score' is not an aggregate or scalar function." {
		t.Fatalf("unexpected KS236 diagnostic: %#v", d)
	}
	if d := DiagnosticNoCommonArgumentType(); d.Code != "KS232" || d.Message != "No common type can be determined from the arguments." {
		t.Fatalf("unexpected KS232 diagnostic: %#v", d)
	}
	if d := DiagnosticPluginFunctionNotDefined("unknown_plugin"); d.Code != "KS124" || d.Message != "The plug-in function 'unknown_plugin' is not defined." {
		t.Fatalf("unexpected KS124 diagnostic: %#v", d)
	}
	if d := DiagnosticFunctionRequiresArgumentList("not"); d.Code != "KS117" || d.Message != "The function 'not' requires an argument list." {
		t.Fatalf("unexpected KS117 diagnostic: %#v", d)
	}
	if d := DiagnosticNameRefersToMoreThanOneItem("Shared"); d.Code != "KS146" || d.Message != "The name 'Shared' refers to more than one column or variable" {
		t.Fatalf("unexpected KS146 diagnostic: %#v", d)
	}
}

func TestBinderResolveAndValidate(t *testing.T) {
	table := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	fn := NewScalarFunctionSymbol("MyScalar", ScalarTypes.Bool.Type(), NewParameter("x", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table, fn))
	binder := NewBinder(globals)

	tableInfo := binder.ResolveTable("SigninLogs")
	if tableInfo.ReferencedSymbol() != table || len(tableInfo.Diagnostics) != 0 {
		t.Fatalf("table info = %#v", tableInfo)
	}
	fnInfo := binder.ResolveFunction("MyScalar")
	if fnInfo.ReferencedSymbol() != fn || fnInfo.ReferencedSignature() == nil || len(fnInfo.Diagnostics) != 0 {
		t.Fatalf("function info = %#v", fnInfo)
	}
	colInfo := binder.ResolveColumn(table, "UserPrincipalName")
	if colInfo.ReferencedSymbol() == nil || colInfo.ResultType.SymbolName() != "string" {
		t.Fatalf("column info = %#v", colInfo)
	}
	if diags := binder.ValidateFunctionArgumentCount("MyScalar", 0); len(diags) != 1 || diags[0].Code != "KS119" {
		t.Fatalf("argument diagnostics = %#v", diags)
	}
}

func TestBinderResolveNameFindsLocalAmbientDatabaseAndBuiltIns(t *testing.T) {
	table := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	databaseFn := NewScalarFunctionSymbol("MyScalar", ScalarTypes.Bool.Type(), NewParameter("x", ScalarTypes.String.Type()))
	ambient := NewVariableSymbol("AmbientScalar", ScalarTypes.Long.Type(), true, int64(5), nil)
	globals := NewGlobalState(NewDatabaseSymbol("db", table, databaseFn)).
		WithAmbientSymbols([]Symbol{ambient})
	binder := NewBinder(globals)
	local := NewVariableSymbol("SigninLogs", ScalarTypes.String.Type(), true, "shadowed", nil)
	binder.LocalScope.AddSymbol(local)

	localInfo := binder.ResolveName("SigninLogs", SymbolMatchDefault, false)
	if localInfo.ReferencedSymbol() != local || localInfo.ResultType != ScalarTypes.String.Type() || !localInfo.IsConstant || len(localInfo.Diagnostics) != 0 {
		t.Fatalf("local name semantic info = %#v", localInfo)
	}

	tableInfo := binder.ResolveName("SigninLogs", SymbolMatchTable, false)
	if tableInfo.ReferencedSymbol() != table || tableInfo.ResultType != &table.TypeSymbol || len(tableInfo.Diagnostics) != 0 {
		t.Fatalf("table name semantic info = %#v", tableInfo)
	}

	fnInfo := binder.ResolveName("MyScalar", SymbolMatchFunction, false)
	if fnInfo.ReferencedSymbol() != databaseFn || fnInfo.ReferencedSignature() == nil || len(fnInfo.Diagnostics) != 0 {
		t.Fatalf("database function name semantic info = %#v", fnInfo)
	}

	ambientInfo := binder.ResolveName("AmbientScalar", SymbolMatchLocal|SymbolMatchScalar, false)
	if ambientInfo.ReferencedSymbol() != ambient || ambientInfo.ResultType != ScalarTypes.Long.Type() || !ambientInfo.IsConstant || len(ambientInfo.Diagnostics) != 0 {
		t.Fatalf("ambient name semantic info = %#v", ambientInfo)
	}

	builtInInfo := NewBinder(nil).ResolveName("strlen", SymbolMatchFunction, false)
	if builtInInfo.ReferencedSymbol() == nil || builtInInfo.ReferencedSymbol().SymbolName() != "strlen" || builtInInfo.ReferencedSignature() == nil || len(builtInInfo.Diagnostics) != 0 {
		t.Fatalf("built-in function name semantic info = %#v", builtInInfo)
	}
}

func TestBinderResolveNameInRowScopeUsesColumnFirstOrdering(t *testing.T) {
	table := NewTableSymbol("SigninLogs", NewColumnSymbol("UserPrincipalName", ScalarTypes.String.Type()))
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	binder := NewBinder(globals)
	local := NewVariableSymbol("SigninLogs", ScalarTypes.String.Type(), true, "shadowed", nil)
	binder.LocalScope.AddSymbol(local)
	rowColumn := NewColumnSymbol("SigninLogs", ScalarTypes.Long.Type())
	rowScope := NewTableSymbol("Rows", rowColumn)

	columnInfo := binder.ResolveNameInRowScope(rowScope, "SigninLogs", SymbolMatchDefault, false)
	if columnInfo.ReferencedSymbol() != rowColumn || columnInfo.ResultType != ScalarTypes.Long.Type() || len(columnInfo.Diagnostics) != 0 {
		t.Fatalf("row-scope name semantic info = %#v", columnInfo)
	}

	tableInfo := binder.ResolveNameInRowScope(rowScope, "SigninLogs", SymbolMatchTable, false)
	if tableInfo.ReferencedSymbol() != table || tableInfo.ResultType != &table.TypeSymbol || len(tableInfo.Diagnostics) != 0 {
		t.Fatalf("table-filtered row-scope name semantic info = %#v", tableInfo)
	}
}

func TestBinderResolveNameReportsAmbiguousMatches(t *testing.T) {
	table := NewTableSymbol("Shared", NewColumnSymbol("Id", ScalarTypes.Long.Type()))
	fn := NewScalarFunctionSymbol("Shared", ScalarTypes.String.Type())
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, fn)).WithFunctions(nil).WithAggregates(nil).WithPlugIns(nil))

	info := binder.ResolveName("Shared", SymbolMatchDefault, false)
	group, ok := info.ReferencedSymbol().(*GroupSymbol)
	if !ok || info.ResultType != &Error.TypeSymbol || len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS146" {
		t.Fatalf("ambiguous name semantic info = %#v", info)
	}
	members := group.Members()
	if len(members) != 2 || members[0] != table || members[1] != fn {
		t.Fatalf("ambiguous members = %#v", members)
	}

	tableOnly := binder.ResolveName("Shared", SymbolMatchTable, false)
	if tableOnly.ReferencedSymbol() != table || len(tableOnly.Diagnostics) != 0 {
		t.Fatalf("table-filtered semantic info = %#v", tableOnly)
	}
}

func TestBinderResolveNameInRowScopesMergesCommonColumns(t *testing.T) {
	left := NewColumnSymbolFull("Account", ScalarTypes.String.Type(), "left account", nil, nil, []string{"alice"})
	right := NewColumnSymbol("Account", ScalarTypes.Long.Type())
	leftScope := NewTableSymbol("Left", left)
	rightScope := NewTableSymbol("Right", right)
	binder := NewBinder(NewGlobalState(nil))

	info := binder.ResolveNameInRowScopes(leftScope, rightScope, true, "Account", SymbolMatchDefault, false)
	common, ok := info.ReferencedSymbol().(*ColumnSymbol)
	if !ok || common.SymbolName() != "Account" || common.Type != ScalarTypes.String.Type() || common.Description != "left account" || len(info.Diagnostics) != 0 {
		t.Fatalf("common column semantic info = %#v", info)
	}
	if len(common.OriginalColumns) != 2 || common.OriginalColumns[0] != left || common.OriginalColumns[1] != right {
		t.Fatalf("common original columns = %#v", common.OriginalColumns)
	}

	leftOnly := binder.ResolveNameInRowScopes(leftScope, rightScope, false, "Account", SymbolMatchDefault, false)
	if leftOnly.ReferencedSymbol() != left || len(leftOnly.Diagnostics) != 0 {
		t.Fatalf("non-common row-scope semantic info = %#v", leftOnly)
	}
}

func TestBinderResolveNameInRowScopesUsesRightScopeBeforeLocalsAndCatalog(t *testing.T) {
	table := NewTableSymbol("RightOnly", NewColumnSymbol("Id", ScalarTypes.Long.Type()))
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table)))
	local := NewVariableSymbol("RightOnly", ScalarTypes.String.Type(), true, "shadowed", nil)
	binder.LocalScope.AddSymbol(local)
	rightColumn := NewColumnSymbol("RightOnly", ScalarTypes.Bool.Type())
	rightScope := NewTableSymbol("Right", rightColumn)

	info := binder.ResolveNameInRowScopes(nil, rightScope, false, "RightOnly", SymbolMatchDefault, false)
	if info.ReferencedSymbol() != rightColumn || info.ResultType != ScalarTypes.Bool.Type() || len(info.Diagnostics) != 0 {
		t.Fatalf("right row-scope semantic info = %#v", info)
	}

	tableInfo := binder.ResolveNameInRowScopes(nil, rightScope, false, "RightOnly", SymbolMatchTable, false)
	if tableInfo.ReferencedSymbol() != table || tableInfo.ResultType != &table.TypeSymbol || len(tableInfo.Diagnostics) != 0 {
		t.Fatalf("right row-scope table-filtered semantic info = %#v", tableInfo)
	}
}

func TestBinderResolveNameInRowScopesBindsLeftAndRightTuples(t *testing.T) {
	leftColumn := NewColumnSymbol("LeftAccount", ScalarTypes.String.Type())
	rightColumn := NewColumnSymbol("RightAccount", ScalarTypes.Bool.Type())
	leftScope := NewTableSymbol("Left", leftColumn)
	rightScope := NewTableSymbol("Right", rightColumn)
	binder := NewBinder(NewGlobalState(nil))
	binder.LocalScope.AddSymbol(NewVariableSymbol("$right", ScalarTypes.String.Type(), true, "shadowed", nil))

	leftInfo := binder.ResolveNameInRowScopes(leftScope, rightScope, false, "$left", SymbolMatchDefault, false)
	leftTuple := assertTupleInfo(t, leftInfo, leftScope, leftColumn)
	if cached := binder.ResolveNameInRowScopes(leftScope, rightScope, false, "$left", SymbolMatchDefault, false); cached.ReferencedSymbol() != leftTuple {
		t.Fatalf("$left tuple was not cached: %#v", cached)
	}

	rightInfo := binder.ResolveNameInRowScopes(leftScope, rightScope, false, "$right", SymbolMatchDefault, false)
	rightTuple := assertTupleInfo(t, rightInfo, rightScope, rightColumn)
	if rightTuple.RelatedTable != rightScope {
		t.Fatalf("$right related table = %#v", rightTuple.RelatedTable)
	}

	rightOnly := binder.ResolveNameInRowScopes(nil, rightScope, false, "$right", SymbolMatchDefault, false)
	if rightOnly.ReferencedSymbol() != rightTuple || len(rightOnly.Diagnostics) != 0 {
		t.Fatalf("$right without left scope semantic info = %#v", rightOnly)
	}

	leftWithoutRight := binder.ResolveNameInRowScopes(leftScope, nil, false, "$left", SymbolMatchDefault, false)
	if leftWithoutRight.ReferencedSymbol() != Error || len(leftWithoutRight.Diagnostics) != 1 || leftWithoutRight.Diagnostics[0].Code != "KS142" {
		t.Fatalf("$left without right scope semantic info = %#v", leftWithoutRight)
	}

	leftReference := MustNewNameReferenceTyped(MustNewTokenNameTyped(syntaxToken(SyntaxKindIdentifierToken, "$left")), SymbolMatchDefault)
	referenceInfo := binder.ResolveNameReferenceInRowScopes(leftScope, rightScope, false, leftReference, false)
	if referenceInfo.ReferencedSymbol() != leftTuple || len(referenceInfo.Diagnostics) != 0 {
		t.Fatalf("$left name-reference semantic info = %#v", referenceInfo)
	}
}

func assertTupleInfo(t *testing.T, info SemanticInfo, relatedTable *TableSymbol, columns ...*ColumnSymbol) *TupleSymbol {
	t.Helper()
	tuple, ok := info.ReferencedSymbol().(*TupleSymbol)
	if !ok || tuple.RelatedTable != relatedTable || info.ResultType != &tuple.TypeSymbol || len(info.Diagnostics) != 0 {
		t.Fatalf("tuple semantic info = %#v", info)
	}
	got := tuple.Columns()
	if len(got) != len(columns) {
		t.Fatalf("tuple columns = %#v", got)
	}
	for i := range columns {
		if got[i] != columns[i] {
			t.Fatalf("tuple column %d = %#v, want %#v", i, got[i], columns[i])
		}
	}
	return tuple
}

func TestBinderResolveNameInOpenRowScopeInfersColumnsAfterCatalogLookup(t *testing.T) {
	fn := NewScalarFunctionSymbol("KnownFunction", ScalarTypes.String.Type())
	table := NewTableSymbol("KnownTable", NewColumnSymbol("Id", ScalarTypes.Long.Type()))
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", table, fn)))
	rowScope := NewTableSymbol("OpenRows").WithIsOpen(true)

	inferred := binder.ResolveNameInRowScope(rowScope, "NewColumn", SymbolMatchDefault, false)
	column, ok := inferred.ReferencedSymbol().(*ColumnSymbol)
	if !ok || column.SymbolName() != "NewColumn" || inferred.ResultType != ScalarTypes.Unknown.Type() || len(inferred.Diagnostics) != 0 {
		t.Fatalf("open row-scope inferred column semantic info = %#v", inferred)
	}

	tableInfo := binder.ResolveNameInRowScope(rowScope, "KnownTable", SymbolMatchDefault, false)
	if tableInfo.ReferencedSymbol() != table || tableInfo.ResultType != &table.TypeSymbol || len(tableInfo.Diagnostics) != 0 {
		t.Fatalf("open row-scope known table semantic info = %#v", tableInfo)
	}

	functionInfo := binder.ResolveNameInRowScope(rowScope, "KnownFunction", SymbolMatchDefault, false)
	if functionInfo.ReferencedSymbol() != fn || functionInfo.ReferencedSignature() == nil || len(functionInfo.Diagnostics) != 0 {
		t.Fatalf("open row-scope known function semantic info = %#v", functionInfo)
	}

	tableOnlyMissing := binder.ResolveNameInRowScope(rowScope, "MissingTable", SymbolMatchTable, false)
	openTable, ok := tableOnlyMissing.ReferencedSymbol().(*TableSymbol)
	if !ok || openTable.SymbolName() != "MissingTable" || !openTable.IsOpen() || len(tableOnlyMissing.Diagnostics) != 1 || tableOnlyMissing.Diagnostics[0].Code != "KS204" {
		t.Fatalf("open row-scope table-only missing semantic info = %#v", tableOnlyMissing)
	}
}

func TestBinderResolveNameInClosedRowScopeReportsUnknownColumn(t *testing.T) {
	binder := NewBinder(NewGlobalState(nil))
	rowScope := NewTableSymbol("Rows", NewColumnSymbol("Known", ScalarTypes.String.Type()))

	missing := binder.ResolveNameInRowScope(rowScope, "Missing", SymbolMatchDefault, false)
	if missing.ReferencedSymbol() != Error || missing.ResultType != &Error.TypeSymbol || len(missing.Diagnostics) != 1 || missing.Diagnostics[0].Code != "KS142" {
		t.Fatalf("closed row-scope missing semantic info = %#v", missing)
	}
}

func TestBinderResolveNameReportsUnknownScalarAndTabularNames(t *testing.T) {
	binder := NewBinder(NewGlobalState(nil))

	scalar := binder.ResolveName("MissingScalar", SymbolMatchDefault, false)
	if scalar.ReferencedSymbol() != Error || scalar.ResultType != &Error.TypeSymbol {
		t.Fatalf("unknown scalar semantic info = %#v", scalar)
	}
	if len(scalar.Diagnostics) != 1 || scalar.Diagnostics[0].Code != "KS142" {
		t.Fatalf("unknown scalar diagnostics = %#v", scalar.Diagnostics)
	}

	tabular := binder.ResolveName("MissingRows", SymbolMatchDefault, true)
	table, ok := tabular.ReferencedSymbol().(*TableSymbol)
	if !ok || table.SymbolName() != "MissingRows" || !table.IsOpen() || tabular.ResultType != &table.TypeSymbol {
		t.Fatalf("unknown tabular semantic info = %#v", tabular)
	}
	if len(tabular.Diagnostics) != 1 || tabular.Diagnostics[0].Code != "KS204" {
		t.Fatalf("unknown tabular diagnostics = %#v", tabular.Diagnostics)
	}

	tableOrFunction := binder.ResolveName("MissingTableOrFunction", SymbolMatchTable|SymbolMatchFunction, false)
	table, ok = tableOrFunction.ReferencedSymbol().(*TableSymbol)
	if !ok || table.SymbolName() != "MissingTableOrFunction" || !table.IsOpen() || len(tableOrFunction.Diagnostics) != 1 || tableOrFunction.Diagnostics[0].Code != "KS204" {
		t.Fatalf("unknown table-or-function semantic info = %#v", tableOrFunction)
	}

	external := binder.ResolveName("MissingExternal", SymbolMatchExternalTable, false)
	if external.ReferencedSymbol() != Error || len(external.Diagnostics) != 1 || external.Diagnostics[0].Code != "KS209" {
		t.Fatalf("unknown external-table semantic info = %#v", external)
	}

	function := binder.ResolveName("MissingFn", SymbolMatchFunction, false)
	if function.ReferencedSymbol() != Error || len(function.Diagnostics) != 1 || function.Diagnostics[0].Code != "KS211" {
		t.Fatalf("unknown function semantic info = %#v", function)
	}

	database := binder.ResolveName("MissingDb", SymbolMatchDatabase, false)
	if database.ReferencedSymbol() != Error || len(database.Diagnostics) != 1 || database.Diagnostics[0].Code != "KS208" {
		t.Fatalf("unknown database semantic info = %#v", database)
	}
}

func TestBinderResolveNameReferenceUsesReferenceMatchAndClientParameterBranch(t *testing.T) {
	fn := NewScalarFunctionSymbol("MyScalar", ScalarTypes.Bool.Type())
	binder := NewBinder(NewGlobalState(NewDatabaseSymbol("db", fn)))
	reference := MustNewNameReferenceTyped(MustNewTokenNameTyped(syntaxToken(SyntaxKindIdentifierToken, "MyScalar")), SymbolMatchFunction)

	info := binder.ResolveNameReference(reference, false)
	if info.ReferencedSymbol() != fn || info.ReferencedSignature() == nil || len(info.Diagnostics) != 0 {
		t.Fatalf("ordinary name-reference semantic info = %#v", info)
	}

	client := binder.ResolveNameReference(bracedNameReference("tenant"), false)
	if client.ResultType != ScalarTypes.Unknown.Type() || len(client.Diagnostics) != 1 || client.Diagnostics[0].Code != "KS213" {
		t.Fatalf("client name-reference semantic info = %#v", client)
	}

	rowColumn := NewColumnSymbol("Account", ScalarTypes.String.Type())
	rowScope := NewTableSymbol("Rows", rowColumn)
	rowReference := MustNewNameReferenceTyped(MustNewTokenNameTyped(syntaxToken(SyntaxKindIdentifierToken, "Account")), SymbolMatchDefault)
	rowInfo := binder.ResolveNameReferenceInRowScope(rowScope, rowReference, false)
	if rowInfo.ReferencedSymbol() != rowColumn || rowInfo.ResultType != ScalarTypes.String.Type() || len(rowInfo.Diagnostics) != 0 {
		t.Fatalf("row-scope name-reference semantic info = %#v", rowInfo)
	}

	rightColumn := NewColumnSymbol("RightAccount", ScalarTypes.Bool.Type())
	rightScope := NewTableSymbol("Right", rightColumn)
	rightReference := MustNewNameReferenceTyped(MustNewTokenNameTyped(syntaxToken(SyntaxKindIdentifierToken, "RightAccount")), SymbolMatchDefault)
	rightInfo := binder.ResolveNameReferenceInRowScopes(rowScope, rightScope, false, rightReference, false)
	if rightInfo.ReferencedSymbol() != rightColumn || rightInfo.ResultType != ScalarTypes.Bool.Type() || len(rightInfo.Diagnostics) != 0 {
		t.Fatalf("right row-scope name-reference semantic info = %#v", rightInfo)
	}
}

func TestBinderResolveClientParameterReferenceMirrorsUpstream(t *testing.T) {
	disabled := NewBinder(NewGlobalState(nil)).ResolveClientParameterReference("tenant", false)
	if disabled.ResultType != ScalarTypes.Unknown.Type() || disabled.ReferencedSymbol() != nil {
		t.Fatalf("disabled client parameter semantic info = %#v", disabled)
	}
	if len(disabled.Diagnostics) != 1 || disabled.Diagnostics[0].Code != "KS213" {
		t.Fatalf("disabled diagnostics = %#v", disabled.Diagnostics)
	}

	tenant := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")
	enabledGlobals := GlobalStateWithProperty(NewGlobalState(nil), Properties.AllowClientParameters, true).
		WithClientSymbols([]Symbol{tenant})
	binder := NewBinder(enabledGlobals)

	known := binder.ResolveClientParameterReference("tenant", false)
	if known.ReferencedSymbol() != tenant || known.ResultType != ScalarTypes.String.Type() || !known.IsConstant || len(known.Diagnostics) != 0 {
		t.Fatalf("known client parameter semantic info = %#v", known)
	}

	unknownScalar := binder.ResolveClientParameterReference("missing", false)
	if unknownScalar.ReferencedSymbol() != nil || unknownScalar.ResultType != ScalarTypes.Unknown.Type() || len(unknownScalar.Diagnostics) != 0 {
		t.Fatalf("unknown scalar client parameter semantic info = %#v", unknownScalar)
	}

	unknownTabular := binder.ResolveClientParameterReference("Rows", true)
	table, ok := unknownTabular.ReferencedSymbol().(*TableSymbol)
	if !ok || table.SymbolName() != "Rows" || !table.IsOpen() || unknownTabular.ResultType != &table.TypeSymbol || len(unknownTabular.Diagnostics) != 0 {
		t.Fatalf("unknown tabular client parameter semantic info = %#v", unknownTabular)
	}
}

func TestBinderBindClientParameterReferencesAttachesDisabledDiagnostics(t *testing.T) {
	code := Parse("print x={tenant}")
	binder := NewBinder(code.Globals)
	if count := binder.BindClientParameterReferences(code.Syntax()); count != 1 {
		t.Fatalf("bound client parameter count = %d", count)
	}
	reference := singleBracedNameReference(t, code.Syntax())
	info, ok := reference.GetSemanticInfo()
	if !ok {
		t.Fatal("client parameter reference was not bound")
	}
	if info.ResultType != ScalarTypes.Unknown.Type() || info.ReferencedSymbol() != nil {
		t.Fatalf("disabled semantic info = %#v", info)
	}
	if len(info.Diagnostics) != 1 || info.Diagnostics[0].Code != "KS213" || !info.Diagnostics[0].HasLocation() {
		t.Fatalf("disabled diagnostics = %#v", info.Diagnostics)
	}
}

func TestBinderBindClientParameterReferencesUsesKnownClientSymbols(t *testing.T) {
	tenant := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")
	globals := GlobalStateWithProperty(NewGlobalState(nil), Properties.AllowClientParameters, true).
		WithClientSymbols([]Symbol{tenant})
	code := ParseWithGlobals("print x={tenant}", globals)
	binder := NewBinder(code.Globals)
	if count := binder.BindClientParameterReferences(code.Syntax()); count != 1 {
		t.Fatalf("bound client parameter count = %d", count)
	}
	info, ok := singleBracedNameReference(t, code.Syntax()).GetSemanticInfo()
	if !ok || info.ReferencedSymbol() != tenant || info.ResultType != ScalarTypes.String.Type() || !info.IsConstant || len(info.Diagnostics) != 0 {
		t.Fatalf("known client parameter semantic info = %#v/%t", info, ok)
	}
}

func TestBinderBindClientParameterReferencesInfersUnknownTabularContext(t *testing.T) {
	globals := GlobalStateWithProperty(NewGlobalState(nil), Properties.AllowClientParameters, true)

	pipeCode := ParseWithGlobals("{Rows} | count", globals)
	if count := NewBinder(pipeCode.Globals).BindClientParameterReferences(pipeCode.Syntax()); count != 1 {
		t.Fatalf("pipe bound client parameter count = %d", count)
	}
	pipeInfo, ok := singleBracedNameReference(t, pipeCode.Syntax()).GetSemanticInfo()
	assertClientParameterOpenTable(t, "pipe", pipeInfo, ok, "Rows")

	parenthesizedCode := ParseWithGlobals("({Rows}) | count", globals)
	if count := NewBinder(parenthesizedCode.Globals).BindClientParameterReferences(parenthesizedCode.Syntax()); count != 1 {
		t.Fatalf("parenthesized bound client parameter count = %d", count)
	}
	parenthesizedInfo, ok := singleBracedNameReference(t, parenthesizedCode.Syntax()).GetSemanticInfo()
	assertClientParameterOpenTable(t, "parenthesized pipe", parenthesizedInfo, ok, "Rows")

	callNameCode := ParseWithGlobals("{Rows}() | count", globals)
	if count := NewBinder(callNameCode.Globals).BindClientParameterReferences(callNameCode.Syntax()); count != 1 {
		t.Fatalf("function-call-name bound client parameter count = %d", count)
	}
	callNameInfo, ok := singleBracedNameReference(t, callNameCode.Syntax()).GetSemanticInfo()
	assertClientParameterOpenTable(t, "function-call-name pipe", callNameInfo, ok, "Rows")

	unionCode := ParseWithGlobals("union {Rows}", globals)
	if count := NewBinder(unionCode.Globals).BindClientParameterReferences(unionCode.Syntax()); count != 1 {
		t.Fatalf("union bound client parameter count = %d", count)
	}
	unionInfo, ok := singleBracedNameReference(t, unionCode.Syntax()).GetSemanticInfo()
	assertClientParameterOpenTable(t, "union", unionInfo, ok, "Rows")
}

func TestBinderBindClientParameterReferencesInfersDatabasePathSelectorContext(t *testing.T) {
	globals := GlobalStateWithProperty(NewGlobalState(nil), Properties.AllowClientParameters, true)
	databaseVariable := NewVariableSymbol("database", NewDatabaseSymbol("db"), false, nil, nil)
	left := nameReference("database")
	left.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(databaseVariable))
	path := MustNewPathExpressionTyped(
		left,
		syntaxToken(SyntaxKindDotToken, "."),
		bracedNameReference("Rows"),
	)

	root := path
	if count := NewBinder(globals).BindClientParameterReferences(root); count != 1 {
		t.Fatalf("database path bound client parameter count = %d", count)
	}
	info, ok := singleBracedNameReference(t, root).GetSemanticInfo()
	assertClientParameterOpenTable(t, "database path selector", info, ok, "Rows")
}

func TestBinderBindClientParameterReferencesInfersUnknownScalarContext(t *testing.T) {
	globals := GlobalStateWithProperty(NewGlobalState(nil), Properties.AllowClientParameters, true)
	code := ParseWithGlobals("print x={tenant}", globals)
	if count := NewBinder(code.Globals).BindClientParameterReferences(code.Syntax()); count != 1 {
		t.Fatalf("bound client parameter count = %d", count)
	}
	info, ok := singleBracedNameReference(t, code.Syntax()).GetSemanticInfo()
	if !ok || info.ResultType != ScalarTypes.Unknown.Type() || info.ReferencedSymbol() != nil || len(info.Diagnostics) != 0 {
		t.Fatalf("unknown scalar semantic info = %#v/%t", info, ok)
	}
}

func singleBracedNameReference(t *testing.T, root SyntaxElement) *NameReference {
	t.Helper()
	var references []*NameReference
	WalkSyntax(root, func(element SyntaxElement) {
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if !ok {
			return
		}
		if _, ok := reference.Name().(*BracedName); ok {
			references = append(references, reference)
		}
	})
	if len(references) != 1 {
		t.Fatalf("braced name references = %d, want 1", len(references))
	}
	return references[0]
}

func assertClientParameterOpenTable(t *testing.T, label string, info SemanticInfo, ok bool, name string) {
	t.Helper()
	table, tableOK := info.ReferencedSymbol().(*TableSymbol)
	if !ok || !tableOK || table.SymbolName() != name || !table.IsOpen() || info.ResultType != &table.TypeSymbol || len(info.Diagnostics) != 0 {
		t.Fatalf("%s tabular semantic info = %#v/%t", label, info, ok)
	}
}

func bracedNameReference(name string) *NameReference {
	return MustNewNameReferenceTyped(
		MustNewBracedNameTyped(
			syntaxToken(SyntaxKindOpenBraceToken, "{"),
			syntaxToken(SyntaxKindIdentifierToken, name),
			syntaxToken(SyntaxKindCloseBraceToken, "}"),
		),
		SymbolMatchAny,
	)
}

func nameReference(name string) *NameReference {
	return MustNewNameReferenceTyped(MustNewTokenNameTyped(syntaxToken(SyntaxKindIdentifierToken, name)), SymbolMatchAny)
}

func syntaxToken(kind SyntaxKind, text string) *SyntaxToken {
	return SyntaxTokenFromLexical(NewLexicalToken(kind, "", text))
}
