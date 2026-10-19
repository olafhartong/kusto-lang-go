package kustolang

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Ported from src/Kusto.Language/Symbols/SchemaDisplay.cs plus the identifier
// and string-literal helpers from src/Kusto.Language/Parser/KustoFacts.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func SchemaDisplayText(symbol Symbol) string {
	switch s := symbol.(type) {
	case nil:
		return ""
	case *ColumnSymbol:
		return BracketNameIfNecessary(s.SymbolName()) + ": " + SchemaDisplayScalarTypeText(s.Type)
	case *ParameterSymbol:
		return BracketNameIfNecessary(s.SymbolName()) + ": " + SchemaDisplaySymbolTypeText(s.Type)
	case *TableSymbol:
		return schemaDisplayTableText(s)
	case *MaterializedViewSymbol:
		if s == nil {
			return ""
		}
		return schemaDisplayTableText(&s.TableSymbol)
	case *ExternalTableSymbol:
		if s == nil {
			return ""
		}
		return schemaDisplayTableText(&s.TableSymbol)
	case *StoredQueryResultSymbol:
		if s == nil {
			return ""
		}
		return schemaDisplayTableText(&s.TableSymbol)
	case *FunctionSymbol:
		signatures := s.Signatures()
		if len(signatures) == 0 || signatures[0] == nil {
			return "()"
		}
		parameters := signatures[0].Parameters
		parts := make([]string, 0, len(parameters))
		for _, parameter := range parameters {
			if parameter != nil {
				parts = append(parts, BracketNameIfNecessary(parameter.Name)+": "+SchemaDisplayParameterTypeText(parameter))
			}
		}
		return "(" + strings.Join(parts, ", ") + ")"
	default:
		return symbol.SymbolName()
	}
}

func SchemaDisplaySymbolTypeText(symbol Symbol) string {
	switch s := symbol.(type) {
	case nil:
		return ""
	case *TableSymbol:
		return SchemaDisplayText(s)
	case *MaterializedViewSymbol:
		if s == nil {
			return ""
		}
		return SchemaDisplayText(s)
	case *ExternalTableSymbol:
		if s == nil {
			return ""
		}
		return SchemaDisplayText(s)
	case *StoredQueryResultSymbol:
		if s == nil {
			return ""
		}
		return SchemaDisplayText(s)
	default:
		return SchemaDisplayScalarTypeText(typeSymbolFromSymbol(s))
	}
}

