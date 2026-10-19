package kustolang

import "fmt"

// Ported from src/Kusto.Language/Utils/Ensure.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type InvalidOperationError struct {
	Message string
}

func (e InvalidOperationError) Error() string {
	return e.Message
}

func EnsureIsTrue(value bool, message ...string) {
	if !value {
		panic(InvalidOperationError{Message: ensureMessage("Expected true", message)})
	}
}

func EnsureAreEqual[T comparable](expected T, actual T, message ...string) {
	if expected != actual {
		defaultMessage := fmt.Sprintf("Expected: %v actual: %v", expected, actual)
		panic(InvalidOperationError{Message: ensureMessage(defaultMessage, message)})
	}
}

func EnsureNotNil(value any, message ...string) {
	if listItemIsNil(value) {
		panic(InvalidOperationError{Message: ensureMessage("Expected not null", message)})
	}
}

func EnsureIsNil(value any, message ...string) {
	if !listItemIsNil(value) {
		panic(InvalidOperationError{Message: ensureMessage("Expected null", message)})
	}
}

func EnsureArgumentNotNil(value any, paramName string) {
	if listItemIsNil(value) {
		panic(ArgumentNullError{ParameterName: paramName})
	}
}

func EnsureElementsNotNil[T any](list []T, listName string) {
	for i, item := range list {
		if listItemIsNil(item) {
			panic(InvalidOperationError{Message: fmt.Sprintf("%s[%d] is nil", listName, i)})
		}
	}
}

func ensureMessage(defaultMessage string, messages []string) string {
	if len(messages) > 0 && messages[0] != "" {
		return messages[0]
	}
	return defaultMessage
}
