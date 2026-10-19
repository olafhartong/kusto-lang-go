package kustolang

import (
	"reflect"
	"testing"
)

func TestFunctionBodyFactsMirrorsUpstreamShape(t *testing.T) {
	first := NewParameter("first", ScalarTypes.String.Type())
	second := NewParameter("second", ScalarTypes.Long.Type())

	facts := FunctionBodyFactsDefault.
		WithNonVariableReturnType(ScalarTypes.String.Type()).
		WithHasSyntaxErrors(true).
		AddDependentParameter(first).
		AddDependentParameter(first).
		AddDependentParameters([]*Parameter{first, second}).
		AddUnqualifiedTableName("SigninLogs").
		AddUnqualifiedTableName("SigninLogs").
		WithHasClusterCall(true).
		WithHasGraphCall(true)

	if facts == FunctionBodyFactsDefault {
		t.Fatal("FunctionBodyFacts mutators should return a new value when facts change")
	}
	if FunctionBodyFactsDefault.NonVariableReturnType() != nil || FunctionBodyFactsDefault.HasSyntaxErrors() || FunctionBodyFactsDefault.HasVariableReturnType() {
		t.Fatalf("default facts mutated: %#v", FunctionBodyFactsDefault)
	}
	if facts.NonVariableReturnType() != ScalarTypes.String.Type() || !facts.HasSyntaxErrors() {
		t.Fatalf("facts scalar fields not preserved: %#v", facts)
	}
	if len(facts.DependentParameters()) != 2 || facts.DependentParameters()[0] != first || facts.DependentParameters()[1] != second {
		t.Fatalf("dependent parameters = %#v", facts.DependentParameters())
	}
	if len(facts.UnqualifiedTableNames()) != 1 || facts.UnqualifiedTableNames()[0] != "SigninLogs" {
		t.Fatalf("unqualified names = %#v", facts.UnqualifiedTableNames())
	}
	if !facts.HasVariableReturnType() || !facts.HasClusterCall() || !facts.HasGraphCall() || !facts.IsInteresting() {
		t.Fatalf("facts flags not preserved: %#v", facts)
	}

	called := FunctionBodyFactsDefault.
		WithHasDatabaseCall(true).
		WithHasUnqualifiedTableCall(true).
		AddUnqualifiedTableName("SecurityEvent")
	combined := facts.CombineCalledFunction(called)
	if !combined.HasClusterCall() || !combined.HasDatabaseCall() || !combined.HasUnqualifiedTableCall() || !combined.HasVariableReturnType() {
		t.Fatalf("combined facts flags not preserved: %#v", combined)
	}
	if len(combined.UnqualifiedTableNames()) != 2 || combined.UnqualifiedTableNames()[0] != "SigninLogs" || combined.UnqualifiedTableNames()[1] != "SecurityEvent" {
		t.Fatalf("combined table names = %#v", combined.UnqualifiedTableNames())
	}
}

