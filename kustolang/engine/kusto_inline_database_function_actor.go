package kustolang

import (
	"strconv"
	"strings"
)

// Ported from:
//   src/Kusto.Language/Editor/Kusto/Actors/InlineDatabaseFunctionActor.cs
//   src/Kusto.Language/Editor/Kusto/Actors/ActorUtilities.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	InlineOnceKind         = "InlineOnce"
	InlineRecursiveKind    = "InlineRecursive"
	InlineAllOnceKind      = "InlineAllOnce"
	InlineAllRecursiveKind = "InlineAllRecursive"
)

// InlineDatabaseFunctionActor exposes the upstream inlining actor independently
// of KustoCodeService dispatch. Native methods use byte offsets; UTF16 methods
// translate the editor boundary explicitly.
type InlineDatabaseFunctionActor struct{}

func NewInlineDatabaseFunctionActor() *InlineDatabaseFunctionActor {
	return &InlineDatabaseFunctionActor{}
}

func (a *InlineDatabaseFunctionActor) GetActions(code *KustoCode, position, selectionStart, selectionLength int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionInfo {
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
	if token == nil || position < token.ByteTextStart() || position > token.ByteEnd() {
		return CodeActionInfoNoActions
	}
	reference := GetFirstAncestor[*NameReference](token)
	if reference == nil || !inlineIsDatabaseFunctionReference(reference, code.Globals) {
		return CodeActionInfoNoActions
	}

	if selectionLength == 0 {
		selectionStart = 0
		selectionLength = len(code.Text)
	}
	menuActions := []CodeAction{
		NewCodeAction(InlineOnceKind, "Inline", "Inline this function only.", []string{strconv.Itoa(position)}),
		NewCodeAction(InlineRecursiveKind, "Inline Recursive", "Inline this function recursively.", []string{strconv.Itoa(position)}),
	}
	if len(inlineDatabaseFunctionsInRange(code, selectionStart, selectionLength)) > 1 {
		menuActions = append(menuActions,
			NewCodeAction(InlineAllOnceKind, "Inline All", "Inline all functions in selection or query.", []string{strconv.Itoa(selectionStart), strconv.Itoa(selectionLength)}),
			NewCodeAction(InlineAllRecursiveKind, "Inline All, Recursive", "Inline all functions in selection or query, recursively.", []string{strconv.Itoa(selectionStart), strconv.Itoa(selectionLength)}),
		)
	}
	menu := NewCodeActionMenu("Inline Function", "Copy database function into query.", menuActions)
	return NewCodeActionInfo([]CodeAction{menu})
}

func (a *InlineDatabaseFunctionActor) ApplyAction(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || code.Syntax() == nil || action == nil || caretPosition < 0 || caretPosition > len(code.Text) {
		return NewCodeActionFailure("Bad action data")
	}
	if !code.HasSemantics() {
		code = code.Analyze()
	}
	if code == nil || !code.HasSemantics() {
		return NewCodeActionFailure("Bad action data")
	}
	data := action.ApplyData()
	var start, length int
	var ok bool
	switch action.Base().Kind {
	case InlineOnceKind, InlineRecursiveKind:
		if len(data) != 1 {
			return NewCodeActionFailure("Bad action data")
		}
		start, ok = parseActorPosition(data[0], len(code.Text))
	case InlineAllOnceKind, InlineAllRecursiveKind:
		if len(data) != 2 {
			return NewCodeActionFailure("Bad action data")
		}
		start, ok = parseActorPosition(data[0], len(code.Text))
		if ok {
			length, ok = parseActorLength(data[1], len(code.Text)-start)
		}
	default:
		return NewCodeActionFailure("Bad action data")
	}
	if !ok {
		return NewCodeActionFailure("Bad action data")
	}

	functions := inlineDatabaseFunctionsInRange(code, start, length)
	recursive := action.Base().Kind == InlineRecursiveKind || action.Base().Kind == InlineAllRecursiveKind
	return inlineDatabaseFunctions(code, functions, caretPosition, recursive, cancellationToken)
}

func (a *InlineDatabaseFunctionActor) GetActionsUTF16(code *KustoCode, position, selectionStart, selectionLength int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionInfo {
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
	converted := make([]CodeAction, 0, len(info.Actions))
	for _, action := range info.Actions {
		converted = append(converted, inlineActionToUTF16(code.Text, action))
	}
	return NewCodeActionInfo(converted)
}

func (a *InlineDatabaseFunctionActor) ApplyActionUTF16(code *KustoCode, action ApplyAction, caretPosition int, options CodeActionOptions, cancellationToken CancellationToken) CodeActionResult {
	cancellationToken.ThrowIfCancellationRequested()
	if code == nil || action == nil {
		return NewCodeActionFailure("Bad action data")
	}
	byteCaret, caretOK := utf16RangeStartToByte(code.Text, caretPosition)
	byteAction, actionOK := inlineApplyActionToBytes(code.Text, action)
	if !caretOK || !actionOK {
		return NewCodeActionFailure("Bad action data")
	}
	result := a.ApplyAction(code, byteAction, byteCaret, options, cancellationToken)
	return extractionResultToUTF16(code.Text, result)
}

func inlineIsDatabaseFunctionReference(reference *NameReference, globals *GlobalState) bool {
	if reference == nil || globals == nil {
		return false
	}
	function, ok := reference.ReferencedSymbol().(*FunctionSymbol)
	if !ok || !globals.IsDatabaseFunction(function) {
		return false
	}
	if _, call := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); call {
		return true
	}
	return function.MinArgumentCount() == 0
}

func inlineDatabaseFunctionsInRange(code *KustoCode, start, length int) []*FunctionSymbol {
	if code == nil || code.Syntax() == nil {
		return nil
	}
	seen := make(map[*FunctionSymbol]bool)
	functions := make([]*FunctionSymbol, 0)
	for _, reference := range GetDescendants[*NameReference](code.Syntax(), func(reference *NameReference) bool {
		return inlineIsDatabaseFunctionReference(reference, code.Globals)
	}) {
		if !inlineRangesOverlap(reference.ByteTextStart(), reference.ByteWidth(), start, length) {
			continue
		}
		function := reference.ReferencedSymbol().(*FunctionSymbol)
		if !seen[function] {
			seen[function] = true
			functions = append(functions, function)
		}
	}
	return functions
}

func inlineRangesOverlap(start1, length1, start2, length2 int) bool {
	end := min(start1+length1, start2+length2)
	start := max(start1, start2)
	return end >= start
}

func inlineDatabaseFunctions(code *KustoCode, functions []*FunctionSymbol, caretPosition int, recursive bool, cancellationToken CancellationToken) CodeActionResult {
	if len(functions) == 0 {
		return NewCodeActionFailure("Reference to database function not found")
	}
	originalText := code.Text
	remaining := append([]*FunctionSymbol{}, functions...)
	alreadyInlined := make(map[*FunctionSymbol]bool)
	var result CodeActionResult
	steps := 0

	for len(remaining) > 0 {
		cancellationToken.ThrowIfCancellationRequested()
		function := remaining[0]
		remaining = remaining[1:]
		if alreadyInlined[function] {
			continue
		}
		alreadyInlined[function] = true

		step := inlineDatabaseFunction(code, function, caretPosition)
		if step.HasFailureReason() {
			if steps == 0 {
				return step
			}
			break
		}
		result = step
		steps++
		change, move, ok := inlineResultChangeAndMove(step)
		if !ok {
			return NewCodeActionFailure("Action failed")
		}
		changedRange := NewEditString(code.Text).ApplyAll(change.Changes).GetChangeRange()
		caretPosition = move.NewCaretPosition
		code = code.WithTextWithCancellation(change.ChangedText, cancellationToken)
		if code == nil || !code.HasSemantics() {
			return NewCodeActionFailure("Could not analyze inlined function")
		}

		if recursive {
			length := changedRange.Length - 1
			if length < 0 {
				length = 0
			}
			for _, candidate := range inlineDatabaseFunctionsInRange(code, changedRange.Start, length) {
				if !alreadyInlined[candidate] && !inlineFunctionQueued(remaining, candidate) {
					remaining = append(remaining, candidate)
				}
			}
		}
	}
	if steps == 0 {
		return NewCodeActionFailure("Reference to database function not found")
	}
	if steps == 1 {
		return result
	}
	finalText := inlineChangedText(result)
	edit := NewEditString(originalText).ReplaceAt(0, len(originalText), finalText)
	return NewCodeActionChangeAndMoveEditString(edit, caretPosition)
}

func inlineDatabaseFunction(code *KustoCode, function *FunctionSymbol, caretPosition int) CodeActionResult {
	reference := actorGetFirstReference(code.Syntax(), function)
	if reference == nil {
		return NewCodeActionFailure("Reference to database function not found")
	}
	insertPosition, ok := inlineTopLevelInsertionPosition(code, reference.ByteTextStart())
	if !ok {
		return NewCodeActionFailure("Reference to database function not found")
	}

	body := reference.GetCalledFunctionBody()
	if call, isCall := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); isCall {
		body = call.GetCalledFunctionBody()
	}
	if body == nil {
		return NewCodeActionFailure("Could not access definition of referenced function")
	}
	bodyText := body.Text(IncludeTriviaAll)
	requalified := kustoQualifierRequalifyEdit(NewEditString(bodyText), body, code.Globals, code.Globals.Cluster, code.Globals.Database).CurrentText
	declaration, ok := inlineFunctionLetStatement(function, inlineBodyWithBraces(requalified))
	if !ok {
		return NewCodeActionFailure("Could not access definition of referenced function")
	}
	query := actorQueryWithDatabaseQualifiersRemoved(code.Syntax(), function)
	newText := query.Insert(insertPosition, declaration+"\n")
	return NewCodeActionChangeAndMoveEditString(newText, newText.GetCurrentPosition(caretPosition))
}

