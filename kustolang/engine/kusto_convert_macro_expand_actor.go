package kustolang

import (
	"strconv"
	"strings"
)

// Ported from:
//   src/Kusto.Language/Editor/Kusto/Actors/ConvertToMacroExpandActor.cs
//   src/Kusto.Language/Editor/Kusto/Actors/ActorUtilities.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	ConvertSelectionKind = "ConvertSelection"
	ConvertUnionKind     = "ConvertUnion"

	macroExpandInitialEntityName = "_scope"
)

// ConvertToMacroExpandActor exposes the upstream actor independently of
// KustoCodeService dispatch. Native methods use byte offsets; UTF16 variants
// translate the editor boundary explicitly.
type ConvertToMacroExpandActor struct{}

func NewConvertToMacroExpandActor() *ConvertToMacroExpandActor {
	return &ConvertToMacroExpandActor{}
}

func (a *ConvertToMacroExpandActor) GetActions(code *KustoCode, position, selectionStart, selectionLength int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionInfo {
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
	token := code.Syntax().GetTokenAtByte(position)
	if token == nil || position < token.ByteTextStart() || position > token.ByteEnd() || macroExpandInsideOperator(token) {
		return CodeActionInfoNoActions
	}

	actions := make([]CodeAction, 0, 2)
	if actorGetEntityExpressionStart(token, code.Globals) != nil {
		start, length := selectionStart, selectionLength
		if length == 0 {
			start, length = 0, len(code.Text)
		}
		actions = append(actions, NewCodeAction(
			ConvertSelectionKind,
			"Convert selection to macro-expand",
			"Move the selection into macro-expand body.",
			[]string{strconv.Itoa(start), strconv.Itoa(length), strconv.Itoa(position)},
		))
	}
	if unionQuery := macroExpandUnionQuery(token); unionQuery != nil {
		if _, ok := macroExpandUnionEntities(unionQuery, code.Globals, cancellationToken); ok {
			actions = append(actions, NewCodeAction(
				ConvertUnionKind,
				"Convert union to macro-expand",
				"Convert union inputs into entity group and move remaining query into a macro-expand expression.",
				[]string{strconv.Itoa(position)},
			))
		}
	}
	if len(actions) == 0 {
		return CodeActionInfoNoActions
	}
	return NewCodeActionInfo(actions)
}

func (a *ConvertToMacroExpandActor) ApplyAction(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || code.Syntax() == nil || action == nil {
		return NewCodeActionFailure("Bad action data")
	}
	if !code.HasSemantics() {
		code = code.Analyze()
	}
	if code == nil || !code.HasSemantics() {
		return NewCodeActionFailure("Analysis unavailable")
	}
	switch action.Base().Kind {
	case ConvertSelectionKind:
		return a.convertSelection(code, action, options, cancellationToken)
	case ConvertUnionKind:
		return a.convertUnion(code, action, options, cancellationToken)
	default:
		return NewCodeActionFailure("Unknown action kind: " + action.Base().Kind)
	}
}

func (a *ConvertToMacroExpandActor) GetActionsUTF16(code *KustoCode, position, selectionStart, selectionLength int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionInfo {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil {
		return CodeActionInfoNoActions
	}
	bytePosition, positionOK := utf16RangeStartToByte(code.Text, position)
	byteStart, byteLength, selectionOK := utf16RangeToBytes(code.Text, selectionStart, selectionLength)
	if !positionOK || !selectionOK {
		return CodeActionInfoNoActions
	}
	info := a.GetActions(code, bytePosition, byteStart, byteLength, options, cancellationToken)
	actions := make([]CodeAction, 0, len(info.Actions))
	for _, item := range info.Actions {
		apply, ok := item.(ApplyAction)
		if !ok {
			continue
		}
		data := apply.ApplyData()
		switch apply.Base().Kind {
		case ConvertSelectionKind:
			if len(data) != 3 {
				continue
			}
			start, _ := strconv.Atoi(data[0])
			length, _ := strconv.Atoi(data[1])
			pos, _ := strconv.Atoi(data[2])
			utfStart := byteOffsetToUTF16Position(code.Text, start)
			actions = append(actions, apply.WithDataStrings(
				strconv.Itoa(utfStart),
				strconv.Itoa(byteOffsetToUTF16Position(code.Text, start+length)-utfStart),
				strconv.Itoa(byteOffsetToUTF16Position(code.Text, pos)),
			))
		case ConvertUnionKind:
			if len(data) != 1 {
				continue
			}
			pos, _ := strconv.Atoi(data[0])
			actions = append(actions, apply.WithDataStrings(strconv.Itoa(byteOffsetToUTF16Position(code.Text, pos))))
		}
	}
	return NewCodeActionInfo(actions)
}

func (a *ConvertToMacroExpandActor) ApplyActionUTF16(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || action == nil {
		return NewCodeActionFailure("Bad action data")
	}
	data := action.ApplyData()
	byteData := make([]string, len(data))
	copy(byteData, data)
	switch action.Base().Kind {
	case ConvertSelectionKind:
		if len(data) != 3 {
			return NewCodeActionFailure("Bad action data")
		}
		start, startErr := strconv.Atoi(data[0])
		length, lengthErr := strconv.Atoi(data[1])
		position, positionErr := strconv.Atoi(data[2])
		byteStart, byteLength, rangeOK := utf16RangeToBytes(code.Text, start, length)
		bytePosition, positionOK := utf16RangeStartToByte(code.Text, position)
		if startErr != nil || lengthErr != nil || positionErr != nil || !rangeOK || !positionOK {
			return NewCodeActionFailure("Bad action data")
		}
		byteData = []string{strconv.Itoa(byteStart), strconv.Itoa(byteLength), strconv.Itoa(bytePosition)}
	case ConvertUnionKind:
		if len(data) != 1 {
			return NewCodeActionFailure("Bad action data")
		}
		position, err := strconv.Atoi(data[0])
		bytePosition, ok := utf16RangeStartToByte(code.Text, position)
		if err != nil || !ok {
			return NewCodeActionFailure("Bad action data")
		}
		byteData = []string{strconv.Itoa(bytePosition)}
	default:
		return NewCodeActionFailure("Unknown action kind: " + action.Base().Kind)
	}
	byteCaret, ok := utf16RangeStartToByte(code.Text, caretPosition)
	if !ok {
		return NewCodeActionFailure("Bad action data")
	}
	result := a.ApplyAction(code, action.WithDataStrings(byteData...), byteCaret, options, cancellationToken)
	return extractionResultToUTF16(code.Text, result)
}

func (a *ConvertToMacroExpandActor) convertSelection(code *KustoCode, action ApplyAction, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	data := action.ApplyData()
	if len(data) < 3 {
		return NewCodeActionFailure("Bad action data")
	}
	selectionStart, startOK := parseActorPosition(data[0], len(code.Text))
	selectionLength, lengthOK := parseActorLength(data[1], len(code.Text)-selectionStart)
	position, positionOK := parseActorPosition(data[2], len(code.Text))
	if !startOK || !lengthOK || !positionOK {
		return NewCodeActionFailure("Bad action data")
	}
	token := code.Syntax().GetTokenAtByte(position)
	entityStart := actorGetEntityExpressionStart(token, code.Globals)
	entityType := ExpressionResultType(entityStart)
	if token == nil || position < token.ByteTextStart() || position > token.ByteEnd() || entityStart == nil || entityType == nil || macroExpandInsideOperator(token) {
		return NewCodeActionFailure()
	}

	selectionLength = macroExpandTrimRangeEnd(code.Text, selectionStart, selectionLength)
	selectionStart, selectionLength = actorAdjustRangeToNode(code, selectionStart, selectionLength)
	if selectionLength <= 0 {
		return NewCodeActionFailure("Invalid selection range")
	}
	cancellationToken.ThrowIfCancellationRequested()
	original := NewEditString(code.Text)
	prefix := "macro-expand entity_group [" + entityStart.Text(IncludeTriviaInterior) + "] as " + macroExpandInitialEntityName
	withBrackets := actorInsertBrackets(original, selectionStart, prefix, "(", selectionStart+selectionLength, ")", options.FormattingOptions)

	references := macroExpandEntityExpressionStarts(code, entityType, cancellationToken)
	edits := make([]TextEdit, 0, len(references))
	for _, expression := range references {
		current := withBrackets.GetCurrentRange(expression.ByteTextStart(), expression.ByteWidth())
		edits = append(edits, NewTextReplacement(current.Start, current.Length, macroExpandInitialEntityName))
	}
	if len(edits) == 0 || !withBrackets.CanApplyAll(edits) {
		return NewCodeActionFailure("Entity references unavailable")
	}
	changed := withBrackets.ApplyAll(edits)
	return NewCodeActionChangeAndMoveEditString(changed, changed.IndexOf(macroExpandInitialEntityName)).WithAction(RenameActionInstance)
}

func (a *ConvertToMacroExpandActor) convertUnion(code *KustoCode, action ApplyAction, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	data := action.ApplyData()
	if len(data) < 1 {
		return NewCodeActionFailure("Bad action data")
	}
	position, ok := parseActorPosition(data[0], len(code.Text))
	if !ok {
		return NewCodeActionFailure("Bad action data")
	}
	token := code.Syntax().GetTokenAtByte(position)
	if token == nil || position < token.ByteTextStart() || position > token.ByteEnd() || macroExpandInsideOperator(token) {
		return NewCodeActionFailure()
	}
	unionQuery := macroExpandUnionQuery(token)
	entities, entitiesOK := macroExpandUnionEntities(unionQuery, code.Globals, cancellationToken)
	query := macroExpandContainingQuery(token)
	if !entitiesOK || query == nil || unionQuery == nil || len(unionQuery.ContainedDiagnostics()) > 0 {
		return NewCodeActionFailure()
	}

	entityText := make([]string, 0, len(entities))
	for _, entity := range entities {
		entityText = append(entityText, entity.Text(IncludeTriviaInterior))
	}
	prefix := "macro-expand entity_group [" + strings.Join(entityText, ", ") + "] as " + macroExpandInitialEntityName
	withBrackets := actorInsertBrackets(NewEditString(code.Text), query.ByteTextStart(), prefix, "(", query.ByteEnd(), ")", options.FormattingOptions)
	currentUnion := withBrackets.GetCurrentRange(unionQuery.ByteTextStart(), unionQuery.ByteEnd()-unionQuery.ByteTextStart())
	changed := withBrackets.ReplaceAt(currentUnion.Start, currentUnion.Length, macroExpandInitialEntityName)
	return NewCodeActionChangeAndMoveEditString(changed, changed.IndexOf(macroExpandInitialEntityName)).WithAction(RenameActionInstance)
}

func macroExpandInsideOperator(element SyntaxElement) bool {
	for current := element; current != nil; current = current.Parent() {
		if _, ok := WrapGeneratedSyntaxElement(current).(*MacroExpandOperator); ok {
			return true
		}
	}
	return false
}

func macroExpandEntityExpressionStarts(code *KustoCode, entityType *TypeSymbol, cancellationToken CancellationToken) []Expression {
	result := make([]Expression, 0)
	seen := make(map[*SyntaxNode]struct{})
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, _ int) {
		cancellationToken.ThrowIfCancellationRequested()
		expression, ok := WrapGeneratedSyntaxElement(element).(Expression)
		if !ok || ExpressionResultType(expression) != entityType {
			return
		}
		if path, pathOK := WrapGeneratedSyntaxElement(expression.Parent()).(*PathExpression); pathOK && extractSameElement(path.Selector(), expression) {
			return
		}
		start := actorGetEntityExpressionStart(expression, code.Globals)
		if start == nil {
			return
		}
		node := simpleBinderSyntaxNode(start)
		if node == nil {
			return
		}
		if _, duplicate := seen[node]; !duplicate {
			seen[node] = struct{}{}
			result = append(result, start)
		}
	})
	return result
}

