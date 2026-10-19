package kustolang

// Ported from src/Kusto.Language/Binder/Binder_API.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// TryBindSyntaxTree binds a complete syntax tree. The distinct name replaces
// the reference implementation TryBind overload while retaining the existing Go SyntaxElement API.
func TryBindSyntaxTree(tree *SyntaxTree, globals *GlobalState, cancellationTokens ...CancellationToken) bool {
	if tree == nil || tree.Root == nil || !tree.IsSafeToRecurse(globals) {
		return false
	}
	cancellationToken := binderAPICancellationToken(cancellationTokens)
	cancellationToken.ThrowIfCancellationRequested()
	bound := NewBinder(globals).TryBindWithCancellation(tree.Root, cancellationToken)
	throwBinderAPICancellation(cancellationTokens)
	return bound
}

// TryBindCalledFunctionBody binds a function-body tree in the supplied outer
// lexical and database context.
func TryBindCalledFunctionBody(
	bodyTree *SyntaxTree,
	outer *Binder,
	currentCluster *ClusterSymbol,
	currentDatabase *DatabaseSymbol,
	currentFunction *FunctionSymbol,
	outerScope *LocalScope,
	locals []Symbol,
	cancellationTokens ...CancellationToken,
) bool {
	if bodyTree == nil || bodyTree.Root == nil || outer == nil || !bodyTree.IsSafeToRecurse(outer.Globals) {
		return false
	}
	throwBinderAPICancellation(cancellationTokens)

	globals := outer.Globals
	if globals == nil {
		globals = DefaultGlobalState
	}
	if currentCluster != nil {
		globals = globals.WithCluster(currentCluster)
	}
	if currentDatabase != nil {
		globals = globals.WithDatabase(currentDatabase)
	}
	called := NewBinder(globals)
	called.currentFunction = currentFunction
	called.LocalScope = NewLocalScope(outerScope)
	called.functionExpansions = outer.functionExpansions
	called.entityGroupExpansions = outer.entityGroupExpansions
	called.functionStaticScopes = outer.functionStaticScopes
	for _, local := range locals {
		called.LocalScope.AddSymbol(local)
	}
	cancellationToken := binderAPICancellationToken(cancellationTokens)
	if len(cancellationTokens) == 0 {
		cancellationToken = outer.cancellationToken
	}
	bound := called.TryBindWithCancellation(bodyTree.Root, cancellationToken)
	throwBinderAPICancellation(cancellationTokens)
	return bound
}

// GetComputedReturnType binds a computed signature body and returns its result
// type. Type-only arguments populate the expansion's parameter variables,
// matching Binder.GetComputedReturnType in the source API.
func GetComputedReturnType(signature *Signature, globals *GlobalState, argumentTypes []*TypeSymbol) *TypeSymbol {
	if signature == nil || signature.ReturnKind != ReturnTypeComputed {
		return ScalarTypes.Unknown.Type()
	}
	if globals == nil {
		globals = DefaultGlobalState
	}
	binder := NewBinder(globals)
	parameters := signature.ArgumentParametersForTypes(argumentTypes)
	if cache := binder.globalBindingCache; cache != nil {
		cache.expansionMu.Lock()
		defer cache.expansionMu.Unlock()
	}
	result, _ := newSimpleNodeBinder(binder).bindComputedFunction(signature, nil, argumentTypes, parameters)
	if result == nil {
		return ScalarTypes.Unknown.Type()
	}
	if typ := returnTypeFromSymbol(result); typ != nil && !isErrorTypeSymbol(typ) {
		return typ
	}
	return ScalarTypes.Unknown.Type()
}

// GetReferencedSymbol gets the symbol that a name would reference at a tree
// position. The tree is expected to have been bound, as in the upstream API.
func GetReferencedSymbol(tree *SyntaxTree, position int, name string, globals *GlobalState, match SymbolMatch, cancellationTokens ...CancellationToken) Symbol {
	throwBinderAPICancellation(cancellationTokens)
	code := binderAPICode(tree, globals)
	if code == nil {
		return nil
	}
	result := code.GetSpeculativeReferencedSymbol(position, name, match)
	throwBinderAPICancellation(cancellationTokens)
	return result
}

// GetRowScope gets the implicit table of columns available at a tree position.
func GetRowScope(tree *SyntaxTree, position int, globals *GlobalState, cancellationTokens ...CancellationToken) *TableSymbol {
	throwBinderAPICancellation(cancellationTokens)
	code := binderAPICode(tree, globals)
	if code == nil {
		return EmptyTableSymbol
	}
	row := code.GetColumnsInScope(position)
	throwBinderAPICancellation(cancellationTokens)
	if row == nil {
		return EmptyTableSymbol
	}
	return row
}

// GetSymbolsInScope gets all symbols visible at a tree position.
func GetSymbolsInScope(tree *SyntaxTree, position int, globals *GlobalState, match SymbolMatch, include IncludeFunctionKind, cancellationTokens ...CancellationToken) []Symbol {
	throwBinderAPICancellation(cancellationTokens)
	code := binderAPICode(tree, globals)
	if code == nil {
		return []Symbol{}
	}
	result := code.GetSymbolsInScope(position, match, include)
	throwBinderAPICancellation(cancellationTokens)
	return result
}

func binderAPICode(tree *SyntaxTree, globals *GlobalState) *KustoCode {
	if tree == nil || tree.Root == nil || !tree.IsSafeToRecurse(globals) {
		return nil
	}
	if globals == nil {
		globals = DefaultGlobalState
	}
	text := tree.Root.Text(IncludeTriviaAll)
	tokens := lexForGlobalState(text, globals)
	var query *QueryBlock
	if tree.Root.Kind() == SyntaxKindQueryBlock {
		query = &QueryBlock{SyntaxNode: tree.Root}
	}
	return &KustoCode{
		Text:             text,
		Kind:             CodeKindOf(text),
		Globals:          globals,
		Tree:             tree,
		Tokens:           tokens,
		TokenStarts:      tokenStarts(tokens),
		Query:            query,
		analysisState:    kustoCodeAnalysisPerformed,
		lineStartsCache:  newKustoCodeLineStartsCache(),
		diagnosticsCache: &kustoCodeDiagnosticsCache{},
	}
}

func throwBinderAPICancellation(tokens []CancellationToken) {
	for _, token := range tokens {
		token.ThrowIfCancellationRequested()
	}
}

func binderAPICancellationToken(tokens []CancellationToken) CancellationToken {
	if len(tokens) == 0 {
		return CancellationToken{}
	}
	if len(tokens) == 1 {
		return tokens[0]
	}
	return NewCancellationTokenFunc(func() bool {
		for _, token := range tokens {
			if token.Err() != nil {
				return true
			}
		}
		return false
	})
}
