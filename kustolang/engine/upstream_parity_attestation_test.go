package kustolang

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

type attestedDiagnostic struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
	Start   int    `json:"start"`
	Length  int    `json:"length"`
}

type attestedColumn struct {
	Name string `json:"Name"`
	Type string `json:"type"`
}

type attestedLanguageResult struct {
	Query       string               `json:"query"`
	Diagnostics []attestedDiagnostic `json:"diagnostics"`
	Columns     []attestedColumn     `json:"columns"`
	IsOpen      *bool                `json:"isOpen"`
	ResultType  *string              `json:"resultType"`
}

func TestCurrentLanguageMatchesCommitBoundParityAttestations(t *testing.T) {
	for _, artifact := range []string{
		"kql-v7-native.canonical.json",
		"kql-command-native.canonical.json",
	} {
		t.Run(artifact, func(t *testing.T) {
			expected := readAttestedLanguageResults(t, artifact)
			for index, want := range expected {
				got := currentAttestedLanguageResult(want.Query)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("query %d %q\nnative:   %#v\nattested: %#v", index, want.Query, got, want)
				}
			}
		})
	}
}

func readAttestedLanguageResults(t *testing.T, name string) []attestedLanguageResult {
	t.Helper()
	path := filepath.Join("..", "..", "parity-artifacts", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var results []attestedLanguageResult
	if err := json.Unmarshal(data, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatalf("attestation %s is empty", name)
	}
	for index := range results {
		if results[index].Diagnostics == nil {
			results[index].Diagnostics = []attestedDiagnostic{}
		}
		if results[index].Columns == nil {
			results[index].Columns = []attestedColumn{}
		}
	}
	return results
}

func currentAttestedLanguageResult(query string, globals ...*GlobalState) attestedLanguageResult {
	code := ParseAndAnalyze(query, globals...)
	result := attestedLanguageResult{
		Query:       query,
		Diagnostics: []attestedDiagnostic{},
		Columns:     []attestedColumn{},
	}
	for _, diagnostic := range code.GetDiagnostics() {
		result.Diagnostics = append(result.Diagnostics, attestedDiagnostic{
			Code: diagnostic.Code, Message: diagnostic.Message,
			Start: diagnostic.Start, Length: diagnostic.Length,
		})
	}
	if code.ResultType != nil {
		name := code.ResultType.SymbolName()
		result.ResultType = &name
	}
	if table, ok := TableSymbolForType(code.ResultType); ok {
		open := table.IsOpen()
		result.IsOpen = &open
		for _, column := range table.Columns() {
			result.Columns = append(result.Columns, attestedColumn{Name: column.SymbolName(), Type: column.Type.SymbolName()})
		}
		sort.Slice(result.Columns, func(i, j int) bool { return result.Columns[i].Name < result.Columns[j].Name })
	}
	return result
}
