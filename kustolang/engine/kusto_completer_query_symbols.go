package kustolang

const queryCompletionAfterQueryStart = "\n| "

func (c *kustoCompleter) getQuerySymbolCompletions(position int, builder *kustoCompletionBuilder) completionMode {
	if c == nil || c.code == nil || builder == nil || !c.options.IncludeSymbols() {
		return completionModeCombined
	}
	if c.addSetOptionNameCompletions(position, builder) {
		return completionModeIsolated
	}
	if mode := c.getFunctionArgumentCompletions(position, builder); mode == completionModeIsolated {
		return mode
	}
	if mode := c.getSetOptionCompletions(position, builder); mode == completionModeIsolated {
		return mode
	}
	c.getColumnExampleCompletions(position, builder)
	c.getDeconstructCompletions(position, builder)
	hint := c.getCompletionHint(position)
	match := c.getSymbolMatch(position)
	if path := c.code.getPathScopeContext(position); path.bounded {
		seen := map[string]struct{}{}
		for _, member := range c.code.GetSymbolsInScope(position, match, c.options.IncludeFunctions) {
			if kind, ok := c.queryCompletionKindForSymbol(member); ok {
				c.addQuerySymbolCompletion(builder, member, kind, queryCompletionScope{hint: hint}, seen)
			}
		}
		return completionModeCombined
	}
	scope := getQueryCompletionScope(c.code, position)
	scope.match = match
	scope.hint = hint
	if match != SymbolMatchNone {
		c.addMatchedQuerySymbols(position, builder, scope)
		return completionModeCombined
	}
	globals := scope.globals
	if globals == nil {
		globals = c.code.Globals
	}
	seen := map[string]struct{}{}
	c.addVisibleLocalQuerySymbols(position, builder, scope, seen)
	if scope.root || scope.tableSource || querySymbolMatchRequestsDatabaseEntities(scope.match) ||
		(scope.bounded && scope.catalogScope) {
		c.addDatabaseEntityQuerySymbols(builder, globals, scope, seen)
	}
	if !scope.root && scope.rowScope == nil {
		if scope.bounded && scope.scopeKind == ScopeKindNormal && (scope.inputRowScope != nil || scope.catalogScope) {
			c.addApplicableQueryFunctions(builder, scope, seen)
		}
		return completionModeCombined
	}
	if scope.scopeKind == ScopeKindNormal && scope.rowScope != nil {
		columns := scope.rowScope.Columns()
		if scope.rightRowScope != nil {
			left := NewVariableSymbol("$left", NewTupleSymbol(scope.rowScope.Columns(), scope.rowScope), false, nil, nil)
			right := NewVariableSymbol("$right", NewTupleSymbol(scope.rightRowScope.Columns(), scope.rightRowScope), false, nil, nil)
			c.addQuerySymbolCompletion(builder, left, CompletionKindVariable, scope, seen)
			c.addQuerySymbolCompletion(builder, right, CompletionKindVariable, scope, seen)
			columns = kustoCodeCommonColumns(scope.rowScope, scope.rightRowScope)
		}
		for _, column := range columns {
			c.addQuerySymbolCompletion(builder, column, CompletionKindColumn, scope, seen)
		}
	}

	c.addApplicableQueryFunctions(builder, scope, seen)
	return completionModeCombined
}

func (c *kustoCompleter) addMatchedQuerySymbols(position int, builder *kustoCompletionBuilder, scope queryCompletionScope) {
	seen := map[string]struct{}{}
	for _, symbol := range c.code.GetSymbolsInScope(position, scope.match, c.options.IncludeFunctions) {
		if scope.invoke && scope.inputRowScope != nil && !scope.inputRowScope.IsOpen() {
			switch typed := symbol.(type) {
			case *FunctionSymbol:
				if !queryInvokeFunctionApplies(typed, scope.inputRowScope) {
					continue
				}
			case *VariableSymbol:
				if function, ok := typed.Type.(*FunctionSymbol); ok && !queryInvokeFunctionApplies(function, scope.inputRowScope) {
					continue
				}
			}
		}
		kind, ok := c.queryCompletionKindForSymbol(symbol)
		if !ok {
			continue
		}
		c.addQuerySymbolCompletion(builder, symbol, kind, scope, seen)
	}
	c.addContextRecoveryQuerySymbols(position, builder, scope, seen)
}