func schemaDisplayTableText(table *TableSymbol) string {
	if table == nil {
		return "()"
	}
	columns := table.Columns()
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		if column != nil {
			parts = append(parts, SchemaDisplayText(column))
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func SchemaDisplayScalarTypeText(typ *TypeSymbol) string {
	if typ == nil {
		return ""
	}
	if isDynamicTypedSymbol(typ) {
		return ScalarTypes.Dynamic.SymbolName()
	}
	if scalar := scalarFromType(typ); scalar != nil {
		return scalar.SymbolName()
	}
	return ""
}

func SchemaDisplayParameterTypeText(parameter *Parameter) string {
	if parameter == nil {
		return ScalarTypes.Dynamic.SymbolName()
	}
	if parameter.TypeKind == ParameterTypeTabular {
		return "(*)"
	}
	if parameter.TypeKind == ParameterTypeDeclared && len(parameter.DeclaredTypes) == 1 {
		if table := tableSymbolFromDeclaredSymbol(parameter.DeclaredSymbol(0)); table != nil {
			return schemaDisplayTableText(table)
		}
		if text := SchemaDisplayScalarTypeText(parameter.DeclaredTypes[0]); text != "" {
			return text
		}
	}
	return ScalarTypes.Dynamic.SymbolName()
}

func BracketNameIfNecessary(name string, dialect ...KustoDialect) string {
	if CanBeIdentifier(name, dialect...) {
		return name
	}
	return GetBracketedName(name)
}

func CanBeIdentifier(name string, dialect ...KustoDialect) bool {
	if name == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(name)
	if first == utf8.RuneError || !isIdentifierStart(first) {
		return false
	}
	tokens := Lex(name)
	if len(tokens) != 2 {
		return false
	}
	token := tokens[0]
	if token.Trivia != "" || token.Text != name || len(token.Diagnostics) > 0 {
		return false
	}

	if token.Kind == SyntaxKindBooleanLiteralToken {
		return false
	}

	selectedDialect := KustoDialectQuery
	if len(dialect) > 0 {
		selectedDialect = dialect[0]
	}
	switch selectedDialect {
	case KustoDialectClusterManagerCommand, KustoDialectDataManagerCommand:
		return false
	case KustoDialectEngineCommand:
		return !KustoFactsEngineCommandKeywordNeedsBrackets(name)
	case KustoDialectQuery:
		if token.Kind == SyntaxKindIdentifierToken {
			return true
		}
		return IsKeywordKind(token.Kind) && SyntaxKindCanBeIdentifier(token.Kind)
	default:
		return true
	}
}

// CanBeIdentifierName is retained for callers of the original Go API.
func CanBeIdentifierName(name string, dialect ...KustoDialect) bool {
	return CanBeIdentifier(name, dialect...)
}

func GetBracketedName(name string) string {
	return "[" + GetStringLiteral(name) + "]"
}

func GetStringLiteral(text string) string {
	if strings.Contains(text, "'") {
		return GetDoubleQuotedStringLiteral(text)
	}
	return GetSingleQuotedStringLiteral(text)
}

func GetSingleQuotedStringLiteral(text string) string {
	return "'" + escapeStringLiteral(text, '\'') + "'"
}

func GetDoubleQuotedStringLiteral(text string) string {
	return `"` + escapeStringLiteral(text, '"') + `"`
}

func escapeStringLiteral(text string, quote rune) string {
	var b strings.Builder
	for _, ch := range text {
		switch ch {
		case '\\':
			b.WriteString(`\\`)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if ch == quote {
				b.WriteRune('\\')
			}
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func TableSymbolFromSchema(schema string) (*TableSymbol, error) {
	text := strings.TrimSpace(schema)
	if len(text) > 0 && text[0] != '(' {
		text = "(" + text
	}
	if len(text) > 0 && text[len(text)-1] != ')' {
		text += ")"
	}
	rowSchema := ParseRowSchema(text)
	if rowSchema == nil {
		return nil, fmt.Errorf("invalid table schema %q", schema)
	}
	columns, err := ColumnsFromRowSchema(rowSchema)
	if err != nil {
		return nil, err
	}
	return NewTableSymbol("", columns...), nil
}

func NewTableSymbolFromSchema(name string, schema string, description ...string) (*TableSymbol, error) {
	table, err := TableSymbolFromSchema(schema)
	if err != nil {
		return nil, err
	}
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	return NewTableSymbolWithDescription(name, table.Columns(), desc), nil
}

func NewExternalTableSymbolFromSchema(name string, schema string, description ...string) (*ExternalTableSymbol, error) {
	table, err := TableSymbolFromSchema(schema)
	if err != nil {
		return nil, err
	}
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	return NewExternalTableSymbolWithDescription(name, table.Columns(), desc), nil
}

func NewMaterializedViewSymbolFromSchema(name string, schema string, query string, description ...string) (*MaterializedViewSymbol, error) {
	table, err := TableSymbolFromSchema(schema)
	if err != nil {
		return nil, err
	}
	return NewMaterializedViewSymbol(name, table.Columns(), query, description...), nil
}

func ColumnsFromRowSchema(schema *RowSchema) ([]*ColumnSymbol, error) {
	if schema == nil {
		return nil, fmt.Errorf("invalid row schema")
	}
	declared := map[string]struct{}{}
	var columns []*ColumnSymbol
	for _, element := range schema.Columns().Elements() {
		if element == nil || element.Element() == nil {
			continue
		}
		declaration := element.Element()
		typExpression, ok := declaration.Type().(*PrimitiveTypeExpression)
		if !ok || typExpression == nil {
			continue
		}
		name, err := rowSchemaColumnName(declaration.Name())
		if err != nil {
			name = NameDeclarationSimpleName(declaration.Name())
		}
		if _, ok := declared[name]; ok {
			continue
		}
		declared[name] = struct{}{}
		typ, err := rowSchemaColumnType(typExpression)
		if err != nil {
			typ = &Error.TypeSymbol
		}
		columns = append(columns, NewColumnSymbol(name, typ))
	}
	return columns, nil
}

func validRowSchemaColumnPrefix(schema *RowSchema) []*ColumnSymbol {
	if schema == nil || schema.Columns() == nil {
		return nil
	}
	declared := map[string]struct{}{}
	var columns []*ColumnSymbol
	for _, element := range schema.Columns().Elements() {
		if element == nil || element.Element() == nil {
			break
		}
		declaration := element.Element()
		typExpression, ok := declaration.Type().(*PrimitiveTypeExpression)
		if !ok || typExpression == nil || typExpression.Type() == nil || len(declaration.ContainedDiagnostics()) > 0 {
			break
		}
		name, nameErr := rowSchemaColumnName(declaration.Name())
		typ, typeErr := rowSchemaColumnType(typExpression)
		if nameErr != nil || typeErr != nil {
			break
		}
		if _, exists := declared[name]; exists {
			break
		}
		declared[name] = struct{}{}
		columns = append(columns, NewColumnSymbol(name, typ))
	}
	return columns
}

func rowSchemaColumnType(expression *PrimitiveTypeExpression) (*TypeSymbol, error) {
	token := expression.Type()
	if token == nil {
		return nil, fmt.Errorf("missing column type")
	}
	scalar := ScalarTypes.Get(token.ValueText())
	if scalar == nil {
		return nil, fmt.Errorf("unknown table schema type %q", token.ValueText())
	}
	return scalar.Type(), nil
}

func rowSchemaColumnName(declaration *NameDeclaration) (string, error) {
	if declaration == nil {
		return "", fmt.Errorf("missing column name")
	}
	switch name := declaration.Name().(type) {
	case *TokenName:
		if name.Name() == nil {
			return "", fmt.Errorf("missing column name")
		}
		return name.Name().ValueText(), nil
	case *BracketedName:
		literal, ok := name.Name().(*LiteralExpression)
		if !ok || literal.Token() == nil {
			return "", fmt.Errorf("invalid bracketed column name")
		}
		return literal.Token().ValueText(), nil
	case *BracedName:
		if name.Name() == nil {
			return "", fmt.Errorf("missing column name")
		}
		return name.Name().ValueText(), nil
	default:
		return "", fmt.Errorf("unsupported column name syntax %T", name)
	}
}
