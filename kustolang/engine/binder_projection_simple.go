package kustolang

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Ported from the direct project/extend paths in Binder_Projection.cs and
// Binder_NodeBinder.cs. Unsupported expression shapes fail transactionally.

type simpleProjectionBinder struct {
	binder                  *Binder
	defaultColumnNameSuffix *int
	inPipe                  *bool
}

type compoundProjectionStyle int

const (
	compoundProjectionDefault compoundProjectionStyle = iota
	compoundProjectionExtend
	compoundProjectionPrint
)

func newSimpleProjectionBinder(binder *Binder, defaultColumnNameSuffix *int, inPipe *bool) *simpleProjectionBinder {
	return &simpleProjectionBinder{binder: binder, defaultColumnNameSuffix: defaultColumnNameSuffix, inPipe: inPipe}
}

func (b *simpleProjectionBinder) bindPrint(node *PrintOperator, inPipe bool) *SemanticInfo {
	if node == nil || node.Expressions() == nil {
		return nil
	}

	builder := NewProjectionBuilder()
	var diagnostics []Diagnostic
	if inPipe {
		diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorMustBeFirst(), node.PrintKeyword()))
	}
	for i, separated := range node.Expressions().Elements() {
		expression := separated.Element()
		typ := printExpressionType(expression)
		if typ != nil && !isErrorTypeSymbol(typ) && !printExpressionIsScalar(expression) {
			diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
		}
		if typ == nil || (isErrorTypeSymbol(typ)) {
			typ = ScalarTypes.Unknown.Type()
		}
		if compound, ok := expression.(*CompoundNamedExpression); ok {
			if b.addCompoundNamedExpression(compound, builder, &diagnostics, compoundProjectionPrint) {
				continue
			}
		}

		if named, ok := expression.(*SimpleNamedExpression); ok {
			if b.addPrintNamedTupleExpression(named, builder, &diagnostics) {
				continue
			}
			column := b.newProjectionColumn(NameDeclarationSimpleName(named.Name()), typ, named.Expression())
			builder.Declare(column, &diagnostics, named.Name())
			named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
			continue
		}
		if column := projectionResultColumn(expression); column != nil {
			builder.Declare(column.WithType(typ), &diagnostics, expression)
			continue
		}

		name := "print_" + strconv.Itoa(i)
		if call, ok := expression.(*FunctionCallExpression); ok && b.addPrintFunctionTupleExpression(call, name, builder) {
			continue
		}
		column := b.newProjectionColumn(name, typ, expression)
		builder.Add(column, ProjectionAddOptions{BaseName: name, HasBaseName: true})
	}

	table := NewTableSymbol("", builder.GetProjection()...)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (b *simpleProjectionBinder) addPrintNamedTupleExpression(expression *SimpleNamedExpression, builder *ProjectionBuilder, diagnostics *[]Diagnostic) bool {
	if expression == nil || expression.Name() == nil || expression.Expression() == nil {
		return false
	}
	tuple, ok := TupleForType(ExpressionRawResultType(expression.Expression()))
	if !ok || tuple == nil {
		return false
	}
	columns := tuple.Columns()
	if len(columns) == 0 || columns[0] == nil {
		return true
	}
	first := columns[0]
	column := NewColumnSymbolFull(NameDeclarationSimpleName(expression.Name()), first.Type, "", []*ColumnSymbol{first}, first.Source, nil)
	b.declareProjectionColumn(builder, column, diagnostics, expression.Name(), true)
	expression.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	return true
}

func (b *simpleProjectionBinder) addPrintFunctionTupleExpression(call *FunctionCallExpression, defaultName string, builder *ProjectionBuilder) bool {
	if call == nil {
		return false
	}
	tuple, ok := TupleForType(ExpressionRawResultType(call))
	if !ok || tuple == nil {
		return false
	}
	columns := tuple.Columns()
	if len(columns) == 1 && columns[0] != nil {
		column := b.newProjectionColumn(defaultName, columns[0].Type, call)
		builder.Add(column, ProjectionAddOptions{BaseName: defaultName, HasBaseName: true})
		return true
	}
	function, ok := expressionReferencedSymbol(call).(*FunctionSymbol)
	if !ok || function == nil {
		return false
	}
	for _, column := range columns {
		b.addFunctionTupleResultColumn(function, column, builder)
	}
	return true
}

func printExpressionType(expression Expression) *TypeSymbol {
	switch named := expression.(type) {
	case *SimpleNamedExpression:
		expression = named.Expression()
	case *CompoundNamedExpression:
		expression = named.Expression()
	}
	return ExpressionRawResultType(expression)
}

func printExpressionIsScalar(expression Expression) bool {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return printExpressionIsScalar(expression.Expression())
	case *CompoundNamedExpression:
		return printExpressionIsScalar(expression.Expression())
	case *ParenthesizedExpression:
		return printExpressionIsScalar(expression.Expression())
	default:
		if typ := ExpressionRawResultType(expression); typ != nil && typ.SymbolTabularity() == TabularityTabular {
			return false
		}
		return expressionResultIsScalar(expression)
	}
}

func (b *simpleProjectionBinder) bindProjectAway(node *ProjectAwayOperator) *SemanticInfo {
	if node == nil || node.ProjectAwayKeyword() == nil || node.ProjectAwayKeyword().Width() == 0 {
		return nil
	}
	return b.bindProjectColumns(node, node.Expressions(), true)
}

func (b *simpleProjectionBinder) bindProjectKeep(node *ProjectKeepOperator) *SemanticInfo {
	if node == nil || node.ProjectKeepKeyword() == nil || node.ProjectKeepKeyword().Width() == 0 {
		return nil
	}
	return b.bindProjectColumns(node, node.Expressions(), false)
}

