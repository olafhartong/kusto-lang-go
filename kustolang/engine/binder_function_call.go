package kustolang

import (
	"strings"
	"sync"
)

const functionCallResultSymbolPayload = "FunctionCallResultSymbol"

// Function calls are ordered like the upstream tree binder: resolve the name
// first, bind arguments under signature-specific scopes, then bind the call.
func (t *treeBinder) visitFunctionCall(call *FunctionCallExpression) {
	if call == nil || call.Name() == nil || call.ArgumentList() == nil {
		return
	}
	call.Name().SetSemanticInfo(t.binder.resolveFunctionCallName(call.Name()))
	oldPath, oldScope, oldRow, oldAggregateArgument := t.binder.pathScope, t.binder.scopeKind, t.binder.rowScope, t.binder.aggregateArgumentScope
	defer func() {
		t.binder.pathScope, t.binder.scopeKind, t.binder.rowScope, t.binder.aggregateArgumentScope = oldPath, oldScope, oldRow, oldAggregateArgument
	}()
	t.binder.pathScope = nil

	outer := oldScope
	fn, _ := call.Name().ReferencedSymbol().(*FunctionSymbol)
	argumentScope := ArgumentScopeForFunction(t.binder.Globals, fn, outer)
	args := call.ArgumentList().Expressions().Elements()
	arguments := make([]Expression, len(args))
	for i, separated := range args {
		if separated != nil {
			arguments[i] = separated.Element()
		}
	}
	var signature *Signature
	var argumentParameters []*Parameter
	if fn != nil && len(fn.Signatures()) == 1 {
		signature = fn.Signatures()[0]
		argumentParameters = signature.ArgumentParametersForExpressions(arguments, nil, t.binder.Globals)
	}
	for i, separated := range args {
		if separated == nil || separated.Element() == nil {
			continue
		}
		arg := separated.Element()
		if syntaxElementIsMissing(arg) {
			if reference, ok := arg.(*NameReference); ok {
				reference.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error))
				continue
			}
		}
		var parameter *Parameter
		if i < len(argumentParameters) {
			parameter = argumentParameters[i]
		}
		if signature == nil {
			// Upstream uses the broad function argument scope until overload
			// resolution selects a single signature.
			t.binder.scopeKind = argumentScope
		} else if parameter != nil && parameter.ArgumentKind == ArgumentKindAggregate {
			t.binder.scopeKind = ScopeKindAggregate
		} else {
			t.binder.scopeKind = argumentScope
		}
		t.binder.aggregateArgumentScope = t.binder.scopeKind == ScopeKindAggregate
		t.binder.rowScope = oldRow
		if signature != nil {
			if parameterAcceptsTabularArgument(parameter) {
				t.binder.rowScope = nil
				if reference, ok := arg.(*NameReference); ok {
					info := t.binder.ResolveNameReferenceInRowScope(nil, reference, false)
					if _, unresolved := info.ReferencedSymbol().(*ErrorSymbol); !unresolved && len(info.Diagnostics) == 0 {
						reference.SetSemanticInfo(info)
						t.bindImplicitFunctionSource(reference)
						continue
					}
					t.binder.rowScope = oldRow
				}
			}
			if parameter != nil && i > 0 {
				var firstArgument Expression
				if args[0] != nil {
					firstArgument = args[0].Element()
				}
				switch parameter.ArgumentKind {
				case ArgumentKindColumnParameter0:
					if table := expressionTableResult(firstArgument); table != nil {
						t.binder.rowScope = table
					} else {
						t.binder.scopeKind = ScopeKindAggregate
					}
				case ArgumentKindColumnParameter0Common:
					if table := expressionTableResult(firstArgument); table != nil {
						t.binder.rowScope = NewTableSymbol("", commonColumns(oldRow, table)...)
					} else {
						t.binder.scopeKind = ScopeKindAggregate
					}
				case ArgumentKindExpressionParameter0Element:
					if tuple, ok := TupleForType(ExpressionResultType(firstArgument)); ok {
						t.binder.rowScope = NewTableSymbol("", tuple.Columns()...)
					}
				}
				t.binder.aggregateArgumentScope = t.binder.scopeKind == ScopeKindAggregate
			}
		}
		t.visitElement(arg)
	}
	t.binder.pathScope = oldPath
	t.binder.aggregateArgumentScope = oldAggregateArgument
	t.bindNode(call.SyntaxNode)
	if info, ok := call.GetSemanticInfo(); ok {
		if nameInfo, bound := call.Name().GetSemanticInfo(); !bound || nameInfo.ResultType == nil || !isErrorTypeSymbol(nameInfo.ResultType) {
			referenced := info.ReferencedSymbol()
			if signature := info.ReferencedSignature(); signature != nil && signature.Symbol != nil {
				referenced = signature.Symbol
			}
			call.Name().SetSemanticInfo(NewSemanticInfoForSymbol(referenced, info.ResultType).WithReferencedSignature(info.ReferencedSignature()))
		}
	}
}

func parameterAcceptsTabularArgument(parameter *Parameter) bool {
	if parameter == nil {
		return false
	}
	if parameter.TypeKind == ParameterTypeTabular {
		return true
	}
	if parameter.TypeKind == ParameterTypeDeclared {
		for _, symbol := range parameter.DeclaredSymbols {
			if IsTabular(symbol) {
				return true
			}
		}
	}
	return false
}

func commonColumns(left, right *TableSymbol) []*ColumnSymbol {
	if left == nil || right == nil {
		return nil
	}
	var out []*ColumnSymbol
	for _, l := range left.Columns() {
		for _, r := range right.Columns() {
			if l.SymbolName() == r.SymbolName() {
				out = append(out, NewColumnSymbolFull(l.SymbolName(), l.Type, l.Description, []*ColumnSymbol{l, r}, l.Source, l.Examples))
				break
			}
		}
	}
	return out
}

func (n *simpleNodeBinder) bindFunctionCall(call *FunctionCallExpression) *SemanticInfo {
	nameInfo, ok := call.Name().GetSemanticInfo()
	if !ok {
		return nil
	}
	callable := GetResultType(nameInfo.ReferencedSymbol())
	fn, functionCall := callable.(*FunctionSymbol)
	pattern, patternCall := callable.(*PatternSymbol)
	if !functionCall && !patternCall {
		if nameInfo.ResultType != nil && !isErrorTypeSymbol(nameInfo.ResultType) {
			return semanticInfoPtr(nameInfo)
		}
		return semanticInfoPtr(NewSemanticInfoForReferencedSymbol(Error))
	}
	args := call.ArgumentList().Expressions().Elements()
	expressions := make([]Expression, 0, len(args))
	types := make([]*TypeSymbol, 0, len(args))
	if implicit := n.binder.implicitFunctionArgument; implicit != nil && implicit.call == call.SyntaxNode {
		expressions = append(expressions, implicit.expression)
		types = append(types, implicit.typ)
	}
	for _, arg := range args {
		if arg != nil && arg.Element() != nil {
			expressions = append(expressions, arg.Element())
			types = append(types, resultTypeOrError(arg.Element()))
		}
	}
	if patternCall {
		return n.bindPatternFunctionCall(pattern, call, expressions, types)
	}
	return n.bindFunction(fn, call.Name(), call.SyntaxNode, expressions, types)
}

