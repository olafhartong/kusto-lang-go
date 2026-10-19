package kustolang

// Ported from src/Kusto.Language/Symbols/DynamicSymbol.cs,
// TupleSymbol.cs, ErrorSymbol.cs and VoidSymbol.cs.

type DynamicSymbol struct {
	ScalarSymbol
}

func newDynamicSymbol() DynamicSymbol {
	return DynamicSymbol{ScalarSymbol: *NewScalarSymbol("dynamic", nil, ScalarFlagAll)}
}

type DynamicAnySymbol struct {
	DynamicSymbol
}

func NewDynamicAnySymbol() *DynamicAnySymbol {
	return &DynamicAnySymbol{DynamicSymbol: newDynamicSymbol()}
}

func (s *DynamicAnySymbol) SymbolKind() SymbolKind {
	return SymbolKindPrimitive
}

type DynamicPrimitiveSymbol struct {
	DynamicSymbol
	UnderlyingType *ScalarSymbol
}

func NewDynamicPrimitiveSymbol(underlyingType *ScalarSymbol) *DynamicPrimitiveSymbol {
	primitive := &DynamicPrimitiveSymbol{
		DynamicSymbol:  newDynamicSymbol(),
		UnderlyingType: underlyingType,
	}
	primitive.ScalarSymbol.TypeSymbol.dynamicPrimitive = primitive
	return primitive
}

func (s *DynamicPrimitiveSymbol) SymbolKind() SymbolKind {
	return SymbolKindPrimitive
}

type DynamicArraySymbol struct {
	DynamicSymbol
	ElementType *TypeSymbol
}

func NewDynamicArraySymbol(elementType *TypeSymbol) *DynamicArraySymbol {
	array := &DynamicArraySymbol{
		DynamicSymbol: newDynamicSymbol(),
		ElementType:   elementType,
	}
	array.ScalarSymbol.TypeSymbol.dynamicArray = array
	return array
}

func (s *DynamicArraySymbol) SymbolKind() SymbolKind {
	return SymbolKindArray
}

type DynamicBagSymbol struct {
	DynamicSymbol
	properties []*ColumnSymbol
}

func NewDynamicBagSymbol(properties ...*ColumnSymbol) *DynamicBagSymbol {
	bag := &DynamicBagSymbol{
		DynamicSymbol: newDynamicSymbol(),
		properties:    append([]*ColumnSymbol{}, properties...),
	}
	bag.ScalarSymbol.TypeSymbol.dynamicBag = bag
	return bag
}

func (s *DynamicBagSymbol) SymbolKind() SymbolKind {
	return SymbolKindBag
}

func (s *DynamicBagSymbol) SymbolTabularity() Tabularity {
	return TabularityScalar
}

func (s *DynamicBagSymbol) Members() []Symbol {
	if s == nil {
		return nil
	}
	out := make([]Symbol, 0, len(s.properties))
	for _, property := range s.properties {
		out = append(out, property)
	}
	return out
}

func (s *DynamicBagSymbol) Properties() []*ColumnSymbol {
	if s == nil {
		return nil
	}
	return append([]*ColumnSymbol{}, s.properties...)
}

func (s *DynamicBagSymbol) TryGetProperty(name string) (*ColumnSymbol, bool) {
	if s == nil {
		return nil, false
	}
	var found *ColumnSymbol
	for _, property := range s.properties {
		if property != nil && property.SymbolName() == name {
			found = property
		}
	}
	return found, found != nil
}

func (s *DynamicBagSymbol) WithProperties(properties []*ColumnSymbol) *DynamicBagSymbol {
	if s == nil {
		return NewDynamicBagSymbol(properties...)
	}
	if sameColumnSlice(s.properties, properties) {
		return s
	}
	return NewDynamicBagSymbol(properties...)
}

func (s *DynamicBagSymbol) AddOrUpdateProperty(property *ColumnSymbol) *DynamicBagSymbol {
	if s == nil {
		return NewDynamicBagSymbol(property)
	}
	if property == nil {
		return s
	}
	properties := s.Properties()
	if existingProperty, ok := s.TryGetProperty(property.SymbolName()); ok {
		for i, existing := range properties {
			if existing == existingProperty {
				properties[i] = property
			}
		}
		return NewDynamicBagSymbol(properties...)
	}
	properties = append(properties, property)
	return NewDynamicBagSymbol(properties...)
}

func (s *DynamicBagSymbol) WithSource(source SyntaxElement) *DynamicBagSymbol {
	if s == nil {
		return nil
	}
	properties := make([]*ColumnSymbol, 0, len(s.properties))
	for _, property := range s.properties {
		if property == nil {
			properties = append(properties, nil)
			continue
		}
		properties = append(properties, property.WithSource(source))
	}
	return NewDynamicBagSymbol(properties...)
}

