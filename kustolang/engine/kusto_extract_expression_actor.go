package kustolang

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Ported from:
//   src/Kusto.Language/Editor/Kusto/Actors/ExtractValueToLetStatementActor.cs
//   src/Kusto.Language/Editor/Kusto/Actors/ActorUtilities.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	ExtractValueKind      = "Extract Value"
	ExtractExpressionKind = "Extract Expression"
	ExtractFunctionKind   = "Extract Function"
)

// ExtractExpressionActor exposes the upstream extraction actor independently
// of KustoCodeService dispatch. Native methods use the package's byte offsets;
// the UTF16 variants translate the public editor boundary explicitly.
type ExtractExpressionActor struct{}

func NewExtractExpressionActor() *ExtractExpressionActor {
	return &ExtractExpressionActor{}
}

func (a *ExtractExpressionActor) GetActions(code *KustoCode, position, selectionStart, selectionLength int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionInfo {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || code.Syntax() == nil || position < 0 || position > len(code.Text) || selectionStart < 0 || selectionLength < 0 || selectionStart+selectionLength > len(code.Text) {
		return CodeActionInfoNoActions
	}
	if !code.HasSemantics() {
		code = code.Analyze()
	}
	if code == nil || !code.HasSemantics() {
		return CodeActionInfoNoActions
	}

	actions := make([]CodeAction, 0, 2)
	if selectionLength > 0 {
		if a.canExtractExpression(code, selectionStart, selectionLength, cancellationToken) {
			actions = append(actions, NewCodeAction(ExtractExpressionKind, ExtractExpressionKind, "Extract expression into new let statement variable declaration.", []string{
				strconv.Itoa(selectionStart), strconv.Itoa(selectionLength),
			}))
		}
		if a.canExtractFunction(code, selectionStart, selectionLength, cancellationToken) {
			actions = append(actions, NewCodeAction(ExtractFunctionKind, ExtractFunctionKind, "Extract expression into new let statement function declaration.", []string{
				strconv.Itoa(selectionStart), strconv.Itoa(selectionLength),
			}))
		}
	} else if a.canExtractValue(code, position) {
		actions = append(actions, NewCodeAction(ExtractValueKind, ExtractValueKind, "Extract value into new let statement variable declaration.", []string{
			strconv.Itoa(position), "0",
		}))
	}
	if len(actions) == 0 {
		return CodeActionInfoNoActions
	}
	return NewCodeActionInfo(actions)
}

func (a *ExtractExpressionActor) ApplyAction(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || code.Syntax() == nil || action == nil {
		return NewCodeActionFailure("Bad action data")
	}
	if !code.HasSemantics() {
		code = code.Analyze()
	}
	data := action.ApplyData()
	if len(data) != 2 {
		return NewCodeActionFailure("Bad action data")
	}
	position, positionOK := parseActorPosition(data[0], len(code.Text))
	length, lengthOK := parseActorLength(data[1], len(code.Text)-position)
	if !positionOK || !lengthOK {
		return NewCodeActionFailure("Bad action data")
	}

	switch action.Base().Kind {
	case ExtractExpressionKind:
		return a.expressionResult(code, position, length, cancellationToken)
	case ExtractFunctionKind:
		return a.functionResult(code, position, length, options, cancellationToken)
	case ExtractValueKind:
		return a.valueResult(code, position, cancellationToken)
	default:
		return NewCodeActionFailure("Unknown action")
	}
}

func (a *ExtractExpressionActor) GetActionsUTF16(code *KustoCode, position, selectionStart, selectionLength int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionInfo {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil {
		return CodeActionInfoNoActions
	}
	bytePosition, ok := utf16RangeStartToByte(code.Text, position)
	if !ok {
		return CodeActionInfoNoActions
	}
	byteSelectionStart, byteSelectionLength, ok := utf16RangeToBytes(code.Text, selectionStart, selectionLength)
	if !ok {
		return CodeActionInfoNoActions
	}
	info := a.GetActions(code, bytePosition, byteSelectionStart, byteSelectionLength, options, cancellationToken)
	actions := make([]CodeAction, 0, len(info.Actions))
	for _, action := range info.Actions {
		apply, ok := action.(ApplyAction)
		if !ok || len(apply.ApplyData()) != 2 {
			continue
		}
		data := apply.ApplyData()
		start, _ := strconv.Atoi(data[0])
		length, _ := strconv.Atoi(data[1])
		utf16Start := byteOffsetToUTF16Position(code.Text, start)
		utf16Length := byteOffsetToUTF16Position(code.Text, start+length) - utf16Start
		actions = append(actions, apply.WithDataStrings(strconv.Itoa(utf16Start), strconv.Itoa(utf16Length)))
	}
	return NewCodeActionInfo(actions)
}

func (a *ExtractExpressionActor) ApplyActionUTF16(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || action == nil || len(action.ApplyData()) != 2 {
		return NewCodeActionFailure("Bad action data")
	}
	data := action.ApplyData()
	start, errStart := strconv.Atoi(data[0])
	length, errLength := strconv.Atoi(data[1])
	byteStart, byteLength, ok := utf16RangeToBytes(code.Text, start, length)
	byteCaret, caretOK := utf16RangeStartToByte(code.Text, caretPosition)
	if errStart != nil || errLength != nil || !ok || !caretOK {
		return NewCodeActionFailure("Bad action data")
	}
	result := a.ApplyAction(code, action.WithDataStrings(strconv.Itoa(byteStart), strconv.Itoa(byteLength)), byteCaret, options, cancellationToken)
	return extractionResultToUTF16(code.Text, result)
}

func (a *ExtractExpressionActor) canExtractValue(code *KustoCode, position int) bool {
	token := extractTokenNear(code, position)
	if token == nil || position < token.ByteTextStart() || position > token.ByteEnd() {
		return false
	}
	literal, ok := WrapGeneratedSyntaxElement(token.Parent()).(*LiteralExpression)
	return ok && len(literal.ContainedDiagnostics()) == 0 && !extractIsLetExpression(literal)
}

func (a *ExtractExpressionActor) canExtractExpression(code *KustoCode, position, length int, cancellationToken CancellationToken) bool {
	position, length, ok := extractTrimRangeTrivia(code, position, length)
	if !ok {
		return false
	}
	expression := extractExpressionInRange(code, position, length)
	if expression == nil || expression.ByteTextStart() < position || expression.ByteEnd() > position+length || len(expression.ContainedDiagnostics()) > 0 {
		return false
	}
	if extractIsLetExpression(expression) || !extractIsFreeStandingExpression(expression) {
		return false
	}
	return len(extractCandidateParameters(code, position, length, cancellationToken)) == 0
}

func (a *ExtractExpressionActor) canExtractFunction(code *KustoCode, position, length int, cancellationToken CancellationToken) bool {
	position, length, ok := extractTrimRangeTrivia(code, position, length)
	if !ok {
		return false
	}
	expression := extractExpressionInRange(code, position, length)
	if expression == nil || len(expression.ContainedDiagnostics()) > 0 {
		return false
	}
	if _, pipe := expression.(*PipeExpression); pipe {
		_, _, adjusted := extractAdjustedSubqueryRange(code, position, length)
		return adjusted && len(extractColumnsInScope(code, position)) > 0
	}
	if _, queryOperator := expression.(QueryOperator); queryOperator && len(extractColumnsInScope(code, position)) > 0 {
		return true
	}
	typ := ExpressionResultType(expression)
	return !extractIsLetExpression(expression) && typ != nil && typ.SymbolTabularity() == TabularityScalar && len(extractCandidateParameters(code, position, length, cancellationToken)) > 0
}

func (a *ExtractExpressionActor) valueResult(code *KustoCode, position int, cancellationToken CancellationToken) CodeActionResult {
	token := extractTokenNear(code, position)
	if token == nil || position < token.ByteTextStart() || position > token.ByteEnd() {
		return NewCodeActionFailure("Invalid literal value")
	}
	literal, ok := WrapGeneratedSyntaxElement(token.Parent()).(*LiteralExpression)
	if !ok || len(literal.ContainedDiagnostics()) > 0 || extractIsLetExpression(literal) {
		return NewCodeActionFailure("Invalid literal value")
	}
	return a.expressionResult(code, token.ByteTextStart(), token.ByteWidth(), cancellationToken)
}

func (a *ExtractExpressionActor) expressionResult(code *KustoCode, position, length int, cancellationToken CancellationToken) CodeActionResult {
	position, length, ok := extractTrimRangeTrivia(code, position, length)
	if !ok {
		return NewCodeActionFailure("Invalid selection range")
	}
	expression := extractExpressionInRange(code, position, length)
	if expression == nil || expression.ByteTextStart() < position || expression.ByteEnd() > position+length || len(expression.ContainedDiagnostics()) > 0 {
		return NewCodeActionFailure("Invalid selection range")
	}
	insertionPosition, ok := extractNearestStatementInsertionPosition(code, position)
	if !ok {
		return NewCodeActionFailure("No insertion location available.")
	}
	cancellationToken.ThrowIfCancellationRequested()
	baseName := "varName"
	if _, pipe := expression.(*PipeExpression); pipe || ExpressionResultType(expression) != nil && ExpressionResultType(expression).SymbolTabularity() == TabularityTabular {
		baseName = "queryName"
	}
	newName := extractNewName(code, position, baseName)
	indentation := extractWhitespaceAt(code.Text, insertionPosition)
	expressionLength := expression.ByteEnd() - expression.ByteTextStart()
	expressionText := code.Text[expression.ByteTextStart():expression.ByteEnd()]
	letStatement := indentation + "let " + newName + " = " + expressionText + ";\n"
	newText := NewEditString(code.Text).ReplaceAt(expression.ByteTextStart(), expressionLength, newName).Insert(insertionPosition, letStatement)
	newPosition := insertionPosition + len(indentation) + len("let ")
	return NewCodeActionChangeAndMoveEditString(newText, newPosition).WithAction(RenameActionInstance)
}

func (a *ExtractExpressionActor) functionResult(code *KustoCode, position, length int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	position, length, ok := extractTrimRangeTrivia(code, position, length)
	if !ok {
		return NewCodeActionFailure("Invalid selection range")
	}
	expression := extractExpressionInRange(code, position, length)
	if expression == nil || len(expression.ContainedDiagnostics()) > 0 {
		return NewCodeActionFailure("Invalid selection range")
	}
	typ := ExpressionResultType(expression)
	if _, pipe := expression.(*PipeExpression); pipe || typ != nil && typ.SymbolTabularity() == TabularityTabular {
		return a.tabularFunctionResult(code, position, length, options, cancellationToken)
	}
	if typ != nil && typ.SymbolTabularity() == TabularityScalar {
		return a.scalarFunctionResult(code, position, length, cancellationToken)
	}
	return NewCodeActionFailure("Invalid selection range")
}

func (a *ExtractExpressionActor) scalarFunctionResult(code *KustoCode, position, length int, cancellationToken CancellationToken) CodeActionResult {
	insertionPosition, ok := extractNearestStatementInsertionPosition(code, position)
	if !ok {
		return NewCodeActionFailure("No insertion location available.")
	}
	parameters := extractCandidateParameters(code, position, length, cancellationToken)
	parameterList := extractSchema(parameters)
	arguments := make([]string, 0, len(parameters))
	for _, parameter := range parameters {
		arguments = append(arguments, BracketNameIfNecessary(parameter.SymbolName()))
	}
	newName := extractNewName(code, position, "functionName")
	indentation := extractWhitespaceAt(code.Text, insertionPosition)
	letStatement := indentation + "let " + newName + " = " + parameterList + " { " + code.Text[position:position+length] + " };\n"
	invocation := newName + "(" + strings.Join(arguments, ", ") + ")"
	newText := NewEditString(code.Text).ReplaceAt(position, length, invocation).Insert(insertionPosition, letStatement)
	newPosition := insertionPosition + len(indentation) + len("let ")
	return NewCodeActionChangeAndMoveEditString(newText, newPosition).WithAction(RenameActionInstance)
}

func (a *ExtractExpressionActor) tabularFunctionResult(code *KustoCode, position, length int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	position, length, ok := extractAdjustedSubqueryRange(code, position, length)
	if !ok {
		return NewCodeActionFailure("No insertion location available.")
	}
	insertionPosition, ok := extractNearestStatementInsertionPosition(code, position)
	if !ok {
		return NewCodeActionFailure("No insertion location available.")
	}
	parameters := extractCandidateParameters(code, position, length, cancellationToken)
	tableType := "(*)"
	if len(parameters) > 0 {
		tableType = extractSchema(parameters)
	}
	tableName := extractNewName(code, position, "inputTable")
	functionName := extractNewName(code, position, "functionName")
	lineIndentation := extractWhitespaceAt(code.Text, insertionPosition)
	nestedIndentation := lineIndentation + strings.Repeat(" ", maxInt(0, options.FormattingOptions.IndentationSize))
	insertionText := extractChangeIndentation(code.Text[position:position+length], nestedIndentation, false)
	letStatement := lineIndentation + "let " + functionName + " = (" + tableName + ": " + tableType + ") {\n" + nestedIndentation + tableName + "\n" + nestedIndentation + "| " + insertionText + "\n" + lineIndentation + "};\n"
	newText := NewEditString(code.Text).ReplaceAt(position, length, "invoke "+functionName+"()").Insert(insertionPosition, letStatement)
	newPosition := insertionPosition + len(lineIndentation) + len("let ")
	return NewCodeActionChangeAndMoveEditString(newText, newPosition).WithAction(RenameActionInstance)
}

func extractTokenNear(code *KustoCode, position int) *SyntaxToken {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	token := code.Syntax().GetTokenAtByte(position)
	if token != nil && position == token.ByteTriviaStart() && token.ByteTriviaWidth() > 0 {
		if previous := token.GetPreviousToken(); previous != nil {
			return previous
		}
	}
	if token != nil && position == token.ByteTextStart() {
		if _, literal := WrapGeneratedSyntaxElement(token.Parent()).(*LiteralExpression); !literal {
			if previous := token.GetPreviousToken(); previous != nil && previous.ByteEnd() == position {
				return previous
			}
		}
	}
	return token
}

func extractTrimRangeTrivia(code *KustoCode, start, length int) (int, int, bool) {
	if code == nil || code.Syntax() == nil || start < 0 || length < 0 || start+length > len(code.Text) {
		return 0, 0, false
	}
	expectedEnd := start + length
	if endToken := code.Syntax().GetTokenAtByte(expectedEnd); endToken != nil && start < endToken.ByteTextStart() {
		length = endToken.ByteTriviaStart() - start
		expectedEnd = start + length
	}
	if startToken := code.Syntax().GetTokenAtByte(start); startToken != nil && start < startToken.ByteTextStart() {
		start = startToken.ByteTextStart()
		length = expectedEnd - start
	}
	return start, length, length >= 0
}

func extractExpressionInRange(code *KustoCode, start, length int) Expression {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	node := code.Syntax().GetNodeAtByte(start, length)
	if node == nil {
		return nil
	}
	expression, _ := WrapGeneratedSyntaxElement(node).(Expression)
	return expression
}

func extractIsLetExpression(expression Expression) bool {
	if expression == nil || expression.Parent() == nil {
		return false
	}
	statement, ok := WrapGeneratedSyntaxElement(expression.Parent()).(*LetStatement)
	return ok && extractSameElement(statement.Expression(), expression)
}

func extractIsFreeStandingExpression(expression Expression) bool {
	if expression == nil || expression.Parent() == nil {
		return true
	}
	switch parent := WrapGeneratedSyntaxElement(expression.Parent()).(type) {
	case *PathExpression:
		return !extractSameElement(parent.Selector(), expression)
	case *PipeExpression:
		return !extractSameElement(parent.Operator(), expression)
	default:
		return true
	}
}

func extractSameElement(left, right SyntaxElement) bool {
	if left == nil || right == nil {
		return left == right
	}
	leftInfo, rightInfo := syntaxOwnershipOf(left), syntaxOwnershipOf(right)
	return leftInfo != nil && rightInfo != nil && leftInfo.self == rightInfo.self
}

func extractNearestStatementInsertionPosition(code *KustoCode, position int) (int, bool) {
	token := extractTokenNear(code, position)
	if token == nil {
		return 0, false
	}
	for element := token.Parent(); element != nil; element = element.Parent() {
		wrapped := WrapGeneratedSyntaxElement(element)
		if _, statement := wrapped.(Statement); statement {
			return extractInsertionPositionBefore(element), true
		}
		if body, ok := WrapGeneratedSyntaxElement(element.Parent()).(*FunctionBody); ok && extractSameElement(body.Expression(), wrapped) {
			return extractInsertionPositionBefore(element), true
		}
	}
	return 0, false
}

func extractInsertionPositionBefore(node SyntaxElement) int {
	firstToken := node.GetFirstToken()
	if firstToken == nil {
		return node.ByteTriviaStart()
	}
	trivia := firstToken.Trivia()
	if node.ByteTriviaStart() == 0 {
		if end := LastLineBreakEnd(trivia); end >= 0 {
			return node.ByteTriviaStart() + end
		}
		return node.ByteTriviaStart()
	}
	if end := FirstLineBreakEnd(trivia); end >= 0 {
		return node.ByteTriviaStart() + end
	}
	return node.ByteTriviaStart()
}

func extractNewName(code *KustoCode, position int, baseName string) string {
	for suffix := 0; ; suffix++ {
		name := baseName
		if suffix > 0 {
			name += strconv.Itoa(suffix)
		}
		if code.GetSpeculativeReferencedSymbol(position, name, SymbolMatchAny) == nil {
			return name
		}
	}
}

func extractCandidateParameters(code *KustoCode, start, length int, cancellationToken CancellationToken) []*ColumnSymbol {
	if code == nil || !code.HasSemantics() {
		return nil
	}
	inScope := make(map[Symbol]struct{})
	for _, symbol := range code.GetSymbolsInScope(start, SymbolMatchAny, IncludeFunctionKindAll) {
		inScope[symbol] = struct{}{}
	}
	external := make([]*ColumnSymbol, 0)
	seen := make(map[*ColumnSymbol]struct{})
	end := start + length
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, triviaStart int) {
		if triviaStart > end || triviaStart+element.ByteFullWidth() < start {
			return
		}
		cancellationToken.ThrowIfCancellationRequested()
		expression, ok := WrapGeneratedSyntaxElement(element).(Expression)
		if !ok {
			return
		}
		column, ok := expressionReferencedSymbol(expression).(*ColumnSymbol)
		if !ok {
			return
		}
		if _, visible := inScope[column]; !visible {
			return
		}
		if _, duplicate := seen[column]; !duplicate {
			seen[column] = struct{}{}
			external = append(external, column)
		}
	})
	if len(external) == 0 {
		return nil
	}
	insertionPosition, ok := extractNearestStatementInsertionPosition(code, start)
	if !ok {
		return nil
	}
	atInsertion := make(map[Symbol]struct{})
	for _, symbol := range code.GetSymbolsInScope(insertionPosition, SymbolMatchAny, IncludeFunctionKindAll) {
		atInsertion[symbol] = struct{}{}
	}
	remaining := external[:0]
	for _, column := range external {
		if _, alreadyAvailable := atInsertion[column]; !alreadyAvailable {
			remaining = append(remaining, column)
		}
	}
	return remaining
}

