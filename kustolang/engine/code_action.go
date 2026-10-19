package kustolang

// Ported from src/Kusto.Language/Editor/CodeAction.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	CodeActionFixAllKind = "FixAll"
	CodeActionMenuKind   = "Menu"
)

type CodeAction interface {
	codeActionMarker()
	Base() CodeActionBase
	WithTitle(title string) CodeAction
	WithDescription(description string) CodeAction
}

type CodeActionBase struct {
	Kind        string
	Title       string
	Description string
}

func (CodeActionBase) codeActionMarker() {}

func newCodeActionBase(kind string, title string, description string) CodeActionBase {
	return CodeActionBase{Kind: kind, Title: title, Description: description}
}

type ApplyAction interface {
	CodeAction
	applyActionMarker()
	ApplyData() []string
	WithData(data []string) ApplyAction
	WithDataStrings(data ...string) ApplyAction
	AddData(data []string) ApplyAction
	AddDataStrings(data ...string) ApplyAction
	RemoveData(count int) ApplyAction
}

// ApplyActionBase can be embedded by implementations outside this package.
type ApplyActionBase struct {
	CodeActionBase
}

func (ApplyActionBase) applyActionMarker() {}

type SingleAction struct {
	CodeActionBase
	Data []string
}

func NewSingleAction(kind string, title string, description string, data []string) *SingleAction {
	return &SingleAction{
		CodeActionBase: newCodeActionBase(kind, title, description),
		Data:           ListToReadOnly(data),
	}
}

func NewCodeAction(kind string, title string, description string, data []string) *SingleAction {
	return NewSingleAction(kind, title, description, data)
}

func NewCodeActionTitle(title string, description string, data []string) *SingleAction {
	return NewSingleAction(title, title, description, data)
}

func (a *SingleAction) codeActionMarker()  {}
func (a *SingleAction) applyActionMarker() {}

func (a *SingleAction) Base() CodeActionBase {
	if a == nil {
		return CodeActionBase{}
	}
	return a.CodeActionBase
}

func (a *SingleAction) ApplyData() []string {
	if a == nil {
		return EmptyReadOnlyList[string]()
	}
	return ListToReadOnly(a.Data)
}

func (a *SingleAction) WithTitle(title string) CodeAction {
	if a == nil {
		return NewSingleAction("", title, "", nil)
	}
	return NewSingleAction(a.Kind, title, a.Description, a.Data)
}

func (a *SingleAction) WithDescription(description string) CodeAction {
	if a == nil {
		return NewSingleAction("", "", description, nil)
	}
	return NewSingleAction(a.Kind, a.Title, description, a.Data)
}

func (a *SingleAction) WithData(data []string) ApplyAction {
	if a == nil {
		return NewSingleAction("", "", "", data)
	}
	if readOnlyStringListEquals(a.Data, data) {
		return a
	}
	return NewSingleAction(a.Kind, a.Title, a.Description, data)
}

func (a *SingleAction) WithDataStrings(data ...string) ApplyAction {
	return a.WithData(data)
}

func (a *SingleAction) AddData(data []string) ApplyAction {
	current := []string(nil)
	if a != nil {
		current = a.Data
	}
	combined := append(ListToReadOnly(current), data...)
	return a.WithData(combined)
}

func (a *SingleAction) AddDataStrings(data ...string) ApplyAction {
	return a.AddData(data)
}

func (a *SingleAction) RemoveData(count int) ApplyAction {
	if a == nil {
		return NewSingleAction("", "", "", nil)
	}
	keep := len(a.Data) - count
	if keep < 0 {
		keep = 0
	}
	if keep > len(a.Data) {
		keep = len(a.Data)
	}
	return a.WithData(a.Data[:keep])
}

type MultiAction struct {
	CodeActionBase
	Data    []string
	Actions []ApplyAction
}

func NewMultiAction(kind string, title string, description string, actions []ApplyAction) *MultiAction {
	return newMultiAction(kind, title, description, nil, actions)
}

func NewCodeActionFixAll(title string, description string, actions []ApplyAction) *MultiAction {
	return NewMultiAction(CodeActionFixAllKind, title, description, actions)
}

