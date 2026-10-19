package kustolang

import "fmt"

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/CalledFunctionHasErrorsAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var calledFunctionHasErrorsDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeCalledFunctionHasErrors,
	DiagnosticCategoryGeneral,
	DiagnosticSeverityWarning,
	"A called function has errors in its definition",
	"",
)

func NewCalledFunctionHasErrorsAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "CalledFunctionHasErrorsAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{calledFunctionHasErrorsDiagnostic}
		},
		Analyze: analyzeCalledFunctionHasErrors,
	})
}

func analyzeCalledFunctionHasErrors(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStartAndAncestors(code.Syntax(), 0, func(element SyntaxElement, triviaStart int, ancestors []SyntaxElement) {
		cancellationToken.ThrowIfCancellationRequested()
		switch node := WrapGeneratedSyntaxElement(element).(type) {
		case *NameReference:
			if calledFunctionNameReferenceIsFunctionCallName(node, ancestors) || !node.CalledFunctionHasErrors() {
				return
			}
			if signature := node.ReferencedSignature(); signature != nil {
				diagnostics = addCalledFunctionErrors(code, signature, node.GetCalledFunctionDiagnostics(), node, triviaStart, diagnostics)
			}
		case *FunctionCallExpression:
			if !node.CalledFunctionHasErrors() || node.Name() == nil {
				return
			}
			if signature := calledFunctionExpressionSignature(node); signature != nil {
				diagnostics = addCalledFunctionErrors(code, signature, node.GetCalledFunctionDiagnostics(), node.Name(), triviaStart, diagnostics)
			}
		}
	})
	return diagnostics
}

func calledFunctionNameReferenceIsFunctionCallName(name *NameReference, ancestors []SyntaxElement) bool {
	if name == nil || len(ancestors) == 0 {
		return false
	}
	parent, ok := WrapGeneratedSyntaxElement(ancestors[len(ancestors)-1]).(*FunctionCallExpression)
	return ok && parent.Name() != nil && parent.Name().SyntaxNode == name.SyntaxNode
}

func calledFunctionExpressionSignature(call *FunctionCallExpression) *Signature {
	if call == nil {
		return nil
	}
	if call.Name() != nil {
		if signature := call.Name().ReferencedSignature(); signature != nil {
			return signature
		}
	}
	return call.ReferencedSignature()
}

func addCalledFunctionErrors(code *KustoCode, signature *Signature, calledFunctionDiagnostics []Diagnostic, location SyntaxElement, locationTriviaStart int, diagnostics []Diagnostic) []Diagnostic {
	if signature == nil {
		return diagnostics
	}
	if signature.Symbol != nil && calledFunctionSignatureDeclaredInCode(signature, code) {
		for _, diagnostic := range calledFunctionDiagnostics {
			if diagnostic.Severity == DiagnosticSeverityError {
				diagnostics = append(diagnostics, diagnostic)
			}
		}
	}
	name := ""
	if signature.Symbol != nil {
		name = signature.Symbol.SymbolName()
	}
	return append(diagnostics, calledFunctionHasErrorsDiagnostic.
		WithMessage(fmt.Sprintf("The function '%s' has errors in its definition", name)).
		WithLocation(elementTextStart(location, locationTriviaStart), location.Width()))
}

func calledFunctionSignatureDeclaredInCode(signature *Signature, code *KustoCode) bool {
	if signature == nil || signature.Declaration == nil || signature.Declaration.SyntaxNode == nil || code == nil || code.Syntax() == nil {
		return false
	}
	target := signature.Declaration.SyntaxNode
	found := false
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if found {
			return
		}
		if body, ok := WrapGeneratedSyntaxElement(element).(*FunctionBody); ok && body.SyntaxNode == target {
			found = true
			return
		}
		if node, ok := element.(*SyntaxNode); ok && node == target {
			found = true
		}
	})
	return found
}
