package kustolang

// Ported from the public purpose of src/Kusto.Language/Binder/Binder_Names.cs:
// resolve ordinary name references against the currently modeled scopes. This
// deliberately stays API-level until the full tree binder/scope builder is
// ported.

func (b *Binder) ResolveNameReference(reference *NameReference, inTabularContext bool) SemanticInfo {
	return b.ResolveNameReferenceInRowScope(nil, reference, inTabularContext)
}

func (b *Binder) resolveFunctionCallName(reference *NameReference) SemanticInfo {
	if b != nil && b.pathScope != nil {
		name := NameReferenceSimpleName(reference)
		switch scope := b.pathScope.(type) {
		case *ClusterSymbol:
			if name == Functions.Database.SymbolName() {
				return NewSemanticInfoForReferencedSymbol(Functions.Database)
			}
		case *DatabaseSymbol:
			if source := databaseSourceFunction(name); source != nil {
				return NewSemanticInfoForReferencedSymbol(source)
			}
		case *GroupSymbol:
			for _, member := range scope.Members() {
				switch member.(type) {
				case *ClusterSymbol:
					if name == Functions.Database.SymbolName() {
						return NewSemanticInfoForReferencedSymbol(Functions.Database)
					}
				case *DatabaseSymbol:
					if source := databaseSourceFunction(name); source != nil {
						return NewSemanticInfoForReferencedSymbol(source)
					}
				}
			}
		}
		members := FindMembers(b.pathScope, name, SymbolMatchFunction, false)
		if len(members) == 1 {
			return NewSemanticInfoForReferencedSymbol(members[0])
		}
		if len(members) > 1 {
			diagnostic := b.diagnosticAtSyntaxElement(DiagnosticNameRefersToMoreThanOneItem(name), reference)
			return NewSemanticInfoForSymbol(NewGroupSymbol(members...), &Error.TypeSymbol, diagnostic)
		}
	}
	if b != nil && b.scopeKind == ScopeKindPlugIn {
		name := NameReferenceSimpleName(reference)
		globals := b.Globals
		if globals == nil {
			globals = DefaultGlobalState
		}
		if plugin := globals.GetPlugIn(name); plugin != nil {
			return NewSemanticInfoForReferencedSymbol(plugin)
		}
		known := PlugIns.GetPlugIn(name)
		if known != nil {
			return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(DiagnosticPluginFunctionIsNotEnabled(name), reference))
		}
		return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(DiagnosticPluginFunctionNotDefined(name), reference))
	}
	if b != nil && b.scopeKind == ScopeKindAggregate {
		globals := b.Globals
		if globals == nil {
			globals = DefaultGlobalState
		}
		if aggregate := globals.GetAggregate(NameReferenceSimpleName(reference)); aggregate != nil {
			return NewSemanticInfoForReferencedSymbol(aggregate)
		}
		if plugin := globals.GetPlugIn(NameReferenceSimpleName(reference)); plugin != nil {
			diagnostic := b.diagnosticAtSyntaxElement(DiagnosticPluginNotAllowedInThisContext(plugin.SymbolName()), reference)
			return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
		}
		info := b.locateNameReferenceDiagnostics(
			b.ResolveNameInRowScope(b.rowScope, NameReferenceSimpleName(reference), SymbolMatchFunction|SymbolMatchView, false),
			reference,
		)
		return b.enforceFunctionAvailability(info, reference)
	}
	if call, ok := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); ok && SyntaxHasMissingChildren(call) {
		if b.LocalScope != nil {
			var locals []Symbol
			b.LocalScope.GetSymbols(NameReferenceSimpleName(reference), SymbolMatchLocal, &locals)
			for _, local := range locals {
				if _, pattern := GetResultType(local).(*PatternSymbol); pattern {
					return NewSemanticInfoForReferencedSymbol(local)
				}
			}
		}
		info := b.locateNameReferenceDiagnostics(
			b.ResolveNameInRowScope(b.rowScope, NameReferenceSimpleName(reference), SymbolMatchFunction|SymbolMatchView, false),
			reference,
		)
		if _, unresolved := info.ReferencedSymbol().(*ErrorSymbol); unresolved {
			globals := b.Globals
			if globals == nil {
				globals = DefaultGlobalState
			}
			if aggregate := globals.GetAggregate(NameReferenceSimpleName(reference)); aggregate != nil {
				diagnostic := b.diagnosticAtSyntaxElement(DiagnosticAggregateNotAllowedInThisContext(aggregate.SymbolName()), reference)
				return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
			}
			if plugin := globals.GetPlugIn(NameReferenceSimpleName(reference)); plugin != nil {
				diagnostic := b.diagnosticAtSyntaxElement(DiagnosticPluginNotAllowedInThisContext(plugin.SymbolName()), reference)
				return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
			}
		}
		return b.enforceFunctionAvailability(info, reference)
	}
	info := b.ResolveNameReferenceInRowScope(b.rowScope, reference, false)
	if _, unresolved := info.ReferencedSymbol().(*ErrorSymbol); unresolved {
		globals := b.Globals
		if globals == nil {
			globals = DefaultGlobalState
		}
		if aggregate := globals.GetAggregate(NameReferenceSimpleName(reference)); aggregate != nil {
			diagnostic := b.diagnosticAtSyntaxElement(DiagnosticAggregateNotAllowedInThisContext(aggregate.SymbolName()), reference)
			return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
		}
		if plugin := globals.GetPlugIn(NameReferenceSimpleName(reference)); plugin != nil {
			diagnostic := b.diagnosticAtSyntaxElement(DiagnosticPluginNotAllowedInThisContext(plugin.SymbolName()), reference)
			return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
		}
		name := NameReferenceSimpleName(reference)
		if syntaxElementInTabularContext(reference) {
			table := NewTableSymbol("").WithIsOpen(true)
			diagnostic := b.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownFunction(name), reference)
			return NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostic)
		}
		if len(name) >= 2 && name[:2] == "__" {
			return NewSemanticInfoForType(ScalarTypes.Unknown.Type())
		}
		return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownFunction(name), reference))
	}
	return b.enforceFunctionAvailability(info, reference)
}