func inlineFunctionLetStatement(function *FunctionSymbol, body string) (string, bool) {
	if function == nil {
		return "", false
	}
	signatures := function.Signatures()
	if len(signatures) == 0 || signatures[0] == nil {
		return "", false
	}
	return "let " + BracketNameIfNecessary(function.SymbolName()) + " = " + ParameterListDeclaration(signatures[0].Parameters) + " " + body + ";", true
}

func inlineBodyWithBraces(body string) string {
	if !strings.HasPrefix(body, "{") {
		body = "{\n" + body
	}
	if !strings.HasSuffix(body, "}") {
		body += "\n}"
	}
	return body
}

func inlineTopLevelInsertionPosition(code *KustoCode, position int) (int, bool) {
	if insertion, ok := actorNearestTopLevelStatementInsertionPosition(code, position); ok {
		return insertion, true
	}
	token := actorTokenNear(code, position)
	if token == nil {
		return 0, false
	}
	for element := token.Parent(); element != nil; element = element.Parent() {
		if _, ok := WrapGeneratedSyntaxElement(element).(Statement); !ok {
			continue
		}
		separated := element.Parent()
		list := SyntaxElement(nil)
		if separated != nil {
			list = separated.Parent()
		}
		if list != nil && list.Parent() != nil && list.Parent().Kind() == SyntaxKindQueryBlock {
			return extractInsertionPositionBefore(element), true
		}
	}
	return 0, false
}

