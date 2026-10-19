package kustolang

import "testing"

func TestParseLiteralBuildsLiteralExpressions(t *testing.T) {
	tests := []struct {
		text string
		kind SyntaxKind
	}{
		{"'hello'", SyntaxKindStringLiteralExpression},
		{"true", SyntaxKindBooleanLiteralExpression},
		{"42", SyntaxKindLongLiteralExpression},
		{"0x10", SyntaxKindLongLiteralExpression},
		{"int(42)", SyntaxKindIntLiteralExpression},
		{"long(42)", SyntaxKindLongLiteralExpression},
		{"3.14", SyntaxKindRealLiteralExpression},
		{"-42", SyntaxKindLongLiteralExpression},
	}
	for _, tt := range tests {
		expr := ParseLiteral(tt.text)
		if expr == nil || expr.Kind() != tt.kind {
			t.Fatalf("ParseLiteral(%q) = %#v, want kind %v", tt.text, expr, tt.kind)
		}
	}
}

func TestParseLiteralUsesUpstreamLiteralBoundaries(t *testing.T) {
	for _, text := range []string{"null", "-int(1)", "+long(1)", "-real(1)", "+timespan(1s)"} {
		if expr := ParseLiteral(text); expr != nil {
			t.Fatalf("ParseLiteral(%q) = %#v, want nil", text, expr)
		}
	}

	for _, text := range []string{"-1", "+1.5", "-1s"} {
		if expr := ParseLiteral(text); expr == nil {
			t.Fatalf("ParseLiteral(%q) returned nil", text)
		}
	}
}

func TestParseDynamicLiteralRejectsTypedIntJsonNumber(t *testing.T) {
	expr := ParseExpression(`dynamic(int(1))`)
	dynamic, ok := expr.(*DynamicExpression)
	if !ok {
		t.Fatalf("dynamic expression = %#v", expr)
	}
	if !diagnosticsContainMessage(dynamic.ContainedDiagnostics(), "KS006", "Missing json value") {
		t.Fatalf("dynamic diagnostics = %#v", dynamic.ContainedDiagnostics())
	}
	if !diagnosticsContainMessage(dynamic.ContainedDiagnostics(), "KS005", "Expected: )") {
		t.Fatalf("dynamic diagnostics = %#v", dynamic.ContainedDiagnostics())
	}

	for _, text := range []string{`dynamic(1)`, `dynamic(-1)`, `dynamic(1.5)`, `dynamic(-1.5)`, `dynamic(null)`} {
		parsed := ParseExpression(text)
		if parsed == nil {
			t.Fatalf("ParseExpression(%q) returned nil", text)
		}
		if diagnostics := parsed.ContainedDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("ParseExpression(%q) diagnostics = %#v", text, diagnostics)
		}
	}
}

func TestParseLiteralCompoundString(t *testing.T) {
	expr := ParseLiteral("'a' 'b'")
	compound, ok := expr.(*CompoundStringLiteralExpression)
	if !ok {
		t.Fatalf("ParseLiteral compound = %#v", expr)
	}
	if compound.Tokens().Count() != 2 {
		t.Fatalf("compound token count = %d", compound.Tokens().Count())
	}
}

func TestParseLiteralStringForms(t *testing.T) {
	tests := []struct {
		text      string
		valueText string
	}{
		{text: `@"C:\Temp\file.txt"`, valueText: `C:\Temp\file.txt`},
		{text: `h"secret"`, valueText: "secret"},
		{text: "```line1\nline2```", valueText: "line1\nline2"},
		{text: "~~~line1\nline2~~~", valueText: "line1\nline2"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			expr := ParseLiteral(tt.text)
			literal, ok := expr.(*LiteralExpression)
			if !ok || literal.Kind() != SyntaxKindStringLiteralExpression {
				t.Fatalf("ParseLiteral(%q) = %#v", tt.text, expr)
			}
			if literal.Token().ValueText() != tt.valueText {
				t.Fatalf("value text = %q, want %q", literal.Token().ValueText(), tt.valueText)
			}
		})
	}
}

func TestParseTypeOfLiteralRecoveryDiagnostics(t *testing.T) {
	tests := []struct {
		name string
		text string
		code string
	}{
		{
			name: "missing comma",
			text: `typeof(* Score:long)`,
			code: "KS005",
		},
		{
			name: "missing type",
			text: `typeof(*, Score:)`,
			code: "KS006",
		},
		{
			name: "missing first element",
			text: `typeof(, Score:long)`,
			code: "KS006",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := ParseExpression(tt.text)
			if expr == nil {
				t.Fatalf("ParseExpression(%q) returned nil", tt.text)
			}
			if !diagnosticsContain(expr.ContainedDiagnostics(), tt.code) {
				t.Fatalf("expected %s for %q, got %#v", tt.code, tt.text, expr.ContainedDiagnostics())
			}
		})
	}
}

