package kustolang

import "testing"

func TestQueryParserQueryBlockPreservesTypedStatementPrefix(t *testing.T) {
	tests := []struct {
		name              string
		text              string
		wantStatements    int
		wantSkipped       string
		wantFirstExplicit bool
	}{
		{
			name:              "valid multiple statements",
			text:              "print x=1; print y=2",
			wantStatements:    2,
			wantFirstExplicit: true,
		},
		{
			name:           "missing semicolon",
			text:           "print x=1 print y=2",
			wantStatements: 2,
		},
		{
			name:           "incomplete where",
			text:           "print x=1 | where",
			wantStatements: 1,
		},
		{
			name:           "incomplete project comma tail",
			text:           "print x=1 | project x,",
			wantStatements: 1,
		},
		{
			name:           "unknown suffix",
			text:           "print x=1 ???",
			wantStatements: 1,
			wantSkipped:    " ???",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := ParseQuery(tt.text)
			statements, ok := block.GetChild(1).(*SyntaxList)
			if !ok || statements.Count() != tt.wantStatements {
				t.Fatalf("statements = %T count %d, want %d", block.GetChild(1), syntaxListCount(statements), tt.wantStatements)
			}

			for i := 0; i < statements.Count(); i++ {
				separated, ok := statements.At(i).(*SeparatedElement)
				if !ok {
					t.Fatalf("statement %d = %T, want *SeparatedElement", i, statements.At(i))
				}
				statement, ok := WrapGeneratedSyntaxElement(separated.Element()).(*ExpressionStatement)
				if !ok {
					t.Fatalf("statement %d = %T, want *ExpressionStatement", i, separated.Element())
				}
				expression := statement.Expression()
				if expression == nil {
					t.Fatalf("statement %d expression is nil, want typed expression", i)
				}
				if expression.Kind() == SyntaxKindNone {
					t.Fatalf("statement %d expression = %T kind %v, want typed expression", i, expression, expression.Kind())
				}
			}

			first := statements.At(0).(*SeparatedElement)
			if tt.wantStatements == 2 {
				separator := first.Separator()
				if separator == nil || separator.Kind() != SyntaxKindSemicolonToken {
					t.Fatalf("first separator = %#v, want semicolon token", separator)
				}
				if gotExplicit := separator.Width() > 0; gotExplicit != tt.wantFirstExplicit {
					t.Fatalf("first separator explicit = %v, want %v", gotExplicit, tt.wantFirstExplicit)
				}
			}

			skipped, _ := WrapGeneratedSyntaxElement(block.GetChild(2)).(*SkippedTokens)
			if tt.wantSkipped == "" {
				if skipped != nil {
					t.Fatalf("skipped tokens = %q, want none", skipped.Text(IncludeTriviaAll))
				}
				return
			}
			if skipped == nil || skipped.Text(IncludeTriviaAll) != tt.wantSkipped {
				got := ""
				if skipped != nil {
					got = skipped.Text(IncludeTriviaAll)
				}
				t.Fatalf("skipped tokens = %q, want %q", got, tt.wantSkipped)
			}
			if firstSkipped := skipped.Tokens().At(0); !queryParserPrefixDiagnosticsContain(firstSkipped.Diagnostics(), "KS198") {
				t.Fatalf("first skipped token diagnostics = %#v, want KS198", firstSkipped.Diagnostics())
			}
		})
	}
}

func queryParserPrefixDiagnosticsContain(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func syntaxListCount(list *SyntaxList) int {
	if list == nil {
		return 0
	}
	return list.Count()
}