func extractColumnsInScope(code *KustoCode, position int) []*ColumnSymbol {
	table := code.GetColumnsInScope(position)
	if table == nil {
		return nil
	}
	return table.Columns()
}

func extractSchema(columns []*ColumnSymbol) string {
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		parts = append(parts, BracketNameIfNecessary(column.SymbolName())+": "+SchemaDisplayScalarTypeText(column.Type))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func extractAdjustedSubqueryRange(code *KustoCode, start, length int) (int, int, bool) {
	start, length, ok := extractTrimRangeTrivia(code, start, length)
	if !ok {
		return 0, 0, false
	}
	expression := extractExpressionInRange(code, start, length)
	pipe, isPipe := expression.(*PipeExpression)
	if expression == nil {
		return 0, 0, false
	}
	if !isPipe {
		return start, length, true
	}
	first := extractPipePartAtStart(pipe, start)
	if first == nil {
		return 0, 0, false
	}
	end := start + length
	if first.ByteTextStart() > start {
		start = first.ByteTextStart()
		length = end - start
	}
	last := extractPipePartAtEnd(pipe, start+length)
	if last == nil {
		return 0, 0, false
	}
	if last.ByteEnd() != start+length {
		length = last.ByteEnd() - start
	}
	return start, length, true
}

func extractPipePartAtStart(pipe *PipeExpression, start int) Expression {
	for pipe != nil {
		left, operator := pipe.Expression(), pipe.Operator()
		if left == nil || operator == nil {
			return nil
		}
		if start >= left.ByteEnd() && start <= operator.ByteTextStart() {
			expression, ok := operator.(Expression)
			if !ok {
				return nil
			}
			return expression
		}
		if start <= left.ByteEnd() {
			if nested, ok := left.(*PipeExpression); ok {
				pipe = nested
				continue
			}
		}
		if start >= left.ByteTriviaStart() && start <= left.ByteTextStart() {
			return left
		}
		return nil
	}
	return nil
}

func extractPipePartAtEnd(pipe *PipeExpression, end int) Expression {
	for pipe != nil {
		left, operator := pipe.Expression(), pipe.Operator()
		if left == nil || operator == nil {
			return nil
		}
		if end == pipe.ByteEnd() {
			expression, ok := operator.(Expression)
			if !ok {
				return nil
			}
			return expression
		}
		if end < operator.ByteTextStart() && end >= left.ByteEnd() {
			return left
		}
		if nested, ok := left.(*PipeExpression); ok {
			pipe = nested
			continue
		}
		return nil
	}
	return nil
}

func extractWhitespaceAt(text string, position int) string {
	if position < 0 || position > len(text) {
		return ""
	}
	end := position
	for end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		if r == '\r' || r == '\n' || !IsWhitespace(r) {
			break
		}
		end += size
	}
	return text[position:end]
}

