package environment

import (
	"reflect"
	"sort"
	"strings"

	kustolang "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

type Column struct {
	Type                     string
	ExactType                *kustolang.TypeSymbol
	DynamicSelectorShape     string
	NonScalar                bool
	Origin                   bool
	ReferenceID              string
	ElementType              string
	ElementElementType       string
	Properties               DynamicProperties
	ElementProperties        DynamicProperties
	ElementElementProperties DynamicProperties
}

type DynamicProperty struct {
	Type              string
	ElementType       string
	Properties        DynamicProperties
	ElementProperties DynamicProperties
}

type DynamicProperties map[string]DynamicProperty

type terminalResultKind uint8

const (
	terminalResultUnknown terminalResultKind = iota
	terminalResultTabular
	terminalResultNonTabular
)

type TableSchema struct {
	Columns           map[string]Column
	ColumnCandidates  map[string][]Column
	Order             []string
	Open              bool
	Serialized        bool
	Sorted            bool
	MissingColumnType string
	TerminalTables    []ResultTable
	resultKind        terminalResultKind
}

func NewTableSchema() TableSchema {
	return TableSchema{Columns: map[string]Column{}, ColumnCandidates: map[string][]Column{}, Order: []string{}, resultKind: terminalResultTabular}
}

func NewOpenTableSchema() TableSchema {
	table := NewTableSchema()
	table.Open = true
	return table
}

func NewNonTabularSchema() TableSchema {
	schema := NewTableSchema()
	schema.resultKind = terminalResultNonTabular
	return schema
}

func NewNonTabularOpenSchema() TableSchema {
	schema := NewOpenTableSchema()
	schema.resultKind = terminalResultNonTabular
	return schema
}

func (t TableSchema) Clone() TableSchema {
	out := NewTableSchema()
	out.Open = t.Open
	out.Serialized = t.Serialized
	out.Sorted = t.Sorted
	out.MissingColumnType = t.MissingColumnType
	out.Order = append(out.Order, t.Order...)
	out.TerminalTables = cloneResultTables(t.TerminalTables)
	out.resultKind = t.resultKind
	for name, col := range t.Columns {
		out.Columns[name] = col.Clone()
	}
	for name, candidates := range t.ColumnCandidates {
		out.ColumnCandidates[name] = cloneColumns(candidates)
	}
	return out
}

func (t TableSchema) ResultTables() []ResultTable {
	if t.resultKind != terminalResultTabular {
		return []ResultTable{}
	}
	if t.TerminalTables != nil {
		return cloneResultTables(t.TerminalTables)
	}
	return []ResultTable{resultTableFromSchema("", t)}
}

func (t *TableSchema) SetResultTables(tables []ResultTable) {
	t.TerminalTables = cloneResultTables(tables)
	t.resultKind = terminalResultTabular
}

func (t *TableSchema) MarkNonTabularResult() {
	t.TerminalTables = nil
	t.resultKind = terminalResultNonTabular
}

func resultTableFromSchema(name string, schema TableSchema) ResultTable {
	columns := make([]ResultColumn, 0, len(schema.Order))
	candidateIndexes := make(map[string]int, len(schema.ColumnCandidates))
	for _, columnName := range schema.Order {
		column, ok := schema.Columns[columnName]
		if candidates := schema.ColumnCandidates[columnName]; len(candidates) > 0 {
			index := candidateIndexes[columnName]
			if index < len(candidates) {
				column = candidates[index]
				ok = true
				candidateIndexes[columnName] = index + 1
			}
		}
		if ok {
			columns = append(columns, ResultColumn{Name: columnName, Type: column.Type})
		}
	}
	if len(columns) == 0 && len(schema.Columns) > 0 {
		names := make([]string, 0, len(schema.Columns))
		for columnName := range schema.Columns {
			names = append(names, columnName)
		}
		sort.Strings(names)
		for _, columnName := range names {
			columns = append(columns, ResultColumn{Name: columnName, Type: schema.Columns[columnName].Type})
		}
	}
	return ResultTable{Name: name, Columns: columns, IsOpen: schema.Open}
}

func resultTableFromTableSymbol(table *kustolang.TableSymbol) ResultTable {
	result := ResultTable{Columns: []ResultColumn{}}
	if table == nil {
		return result
	}
	result.Name = table.SymbolName()
	result.IsOpen = table.IsOpen()
	for _, column := range table.Columns() {
		if column != nil {
			result.Columns = append(result.Columns, ResultColumn{Name: column.SymbolName(), Type: publicTypeName(column.Type)})
		}
	}
	return result
}

func cloneResultTables(tables []ResultTable) []ResultTable {
	if tables == nil {
		return nil
	}
	out := make([]ResultTable, len(tables))
	for i, table := range tables {
		out[i] = table
		out[i].Columns = append([]ResultColumn{}, table.Columns...)
	}
	return out
}

func (t *TableSchema) Set(name, typ string, origin bool) {
	t.SetWithProperties(name, typ, origin, nil)
}

func (t *TableSchema) SetUnnamed(typ string, origin bool) {
	t.SetUnnamedColumn(Column{Type: normalizeType(typ), Origin: origin})
}

func (t *TableSchema) SetUnnamedColumn(column Column) {
	if _, exists := t.Columns[""]; !exists {
		t.Order = append(t.Order, "")
	}
	t.Columns[""] = column.Clone()
}

func (t *TableSchema) SetExactType(name string, typ *kustolang.TypeSymbol) {
	name = strings.TrimSpace(name)
	column, ok := t.Columns[name]
	if !ok {
		return
	}
	column.ExactType = typ
	t.Columns[name] = column
}

func (t *TableSchema) SetColumn(name string, column Column) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	t.SetWithNestedDynamicProperties(
		name,
		typeNameForSchemaCopy(column.Type),
		column.Origin,
		column.Properties,
		column.ElementProperties,
		column.ElementElementProperties,
		column.ElementType,
		column.ElementElementType,
	)
	t.SetExactType(name, column.ExactType)
	stored := t.Columns[name]
	stored.ReferenceID = column.ReferenceID
	stored.NonScalar = column.NonScalar
	stored.DynamicSelectorShape = column.DynamicSelectorShape
	t.Columns[name] = stored
}

