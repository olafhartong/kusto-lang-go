package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicNoDiagnosticsSharedEmptyCollection(t *testing.T) {
	if kustolang.NoDiagnostics == nil || len(kustolang.NoDiagnostics) != 0 {
		t.Fatalf("NoDiagnostics = %#v", kustolang.NoDiagnostics)
	}
}

func TestPublicDiagnosticMutationAndEqualityParity(t *testing.T) {
	original := kustolang.NewDiagnosticDetailed(
		"APP001",
		kustolang.DiagnosticCategoryCorrectness,
		kustolang.DiagnosticSeverityWarning,
		"Original description",
		"Original message",
	).WithLocation(7, 3)

	tests := []struct {
		name string
		got  kustolang.Diagnostic
		want kustolang.Diagnostic
	}{
		{"code", original.WithCode("APP002"), diagnosticWithCode(original, "APP002")},
		{"category", original.WithCategory(kustolang.DiagnosticCategoryPerformance), diagnosticWithCategory(original, kustolang.DiagnosticCategoryPerformance)},
		{"severity", original.WithSeverity(kustolang.DiagnosticSeveritySuggestion), diagnosticWithSeverity(original, kustolang.DiagnosticSeveritySuggestion)},
		{"description", original.WithDescription("Changed description"), diagnosticWithDescription(original, "Changed description")},
		{"message", original.WithMessage("Changed message"), diagnosticWithMessage(original, "Changed message")},
		{"location kind", original.WithLocationKind(kustolang.DiagnosticLocationRelativeEnd), diagnosticWithLocationKind(original, kustolang.DiagnosticLocationRelativeEnd)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !reflect.DeepEqual(test.got, test.want) {
				t.Fatalf("diagnostic = %#v, want %#v", test.got, test.want)
			}
		})
	}
	if original.Code != "APP001" || original.Category != kustolang.DiagnosticCategoryCorrectness ||
		original.Severity != kustolang.DiagnosticSeverityWarning || original.Description != "Original description" ||
		original.Message != "Original message" || original.LocationKind != kustolang.DiagnosticLocationAbsolute ||
		original.Start != 7 || original.Length != 3 {
		t.Fatalf("original was mutated: %#v", original)
	}

	equivalent := original.WithCategory(kustolang.DiagnosticCategoryPerformance).
		WithSeverity(kustolang.DiagnosticSeverityHidden).
		WithDescription("Ignored by equality")
	if !original.Equals(equivalent) || !equivalent.Equals(original) {
		t.Fatalf("absolute diagnostics should be equal: %#v / %#v", original, equivalent)
	}
	if original.Equals(original.WithMessage("different")) || original.Equals(original.WithLocation(7, 4)) ||
		original.WithLocationKind(kustolang.DiagnosticLocationRelative).Equals(original.WithLocationKind(kustolang.DiagnosticLocationRelative)) {
		t.Fatal("diagnostic equality accepted a different identity or a relative location")
	}
}

func TestPublicDiagnosticDescriptionAndElementLocation(t *testing.T) {
	diagnostic := kustolang.NewDiagnosticWithDescription(
		"APP003",
		kustolang.DiagnosticCategoryPerformance,
		kustolang.DiagnosticSeverityInformation,
		"Description only",
	)
	if diagnostic.Description != "Description only" || diagnostic.Message != "Description only" || diagnostic.HasLocation() {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}

	expression := kustolang.ParseExpression("  value")
	located := diagnostic.WithElementLocation(expression)
	if !located.HasLocation() || located.Start != expression.UTF16TextStart() || located.Length != expression.UTF16Width() || located.End() != located.Start+located.Length {
		t.Fatalf("located diagnostic = %#v, expression = %d+%d", located, expression.UTF16TextStart(), expression.UTF16Width())
	}

	unicodeExpression := kustolang.ParseExpression(`  "😀"`)
	unicodeLocated := diagnostic.WithElementLocation(unicodeExpression)
	if unicodeLocated.Start != unicodeExpression.UTF16TextStart() || unicodeLocated.Length != unicodeExpression.UTF16Width() {
		t.Fatalf("UTF-16 located diagnostic = %#v, expression = %d+%d", unicodeLocated, unicodeExpression.UTF16TextStart(), unicodeExpression.UTF16Width())
	}
	if unicodeExpression.ByteWidth() == unicodeExpression.UTF16Width() {
		t.Fatal("unicode fixture does not distinguish byte and UTF-16 widths")
	}

	first := kustolang.EmptyDiagnostics()
	second := kustolang.EmptyDiagnostics()
	first = append(first, diagnostic)
	if len(second) != 0 {
		t.Fatal("EmptyDiagnostics returned shared mutable storage")
	}
}

func TestPublicDiagnosticEmptyMutationAndNegativeLocationSentinels(t *testing.T) {
	detailed := kustolang.NewDiagnosticDetailed("APP004", "", "", "", "message")
	if detailed.Category != kustolang.DiagnosticCategoryGeneral || detailed.Severity != kustolang.DiagnosticSeverityError ||
		detailed.Description != "message" || detailed.Message != "message" {
		t.Fatalf("detailed diagnostic = %#v", detailed)
	}
	updated := detailed.WithCategory("").WithSeverity("").WithDescription("").WithMessage("").WithLocation(8, 3)
	if updated.Category != "" || updated.Severity != "" || updated.Description != "" || updated.Message != "" {
		t.Fatalf("explicit empty mutation = %#v", updated)
	}
	if retained := updated.WithLocation(-1, -1); retained.Start != 8 || retained.Length != 3 || !retained.HasLocation() {
		t.Fatalf("negative location sentinel = %#v", retained)
	}
}

func diagnosticWithCode(d kustolang.Diagnostic, value string) kustolang.Diagnostic {
	d.Code = value
	return d
}

func diagnosticWithCategory(d kustolang.Diagnostic, value string) kustolang.Diagnostic {
	d.Category = value
	return d
}

func diagnosticWithSeverity(d kustolang.Diagnostic, value string) kustolang.Diagnostic {
	d.Severity = value
	return d
}

func diagnosticWithDescription(d kustolang.Diagnostic, value string) kustolang.Diagnostic {
	d.Description = value
	return d
}

func diagnosticWithMessage(d kustolang.Diagnostic, value string) kustolang.Diagnostic {
	d.Message = value
	return d
}

func diagnosticWithLocationKind(d kustolang.Diagnostic, value kustolang.DiagnosticLocationKind) kustolang.Diagnostic {
	d.LocationKind = value
	return d
}
