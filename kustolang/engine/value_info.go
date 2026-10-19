package kustolang

// Ported from src/Kusto.Language/Syntax/ValueInfo.cs.
// ValueInfo preserves both the source spelling and the interpreted value text
// for syntax literals.

type ValueInfo struct {
	RawText   string
	ValueText string
	Value     any
}

func NewValueInfo(rawText, valueText string, value any) *ValueInfo {
	return &ValueInfo{RawText: rawText, ValueText: valueText, Value: value}
}
