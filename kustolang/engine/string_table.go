package kustolang

// Ported from src/Kusto.Language/Utils/StringTable.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type StringTable struct {
	values *TextKeyedDictionary[string]
}

func NewStringTable() *StringTable {
	return &StringTable{values: NewTextKeyedDictionary[string]()}
}

func (t *StringTable) Add(text string) string {
	if result, ok := t.values.TryGetValue(text, 0, len(text)); ok {
		return result
	}
	return t.values.GetOrAddValueRangeFunc(text, 0, len(text), func(text string, start int, length int) string {
		if start > 0 || length < len(text) {
			return text[start : start+length]
		}
		return text
	})
}

func (t *StringTable) AddAll(strings []string) {
	for _, value := range strings {
		t.Add(value)
	}
}

func (t *StringTable) AddRange(text string, start int, length int) string {
	if result, ok := t.values.TryGetValue(text, start, length); ok {
		return result
	}
	return t.values.GetOrAddValueRangeFunc(text, start, length, func(text string, start int, length int) string {
		return text[start : start+length]
	})
}

func (t *StringTable) Contains(text string) bool {
	return t.values.ContainsKey(text)
}

func (t *StringTable) ContainsKey(text string, start int, length int) bool {
	return t.values.ContainsKeyRange(text, start, length)
}

func (t *StringTable) Values() []string {
	return t.values.Values()
}
