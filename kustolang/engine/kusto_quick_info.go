package kustolang

import (
	"fmt"
	"strings"
)

// Ported from src/Kusto.Language/Editor/Kusto/KustoQuickInfoBuilder.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type kustoQuickInfoBuilder struct {
	service *KustoCodeService
	code    *KustoCode
	options QuickInfoOptions
}

func newKustoQuickInfoBuilder(service *KustoCodeService, code *KustoCode, options QuickInfoOptions) *kustoQuickInfoBuilder {
	return &kustoQuickInfoBuilder{service: service, code: code, options: options}
}

func (b *kustoQuickInfoBuilder) getQuickInfo(position int, cancellationToken CancellationToken) QuickInfo {
	if b == nil || b.code == nil || !b.code.HasSemantics() {
		return QuickInfoEmpty
	}
	var items []QuickInfoItem
	if symbol, ok := b.getSymbolInfo(position); ok {
		items = append(items, symbol)
	}
	if b.options.ShowDiagnostics {
		if diagnostic, ok := b.getDiagnosticInfo(position, cancellationToken); ok {
			items = append(items, diagnostic)
		}
	}
	return NewQuickInfoItems(items)
}

func (b *kustoQuickInfoBuilder) getSymbolInfo(position int) (QuickInfoItem, bool) {
	ref, ok := quickInfoTokenRefWithAffinity(b.code, position)
	if !ok || position < ref.positionedToken().TextStart() {
		return QuickInfoItem{}, false
	}
	for _, ancestor := range ref.ancestors {
		bracketed, ok := WrapGeneratedSyntaxElement(ancestor.element).(*BracketedExpression)
		if ok && bracketed.CloseBracket() != nil && bracketed.CloseBracket().Width() == 0 {
			return QuickInfoItem{}, false
		}
	}

	for i := len(ref.ancestors) - 1; i >= 0; i-- {
		wrapped := WrapGeneratedSyntaxElement(ref.ancestors[i].element)
		expression, isExpression := wrapped.(Expression)
		if !isExpression {
			continue
		}
		expressionAncestorIndex := i

		// A bracketed name's string literal is syntax, not the value being
		// referenced. Promote to the enclosing expression so ['column name']
		// displays column QuickInfo.
		if i >= 2 {
			if _, isBracketedName := WrapGeneratedSyntaxElement(ref.ancestors[i-1].element).(*BracketedName); isBracketedName {
				if enclosing, ok := WrapGeneratedSyntaxElement(ref.ancestors[i-2].element).(Expression); ok {
					expression = enclosing
					expressionAncestorIndex = i - 2
				}
			}
		}

		// Function names display the selected call signature, not just the
		// FunctionSymbol attached to the name reference.
		if expressionAncestorIndex > 0 {
			if call, isCall := WrapGeneratedSyntaxElement(ref.ancestors[expressionAncestorIndex-1].element).(*FunctionCallExpression); isCall {
				expression = call
			}
		}

		if ExpressionIsLiteral(expression) {
			if typ := quickInfoConcreteType(ExpressionResultType(expression)); typ != nil && IsScalar(typ) {
				return b.symbolInfo(typ, nil, nil, nil, QuickInfoKindLiteral)
			}
		}

		symbol := expressionReferencedSymbol(expression)
		if symbol == nil {
			continue
		}
		return b.symbolInfo(symbol, expressionReferencedSignature(expression), quickInfoExpressionResultSymbol(expression), ExpressionLiteralValueInfo(expression), -1)
	}

	for i := len(ref.ancestors) - 1; i >= 0; i-- {
		node := simpleBinderSyntaxNode(WrapGeneratedSyntaxElement(ref.ancestors[i].element))
		if node != nil && node.ReferencedSymbol() != nil {
			return b.symbolInfo(node.ReferencedSymbol(), node.ReferencedSignature(), nil, nil, -1)
		}
	}
	return QuickInfoItem{}, false
}

func quickInfoTokenRefWithAffinity(code *KustoCode, position int) (kustoServiceTokenRef, bool) {
	refs, index, ok := kustoServiceTokenRefAtPosition(code, position)
	if !ok {
		return kustoServiceTokenRef{}, false
	}
	if HasAffinity(refs[index].positionedToken(), position) {
		return refs[index], true
	}
	if index > 0 && HasAffinity(refs[index-1].positionedToken(), position) {
		return refs[index-1], true
	}
	return kustoServiceTokenRef{}, false
}