func (b *Binder) enforceFunctionAvailability(info SemanticInfo, location SyntaxElement) SemanticInfo {
	function, ok := info.ReferencedSymbol().(*FunctionSymbol)
	if !ok {
		return info
	}
	if function.IsAvailable(customAvailabilityContextAt(location)) {
		return info
	}
	name := function.SymbolName()
	return NewSemanticInfoForReferencedSymbol(Error, b.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownFunction(name), location))
}

func customAvailabilityContextAt(location SyntaxElement) CustomAvailabilityContext {
	context := CustomAvailabilityContext{Location: location}
	if location == nil {
		return context
	}
	if graphMatch := GetFirstAncestor[*GraphMatchOperator](location); graphMatch != nil {
		context.InGraphMatchOrShortestPaths = true
		context.InGraphWhereOrProjectClause = syntaxElementWithin(location, graphMatch.WhereClause()) || syntaxElementWithin(location, graphMatch.ProjectClause())
	}
	if shortest := GetFirstAncestor[*GraphShortestPathsOperator](location); shortest != nil {
		context.InGraphMatchOrShortestPaths = true
		context.InGraphWhereOrProjectClause = syntaxElementWithin(location, shortest.WhereClause()) || syntaxElementWithin(location, shortest.ProjectClause())
	}
	if context.InGraphMatchOrShortestPaths {
		for call := GetFirstAncestor[*FunctionCallExpression](location); call != nil; call = GetFirstAncestor[*FunctionCallExpression](call) {
			name := NameReferenceSimpleName(call.Name())
			if name != Functions.GraphAll.SymbolName() && name != Functions.Any.SymbolName() && name != Functions.Map.SymbolName() {
				continue
			}
			if list := call.ArgumentList(); list != nil && list.Expressions() != nil && list.Expressions().Count() > 0 {
				first := list.Expressions().At(0)
				context.InGraphAllAnyMapFirstArgument = first != nil && syntaxElementWithin(location, first.Element())
			}
			break
		}
	}
	return context
}

func syntaxElementWithin(location, container SyntaxElement) bool {
	return location != nil && container != nil && (location == container || SyntaxIsAncestorOf(container, location))
}

func databaseSourceFunction(name string) *FunctionSymbol {
	for _, function := range []*FunctionSymbol{
		Functions.Table,
		Functions.ExternalTable,
		Functions.MaterializedView,
		Functions.EntityGroup,
		Functions.StoredQueryResult,
		Functions.Graph,
	} {
		if function != nil && name == function.SymbolName() {
			return function
		}
	}
	return nil
}

func (b *Binder) ResolveNameReferenceInRowScope(rowScope *TableSymbol, reference *NameReference, inTabularContext bool) SemanticInfo {
	return b.ResolveNameReferenceInRowScopes(rowScope, nil, false, reference, inTabularContext)
}

