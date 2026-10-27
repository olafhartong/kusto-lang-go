package kustolang

import (
	"fmt"
	"strings"
	"sync"
)

// Ported from the core symbol model under src/Kusto.Language/Symbols.
// This is the initial Go foundation for the official Kusto.Language port.

type SymbolKind string

const (
	SymbolKindNone                   SymbolKind = "None"
	SymbolKindDatabase               SymbolKind = "Database"
	SymbolKindTable                  SymbolKind = "Table"
	SymbolKindColumn                 SymbolKind = "Column"
	SymbolKindFunction               SymbolKind = "Function"
	SymbolKindVariable               SymbolKind = "Variable"
	SymbolKindParameter              SymbolKind = "Parameter"
	SymbolKindOption                 SymbolKind = "Option"
	SymbolKindOperator               SymbolKind = "Operator"
	SymbolKindPattern                SymbolKind = "Pattern"
	SymbolKindTuple                  SymbolKind = "Tuple"
	SymbolKindBag                    SymbolKind = "Bag"
	SymbolKindArray                  SymbolKind = "Array"
	SymbolKindCluster                SymbolKind = "Cluster"
	SymbolKindGroup                  SymbolKind = "Group"
	SymbolKindVoid                   SymbolKind = "Void"
	SymbolKindCommand                SymbolKind = "Command"
	SymbolKindGraph                  SymbolKind = "Graph"
	SymbolKindPrimitive              SymbolKind = "Primitive"
	SymbolKindMaterializedView       SymbolKind = "MaterializedView"
	SymbolKindQueryOperatorParameter SymbolKind = "QueryOperatorParameter"
	SymbolKindEntityGroup            SymbolKind = "EntityGroup"
	SymbolKindEntityGroupElement     SymbolKind = "EntityGroupElement"
	SymbolKindStoredQueryResult      SymbolKind = "StoredQueryResult"
	SymbolKindGraphModel             SymbolKind = "GraphModel"
	SymbolKindGraphSnapshot          SymbolKind = "GraphSnapshot"
	SymbolKindType                   SymbolKind = "Type"
	SymbolKindScalar                 SymbolKind = "Scalar"
	SymbolKindError                  SymbolKind = "Error"
)

type Tabularity string

const (
	TabularityNone        Tabularity = "None"
	TabularityScalar      Tabularity = "Scalar"
	TabularityTabular     Tabularity = "Tabular"
	TabularityOther       Tabularity = "Other"
	TabularityUnknown     Tabularity = "Unknown"
	TabularityUnspecified Tabularity = "Unspecified"
)

type Symbol interface {
	SymbolName() string
	AlternateName() string
	SymbolKind() SymbolKind
	SymbolTabularity() Tabularity
	Members() []Symbol
	IsHidden() bool
	IsError() bool
}

type BaseSymbol struct {
	name string
}

func NewBaseSymbol(name string) BaseSymbol {
	return BaseSymbol{name: name}
}

func (s BaseSymbol) SymbolName() string {
	return s.name
}

func (s BaseSymbol) AlternateName() string {
	return ""
}

func (s BaseSymbol) SymbolKind() SymbolKind {
	return SymbolKindNone
}

func (s BaseSymbol) SymbolTabularity() Tabularity {
	return TabularityUnknown
}

func (s BaseSymbol) Members() []Symbol {
	return nil
}

func (s BaseSymbol) IsHidden() bool {
	return strings.HasPrefix(s.name, "__")
}

func (s BaseSymbol) IsError() bool {
	return false
}

func IsScalar(symbol Symbol) bool {
	if symbol == nil {
		return false
	}
	switch symbol.SymbolTabularity() {
	case TabularityScalar, TabularityUnknown:
		return true
	default:
		return false
	}
}

func IsTabular(symbol Symbol) bool {
	if symbol == nil {
		return false
	}
	switch symbol.SymbolTabularity() {
	case TabularityTabular, TabularityUnknown:
		return true
	default:
		return false
	}
}

// GetResultType mirrors Symbol.GetResultType from the upstream reference implementation model. It
// returns the type produced by referencing a symbol, with GroupSymbol results
// collapsed when only one member contributes a result type.
func GetResultType(symbol Symbol) Symbol {
	switch s := symbol.(type) {
	case nil:
		return nil
	case *ColumnSymbol:
		return s.Type
	case *VariableSymbol:
		return GetResultType(s.Type)
	case *EntityGroupElementSymbol:
		return GetResultType(s.UnderlyingSymbol)
	case *ParameterSymbol:
		return s.Type
	case *GroupSymbol:
		var results []Symbol
		for _, member := range s.Members() {
			if result := GetResultType(member); result != nil {
				results = append(results, result)
			}
		}
		switch len(results) {
		case 0:
			return nil
		case 1:
			if isResultTypeSymbol(results[0]) {
				return results[0]
			}
			return nil
		default:
			return NewGroupSymbol(results...)
		}
	default:
		if isResultTypeSymbol(symbol) {
			return symbol
		}
		return nil
	}
}