func quickInfoExpressionResultSymbol(expression Expression) Symbol {
	if call, ok := expression.(*FunctionCallExpression); ok {
		if result, found := call.SyntaxNode.Payload(functionCallResultSymbolPayload); found {
			if symbol, ok := result.(Symbol); ok {
				return symbol
			}
		}
	}
	if symbol := expressionReferencedSymbol(expression); symbol != nil {
		switch symbol := symbol.(type) {
		case *TableSymbol, *MaterializedViewSymbol, *ExternalTableSymbol, *StoredQueryResultSymbol,
			*DatabaseSymbol, *ClusterSymbol, *GraphSymbol, *EntityGroupElementSymbol:
			return symbol
		case *ColumnSymbol:
			return quickInfoConcreteType(symbol.Type)
		case *VariableSymbol:
			return GetResultType(symbol.Type)
		case *ParameterSymbol:
			return GetResultType(symbol.Type)
		}
	}
	return quickInfoConcreteType(ExpressionResultType(expression))
}

func quickInfoConcreteType(typ *TypeSymbol) Symbol {
	if typ == nil {
		return nil
	}
	if typ.dynamicPrimitive != nil {
		return typ.dynamicPrimitive
	}
	if typ.dynamicBag != nil {
		return typ.dynamicBag
	}
	if typ.dynamicArray != nil {
		return typ.dynamicArray
	}
	if typ.tuple != nil {
		return typ.tuple
	}
	return typ
}

func (b *kustoQuickInfoBuilder) symbolInfo(symbol Symbol, signature *Signature, result Symbol, value *ValueInfo, forcedKind QuickInfoKind) (QuickInfoItem, bool) {
	if symbol == nil {
		return QuickInfoItem{}, false
	}
	parts := make([]ClassifiedText, 0, 8)
	b.appendSymbolDisplay(symbol, signature, result, value, &parts)
	kind := forcedKind
	if kind < 0 {
		kind = b.itemKind(symbol)
	}
	return NewQuickInfoItem(kind, parts...), true
}

func (b *kustoQuickInfoBuilder) itemKind(symbol Symbol) QuickInfoKind {
	switch symbol := symbol.(type) {
	case *ClusterSymbol:
		return QuickInfoKindCluster
	case *ColumnSymbol:
		return QuickInfoKindColumn
	case *DatabaseSymbol:
		return QuickInfoKindDatabase
	case *FunctionSymbol:
		if b.code.Globals.IsBuiltInFunction(symbol) {
			return QuickInfoKindBuiltInFunction
		}
		if b.code.Globals.IsDatabaseFunction(symbol) {
			return QuickInfoKindDatabaseFunction
		}
		return QuickInfoKindLocalFunction
	case *OperatorSymbol:
		return QuickInfoKindOperator
	case *ParameterSymbol:
		return QuickInfoKindParameter
	case *PatternSymbol:
		return QuickInfoKindPattern
	case *TableSymbol, *MaterializedViewSymbol, *ExternalTableSymbol, *StoredQueryResultSymbol:
		return QuickInfoKindTable
	case *VariableSymbol:
		switch symbol.Type.(type) {
		case *FunctionSymbol, *PatternSymbol,
			*TableSymbol, *MaterializedViewSymbol, *ExternalTableSymbol, *StoredQueryResultSymbol,
			*GraphSymbol,
			*ScalarSymbol, *DynamicAnySymbol, *DynamicPrimitiveSymbol, *DynamicArraySymbol, *DynamicBagSymbol, *TupleSymbol, *TypeSymbol:
			return b.itemKind(symbol.Type)
		default:
			return QuickInfoKindVariable
		}
	case *ScalarSymbol, *DynamicAnySymbol, *DynamicPrimitiveSymbol, *DynamicArraySymbol, *DynamicBagSymbol, *TupleSymbol, *TypeSymbol:
		return QuickInfoKindType
	case *CommandSymbol:
		return QuickInfoKindCommand
	case *OptionSymbol:
		return QuickInfoKindOption
	case *GraphSymbol:
		return QuickInfoKindGraph
	case *EntityGroupElementSymbol:
		return b.itemKind(symbol.UnderlyingSymbol)
	case *GroupSymbol:
		members := symbol.Members()
		if len(members) > 0 {
			return b.itemKind(members[0])
		}
	}
	return QuickInfoKindText
}

