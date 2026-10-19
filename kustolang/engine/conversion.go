package kustolang

// Ported from src/Kusto.Language/Symbols/Conversion.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type Conversion int

const (
	ConversionNone Conversion = iota
	ConversionPromotable
	ConversionDynamic
	ConversionCompatible
	ConversionAny
)

func IsConversionAllowed(conversion Conversion, allowedConversions Conversion) bool {
	return conversion <= allowedConversions
}