func isResultTypeSymbol(symbol Symbol) bool {
	switch symbol.(type) {
	case *TypeSymbol,
		*ScalarSymbol,
		*DynamicAnySymbol,
		*DynamicPrimitiveSymbol,
		*DynamicArraySymbol,
		*DynamicBagSymbol,
		*TupleSymbol,
		*ErrorSymbol,
		*VoidSymbol,
		*TableSymbol,
		*MaterializedViewSymbol,
		*ExternalTableSymbol,
		*StoredQueryResultSymbol,
		*DatabaseSymbol,
		*ClusterSymbol,
		*EntityGroupSymbol,
		*EntityGroupElementSymbol,
		*GraphSymbol,
		*GraphModelSymbol,
		*GroupSymbol,
		*PatternSymbol,
		*FunctionSymbol:
		return true
	default:
		return false
	}
}

type TypeSymbol struct {
	BaseSymbol
	dynamicPrimitive *DynamicPrimitiveSymbol
	dynamicBag       *DynamicBagSymbol
	dynamicArray     *DynamicArraySymbol
	tuple            *TupleSymbol
	table            *TableSymbol
}

func NewTypeSymbol(name string) *TypeSymbol {
	return &TypeSymbol{BaseSymbol: NewBaseSymbol(name)}
}

func (s *TypeSymbol) SymbolKind() SymbolKind {
	if table, ok := TableSymbolForType(s); ok {
		return table.SymbolKind()
	}
	return SymbolKindType
}

func (s *TypeSymbol) SymbolTabularity() Tabularity {
	if _, ok := TableSymbolForType(s); ok {
		return TabularityTabular
	}
	return TabularityScalar
}

func (s *TypeSymbol) Members() []Symbol {
	if table, ok := TableSymbolForType(s); ok {
		return table.Members()
	}
	return nil
}

type ScalarFlag uint64

const (
	ScalarFlagNone    ScalarFlag = 0
	ScalarFlagInteger ScalarFlag = 1 << (iota - 1)
	ScalarFlagNumeric
	ScalarFlagInterval
	ScalarFlagSummable
	ScalarFlagOrderable
	ScalarFlagAll = ScalarFlagInteger | ScalarFlagNumeric | ScalarFlagInterval | ScalarFlagSummable | ScalarFlagOrderable
)

type ScalarSymbol struct {
	TypeSymbol
	aliases   []string
	flags     ScalarFlag
	widerThan []*ScalarSymbol
}

func NewScalarSymbol(name string, aliases []string, flags ScalarFlag, widerThan ...*ScalarSymbol) *ScalarSymbol {
	for _, scalar := range widerThan {
		if scalar == nil {
			panic(ArgumentError{ParameterName: "widerThan", Message: "widerThan contains nil"})
		}
	}
	return &ScalarSymbol{
		TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
		aliases:    append([]string{}, aliases...),
		flags:      flags,
		widerThan:  append([]*ScalarSymbol{}, widerThan...),
	}
}

func (s *ScalarSymbol) SymbolKind() SymbolKind {
	return SymbolKindPrimitive
}

func (s *ScalarSymbol) Aliases() []string {
	return append([]string{}, s.aliases...)
}

func (s *ScalarSymbol) Flags() ScalarFlag {
	if s == nil {
		return ScalarFlagNone
	}
	return s.flags
}

func (s *ScalarSymbol) IsInteger() bool {
	return s.Flags()&ScalarFlagInteger != 0
}

func (s *ScalarSymbol) IsNumeric() bool {
	return s.Flags()&ScalarFlagNumeric != 0
}

func (s *ScalarSymbol) IsInterval() bool {
	return s.Flags()&ScalarFlagInterval != 0
}

func (s *ScalarSymbol) IsSummable() bool {
	return s.Flags()&ScalarFlagSummable != 0
}

func (s *ScalarSymbol) IsOrderable() bool {
	return s.Flags()&ScalarFlagOrderable != 0
}

func (s *ScalarSymbol) IsMultiValue() bool {
	return false
}

func (s *ScalarSymbol) IsWiderThan(other *ScalarSymbol) bool {
	if s == nil || other == nil {
		return false
	}
	for _, candidate := range s.widerThan {
		if candidate == other {
			return true
		}
	}
	return false
}

type ColumnSymbol struct {
	BaseSymbol
	Type            *TypeSymbol
	Description     string
	OriginalColumns []*ColumnSymbol
	Source          SyntaxElement
	Examples        []string
}

func NewColumnSymbol(name string, typ *TypeSymbol) *ColumnSymbol {
	if typ == nil || isErrorTypeSymbol(typ) {
		typ = ScalarTypes.Unknown.Type()
	}
	return &ColumnSymbol{BaseSymbol: NewBaseSymbol(name), Type: typ}
}

func NewColumnSymbolFull(name string, typ *TypeSymbol, description string, originalColumns []*ColumnSymbol, source SyntaxElement, examples []string) *ColumnSymbol {
	if typ == nil || isErrorTypeSymbol(typ) {
		typ = ScalarTypes.Unknown.Type()
	}
	return &ColumnSymbol{
		BaseSymbol:      NewBaseSymbol(name),
		Type:            typ,
		Description:     description,
		OriginalColumns: trulyOriginalColumns(originalColumns),
		Source:          source,
		Examples:        append([]string{}, examples...),
	}
}

func isErrorTypeSymbol(typ *TypeSymbol) bool {
	return typ != nil && (typ.IsError() || typ == &Error.TypeSymbol)
}

func NewColumnSymbolWithSource(name string, typ *TypeSymbol, source SyntaxElement) *ColumnSymbol {
	column := NewColumnSymbol(name, typ)
	column.Source = source
	return column
}

