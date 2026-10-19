// Command gen-kusto-parity-corpus emits the stable current-profile structural
// mutation corpus used for Kusto language parity testing.
//
// The Kusto syntax exercised here is derived from Microsoft Kusto.Language,
// licensed under the Apache License 2.0. See third_party/kusto/LICENSE.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const structuralRunes = "()[]{},;:=|"

type queryFamily struct {
	seed      string
	malformed []string
}

var queryFamilies = []queryFamily{
	{
		seed: `datatable(A:long,B:string)[1,"x"] | as hint.materialized=true Rows; Rows | project A, B`,
		malformed: []string{
			`datatable(A:long)[1] | as as Rows`,
			`datatable(A:long)[1] | as hint.materialized= Rows`,
		},
	},
	{
		seed: `datatable(A:long,B:string,C:real)[1,"x",2.5] | filter A > 0 | project-by-names dynamic(["A","B"])`,
		malformed: []string{
			`datatable(A:long)[1] | filter filter`,
			`datatable(A:long)[1] | filter A > 0 | project-by-names,`,
		},
	},
	{
		seed: `datatable(Message:string)["alpha beta","alpha gamma"] | reduce by Message with threshold=0.5, characters="abcdefghijklmnopqrstuvwxyz" | project Pattern, Count, Representative`,
		malformed: []string{
			`datatable(Message:string)["x"] | reduce by with threshold=`,
			`datatable(Message:string)["x"] | reduce by Message with threshold=0.5,`,
		},
	},
	{
		seed: `datatable(User:string,Score:long)["alice",10,"bob",5] | top-hitters 2 of User by Score`,
		malformed: []string{
			`datatable(User:string,Score:long)["a",1] | top-hitters 2 User by Score`,
			`datatable(User:string,Score:long)["a",1] | top-hitters 2 of User Score`,
		},
	},
	{
		seed: `datatable(Source:long,Target:long,Weight:long)[1,2,3] | make-graph Source --> Target | graph-mark-components with_component_id=Component | graph-where-nodes true | graph-where-edges true | graph-to-table edges`,
		malformed: []string{
			`datatable(Source:long,Target:long)[1,2] | make-graph Source --> Target | graph-where-nodes`,
			`datatable(Source:long,Target:long)[1,2] | make-graph Source --> Target | graph-mark-components with_component_id=`,
		},
	},
	{
		seed: `alias database Current = database(""); print x=1`,
		malformed: []string{
			`alias database = database("")`,
			`alias database Current database("")`,
		},
	},
	{
		seed: `set query_take_max_records=10; range Value from 1 to 3 step 1`,
		malformed: []string{
			`set query_take_max_records=`,
			`set query_take_max_records==10`,
		},
	},
	{
		seed: `datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant (summarize Total=sum(Value))`,
		malformed: []string{
			`datatable(Tenant:string,Value:long)["a",1] | __partitionby (summarize Total=sum(Value))`,
			`datatable(Tenant:string,Value:long)["a",1] | __partitionby Tenant ()`,
		},
	},
}

func main() {
	if err := writeCorpus(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "gen-kusto-parity-corpus: %v\n", err)
		os.Exit(1)
	}
}

func writeCorpus(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(buildCorpus())
}

func buildCorpus() []string {
	queries := make([]string, 0)
	for _, family := range queryFamilies {
		queries = append(queries, family.seed)
		queries = append(queries, structuralMutations(family.seed)...)
		queries = append(queries, family.malformed...)
	}

	unique := make(map[string]struct{}, len(queries))
	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query != "" {
			unique[query] = struct{}{}
		}
	}

	corpus := make([]string, 0, len(unique))
	for query := range unique {
		corpus = append(corpus, query)
	}
	sort.Strings(corpus)
	return corpus
}

func structuralMutations(seed string) []string {
	runes := []rune(seed)
	mutations := make([]string, 0, len(runes))
	for i, r := range runes {
		if isStructuralRune(r) {
			mutation := make([]rune, 0, len(runes)-1)
			mutation = append(mutation, runes[:i]...)
			mutation = append(mutation, runes[i+1:]...)
			mutations = append(mutations, string(mutation))
		}
		if r == ' ' || isStructuralRune(r) {
			mutations = append(mutations, string(runes[:i]))
		}
	}
	return mutations
}

func isStructuralRune(r rune) bool {
	return strings.ContainsRune(structuralRunes, r)
}
