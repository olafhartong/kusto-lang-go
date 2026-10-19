package kustolang

// Ported from the public purpose of src/Kusto.Language/Binder/Binder_API.cs:
// resolve names against GlobalState and produce SemanticInfo. This is a small
// facade over database symbols, not the full syntax-tree binder.

type Binder struct {
	Globals                  *GlobalState
	LocalScope               *LocalScope
	initialGlobals           *GlobalState
	currentCluster           *ClusterSymbol
	currentDatabase          *DatabaseSymbol
	currentFunction          *FunctionSymbol
	globalBindingCache       *GlobalBindingCache
	localBindingCache        *LocalBindingCache
	tupleMap                 map[*TableSymbol]*TupleSymbol
	openTableInferredColumns map[*TableSymbol]map[string]*ColumnSymbol
	openTableInferredOrder   map[*TableSymbol][]*ColumnSymbol
	syntaxNodeLocations      map[*SyntaxNode]TextRange
	syntaxTokenLocations     map[*SyntaxToken]TextRange
	rowScope                 *TableSymbol
	pathScope                Symbol
	scopeKind                ScopeKind
	aggregateArgumentScope   bool
	dynamicDepth             int
	implicitFunctionArgument *implicitFunctionArgument
	openClusters             map[string]*ClusterSymbol
	openDatabases            map[*ClusterSymbol]map[string]*DatabaseSymbol
	openTables               map[*DatabaseSymbol]map[string]*TableSymbol
	aliasedDatabases         map[string]*DatabaseSymbol
	graphModels              map[*GraphModelSymbol]*GraphSymbol
	entityGroupExpansions    map[*Signature]bool
	functionExpansions       map[*Signature]bool
	functionStaticScopes     map[*Signature]*LocalScope
	isFuzzy                  bool
	cancellationToken        CancellationToken
}

func (b *Binder) getOpenCluster(name string) *ClusterSymbol {
	if b.openClusters == nil {
		b.openClusters = map[string]*ClusterSymbol{}
	}
	if cluster := b.openClusters[name]; cluster != nil {
		return cluster
	}
	cluster := NewClusterSymbolWithMembers(name, nil, true)
	b.openClusters[name] = cluster
	return cluster
}

func (b *Binder) getOpenDatabase(name string, cluster *ClusterSymbol) *DatabaseSymbol {
	if cluster == nil && b.Globals != nil {
		cluster = b.Globals.Cluster
	}
	if b.openDatabases == nil {
		b.openDatabases = map[*ClusterSymbol]map[string]*DatabaseSymbol{}
	}
	databases := b.openDatabases[cluster]
	if databases == nil {
		databases = map[string]*DatabaseSymbol{}
		b.openDatabases[cluster] = databases
	}
	if database := databases[name]; database != nil {
		return database
	}
	database := NewDatabaseSymbolWithAlternateName(name, "", nil, true)
	databases[name] = database
	return database
}

func (b *Binder) getOpenTable(name string, database *DatabaseSymbol) *TableSymbol {
	if database == nil && b.Globals != nil {
		database = b.Globals.Database
	}
	if b.openTables == nil {
		b.openTables = map[*DatabaseSymbol]map[string]*TableSymbol{}
	}
	tables := b.openTables[database]
	if tables == nil {
		tables = map[string]*TableSymbol{}
		b.openTables[database] = tables
	}
	if table := tables[name]; table != nil {
		return table
	}
	table := NewTableSymbol(name).WithIsOpen(true)
	tables[name] = table
	return table
}

func (b *Binder) getOpenTableInferredColumn(name string, table *TableSymbol) *ColumnSymbol {
	if b == nil || table == nil || name == "" {
		return nil
	}
	if b.openTableInferredColumns == nil {
		b.openTableInferredColumns = map[*TableSymbol]map[string]*ColumnSymbol{}
		b.openTableInferredOrder = map[*TableSymbol][]*ColumnSymbol{}
	}
	columns := b.openTableInferredColumns[table]
	if columns == nil {
		columns = map[string]*ColumnSymbol{}
		b.openTableInferredColumns[table] = columns
	}
	if column := columns[name]; column != nil {
		return column
	}
	column := NewColumnSymbol(name, ScalarTypes.Unknown.Type())
	columns[name] = column
	b.openTableInferredOrder[table] = append(b.openTableInferredOrder[table], column)
	return column
}

func (b *Binder) declaredAndInferredColumns(table *TableSymbol) []*ColumnSymbol {
	if table == nil {
		return []*ColumnSymbol{}
	}
	columns := append([]*ColumnSymbol{}, table.Columns()...)
	if table.IsOpen() && b != nil {
		columns = append(columns, b.openTableInferredOrder[table]...)
	}
	return columns
}

