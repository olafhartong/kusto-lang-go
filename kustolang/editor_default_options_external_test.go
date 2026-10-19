package kustolang_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicCodeServiceDefaultOptionEntryPointsMatchExplicitDefaults(t *testing.T) {
	const text = `Events|where Account=="alice"|project Account`
	service := newEventsService(t, text)
	token := kustolang.CancellationToken{}

	if got, want := service.GetOutlinesDefault(token), service.GetOutlines(kustolang.OutliningOptionsDefault, token); !reflect.DeepEqual(got, want) {
		t.Fatalf("default outlines = %#v, want %#v", got, want)
	}
	completionPosition := strings.Index(text, "Account") + len("Account")
	if got, want := service.GetCompletionItemsDefault(completionPosition, token), service.GetCompletionItems(completionPosition, kustolang.CompletionOptionsDefault, token); !reflect.DeepEqual(got, want) {
		t.Fatalf("default completion = %#v, want %#v", got, want)
	}
	quickInfoPosition := strings.Index(text, "Account") + 1
	if got, want := service.GetQuickInfoDefault(quickInfoPosition, token), service.GetQuickInfo(quickInfoPosition, kustolang.QuickInfoOptionsDefault, token); !reflect.DeepEqual(got, want) {
		t.Fatalf("default quick info = %#v, want %#v", got, want)
	}
	if got, want := service.GetFormattedTextDefault(len(text), token), service.GetFormattedText(kustolang.FormattingOptionsDefault, len(text), token); !reflect.DeepEqual(got, want) {
		t.Fatalf("default formatted text = %#v, want %#v", got, want)
	}
}

func TestPublicCodeServiceDefaultCallFormsMatchPinnedDefaults(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{name: "valid", text: `Events | where Account contains "failure"`},
		{name: "malformed", text: `Events | where Account ==`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newEventsService(t, test.text)
			position := strings.Index(test.text, "Account") + 1

			if got, want := service.IsFeatureSupportedDefault(kustolang.CodeServiceFeatureDiagnostics), service.IsFeatureSupported(kustolang.CodeServiceFeatureDiagnostics, -1); got != want {
				t.Fatalf("default feature support = %t, want %t", got, want)
			}
			diagnostics := service.GetDiagnosticsDefault()
			if want := service.GetDiagnostics(true, kustolang.CancellationToken{}); !reflect.DeepEqual(diagnostics, want) {
				t.Fatalf("default diagnostics = %#v, want %#v", diagnostics, want)
			}
			if test.name == "malformed" && !hasDiagnostic(diagnostics, "KS006") {
				t.Fatalf("malformed default diagnostics = %#v, want KS006", diagnostics)
			}
			if got, want := service.GetAnalyzerDiagnosticsDefault(), service.GetAnalyzerDiagnostics(true, kustolang.CancellationToken{}); !reflect.DeepEqual(got, want) {
				t.Fatalf("default analyzer diagnostics = %#v, want %#v", got, want)
			}
			if got, want := service.GetCombinedDiagnosticsDefault(), service.GetCombinedDiagnostics(true, nil, kustolang.CancellationToken{}); !reflect.DeepEqual(got, want) {
				t.Fatalf("default combined diagnostics = %#v, want %#v", got, want)
			}
			if got, want := service.GetCodeActionsDefault(position, 0, 0), service.GetCodeActions(position, 0, 0, kustolang.CodeActionOptionsDefault, true, "", kustolang.CancellationToken{}); !reflect.DeepEqual(got, want) {
				t.Fatalf("default code actions = %#v, want %#v", got, want)
			}
			classificationStart, classificationLength := position, 3
			classifications := service.GetClassificationsDefault(classificationStart, classificationLength)
			if want := service.GetClassifications(classificationStart, classificationLength, true, true, kustolang.CancellationToken{}); !reflect.DeepEqual(classifications, want) {
				t.Fatalf("default classifications = %#v, want %#v", classifications, want)
			}
			for _, classification := range classifications.Classifications {
				if classification.Start < classificationStart || classification.End() > classificationStart+classificationLength {
					t.Fatalf("default classification was not clipped: %#v", classification)
				}
			}
		})
	}

	service := newEventsService(t, tests[0].text)
	position := strings.Index(tests[0].text, "contains")
	actions := service.GetCodeActionsDefault(position, 0, 0)
	for _, action := range actions.Actions {
		apply, ok := action.(kustolang.ApplyAction)
		if !ok {
			continue
		}
		if got, want := service.ApplyCodeActionDefault(apply, position), service.ApplyCodeAction(apply, position, kustolang.CodeActionOptionsDefault, kustolang.CancellationToken{}); !reflect.DeepEqual(got, want) {
			t.Fatalf("default applied action = %#v, want %#v", got, want)
		}
		return
	}
	t.Fatal("default code actions contained no applicable action")
}