func (b *simpleProjectionBinder) bindProjectReorder(node *ProjectReorderOperator) *SemanticInfo {
	if node == nil || node.ProjectReorderKeyword() == nil || node.ProjectReorderKeyword().Width() == 0 || node.Expressions() == nil {
		return nil
	}

	diagnostics := b.notFirstDiagnostics(node)
	var columns []*ColumnSymbol
	selected := map[*ColumnSymbol]struct{}{}
	exact := map[*ColumnSymbol]struct{}{}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil {
			return nil
		}
		expression := separated.Element()
		if syntaxElementIsMissing(expression) {
			continue
		}
		var ordering *OrderingClause
		if ordered, ok := expression.(*OrderedExpression); ok {
			ordering = ordered.Ordering()
			expression = ordered.Expression()
		}
		reference, ok := expression.(*NameReference)
		if !ok || reference == nil {
			return nil
		}
		matches, resolved := projectReorderColumns(reference.ReferencedSymbol())
		if !resolved {
			name := NameReferenceSimpleName(reference)
			if name == "" {
				continue
			}
			matches = []*ColumnSymbol{NewColumnSymbolWithSource(name, ScalarTypes.Unknown.Type(), reference)}
		}
		wildcard := isProjectColumnWildcard(reference)
		if wildcard && ordering != nil {
			matches = append([]*ColumnSymbol(nil), matches...)
			projectReorderSort(matches, ordering.AscOrDescKeyword().Kind())
		}
		for _, column := range matches {
			if wildcard {
				if _, exists := selected[column]; exists {
					continue
				}
			} else {
				if _, exists := exact[column]; exists {
					diagnostics = append(diagnostics, DiagnosticDuplicateColumnDeclaration(column.SymbolName()))
					continue
				}
				exact[column] = struct{}{}
				if _, exists := selected[column]; exists {
					continue
				}
			}
			selected[column] = struct{}{}
			columns = append(columns, column)
		}
	}
	if b.binder.rowScope != nil {
		for _, column := range b.binder.rowScope.Columns() {
			if _, exists := selected[column]; !exists {
				selected[column] = struct{}{}
				columns = append(columns, column)
			}
		}
	}
	table := NewTableSymbol("", columns...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (b *simpleProjectionBinder) bindProjectByNames(node *ProjectByNamesOperator) *SemanticInfo {
	if node == nil || node.ProjectByNamesKeyword() == nil || node.ProjectByNamesKeyword().Width() == 0 || node.Expressions() == nil {
		return nil
	}

	diagnostics := b.notFirstDiagnostics(node)
	projection := NewProjectionBuilder()
	for _, separated := range node.Expressions().Elements() {
		if separated == nil {
			return nil
		}
		if b.binder.rowScope == nil {
			b.addExpression(separated.Element(), projection, &diagnostics, false)
			continue
		}
		for _, name := range projectByNamesSelectors(separated.Element()) {
			if name == "" {
				continue
			}
			var matching []*ColumnSymbol
			for _, column := range b.binder.rowScope.Columns() {
				if column == nil || !KustoFactsMatches(name, column.SymbolName()) {
					continue
				}
				matching = append(matching, column)
			}
			projection.AddRange(matching, ProjectionAddRangeOptions{DoNotRepeat: true})
		}
	}
	columns := projection.GetProjection()
	table := NewTableSymbol("", columns...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func projectByNamesSelectors(expression Expression) []string {
	for expression != nil {
		variable, ok := expressionReferencedSymbol(expression).(*VariableSymbol)
		if !ok || variable.Source == nil {
			break
		}
		expression = variable.Source
	}
	if expression == nil {
		return nil
	}
	if dynamic, ok := expression.(*DynamicExpression); ok {
		if value, ok := ExpressionConstantValue(dynamic.Expression()).(string); ok {
			return []string{value}
		}
		if array, ok := dynamic.Expression().(*JsonArrayExpression); ok && array.Values() != nil {
			names := make([]string, 0, array.Values().Count())
			for _, separated := range array.Values().Elements() {
				if separated != nil {
					if value, ok := ExpressionConstantValue(separated.Element()).(string); ok {
						names = append(names, value)
					}
				}
			}
			return names
		}
	}
	if value, ok := ExpressionConstantValue(expression).(string); ok {
		return []string{value}
	}
	if call, ok := expression.(*FunctionCallExpression); ok && expressionReferencedSymbol(call) == Functions.ColumnNamesOf {
		arguments := functionCallArguments(call)
		if len(arguments) > 0 {
			if table := expressionTableResult(arguments[0]); table != nil {
				names := make([]string, 0, len(table.Columns()))
				for _, column := range table.Columns() {
					if column != nil {
						names = append(names, column.SymbolName())
					}
				}
				return names
			}
		}
	}
	return nil
}

func projectReorderColumns(symbol Symbol) ([]*ColumnSymbol, bool) {
	switch symbol := symbol.(type) {
	case *ColumnSymbol:
		return []*ColumnSymbol{symbol}, true
	case *GroupSymbol:
		columns := make([]*ColumnSymbol, 0, len(symbol.Members()))
		for _, member := range symbol.Members() {
			column, ok := member.(*ColumnSymbol)
			if !ok {
				return nil, false
			}
			columns = append(columns, column)
		}
		return columns, true
	default:
		return nil, false
	}
}

func projectReorderSort(columns []*ColumnSymbol, kind SyntaxKind) {
	descending := kind == SyntaxKindDescKeyword || kind == SyntaxKindGrannyDescKeyword
	granny := kind == SyntaxKindGrannyAscKeyword || kind == SyntaxKindGrannyDescKeyword
	sort.SliceStable(columns, func(i, j int) bool {
		var comparison int
		if granny {
			comparison = StringAndNumberCompareOrdinalIgnoreCase(columns[i].SymbolName(), columns[j].SymbolName())
		} else {
			comparison = ordinalCompare(columns[i].SymbolName(), columns[j].SymbolName(), true)
		}
		if descending {
			return comparison > 0
		}
		return comparison < 0
	})
}

func (b *simpleProjectionBinder) bindProjectColumns(operator SyntaxElement, expressions *SeparatedSyntaxListOf[Expression], away bool) *SemanticInfo {
	if expressions == nil {
		return nil
	}

	diagnostics := b.notFirstDiagnostics(operator)
	selectedNames := map[string]struct{}{}
	for _, separated := range expressions.Elements() {
		if separated == nil {
			return nil
		}
		reference, ok := separated.Element().(*NameReference)
		if !ok || reference == nil {
			return nil
		}
		if syntaxElementIsMissing(reference) {
			diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), reference))
			continue
		}
		switch symbol := reference.ReferencedSymbol().(type) {
		case *ColumnSymbol:
			selectedNames[symbol.SymbolName()] = struct{}{}
		case *GroupSymbol:
			allColumns := true
			for _, member := range symbol.Members() {
				column, ok := member.(*ColumnSymbol)
				if !ok {
					allColumns = false
					continue
				}
				selectedNames[column.SymbolName()] = struct{}{}
			}
			if !allColumns {
				diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), reference))
			}
		case nil, *ErrorSymbol:
			diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), reference))
		default:
			diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticColumnExpected(), reference))
		}
	}

	var columns []*ColumnSymbol
	if b.binder.rowScope != nil && expressions.Count() == 0 {
		if away {
			columns = append(columns, b.binder.rowScope.Columns()...)
		}
	} else if b.binder.rowScope != nil {
		for _, column := range b.binder.rowScope.Columns() {
			_, selected := selectedNames[column.SymbolName()]
			if selected != away {
				columns = append(columns, column)
			}
		}
	}
	table := NewTableSymbol("", columns...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (b *simpleProjectionBinder) bindProject(node *ProjectOperator) *SemanticInfo {
	if node == nil || !b.supportsExpressions(node.Expressions(), true) {
		return nil
	}

	diagnostics := b.notFirstDiagnostics(node)
	b.addProjectionScalarDiagnostic(node.Expressions(), &diagnostics)
	builder := NewProjectionBuilder()
	for _, separated := range node.Expressions().Elements() {
		b.addExpression(separated.Element(), builder, &diagnostics, false)
	}
	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (b *simpleProjectionBinder) bindExtend(node *ExtendOperator) *SemanticInfo {
	if node == nil || !b.supportsExpressions(node.Expressions(), false) {
		return nil
	}

	diagnostics := b.notFirstDiagnostics(node)
	b.addProjectionScalarDiagnostic(node.Expressions(), &diagnostics)
	builder := NewProjectionBuilder()
	if b.binder.rowScope != nil {
		builder.AddRange(b.binder.rowScope.Columns(), ProjectionAddRangeOptions{DoNotRepeat: true})
	}
	for _, separated := range node.Expressions().Elements() {
		b.addExpression(separated.Element(), builder, &diagnostics, true)
	}
	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (b *simpleProjectionBinder) bindMvExpand(node *MvExpandOperator, inPipe bool) *SemanticInfo {
	if node == nil || node.Expressions() == nil {
		return nil
	}
	diagnostics := b.binderNodeMvExpandDiagnostics(node, inPipe)
	builder := NewProjectionBuilder()
	if b.binder.rowScope != nil {
		builder.AddRange(b.binder.declaredAndInferredColumns(b.binder.rowScope), ProjectionAddRangeOptions{DoNotRepeat: true})
	}
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || separated.Element() == nil {
			return nil
		}
		mvExpression := separated.Element()
		expression := mvExpression.Expression()
		if expression == nil || expression.Width() == 0 {
			continue
		}
		underlying := mvExpandUnderlyingExpression(expression)
		name := mvExpandExpressionName(expression)
		if name == "" {
			name = b.nextDefaultColumnName(builder)
		}
		column := NewColumnSymbolWithSource(name, mvExpandExpressionResultType(mvExpression), underlying)
		if source, ok := directMvExpandColumn(underlying); ok {
			column.OriginalColumns = trulyOriginalColumns([]*ColumnSymbol{source})
		}
		builder.Add(column, ProjectionAddOptions{Replace: true})
		if named := mvExpandNamedExpression(expression); named != nil && named.Name() != nil {
			named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		}
	}
	if indexName := mvExpandItemIndexName(node); indexName != "" {
		index := NewColumnSymbol(indexName, ScalarTypes.Long.Type())
		builder.Add(index)
		for _, parameter := range node.Parameters().Elements() {
			if parameter != nil && findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), QueryOperatorParameters.MvExpandParameters) == QueryOperatorParameters.WithItemIndex {
				if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
					declaration.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(index))
				}
			}
		}
	}
	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func mvExpandNamedExpression(expression Expression) *SimpleNamedExpression {
	for {
		parenthesized, ok := expression.(*ParenthesizedExpression)
		if !ok {
			named, _ := expression.(*SimpleNamedExpression)
			return named
		}
		expression = parenthesized.Expression()
	}
}

func (b *simpleProjectionBinder) binderNodeMvExpandDiagnostics(node *MvExpandOperator, inPipe bool) []Diagnostic {
	nodeBinder := simpleNodeBinder{binder: b.binder}
	return nodeBinder.mvExpandDiagnostics(node, inPipe)
}

func mvExpandUnderlyingExpression(expression Expression) Expression {
	for {
		switch current := expression.(type) {
		case *ParenthesizedExpression:
			expression = current.Expression()
		case *SimpleNamedExpression:
			expression = current.Expression()
		default:
			return expression
		}
	}
}

func directMvExpandColumn(expression Expression) (*ColumnSymbol, bool) {
	reference, ok := expression.(*NameReference)
	if !ok || reference == nil {
		return nil, false
	}
	column, ok := reference.ReferencedSymbol().(*ColumnSymbol)
	return column, ok
}

func mvExpandExpressionName(expression Expression) string {
	return expressionResultName(expression, "", nil, nil)
}

func mvExpandExpressionResultType(expression *MvExpandExpression) *TypeSymbol {
	if expression == nil {
		return ScalarTypes.Unknown.Type()
	}
	if clause := expression.ToTypeOf(); clause != nil {
		if clause.TypeOf() != nil {
			if typ := returnTypeFromSymbol(expressionReferencedSymbol(clause.TypeOf())); typ != nil {
				return typ
			}
		}
		return &Error.TypeSymbol
	}
	typ := ExpressionResultType(mvExpandUnderlyingExpression(expression.Expression()))
	if elementType := GetElementType(typ); elementType != nil {
		return elementType
	}
	if isDynamicType(typ) {
		return ScalarTypes.Dynamic.Type()
	}
	if typ == nil || typ == &Error.TypeSymbol {
		return ScalarTypes.Unknown.Type()
	}
	return typ
}

func mvExpandItemIndexName(node *MvExpandOperator) string {
	if node == nil || node.Parameters() == nil {
		return ""
	}
	for _, parameter := range node.Parameters().Elements() {
		if parameter == nil || findSpecificQueryOperatorParameter(NameDeclarationSimpleName(parameter.Name()), QueryOperatorParameters.MvExpandParameters) != QueryOperatorParameters.WithItemIndex {
			continue
		}
		if declaration, ok := parameter.Expression().(*NameDeclaration); ok {
			return NameDeclarationSimpleName(declaration)
		}
	}
	return ""
}

func (b *simpleProjectionBinder) supportsSerializeExpressions(expressions *SeparatedSyntaxListOf[Expression]) bool {
	if expressions == nil {
		return false
	}
	for _, separated := range expressions.Elements() {
		if separated == nil || !b.supportsSerializeExpression(separated.Element()) {
			return false
		}
	}
	return true
}

