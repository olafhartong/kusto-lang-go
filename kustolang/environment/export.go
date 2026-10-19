// Package-level exported bridges for the internal/analyzer compatibility shim.
package environment

// Exported bridges to unexported helpers used by internal/analyzer.
var (
	AllowDefenderTimeAliases               = allowDefenderTimeAliases
	ApplyColumnTypeDetailsToSchema         = applyColumnTypeDetailsToSchema
	CloneColumnProperties                  = cloneColumnProperties
	CloneEnvironmentDefinition             = cloneEnvironmentDefinition
	CloneLocalData                         = cloneLocalData
	CloneColumnTypeDetails                 = copyColumnTypeDetails
	CloneColumnTypeDetailsMap              = copyColumnTypeDetailsMap
	CloneResultTables                      = cloneResultTables
	CloneStringMap                         = cloneStringMap
	ColumnFromTypeSymbol                   = columnFromTypeSymbol
	ColumnMetadataConsensus                = columnMetadataConsensus
	CopyTableOptionsMap                    = copyTableOptionsMap
	CurrentEnvironmentDefinition           = currentEnvironmentDefinition
	DecodeOrderedJSONObject                = decodeOrderedJSONObject
	FoldName                               = foldName
	IntSuffix                              = intSuffix
	MatchingDatabaseSymbol                 = matchingDatabaseSymbol
	NewOwnedCatalog                        = newOwnedCatalog
	NormalizeColumnTypeDetails             = normalizeColumnTypeDetails
	NormalizeColumnTypeDetailsMap          = normalizeColumnTypeDetailsMap
	NormalizeType                          = normalizeType
	ParametersFromArguments                = parametersFromArguments
	PublicTypeName                         = publicTypeName
	ResultTableFromSchema                  = resultTableFromSchema
	ResultTableFromTableSymbol             = resultTableFromTableSymbol
	ScalarFunctionPublicType               = scalarFunctionPublicType
	ScalarFunctionTypeSymbol               = scalarFunctionTypeSymbol
	ScalarType                             = scalarType
	SchemaFromTableDetails                 = schemaFromTableDetails
	TableIsOpen                            = tableIsOpen
	TableSchemaFromSymbol                  = tableSchemaFromSymbol
	TabularFunctionSchema                  = tabularFunctionSchema
	TabularFunctionSymbolFromDetails       = tabularFunctionSymbolFromDetails
	TypeNameForSchemaCopy                  = typeNameForSchemaCopy
	TypeSymbolFromColumnTypeDetails        = typeSymbolFromColumnTypeDetails
	UniqueColumnName                       = uniqueColumnName
	IsIdentifierLike                       = isIdentifierLike
	CleanIdentifier                        = cleanIdentifier
	DynamicPropertiesFromTypeSymbol        = dynamicPropertiesFromTypeSymbol
	DynamicScalarElementTypeFromTypeSymbol = dynamicScalarElementTypeFromTypeSymbol
	DynamicBagPropertiesFromSymbol         = dynamicBagPropertiesFromSymbol
)

const (
	EmptyTableTypeName        = emptyTableTypeName
	NonScalarNamedTypePrefix  = nonScalarNamedTypePrefix
	MaxColumnTypeDetailsDepth = maxColumnTypeDetailsDepth
)

type OrderedJSONField = orderedJSONField
