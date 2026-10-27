package kustolang

import (
	"sort"
	"strings"
)

// Ported from the function-argument completion path in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type functionArgumentCompletionContext struct {
	call          *FunctionCallExpression
	function      *FunctionSymbol
	operator      *OperatorSymbol
	signatures    []*Signature
	arguments     []Expression
	argumentIndex int
	parameterName string
}

func (c *kustoCompleter) getFunctionArgumentCompletions(position int, builder *kustoCompletionBuilder) completionMode {
	context, ok := c.functionArgumentCompletionContext(position)
	if !ok {
		context, ok = c.binaryOperatorArgumentCompletionContext(position)
	}
	if !ok {
		return completionModeCombined
	}

	if context.function != nil && !c.code.Globals.IsBuiltInFunction(context.function) && context.function.MaxArgumentCount() >= 2 {
		suggestions := functionArgumentNameSuggestions(context.signatures, context.arguments, context.argumentIndex, c.code.Globals)
		for _, name := range suggestions {
			builder.add(NewCompletionItem(CompletionKindParameter, name+"=",
				WithCompletionBeforeText(name+"="),
				WithCompletionMatchText(name)))
		}
		if len(suggestions) > 0 && functionArgumentNamesRequired(context.signatures, context.arguments, context.argumentIndex, c.code.Globals) {
			return completionModeIsolated
		}
	}
	for _, signature := range context.signatures {
		if signature != nil && c.addReturnTypeKindCompletions(signature.ReturnKind, context, position, builder) {
			return completionModeIsolated
		}
	}

	possible := functionPossibleParameters(context, c.code.Globals)
	if len(possible) == 0 {
		return completionModeIsolated
	}

	examples := map[string]struct{}{}
	literalOnly := true
	for _, candidate := range possible {
		parameter := candidate.parameter
		if parameter == nil || parameter == UnknownParameter {
			continue
		}
		if parameter.ArgumentKind != ArgumentKindLiteral && parameter.ArgumentKind != ArgumentKindLiteralNotEmpty {
			literalOnly = false
		}
		c.addFunctionParameterExamples(candidate.signature, parameter, context.arguments, position, examples)
	}
	for example := range examples {
		builder.add(NewCompletionItem(CompletionKindExample, example))
	}
	if literalOnly {
		return completionModeIsolated
	}
	return completionModeCombined
}

func (c *kustoCompleter) binaryOperatorArgumentCompletionContext(position int) (functionArgumentCompletionContext, bool) {
	if c == nil || c.code == nil || c.code.Syntax() == nil || position < 0 || position > len(c.code.Text) {
		return functionArgumentCompletionContext{}, false
	}
	var selected *BinaryExpression
	selectedStart := -1
	walkSyntaxWithTriviaStartAndAncestors(c.code.Syntax(), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		binary, ok := WrapGeneratedSyntaxElement(element).(*BinaryExpression)
		if !ok || binary == nil || binary.Operator() == nil || binary.Left() == nil || binary.Right() == nil {
			return
		}
		operator := binary.Operator()
		if operator.Width() == 0 || position < operator.End() {
			return
		}
		end := binary.End()
		if end < 0 || end > len(c.code.Text) || end > position {
			return
		}
		if end < position && strings.TrimSpace(c.code.Text[end:position]) != "" {
			return
		}
		if operator.TextStart() >= selectedStart {
			selected = binary
			selectedStart = operator.TextStart()
		}
	})
	if selected == nil {
		return functionArgumentCompletionContext{}, false
	}
	operator, _ := selected.ReferencedSymbol().(*OperatorSymbol)
	if operator == nil {
		operator = completionOperatorForSyntaxKind(selected.Kind())
	}
	if operator == nil || len(operator.Signatures()) == 0 {
		return functionArgumentCompletionContext{}, false
	}
	return functionArgumentCompletionContext{
		operator:      operator,
		signatures:    operator.Signatures(),
		arguments:     []Expression{selected.Left(), selected.Right()},
		argumentIndex: 1,
	}, true
}