func TestParseTypeOfLiteralBareTypeAfterStarUsesMissingTypeRecovery(t *testing.T) {
	expr := ParseExpression(`typeof(*, long)`)
	if expr == nil {
		t.Fatal("ParseExpression returned nil")
	}
	diagnostics := expr.ContainedDiagnostics()
	if !diagnosticsContainMessage(diagnostics, "KS006", "Missing type name") {
		t.Fatalf("missing-type diagnostics = %#v", diagnostics)
	}
	if diagnosticsContainMessage(diagnostics, "KS006", "Missing expression: If the keyword 'long' is intended to be part of an expression it needs to be bracketted as ['long'].") {
		t.Fatalf("generic expression recovery leaked into typeof list: %#v", diagnostics)
	}
}

func TestParseEmptyTableTypeOfLiteral(t *testing.T) {
	expr := ParseExpression(`typeof()`)
	literal, ok := expr.(*TypeOfLiteralExpression)
	if !ok {
		t.Fatalf("typeof() expression = %#v", expr)
	}
	if diagnostics := literal.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("unexpected typeof() diagnostics: %#v", diagnostics)
	}
	if literal.Types() == nil || literal.Types().Count() != 0 {
		t.Fatalf("typeof() elements = %#v", literal.Types())
	}
}

func TestParseTypeOfKeywordCanBeIdentifier(t *testing.T) {
	expr := ParseExpression(`typeof`)
	reference, ok := expr.(*NameReference)
	if !ok {
		t.Fatalf("typeof expression = %#v", expr)
	}
	name, ok := reference.Name().(*TokenName)
	if !ok || name.Name().Kind() != SyntaxKindTypeOfKeyword {
		t.Fatalf("typeof name = %#v", reference.Name())
	}

	literal := ParseExpression(`typeof(long)`)
	if literal == nil || literal.Kind() != SyntaxKindTypeOfLiteralExpression {
		t.Fatalf("typeof literal expression = %#v", literal)
	}

	tableLiteral := ParseExpression(`typeof(access:string)`)
	if tableLiteral == nil || tableLiteral.Kind() != SyntaxKindTypeOfLiteralExpression {
		t.Fatalf("typeof keyword table literal expression = %#v", tableLiteral)
	}

	invalidScalar := ParseExpression(`typeof(access)`)
	if invalidScalar == nil || invalidScalar.Kind() != SyntaxKindTypeOfLiteralExpression {
		t.Fatalf("typeof invalid scalar expression = %#v", invalidScalar)
	}
	if !diagnosticsContain(invalidScalar.ContainedDiagnostics(), "KS005") || !diagnosticsContain(invalidScalar.ContainedDiagnostics(), "KS006") {
		t.Fatalf("expected table-shaped typeof recovery diagnostics, got %#v", invalidScalar.ContainedDiagnostics())
	}
}

func TestParseDynamicLiteralBuildsJsonTree(t *testing.T) {
	expr := ParseExpression(`dynamic({"user":"alice","scores":[1,-2.5,true,null],"nested":{"id":"00000000-0000-0000-0000-000000000001"}})`)
	dynamic, ok := expr.(*DynamicExpression)
	if !ok {
		t.Fatalf("dynamic expression = %#v", expr)
	}
	object, ok := dynamic.Expression().(*JsonObjectExpression)
	if !ok {
		t.Fatalf("dynamic value = %#v", dynamic.Expression())
	}
	if object.Pairs().Count() != 3 {
		t.Fatalf("object pairs = %d", object.Pairs().Count())
	}
	second := object.Pairs().At(1).Element()
	if second.Name().ValueText() != "scores" {
		t.Fatalf("second pair name = %q", second.Name().ValueText())
	}
	array, ok := second.Value().(*JsonArrayExpression)
	if !ok {
		t.Fatalf("scores value = %#v", second.Value())
	}
	if array.Values().Count() != 4 {
		t.Fatalf("array values = %d", array.Values().Count())
	}
	if _, ok := array.Values().At(1).Element().(*PrefixUnaryExpression); !ok {
		t.Fatalf("negative json number = %#v", array.Values().At(1).Element())
	}
	if got := dynamic.Text(IncludeTriviaAll); got != `dynamic({"user":"alice","scores":[1,-2.5,true,null],"nested":{"id":"00000000-0000-0000-0000-000000000001"}})` {
		t.Fatalf("dynamic text = %q", got)
	}
}

