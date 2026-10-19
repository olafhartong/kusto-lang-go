package kustolang

import (
	"testing"
)

func firstPatternStatementInCode(t *testing.T, code *KustoCode) *PatternStatement {
	t.Helper()
	if code == nil || code.Query == nil || code.Query.Statements() == nil {
		t.Fatal("code has no query statements")
	}
	for _, element := range code.Query.Statements().Elements() {
		if element == nil {
			continue
		}
		if statement, ok := element.Element().(*PatternStatement); ok {
			return statement
		}
	}
	t.Fatal("pattern statement not found")
	return nil
}
