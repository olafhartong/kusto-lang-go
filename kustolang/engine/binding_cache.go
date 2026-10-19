package kustolang

import "sync"

// Ported from src/Kusto.Language/Binder/GlobalBindingCache.cs and
// src/Kusto.Language/Binder/LocalBindingCache.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type GlobalBindingCache struct {
	mu                           sync.RWMutex
	expansionMu                  sync.Mutex
	UnifiedNameColumnsMap        *TableListCache
	UnifiedNameAndTypeColumnsMap *TableListCache
	CommonColumnsMap             *TableListCache
	CallSiteToExpansionMap       map[*Signature]*MostRecentlyUsedCache[CallSiteInfo, *FunctionCallExpansion]
	CallSiteToResultTypeMap      map[*Signature]*MostRecentlyUsedCache[CallSiteInfo, *TypeSymbol]
	CallSiteToResultSymbolMap    map[*Signature]*MostRecentlyUsedCache[CallSiteInfo, Symbol]
	DatabaseFunctionBodyFacts    map[*Signature]*FunctionBodyFacts
}

func NewGlobalBindingCache() *GlobalBindingCache {
	return &GlobalBindingCache{
		UnifiedNameColumnsMap:        NewTableListCache(),
		UnifiedNameAndTypeColumnsMap: NewTableListCache(),
		CommonColumnsMap:             NewTableListCache(),
		CallSiteToExpansionMap:       map[*Signature]*MostRecentlyUsedCache[CallSiteInfo, *FunctionCallExpansion]{},
		CallSiteToResultTypeMap:      map[*Signature]*MostRecentlyUsedCache[CallSiteInfo, *TypeSymbol]{},
		CallSiteToResultSymbolMap:    map[*Signature]*MostRecentlyUsedCache[CallSiteInfo, Symbol]{},
		DatabaseFunctionBodyFacts:    map[*Signature]*FunctionBodyFacts{},
	}
}

type LocalBindingCache struct {
	SignaturesComputingExpansion map[*Signature]bool
	CallSiteToExpansionMap       *CallSiteValueMap[*FunctionCallExpansion]
	CallSiteToResultTypeMap      *CallSiteValueMap[*TypeSymbol]
	CallSiteToResultSymbolMap    *CallSiteValueMap[Symbol]
	NonDatabaseFunctionBodyFacts map[*Signature]*FunctionBodyFacts
	FunctionExpansionCounts      map[*Signature]int
}

func NewLocalBindingCache() *LocalBindingCache {
	return &LocalBindingCache{
		SignaturesComputingExpansion: map[*Signature]bool{},
		CallSiteToExpansionMap:       NewCallSiteValueMap[*FunctionCallExpansion](),
		CallSiteToResultTypeMap:      NewCallSiteValueMap[*TypeSymbol](),
		CallSiteToResultSymbolMap:    NewCallSiteValueMap[Symbol](),
		NonDatabaseFunctionBodyFacts: map[*Signature]*FunctionBodyFacts{},
		FunctionExpansionCounts:      map[*Signature]int{},
	}
}

func IsDatabaseSymbolSignature(globals *GlobalState, signature *Signature) bool {
	return signature != nil &&
		signature.Symbol != nil &&
		signature.Declaration == nil &&
		signature.Body != "" &&
		globals != nil &&
		globals.IsDatabaseSymbol(signature.Symbol)
}

func TryGetFunctionBodyFacts(global *GlobalBindingCache, local *LocalBindingCache, signature *Signature) (*FunctionBodyFacts, bool) {
	if global != nil {
		if facts, ok := global.tryGetDatabaseFunctionBodyFacts(signature); ok {
			return facts, true
		}
	}
	if local != nil {
		if facts, ok := local.NonDatabaseFunctionBodyFacts[signature]; ok {
			return facts, true
		}
	}
	return nil, false
}

func SetFunctionBodyFacts(globals *GlobalState, global *GlobalBindingCache, local *LocalBindingCache, signature *Signature, facts *FunctionBodyFacts) {
	if IsDatabaseSymbolSignature(globals, signature) {
		if global != nil {
			global.setDatabaseFunctionBodyFacts(signature, facts)
		}
		return
	}
	if local != nil {
		local.NonDatabaseFunctionBodyFacts[signature] = facts
	}
}

func CanCacheResultType(localScope *LocalScope, global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo) bool {
	facts, ok := TryGetFunctionBodyFacts(global, local, callSite.Signature)
	if !ok || facts == nil {
		return false
	}
	if !facts.HasVariableReturnType() {
		return false
	}
	if facts.HasUnqualifiedTableCall() {
		if isLocalTabularVariable(localScope, facts.UnqualifiedTableNames()...) {
			return false
		}
		for _, value := range callSite.Values {
			if possibleTableName, ok := value.(string); ok && isLocalTabularVariable(localScope, possibleTableName) {
				return false
			}
		}
	}
	return true
}

