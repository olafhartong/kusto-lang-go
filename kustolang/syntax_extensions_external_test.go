package kustolang_test

import (
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicNameConvenienceConstructors(t *testing.T) {
	declarationDiagnostic := kustolang.NewDiagnostic("TEST001", "declaration")
	declaration := kustolang.MustNewNameDeclarationFromToken(
		kustolang.NewIdentifierToken(" ", "Account"),
		declarationDiagnostic,
	)
	if got := declaration.SimpleName(); got != "Account" {
		t.Fatalf("declaration SimpleName() = %q", got)
	}
	if got := kustolang.NameDeclarationSimpleName(declaration); got != "Account" {
		t.Fatalf("NameDeclarationSimpleName() = %q", got)
	}
	if got := declaration.Diagnostics(); len(got) != 1 || got[0].Code != declarationDiagnostic.Code {
		t.Fatalf("declaration diagnostics = %#v", got)
	}

	name := kustolang.MustNewTokenNameTyped(kustolang.NewIdentifierToken("", "Device"))
	reference, err := kustolang.NewNameReferenceFromName(name)
	if err != nil {
		t.Fatal(err)
	}
	if reference.Match() != kustolang.SymbolMatchDefault {
		t.Fatalf("reference Match() = %v, want %v", reference.Match(), kustolang.SymbolMatchDefault)
	}
	if got := reference.SimpleName(); got != "Device" {
		t.Fatalf("reference SimpleName() = %q", got)
	}
	if name.Parent() != reference.SyntaxNode {
		t.Fatal("name reference did not attach the supplied Name")
	}

	tokenReference := kustolang.MustNewNameReferenceFromToken(kustolang.NewIdentifierToken("", "Host"))
	if tokenReference.Match() != kustolang.SymbolMatchDefault || tokenReference.SimpleName() != "Host" {
		t.Fatalf("token reference = match:%v name:%q", tokenReference.Match(), tokenReference.SimpleName())
	}
}

func TestPublicNameConvenienceConstructorsPreserveMalformedAndMissingTokens(t *testing.T) {
	if _, err := kustolang.NewNameDeclarationFromToken(nil); err == nil {
		t.Fatal("nil declaration token did not fail typed construction")
	}
	if _, err := kustolang.NewNameReferenceFromToken(nil); err == nil {
		t.Fatal("nil reference token did not fail typed construction")
	}

	tokenDiagnostic := kustolang.NewDiagnostic("TEST002", "missing name")
	nodeDiagnostic := kustolang.NewDiagnostic("TEST003", "recovered declaration")
	missing := kustolang.NewMissingTokenWithTrivia(" /* recovery */", kustolang.SyntaxKindIdentifierToken, tokenDiagnostic)
	declaration, err := kustolang.NewNameDeclarationFromToken(missing, nodeDiagnostic)
	if err != nil {
		t.Fatal(err)
	}
	tokenName, ok := declaration.Name().(*kustolang.TokenName)
	if !ok || tokenName.Name() != missing {
		t.Fatalf("missing declaration name = %#v", declaration.Name())
	}
	if !missing.IsMissing() || declaration.SimpleName() != "" {
		t.Fatalf("missing declaration = missing:%t name:%q", missing.IsMissing(), declaration.SimpleName())
	}
	if got := diagnosticCodes(declaration.GetContainedSyntaxDiagnostics()); !strings.Contains(got, tokenDiagnostic.Code) || !strings.Contains(got, nodeDiagnostic.Code) {
		t.Fatalf("contained diagnostic codes = %q", got)
	}

	malformedDiagnostic := kustolang.NewDiagnostic("TEST004", "punctuation used as name")
	comma, err := kustolang.NewPunctuationToken("", kustolang.SyntaxKindCommaToken, malformedDiagnostic)
	if err != nil {
		t.Fatal(err)
	}
	reference := kustolang.MustNewNameReferenceFromToken(comma)
	if reference.SimpleName() != "," || reference.Match() != kustolang.SymbolMatchDefault {
		t.Fatalf("malformed reference = match:%v name:%q", reference.Match(), reference.SimpleName())
	}
	if got := diagnosticCodes(reference.GetContainedSyntaxDiagnostics()); !strings.Contains(got, malformedDiagnostic.Code) {
		t.Fatalf("malformed reference diagnostic codes = %q", got)
	}
}

func TestPublicNameConvenienceConstructorsPreserveOwnershipValidation(t *testing.T) {
	token := kustolang.NewIdentifierToken("", "Owned")
	owner := kustolang.MustNewTokenNameTyped(token)

	defer func() {
		recovered := recover()
		err, ok := recovered.(kustolang.InvalidOperationError)
		if !ok || err.Message != "The syntax element already has a parent." {
			t.Fatalf("panic = %#v, want InvalidOperationError", recovered)
		}
		if token.Parent() != owner.SyntaxNode {
			t.Fatal("failed convenience construction changed token ownership")
		}
	}()

	kustolang.MustNewNameReferenceFromToken(token)
	t.Fatal("convenience constructor silently reparented a reused token")
}

func TestPublicQueryOperatorParameterSyntaxExtensions(t *testing.T) {
	declaration := kustolang.NewQueryOperatorParameter(
		"enabled",
		kustolang.QueryOperatorParameterValueBoolLiteral,
		kustolang.WithQueryOperatorParameterAliases([]string{"on"}),
	)
	parameter := kustolang.MustNewNamedParameterTyped(
		kustolang.MustNewNameDeclarationTyped(kustolang.MustNewTokenNameTyped(kustolang.NewIdentifierToken("", "on"))),
		mustPunctuationToken(t, kustolang.SyntaxKindEqualToken),
		kustolang.MustNewLiteralExpressionTyped(kustolang.SyntaxKindBooleanLiteralExpression, mustLiteralToken(t, "true", kustolang.SyntaxKindBooleanLiteralToken)),
		kustolang.CompletionHintScalar,
	)
	list := kustolang.AsSyntaxListOf[*kustolang.NamedParameter](kustolang.NewTypedSyntaxList("NamedParameter", parameter))

	if got := kustolang.GetParameter(list, declaration); got != parameter {
		t.Fatalf("parameter = %#v", got)
	}
	if got := kustolang.GetParameterLiteralValue(list, declaration, false); !got {
		t.Fatal("literal value = false, want true")
	}
	missing := kustolang.NewQueryOperatorParameter("missing", kustolang.QueryOperatorParameterValueIntegerLiteral)
	if got := kustolang.GetParameterLiteralValue(list, missing, int64(42)); got != 42 {
		t.Fatalf("missing literal default = %d", got)
	}

	nameDeclaration := kustolang.NewQueryOperatorParameter("target", kustolang.QueryOperatorParameterValueNameDeclaration)
	nameParameter := kustolang.MustNewNamedParameterTyped(
		kustolang.MustNewNameDeclarationTyped(kustolang.MustNewTokenNameTyped(kustolang.NewIdentifierToken("", "target"))),
		mustPunctuationToken(t, kustolang.SyntaxKindEqualToken),
		kustolang.MustNewNameReferenceTyped(kustolang.MustNewTokenNameTyped(kustolang.NewIdentifierToken("", "Account")), kustolang.SymbolMatchAny),
		kustolang.CompletionHintScalar,
	)
	nameList := kustolang.AsSyntaxListOf[*kustolang.NamedParameter](kustolang.NewTypedSyntaxList("NamedParameter", nameParameter))
	if got := kustolang.GetParameterNameValue(nameList, nameDeclaration); got != "Account" {
		t.Fatalf("name value = %q", got)
	}
	if got := kustolang.NameReferenceSimpleName(nameParameter.Expression().(*kustolang.NameReference)); got != "Account" {
		t.Fatalf("simple name = %q", got)
	}
}

func mustPunctuationToken(t *testing.T, kind kustolang.SyntaxKind) *kustolang.SyntaxToken {
	t.Helper()
	token, err := kustolang.NewPunctuationToken("", kind)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func mustLiteralToken(t *testing.T, text string, kind kustolang.SyntaxKind) *kustolang.SyntaxToken {
	t.Helper()
	token, err := kustolang.NewLiteralToken("", text, kind)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func diagnosticCodes(diagnostics []kustolang.Diagnostic) string {
	codes := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		codes = append(codes, diagnostic.Code)
	}
	return strings.Join(codes, ",")
}
