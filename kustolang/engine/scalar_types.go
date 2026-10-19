package kustolang

import "strings"

// Ported from src/Kusto.Language/Symbols/ScalarTypes.cs. The current Go port
// includes primitive scalars plus the dynamic primitive/array/bag catalog used
// by function, operator and plugin signatures.

type scalarTypeSet struct {
	DynamicAny *DynamicAnySymbol
	Dynamic    *ScalarSymbol
	Bool       *ScalarSymbol
	Int        *ScalarSymbol
	Long       *ScalarSymbol
	Real       *ScalarSymbol
	Decimal    *ScalarSymbol
	DateTime   *ScalarSymbol
	TimeSpan   *ScalarSymbol
	Guid       *ScalarSymbol
	Type       *ScalarSymbol
	TypeLit    *ScalarSymbol
	String     *ScalarSymbol
	Null       *ScalarSymbol
	Unknown    *ScalarSymbol

	DynamicBool                 *DynamicPrimitiveSymbol
	DynamicLong                 *DynamicPrimitiveSymbol
	DynamicReal                 *DynamicPrimitiveSymbol
	DynamicDateTime             *DynamicPrimitiveSymbol
	DynamicTimeSpan             *DynamicPrimitiveSymbol
	DynamicGuid                 *DynamicPrimitiveSymbol
	DynamicString               *DynamicPrimitiveSymbol
	DynamicBag                  *DynamicBagSymbol
	DynamicArray                *DynamicArraySymbol
	DynamicArrayOfBool          *DynamicArraySymbol
	DynamicArrayOfLong          *DynamicArraySymbol
	DynamicArrayOfReal          *DynamicArraySymbol
	DynamicArrayOfDateTime      *DynamicArraySymbol
	DynamicArrayOfTimeSpan      *DynamicArraySymbol
	DynamicArrayOfGuid          *DynamicArraySymbol
	DynamicArrayOfString        *DynamicArraySymbol
	DynamicArrayOfArray         *DynamicArraySymbol
	DynamicArrayOfBag           *DynamicArraySymbol
	DynamicArrayOfArrayOfReal   *DynamicArraySymbol
	DynamicArrayOfArrayOfString *DynamicArraySymbol
	GeoShape                    *DynamicBagSymbol

	all          []*ScalarSymbol
	byName       map[string]*ScalarSymbol
	symbolByName map[string]*ScalarSymbol
}

var ScalarTypes = newScalarTypes()

