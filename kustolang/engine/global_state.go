package kustolang

import (
	"strings"
	"sync"
)

// Ported from src/Kusto.Language/GlobalState.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type GlobalState struct {
	Domain          string
	Clusters        []*ClusterSymbol
	Cluster         *ClusterSymbol
	Database        *DatabaseSymbol
	Functions       []*FunctionSymbol
	Aggregates      []*FunctionSymbol
	PlugIns         []*FunctionSymbol
	Operators       []*OperatorSymbol
	Options         []*OptionSymbol
	ServerKind      string
	ParseOptions    ParseOptions
	Cache           *KustoCache
	AmbientSymbols  []Symbol
	ClientSymbols   []Symbol
	PropertyValues  []GlobalStatePropertyAndValue
	lookupCache     *globalStateLookupCache
	plugInsExplicit bool
}

// globalStateLookupCache holds immutable GlobalState's lazily constructed
// lookup maps. It replaces the source's Interlocked.CompareExchange caches.
// A state clone gets a fresh cache so its maps cannot describe an older catalog.
type globalStateLookupCache struct {
	mu           sync.Mutex
	commandMap   map[string]*CommandSymbol
	operatorMap  map[OperatorKind]*OperatorSymbol
	functionMap  map[string]*FunctionSymbol
	aggregateMap map[string]*FunctionSymbol
	plugInMap    map[string]*FunctionSymbol
	optionMap    map[string]*OptionSymbol
	ambientMap   map[string]Symbol
	clientMap    map[string]Symbol
}

func newGlobalStateLookupCache() *globalStateLookupCache {
	return &globalStateLookupCache{}
}

const KustoWindowsNet = ".kusto.windows.net"

var DefaultGlobalState = newDefaultGlobalState()

func newDefaultGlobalState() *GlobalState {
	return &GlobalState{
		Domain:       KustoWindowsNet,
		Clusters:     EmptyReadOnlyList[*ClusterSymbol](),
		Cluster:      UnknownClusterSymbol,
		Database:     UnknownDatabaseSymbol,
		Functions:    defaultFunctions(),
		Aggregates:   defaultAggregates(),
		PlugIns:      defaultPlugIns(),
		Operators:    defaultOperators(),
		Options:      defaultOptions(),
		ServerKind:   ServerKindEngine,
		ParseOptions: DefaultParseOptions,
		lookupCache:  newGlobalStateLookupCache(),
	}
}

// NewGlobalState constructs the Go compatibility state used by analyzer
// environments. A nil database retains the compatibility closed "db" namespace;
// use DefaultGlobalState for the source-compatible unknown environment.
func NewGlobalState(database *DatabaseSymbol) *GlobalState {
	if database == nil {
		database = NewDatabaseSymbol("db")
	}
	return newGlobalState(database)
}

func newGlobalState(database *DatabaseSymbol) *GlobalState {
	cluster := NewClusterSymbol(database.SymbolName()+":cluster", database)
	return &GlobalState{
		Domain:       KustoWindowsNet,
		Clusters:     []*ClusterSymbol{cluster},
		Cluster:      cluster,
		Database:     database,
		Functions:    defaultFunctions(),
		Aggregates:   defaultAggregates(),
		PlugIns:      defaultPlugIns(),
		Operators:    defaultOperators(),
		Options:      defaultOptions(),
		ServerKind:   ServerKindEngine,
		ParseOptions: DefaultParseOptions,
		lookupCache:  newGlobalStateLookupCache(),
	}
}

// Copy returns a distinct GlobalState with the same immutable symbol lists.
// When caching is enabled, the copy receives a fresh cache bound to itself.
func (g *GlobalState) Copy() *GlobalState {
	if g == nil {
		return nil
	}
	clone := *g
	clone.lookupCache = newGlobalStateLookupCache()
	if clone.Cache != nil {
		clone.Cache = clone.Cache.WithGlobals(&clone)
	}
	return &clone
}