func TestFunctionCallInfoLazyExpansionAndDiagnostics(t *testing.T) {
	diagnostic := NewDiagnostic("KS999", "bad expansion")
	root := NewSyntaxNode(SyntaxKindNameReference, nil, diagnostic)
	expansion := NewFunctionCallExpansion(root, nil, 0)
	calls := 0
	info := NewDeferredFunctionCallInfo(func() *FunctionCallExpansion {
		calls++
		return expansion
	}, FunctionBodyFactsDefault)

	if info.Expansion() != expansion || info.Expansion() != expansion {
		t.Fatal("deferred expansion should return the cached expansion")
	}
	if calls != 1 {
		t.Fatalf("expander calls = %d, want 1", calls)
	}
	diagnostics := info.Diagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS999" {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !info.HasErrors() {
		t.Fatal("function call info should report errors from expansion diagnostics")
	}

	semanticRoot := NewSyntaxNode(SyntaxKindNameReference, nil)
	semanticRoot.SetSemanticInfo(NewSemanticInfoWithDiagnostics(NewDiagnostic("KS998", "semantic expansion")))
	semanticInfo := NewFunctionCallInfo(NewFunctionCallExpansion(semanticRoot, nil, 0), FunctionBodyFactsDefault)
	if diagnostics := semanticInfo.Diagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS998" {
		t.Fatalf("semantic expansion diagnostics = %#v", diagnostics)
	}

	original := NewSyntaxTree(NewSyntaxNode(SyntaxKindQueryBlock, nil), nil, 0)
	locatedRoot := NewSyntaxNode(SyntaxKindNameReference, nil, NewDiagnostic("KS997", "located expansion").WithLocation(3, 4))
	locatedInfo := NewFunctionCallInfo(NewFunctionCallExpansion(locatedRoot, original, 17), FunctionBodyFactsDefault)
	if diagnostics := locatedInfo.Diagnostics(); len(diagnostics) != 1 || diagnostics[0].Start != 20 || diagnostics[0].Length != 4 || diagnostics[0].LocationKind != DiagnosticLocationAbsolute {
		t.Fatalf("relocated expansion diagnostics = %#v", diagnostics)
	}

	syntaxErrorInfo := NewFunctionCallInfo(nil, FunctionBodyFactsDefault.WithHasSyntaxErrors(true))
	if !syntaxErrorInfo.HasErrors() {
		t.Fatal("function call info should report errors from function body facts")
	}
}

func TestSemanticInfoCarriesFunctionCallInfoAndAlternates(t *testing.T) {
	info := NewSemanticInfoForType(ScalarTypes.String.Type())
	callInfo := NewFunctionCallInfo(nil, FunctionBodyFactsDefault.WithHasGraphCall(true))
	alternate := NewSyntaxNode(SyntaxKindNameReference, nil)

	updated := info.WithCalledFunctionInfo(callInfo).WithAlternates([]*SyntaxNode{alternate})
	if updated.CalledFunctionInfo != callInfo {
		t.Fatalf("called function info = %#v, want original", updated.CalledFunctionInfo)
	}
	if len(updated.Alternates) != 1 || updated.Alternates[0] != alternate {
		t.Fatalf("alternates = %#v", updated.Alternates)
	}
	if info.CalledFunctionInfo != nil || len(info.Alternates) != 0 {
		t.Fatalf("original semantic info mutated: %#v", info)
	}
}

func TestSemanticInfoReferencedSignatureMirrorsUpstreamCases(t *testing.T) {
	signature := NewSignature(ScalarTypes.String.Type())
	fn := NewFunctionSymbol("single", signature)
	info := NewSemanticInfoForSymbol(fn, &fn.TypeSymbol)
	if info.ReferencedSignature() != signature {
		t.Fatalf("single function referenced signature = %#v", info.ReferencedSignature())
	}

	variable := NewVariableSymbol("callable", fn, false, nil, nil)
	info = NewSemanticInfoForSymbol(variable, &fn.TypeSymbol)
	if info.ReferencedSignature() != signature {
		t.Fatalf("function-valued variable referenced signature = %#v", info.ReferencedSignature())
	}

	entityGroup := NewEntityGroupSymbolFromDefinition("Group", "database('d')", "")
	info = NewSemanticInfoForSymbol(entityGroup, &entityGroup.TypeSymbol)
	if info.ReferencedSignature() != entityGroup.Signature() {
		t.Fatalf("entity-group referenced signature = %#v", info.ReferencedSignature())
	}

	overloaded := NewFunctionSymbol("overloaded", NewSignature(ScalarTypes.String.Type()), NewSignature(ScalarTypes.Long.Type()))
	info = NewSemanticInfoForSymbol(overloaded, &overloaded.TypeSymbol)
	if info.ReferencedSignature() != nil {
		t.Fatalf("overloaded function referenced signature = %#v, want nil", info.ReferencedSignature())
	}
}