func (c *kustoCompleter) addContextRecoveryQuerySymbols(position int, builder *kustoCompletionBuilder, scope queryCompletionScope, seen map[string]struct{}) {
	for _, symbol := range c.code.localSymbolsInScope(position) {
		switch typed := symbol.(type) {
		case *EntityGroupElementSymbol:
			c.addQuerySymbolCompletion(builder, typed, CompletionKindVariable, scope, seen)
		case *VariableSymbol, *ParameterSymbol:
			if scope.root {
				kind, ok := c.queryCompletionKindForSymbol(symbol)
				if ok {
					c.addQuerySymbolCompletion(builder, symbol, kind, scope, seen)
				}
			}
		case *FunctionSymbol:
			if scope.invoke && queryInvokeFunctionApplies(typed, scope.inputRowScope) {
				c.addQuerySymbolCompletion(builder, typed, CompletionKindLocalFunction, scope, seen)
			}
		}
	}
	if scope.invoke {
		c.addApplicableQueryFunctions(builder, scope, seen)
	}
	if c.isNestedTabularExpressionContext(position) {
		globals := scope.globals
		if globals == nil {
			globals = c.code.Globals
		}
		c.addDatabaseEntityQuerySymbols(builder, globals, scope, seen)
	}
}

func (c *kustoCompleter) isNestedTabularExpressionContext(position int) bool {
	context, ok := c.tryGetCompletionContext(position)
	for current := context.node; ok && current != nil; current = current.Parent() {
		switch WrapGeneratedSyntaxElement(current).(type) {
		case *ToScalarExpression, *ToTableExpression:
			return true
		}
	}
	return false
}

func (c *kustoCompleter) queryCompletionKindForSymbol(symbol Symbol) (CompletionKind, bool) {
	if symbol == nil {
		return CompletionKindUnknown, false
	}
	switch typed := symbol.(type) {
	case *ClusterSymbol:
		return CompletionKindCluster, true
	case *ColumnSymbol:
		return CompletionKindColumn, true
	case *DatabaseSymbol:
		return CompletionKindDatabase, true
	case *FunctionSymbol:
		globals := c.code.Globals
		switch {
		case globals != nil && globals.IsAggregateFunction(typed):
			return CompletionKindAggregateFunction, true
		case globals != nil && globals.IsBuiltInFunction(typed):
			return CompletionKindBuiltInFunction, true
		case globals != nil && globals.IsDatabaseFunction(typed):
			return CompletionKindDatabaseFunction, true
		default:
			return CompletionKindLocalFunction, true
		}
	case *PatternSymbol:
		return CompletionKindLocalFunction, true
	case *VariableSymbol:
		if typed.Type != nil {
			switch typed.Type.SymbolKind() {
			case SymbolKindTable:
				return CompletionKindTable, true
			case SymbolKindFunction, SymbolKindPattern, SymbolKindGraph:
				return c.queryCompletionKindForSymbol(typed.Type)
			}
		}
		return CompletionKindVariable, true
	case *ParameterSymbol:
		return CompletionKindParameter, true
	case *TableSymbol, *ExternalTableSymbol:
		return CompletionKindTable, true
	case *MaterializedViewSymbol:
		return CompletionKindMaterialiedView, true
	case *EntityGroupSymbol:
		return CompletionKindEntityGroup, true
	case *EntityGroupElementSymbol:
		return CompletionKindVariable, true
	case *GraphModelSymbol, *GraphSnapshotSymbol, *GraphSymbol:
		return CompletionKindGraph, true
	case *StoredQueryResultSymbol:
		return CompletionKindStoredQueryResult, true
	default:
		return CompletionKindUnknown, false
	}
}

func querySymbolMatchRequestsDatabaseEntities(match SymbolMatch) bool {
	return match&(SymbolMatchTable|SymbolMatchExternalTable|SymbolMatchMaterializedView|
		SymbolMatchStoredQueryResult|SymbolMatchEntityGroup|SymbolMatchGraph|SymbolMatchGraphModel|
		SymbolMatchGraphSnapshot) != 0
}

