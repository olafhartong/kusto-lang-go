package kustolang

import "testing"

func TestAttachKnownAnalyzerSemanticsBindsFunctionsAggregatesAndSignatures(t *testing.T) {
	code := Parse(`T | where format_datetime(Timestamp, "yyyy") == "2024" | summarize N=count_distinct(User)`)
	AttachKnownAnalyzerSemantics(code)

	format := parsedFunctionCallByName(t, code, "format_datetime")
	if format.ReferencedSymbol() != Functions.FormatDatetime || format.Name().ReferencedSymbol() != Functions.FormatDatetime || format.ReferencedSignature() == nil {
		t.Fatalf("format_datetime semantics = call %#v name %#v signature %#v", format.ReferencedSymbol(), format.Name().ReferencedSymbol(), format.ReferencedSignature())
	}
	count := parsedFunctionCallByName(t, code, "count_distinct")
	if count.ReferencedSymbol() != Aggregates.CountDistinct || count.Name().ReferencedSymbol() != Aggregates.CountDistinct {
		t.Fatalf("count_distinct semantics = call %#v name %#v", count.ReferencedSymbol(), count.Name().ReferencedSymbol())
	}
}

func TestAttachKnownAnalyzerSemanticsDoesNotGuessUnknownOrAmbiguousSignature(t *testing.T) {
	code := Parse(`print A=missing_function(1), B=case(true, 1, 0)`)
	AttachKnownAnalyzerSemantics(code)
	missing := parsedFunctionCallByName(t, code, "missing_function")
	if missing.ReferencedSymbol() != nil || missing.ReferencedSignature() != nil {
		t.Fatalf("unknown function semantics = %#v/%#v", missing.ReferencedSymbol(), missing.ReferencedSignature())
	}
	caseCall := parsedFunctionCallByName(t, code, "case")
	if caseCall.ReferencedSymbol() != Functions.Case {
		t.Fatalf("case symbol = %#v, want catalog function", caseCall.ReferencedSymbol())
	}
}
