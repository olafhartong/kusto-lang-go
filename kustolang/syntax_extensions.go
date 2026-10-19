package kustolang

// Ported from src/Kusto.Language/Syntax/SyntaxExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

func NewNameDeclarationFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) (*NameDeclaration, error) {
	return internal.NewNameDeclarationFromToken(nameToken, diagnostics...)
}

func MustNewNameDeclarationFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) *NameDeclaration {
	return internal.MustNewNameDeclarationFromToken(nameToken, diagnostics...)
}

func NewNameReferenceFromName(name Name, diagnostics ...Diagnostic) (*NameReference, error) {
	return internal.NewNameReferenceFromName(name, diagnostics...)
}

func MustNewNameReferenceFromName(name Name, diagnostics ...Diagnostic) *NameReference {
	return internal.MustNewNameReferenceFromName(name, diagnostics...)
}

func NewNameReferenceFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) (*NameReference, error) {
	return internal.NewNameReferenceFromToken(nameToken, diagnostics...)
}

func MustNewNameReferenceFromToken(nameToken *SyntaxToken, diagnostics ...Diagnostic) *NameReference {
	return internal.MustNewNameReferenceFromToken(nameToken, diagnostics...)
}

func NameSimpleName(name Name) string {
	return internal.NameSimpleName(name)
}

func NameDeclarationSimpleName(declaration *NameDeclaration) string {
	return internal.NameDeclarationSimpleName(declaration)
}

func NameReferenceSimpleName(reference *NameReference) string {
	return internal.NameReferenceSimpleName(reference)
}

func GetParameter(list *SyntaxListOf[*NamedParameter], parameter *QueryOperatorParameter) *NamedParameter {
	return internal.GetParameter(list, parameter)
}

func GetParameterLiteralValue[T any](list *SyntaxListOf[*NamedParameter], parameter *QueryOperatorParameter, defaultValue T) T {
	return internal.GetParameterLiteralValue(list, parameter, defaultValue)
}

func GetParameterNameValue(list *SyntaxListOf[*NamedParameter], parameter *QueryOperatorParameter) string {
	return internal.GetParameterNameValue(list, parameter)
}
