package kustolang

// Ported from the following files at upstream commit
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856:
//
//   src/Kusto.Language/Utils/ConnectionInfo.cs
//   src/Kusto.Language/Utils/MostRecentlyUsedCache.cs
//   src/Kusto.Language/Utils/ObjectPool.cs
//   src/Kusto.Language/Utils/Optional.cs
//   src/Kusto.Language/Utils/StringAndNumberComparer.cs
//   src/Kusto.Language/Utils/StringTable.cs
//   src/Kusto.Language/Utils/TextKeyedDictionary.cs
//   src/Kusto.Language/Utils/ThreadSafeDictionary.cs
//   src/Kusto.Language/Utils/UniqueNameTable.cs
//   src/Kusto.Language/Utils/ValueComparer.cs
//
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import (
	"encoding/binary"
	"reflect"
	"sync"

	internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"
)

// ConnectionInfo is the immutable parsed form of a Kusto connection string.
type ConnectionInfo struct {
	value *internal.ConnectionInfo
}

func ParseConnectionInfo(text string) *ConnectionInfo {
	return &ConnectionInfo{value: internal.ParseConnectionInfo(text)}
}

func (c *ConnectionInfo) GetPart(name string) string {
	if c == nil {
		return ""
	}
	return c.value.GetPart(name)
}

func (c *ConnectionInfo) Text() string {
	if c == nil {
		return ""
	}
	return c.value.Text()
}

func (c *ConnectionInfo) DataSource() string {
	if c == nil {
		return ""
	}
	return c.value.DataSource()
}

// Optional is the Go adaptation of the source's immutable Optional<T> struct.
type Optional[T any] struct {
	value    T
	hasValue bool
}

func NewOptional[T any](value T) Optional[T] {
	return Optional[T]{value: value, hasValue: true}
}

func (o Optional[T]) Value() T       { return o.value }
func (o Optional[T]) HasValue() bool { return o.hasValue }

// UniqueNameTable allocates stable numeric suffixes for duplicate names.
type UniqueNameTable struct {
	value *internal.UniqueNameTable
}

func NewUniqueNameTable() *UniqueNameTable {
	return &UniqueNameTable{value: internal.NewUniqueNameTable()}
}

func (t *UniqueNameTable) AddNames(names []string) { t.internalValue().AddNames(names) }
func (t *UniqueNameTable) AddName(name string)     { t.internalValue().AddName(name) }

func (t *UniqueNameTable) GetOrAddName(name string, baseName ...string) string {
	return t.internalValue().GetOrAddName(name, baseName...)
}

func (t *UniqueNameTable) Clear() { t.internalValue().Clear() }

func (t *UniqueNameTable) internalValue() *internal.UniqueNameTable {
	if t == nil {
		panic("kustolang: nil UniqueNameTable")
	}
	if t.value == nil {
		t.value = internal.NewUniqueNameTable()
	}
	return t.value
}

type textKeyedDictionaryEntry[V any] struct {
	key   string
	value V
}

// TextKeyedDictionary indexes values by UTF-16 subranges without changing the
// parser's existing byte-oriented internal dictionary contract.
type TextKeyedDictionary[V any] struct {
	entries map[uint32][]textKeyedDictionaryEntry[V]
	order   []textKeyedDictionaryEntry[V]
}

func NewTextKeyedDictionary[V any]() *TextKeyedDictionary[V] {
	return &TextKeyedDictionary[V]{entries: make(map[uint32][]textKeyedDictionaryEntry[V])}
}

func (d *TextKeyedDictionary[V]) TryGetValue(text string, start, length int) (V, bool) {
	key := utf16RangeKey(text, start, length)
	hash := utf16RangeHash(key)
	for _, entry := range d.internalEntries()[hash] {
		if entry.key == key {
			return entry.value, true
		}
	}
	var zero V
	return zero, false
}

func (d *TextKeyedDictionary[V]) GetOrAddValue(text string, newValue V) V {
	return d.GetOrAddValueRange(text, 0, len(UTF16CodeUnits(text)), newValue)
}

func (d *TextKeyedDictionary[V]) GetOrAddValueRange(text string, start, length int, newValue V) V {
	if value, ok := d.TryGetValue(text, start, length); ok {
		return value
	}
	d.add(utf16RangeKey(text, start, length), newValue)
	return newValue
}

func (d *TextKeyedDictionary[V]) GetOrAddValueFunc(text string, evaluator func(string, int, int) V) V {
	return d.GetOrAddValueRangeFunc(text, 0, len(UTF16CodeUnits(text)), evaluator)
}