func TestSyntaxNodeSemanticAccessorsAndDiagnosticsInclude(t *testing.T) {
	syntaxDiagnostic := NewDiagnostic("KS900", "syntax")
	semanticDiagnostic := NewDiagnostic("KS901", "semantic")
	expansionDiagnostic := NewDiagnostic("KS902", "expansion")
	alternateDiagnostic := NewDiagnostic("KS903", "alternate")

	signature := NewSignature(ScalarTypes.String.Type())
	fn := NewFunctionSymbol("f", signature)
	calledBody := NewSyntaxNode(SyntaxKindNameReference, nil, expansionDiagnostic)
	callInfo := NewFunctionCallInfo(
		NewFunctionCallExpansion(calledBody, nil, 0),
		FunctionBodyFactsDefault.WithHasGraphCall(true),
	)
	alternate := NewSyntaxNode(SyntaxKindNameReference, nil, alternateDiagnostic)
	node := MustNewNameReferenceTyped(
		MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "f"))),
		SymbolMatchAny,
		syntaxDiagnostic,
	)
	node.SetSemanticInfo(
		NewSemanticInfoForSymbol(fn, ScalarTypes.String.Type(), semanticDiagnostic).
			WithIsConstant(true).
			WithCalledFunctionInfo(callInfo).
			WithAlternates([]*SyntaxNode{alternate}),
	)

	if !node.IsBound() {
		t.Fatal("node should report bound after semantic info is attached")
	}
	if info, ok := node.GetSemanticInfo(); !ok || info.ReferencedSymbol() != fn {
		t.Fatalf("semantic info = %#v/%t", info, ok)
	}
	if node.ReferencedSymbol() != fn || node.ReferencedSignature() != signature {
		t.Fatalf("referenced symbol/signature = %#v/%#v", node.ReferencedSymbol(), node.ReferencedSignature())
	}
	if node.RawResultType() != ScalarTypes.String.Type() || !node.IsConstant() {
		t.Fatalf("raw result/constant = %#v/%t", node.RawResultType(), node.IsConstant())
	}
	if node.GetCalledFunctionBody() != calledBody || node.GetCalledFunctionFacts() != callInfo.Facts || !node.CalledFunctionHasErrors() {
		t.Fatalf("called function info not preserved")
	}
	if diagnostics := node.GetCalledFunctionDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS902" {
		t.Fatalf("called diagnostics = %#v", diagnostics)
	}
	if alternates := node.Alternates(); len(alternates) != 1 || alternates[0] != alternate {
		t.Fatalf("alternates = %#v", alternates)
	}
	if diagnostics := node.SemanticDiagnostics(); len(diagnostics) != 1 || diagnostics[0].Code != "KS901" {
		t.Fatalf("semantic diagnostics = %#v", diagnostics)
	}

	syntactic := GetContainedDiagnostics(node, DiagnosticsIncludeSyntactic)
	if !diagnosticsContain(syntactic, "KS900") || diagnosticsContain(syntactic, "KS901") || diagnosticsContain(syntactic, "KS902") {
		t.Fatalf("syntactic diagnostics = %#v", syntactic)
	}
	semantic := GetContainedDiagnostics(node, DiagnosticsIncludeDefault)
	if !diagnosticsContain(semantic, "KS900") || !diagnosticsContain(semantic, "KS901") || !diagnosticsContain(semantic, "KS903") || diagnosticsContain(semantic, "KS902") {
		t.Fatalf("default diagnostics = %#v", semantic)
	}
	if got := semanticInfoDiagnosticCodes(semantic); !reflect.DeepEqual(got, []string{"KS900", "KS901", "KS903"}) {
		t.Fatalf("default diagnostic order = %#v", got)
	}
	withExpansion := GetContainedDiagnostics(node, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion)
	if !diagnosticsContain(withExpansion, "KS200") || diagnosticsContain(withExpansion, "KS902") {
		t.Fatalf("diagnostics with expansion = %#v", withExpansion)
	}
	if got := semanticInfoDiagnosticCodes(withExpansion); !reflect.DeepEqual(got, []string{"KS900", "KS901", "KS200", "KS903"}) {
		t.Fatalf("expansion diagnostic order = %#v", got)
	}
	if diagnostics := GetContainedDiagnostics(node, DiagnosticsInclude(0)); len(diagnostics) != 0 {
		t.Fatalf("zero include diagnostics = %#v", diagnostics)
	}
}