func newScalarTypes() *scalarTypeSet {
	st := &scalarTypeSet{}
	dynamicAny := NewDynamicAnySymbol()
	st.DynamicAny = dynamicAny
	st.Dynamic = &dynamicAny.ScalarSymbol
	st.Bool = NewScalarSymbol("bool", []string{"boolean"}, ScalarFlagOrderable)
	st.Int = NewScalarSymbol("int", []string{"int32", "uint", "uint32", "int8", "uint8", "int16", "uint16"}, ScalarFlagInteger|ScalarFlagNumeric|ScalarFlagInterval|ScalarFlagSummable|ScalarFlagOrderable)
	st.Long = NewScalarSymbol("long", []string{"int64", "ulong", "uint64"}, ScalarFlagInteger|ScalarFlagNumeric|ScalarFlagInterval|ScalarFlagSummable|ScalarFlagOrderable, st.Int)
	st.Real = NewScalarSymbol("real", []string{"double", "float"}, ScalarFlagNumeric|ScalarFlagInterval|ScalarFlagSummable|ScalarFlagOrderable, st.Int, st.Long)
	st.Decimal = NewScalarSymbol("decimal", nil, ScalarFlagNumeric|ScalarFlagInterval|ScalarFlagSummable|ScalarFlagOrderable, st.Int, st.Long, st.Real)
	st.DateTime = NewScalarSymbol("datetime", []string{"date"}, ScalarFlagInterval|ScalarFlagSummable|ScalarFlagOrderable)
	st.TimeSpan = NewScalarSymbol("timespan", []string{"time"}, ScalarFlagInterval|ScalarFlagSummable|ScalarFlagOrderable)
	st.Guid = NewScalarSymbol("guid", []string{"uuid", "uniqueid"}, ScalarFlagNone)
	st.Type = NewScalarSymbol("type", nil, ScalarFlagNone)
	st.TypeLit = st.Type
	st.String = NewScalarSymbol("string", nil, ScalarFlagOrderable, st.Dynamic)
	st.Null = NewScalarSymbol("null", nil, ScalarFlagNone)
	st.Unknown = NewScalarSymbol("unknown", nil, ScalarFlagAll)

	st.DynamicBool = NewDynamicPrimitiveSymbol(st.Bool)
	st.DynamicLong = NewDynamicPrimitiveSymbol(st.Long)
	st.DynamicReal = NewDynamicPrimitiveSymbol(st.Real)
	st.DynamicDateTime = NewDynamicPrimitiveSymbol(st.DateTime)
	st.DynamicTimeSpan = NewDynamicPrimitiveSymbol(st.TimeSpan)
	st.DynamicGuid = NewDynamicPrimitiveSymbol(st.Guid)
	st.DynamicString = NewDynamicPrimitiveSymbol(st.String)
	st.DynamicBag = NewDynamicBagSymbol()
	st.DynamicArray = NewDynamicArraySymbol(scalarTypeRef(st.Dynamic))
	st.DynamicArrayOfBool = NewDynamicArraySymbol(scalarTypeRef(st.Bool))
	st.DynamicArrayOfLong = NewDynamicArraySymbol(scalarTypeRef(st.Long))
	st.DynamicArrayOfReal = NewDynamicArraySymbol(scalarTypeRef(st.Real))
	st.DynamicArrayOfDateTime = NewDynamicArraySymbol(scalarTypeRef(st.DateTime))
	st.DynamicArrayOfTimeSpan = NewDynamicArraySymbol(scalarTypeRef(st.TimeSpan))
	st.DynamicArrayOfGuid = NewDynamicArraySymbol(scalarTypeRef(st.Guid))
	st.DynamicArrayOfString = NewDynamicArraySymbol(scalarTypeRef(st.String))
	st.DynamicArrayOfArray = NewDynamicArraySymbol(dynamicArrayTypeRef(st.DynamicArray))
	st.DynamicArrayOfBag = NewDynamicArraySymbol(dynamicBagTypeRef(st.DynamicBag))
	st.DynamicArrayOfArrayOfReal = NewDynamicArraySymbol(dynamicArrayTypeRef(st.DynamicArrayOfReal))
	st.DynamicArrayOfArrayOfString = NewDynamicArraySymbol(dynamicArrayTypeRef(st.DynamicArrayOfString))
	st.GeoShape = NewDynamicBagSymbol(
		newColumnSymbolWithoutScalarFallback("type", scalarTypeRef(st.String)),
		newColumnSymbolWithoutScalarFallback("coordinates", dynamicArrayTypeRef(st.DynamicArray)),
	)

	st.all = []*ScalarSymbol{
		st.Bool, st.Int, st.Long, st.Real, st.Decimal, st.String, st.DateTime,
		st.TimeSpan, st.Guid, st.Type, st.Dynamic, st.Null,
	}
	st.byName = map[string]*ScalarSymbol{}
	st.symbolByName = map[string]*ScalarSymbol{}
	for _, typ := range []*ScalarSymbol{
		st.Dynamic, st.Bool, st.Int, st.Long, st.Real, st.Decimal, st.DateTime,
		st.TimeSpan, st.Guid, st.Type, st.String, st.Null, st.Unknown,
	} {
		st.register(typ)
	}
	for _, typ := range st.all {
		st.registerSymbolName(typ)
	}
	return st
}