func newMultiAction(kind string, title string, description string, data []string, actions []ApplyAction) *MultiAction {
	return &MultiAction{
		CodeActionBase: newCodeActionBase(kind, title, description),
		Data:           ListToReadOnly(data),
		Actions:        ListToReadOnly(actions),
	}
}

func (a *MultiAction) codeActionMarker()  {}
func (a *MultiAction) applyActionMarker() {}

func (a *MultiAction) Base() CodeActionBase {
	if a == nil {
		return CodeActionBase{}
	}
	return a.CodeActionBase
}

func (a *MultiAction) ApplyData() []string {
	if a == nil {
		return EmptyReadOnlyList[string]()
	}
	return ListToReadOnly(a.Data)
}

func (a *MultiAction) WithTitle(title string) CodeAction {
	if a == nil {
		return newMultiAction("", title, "", nil, nil)
	}
	return newMultiAction(a.Kind, title, a.Description, a.Data, a.Actions)
}

func (a *MultiAction) WithDescription(description string) CodeAction {
	if a == nil {
		return newMultiAction("", "", description, nil, nil)
	}
	return newMultiAction(a.Kind, a.Title, description, a.Data, a.Actions)
}

func (a *MultiAction) WithData(data []string) ApplyAction {
	if a == nil {
		return newMultiAction("", "", "", data, nil)
	}
	if readOnlyStringListEquals(a.Data, data) {
		return a
	}
	return newMultiAction(a.Kind, a.Title, a.Description, data, a.Actions)
}

func (a *MultiAction) WithDataStrings(data ...string) ApplyAction {
	return a.WithData(data)
}

func (a *MultiAction) AddData(data []string) ApplyAction {
	current := []string(nil)
	if a != nil {
		current = a.Data
	}
	combined := append(ListToReadOnly(current), data...)
	return a.WithData(combined)
}

func (a *MultiAction) AddDataStrings(data ...string) ApplyAction {
	return a.AddData(data)
}

func (a *MultiAction) RemoveData(count int) ApplyAction {
	if a == nil {
		return newMultiAction("", "", "", nil, nil)
	}
	keep := len(a.Data) - count
	if keep < 0 {
		keep = 0
	}
	if keep > len(a.Data) {
		keep = len(a.Data)
	}
	return a.WithData(a.Data[:keep])
}

type MenuAction struct {
	CodeActionBase
	Actions []CodeAction
}

func NewMenuAction(title string, description string, actions []CodeAction) *MenuAction {
	return newMenuAction(CodeActionMenuKind, title, description, actions)
}

func NewCodeActionMenu(title string, description string, actions []CodeAction) *MenuAction {
	return NewMenuAction(title, description, actions)
}

func newMenuAction(kind string, title string, description string, actions []CodeAction) *MenuAction {
	return &MenuAction{
		CodeActionBase: newCodeActionBase(kind, title, description),
		Actions:        ListToReadOnly(actions),
	}
}

func (a *MenuAction) codeActionMarker() {}

func (a *MenuAction) Base() CodeActionBase {
	if a == nil {
		return CodeActionBase{}
	}
	return a.CodeActionBase
}

func (a *MenuAction) WithTitle(title string) CodeAction {
	if a == nil {
		return newMenuAction(CodeActionMenuKind, title, "", nil)
	}
	return newMenuAction(a.Kind, title, a.Description, a.Actions)
}

func (a *MenuAction) WithDescription(description string) CodeAction {
	if a == nil {
		return newMenuAction(CodeActionMenuKind, "", description, nil)
	}
	return newMenuAction(a.Kind, a.Title, description, a.Actions)
}

func (a *MenuAction) WithActions(actions []CodeAction) *MenuAction {
	if a == nil {
		return NewMenuAction("", "", actions)
	}
	return newMenuAction(a.Kind, a.Title, a.Description, actions)
}

func (a *MenuAction) WithActionsList(actions ...CodeAction) *MenuAction {
	return a.WithActions(actions)
}

type CodeActionInfo struct {
	Actions []CodeAction
}

func NewCodeActionInfo(actions []CodeAction) CodeActionInfo {
	return CodeActionInfo{Actions: ListToReadOnly(actions)}
}