func completionOperatorForSyntaxKind(kind SyntaxKind) *OperatorSymbol {
	operatorKind := operatorKindForSyntax(kind)
	if operatorKind == OperatorKindNone {
		return nil
	}
	for _, operator := range Operators.All {
		if operator != nil && operator.OperatorKind == operatorKind {
			return operator
		}
	}
	return nil
}

func (c *kustoCompleter) functionArgumentCompletionContext(position int) (functionArgumentCompletionContext, bool) {
	if c == nil || c.code == nil || c.code.Syntax() == nil || position < 0 || position > len(c.code.Text) {
		return functionArgumentCompletionContext{}, false
	}
	var selected *FunctionCallExpression
	selectedStart := -1
	walkSyntaxWithTriviaStartAndAncestors(c.code.Syntax(), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call == nil || call.ArgumentList() == nil || call.ArgumentList().OpenParen() == nil {
			return
		}
		list := call.ArgumentList()
		open := list.OpenParen()
		if position < open.End() {
			return
		}
		if close := list.CloseParen(); close != nil && position > close.TextStart() {
			if close.Width() > 0 || close.TextStart() < 0 || close.TextStart() > len(c.code.Text) || strings.TrimSpace(c.code.Text[close.TextStart():position]) != "" {
				return
			}
		}
		if open.TextStart() >= selectedStart {
			selected = call
			selectedStart = open.TextStart()
		}
	})
	if selected == nil {
		return functionArgumentCompletionContext{}, false
	}
	function, ok := selected.ReferencedSymbol().(*FunctionSymbol)
	if !ok || function == nil {
		if selected.Name() != nil {
			function, _ = selected.Name().ReferencedSymbol().(*FunctionSymbol)
		}
	}
	if function == nil {
		function = c.resolveCompletionFunction(selected)
	}
	if function == nil || len(function.Signatures()) == 0 {
		return functionArgumentCompletionContext{}, false
	}

	arguments := functionCallCompletionArguments(selected)
	argumentIndex := functionCallCompletionArgumentIndex(selected, position)
	parameterName := ""
	if argumentIndex >= 0 && argumentIndex < len(arguments) {
		if named, ok := arguments[argumentIndex].(*SimpleNamedExpression); ok && named.EqualToken() != nil && position >= named.EqualToken().End() {
			parameterName = NameDeclarationSimpleName(named.Name())
		}
	}
	return functionArgumentCompletionContext{
		call:          selected,
		function:      function,
		signatures:    function.Signatures(),
		arguments:     arguments,
		argumentIndex: argumentIndex,
		parameterName: parameterName,
	}, true
}