func (b *Binder) resolveAggregateNameReference(rowScope *TableSymbol, reference *NameReference) SemanticInfo {
	name := NameReferenceSimpleName(reference)
	if name == "" {
		return NewSemanticInfoForReferencedSymbol(Error)
	}
	if call, ok := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); ok && call.Name() == reference {
		globals := DefaultGlobalState
		if b != nil && b.Globals != nil {
			globals = b.Globals
		}
		if aggregate := globals.GetAggregate(name); aggregate != nil {
			diagnostic := DiagnosticAggregateNotAllowedInThisContext(aggregate.SymbolName())
			if b != nil {
				diagnostic = b.diagnosticAtSyntaxElement(diagnostic, reference)
			}
			return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
		}
	}
	// Ordinary local/database names are considered before aggregate built-ins
	// for a bare reference. Function-call names use aggregate-first lookup.
	info := b.ResolveNameReferenceInRowScope(nil, reference, false)
	if _, unresolved := info.ReferencedSymbol().(*ErrorSymbol); !unresolved || len(info.Diagnostics) == 0 {
		return info
	}

	// Aggregate scope deliberately excludes row columns. Resolve once in normal
	// scope only to distinguish an invalid aggregate-context name (KS236) from
	// a genuinely unknown name.
	if symbols := b.matchingNameSymbolsInRowScope(rowScope, name, SymbolMatchColumn); len(symbols) > 0 {
		diagnostic := DiagnosticInvalidNameInAggregateContext(name)
		if b != nil {
			diagnostic = b.diagnosticAtSyntaxElement(diagnostic, reference)
		}
		return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
	}
	return info
}

func (b *Binder) resolvePlugInNameReference(rowScope *TableSymbol, reference *NameReference) SemanticInfo {
	name := NameReferenceSimpleName(reference)
	if name == "" {
		return NewSemanticInfoForReferencedSymbol(Error)
	}
	globals := DefaultGlobalState
	if b != nil && b.Globals != nil {
		globals = b.Globals
	}
	if plugin := globals.GetPlugIn(name); plugin != nil {
		return NewSemanticInfoForReferencedSymbol(plugin)
	}

	// Plug-in scope excludes ordinary row/local/global names. Resolve once in
	// normal scope to distinguish an invalid plug-in-context name from an
	// actually unknown name, matching Binder_Names.cs.
	info := b.ResolveNameReferenceInRowScope(rowScope, reference, false)
	if _, unresolved := info.ReferencedSymbol().(*ErrorSymbol); !unresolved || len(info.Diagnostics) == 0 {
		diagnostic := DiagnosticInvalidNameInPlugInContext(name)
		if b != nil {
			diagnostic = b.diagnosticAtSyntaxElement(diagnostic, reference)
		}
		return NewSemanticInfoForReferencedSymbol(Error, diagnostic)
	}
	return info
}

func (b *Binder) ResolveNameReferenceInRowScopes(rowScope, rightRowScope *TableSymbol, commonColumnsOnly bool, reference *NameReference, inTabularContext bool) SemanticInfo {
	if reference == nil {
		return b.ResolveNameInRowScopes(rowScope, rightRowScope, commonColumnsOnly, "", SymbolMatchNone, inTabularContext)
	}
	if _, ok := reference.Name().(*BracedName); ok {
		return b.locateNameReferenceDiagnostics(b.ResolveClientParameterReference(NameReferenceSimpleName(reference), inTabularContext), reference)
	}
	var resolved SemanticInfo
	if b != nil && b.pathScope != nil {
		resolved = b.resolveNameInPathScope(
			NameReferenceSimpleName(reference),
			b.normalizeNameMatch(reference.Match()),
			nameReferenceIsCommandButNotQueryPart(reference),
		)
	} else {
		resolved = b.ResolveNameInRowScopes(rowScope, rightRowScope, commonColumnsOnly, NameReferenceSimpleName(reference), reference.Match(), inTabularContext)
	}
	if nameReferenceIsCommandButNotQueryPart(reference) &&
		matchRequestsOnly(b.normalizeNameMatch(reference.Match()), SymbolMatchDatabase) {
		if _, unresolved := resolved.ReferencedSymbol().(*ErrorSymbol); unresolved && semanticDiagnosticsContainCode(resolved.Diagnostics, "KS208") {
			resolved = NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownItem(NameReferenceSimpleName(reference), ""))
		}
	}
	if nameReferenceIsCommandButNotQueryPart(reference) &&
		matchRequestsOnly(b.normalizeNameMatch(reference.Match()), SymbolMatchFunction|SymbolMatchView) {
		if _, unresolved := resolved.ReferencedSymbol().(*ErrorSymbol); unresolved && semanticDiagnosticsContainCode(resolved.Diagnostics, "KS211") {
			resolved = NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownItem(NameReferenceSimpleName(reference), ""))
		}
	}
	info := b.locateNameReferenceDiagnostics(resolved, reference)
	return b.bindBareFunctionReference(reference, info)
}