func (n *simpleNodeBinder) bindPatternFunctionCall(pattern *PatternSymbol, call *FunctionCallExpression, expressions []Expression, types []*TypeSymbol) *SemanticInfo {
	result := BindPatternCallWithArgumentTypes(pattern, expressions, types, n.binder.Globals)
	diagnostics := make([]Diagnostic, 0, len(result.Diagnostics))
	arguments := call.ArgumentList().Expressions().Elements()
	for i, diagnostic := range result.Diagnostics {
		location := SyntaxElement(call.Name())
		if i < len(result.DiagnosticArgumentIndices) {
			argumentIndex := result.DiagnosticArgumentIndices[i]
			if argumentIndex >= 0 && argumentIndex < len(arguments) && arguments[argumentIndex] != nil && arguments[argumentIndex].Element() != nil {
				location = arguments[argumentIndex].Element()
			}
		}
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(diagnostic, location))
	}
	call.SyntaxNode.setPayload(functionCallResultSymbolPayload, result.Result)
	return semanticInfoPtr(NewSemanticInfoForSymbol(pattern, result.ResultType(), diagnostics...))
}

// bindFunction is shared by ordinary calls and calls carrying a synthetic
// argument (currently invoke's input table).
func (n *simpleNodeBinder) bindFunction(fn *FunctionSymbol, name SyntaxElement, resultNode *SyntaxNode, expressions []Expression, types []*TypeSymbol) *SemanticInfo {
	return n.bindFunctionWithDynamicRetry(fn, name, resultNode, expressions, types, true)
}

func (n *simpleNodeBinder) bindFunctionWithDynamicRetry(fn *FunctionSymbol, name SyntaxElement, resultNode *SyntaxNode, expressions []Expression, types []*TypeSymbol, allowDynamicRetry bool) *SemanticInfo {
	matches := n.getBestMatchingFunctionSignatures(fn.Signatures(), expressions, symbolsFromTypeSymbols(types), false)
	diagnostics := []Diagnostic{}
	if len(matches) != 1 {
		if len(expressions) == 0 && fn.MinArgumentCount() > 0 {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticFunctionExpectsArgumentCountRange(fn.SymbolName(), fn.MinArgumentCount(), fn.MaxArgumentCount()), name))
		} else if !argumentTypesHaveErrorsOrUnknown(types) {
			names := make([]string, len(types))
			for i, typ := range types {
				names[i] = typ.SymbolName()
			}
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticFunctionNotDefinedWithMatchingParameters(fn.SymbolName(), names...), name))
		}
		result := GetCommonSignatureReturnType(matches, types, func(signature *Signature) *TypeSymbol {
			parameters := signature.ArgumentParametersForExpressions(expressions, symbolsFromTypeSymbols(types), n.binder.Globals)
			_, result, _, _ := n.functionCallResult(signature, resultNode, expressions, types, parameters, nil, nil)
			return result
		})
		return semanticInfoPtr(NewSemanticInfoForSymbol(fn, result, diagnostics...))
	}
	sig := matches[0]
	params := sig.ArgumentParametersForExpressions(expressions, symbolsFromTypeSymbols(types), n.binder.Globals)
	diagnostics = append(diagnostics, n.functionSignatureDiagnostics(sig, name, expressions, types, params)...)
	resultSymbol, result, calledFunctionInfo, resultDiagnostics := n.functionCallResult(sig, resultNode, expressions, types, params, nil, nil)
	diagnostics = append(diagnostics, resultDiagnostics...)
	initialResult := result
	if allowDynamicRetry {
		result = n.refineDynamicFunctionCallResult(fn, resultNode, expressions, types, result, nil, nil)
	}
	if result != initialResult && resultSymbol == initialResult {
		resultSymbol = result
	}
	resultNode.setPayload(functionCallResultSymbolPayload, resultSymbol)
	constant := fn.IsConstantFoldable()
	for _, expression := range expressions {
		if node := simpleBinderSyntaxNode(expression); node == nil || !node.IsConstant() {
			constant = false
		}
	}
	info := NewSemanticInfoForSignature(sig, result, diagnostics...).WithIsConstant(constant)
	if sig.ReturnKind == ReturnTypeParameter0EntityGroup {
		// reference implementation can carry the concrete EntityGroupSymbol as the TypeSymbol result.
		// Keep that identity directly on the Go call while retaining the
		// intrinsic function signature for call analysis.
		info.referencedSymbol = resultSymbol
		info.referencedSignature = sig
	}
	if calledFunctionInfo != nil {
		info = info.WithCalledFunctionInfo(calledFunctionInfo)
	}
	return semanticInfoPtr(info)
}

func (n *simpleNodeBinder) refineDynamicFunctionCallResult(fn *FunctionSymbol, resultNode *SyntaxNode, expressions []Expression, types []*TypeSymbol, result *TypeSymbol, argumentSemantics []SemanticInfo, summarizeByColumns []*ColumnSymbol) *TypeSymbol {
	if n == nil || n.binder == nil || !sameType(result, ScalarTypes.Dynamic.Type()) || !HasDynamicPrimitives(types) {
		return result
	}
	unwrapped := GetUnwrappedDynamicPrimitives(types)
	rebound := n.functionCallReturnTypeForTypes(fn, resultNode, expressions, unwrapped, argumentSemantics, summarizeByColumns)
	return RefineDynamicFunctionResult(result, rebound)
}

func (n *simpleNodeBinder) functionCallReturnTypeForTypes(fn *FunctionSymbol, resultNode *SyntaxNode, expressions []Expression, types []*TypeSymbol, argumentSemantics []SemanticInfo, summarizeByColumns []*ColumnSymbol) *TypeSymbol {
	if fn == nil {
		return &Error.TypeSymbol
	}
	matches := n.getBestMatchingFunctionSignatures(fn.Signatures(), expressions, symbolsFromTypeSymbols(types), false)
	return GetCommonSignatureReturnType(matches, types, func(signature *Signature) *TypeSymbol {
		parameters := signature.ArgumentParametersForExpressions(expressions, symbolsFromTypeSymbols(types), n.binder.Globals)
		_, candidateResult, _, _ := n.functionCallResult(
			signature, resultNode, expressions, types, parameters, argumentSemantics, summarizeByColumns,
		)
		return candidateResult
	})
}

func (n *simpleNodeBinder) functionCallResult(sig *Signature, resultNode *SyntaxNode, expressions []Expression, types []*TypeSymbol, params []*Parameter, argumentSemantics []SemanticInfo, summarizeByColumns []*ColumnSymbol) (Symbol, *TypeSymbol, *FunctionCallInfo, []Diagnostic) {
	diagnostics := []Diagnostic{}
	// Computed calls are bound by the branch below so its expansion guard can
	// stop recursion. Calling Signature.GetReturnType here would enter the
	// public type-only API before that guard is active.
	result := ScalarTypes.Unknown.Type()
	if sig.ReturnKind != ReturnTypeComputed {
		result = sig.GetReturnTypeForArgumentExpressions(n.binder.Globals, expressions, types, params)
	}
	var resultSymbol Symbol = result
	var calledFunctionInfo *FunctionCallInfo
	if symbol, sourceDiagnostics, sourceCallInfo, ok := n.bindSourceReferenceFunctionReturn(sig, expressions, params); ok {
		resultSymbol = symbol
		result = returnTypeFromSymbol(symbol)
		diagnostics = append(diagnostics, sourceDiagnostics...)
		calledFunctionInfo = sourceCallInfo
	} else if sig.ReturnKind == ReturnTypeCustom && sig.CustomReturnType != nil {
		semantics := argumentSemantics
		if len(semantics) != len(expressions) {
			semantics = make([]SemanticInfo, len(expressions))
			for i, expression := range expressions {
				semantics[i], _ = simpleBinderSyntaxNode(expression).GetSemanticInfo()
			}
		}
		context := n.customReturnTypeContext(resultNode, sig, expressions, types, params, semantics)
		context.SummarizeByColumns = append([]*ColumnSymbol{}, summarizeByColumns...)
		resultSymbol = sig.CustomReturnType(context)
		result = returnTypeFromSymbol(resultSymbol)
	} else if sig.ReturnKind == ReturnTypeComputed {
		if computedResult, computedInfo := n.bindComputedFunction(sig, expressions, types, params); computedInfo != nil {
			calledFunctionInfo = computedInfo
			if computedResult != nil {
				resultSymbol = computedResult
				result = returnTypeFromSymbol(computedResult)
			}
		}
	} else if sig.ReturnKind == ReturnTypeDeclared && sig.DeclaredReturnSymbol != nil {
		resultSymbol = sig.DeclaredReturnSymbol
	}
	return resultSymbol, result, calledFunctionInfo, diagnostics
}

