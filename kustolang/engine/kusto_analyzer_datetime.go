package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/Analyzers/DateTimeLiteralAnalyzer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"fmt"
	"strings"
)

var invalidDateTimeDiagnostic = NewDiagnosticDetailed(KustoAnalyzerCodeInvalidDateTimeLiteral,
	DiagnosticCategoryCorrectness, DiagnosticSeverityWarning, "Invalid datetime literal",
	"This datetime literal cannot be parsed and will be rejected by the service")
var numericDateTimeDiagnostic = NewDiagnosticDetailed(KustoAnalyzerCodeAmbiguousNumericDateTimeLiteral,
	DiagnosticCategoryCorrectness, DiagnosticSeverityWarning, "Ambiguous numeric datetime literal",
	"A numeric datetime literal is interpreted by its digit count, use an explicit date instead")

func NewDateTimeLiteralAnalyzer() KustoAnalyzer {
	return NewKustoAnalyzer(KustoAnalyzerCallbacks{
		Name:        "DateTimeLiteralAnalyzer",
		Diagnostics: func() []Diagnostic { return []Diagnostic{invalidDateTimeDiagnostic, numericDateTimeDiagnostic} },
		Analyze:     analyzeDateTimeLiterals,
	})
}
func analyzeDateTimeLiterals(code *KustoCode, diagnostics []Diagnostic, token CancellationToken) []Diagnostic {
	if code == nil || code.Syntax() == nil {
		return diagnostics
	}
	walkSyntaxWithTriviaStart(code.Syntax(), 0, func(element SyntaxElement, start int) {
		token.ThrowIfCancellationRequested()
		literal, ok := WrapGeneratedSyntaxElement(element).(*LiteralExpression)
		if !ok || literal.Kind() != SyntaxKindDateTimeLiteralExpression {
			return
		}
		syntaxToken := literal.Token()
		if syntaxToken == nil || syntaxToken.Value() == nil {
			return
		}
		// The source treats an absent argument as null; quoted empty text is invalid.
		raw := syntaxToken.RawText()
		if open := strings.IndexByte(raw, '('); open >= 0 {
			interior := strings.TrimSpace(strings.TrimSuffix(raw[open+1:], ")"))
			if interior == "" {
				return
			}
		}
		text := syntaxToken.ValueText()
		interpretation, value := parseKustoDateTime(text)
		var diagnostic Diagnostic
		if interpretation == datetimeInvalid {
			diagnostic = invalidDateTimeDiagnostic.WithMessage(fmt.Sprintf("The datetime literal '%s' cannot be parsed and will be rejected by the service", text))
		} else if interpretation == datetimeTicks || interpretation == datetimeEpochSeconds || (interpretation == datetimeFixedWidth && len(strings.TrimSpace(text)) == 4) {
			meaning := "a year"
			if interpretation == datetimeTicks {
				meaning = "a count of ticks since year one"
			} else if interpretation == datetimeEpochSeconds {
				meaning = "a count of seconds since 1970-01-01"
			}
			fraction := ""
			if ticks := value.Nanosecond() / 100; ticks != 0 {
				fraction = fmt.Sprintf(".%07d", ticks)
			}
			date := value.Format("2006-01-02")
			instant := value.Format("2006-01-02T15:04:05") + fraction + "Z"
			diagnostic = numericDateTimeDiagnostic.WithMessage(fmt.Sprintf("The datetime literal '%s' is interpreted as %s, giving %s. Use an explicit date such as datetime(%s) to say what is meant", text, meaning, instant, date))
		} else {
			return
		}
		diagnostics = append(diagnostics, diagnostic.WithByteLocation(elementTextStart(literal, start), literal.Width()))
	})
	return diagnostics
}