func (b *simpleProjectionBinder) supportsSerializeExpression(expression Expression) bool {
	switch expression := expression.(type) {
	case *NameReference:
		if isRecoverableMissingProjectionExpression(expression) {
			return true
		}
		if _, unresolved := expression.ReferencedSymbol().(*ErrorSymbol); unresolved {
			return isRecoverableBareErrorProjectionReference(expression)
		}
		return b.supportsUnnamedProjectionExpression(expression)
	case *SimpleNamedExpression:
		return expression.Name() != nil && NameDeclarationSimpleName(expression.Name()) != "" &&
			serializeNameDeclarationIsSupported(expression.Name()) &&
			expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			b.supportsSerializeExpression(expression.Expression())
	case *CompoundNamedExpression:
		return compoundRenameListIsSupported(expression.Names()) && expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			expression.Expression() != nil && expressionIsBound(expression.Expression()) && compoundProjectionTypeIsSupported(expression.Expression())
	case *PathExpression, *ElementExpression:
		return expressionIsBound(expression) && ExpressionRawResultType(expression) != nil
	default:
		return b.supportsUnnamedProjectionExpression(expression)
	}
}

func serializeNameDeclarationIsSupported(declaration *NameDeclaration) bool {
	if declaration == nil {
		return false
	}
	switch name := declaration.Name().(type) {
	case *TokenName:
		return name.Name() != nil && name.Name().Width() != 0 && !strings.Contains(name.Name().RawText(), ".")
	case *BracketedName:
		return name.OpenBracket() != nil && name.OpenBracket().Width() != 0 && name.Name() != nil && name.Name().Width() != 0 &&
			name.CloseBracket() != nil && name.CloseBracket().Width() != 0
	default:
		return false
	}
}

func (b *simpleProjectionBinder) addSerializeExpressions(expressions *SeparatedSyntaxListOf[Expression], builder *ProjectionBuilder, diagnostics *[]Diagnostic) {
	for _, separated := range expressions.Elements() {
		b.addSerializeExpression(separated.Element(), builder, diagnostics)
	}
}

func (b *simpleProjectionBinder) addSerializeExpression(expression Expression, builder *ProjectionBuilder, diagnostics *[]Diagnostic) {
	if reference, ok := expression.(*NameReference); ok && isRecoverableMissingProjectionExpression(reference) {
		return
	}
	if compound, ok := expression.(*CompoundNamedExpression); ok {
		if b.addCompoundNamedExpression(compound, builder, diagnostics, compoundProjectionDefault) {
			return
		}
	}
	if named, ok := expression.(*SimpleNamedExpression); ok {
		name := NameDeclarationSimpleName(named.Name())
		inner := named.Expression()
		if tuple, ok := TupleForType(ExpressionResultType(inner)); ok {
			b.addSimpleNamedTupleExpression(name, named.Name(), inner, tuple, builder, diagnostics)
			return
		}
		var column *ColumnSymbol
		if source := serializeResultColumn(inner); source != nil {
			column = b.newProjectionColumn(name, source.Type, inner)
			b.declareSerializeColumn(builder, column, diagnostics, named.Name(), true)
		} else {
			column = b.newProjectionColumn(name, serializeProjectionType(inner, false), inner)
			b.declareSerializeColumn(builder, column, diagnostics, named.Name(), false)
		}
		named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		return
	}

	if column := serializeResultColumn(expression); column != nil {
		b.declareSerializeColumn(builder, column, diagnostics, expression, false)
		return
	}
	if reference, ok := expression.(*NameReference); ok && isRecoverableBareErrorProjectionReference(reference) {
		name := NameReferenceSimpleName(reference)
		builder.Add(b.newProjectionColumn(name, ExpressionRawResultType(reference), reference))
		return
	}
	if call, ok := expression.(*FunctionCallExpression); ok {
		if b.addFunctionTupleExpression(call, builder) {
			return
		}
	}

	typ := b.unnamedProjectionType(expression, diagnostics)
	name := expressionResultName(expression, "", b.binder.rowScope, b.binder.Globals)
	if name == "" {
		name = b.nextDefaultColumnName(builder)
	}
	builder.Add(b.newProjectionColumn(name, typ, expression))
}

func (b *simpleProjectionBinder) declareSerializeColumn(builder *ProjectionBuilder, column *ColumnSymbol, diagnostics *[]Diagnostic, location SyntaxElement, replace bool) {
	var local []Diagnostic
	builder.Declare(column, &local, location, replace)
	for range local {
		*diagnostics = append(*diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticDuplicateColumnDeclaration(column.SymbolName()), location))
	}
}

func serializeResultColumn(expression Expression) *ColumnSymbol {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return serializeResultColumn(expression.Expression())
	case *NameReference:
		column, _ := expression.ReferencedSymbol().(*ColumnSymbol)
		return column
	default:
		return nil
	}
}

func serializeProjectionType(expression Expression, checkScalar bool) *TypeSymbol {
	typ := ExpressionRawResultType(expression)
	if typ == nil {
		return ScalarTypes.Unknown.Type()
	}
	if checkScalar && !expressionResultIsScalar(expression) {
		return ScalarTypes.Unknown.Type()
	}
	return typ
}

func serializeExpressionIsTabular(expression Expression) bool {
	typ := ExpressionResultType(expression)
	return typ != nil && typ != &Error.TypeSymbol && !expressionResultIsScalar(expression)
}

func serializeExpressionResultName(expression Expression) string {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return serializeExpressionResultName(expression.Expression())
	case *NameReference:
		return NameReferenceSimpleName(expression)
	default:
		return ""
	}
}

func expressionResultName(expression Expression, defaultName string, row *TableSymbol, globals *GlobalState) string {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return expressionResultName(expression.Expression(), defaultName, row, globals)
	case *NameReference:
		return NameReferenceSimpleName(expression)
	case *BracketedExpression:
		inner := expression.Expression()
		typ := ExpressionResultType(inner)
		if ExpressionIsLiteral(inner) && (sameType(typ, ScalarTypes.String.Type()) || sameType(typ, ScalarTypes.Long.Type()) || sameType(typ, ScalarTypes.Int.Type())) {
			if value := ExpressionLiteralValue(inner); value != nil {
				return fmt.Sprint(value)
			}
		}
		return defaultName
	case *PathExpression:
		return selectorExpressionResultName(expression.Expression(), expression.Selector(), defaultName, row, globals)
	case *ElementExpression:
		return selectorExpressionResultName(expression.Expression(), expression.Selector(), defaultName, row, globals)
	case *OrderedExpression:
		return expressionResultName(expression.Expression(), defaultName, row, globals)
	case *SimpleNamedExpression:
		return NameDeclarationSimpleName(expression.Name())
	case *FunctionCallExpression:
		return functionExpressionResultName(expression, defaultName, row, globals)
	default:
		return defaultName
	}
}

func selectorExpressionResultName(left, selector Expression, defaultName string, row *TableSymbol, globals *GlobalState) string {
	typ := ExpressionResultType(left)
	if IsDynamicArrayOrBag(typ) || sameType(typ, ScalarTypes.Unknown.Type()) {
		leftName := expressionResultName(left, "", row, globals)
		rightName := expressionResultName(selector, "", row, globals)
		if strings.TrimSpace(leftName) != "" {
			return leftName + "_" + rightName
		}
		return rightName
	}
	return expressionResultName(selector, defaultName, row, globals)
}

func functionExpressionResultName(call *FunctionCallExpression, defaultName string, row *TableSymbol, globals *GlobalState) string {
	function := functionCallReferencedFunction(call)
	if function == nil && globals != nil {
		function = globals.GetFunction(NameReferenceSimpleName(call.Name()))
	}
	if function == nil {
		return defaultName
	}
	kind := function.ResultNameKind()
	prefix := function.ResultNamePrefix()
	if kind == ResultNameNameAndFirstArgument {
		prefix, kind = function.SymbolName(), ResultNamePrefixAndFirstArgument
	} else if kind == ResultNameNameAndOnlyArgument {
		prefix, kind = function.SymbolName(), ResultNamePrefixAndOnlyArgument
	}
	arguments := call.ArgumentList().Expressions()
	count := 0
	if arguments != nil {
		count = arguments.Count()
	}
	firstName := defaultName
	if count > 0 {
		firstName = expressionResultName(arguments.At(0).Element(), defaultName, row, globals)
	}
	switch kind {
	case ResultNamePrefixAndFirstArgument:
		if count > 0 {
			return prefix + "_" + firstName
		}
		if prefix != "" {
			return prefix + "_"
		}
		return ""
	case ResultNamePrefixAndOnlyArgument:
		if count == 1 {
			return prefix + "_" + firstName
		}
	case ResultNameFirstArgumentValueIfColumn:
		if count > 0 {
			if name, ok := ExpressionConstantValue(arguments.At(0).Element()).(string); ok && row != nil && row.GetColumn(name) != nil {
				return name
			}
		}
	case ResultNameFirstArgument:
		if count > 0 {
			return firstName
		}
		return ""
	case ResultNamePrefixOnly:
		if prefix != "" {
			return prefix
		}
	case ResultNameOnlyArgument:
		if count == 1 {
			return firstName
		}
	}
	return defaultName
}

func (b *simpleProjectionBinder) nextDefaultColumnName(builder *ProjectionBuilder) string {
	if b == nil || b.defaultColumnNameSuffix == nil {
		return ""
	}
	for {
		name := "Column" + strconv.Itoa(*b.defaultColumnNameSuffix)
		*b.defaultColumnNameSuffix++
		if builder != nil {
			if _, exists := builder.columnIndexMap[name]; exists {
				continue
			}
		}
		if b.binder != nil && len(b.binder.matchingNameSymbols(name, SymbolMatchAny)) != 0 {
			continue
		}
		return name
	}
}

// Kept for projection helpers outside this tranche that still own a local suffix.
func serializeDefaultColumnName(builder *ProjectionBuilder, suffix *int) string {
	for {
		name := "Column" + strconv.Itoa(*suffix)
		*suffix++
		if builder == nil {
			return name
		}
		if _, exists := builder.columnIndexMap[name]; !exists {
			return name
		}
	}
}

