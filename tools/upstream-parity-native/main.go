// Command upstream-parity-native replays a query corpus for the development oracle.
package main

import (
	"encoding/json"
	"fmt"
	k "github.com/olafhartong/kusto-lang-go/kustolang"
	"os"
	"sort"
	"unicode/utf16"
)

type diagnostic struct {
	Code, Message string
	Start         int `json:"start"`
	Length        int `json:"length"`
}
type column struct {
	Name string
	Type string `json:"type"`
}
type result struct {
	Query       string       `json:"query"`
	Diagnostics []diagnostic `json:"diagnostics"`
	Columns     []column     `json:"columns"`
	IsOpen      *bool        `json:"isOpen"`
	ResultType  *string      `json:"resultType"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: upstream-parity-native corpus.json result.json")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var queries []string
	if err := json.Unmarshal(data, &queries); err != nil {
		panic(err)
	}
	results := make([]result, 0, len(queries))
	for _, query := range queries {
		code := k.ParseAndAnalyze(query)
		r := result{Query: query, Diagnostics: []diagnostic{}}
		for _, d := range code.GetDiagnostics() {
			r.Diagnostics = append(r.Diagnostics, diagnostic{d.Code, d.Message, d.Start, d.Length})
		}
		if code.ResultType != nil {
			name := code.ResultType.SymbolName()
			r.ResultType = &name
		}
		if table, ok := k.TableSymbolForType(code.ResultType); ok {
			open := table.IsOpen()
			r.IsOpen = &open
			r.Columns = []column{}
			for _, c := range table.Columns() {
				r.Columns = append(r.Columns, column{c.SymbolName(), c.Type.SymbolName()})
			}
			sort.SliceStable(r.Columns, func(i, j int) bool {
				l, r := utf16.Encode([]rune(r.Columns[i].Name)), utf16.Encode([]rune(r.Columns[j].Name))
				for n := 0; n < len(l) && n < len(r); n++ {
					if l[n] != r[n] {
						return l[n] < r[n]
					}
				}
				return len(l) < len(r)
			})
		}
		results = append(results, r)
	}
	data, err = json.Marshal(results)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(os.Args[2], append(data, '\n'), 0644); err != nil {
		panic(err)
	}
}
