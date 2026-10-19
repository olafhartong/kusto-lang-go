package kustolang

import "reflect"

// Ported from src/Kusto.Language/Properties.cs and the GlobalState property
// helpers in src/Kusto.Language/GlobalState.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type GlobalStateProperty interface {
	PropertyName() string
	globalStateProperty()
}

type TypedGlobalStateProperty[T any] struct {
	name         string
	defaultValue T
}

func NewGlobalStateProperty[T any](name string, defaultValue ...T) *TypedGlobalStateProperty[T] {
	var value T
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}
	return &TypedGlobalStateProperty[T]{name: name, defaultValue: value}
}

func (p *TypedGlobalStateProperty[T]) Name() string {
	return p.PropertyName()
}

func (p *TypedGlobalStateProperty[T]) PropertyName() string {
	if p == nil {
		return ""
	}
	return p.name
}

func (p *TypedGlobalStateProperty[T]) DefaultValue() T {
	var zero T
	if p == nil {
		return zero
	}
	return p.defaultValue
}

func (*TypedGlobalStateProperty[T]) globalStateProperty() {}

type GlobalStatePropertyAndValue struct {
	Property GlobalStateProperty
	Value    any
}

type kustoProperties struct {
	AllowClientParameters *TypedGlobalStateProperty[bool]
	MaxParseTextSize      *TypedGlobalStateProperty[int]
	MaxAnalysisDepth      *TypedGlobalStateProperty[int]
	MaxCachedExpansions   *TypedGlobalStateProperty[int]
	MaxCachedResultTypes  *TypedGlobalStateProperty[int]
}

var Properties = kustoProperties{
	AllowClientParameters: NewGlobalStateProperty[bool]("AllowClientParameters"),
	MaxParseTextSize:      NewGlobalStateProperty[int]("MaxParseTextSize", 4*1024*1024),
	MaxAnalysisDepth:      NewGlobalStateProperty[int]("MaxAnalysisDepth", 500),
	MaxCachedExpansions:   NewGlobalStateProperty[int]("MaxCachedExpansions", 10),
	MaxCachedResultTypes:  NewGlobalStateProperty[int]("MaxCachedResultTypes", 50),
}

func globalStatePropertyValuesEqual(left any, right any) bool {
	return reflect.DeepEqual(left, right)
}