func (b *simpleProjectionBinder) bindDistinct(node *DistinctOperator) *SemanticInfo {
	if node == nil || node.DistinctKeyword() == nil || node.DistinctKeyword().Width() == 0 ||
		node.Parameters() == nil || node.Expressions() == nil || node.Expressions().Count() == 0 {
		return nil
	}

	diagnostics := b.notFirstDiagnostics(node)
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(b.binder, node.Parameters(), QueryOperatorParameters.DistinctParameters)...)
	builder := NewProjectionBuilder()
	var star SyntaxElement
	for _, separated := range node.Expressions().Elements() {
		if separated == nil || separated.Element() == nil {
			return nil
		}
		expression := separated.Element()
		switch expression.(type) {
		case *StarExpression:
			star = expression
			if b.binder.rowScope != nil {
				builder.AddRange(b.binder.declaredAndInferredColumns(b.binder.rowScope))
			}
		default:
			if !b.addDistinctExpression(expression, builder, &diagnostics) {
				return nil
			}
		}
	}

	projection := builder.GetProjection()
	for _, separated := range node.Expressions().Elements() {
		expression := separated.Element()
		if _, ok := expression.(*StarExpression); ok {
			continue
		}
		check := distinctValidationExpression(expression)
		typ := ExpressionResultType(check)
		if typ != nil && typ != &Error.TypeSymbol && !IsScalar(typ) {
			appendProjectionDiagnostic(&diagnostics, DiagnosticScalarTypeExpected(), expression)
			break
		}
	}
	if !semanticDiagnosticsContainCode(diagnostics, "KS108") {
		for _, separated := range node.Expressions().Elements() {
			expression := separated.Element()
			if _, ok := expression.(*StarExpression); ok {
				continue
			}
			check := distinctValidationExpression(expression)
			if typ := ExpressionResultType(check); isDynamicType(typ) {
				appendProjectionDiagnostic(&diagnostics, DiagnosticTypeNotAllowed(typ.SymbolName()), expression)
				break
			}
		}
		if star != nil && !semanticDiagnosticsContainCode(diagnostics, "KS116") {
			for _, column := range projection {
				if column != nil && isDynamicType(column.Type) {
					appendProjectionDiagnostic(&diagnostics, DiagnosticTypeNotAllowed(column.Type.SymbolName()), star)
					break
				}
			}
		}
	}

	table := NewTableSymbol("", projection...).WithInheritableProperties(b.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func (b *simpleProjectionBinder) addDistinctExpression(expression Expression, builder *ProjectionBuilder, diagnostics *[]Diagnostic) bool {
	if compound, ok := expression.(*CompoundNamedExpression); ok {
		return b.addCompoundNamedExpression(compound, builder, diagnostics, compoundProjectionDefault)
	}
	if named, ok := expression.(*SimpleNamedExpression); ok {
		if tuple, function := summarizeTupleResult(named.Expression()); tuple != nil {
			columns := tuple.Columns()
			if len(columns) > 0 {
				first := columns[0]
				column := NewColumnSymbolFull(NameDeclarationSimpleName(named.Name()), first.Type, "", []*ColumnSymbol{first}, first.Source, nil)
				builder.Declare(column, diagnostics, named.Name(), true)
				named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
				addSummarizeTupleColumns(columns[1:], function, builder)
			}
			return true
		}
		name := NameDeclarationSimpleName(named.Name())
		if name == "" {
			return false
		}
		column := NewColumnSymbolWithSource(name, distinctExpressionType(named.Expression()), named.Expression())
		builder.Declare(column, diagnostics, named.Name())
		named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		return true
	}
	if tuple, function := summarizeTupleResult(expression); tuple != nil {
		addSummarizeTupleColumns(tuple.Columns(), function, builder)
		return true
	}
	if reference, ok := expression.(*NameReference); ok {
		switch symbol := reference.ReferencedSymbol().(type) {
		case *ColumnSymbol:
			builder.Declare(symbol, diagnostics, expression)
			return true
		case *ErrorSymbol:
			builder.Declare(NewColumnSymbolWithSource(NameReferenceSimpleName(reference), &Error.TypeSymbol, reference), diagnostics, expression)
			return true
		}
	}
	if column := projectionResultColumn(expression); column != nil {
		builder.Declare(column, diagnostics, expression)
		return true
	}
	name := expressionResultName(expression, "", b.binder.rowScope, b.binder.Globals)
	if name == "" {
		name = b.nextDefaultColumnName(builder)
	}
	builder.Add(NewColumnSymbolWithSource(name, distinctExpressionType(expression), expression))
	return true
}

func distinctExpressionType(expression Expression) *TypeSymbol {
	typ := ExpressionResultType(expression)
	if typ == nil {
		return ScalarTypes.Unknown.Type()
	}
	return typ
}

func distinctValidationExpression(expression Expression) Expression {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return expression.Expression()
	case *CompoundNamedExpression:
		return expression.Expression()
	default:
		return expression
	}
}

func projectionResultColumn(expression Expression) *ColumnSymbol {
	return GetResultColumn(expression)
}

// GetResultColumn mirrors Binder.GetResultColumn. A no-op conversion retains
// the identity of its single source column.
func GetResultColumn(expression Expression) *ColumnSymbol {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return GetResultColumn(expression.Expression())
	case *FunctionCallExpression:
		if functionCallArgumentCount(expression) != 1 || !isTypeConversionFunction(expressionReferencedSymbol(expression)) {
			return nil
		}
		column, _ := expressionReferencedSymbol(functionCallArgumentAt(expression, 0)).(*ColumnSymbol)
		if column != nil && sameType(column.Type, ExpressionResultType(expression)) {
			return column
		}
	default:
		column, _ := expressionReferencedSymbol(expression).(*ColumnSymbol)
		return column
	}
	return nil
}

// GetUnderlyingExpression removes the naming and ordering adornments used by
// summarize and custom-return helpers.
func GetUnderlyingExpression(expression Expression) Expression {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return GetUnderlyingExpression(expression.Expression())
	case *OrderedExpression:
		return GetUnderlyingExpression(expression.Expression())
	default:
		return expression
	}
}

func GetExpressionDeclaredName(expression Expression) string {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return NameDeclarationSimpleName(expression.Name())
	case *OrderedExpression:
		return GetExpressionDeclaredName(expression.Expression())
	default:
		return ""
	}
}

func GetGraphExpressionResultName(expression Expression) string {
	if expression == nil {
		return ""
	}
	var left Expression
	var selector Expression
	switch expression := expression.(type) {
	case *PathExpression:
		left, selector = expression.Expression(), expression.Selector()
	case *ElementExpression:
		left, selector = expression.Expression(), expression.Selector()
	default:
		return expressionResultName(expression, "", nil, nil)
	}
	leftName := ""
	switch left.(type) {
	case *PathExpression, *ElementExpression:
		leftName = GetGraphExpressionResultName(left)
	default:
		leftName = expressionResultName(left, "", nil, nil)
	}
	rightName := expressionResultName(selector, "", nil, nil)
	if strings.TrimSpace(leftName) != "" {
		return leftName + "_" + rightName
	}
	return rightName
}

func IsConversionFunction(function *FunctionSymbol) bool {
	return isTypeConversionFunction(function)
}

func IsConversionFunctionExpression(expression Expression) bool {
	function, _ := expressionReferencedSymbol(expression).(*FunctionSymbol)
	return IsConversionFunction(function)
}

func isTypeConversionFunction(symbol Symbol) bool {
	function, _ := symbol.(*FunctionSymbol)
	switch function {
	case Functions.ToString, Functions.ToLong, Functions.ToInt, Functions.ToReal, Functions.ToDouble,
		Functions.ToDateTime, Functions.ToTimespan, Functions.ToTime, Functions.ToBool,
		Functions.ToDecimal, Functions.ToGuid, Functions.ToDynamic:
		return true
	default:
		return false
	}
}

func (b *simpleProjectionBinder) bindSummarize(node *SummarizeOperator) *SemanticInfo {
	if node == nil || node.Parameters() == nil || node.Aggregates() == nil {
		return nil
	}
	var diagnostics []Diagnostic
	if !b.hasPipedInput(node) {
		diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.SummarizeKeyword()))
	}
	diagnostics = append(diagnostics, queryOperatorParameterDiagnostics(b.binder, node.Parameters(), QueryOperatorParameters.SummarizeParameters)...)
	for _, separated := range node.Aggregates().Elements() {
		expression := separated.Element()
		check := summarizeValidationExpression(expression)
		if typ := ExpressionResultType(check); typ != nil && typ != &Error.TypeSymbol && !IsScalar(typ) {
			diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
			break
		}
	}
	if by := node.ByClause(); by != nil && !semanticDiagnosticsContainCode(diagnostics, "KS108") {
		for _, separated := range by.Expressions().Elements() {
			expression := separated.Element()
			check := summarizeValidationExpression(expression)
			if typ := ExpressionResultType(check); isDynamicType(typ) {
				diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticTypeNotAllowed(typ.SymbolName()), expression))
				break
			}
		}
	}

	builder := NewProjectionBuilder()
	if by := node.ByClause(); by != nil {
		for _, separated := range by.Expressions().Elements() {
			expression := separated.Element()
			if syntaxElementIsMissing(expression) {
				continue
			}
			b.addSummarizeExpression(expression, builder, &diagnostics, false)
		}
		builder.DoNotAddAny(builder.GetProjection())
	}
	for _, separated := range node.Aggregates().Elements() {
		b.addSummarizeExpression(separated.Element(), builder, &diagnostics, true)
	}
	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(b.binder.rowScope).WithIsSorted(false)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func summarizeValidationExpression(expression Expression) Expression {
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return expression.Expression()
	case *CompoundNamedExpression:
		return expression.Expression()
	default:
		return expression
	}
}