func (b *Binder) bindBareFunctionReference(reference *NameReference, info SemanticInfo) SemanticInfo {
	fn, ok := info.ReferencedSymbol().(*FunctionSymbol)
	if !ok || reference == nil {
		return info
	}
	if _, isCallName := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); isCallName {
		return info
	}
	if !fn.IsAvailable(customAvailabilityContextAt(reference)) {
		return info
	}
	globals := DefaultGlobalState
	if b != nil && b.Globals != nil {
		globals = b.Globals
	}
	allowZeroArgumentInvocation := !globals.IsBuiltInFunction(fn)
	if fn.MinArgumentCount() == 0 && allowZeroArgumentInvocation {
		return NewSemanticInfoForSymbol(fn, b.commonBareFunctionReturnType(fn))
	}
	diagnostic := DiagnosticFunctionRequiresArgumentList(fn.SymbolName())
	if b != nil {
		diagnostic = b.diagnosticAtSyntaxElement(diagnostic, reference)
	}
	return NewSemanticInfoForSymbol(fn, b.commonBareFunctionReturnType(fn), diagnostic)
}

func (b *Binder) commonBareFunctionReturnType(fn *FunctionSymbol) *TypeSymbol {
	if fn == nil {
		return &Error.TypeSymbol
	}
	globals := DefaultGlobalState
	if b != nil && b.Globals != nil {
		globals = b.Globals
	}
	return GetCommonSignatureReturnType(fn.Signatures(), nil, func(signature *Signature) *TypeSymbol {
		if signature != nil && signature.ReturnKind == ReturnTypeCustom && signature.CustomReturnType != nil {
			result := signature.CustomReturnType(CustomReturnTypeContext{
				Location:      nil,
				Arguments:     []Expression{},
				ArgumentTypes: []*TypeSymbol{},
				ArgumentParameters: signature.ArgumentParametersForExpressions(
					[]Expression{}, nil, globals,
				),
				RowScope:  b.rowScope,
				Signature: signature,
				Globals:   globals,
			})
			return returnTypeFromSymbol(result)
		}
		if signature != nil && signature.ReturnKind == ReturnTypeComputed {
			if result, _ := newSimpleNodeBinder(b).bindComputedFunction(signature, nil, nil, nil); result != nil {
				return returnTypeFromSymbol(result)
			}
		}
		return signature.GetReturnTypeForArguments(globals, nil, nil)
	})
}

func (b *Binder) locateNameReferenceDiagnostics(info SemanticInfo, reference *NameReference) SemanticInfo {
	if b == nil || reference == nil || len(info.Diagnostics) == 0 {
		return info
	}
	diagnostics := append([]Diagnostic{}, info.Diagnostics...)
	for i, diagnostic := range diagnostics {
		if !diagnostic.HasLocation() {
			diagnostics[i] = b.diagnosticAtSyntaxElement(diagnostic, reference)
		}
	}
	info.Diagnostics = diagnostics
	return info
}

func (b *Binder) ResolveName(name string, match SymbolMatch, inTabularContext bool) SemanticInfo {
	return b.ResolveNameInRowScope(nil, name, match, inTabularContext)
}

func (b *Binder) ResolveNameInRowScope(rowScope *TableSymbol, name string, match SymbolMatch, inTabularContext bool) SemanticInfo {
	return b.ResolveNameInRowScopes(rowScope, nil, false, name, match, inTabularContext)
}

