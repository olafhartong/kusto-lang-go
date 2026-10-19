package kustolang

// Ported from src/Kusto.Language/Symbols/ScalarTypes.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// ScalarTypesAPI is the named public facade for Kusto.Language's static
// ScalarTypes catalog. The symbols retain the canonical identities used by the
// parser and binder.
type ScalarTypesAPI struct {
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
}

func newScalarTypesAPI() *ScalarTypesAPI {
	source := internal.ScalarTypes
	return &ScalarTypesAPI{
		DynamicAny:                  source.DynamicAny,
		Dynamic:                     source.Dynamic,
		Bool:                        source.Bool,
		Int:                         source.Int,
		Long:                        source.Long,
		Real:                        source.Real,
		Decimal:                     source.Decimal,
		DateTime:                    source.DateTime,
		TimeSpan:                    source.TimeSpan,
		Guid:                        source.Guid,
		Type:                        source.Type,
		TypeLit:                     source.TypeLit,
		String:                      source.String,
		Null:                        source.Null,
		Unknown:                     source.Unknown,
		DynamicBool:                 source.DynamicBool,
		DynamicLong:                 source.DynamicLong,
		DynamicReal:                 source.DynamicReal,
		DynamicDateTime:             source.DynamicDateTime,
		DynamicTimeSpan:             source.DynamicTimeSpan,
		DynamicGuid:                 source.DynamicGuid,
		DynamicString:               source.DynamicString,
		DynamicBag:                  source.DynamicBag,
		DynamicArray:                source.DynamicArray,
		DynamicArrayOfBool:          source.DynamicArrayOfBool,
		DynamicArrayOfLong:          source.DynamicArrayOfLong,
		DynamicArrayOfReal:          source.DynamicArrayOfReal,
		DynamicArrayOfDateTime:      source.DynamicArrayOfDateTime,
		DynamicArrayOfTimeSpan:      source.DynamicArrayOfTimeSpan,
		DynamicArrayOfGuid:          source.DynamicArrayOfGuid,
		DynamicArrayOfString:        source.DynamicArrayOfString,
		DynamicArrayOfArray:         source.DynamicArrayOfArray,
		DynamicArrayOfBag:           source.DynamicArrayOfBag,
		DynamicArrayOfArrayOfReal:   source.DynamicArrayOfArrayOfReal,
		DynamicArrayOfArrayOfString: source.DynamicArrayOfArrayOfString,
		GeoShape:                    source.GeoShape,
	}
}

// Get performs the port's compatibility lookup, including aliases, trimming,
// and case folding.
func (s *ScalarTypesAPI) Get(name string) *ScalarSymbol {
	if s != nil {
		return internal.ScalarTypes.Get(name)
	}
	return nil
}

// GetSymbol performs the pinned source's ordinal name and alias lookup.
func (s *ScalarTypesAPI) GetSymbol(name string) *ScalarSymbol {
	if s != nil {
		return internal.ScalarTypes.GetSymbol(name)
	}
	return nil
}

// All returns the public scalar catalog in pinned source order.
func (s *ScalarTypesAPI) All() []*ScalarSymbol {
	if s != nil {
		return internal.ScalarTypes.All()
	}
	return []*ScalarSymbol{}
}

func (s *ScalarTypesAPI) GetDynamic(underlyingType *TypeSymbol) *TypeSymbol {
	if s != nil {
		return internal.ScalarTypes.GetDynamic(underlyingType)
	}
	return nil
}

func (s *ScalarTypesAPI) GetDynamicArray(elementType *TypeSymbol) *DynamicArraySymbol {
	if s != nil {
		return internal.ScalarTypes.GetDynamicArray(elementType)
	}
	return nil
}

func (s *ScalarTypesAPI) GetDynamicBag(properties ...*ColumnSymbol) *DynamicBagSymbol {
	if s != nil {
		return internal.ScalarTypes.GetDynamicBag(properties...)
	}
	return nil
}

func (s *ScalarTypesAPI) GetDynamicBagFromSchema(schema string) (*DynamicBagSymbol, error) {
	if s != nil {
		return internal.ScalarTypes.GetDynamicBagFromSchema(schema)
	}
	return nil, nil
}

func (s *ScalarTypesAPI) GetTuple(columns ...*ColumnSymbol) *TupleSymbol {
	if s != nil {
		return internal.ScalarTypes.GetTuple(columns...)
	}
	return nil
}

func (s *ScalarTypesAPI) GetTupleFromSchema(schema string) (*TupleSymbol, error) {
	if s != nil {
		return internal.ScalarTypes.GetTupleFromSchema(schema)
	}
	return nil, nil
}

func (s *ScalarTypesAPI) CanonicalName(name string) string {
	if s != nil {
		return internal.ScalarTypes.CanonicalName(name)
	}
	return ""
}