func (c *kustoCompleter) addReturnTypeKindCompletions(kind ReturnTypeKind, context functionArgumentCompletionContext, position int, builder *kustoCompletionBuilder) bool {
	if c == nil || c.code == nil || c.code.Globals == nil {
		return false
	}
	var symbols []Symbol
	switch kind {
	case ReturnTypeParameter0Cluster:
		if context.argumentIndex != 0 {
			return false
		}
		for _, cluster := range c.code.Globals.Clusters {
			symbols = append(symbols, cluster)
		}
	case ReturnTypeParameter0Database:
		if context.argumentIndex != 0 {
			return false
		}
		cluster := c.completionClusterForFunctionCall(context.call)
		if cluster == nil {
			cluster = c.code.Globals.Cluster
		}
		if cluster != nil {
			for _, database := range cluster.Databases() {
				symbols = append(symbols, database)
			}
		}
	case ReturnTypeParameter0Table:
		if context.argumentIndex != 0 {
			return false
		}
		if database := c.completionDatabaseForFunctionCall(context.call); database != nil {
			for _, table := range database.Tables() {
				symbols = append(symbols, table)
			}
		}
	case ReturnTypeParameter0ExternalTable:
		if context.argumentIndex != 0 {
			return false
		}
		if database := c.completionDatabaseForFunctionCall(context.call); database != nil {
			for _, table := range database.ExternalTables() {
				symbols = append(symbols, table)
			}
		}
	case ReturnTypeParameter0MaterializedView:
		if context.argumentIndex != 0 {
			return false
		}
		if database := c.completionDatabaseForFunctionCall(context.call); database != nil {
			for _, view := range database.MaterializedViews() {
				symbols = append(symbols, view)
			}
		}
	case ReturnTypeParameter0EntityGroup:
		if context.argumentIndex != 0 {
			return false
		}
		if database := c.completionDatabaseForFunctionCall(context.call); database != nil {
			for _, group := range database.EntityGroups() {
				symbols = append(symbols, group)
			}
		}
	case ReturnTypeParameter0StoredQueryResult:
		if context.argumentIndex != 0 {
			return false
		}
		if database := c.completionDatabaseForFunctionCall(context.call); database != nil {
			for _, result := range database.StoredQueryResults() {
				symbols = append(symbols, result)
			}
		}
	case ReturnTypeParameter0Graph:
		database := c.completionDatabaseForFunctionCall(context.call)
		if database == nil {
			return false
		}
		switch context.argumentIndex {
		case 0:
			for _, model := range database.GraphModels() {
				symbols = append(symbols, model)
			}
		case 1:
			if len(context.arguments) > 0 {
				if modelName, ok := ExpressionConstantValue(context.arguments[0]).(string); ok {
					if model := database.GetGraphModel(modelName); model != nil {
						for _, snapshot := range model.Snapshots {
							symbols = append(symbols, snapshot)
						}
					}
				}
			}
		default:
			return false
		}
	default:
		return false
	}
	c.addMemberNameExamples(symbols, builder)
	c.addMatchingStringSymbolCompletions(position, builder)
	return true
}

func (c *kustoCompleter) completionDatabaseForFunctionCall(call *FunctionCallExpression) *DatabaseSymbol {
	if receiver := functionCallCompletionReceiver(call); receiver != nil {
		if database := c.completionDatabaseForExpression(receiver); database != nil {
			return database
		}
	}
	if c != nil && c.code != nil && c.code.Globals != nil {
		return c.code.Globals.Database
	}
	return nil
}

func (c *kustoCompleter) completionClusterForFunctionCall(call *FunctionCallExpression) *ClusterSymbol {
	if receiver := functionCallCompletionReceiver(call); receiver != nil {
		if cluster := c.completionClusterForExpression(receiver); cluster != nil {
			return cluster
		}
	}
	if c != nil && c.code != nil && c.code.Globals != nil {
		return c.code.Globals.Cluster
	}
	return nil
}

func (c *kustoCompleter) completionClusterForExpression(expression Expression) *ClusterSymbol {
	if cluster, ok := expressionReferencedSymbol(expression).(*ClusterSymbol); ok {
		return cluster
	}
	call, ok := expression.(*FunctionCallExpression)
	if !ok || c.resolveCompletionFunction(call) != Functions.Cluster || c.code == nil || c.code.Globals == nil {
		return nil
	}
	arguments := functionCallCompletionArguments(call)
	if len(arguments) == 0 {
		return nil
	}
	name, ok := ExpressionConstantValue(arguments[0]).(string)
	if !ok {
		return nil
	}
	return c.code.Globals.GetCluster(name)
}

func (c *kustoCompleter) completionDatabaseForExpression(expression Expression) *DatabaseSymbol {
	if database, ok := expressionReferencedSymbol(expression).(*DatabaseSymbol); ok {
		return database
	}
	path, ok := expression.(*PathExpression)
	if !ok || path.Selector() == nil {
		return nil
	}
	call, ok := path.Selector().(*FunctionCallExpression)
	if !ok || c.resolveCompletionFunction(call) != Functions.Database {
		return nil
	}
	cluster := c.completionClusterForExpression(path.Expression())
	if cluster == nil {
		return nil
	}
	arguments := functionCallCompletionArguments(call)
	if len(arguments) == 0 {
		if c.code != nil && c.code.Globals != nil && cluster == c.code.Globals.Cluster {
			return c.code.Globals.Database
		}
		return nil
	}
	name, ok := ExpressionConstantValue(arguments[0]).(string)
	if !ok {
		return nil
	}
	return cluster.GetDatabase(name)
}