func (n *simpleNodeBinder) customReturnTypeContext(location *SyntaxNode, signature *Signature, expressions []Expression, types []*TypeSymbol, parameters []*Parameter, semantics []SemanticInfo) CustomReturnTypeContext {
	constants := make([]bool, len(expressions))
	values := make([]any, len(expressions))
	names := make([]string, len(expressions))
	resultSymbols := make([]Symbol, len(expressions))
	for i, expression := range expressions {
		var semantic SemanticInfo
		if i < len(semantics) {
			semantic = semantics[i]
		}
		constants[i], values[i] = customReturnConstant(expression, semantic)
		names[i] = serializeExpressionResultName(expression)
		resultSymbols[i] = customReturnArgumentResultSymbol(expression, semantic)
	}
	rowScope := n.binder.rowScope
	if rowScope == nil {
		rowScope = EmptyTableSymbol
	}
	return CustomReturnTypeContext{
		Location: location, Arguments: expressions, ArgumentTypes: types, ArgumentResultSymbols: resultSymbols,
		ArgumentParameters: parameters, ArgumentIsConstant: constants, ArgumentSemantics: semantics,
		ArgumentResultNames: names, ArgumentConstantValues: values, RowScope: rowScope,
		Signature: signature, Globals: n.binder.Globals, CurrentCluster: n.binder.CurrentCluster(),
		CurrentDatabase: n.binder.CurrentDatabase(), CurrentFunction: n.binder.CurrentFunction(),
		ReferencedSymbolByName: func(name string) Symbol {
			return n.binder.ResolveName(name, SymbolMatchDefault, false).ReferencedSymbol()
		},
		ResultTypeByName: func(name string) *TypeSymbol {
			return n.binder.ResolveName(name, SymbolMatchDefault, false).ResultType
		},
		ResultNameForExpression: func(expression Expression, defaultName string) string {
			return expressionResultName(expression, defaultName, rowScope, n.binder.Globals)
		},
	}
}

func customReturnConstant(expression Expression, semantic SemanticInfo) (bool, any) {
	if info := ExpressionLiteralValueInfo(expression); info != nil {
		return true, info.Value
	}
	if variable, ok := semantic.ReferencedSymbol().(*VariableSymbol); ok && variable.IsConstant {
		if variable.ConstantValueInfo != nil {
			return true, variable.ConstantValueInfo.Value
		}
		return true, nil
	}
	return semantic.IsConstant, nil
}

func customReturnArgumentResultSymbol(expression Expression, semantic SemanticInfo) Symbol {
	if call, ok := expression.(*FunctionCallExpression); ok {
		if result, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok && symbol != nil {
				return symbol
			}
		}
	}
	if symbol := semantic.ReferencedSymbol(); symbol != nil {
		return symbol
	}
	if semantic.ResultType != nil {
		return semantic.ResultType
	}
	return resultTypeOrError(expression)
}

func (n *simpleNodeBinder) functionSignatureDiagnostics(signature *Signature, location SyntaxElement, arguments []Expression, argumentTypes []*TypeSymbol, parameters []*Parameter) []Diagnostic {
	if signature == nil {
		return nil
	}

	var diagnostics []Diagnostic
	if !signature.IsValidArgumentCount(len(arguments)) {
		var diagnostic Diagnostic
		if signature.HasRepeatableParameters {
			if len(arguments) < signature.MinArgumentCount || len(arguments) > signature.MaxArgumentCount {
				diagnostic = DiagnosticFunctionExpectsArgumentCountRange(signature.Symbol.SymbolName(), signature.MinArgumentCount, signature.MaxArgumentCount)
			} else {
				diagnostic = DiagnosticFunctionHasIncorrectNumberOfArguments()
			}
		} else {
			diagnostic = DiagnosticFunctionExpectsArgumentCountExact(signature.Symbol.SymbolName(), len(signature.Parameters))
		}
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(diagnostic, location))
	}

	namedArgumentsAllowed := signature.AllowsNamedArguments(n.binder.Globals)
	if namedArgumentsAllowed && len(diagnostics) == 0 {
		hadOutOfOrderNamedArgument := false
		reportedUnnamedArgument := false
		for i, argument := range arguments {
			named, isNamed := argument.(*SimpleNamedExpression)
			var namedParameter *Parameter
			if isNamed && named.Name() != nil {
				namedParameter = signature.GetParameter(parameterNameFromDeclaration(named.Name()))
				if namedParameter == nil {
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticUnknownArgumentName(), named.Name()))
				}
			}

			if isNamed && !hadOutOfOrderNamedArgument {
				var orderedParameter *Parameter
				if i < len(signature.Parameters) {
					orderedParameter = signature.Parameters[i]
				}
				hadOutOfOrderNamedArgument = orderedParameter != namedParameter
			} else if !isNamed && hadOutOfOrderNamedArgument && !reportedUnnamedArgument {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticUnnamedArgumentAfterOutOfOrderNamedArgument(), argument))
				reportedUnnamedArgument = true
			}
		}
	}

	if len(diagnostics) == 0 && signatureHasCommonParameter(parameters) &&
		GetCommonArgumentType(parameters, argumentTypes) == nil && !argumentTypesHaveErrorsOrUnknown(argumentTypes) {
		diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticNoCommonArgumentType(), location))
	}
	if len(diagnostics) == 0 {
		diagnostics = append(diagnostics, n.functionArgumentDiagnostics(signature, arguments, argumentTypes, parameters)...)
	}
	if namedArgumentsAllowed && len(diagnostics) == 0 {
		for _, parameter := range signature.Parameters {
			if parameter == nil || parameter.IsOptional() || parameterOccursInArguments(parameter, parameters) {
				continue
			}
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticMissingArgumentForParameter(parameter.Name), location))
		}
	}
	return diagnostics
}

func signatureHasCommonParameter(parameters []*Parameter) bool {
	for _, parameter := range parameters {
		if parameter == nil {
			continue
		}
		switch parameter.TypeKind {
		case ParameterTypeCommonScalar, ParameterTypeCommonNumber, ParameterTypeCommonSummable,
			ParameterTypeCommonOrderable, ParameterTypeCommonScalarOrDynamic:
			return true
		}
	}
	return false
}

func parameterOccursInArguments(parameter *Parameter, argumentParameters []*Parameter) bool {
	for _, candidate := range argumentParameters {
		if candidate == parameter {
			return true
		}
	}
	return false
}

func (n *simpleNodeBinder) bindComputedFunction(signature *Signature, arguments []Expression, argumentTypes []*TypeSymbol, argumentParameters []*Parameter) (Symbol, *FunctionCallInfo) {
	return n.bindComputedFunctionCore(signature, arguments, argumentTypes, argumentParameters, true)
}