func (g *GlobalState) WithDatabase(database *DatabaseSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if database == nil {
		database = UnknownDatabaseSymbol
	}
	if g.Database == database {
		return g
	}
	if database == UnknownDatabaseSymbol || clusterContainsDatabase(g.Cluster, database) {
		clone := *g
		clone.Database = database
		return clone.rebindCache()
	}
	if cluster := g.GetClusterForDatabase(database); cluster != nil {
		clone := *g
		clone.Cluster = cluster
		clone.Database = database
		return clone.rebindCache()
	}
	cluster := NewClusterSymbol(database.SymbolName()+":cluster", database)
	return g.WithCluster(cluster).WithDatabase(database)
}

func (g *GlobalState) WithDatabaseName(databaseName string) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if g.Cluster == nil {
		return g.WithDatabase(UnknownDatabaseSymbol)
	}
	return g.WithDatabase(g.Cluster.GetDatabase(databaseName))
}

func (g *GlobalState) WithClusterList(clusters []*ClusterSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.Clusters = append([]*ClusterSymbol{}, clusters...)
	clone.Cluster = UnknownClusterSymbol
	clone.Database = UnknownDatabaseSymbol
	if g.Cluster != nil {
		for _, cluster := range clone.Clusters {
			if cluster != nil && cluster.SymbolName() == g.Cluster.SymbolName() {
				clone.Cluster = cluster
				if g.Database != nil {
					clone.Database = cluster.GetDatabase(g.Database.SymbolName())
				}
				if clone.Database == nil {
					clone.Database = UnknownDatabaseSymbol
				}
				break
			}
		}
	}
	return clone.rebindCache()
}

func (g *GlobalState) AddOrReplaceCluster(cluster *ClusterSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if cluster == nil || cluster == UnknownClusterSymbol || cluster == g.Cluster || clusterInList(g.Clusters, cluster) {
		return g
	}
	return g.WithClusterList(addOrReplaceCluster(g.Clusters, cluster))
}

func (g *GlobalState) WithCluster(cluster *ClusterSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if g.Cluster == cluster {
		return g
	}
	if cluster == nil || cluster == UnknownClusterSymbol {
		clone := *g
		clone.Cluster = UnknownClusterSymbol
		clone.Database = UnknownDatabaseSymbol
		return clone.rebindCache()
	}
	if !clusterInList(g.Clusters, cluster) {
		withList := g.WithClusterList(addOrReplaceCluster(g.Clusters, cluster))
		return withList.WithCluster(cluster)
	}
	clone := *g
	clone.Cluster = cluster
	clone.Database = cluster.GetDatabase(g.Database.SymbolName())
	if clone.Database == nil {
		clone.Database = UnknownDatabaseSymbol
	}
	return clone.rebindCache()
}

func (g *GlobalState) WithClusterName(clusterName string) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	return g.WithCluster(g.GetCluster(clusterName))
}

func (g *GlobalState) WithDomain(domain string) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if domain != "" && !strings.HasPrefix(domain, ".") {
		domain = "." + domain
	}
	if g.Domain == domain {
		return g
	}
	clone := *g
	clone.Domain = domain
	return clone.rebindCache()
}

func (g *GlobalState) GetCluster(name string) *ClusterSymbol {
	if g == nil {
		g = DefaultGlobalState
	}
	name = GetHostName(name)
	if name == "" {
		return nil
	}
	if g.Cluster != nil && g.Cluster != UnknownClusterSymbol &&
		(clusterHostNamesEqual(name, g.Cluster.SymbolName()) || isShortClusterHostName(name, g.Cluster.SymbolName(), g.Domain)) {
		return g.Cluster
	}
	for _, cluster := range g.Clusters {
		if cluster != nil && clusterHostNamesEqual(name, cluster.SymbolName()) {
			return cluster
		}
	}
	for _, cluster := range g.Clusters {
		if cluster != nil && isShortClusterHostName(name, cluster.SymbolName(), g.Domain) {
			return cluster
		}
	}
	return nil
}

