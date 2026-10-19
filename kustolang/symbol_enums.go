package kustolang

// Ported from the public enums under src/Kusto.Language/Symbols.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

type ParameterLayout = internal.ParameterLayout
type ParameterLayoutBuilder = internal.ParameterLayoutBuilder
type TableState = internal.TableState
type CombineKind = internal.CombineKind
type ResultNameKind = internal.ResultNameKind

const (
	CombineKindUnifySameNameAndType = internal.CombineKindUnifySameNameAndType
	CombineKindUnifySameName        = internal.CombineKindUnifySameName
	CombineKindUniqueNames          = internal.CombineKindUniqueNames

	ResultNameNone                       = internal.ResultNameNone
	ResultNameNameAndFirstArgument       = internal.ResultNameNameAndFirstArgument
	ResultNameNameAndOnlyArgument        = internal.ResultNameNameAndOnlyArgument
	ResultNamePrefixAndFirstArgument     = internal.ResultNamePrefixAndFirstArgument
	ResultNamePrefixAndOnlyArgument      = internal.ResultNamePrefixAndOnlyArgument
	ResultNameFirstArgumentValueIfColumn = internal.ResultNameFirstArgumentValueIfColumn
	ResultNamePrefixOnly                 = internal.ResultNamePrefixOnly
	ResultNameFirstArgument              = internal.ResultNameFirstArgument
	ResultNameOnlyArgument               = internal.ResultNameOnlyArgument
	ResultNameDefault                    = internal.ResultNameDefault

	ArgumentKindExpression                  = internal.ArgumentKindExpression
	ArgumentKindExpressionParameter0Element = internal.ArgumentKindExpressionParameter0Element
	ArgumentKindStarOnly                    = internal.ArgumentKindStarOnly
	ArgumentKindStarAllowed                 = internal.ArgumentKindStarAllowed
	ArgumentKindAggregate                   = internal.ArgumentKindAggregate
	ArgumentKindColumn                      = internal.ArgumentKindColumn
	ArgumentKindColumnOrLiteral             = internal.ArgumentKindColumnOrLiteral
	ArgumentKindColumnParameter0            = internal.ArgumentKindColumnParameter0
	ArgumentKindColumnParameter0Common      = internal.ArgumentKindColumnParameter0Common
	ArgumentKindLiteral                     = internal.ArgumentKindLiteral
	ArgumentKindLiteralNotEmpty             = internal.ArgumentKindLiteralNotEmpty
	ArgumentKindConstant                    = internal.ArgumentKindConstant
	ArgumentKindStar                        = internal.ArgumentKindStar

	ParameterTypeAny                   = internal.ParameterTypeAny
	ParameterTypeDeclared              = internal.ParameterTypeDeclared
	ParameterTypeScalar                = internal.ParameterTypeScalar
	ParameterTypeTabular               = internal.ParameterTypeTabular
	ParameterTypeDatabase              = internal.ParameterTypeDatabase
	ParameterTypeCluster               = internal.ParameterTypeCluster
	ParameterTypeInteger               = internal.ParameterTypeInteger
	ParameterTypeRealOrDecimal         = internal.ParameterTypeRealOrDecimal
	ParameterTypeStringOrDynamic       = internal.ParameterTypeStringOrDynamic
	ParameterTypeStringOrArray         = internal.ParameterTypeStringOrArray
	ParameterTypeIntegerOrArray        = internal.ParameterTypeIntegerOrArray
	ParameterTypeDynamicArray          = internal.ParameterTypeDynamicArray
	ParameterTypeDynamicBag            = internal.ParameterTypeDynamicBag
	ParameterTypeNumber                = internal.ParameterTypeNumber
	ParameterTypeNumberOrBool          = internal.ParameterTypeNumberOrBool
	ParameterTypeSummable              = internal.ParameterTypeSummable
	ParameterTypeOrderable             = internal.ParameterTypeOrderable
	ParameterTypeNotRealOrBool         = internal.ParameterTypeNotRealOrBool
	ParameterTypeNotDynamic            = internal.ParameterTypeNotDynamic
	ParameterTypeNotBool               = internal.ParameterTypeNotBool
	ParameterTypeParameter0            = internal.ParameterTypeParameter0
	ParameterTypeParameter1            = internal.ParameterTypeParameter1
	ParameterTypeParameter2            = internal.ParameterTypeParameter2
	ParameterTypeCommonScalar          = internal.ParameterTypeCommonScalar
	ParameterTypeCommonNumber          = internal.ParameterTypeCommonNumber
	ParameterTypeCommonSummable        = internal.ParameterTypeCommonSummable
	ParameterTypeCommonOrderable       = internal.ParameterTypeCommonOrderable
	ParameterTypeCommonScalarOrDynamic = internal.ParameterTypeCommonScalarOrDynamic

	ReturnTypeDeclared                    = internal.ReturnTypeDeclared
	ReturnTypeComputed                    = internal.ReturnTypeComputed
	ReturnTypeCustom                      = internal.ReturnTypeCustom
	ReturnTypeParameter0                  = internal.ReturnTypeParameter0
	ReturnTypeParameter1                  = internal.ReturnTypeParameter1
	ReturnTypeParameter2                  = internal.ReturnTypeParameter2
	ReturnTypeParameterN                  = internal.ReturnTypeParameterN
	ReturnTypeParameter0Literal           = internal.ReturnTypeParameter0Literal
	ReturnTypeParameter1Literal           = internal.ReturnTypeParameter1Literal
	ReturnTypeParameterNLiteral           = internal.ReturnTypeParameterNLiteral
	ReturnTypeParameter0Promoted          = internal.ReturnTypeParameter0Promoted
	ReturnTypeParameter0Cluster           = internal.ReturnTypeParameter0Cluster
	ReturnTypeParameter0Database          = internal.ReturnTypeParameter0Database
	ReturnTypeParameter0Table             = internal.ReturnTypeParameter0Table
	ReturnTypeParameter0ExternalTable     = internal.ReturnTypeParameter0ExternalTable
	ReturnTypeParameter0Array             = internal.ReturnTypeParameter0Array
	ReturnTypeCommon                      = internal.ReturnTypeCommon
	ReturnTypeCommonNonDynamic            = internal.ReturnTypeCommonNonDynamic
	ReturnTypeWidest                      = internal.ReturnTypeWidest
	ReturnTypeParameter0MaterializedView  = internal.ReturnTypeParameter0MaterializedView
	ReturnTypeParameter0EntityGroup       = internal.ReturnTypeParameter0EntityGroup
	ReturnTypeParameter0StoredQueryResult = internal.ReturnTypeParameter0StoredQueryResult
	ReturnTypeParameter0Graph             = internal.ReturnTypeParameter0Graph

	ParameterLayoutFixed          = internal.ParameterLayoutFixed
	ParameterLayoutRepeating      = internal.ParameterLayoutRepeating
	ParameterLayoutRepeatingSkip  = internal.ParameterLayoutRepeatingSkip
	ParameterLayoutBlockRepeating = internal.ParameterLayoutBlockRepeating
	ParameterLayoutCustom         = internal.ParameterLayoutCustom
	MaxRepeat                     = internal.MaxRepeat
	TableStateNone                = internal.TableStateNone
	TableStateSerialized          = internal.TableStateSerialized
	TableStateSorted              = internal.TableStateSorted
	TableStateOpen                = internal.TableStateOpen

	ConversionNone       = internal.ConversionNone
	ConversionPromotable = internal.ConversionPromotable
	ConversionDynamic    = internal.ConversionDynamic
	ConversionCompatible = internal.ConversionCompatible
	ConversionAny        = internal.ConversionAny

	OperatorKindNone               = internal.OperatorKindNone
	OperatorKindUnaryMinus         = internal.OperatorKindUnaryMinus
	OperatorKindUnaryPlus          = internal.OperatorKindUnaryPlus
	OperatorKindAdd                = internal.OperatorKindAdd
	OperatorKindSubtract           = internal.OperatorKindSubtract
	OperatorKindMultiply           = internal.OperatorKindMultiply
	OperatorKindDivide             = internal.OperatorKindDivide
	OperatorKindModulo             = internal.OperatorKindModulo
	OperatorKindEqualTilde         = internal.OperatorKindEqualTilde
	OperatorKindBangTilde          = internal.OperatorKindBangTilde
	OperatorKindHas                = internal.OperatorKindHas
	OperatorKindHasCs              = internal.OperatorKindHasCs
	OperatorKindNotHas             = internal.OperatorKindNotHas
	OperatorKindNotHasCs           = internal.OperatorKindNotHasCs
	OperatorKindHasPrefix          = internal.OperatorKindHasPrefix
	OperatorKindHasPrefixCs        = internal.OperatorKindHasPrefixCs
	OperatorKindNotHasPrefix       = internal.OperatorKindNotHasPrefix
	OperatorKindNotHasPrefixCs     = internal.OperatorKindNotHasPrefixCs
	OperatorKindHasSuffix          = internal.OperatorKindHasSuffix
	OperatorKindHasSuffixCs        = internal.OperatorKindHasSuffixCs
	OperatorKindNotHasSuffix       = internal.OperatorKindNotHasSuffix
	OperatorKindNotHasSuffixCs     = internal.OperatorKindNotHasSuffixCs
	OperatorKindLike               = internal.OperatorKindLike
	OperatorKindLikeCs             = internal.OperatorKindLikeCs
	OperatorKindNotLike            = internal.OperatorKindNotLike
	OperatorKindNotLikeCs          = internal.OperatorKindNotLikeCs
	OperatorKindContains           = internal.OperatorKindContains
	OperatorKindContainsCs         = internal.OperatorKindContainsCs
	OperatorKindNotContains        = internal.OperatorKindNotContains
	OperatorKindNotContainsCs      = internal.OperatorKindNotContainsCs
	OperatorKindStartsWith         = internal.OperatorKindStartsWith
	OperatorKindStartsWithCs       = internal.OperatorKindStartsWithCs
	OperatorKindNotStartsWith      = internal.OperatorKindNotStartsWith
	OperatorKindNotStartsWithCs    = internal.OperatorKindNotStartsWithCs
	OperatorKindEndsWith           = internal.OperatorKindEndsWith
	OperatorKindEndsWithCs         = internal.OperatorKindEndsWithCs
	OperatorKindNotEndsWith        = internal.OperatorKindNotEndsWith
	OperatorKindNotEndsWithCs      = internal.OperatorKindNotEndsWithCs
	OperatorKindMatchRegex         = internal.OperatorKindMatchRegex
	OperatorKindSearch             = internal.OperatorKindSearch
	OperatorKindLessThan           = internal.OperatorKindLessThan
	OperatorKindLessThanOrEqual    = internal.OperatorKindLessThanOrEqual
	OperatorKindGreaterThan        = internal.OperatorKindGreaterThan
	OperatorKindGreaterThanOrEqual = internal.OperatorKindGreaterThanOrEqual
	OperatorKindEqual              = internal.OperatorKindEqual
	OperatorKindNotEqual           = internal.OperatorKindNotEqual
	OperatorKindAnd                = internal.OperatorKindAnd
	OperatorKindOr                 = internal.OperatorKindOr
	OperatorKindIn                 = internal.OperatorKindIn
	OperatorKindInCs               = internal.OperatorKindInCs
	OperatorKindNotIn              = internal.OperatorKindNotIn
	OperatorKindNotInCs            = internal.OperatorKindNotInCs
	OperatorKindBetween            = internal.OperatorKindBetween
	OperatorKindNotBetween         = internal.OperatorKindNotBetween
	OperatorKindHasAny             = internal.OperatorKindHasAny
	OperatorKindHasAll             = internal.OperatorKindHasAll
)
