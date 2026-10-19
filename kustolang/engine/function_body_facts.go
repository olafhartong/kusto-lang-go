package kustolang

// Ported from src/Kusto.Language/FunctionBodyFacts.cs. The Go model keeps the
// same immutable-by-return convention used by the rest of this port.

type functionBodyFactsFlag uint16

const (
	functionBodyFactsNone    functionBodyFactsFlag = 0
	functionBodyFactsCluster functionBodyFactsFlag = 1 << iota
	functionBodyFactsDatabase
	functionBodyFactsUnqualifiedTable
	functionBodyFactsQualifiedTable
	functionBodyFactsExternalTable
	functionBodyFactsMaterializedView
	functionBodyFactsStoredQueryResult
	functionBodyFactsGraph
)

type FunctionBodyFacts struct {
	flags                 functionBodyFactsFlag
	nonVariableReturnType *TypeSymbol
	dependentParameters   []*Parameter
	unqualifiedTableNames []string
	hasSyntaxErrors       bool
}

var FunctionBodyFactsDefault = newFunctionBodyFacts(functionBodyFactsNone, nil, false, nil, nil)

func newFunctionBodyFacts(flags functionBodyFactsFlag, nonVariableReturnType *TypeSymbol, hasSyntaxErrors bool, dependentParameters []*Parameter, unqualifiedTableNames []string) *FunctionBodyFacts {
	return &FunctionBodyFacts{
		flags:                 flags,
		nonVariableReturnType: nonVariableReturnType,
		hasSyntaxErrors:       hasSyntaxErrors,
		dependentParameters:   uniqueParameters(dependentParameters),
		unqualifiedTableNames: uniqueStrings(unqualifiedTableNames),
	}
}

func (f *FunctionBodyFacts) NonVariableReturnType() *TypeSymbol {
	return functionBodyFactsOrDefault(f).nonVariableReturnType
}

func (f *FunctionBodyFacts) DependentParameters() []*Parameter {
	return append([]*Parameter{}, functionBodyFactsOrDefault(f).dependentParameters...)
}

func (f *FunctionBodyFacts) UnqualifiedTableNames() []string {
	return append([]string{}, functionBodyFactsOrDefault(f).unqualifiedTableNames...)
}

func (f *FunctionBodyFacts) HasSyntaxErrors() bool {
	return functionBodyFactsOrDefault(f).hasSyntaxErrors
}

func functionBodyFactsOrDefault(facts *FunctionBodyFacts) *FunctionBodyFacts {
	if facts == nil {
		return FunctionBodyFactsDefault
	}
	return facts
}

func (f *FunctionBodyFacts) with(flags functionBodyFactsFlag, nonVariableReturnType *TypeSymbol, hasSyntaxErrors bool, dependentParameters []*Parameter, unqualifiedTableNames []string) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	if f.flags == flags &&
		f.nonVariableReturnType == nonVariableReturnType &&
		f.hasSyntaxErrors == hasSyntaxErrors &&
		sameParameterSlice(f.dependentParameters, dependentParameters) &&
		sameStringSlice(f.unqualifiedTableNames, unqualifiedTableNames) {
		return f
	}
	return newFunctionBodyFacts(flags, nonVariableReturnType, hasSyntaxErrors, dependentParameters, unqualifiedTableNames)
}

func (f *FunctionBodyFacts) WithNonVariableReturnType(typ *TypeSymbol) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	return f.with(f.flags, typ, f.hasSyntaxErrors, f.dependentParameters, f.unqualifiedTableNames)
}

func (f *FunctionBodyFacts) WithHasSyntaxErrors(value bool) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	return f.with(f.flags, f.nonVariableReturnType, value, f.dependentParameters, f.unqualifiedTableNames)
}

func (f *FunctionBodyFacts) WithDependentParameters(parameters []*Parameter) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	unique := uniqueParameters(parameters)
	return f.with(f.flags, f.nonVariableReturnType, f.hasSyntaxErrors, unique, f.unqualifiedTableNames)
}

func (f *FunctionBodyFacts) AddDependentParameter(parameter *Parameter) *FunctionBodyFacts {
	return f.AddDependentParameters([]*Parameter{parameter})
}