func (g *GlobalState) GetClusterForDatabase(database *DatabaseSymbol) *ClusterSymbol {
	if g == nil || database == nil {
		return nil
	}
	for _, cluster := range g.Clusters {
		if clusterContainsDatabase(cluster, database) {
			return cluster
		}
	}
	return nil
}

func clusterContainsDatabase(cluster *ClusterSymbol, database *DatabaseSymbol) bool {
	if cluster == nil || database == nil {
		return false
	}
	for _, candidate := range cluster.Databases() {
		if candidate == database {
			return true
		}
	}
	return false
}

func clusterInList(clusters []*ClusterSymbol, target *ClusterSymbol) bool {
	for _, cluster := range clusters {
		if cluster == target {
			return true
		}
	}
	return false
}

func addOrReplaceCluster(clusters []*ClusterSymbol, replacement *ClusterSymbol) []*ClusterSymbol {
	out := append([]*ClusterSymbol{}, clusters...)
	for i, cluster := range out {
		if cluster != nil && cluster.SymbolName() == replacement.SymbolName() {
			out[i] = replacement
			return out
		}
	}
	return append(out, replacement)
}

func clusterHostNamesEqual(left, right string) bool {
	return strings.EqualFold(left, right)
}

func isShortClusterHostName(name, hostName, domain string) bool {
	if domain == "" || !strings.HasPrefix(domain, ".") || len(hostName) <= len(domain) ||
		!strings.HasSuffix(strings.ToLower(hostName), strings.ToLower(domain)) {
		return false
	}
	shortLength := len(hostName) - len(domain)
	short := hostName[:shortLength]
	if short == "" || strings.HasSuffix(short, ".") || strings.Count(short, ".") >= 2 {
		return false
	}
	return len(name)+len(domain) == len(hostName) && strings.EqualFold(name, short)
}

func (g *GlobalState) WithOptions(options []*OptionSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.Options = append([]*OptionSymbol{}, options...)
	return clone.rebindCache()
}

func (g *GlobalState) WithFunctions(functions []*FunctionSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.Functions = append([]*FunctionSymbol{}, functions...)
	return clone.rebindCache()
}

func (g *GlobalState) WithAggregates(aggregates []*FunctionSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.Aggregates = append([]*FunctionSymbol{}, aggregates...)
	return clone.rebindCache()
}

func (g *GlobalState) WithPlugIns(plugIns []*FunctionSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.PlugIns = append([]*FunctionSymbol{}, plugIns...)
	clone.plugInsExplicit = true
	return clone.rebindCache()
}

// WithOperators returns a state whose scalar operator catalog is operators.
func (g *GlobalState) WithOperators(operators []*OperatorSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.Operators = append([]*OperatorSymbol{}, operators...)
	return clone.rebindCache()
}

func (g *GlobalState) WithServerKind(serverKind string) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if g.ServerKind == serverKind {
		return g
	}
	clone := *g
	clone.ServerKind = serverKind
	return clone.rebindCache()
}

func (g *GlobalState) WithParseOptions(parseOptions ParseOptions) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if g.ParseOptions.Equal(parseOptions) {
		return g
	}
	clone := *g
	clone.ParseOptions = parseOptions
	return clone.rebindCache()
}

func (g *GlobalState) WithAmbientSymbols(symbols []Symbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.AmbientSymbols = ListToReadOnly(symbols)
	return clone.rebindCache()
}

func (g *GlobalState) AddOrUpdateAmbientSymbols(symbols ...Symbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if !hasAnySymbol(symbols) {
		return g
	}
	return g.WithAmbientSymbols(ListAddOrUpdate(g.ambientSymbols(), symbols, symbolNameKey))
}

func (g *GlobalState) GetAmbientSymbol(name string) Symbol {
	if g == nil {
		g = DefaultGlobalState
	}
	if len(g.AmbientSymbols) == 0 {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.ambientMap == nil {
			cache.ambientMap = symbolsByName(g.AmbientSymbols)
		}
		return cache.ambientMap[name]
	}
	return lastNamedSymbol(g.AmbientSymbols, name)
}