func inlineResultChangeAndMove(result CodeActionResult) (ChangeTextAction, MoveCaretAction, bool) {
	var change ChangeTextAction
	var move MoveCaretAction
	changeOK, moveOK := false, false
	for _, action := range result.Actions {
		switch typed := action.(type) {
		case ChangeTextAction:
			change, changeOK = typed, true
		case MoveCaretAction:
			move, moveOK = typed, true
		}
	}
	return change, move, changeOK && moveOK
}

func inlineChangedText(result CodeActionResult) string {
	change, _, ok := inlineResultChangeAndMove(result)
	if !ok {
		return ""
	}
	return change.ChangedText
}

func inlineFunctionQueued(functions []*FunctionSymbol, function *FunctionSymbol) bool {
	for _, candidate := range functions {
		if candidate == function {
			return true
		}
	}
	return false
}

func inlineActionToUTF16(text string, action CodeAction) CodeAction {
	if menu, ok := action.(*MenuAction); ok {
		actions := make([]CodeAction, 0, len(menu.Actions))
		for _, child := range menu.Actions {
			actions = append(actions, inlineActionToUTF16(text, child))
		}
		return menu.WithActions(actions)
	}
	apply, ok := action.(ApplyAction)
	if !ok {
		return action
	}
	data := apply.ApplyData()
	switch apply.Base().Kind {
	case InlineOnceKind, InlineRecursiveKind:
		if len(data) == 1 {
			start, err := strconv.Atoi(data[0])
			if err == nil {
				return apply.WithDataStrings(strconv.Itoa(byteOffsetToUTF16Position(text, start)))
			}
		}
	case InlineAllOnceKind, InlineAllRecursiveKind:
		if len(data) == 2 {
			start, startErr := strconv.Atoi(data[0])
			length, lengthErr := strconv.Atoi(data[1])
			if startErr == nil && lengthErr == nil {
				utfStart := byteOffsetToUTF16Position(text, start)
				utfLength := byteOffsetToUTF16Position(text, start+length) - utfStart
				return apply.WithDataStrings(strconv.Itoa(utfStart), strconv.Itoa(utfLength))
			}
		}
	}
	return action
}

func inlineApplyActionToBytes(text string, action ApplyAction) (ApplyAction, bool) {
	data := action.ApplyData()
	switch action.Base().Kind {
	case InlineOnceKind, InlineRecursiveKind:
		if len(data) != 1 {
			return nil, false
		}
		position, err := strconv.Atoi(data[0])
		bytePosition, ok := utf16RangeStartToByte(text, position)
		if err != nil || !ok {
			return nil, false
		}
		return action.WithDataStrings(strconv.Itoa(bytePosition)), true
	case InlineAllOnceKind, InlineAllRecursiveKind:
		if len(data) != 2 {
			return nil, false
		}
		start, startErr := strconv.Atoi(data[0])
		length, lengthErr := strconv.Atoi(data[1])
		byteStart, byteLength, ok := utf16RangeToBytes(text, start, length)
		if startErr != nil || lengthErr != nil || !ok {
			return nil, false
		}
		return action.WithDataStrings(strconv.Itoa(byteStart), strconv.Itoa(byteLength)), true
	default:
		return nil, false
	}
}