func (b *simpleProjectionBinder) bindTopNested(node *TopNestedOperator) *SemanticInfo {
	if node == nil || node.Clauses() == nil || node.Clauses().Count() == 0 {
		return nil
	}

	var diagnostics []Diagnostic
	if !b.hasPipedInput(node) {
		diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node))
	}
	columns := make([]*ColumnSymbol, 0, node.Clauses().Count()*2)
	uniqueNames := NewUniqueNameTable()
	for _, separated := range node.Clauses().Elements() {
		clause := separated.Element()
		if clause.Expression() != nil {
			diagnostics = append(diagnostics, topNestedIntegerDiagnostics(b.binder, clause.Expression())...)
		}
		topNestedAppendScalarDiagnostic(b.binder, &diagnostics, clause.OfExpression())
		if others := clause.WithOthersClause(); others != nil {
			topNestedAppendScalarDiagnostic(b.binder, &diagnostics, others.Expression())
		}
		topNestedAppendScalarDiagnostic(b.binder, &diagnostics, clause.ByExpression())

		ofName := uniqueNames.GetOrAddName(topNestedExpressionResultName(clause.OfExpression()))
		columns = append(columns, topNestedColumnForExpression(clause.OfExpression(), ofName, false))
		byName := topNestedExpressionDeclaredName(clause.ByExpression())
		if byName == "" {
			byName = "aggregated_" + ofName
		}
		byName = uniqueNames.GetOrAddName(byName)
		columns = append(columns, topNestedColumnForExpression(clause.ByExpression(), byName, false))
	}

	table := NewTableSymbol("", columns...).WithInheritableProperties(b.binder.rowScope).WithIsSorted(true)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func topNestedIntegerDiagnostics(binder *Binder, expression Expression) []Diagnostic {
	typ := ExpressionResultType(expression)
	if !topDiagnosticTypeKnown(typ) || IsInteger(typ) {
		return nil
	}
	return []Diagnostic{binder.diagnosticAtSyntaxElement(DiagnosticExpressionMustBeInteger(), expression)}
}

func topNestedAppendScalarDiagnostic(binder *Binder, diagnostics *[]Diagnostic, expression Expression) {
	typ := ExpressionResultType(expression)
	if topDiagnosticTypeKnown(typ) && !expressionResultIsScalar(expression) {
		*diagnostics = append(*diagnostics, binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
	}
}

func topNestedUnderlyingExpression(expression Expression) Expression {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return topNestedUnderlyingExpression(expression.Expression())
	case *OrderedExpression:
		return topNestedUnderlyingExpression(expression.Expression())
	default:
		return expression
	}
}

func topNestedExpressionResultName(expression Expression) string {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return topNestedExpressionResultName(expression.Expression())
	case *OrderedExpression:
		return topNestedExpressionResultName(expression.Expression())
	case *NameReference:
		return NameReferenceSimpleName(expression)
	case *SimpleNamedExpression:
		return NameDeclarationSimpleName(expression.Name())
	case *FunctionCallExpression:
		return summarizeFunctionResultName(expression)
	default:
		return ""
	}
}

func topNestedExpressionDeclaredName(expression Expression) string {
	switch expression := expression.(type) {
	case *OrderedExpression:
		return topNestedExpressionDeclaredName(expression.Expression())
	case *SimpleNamedExpression:
		return NameDeclarationSimpleName(expression.Name())
	default:
		return ""
	}
}

func topNestedColumnForExpression(expression Expression, name string, compatibility bool) *ColumnSymbol {
	declared, _ := topNestedUnderlyingExpression(expression).(*SimpleNamedExpression)
	underlying := topNestedUnderlyingExpression(expression)
	if declared != nil {
		named := declared
		underlying = named.Expression()
	}
	typ := summarizeExpressionType(underlying, compatibility)
	var column *ColumnSymbol
	if sourceColumn := summarizeResultColumn(underlying); sourceColumn != nil {
		if sourceColumn.SymbolName() == name {
			if declared != nil && declared.Name() != nil {
				declared.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(sourceColumn))
			}
			return sourceColumn
		}
		column = NewColumnSymbolFull(name, typ, "", []*ColumnSymbol{sourceColumn}, expression, nil)
	} else {
		if compatibility && isErrorTypeSymbol(typ) {
			column = &ColumnSymbol{BaseSymbol: NewBaseSymbol(name), Type: typ, Source: expression}
		} else {
			column = NewColumnSymbolWithSource(name, typ, expression)
		}
	}
	if declared != nil && declared.Name() != nil {
		declared.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
	}
	return column
}

func (b *simpleProjectionBinder) addSummarizeExpression(expression Expression, builder *ProjectionBuilder, diagnostics *[]Diagnostic, aggregate bool, _ ...*int) {
	if syntaxElementIsMissing(expression) {
		return
	}
	if compound, ok := expression.(*CompoundNamedExpression); ok {
		if b.addCompoundNamedExpression(compound, builder, diagnostics, compoundProjectionDefault) {
			return
		}
	}
	if named, ok := expression.(*SimpleNamedExpression); ok {
		if tuple, function := summarizeTupleResult(named.Expression()); tuple != nil {
			columns := tuple.Columns()
			if len(columns) > 0 {
				first := columns[0]
				column := b.newSummarizeProjectionColumnFull(NameDeclarationSimpleName(named.Name()), first.Type, first.Source, first)
				builder.Declare(column, diagnostics, named.Name(), true)
				named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
				addSummarizeTupleColumns(columns[1:], function, builder)
			}
			return
		}
		name := NameDeclarationSimpleName(named.Name())
		column := b.newSummarizeProjectionColumn(name, summarizeExpressionType(named.Expression(), false), named.Expression())
		builder.Declare(column, diagnostics, named.Name())
		named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		return
	}
	if tuple, function := summarizeTupleResult(expression); tuple != nil {
		addSummarizeTupleColumns(tuple.Columns(), function, builder)
		return
	}
	if column := summarizeResultColumn(expression); column != nil {
		builder.Declare(column, diagnostics, expression)
		return
	}
	name := summarizeExpressionName(expression, aggregate)
	if name == "" {
		name = b.nextDefaultColumnName(builder)
	}
	builder.Add(b.newSummarizeProjectionColumn(name, summarizeExpressionType(expression, false), expression))
}

func summarizeTupleResult(expression Expression) (*TupleSymbol, *FunctionSymbol) {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return summarizeTupleResult(expression.Expression())
	case *FunctionCallExpression:
		tuple, _ := TupleForType(ExpressionRawResultType(expression))
		function := functionCallReferencedFunction(expression)
		return tuple, function
	default:
		return nil, nil
	}
}

func addSummarizeTupleColumns(columns []*ColumnSymbol, function *FunctionSymbol, builder *ProjectionBuilder) {
	for _, source := range columns {
		if source == nil {
			continue
		}
		column := source
		if function != nil && function.ResultNamePrefix() != "" {
			column = source.WithName(function.ResultNamePrefix() + "_" + source.SymbolName())
		}
		builder.Add(column)
	}
}

func summarizeResultColumn(expression Expression) *ColumnSymbol {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return summarizeResultColumn(expression.Expression())
	case *NameReference:
		column, _ := expression.ReferencedSymbol().(*ColumnSymbol)
		return column
	default:
		return nil
	}
}

func summarizeExpressionType(expression Expression, compatibility bool) *TypeSymbol {
	if call, ok := expression.(*FunctionCallExpression); ok && call.Name() != nil && call.Name().ReferencedSymbol() == Error {
		return &Error.TypeSymbol
	}
	typ := ExpressionResultType(expression)
	if _, ok := expression.(*FunctionCallExpression); ok && isErrorTypeSymbol(typ) {
		return &Error.TypeSymbol
	}
	if compatibility && isErrorTypeSymbol(typ) {
		return &Error.TypeSymbol
	}
	if typ == nil || isErrorTypeSymbol(typ) || !IsScalar(typ) {
		return ScalarTypes.Unknown.Type()
	}
	return typ
}

func summarizeExpressionName(expression Expression, aggregate bool) string {
	switch expression := expression.(type) {
	case *ParenthesizedExpression:
		return summarizeExpressionName(expression.Expression(), aggregate)
	case *NameReference:
		return NameReferenceSimpleName(expression)
	case *FunctionCallExpression:
		if aggregate {
			return summarizeFunctionResultName(expression)
		}
		return expressionResultName(expression, "", nil, nil)
	case *PathExpression, *ElementExpression:
		return expressionResultName(expression, "", nil, nil)
	}
	return ""
}

func summarizeFunctionResultName(call *FunctionCallExpression) string {
	function := functionCallReferencedFunction(call)
	if function == nil {
		return ""
	}
	kind := function.ResultNameKind()
	prefix := function.ResultNamePrefix()
	if kind == ResultNameNameAndFirstArgument {
		prefix, kind = function.SymbolName(), ResultNamePrefixAndFirstArgument
	} else if kind == ResultNameNameAndOnlyArgument {
		prefix, kind = function.SymbolName(), ResultNamePrefixAndOnlyArgument
	}
	arguments := call.ArgumentList().Expressions()
	firstName := ""
	if arguments != nil && arguments.Count() > 0 {
		firstName = summarizeExpressionName(arguments.At(0).Element(), false)
	}
	switch kind {
	case ResultNamePrefixAndFirstArgument:
		if arguments != nil && arguments.Count() > 0 {
			return prefix + "_" + firstName
		}
		if prefix != "" {
			return prefix + "_"
		}
	case ResultNamePrefixAndOnlyArgument:
		if arguments != nil && arguments.Count() == 1 {
			return prefix + "_" + firstName
		}
	case ResultNameFirstArgument:
		return firstName
	case ResultNamePrefixOnly:
		return prefix
	case ResultNameOnlyArgument:
		if arguments != nil && arguments.Count() == 1 {
			return firstName
		}
	}
	return ""
}