// Parameters returns ambient ParameterSymbol values. It preserves the
// obsolete upstream compatibility property in terms of AmbientSymbols.
func (g *GlobalState) Parameters() []*ParameterSymbol {
	if g == nil {
		return EmptyReadOnlyList[*ParameterSymbol]()
	}
	parameters := make([]*ParameterSymbol, 0)
	for _, symbol := range g.AmbientSymbols {
		if parameter, ok := symbol.(*ParameterSymbol); ok && parameter != nil {
			parameters = append(parameters, parameter)
		}
	}
	return ListToReadOnly(parameters)
}

// WithParameters replaces ambient ParameterSymbol values while retaining
// every non-parameter ambient symbol.
func (g *GlobalState) WithParameters(parameters []*ParameterSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	symbols := make([]Symbol, 0, len(g.AmbientSymbols)+len(parameters))
	for _, symbol := range g.AmbientSymbols {
		if _, isParameter := symbol.(*ParameterSymbol); !isParameter && !listItemIsNil(symbol) {
			symbols = append(symbols, symbol)
		}
	}
	for _, parameter := range parameters {
		if parameter != nil {
			symbols = append(symbols, parameter)
		}
	}
	return g.WithAmbientSymbols(symbols)
}

// AddParameters preserves the obsolete upstream convenience overload.
func (g *GlobalState) AddParameters(parameters ...*ParameterSymbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	symbols := make([]Symbol, 0, len(parameters))
	for _, parameter := range parameters {
		if parameter != nil {
			symbols = append(symbols, parameter)
		}
	}
	return g.AddOrUpdateAmbientSymbols(symbols...)
}

func (g *GlobalState) WithClientSymbols(symbols []Symbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	clone := *g
	clone.ClientSymbols = ListToReadOnly(symbols)
	return clone.rebindCache()
}

func (g *GlobalState) AddOrUpdateClientSymbols(symbols ...Symbol) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if !hasAnySymbol(symbols) {
		return g
	}
	return g.WithClientSymbols(ListAddOrUpdate(g.clientSymbols(), symbols, symbolNameKey))
}

func (g *GlobalState) GetClientSymbol(name string) Symbol {
	if g == nil {
		g = DefaultGlobalState
	}
	if len(g.ClientSymbols) == 0 {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.clientMap == nil {
			cache.clientMap = symbolsByName(g.ClientSymbols)
		}
		return cache.clientMap[name]
	}
	return lastNamedSymbol(g.ClientSymbols, name)
}

func (g *GlobalState) GetProperty(property GlobalStateProperty) any {
	if property == nil {
		return nil
	}
	switch p := property.(type) {
	case *TypedGlobalStateProperty[bool]:
		return GlobalStateGetProperty(g, p)
	case *TypedGlobalStateProperty[int]:
		return GlobalStateGetProperty(g, p)
	default:
		for i := len(g.propertyValues()) - 1; i >= 0; i-- {
			pv := g.propertyValues()[i]
			if pv.Property == property {
				return pv.Value
			}
		}
		return nil
	}
}

func GlobalStateGetProperty[T any](g *GlobalState, property *TypedGlobalStateProperty[T]) T {
	var zero T
	if property == nil {
		return zero
	}
	values := g.propertyValues()
	for i := len(values) - 1; i >= 0; i-- {
		pv := values[i]
		if pv.Property == property {
			if value, ok := pv.Value.(T); ok {
				return value
			}
			return zero
		}
	}
	return property.DefaultValue()
}