func (info CodeActionInfo) ActionsList() []CodeAction {
	return ListToReadOnly(info.Actions)
}

func CloneCodeActionInfo(info CodeActionInfo) CodeActionInfo {
	owned := make([]CodeAction, len(info.Actions))
	for i, action := range info.Actions {
		owned[i] = cloneCodeAction(action)
	}
	return CodeActionInfo{Actions: ListToReadOnly(owned)}
}

var CodeActionInfoNoActions = NewCodeActionInfo(nil)

type CodeActionOptions struct {
	DiagnosticFilter  DisabledDiagnostics
	FormattingOptions FormattingOptions
}

func NewCodeActionOptions(diagnosticFilter *DisabledDiagnostics, formattingOptions *FormattingOptions) CodeActionOptions {
	filter := DisabledDiagnosticsDefault
	if diagnosticFilter != nil {
		filter = *diagnosticFilter
	}
	formatting := FormattingOptionsDefault
	if formattingOptions != nil {
		formatting = *formattingOptions
	}
	return CodeActionOptions{DiagnosticFilter: filter, FormattingOptions: formatting}
}

func (options CodeActionOptions) WithFormattingOptions(formattingOptions *FormattingOptions) CodeActionOptions {
	return NewCodeActionOptions(&options.DiagnosticFilter, formattingOptions)
}

func (options CodeActionOptions) WithDiagnosticFilter(filter *DisabledDiagnostics) CodeActionOptions {
	return NewCodeActionOptions(filter, &options.FormattingOptions)
}

var CodeActionOptionsDefault = NewCodeActionOptions(nil, nil)

type ResultAction interface {
	resultActionMarker()
}

// ResultActionBase can be embedded by implementations outside this package.
type ResultActionBase struct{}

func (ResultActionBase) resultActionMarker() {}

type ChangeTextAction struct {
	Changes     []TextEdit
	ChangedText string
}

func NewChangeTextAction(changes []TextEdit, changedText string) ChangeTextAction {
	return ChangeTextAction{Changes: ListToReadOnly(changes), ChangedText: changedText}
}

func NewChangeTextActionEditString(editString EditString) ChangeTextAction {
	return NewChangeTextAction(editString.GetChanges(), editString.CurrentText)
}

func (a ChangeTextAction) resultActionMarker() {}

type MoveCaretAction struct {
	NewCaretPosition int
}

func NewMoveCaretAction(newCaretPosition int) MoveCaretAction {
	return MoveCaretAction{NewCaretPosition: newCaretPosition}
}

func (a MoveCaretAction) resultActionMarker() {}

type RenameAction struct{}

func NewRenameAction() RenameAction {
	return RenameAction{}
}

func (a RenameAction) resultActionMarker() {}

var RenameActionInstance = NewRenameAction()

type CodeActionResult struct {
	Actions          []ResultAction
	FailureReason    string
	hasFailureReason bool
}

func NewCodeActionResult(actions ...ResultAction) CodeActionResult {
	return NewCodeActionResultActions(actions)
}

func NewCodeActionResultActions(actions []ResultAction) CodeActionResult {
	return newCodeActionResult(actions, "", false)
}

func newCodeActionResult(actions []ResultAction, failureReason string, hasFailureReason bool) CodeActionResult {
	return CodeActionResult{
		Actions:          ListToReadOnly(actions),
		FailureReason:    failureReason,
		hasFailureReason: hasFailureReason,
	}
}

func CloneCodeActionResult(result CodeActionResult) CodeActionResult {
	owned := make([]ResultAction, len(result.Actions))
	for i, action := range result.Actions {
		owned[i] = cloneResultAction(action)
	}
	return CodeActionResult{
		Actions:          ListToReadOnly(owned),
		FailureReason:    result.FailureReason,
		hasFailureReason: result.hasFailureReason,
	}
}

