package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicOperatorSymbolDirectResultConstructor(t *testing.T) {
	resultType := kustolang.ScalarTypes.Long.Type()
	operator := kustolang.NewOperatorSymbolWithResult(kustolang.OperatorKindAdd, resultType)
	signatures := operator.Signatures()

	if operator.Result != resultType || len(signatures) != 1 || signatures[0].DeclaredReturnType != resultType {
		t.Fatalf("direct-result operator = result:%#v signatures:%#v", operator.Result, signatures)
	}
	if signatures[0].Symbol != operator {
		t.Fatalf("signature owner = %#v, want operator %#v", signatures[0].Symbol, operator)
	}
}
