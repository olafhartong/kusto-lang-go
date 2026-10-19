package kustolang

import "strconv"

// Ported from src/Kusto.Language/Editor/Kusto/Actors/RenameActor.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func kustoRenameActorGetAction(service *KustoCodeService, code *KustoCode, position int, cancellationToken CancellationToken) ApplyAction {
	refs, index, ok := kustoServiceTokenRefAtPosition(code, position)
	if !ok {
		return nil
	}
	ref := refs[index]
	token := ref.positionedToken()
	if token == nil || position > token.ByteEnd() {
		return nil
	}
	name, _, ok := relatedBoundName(ref)
	if !ok || name == nil {
		return nil
	}
	location := relatedAncestorLocation(ref, name)
	if location == nil {
		return nil
	}
	nameStart := elementTextStart(name, location.triviaStart)
	related := service.GetRelatedElements(byteOffsetToUTF16Position(code.Text, nameStart), FindRelatedOptionsNone, cancellationToken)
	canRename := false
	for _, element := range related.Elements {
		if element.Kind == RelatedElementKindDeclaration {
			canRename = true
			break
		}
	}
	if !canRename {
		return nil
	}
	return NewCodeActionTitle("Rename", "Rename selected item", []string{strconv.Itoa(nameStart)})
}

func kustoRenameActorApplyAction(action ApplyAction) CodeActionResult {
	if action == nil {
		return NewCodeActionFailure()
	}
	data := action.ApplyData()
	if len(data) == 0 {
		return NewCodeActionFailure()
	}
	parsed, err := strconv.ParseInt(data[0], 10, 32)
	if err != nil {
		return NewCodeActionFailure()
	}
	position := int(parsed)
	return NewCodeActionResult(NewMoveCaretAction(position), RenameActionInstance)
}
