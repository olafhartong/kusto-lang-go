// Command gen-kusto-command-parity-corpus emits a stable structural mutation
// corpus for Kusto function-management commands.
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

const commandStructuralRunes = `.()[]{},;:=|"`

type commandFamily struct {
	seed      string
	malformed []string
}

var commandFamilies = []commandFamily{
	{
		seed: `.show functions`,
		malformed: []string{
			`.show`,
			`.show function`,
		},
	},
	{
		seed: `.show function F with (folder="Ops",docstring="seed")`,
		malformed: []string{
			`.show function F with (folder="Ops",docstring="seed"`,
			`.show function F with (folder="Ops",)`,
		},
	},
	{
		seed: `.create function ifnotexists with (folder="Ops",docstring="seed") F(x:long,y:string) { print Value=x, Label=y }`,
		malformed: []string{
			`.create function view () { print x=1 }`,
			`.create function ifnotexists with (folder="Ops") F(x:long,y:string { print Value=x }`,
		},
	},
	{
		seed: `.alter function with (folder="Ops") F(x:long, label:string="x") { print Value=x, Label=label }`,
		malformed: []string{
			`.alter function F view (x:long) { print Value=x }`,
			`.alter function F(x:long) { print Value= }`,
		},
	},
	{
		seed: `.create-or-alter function with (folder="Ops") G() { range x from 1 to 3 step 1 }`,
		malformed: []string{
			`.create-or-alter function with (folder) G() { print x=1 }`,
			`.create-or-alter function with (folder="Ops") G() { range x from 1 to 3 step 1`,
		},
	},
	{
		seed: `.drop functions (F,G) ifexists`,
		malformed: []string{
			`.drop functions (F,,G) ifexists`,
			`.drop functions (F,G ifexists`,
		},
	},
	{
		seed: `.alter function F docstring "seed docs"`,
		malformed: []string{
			`.alter function F docstring`,
			`.alter function F docstring 42`,
		},
	},
	{
		seed: `.show function F principal "aaduser=user@example.com" roles with (scope="all")`,
		malformed: []string{
			`.show function F principal "aaduser=user@example.com"`,
			`.show function F principal roles with (scope=)`,
		},
	},
	{
		seed: `.add function F admins ("aaduser=user@example.com","aadapp=00000000-0000-0000-0000-000000000001") skip-results "seed"`,
		malformed: []string{
			`.add function F admins ("aaduser=user@example.com",) skip-results "seed"`,
			`.add function F admins ("aaduser=user@example.com" skip-results "seed"`,
		},
	},
	{
		seed: `.set function F admins none skip-results`,
		malformed: []string{
			`.set function F admins none "notes"`,
			`.set function F admins (`,
		},
	},
}

func main() {
	if err := writeCorpus(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "gen-kusto-command-parity-corpus: %v\n", err)
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
	for _, family := range commandFamilies {
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
	return strings.ContainsRune(commandStructuralRunes, r)
}
