package kustolang

import "testing"

func TestParseOptionsWithMethodsMirrorUpstreamImmutability(t *testing.T) {
	defaults := DefaultParseOptions
	if !defaults.AlwaysProduceEndToken {
		t.Fatal("default parse options should always produce an end token")
	}
	if defaults.AllowLiteralsWithLineBreaks || defaults.AllowNonAdjacentWildcardParts || defaults.ParserKind != ParserKindDefault {
		t.Fatalf("default parse options = %#v", defaults)
	}

	changed := defaults.
		WithAllowLiteralsWithLineBreaks(true).
		WithAllowNonAdjacentWildcardParts(true).
		WithParserKind(ParserKindGrammar)
	if !changed.AllowLiteralsWithLineBreaks || !changed.AllowNonAdjacentWildcardParts || changed.ParserKind != ParserKindGrammar {
		t.Fatalf("changed parse options = %#v", changed)
	}
	if defaults.AllowLiteralsWithLineBreaks || defaults.AllowNonAdjacentWildcardParts || defaults.ParserKind != ParserKindDefault {
		t.Fatalf("default parse options were mutated = %#v", defaults)
	}
	if !defaults.EqualExceptForParseKind(defaults.WithParserKind(ParserKindGrammar)) {
		t.Fatal("EqualExceptForParseKind should ignore parser kind")
	}
	if defaults.Equal(changed) {
		t.Fatal("Equal should include parser kind and option flags")
	}
}

func TestGlobalStateCarriesParseOptionsAndServerKind(t *testing.T) {
	globals := NewGlobalState(nil)
	if globals.ServerKind != ServerKindEngine {
		t.Fatalf("server kind = %q", globals.ServerKind)
	}
	if !globals.ParseOptions.Equal(DefaultParseOptions) {
		t.Fatalf("parse options = %#v", globals.ParseOptions)
	}

	customOptions := globals.ParseOptions.WithAllowNonAdjacentWildcardParts(true)
	changed := globals.WithServerKind(ServerKindDataManager).WithParseOptions(customOptions)
	if changed == globals {
		t.Fatal("With helpers should return a cloned global state")
	}
	if changed.ServerKind != ServerKindDataManager {
		t.Fatalf("changed server kind = %q", changed.ServerKind)
	}
	if !changed.ParseOptions.Equal(customOptions) {
		t.Fatalf("changed parse options = %#v", changed.ParseOptions)
	}
	if globals.ServerKind != ServerKindEngine || globals.ParseOptions.AllowNonAdjacentWildcardParts {
		t.Fatalf("original global state was mutated = %#v", globals)
	}

	cached := changed.WithCache()
	if cached.WithParseOptions(cached.ParseOptions) != cached {
		t.Fatal("WithParseOptions should preserve identity when options are unchanged")
	}
	if cached.WithParseOptions(cached.ParseOptions).Cache != cached.Cache {
		t.Fatal("unchanged parse options should preserve the enabled cache")
	}
}

func TestKustoCodeDialectUsesCodeKindAndServerKind(t *testing.T) {
	if got := Parse("T | count").Dialect(); got != KustoDialectQuery {
		t.Fatalf("query dialect = %v", got)
	}
	if got := Parse(".show tables").Dialect(); got != KustoDialectEngineCommand {
		t.Fatalf("default command dialect = %v", got)
	}
	if got := ParseWithGlobals(".show tables", NewGlobalState(nil).WithServerKind(ServerKindDataManager)).Dialect(); got != KustoDialectDataManagerCommand {
		t.Fatalf("data manager command dialect = %v", got)
	}
	if got := ParseWithGlobals(".show tables", NewGlobalState(nil).WithServerKind(ServerKindClusterManager)).Dialect(); got != KustoDialectClusterManagerCommand {
		t.Fatalf("cluster manager command dialect = %v", got)
	}
}

func TestCodeKindSkipsLeadingDirectiveLines(t *testing.T) {
	if got := CodeKindOf("#crp\n.show tables"); got != CodeKindCommand {
		t.Fatalf("directive before command kind = %q", got)
	}
	if got := CodeKindOf("#crp\nT | count"); got != CodeKindQuery {
		t.Fatalf("directive before query kind = %q", got)
	}
	if got := CodeKindOf("#crp"); got != CodeKindQuery {
		t.Fatalf("standalone directive kind = %q", got)
	}
}

func TestKustoCodeDispatchesParserKind(t *testing.T) {
	text := "print x=1 | project x"
	defaults := NewGlobalState(nil)
	grammarGlobals := defaults.WithParseOptions(defaults.ParseOptions.WithParserKind(ParserKindGrammar))

	native := ParseAndAnalyze(text, defaults)
	grammar := ParseAndAnalyze(text, grammarGlobals)
	if native.Query == nil || grammar.Query == nil || len(native.GetDiagnostics()) != 0 || len(grammar.GetDiagnostics()) != 0 {
		t.Fatalf("parser kind results = native:%#v grammar:%#v", native.GetDiagnostics(), grammar.GetDiagnostics())
	}
	if len(native.Query.Segments) == 0 {
		t.Fatal("default QueryParser path did not retain pipe segments")
	}
	if len(grammar.Query.Segments) != 0 {
		t.Fatalf("QueryGrammar path unexpectedly used QueryParser segments: %#v", grammar.Query.Segments)
	}
	if native.ResultType == nil || grammar.ResultType == nil {
		t.Fatalf("parser kind result types = native:%#v grammar:%#v", native.ResultType, grammar.ResultType)
	}
}