func (s *scalarTypeSet) register(symbol *ScalarSymbol) {
	s.byName[foldTypeName(symbol.SymbolName())] = symbol
	for _, alias := range symbol.Aliases() {
		s.byName[foldTypeName(alias)] = symbol
	}
}

func (s *scalarTypeSet) registerSymbolName(symbol *ScalarSymbol) {
	s.symbolByName[symbol.SymbolName()] = symbol
	for _, alias := range symbol.Aliases() {
		s.symbolByName[alias] = symbol
	}
}

func (s *scalarTypeSet) Get(name string) *ScalarSymbol {
	if s == nil {
		return nil
	}
	return s.byName[foldTypeName(name)]
}

// GetSymbol mirrors the pinned source's ordinal type-name lookup.
func (s *scalarTypeSet) GetSymbol(name string) *ScalarSymbol {
	if s == nil {
		return nil
	}
	return s.symbolByName[name]
}

// All returns the pinned public scalar catalog in source order.
func (s *scalarTypeSet) All() []*ScalarSymbol {
	if s == nil {
		return nil
	}
	return append([]*ScalarSymbol{}, s.all...)
}

func (s *scalarTypeSet) GetDynamic(underlyingType *TypeSymbol) *TypeSymbol {
	if s == nil {
		return nil
	}
	switch {
	case underlyingType == nil,
		sameType(underlyingType, s.Dynamic.Type()),
		sameType(underlyingType, s.Null.Type()),
		sameType(underlyingType, s.Unknown.Type()):
		return s.Dynamic.Type()
	case isDynamicTypeSymbol(underlyingType):
		return underlyingType
	case sameType(underlyingType, s.Bool.Type()):
		return s.DynamicBool.Type()
	case sameType(underlyingType, s.Int.Type()), sameType(underlyingType, s.Long.Type()):
		return s.DynamicLong.Type()
	case sameType(underlyingType, s.Real.Type()):
		return s.DynamicReal.Type()
	case sameType(underlyingType, s.DateTime.Type()):
		return s.DynamicDateTime.Type()
	case sameType(underlyingType, s.TimeSpan.Type()):
		return s.DynamicTimeSpan.Type()
	case sameType(underlyingType, s.Guid.Type()):
		return s.DynamicGuid.Type()
	case sameType(underlyingType, s.String.Type()):
		return s.DynamicString.Type()
	default:
		return s.Dynamic.Type()
	}
}

func (s *scalarTypeSet) GetDynamicArray(elementType *TypeSymbol) *DynamicArraySymbol {
	if s == nil {
		return nil
	}
	elementType = s.unwrapDynamicPrimitive(elementType)
	switch {
	case elementType == nil,
		sameType(elementType, s.Dynamic.Type()),
		sameType(elementType, s.Null.Type()),
		sameType(elementType, s.Unknown.Type()):
		return s.DynamicArray
	case sameType(elementType, s.Bool.Type()):
		return s.DynamicArrayOfBool
	case sameType(elementType, s.Int.Type()), sameType(elementType, s.Long.Type()):
		return s.DynamicArrayOfLong
	case sameType(elementType, s.Real.Type()):
		return s.DynamicArrayOfReal
	case sameType(elementType, s.DateTime.Type()):
		return s.DynamicArrayOfDateTime
	case sameType(elementType, s.TimeSpan.Type()):
		return s.DynamicArrayOfTimeSpan
	case sameType(elementType, s.Guid.Type()):
		return s.DynamicArrayOfGuid
	case sameType(elementType, s.String.Type()):
		return s.DynamicArrayOfString
	case sameType(elementType, s.DynamicArray.Type()):
		return s.DynamicArrayOfArray
	case sameType(elementType, s.DynamicBag.Type()):
		return s.DynamicArrayOfBag
	case sameType(elementType, s.DynamicArrayOfReal.Type()):
		return s.DynamicArrayOfArrayOfReal
	case sameType(elementType, s.DynamicArrayOfString.Type()):
		return s.DynamicArrayOfArrayOfString
	default:
		if _, ok := DynamicArrayForType(elementType); ok {
			return NewDynamicArraySymbol(elementType)
		}
		if _, ok := DynamicBagForType(elementType); ok {
			return NewDynamicArraySymbol(elementType)
		}
		return s.DynamicArray
	}
}