func (b *Binder) ResolveNameInRowScopes(rowScope, rightRowScope *TableSymbol, commonColumnsOnly bool, name string, match SymbolMatch, inTabularContext bool) SemanticInfo {
	match = b.normalizeNameMatch(match)
	if name == "" {
		return NewSemanticInfoForReferencedSymbol(Error)
	}
	if b != nil && b.pathScope != nil {
		return b.resolveNameInPathScope(name, match, false)
	}
	if name == "$left" && rowScope != nil && rightRowScope != nil {
		tuple := b.getTuple(rowScope)
		return NewSemanticInfoForSymbol(tuple, &tuple.TypeSymbol)
	}
	if name == "$right" && rightRowScope != nil {
		tuple := b.getTuple(rightRowScope)
		return NewSemanticInfoForSymbol(tuple, &tuple.TypeSymbol)
	}
	if symbols := b.matchingNameSymbolsInRowScopes(rowScope, rightRowScope, commonColumnsOnly, name, match); len(symbols) > 0 {
		if len(symbols) == 1 {
			return NewSemanticInfoForReferencedSymbol(symbols[0])
		}
		return NewSemanticInfoForSymbol(NewGroupSymbol(symbols...), &Error.TypeSymbol, DiagnosticNameRefersToMoreThanOneItem(name))
	}
	if b != nil && b.isFuzzy && (inTabularContext || matchRequestsTabularName(match)) {
		table := NewTableSymbol(name).WithIsOpen(true)
		return NewSemanticInfoForSymbol(table, &table.TypeSymbol, DiagnosticFuzzyTableNotDefined(name))
	}
	return unknownNameSemanticInfo(name, match, inTabularContext)
}

func (b *Binder) getTuple(table *TableSymbol) *TupleSymbol {
	if table == nil {
		return NewTupleSymbol(nil)
	}
	if b == nil {
		return NewTupleSymbol(table.Columns(), table)
	}
	if b.tupleMap == nil {
		b.tupleMap = map[*TableSymbol]*TupleSymbol{}
	}
	if tuple, ok := b.tupleMap[table]; ok {
		return tuple
	}
	tuple := NewTupleSymbol(table.Columns(), table)
	b.tupleMap[table] = tuple
	return tuple
}

func normalizeNameMatch(match SymbolMatch) SymbolMatch {
	if match == SymbolMatchNone {
		return SymbolMatchDefault
	}
	return match
}

func (b *Binder) normalizeNameMatch(match SymbolMatch) SymbolMatch {
	return normalizeNameMatch(match)
}

func (b *Binder) matchingNameSymbolsInRowScope(rowScope *TableSymbol, name string, match SymbolMatch) []Symbol {
	return b.matchingNameSymbolsInRowScopes(rowScope, nil, false, name, match)
}

func (b *Binder) matchingNameSymbolsInRowScopes(rowScope, rightRowScope *TableSymbol, commonColumnsOnly bool, name string, match SymbolMatch) []Symbol {
	if name == "" || match == SymbolMatchNone {
		return nil
	}
	if rowScope != nil && match.Has(SymbolMatchColumn) {
		if symbols := FindMembers(rowScope, name, match, false); len(symbols) > 0 {
			if len(symbols) == 1 && rightRowScope != nil && commonColumnsOnly {
				if common := commonColumnSymbol(symbols[0], firstRowScopeColumn(rightRowScope, name, match)); common != nil {
					return []Symbol{common}
				}
			}
			return distinctSymbols(symbols)
		}
	}
	if rightRowScope != nil && match.Has(SymbolMatchColumn) {
		if symbols := FindMembers(rightRowScope, name, match, false); len(symbols) > 0 {
			return distinctSymbols(symbols)
		}
	}
	if symbols := b.matchingNameSymbols(name, match); len(symbols) > 0 {
		return symbols
	}
	if rowScope != nil && rowScope.IsOpen() && match.Has(SymbolMatchColumn) {
		return []Symbol{b.getOpenTableInferredColumn(name, rowScope)}
	}
	return nil
}

func firstRowScopeColumn(rowScope *TableSymbol, name string, match SymbolMatch) Symbol {
	if rowScope == nil || !match.Has(SymbolMatchColumn) {
		return nil
	}
	symbols := FindMembers(rowScope, name, match, false)
	if len(symbols) != 1 {
		return nil
	}
	return symbols[0]
}

func commonColumnSymbol(left Symbol, right Symbol) *ColumnSymbol {
	leftColumn, leftOK := left.(*ColumnSymbol)
	rightColumn, rightOK := right.(*ColumnSymbol)
	if !leftOK || !rightOK {
		return nil
	}
	return NewColumnSymbolFull(leftColumn.SymbolName(), leftColumn.Type, leftColumn.Description, []*ColumnSymbol{leftColumn, rightColumn}, leftColumn.Source, leftColumn.Examples)
}