func (s *ColumnSymbol) SymbolKind() SymbolKind {
	return SymbolKindColumn
}

func (s *ColumnSymbol) SymbolTabularity() Tabularity {
	return TabularityScalar
}

func (s *ColumnSymbol) WithName(name string) *ColumnSymbol {
	if s == nil {
		return nil
	}
	if name == s.SymbolName() {
		return s
	}
	return NewColumnSymbolFull(name, s.Type, s.Description, s.OriginalColumns, s.Source, s.Examples)
}

func (s *ColumnSymbol) WithType(typ *TypeSymbol) *ColumnSymbol {
	if s == nil {
		return nil
	}
	if typ == s.Type {
		return s
	}
	return NewColumnSymbolFull(s.SymbolName(), typ, s.Description, s.OriginalColumns, s.Source, s.Examples)
}

func (s *ColumnSymbol) WithDescription(description string) *ColumnSymbol {
	if s == nil {
		return nil
	}
	if description == s.Description {
		return s
	}
	return NewColumnSymbolFull(s.SymbolName(), s.Type, description, s.OriginalColumns, s.Source, s.Examples)
}

func (s *ColumnSymbol) WithOriginalColumns(originalColumns ...*ColumnSymbol) *ColumnSymbol {
	if s == nil {
		return nil
	}
	return NewColumnSymbolFull(s.SymbolName(), s.Type, s.Description, originalColumns, s.Source, s.Examples)
}

func (s *ColumnSymbol) WithSource(source SyntaxElement) *ColumnSymbol {
	if s == nil {
		return nil
	}
	if source == s.Source {
		return s
	}
	return NewColumnSymbolFull(s.SymbolName(), s.Type, s.Description, s.OriginalColumns, source, s.Examples)
}

func (s *ColumnSymbol) WithExamples(examples []string) *ColumnSymbol {
	if s == nil {
		return nil
	}
	return NewColumnSymbolFull(s.SymbolName(), s.Type, s.Description, s.OriginalColumns, s.Source, examples)
}

func trulyOriginalColumns(columns []*ColumnSymbol) []*ColumnSymbol {
	if len(columns) == 0 {
		return nil
	}
	out := []*ColumnSymbol{}
	for _, column := range columns {
		if column == nil {
			continue
		}
		if len(column.OriginalColumns) > 0 {
			out = append(out, column.OriginalColumns...)
		} else {
			out = append(out, column)
		}
	}
	return out
}

type VariableSymbol struct {
	BaseSymbol
	Type              Symbol
	IsConstant        bool
	ConstantValueInfo *ValueInfo
	Source            Expression
}

func NewVariableSymbol(name string, typ Symbol, isConstant bool, constantValueInfo any, source Expression) *VariableSymbol {
	if typ == nil {
		typ = ScalarTypes.Unknown.Type()
	}
	return &VariableSymbol{
		BaseSymbol:        NewBaseSymbol(name),
		Type:              typ,
		IsConstant:        isConstant,
		ConstantValueInfo: normalizeVariableConstantValueInfo(constantValueInfo),
		Source:            source,
	}
}

// NewVariableSymbolWithValueInfo is the source-shaped constructor form. The
// any-typed constructor remains as a compatibility adapter for existing Go
// callers that supplied raw constant values before ValueInfo was public.
func NewVariableSymbolWithValueInfo(name string, typ Symbol, isConstant bool, constantValueInfo *ValueInfo, source Expression) *VariableSymbol {
	return NewVariableSymbol(name, typ, isConstant, constantValueInfo, source)
}

func normalizeVariableConstantValueInfo(value any) *ValueInfo {
	switch value := value.(type) {
	case nil:
		return nil
	case *ValueInfo:
		return value
	default:
		text := fmt.Sprint(value)
		return NewValueInfo(text, text, value)
	}
}

func (s *VariableSymbol) SymbolKind() SymbolKind {
	return SymbolKindVariable
}

func (s *VariableSymbol) SymbolTabularity() Tabularity {
	if s == nil || s.Type == nil {
		return TabularityUnknown
	}
	return s.Type.SymbolTabularity()
}

func (s *VariableSymbol) ConstantValue() any {
	if s == nil || s.ConstantValueInfo == nil {
		return nil
	}
	return s.ConstantValueInfo.Value
}

type ParameterSymbol struct {
	BaseSymbol
	Type        Symbol
	description string
}

func NewParameterSymbol(name string, typ Symbol, description string) *ParameterSymbol {
	if typ == nil {
		typ = ScalarTypes.Unknown.Type()
	}
	return &ParameterSymbol{BaseSymbol: NewBaseSymbol(name), Type: typ, description: description}
}

func (s *ParameterSymbol) SymbolKind() SymbolKind {
	return SymbolKindParameter
}

func (s *ParameterSymbol) SymbolTabularity() Tabularity {
	if s == nil || s.Type == nil {
		return TabularityUnknown
	}
	return s.Type.SymbolTabularity()
}

func (s *ParameterSymbol) Description() string {
	if s == nil {
		return ""
	}
	return s.description
}

type OptionSymbol struct {
	BaseSymbol
	description string
	types       []*ScalarSymbol
	examples    []string
}