func extractChangeIndentation(text, indentation string, includeFirstLine bool) string {
	var builder strings.Builder
	lineStart := 0
	if !includeFirstLine {
		lineLength := LineLength(text, lineStart, true)
		builder.WriteString(text[:lineLength])
		lineStart += lineLength
	}
	for lineStart < len(text) {
		lineLength := LineLength(text, lineStart, true)
		whitespace := ScanWhitespace(text, lineStart)
		builder.WriteString(indentation)
		builder.WriteString(text[lineStart+whitespace : lineStart+lineLength])
		lineStart += lineLength
	}
	return builder.String()
}

func parseActorPosition(text string, maximum int) (int, bool) {
	value, err := strconv.ParseInt(text, 10, 32)
	return int(value), err == nil && value >= 0 && value <= int64(maximum)
}

func parseActorLength(text string, maximum int) (int, bool) {
	value, err := strconv.ParseInt(text, 10, 32)
	return int(value), err == nil && value >= 0 && value <= int64(maximum)
}

func utf16RangeStartToByte(text string, position int) (int, bool) {
	if position < 0 {
		return 0, false
	}
	units := 0
	for offset := 0; offset < len(text); {
		if units == position {
			return offset, true
		}
		r, size := utf8.DecodeRuneInString(text[offset:])
		width := 1
		if r > 0xffff {
			width = 2
		}
		if units+width > position {
			return 0, false
		}
		units += width
		offset += size
	}
	return len(text), units == position
}

