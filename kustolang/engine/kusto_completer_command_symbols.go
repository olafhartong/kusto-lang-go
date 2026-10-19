package kustolang

// Ported from the symbol-completion phase in
// src/Kusto.Language/Editor/Kusto/KustoCompleter.cs, scoped to command parser
// rule hints and catalog-backed table-like database members currently present
// in the Go port. See third_party/kusto/LICENSE for the preserved Apache-2.0
// license.

func (c *kustoCompleter) getSymbolCompletions(position int, builder *kustoCompletionBuilder) completionMode {
	if c == nil || c.code == nil || builder == nil || !c.options.IncludeSymbols() {
		return completionModeCombined
	}
	if kustoCodeIsQueryCompletionPosition(c.code, position) {
		return c.getQuerySymbolCompletions(position, builder)
	}
	if c.code.Kind == CodeKindCommand {
		c.getCommandSymbolCompletions(position, builder)
	}
	return completionModeCombined
}

func (c *kustoCompleter) getCommandSymbolCompletions(position int, builder *kustoCompletionBuilder) {
	context, ok := c.commandCompletionContext(position)
	if !ok {
		return
	}
	grammar := CommandGrammarFrom(c.code.Globals)
	if grammar == nil {
		return
	}
	parser := grammar.CommandCompletionParser()
	if parser == nil {
		return
	}
	parsers := FindAnnotatedParsers(NewArraySource(c.code.Tokens), context.findOffset, parser, context.startIndex)
	for _, parser := range parsers {
		hints := syntaxCompletionHintFromAnnotations(parser.Annotations())
		for _, symbol := range c.commandSymbolsForCompletionHint(position, hints) {
			for _, item := range commandSymbolCompletionItems(symbol) {
				if commandCompletionItemMatchesPartial(item, context.partialText) {
					builder.add(item)
				}
			}
		}
	}
}

func (c *kustoCompleter) commandSymbolsForCompletionHint(position int, hints CompletionHint) []Symbol {
	if c == nil || c.code == nil || c.code.Globals == nil {
		return nil
	}
	globals := c.code.globalsInScope(position)
	symbols := []Symbol{}
	if hints&CompletionHintColumn != 0 {
		if rowScope := c.code.GetColumnsInScope(position); rowScope != nil {
			for _, column := range rowScope.Columns() {
				symbols = append(symbols, column)
			}
		}
	}
	if database := globals.Database; database != nil {
		if hints&CompletionHintTable != 0 {
			for _, table := range database.Tables() {
				symbols = append(symbols, table)
			}
		}
		if hints&CompletionHintExternalTable != 0 {
			for _, table := range database.ExternalTables() {
				symbols = append(symbols, table)
			}
		}
		if hints&CompletionHintMaterializedView != 0 {
			for _, view := range database.MaterializedViews() {
				symbols = append(symbols, view)
			}
		}
		if hints&CompletionHintEntityGroup != 0 {
			for _, group := range database.EntityGroups() {
				symbols = append(symbols, group)
			}
		}
		if hints&CompletionHintStoredQueryResult != 0 {
			for _, result := range database.StoredQueryResults() {
				symbols = append(symbols, result)
			}
		}
		if hints&CompletionHintGraphModel != 0 {
			for _, model := range database.GraphModels() {
				symbols = append(symbols, model)
			}
		}
		if hints&CompletionHintGraphSnapshot != 0 {
			for _, model := range database.GraphModels() {
				for _, snapshot := range model.Snapshots {
					symbols = append(symbols, snapshot)
				}
			}
		}
		if hints&(CompletionHintFunction|CompletionHintDatabaseFunction|CompletionHintTabularFunction|CompletionHintScalarFunction) != 0 {
			for _, function := range database.Functions() {
				symbols = append(symbols, function)
			}
		}
	}
	if hints&CompletionHintDatabase != 0 && globals.Cluster != nil {
		for _, database := range globals.Cluster.Databases() {
			symbols = append(symbols, database)
		}
	}
	if hints&CompletionHintCluster != 0 {
		for _, cluster := range globals.Clusters {
			symbols = append(symbols, cluster)
		}
	}
	if hints&CompletionHintOption != 0 {
		for _, option := range globals.Options {
			symbols = append(symbols, option)
		}
	}
	return symbols
}

func commandSymbolCompletionItems(symbol Symbol) []CompletionItem {
	if symbol == nil || symbol.SymbolName() == "" || symbol.IsHidden() {
		return nil
	}
	name := symbol.SymbolName()
	display := CompletionDisplayText(symbol, true)
	before := BracketNameIfNecessary(name)
	kind := CompletionKindUnknown
	switch typed := symbol.(type) {
	case *ColumnSymbol:
		kind = CompletionKindColumn
	case *TableSymbol, *ExternalTableSymbol:
		kind = CompletionKindTable
	case *MaterializedViewSymbol:
		kind = CompletionKindMaterialiedView
	case *DatabaseSymbol:
		items := []CompletionItem{NewCompletionItem(CompletionKindDatabase, display,
			WithCompletionBeforeText(before), WithCompletionMatchText(name))}
		if alternate := typed.AlternateName(); alternate != "" {
			items = append(items, NewCompletionItem(CompletionKindDatabase, alternate,
				WithCompletionBeforeText(BracketNameIfNecessary(alternate)), WithCompletionMatchText(alternate)))
		}
		return items
	case *ClusterSymbol:
		kind = CompletionKindCluster
		before = GetBracketedName(name)
	case *FunctionSymbol:
		kind = CompletionKindDatabaseFunction
	case *EntityGroupSymbol:
		kind = CompletionKindEntityGroup
	case *StoredQueryResultSymbol:
		kind = CompletionKindStoredQueryResult
	case *GraphModelSymbol, *GraphSnapshotSymbol, *GraphSymbol:
		kind = CompletionKindGraph
	case *OptionSymbol:
		kind = CompletionKindOption
	default:
		return nil
	}
	return []CompletionItem{NewCompletionItem(kind, display,
		WithCompletionBeforeText(before), WithCompletionMatchText(name))}
}