func semanticInfoDiagnosticCodes(diagnostics []Diagnostic) []string {
	codes := make([]string, len(diagnostics))
	for index, diagnostic := range diagnostics {
		codes[index] = diagnostic.Code
	}
	return codes
}

func TestContainedDiagnosticsAssignsElementLocations(t *testing.T) {
	first := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a"))
	relative := SyntaxTokenFromLexical(
		NewLexicalToken(SyntaxKindIdentifierToken, "  ", "b"),
		NewDiagnostic("KS910", "relative"),
	)
	relativeEnd := SyntaxTokenFromLexical(
		NewLexicalToken(SyntaxKindIdentifierToken, "", "c"),
		NewDiagnostic("KS911", "relative end").WithLocationKind(DiagnosticLocationRelativeEnd),
	)
	root := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{first, relative, relativeEnd})

	diagnostics := GetContainedDiagnostics(root, DiagnosticsIncludeSyntactic)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Code != "KS910" || diagnostics[0].Start != 3 || diagnostics[0].Length != 1 || diagnostics[0].LocationKind != DiagnosticLocationAbsolute {
		t.Fatalf("relative diagnostic location = %#v", diagnostics[0])
	}
	if diagnostics[1].Code != "KS911" || diagnostics[1].Start != 5 || diagnostics[1].Length != 0 || diagnostics[1].LocationKind != DiagnosticLocationAbsolute {
		t.Fatalf("relative-end diagnostic location = %#v", diagnostics[1])
	}
	direct := root.GetContainedSyntaxDiagnostics()
	if !reflect.DeepEqual(direct, diagnostics) {
		t.Fatalf("direct syntax diagnostics = %#v, want %#v", direct, diagnostics)
	}

	treeDiagnostics := NewSyntaxTree(root, nil, 0).Diagnostics()
	if len(treeDiagnostics) != 2 || treeDiagnostics[0].Start != diagnostics[0].Start || treeDiagnostics[1].Start != diagnostics[1].Start {
		t.Fatalf("tree diagnostics = %#v", treeDiagnostics)
	}
}

func TestContainedDiagnosticsAnchorMissingTokenChainToNextRealToken(t *testing.T) {
	first := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "a"))
	missingClose := NewMissingToken(SyntaxKindCloseBraceToken, DiagnosticTermsExpected("}"))
	missingSeparator := NewMissingToken(SyntaxKindSemicolonToken, DiagnosticTermsExpected(";"))
	next := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "b"))
	root := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{first, missingClose, missingSeparator, next})

	diagnostics := GetContainedDiagnostics(root, DiagnosticsIncludeSyntactic)
	if len(diagnostics) != 2 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, diagnostic := range diagnostics {
		if diagnostic.Start != 1 || diagnostic.Length != 1 || diagnostic.LocationKind != DiagnosticLocationAbsolute {
			t.Fatalf("missing-token diagnostic location = %#v", diagnostic)
		}
	}
}

func TestSemanticInfoForReferencedSymbolMirrorsUpstreamKinds(t *testing.T) {
	scalarParameter := NewParameterSymbol("tenant", ScalarTypes.String.Type(), "")
	scalarInfo := NewSemanticInfoForReferencedSymbol(scalarParameter)
	if scalarInfo.ReferencedSymbol() != scalarParameter || scalarInfo.ResultType != ScalarTypes.String.Type() || !scalarInfo.IsConstant {
		t.Fatalf("scalar parameter semantic info = %#v", scalarInfo)
	}

	table := NewTableSymbol("Rows", NewColumnSymbol("Account", ScalarTypes.String.Type()))
	tableParameter := NewParameterSymbol("rows", table, "")
	tableInfo := NewSemanticInfoForReferencedSymbol(tableParameter)
	if tableInfo.ReferencedSymbol() != tableParameter || tableInfo.ResultType != &table.TypeSymbol || tableInfo.IsConstant {
		t.Fatalf("table parameter semantic info = %#v", tableInfo)
	}

	variable := NewVariableSymbol("limit", ScalarTypes.Long.Type(), true, nil, nil)
	variableInfo := NewSemanticInfoForReferencedSymbol(variable)
	if variableInfo.ReferencedSymbol() != variable || variableInfo.ResultType != ScalarTypes.Long.Type() || !variableInfo.IsConstant {
		t.Fatalf("variable semantic info = %#v", variableInfo)
	}

	typeInfo := NewSemanticInfoForReferencedSymbol(ScalarTypes.Bool)
	if typeInfo.ReferencedSymbol() != nil || typeInfo.ResultType != ScalarTypes.Bool.Type() {
		t.Fatalf("primitive type semantic info = %#v", typeInfo)
	}
}