func GlobalStateWithProperty[T any](g *GlobalState, property *TypedGlobalStateProperty[T], value T) *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if property == nil {
		return g
	}

	values := g.propertyValues()
	next := make([]GlobalStatePropertyAndValue, 0, len(values)+1)
	replaced := false
	for _, current := range values {
		if current.Property == property && !replaced {
			if globalStatePropertyValuesEqual(current.Value, value) {
				return g
			}
			next = append(next, GlobalStatePropertyAndValue{Property: property, Value: value})
			replaced = true
			continue
		}
		next = append(next, current)
	}
	if !replaced {
		next = append(next, GlobalStatePropertyAndValue{Property: property, Value: value})
	}

	clone := *g
	clone.PropertyValues = ListToReadOnly(next)
	return clone.rebindCache()
}

func (g *GlobalState) propertyValues() []GlobalStatePropertyAndValue {
	if g == nil {
		g = DefaultGlobalState
	}
	if len(g.PropertyValues) == 0 {
		return EmptyReadOnlyList[GlobalStatePropertyAndValue]()
	}
	return g.PropertyValues
}

func (g *GlobalState) ambientSymbols() []Symbol {
	if g == nil {
		g = DefaultGlobalState
	}
	if len(g.AmbientSymbols) == 0 {
		return EmptyReadOnlyList[Symbol]()
	}
	return g.AmbientSymbols
}

func (g *GlobalState) clientSymbols() []Symbol {
	if g == nil {
		g = DefaultGlobalState
	}
	if len(g.ClientSymbols) == 0 {
		return EmptyReadOnlyList[Symbol]()
	}
	return g.ClientSymbols
}

func (g *GlobalState) WithCache() *GlobalState {
	if g == nil {
		g = DefaultGlobalState
	}
	if g.Cache != nil {
		return g
	}
	clone := *g
	clone.lookupCache = newGlobalStateLookupCache()
	clone.Cache = NewKustoCache(&clone)
	return &clone
}

func (g *GlobalState) rebindCache() *GlobalState {
	if g != nil {
		g.lookupCache = newGlobalStateLookupCache()
	}
	if g.Cache != nil {
		g.Cache = g.Cache.WithGlobals(g)
	}
	return g
}

func (g *GlobalState) GetAggregate(name string) *FunctionSymbol {
	if g == nil || name == "" {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.aggregateMap == nil {
			cache.aggregateMap = functionsByName(g.Aggregates)
		}
		return cache.aggregateMap[name]
	}
	return lastNamedFunction(g.Aggregates, name)
}

func (g *GlobalState) IsAggregateFunction(fn *FunctionSymbol) bool {
	return fn != nil && g != nil && g.GetAggregate(fn.SymbolName()) == fn
}

func (g *GlobalState) IsBuiltInFunction(fn *FunctionSymbol) bool {
	return fn != nil && g != nil &&
		(g.GetFunction(fn.SymbolName()) == fn ||
			g.GetAggregate(fn.SymbolName()) == fn ||
			g.GetPlugIn(fn.SymbolName()) == fn)
}

func (g *GlobalState) IsBuiltInFunctionName(name string) bool {
	return g != nil && name != "" &&
		(g.GetFunction(name) != nil || g.GetAggregate(name) != nil || g.GetPlugIn(name) != nil)
}

// GetFunction returns the last ordinary built-in function with the exact name.
// Aggregates and plug-ins are available through GetAggregate and GetPlugIn.
func (g *GlobalState) GetFunction(name string) *FunctionSymbol {
	if g == nil || name == "" {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.functionMap == nil {
			cache.functionMap = functionsByName(g.Functions)
		}
		return cache.functionMap[name]
	}
	return lastNamedFunction(g.Functions, name)
}

// GetAnyFunction preserves the port's original convenience lookup: current
// database functions first, followed by ordinary functions, aggregates, and
// plug-ins. Binding code should prefer the scope-specific source APIs.
func (g *GlobalState) GetAnyFunction(name string) *FunctionSymbol {
	if g == nil || name == "" {
		return nil
	}
	if g.Database != nil {
		if fn := g.Database.GetFunction(name); fn != nil {
			return fn
		}
	}
	for _, groups := range [][]*FunctionSymbol{g.Functions, g.Aggregates, g.PlugIns} {
		for _, fn := range groups {
			if matchesSymbolName(fn, name, false) {
				return fn
			}
		}
	}
	return nil
}

