package kustolang

import (
	"strings"
	"sync"
)

// Ported from src/Kusto.Language/Symbols/CommandSymbol.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type CommandSymbol struct {
	BaseSymbol
	resultSchema     string
	hasResultSchema  bool
	resultSchemaOnce sync.Once
	resultType       *TableSymbol
	resultTypeOnce   sync.Once
	construction     string
}

func NewCommandSymbol(name, resultSchema string, construction ...string) *CommandSymbol {
	return &CommandSymbol{
		BaseSymbol:      NewBaseSymbol(name),
		resultSchema:    resultSchema,
		hasResultSchema: true,
		construction:    optionalConstruction(construction),
	}
}

func NewCommandSymbolWithResultType(name string, resultType *TableSymbol, construction ...string) *CommandSymbol {
	return &CommandSymbol{
		BaseSymbol:   NewBaseSymbol(name),
		resultType:   resultType,
		construction: optionalConstruction(construction),
	}
}

func (s *CommandSymbol) SymbolKind() SymbolKind {
	return SymbolKindCommand
}

func (s *CommandSymbol) ResultSchema() string {
	if s == nil {
		return ""
	}
	s.resultSchemaOnce.Do(func() {
		if !s.hasResultSchema && s.ResultType() != nil {
			s.resultSchema = SchemaDisplayText(s.resultType)
			s.hasResultSchema = true
		}
	})
	if s.hasResultSchema {
		return s.resultSchema
	}
	return ""
}

func (s *CommandSymbol) ResultType() *TableSymbol {
	if s == nil {
		return nil
	}
	s.resultTypeOnce.Do(func() {
		if s.resultType != nil || !s.hasResultSchema {
			return
		}
		schema := strings.TrimSpace(s.resultSchema)
		if schema == "()" || schema == "(*)" {
			s.resultType = NewTableSymbol("").WithIsOpen(true)
			return
		}
		table, err := TableSymbolFromSchema(schema)
		if err != nil {
			rowSchema := ParseRowSchema(schema)
			columns := validRowSchemaColumnPrefix(rowSchema)
			if len(columns) == 0 {
				return
			}
			table = NewTableSymbol("", columns...)
		}
		s.resultType = table
	})
	return s.resultType
}

func (s *CommandSymbol) Construction() string {
	if s == nil {
		return ""
	}
	return s.construction
}

func optionalConstruction(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
