package kustolang

import "testing"

func TestAnalyzerUtilityDbColumnAndOperandPredicates(t *testing.T) {
	column := NewColumnSymbol("Account", ScalarTypes.String.Type())
	table := NewTableSymbol("SigninLogs", column)
	globals := NewGlobalState(NewDatabaseSymbol("db", table))
	expr := ParseExpression(`Account has_any ("admin", "root")`).(*HasAnyExpression)
	account := firstNameReferenceInElement(t, expr, "Account")
	account.SetSemanticInfo(NewSemanticInfoForSymbol(column, column.Type))

	if !analyzerUtilityIsDbColumn(account, globals) {
		t.Fatal("Account should be recognized as a database column")
	}
	if !analyzerUtilityHasDbColumnOperand(expr, globals) {
		t.Fatal("has_any left operand should be recognized as a database column")
	}
	if analyzerUtilityGetFirstMatchingOperand(expr, func(candidate Expression) bool {
		return analyzerUtilityIsShortStringConstant(candidate, globals)
	}) != nil {
		t.Fatal("list operands should not be inspected by default")
	}
	firstShort := analyzerUtilityGetFirstMatchingOperand(expr, func(candidate Expression) bool {
		return analyzerUtilityIsShortStringConstant(candidate, globals, 6)
	}, true)
	if firstShort == nil || firstShort.Text(IncludeTriviaMinimal) != `"admin"` {
		t.Fatalf("first short list operand = %#v", firstShort)
	}
	if !analyzerUtilityHasAnyMatchingElement(expr.Right().Expressions(), func(candidate Expression) bool {
		return candidate.Text(IncludeTriviaMinimal) == `"root"`
	}) {
		t.Fatal("right expression list should contain root")
	}
}

func TestAnalyzerUtilityGetFirstMatchingOperandLexicalOrder(t *testing.T) {
	binary := ParseExpression("Left == Right").(*BinaryExpression)
	if got := analyzerUtilityGetFirstMatchingOperand(binary, func(candidate Expression) bool {
		return candidate.Text(IncludeTriviaMinimal) == "Left" || candidate.Text(IncludeTriviaMinimal) == "Right"
	}); got == nil || got.Text(IncludeTriviaMinimal) != "Left" {
		t.Fatalf("binary first operand = %#v", got)
	}
	prefix := ParseExpression("-Duration").(*PrefixUnaryExpression)
	if got := analyzerUtilityGetFirstMatchingOperand(prefix, func(candidate Expression) bool {
		return candidate.Text(IncludeTriviaMinimal) == "Duration"
	}); got == nil || got.Text(IncludeTriviaMinimal) != "Duration" {
		t.Fatalf("prefix operand = %#v", got)
	}
}

func TestAnalyzerUtilityRenameAndFunctionCallEditHelpers(t *testing.T) {
	name := ParseExpression("OldName").(*NameReference).Name()
	rename := analyzerUtilityAddRenameEdits(nil, name, 3, "NewName")
	if len(rename) != 1 || rename[0] != NewTextReplacement(3, len("OldName"), "NewName") {
		t.Fatalf("rename edits = %#v", rename)
	}

	call := ParseExpression("fn(a, b, c)").(*FunctionCallExpression)
	removal := analyzerUtilityAddRemoveArgumentEdits(nil, call, 0, 1)
	if got := NewEditString("fn(a, b, c)").ApplyAll(removal).CurrentText; got != "fn(a, c)" {
		t.Fatalf("remove argument result = %q edits=%#v", got, removal)
	}
	insert := analyzerUtilityAddInsertArgumentEdits(nil, call, 0, 1, "x")
	if got := NewEditString("fn(a, b, c)").ApplyAll(insert).CurrentText; got != "fn(a, x, b, c)" {
		t.Fatalf("insert argument result = %q edits=%#v", got, insert)
	}
	add := analyzerUtilityAddAddArgumentEdits(nil, ParseExpression("fn(a)").(*FunctionCallExpression), 0, "z")
	if got := NewEditString("fn(a)").ApplyAll(add).CurrentText; got != "fn(a, z)" {
		t.Fatalf("add argument result = %q edits=%#v", got, add)
	}
}

func TestAnalyzerUtilityGetFunctionCallAndRemoveOuterCall(t *testing.T) {
	call := ParseExpression("tolower(Account)").(*FunctionCallExpression)
	argument, positions, ok := analyzerUtilityFunctionCallArgumentPosition(call, 0, 0)
	if !ok {
		t.Fatal("argument position not found")
	}
	if got := analyzerUtilityGetFunctionCall(call, argument); got == nil || got.SyntaxNode != call.SyntaxNode {
		t.Fatalf("function call = %#v, want original", got)
	}
	edits := analyzerUtilityAddRemoveOuterFunctionCallEdits(nil, call, 0, argument, positions.ArgumentTriviaStart)
	if got := NewEditString("tolower(Account)").ApplyAll(edits).CurrentText; got != "Account" {
		t.Fatalf("remove outer call result = %q edits=%#v", got, edits)
	}
}

func firstNameReferenceInElement(t *testing.T, root SyntaxElement, name string) *NameReference {
	t.Helper()
	var reference *NameReference
	WalkSyntax(root, func(element SyntaxElement) {
		if reference != nil {
			return
		}
		if found, ok := WrapGeneratedSyntaxElement(element).(*NameReference); ok && NameReferenceSimpleName(found) == name {
			reference = found
		}
	})
	if reference == nil {
		t.Fatalf("name reference %q not found in %q", name, root.Text(IncludeTriviaAll))
	}
	return reference
}