func TryGetResultTypeFromCache(global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo) (*TypeSymbol, bool) {
	if local != nil {
		if typ, ok := local.CallSiteToResultTypeMap.TryGetValue(callSite); ok {
			return typ, true
		}
	}
	if global != nil {
		if mru, ok := global.tryGetResultTypeCache(callSite.Signature); ok {
			return mru.TryGetValue(callSite)
		}
	}
	return nil, false
}

func AddResultTypeToCache(globals *GlobalState, global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo, resultType *TypeSymbol) {
	if IsDatabaseSymbolSignature(globals, callSite.Signature) {
		if global == nil {
			return
		}
		mru := global.getOrCreateResultTypeCache(callSite.Signature, GlobalStateGetProperty(globals, Properties.MaxCachedResultTypes))
		mru.AddOrUpdate(callSite, resultType)
		return
	}
	if local != nil {
		local.CallSiteToResultTypeMap.Add(callSite, resultType)
	}
}

// The source hierarchy models tabular schemas as TypeSymbol instances. Go
// keeps TableSymbol separate so callers can retain its columns. This companion
// cache mirrors the source result-type cache while preserving the full result
// symbol for computed function calls.
func TryGetResultSymbolFromCache(global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo) (Symbol, bool) {
	if local != nil {
		if symbol, ok := local.CallSiteToResultSymbolMap.TryGetValue(callSite); ok {
			return symbol, true
		}
	}
	if global != nil {
		if mru, ok := global.tryGetResultSymbolCache(callSite.Signature); ok {
			return mru.TryGetValue(callSite)
		}
	}
	return nil, false
}

func AddResultSymbolToCache(globals *GlobalState, global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo, resultSymbol Symbol) {
	if resultSymbol == nil {
		return
	}
	if IsDatabaseSymbolSignature(globals, callSite.Signature) {
		if global == nil {
			return
		}
		mru := global.getOrCreateResultSymbolCache(callSite.Signature, GlobalStateGetProperty(globals, Properties.MaxCachedResultTypes))
		mru.AddOrUpdate(callSite, resultSymbol)
		return
	}
	if local != nil {
		local.CallSiteToResultSymbolMap.Add(callSite, resultSymbol)
	}
}

func CanCacheExpansion(localScope *LocalScope, global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo) bool {
	facts, ok := TryGetFunctionBodyFacts(global, local, callSite.Signature)
	if !ok || facts == nil {
		return false
	}
	if !facts.HasUnqualifiedTableCall() {
		return true
	}
	if isLocalTabularVariable(localScope, facts.UnqualifiedTableNames()...) {
		return false
	}
	for _, dependentParameter := range facts.DependentParameters() {
		index := indexOfCallSiteParameter(callSite.Parameters, dependentParameter)
		if index >= 0 && index < len(callSite.Values) {
			if possibleTableName, ok := callSite.Values[index].(string); ok && isLocalTabularVariable(localScope, possibleTableName) {
				return false
			}
		}
	}
	return true
}

func TryGetExpansionFromCache(global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo) (*FunctionCallExpansion, bool) {
	if local != nil {
		if expansion, ok := local.CallSiteToExpansionMap.TryGetValue(callSite); ok {
			return expansion, true
		}
	}
	if global != nil {
		if mru, ok := global.tryGetExpansionCache(callSite.Signature); ok {
			return mru.TryGetValue(callSite)
		}
	}
	return nil, false
}

func AddExpansionToCache(globals *GlobalState, global *GlobalBindingCache, local *LocalBindingCache, callSite CallSiteInfo, expansion *FunctionCallExpansion) {
	if IsDatabaseSymbolSignature(globals, callSite.Signature) {
		if global == nil {
			return
		}
		mru := global.getOrCreateExpansionCache(callSite.Signature, GlobalStateGetProperty(globals, Properties.MaxCachedExpansions))
		mru.AddOrUpdate(callSite, expansion)
		return
	}
	if local != nil {
		local.CallSiteToExpansionMap.Add(callSite, expansion)
	}
}

// GlobalBindingCache mirrors the upstream ThreadSafeDictionary-backed cache.
// Each returned MRU cache is independently synchronized, so the dictionary
// lock only protects map lookup and creation.
func (c *GlobalBindingCache) tryGetDatabaseFunctionBodyFacts(signature *Signature) (*FunctionBodyFacts, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	facts, ok := c.DatabaseFunctionBodyFacts[signature]
	c.mu.RUnlock()
	return facts, ok
}

func (c *GlobalBindingCache) setDatabaseFunctionBodyFacts(signature *Signature, facts *FunctionBodyFacts) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.DatabaseFunctionBodyFacts[signature] = facts
	c.mu.Unlock()
}