func TestContainedDiagnosticsRelocatesEmptyTokenToNearbyNextToken(t *testing.T) {
	missing := NewMissingToken(SyntaxKindIdentifierToken, NewDiagnostic("KS912", "missing"))
	next := SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "next"))
	root := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{missing, next})

	diagnostics := GetContainedDiagnostics(root, DiagnosticsIncludeSyntactic)
	if len(diagnostics) != 1 || diagnostics[0].Start != 1 || diagnostics[0].Length != 4 {
		t.Fatalf("nearby next-token diagnostic = %#v", diagnostics)
	}

	twoSpaces := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{
		NewMissingToken(SyntaxKindIdentifierToken, NewDiagnostic("KS913", "missing")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "  ", "next")),
	})
	if diagnostics := GetContainedDiagnostics(twoSpaces, DiagnosticsIncludeSyntactic); len(diagnostics) != 1 || diagnostics[0].Start != 0 || diagnostics[0].Length != 0 {
		t.Fatalf("two-space next-token diagnostic = %#v", diagnostics)
	}

	lineBreak := NewSyntaxNode(SyntaxKindNameReference, []SyntaxElement{
		NewMissingToken(SyntaxKindIdentifierToken, NewDiagnostic("KS914", "missing")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "\n", "next")),
	})
	if diagnostics := GetContainedDiagnostics(lineBreak, DiagnosticsIncludeSyntactic); len(diagnostics) != 1 || diagnostics[0].Start != 0 || diagnostics[0].Length != 0 {
		t.Fatalf("line-break next-token diagnostic = %#v", diagnostics)
	}
}

func TestContainedDiagnosticsLocatesExpansionSummaryOnCallName(t *testing.T) {
	fn := NewFunctionSymbol("f", NewSignature(ScalarTypes.String.Type()))
	nameReference := MustNewNameReferenceTyped(
		MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, " ", "f"))),
		SymbolMatchFunction,
	)
	argumentList := MustNewExpressionListTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenParenToken, "", "(")),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseParenToken, "", ")")),
	)
	call := MustNewFunctionCallExpressionTyped(nameReference, argumentList)
	call.SetSemanticInfo(
		NewSemanticInfoForSymbol(fn, ScalarTypes.String.Type()).
			WithCalledFunctionInfo(NewFunctionCallInfo(
				NewFunctionCallExpansion(NewSyntaxNode(SyntaxKindNameReference, nil, NewDiagnostic("KS912", "expanded bad")), nil, 0),
				FunctionBodyFactsDefault,
			)),
	)
	root := NewSyntaxNode(SyntaxKindQueryBlock, []SyntaxElement{
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "T")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindBarToken, "", "|")),
		call,
	})

	diagnostics := GetContainedDiagnostics(root, DiagnosticsIncludeDefault|DiagnosticsIncludeExpansion)
	if len(diagnostics) != 1 || diagnostics[0].Code != "KS200" {
		t.Fatalf("expansion diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Start != 3 || diagnostics[0].Length != 1 || diagnostics[0].LocationKind != DiagnosticLocationAbsolute {
		t.Fatalf("expansion diagnostic location = %#v", diagnostics[0])
	}
}