func NewOptionSymbol(name, description string, types []*ScalarSymbol, examples []string) *OptionSymbol {
	return &OptionSymbol{
		BaseSymbol:  NewBaseSymbol(name),
		description: description,
		types:       append([]*ScalarSymbol{}, types...),
		examples:    append([]string{}, examples...),
	}
}

func NewOptionSymbolWithType(name, description string, typ *ScalarSymbol, examples []string) *OptionSymbol {
	var types []*ScalarSymbol
	if typ != nil {
		types = []*ScalarSymbol{typ}
	}
	return NewOptionSymbol(name, description, types, examples)
}

func (s *OptionSymbol) SymbolKind() SymbolKind {
	return SymbolKindOption
}

func (s *OptionSymbol) SymbolTabularity() Tabularity {
	return TabularityScalar
}

func (s *OptionSymbol) Description() string {
	if s == nil {
		return ""
	}
	return s.description
}

func (s *OptionSymbol) Types() []*ScalarSymbol {
	if s == nil {
		return nil
	}
	return append([]*ScalarSymbol{}, s.types...)
}

func (s *OptionSymbol) Examples() []string {
	if s == nil {
		return nil
	}
	return append([]string{}, s.examples...)
}

type TableSymbol struct {
	TypeSymbol
	columns             []*ColumnSymbol
	description         string
	isOpen              bool
	isSerialized        bool
	isSorted            bool
	isExternal          bool
	isMaterializedView  bool
	isStoredQueryResult bool
}

func NewTableSymbol(name string, columns ...*ColumnSymbol) *TableSymbol {
	return newTableSymbol(name, columns, "")
}

func NewTableSymbolWithDescription(name string, columns []*ColumnSymbol, description string) *TableSymbol {
	return newTableSymbol(name, columns, description)
}

func newTableSymbol(name string, columns []*ColumnSymbol, description string) *TableSymbol {
	table := &TableSymbol{TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)}, columns: append([]*ColumnSymbol{}, columns...), description: description}
	table.TypeSymbol.table = table
	return table
}

func (s *TableSymbol) SymbolKind() SymbolKind {
	if s != nil && s.isStoredQueryResult {
		return SymbolKindStoredQueryResult
	}
	if s != nil && s.isMaterializedView {
		return SymbolKindMaterializedView
	}
	return SymbolKindTable
}

func (s *TableSymbol) SymbolTabularity() Tabularity {
	return TabularityTabular
}

func (s *TableSymbol) Columns() []*ColumnSymbol {
	return append([]*ColumnSymbol{}, s.columns...)
}

func (s *TableSymbol) GetColumn(name string) *ColumnSymbol {
	column, _ := s.TryGetColumn(name)
	return column
}

func (s *TableSymbol) TryGetColumn(name string) (*ColumnSymbol, bool) {
	if s == nil {
		return nil, false
	}
	for _, column := range s.columns {
		if column != nil && column.SymbolName() == name {
			return column, true
		}
	}
	return nil, false
}

func (s *TableSymbol) GetMatchingColumns(pattern string) []*ColumnSymbol {
	if s == nil {
		return nil
	}
	if !strings.Contains(pattern, "*") {
		if column, ok := s.TryGetColumn(pattern); ok {
			return []*ColumnSymbol{column}
		}
		return nil
	}
	var columns []*ColumnSymbol
	for _, column := range s.columns {
		if column != nil && KustoFactsMatches(pattern, column.SymbolName()) {
			columns = append(columns, column)
		}
	}
	return columns
}

func (s *TableSymbol) Description() string {
	return s.description
}

func (s *TableSymbol) WithName(name string) *TableSymbol {
	if s == nil {
		return nil
	}
	if name == s.SymbolName() {
		return s
	}
	clone := s.clone()
	clone.TypeSymbol = TypeSymbol{BaseSymbol: NewBaseSymbol(name)}
	clone.TypeSymbol.table = clone
	return clone
}

func (s *TableSymbol) WithDescription(description string) *TableSymbol {
	if s == nil {
		return nil
	}
	if description == s.description {
		return s
	}
	clone := s.clone()
	clone.description = description
	return clone
}

// WithDescripton retains the misspelled method name exposed by Kusto.Language.
func (s *TableSymbol) WithDescripton(description string) *TableSymbol {
	return s.WithDescription(description)
}

func (s *TableSymbol) WithColumns(columns []*ColumnSymbol) *TableSymbol {
	if s == nil {
		return nil
	}
	if sameColumnSlice(s.columns, columns) {
		return s
	}
	clone := s.clone()
	clone.columns = append([]*ColumnSymbol{}, columns...)
	return clone
}

func (s *TableSymbol) AddColumns(columns ...*ColumnSymbol) *TableSymbol {
	if s == nil {
		return nil
	}
	if len(columns) == 0 {
		return s
	}
	newColumns := append(s.Columns(), columns...)
	return s.WithColumns(newColumns)
}

func (s *TableSymbol) WithSource(source SyntaxElement) *TableSymbol {
	if s == nil {
		return nil
	}
	columns := make([]*ColumnSymbol, 0, len(s.columns))
	changed := false
	for _, column := range s.columns {
		if column == nil {
			columns = append(columns, nil)
			continue
		}
		sourced := column.WithSource(source)
		if sourced != column {
			changed = true
		}
		columns = append(columns, sourced)
	}
	if !changed {
		return s
	}
	return s.WithColumns(columns)
}