func (b *kustoQuickInfoBuilder) appendSymbolDisplay(symbol Symbol, signature *Signature, result Symbol, value *ValueInfo, parts *[]ClassifiedText) {
	if group, ok := symbol.(*GroupSymbol); ok {
		members := group.Members()
		limit := minInt(len(members), 5)
		for i := 0; i < limit; i++ {
			if i > 0 {
				*parts = append(*parts, NewClassifiedText("\n"))
			}
			b.appendSymbolDisplay(members[i], nil, GetResultType(members[i]), nil, parts)
		}
		if limit < len(members) {
			*parts = append(*parts, NewClassifiedText(fmt.Sprintf("\n+ (%d) additional", len(members)-limit)))
		}
		return
	}

	if scalar, ok := symbol.(*ScalarSymbol); ok {
		*parts = append(*parts, NewClassifiedTextKind(ClassificationKindIdentifier, scalar.SymbolName()))
		return
	}
	if variable, ok := symbol.(*VariableSymbol); ok {
		switch variable.Type.(type) {
		case *FunctionSymbol, *PatternSymbol:
			b.appendSymbolDisplay(variable.Type, nil, result, value, parts)
			return
		}
	}

	*parts = append(*parts, NewClassifiedTextKind(quickInfoNameClassification(symbol), symbol.SymbolName()))
	if function, ok := symbol.(*FunctionSymbol); ok {
		if signature == nil {
			signatures := function.Signatures()
			if len(signatures) > 0 {
				signature = signatures[0]
			}
		}
		if signature != nil {
			appendQuickInfoSignature(signature, parts)
		}
	}
	if result == nil {
		if variable, ok := symbol.(*VariableSymbol); ok {
			result = variable.Type
		}
	}
	if result != nil && !quickInfoUnknownOrError(result) {
		*parts = append(*parts,
			NewClassifiedTextKind(ClassificationKindPunctuation, ": "),
			NewClassifiedTextKind(ClassificationKindType, quickInfoTypeText(result)))
	}
	if value != nil && value.RawText != "" {
		text := value.RawText
		if utf16TextWidth(text) > 40 {
			text = kustoOutlinerUTF16Prefix(text, 40) + " ..."
		}
		kind := ClassificationKindLiteral
		if _, ok := value.Value.(string); ok {
			kind = ClassificationKindStringLiteral
		}
		*parts = append(*parts,
			NewClassifiedTextKind(ClassificationKindPunctuation, " = "),
			NewClassifiedTextKind(kind, text))
	}
	if b.options.ShowDescriptions {
		if description := quickInfoDescription(symbol); description != "" {
			*parts = append(*parts, NewClassifiedTextKind(ClassificationKindComment, "\n\n"+description))
		}
	}
}

func appendQuickInfoSignature(signature *Signature, parts *[]ClassifiedText) {
	*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, "("))
	for i, parameter := range signature.Parameters {
		if i > 0 {
			*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, ", "))
		}
		if parameter.IsOptional() {
			*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, "["))
		}
		*parts = append(*parts, NewClassifiedTextKind(ClassificationKindSignatureParameter, parameter.Name))
		if parameter.IsOptional() {
			*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, "]"))
		}
	}
	if signature.HasRepeatableParameters {
		if len(signature.Parameters) > 0 {
			*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, ", "))
		}
		*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, "..."))
	}
	*parts = append(*parts, NewClassifiedTextKind(ClassificationKindPunctuation, ")"))
}

func quickInfoNameClassification(symbol Symbol) ClassificationKind {
	switch symbol := symbol.(type) {
	case *ColumnSymbol:
		return ClassificationKindColumn
	case *DatabaseSymbol:
		return ClassificationKindDatabase
	case *FunctionSymbol, *PatternSymbol:
		return ClassificationKindFunction
	case *OperatorSymbol:
		return ClassificationKindMathOperator
	case *ParameterSymbol:
		return ClassificationKindParameter
	case *TableSymbol, *ExternalTableSymbol, *StoredQueryResultSymbol:
		return ClassificationKindTable
	case *MaterializedViewSymbol:
		return ClassificationKindMaterializedView
	case *VariableSymbol:
		return ClassificationKindVariable
	case *CommandSymbol:
		return ClassificationKindCommand
	case *OptionSymbol:
		return ClassificationKindOption
	case *EntityGroupElementSymbol:
		return quickInfoNameClassification(symbol.UnderlyingSymbol)
	case *ScalarSymbol:
		if symbol.SymbolName() == "string" {
			return ClassificationKindStringLiteral
		}
		return ClassificationKindLiteral
	}
	return ClassificationKindPlainText
}

func quickInfoDescription(symbol Symbol) string {
	switch symbol := symbol.(type) {
	case *TableSymbol:
		return symbol.Description()
	case *ExternalTableSymbol:
		return symbol.Description()
	case *MaterializedViewSymbol:
		return symbol.Description()
	case *StoredQueryResultSymbol:
		return symbol.Description()
	case *ColumnSymbol:
		return symbol.Description
	case *ParameterSymbol:
		return symbol.Description()
	case *FunctionSymbol:
		return symbol.Description()
	case *OptionSymbol:
		return symbol.Description()
	case *EntityGroupSymbol:
		return "Entity Group"
	case *EntityGroupElementSymbol:
		return quickInfoDescription(symbol.UnderlyingSymbol)
	}
	return ""
}