func TestExpressionSemanticHelpersMirrorUpstreamConveniences(t *testing.T) {
	stringColumn := NewColumnSymbol("Only", ScalarTypes.String.Type())
	tuple := NewTupleSymbolOf(stringColumn)
	tupleVariable := NewVariableSymbol("tupleValue", tuple, false, nil, nil)
	tupleExpr := MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "tupleValue"))), SymbolMatchAny)
	tupleExpr.SetSemanticInfo(NewSemanticInfoForSymbol(tupleVariable, tuple.Type()))

	if got := ExpressionRawResultType(tupleExpr); got != tuple.Type() {
		t.Fatalf("raw tuple result = %s, want tuple", typeName(got))
	}
	if got := ExpressionResultType(tupleExpr); got != ScalarTypes.String.Type() {
		t.Fatalf("adjusted tuple result = %s, want string", typeName(got))
	}

	constVariable := NewVariableSymbol("Limit", ScalarTypes.Long.Type(), true, int64(42), nil)
	constExpr := MustNewNameReferenceTyped(MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "Limit"))), SymbolMatchAny)
	constExpr.SetSemanticInfo(NewSemanticInfoForSymbol(constVariable, ScalarTypes.Long.Type()).WithIsConstant(true))
	if info := ExpressionConstantValueInfo(constExpr); !ExpressionIsConstant(constExpr) || info == nil || info.Value != int64(42) || ExpressionConstantValue(constExpr) != int64(42) {
		t.Fatalf("constant variable helpers = const:%t info:%#v value:%#v", ExpressionIsConstant(constExpr), ExpressionConstantValueInfo(constExpr), ExpressionConstantValue(constExpr))
	}

	literal := MustNewLiteralExpressionTyped(
		SyntaxKindStringLiteralExpression,
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'abc'")),
	)
	literalInfo := ExpressionLiteralValueInfo(literal)
	if !ExpressionIsLiteral(literal) || literalInfo == nil || literalInfo.RawText != "'abc'" || literalInfo.ValueText != "abc" || literalInfo.Value != "abc" {
		t.Fatalf("literal info = %#v literal:%t", literalInfo, ExpressionIsLiteral(literal))
	}
	if info := ExpressionConstantValueInfo(literal); info == nil || info.RawText != literalInfo.RawText || info.ValueText != literalInfo.ValueText || info.Value != literalInfo.Value {
		t.Fatalf("literal constant info = %#v, want literal ValueInfo", ExpressionConstantValueInfo(literal))
	}
	if got := ExpressionConstantValue(literal); got != "abc" {
		t.Fatalf("literal constant value = %#v, want abc", got)
	}

	compound := MustNewCompoundStringLiteralExpressionTyped(
		AsSyntaxListOf[*SyntaxToken](NewTypedSyntaxList("SyntaxToken",
			SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'a'")),
			SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'b'")),
		)),
	)
	if got := ExpressionConstantValue(compound); got != "ab" {
		t.Fatalf("compound literal constant value = %#v, want ab", got)
	}
	if info := ExpressionLiteralValueInfo(compound); info == nil || info.RawText != "'a''b'" || info.ValueText != "ab" || info.Value != "ab" {
		t.Fatalf("compound literal info = %#v", info)
	}

	dynamicLiteral := MustNewDynamicExpressionTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindDynamicKeyword, "", "dynamic")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenParenToken, "", "(")),
		literal,
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseParenToken, "", ")")),
	)
	if info := ExpressionLiteralValueInfo(dynamicLiteral); info == nil || info.RawText != "dynamic('abc')" || info.ValueText != "'abc'" || info.Value != "'abc'" {
		t.Fatalf("dynamic literal info = %#v", info)
	}

	typeOfLiteral := MustNewTypeOfLiteralExpressionTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindTypeOfKeyword, "", "typeof")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenParenToken, "", "(")),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseParenToken, "", ")")),
	)
	if !ExpressionIsLiteral(typeOfLiteral) || ExpressionLiteralValueInfo(typeOfLiteral) != nil || ExpressionLiteralValue(typeOfLiteral) != nil {
		t.Fatalf("typeof literal state = literal:%t info:%#v value:%#v", ExpressionIsLiteral(typeOfLiteral), ExpressionLiteralValueInfo(typeOfLiteral), ExpressionLiteralValue(typeOfLiteral))
	}
}