func (f *FunctionBodyFacts) AddDependentParameters(parameters []*Parameter) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	merged := append([]*Parameter{}, f.dependentParameters...)
	changed := false
	for _, parameter := range parameters {
		if containsParameter(merged, parameter) {
			continue
		}
		merged = append(merged, parameter)
		changed = true
	}
	if !changed {
		return f
	}
	return f.with(f.flags, f.nonVariableReturnType, f.hasSyntaxErrors, merged, f.unqualifiedTableNames)
}

func (f *FunctionBodyFacts) WithUnqualifiedTableNames(names []string) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	unique := uniqueStrings(names)
	return f.with(f.flags, f.nonVariableReturnType, f.hasSyntaxErrors, f.dependentParameters, unique)
}

func (f *FunctionBodyFacts) AddUnqualifiedTableName(name string) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	names := append([]string{}, f.unqualifiedTableNames...)
	if containsString(names, name) {
		return f
	}
	names = append(names, name)
	return f.with(f.flags, f.nonVariableReturnType, f.hasSyntaxErrors, f.dependentParameters, names)
}

func (f *FunctionBodyFacts) HasVariableReturnType() bool {
	f = functionBodyFactsOrDefault(f)
	return len(f.dependentParameters) > 0 || f.HasUnqualifiedTableCall()
}

func (f *FunctionBodyFacts) HasClusterCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsCluster)
}

func (f *FunctionBodyFacts) WithHasClusterCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsCluster, value)
}

func (f *FunctionBodyFacts) HasDatabaseCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsDatabase)
}

func (f *FunctionBodyFacts) WithHasDatabaseCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsDatabase, value)
}

func (f *FunctionBodyFacts) HasUnqualifiedTableCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsUnqualifiedTable)
}

func (f *FunctionBodyFacts) WithHasUnqualifiedTableCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsUnqualifiedTable, value)
}

func (f *FunctionBodyFacts) HasQualifiedTableCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsQualifiedTable)
}

func (f *FunctionBodyFacts) WithHasQualifiedTableCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsQualifiedTable, value)
}

func (f *FunctionBodyFacts) HasExternalTableCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsExternalTable)
}

func (f *FunctionBodyFacts) WithHasExternalTableCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsExternalTable, value)
}

func (f *FunctionBodyFacts) HasMaterializedViewCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsMaterializedView)
}

func (f *FunctionBodyFacts) WithHasMaterializedViewCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsMaterializedView, value)
}

func (f *FunctionBodyFacts) HasStoredQueryResultCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsStoredQueryResult)
}

func (f *FunctionBodyFacts) WithHasStoredQueryResultCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsStoredQueryResult, value)
}

func (f *FunctionBodyFacts) HasGraphCall() bool {
	return functionBodyFactsOrDefault(f).hasFlag(functionBodyFactsGraph)
}

func (f *FunctionBodyFacts) WithHasGraphCall(value bool) *FunctionBodyFacts {
	return f.withFlag(functionBodyFactsGraph, value)
}

func (f *FunctionBodyFacts) IsInteresting() bool {
	return functionBodyFactsOrDefault(f).flags != functionBodyFactsNone
}

func (f *FunctionBodyFacts) WithIsInteresting(value bool) *FunctionBodyFacts {
	return functionBodyFactsOrDefault(f)
}

func (f *FunctionBodyFacts) CombineCalledFunction(facts *FunctionBodyFacts) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	facts = functionBodyFactsOrDefault(facts)
	names := append([]string{}, f.unqualifiedTableNames...)
	for _, name := range facts.unqualifiedTableNames {
		if !containsString(names, name) {
			names = append(names, name)
		}
	}
	return f.with(f.flags|facts.flags, f.nonVariableReturnType, f.hasSyntaxErrors, f.dependentParameters, names)
}