func TestParseDynamicLiteralArrayAndRecovery(t *testing.T) {
	arrayExpr := ParseLiteral(`dynamic([1,,{"k":"v"}])`)
	dynamic, ok := arrayExpr.(*DynamicExpression)
	if !ok {
		t.Fatalf("dynamic array expression = %#v", arrayExpr)
	}
	array, ok := dynamic.Expression().(*JsonArrayExpression)
	if !ok {
		t.Fatalf("dynamic array value = %#v", dynamic.Expression())
	}
	if array.Values().Count() != 2 {
		t.Fatalf("array values = %d", array.Values().Count())
	}
	if len(array.Values().At(1).Element().ContainedDiagnostics()) == 0 {
		t.Fatalf("missing json array element should carry diagnostics")
	}

	objectExpr := ParseExpression(`dynamic({:"value", "ok": true})`)
	dynamic, ok = objectExpr.(*DynamicExpression)
	if !ok {
		t.Fatalf("dynamic object expression = %#v", objectExpr)
	}
	object, ok := dynamic.Expression().(*JsonObjectExpression)
	if !ok {
		t.Fatalf("dynamic object value = %#v", dynamic.Expression())
	}
	if object.Pairs().Count() != 2 {
		t.Fatalf("object pairs = %d", object.Pairs().Count())
	}
	if len(object.Pairs().At(0).Element().Name().ContainedDiagnostics()) == 0 {
		t.Fatalf("missing json object name should carry diagnostics")
	}

	trailingArrayExpr := ParseLiteral(`dynamic([1,])`)
	dynamic, ok = trailingArrayExpr.(*DynamicExpression)
	if !ok {
		t.Fatalf("trailing dynamic array expression = %#v", trailingArrayExpr)
	}
	trailingArray, ok := dynamic.Expression().(*JsonArrayExpression)
	if !ok {
		t.Fatalf("trailing dynamic array value = %#v", dynamic.Expression())
	}
	if trailingArray.Values().Count() != 2 {
		t.Fatalf("trailing array values = %d", trailingArray.Values().Count())
	}
	if !diagnosticsContainMessage(trailingArray.Values().At(1).Element().ContainedDiagnostics(), "KS006", "Missing json value") {
		t.Fatalf("trailing array missing value diagnostics = %#v", trailingArray.Values().At(1).Element().ContainedDiagnostics())
	}

	trailingObjectExpr := ParseExpression(`dynamic({"a":1,})`)
	dynamic, ok = trailingObjectExpr.(*DynamicExpression)
	if !ok {
		t.Fatalf("trailing dynamic object expression = %#v", trailingObjectExpr)
	}
	trailingObject, ok := dynamic.Expression().(*JsonObjectExpression)
	if !ok {
		t.Fatalf("trailing dynamic object value = %#v", dynamic.Expression())
	}
	if trailingObject.Pairs().Count() != 2 {
		t.Fatalf("trailing object pairs = %d", trailingObject.Pairs().Count())
	}
	if !diagnosticsContainMessage(trailingObject.Pairs().At(1).Element().ContainedDiagnostics(), "KS006", "Missing json key:value pair") {
		t.Fatalf("trailing object missing pair diagnostics = %#v", trailingObject.Pairs().At(1).Element().ContainedDiagnostics())
	}
}

func TestParseDynamicLiteralMissingValueRecovery(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		value func(*DynamicExpression) Expression
	}{
		{
			name: "top-level",
			text: `dynamic()`,
			value: func(dynamic *DynamicExpression) Expression {
				return dynamic.Expression()
			},
		},
		{
			name: "object value",
			text: `dynamic({"missing_value":})`,
			value: func(dynamic *DynamicExpression) Expression {
				object := dynamic.Expression().(*JsonObjectExpression)
				return object.Pairs().At(0).Element().Value()
			},
		},
		{
			name: "array value",
			text: `dynamic([1,])`,
			value: func(dynamic *DynamicExpression) Expression {
				array := dynamic.Expression().(*JsonArrayExpression)
				return array.Values().At(1).Element()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dynamic, ok := ParseLiteral(test.text).(*DynamicExpression)
			if !ok {
				t.Fatalf("dynamic expression = %#v", ParseLiteral(test.text))
			}
			value := test.value(dynamic)
			literal, ok := value.(*LiteralExpression)
			if !ok || literal.Kind() != SyntaxKindStringLiteralExpression || literal.Token().Width() != 0 {
				t.Fatalf("missing json value = %#v", value)
			}
			diagnostics := literal.ContainedDiagnostics()
			if len(diagnostics) != 1 || diagnostics[0].Code != "KS006" || diagnostics[0].Message != "Missing json value" {
				t.Fatalf("missing json value diagnostics = %#v", diagnostics)
			}
		})
	}
}

