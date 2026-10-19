package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicExpressionExposesInstanceSemantics(t *testing.T) {
	var literal kustolang.Expression = kustolang.ParseLiteral("'public'")
	var info *kustolang.ValueInfo = literal.ConstantValueInfo()
	if !literal.IsLiteral() || info == nil || info.Value != "public" || literal.LiteralValue() != "public" || literal.ConstantValue() != "public" {
		t.Fatalf("public literal semantics = literal:%t info:%#v literal-value:%#v constant-value:%#v", literal.IsLiteral(), info, literal.LiteralValue(), literal.ConstantValue())
	}

	var dynamicNull kustolang.Expression = kustolang.ParseLiteral("dynamic(null)")
	if info := dynamicNull.LiteralValueInfo(); info == nil || info.ValueText != "null" || info.Value != nil {
		t.Fatalf("public dynamic null semantics = %#v", info)
	}

	var nonliteral kustolang.Expression = kustolang.ParseExpression("value + 1")
	if nonliteral.IsLiteral() || nonliteral.LiteralValueInfo() != nil || nonliteral.ConstantValueInfo() != nil {
		t.Fatalf("public nonliteral semantics = literal:%t literal-info:%#v constant-info:%#v", nonliteral.IsLiteral(), nonliteral.LiteralValueInfo(), nonliteral.ConstantValueInfo())
	}
}