func (s *TableSymbol) WithInheritableProperties(table *TableSymbol) *TableSymbol {
	if s == nil || table == nil {
		return s
	}
	if s.isSerialized == table.isSerialized && s.isSorted == table.isSorted && s.isOpen == table.isOpen {
		return s
	}
	clone := s.clone()
	clone.isSerialized = table.isSerialized
	clone.isSorted = table.isSorted
	clone.isOpen = table.isOpen
	return clone
}

func (s *TableSymbol) IsOpen() bool {
	return s.isOpen
}

func (s *TableSymbol) WithIsOpen(isOpen bool) *TableSymbol {
	if s == nil || s.isOpen == isOpen {
		return s
	}
	clone := s.clone()
	clone.isOpen = isOpen
	return clone
}

func (s *TableSymbol) IsSerialized() bool {
	return s != nil && s.isSerialized
}

func (s *TableSymbol) WithIsSerialized(isSerialized bool) *TableSymbol {
	if s == nil || s.isSerialized == isSerialized {
		return s
	}
	clone := s.clone()
	clone.isSerialized = isSerialized
	return clone
}

func (s *TableSymbol) IsSorted() bool {
	return s != nil && s.isSorted
}

func (s *TableSymbol) WithIsSorted(isSorted bool) *TableSymbol {
	if s == nil || s.isSorted == isSorted {
		return s
	}
	clone := s.clone()
	clone.isSorted = isSorted
	return clone
}

func (s *TableSymbol) IsExternal() bool {
	return s != nil && s.isExternal
}

func (s *TableSymbol) WithIsExternal(isExternal bool) *TableSymbol {
	if s == nil || s.isExternal == isExternal {
		return s
	}
	clone := s.clone()
	clone.isExternal = isExternal
	if isExternal {
		clone.isMaterializedView = false
		clone.isStoredQueryResult = false
	}
	return clone
}

func (s *TableSymbol) IsMaterializedView() bool {
	return s != nil && s.isMaterializedView
}

func (s *TableSymbol) WithIsMaterializedView(isMaterializedView bool) *TableSymbol {
	if s == nil || s.isMaterializedView == isMaterializedView {
		return s
	}
	clone := s.clone()
	clone.isMaterializedView = isMaterializedView
	if isMaterializedView {
		clone.isExternal = false
		clone.isStoredQueryResult = false
	}
	return clone
}

func (s *TableSymbol) IsStoredQueryResult() bool {
	return s != nil && s.isStoredQueryResult
}

func (s *TableSymbol) WithIsStoredQueryResult(isStoredQueryResult bool) *TableSymbol {
	if s == nil || s.isStoredQueryResult == isStoredQueryResult {
		return s
	}
	clone := s.clone()
	clone.isStoredQueryResult = isStoredQueryResult
	if isStoredQueryResult {
		clone.isExternal = false
		clone.isMaterializedView = false
	}
	return clone
}

func AreTablesEquivalent(x Symbol, y Symbol) bool {
	if x == y {
		return true
	}
	xTable, xOK := tableSymbolFromSymbol(x)
	yTable, yOK := tableSymbolFromSymbol(y)
	if !xOK || !yOK || xTable == nil || yTable == nil {
		return false
	}
	if tableEquivalenceKind(xTable) != tableEquivalenceKind(yTable) {
		return false
	}
	if xTable.SymbolName() != yTable.SymbolName() ||
		xTable.AlternateName() != yTable.AlternateName() ||
		xTable.Description() != yTable.Description() {
		return false
	}
	return AreTableResultsEquivalent(xTable, yTable)
}

func AreTableResultsEquivalent(x Symbol, y Symbol) bool {
	if x == y {
		return true
	}
	xTable, xOK := tableSymbolFromSymbol(x)
	yTable, yOK := tableSymbolFromSymbol(y)
	if !xOK || !yOK || xTable == nil || yTable == nil {
		return false
	}
	if xTable.isSerialized != yTable.isSerialized ||
		xTable.isSorted != yTable.isSorted ||
		xTable.isOpen != yTable.isOpen {
		return false
	}
	return AreTableColumnsEquivalent(xTable, yTable)
}

func AreTableColumnsEquivalent(x Symbol, y Symbol) bool {
	if x == y {
		return true
	}
	xTable, xOK := tableSymbolFromSymbol(x)
	yTable, yOK := tableSymbolFromSymbol(y)
	if !xOK || !yOK || xTable == nil || yTable == nil {
		return false
	}
	if len(xTable.columns) != len(yTable.columns) {
		return false
	}
	for i, xColumn := range xTable.columns {
		yColumn := yTable.columns[i]
		if xColumn == yColumn {
			continue
		}
		if xColumn == nil || yColumn == nil ||
			xColumn.SymbolName() != yColumn.SymbolName() ||
			xColumn.Type != yColumn.Type {
			return false
		}
	}
	return true
}

func tableEquivalenceKind(table *TableSymbol) string {
	switch {
	case table == nil:
		return string(SymbolKindNone)
	case table.IsStoredQueryResult():
		return string(SymbolKindStoredQueryResult)
	case table.IsMaterializedView():
		return string(SymbolKindMaterializedView)
	case table.IsExternal():
		return "ExternalTable"
	default:
		return string(SymbolKindTable)
	}
}