// bindComputedFunctionCore follows GetComputedFunctionCallResult: a cached
// variable result can avoid body binding, but its expansion remains available
// on demand through FunctionCallInfo.
func (n *simpleNodeBinder) bindComputedFunctionCore(signature *Signature, arguments []Expression, argumentTypes []*TypeSymbol, argumentParameters []*Parameter, useCachedResult bool) (Symbol, *FunctionCallInfo) {
	if n == nil || n.binder == nil || signature == nil || signature.ReturnKind != ReturnTypeComputed || (signature.Body == "" && signature.Declaration == nil) {
		return nil, nil
	}
	if n.binder.functionExpansions == nil {
		n.binder.functionExpansions = map[*Signature]bool{}
	}
	if n.binder.functionExpansions[signature] {
		return nil, nil
	}
	n.binder.functionExpansions[signature] = true
	defer delete(n.binder.functionExpansions, signature)

	if facts, ok := TryGetFunctionBodyFacts(n.binder.globalBindingCache, n.binder.localBindingCache, signature); ok {
		if useCachedResult && facts.HasVariableReturnType() {
			if callSite, ok := computedResultCallSite(signature, facts.DependentParameters(), arguments, argumentParameters); ok &&
				CanCacheResultType(n.binder.LocalScope, n.binder.globalBindingCache, n.binder.localBindingCache, callSite) {
				if result, ok := TryGetResultSymbolFromCache(n.binder.globalBindingCache, n.binder.localBindingCache, callSite); ok && result != nil {
					return result, NewDeferredFunctionCallInfo(n.deferredComputedFunctionExpansion(signature, arguments, argumentTypes, argumentParameters), facts)
				}
			}
		}
		if callSite, ok := computedExpansionCallSite(signature, arguments, argumentParameters); ok &&
			CanCacheExpansion(n.binder.LocalScope, n.binder.globalBindingCache, n.binder.localBindingCache, callSite) {
			if expansion, ok := TryGetExpansionFromCache(n.binder.globalBindingCache, n.binder.localBindingCache, callSite); ok && expansion != nil {
				return computedFunctionExpansionResult(expansion), NewFunctionCallInfo(expansion, facts)
			}
		}
	}

	var body *FunctionBody
	if signature.Declaration != nil {
		body = signature.Declaration.CloneTyped(true)
	} else {
		body = ParseFunctionBody("{" + signature.Body + "}")
	}
	if body == nil || body.SyntaxNode == nil {
		return nil, nil
	}
	globals := n.binder.Globals
	if globals == nil {
		globals = DefaultGlobalState
	}
	if signature.Symbol != nil {
		if database := globals.GetDatabase(signature.Symbol); database != nil {
			globals = globals.WithDatabase(database)
		}
	}
	calledBinder := NewBinder(globals)
	calledBinder.currentFunction, _ = signature.Symbol.(*FunctionSymbol)
	calledBinder.functionExpansions = n.binder.functionExpansions
	calledBinder.entityGroupExpansions = n.binder.entityGroupExpansions
	calledBinder.functionStaticScopes = n.binder.functionStaticScopes
	if outer := n.binder.functionStaticScopes[signature]; outer != nil {
		calledBinder.LocalScope = NewLocalScope(outer)
	}
	for parameterIndex, parameter := range signature.Parameters {
		if parameter == nil {
			continue
		}
		argumentIndex := -1
		for i, mapped := range argumentParameters {
			if mapped == parameter {
				argumentIndex = i
				break
			}
		}
		var typ Symbol = ScalarTypes.Unknown.Type()
		var source Expression
		constant := false
		var constantValue any
		if argumentIndex >= 0 && argumentIndex < len(arguments) {
			source = arguments[argumentIndex]
			constant = ExpressionIsConstant(source)
			constantValue = ExpressionConstantValueInfo(source)
			if len(parameter.DeclaredSymbols) == 1 && parameter.DeclaredSymbols[0] != nil && IsScalar(parameter.DeclaredSymbols[0]) &&
				!sameType(returnTypeFromSymbol(parameter.DeclaredSymbols[0]), ScalarTypes.Unknown.Type()) {
				typ = parameter.DeclaredSymbols[0]
			} else if table := expressionTableResult(source); table != nil {
				typ = table
			} else if argumentIndex < len(argumentTypes) && argumentTypes[argumentIndex] != nil {
				typ = argumentTypes[argumentIndex]
			}
		} else if argumentIndex >= 0 && argumentIndex < len(argumentTypes) && argumentTypes[argumentIndex] != nil {
			// Binder.GetComputedReturnType supplies types without syntax
			// expressions. Preserve a backing table symbol when the type carries
			// one, just as the source TypeSymbol hierarchy does.
			typ = argumentTypes[argumentIndex]
		} else {
			if parameter.DefaultValue != nil {
				source = parameter.DefaultValue
				constant = ExpressionIsConstant(source)
				constantValue = ExpressionConstantValueInfo(source)
			}
			if parameterIndex < len(signature.Parameters) && parameter.DeclaredTypes != nil && len(parameter.DeclaredTypes) == 1 {
				typ = parameter.DeclaredTypes[0]
			}
		}
		calledBinder.LocalScope.AddSymbol(NewVariableSymbol(parameter.Name, typ, constant, constantValue, source))
	}
	calledBinder.TryBindWithCancellation(body, n.binder.cancellationToken)
	n.binder.cancellationToken.ThrowIfCancellationRequested()
	expansion := NewFunctionCallExpansion(body.SyntaxNode, nil, 0)
	facts := GetOrComputeFunctionBodyFacts(globals, n.binder.globalBindingCache, n.binder.localBindingCache, signature, body)
	n.binder.cancellationToken.ThrowIfCancellationRequested()
	result := computedFunctionExpansionResult(expansion)
	if n.binder.localBindingCache != nil {
		n.binder.localBindingCache.FunctionExpansionCounts[signature]++
	}
	if facts.HasVariableReturnType() {
		if callSite, ok := computedResultCallSite(signature, facts.DependentParameters(), arguments, argumentParameters); ok &&
			CanCacheResultType(n.binder.LocalScope, n.binder.globalBindingCache, n.binder.localBindingCache, callSite) {
			n.binder.cancellationToken.ThrowIfCancellationRequested()
			AddResultSymbolToCache(globals, n.binder.globalBindingCache, n.binder.localBindingCache, callSite, result)
		}
	}
	if callSite, ok := computedExpansionCallSite(signature, arguments, argumentParameters); ok &&
		CanCacheExpansion(n.binder.LocalScope, n.binder.globalBindingCache, n.binder.localBindingCache, callSite) {
		n.binder.cancellationToken.ThrowIfCancellationRequested()
		AddExpansionToCache(globals, n.binder.globalBindingCache, n.binder.localBindingCache, callSite, expansion)
	}
	return result, NewFunctionCallInfo(expansion, facts)
}

// deferredComputedFunctionExpansion reuses the source synchronization point
// because callers may request an expansion after the original bind completed.
func (n *simpleNodeBinder) deferredComputedFunctionExpansion(signature *Signature, arguments []Expression, argumentTypes []*TypeSymbol, argumentParameters []*Parameter) FunctionCallExpander {
	arguments = append([]Expression{}, arguments...)
	argumentTypes = append([]*TypeSymbol{}, argumentTypes...)
	argumentParameters = append([]*Parameter{}, argumentParameters...)
	var expansion *FunctionCallExpansion
	var once sync.Once
	return func() *FunctionCallExpansion {
		once.Do(func() {
			if n == nil || n.binder == nil {
				return
			}
			cache := n.binder.globalBindingCache
			if cache != nil {
				cache.expansionMu.Lock()
				defer cache.expansionMu.Unlock()
			}
			_, info := n.bindComputedFunctionCore(signature, arguments, argumentTypes, argumentParameters, false)
			if info != nil {
				expansion = info.Expansion()
			}
		})
		return expansion
	}
}

