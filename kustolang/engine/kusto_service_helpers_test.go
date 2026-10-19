package kustolang

import (
	"strings"
	"testing"
)

func TestKustoServiceHelpersTokenAffinity(t *testing.T) {
	text := "Table | where Name == 'x'"
	code := Parse(text)

	tableStart := strings.Index(text, "Table")
	tableEnd := tableStart + len("Table")
	pipeStart := strings.Index(text, "|")
	whereStart := strings.Index(text, "where")
	nameStart := strings.Index(text, "Name")
	nameEnd := nameStart + len("Name")

	if token := GetTokenWithAffinity(code, tableStart); token == nil || token.RawText() != "Table" || token.TextStart() != tableStart {
		t.Fatalf("table-start affinity = %#v", token)
	}
	if token := GetTokenWithAffinity(code, tableEnd); token == nil || token.RawText() != "Table" || token.End() != tableEnd {
		t.Fatalf("table-end affinity = %#v", token)
	}
	if token := GetTokenWithAffinity(code, pipeStart); token != nil {
		t.Fatalf("pipe-start affinity = %#v, want nil", token)
	}
	if token := GetTokenWithAffinity(code, pipeStart+1); token != nil {
		t.Fatalf("post-pipe whitespace affinity = %#v, want nil", token)
	}
	if token := GetTokenWithAffinity(code, whereStart); token == nil || token.RawText() != "where" {
		t.Fatalf("keyword-start affinity = %#v", token)
	}
	if token := GetTokenWithAffinity(code, nameEnd); token == nil || token.RawText() != "Name" {
		t.Fatalf("name-end affinity = %#v", token)
	}
	if !IsNameLikeToken(SyntaxKindBangToken) || !IsNameLikeToken(SyntaxKindWhereKeyword) || !IsNameLikeToken(SyntaxKindIdentifierToken) {
		t.Fatalf("expected bang, keyword and identifier to be name-like")
	}
	if IsNameLikeToken(SyntaxKindBarToken) {
		t.Fatalf("bar token should not be name-like")
	}
}

func TestKustoServiceHelpersTokenLeftOfPosition(t *testing.T) {
	text := "Table | where Name"
	code := Parse(text)

	tableEnd := strings.Index(text, "Table") + len("Table")
	pipeStart := strings.Index(text, "|")
	whereStart := strings.Index(text, "where")
	nameStart := strings.Index(text, "Name")

	if token := GetTokenLeftOfPosition(code, tableEnd); token == nil || token.RawText() != "Table" {
		t.Fatalf("token left at table end = %#v", token)
	}
	if token := GetTokenLeftOfPosition(code, pipeStart); token == nil || token.RawText() != "Table" {
		t.Fatalf("token left at pipe start = %#v", token)
	}
	if token := GetTokenLeftOfPosition(code, whereStart); token == nil || token.RawText() != "|" {
		t.Fatalf("token left at where start = %#v", token)
	}
	if token := GetTokenLeftOfPosition(code, nameStart+2); token == nil || token.RawText() != "Name" {
		t.Fatalf("token left inside name = %#v", token)
	}
	if token := GetTokenLeftOfPosition(code, len(text)); token == nil || token.RawText() != "Name" {
		t.Fatalf("token left at end = %#v", token)
	}
}

func TestKustoServiceHelpersCompleteExpressionLeftOfPosition(t *testing.T) {
	code := Parse("count(User)")
	expr := GetCompleteExpressionLeftOfPosition(code, len(code.Text))
	if _, ok := expr.(*FunctionCallExpression); !ok {
		t.Fatalf("complete expression = %#v, want FunctionCallExpression", expr)
	}

	trailing := Parse("count(User)   ")
	expr = GetCompleteExpressionLeftOfPosition(trailing, len(trailing.Text))
	if _, ok := expr.(*FunctionCallExpression); !ok {
		t.Fatalf("complete trailing expression = %#v, want FunctionCallExpression", expr)
	}

	incomplete := Parse("count(")
	if expr := GetCompleteExpressionLeftOfPosition(incomplete, len(incomplete.Text)); expr != nil {
		t.Fatalf("incomplete expression = %#v, want nil", expr)
	}
}
