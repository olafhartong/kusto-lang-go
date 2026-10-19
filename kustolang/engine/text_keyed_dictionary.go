package kustolang

// Ported from src/Kusto.Language/Utils/TextKeyedDictionary.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type TextKeyedDictionary[V any] struct {
	values map[uint32][]textKeyedEntry[V]
}

type textKeyedEntry[V any] struct {
	text   string
	start  int
	length int
	value  V
}

func NewTextKeyedDictionary[V any]() *TextKeyedDictionary[V] {
	return &TextKeyedDictionary[V]{values: map[uint32][]textKeyedEntry[V]{}}
}

func (d *TextKeyedDictionary[V]) TryGetValue(text string, start int, length int) (V, bool) {
	var zero V
	if d == nil {
		return zero, false
	}
	hash := textKeyedFNVHash(text, start, length)
	for _, entry := range d.values[hash] {
		if entry.matches(text, start, length) {
			return entry.value, true
		}
	}
	return zero, false
}

func (d *TextKeyedDictionary[V]) GetOrAddValue(text string, newValue V) V {
	return d.GetOrAddValueRange(text, 0, len(text), newValue)
}

func (d *TextKeyedDictionary[V]) GetOrAddValueRange(text string, start int, length int, newValue V) V {
	if value, ok := d.TryGetValue(text, start, length); ok {
		return value
	}
	d.add(text, start, length, newValue)
	return newValue
}

func (d *TextKeyedDictionary[V]) GetOrAddValueFunc(text string, evaluator func(string, int, int) V) V {
	return d.GetOrAddValueRangeFunc(text, 0, len(text), evaluator)
}

func (d *TextKeyedDictionary[V]) GetOrAddValueRangeFunc(text string, start int, length int, evaluator func(string, int, int) V) V {
	if value, ok := d.TryGetValue(text, start, length); ok {
		return value
	}
	value := evaluator(text, start, length)
	d.add(text, start, length, value)
	return value
}

func (d *TextKeyedDictionary[V]) ContainsKey(text string) bool {
	return d.ContainsKeyRange(text, 0, len(text))
}

func (d *TextKeyedDictionary[V]) ContainsKeyRange(text string, start int, length int) bool {
	if d == nil || start < 0 || length < 1 || start+length > len(text) {
		return false
	}
	_, ok := d.TryGetValue(text, start, length)
	return ok
}

func (d *TextKeyedDictionary[V]) Values() []V {
	if d == nil {
		return nil
	}
	values := make([]V, 0)
	for _, bucket := range d.values {
		for _, entry := range bucket {
			values = append(values, entry.value)
		}
	}
	return values
}

func ToTextKeyedDictionary[S any, V any](source []S, keySelector func(S) string, valueSelector func(S) V) *TextKeyedDictionary[V] {
	dictionary := NewTextKeyedDictionary[V]()
	for _, item := range source {
		dictionary.GetOrAddValue(keySelector(item), valueSelector(item))
	}
	return dictionary
}

func (d *TextKeyedDictionary[V]) add(text string, start int, length int, value V) {
	hash := textKeyedFNVHash(text, start, length)
	d.values[hash] = append(d.values[hash], textKeyedEntry[V]{
		text:   text,
		start:  start,
		length: length,
		value:  value,
	})
}

func (e textKeyedEntry[V]) matches(text string, start int, length int) bool {
	if e.length != length {
		return false
	}
	for i := 0; i < length; i++ {
		if e.text[e.start+i] != text[start+i] {
			return false
		}
	}
	return true
}

const (
	textKeyedFNVOffsetBias uint32 = 2166136261
	textKeyedFNVPrime      uint32 = 16777619
)

func textKeyedFNVHash(text string, start int, length int) uint32 {
	hash := textKeyedFNVOffsetBias
	end := start + length
	for i := start; i < end; i++ {
		hash = (hash ^ uint32(text[i])) * textKeyedFNVPrime
	}
	return hash
}
