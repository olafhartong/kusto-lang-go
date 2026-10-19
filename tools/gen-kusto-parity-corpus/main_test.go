// The corpus targets Microsoft Kusto.Language syntax, licensed under the
// Apache License 2.0. See third_party/kusto/LICENSE for source attribution.
package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const expectedCorpusCount = 258

func TestBuildCorpusDeterministicSortedUnique(t *testing.T) {
	first := buildCorpus()
	second := buildCorpus()

	if !reflect.DeepEqual(first, second) {
		t.Fatal("buildCorpus returned different results across calls")
	}
	if len(first) != expectedCorpusCount {
		t.Fatalf("corpus query count = %d, want %d", len(first), expectedCorpusCount)
	}
	if !sort.StringsAreSorted(first) {
		t.Fatal("corpus is not sorted")
	}
	for i, query := range first {
		if query == "" {
			t.Fatalf("corpus[%d] is empty", i)
		}
		if query != strings.TrimSpace(query) {
			t.Fatalf("corpus[%d] is not trimmed: %q", i, query)
		}
		if i > 0 && first[i-1] == query {
			t.Fatalf("corpus contains duplicate query %q", query)
		}
	}
}

func TestCorpusContainsValidSeedsAndMalformedCases(t *testing.T) {
	corpus := corpusSet(buildCorpus())
	validCount := 0
	invalidCount := 0

	for familyIndex, family := range queryFamilies {
		if _, ok := corpus[family.seed]; !ok {
			t.Errorf("family %d valid seed is absent: %q", familyIndex, family.seed)
		} else {
			validCount++
		}
		if len(family.malformed) < 2 {
			t.Errorf("family %d has %d explicit malformed cases, want at least 2", familyIndex, len(family.malformed))
		}
		for _, query := range family.malformed {
			if _, ok := corpus[query]; !ok {
				t.Errorf("family %d malformed case is absent: %q", familyIndex, query)
			} else {
				invalidCount++
			}
		}
	}

	if validCount == 0 || invalidCount == 0 {
		t.Fatalf("corpus mix = %d valid seeds and %d malformed cases", validCount, invalidCount)
	}
}

func TestCorpusContainsEveryStructuralMutation(t *testing.T) {
	corpus := corpusSet(buildCorpus())

	for familyIndex, family := range queryFamilies {
		runes := []rune(family.seed)
		for i, r := range runes {
			if isStructuralRune(r) {
				deleted := append([]rune(nil), runes[:i]...)
				deleted = append(deleted, runes[i+1:]...)
				assertCorpusContains(t, corpus, familyIndex, "deletion", string(deleted))
			}
			if r == ' ' || isStructuralRune(r) {
				assertCorpusContains(t, corpus, familyIndex, "truncation", string(runes[:i]))
			}
		}
	}
}

func TestWriteCorpusProducesStableValidJSON(t *testing.T) {
	var first bytes.Buffer
	if err := writeCorpus(&first); err != nil {
		t.Fatalf("writeCorpus: %v", err)
	}
	if !json.Valid(first.Bytes()) {
		t.Fatalf("writeCorpus emitted invalid JSON: %s", first.Bytes())
	}

	var decoded []string
	if err := json.Unmarshal(first.Bytes(), &decoded); err != nil {
		t.Fatalf("decode corpus: %v", err)
	}
	if !reflect.DeepEqual(decoded, buildCorpus()) {
		t.Fatal("JSON corpus differs from buildCorpus result")
	}

	var second bytes.Buffer
	if err := writeCorpus(&second); err != nil {
		t.Fatalf("second writeCorpus: %v", err)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("writeCorpus returned different JSON across calls")
	}
}

func corpusSet(corpus []string) map[string]struct{} {
	set := make(map[string]struct{}, len(corpus))
	for _, query := range corpus {
		set[query] = struct{}{}
	}
	return set
}

func assertCorpusContains(t *testing.T, corpus map[string]struct{}, familyIndex int, mutation string, query string) {
	t.Helper()
	query = strings.TrimSpace(query)
	if query == "" {
		return
	}
	if _, ok := corpus[query]; !ok {
		t.Errorf("family %d %s is absent: %q", familyIndex, mutation, query)
	}
}
