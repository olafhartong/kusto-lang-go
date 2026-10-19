package kustolang

// Ported from src/Kusto.Language/Utils/DictionaryExtensions.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

func DictionaryGetOrAdd[K comparable, V any](dictionary map[K]V, key K, value V) V {
	if existingValue, ok := dictionary[key]; ok {
		return existingValue
	}
	dictionary[key] = value
	return value
}

func DictionaryGetOrAddFunc[K comparable, V any](dictionary map[K]V, key K, valueFactory func(K) V) V {
	if existingValue, ok := dictionary[key]; ok {
		return existingValue
	}
	value := valueFactory(key)
	dictionary[key] = value
	return value
}

func DictionaryAddOrUpdate[K comparable, V any](dictionary map[K]V, key K, value V) {
	dictionary[key] = value
}

func DictionaryAddOrUpdateFunc[K comparable, V any](dictionary map[K]V, key K, addFactory func(K) V, updateFactory func(K, V) V) {
	if oldValue, ok := dictionary[key]; ok {
		dictionary[key] = updateFactory(key, oldValue)
	} else {
		dictionary[key] = addFactory(key)
	}
}

func DictionaryAddOrUpdateValueFunc[K comparable, V any](dictionary map[K]V, key K, addFactory func(K, V) V, updateFactory func(K, V, V) V, value V) {
	if oldValue, ok := dictionary[key]; ok {
		dictionary[key] = updateFactory(key, oldValue, value)
	} else {
		dictionary[key] = addFactory(key, value)
	}
}