func DynamicBagForType(typ *TypeSymbol) (*DynamicBagSymbol, bool) {
	if typ == nil || typ.dynamicBag == nil || &typ.dynamicBag.ScalarSymbol.TypeSymbol != typ {
		return nil, false
	}
	return typ.dynamicBag, true
}

func DynamicPrimitiveForType(typ *TypeSymbol) (*DynamicPrimitiveSymbol, bool) {
	if typ == nil || typ.dynamicPrimitive == nil || &typ.dynamicPrimitive.ScalarSymbol.TypeSymbol != typ {
		return nil, false
	}
	return typ.dynamicPrimitive, true
}

func DynamicArrayForType(typ *TypeSymbol) (*DynamicArraySymbol, bool) {
	if typ == nil || typ.dynamicArray == nil || &typ.dynamicArray.ScalarSymbol.TypeSymbol != typ {
		return nil, false
	}
	return typ.dynamicArray, true
}

type TupleSymbol struct {
	ScalarSymbol
	columns      []*ColumnSymbol
	RelatedTable *TableSymbol
}

func NewTupleSymbol(columns []*ColumnSymbol, relatedTable ...*TableSymbol) *TupleSymbol {
	var table *TableSymbol
	if len(relatedTable) > 0 {
		table = relatedTable[0]
	}
	tuple := &TupleSymbol{
		ScalarSymbol: *NewScalarSymbol("tuple", nil, ScalarFlagNone),
		columns:      append([]*ColumnSymbol{}, columns...),
		RelatedTable: table,
	}
	tuple.ScalarSymbol.TypeSymbol.tuple = tuple
	return tuple
}

func NewTupleSymbolOf(columns ...*ColumnSymbol) *TupleSymbol {
	return NewTupleSymbol(columns)
}

func TupleSymbolFromSchema(schema string) (*TupleSymbol, error) {
	table, err := TableSymbolFromSchema(schema)
	if err != nil {
		return nil, err
	}
	return NewTupleSymbol(table.Columns()), nil
}

var EmptyTupleSymbol = NewTupleSymbol(nil)

func (s *TupleSymbol) SymbolKind() SymbolKind {
	return SymbolKindTuple
}

func (s *TupleSymbol) Members() []Symbol {
	if s == nil {
		return nil
	}
	out := make([]Symbol, 0, len(s.columns))
	for _, column := range s.columns {
		out = append(out, column)
	}
	return out
}

func (s *TupleSymbol) Columns() []*ColumnSymbol {
	if s == nil {
		return nil
	}
	return append([]*ColumnSymbol{}, s.columns...)
}

func (s *TupleSymbol) IsReducibleToScalar() bool {
	return s != nil && len(s.columns) == 1 && s.RelatedTable == nil
}

func (s *TupleSymbol) WithColumns(columns []*ColumnSymbol) *TupleSymbol {
	if s == nil {
		return NewTupleSymbol(columns)
	}
	return NewTupleSymbol(columns, s.RelatedTable)
}

func (s *TupleSymbol) WithSource(source SyntaxElement) *TupleSymbol {
	if s == nil {
		return nil
	}
	columns := make([]*ColumnSymbol, 0, len(s.columns))
	for _, column := range s.columns {
		if column == nil {
			columns = append(columns, nil)
			continue
		}
		columns = append(columns, column.WithSource(source))
	}
	return s.WithColumns(columns)
}

func TupleForType(typ *TypeSymbol) (*TupleSymbol, bool) {
	if typ == nil || typ.tuple == nil || &typ.tuple.ScalarSymbol.TypeSymbol != typ {
		return nil, false
	}
	return typ.tuple, true
}

type ErrorSymbol struct {
	TypeSymbol
}

func NewErrorSymbol() *ErrorSymbol {
	return &ErrorSymbol{TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol("error")}}
}

func (s *ErrorSymbol) SymbolKind() SymbolKind {
	return SymbolKindError
}

func (s *ErrorSymbol) SymbolTabularity() Tabularity {
	return TabularityNone
}

func (s *ErrorSymbol) IsError() bool {
	return true
}

type VoidSymbol struct {
	TypeSymbol
}

func NewVoidSymbol() *VoidSymbol {
	return &VoidSymbol{TypeSymbol: TypeSymbol{BaseSymbol: NewBaseSymbol("void")}}
}

func (s *VoidSymbol) SymbolKind() SymbolKind {
	return SymbolKindVoid
}

func (s *VoidSymbol) SymbolTabularity() Tabularity {
	return TabularityNone
}

var (
	Error = NewErrorSymbol()
	Void  = NewVoidSymbol()
)

func sameColumnSlice(left, right []*ColumnSymbol) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