func (g *GlobalState) GetCommand(name string) *CommandSymbol {
	if g == nil {
		g = DefaultGlobalState
	}
	if name == "" {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.commandMap == nil {
			cache.commandMap = commandsByName(globalStateCommands(g.ServerKind))
		}
		return cache.commandMap[name]
	}
	return lastNamedCommand(globalStateCommands(g.ServerKind), name)
}

func commandsByName(commands []*CommandSymbol) map[string]*CommandSymbol {
	commandMap := make(map[string]*CommandSymbol, len(commands))
	for _, command := range commands {
		if command != nil {
			commandMap[command.SymbolName()] = command
		}
	}
	return commandMap
}

func globalStateCommands(serverKind string) []*CommandSymbol {
	switch serverKind {
	case ServerKindEngine:
		return EngineCommands.All
	case ServerKindDataManager:
		return DataManagerCommands.All
	case ServerKindClusterManager:
		return ClusterManagerCommands.All
	case ServerKindAriaBridge:
		return AriaBridgeCommands.All
	default:
		return EmptyReadOnlyList[*CommandSymbol]()
	}
}

func (g *GlobalState) GetPlugIn(name string) *FunctionSymbol {
	if g == nil || name == "" {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.plugInMap == nil {
			cache.plugInMap = functionsByName(g.PlugIns)
		}
		return cache.plugInMap[name]
	}
	return lastNamedFunction(g.PlugIns, name)
}

func (g *GlobalState) GetOperator(kind OperatorKind) *OperatorSymbol {
	if g == nil || len(g.Operators) == 0 {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.operatorMap == nil {
			cache.operatorMap = operatorsByKind(g.Operators)
		}
		return cache.operatorMap[kind]
	}
	return lastOperatorByKind(g.Operators, kind)
}

func (g *GlobalState) GetOption(name string) *OptionSymbol {
	if g == nil || name == "" {
		return nil
	}
	if cache := g.lookupCache; cache != nil {
		cache.mu.Lock()
		defer cache.mu.Unlock()
		if cache.optionMap == nil {
			cache.optionMap = optionsByName(g.Options)
		}
		return cache.optionMap[name]
	}
	return lastNamedOption(g.Options, name)
}

func (g *GlobalState) IsDatabaseTable(symbol Symbol) bool {
	_, ok := tableSymbolFromSymbol(symbol)
	return ok && g.GetDatabase(symbol) != nil
}

func (g *GlobalState) IsDatabaseFunction(symbol Symbol) bool {
	_, ok := symbol.(*FunctionSymbol)
	return ok && g.GetDatabase(symbol) != nil
}

func (g *GlobalState) IsDatabaseSymbol(symbol Symbol) bool {
	return g.GetDatabase(symbol) != nil
}

func (g *GlobalState) GetDatabase(symbol Symbol) *DatabaseSymbol {
	if g == nil || symbol == nil {
		return nil
	}
	for _, cluster := range g.Clusters {
		if cluster == nil {
			continue
		}
		// GlobalState and the symbol implementations share a package, so these
		// read-only searches can use the immutable backing slices directly. The
		// public accessors intentionally return defensive copies and are much too
		// costly on this binding hot path.
		for _, clusterMember := range cluster.members {
			database, ok := clusterMember.(*DatabaseSymbol)
			if !ok {
				continue
			}
			if database == nil {
				continue
			}
			for _, member := range database.members {
				if member == symbol {
					return database
				}
				memberTable, memberIsTable := tableSymbolFromSymbol(member)
				symbolTable, symbolIsTable := tableSymbolFromSymbol(symbol)
				if memberIsTable && symbolIsTable && memberTable == symbolTable {
					return database
				}
			}
		}
	}
	return nil
}