// ComputeFunctionBodyFacts ports the Binder_FunctionCalls.cs function-body
// fact scan for the semantic payload that is already represented in Go. It
// records source-lookup calls, dependent outer parameters and syntax errors;
// full computed-function expansion is still handled by future tree-binder work.
func ComputeFunctionBodyFacts(signature *Signature, body SyntaxElement) *FunctionBodyFacts {
	result := FunctionBodyFactsDefault
	isTabular := functionBodyIsTabular(signature, body)

	if isTabular && signature != nil {
		for _, parameter := range signature.Parameters {
			if parameter != nil && parameter.IsTabular() {
				result = result.AddDependentParameter(parameter)
			}
		}
	}

	walkFunctionBodyFacts(body, nil, body, func(element SyntaxElement, parent SyntaxElement) {
		wrapped := WrapGeneratedSyntaxElement(element)
		switch node := wrapped.(type) {
		case *FunctionCallExpression:
			result = computeFunctionCallFacts(result, signature, node, parent, isTabular)
		case *ProjectByNamesOperator:
			result = AddReferencedParametersAsDependentParameters(result, signature, node)
		default:
			expression, ok := wrapped.(Expression)
			if !ok {
				return
			}
			referencedSignature := expressionReferencedSignature(expression)
			if referencedSignature == nil || IsSymbolLookupFunction(referencedSignature.Symbol) {
				return
			}
			facts := functionBodyFactsForExpression(expression)
			result = result.CombineCalledFunction(facts)
			if isTabular {
				if call, ok := expression.(*FunctionCallExpression); ok && len(facts.dependentParameters) > 0 {
					result = addCalledDependentArguments(result, signature, call, facts)
				}
			}
		}
	}, func(element SyntaxElement) {
		if semantic, ok := WrapGeneratedSyntaxElement(element).(interface{ Alternates() []*SyntaxNode }); ok {
			for _, alternate := range semantic.Alternates() {
				result = result.CombineCalledFunction(ComputeFunctionBodyFacts(signature, alternate))
			}
		}
	})

	if !result.HasVariableReturnType() {
		result = result.WithNonVariableReturnType(functionBodyResultType(body))
	}
	result = result.WithHasSyntaxErrors(HasSyntaxErrors(body))
	return result
}

func GetOrComputeFunctionBodyFacts(globals *GlobalState, global *GlobalBindingCache, local *LocalBindingCache, signature *Signature, body SyntaxElement) *FunctionBodyFacts {
	if facts, ok := TryGetFunctionBodyFacts(global, local, signature); ok {
		return facts
	}
	facts := ComputeFunctionBodyFacts(signature, body)
	SetFunctionBodyFacts(globals, global, local, signature, facts)
	return facts
}

func AddReferencedParametersAsDependentParameters(facts *FunctionBodyFacts, signature *Signature, root SyntaxElement) *FunctionBodyFacts {
	return functionBodyFactsOrDefault(facts).AddDependentParameters(GetReferencedParameters(signature, root))
}

func GetReferencedParameters(signature *Signature, root SyntaxElement) []*Parameter {
	if signature == nil || root == nil {
		return []*Parameter{}
	}
	var parameters []*Parameter
	WalkSyntax(root, func(element SyntaxElement) {
		wrapped := WrapGeneratedSyntaxElement(element)
		if named, ok := wrapped.(*SimpleNamedExpression); ok {
			wrapped = named.Expression()
		}
		expression, ok := wrapped.(Expression)
		if !ok {
			return
		}
		switch symbol := expressionReferencedSymbol(expression).(type) {
		case *ParameterSymbol:
			parameters = addReferencedParameterByName(parameters, signature, symbol.SymbolName())
		case *VariableSymbol:
			if symbol.Source != nil {
				if expression.Tree() != symbol.Source.Tree() {
					// Computed-function expansion binds parameters as variables whose
					// source is an argument in the calling tree. Follow the source only
					// for ordinary same-tree lets; a cross-tree source represents this
					// signature parameter, matching the upstream binder distinction.
					parameters = addReferencedParameterByName(parameters, signature, symbol.SymbolName())
				} else {
					parameters = append(parameters, GetReferencedParameters(signature, symbol.Source)...)
				}
			}
		}
	})
	return uniqueParameters(parameters)
}

func IsSymbolLookupFunction(symbol Symbol) bool {
	return symbol == Functions.Table ||
		symbol == Functions.ExternalTable ||
		symbol == Functions.MaterializedView ||
		symbol == Functions.EntityGroup ||
		symbol == Functions.StoredQueryResult ||
		symbol == Functions.Database ||
		symbol == Functions.Cluster ||
		symbol == Functions.Graph
}

