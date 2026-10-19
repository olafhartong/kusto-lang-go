package kustolang

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestSyntaxNodeCompactMatchPayload(t *testing.T) {
	node := NewSyntaxNode(SyntaxKindNameReference, nil)
	if _, ok := node.Payload("Match"); ok {
		t.Fatal("absent match reported present")
	}
	for _, match := range []SymbolMatch{SymbolMatchNone, SymbolMatchColumn, SymbolMatchAny, ^SymbolMatch(0)} {
		node.setPayload("Match", match)
		if value, ok := node.Payload("Match"); !ok || value != match {
			t.Fatalf("match = %v, %t", value, ok)
		}
		if node.payloads != nil {
			t.Fatal("match allocated a payload map")
		}
		clone := node.Clone(true).(*SyntaxNode)
		node.setPayload("Match", SymbolMatchTable)
		if value, ok := clone.Payload("Match"); !ok || value != match {
			t.Fatal("cloned match changed")
		}
	}
	// Preserve arbitrary payload types, including a present nil value.
	for _, value := range []any{"custom", nil, []string{"original"}} {
		node.setPayload("Match", value)
		if got, ok := node.Payload("Match"); !ok || !reflect.DeepEqual(got, value) {
			t.Fatalf("fallback payload = %#v, %t", got, ok)
		}
		node.setPayload("Match", SymbolMatchNone)
		if _, exists := node.payloads["Match"]; exists {
			t.Fatal("stale fallback match retained")
		}
	}
	node.setPayload("CommandKinds", []string{"one"})
	clone := node.Clone(false).(*SyntaxNode)
	node.setPayload("CommandKinds", []string{"two"})
	value, _ := clone.Payload("CommandKinds")
	value.([]string)[0] = "changed"
	if got, _ := clone.Payload("CommandKinds"); !reflect.DeepEqual(got, []string{"one"}) {
		t.Fatal("payload slice ownership changed")
	}
	parsed := ParseExpression("x").(*NameReference)
	if got := parsed.CloneTyped(true).Match(); got != parsed.Match() {
		t.Fatal("typed clone lost match")
	}
}

func TestSyntaxNodeCompactMatchDoesNotGrow64BitNodes(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("padding optimization targets 64-bit layouts")
	}
	// Original layout, before compacting Match into the padding after kind.
	type originalNode struct {
		syntaxOwnership
		kind              SyntaxKind
		generatedNodeName string
		children          []SyntaxElement
		diagnostics       []Diagnostic
		payloads          map[string]any
	}
	got, want := unsafe.Sizeof(SyntaxNode{}), unsafe.Sizeof(originalNode{})
	t.Logf("SyntaxNode size: before %d bytes, after %d bytes", want, got)
	if got != want {
		t.Fatalf("node size = %d, want %d", got, want)
	}
}
