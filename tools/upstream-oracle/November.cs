using System;
using System.Linq;
using Kusto.Language;
using Kusto.Language.Parsing;
using Kusto.Language.Editor;
using Kusto.Language.Syntax;

static class November
{
    public static object Build(string commit)
    {
        var texts = new[] {
            "", " ", "null", "NULL", "min", "max", "MIN", "MAX", "now", "NOW", "nOw", "bad", "0", "+0", "-0", "-1", "1", "999", "2026", "0000", "9999", "0001", "202611", "20261108", "2026110812", "202611081234", "20261108123456", "20260230", "20261301", "20261108 12:34:56", "20261108 12:34:56.1", "20261108 12:34:56.1234567", "20261108 12:34:56.12345678",
            "599266080000000000", "599266079999999999", "621355968000000000", "638000000000000001", "3155378975999999999", "3155378976000000000", "253402300799", "253402300800", "9223372036854775807", "9223372036854775808", "-9223372036854775808", "1,000", "1e3", "0x10", "+2026",
            "2026-11-08", "2026/11/08", "2026/11-08", "2026-11/08", "2026-1-2", "2026-11-08T12:34:56Z", "2026-11-08t12:34Z", "2026-11-08 T 12 : 34 : 56 Z", "2026-11-08T1:02", "2026-11-08T12:3", "2026-11-08T12:34:56.1234567Z", "2026-11-08T12:34:56.12345675Z", "2026-11-08T12:34:56,12345678Z", "2026-11-08T23:59:59.99999995Z", "9999-12-31T23:59:59.99999995Z", "2026-11-08T12:34:56+14:00", "2026-11-08T12:34:56+14:01", "2026-11-08T12:34:56-1400", "2026-11-08T12:34:56+1", "2026-11-08T12:34:56+1:2", "2026-11-08T12:34:56+123", "2026-11-08T12:34:56+01:60", "0001-01-01T00:00:00+14:00", "0001-01-01T00:00:00-14:00", "9999-12-31T23:59:59-14:00", "0000-01-01", "2026-02-29", "2024-02-29", "2026-01-32", "2026-11-08 24:00:00", "2026-11-08 23:59:60", "2026-11-08Z", "2026-11-08 GMT", "2026-11-08 PM", "2026-11-08 1 PM", "2026-11-08 0 AM", "2026-11-08 12 AM", "2026-11-08 23 PM", "2026-11-08 1:2:3.12345675 PM", "2026-11-08 Sunday", "2026-11-08 Monday", "2026-11-08 Sunday 12:00 GMT",
            "11/08/2026", "11/8/26", "11/8/49", "11/8/50", "11/8/0001", "11/8/2026T12:00", "11/8/2026 1:2 PM", "Nov 8 2026", "November 8, 2026", "8 Nov 2026", "2026 November 8", "Nov-8-2026", "Nov.8.2026", "Nov..8.2026", "Nov..8.2099", "Nov 31 2026", "Feb 29 2026", "February 29 2024", "Sun, 08 Nov 2026 12:34:56 GMT", "Sunday, November 8, 2026", "2026-11", "November 2026", "11/2026", "\u00a02026-11-08\u00a0"
        };
        var dateTimes = texts.Select(text => {
            var interpretation = KustoDateTimeParser.Parse(text, out var value);
            return new {text, interpretation = interpretation.ToString(), ticks = KustoDateTimeParser.HasConstantValue(interpretation) ? (long?)value.Ticks : null};
        }).ToArray();
        var analyzerQueries = texts.Select(t => "// 😀\nprint D=datetime('" + t.Replace("'", "\\'") + "')")
            .Concat(new[] {"print datetime()", "print datetime(null)", "print date(2026)", "print datetime('')", "print datetime(   )", "print datetime('Nov 8')", "print datetime('11/8')", "print datetime('12:30')", "print datetime('12 PM')", "print datetime('8 Nov')", "print datetime('November')", "print datetime('November 8, 26')", "print datetime('2026.11.08')", "print datetime('Feb 30')", "print datetime('25:00')"}).ToArray();
        var analyzer = new DateTimeLiteralAnalyzer();
        var diagnostics = analyzerQueries.Select(query => {
            var code = KustoCode.ParseAndAnalyze(query);
            var ds = new System.Collections.Generic.List<Diagnostic>();
            analyzer.Analyze(code, ds, default);
            return new {query, diagnostics = ds.Select(d => new {d.Code,d.Message,start=d.Start,length=d.Length}).ToArray()};
        }).ToArray();
        var parserQueries = new[] {"let partition=(x:long){x}; print partition(1)", "let render=(x:long){x}; print render(1)", "print 'x' has_any ('a')", "print 'x' has_all ('a')", "print typeof(long)", "inline_external_table(A:long) kind=storage dataformat=csv ('u')"};
        var parsers = parserQueries.Select(query => {
            var code = KustoCode.ParseAndAnalyze(query, GlobalState.Default.WithParseOptions(ParseOptions.Default.WithParserKind(ParserKind.Grammar)));
            return new {query, diagnostics=code.GetDiagnostics().Select(d=>new {d.Code,d.Message,start=d.Start,length=d.Length}).ToArray(),kinds=code.Syntax.GetDescendants<SyntaxNode>().Select(n=>n.Kind.ToString()).ToArray()};
        }).ToArray();
        var commands = new[] {"snapshot_indexing","graph_query"}.SelectMany(policy => new[] {
            ".show graph_model M policy " + policy,
            ".show graph_model * policy " + policy,
            ".alter graph_model M policy " + policy + " '{}'",
            ".delete graph_model M policy " + policy,
            ".show graph_model db.M policy " + policy,
            ".alter graph_model ['model name'] policy " + policy + " '{}'",
            ".alter graph_model M policy " + policy,
            ".alter graph_model M policy " + policy + " 1",
            ".delete graph_model * policy " + policy,
        }).Select(query => {
            var code=KustoCode.ParseAndAnalyze(query);
            var table=code.ResultType as Kusto.Language.Symbols.TableSymbol;
            return new {query,diagnostics=code.GetDiagnostics().Select(d=>new {d.Code,d.Message,start=d.Start,length=d.Length}).ToArray(),columns=table?.Columns.OrderBy(c=>c.Name,StringComparer.Ordinal).Select(c=>new {c.Name,type=c.Type.Name}).ToArray(),isOpen=table==null?(bool?)null:table.IsOpen,resultType=code.ResultType?.Name};
        }).ToArray();
        var completions = new[] {"print 1 ", "let ", "query_parameters ", "print \"😀\"; let ", "// 😀\nlet "}.Select(text => new {
            text, items = new KustoCodeService(text).GetCompletionItems(text.Length, CompletionOptions.Default.WithIncludeSymbols(false).WithIncludeExtendedSyntax(false)).Items.Select(i=>i.DisplayText).OrderBy(x=>x,StringComparer.Ordinal).ToArray()
        }).ToArray();
        var grammar = QueryGrammar.From(GlobalState.Default);
        var prefixes = new[] {"fork (count)", "__partitionby Tenant id (count)"}.Select(text => new {
            text, length = (text.StartsWith("fork") ? (Parser<LexicalToken>)grammar.Expression : grammar.QueryOperator).Scan(new ArraySource<LexicalToken>(TokenParser.ParseTokens(text)), 0)
        }).ToArray();
        return new {commit,dateTimes,diagnostics,parsers,commands,completions,prefixes};
    }
}