func (s *TableSymbol) Members() []Symbol {
	out := make([]Symbol, 0, len(s.columns))
	for _, col := range s.columns {
		out = append(out, col)
	}
	return out
}

func (s *TableSymbol) clone() *TableSymbol {
	if s == nil {
		return nil
	}
	clone := *s
	clone.columns = append([]*ColumnSymbol{}, s.columns...)
	clone.TypeSymbol.table = &clone
	return &clone
}

func TableSymbolForType(typ *TypeSymbol) (*TableSymbol, bool) {
	if typ == nil || typ.table == nil || &typ.table.TypeSymbol != typ {
		return nil, false
	}
	return typ.table, true
}

type MaterializedViewSymbol struct {
	TableSymbol
	materializedViewQuery string
	materializedViewKind  MaterializedViewKind
}

type MaterializedViewKind string

const (
	MaterializedViewKindUnknown      MaterializedViewKind = "Unknown"
	MaterializedViewKindDownsampling MaterializedViewKind = "Downsampling"
	MaterializedViewKindOther        MaterializedViewKind = "Other"
)

func NewMaterializedViewSymbol(name string, columns []*ColumnSymbol, query string, description ...string) *MaterializedViewSymbol {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}
	view := &MaterializedViewSymbol{
		TableSymbol: TableSymbol{
			TypeSymbol:         TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
			columns:            append([]*ColumnSymbol{}, columns...),
			description:        desc,
			isMaterializedView: true,
		},
		materializedViewQuery: query,
		materializedViewKind:  MaterializedViewKindUnknown,
	}
	view.TypeSymbol.table = &view.TableSymbol
	return view
}

func (s *MaterializedViewSymbol) SymbolKind() SymbolKind {
	return SymbolKindMaterializedView
}

func (s *MaterializedViewSymbol) MaterializedViewQuery() string {
	if s == nil {
		return ""
	}
	return s.materializedViewQuery
}

func (s *MaterializedViewSymbol) MaterializedViewKind() MaterializedViewKind {
	if s == nil || s.materializedViewKind == "" {
		return MaterializedViewKindUnknown
	}
	return s.materializedViewKind
}

func (s *MaterializedViewSymbol) WithMaterializedViewKind(kind MaterializedViewKind) *MaterializedViewSymbol {
	if kind == "" {
		kind = MaterializedViewKindUnknown
	}
	if s == nil || s.MaterializedViewKind() == kind {
		return s
	}
	clone := s.clone()
	clone.materializedViewKind = kind
	return clone
}

func (s *MaterializedViewSymbol) WithName(name string) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithName(name))
}

func (s *MaterializedViewSymbol) WithDescription(description string) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithDescription(description))
}

func (s *MaterializedViewSymbol) WithColumns(columns []*ColumnSymbol) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithColumns(columns))
}

func (s *MaterializedViewSymbol) AddColumns(columns ...*ColumnSymbol) *MaterializedViewSymbol {
	if s == nil || len(columns) == 0 {
		return s
	}
	return s.WithColumns(append(s.Columns(), columns...))
}

func (s *MaterializedViewSymbol) WithSource(source SyntaxElement) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithSource(source))
}

func (s *MaterializedViewSymbol) WithInheritableProperties(table *TableSymbol) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithInheritableProperties(table))
}

func (s *MaterializedViewSymbol) WithIsOpen(isOpen bool) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithIsOpen(isOpen))
}

func (s *MaterializedViewSymbol) WithIsSerialized(isSerialized bool) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithIsSerialized(isSerialized))
}

func (s *MaterializedViewSymbol) WithIsSorted(isSorted bool) *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithIsSorted(isSorted))
}

func (s *MaterializedViewSymbol) withTableSymbol(table *TableSymbol) *MaterializedViewSymbol {
	if s == nil || table == nil {
		return nil
	}
	if table == &s.TableSymbol {
		return s
	}
	clone := *s
	clone.TableSymbol = *table.clone()
	clone.TypeSymbol.table = &clone.TableSymbol
	return &clone
}

func (s *MaterializedViewSymbol) clone() *MaterializedViewSymbol {
	if s == nil {
		return nil
	}
	clone := *s
	clone.TableSymbol = *s.TableSymbol.clone()
	clone.TypeSymbol.table = &clone.TableSymbol
	return &clone
}

type ExternalTableSymbol struct {
	TableSymbol
}

func NewExternalTableSymbol(name string, columns ...*ColumnSymbol) *ExternalTableSymbol {
	return NewExternalTableSymbolWithDescription(name, columns, "")
}

func NewExternalTableSymbolWithDescription(name string, columns []*ColumnSymbol, description string) *ExternalTableSymbol {
	table := &ExternalTableSymbol{
		TableSymbol: TableSymbol{
			TypeSymbol:  TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
			columns:     append([]*ColumnSymbol{}, columns...),
			description: description,
			isExternal:  true,
		},
	}
	table.TypeSymbol.table = &table.TableSymbol
	return table
}

func (s *ExternalTableSymbol) WithName(name string) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithName(name))
}

func (s *ExternalTableSymbol) WithDescription(description string) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithDescription(description))
}

func (s *ExternalTableSymbol) WithColumns(columns []*ColumnSymbol) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithColumns(columns))
}