func NewBinder(globals *GlobalState) *Binder {
	if globals == nil {
		globals = DefaultGlobalState
	}
	globals = globals.WithCache()
	globalBindingCache := KustoCacheGetOrCreate(globals.Cache, NewGlobalBindingCache)
	return &Binder{
		Globals:               globals,
		initialGlobals:        globals,
		currentCluster:        globals.Cluster,
		currentDatabase:       globals.Database,
		globalBindingCache:    globalBindingCache,
		localBindingCache:     NewLocalBindingCache(),
		LocalScope:            NewLocalScope(nil),
		entityGroupExpansions: map[*Signature]bool{},
		functionExpansions:    map[*Signature]bool{},
		functionStaticScopes:  map[*Signature]*LocalScope{},
	}
}

// CurrentCluster returns the cluster selected for subsequent binding.
func (b *Binder) CurrentCluster() *ClusterSymbol {
	if b == nil {
		return nil
	}
	return b.currentCluster
}

// CurrentDatabase returns the database selected for subsequent binding.
func (b *Binder) CurrentDatabase() *DatabaseSymbol {
	if b == nil {
		return nil
	}
	return b.currentDatabase
}

// CurrentFunction returns the function whose body is being bound.
func (b *Binder) CurrentFunction() *FunctionSymbol {
	if b == nil {
		return nil
	}
	return b.currentFunction
}

func (b *Binder) resetCurrentDatabaseContext() {
	if b == nil {
		return
	}
	if b.initialGlobals == nil {
		b.initialGlobals = DefaultGlobalState
	}
	b.Globals = b.initialGlobals
	b.currentCluster = b.initialGlobals.Cluster
	b.currentDatabase = b.initialGlobals.Database
}

func (b *Binder) updateCurrentDatabaseContext() {
	if b == nil || b.initialGlobals == nil {
		return
	}
	if b.currentCluster == b.initialGlobals.Cluster && b.currentDatabase == b.initialGlobals.Database {
		b.Globals = b.initialGlobals
		return
	}
	globals := *b.initialGlobals
	globals.Cluster = b.currentCluster
	globals.Database = b.currentDatabase
	b.Globals = globals.rebindCache()
}

// TryBind attaches semantic information for the syntax-tree surface currently
// supported by the native tree binder. Unsupported nodes remain unbound.
func (b *Binder) TryBind(root SyntaxElement) bool {
	return b.TryBindWithCancellation(root, CancellationToken{})
}

// TryBindWithCancellation binds the syntax tree and checks cancellation while
// capturing locations and visiting nodes.
func (b *Binder) TryBindWithCancellation(root SyntaxElement, cancellationToken CancellationToken) bool {
	if b == nil || root == nil {
		return false
	}
	b.cancellationToken = cancellationToken
	b.cancellationToken.ThrowIfCancellationRequested()
	b.resetCurrentDatabaseContext()
	if node := syntaxNodeForBinderLocation(root); node != nil {
		if node.Tree() == nil {
			NewSyntaxTree(node, nil, 0)
		}
	}
	b.captureSyntaxLocations(root)
	newTreeBinder(b).bind(root)
	b.cancellationToken.ThrowIfCancellationRequested()
	return true
}

func (b *Binder) captureSyntaxLocations(root SyntaxElement) {
	b.syntaxNodeLocations = map[*SyntaxNode]TextRange{}
	b.syntaxTokenLocations = map[*SyntaxToken]TextRange{}
	b.appendSyntaxLocations(root)
}

func (b *Binder) appendSyntaxLocations(root SyntaxElement) {
	if b.syntaxNodeLocations == nil {
		b.syntaxNodeLocations = map[*SyntaxNode]TextRange{}
	}
	if b.syntaxTokenLocations == nil {
		b.syntaxTokenLocations = map[*SyntaxToken]TextRange{}
	}
	var walk func(SyntaxElement, int)
	walk = func(element SyntaxElement, triviaStart int) {
		if element == nil {
			return
		}
		b.cancellationToken.ThrowIfCancellationRequested()
		start := triviaStart + element.ByteTriviaWidth()
		location := TextRangeFromBounds(start, triviaStart+element.ByteFullWidth())
		if token, ok := element.(*SyntaxToken); ok {
			b.syntaxTokenLocations[token] = location
		}
		if node := syntaxNodeForBinderLocation(element); node != nil {
			b.syntaxNodeLocations[node] = location
		}
		childStart := triviaStart
		for i := 0; i < element.ChildCount(); i++ {
			child := element.GetChild(i)
			walk(child, childStart)
			if child != nil {
				childStart += child.ByteFullWidth()
			}
		}
	}
	walk(root, 0)
}

func syntaxNodeForBinderLocation(element SyntaxElement) *SyntaxNode {
	if node, ok := element.(*SyntaxNode); ok {
		return node
	}
	return simpleBinderSyntaxNode(element)
}