func queryOperatorParameterDiagnostics(binder *Binder, parameters *SyntaxListOf[*NamedParameter], declarations []*QueryOperatorParameter) []Diagnostic {
	if parameters == nil {
		return nil
	}
	return queryOperatorParameterElementDiagnostics(binder, parameters.Elements(), declarations)
}

func queryOperatorParameterElementDiagnostics(binder *Binder, parameters []*NamedParameter, declarations []*QueryOperatorParameter) []Diagnostic {
	var diagnostics []Diagnostic
	seen := map[*QueryOperatorParameter]bool{}
	appendDiagnostic := func(diagnostic Diagnostic, location SyntaxElement) {
		if binder != nil {
			diagnostic = binder.diagnosticAtSyntaxElement(diagnostic, location)
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	for _, parameter := range parameters {
		if parameter == nil || parameter.Name() == nil {
			continue
		}
		name := NameDeclarationSimpleName(parameter.Name())
		if name == "" {
			continue
		}
		declaration := findSpecificQueryOperatorParameter(name, declarations)
		if declaration == nil {
			appendDiagnostic(DiagnosticUnknownQueryOperatorParameterName(name), queryOperatorParameterNameLocation(parameter.Name()))
			continue
		}
		parameter.Name().SetSemanticInfo(NewSemanticInfoForSymbol(declaration, ScalarTypes.Unknown.Type()))
		if seen[declaration] && !declaration.IsRepeatable() {
			appendDiagnostic(DiagnosticParameterAlreadySpecified(name), queryOperatorParameterNameLocation(parameter.Name()))
		}
		seen[declaration] = true
		expression := parameter.Expression()
		if expression == nil || expression.ContainsSyntaxDiagnostics() {
			continue
		}
		if queryOperatorParameterKindIsValid(expression, declaration) {
			continue
		}
		typ := ExpressionResultType(expression)
		suppressUnknown := sameType(typ, ScalarTypes.Unknown.Type())
		suppressError := isErrorTypeSymbol(typ)
		valueKind := declaration.ValueKind()
		values := declaration.Values()
		switch valueKind {
		case QueryOperatorParameterValueIntegerLiteral:
			if !suppressError && !suppressUnknown && (!ExpressionIsLiteral(expression) || !IsInteger(typ)) {
				appendDiagnostic(DiagnosticIntegerLiteralExpected(), expression)
			}
		case QueryOperatorParameterValueNumericLiteral, QueryOperatorParameterValueForcedRealLiteral:
			if !suppressError && !suppressUnknown && (!ExpressionIsLiteral(expression) || !IsNumeric(typ)) {
				appendDiagnostic(DiagnosticSummableLiteralExpected(), expression)
			}
		case QueryOperatorParameterValueScalarLiteral:
			if !suppressError && !suppressUnknown && (!ExpressionIsLiteral(expression) || !queryOperatorParameterIsScalarType(typ)) {
				appendDiagnostic(DiagnosticScalarLiteralExpected(), expression)
			}
		case QueryOperatorParameterValueSummableLiteral:
			if !suppressError && !suppressUnknown && (!ExpressionIsLiteral(expression) || !IsSummable(typ)) {
				appendDiagnostic(DiagnosticSummableLiteralExpected(), expression)
			}
		case QueryOperatorParameterValueStringLiteral:
			if !suppressError && !suppressUnknown && (!ExpressionIsLiteral(expression) || !sameType(typ, ScalarTypes.String.Type())) {
				appendDiagnostic(DiagnosticStringLiteralExpected(), expression)
			}
		case QueryOperatorParameterValueBoolLiteral:
			if !suppressError && !suppressUnknown && (!ExpressionIsLiteral(expression) || !sameType(typ, ScalarTypes.Bool.Type())) {
				appendDiagnostic(DiagnosticBooleanLiteralExpected(), expression)
			}
		case QueryOperatorParameterValueString:
			if !suppressError && !suppressUnknown && !sameType(typ, ScalarTypes.String.Type()) && !sameType(typ, ScalarTypes.Dynamic.Type()) {
				appendDiagnostic(DiagnosticExpressionMustHaveType("string", "dynamic"), expression)
			}
		case QueryOperatorParameterValueColumn:
			if !suppressError {
				appendDiagnostic(DiagnosticColumnExpected(), expression)
			}
		case QueryOperatorParameterValueWord, QueryOperatorParameterValueWordOrNumber:
			if !suppressError {
				appendDiagnostic(DiagnosticTermsExpected(values...), expression)
			}
		}

		if valueKind != QueryOperatorParameterValueWord &&
			valueKind != QueryOperatorParameterValueWordOrNumber && len(values) > 0 && !suppressError {
			if !ExpressionIsLiteral(expression) {
				appendDiagnostic(DiagnosticExpressionMustBeLiteral(), expression)
			}
			if !queryOperatorLiteralValueIsAllowed(expression, declaration) {
				appendDiagnostic(DiagnosticExpressionMustHaveValue(values...), expression)
			}
		}
	}
	return diagnostics
}

func queryOperatorParameterKindIsValid(expression Expression, declaration *QueryOperatorParameter) bool {
	if declaration == nil {
		return true
	}
	typ := ExpressionResultType(expression)
	valid := true
	valueKind := declaration.ValueKind()
	values := declaration.Values()
	switch valueKind {
	case QueryOperatorParameterValueIntegerLiteral:
		valid = ExpressionIsLiteral(expression) && IsInteger(typ)
	case QueryOperatorParameterValueNumericLiteral, QueryOperatorParameterValueForcedRealLiteral:
		valid = ExpressionIsLiteral(expression) && IsNumeric(typ)
	case QueryOperatorParameterValueScalarLiteral:
		valid = ExpressionIsLiteral(expression) && queryOperatorParameterIsScalarType(typ)
	case QueryOperatorParameterValueSummableLiteral:
		valid = ExpressionIsLiteral(expression) && IsSummable(typ)
	case QueryOperatorParameterValueStringLiteral:
		valid = ExpressionIsLiteral(expression) && sameType(typ, ScalarTypes.String.Type())
	case QueryOperatorParameterValueBoolLiteral:
		valid = ExpressionIsLiteral(expression) && sameType(typ, ScalarTypes.Bool.Type())
	case QueryOperatorParameterValueString:
		// The upstream fallback checker accepts dynamic even though the fast-path
		// predicate only names string.
		valid = sameType(typ, ScalarTypes.String.Type()) || sameType(typ, ScalarTypes.Dynamic.Type())
	case QueryOperatorParameterValueColumn:
		var symbol Symbol
		if accessor, ok := expression.(interface{ ReferencedSymbol() Symbol }); ok {
			symbol = accessor.ReferencedSymbol()
		}
		_, valid = symbol.(*ColumnSymbol)
	case QueryOperatorParameterValueWord:
		valid = queryOperatorTokenLiteralIsAllowed(expression, declaration)
	case QueryOperatorParameterValueWordOrNumber:
		valid = IsNumeric(typ) || queryOperatorTokenLiteralIsAllowed(expression, declaration)
	}
	if valid && valueKind != QueryOperatorParameterValueWord &&
		valueKind != QueryOperatorParameterValueWordOrNumber && len(values) > 0 {
		valid = queryOperatorLiteralValueIsAllowed(expression, declaration)
	}
	return valid
}

func queryOperatorParameterIsScalarType(typ *TypeSymbol) bool {
	return typ != nil && typ != &Void.TypeSymbol && !isErrorTypeSymbol(typ) && IsScalar(typ)
}

func queryOperatorTokenLiteralIsAllowed(expression Expression, declaration *QueryOperatorParameter) bool {
	if expression == nil || expression.Kind() != SyntaxKindTokenLiteralExpression {
		return false
	}
	if len(declaration.Values()) == 0 {
		return true
	}
	return queryOperatorLiteralValueIsAllowed(expression, declaration)
}

func queryOperatorLiteralValueIsAllowed(expression Expression, declaration *QueryOperatorParameter) bool {
	if !ExpressionIsLiteral(expression) {
		return false
	}
	info := ExpressionLiteralValueInfo(expression)
	return info != nil && stringInParameterValues(info.ValueText, declaration.Values(), declaration.IsCaseSensitive())
}

func queryOperatorParameterNameLocation(declaration *NameDeclaration) SyntaxElement {
	if declaration == nil {
		return nil
	}
	if name, ok := declaration.Name().(*TokenName); ok && name.Name() != nil {
		return name.Name()
	}
	return declaration.SyntaxNode
}

func stringInParameterValues(value string, values []string, caseSensitive bool) bool {
	for _, candidate := range values {
		if caseSensitive && value == candidate || !caseSensitive && ordinalEqualFold(value, candidate) {
			return true
		}
	}
	return false
}

func semanticDiagnosticsContainCode(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func (b *simpleProjectionBinder) bindProjectRename(node *ProjectRenameOperator) *SemanticInfo {
	if node == nil || !renameExpressionsAreStructurallyPossible(node.Expressions()) {
		return nil
	}

	var diagnostics []Diagnostic
	if !b.hasPipedInput(node) {
		diagnostics = append(diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), node.ProjectRenameKeyword()))
	}
	builder := NewProjectionBuilder()
	if b.binder.rowScope != nil {
		builder.AddRange(b.binder.declaredAndInferredColumns(b.binder.rowScope), ProjectionAddRangeOptions{Declare: true, DoNotRepeat: true})
	}
	for _, separated := range node.Expressions().Elements() {
		expression := separated.Element()
		named, simple := expression.(*SimpleNamedExpression)
		if !simple || syntaxElementIsMissing(expression) {
			diagnostics = append(diagnostics, b.renameDiagnosticAt(DiagnosticRenameAssignmentExpected(), expression))
			continue
		}

		reference, referenceOK := named.Expression().(*NameReference)
		column, columnOK := referencedColumn(reference)
		if !referenceOK || !columnOK {
			diagnostics = append(diagnostics, b.renameDiagnosticAt(DiagnosticColumnExpected(), named.Expression()))
			continue
		}

		var renameDiagnostics []Diagnostic
		targetName := NameDeclarationSimpleName(named.Name())
		renamed := builder.Rename(column.SymbolName(), targetName, &renameDiagnostics, named.Name())
		for range renameDiagnostics {
			diagnostics = append(diagnostics, b.renameDiagnosticAt(DiagnosticDuplicateColumnDeclaration(targetName), named.Name()))
		}
		if renamed != nil && named.Name() != nil {
			named.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(renamed))
		}
	}

	table := NewTableSymbol("", builder.GetProjection()...).WithInheritableProperties(b.binder.rowScope)
	return semanticInfoPtr(NewSemanticInfoForSymbol(table, &table.TypeSymbol, diagnostics...))
}