func (c *kustoCompleter) addDatabaseEntityQuerySymbols(builder *kustoCompletionBuilder, globals *GlobalState, scope queryCompletionScope, seen map[string]struct{}) {
	if globals == nil || globals.Database == nil {
		return
	}
	if scope.match == SymbolMatchNone {
		for _, table := range globals.Database.Tables() {
			c.addQuerySymbolCompletion(builder, table, CompletionKindTable, scope, seen)
		}
		return
	}
	for _, member := range globals.Database.Members() {
		if _, isFunction := member.(*FunctionSymbol); isFunction {
			continue
		}
		if kind, ok := queryPathCompletionKind(member); ok {
			c.addQuerySymbolCompletion(builder, member, kind, scope, seen)
		}
	}
}

func queryPathCompletionKind(symbol Symbol) (CompletionKind, bool) {
	switch symbol.(type) {
	case *ColumnSymbol:
		return CompletionKindColumn, true
	case *TableSymbol, *ExternalTableSymbol:
		return CompletionKindTable, true
	case *MaterializedViewSymbol:
		return CompletionKindMaterialiedView, true
	case *StoredQueryResultSymbol:
		return CompletionKindStoredQueryResult, true
	case *FunctionSymbol:
		return CompletionKindDatabaseFunction, true
	case *DatabaseSymbol:
		return CompletionKindDatabase, true
	case *ClusterSymbol:
		return CompletionKindCluster, true
	case *EntityGroupSymbol:
		return CompletionKindEntityGroup, true
	case *GraphModelSymbol, *GraphSnapshotSymbol, *GraphSymbol:
		return CompletionKindGraph, true
	default:
		return CompletionKindUnknown, false
	}
}

func (c *kustoCompleter) addVisibleLocalQuerySymbols(position int, builder *kustoCompletionBuilder, scope queryCompletionScope, seen map[string]struct{}) {
	if c == nil || c.code == nil {
		return
	}
	include := c.options.IncludeFunctions & (IncludeFunctionKindLocalFunctions | IncludeFunctionKindLocalViews)
	for _, visible := range c.code.GetSymbolsInScope(position, SymbolMatchLocal|SymbolMatchFunction|SymbolMatchView|SymbolMatchEntityGroupElement, include) {
		switch symbol := visible.(type) {
		case *VariableSymbol:
			if function, ok := symbol.Type.(*FunctionSymbol); ok && scope.invoke && !queryInvokeFunctionApplies(function, scope.inputRowScope) {
				continue
			}
			if kind, ok := c.queryCompletionKindForSymbol(symbol); ok {
				c.addQuerySymbolCompletion(builder, symbol, kind, scope, seen)
			}
		case *ParameterSymbol:
			c.addQuerySymbolCompletion(builder, symbol, CompletionKindParameter, scope, seen)
		case *FunctionSymbol:
			if scope.invoke && !queryInvokeFunctionApplies(symbol, scope.inputRowScope) {
				continue
			}
			c.addQuerySymbolCompletion(builder, symbol, CompletionKindLocalFunction, scope, seen)
		case *EntityGroupElementSymbol:
			c.addQuerySymbolCompletion(builder, symbol, CompletionKindVariable, scope, seen)
		}
	}
	for _, visible := range c.code.localSymbolsInScope(position) {
		if column, ok := visible.(*ColumnSymbol); ok {
			c.addQuerySymbolCompletion(builder, column, CompletionKindColumn, scope, seen)
		}
	}
}

func (c *kustoCompleter) addApplicableQueryFunctions(builder *kustoCompletionBuilder, scope queryCompletionScope, seen map[string]struct{}) {
	globals := scope.globals
	if globals == nil {
		globals = c.code.Globals
	}
	include := c.options.IncludeFunctions
	if scope.scopeKind == ScopeKindPlugIn {
		if include&IncludeFunctionKindBuiltInFunctions != 0 {
			for _, function := range globals.PlugIns {
				c.addQuerySymbolCompletion(builder, function, CompletionKindBuiltInFunction, scope, seen)
			}
		}
		return
	}
	if scope.scopeKind == ScopeKindOption {
		return
	}
	if scope.scopeKind == ScopeKindAggregate && include&IncludeFunctionKindBuiltInFunctions != 0 {
		for _, function := range globals.Aggregates {
			if queryFunctionApplies(function, scope) {
				if kind, ok := c.queryCompletionKindForSymbol(function); ok {
					c.addQuerySymbolCompletion(builder, function, kind, scope, seen)
				}
			}
		}
	}
	if include&IncludeFunctionKindDatabaseFunctions != 0 && globals.Database != nil {
		for _, function := range globals.Database.Functions() {
			if queryFunctionApplies(function, scope) {
				c.addQuerySymbolCompletion(builder, function, CompletionKindDatabaseFunction, scope, seen)
			}
		}
	}
	if include&IncludeFunctionKindBuiltInFunctions != 0 {
		for _, function := range globals.Functions {
			if queryFunctionApplies(function, scope) {
				c.addQuerySymbolCompletion(builder, function, CompletionKindBuiltInFunction, scope, seen)
			}
		}
	}
}