func (d *TextKeyedDictionary[V]) GetOrAddValueRangeFunc(text string, start, length int, evaluator func(string, int, int) V) V {
	if value, ok := d.TryGetValue(text, start, length); ok {
		return value
	}
	value := evaluator(text, start, length)
	d.add(utf16RangeKey(text, start, length), value)
	return value
}

func (d *TextKeyedDictionary[V]) ContainsKey(text string) bool {
	return d.ContainsKeyRange(text, 0, len(UTF16CodeUnits(text)))
}

func (d *TextKeyedDictionary[V]) ContainsKeyRange(text string, start, length int) bool {
	units := UTF16CodeUnits(text)
	if start < 0 || length < 1 || start > len(units) || length > len(units)-start {
		return false
	}
	_, ok := d.TryGetValue(text, start, length)
	return ok
}

// Values returns a defensive insertion-ordered snapshot.
func (d *TextKeyedDictionary[V]) Values() []V {
	entries := d.internalOrder()
	values := make([]V, len(entries))
	for i, entry := range entries {
		values[i] = entry.value
	}
	return values
}

func (d *TextKeyedDictionary[V]) add(key string, value V) {
	d.ensureInitialized()
	entry := textKeyedDictionaryEntry[V]{key: key, value: value}
	hash := utf16RangeHash(key)
	d.entries[hash] = append(d.entries[hash], entry)
	d.order = append(d.order, entry)
}

func (d *TextKeyedDictionary[V]) ensureInitialized() {
	if d == nil {
		panic("kustolang: nil TextKeyedDictionary")
	}
	if d.entries == nil {
		d.entries = make(map[uint32][]textKeyedDictionaryEntry[V])
	}
}

func (d *TextKeyedDictionary[V]) internalEntries() map[uint32][]textKeyedDictionaryEntry[V] {
	d.ensureInitialized()
	return d.entries
}

func (d *TextKeyedDictionary[V]) internalOrder() []textKeyedDictionaryEntry[V] {
	if d == nil {
		return nil
	}
	return d.order
}

func utf16RangeKey(text string, start, length int) string {
	units := UTF16CodeUnits(text)
	if start < 0 || length < 0 || start > len(units) || length > len(units)-start {
		panic("kustolang: UTF-16 range outside string")
	}
	key := make([]byte, length*2)
	for i, unit := range units[start : start+length] {
		binary.LittleEndian.PutUint16(key[i*2:], unit)
	}
	return string(key)
}

func utf16RangeHash(key string) uint32 {
	const (
		fnvOffsetBias uint32 = 2166136261
		fnvPrime      uint32 = 16777619
	)
	hash := fnvOffsetBias
	for i := 0; i < len(key); i += 2 {
		hash = (hash ^ uint32(binary.LittleEndian.Uint16([]byte(key[i:i+2])))) * fnvPrime
	}
	return hash
}

func ToTextKeyedDictionary[S any, V any](source []S, keySelector func(S) string, valueSelector func(S) V) *TextKeyedDictionary[V] {
	dictionary := NewTextKeyedDictionary[V]()
	for _, item := range source {
		dictionary.GetOrAddValue(keySelector(item), valueSelector(item))
	}
	return dictionary
}

// StringTable interns strings and accepts source-compatible UTF-16 ranges.
type StringTable struct {
	values *TextKeyedDictionary[string]
}

func NewStringTable() *StringTable {
	return &StringTable{values: NewTextKeyedDictionary[string]()}
}

func (t *StringTable) Add(text string) string {
	return t.internalValues().GetOrAddValueFunc(text, func(_ string, _, _ int) string { return text })
}

func (t *StringTable) AddAll(values []string) {
	for _, value := range values {
		t.Add(value)
	}
}

func (t *StringTable) AddRange(text string, start, length int) string {
	return t.internalValues().GetOrAddValueRangeFunc(text, start, length, func(text string, start, length int) string {
		return textFactsSubstring(UTF16CodeUnits(text), start, length)
	})
}

func (t *StringTable) Contains(text string) bool {
	return t.internalValues().ContainsKey(text)
}

func (t *StringTable) ContainsKey(text string, start, length int) bool {
	return t.internalValues().ContainsKeyRange(text, start, length)
}

func (t *StringTable) Values() []string { return t.internalValues().Values() }

func (t *StringTable) internalValues() *TextKeyedDictionary[string] {
	if t == nil {
		panic("kustolang: nil StringTable")
	}
	if t.values == nil {
		t.values = NewTextKeyedDictionary[string]()
	}
	return t.values
}

type threadSafeDictionaryEntry[K any, V any] struct {
	key     K
	value   V
	version uint64
}

// ThreadSafeDictionary adapts ConcurrentDictionary and supports a custom key
// comparer without requiring Go keys to be comparable.
type ThreadSafeDictionary[K any, V any] struct {
	mu          sync.RWMutex
	equal       func(K, K) bool
	entries     []threadSafeDictionaryEntry[K, V]
	nextVersion uint64
}