func (s *ExternalTableSymbol) AddColumns(columns ...*ColumnSymbol) *ExternalTableSymbol {
	if s == nil || len(columns) == 0 {
		return s
	}
	return s.WithColumns(append(s.Columns(), columns...))
}

func (s *ExternalTableSymbol) WithSource(source SyntaxElement) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithSource(source))
}

func (s *ExternalTableSymbol) WithInheritableProperties(table *TableSymbol) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithInheritableProperties(table))
}

func (s *ExternalTableSymbol) WithIsOpen(isOpen bool) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithIsOpen(isOpen))
}

func (s *ExternalTableSymbol) WithIsSerialized(isSerialized bool) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithIsSerialized(isSerialized))
}

func (s *ExternalTableSymbol) WithIsSorted(isSorted bool) *ExternalTableSymbol {
	if s == nil {
		return nil
	}
	return s.withTableSymbol(s.TableSymbol.WithIsSorted(isSorted))
}

func (s *ExternalTableSymbol) withTableSymbol(table *TableSymbol) *ExternalTableSymbol {
	if s == nil || table == nil {
		return nil
	}
	if table == &s.TableSymbol {
		return s
	}
	clone := *s
	clone.TableSymbol = *table.clone()
	clone.TypeSymbol.table = &clone.TableSymbol
	return &clone
}

type StoredQueryResultSymbol struct {
	TableSymbol
}

func NewStoredQueryResultSymbol(name string, columns ...*ColumnSymbol) *StoredQueryResultSymbol {
	result := &StoredQueryResultSymbol{
		TableSymbol: TableSymbol{
			TypeSymbol:          TypeSymbol{BaseSymbol: NewBaseSymbol(name)},
			columns:             append([]*ColumnSymbol{}, columns...),
			isStoredQueryResult: true,
		},
	}
	result.TypeSymbol.table = &result.TableSymbol
	return result
}

func NewStoredQueryResultSymbolFromColumns(name string, columns []*ColumnSymbol) *StoredQueryResultSymbol {
	return NewStoredQueryResultSymbol(name, columns...)
}

func (s *StoredQueryResultSymbol) SymbolKind() SymbolKind {
	return SymbolKindStoredQueryResult
}

func tableSymbolFromSymbol(symbol Symbol) (*TableSymbol, bool) {
	switch s := symbol.(type) {
	case *TypeSymbol:
		return TableSymbolForType(s)
	case *TableSymbol:
		return s, s != nil
	case *MaterializedViewSymbol:
		if s == nil {
			return nil, false
		}
		return &s.TableSymbol, true
	case *ExternalTableSymbol:
		if s == nil {
			return nil, false
		}
		return &s.TableSymbol, true
	case *StoredQueryResultSymbol:
		if s == nil {
			return nil, false
		}
		return &s.TableSymbol, true
	default:
		return nil, false
	}
}

type DatabaseSymbol struct {
	TypeSymbol
	alternateName string
	members       []Symbol
	isOpen        bool
	containsMu    sync.Mutex
	symbolSet     map[Symbol]struct{}
}

func NewDatabaseSymbol(name string, members ...Symbol) *DatabaseSymbol {
	return &DatabaseSymbol{TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)}, members: append([]Symbol{}, members...)}
}

func NewDatabaseSymbolWithAlternateName(name, alternateName string, members []Symbol, isOpen bool) *DatabaseSymbol {
	return &DatabaseSymbol{TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol(name)}, alternateName: alternateName, members: append([]Symbol{}, members...), isOpen: isOpen}
}

func (s *DatabaseSymbol) AlternateName() string {
	return s.alternateName
}

func (s *DatabaseSymbol) SymbolKind() SymbolKind {
	return SymbolKindDatabase
}

func (s *DatabaseSymbol) SymbolTabularity() Tabularity {
	return TabularityOther
}

func (s *DatabaseSymbol) Members() []Symbol {
	return append([]Symbol{}, s.members...)
}

func (s *DatabaseSymbol) IsOpen() bool {
	return s.isOpen
}

func (s *DatabaseSymbol) Tables() []*TableSymbol {
	var tables []*TableSymbol
	for _, member := range s.members {
		if table, ok := member.(*TableSymbol); ok && !table.IsExternal() && !table.IsMaterializedView() && !table.IsStoredQueryResult() {
			tables = append(tables, table)
		}
	}
	return tables
}

func (s *DatabaseSymbol) ExternalTables() []*ExternalTableSymbol {
	var externalTables []*ExternalTableSymbol
	for _, member := range s.members {
		if table, ok := member.(*ExternalTableSymbol); ok {
			externalTables = append(externalTables, table)
		}
	}
	return externalTables
}

func (s *DatabaseSymbol) MaterializedViews() []*MaterializedViewSymbol {
	var materializedViews []*MaterializedViewSymbol
	for _, member := range s.members {
		if table, ok := member.(*MaterializedViewSymbol); ok {
			materializedViews = append(materializedViews, table)
		}
	}
	return materializedViews
}

func (s *DatabaseSymbol) EntityGroups() []*EntityGroupSymbol {
	var entityGroups []*EntityGroupSymbol
	for _, member := range s.members {
		if entityGroup, ok := member.(*EntityGroupSymbol); ok {
			entityGroups = append(entityGroups, entityGroup)
		}
	}
	return entityGroups
}