func TestParseDynamicLiteralTrailingCommaRecoveryControls(t *testing.T) {
	array := ParseLiteral(`dynamic([1,])`).(*DynamicExpression).Expression().(*JsonArrayExpression)
	if array.Values().Count() != 2 || array.Values().At(0).Separator().Kind() != SyntaxKindCommaToken {
		t.Fatalf("trailing array values = %#v", array.Values())
	}

	object := ParseLiteral(`dynamic({"a":1,})`).(*DynamicExpression).Expression().(*JsonObjectExpression)
	if object.Pairs().Count() != 2 || object.Pairs().At(0).Separator().Kind() != SyntaxKindCommaToken ||
		!diagnosticsContainMessage(object.Pairs().At(1).Element().ContainedDiagnostics(), "KS006", "Missing json key:value pair") {
		t.Fatalf("trailing object pairs = %#v diagnostics:%#v", object.Pairs(), object.ContainedDiagnostics())
	}
}

func TestParseExpressionFunctionCall(t *testing.T) {
	expr := ParseExpression("count(UserPrincipalName, 5)")
	call, ok := expr.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("ParseExpression function = %#v", expr)
	}
	name, ok := call.Name().Name().(*TokenName)
	if !ok || name.Name().ValueText() != "count" {
		t.Fatalf("function name = %#v", call.Name().Name())
	}
	if call.ArgumentList().Expressions().Count() != 2 {
		t.Fatalf("argument count = %d", call.ArgumentList().Expressions().Count())
	}
	if got := call.Text(IncludeTriviaAll); got != "count(UserPrincipalName, 5)" {
		t.Fatalf("function text = %q", got)
	}
}

func TestParseExpressionBracedClientParameterSupportsIndexedForm(t *testing.T) {
	tests := []struct {
		text string
		name string
	}{
		{text: "{tenant}", name: "tenant"},
		{text: "{workspace[2]}", name: "workspace[2]"},
		{text: "{where[-1]}", name: "where[-1]"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			expr := ParseExpression(tt.text)
			reference, ok := expr.(*NameReference)
			if !ok {
				t.Fatalf("ParseExpression(%q) = %#v", tt.text, expr)
			}
			name, ok := reference.Name().(*BracedName)
			if !ok {
				t.Fatalf("name = %#v, want BracedName", reference.Name())
			}
			if name.Name().ValueText() != tt.name || NameReferenceSimpleName(reference) != tt.name {
				t.Fatalf("braced name = token:%q simple:%q, want %q", name.Name().ValueText(), NameReferenceSimpleName(reference), tt.name)
			}
			if got := reference.Text(IncludeTriviaAll); got != tt.text {
				t.Fatalf("braced reference text = %q, want %q", got, tt.text)
			}
		})
	}

	if expr := ParseExpression("{workspace [2]}"); expr != nil {
		if reference, ok := expr.(*NameReference); ok {
			if _, ok := reference.Name().(*BracedName); ok {
				t.Fatalf("whitespace-separated indexed client parameter parsed as braced name: %#v", expr)
			}
		}
	}
}

func TestParseExpressionBuildsToScalarAndToTableExpressions(t *testing.T) {
	scalar, ok := ParseExpression("toscalar(SigninLogs | summarize Count=count())").(*ToScalarExpression)
	if !ok {
		t.Fatalf("toscalar expression = %#v", scalar)
	}
	if scalar.ToScalar().Kind() != SyntaxKindToScalarKeyword {
		t.Fatalf("toscalar keyword = %v", scalar.ToScalar().Kind())
	}
	if scalar.KindParameter() != nil {
		t.Fatalf("unexpected toscalar kind parameter = %#v", scalar.KindParameter())
	}
	if _, ok := scalar.Expression().(*PipeExpression); !ok {
		t.Fatalf("toscalar inner expression = %#v", scalar.Expression())
	}

	table, ok := ParseExpression("totable kind=nooptimization (SigninLogs | take 5)").(*ToTableExpression)
	if !ok {
		t.Fatalf("totable expression = %#v", table)
	}
	if table.KindParameter() == nil || namedParameterNameText(t, table.KindParameter()) != "kind" {
		t.Fatalf("totable kind parameter = %#v", table.KindParameter())
	}
	if _, ok := table.Expression().(*PipeExpression); !ok {
		t.Fatalf("totable inner expression = %#v", table.Expression())
	}

	tablePrefix, ok := ParseExpression("totable(T | project A).A").(*ToTableExpression)
	if !ok {
		t.Fatalf("totable prefix expression = %#v", tablePrefix)
	}
	if _, ok := tablePrefix.Expression().(*PipeExpression); !ok {
		t.Fatalf("totable prefix body = %T", tablePrefix.Expression())
	}
}

