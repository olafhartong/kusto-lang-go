package kustolang

// Ported from src/Kusto.Language/Editor/Kusto/KustoFixers.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type KustoFixersCatalog struct {
	all []KustoFixer
}

var KustoFixers = newKustoFixersCatalog()

func newKustoFixersCatalog() *KustoFixersCatalog {
	analyzers := KustoAnalyzers.All()
	fixers := make([]KustoFixer, 0, len(analyzers))
	for _, analyzer := range analyzers {
		if analyzer != nil {
			fixers = append(fixers, analyzer.Fixer())
		}
	}
	return &KustoFixersCatalog{all: ListToReadOnly(fixers)}
}

func (c *KustoFixersCatalog) All() []KustoFixer {
	if c == nil {
		return EmptyReadOnlyList[KustoFixer]()
	}
	return ListToReadOnly(c.all)
}
