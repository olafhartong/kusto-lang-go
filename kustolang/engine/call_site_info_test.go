package kustolang

import (
	"strings"
	"testing"
)

func TestCallSiteInfoEqualsUsesSignatureAndParameterIdentity(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	limit := NewParameter("limit", ScalarTypes.Long.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), value, limit)
	NewFunctionSymbol("Resolve", signature)

	left := NewCallSiteInfo(signature, []*Parameter{value, limit}, []any{"alice", int64(1)})
	same := NewCallSiteInfo(signature, []*Parameter{value, limit}, []any{"alice", int64(1)})
	if !left.Equals(same) {
		t.Fatal("matching call-site info should be equal")
	}

	otherParameter := NewParameter("value", ScalarTypes.String.Type())
	if left.Equals(NewCallSiteInfo(signature, []*Parameter{otherParameter, limit}, []any{"alice", int64(1)})) {
		t.Fatal("equal parameter names should not replace parameter identity")
	}

	otherSignature := NewSignature(ScalarTypes.Bool.Type(), value, limit)
	NewFunctionSymbol("Resolve", otherSignature)
	if left.Equals(NewCallSiteInfo(otherSignature, []*Parameter{value, limit}, []any{"alice", int64(1)})) {
		t.Fatal("equal signatures by shape should not replace signature identity")
	}
}

func TestCallSiteInfoEqualsUsesTableResultEquivalenceForValues(t *testing.T) {
	tableParameter := NewParameterSymbolType("rows", NewTableSymbol("Rows"))
	signature := NewSignature(ScalarTypes.Bool.Type(), tableParameter)
	NewFunctionSymbol("ExpandRows", signature)

	first := NewTableSymbol("Events", NewColumnSymbol("Account", ScalarTypes.String.Type())).
		WithIsSerialized(true)
	second := NewTableSymbol("OtherName", NewColumnSymbol("Account", ScalarTypes.String.Type())).
		WithIsSerialized(true)
	differentResultState := second.WithIsSerialized(false)

	left := NewCallSiteInfo(signature, []*Parameter{tableParameter}, []any{first})
	if !left.Equals(NewCallSiteInfo(signature, []*Parameter{tableParameter}, []any{second})) {
		t.Fatal("table values should use result equivalence, not table name identity")
	}
	if left.Equals(NewCallSiteInfo(signature, []*Parameter{tableParameter}, []any{differentResultState})) {
		t.Fatal("table values with different result state should not be equal")
	}
}

func TestCallSiteInfoStringMirrorsUpstreamShape(t *testing.T) {
	value := NewParameter("value", ScalarTypes.String.Type())
	signature := NewSignature(ScalarTypes.Bool.Type(), value)
	NewFunctionSymbol("Resolve", signature)

	text := NewCallSiteInfo(signature, []*Parameter{value}, []any{"alice"}).String()
	if !strings.HasPrefix(text, "Resolve(") || !strings.Contains(text, "value=alice") {
		t.Fatalf("call-site string = %q", text)
	}
}