func queryFunctionApplies(function *FunctionSymbol, scope queryCompletionScope) bool {
	if function == nil {
		return false
	}
	if scope.root {
		return function.SymbolTabularity() == TabularityTabular
	}
	if scope.invoke {
		return queryInvokeFunctionApplies(function, scope.inputRowScope)
	}
	if scope.tableSource {
		return function.SymbolTabularity() == TabularityTabular || function.SymbolTabularity() == TabularityUnknown
	}
	return function.SymbolTabularity() == TabularityScalar || function.SymbolTabularity() == TabularityUnknown
}

func queryInvokeFunctionApplies(function *FunctionSymbol, rowScope *TableSymbol) bool {
	if function == nil {
		return false
	}
	if function == Functions.Cluster || function == Functions.Database {
		return true
	}
	if rowScope == nil {
		return false
	}
	if rowScope.IsOpen() {
		return true
	}
	for _, signature := range function.Signatures() {
		if signature == nil || len(signature.Parameters) == 0 || signature.Parameters[0] == nil {
			continue
		}
		parameter := signature.Parameters[0]
		switch parameter.TypeKind {
		case ParameterTypeDeclared:
			for _, declared := range parameter.DeclaredSymbols {
				if table, ok := tableSymbolFromSymbol(declared); ok && SymbolAssignableTo(rowScope, table) {
					return true
				}
			}
			for _, declared := range parameter.DeclaredTypes {
				if table, ok := tableSymbolFromSymbol(declared); ok && SymbolAssignableTo(rowScope, table) {
					return true
				}
			}
		case ParameterTypeTabular:
			return true
		}
	}
	return false
}

func (c *kustoCompleter) addQuerySymbolCompletion(builder *kustoCompletionBuilder, symbol Symbol, kind CompletionKind, scope queryCompletionScope, seen map[string]struct{}) {
	if symbol == nil || symbol.SymbolName() == "" || symbol.IsHidden() {
		return
	}
	key := queryCompletionSymbolKey(symbol)
	if _, exists := seen[key]; exists {
		return
	}
	seen[key] = struct{}{}
	c.cancellationToken.ThrowIfCancellationRequested()

	item, ok := c.querySymbolCompletionItemInContext(symbol, kind, scope.root, scope.invoke)
	if ok {
		shouldAddSpace := querySymbolCompletionShouldAddSpace(symbol, scope.hint)
		if shouldAddSpace {
			item = addSpaceToCompletionItem(item)
		}
		builder.add(item)
		if database, isDatabase := symbol.(*DatabaseSymbol); isDatabase && database.AlternateName() != "" {
			alternate := database.AlternateName()
			alternateItem := NewCompletionItem(CompletionKindDatabase, alternate,
				WithCompletionBeforeText(BracketNameIfNecessary(alternate)))
			if shouldAddSpace {
				alternateItem = addSpaceToCompletionItem(alternateItem)
			}
			builder.add(alternateItem)
		}
	}
}

func querySymbolCompletionShouldAddSpace(symbol Symbol, hint CompletionHint) bool {
	return hint&CompletionHintBoolean != 0 && GetResultType(symbol) != ScalarTypes.Bool.Type()
}

func queryCompletionSymbolKey(symbol Symbol) string {
	prefix := string(symbol.SymbolKind())
	if _, ok := symbol.(*FunctionSymbol); ok {
		prefix = "Function"
	}
	return prefix + ":" + symbol.SymbolName()
}

func (c *kustoCompleter) querySymbolCompletionItem(symbol Symbol, kind CompletionKind, atQueryStart bool) (CompletionItem, bool) {
	return c.querySymbolCompletionItemInContext(symbol, kind, atQueryStart, false)
}

