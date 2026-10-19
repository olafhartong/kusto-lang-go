package kustolang

import "fmt"

// Ported from src/Kusto.Language/Utils/ArgumentCheckers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ArgumentNullError struct {
	ParameterName string
}

func (e ArgumentNullError) Error() string {
	if e.ParameterName == "" {
		return "argument is nil"
	}
	return fmt.Sprintf("argument %q is nil", e.ParameterName)
}

type ArgumentError struct {
	ParameterName string
	Message       string
}

func (e ArgumentError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.ParameterName == "" {
		return "invalid argument"
	}
	return fmt.Sprintf("invalid argument %q", e.ParameterName)
}

func CheckArgumentNull[T any](value T, parameterName string) T {
	if listItemIsNil(value) {
		panic(ArgumentNullError{ParameterName: parameterName})
	}
	return value
}

func CheckArgumentNullOrEmpty[T any](list []T, parameterName string) []T {
	if list == nil {
		panic(ArgumentNullError{ParameterName: parameterName})
	}
	if len(list) == 0 {
		panic(ArgumentError{
			ParameterName: parameterName,
			Message:       fmt.Sprintf("Parameter %q must have at least one element.", parameterName),
		})
	}
	return list
}

func CheckArgumentNullOrElementNull[T any](list []T, parameterName string) []T {
	if list == nil {
		panic(ArgumentNullError{ParameterName: parameterName})
	}
	for i, item := range list {
		if listItemIsNil(item) {
			panic(ArgumentError{
				ParameterName: parameterName,
				Message:       fmt.Sprintf("Element %d of parameter %q is nil.", i, parameterName),
			})
		}
	}
	return list
}

func CheckArgumentNullOrEmptyOrElementNull[T any](list []T, parameterName string) []T {
	CheckArgumentNullOrElementNull(list, parameterName)
	CheckArgumentNullOrEmpty(list, parameterName)
	return list
}