func HasSyntaxErrors(root SyntaxElement) bool {
	for _, diagnostic := range GetContainedDiagnostics(root, DiagnosticsIncludeSyntactic) {
		if diagnostic.Severity == DiagnosticSeverityError {
			return true
		}
	}
	return false
}

func (f *FunctionBodyFacts) hasFlag(flag functionBodyFactsFlag) bool {
	return f.flags&flag != 0
}

func (f *FunctionBodyFacts) withFlag(flag functionBodyFactsFlag, value bool) *FunctionBodyFacts {
	f = functionBodyFactsOrDefault(f)
	flags := f.flags
	if value {
		flags |= flag
	} else {
		flags &^= flag
	}
	return f.with(flags, f.nonVariableReturnType, f.hasSyntaxErrors, f.dependentParameters, f.unqualifiedTableNames)
}

func computeFunctionCallFacts(facts *FunctionBodyFacts, signature *Signature, call *FunctionCallExpression, parent SyntaxElement, isTabular bool) *FunctionBodyFacts {
	facts = functionBodyFactsOrDefault(facts)
	symbol := expressionReferencedSymbol(call)
	if !IsSymbolLookupFunction(symbol) {
		calledFacts := functionBodyFactsForExpression(call)
		facts = facts.CombineCalledFunction(calledFacts)
		if isTabular && len(calledFacts.dependentParameters) > 0 {
			facts = addCalledDependentArguments(facts, signature, call, calledFacts)
		}
		return facts
	}

	isUnqualifiedTableCall := false

	switch symbol {
	case Functions.Table:
		if parentPath, ok := WrapGeneratedSyntaxElement(parent).(*PathExpression); ok && sameSyntaxExpression(parentPath.Selector(), call) {
			facts = facts.WithHasQualifiedTableCall(true)
		} else {
			facts = facts.WithHasUnqualifiedTableCall(true)
			isUnqualifiedTableCall = true
		}
	case Functions.ExternalTable:
		facts = facts.WithHasExternalTableCall(true)
	case Functions.MaterializedView:
		facts = facts.WithHasMaterializedViewCall(true)
	case Functions.Database:
		facts = facts.WithHasDatabaseCall(true)
	case Functions.Cluster:
		facts = facts.WithHasClusterCall(true)
	case Functions.EntityGroup:
		if calledFacts := call.GetCalledFunctionFacts(); calledFacts != nil {
			facts = facts.CombineCalledFunction(calledFacts)
		}
	case Functions.StoredQueryResult:
		facts = facts.WithHasStoredQueryResultCall(true)
	case Functions.Graph:
		facts = facts.WithHasGraphCall(true)
	}

	for _, argument := range functionCallArguments(call) {
		facts = AddReferencedParametersAsDependentParameters(facts, signature, argument)
	}
	if isUnqualifiedTableCall && isTabular {
		if arguments := functionCallArguments(call); len(arguments) > 0 {
			if tableName, ok := ExpressionConstantValue(arguments[0]).(string); ok {
				facts = facts.AddUnqualifiedTableName(tableName)
			}
		}
	}
	return facts
}

func addCalledDependentArguments(facts *FunctionBodyFacts, signature *Signature, call *FunctionCallExpression, calledFacts *FunctionBodyFacts) *FunctionBodyFacts {
	arguments := functionCallArguments(call)
	if len(arguments) == 0 || call == nil || signature == nil {
		return functionBodyFactsOrDefault(facts)
	}
	calledSignature := expressionReferencedSignature(call)
	if calledSignature == nil {
		return functionBodyFactsOrDefault(facts)
	}
	argumentParameters := calledSignature.ArgumentParametersForExpressions(arguments, nil)
	for i, argument := range arguments {
		if i < len(argumentParameters) && containsParameter(calledFacts.dependentParameters, argumentParameters[i]) {
			facts = AddReferencedParametersAsDependentParameters(facts, signature, argument)
		}
	}
	return facts
}