func utf16RangeToBytes(text string, start, length int) (int, int, bool) {
	if length < 0 {
		return 0, 0, false
	}
	byteStart, ok := utf16RangeStartToByte(text, start)
	if !ok {
		return 0, 0, false
	}
	byteEnd, ok := utf16RangeStartToByte(text, start+length)
	if !ok || byteEnd < byteStart {
		return 0, 0, false
	}
	return byteStart, byteEnd - byteStart, true
}

func extractionResultToUTF16(originalText string, result CodeActionResult) CodeActionResult {
	if result.HasFailureReason() || len(result.Actions) == 0 {
		return result
	}
	changedText := originalText
	for _, resultAction := range result.Actions {
		switch action := resultAction.(type) {
		case ChangeTextAction:
			changedText = action.ChangedText
		case *ChangeTextAction:
			if action != nil {
				changedText = action.ChangedText
			}
		}
	}
	converted := make([]ResultAction, 0, len(result.Actions))
	for _, resultAction := range result.Actions {
		switch action := resultAction.(type) {
		case ChangeTextAction:
			converted = append(converted, extractChangeTextToUTF16(originalText, action))
		case *ChangeTextAction:
			if action == nil {
				converted = append(converted, resultAction)
			} else {
				converted = append(converted, extractChangeTextToUTF16(originalText, *action))
			}
		case MoveCaretAction:
			converted = append(converted, NewMoveCaretAction(byteOffsetToUTF16Position(changedText, action.NewCaretPosition)))
		case *MoveCaretAction:
			if action == nil {
				converted = append(converted, resultAction)
			} else {
				converted = append(converted, NewMoveCaretAction(byteOffsetToUTF16Position(changedText, action.NewCaretPosition)))
			}
		default:
			converted = append(converted, resultAction)
		}
	}
	return NewCodeActionResultActions(converted)
}

func extractChangeTextToUTF16(originalText string, action ChangeTextAction) ChangeTextAction {
	edits := make([]TextEdit, 0, len(action.Changes))
	for _, edit := range action.Changes {
		start := byteOffsetToUTF16Position(originalText, edit.Start)
		end := byteOffsetToUTF16Position(originalText, edit.Start+edit.DeleteLength)
		edits = append(edits, NewTextReplacement(start, end-start, edit.InsertText))
	}
	return NewChangeTextAction(edits, action.ChangedText)
}