func (b *Binder) matchingNameSymbols(name string, match SymbolMatch) []Symbol {
	if name == "" || match == SymbolMatchNone {
		return nil
	}
	globals := DefaultGlobalState
	if b != nil && b.Globals != nil {
		globals = b.Globals
	}

	if b != nil && b.LocalScope != nil {
		var local []Symbol
		b.LocalScope.GetSymbols(name, match, &local)
		if len(local) > 0 {
			return distinctSymbols(local)
		}
	}

	var symbols []Symbol
	if ambient := globals.GetAmbientSymbol(name); SymbolMatchesName(ambient, name, match, false) {
		symbols = append(symbols, ambient)
	}
	if globals.Database != nil {
		symbols = append(symbols, FindMembers(globals.Database, name, match, false)...)
	}
	// Database-qualified command table names can resolve a database in the
	// current cluster when no member of the current database matches.
	if len(symbols) == 0 && globals.Cluster != nil && match.Has(SymbolMatchDatabase) {
		symbols = append(symbols, FindMembers(globals.Cluster, name, SymbolMatchDatabase, false)...)
	}
	functionGroups := [][]*FunctionSymbol{globals.Functions}
	if b != nil {
		switch b.scopeKind {
		case ScopeKindAggregate:
			functionGroups = [][]*FunctionSymbol{globals.Aggregates, globals.Functions}
		case ScopeKindPlugIn:
			functionGroups = [][]*FunctionSymbol{globals.PlugIns}
		case ScopeKindOption:
			functionGroups = nil
		}
	}
	for _, group := range functionGroups {
		matched := false
		for _, fn := range group {
			if SymbolMatchesName(fn, name, match, false) {
				symbols = append(symbols, fn)
				matched = true
			}
		}
		if matched {
			break
		}
	}
	for _, option := range globals.Options {
		if SymbolMatchesName(option, name, match, false) {
			symbols = append(symbols, option)
		}
	}
	return distinctSymbols(symbols)
}

func distinctSymbols(symbols []Symbol) []Symbol {
	if len(symbols) < 2 {
		return symbols
	}
	seen := map[Symbol]struct{}{}
	distinct := make([]Symbol, 0, len(symbols))
	for _, symbol := range symbols {
		if symbol == nil {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		distinct = append(distinct, symbol)
	}
	return distinct
}

func unknownNameSemanticInfo(name string, match SymbolMatch, inTabularContext bool) SemanticInfo {
	switch {
	case matchRequestsOnly(match, SymbolMatchFunction|SymbolMatchView):
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownFunction(name))
	case matchRequestsOnly(match, SymbolMatchDatabase):
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownDatabase(name))
	case matchRequestsOnly(match, SymbolMatchExternalTable):
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownExternalTable(name))
	case matchRequestsOnly(match, SymbolMatchMaterializedView):
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownMaterializedView(name))
	case matchRequestsOnly(match, SymbolMatchStoredQueryResult):
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownStoredQueryResult(name))
	case matchRequestsOnly(match, SymbolMatchGraphModel):
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownGraphModel(name))
	case inTabularContext || matchRequestsTabularName(match):
		table := NewTableSymbol(name).WithIsOpen(true)
		return NewSemanticInfoForSymbol(table, &table.TypeSymbol, DiagnosticNameDoesNotReferToAnyKnownTableOrFunction(name))
	default:
		return NewSemanticInfoForReferencedSymbol(Error, DiagnosticNameDoesNotReferToAnyKnownItem(name, ""))
	}
}

func matchRequestsOnly(match SymbolMatch, allowed SymbolMatch) bool {
	return match != SymbolMatchNone && match&^allowed == 0 && match&allowed != 0
}

func matchRequestsTabularName(match SymbolMatch) bool {
	if match == SymbolMatchNone {
		return false
	}
	if match.Has(SymbolMatchColumn | SymbolMatchScalar | SymbolMatchLocal | SymbolMatchDatabase | SymbolMatchCluster | SymbolMatchOption | SymbolMatchGraph | SymbolMatchGraphModel | SymbolMatchGraphSnapshot | SymbolMatchEntityGroup | SymbolMatchEntityGroupElement) {
		return false
	}
	return match.Has(SymbolMatchTable | SymbolMatchExternalTable | SymbolMatchMaterializedView | SymbolMatchStoredQueryResult | SymbolMatchTabular | SymbolMatchNonScalar | SymbolMatchFunction | SymbolMatchView)
}