func computedFunctionExpansionResult(expansion *FunctionCallExpansion) Symbol {
	if expansion == nil {
		return ScalarTypes.Unknown.Type()
	}
	result := functionBodyResultSymbol(expansion.Root)
	if result == nil || isErrorTypeSymbol(returnTypeFromSymbol(result)) {
		return ScalarTypes.Unknown.Type()
	}
	return result
}

// computedExpansionCallSite mirrors Binder_FunctionCalls.TryGetExpansionCallSiteInfo.
// The cache key contains every argument/default value that can influence a
// database function's bound body. Nonconstant scalar expressions are omitted,
// which is the same conservative source behavior.
func computedExpansionCallSite(signature *Signature, arguments []Expression, argumentParameters []*Parameter) (CallSiteInfo, bool) {
	if signature == nil || arguments == nil {
		return CallSiteInfo{}, false
	}
	parameters := make([]*Parameter, 0, len(signature.Parameters))
	values := make([]any, 0, len(signature.Parameters))
	for _, parameter := range signature.Parameters {
		if parameter == nil {
			continue
		}
		index := indexOfCallSiteParameter(argumentParameters, parameter)
		if index >= 0 && index < len(arguments) {
			if value, ok := computedCallSiteArgumentValue(arguments[index]); ok {
				parameters = append(parameters, parameter)
				values = append(values, value)
			}
			continue
		}
		if parameter.DefaultValue != nil {
			if value, ok := computedCallSiteArgumentValue(parameter.DefaultValue); ok {
				parameters = append(parameters, parameter)
				values = append(values, value)
			}
		}
	}
	return NewCallSiteInfo(signature, parameters, values), true
}

// computedResultCallSite mirrors TryGetResultTypeCallSite. Only parameters
// proven to affect the return schema are part of the key, and every one must
// have a constant scalar or table value at this call site.
func computedResultCallSite(signature *Signature, dependentParameters []*Parameter, arguments []Expression, argumentParameters []*Parameter) (CallSiteInfo, bool) {
	if signature == nil || arguments == nil {
		return CallSiteInfo{}, false
	}
	parameters := make([]*Parameter, 0, len(dependentParameters))
	values := make([]any, 0, len(dependentParameters))
	for _, parameter := range dependentParameters {
		if parameter == nil {
			continue
		}
		index := indexOfCallSiteParameter(argumentParameters, parameter)
		if index < 0 || index >= len(arguments) {
			return CallSiteInfo{}, false
		}
		value, ok := computedCallSiteArgumentValue(arguments[index])
		if !ok {
			return CallSiteInfo{}, false
		}
		parameters = append(parameters, parameter)
		values = append(values, value)
	}
	return NewCallSiteInfo(signature, parameters, values), true
}

func computedCallSiteArgumentValue(expression Expression) (any, bool) {
	if expression == nil {
		return nil, false
	}
	if value := ExpressionConstantValueInfo(expression); value != nil {
		return value.Value, true
	}
	if table := expressionTableResult(expression); table != nil {
		return table, true
	}
	return nil, false
}

