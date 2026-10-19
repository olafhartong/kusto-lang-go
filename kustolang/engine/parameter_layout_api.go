package kustolang

// Ported from src/Kusto.Language/Symbols/ParameterLayout.cs and
// src/Kusto.Language/Symbols/ParameterLayouts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// ParameterLayoutStrategy is the Go interface adaptation of the upstream
// extensible ParameterLayout base class.
type ParameterLayoutStrategy interface {
	Kind() ParameterLayout
	GetArgumentParameters(signature *Signature, arguments []Expression, globals ...*GlobalState) []*Parameter
	GetArgumentParametersForTypes(signature *Signature, argumentTypes []*TypeSymbol) []*Parameter
	GetNextPossibleParameters(signature *Signature, arguments []Expression) []*Parameter
	IsValidArgumentCount(signature *Signature, argumentCount int) bool
}

// ParameterLayoutDefinition is the canonical strategy used by the upstream
// ParameterLayouts catalog. ParameterLayout remains the existing comparable
// enum used by Signature.Layout.
type ParameterLayoutDefinition struct {
	kind    ParameterLayout
	builder ParameterLayoutBuilder
}

// ParameterLayoutsAPI mirrors the upstream static ParameterLayouts catalog.
type ParameterLayoutsAPI struct {
	Fixed             ParameterLayoutDefinition
	Repeating         ParameterLayoutDefinition
	RepeatingSkipping ParameterLayoutDefinition
	BlockRepeating    ParameterLayoutDefinition
}

// ParameterLayouts exposes the four shared layouts and constructs custom
// layout definitions without changing the existing Signature.Layout contract.
var ParameterLayouts = ParameterLayoutsAPI{
	Fixed:             ParameterLayoutDefinition{kind: ParameterLayoutFixed},
	Repeating:         ParameterLayoutDefinition{kind: ParameterLayoutRepeating},
	RepeatingSkipping: ParameterLayoutDefinition{kind: ParameterLayoutRepeatingSkip},
	BlockRepeating:    ParameterLayoutDefinition{kind: ParameterLayoutBlockRepeating},
}

func (ParameterLayoutsAPI) Custom(builder ParameterLayoutBuilder) ParameterLayoutDefinition {
	return ParameterLayoutDefinition{kind: ParameterLayoutCustom, builder: builder}
}

func (d ParameterLayoutDefinition) Kind() ParameterLayout {
	if d.kind == "" {
		return ParameterLayoutFixed
	}
	return d.kind
}

func (d ParameterLayoutDefinition) Apply(signature *Signature) *Signature {
	if signature == nil {
		return nil
	}
	if d.Kind() == ParameterLayoutCustom {
		return signature.WithCustomLayout(d.builder)
	}
	return signature.WithLayout(d.Kind())
}

func (d ParameterLayoutDefinition) GetArgumentParameters(signature *Signature, arguments []Expression, globals ...*GlobalState) []*Parameter {
	return d.Apply(signature).GetArgumentParameters(arguments, globals...)
}

func (d ParameterLayoutDefinition) GetArgumentParametersForTypes(signature *Signature, argumentTypes []*TypeSymbol) []*Parameter {
	return d.Apply(signature).GetArgumentParametersForTypes(argumentTypes)
}

func (d ParameterLayoutDefinition) GetNextPossibleParameters(signature *Signature, arguments []Expression) []*Parameter {
	return d.Apply(signature).GetNextPossibleParameters(arguments)
}

func (d ParameterLayoutDefinition) IsValidArgumentCount(signature *Signature, argumentCount int) bool {
	applied := d.Apply(signature)
	return applied != nil && applied.IsValidArgumentCount(argumentCount)
}

// WithParameterLayout applies any layout strategy while preserving the
// existing WithLayout and WithCustomLayout APIs.
func (s *Signature) WithParameterLayout(layout ParameterLayoutStrategy) *Signature {
	if s == nil || layout == nil {
		return s
	}
	if definition, ok := layout.(ParameterLayoutDefinition); ok {
		return definition.Apply(s)
	}
	if definition, ok := layout.(*ParameterLayoutDefinition); ok && definition != nil {
		return definition.Apply(s)
	}
	clone := *s
	clone.Parameters = append([]*Parameter{}, s.Parameters...)
	clone.Layout = layout.Kind()
	if clone.Layout == "" {
		clone.Layout = ParameterLayoutCustom
	}
	clone.CustomLayoutBuilder = nil
	clone.layoutStrategy = layout
	return &clone
}

// LayoutStrategy returns the behavior object represented by the signature's
// compatible Layout and CustomLayoutBuilder fields.
func (s *Signature) LayoutStrategy() ParameterLayoutStrategy {
	if s == nil {
		return nil
	}
	if s.layoutStrategy != nil {
		return s.layoutStrategy
	}
	switch s.Layout {
	case ParameterLayoutRepeating:
		return ParameterLayouts.Repeating
	case ParameterLayoutRepeatingSkip:
		return ParameterLayouts.RepeatingSkipping
	case ParameterLayoutBlockRepeating:
		return ParameterLayouts.BlockRepeating
	case ParameterLayoutCustom:
		return ParameterLayouts.Custom(s.CustomLayoutBuilder)
	default:
		return ParameterLayouts.Fixed
	}
}

func (s *Signature) HasCustomParameterLayout() bool {
	return s != nil && s.Layout == ParameterLayoutCustom && (s.CustomLayoutBuilder != nil || s.layoutStrategy != nil)
}

// GetArgumentParameters is the source-shaped expression overload. The
// existing explicit method remains available when callers already have richer
// result symbols or need a specific GlobalState for named-argument ownership.
func (s *Signature) GetArgumentParameters(arguments []Expression, globals ...*GlobalState) []*Parameter {
	return s.ArgumentParametersForExpressions(arguments, expressionResultSymbols(arguments), globals...)
}

func (s *Signature) GetArgumentParametersForTypes(argumentTypes []*TypeSymbol) []*Parameter {
	return s.ArgumentParametersForTypes(argumentTypes)
}

func (s *Signature) GetNextPossibleParameters(arguments []Expression) []*Parameter {
	return s.NextPossibleParametersForExpressions(arguments, expressionResultSymbols(arguments))
}

func expressionResultSymbols(arguments []Expression) []Symbol {
	result := make([]Symbol, len(arguments))
	for i, argument := range arguments {
		if argument != nil && argument.ResultType() != nil {
			result[i] = argument.ResultType()
		} else {
			result[i] = ScalarTypes.Unknown.Type()
		}
	}
	return result
}