func (c *GlobalBindingCache) tryGetResultTypeCache(signature *Signature) (*MostRecentlyUsedCache[CallSiteInfo, *TypeSymbol], bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	cache, ok := c.CallSiteToResultTypeMap[signature]
	c.mu.RUnlock()
	return cache, ok
}

func (c *GlobalBindingCache) getOrCreateResultTypeCache(signature *Signature, maxSize int) *MostRecentlyUsedCache[CallSiteInfo, *TypeSymbol] {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.CallSiteToResultTypeMap[signature]
	if cache == nil {
		cache = NewMostRecentlyUsedCache[CallSiteInfo, *TypeSymbol](maxSize)
		c.CallSiteToResultTypeMap[signature] = cache
	}
	return cache
}

func (c *GlobalBindingCache) tryGetResultSymbolCache(signature *Signature) (*MostRecentlyUsedCache[CallSiteInfo, Symbol], bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	cache, ok := c.CallSiteToResultSymbolMap[signature]
	c.mu.RUnlock()
	return cache, ok
}

func (c *GlobalBindingCache) getOrCreateResultSymbolCache(signature *Signature, maxSize int) *MostRecentlyUsedCache[CallSiteInfo, Symbol] {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.CallSiteToResultSymbolMap[signature]
	if cache == nil {
		cache = NewMostRecentlyUsedCache[CallSiteInfo, Symbol](maxSize)
		c.CallSiteToResultSymbolMap[signature] = cache
	}
	return cache
}

func (c *GlobalBindingCache) tryGetExpansionCache(signature *Signature) (*MostRecentlyUsedCache[CallSiteInfo, *FunctionCallExpansion], bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	cache, ok := c.CallSiteToExpansionMap[signature]
	c.mu.RUnlock()
	return cache, ok
}

func (c *GlobalBindingCache) getOrCreateExpansionCache(signature *Signature, maxSize int) *MostRecentlyUsedCache[CallSiteInfo, *FunctionCallExpansion] {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache := c.CallSiteToExpansionMap[signature]
	if cache == nil {
		cache = NewMostRecentlyUsedCache[CallSiteInfo, *FunctionCallExpansion](maxSize)
		c.CallSiteToExpansionMap[signature] = cache
	}
	return cache
}

type TableListCache struct {
	mu      sync.RWMutex
	entries []tableListCacheEntry
}

type tableListCacheEntry struct {
	tables []*TableSymbol
	value  *TableSymbol
}

func NewTableListCache() *TableListCache {
	return &TableListCache{}
}

func (c *TableListCache) TryGetValue(tables []*TableSymbol) (*TableSymbol, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, entry := range c.entries {
		if sameTableList(entry.tables, tables) {
			return entry.value, true
		}
	}
	return nil, false
}

func (c *TableListCache) AddOrUpdate(tables []*TableSymbol, value *TableSymbol) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	key := append([]*TableSymbol{}, tables...)
	for i := range c.entries {
		if sameTableList(c.entries[i].tables, tables) {
			c.entries[i] = tableListCacheEntry{tables: key, value: value}
			return
		}
	}
	c.entries = append(c.entries, tableListCacheEntry{tables: key, value: value})
}

func (c *TableListCache) Count() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

type CallSiteValueMap[V any] struct {
	entries []callSiteValueEntry[V]
}

type callSiteValueEntry[V any] struct {
	key   CallSiteInfo
	value V
}

func NewCallSiteValueMap[V any]() *CallSiteValueMap[V] {
	return &CallSiteValueMap[V]{}
}

func (m *CallSiteValueMap[V]) TryGetValue(key CallSiteInfo) (V, bool) {
	var zero V
	if m == nil {
		return zero, false
	}
	for _, entry := range m.entries {
		if entry.key.Equals(key) {
			return entry.value, true
		}
	}
	return zero, false
}

func (m *CallSiteValueMap[V]) Add(key CallSiteInfo, value V) {
	if m == nil {
		return
	}
	for _, entry := range m.entries {
		if entry.key.Equals(key) {
			panic("duplicate CallSiteInfo key")
		}
	}
	m.entries = append(m.entries, callSiteValueEntry[V]{key: key, value: value})
}

func (m *CallSiteValueMap[V]) Count() int {
	if m == nil {
		return 0
	}
	return len(m.entries)
}

func sameTableList(left []*TableSymbol, right []*TableSymbol) bool {
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

func isLocalTabularVariable(scope *LocalScope, names ...string) bool {
	for _, name := range names {
		if name != "" && scope != nil && scope.ContainsSymbolMatch(name, SymbolMatchTabular|SymbolMatchLocal) {
			return true
		}
	}
	return false
}

func indexOfCallSiteParameter(parameters []*Parameter, target *Parameter) int {
	for i, parameter := range parameters {
		if parameter == target {
			return i
		}
	}
	return -1
}
