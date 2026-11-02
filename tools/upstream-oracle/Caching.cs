using System;
using System.Linq;
using Kusto.Language;
using Kusto.Language.Parsing;
using Kusto.Language.Symbols;

static class Caching
{
    public static object Build(string commit)
    {
        var sourceCache = new SourceCache();
        var sourceCalls = 0;
        var first = sourceCache.GetOrCreate(() => { sourceCalls++; return new object(); });
        var again = sourceCache.GetOrCreate(() => { sourceCalls++; return new object(); });
        var kustoCache = new KustoCache(GlobalState.Default);
        var kustoCalls = 0;
        var kustoFirst = kustoCache.GetOrCreate(() => { kustoCalls++; return new object(); });
        var kustoAgain = kustoCache.GetOrCreate(() => { kustoCalls++; return new object(); });
        var retryCache = new SourceCache();
        var retryCalls = 0;
        try { retryCache.GetOrCreate<object>(() => { retryCalls++; throw new InvalidOperationException(); }); }
        catch (InvalidOperationException) { }
        var retryValue = retryCache.GetOrCreate(() => { retryCalls++; return new object(); });
        var grammars = Enumerable.Range(0, 8).Select(index => {
            var options = ParseOptions.Default.WithAlwaysProduceEndTokens((index & 1) != 0)
                .WithAllowLiteralsWithLineBreaks((index & 2) != 0)
                .WithAllowNonAdjacentWildcardParts((index & 4) != 0);
            var firstGrammar = QueryGrammar.From(GlobalState.Default.WithParseOptions(options.WithParserKind(ParserKind.Grammar)));
            QueryGrammar.From(GlobalState.Default.WithParseOptions(options.WithAlwaysProduceEndTokens((index & 1) == 0)));
            var otherGlobals = GlobalState.Default.WithDatabase(new DatabaseSymbol("other")).WithCache().WithParseOptions(options);
            return new { index,
                reusedAfterOptionSwitch = ReferenceEquals(firstGrammar, QueryGrammar.From(GlobalState.Default.WithParseOptions(options))),
                reusedAcrossGlobals = ReferenceEquals(firstGrammar, QueryGrammar.From(otherGlobals)),
                reusedAcrossParserKinds = ReferenceEquals(firstGrammar, QueryGrammar.From(otherGlobals.WithParseOptions(options.WithParserKind(ParserKind.Grammar)))) };
        }).ToArray();
        var variants = Enumerable.Range(0, 8).Select(index => QueryGrammar.From(GlobalState.Default.WithParseOptions(
            ParseOptions.Default.WithAlwaysProduceEndTokens((index & 1) != 0)
                .WithAllowLiteralsWithLineBreaks((index & 2) != 0)
                .WithAllowNonAdjacentWildcardParts((index & 4) != 0)))).ToArray();
        return new { commit, sourceCalls, sourceReused = ReferenceEquals(first, again),
            kustoCalls, kustoReused = ReferenceEquals(kustoFirst, kustoAgain),
            retryCalls, retryStored = retryCache.TryGetValue<object>(out var stored) && ReferenceEquals(stored, retryValue),
            distinctGrammars = variants.Distinct().Count(), grammars };
    }
}