func NewThreadSafeDictionary[K any, V any]() *ThreadSafeDictionary[K, V] {
	return NewThreadSafeDictionaryWithComparer[K, V](func(left, right K) bool {
		return reflect.DeepEqual(left, right)
	})
}

func NewThreadSafeDictionaryWithComparer[K any, V any](equal func(K, K) bool) *ThreadSafeDictionary[K, V] {
	if equal == nil {
		panic("kustolang: nil dictionary comparer")
	}
	return &ThreadSafeDictionary[K, V]{equal: equal}
}

func (d *ThreadSafeDictionary[K, V]) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.entries)
}

func (d *ThreadSafeDictionary[K, V]) Keys() []K {
	d.mu.RLock()
	defer d.mu.RUnlock()
	keys := make([]K, len(d.entries))
	for i, entry := range d.entries {
		keys[i] = entry.key
	}
	return keys
}

func (d *ThreadSafeDictionary[K, V]) Values() []V {
	d.mu.RLock()
	defer d.mu.RUnlock()
	values := make([]V, len(d.entries))
	for i, entry := range d.entries {
		values[i] = entry.value
	}
	return values
}

func (d *ThreadSafeDictionary[K, V]) TryGetValue(key K) (V, bool) {
	requireDictionaryKey(key)
	d.mu.RLock()
	defer d.mu.RUnlock()
	if index := d.indexOf(key); index >= 0 {
		return d.entries[index].value, true
	}
	var zero V
	return zero, false
}

func (d *ThreadSafeDictionary[K, V]) TryAdd(key K, value V) bool {
	requireDictionaryKey(key)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.indexOf(key) >= 0 {
		return false
	}
	d.entries = append(d.entries, threadSafeDictionaryEntry[K, V]{key: key, value: value, version: d.newVersion()})
	return true
}

func (d *ThreadSafeDictionary[K, V]) GetOrAdd(key K, value V) V {
	requireDictionaryKey(key)
	d.mu.Lock()
	defer d.mu.Unlock()
	if index := d.indexOf(key); index >= 0 {
		return d.entries[index].value
	}
	d.entries = append(d.entries, threadSafeDictionaryEntry[K, V]{key: key, value: value, version: d.newVersion()})
	return value
}

func (d *ThreadSafeDictionary[K, V]) GetOrAddFunc(key K, valueFactory func(K) V) V {
	requireDictionaryKey(key)
	if value, ok := d.TryGetValue(key); ok {
		return value
	}
	value := valueFactory(key)
	d.mu.Lock()
	defer d.mu.Unlock()
	if index := d.indexOf(key); index >= 0 {
		return d.entries[index].value
	}
	d.entries = append(d.entries, threadSafeDictionaryEntry[K, V]{key: key, value: value, version: d.newVersion()})
	return value
}

func (d *ThreadSafeDictionary[K, V]) AddOrUpdate(key K, value V) {
	requireDictionaryKey(key)
	d.mu.Lock()
	defer d.mu.Unlock()
	if index := d.indexOf(key); index >= 0 {
		d.entries[index].value = value
		d.entries[index].version = d.newVersion()
		return
	}
	d.entries = append(d.entries, threadSafeDictionaryEntry[K, V]{key: key, value: value, version: d.newVersion()})
}

func (d *ThreadSafeDictionary[K, V]) AddOrUpdateFunc(key K, addFactory func(K) V, updateFactory func(K, V) V) {
	requireDictionaryKey(key)
	for {
		oldValue, version, exists := d.valueAndVersion(key)
		if exists {
			newValue := updateFactory(key, oldValue)
			d.mu.Lock()
			index := d.indexOf(key)
			if index >= 0 && d.entries[index].version == version {
				d.entries[index].value = newValue
				d.entries[index].version = d.newVersion()
				d.mu.Unlock()
				return
			}
			d.mu.Unlock()
			continue
		}

		newValue := addFactory(key)
		d.mu.Lock()
		if d.indexOf(key) < 0 {
			d.entries = append(d.entries, threadSafeDictionaryEntry[K, V]{key: key, value: newValue, version: d.newVersion()})
			d.mu.Unlock()
			return
		}
		d.mu.Unlock()
	}
}

func (d *ThreadSafeDictionary[K, V]) AddOrUpdateValueFunc(key K, addFactory func(K, V) V, updateFactory func(K, V, V) V, value V) {
	requireDictionaryKey(key)
	d.AddOrUpdateFunc(
		key,
		func(key K) V { return addFactory(key, value) },
		func(key K, oldValue V) V { return updateFactory(key, oldValue, value) },
	)
}