func (g *GlobalState) GetTable(column *ColumnSymbol) *TableSymbol {
	if g == nil || column == nil {
		return nil
	}
	for _, cluster := range g.Clusters {
		if cluster == nil {
			continue
		}
		for _, clusterMember := range cluster.members {
			database, ok := clusterMember.(*DatabaseSymbol)
			if !ok {
				continue
			}
			if database == nil {
				continue
			}
			for _, member := range database.members {
				table, ok := member.(*TableSymbol)
				if !ok || table.IsExternal() || table.IsMaterializedView() || table.IsStoredQueryResult() {
					continue
				}
				for _, tableColumn := range table.columns {
					if tableColumn == column {
						return table
					}
				}
			}
		}
	}
	return nil
}

func functionsByName(functions []*FunctionSymbol) map[string]*FunctionSymbol {
	result := make(map[string]*FunctionSymbol, len(functions))
	for _, function := range functions {
		if function != nil {
			result[function.SymbolName()] = function
		}
	}
	return result
}

func optionsByName(options []*OptionSymbol) map[string]*OptionSymbol {
	result := make(map[string]*OptionSymbol, len(options))
	for _, option := range options {
		if option != nil {
			result[option.SymbolName()] = option
		}
	}
	return result
}

func operatorsByKind(operators []*OperatorSymbol) map[OperatorKind]*OperatorSymbol {
	result := make(map[OperatorKind]*OperatorSymbol, len(operators))
	for _, operator := range operators {
		if operator != nil {
			result[operator.OperatorKind] = operator
		}
	}
	return result
}

func lastNamedSymbol(symbols []Symbol, name string) Symbol {
	for index := len(symbols) - 1; index >= 0; index-- {
		if symbol := symbols[index]; !listItemIsNil(symbol) && symbol.SymbolName() == name {
			return symbol
		}
	}
	return nil
}

func lastNamedFunction(functions []*FunctionSymbol, name string) *FunctionSymbol {
	for index := len(functions) - 1; index >= 0; index-- {
		if function := functions[index]; function != nil && function.SymbolName() == name {
			return function
		}
	}
	return nil
}

func lastNamedCommand(commands []*CommandSymbol, name string) *CommandSymbol {
	for index := len(commands) - 1; index >= 0; index-- {
		if command := commands[index]; command != nil && command.SymbolName() == name {
			return command
		}
	}
	return nil
}

func lastNamedOption(options []*OptionSymbol, name string) *OptionSymbol {
	for index := len(options) - 1; index >= 0; index-- {
		if option := options[index]; option != nil && option.SymbolName() == name {
			return option
		}
	}
	return nil
}

func lastOperatorByKind(operators []*OperatorSymbol, kind OperatorKind) *OperatorSymbol {
	for index := len(operators) - 1; index >= 0; index-- {
		if operator := operators[index]; operator != nil && operator.OperatorKind == kind {
			return operator
		}
	}
	return nil
}

func defaultOptions() []*OptionSymbol {
	return append([]*OptionSymbol{}, Options.All...)
}

func defaultFunctions() []*FunctionSymbol {
	return append([]*FunctionSymbol{}, Functions.All...)
}

func defaultAggregates() []*FunctionSymbol {
	return append([]*FunctionSymbol{}, Aggregates.All...)
}

func defaultPlugIns() []*FunctionSymbol {
	return append([]*FunctionSymbol{}, PlugIns.KnownCurrent...)
}

func defaultOperators() []*OperatorSymbol {
	return append([]*OperatorSymbol{}, Operators.All...)
}

func symbolNameKey(symbol Symbol) string {
	if listItemIsNil(symbol) {
		return ""
	}
	return symbol.SymbolName()
}

func symbolsByName(symbols []Symbol) map[string]Symbol {
	out := make(map[string]Symbol, len(symbols))
	for _, symbol := range symbols {
		if listItemIsNil(symbol) {
			continue
		}
		out[symbol.SymbolName()] = symbol
	}
	return out
}

func hasAnySymbol(symbols []Symbol) bool {
	for _, symbol := range symbols {
		if !listItemIsNil(symbol) {
			return true
		}
	}
	return false
}
