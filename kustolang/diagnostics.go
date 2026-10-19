package kustolang

// Ported from src/Kusto.Language/Diagnostics/Diagnostic.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

//go:generate go run ../tools/gen-public-diagnostic-facts

// NoDiagnostics is the source-shaped shared empty diagnostic collection.
var NoDiagnostics = EmptyReadOnlyList[Diagnostic]()

func NewDiagnostic(code, message string) Diagnostic {
	return internal.NewDiagnostic(code, message)
}

func NewDiagnosticWithDescription(code, category, severity, description string) Diagnostic {
	return internal.NewDiagnosticWithDescription(code, category, severity, description)
}

func NewDiagnosticDetailed(code, category, severity, description, message string) Diagnostic {
	return internal.NewDiagnosticDetailed(code, category, severity, description, message)
}

func EmptyDiagnostics() []Diagnostic {
	return []Diagnostic{}
}
