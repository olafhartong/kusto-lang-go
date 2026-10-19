package kustolang_test

import (
	"context"
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func newEventsService(t *testing.T, text string) *kustolang.CodeService {
	t.Helper()
	service, err := kustolang.NewCodeServiceForEnvironment(text, nil, environment.LocalData{
		Tables: map[string]environment.TableDetails{
			"Events": {
				"Account":   "string",
				"Timestamp": "datetime",
				"Message":   "string",
			},
		},
	})
	if err != nil {
		t.Fatalf("NewCodeServiceForEnvironment: %v", err)
	}
	return service
}

func TestPublicCodeServiceValidAndMalformedDiagnostics(t *testing.T) {
	valid := newEventsService(t, `Events | where Account == "alice" | project Account`)
	if diagnostics := valid.GetDiagnostics(true, kustolang.CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("valid diagnostics = %#v", diagnostics)
	}

	malformed := newEventsService(t, `Events | where Account ==`)
	diagnostics := malformed.GetDiagnostics(true, kustolang.CancellationToken{})
	if !hasDiagnostic(diagnostics, "KS006") {
		t.Fatalf("malformed diagnostics = %#v, want KS006", diagnostics)
	}

	cached, ok := malformed.TryGetCachedDiagnostics()
	if !ok || !hasDiagnostic(cached, "KS006") {
		t.Fatalf("cached malformed diagnostics = %#v/%t", cached, ok)
	}
}

func TestPublicCodeServiceEditorFeatures(t *testing.T) {
	completionText := `Events | where Acc`
	completionService := newEventsService(t, completionText)
	completionOptions := kustolang.CompletionOptionsDefault.WithIncludeFunctions(kustolang.IncludeFunctionKindAll)
	completion := completionService.GetCompletionItems(len(completionText), completionOptions, kustolang.CancellationToken{})
	var accountCompletion *kustolang.CompletionItem
	for i := range completion.Items {
		if completion.Items[i].DisplayText == "Account" {
			accountCompletion = &completion.Items[i]
			break
		}
	}
	if accountCompletion == nil || accountCompletion.Kind != kustolang.CompletionKindColumn {
		t.Fatalf("Account completion not found: %#v", completion.Items)
	}
	if accountCompletion.Rank() != kustolang.CompletionRankDefault || accountCompletion.Priority() == 0 {
		t.Fatalf("completion ordering metadata = rank:%v priority:%v", accountCompletion.Rank(), accountCompletion.Priority())
	}

	text := `Events|where Account=="alice"|project Account`
	service := newEventsService(t, text)
	formatted := service.GetFormattedText(
		kustolang.FormattingOptionsDefault.WithPipeOperatorStyle(kustolang.PlacementStyleNewLine),
		len(text),
		kustolang.CancellationToken{},
	)
	wantFormatted := "Events\n| where Account == \"alice\"\n| project Account"
	if formatted.Text != wantFormatted || formatted.Position != len(wantFormatted) {
		t.Fatalf("formatted = %#v, want %q", formatted, wantFormatted)
	}

	accountPosition := strings.Index(text, "Account") + 1
	quickInfo := service.GetQuickInfo(accountPosition, kustolang.QuickInfoOptionsDefault, kustolang.CancellationToken{})
	if len(quickInfo.Items) == 0 || quickInfo.Items[0].Kind != kustolang.QuickInfoKindColumn || !strings.Contains(quickInfo.Items[0].Text(), "Account: string") {
		t.Fatalf("Account quick info = %#v", quickInfo)
	}

	classifications := service.GetClassifications(0, len(text), true, true, kustolang.CancellationToken{}).Classifications
	for _, kind := range []kustolang.ClassificationKind{
		kustolang.ClassificationKindTable,
		kustolang.ClassificationKindColumn,
		kustolang.ClassificationKindQueryOperator,
		kustolang.ClassificationKindStringLiteral,
	} {
		if !hasClassification(classifications, kind) {
			t.Fatalf("classification %v missing from %#v", kind, classifications)
		}
	}
}

func TestPublicCodeServiceAnalyzerFixAndCancellation(t *testing.T) {
	text := `Events | where Account contains "failure"`
	service := newEventsService(t, text)
	analyzerDiagnostics := service.GetAnalyzerDiagnostics(true, kustolang.CancellationToken{})
	if !hasDiagnostic(analyzerDiagnostics, "KS500") {
		t.Fatalf("analyzer diagnostics = %#v, want KS500", analyzerDiagnostics)
	}

	containsPosition := strings.Index(text, "contains")
	actions := service.GetCodeActions(containsPosition, 0, 0, kustolang.CodeActionOptionsDefault, true, kustolang.KustoActorNameDiagnosticFixer, kustolang.CancellationToken{})
	if len(actions.Actions) == 0 {
		t.Fatal("expected a diagnostic fix action")
	}
	apply, ok := actions.Actions[0].(kustolang.ApplyAction)
	if !ok {
		t.Fatalf("action = %T, want ApplyAction", actions.Actions[0])
	}
	if apply.Base().Title == "" {
		t.Fatalf("action base = %#v", apply.Base())
	}
	result := service.ApplyCodeAction(apply, containsPosition, kustolang.CodeActionOptionsDefault, kustolang.CancellationToken{})
	if result.FailureReason != "" || len(result.Actions) == 0 {
		t.Fatalf("apply result = %#v", result)
	}
	change, ok := result.Actions[0].(kustolang.ChangeTextAction)
	if !ok || !strings.Contains(change.ChangedText, `Account has "failure"`) {
		t.Fatalf("change result = %#v", result.Actions)
	}

	disabled := kustolang.ParseDisabledDiagnostics("KS500")
	filtered := kustolang.NewCodeActionOptions(&disabled, nil)
	if got := service.GetCodeActions(containsPosition, 0, 0, filtered, true, kustolang.DiagnosticFixActorName, kustolang.CancellationToken{}); len(got.Actions) != 0 {
		t.Fatalf("disabled KS500 actions = %#v", got.Actions)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	token := kustolang.NewCancellationToken(ctx)
	if got := service.GetCompletionItems(len(text), kustolang.CompletionOptionsDefault, token); len(got.Items) != 0 {
		t.Fatalf("canceled completion = %#v", got.Items)
	}

	if len(kustolang.KustoActorNamesAll()) < 5 || kustolang.CompletionKindMaterializedView != kustolang.CompletionKindMaterialiedView {
		t.Fatal("public compatibility constants are incomplete")
	}
}

func TestPublicCodeServiceTypedClientParameters(t *testing.T) {
	tenant := kustolang.NewParameterSymbol("Tenant", kustolang.ScalarTypes.String.Type(), "tenant identifier")
	globals := kustolang.WithClientParameters(kustolang.NewGlobalState(nil), tenant)
	text := `print tenant={Tenant}`
	service := kustolang.NewCodeServiceWithGlobals(text, globals)
	if diagnostics := service.GetDiagnostics(true, kustolang.CancellationToken{}); len(diagnostics) != 0 {
		t.Fatalf("typed client diagnostics = %#v", diagnostics)
	}
	parameters := service.GetClientParameters()
	if len(parameters) != 1 || parameters[0].Name != "Tenant" {
		t.Fatalf("client parameter references = %#v", parameters)
	}
	classifications := service.GetClassifications(0, len(text), true, true, kustolang.CancellationToken{}).Classifications
	if !hasClassification(classifications, kustolang.ClassificationKindClientParameter) {
		t.Fatalf("client parameter classifications = %#v", classifications)
	}
}

func hasDiagnostic(diagnostics []kustolang.Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func hasClassification(classifications []kustolang.ClassifiedRange, kind kustolang.ClassificationKind) bool {
	for _, classification := range classifications {
		if classification.Kind == kind {
			return true
		}
	}
	return false
}