func renameExpressionsAreStructurallyPossible(expressions *SeparatedSyntaxListOf[Expression]) bool {
	if expressions == nil {
		return false
	}
	for _, separated := range expressions.Elements() {
		if separated == nil || separated.Element() == nil {
			return false
		}
	}
	return true
}

func referencedColumn(reference *NameReference) (*ColumnSymbol, bool) {
	if reference == nil {
		return nil, false
	}
	column, ok := reference.ReferencedSymbol().(*ColumnSymbol)
	return column, ok
}

func (b *simpleProjectionBinder) renameDiagnosticAt(diagnostic Diagnostic, element SyntaxElement) Diagnostic {
	located := b.binder.diagnosticAtSyntaxElement(diagnostic, element)
	if located.LocationKind == DiagnosticLocationAbsolute || element == nil {
		return located
	}
	if token := firstProjectRenameToken(element); token != nil {
		if location, ok := b.binder.syntaxTokenLocations[token]; ok {
			end := location.End()
			if last := lastProjectRenameToken(element); last != nil {
				if lastLocation, found := b.binder.syntaxTokenLocations[last]; found {
					end = lastLocation.End()
				}
			}
			return diagnostic.WithByteLocation(location.Start, end-location.Start)
		}
	}
	return located
}

func firstProjectRenameToken(element SyntaxElement) *SyntaxToken {
	if token, ok := element.(*SyntaxToken); ok {
		return token
	}
	for i := 0; element != nil && i < element.ChildCount(); i++ {
		if token := firstProjectRenameToken(element.GetChild(i)); token != nil {
			return token
		}
	}
	return nil
}

func lastProjectRenameToken(element SyntaxElement) *SyntaxToken {
	if element == nil {
		return nil
	}
	if token, ok := element.(*SyntaxToken); ok {
		return token
	}
	for i := element.ChildCount() - 1; i >= 0; i-- {
		if token := lastProjectRenameToken(element.GetChild(i)); token != nil {
			return token
		}
	}
	return nil
}

func (b *simpleProjectionBinder) supportsExpressions(expressions *SeparatedSyntaxListOf[Expression], allowStar bool) bool {
	if expressions == nil {
		return false
	}
	for _, separated := range expressions.Elements() {
		if separated == nil || !b.supportsExpression(separated.Element(), allowStar) {
			return false
		}
	}
	return true
}

func (b *simpleProjectionBinder) supportsExpression(expression Expression, allowStar bool) bool {
	switch expression := expression.(type) {
	case *StarExpression:
		return allowStar
	case *NameReference:
		if isRecoverableMissingProjectionExpression(expression) {
			return true
		}
		switch expression.ReferencedSymbol().(type) {
		case *ColumnSymbol:
			return projectionExpressionIsStructurallySupported(expression)
		case *ErrorSymbol:
			return isRecoverableBareErrorProjectionReference(expression)
		default:
			return b.supportsUnnamedProjectionExpression(expression)
		}
	case *SimpleNamedExpression:
		return expression.Name() != nil && serializeNameDeclarationIsSupported(expression.Name()) &&
			expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			b.supportsNamedProjectionExpression(expression.Expression())
	case *CompoundNamedExpression:
		return compoundRenameListIsSupported(expression.Names()) && expression.EqualToken() != nil && expression.EqualToken().Width() != 0 &&
			expression.Expression() != nil && expressionIsBound(expression.Expression()) && compoundProjectionTypeIsSupported(expression.Expression())
	default:
		return b.supportsUnnamedProjectionExpression(expression)
	}
}

func (b *simpleProjectionBinder) supportsNamedProjectionExpression(expression Expression) bool {
	if reference, ok := expression.(*NameReference); ok && isRecoverableMissingProjectionExpression(reference) {
		return true
	}
	if !projectionExpressionIsStructurallySupported(expression) || !expressionIsBound(expression) {
		return false
	}
	typ := ExpressionRawResultType(expression)
	return typ != nil
}

func (b *simpleProjectionBinder) addProjectionScalarDiagnostic(expressions *SeparatedSyntaxListOf[Expression], diagnostics *[]Diagnostic) {
	if expressions == nil {
		return
	}
	for _, separated := range expressions.Elements() {
		if separated == nil {
			continue
		}
		expression := separated.Element()
		typ := projectionScalarCheckType(expression)
		if typ != nil && !isErrorTypeSymbol(typ) && !IsScalar(typ) {
			*diagnostics = append(*diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression))
			return
		}
	}
}

func projectionScalarCheckType(expression Expression) *TypeSymbol {
	if typ := ExpressionResultType(expression); typ != nil {
		return typ
	}
	switch expression := expression.(type) {
	case *SimpleNamedExpression:
		return ExpressionResultType(expression.Expression())
	case *CompoundNamedExpression:
		return ExpressionResultType(expression.Expression())
	default:
		return nil
	}
}

func (b *simpleProjectionBinder) supportsUnnamedProjectionExpression(expression Expression) bool {
	return projectionExpressionIsStructurallySupported(expression) && expressionIsBound(expression) && ExpressionRawResultType(expression) != nil
}

func projectionExpressionIsStructurallySupported(element SyntaxElement) bool {
	if element == nil {
		return false
	}
	if expression, ok := element.(*ElementExpression); ok {
		if selector, ok := expression.Selector().(*BracketedExpression); ok && parserRecoveredEmptyBracketedExpression(selector) {
			return projectionExpressionIsStructurallySupported(expression.Expression())
		}
	}
	if token, ok := element.(*SyntaxToken); ok {
		if token.Width() != 0 {
			return true
		}
		switch token.Kind() {
		case SyntaxKindCloseParenToken, SyntaxKindCloseBracketToken, SyntaxKindCloseBraceToken:
			return semanticDiagnosticsContainCode(token.ContainedDiagnostics(), "KS005")
		default:
			return false
		}
	}
	for i := 0; i < element.ChildCount(); i++ {
		child := element.GetChild(i)
		if child != nil && !projectionExpressionIsStructurallySupported(child) {
			return false
		}
	}
	return true
}

func (b *simpleProjectionBinder) addExpression(expression Expression, builder *ProjectionBuilder, diagnostics *[]Diagnostic, extend bool) {
	switch expression := expression.(type) {
	case *StarExpression:
		if b.binder.rowScope != nil {
			for _, column := range b.binder.rowScope.Columns() {
				builder.Add(column, ProjectionAddOptions{Replace: true})
			}
		}
		return
	case *NameReference:
		if isRecoverableMissingProjectionExpression(expression) {
			return
		}
		switch symbol := expression.ReferencedSymbol().(type) {
		case *ColumnSymbol:
			b.declareProjectionColumn(builder, symbol, diagnostics, expression)
			return
		case *ErrorSymbol:
			column := b.newProjectionColumn(NameReferenceSimpleName(expression), ExpressionRawResultType(expression), expression)
			b.declareProjectionColumn(builder, column, diagnostics, expression)
			return
		}
	case *SimpleNamedExpression:
		name := NameDeclarationSimpleName(expression.Name())
		if tuple, ok := TupleForType(ExpressionResultType(expression.Expression())); ok {
			b.addSimpleNamedTupleExpression(name, expression.Name(), expression.Expression(), tuple, builder, diagnostics)
			return
		}
		typ := ExpressionResultType(expression.Expression())
		column := b.newProjectionColumn(name, typ, expression.Expression())
		b.declareProjectionColumn(builder, column, diagnostics, expression.Name(), extend)
		if expression.Name() != nil {
			expression.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		}
		return
	case *CompoundNamedExpression:
		style := compoundProjectionDefault
		if extend {
			style = compoundProjectionExtend
		}
		b.addCompoundNamedExpression(expression, builder, diagnostics, style)
		return
	}

	if column := projectionResultColumn(expression); column != nil {
		b.declareProjectionColumn(builder, column, diagnostics, expression)
		return
	}
	if call, ok := expression.(*FunctionCallExpression); ok {
		if b.addFunctionTupleExpression(call, builder) {
			return
		}
	}
	typ := b.unnamedProjectionType(expression, nil)
	name := expressionResultName(expression, "", b.binder.rowScope, b.binder.Globals)
	if name == "" {
		name = b.nextDefaultColumnName(builder)
	}
	column := b.newProjectionColumn(name, typ, expression)
	builder.Add(column, ProjectionAddOptions{Replace: extend})
}