func (t *TableSchema) AppendColumn(name string, column Column) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if existing, ok := t.Columns[name]; ok {
		if len(t.ColumnCandidates[name]) == 0 {
			t.ColumnCandidates[name] = []Column{existing.Clone()}
		}
		t.ColumnCandidates[name] = append(t.ColumnCandidates[name], column.Clone())
		t.Order = append(t.Order, name)
		return
	}
	t.SetColumn(name, column)
}

func (t TableSchema) Candidates(name string) []Column {
	name = strings.TrimSpace(name)
	if candidates := t.ColumnCandidates[name]; len(candidates) > 0 {
		return cloneColumns(candidates)
	}
	if column, ok := t.Columns[name]; ok {
		return []Column{column.Clone()}
	}
	return nil
}

func (t TableSchema) IsAmbiguous(name string) bool {
	return len(t.ColumnCandidates[strings.TrimSpace(name)]) > 1
}

func (t *TableSchema) SetWithProperties(name, typ string, origin bool, properties DynamicProperties) {
	t.SetWithDynamicProperties(name, typ, origin, properties, nil)
}

func (t *TableSchema) SetWithDynamicProperties(name, typ string, origin bool, properties DynamicProperties, elementProperties DynamicProperties, elementType ...string) {
	t.SetWithNestedDynamicProperties(name, typ, origin, properties, elementProperties, nil, elementType...)
}