func TestParseEntityGroupBuildsExplicitAndFunctionForms(t *testing.T) {
	explicit, ok := ParseEntityGroup("entity_group[cluster('c').database('db'), database('local')]").(*EntityGroup)
	if !ok {
		t.Fatalf("explicit entity group = %#v", explicit)
	}
	if explicit.Entities().Count() != 2 {
		t.Fatalf("entity count = %d", explicit.Entities().Count())
	}
	if explicit.Entities().At(0).Element().Kind() != SyntaxKindPathExpression {
		t.Fatalf("first entity kind = %v", explicit.Entities().At(0).Element().Kind())
	}

	expression, ok := ParseExpression("entity_group[cluster('c').database('db')]").(*EntityGroup)
	if !ok {
		t.Fatalf("expression entity group = %#v", expression)
	}
	if expression.Entities().Count() != 1 {
		t.Fatalf("expression entity count = %d", expression.Entities().Count())
	}
	if expression.Entities().At(0).Element().Kind() != SyntaxKindPathExpression {
		t.Fatalf("expression entity kind = %v", expression.Entities().At(0).Element().Kind())
	}

	call, ok := ParseEntityGroup(`entity_group("prod")`).(*FunctionCallExpression)
	if !ok {
		t.Fatalf("function-call entity group = %#v", call)
	}
	name, ok := call.Name().Name().(*TokenName)
	if !ok || name.Name().ValueText() != "entity_group" {
		t.Fatalf("entity_group call name = %#v", call.Name().Name())
	}
	if call.ArgumentList().Expressions().Count() != 1 {
		t.Fatalf("entity_group call args = %d", call.ArgumentList().Expressions().Count())
	}

	expressionCall, ok := ParseExpression(`entity_group("prod")`).(*FunctionCallExpression)
	if !ok {
		t.Fatalf("expression entity_group call = %#v", expressionCall)
	}
	bare, ok := ParseExpression("entity_group").(*NameReference)
	if !ok {
		t.Fatalf("bare entity_group expression = %#v", bare)
	}
}

func TestParsePackStarBuildsOrdinaryFunctionCall(t *testing.T) {
	pack, ok := ParseExpression("pack(*)").(*FunctionCallExpression)
	if !ok {
		t.Fatalf("pack star expression = %#v", pack)
	}
	if pack.Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("pack star kind = %v", pack.Kind())
	}
	if diagnostics := pack.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("pack star diagnostics = %#v", diagnostics)
	}

	call, ok := ParseExpression("pack(Value)").(*FunctionCallExpression)
	if !ok {
		t.Fatalf("pack function call = %#v", call)
	}
	if call.Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("pack call kind = %v", call.Kind())
	}
}

func TestParseExpressionParenthesizedPathAndElement(t *testing.T) {
	expr := ParseExpression("database('x').table('y')['Column']")
	path, ok := expr.(*ElementExpression)
	if !ok {
		t.Fatalf("element expression = %#v", expr)
	}
	if _, ok := path.Expression().(*PathExpression); !ok {
		t.Fatalf("element receiver = %#v", path.Expression())
	}
	if _, ok := path.Selector().(*BracketedExpression); !ok {
		t.Fatalf("element selector = %#v", path.Selector())
	}

	parenthesized := ParseExpression("(UserPrincipalName)")
	if _, ok := parenthesized.(*ParenthesizedExpression); !ok {
		t.Fatalf("parenthesized = %#v", parenthesized)
	}
}

func TestParseEntityPathPublicFacade(t *testing.T) {
	expr := ParseEntityPath(`cluster("prod").database("sec").SigninLogs`)
	path, ok := expr.(*PathExpression)
	if !ok {
		t.Fatalf("entity path = %#v", expr)
	}
	if diagnostics := path.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("entity path diagnostics = %#v", diagnostics)
	}
	table, ok := path.Selector().(*NameReference)
	if !ok || tokenNameRawText(table.Name()) != "SigninLogs" {
		t.Fatalf("entity path table selector = %#v", path.Selector())
	}
	databasePath, ok := path.Expression().(*PathExpression)
	if !ok {
		t.Fatalf("entity path database receiver = %#v", path.Expression())
	}
	clusterCall, ok := databasePath.Expression().(*FunctionCallExpression)
	if !ok {
		t.Fatalf("entity path cluster selector = %#v", databasePath.Expression())
	}
	if tokenNameRawText(clusterCall.Name().Name()) != "cluster" {
		t.Fatalf("entity path cluster call = %#v", clusterCall.Name().Name())
	}
	databaseCall, ok := databasePath.Selector().(*FunctionCallExpression)
	if !ok {
		t.Fatalf("entity path database selector = %#v", databasePath.Selector())
	}
	if tokenNameRawText(databaseCall.Name().Name()) != "database" {
		t.Fatalf("entity path database call = %#v", databaseCall.Name().Name())
	}
}