func functionCallCompletionReceiver(call *FunctionCallExpression) Expression {
	if call == nil || call.Parent() == nil {
		return nil
	}
	path, ok := WrapGeneratedSyntaxElement(call.Parent()).(*PathExpression)
	if !ok || path == nil || path.Selector() == nil {
		return nil
	}
	selector, ok := path.Selector().(*FunctionCallExpression)
	if !ok || selector == nil || selector.SyntaxNode != call.SyntaxNode {
		return nil
	}
	return path.Expression()
}

func (c *kustoCompleter) addMemberNameExamples(symbols []Symbol, builder *kustoCompletionBuilder) {
	for _, symbol := range symbols {
		if symbol == nil || symbol.SymbolName() == "" || symbol.IsHidden() {
			continue
		}
		builder.add(NewCompletionItem(CompletionKindExample, GetSingleQuotedStringLiteral(symbol.SymbolName())))
		if alternate := symbol.AlternateName(); alternate != "" {
			builder.add(NewCompletionItem(CompletionKindExample, GetSingleQuotedStringLiteral(alternate)))
		}
	}
}

func (c *kustoCompleter) addMatchingStringSymbolCompletions(position int, builder *kustoCompletionBuilder) {
	if c == nil || c.code == nil {
		return
	}
	for _, symbol := range c.code.GetSymbolsInScope(position, SymbolMatchLocal, IncludeFunctionKindAll) {
		switch typed := symbol.(type) {
		case *VariableSymbol:
			if scalarSymbolFromSymbol(typed.Type) != ScalarTypes.String {
				continue
			}
			if item, ok := c.querySymbolCompletionItem(typed, CompletionKindVariable, false); ok {
				builder.add(item)
			}
		case *ParameterSymbol:
			if scalarSymbolFromSymbol(typed.Type) != ScalarTypes.String {
				continue
			}
			if item, ok := c.querySymbolCompletionItem(typed, CompletionKindParameter, false); ok {
				builder.add(item)
			}
		}
	}
}

func (c *kustoCompleter) resolveCompletionFunction(call *FunctionCallExpression) *FunctionSymbol {
	if c == nil || c.code == nil || call == nil || call.Name() == nil {
		return nil
	}
	name := NameReferenceSimpleName(call.Name())
	if name == "" {
		return nil
	}
	for _, symbol := range c.code.GetSymbolsInScope(call.Name().TextStart(), SymbolMatchLocal|SymbolMatchFunction|SymbolMatchView, IncludeFunctionKindAll) {
		if function, ok := symbol.(*FunctionSymbol); ok && function.SymbolName() == name {
			return function
		}
	}
	if c.code.Globals != nil {
		return c.code.Globals.GetFunction(name)
	}
	return nil
}

func functionCallCompletionArguments(call *FunctionCallExpression) []Expression {
	if call == nil || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
		return nil
	}
	elements := call.ArgumentList().Expressions().Elements()
	arguments := make([]Expression, 0, len(elements))
	for _, element := range elements {
		if element != nil {
			arguments = append(arguments, element.Element())
		}
	}
	return arguments
}

func functionCallCompletionArgumentIndex(call *FunctionCallExpression, position int) int {
	if call == nil || call.ArgumentList() == nil || call.ArgumentList().Expressions() == nil {
		return 0
	}
	index := 0
	for i, element := range call.ArgumentList().Expressions().Elements() {
		if element == nil || element.SeparatedElement == nil {
			continue
		}
		index = i
		separator := element.Separator()
		if separator != nil && separator.Width() > 0 && position >= separator.End() {
			index = i + 1
			continue
		}
		break
	}
	return index
}