func (t *TableSchema) SetWithNestedDynamicProperties(name, typ string, origin bool, properties DynamicProperties, elementProperties DynamicProperties, elementElementProperties DynamicProperties, elementType ...string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	normalizedType := normalizeType(typ)
	if typ == emptyTableTypeName {
		normalizedType = ""
	} else if strings.HasPrefix(typ, nonScalarNamedTypePrefix) {
		normalizedType = strings.TrimPrefix(typ, nonScalarNamedTypePrefix)
	}
	if _, exists := t.Columns[name]; !exists {
		t.Order = append(t.Order, name)
	}
	normalizedElementType := ""
	if len(elementType) > 0 && strings.TrimSpace(elementType[0]) != "" {
		normalizedElementType = normalizeType(elementType[0])
	}
	normalizedElementElementType := ""
	if len(elementType) > 1 && strings.TrimSpace(elementType[1]) != "" {
		normalizedElementElementType = normalizeType(elementType[1])
	}
	t.Columns[name] = Column{
		Type:                     normalizedType,
		NonScalar:                typ == emptyTableTypeName || strings.HasPrefix(typ, nonScalarNamedTypePrefix),
		Origin:                   origin,
		ElementType:              normalizedElementType,
		ElementElementType:       normalizedElementElementType,
		Properties:               cloneColumnProperties(properties),
		ElementProperties:        cloneColumnProperties(elementProperties),
		ElementElementProperties: cloneColumnProperties(elementElementProperties),
	}
}

func (t *TableSchema) Delete(name string) {
	if _, exists := t.Columns[name]; !exists {
		return
	}
	delete(t.Columns, name)
	delete(t.ColumnCandidates, name)
	order := t.Order[:0]
	for _, existing := range t.Order {
		if existing != name {
			order = append(order, existing)
		}
	}
	t.Order = order
}

func (t TableSchema) Get(name string) (Column, bool) {
	name = strings.TrimSpace(name)
	if col, ok := t.Columns[name]; ok {
		return col, true
	}
	if t.Open && isIdentifierLike(name) {
		return Column{Type: "unknown"}, true
	}
	return Column{}, false
}

func cloneColumns(columns []Column) []Column {
	out := make([]Column, len(columns))
	for i, column := range columns {
		out[i] = column.Clone()
	}
	return out
}

func (t TableSchema) Has(name string) bool {
	_, ok := t.Get(name)
	return ok
}

func (t TableSchema) ToMap() map[string]string {
	out := map[string]string{}
	names := append([]string{}, t.Order...)
	if len(names) == 0 && len(t.Columns) > 0 {
		for name := range t.Columns {
			names = append(names, name)
		}
		sort.Strings(names)
	}
	for _, name := range names {
		if col, ok := t.Columns[name]; ok {
			out[name] = col.Type
		}
	}
	return out
}

func (t TableSchema) PublishedColumnsAndTables() (map[string]string, []ResultTable) {
	columns := t.ToMap()
	tables := t.ResultTables()

	errorType := kustolang.Error.SymbolName()
	unknownType := kustolang.ScalarTypes.Unknown.SymbolName()
	for name, typ := range columns {
		if typ == errorType && !schemaColumnPreservesError(t.Columns[name]) {
			columns[name] = unknownType
		}
	}
	for tableIndex := range tables {
		for columnIndex := range tables[tableIndex].Columns {
			column := tables[tableIndex].Columns[columnIndex]
			if column.Type == errorType && !schemaColumnPreservesError(t.Columns[column.Name]) {
				tables[tableIndex].Columns[columnIndex].Type = unknownType
			}
		}
	}
	return columns, tables
}

func schemaColumnPreservesError(column Column) bool {
	return column.ExactType != nil && column.ExactType.SymbolName() == kustolang.Error.SymbolName()
}