func macroExpandUnionQuery(token *SyntaxToken) Expression {
	if token == nil {
		return nil
	}
	var query Expression
	for current := token.Parent(); current != nil; current = current.Parent() {
		wrapped := WrapGeneratedSyntaxElement(current)
		if union, ok := wrapped.(*UnionOperator); ok {
			query = union
			break
		}
	}
	if query == nil {
		for current := token.Parent(); current != nil; current = current.Parent() {
			pipe, ok := WrapGeneratedSyntaxElement(current).(*PipeExpression)
			if ok {
				if _, union := pipe.Operator().(*UnionOperator); union {
					query = pipe
					break
				}
			}
		}
	}
	for query != nil && query.Parent() != nil {
		pipe, ok := WrapGeneratedSyntaxElement(query.Parent()).(*PipeExpression)
		if !ok {
			break
		}
		if _, union := pipe.Operator().(*UnionOperator); !union {
			break
		}
		query = pipe
	}
	return query
}

func macroExpandContainingQuery(token *SyntaxToken) Expression {
	if token == nil {
		return nil
	}
	var query Expression
	for current := token.Parent(); current != nil; current = current.Parent() {
		wrapped := WrapGeneratedSyntaxElement(current)
		if _, ok := wrapped.(QueryOperator); ok {
			query, _ = wrapped.(Expression)
			break
		}
	}
	if query == nil {
		for current := token.Parent(); current != nil; current = current.Parent() {
			if pipe, ok := WrapGeneratedSyntaxElement(current).(*PipeExpression); ok {
				query = pipe
				break
			}
		}
	}
	for query != nil && query.Parent() != nil {
		pipe, ok := WrapGeneratedSyntaxElement(query.Parent()).(*PipeExpression)
		if !ok {
			break
		}
		query = pipe
	}
	return query
}

