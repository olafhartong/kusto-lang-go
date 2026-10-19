package kustolang

// Ported from src/Kusto.Language/Binder/ParameterMatchKind.cs.
// Values are ordered from weakest to strongest match, as in the upstream enum.

type ParameterMatchKind int

const (
	ParameterMatchNone ParameterMatchKind = iota
	ParameterMatchUnknown
	ParameterMatchDynamic
	ParameterMatchCompatible
	ParameterMatchPromoted
	ParameterMatchNotType
	ParameterMatchOneOfMany
	ParameterMatchScalar
	ParameterMatchSummable
	ParameterMatchOrderable
	ParameterMatchNumber
	ParameterMatchInteger
	ParameterMatchTabular
	ParameterMatchTable
	ParameterMatchDatabase
	ParameterMatchCluster
	ParameterMatchExact
)