func functionCallArguments(call *FunctionCallExpression) []Expression {
	if call == nil || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
		return nil
	}
	expressions := call.ArgumentList().Expressions()
	out := make([]Expression, 0, expressions.Count())
	for i := 0; i < expressions.Count(); i++ {
		if item := expressions.At(i); item != nil && item.Element() != nil {
			out = append(out, item.Element())
		}
	}
	return out
}

func functionBodyFactsForExpression(expression Expression) *FunctionBodyFacts {
	if accessor, ok := expression.(interface{ GetCalledFunctionFacts() *FunctionBodyFacts }); ok {
		if facts := accessor.GetCalledFunctionFacts(); facts != nil {
			return facts
		}
	}
	return FunctionBodyFactsDefault
}

func functionBodyIsTabular(signature *Signature, body SyntaxElement) bool {
	if signature != nil && signature.Tabularity == TabularityTabular {
		return true
	}
	return IsTabular(functionBodyResultSymbol(body))
}

func functionBodyResultSymbol(body SyntaxElement) Symbol {
	switch node := WrapGeneratedSyntaxElement(body).(type) {
	case *FunctionBody:
		return functionBodyResultSymbol(node.Expression())
	case Expression:
		if call, ok := node.(*FunctionCallExpression); ok {
			if result, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
				if symbol, ok := result.(Symbol); ok && symbol != nil {
					return symbol
				}
			}
		}
		if symbol := expressionReferencedSymbol(node); symbol != nil && IsTabular(GetResultType(symbol)) {
			return GetResultType(symbol)
		}
		if symbol := expressionReferencedSymbol(node); symbol != nil && IsTabular(symbol) {
			return symbol
		}
		if typ := ExpressionResultType(node); typ != nil {
			return typ
		}
	}
	return nil
}

func functionBodyResultType(body SyntaxElement) *TypeSymbol {
	switch result := functionBodyResultSymbol(body).(type) {
	case nil:
		return &Error.TypeSymbol
	case *TypeSymbol:
		return result
	default:
		if typ := returnTypeFromSymbol(result); typ != nil {
			return typ
		}
	}
	return &Error.TypeSymbol
}

func expressionReferencedSymbol(expression Expression) Symbol {
	if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
		return accessor.ReferencedSymbol()
	}
	return nil
}

func expressionReferencedSignature(expression Expression) *Signature {
	if accessor, ok := expression.(interface{ ReferencedSignature() *Signature }); ok {
		return accessor.ReferencedSignature()
	}
	return nil
}

func addReferencedParameterByName(parameters []*Parameter, signature *Signature, name string) []*Parameter {
	if parameter := signature.GetParameter(name); parameter != nil && !containsParameter(parameters, parameter) {
		parameters = append(parameters, parameter)
	}
	return parameters
}

func walkFunctionBodyFacts(root SyntaxElement, parent SyntaxElement, original SyntaxElement, before func(SyntaxElement, SyntaxElement), after func(SyntaxElement)) {
	if root == nil {
		return
	}
	if before != nil {
		before(root, parent)
	}
	for _, child := range root.Children() {
		if child == nil {
			continue
		}
		if child != original {
			switch WrapGeneratedSyntaxElement(child).(type) {
			case *FunctionDeclaration, *FunctionBody:
				continue
			}
		}
		walkFunctionBodyFacts(child, root, original, before, after)
	}
	if after != nil {
		after(root)
	}
}

func uniqueParameters(parameters []*Parameter) []*Parameter {
	out := []*Parameter{}
	for _, parameter := range parameters {
		if !containsParameter(out, parameter) {
			out = append(out, parameter)
		}
	}
	if out == nil {
		return []*Parameter{}
	}
	return out
}

func containsParameter(parameters []*Parameter, target *Parameter) bool {
	for _, parameter := range parameters {
		if parameter == target {
			return true
		}
	}
	return false
}

func sameParameterSlice(left []*Parameter, right []*Parameter) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func uniqueStrings(values []string) []string {
	out := []string{}
	for _, value := range values {
		if !containsString(out, value) {
			out = append(out, value)
		}
	}
	if out == nil {
		return []string{}
	}
	return out
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sameStringSlice(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