func TestExpressionInstanceSemanticsPreserveGeneratedWrapperShape(t *testing.T) {
	stringColumn := NewColumnSymbol("Only", ScalarTypes.String.Type())
	tuple := NewTupleSymbolOf(stringColumn)
	tupleVariable := NewVariableSymbol("tupleValue", tuple, false, nil, nil)
	tupleReference := MustNewNameReferenceTyped(
		MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", "tupleValue"))),
		SymbolMatchAny,
	)
	tupleReference.SetSemanticInfo(NewSemanticInfoForSymbol(tupleVariable, tuple.Type()))
	var scalarReduction Expression = tupleReference
	if scalarReduction.RawResultType() != tuple.Type() || scalarReduction.ResultType() != ScalarTypes.String.Type() {
		t.Fatalf("instance tuple result = raw:%s adjusted:%s", typeName(scalarReduction.RawResultType()), typeName(scalarReduction.ResultType()))
	}

	var literal Expression = MustNewLiteralExpressionTyped(
		SyntaxKindStringLiteralExpression,
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindStringLiteralToken, "", "'abc'")),
	)
	var literalInfo *ValueInfo = literal.LiteralValueInfo()
	var constantInfo *ValueInfo = literal.ConstantValueInfo()
	if !literal.IsLiteral() || literalInfo == nil || literalInfo.RawText != "'abc'" || literalInfo.ValueText != "abc" || literalInfo.Value != "abc" {
		t.Fatalf("instance literal info = %#v literal:%t", literalInfo, literal.IsLiteral())
	}
	if constantInfo == nil || constantInfo.Value != "abc" || literal.LiteralValue() != "abc" || literal.ConstantValue() != "abc" {
		t.Fatalf("instance literal values = constant:%#v literal:%#v", constantInfo, literal.LiteralValue())
	}

	var typeOfLiteral Expression = MustNewTypeOfLiteralExpressionTyped(
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindTypeOfKeyword, "", "typeof")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindOpenParenToken, "", "(")),
		AsSeparatedSyntaxListOf[Expression](NewTypedSyntaxList("SeparatedElement<Expression>")),
		SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindCloseParenToken, "", ")")),
	)
	if !typeOfLiteral.IsLiteral() || typeOfLiteral.LiteralValueInfo() != nil || typeOfLiteral.LiteralValue() != nil || typeOfLiteral.ConstantValueInfo() != nil || typeOfLiteral.ConstantValue() != nil {
		t.Fatalf("instance typeof literal = literal:%t info:%#v value:%#v constant:%#v", typeOfLiteral.IsLiteral(), typeOfLiteral.LiteralValueInfo(), typeOfLiteral.LiteralValue(), typeOfLiteral.ConstantValueInfo())
	}

	var dynamicNull Expression = ParseLiteral("dynamic(null)")
	dynamicInfo := dynamicNull.LiteralValueInfo()
	if !dynamicNull.IsLiteral() || dynamicInfo == nil || dynamicInfo.RawText != "dynamic(null)" || dynamicInfo.ValueText != "null" || dynamicInfo.Value != nil {
		t.Fatalf("instance dynamic null info = %#v literal:%t", dynamicInfo, dynamicNull.IsLiteral())
	}
	if constant := dynamicNull.ConstantValueInfo(); constant == nil || constant.Value != nil || dynamicNull.ConstantValue() != nil {
		t.Fatalf("instance dynamic null constant = %#v value:%#v", constant, dynamicNull.ConstantValue())
	}

	var nonliteral Expression = ParseExpression("value + 1")
	if nonliteral.IsLiteral() || nonliteral.IsConstant() || nonliteral.LiteralValueInfo() != nil || nonliteral.LiteralValue() != nil || nonliteral.ConstantValueInfo() != nil || nonliteral.ConstantValue() != nil {
		t.Fatalf("instance nonliteral state = literal:%t constant:%t literal-info:%#v constant-info:%#v", nonliteral.IsLiteral(), nonliteral.IsConstant(), nonliteral.LiteralValueInfo(), nonliteral.ConstantValueInfo())
	}
}
