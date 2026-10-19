package kustolang

import (
	"fmt"
	"strconv"
)

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/AvoidUsingObsoleteFunctions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

var avoidUsingObsoleteFunctionsDiagnostic = NewDiagnosticDetailed(
	KustoAnalyzerCodeAvoidUsingObsoleteFunctions,
	DiagnosticCategoryCorrectness,
	DiagnosticSeverityWarning,
	"Avoid using obsolete/deprecated functions.",
	"",
)

func NewAvoidUsingObsoleteFunctionsAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name: "AvoidUsingObsoleteFunctionsAnalyzer",
		Diagnostics: func() []Diagnostic {
			return []Diagnostic{avoidUsingObsoleteFunctionsDiagnostic}
		},
		Analyze:   analyzeAvoidUsingObsoleteFunctions,
		FixAction: avoidUsingObsoleteFunctionsFixAction,
		FixEdits:  avoidUsingObsoleteFunctionsFixEdits,
	})
}

func analyzeAvoidUsingObsoleteFunctions(code *KustoCode, diagnostics []Diagnostic, cancellationToken CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		cancellationToken.ThrowIfCancellationRequested()
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call.Name() == nil {
			return
		}
		functionSymbol := referencedFunctionSymbol(call)
		if functionSymbol == nil {
			return
		}
		signature := analyzerReferencedSignature(call)
		var message string
		switch {
		case functionSymbol.IsObsolete():
			message = fmt.Sprintf("The function '%s' is deprecated; use '%s' instead.", functionSymbol.SymbolName(), functionSymbol.Alternative())
		case signature != nil && signature.IsObsolete():
			message = fmt.Sprintf("This form of the function '%s' is deprecated; use '%s' instead.", functionSymbol.SymbolName(), signature.Alternative)
		default:
			return
		}
		diagnostics = append(diagnostics, avoidUsingObsoleteFunctionsDiagnostic.
			WithMessage(message).
			WithLocation(elementTextStart(call.Name(), triviaStart), call.Name().Width()))
	})
	return diagnostics
}

func avoidUsingObsoleteFunctionsFixAction(code *KustoCode, diagnostic Diagnostic, options CodeActionOptions, actions []CodeAction, cancellationToken CancellationToken) []CodeAction {
	if diagnostic.Code != KustoAnalyzerCodeAvoidUsingObsoleteFunctions {
		return actions
	}
	call, _ := obsoleteFunctionCallAtNameLocation(code, diagnostic)
	if call == nil {
		return actions
	}
	functionSymbol := referencedFunctionSymbol(call)
	if functionSymbol == nil {
		return actions
	}
	signature := analyzerReferencedSignature(call)
	alternativeFunction := ""
	if functionSymbol.IsObsolete() {
		alternativeFunction = functionSymbol.Alternative()
	} else if signature != nil {
		alternativeFunction = signature.Alternative
	}
	if alternativeFunction == "" || !obsoleteFunctionCanBeRenamed(signature, alternativeFunction, code.Globals) {
		return actions
	}
	return append(actions, NewCodeAction(
		"Obsolete",
		fmt.Sprintf("Change to '%s'", alternativeFunction),
		fmt.Sprintf("Convert from using deprecated function '%s' to preferred function '%s'", functionSymbol.SymbolName(), alternativeFunction),
		[]string{"rename", strconv.Itoa(diagnostic.Start), strconv.Itoa(diagnostic.Length), alternativeFunction},
	))
}

func avoidUsingObsoleteFunctionsFixEdits(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) FixEdits {
	data := []string(nil)
	if action != nil {
		data = action.ApplyData()
	}
	if len(data) == 4 && data[0] == "rename" {
		if start, err := strconv.Atoi(data[1]); err == nil {
			if length, err := strconv.Atoi(data[2]); err == nil {
				return NewFixEditsList(caretPosition, NewTextReplacement(start, length, data[3]))
			}
		}
	}
	return NewFixEdits(caretPosition, nil)
}

func analyzerReferencedSignature(call *FunctionCallExpression) *Signature {
	if call == nil {
		return nil
	}
	if signature := call.ReferencedSignature(); signature != nil {
		return signature
	}
	if call.Name() != nil {
		return call.Name().ReferencedSignature()
	}
	return nil
}

func obsoleteFunctionCallAtNameLocation(code *KustoCode, diagnostic Diagnostic) (*FunctionCallExpression, int) {
	if code == nil || code.Syntax() == nil {
		return nil, 0
	}
	var found *FunctionCallExpression
	foundTriviaStart := 0
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if found != nil {
			return
		}
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call.Name() == nil {
			return
		}
		if elementTextStart(call.Name(), triviaStart) == diagnostic.Start && call.Name().Width() == diagnostic.Length {
			found = call
			foundTriviaStart = triviaStart
		}
	})
	return found, foundTriviaStart
}

func obsoleteFunctionCanBeRenamed(signature *Signature, alternativeName string, globals *GlobalState) bool {
	if signature == nil || signature.Symbol == nil {
		return false
	}
	switch signature.Symbol.SymbolName() {
	case "row_rank":
		return false
	default:
		return obsoleteFunctionIsRenameCompatible(signature, alternativeName, globals)
	}
}

func obsoleteFunctionIsRenameCompatible(signature *Signature, alternativeName string, globals *GlobalState) bool {
	alternativeSymbol := obsoleteFunctionAlternativeSymbol(alternativeName, globals)
	if alternativeSymbol == nil {
		return false
	}
	for _, alternativeSignature := range alternativeSymbol.Signatures() {
		if obsoleteFunctionSignaturesCompatible(signature, alternativeSignature) {
			return true
		}
	}
	return false
}

func obsoleteFunctionAlternativeSymbol(name string, globals *GlobalState) *FunctionSymbol {
	if globals == nil || name == "" {
		return nil
	}
	if symbol := globals.GetFunction(name); symbol != nil {
		return symbol
	}
	if symbol := globals.GetAggregate(name); symbol != nil {
		return symbol
	}
	return globals.GetPlugIn(name)
}

func obsoleteFunctionSignaturesCompatible(left *Signature, right *Signature) bool {
	if left == nil || right == nil || len(left.Parameters) != len(right.Parameters) {
		return false
	}
	for i := range left.Parameters {
		if !obsoleteFunctionParametersCompatible(left.Parameters[i], right.Parameters[i]) {
			return false
		}
	}
	return true
}

func obsoleteFunctionParametersCompatible(left *Parameter, right *Parameter) bool {
	if left == nil || right == nil || left.TypeKind != right.TypeKind {
		return false
	}
	if left.TypeKind == ParameterTypeDeclared {
		if len(left.DeclaredTypes) != len(right.DeclaredTypes) {
			return false
		}
		for i := range left.DeclaredTypes {
			if left.DeclaredTypes[i] != right.DeclaredTypes[i] {
				return false
			}
		}
	}
	return true
}
