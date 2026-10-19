package kustolang

import (
	"fmt"
	"reflect"
	"strings"
)

// Ported from src/Kusto.Language/Binder/CallSiteInfo.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type CallSiteInfo struct {
	Signature  *Signature
	Parameters []*Parameter
	Values     []any
}

func NewCallSiteInfo(signature *Signature, parameters []*Parameter, values []any) CallSiteInfo {
	return CallSiteInfo{
		Signature:  signature,
		Parameters: append([]*Parameter{}, parameters...),
		Values:     append([]any{}, values...),
	}
}

func (c CallSiteInfo) String() string {
	name := ""
	if c.Signature != nil && c.Signature.Symbol != nil {
		name = c.Signature.Symbol.SymbolName()
	}
	parts := make([]string, 0, len(c.Values))
	for i, value := range c.Values {
		parameterName := ""
		if i < len(c.Parameters) && c.Parameters[i] != nil {
			parameterName = c.Parameters[i].Name
		}
		parts = append(parts, fmt.Sprintf("%s=%v", parameterName, value))
	}
	return name + "(" + strings.Join(parts, ",") + ")"
}

func (c CallSiteInfo) Equals(other CallSiteInfo) bool {
	if c.Signature != other.Signature {
		return false
	}
	if len(c.Parameters) != len(other.Parameters) {
		return false
	}
	for i := range c.Parameters {
		if c.Parameters[i] != other.Parameters[i] {
			return false
		}
	}
	if len(c.Values) != len(other.Values) {
		return false
	}
	for i := range c.Values {
		if left, leftOK := callSiteTableSymbol(c.Values[i]); leftOK {
			right, rightOK := callSiteTableSymbol(other.Values[i])
			if !rightOK || !AreTableResultsEquivalent(left, right) {
				return false
			}
			continue
		}
		if !reflect.DeepEqual(c.Values[i], other.Values[i]) {
			return false
		}
	}
	return true
}

func (c CallSiteInfo) Equal(other CallSiteInfo) bool {
	return c.Equals(other)
}

func callSiteTableSymbol(value any) (Symbol, bool) {
	switch v := value.(type) {
	case *TableSymbol:
		return v, v != nil
	case TableSymbol:
		return &v, true
	case *ExternalTableSymbol:
		if v == nil {
			return nil, false
		}
		return &v.TableSymbol, true
	case *MaterializedViewSymbol:
		if v == nil {
			return nil, false
		}
		return &v.TableSymbol, true
	case Symbol:
		_, ok := tableSymbolFromSymbol(v)
		return v, ok
	default:
		return nil, false
	}
}