func (c *kustoCompleter) querySymbolCompletionItemInContext(symbol Symbol, kind CompletionKind, atQueryStart bool, invoke bool) (CompletionItem, bool) {
	name := symbol.SymbolName()
	if table, ok := tableSymbolFromSymbol(symbol); ok {
		insertionText := BracketNameIfNecessary(name)
		priority := CompletionPriorityNormal
		if table.IsExternal() {
			insertionText = "external_table('" + name + "')"
			priority = CompletionPriorityLow
		}
		if atQueryStart {
			insertionText += queryCompletionAfterQueryStart
		}
		return NewCompletionItem(kind, CompletionDisplayText(symbol),
			WithCompletionBeforeText(insertionText),
			WithCompletionMatchText(name),
			WithCompletionPriority(priority)), true
	}
	switch typed := symbol.(type) {
	case *EntityGroupSymbol:
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText("entity_group("+GetSingleQuotedStringLiteral(name)+")")), true
	case *GraphModelSymbol, *GraphSnapshotSymbol, *GraphSymbol:
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText("graph("+GetSingleQuotedStringLiteral(name)+")")), true
	case *DatabaseSymbol, *ClusterSymbol:
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText(BracketNameIfNecessary(name))), true
	case *ColumnSymbol:
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText(BracketNameIfNecessary(name))), true
	case *VariableSymbol:
		if function, ok := typed.Type.(*FunctionSymbol); ok {
			return c.queryFunctionCompletionItem(function, kind, atQueryStart, invoke)
		}
		beforeText := BracketNameIfNecessary(name)
		if atQueryStart && typed.Type != nil && typed.Type.SymbolTabularity() == TabularityTabular {
			beforeText += queryCompletionAfterQueryStart
		}
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText(beforeText)), true
	case *ParameterSymbol:
		if function, ok := typed.Type.(*FunctionSymbol); ok {
			return c.queryFunctionCompletionItem(function, kind, atQueryStart, invoke)
		}
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText(BracketNameIfNecessary(name))), true
	case *EntityGroupElementSymbol:
		beforeText := BracketNameIfNecessary(name)
		if atQueryStart && typed.SymbolTabularity() == TabularityTabular {
			beforeText += queryCompletionAfterQueryStart
		}
		return NewCompletionItem(kind, CompletionDisplayText(symbol), WithCompletionBeforeText(beforeText)), true
	case *FunctionSymbol:
		return c.queryFunctionCompletionItem(typed, kind, atQueryStart, invoke)
	case *PatternSymbol:
		return NewCompletionItem(kind, CompletionDisplayText(symbol),
			WithCompletionBeforeText(BracketNameIfNecessary(name)+"("),
			WithCompletionAfterText(")"),
			WithCompletionMatchText(name)), true
	default:
		return CompletionItem{}, false
	}
}

func (c *kustoCompleter) queryFunctionCompletionItem(function *FunctionSymbol, kind CompletionKind, atQueryStart bool, invoke bool) (CompletionItem, bool) {
	if function == nil {
		return CompletionItem{}, false
	}
	name := function.SymbolName()
	builtIn := c != nil && c.code != nil && c.code.Globals != nil && c.code.Globals.IsBuiltInFunction(function)
	editName := BracketNameIfNecessary(name)
	if builtIn {
		editName = name
	}
	displayText := CompletionDisplayText(function)
	if queryFunctionHasNoVisibleParameters(function) {
		if builtIn {
			editName += "()"
		}
		if atQueryStart {
			editName += queryCompletionAfterQueryStart
		}
		return NewCompletionItem(kind, displayText, WithCompletionBeforeText(editName), WithCompletionMatchText(name)), true
	}
	if c != nil && c.options.EnableParameterInjection() && function.MaxArgumentCount() == 1 && !builtIn && !invoke {
		return NewCompletionItem(kind, displayText, WithCompletionBeforeText(editName+"({parameter})"), WithCompletionMatchText(name)), true
	}
	return NewCompletionItem(kind, displayText,
		WithCompletionBeforeText(editName+"("),
		WithCompletionAfterText(")"),
		WithCompletionMatchText(name)), true
}

func queryFunctionHasNoVisibleParameters(function *FunctionSymbol) bool {
	found := false
	for _, signature := range function.Signatures() {
		if signature == nil || signature.IsHidden || signature.IsObsolete() {
			continue
		}
		found = true
		if len(signature.Parameters) > 0 {
			return false
		}
	}
	return found
}