func (s *DatabaseSymbol) StoredQueryResults() []*StoredQueryResultSymbol {
	var storedQueryResults []*StoredQueryResultSymbol
	for _, member := range s.members {
		if stored, ok := member.(*StoredQueryResultSymbol); ok {
			storedQueryResults = append(storedQueryResults, stored)
		}
	}
	return storedQueryResults
}

func (s *DatabaseSymbol) GraphModels() []*GraphModelSymbol {
	var graphModels []*GraphModelSymbol
	for _, member := range s.members {
		if model, ok := member.(*GraphModelSymbol); ok {
			graphModels = append(graphModels, model)
		}
	}
	return graphModels
}

func (s *DatabaseSymbol) Functions() []*FunctionSymbol {
	var functions []*FunctionSymbol
	for _, member := range s.members {
		if fn, ok := member.(*FunctionSymbol); ok {
			functions = append(functions, fn)
		}
	}
	return functions
}

func (s *DatabaseSymbol) GetMember(name string) Symbol {
	if s == nil {
		return nil
	}
	for _, member := range s.members {
		if member != nil && member.SymbolName() == name {
			return member
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetTable(name string) *TableSymbol {
	if s == nil {
		return nil
	}
	for _, member := range s.members {
		if table, ok := member.(*TableSymbol); ok && !table.IsExternal() && !table.IsMaterializedView() && !table.IsStoredQueryResult() && table.SymbolName() == name {
			return table
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetAnyTable(name string) *TableSymbol {
	if symbol := s.GetAnyTableSymbol(name); symbol != nil {
		table, _ := tableSymbolFromSymbol(symbol)
		return table
	}
	return nil
}

// GetAnyTableSymbol preserves the concrete identity of table-derived symbols.
// GetAnyTable remains available for callers that only need the shared table API.
func (s *DatabaseSymbol) GetAnyTableSymbol(name string) Symbol {
	if table := s.GetTable(name); table != nil {
		return table
	}
	if table := s.GetExternalTable(name); table != nil {
		return table
	}
	if table := s.GetMaterializedView(name); table != nil {
		return table
	}
	return nil
}

func (s *DatabaseSymbol) GetExternalTable(name string) *ExternalTableSymbol {
	for _, table := range s.ExternalTables() {
		if table.SymbolName() == name {
			return table
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetMaterializedView(name string) *MaterializedViewSymbol {
	for _, table := range s.MaterializedViews() {
		if table.SymbolName() == name {
			return table
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetEntityGroup(name string) *EntityGroupSymbol {
	for _, entityGroup := range s.EntityGroups() {
		if entityGroup.SymbolName() == name {
			return entityGroup
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetStoredQueryResult(name string) *StoredQueryResultSymbol {
	for _, stored := range s.StoredQueryResults() {
		if stored.SymbolName() == name {
			return stored
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetGraphModel(name string) *GraphModelSymbol {
	for _, model := range s.GraphModels() {
		if model.SymbolName() == name {
			return model
		}
	}
	return nil
}

func (s *DatabaseSymbol) GetFunction(name string) *FunctionSymbol {
	for _, fn := range s.Functions() {
		if fn.SymbolName() == name {
			return fn
		}
	}
	return nil
}

func (s *DatabaseSymbol) WithAlternateName(alternateName string) *DatabaseSymbol {
	if s == nil {
		return nil
	}
	if s.alternateName == alternateName {
		return s
	}
	return NewDatabaseSymbolWithAlternateName(s.SymbolName(), alternateName, s.members, s.isOpen)
}

func (s *DatabaseSymbol) WithMembers(members []Symbol) *DatabaseSymbol {
	if s == nil {
		return nil
	}
	return NewDatabaseSymbolWithAlternateName(s.SymbolName(), s.alternateName, members, s.isOpen)
}

func (s *DatabaseSymbol) AddMembers(members ...Symbol) *DatabaseSymbol {
	if s == nil {
		return nil
	}
	newMembers := append([]Symbol{}, s.members...)
	newMembers = append(newMembers, members...)
	return s.WithMembers(newMembers)
}

// Contains preserves the observable behavior of DatabaseSymbol.Contains in
// the pinned c94 source. That implementation populates its lazy set with the
// first requested symbol once per member, rather than with each member.
func (s *DatabaseSymbol) Contains(symbol Symbol) bool {
	s.containsMu.Lock()
	defer s.containsMu.Unlock()

	if s.symbolSet == nil {
		s.symbolSet = make(map[Symbol]struct{})
		for range s.members {
			s.symbolSet[symbol] = struct{}{}
		}
	}
	_, ok := s.symbolSet[symbol]
	return ok
}

var UnknownDatabaseSymbol = NewDatabaseSymbolWithAlternateName("", "", nil, true)

func FindFirstMember(symbol Symbol, name string, ignoreCase bool) Symbol {
	return FindFirstMemberMatch(symbol, name, SymbolMatchAny, ignoreCase)
}

func matchesSymbolName(symbol Symbol, name string, ignoreCase bool) bool {
	if symbol == nil {
		return false
	}
	if ignoreCase {
		return ordinalEqualFold(symbol.SymbolName(), name) || ordinalEqualFold(symbol.AlternateName(), name)
	}
	return symbol.SymbolName() == name || symbol.AlternateName() == name
}
