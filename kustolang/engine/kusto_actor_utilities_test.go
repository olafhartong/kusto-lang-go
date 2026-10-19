package kustolang

import (
	"strings"
	"testing"
)

func TestActorUtilitiesRecognizeEntityExpressionsAndInsertionPoint(t *testing.T) {
	table := NewTableSymbol("T", NewColumnSymbol("X", ScalarTypes.Long.Type()))
	database := NewDatabaseSymbol("db", table)
	globals := NewGlobalState(database)
	code := ParseAndAnalyze("database('db').table('T') | where X > 0", globals)
	position := strings.LastIndex(code.Text, "table") + 1
	token := actorTokenNear(code, position)
	entity := actorGetEntityExpression(token, globals)
	if token == nil || entity == nil || entity.Text(IncludeTriviaInterior) != "database('db').table('T')" {
		t.Fatalf("entity = token:%#v expression:%#v text:%q", token, entity, func() string {
			if entity == nil {
				return ""
			}
			return entity.Text(IncludeTriviaInterior)
		}())
	}
	if insertion, ok := actorNearestTopLevelStatementInsertionPosition(code, position); !ok || insertion != 0 {
		t.Fatalf("insertion = %d, %t", insertion, ok)
	}
}

func TestActorUtilitiesAdjustRangeAndInsertBrackets(t *testing.T) {
	code := ParseAndAnalyze("print x=1 + 2 ")
	start := strings.Index(code.Text, "1")
	adjustedStart, adjustedLength := actorAdjustRangeToNode(code, start, len("1 + 2"))
	if adjustedStart != start || code.Text[adjustedStart:adjustedStart+adjustedLength] != "1 + 2" {
		t.Fatalf("adjusted range = %d:%d %q", adjustedStart, adjustedLength, code.Text[adjustedStart:adjustedStart+adjustedLength])
	}

	original := NewEditString("T\n| where X > 0")
	wrapped := actorInsertBrackets(original, 0, "macro-expand entity_group [T] as _scope", "(", len(original.CurrentText), ")", FormattingOptionsDefault.WithIndentationSize(2).WithBrackettingStyle(BrackettingStyleDiagonal))
	want := "macro-expand entity_group [T] as _scope (\n  T\n  | where X > 0\n)\n"
	if wrapped.CurrentText != want {
		t.Fatalf("wrapped text = %q, want %q", wrapped.CurrentText, want)
	}
}