func TestParseEntityPathBracketedSelectorsAndRecovery(t *testing.T) {
	expr := ParseEntityPath(`database("db")[ParityWild*]`)
	element, ok := expr.(*ElementExpression)
	if !ok {
		t.Fatalf("entity path bracketed wildcard = %#v", expr)
	}
	wildcard, ok := element.Selector().(*NameReference)
	if !ok {
		t.Fatalf("entity path wildcard selector = %#v", element.Selector())
	}
	if wildcard.Name().Kind() != SyntaxKindBracketedWildcardedName {
		t.Fatalf("entity path wildcard selector kind = %v", wildcard.Name().Kind())
	}
	if diagnostics := element.ContainedDiagnostics(); len(diagnostics) != 0 {
		t.Fatalf("entity path wildcard diagnostics = %#v", diagnostics)
	}

	recovered := ParseEntityPath(`database("db")[ParityWild*`)
	recoveredElement, ok := recovered.(*ElementExpression)
	if !ok {
		t.Fatalf("recovered entity path = %#v", recovered)
	}
	if !diagnosticsContain(recoveredElement.ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected recovered bracket diagnostic, got %#v", recoveredElement.ContainedDiagnostics())
	}

	if got := ParseEntityPath(`1 + 2`); got != nil {
		t.Fatalf("entity path from scalar expression = %#v, want nil", got)
	}
}

func TestParseExpressionBuildsDataScopeExpression(t *testing.T) {
	expr := ParseExpression("SigninLogs datascope=hotcache")
	scoped, ok := expr.(*DataScopeExpression)
	if !ok {
		t.Fatalf("datascope expression = %#v", expr)
	}
	name, ok := scoped.Expression().(*NameReference)
	if !ok {
		t.Fatalf("datascope receiver = %#v", scoped.Expression())
	}
	if token, ok := name.Name().(*TokenName); !ok || token.Name().ValueText() != "SigninLogs" {
		t.Fatalf("datascope receiver name = %#v", name.Name())
	}
	if scoped.DataScopeClause().Value().RawText() != "hotcache" {
		t.Fatalf("datascope value = %q", scoped.DataScopeClause().Value().RawText())
	}

	invalid := ParseExpression("SigninLogs datascope=coldcache")
	scopedInvalid, ok := invalid.(*DataScopeExpression)
	if !ok {
		t.Fatalf("invalid datascope expression = %#v", invalid)
	}
	if !diagnosticsContain(scopedInvalid.DataScopeClause().ContainedDiagnostics(), "KS005") {
		t.Fatalf("expected invalid datascope diagnostic, got %#v", scopedInvalid.DataScopeClause().ContainedDiagnostics())
	}
}

func TestParseExpressionNamedAndMissingArguments(t *testing.T) {
	expr := ParseExpression("series_stats(value=DurationMs,)")
	call, ok := expr.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("call = %#v", expr)
	}
	args := call.ArgumentList().Expressions()
	if args.Count() != 2 {
		t.Fatalf("argument count = %d", args.Count())
	}
	if _, ok := args.At(0).Element().(*SimpleNamedExpression); !ok {
		t.Fatalf("argument = %#v", args.At(0).Element())
	}
	if !diagnosticsContainMessage(args.At(1).Element().ContainedDiagnostics(), "KS006", "Missing expression") {
		t.Fatalf("missing argument diagnostics = %#v", args.At(1).Element().ContainedDiagnostics())
	}
	if diagnosticsContain(call.ContainedDiagnostics(), "KS001") {
		t.Fatalf("trailing argument comma should not add missing-token diagnostics, got %#v", call.ContainedDiagnostics())
	}

	compound := ParseExpression(`strlen((access, from)="abc")`)
	compoundCall, ok := compound.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("compound call = %#v", compound)
	}
	compoundArgs := compoundCall.ArgumentList().Expressions()
	if compoundArgs.Count() != 1 {
		t.Fatalf("compound argument count = %d", compoundArgs.Count())
	}
	if _, ok := compoundArgs.At(0).Element().(*CompoundNamedExpression); !ok {
		t.Fatalf("compound argument = %#v", compoundArgs.At(0).Element())
	}

	namedKeyword := ParseExpression(`strlen(access="abc")`)
	namedKeywordCall, ok := namedKeyword.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("keyword named call = %#v", namedKeyword)
	}
	namedKeywordArgs := namedKeywordCall.ArgumentList().Expressions()
	if namedKeywordArgs.Count() != 1 {
		t.Fatalf("keyword named argument count = %d", namedKeywordArgs.Count())
	}
	if _, ok := namedKeywordArgs.At(0).Element().(*SimpleNamedExpression); !ok {
		t.Fatalf("keyword named argument = %#v", namedKeywordArgs.At(0).Element())
	}

	dashed := ParseExpression(`strlen(foo-bar="abc")`)
	dashedCall, ok := dashed.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("dashed-name call = %#v", dashed)
	}
	dashedArgs := dashedCall.ArgumentList().Expressions()
	if dashedArgs.Count() != 1 {
		t.Fatalf("dashed-name argument count = %d", dashedArgs.Count())
	}
	dashedNamed, ok := dashedArgs.At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("dashed-name argument = %#v", dashedArgs.At(0).Element())
	}
	if got := nameDeclarationText(t, dashedNamed.Name()); got != "foo-bar" {
		t.Fatalf("dashed-name text = %q", got)
	}
	if !diagnosticsContain(dashedNamed.Name().Diagnostics(), "KS228") {
		t.Fatalf("expected KS228 on dashed name, got %#v", dashedNamed.Name().Diagnostics())
	}

	missing := ParseExpression("count(, User)")
	if diagnostics := missing.ContainedDiagnostics(); len(diagnostics) == 0 {
		t.Fatalf("expected missing argument diagnostic")
	}
}