func (d *ThreadSafeDictionary[K, V]) indexOf(key K) int {
	for i, entry := range d.entries {
		if d.equal(entry.key, key) {
			return i
		}
	}
	return -1
}

func (d *ThreadSafeDictionary[K, V]) valueAndVersion(key K) (V, uint64, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if index := d.indexOf(key); index >= 0 {
		entry := d.entries[index]
		return entry.value, entry.version, true
	}
	var zero V
	return zero, 0, false
}

func (d *ThreadSafeDictionary[K, V]) newVersion() uint64 {
	d.nextVersion++
	return d.nextVersion
}

func requireDictionaryKey[K any](key K) {
	value := reflect.ValueOf(key)
	if !value.IsValid() {
		panic("kustolang: nil dictionary key")
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			panic("kustolang: nil dictionary key")
		}
	}
}

// ObjectPool is a bounded concurrent object reuse pool.
type ObjectPool[T any] struct {
	value *internal.ObjectPool[T]
}

func NewObjectPool[T any](creator func() T, resetter func(T), size ...int) *ObjectPool[T] {
	return &ObjectPool[T]{value: internal.NewObjectPool(creator, resetter, size...)}
}

func (p *ObjectPool[T]) AllocateFromPool() T { return p.value.AllocateFromPool() }
func (p *ObjectPool[T]) ReturnToPool(item T) { p.value.ReturnToPool(item) }

type mostRecentlyUsedPair[K any, V any] struct {
	key   K
	value V
}

// MostRecentlyUsedCache retains the most recently accessed entries. The
// pinned source intentionally keeps an existing value on AddOrUpdate and only
// refreshes its recency; this adaptation preserves that observable behavior.
type MostRecentlyUsedCache[K any, V any] struct {
	mu      sync.Mutex
	equal   func(K, K) bool
	pairs   []mostRecentlyUsedPair[K, V]
	maxSize int
}

func NewMostRecentlyUsedCache[K any, V any](maxSize ...int) *MostRecentlyUsedCache[K, V] {
	return NewMostRecentlyUsedCacheWithComparer[K, V](func(left, right K) bool {
		return reflect.DeepEqual(left, right)
	}, maxSize...)
}

func NewMostRecentlyUsedCacheWithComparer[K any, V any](equal func(K, K) bool, maxSize ...int) *MostRecentlyUsedCache[K, V] {
	if equal == nil {
		panic("kustolang: nil cache comparer")
	}
	size := 10
	if len(maxSize) > 0 {
		size = maxSize[0]
	}
	return &MostRecentlyUsedCache[K, V]{equal: equal, maxSize: size}
}

func (c *MostRecentlyUsedCache[K, V]) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.pairs)
}

func (c *MostRecentlyUsedCache[K, V]) TryGetValue(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, pair := range c.pairs {
		if c.equal(pair.key, key) {
			c.pairs = append(c.pairs[:i], c.pairs[i+1:]...)
			c.pairs = append([]mostRecentlyUsedPair[K, V]{pair}, c.pairs...)
			return pair.value, true
		}
	}
	var zero V
	return zero, false
}

func (c *MostRecentlyUsedCache[K, V]) AddOrUpdate(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, pair := range c.pairs {
		if c.equal(pair.key, key) {
			c.pairs = append(c.pairs[:i], c.pairs[i+1:]...)
			c.pairs = append([]mostRecentlyUsedPair[K, V]{{key: key, value: pair.value}}, c.pairs...)
			return
		}
	}
	c.pairs = append([]mostRecentlyUsedPair[K, V]{{key: key, value: value}}, c.pairs...)
	for len(c.pairs) > c.maxSize {
		if len(c.pairs) == 0 {
			panic("kustolang: negative MostRecentlyUsedCache size")
		}
		c.pairs = c.pairs[:len(c.pairs)-1]
	}
}

// ValueComparerAPI mirrors the source's static ValueComparer class.
type ValueComparerAPI struct{}

func (ValueComparerAPI) AreEquivalent(left, right any, caseSensitive ...bool) bool {
	return internal.ValueAreEquivalent(left, right, caseSensitive...)
}

var ValueComparer ValueComparerAPI

type StringAndNumberComparerValue struct {
	caseSensitive bool
}

func (c StringAndNumberComparerValue) Compare(left, right string) int {
	return internal.StringAndNumberCompare(left, right, c.caseSensitive)
}

type StringAndNumberComparerAPI struct {
	Ordinal           StringAndNumberComparerValue
	OrdinalIgnoreCase StringAndNumberComparerValue
}

var StringAndNumberComparer = StringAndNumberComparerAPI{
	Ordinal:           StringAndNumberComparerValue{caseSensitive: true},
	OrdinalIgnoreCase: StringAndNumberComparerValue{caseSensitive: false},
}