func (b *Binder) diagnosticAtSyntaxElement(diagnostic Diagnostic, element SyntaxElement) Diagnostic {
	if token, ok := element.(*SyntaxToken); ok && b != nil {
		if location, found := b.syntaxTokenLocations[token]; found {
			return diagnostic.WithByteLocation(location.Start, location.Length)
		}
	}
	if b != nil {
		if node := syntaxNodeForBinderLocation(element); node != nil {
			if location, found := b.syntaxNodeLocations[node]; found {
				return diagnostic.WithByteLocation(location.Start, location.Length)
			}
		}
	}
	// SyntaxList and SeparatedElement are first-class diagnostic owners in
	// Kusto.Language too. They are attached to the same rooted tree but do not
	// have an underlying SyntaxNode identity for the maps above.
	if element != nil && element.Tree() != nil {
		return diagnostic.WithByteLocation(element.ByteTextStart(), element.ByteWidth())
	}
	return diagnostic
}

// TryBind creates a binder for globals and binds the supported syntax-tree
// surface. It mirrors the entry shape of Binder.TryBind without claiming that
// the complete upstream binder has been ported.
func TryBind(root SyntaxElement, globals *GlobalState) bool {
	return NewBinder(globals).TryBind(root)
}

func TryBindWithCancellation(root SyntaxElement, globals *GlobalState, cancellationToken CancellationToken) bool {
	return NewBinder(globals).TryBindWithCancellation(root, cancellationToken)
}

func (b *Binder) ResolveTable(name string) SemanticInfo {
	if b == nil || b.CurrentDatabase() == nil {
		return NewSemanticInfoWithDiagnostics(DiagnosticNameDoesNotReferToAnyKnownTableOrFunction(name))
	}
	if table := b.CurrentDatabase().GetTable(name); table != nil {
		return NewSemanticInfoForSymbol(table, &table.TypeSymbol)
	}
	return NewSemanticInfoWithDiagnostics(DiagnosticNameDoesNotReferToAnyKnownTableOrFunction(name))
}

func (b *Binder) ResolveFunction(name string) SemanticInfo {
	if b == nil || b.Globals == nil {
		return NewSemanticInfoWithDiagnostics(DiagnosticNameDoesNotReferToAnyKnownFunction(name))
	}
	if fn := b.Globals.GetAnyFunction(name); fn != nil {
		signatures := fn.Signatures()
		if len(signatures) == 1 {
			return NewSemanticInfoForSignature(signatures[0], &fn.TypeSymbol)
		}
		return NewSemanticInfoForSymbol(fn, &fn.TypeSymbol)
	}
	return NewSemanticInfoWithDiagnostics(DiagnosticNameDoesNotReferToAnyKnownFunction(name))
}

func (b *Binder) ResolveColumn(rowScope *TableSymbol, name string) SemanticInfo {
	if rowScope == nil {
		return NewSemanticInfoWithDiagnostics(DiagnosticNameDoesNotReferToAnyKnownItem(name, "column"))
	}
	for _, column := range rowScope.Columns() {
		if column.SymbolName() == name {
			return NewSemanticInfoForSymbol(column, column.Type)
		}
	}
	return NewSemanticInfoWithDiagnostics(DiagnosticNameDoesNotReferToAnyKnownItem(name, "column"))
}

func (b *Binder) ResolveClientParameterReference(name string, inTabularContext bool) SemanticInfo {
	globals := DefaultGlobalState
	if b != nil && b.Globals != nil {
		globals = b.Globals
	}
	if !GlobalStateGetProperty(globals, Properties.AllowClientParameters) {
		return NewSemanticInfoForType(ScalarTypes.Unknown.Type(), DiagnosticClientParametersNotSupported())
	}
	if symbol := globals.GetClientSymbol(name); symbol != nil {
		return NewSemanticInfoForReferencedSymbol(symbol)
	}
	if inTabularContext {
		table := NewTableSymbol(name).WithIsOpen(true)
		return NewSemanticInfoForSymbol(table, &table.TypeSymbol)
	}
	return NewSemanticInfoForType(ScalarTypes.Unknown.Type())
}

func (b *Binder) ValidateFunctionArgumentCount(name string, actual int) []Diagnostic {
	info := b.ResolveFunction(name)
	signature := info.ReferencedSignature()
	if signature == nil {
		return info.Diagnostics
	}
	if signature.IsValidArgumentCount(actual) {
		return []Diagnostic{}
	}
	if signature.MinArgumentCount == signature.MaxArgumentCount {
		return []Diagnostic{DiagnosticFunctionExpectsArgumentCountExact(name, signature.MinArgumentCount)}
	}
	return []Diagnostic{DiagnosticFunctionExpectsArgumentCountRange(name, signature.MinArgumentCount, signature.MaxArgumentCount)}
}