func quickInfoUnknownOrError(symbol Symbol) bool {
	if symbol == nil {
		return true
	}
	if symbol.SymbolName() == ScalarTypes.Unknown.SymbolName() {
		return true
	}
	_, isError := symbol.(*ErrorSymbol)
	return isError || isErrorTypeSymbol(returnTypeFromSymbol(symbol))
}

func quickInfoTypeText(symbol Symbol) string {
	var builder strings.Builder
	appendQuickInfoTypeText(symbol, &builder, false)
	return builder.String()
}

func appendQuickInfoTypeText(symbol Symbol, builder *strings.Builder, nested bool) {
	switch symbol := symbol.(type) {
	case *TableSymbol:
		builder.WriteString("table(")
		appendQuickInfoSchema(symbol.Columns(), builder)
		builder.WriteString(")")
	case *MaterializedViewSymbol:
		appendQuickInfoTypeText(&symbol.TableSymbol, builder, nested)
	case *ExternalTableSymbol:
		appendQuickInfoTypeText(&symbol.TableSymbol, builder, nested)
	case *StoredQueryResultSymbol:
		appendQuickInfoTypeText(&symbol.TableSymbol, builder, nested)
	case *TupleSymbol:
		builder.WriteString("tuple(")
		appendQuickInfoSchema(symbol.Columns(), builder)
		builder.WriteString(")")
	case *ClusterSymbol:
		builder.WriteString("cluster(")
		builder.WriteString(GetSingleQuotedStringLiteral(symbol.SymbolName()))
		builder.WriteString(")")
	case *DatabaseSymbol:
		builder.WriteString("database(")
		builder.WriteString(GetSingleQuotedStringLiteral(symbol.SymbolName()))
		builder.WriteString(")")
	case *GraphSymbol:
		builder.WriteString("graph(edge(")
		if symbol.EdgeShape != nil {
			appendQuickInfoSchema(symbol.EdgeShape.Columns(), builder)
		}
		builder.WriteString(")")
		if symbol.NodeShape != nil {
			builder.WriteString(", node(")
			appendQuickInfoSchema(symbol.NodeShape.Columns(), builder)
			builder.WriteString(")")
		}
		builder.WriteString(")")
	case *EntityGroupElementSymbol:
		appendQuickInfoTypeText(symbol.UnderlyingSymbol, builder, false)
	case *DynamicBagSymbol:
		if !nested {
			builder.WriteString("dynamic(")
		}
		builder.WriteString("bag(")
		appendQuickInfoSchema(symbol.Properties(), builder)
		builder.WriteString(")")
		if !nested {
			builder.WriteString(")")
		}
	case *DynamicArraySymbol:
		if !nested {
			builder.WriteString("dynamic(")
		}
		builder.WriteString("array(")
		appendQuickInfoTypeText(quickInfoConcreteType(symbol.ElementType), builder, true)
		builder.WriteString(")")
		if !nested {
			builder.WriteString(")")
		}
	case *DynamicPrimitiveSymbol:
		if !nested {
			builder.WriteString("dynamic(")
		}
		appendQuickInfoTypeText(symbol.UnderlyingType, builder, true)
		if !nested {
			builder.WriteString(")")
		}
	case *VariableSymbol:
		appendQuickInfoTypeText(symbol.Type, builder, nested)
	default:
		if symbol != nil {
			builder.WriteString(symbol.SymbolName())
		}
	}
}

func appendQuickInfoSchema(columns []*ColumnSymbol, builder *strings.Builder) {
	limit := minInt(len(columns), 7)
	for i := 0; i < limit; i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(columns[i].SymbolName())
	}
	if limit < len(columns) {
		builder.WriteString(", ...")
	}
}

func (b *kustoQuickInfoBuilder) getDiagnosticInfo(position int, cancellationToken CancellationToken) (QuickInfoItem, bool) {
	diagnostics := append(
		b.service.getDiagnosticsBytes(false, cancellationToken),
		b.service.getAnalyzerDiagnosticsBytes(false, cancellationToken)...,
	)
	var best *Diagnostic
	for i := range diagnostics {
		diagnostic := &diagnostics[i]
		if !b.options.DiagnosticFilter.IsDiagnosticEnabled(*diagnostic) {
			continue
		}
		end := diagnostic.End()
		if end <= diagnostic.Start {
			end++
		}
		if position >= diagnostic.Start && position < end && (best == nil || diagnostic.Start > best.Start) {
			best = diagnostic
		}
	}
	if best == nil {
		return QuickInfoItem{}, false
	}
	kind := QuickInfoKindText
	switch best.Severity {
	case DiagnosticSeverityError:
		kind = QuickInfoKindError
	case DiagnosticSeverityWarning:
		kind = QuickInfoKindWarning
	case DiagnosticSeveritySuggestion:
		kind = QuickInfoKindSuggestion
	}
	return NewQuickInfoItemText(kind, best.Message), true
}