func macroExpandUnionEntities(query Expression, globals *GlobalState, cancellationToken CancellationToken) ([]Expression, bool) {
	entities := make([]Expression, 0)
	var gather func(Expression) bool
	gather = func(expression Expression) bool {
		cancellationToken.ThrowIfCancellationRequested()
		switch current := expression.(type) {
		case *PipeExpression:
			operator, ok := current.Operator().(Expression)
			return ok && gather(current.Expression()) && gather(operator)
		case *UnionOperator:
			if current.Expressions() == nil || current.Expressions().Count() == 0 || len(current.ContainedDiagnostics()) > 0 {
				return false
			}
			for index := 0; index < current.Expressions().Count(); index++ {
				item := current.Expressions().At(index)
				if item == nil {
					return false
				}
				expression, ok := WrapGeneratedSyntaxElement(item.Element()).(Expression)
				if !ok {
					return false
				}
				entity := actorGetEntityExpression(expression, globals)
				if entity == nil || len(entity.ContainedDiagnostics()) > 0 {
					return false
				}
				entities = append(entities, entity)
			}
			return true
		default:
			entity := actorGetEntityExpression(expression, globals)
			if entity == nil || len(entity.ContainedDiagnostics()) > 0 {
				return false
			}
			entities = append(entities, entity)
			return true
		}
	}
	if query == nil || !gather(query) || len(entities) == 0 {
		return nil, false
	}
	return entities, true
}

func macroExpandTrimRangeEnd(text string, start, length int) int {
	end := start + length
	for end > start {
		if text[end-1] != ' ' && text[end-1] != '\t' && text[end-1] != '\r' && text[end-1] != '\n' {
			break
		}
		end--
	}
	return end - start
}