func cloneCodeAction(action CodeAction) CodeAction {
	switch typed := action.(type) {
	case *SingleAction:
		if typed == nil {
			return (*SingleAction)(nil)
		}
		return NewSingleAction(typed.Kind, typed.Title, typed.Description, typed.Data)
	case *MultiAction:
		if typed == nil {
			return (*MultiAction)(nil)
		}
		actions := make([]ApplyAction, len(typed.Actions))
		for i, nested := range typed.Actions {
			if cloned, ok := cloneCodeAction(nested).(ApplyAction); ok {
				actions[i] = cloned
			} else {
				actions[i] = nested
			}
		}
		return newMultiAction(typed.Kind, typed.Title, typed.Description, typed.Data, actions)
	case *MenuAction:
		if typed == nil {
			return (*MenuAction)(nil)
		}
		actions := make([]CodeAction, len(typed.Actions))
		for i, nested := range typed.Actions {
			actions[i] = cloneCodeAction(nested)
		}
		return newMenuAction(typed.Kind, typed.Title, typed.Description, actions)
	default:
		return action
	}
}

func cloneResultAction(action ResultAction) ResultAction {
	switch typed := action.(type) {
	case ChangeTextAction:
		return NewChangeTextAction(typed.Changes, typed.ChangedText)
	case *ChangeTextAction:
		if typed == nil {
			return (*ChangeTextAction)(nil)
		}
		clone := NewChangeTextAction(typed.Changes, typed.ChangedText)
		return &clone
	case MoveCaretAction:
		return NewMoveCaretAction(typed.NewCaretPosition)
	case *MoveCaretAction:
		if typed == nil {
			return (*MoveCaretAction)(nil)
		}
		clone := NewMoveCaretAction(typed.NewCaretPosition)
		return &clone
	case RenameAction:
		return NewRenameAction()
	case *RenameAction:
		if typed == nil {
			return (*RenameAction)(nil)
		}
		clone := NewRenameAction()
		return &clone
	default:
		return action
	}
}

// HasFailureReason reports whether FailureReason is present. This preserves
// the upstream distinction between a null reason and an explicit empty one.
func (r CodeActionResult) HasFailureReason() bool {
	return r.hasFailureReason
}

func (r CodeActionResult) ActionsList() []ResultAction {
	return ListToReadOnly(r.Actions)
}

func NewCodeActionChange(changes []TextEdit, changedText string) CodeActionResult {
	return NewCodeActionResult(NewChangeTextAction(changes, changedText))
}

func NewCodeActionChangeEditString(editString EditString) CodeActionResult {
	return NewCodeActionChange(editString.GetChanges(), editString.CurrentText)
}

func NewCodeActionChangeAndMove(changes []TextEdit, changedText string, newCaretPosition int) CodeActionResult {
	return NewCodeActionResult(
		NewChangeTextAction(changes, changedText),
		NewMoveCaretAction(newCaretPosition),
	)
}

func NewCodeActionChangeAndMoveEditString(editString EditString, newCaretPosition int) CodeActionResult {
	return NewCodeActionChangeAndMove(editString.GetChanges(), editString.CurrentText, newCaretPosition)
}

func NewCodeActionFailure(failureReason ...string) CodeActionResult {
	if len(failureReason) == 0 {
		return codeActionResultFailed
	}
	return newCodeActionResult(nil, failureReason[0], true)
}

var codeActionResultFailed = newCodeActionResult(nil, "Action failed", true)

func (r CodeActionResult) WithAction(action ResultAction) CodeActionResult {
	return NewCodeActionResultActions(append(ListToReadOnly(r.Actions), action))
}

func (r CodeActionResult) WithAdjustedPosition(delta int) CodeActionResult {
	if len(r.Actions) == 0 || delta == 0 {
		return r
	}

	adjusted := make([]ResultAction, 0, len(r.Actions))
	for _, action := range r.Actions {
		switch typed := action.(type) {
		case MoveCaretAction:
			adjusted = append(adjusted, NewMoveCaretAction(typed.NewCaretPosition+delta))
		case *MoveCaretAction:
			if typed == nil {
				adjusted = append(adjusted, action)
			} else {
				adjusted = append(adjusted, NewMoveCaretAction(typed.NewCaretPosition+delta))
			}
		default:
			adjusted = append(adjusted, action)
		}
	}
	return NewCodeActionResultActions(adjusted)
}

var CodeActionResultNothing = NewCodeActionResultActions(nil)

func readOnlyStringListEquals(left []string, right []string) bool {
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