func (n *simpleNodeBinder) bindSourceReferenceFunctionReturn(signature *Signature, arguments []Expression, argumentParameters []*Parameter) (Symbol, []Diagnostic, *FunctionCallInfo, bool) {
	if signature == nil || n == nil || n.binder == nil || n.binder.Globals == nil {
		return nil, nil, nil, false
	}
	argumentIndex := signature.argumentIndexForParameter(0, argumentParameters)
	switch signature.ReturnKind {
	case ReturnTypeParameter0Cluster:
		if argumentIndex < 0 || argumentIndex >= len(arguments) {
			return NewClusterSymbolWithMembers("", nil, true), nil, nil, true
		}
		name, literal := ExpressionConstantValue(arguments[argumentIndex]).(string)
		if !literal {
			return NewClusterSymbolWithMembers("", nil, true), nil, nil, true
		}
		if cluster := n.binder.Globals.GetCluster(name); cluster != nil {
			return cluster, nil, nil, true
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownCluster(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyClusterNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, arguments[argumentIndex])
		return n.binder.getOpenCluster(name), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0Database:
		if len(arguments) == 0 {
			return n.binder.Globals.Database, nil, nil, true
		}
		if argumentIndex < 0 || argumentIndex >= len(arguments) {
			return NewDatabaseSymbolWithAlternateName("", "", nil, true), nil, nil, true
		}
		name, literal := ExpressionConstantValue(arguments[argumentIndex]).(string)
		if !literal {
			return NewDatabaseSymbolWithAlternateName("", "", nil, true), nil, nil, true
		}
		if database := n.binder.matchingDatabase(name, n.binder.pathScope); database != nil {
			return database, nil, nil, true
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownDatabase(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyDatabaseNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, arguments[argumentIndex])
		if strings.Contains(name, "*") {
			return Error, []Diagnostic{diagnostic}, nil, true
		}
		return n.binder.getOpenDatabase(name, n.binder.firstDatabaseScopeCluster(n.binder.pathScope)), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0Table:
		name, argument, literal := sourceReferenceLiteralName(arguments, argumentIndex)
		if !literal {
			return NewTableSymbol("").WithIsOpen(true), nil, nil, true
		}
		if n.binder.pathScope == nil {
			if local := n.binder.localTableSource(name); local != nil {
				return local, nil, nil, true
			}
		}
		database := n.sourceReferenceDatabaseScope()
		if strings.Contains(name, "*") {
			var matches []Symbol
			if database != nil {
				for _, table := range database.Tables() {
					if KustoFactsMatches(name, table.SymbolName()) {
						matches = append(matches, table)
					}
				}
			}
			if len(matches) == 1 {
				return matches[0], nil, nil, true
			}
			if len(matches) > 1 {
				return NewGroupSymbol(matches...), nil, nil, true
			}
			return NewGroupSymbol(), nil, nil, true
		}
		if database != nil {
			if table := database.GetTable(name); table != nil {
				return table, nil, nil, true
			}
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownTableOrFunction(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyTableNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, argument)
		return n.binder.getOpenTable(name, database), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0ExternalTable:
		name, argument, literal := sourceReferenceLiteralName(arguments, argumentIndex)
		if !literal {
			return EmptyTableSymbol.WithIsOpen(true), nil, nil, true
		}
		if database := n.sourceReferenceDatabaseScope(); database != nil {
			if table := database.GetExternalTable(name); table != nil {
				return table, nil, nil, true
			}
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownExternalTable(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyExternalTableNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, argument)
		return NewExternalTableSymbol(name), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0MaterializedView:
		name, argument, literal := sourceReferenceLiteralName(arguments, argumentIndex)
		if !literal {
			return NewMaterializedViewSymbol("", nil, ""), nil, nil, true
		}
		if database := n.sourceReferenceDatabaseScope(); database != nil {
			if view := database.GetMaterializedView(name); view != nil {
				return view, nil, nil, true
			}
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownMaterializedView(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyMaterializedViewNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, argument)
		return NewMaterializedViewSymbol(name, nil, ""), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0EntityGroup:
		name, argument, literal := sourceReferenceLiteralName(arguments, argumentIndex)
		if !literal {
			return NewEntityGroupSymbol(""), nil, nil, true
		}
		if database := n.sourceReferenceDatabaseScope(); database != nil {
			if group := database.GetEntityGroup(name); group != nil {
				result, callInfo := n.bindComputedEntityGroup(group)
				return result, nil, callInfo, true
			}
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownEntityGroup(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyEntityGroupNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, argument)
		return NewEntityGroupSymbol(name), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0StoredQueryResult:
		name, argument, literal := sourceReferenceLiteralName(arguments, argumentIndex)
		if !literal {
			return NewStoredQueryResultSymbol(""), nil, nil, true
		}
		if database := n.sourceReferenceDatabaseScope(); database != nil {
			if result := database.GetStoredQueryResult(name); result != nil {
				return result, nil, nil, true
			}
		}
		diagnostic := DiagnosticNameDoesNotReferToAnyKnownStoredQueryResult(name)
		if n.binder.isFuzzy {
			diagnostic = DiagnosticFuzzyStoredQueryResultNotDefined(name)
		}
		diagnostic = n.binder.diagnosticAtSyntaxElement(diagnostic, argument)
		return NewStoredQueryResultSymbol(name), []Diagnostic{diagnostic}, nil, true
	case ReturnTypeParameter0Graph:
		name, argument, literal := sourceReferenceLiteralName(arguments, argumentIndex)
		if !literal {
			return EmptyGraphSymbol, nil, nil, true
		}
		database := n.sourceReferenceDatabaseScope()
		if database == nil || database.GetGraphModel(name) == nil {
			diagnostic := n.binder.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownGraphModel(name), argument)
			return NewGraphSymbol(name, nil, nil), []Diagnostic{diagnostic}, nil, true
		}
		model := database.GetGraphModel(name)
		graph := n.binder.graphForModel(model)
		var diagnostics []Diagnostic
		if len(arguments) > 1 && len(argumentParameters) > 1 && argumentParameters[1] != nil && argumentParameters[1].Name == "snapshot" {
			if snapshot, ok := ExpressionConstantValue(arguments[1]).(string); ok && snapshot != "" {
				if _, found := model.TryGetSnapshot(snapshot); !found {
					diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(
						DiagnosticNameDoesNotReferToAnyKnownGraphSnapshot(snapshot, name), arguments[1]))
				}
			}
		}
		return graph, diagnostics, nil, true
	default:
		return nil, nil, nil, false
	}
}

func (n *simpleNodeBinder) bindComputedEntityGroup(group *EntityGroupSymbol) (Symbol, *FunctionCallInfo) {
	if group == nil || group.Signature() == nil || group.Signature().ReturnKind != ReturnTypeComputed {
		return group, nil
	}

	signature := group.Signature()
	if n.binder.entityGroupExpansions == nil {
		n.binder.entityGroupExpansions = map[*Signature]bool{}
	}
	if n.binder.entityGroupExpansions[signature] {
		return NewEntityGroupSymbol(""), nil
	}
	n.binder.entityGroupExpansions[signature] = true
	defer delete(n.binder.entityGroupExpansions, signature)

	globals := n.binder.Globals
	if database := globals.GetDatabase(group); database != nil {
		globals = globals.WithDatabase(database)
	}
	body := ParseEntityGroup(signature.Body)
	root := simpleBinderSyntaxNode(body)
	if root == nil {
		return NewEntityGroupSymbol(""), nil
	}

	calledBinder := NewBinder(globals)
	calledBinder.entityGroupExpansions = n.binder.entityGroupExpansions
	calledBinder.TryBind(body)

	expansion := NewFunctionCallExpansion(root, nil, 0)
	callInfo := NewFunctionCallInfo(expansion, ComputeFunctionBodyFacts(signature, body))
	if result := boundEntityGroupResult(body); result != nil {
		return result, callInfo
	}
	return NewEntityGroupSymbol(""), callInfo
}

func boundEntityGroupResult(expression Expression) *EntityGroupSymbol {
	if call, ok := expression.(*FunctionCallExpression); ok {
		if result, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if group, ok := result.(*EntityGroupSymbol); ok {
				return group
			}
		}
	}
	if group, ok := expressionReferencedSymbol(expression).(*EntityGroupSymbol); ok {
		return group
	}
	return nil
}

func (b *Binder) localTableSource(name string) Symbol {
	if b == nil || b.LocalScope == nil || name == "" {
		return nil
	}
	var symbols []Symbol
	b.LocalScope.GetSymbols(name, SymbolMatchLocal|SymbolMatchFunction|SymbolMatchView, &symbols)
	for _, symbol := range symbols {
		switch symbol := symbol.(type) {
		case *VariableSymbol:
			if table, ok := tableSymbolFromSymbol(symbol.Type); ok {
				return table
			}
		case *FunctionSymbol:
			for _, signature := range symbol.Signatures() {
				if signature != nil && signature.MinArgumentCount == 0 {
					if table, ok := tableSymbolFromSymbol(signature.DeclaredReturnSymbol); ok {
						return table
					}
				}
			}
		}
	}
	return nil
}

func sourceReferenceLiteralName(arguments []Expression, index int) (string, SyntaxElement, bool) {
	if index < 0 || index >= len(arguments) || arguments[index] == nil {
		return "", nil, false
	}
	name, ok := ExpressionConstantValue(arguments[index]).(string)
	return name, arguments[index], ok
}

func (n *simpleNodeBinder) sourceReferenceDatabaseScope() *DatabaseSymbol {
	if n != nil && n.binder != nil {
		if database, ok := n.binder.pathScope.(*DatabaseSymbol); ok {
			return database
		}
		if n.binder.Globals != nil {
			return n.binder.Globals.Database
		}
	}
	return nil
}

func (b *Binder) graphForModel(model *GraphModelSymbol) *GraphSymbol {
	if model == nil {
		return NewGraphSymbol("", nil, nil)
	}
	if b.graphModels == nil {
		b.graphModels = map[*GraphModelSymbol]*GraphSymbol{}
	}
	if graph := b.graphModels[model]; graph != nil {
		return graph
	}
	edges := declaredSignatureTables(model.Edges)
	nodes := declaredSignatureTables(model.Nodes)
	graph := NewGraphSymbolFromShapes(model.SymbolName(), edges, nodes)
	b.graphModels[model] = graph
	return graph
}

func declaredSignatureTables(signatures []*Signature) []*TableSymbol {
	var tables []*TableSymbol
	for _, signature := range signatures {
		if signature == nil {
			continue
		}
		if table, ok := tableSymbolFromSymbol(signature.DeclaredReturnSymbol); ok {
			tables = append(tables, table)
		}
	}
	return tables
}

func matchingClusterDatabases(cluster *ClusterSymbol, name string, matches []Symbol) []Symbol {
	if cluster == nil {
		return matches
	}
	for _, database := range cluster.Databases() {
		if database != nil && (KustoFactsMatches(name, database.SymbolName(), true) || KustoFactsMatches(name, database.AlternateName(), true)) {
			matches = append(matches, database)
		}
	}
	return matches
}

func (b *Binder) firstDatabaseScopeCluster(scope Symbol) *ClusterSymbol {
	if cluster, ok := scope.(*ClusterSymbol); ok {
		return cluster
	}
	if group, ok := scope.(*GroupSymbol); ok {
		for _, member := range group.Members() {
			if cluster, ok := member.(*ClusterSymbol); ok {
				return cluster
			}
		}
	}
	if b != nil && b.Globals != nil {
		return b.Globals.Cluster
	}
	return nil
}

func (b *Binder) matchingDatabase(name string, scope Symbol) Symbol {
	if b == nil || b.Globals == nil {
		return nil
	}
	if scope == nil || scope == b.Globals.Cluster {
		current := b.Globals.Database
		if current != nil && (name == "" || strings.EqualFold(current.SymbolName(), name) || strings.EqualFold(current.AlternateName(), name)) {
			return current
		}
	}
	if database := b.aliasedDatabases[name]; database != nil {
		return database
	}
	var matches []Symbol
	switch scoped := scope.(type) {
	case nil:
		matches = matchingClusterDatabases(b.Globals.Cluster, name, matches)
	case *ClusterSymbol:
		matches = matchingClusterDatabases(scoped, name, matches)
	case *GroupSymbol:
		for _, member := range scoped.Members() {
			if cluster, ok := member.(*ClusterSymbol); ok {
				matches = matchingClusterDatabases(cluster, name, matches)
			}
		}
	}
	matches = distinctSymbols(matches)
	if len(matches) == 1 {
		return matches[0]
	}
	if len(matches) > 1 {
		return NewGroupSymbol(matches...)
	}
	if strings.Contains(name, "*") {
		return nil
	}
	if cluster := b.firstDatabaseScopeCluster(scope); cluster != nil && cluster.IsOpen() {
		return b.getOpenDatabase(name, cluster)
	}
	return nil
}

func (n *simpleNodeBinder) functionArgumentDiagnostics(signature *Signature, arguments []Expression, argumentTypes []*TypeSymbol, parameters []*Parameter) []Diagnostic {
	var diagnostics []Diagnostic
	argumentSymbols := symbolsFromTypeSymbols(argumentTypes)
	for i, argument := range arguments {
		if i >= len(parameters) || i >= len(argumentTypes) || parameters[i] == nil || argument == nil {
			continue
		}
		parameter := parameters[i]
		argumentType := argumentTypes[i]
		if _, star := argument.(*StarExpression); star {
			if parameter.ArgumentKind != ArgumentKindStarOnly && parameter.ArgumentKind != ArgumentKindStarAllowed {
				diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticStarExpressionNotAllowed(), argument))
			}
			continue
		}
		if isDefaultValueIndicator(parameter, argument, argumentType) {
			continue
		}
		if compound, ok := argument.(*CompoundNamedExpression); ok {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(DiagnosticCompoundNamedArgumentsNotSupported(), compound.Names()))
		}
		argument = functionArgumentUnderlyingExpression(argument)
		diagnostics = append(diagnostics, n.functionArgumentTypeDiagnostics(signature, argument, argumentType, parameter, parameters, argumentSymbols)...)
		if isErrorTypeSymbol(argumentType) {
			continue
		}
		var diagnostic *Diagnostic
		switch parameter.ArgumentKind {
		case ArgumentKindLiteral, ArgumentKindLiteralNotEmpty:
			if !ExpressionIsLiteral(argument) {
				d := DiagnosticExpressionMustBeLiteral()
				diagnostic = &d
			} else if parameter.ArgumentKind == ArgumentKindLiteralNotEmpty {
				if value, ok := ExpressionLiteralValue(argument).(string); ok && value == "" {
					d := DiagnosticExpressionMustNotBeEmpty()
					diagnostic = &d
				}
			} else if len(parameter.Values) > 0 && !functionLiteralValueAllowed(ExpressionLiteralValue(argument), parameter) {
				values := FunctionParameterDiagnosticValues(parameter)
				d := DiagnosticExpressionMustHaveValue(values...)
				diagnostic = &d
			}
		case ArgumentKindConstant:
			_, parameterReference := expressionReferencedSymbol(argument).(*ParameterSymbol)
			if node := simpleBinderSyntaxNode(argument); (node == nil || !node.IsConstant()) && !parameterReference {
				d := DiagnosticExpressionMustBeConstant()
				diagnostic = &d
			}
		case ArgumentKindColumn, ArgumentKindColumnParameter0, ArgumentKindColumnParameter0Common:
			if _, ok := expressionReferencedSymbol(argument).(*ColumnSymbol); !ok {
				d := DiagnosticColumnExpected()
				diagnostic = &d
			}
		case ArgumentKindStarOnly:
			if _, ok := argument.(*StarExpression); !ok {
				d := DiagnosticStarExpressionNotAllowed()
				diagnostic = &d
			}
		}
		if diagnostic != nil {
			diagnostics = append(diagnostics, n.binder.diagnosticAtSyntaxElement(*diagnostic, argument))
		}
	}
	return diagnostics
}

