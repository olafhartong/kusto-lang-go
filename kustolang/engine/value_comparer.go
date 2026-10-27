package kustolang

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// Ported from src/Kusto.Language/Utils/ValueComparer.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func ValueAreEquivalent(x any, y any, caseSensitive ...bool) bool {
	sensitive := false
	if len(caseSensitive) > 0 {
		sensitive = caseSensitive[0]
	}

	xNil := valueIsNilReference(x)
	yNil := valueIsNilReference(y)
	if xNil && yNil {
		return true
	}
	if xNil || yNil {
		return false
	}

	if _, ok := x.(string); ok {
		return compareValueStrings(valueToString(x), valueToString(y), sensitive)
	}
	if _, ok := y.(string); ok {
		return compareValueStrings(valueToString(x), valueToString(y), sensitive)
	}
	if reflect.TypeOf(x) == reflect.TypeOf(y) {
		return sameTypeValuesEqual(x, y)
	}

	if isNumericValue(x) && isNumericValue(y) {
		return numericValuesEquivalent(x, y)
	}
	return false
}

func valueIsNilReference(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func sameTypeValuesEqual(left, right any) bool {
	switch leftValue := left.(type) {
	case float32:
		rightValue := right.(float32)
		return leftValue == rightValue || math.IsNaN(float64(leftValue)) && math.IsNaN(float64(rightValue))
	case float64:
		rightValue := right.(float64)
		return leftValue == rightValue || math.IsNaN(leftValue) && math.IsNaN(rightValue)
	}

	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)
	switch leftValue.Kind() {
	case reflect.Map, reflect.Pointer, reflect.Chan:
		return leftValue.Pointer() == rightValue.Pointer()
	case reflect.Slice:
		return leftValue.Pointer() == rightValue.Pointer() && leftValue.Len() == rightValue.Len() && leftValue.Cap() == rightValue.Cap()
	case reflect.Func:
		// Non-nil Go functions do not expose delegate-instance identity.
		return false
	}
	if leftValue.Type().Comparable() {
		return leftValue.Interface() == rightValue.Interface()
	}
	return reflect.DeepEqual(left, right)
}

func compareValueStrings(left string, right string, caseSensitive bool) bool {
	if caseSensitive {
		return left == right
	}
	return ordinalEqualFold(left, right)
}

func valueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "True"
		}
		return "False"
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'G', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'G', -1, 64)
	default:
		return fmt.Sprint(value)
	}
}

func isNumericValue(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func numericValuesEquivalent(x any, y any) bool {
	if numericRank(x) == numericRankFloat64 || numericRank(y) == numericRankFloat64 {
		return toFloat64(x) == toFloat64(y)
	}
	if numericRank(x) == numericRankFloat32 || numericRank(y) == numericRankFloat32 {
		return float32(toFloat64(x)) == float32(toFloat64(y))
	}
	if numericRank(x) == numericRankUint64 || numericRank(y) == numericRankUint64 {
		if numericValueIsNegative(x) || numericValueIsNegative(y) {
			panic("kustolang: negative value cannot be converted to UInt64")
		}
		return toUint64(x) == toUint64(y)
	}
	return toInt64(x) == toInt64(y)
}

func numericValueIsNegative(value any) bool {
	switch value := value.(type) {
	case int:
		return value < 0
	case int8:
		return value < 0
	case int16:
		return value < 0
	case int32:
		return value < 0
	case int64:
		return value < 0
	case float32:
		return value < 0
	case float64:
		return value < 0
	default:
		return false
	}
}

const (
	numericRankInt64 = iota
	numericRankUint64
	numericRankFloat32
	numericRankFloat64
)

func numericRank(value any) int {
	switch value.(type) {
	case float64:
		return numericRankFloat64
	case float32:
		return numericRankFloat32
	case uint64:
		return numericRankUint64
	default:
		return numericRankInt64
	}
}

func toInt64(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func toUint64(value any) uint64 {
	switch v := value.(type) {
	case int:
		return uint64(v)
	case int8:
		return uint64(v)
	case int16:
		return uint64(v)
	case int32:
		return uint64(v)
	case int64:
		return uint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case float32:
		return uint64(v)
	case float64:
		return uint64(v)
	default:
		return 0
	}
}

func toFloat64(value any) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}
