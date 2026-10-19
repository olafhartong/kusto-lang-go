package kustolang

// Ported from FunctionCallExpansion.cs, FunctionCallInfo.cs and
// FunctionCallResult.cs. Full expansion binding and original-tree diagnostic
// relocation remain part of future tree-binder work.

type FunctionCallExpansion struct {
	*SyntaxTree
}

func NewFunctionCallExpansion(root *SyntaxNode, original *SyntaxTree, offsetInOriginal int) *FunctionCallExpansion {
	return &FunctionCallExpansion{SyntaxTree: NewSyntaxTree(root, original, offsetInOriginal)}
}

type FunctionCallExpander func() *FunctionCallExpansion

type FunctionCallInfo struct {
	expander            FunctionCallExpander
	expansion           *FunctionCallExpansion
	Facts               *FunctionBodyFacts
	hasErrors           *bool
	diagnostics         []Diagnostic
	diagnosticsComputed bool
}

func NewDeferredFunctionCallInfo(expander FunctionCallExpander, facts *FunctionBodyFacts) *FunctionCallInfo {
	return &FunctionCallInfo{expander: expander, Facts: functionBodyFactsOrDefault(facts)}
}

func NewFunctionCallInfo(expansion *FunctionCallExpansion, facts *FunctionBodyFacts) *FunctionCallInfo {
	return &FunctionCallInfo{expansion: expansion, Facts: functionBodyFactsOrDefault(facts)}
}

func (i *FunctionCallInfo) Expansion() *FunctionCallExpansion {
	if i == nil {
		return nil
	}
	if i.expansion == nil && i.expander != nil {
		i.expansion = i.expander()
	}
	return i.expansion
}

func (i *FunctionCallInfo) HasErrors() bool {
	if i == nil {
		return false
	}
	if i.hasErrors == nil {
		hasErrors := functionBodyFactsOrDefault(i.Facts).HasSyntaxErrors()
		if !hasErrors {
			for _, diagnostic := range i.Diagnostics() {
				if diagnostic.Severity == DiagnosticSeverityError {
					hasErrors = true
					break
				}
			}
		}
		i.hasErrors = &hasErrors
	}
	return *i.hasErrors
}

func (i *FunctionCallInfo) Diagnostics() []Diagnostic {
	if i == nil {
		return []Diagnostic{}
	}
	if !i.diagnosticsComputed {
		if expansion := i.Expansion(); expansion != nil && expansion.Root != nil {
			i.diagnostics = relocateDiagnosticsInOriginalTree(expansion.SyntaxTree, GetContainedDiagnostics(expansion.Root, DiagnosticsIncludeDefault))
		} else {
			i.diagnostics = []Diagnostic{}
		}
		i.diagnosticsComputed = true
	}
	return append([]Diagnostic{}, i.diagnostics...)
}

func relocateDiagnosticsInOriginalTree(tree *SyntaxTree, diagnostics []Diagnostic) []Diagnostic {
	if len(diagnostics) == 0 {
		return []Diagnostic{}
	}
	offset := 0
	for current := tree; current != nil && current.Original != nil; current = current.Original {
		offset += current.OffsetInOriginal
	}
	out := make([]Diagnostic, len(diagnostics))
	for i, diagnostic := range diagnostics {
		if offset != 0 {
			diagnostic = diagnostic.WithByteLocation(diagnostic.Start+offset, diagnostic.Length)
		}
		out[i] = diagnostic
	}
	return out
}

type FunctionCallResult struct {
	Type *TypeSymbol
	Info *FunctionCallInfo
}

func NewFunctionCallResult(typ *TypeSymbol, info *FunctionCallInfo) FunctionCallResult {
	return FunctionCallResult{Type: typ, Info: info}
}