func (n *simpleNodeBinder) functionArgumentTypeDiagnostics(signature *Signature, argument Expression, argumentType *TypeSymbol, parameter *Parameter, argumentParameters []*Parameter, argumentTypes []Symbol) []Diagnostic {
	if signature == nil || argument == nil || parameter == nil || argumentType == nil {
		return nil
	}
	argumentHasUnresolvedType := (isErrorTypeSymbol(argumentType) || sameType(argumentType, ScalarTypes.Unknown.Type())) && !SyntaxIsMissing(argument)
	declaredParameterError := parameter.TypeKind == ParameterTypeDeclared && isErrorTypeSymbol(argumentType)
	if argumentHasUnresolvedType && !declaredParameterError {
		return nil
	}
	at := func(diagnostic Diagnostic) []Diagnostic {
		return []Diagnostic{n.binder.diagnosticAtSyntaxElement(diagnostic, argument)}
	}
	isScalar := IsScalar(argumentType)
	switch parameter.TypeKind {
	case ParameterTypeAny:
		return nil
	case ParameterTypeDeclared:
		if requiredTable, ok := declaredParameterTable(parameter); ok {
			argumentTable, tabular := tableSymbolFromSymbol(argumentType)
			if !tabular {
				return at(DiagnosticTabularValueExpected())
			}
			if !tablesAssignable(requiredTable, argumentTable) {
				return at(DiagnosticTabularValueDoesNotHaveRequiredColumns())
			}
			return nil
		}
		match := signature.ParameterMatchKind(argumentParameters, argumentTypes, parameter, argument, argumentType, false)
		if (match == ParameterMatchCompatible || match == ParameterMatchNone) && !n.allowLooseFunctionParameterMatching(signature) {
			return at(DiagnosticTypeExpected(parameterTypeNames(parameter)...))
		}
	case ParameterTypeScalar:
		if !isScalar {
			return at(DiagnosticScalarTypeExpected())
		}
	case ParameterTypeInteger:
		if !isIntegerSymbol(argumentType) {
			return at(DiagnosticExpressionMustBeInteger())
		}
	case ParameterTypeRealOrDecimal:
		if !isRealOrDecimalSymbol(argumentType) {
			return at(DiagnosticExpressionMustBeRealOrDecimal())
		}
	case ParameterTypeStringOrDynamic:
		if !isStringOrDynamicSymbol(argumentType) {
			return at(DiagnosticExpressionMustHaveType("string", "dynamic"))
		}
	case ParameterTypeStringOrArray:
		if !isStringOrArraySymbol(argumentType) {
			return at(DiagnosticExpressionMustBeStringOrArray())
		}
	case ParameterTypeIntegerOrArray:
		if !isIntegerOrArraySymbol(argumentType) {
			return at(DiagnosticExpressionMustBeIntegerOrArray())
		}
	case ParameterTypeDynamicArray:
		if !isDynamicArraySymbol(argumentType) {
			return at(DiagnosticExpressionMustBeDynamicArray())
		}
	case ParameterTypeDynamicBag:
		if !isDynamicBagSymbol(argumentType) {
			return at(DiagnosticExpressionMustBeDynamicBag())
		}
	case ParameterTypeNumber:
		if !isNumericSymbol(argumentType) {
			return at(DiagnosticExpressionMustBeNumeric())
		}
	case ParameterTypeNumberOrBool:
		if !isNumericSymbol(argumentType) && !sameType(argumentType, ScalarTypes.Bool.Type()) {
			return at(DiagnosticExpressionMustBeNumericOrBool())
		}
	case ParameterTypeSummable:
		if !scalarHasFlag(argumentType, ScalarFlagSummable) {
			return at(DiagnosticExpressionMustBeSummable())
		}
	case ParameterTypeOrderable:
		if !scalarHasFlag(argumentType, ScalarFlagOrderable) {
			return at(DiagnosticExpressionMustBeOrderable())
		}
	case ParameterTypeNotBool:
		if !isScalar {
			return at(DiagnosticScalarTypeExpected())
		}
		if SymbolAssignableTo(argumentType, ScalarTypes.Bool.Type()) {
			return at(DiagnosticTypeNotAllowed(argumentType.SymbolName()))
		}
	case ParameterTypeNotRealOrBool:
		if !isScalar {
			return at(DiagnosticScalarTypeExpected())
		}
		if SymbolAssignableTo(argumentType, ScalarTypes.Real.Type()) || SymbolAssignableTo(argumentType, ScalarTypes.Bool.Type()) {
			return at(DiagnosticTypeNotAllowed(argumentType.SymbolName()))
		}
	case ParameterTypeNotDynamic:
		if !isScalar {
			return at(DiagnosticScalarTypeExpected())
		}
		if isDynamicTypedSymbol(argumentType) {
			return at(DiagnosticTypeNotAllowed(argumentType.SymbolName()))
		}
	case ParameterTypeTabular:
		if !IsTabular(argumentType) {
			return at(DiagnosticTabularValueExpected())
		}
	case ParameterTypeDatabase:
		if argumentType.SymbolKind() != SymbolKindDatabase {
			return at(DiagnosticDatabaseExpected())
		}
	case ParameterTypeCluster:
		if argumentType.SymbolKind() != SymbolKindCluster {
			return at(DiagnosticClusterExpected())
		}
	case ParameterTypeParameter0, ParameterTypeParameter1, ParameterTypeParameter2:
		index := int(parameter.TypeKind[len(parameter.TypeKind)-1] - '0')
		if index >= 0 && index < len(argumentTypes) {
			target := returnTypeFromSymbol(argumentTypes[index])
			if target != nil && !isErrorTypeSymbol(target) && !SymbolAssignableTo(argumentType, target) {
				return at(DiagnosticExpressionMustHaveType(target.SymbolName()))
			}
		}
	case ParameterTypeCommonScalar, ParameterTypeCommonNumber, ParameterTypeCommonSummable, ParameterTypeCommonOrderable, ParameterTypeCommonScalarOrDynamic:
		if category := functionCommonArgumentCategoryDiagnostic(parameter.TypeKind, argumentType); category != nil {
			return at(*category)
		}
		if isErrorTypeSymbol(argumentType) || sameType(argumentType, ScalarTypes.Unknown.Type()) {
			return nil
		}
		common := commonArgumentType(argumentParameters, argumentTypes)
		commonType := returnTypeFromSymbol(common)
		if parameter.TypeKind == ParameterTypeCommonScalarOrDynamic && isDynamicType(commonType) {
			return nil
		}
		if commonType != nil && !SymbolAssignableTo(argumentType, commonType, ConversionPromotable) {
			types := []string{commonType.SymbolName()}
			if parameter.TypeKind == ParameterTypeCommonScalarOrDynamic && !sameType(commonType, ScalarTypes.Dynamic.Type()) {
				types = append(types, "dynamic")
			}
			return at(DiagnosticExpressionMustHaveType(types...))
		}
	}
	return nil
}