func (b *simpleProjectionBinder) addSimpleNamedTupleExpression(name string, declaration *NameDeclaration, expression Expression, tuple *TupleSymbol, builder *ProjectionBuilder, diagnostics *[]Diagnostic) {
	columns := tuple.Columns()
	if len(columns) == 0 || columns[0] == nil {
		return
	}
	first := columns[0]
	column := NewColumnSymbolFull(name, first.Type, "", []*ColumnSymbol{first}, first.Source, nil)
	b.declareProjectionColumn(builder, column, diagnostics, declaration, true)
	declaration.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))

	function, _ := expressionReferencedSymbol(expression).(*FunctionSymbol)
	for _, source := range columns[1:] {
		b.addFunctionTupleResultColumn(function, source, builder)
	}
}

func (b *simpleProjectionBinder) addFunctionTupleExpression(call *FunctionCallExpression, builder *ProjectionBuilder) bool {
	tuple, tupleOK := TupleForType(ExpressionResultType(call))
	function, functionOK := expressionReferencedSymbol(call).(*FunctionSymbol)
	if !tupleOK || tuple == nil || !functionOK || function == nil {
		return false
	}
	for _, column := range tuple.Columns() {
		b.addFunctionTupleResultColumn(function, column, builder)
	}
	return true
}

func (b *simpleProjectionBinder) declareProjectionColumn(builder *ProjectionBuilder, column *ColumnSymbol, diagnostics *[]Diagnostic, location SyntaxElement, replace ...bool) {
	var local []Diagnostic
	builder.Declare(column, &local, location, replace...)
	if diagnostics == nil {
		return
	}
	for range local {
		*diagnostics = append(*diagnostics, b.binder.diagnosticAtSyntaxElement(DiagnosticDuplicateColumnDeclaration(column.SymbolName()), location))
	}
}

func (b *simpleProjectionBinder) addFunctionTupleResultColumn(function *FunctionSymbol, column *ColumnSymbol, builder *ProjectionBuilder) {
	if column == nil || !builder.CanAdd(column) {
		return
	}
	if function != nil && function.ResultNamePrefix() != "" {
		column = column.WithName(function.ResultNamePrefix() + "_" + column.SymbolName())
	}
	builder.Add(column)
}

func (b *simpleProjectionBinder) unnamedProjectionType(expression Expression, diagnostics *[]Diagnostic) *TypeSymbol {
	b.normalizeUnnamedProjectionFunctionDiagnostics(expression)
	typ := ExpressionResultType(expression)
	if typ == nil {
		return ScalarTypes.Unknown.Type()
	}
	if !isErrorTypeSymbol(typ) && !expressionResultIsScalar(expression) {
		if diagnostics != nil {
			diagnostic := b.binder.diagnosticAtSyntaxElement(DiagnosticScalarTypeExpected(), expression)
			duplicate := false
			for _, existing := range *diagnostics {
				if existing.Code == diagnostic.Code && existing.Start == diagnostic.Start && existing.Length == diagnostic.Length {
					duplicate = true
					break
				}
			}
			if !duplicate {
				*diagnostics = append(*diagnostics, diagnostic)
			}
		}
		return ScalarTypes.Unknown.Type()
	}
	return typ
}

func (b *simpleProjectionBinder) normalizeUnnamedProjectionFunctionDiagnostics(expression Expression) {
	WalkSyntax(expression, func(element SyntaxElement) {
		call, ok := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !ok || call.Name() == nil {
			return
		}
		if _, unresolved := call.Name().ReferencedSymbol().(*ErrorSymbol); !unresolved ||
			!semanticDiagnosticsContainCode(call.Name().SemanticDiagnostics(), "KS142") {
			return
		}
		name := NameReferenceSimpleName(call.Name())
		diagnostic := b.binder.diagnosticAtSyntaxElement(DiagnosticNameDoesNotReferToAnyKnownFunction(name), call.Name())
		call.Name().SetSemanticInfo(NewSemanticInfoForReferencedSymbol(Error, diagnostic))
	})
}

func (b *simpleProjectionBinder) newProjectionColumn(name string, typ *TypeSymbol, source SyntaxElement) *ColumnSymbol {
	return NewColumnSymbolWithSource(name, typ, source)
}

func (b *simpleProjectionBinder) newSummarizeProjectionColumn(name string, typ *TypeSymbol, source SyntaxElement) *ColumnSymbol {
	return b.newProjectionColumn(name, typ, source)
}

func (b *simpleProjectionBinder) newSummarizeProjectionColumnFull(name string, typ *TypeSymbol, source SyntaxElement, original *ColumnSymbol) *ColumnSymbol {
	return NewColumnSymbolFull(name, typ, "", []*ColumnSymbol{original}, source, nil)
}

func summarizeFunctionSourceHasError(source SyntaxElement) bool {
	switch expression := source.(type) {
	case *ParenthesizedExpression:
		return summarizeFunctionSourceHasError(expression.Expression())
	case *FunctionCallExpression:
		return expression.Name() != nil && expression.Name().ReferencedSymbol() == Error || isErrorTypeSymbol(ExpressionResultType(expression))
	default:
		return false
	}
}

func isRecoverableBareErrorProjectionReference(reference *NameReference) bool {
	if reference == nil || !isCompleteProjectByNamesReference(reference) || len(reference.ContainedDiagnostics()) != 0 || !expressionIsBound(reference) {
		return false
	}
	if _, unresolved := reference.ReferencedSymbol().(*ErrorSymbol); !unresolved {
		return false
	}
	return isErrorTypeSymbol(ExpressionRawResultType(reference))
}

func isRecoverableMissingProjectionExpression(reference *NameReference) bool {
	return reference != nil && reference.Width() == 0 && expressionIsBound(reference) &&
		semanticDiagnosticsContainCode(reference.ContainedDiagnostics(), "KS006") &&
		isErrorTypeSymbol(ExpressionRawResultType(reference))
}

func compoundRenameListIsSupported(list *RenameList) bool {
	if list == nil || list.OpenParen() == nil || list.OpenParen().Width() == 0 || list.CloseParen() == nil || list.CloseParen().Width() == 0 ||
		list.Names() == nil || list.Names().Count() == 0 {
		return false
	}
	for _, separated := range list.Names().Elements() {
		if separated == nil || separated.Element() == nil || NameDeclarationSimpleName(separated.Element()) == "" {
			return false
		}
	}
	return true
}

func compoundProjectionTypeIsSupported(expression Expression) bool {
	if expression == nil {
		return false
	}
	if reference, ok := expression.(*NameReference); ok && isRecoverableMissingProjectionExpression(reference) {
		return true
	}
	typ := ExpressionRawResultType(expression)
	if typ == nil || typ == &Error.TypeSymbol {
		return false
	}
	return true
}

func (b *simpleProjectionBinder) addCompoundNamedExpression(expression *CompoundNamedExpression, builder *ProjectionBuilder, diagnostics *[]Diagnostic, style compoundProjectionStyle) bool {
	if expression == nil || expression.Names() == nil || expression.Names().Names() == nil || expression.Expression() == nil {
		return false
	}
	names := expression.Names().Names()
	tuple, tupleOK := TupleForType(ExpressionRawResultType(expression.Expression()))
	if tupleOK && tuple != nil {
		columns := tuple.Columns()
		for i, source := range columns {
			if source == nil {
				continue
			}
			if i < names.Count() {
				declaration := names.At(i).Element()
				column := NewColumnSymbolFull(NameDeclarationSimpleName(declaration), source.Type, "", source.OriginalColumns, source.Source, nil)
				b.declareProjectionColumn(builder, column, diagnostics, declaration, style == compoundProjectionExtend)
				declaration.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
				continue
			}
			if style == compoundProjectionPrint {
				continue
			}
			column := source
			replace := style == compoundProjectionExtend
			if function, ok := expressionReferencedSymbol(expression.Expression()).(*FunctionSymbol); ok {
				replace = false
				if function.ResultNamePrefix() != "" {
					column = source.WithName(function.ResultNamePrefix() + "_" + source.SymbolName())
				}
			}
			builder.Add(column, ProjectionAddOptions{Replace: replace})
		}
		for i := len(columns); i < names.Count(); i++ {
			declaration := names.At(i).Element()
			*diagnostics = append(*diagnostics, b.renameDiagnosticAt(DiagnosticNameDoesNotHaveCorrespondingExpression(), declaration))
		}
		return true
	}

	if names.Count() == 1 {
		declaration := names.At(0).Element()
		name := NameDeclarationSimpleName(declaration)
		var column *ColumnSymbol
		if source, ok := expressionReferencedSymbol(expression.Expression()).(*ColumnSymbol); ok {
			column = NewColumnSymbolFull(name, source.Type, "", []*ColumnSymbol{source}, nil, nil)
			b.declareProjectionColumn(builder, column, diagnostics, declaration, true)
		} else {
			column = NewColumnSymbolWithSource(name, resultTypeOrError(expression.Expression()), expression.Expression())
			b.declareProjectionColumn(builder, column, diagnostics, declaration, style == compoundProjectionExtend)
		}
		declaration.SetSemanticInfo(NewSemanticInfoForReferencedSymbol(column))
		return true
	}

	*diagnostics = append(*diagnostics, b.renameDiagnosticAt(DiagnosticExpressionDoesNotHaveMultipleValues(), expression.Names()))
	return true
}

func (b *simpleProjectionBinder) notFirstDiagnostics(operator SyntaxElement) []Diagnostic {
	if !b.hasPipedInput(operator) {
		location := operator
		if operator != nil && operator.ChildCount() != 0 && operator.GetChild(0) != nil {
			location = operator.GetChild(0)
		}
		return []Diagnostic{b.binder.diagnosticAtSyntaxElement(DiagnosticQueryOperatorCannotBeFirst(), location)}
	}
	return nil
}

func (b *simpleProjectionBinder) hasPipedInput(operator SyntaxElement) bool {
	queryOperator, _ := WrapGeneratedSyntaxElement(operator).(QueryOperator)
	return (b.inPipe != nil && *b.inPipe) || HasPipedInput(queryOperator)
}