func TestParseExpressionMissingEqualsArgumentsRemainPositional(t *testing.T) {
	expr := ParseExpression(`format_datetime(x, format "yyyy")`)
	call, ok := expr.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("call = %#v", expr)
	}
	args := call.ArgumentList().Expressions()
	if args.Count() != 3 {
		t.Fatalf("argument count = %d, want 3", args.Count())
	}
	if _, ok := args.At(1).Element().(*NameReference); !ok {
		t.Fatalf("missing-equals argument = %#v, want positional name reference", args.At(1).Element())
	}
	separator := args.At(1).Separator()
	if separator == nil || separator.Width() != 0 || !diagnosticsContain(separator.Diagnostics(), "KS005") {
		t.Fatalf("recovered separator = %#v, want zero-width KS005 comma", separator)
	}

	namedExpr := ParseExpression(`format_datetime(datetime=x, format "yyyy")`)
	namedCall, ok := namedExpr.(*FunctionCallExpression)
	if !ok {
		t.Fatalf("named call = %#v", namedExpr)
	}
	namedArgs := namedCall.ArgumentList().Expressions()
	if namedArgs.Count() != 3 {
		t.Fatalf("named argument count = %d, want 3", namedArgs.Count())
	}
	named, ok := namedArgs.At(0).Element().(*SimpleNamedExpression)
	if !ok {
		t.Fatalf("explicit named argument = %#v", namedArgs.At(0).Element())
	}
	if !diagnosticsContain(named.Name().Diagnostics(), "KS228") {
		t.Fatalf("keyword name diagnostics = %#v, want KS228", named.Name().Diagnostics())
	}
	separator = namedArgs.At(1).Separator()
	if separator == nil || separator.Width() != 0 || !diagnosticsContain(separator.Diagnostics(), "KS005") {
		t.Fatalf("named-call recovered separator = %#v, want zero-width KS005 comma", separator)
	}
}

func TestParseExpressionInvalidKeywordNameReferenceRequiresBrackets(t *testing.T) {
	expr := ParseExpression("and == 1")
	binary, ok := expr.(*BinaryExpression)
	if !ok {
		t.Fatalf("expression = %#v", expr)
	}
	left, ok := binary.Left().(*NameReference)
	if !ok {
		t.Fatalf("left expression = %#v", binary.Left())
	}
	if got := tokenNameRawText(left.Name()); got != "and" {
		t.Fatalf("left name = %q", got)
	}
	if !diagnosticsContain(left.Diagnostics(), "KS228") {
		t.Fatalf("expected KS228 on invalid keyword name, got %#v", left.Diagnostics())
	}

	notName := ParseExpression("and(Account)")
	if notName != nil && diagnosticsContain(notName.ContainedDiagnostics(), "KS228") {
		t.Fatalf("keyword before open paren should not use invalid-name recovery: %#v", notName.ContainedDiagnostics())
	}
}

func TestParseExpressionMissingExpressionWithKeywordDiagnostic(t *testing.T) {
	expr := ParseExpression("1 + where")
	binary, ok := expr.(*BinaryExpression)
	if !ok {
		t.Fatalf("expression = %#v", expr)
	}
	right, ok := binary.Right().(*NameReference)
	if !ok {
		t.Fatalf("right expression = %#v", binary.Right())
	}
	const want = "Missing expression: If the keyword 'where' is intended to be part of an expression it needs to be bracketted as ['where']."
	if !diagnosticsContainMessage(right.Diagnostics(), "KS006", want) {
		t.Fatalf("expected missing-expression keyword diagnostic, got %#v", right.Diagnostics())
	}
}