func parameterDeclaresExactDynamic(parameter *Parameter) bool {
	return parameter != nil && parameter.TypeKind == ParameterTypeDeclared && len(parameter.DeclaredTypes) == 1 &&
		parameter.DeclaredTypes[0] == ScalarTypes.Dynamic.Type()
}

func declaredParameterTable(parameter *Parameter) (*TableSymbol, bool) {
	if parameter == nil || parameter.TypeKind != ParameterTypeDeclared || len(parameter.DeclaredSymbols) != 1 {
		return nil, false
	}
	return tableSymbolFromSymbol(parameter.DeclaredSymbols[0])
}

func (n *simpleNodeBinder) allowLooseFunctionParameterMatching(signature *Signature) bool {
	if signature == nil {
		return false
	}
	function, ok := signature.Symbol.(*FunctionSymbol)
	if !ok || function == nil {
		return false
	}
	signatures := function.Signatures()
	isLocalFunction := len(signatures) != 0 && signatures[0].Declaration != nil
	return n.binder.Globals.IsDatabaseFunction(function) || isLocalFunction
}

func (n *simpleNodeBinder) getBestMatchingFunctionSignatures(signatures []*Signature, arguments []Expression, argumentTypes []Symbol, requireAllArgumentsMatch bool) []*Signature {
	return getBestMatchingSignatures(signatures, arguments, argumentTypes, requireAllArgumentsMatch, n.allowLooseFunctionParameterMatching)
}

func parameterTypeNames(parameter *Parameter) []string {
	var names []string
	for _, typ := range parameter.DeclaredTypes {
		if typ != nil {
			names = append(names, typ.SymbolName())
		}
	}
	if len(names) == 0 {
		names = append(names, "unknown")
	}
	return names
}

func functionCommonArgumentCategoryDiagnostic(kind ParameterTypeKind, argumentType Symbol) *Diagnostic {
	var diagnostic Diagnostic
	switch kind {
	case ParameterTypeCommonScalar, ParameterTypeCommonScalarOrDynamic:
		if IsScalar(argumentType) {
			return nil
		}
		diagnostic = DiagnosticScalarTypeExpected()
	case ParameterTypeCommonNumber:
		if isNumericSymbol(argumentType) {
			return nil
		}
		diagnostic = DiagnosticExpressionMustBeNumeric()
	case ParameterTypeCommonSummable:
		if scalarHasFlag(argumentType, ScalarFlagSummable) {
			return nil
		}
		diagnostic = DiagnosticExpressionMustBeSummable()
	case ParameterTypeCommonOrderable:
		if scalarHasFlag(argumentType, ScalarFlagOrderable) {
			return nil
		}
		diagnostic = DiagnosticExpressionMustBeOrderable()
	default:
		return nil
	}
	return &diagnostic
}

func functionLiteralValueAllowed(value any, parameter *Parameter) bool {
	want := literalDiagnosticValue(value)
	for _, candidate := range parameter.Values {
		got := literalDiagnosticValue(candidate)
		if parameter.IsCaseSensitive && got == want || !parameter.IsCaseSensitive && strings.EqualFold(got, want) {
			return true
		}
	}
	return false
}

func literalDiagnosticValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(valueToString(value)), "\"", ""))
}

func functionArgumentUnderlyingExpression(expression Expression) Expression {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return functionArgumentUnderlyingExpression(expression.Expression())
	case *OrderedExpression:
		return functionArgumentUnderlyingExpression(expression.Expression())
	default:
		return expression
	}
}