func functionArgumentNameSuggestions(signatures []*Signature, arguments []Expression, argumentIndex int, globals *GlobalState) []string {
	specified := functionSpecifiedArgumentNames(arguments)
	suggestions := map[string]struct{}{}
	for _, signature := range signatures {
		if signature == nil || !signature.AllowsNamedArguments(globals) || len(arguments) > signature.MaxArgumentCount || !signatureContainsParameterNames(signature, specified) {
			continue
		}
		parameters := signature.ArgumentParametersForExpressions(arguments, functionArgumentSymbols(arguments), globals)
		usedPositionally := map[string]struct{}{}
		for i, parameter := range parameters {
			if i == argumentIndex || parameter == nil || parameter == UnknownParameter || parameter.Name == "" {
				continue
			}
			usedPositionally[parameter.Name] = struct{}{}
		}
		for _, parameter := range signature.Parameters {
			if parameter == nil || parameter.Name == "" {
				continue
			}
			if _, exists := specified[parameter.Name]; exists {
				continue
			}
			if _, exists := usedPositionally[parameter.Name]; exists {
				continue
			}
			suggestions[parameter.Name] = struct{}{}
		}
	}
	names := make([]string, 0, len(suggestions))
	for name := range suggestions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func functionSpecifiedArgumentNames(arguments []Expression) map[string]struct{} {
	names := map[string]struct{}{}
	for _, argument := range arguments {
		if named, ok := argument.(*SimpleNamedExpression); ok {
			if name := NameDeclarationSimpleName(named.Name()); name != "" {
				names[name] = struct{}{}
			}
		}
	}
	return names
}

func signatureContainsParameterNames(signature *Signature, names map[string]struct{}) bool {
	for name := range names {
		if signature.GetParameter(name) == nil {
			return false
		}
	}
	return true
}

func functionArgumentNamesRequired(signatures []*Signature, arguments []Expression, argumentIndex int, globals *GlobalState) bool {
	if argumentIndex <= 0 {
		return false
	}
	for _, signature := range signatures {
		if signature == nil || !signature.AllowsNamedArguments(globals) || argumentIndex >= signature.MaxArgumentCount {
			continue
		}
		parameters := signature.ArgumentParametersForExpressions(arguments, functionArgumentSymbols(arguments), globals)
		for i := 0; i < argumentIndex && i < len(parameters); i++ {
			named, isNamed := arguments[i].(*SimpleNamedExpression)
			if !isNamed || named == nil {
				continue
			}
			parameter := parameters[i]
			if parameter == nil || i >= len(signature.Parameters) || parameter != signature.Parameters[i] {
				return true
			}
		}
	}
	return false
}

type functionPossibleParameter struct {
	signature *Signature
	parameter *Parameter
}

func functionPossibleParameters(context functionArgumentCompletionContext, globals *GlobalState) []functionPossibleParameter {
	possible := []functionPossibleParameter{}
	seen := map[*Parameter]struct{}{}
	for _, signature := range context.signatures {
		if signature == nil {
			continue
		}
		parameters := []*Parameter{}
		switch {
		case context.parameterName != "":
			if parameter := signature.GetParameter(context.parameterName); parameter != nil {
				parameters = append(parameters, parameter)
			}
		case context.argumentIndex < len(context.arguments):
			argument := context.arguments[context.argumentIndex]
			if context.argumentIndex == len(context.arguments)-1 && SyntaxIsMissing(argument) {
				prior := context.arguments[:context.argumentIndex]
				parameters = signature.NextPossibleParametersForExpressions(prior, functionArgumentSymbols(prior))
			} else {
				mapped := signature.ArgumentParametersForExpressions(context.arguments, functionArgumentSymbols(context.arguments), globals)
				if context.argumentIndex < len(mapped) && mapped[context.argumentIndex] != nil && mapped[context.argumentIndex] != UnknownParameter {
					parameters = append(parameters, mapped[context.argumentIndex])
				}
			}
		default:
			parameters = signature.NextPossibleParametersForExpressions(context.arguments, functionArgumentSymbols(context.arguments))
		}
		for _, parameter := range parameters {
			if parameter == nil || parameter == UnknownParameter {
				continue
			}
			if _, exists := seen[parameter]; exists {
				continue
			}
			seen[parameter] = struct{}{}
			possible = append(possible, functionPossibleParameter{signature: signature, parameter: parameter})
		}
	}
	return possible
}

func functionArgumentSymbols(arguments []Expression) []Symbol {
	types := make([]Symbol, len(arguments))
	for i, argument := range arguments {
		if typ := ExpressionResultType(argument); typ != nil {
			types[i] = typ
		} else {
			types[i] = ScalarTypes.Unknown.Type()
		}
	}
	return types
}

func (c *kustoCompleter) addFunctionParameterExamples(signature *Signature, parameter *Parameter, arguments []Expression, position int, examples map[string]struct{}) {
	if parameter == nil {
		return
	}
	if len(parameter.Examples) > 0 {
		for _, example := range parameter.Examples {
			examples[example] = struct{}{}
		}
	} else {
		for _, value := range parameter.Values {
			examples[functionParameterValueExample(parameter, value)] = struct{}{}
		}
	}
	if signature != nil {
		if _, ok := signature.Symbol.(*FunctionSymbol); ok && parameter.TypeKind == ParameterTypeDeclared {
			for _, typ := range parameter.DeclaredTypes {
				if sameType(typ, ScalarTypes.Bool.Type()) {
					examples["false"] = struct{}{}
					examples["true"] = struct{}{}
				}
			}
		}
		if operator, ok := signature.Symbol.(*OperatorSymbol); ok {
			c.addOperatorArgumentExamples(operator, arguments, position, examples)
		}
	}
}

func (c *kustoCompleter) addOperatorArgumentExamples(operator *OperatorSymbol, arguments []Expression, position int, examples map[string]struct{}) {
	if operator == nil || len(arguments) == 0 {
		return
	}
	switch operator {
	case Operators.Equal,
		Operators.NotEqual,
		Operators.EqualTilde,
		Operators.GreaterThan,
		Operators.GreaterThanOrEqual,
		Operators.LessThan,
		Operators.LessThanOrEqual:
		if sameType(c.completionExpressionResultType(arguments[0], position), ScalarTypes.Bool.Type()) {
			examples["false"] = struct{}{}
			examples["true"] = struct{}{}
		}
	}
}

func (c *kustoCompleter) completionExpressionResultType(expression Expression, position int) *TypeSymbol {
	if typ := ExpressionResultType(expression); typ != nil {
		return typ
	}
	switch typed := expression.(type) {
	case *LiteralExpression:
		if typed.Kind() == SyntaxKindBooleanLiteralExpression {
			return ScalarTypes.Bool.Type()
		}
	case *SimpleNamedExpression:
		return c.completionExpressionResultType(typed.Expression(), position)
	case *ParenthesizedExpression:
		return c.completionExpressionResultType(typed.Expression(), position)
	case *BracketedExpression:
		return c.completionExpressionResultType(typed.Expression(), position)
	case *NameReference:
		if c != nil && c.code != nil {
			name := NameReferenceSimpleName(typed)
			if columns := c.code.GetColumnsInScope(position); columns != nil {
				if column := columns.GetColumn(name); column != nil {
					return column.Type
				}
			}
			if rowScope := getQueryCompletionScope(c.code, position).rowScope; rowScope != nil {
				if column := rowScope.GetColumn(name); column != nil {
					return column.Type
				}
			}
		}
	}
	if symbol := expressionReferencedSymbol(expression); symbol != nil {
		return returnTypeFromSymbol(symbol)
	}
	return nil
}

func functionParameterValueExample(parameter *Parameter, value any) string {
	switch typed := value.(type) {
	case string:
		if parameter == nil || !parameter.IsCaseSensitive {
			typed = invariantLower(typed)
		}
		return GetSingleQuotedStringLiteral(typed)
	default:
		return invariantCompletionValue(value)
	}
}
