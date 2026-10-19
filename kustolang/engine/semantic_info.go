package kustolang

// Ported from src/Kusto.Language/Binder/SemanticInfo.cs.

type SemanticInfo struct {
	referencedSymbol    Symbol
	referencedSignature *Signature
	ResultType          *TypeSymbol
	Diagnostics         []Diagnostic
	IsConstant          bool
	CalledFunctionInfo  *FunctionCallInfo
	Alternates          []*SyntaxNode
}

func NewSemanticInfoForSymbol(symbol Symbol, result *TypeSymbol, diagnostics ...Diagnostic) SemanticInfo {
	return SemanticInfo{
		referencedSymbol: symbol,
		ResultType:       result,
		Diagnostics:      nonNilDiagnostics(diagnostics),
	}
}

func NewSemanticInfoForSignature(signature *Signature, result *TypeSymbol, diagnostics ...Diagnostic) SemanticInfo {
	info := SemanticInfo{
		referencedSignature: signature,
		ResultType:          result,
		Diagnostics:         nonNilDiagnostics(diagnostics),
	}
	if signature != nil {
		info.referencedSymbol = signature.Symbol
	}
	return info
}

func NewSemanticInfoForType(result *TypeSymbol, diagnostics ...Diagnostic) SemanticInfo {
	return SemanticInfo{ResultType: result, Diagnostics: nonNilDiagnostics(diagnostics)}
}

func NewSemanticInfoForReferencedSymbol(symbol Symbol, diagnostics ...Diagnostic) SemanticInfo {
	switch s := symbol.(type) {
	case nil:
		return NewSemanticInfoForSymbol(Error, &Error.TypeSymbol, diagnostics...)
	case *ScalarSymbol,
		*DynamicAnySymbol,
		*DynamicPrimitiveSymbol,
		*DynamicArraySymbol,
		*DynamicBagSymbol,
		*TupleSymbol,
		*TypeSymbol:
		return NewSemanticInfoForType(semanticInfoResultType(symbol), diagnostics...)
	case *ParameterSymbol:
		return NewSemanticInfoForSymbol(symbol, semanticInfoResultType(symbol), diagnostics...).WithIsConstant(IsScalar(s))
	case *VariableSymbol:
		return NewSemanticInfoForSymbol(symbol, semanticInfoResultType(symbol), diagnostics...).WithIsConstant(s.IsConstant)
	case *OperatorSymbol,
		*ColumnSymbol,
		*TableSymbol,
		*DatabaseSymbol,
		*ClusterSymbol,
		*FunctionSymbol,
		*PatternSymbol,
		*GroupSymbol,
		*MaterializedViewSymbol,
		*ExternalTableSymbol,
		*GraphSymbol,
		*GraphModelSymbol,
		*GraphSnapshotSymbol,
		*EntityGroupSymbol,
		*EntityGroupElementSymbol,
		*StoredQueryResultSymbol:
		return NewSemanticInfoForSymbol(symbol, semanticInfoResultType(symbol), diagnostics...)
	default:
		return NewSemanticInfoForSymbol(Error, &Error.TypeSymbol, diagnostics...)
	}
}

func NewSemanticInfoWithDiagnostics(diagnostics ...Diagnostic) SemanticInfo {
	return SemanticInfo{Diagnostics: nonNilDiagnostics(diagnostics)}
}

func (s SemanticInfo) ReferencedSymbol() Symbol {
	if s.referencedSymbol != nil {
		return s.referencedSymbol
	}
	if s.referencedSignature != nil {
		return s.referencedSignature.Symbol
	}
	return nil
}

func (s SemanticInfo) ReferencedSignature() *Signature {
	if s.referencedSignature != nil {
		return s.referencedSignature
	}
	if fn, ok := s.referencedSymbol.(*FunctionSymbol); ok {
		signatures := fn.Signatures()
		if len(signatures) == 1 {
			return signatures[0]
		}
	}
	if variable, ok := s.referencedSymbol.(*VariableSymbol); ok {
		if fn, ok := variable.Type.(*FunctionSymbol); ok {
			signatures := fn.Signatures()
			if len(signatures) == 1 {
				return signatures[0]
			}
		}
	}
	if entityGroup, ok := s.referencedSymbol.(*EntityGroupSymbol); ok {
		return entityGroup.Signature()
	}
	return nil
}

func (s SemanticInfo) WithReferencedSymbol(symbol Symbol) SemanticInfo {
	s.referencedSymbol = symbol
	s.referencedSignature = nil
	return s
}

func (s SemanticInfo) WithReferencedSignature(signature *Signature) SemanticInfo {
	s.referencedSignature = signature
	s.referencedSymbol = nil
	if signature != nil {
		s.referencedSymbol = signature.Symbol
	}
	return s
}

func (s SemanticInfo) WithResultType(result *TypeSymbol) SemanticInfo {
	s.ResultType = result
	return s
}

func (s SemanticInfo) WithDiagnostics(diagnostics ...Diagnostic) SemanticInfo {
	s.Diagnostics = nonNilDiagnostics(diagnostics)
	return s
}

func (s SemanticInfo) WithIsConstant(isConstant bool) SemanticInfo {
	s.IsConstant = isConstant
	return s
}

func (s SemanticInfo) WithCalledFunctionInfo(calledFunctionInfo *FunctionCallInfo) SemanticInfo {
	s.CalledFunctionInfo = calledFunctionInfo
	return s
}

func (s SemanticInfo) WithAlternates(alternates []*SyntaxNode) SemanticInfo {
	s.Alternates = append([]*SyntaxNode{}, alternates...)
	return s
}

var EmptySemanticInfo = NewSemanticInfoForType(nil)

func nonNilDiagnostics(diagnostics []Diagnostic) []Diagnostic {
	if diagnostics == nil {
		return []Diagnostic{}
	}
	return append([]Diagnostic{}, diagnostics...)
}

func semanticInfoResultType(symbol Symbol) *TypeSymbol {
	result := GetResultType(symbol)
	if result == nil {
		return ScalarTypes.Unknown.Type()
	}
	return returnTypeFromSymbol(result)
}
