package kustolang

import "testing"

// Reference: Kusto.Language 12.4.1 at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.

func TestNativeBinderFilterRecoveryRejectsDiagnosticOnlyCustomPath(t *testing.T) {
	name := func(value string) *NameReference {
		return MustNewNameReferenceTyped(
			MustNewTokenNameTyped(SyntaxTokenFromLexical(NewLexicalToken(SyntaxKindIdentifierToken, "", value))),
			SymbolMatchAny,
		)
	}
	path := MustNewPathExpressionTyped(
		name("T"),
		createOmittedToken(SyntaxKindDotToken),
		name("A"),
		NewDiagnostic("KS999", "injected diagnostic"),
	)
	tree := newTreeBinder(NewBinder(NewGlobalState(nil)))
	if tree.supportsRecoverableFilterScalar(path) {
		t.Fatal("an arbitrary diagnostic admitted a structurally invalid custom path")
	}
	if path.IsBound() || path.Expression().(*NameReference).IsBound() || path.Selector().(*NameReference).IsBound() {
		t.Fatal("recovery preflight mutated the rejected custom path")
	}
}