func (t *TableSchema) Merge(other TableSchema) {
	if t.MissingColumnType == "" {
		t.MissingColumnType = other.MissingColumnType
	} else if other.MissingColumnType == "" || normalizeType(t.MissingColumnType) != normalizeType(other.MissingColumnType) {
		t.MissingColumnType = ""
	}
	if other.Open {
		t.Open = true
	}
	for _, name := range other.Order {
		col := other.Columns[name]
		if existing, exists := t.Columns[name]; !exists {
			t.SetColumn(name, col)
		} else if existing.Type != col.Type {
			uniqueName := uniqueColumnName(*t, name)
			t.SetColumn(uniqueName, col)
		} else {
			t.Columns[name] = columnMetadataConsensus([]Column{existing, col})
		}
	}
}

func columnMetadataConsensus(columns []Column) Column {
	if len(columns) == 0 {
		return Column{}
	}
	column := columns[0].Clone()
	exactConsensus := true
	elementTypeConsensus := true
	elementElementTypeConsensus := true
	propertiesConsensus := true
	elementPropertiesConsensus := true
	elementElementPropertiesConsensus := true
	for _, candidate := range columns[1:] {
		if candidate.Type != column.Type {
			return Column{Type: "dynamic"}
		}
		if candidate.ReferenceID != column.ReferenceID {
			column.ReferenceID = ""
		}
		if exactConsensus && candidate.ExactType != column.ExactType {
			exactConsensus = false
			column.ExactType = nil
			column.ElementType = ""
			column.ElementElementType = ""
			column.Properties = nil
			column.ElementProperties = nil
			column.ElementElementProperties = nil
			elementTypeConsensus = false
			elementElementTypeConsensus = false
			propertiesConsensus = false
			elementPropertiesConsensus = false
			elementElementPropertiesConsensus = false
			continue
		}
		if elementTypeConsensus && candidate.ElementType != column.ElementType {
			elementTypeConsensus = false
			column.ElementType = ""
		}
		if elementElementTypeConsensus && candidate.ElementElementType != column.ElementElementType {
			elementElementTypeConsensus = false
			column.ElementElementType = ""
		}
		if propertiesConsensus && !reflect.DeepEqual(candidate.Properties, column.Properties) {
			propertiesConsensus = false
			column.Properties = nil
		}
		if elementPropertiesConsensus && !reflect.DeepEqual(candidate.ElementProperties, column.ElementProperties) {
			elementPropertiesConsensus = false
			column.ElementProperties = nil
		}
		if elementElementPropertiesConsensus && !reflect.DeepEqual(candidate.ElementElementProperties, column.ElementElementProperties) {
			elementElementPropertiesConsensus = false
			column.ElementElementProperties = nil
		}
	}
	return column
}

func (c Column) Clone() Column {
	c.Properties = cloneColumnProperties(c.Properties)
	c.ElementProperties = cloneColumnProperties(c.ElementProperties)
	c.ElementElementProperties = cloneColumnProperties(c.ElementElementProperties)
	return c
}

func cloneColumnProperties(properties DynamicProperties) DynamicProperties {
	if len(properties) == 0 {
		return nil
	}
	out := make(DynamicProperties, len(properties))
	for name, property := range properties {
		elementType := ""
		if strings.TrimSpace(property.ElementType) != "" {
			elementType = normalizeType(property.ElementType)
		}
		out[name] = DynamicProperty{
			Type:              normalizeType(property.Type),
			ElementType:       elementType,
			Properties:        cloneColumnProperties(property.Properties),
			ElementProperties: cloneColumnProperties(property.ElementProperties),
		}
	}
	return out
}

func uniqueColumnName(t TableSchema, name string) string {
	if _, ok := t.Columns[name]; !ok {
		return name
	}
	for i := 1; ; i++ {
		candidate := name + intSuffix(i)
		if _, ok := t.Columns[candidate]; !ok {
			return candidate
		}
	}
}

func intSuffix(i int) string {
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	var out []byte
	for i > 0 {
		out = append([]byte{digits[i%10]}, out...)
		i /= 10
	}
	return string(out)
}

func normalizeType(typ string) string {
	return kustolang.ScalarTypes.CanonicalName(typ)
}

func foldName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