func TestParseExpressionArithmeticPrecedence(t *testing.T) {
	expr := ParseExpression("1 + 2 * 3")
	add, ok := expr.(*BinaryExpression)
	if !ok || add.Kind() != SyntaxKindAddExpression {
		t.Fatalf("root = %#v, want AddExpression", expr)
	}
	if add.Left().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("left = %#v", add.Left())
	}
	multiply, ok := add.Right().(*BinaryExpression)
	if !ok || multiply.Kind() != SyntaxKindMultiplyExpression {
		t.Fatalf("right = %#v, want MultiplyExpression", add.Right())
	}
}

func TestParseExpressionLogicalPrecedence(t *testing.T) {
	expr := ParseExpression("a == b or c == d and e != f")
	orExpr, ok := expr.(*BinaryExpression)
	if !ok || orExpr.Kind() != SyntaxKindOrExpression {
		t.Fatalf("root = %#v, want OrExpression", expr)
	}
	if left, ok := orExpr.Left().(*BinaryExpression); !ok || left.Kind() != SyntaxKindEqualExpression {
		t.Fatalf("or left = %#v", orExpr.Left())
	}
	andExpr, ok := orExpr.Right().(*BinaryExpression)
	if !ok || andExpr.Kind() != SyntaxKindAndExpression {
		t.Fatalf("or right = %#v, want AndExpression", orExpr.Right())
	}
	if right, ok := andExpr.Right().(*BinaryExpression); !ok || right.Kind() != SyntaxKindNotEqualExpression {
		t.Fatalf("and right = %#v", andExpr.Right())
	}
}

func TestParseNotEqualMissingRightOperandUsesTokenDiagnostic(t *testing.T) {
	for _, text := range []string{"1 !=", "1 <>"} {
		t.Run(text, func(t *testing.T) {
			expr := ParseExpression(text)
			if expr == nil {
				t.Fatalf("ParseExpression(%q) returned nil", text)
			}
			if !diagnosticsContainMessage(expr.ContainedDiagnostics(), "KS005", "Expected: token") {
				t.Fatalf("expected token diagnostic for %q, got %#v", text, expr.ContainedDiagnostics())
			}
			if diagnosticsContain(expr.ContainedDiagnostics(), "KS006") {
				t.Fatalf("not-equal missing RHS should not use missing-expression recovery for %q: %#v", text, expr.ContainedDiagnostics())
			}
		})
	}
}

func TestParseExpressionStringOperators(t *testing.T) {
	expr := ParseExpression(`UserPrincipalName has "admin"`)
	binary, ok := expr.(*BinaryExpression)
	if !ok || binary.Kind() != SyntaxKindHasExpression {
		t.Fatalf("root = %#v, want HasExpression", expr)
	}
	if binary.Operator().Kind() != SyntaxKindHasKeyword {
		t.Fatalf("operator = %v", binary.Operator().Kind())
	}
}

func TestParseExpressionInAndBetween(t *testing.T) {
	inExpr, ok := ParseExpression("EventID in (4624, 4625)").(*InExpression)
	if !ok || inExpr.Kind() != SyntaxKindInExpression {
		t.Fatalf("in expr = %#v", inExpr)
	}
	if inExpr.Right().Expressions().Count() != 2 {
		t.Fatalf("in list count = %d", inExpr.Right().Expressions().Count())
	}

	queryIn, ok := ParseExpression("Account in (Watchlist | project Account)").(*InExpression)
	if !ok || queryIn.Kind() != SyntaxKindInExpression {
		t.Fatalf("query in expr = %#v", queryIn)
	}
	if queryIn.Right().Expressions().Count() != 1 {
		t.Fatalf("query in list count = %d", queryIn.Right().Expressions().Count())
	}
	if _, ok := queryIn.Right().Expressions().At(0).Element().(*PipeExpression); !ok {
		t.Fatalf("query in list element = %#v", queryIn.Right().Expressions().At(0).Element())
	}

	queryHasAny, ok := ParseExpression("Account has_any (Watchlist | project Account)").(*HasAnyExpression)
	if !ok || queryHasAny.Kind() != SyntaxKindHasAnyExpression {
		t.Fatalf("query has_any expr = %#v", queryHasAny)
	}
	if _, ok := queryHasAny.Right().Expressions().At(0).Element().(*PipeExpression); !ok {
		t.Fatalf("query has_any list element = %#v", queryHasAny.Right().Expressions().At(0).Element())
	}

	between, ok := ParseExpression("Score between (1 .. 10)").(*BetweenExpression)
	if !ok || between.Kind() != SyntaxKindBetweenExpression {
		t.Fatalf("between expr = %#v", between)
	}
	if between.Right().DotDot().Kind() != SyntaxKindDotDotToken {
		t.Fatalf("between dotdot = %v", between.Right().DotDot().Kind())
	}
}