func (s *scalarTypeSet) GetDynamicBag(properties ...*ColumnSymbol) *DynamicBagSymbol {
	if s == nil {
		return nil
	}
	if len(properties) == 0 {
		return s.DynamicBag
	}
	return NewDynamicBagSymbol(properties...)
}

func (s *scalarTypeSet) GetDynamicBagFromSchema(schema string) (*DynamicBagSymbol, error) {
	if s == nil {
		return nil, nil
	}
	table, err := TableSymbolFromSchema(schema)
	if err != nil {
		return nil, err
	}
	return s.GetDynamicBag(table.Columns()...), nil
}

func (s *scalarTypeSet) GetTuple(columns ...*ColumnSymbol) *TupleSymbol {
	return NewTupleSymbol(columns)
}

func (s *scalarTypeSet) GetTupleFromSchema(schema string) (*TupleSymbol, error) {
	if s == nil {
		return nil, nil
	}
	table, err := TableSymbolFromSchema(schema)
	if err != nil {
		return nil, err
	}
	return s.GetTuple(table.Columns()...), nil
}

func (s *scalarTypeSet) CanonicalName(name string) string {
	if symbol := s.Get(name); symbol != nil {
		return symbol.SymbolName()
	}
	if strings.TrimSpace(name) == "" {
		return s.Dynamic.SymbolName()
	}
	return strings.ToLower(strings.TrimSpace(name))
}

func (s *ScalarSymbol) Type() *TypeSymbol {
	if s == nil {
		return ScalarTypes.Dynamic.Type()
	}
	return &s.TypeSymbol
}

func scalarTypeRef(symbol *ScalarSymbol) *TypeSymbol {
	if symbol == nil {
		return nil
	}
	return &symbol.TypeSymbol
}

func dynamicPrimitiveTypeRef(symbol *DynamicPrimitiveSymbol) *TypeSymbol {
	if symbol == nil {
		return nil
	}
	return &symbol.ScalarSymbol.TypeSymbol
}

func dynamicArrayTypeRef(symbol *DynamicArraySymbol) *TypeSymbol {
	if symbol == nil {
		return nil
	}
	return &symbol.ScalarSymbol.TypeSymbol
}

func dynamicBagTypeRef(symbol *DynamicBagSymbol) *TypeSymbol {
	if symbol == nil {
		return nil
	}
	return &symbol.ScalarSymbol.TypeSymbol
}

func newColumnSymbolWithoutScalarFallback(name string, typ *TypeSymbol) *ColumnSymbol {
	return &ColumnSymbol{BaseSymbol: NewBaseSymbol(name), Type: typ}
}

func (s *scalarTypeSet) unwrapDynamicPrimitive(typ *TypeSymbol) *TypeSymbol {
	if s == nil {
		return typ
	}
	if primitive, ok := DynamicPrimitiveForType(typ); ok && primitive != nil && primitive.UnderlyingType != nil {
		return primitive.UnderlyingType.Type()
	}
	return typ
}

func isDynamicTypeSymbol(typ *TypeSymbol) bool {
	if typ == nil {
		return false
	}
	if _, ok := DynamicPrimitiveForType(typ); ok {
		return true
	}
	if _, ok := DynamicArrayForType(typ); ok {
		return true
	}
	if _, ok := DynamicBagForType(typ); ok {
		return true
	}
	return false
}

func sameType(left, right *TypeSymbol) bool {
	return left != nil && right != nil && left == right
}

func foldTypeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
