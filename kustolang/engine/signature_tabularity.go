package kustolang

import "strings"

// ComputeTabularity determines an unspecified computed signature's tabularity
// from its parsed body, falling back to native binding when syntax is
// inconclusive. Explicit tabularity, including Unknown, is left unchanged.
func (s *Signature) ComputeTabularity(globals *GlobalState) {
	if s == nil || s.ReturnKind != ReturnTypeComputed || !s.tabularityUnspecified {
		return
	}

	// Match the source's recursion guard and one-shot computation semantics.
	s.tabularityUnspecified = false
	s.Tabularity = TabularityUnknown

	if tabularity := s.computedSyntaxTabularity(globals); tabularity != TabularityUnknown {
		s.Tabularity = tabularity
		return
	}

	if tabularity := s.computedBindingTabularity(globals); tabularity == TabularityScalar || tabularity == TabularityTabular {
		s.Tabularity = tabularity
	}
}

// NewComputedSignatureInferred is the Go equivalent of the computed-body
// Signature constructor that omits an explicit tabularity.
func NewComputedSignatureInferred(body string, parameters ...*Parameter) *Signature {
	return newSignature(ReturnTypeComputed, nil, nil, nil, body, TabularityUnspecified, parameters)
}

// NewComputedFunctionSymbolInferred is the Go equivalent of the computed-body
// FunctionSymbol constructor that omits an explicit tabularity.
func NewComputedFunctionSymbolInferred(name, body string, parameters ...*Parameter) *FunctionSymbol {
	return NewFunctionSymbol(name, NewComputedSignatureInferred(body, parameters...))
}

func (s *Signature) computedSyntaxTabularity(globals *GlobalState) Tabularity {
	if s.Declaration != nil {
		statements := s.Declaration.Statements()
		if statements == nil {
			return TabularityNone
		}
		return GetSyntaxTabularity(statements.SyntaxList, globals)
	}

	body := strings.TrimSpace(s.Body)
	if strings.HasPrefix(body, "{") {
		body = body[1:]
		if strings.HasSuffix(body, "}") {
			body = body[:len(body)-1]
		}
	}
	code := Parse(body)
	if code == nil || code.Query == nil || code.Query.Statements() == nil {
		return TabularityNone
	}

	statements := code.Query.Statements().SyntaxList
	if statements.Count() == 0 {
		return TabularityNone
	}
	last := statements.At(statements.Count() - 1)
	if separated, ok := last.(*SeparatedElement); ok {
		last = separated.Element()
	}
	expressionStatement, ok := WrapGeneratedSyntaxElement(last).(*ExpressionStatement)
	if !ok || expressionStatement.Expression() == nil {
		return TabularityNone
	}
	if call, ok := expressionStatement.Expression().(*FunctionCallExpression); ok {
		state := globals
		if state == nil {
			state = DefaultGlobalState
		}
		name := NameReferenceSimpleName(call.Name())
		if function := state.GetFunction(name); function != nil {
			return function.SymbolTabularity()
		}
		if state.Database != nil {
			if function := state.Database.GetFunction(name); function != nil {
				return function.SymbolTabularity()
			}
		}
		return TabularityUnknown
	}
	return GetSyntaxTabularity(expressionStatement.Expression(), globals)
}

func (s *Signature) computedBindingTabularity(globals *GlobalState) Tabularity {
	if globals == nil {
		globals = DefaultGlobalState
	}

	bindingSignature := s
	if s.Declaration == nil {
		body := strings.TrimSpace(s.Body)
		if !strings.HasPrefix(body, "{") {
			body = "{" + body
		}
		if !strings.HasSuffix(body, "}") {
			body += "\n}"
		}
		declaration := ParseFunctionBodyWithOptions(body, globals.ParseOptions)
		if declaration == nil {
			return TabularityUnknown
		}
		clone := *s
		clone.Declaration = declaration
		bindingSignature = &clone
	}

	argumentTypes := make([]*TypeSymbol, len(s.Parameters))
	argumentParameters := make([]*Parameter, len(s.Parameters))
	for i, parameter := range s.Parameters {
		argumentParameters[i] = parameter
		argumentTypes[i] = representativeComputedParameterType(parameter)
	}

	binder := NewBinder(globals)
	if cache := binder.globalBindingCache; cache != nil {
		cache.expansionMu.Lock()
		defer cache.expansionMu.Unlock()
	}
	result, info := newSimpleNodeBinder(binder).bindComputedFunction(bindingSignature, nil, argumentTypes, argumentParameters)
	if info != nil && info.HasErrors() {
		return TabularityUnknown
	}
	if info != nil && info.Expansion() != nil {
		if raw := functionBodyResultSymbol(info.Expansion().Root); raw != nil {
			return raw.SymbolTabularity()
		}
	}
	if result != nil {
		return result.SymbolTabularity()
	}
	return TabularityUnknown
}

func representativeComputedParameterType(parameter *Parameter) *TypeSymbol {
	if parameter == nil {
		return ScalarTypes.Dynamic.Type()
	}
	switch parameter.TypeKind {
	case ParameterTypeDeclared:
		if symbol := parameter.DeclaredSymbol(0); symbol != nil {
			return returnTypeFromSymbol(symbol)
		}
	case ParameterTypeTabular:
		table := EmptyTableSymbol.WithIsOpen(true)
		return &table.TypeSymbol
	}
	return ScalarTypes.Dynamic.Type()
}
