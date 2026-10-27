using System;
using System.Globalization;
using System.Linq;
using System.Reflection;
using Kusto.Language.Editor;
using Kusto.Language.Symbols;
using Kusto.Language;
using Kusto.Language.Parsing;
using Kusto.Language.Syntax;
using Kusto.Language.Utils;

static class Compatibility
{
    public static object Build()
    {
        var pairs = new[] {
            ("I", "i"), ("ı", "I"), ("ſ", "S"), ("K", "K"), ("ß", "ẞ"),
            ("µ", "Μ"), ("ς", "Σ"), ("σ", "Σ"), ("İ", "i"), ("Ⱥ", "ⱥ"),
            ("𐐀", "𐐨"), ("𐒰", "𐓘"), ("😀", "\ue000"), ("é", "É"), ("", ""),
            ("a-b", "ab"), ("a\0b", "ab"), ("ch", "h"), ("a", "B")
        };
        var wildcards = new[] {
            ("*SAMPLE*", "Samples"), ("prefix*MID*suffix", "prefix-mId-suffix"),
            ("*ſ*", "S"), ("*K*", "K"), ("*ı*", "I"), ("*𐐀*", "x𐐨y"),
            ("*Ⱥ*", "xⱥy"), ("Ⱥ*", "ⱥx"), ("*Ⱥ", "xⱥ"),
            ("*a*b", "aXab"), ("*", ""), ("", ""), ("a\0b", "ab")
        };
        var literals = new[] {
            (SyntaxKind.RealLiteralToken, "real(1,5)"), (SyntaxKind.RealLiteralToken, "real(1,,5)"),
            (SyntaxKind.RealLiteralToken, "real(1,.5)"), (SyntaxKind.RealLiteralToken, "real(1.5,3)"),
            (SyntaxKind.RealLiteralToken, "real(1e3,5)"), (SyntaxKind.RealLiteralToken, "real(,5)"),
            (SyntaxKind.RealLiteralToken, "real(1 ,5)"), (SyntaxKind.RealLiteralToken, "real(1,234,567.5)"),
            (SyntaxKind.RealLiteralToken, "real(1,)"), (SyntaxKind.RealLiteralToken, "real( NaN )"),
            (SyntaxKind.RealLiteralToken, "real( Infinity )"), (SyntaxKind.RealLiteralToken, "real(1e999)"),
            (SyntaxKind.RealLiteralToken, "real(0x1p2)"), (SyntaxKind.RealLiteralToken, "real(1_0)"),
            (SyntaxKind.DecimalLiteralToken, "decimal(1,5)"), (SyntaxKind.DecimalLiteralToken, "decimal(1.5)"),
            (SyntaxKind.DecimalLiteralToken, "decimal(1e2)"), (SyntaxKind.DecimalLiteralToken, "decimal(1.5-)"),
            (SyntaxKind.IntLiteralToken, "int(012)"), (SyntaxKind.IntLiteralToken, "int(0x10)"),
            (SyntaxKind.IntLiteralToken, "int(2147483648)"), (SyntaxKind.LongLiteralToken, "long(012)"),
            (SyntaxKind.LongLiteralToken, "long(0x10)"), (SyntaxKind.LongLiteralToken, "long(9223372036854775808)"),
            (SyntaxKind.TimespanLiteralToken, "timespan(1,5s)"), (SyntaxKind.TimespanLiteralToken, "timespan(1.5s)"),
            (SyntaxKind.TimespanLiteralToken, "timespan(1.5)"), (SyntaxKind.TimespanLiteralToken, "timespan(1:02:03.5)"),
            (SyntaxKind.DateTimeLiteralToken, "datetime(01/02/2026)"),
            (SyntaxKind.DateTimeLiteralToken, "datetime(31/12/2026)"),
            (SyntaxKind.DateTimeLiteralToken, "datetime(2026-10-22)")
        };
        var examples = new object[] { "Iİ", new DateTime(2026, 10, 22, 13, 14, 15),
            TimeSpan.FromTicks(937840050001), TimeSpan.FromTicks(-1), 1.25m,
            Guid.Parse("00010203-0405-0607-0809-0a0b0c0d0e0f"), 1234.5, true };
        var exampleMethod = typeof(KustoCompleter).GetMethod("CreateValueExamples", BindingFlags.Static | BindingFlags.NonPublic);
        return new {
            completionOrder = new[] { "ab", "a-b", "_a", "é", "z", "\ue000", "😀" }
                .OrderBy(x => CultureUtilities.ToLowerInvariant(x), CultureStringComparer.Ordinal).ToArray(),
            valueExamples = examples.Select((v, i) => new { index = i,
                text = (string)exampleMethod.Invoke(null, new[] { (object)new Parameter("p", ScalarTypes.String), v }) }).ToArray(),
            commit = "328e05d10d6c73654e416675fc75af7ed0e05ffd",
            comparisons = pairs.Select(p => new {
                left = p.Item1, right = p.Item2,
                equal = CultureUtilities.Equals(p.Item1, p.Item2, true),
                order = Math.Sign(CultureUtilities.Compare(p.Item1, p.Item2, true)),
                naturalOrder = Math.Sign(StringAndNumberComparer.OrdinalIgnoreCase.Compare(p.Item1 + "1", p.Item2 + "1")),
                lower = CultureUtilities.ToLowerInvariant(p.Item1)
            }).ToArray(),
            wildcards = wildcards.Select(p => new {
                pattern = p.Item1, text = p.Item2,
                sensitive = KustoFacts.Matches(p.Item1, p.Item2, false),
                insensitive = KustoFacts.Matches(p.Item1, p.Item2, true)
            }).ToArray(),
            literals = literals.Select(p => {
                var value = SyntaxToken.Literal("", p.Item2, p.Item1).Value;
                return new { kind = p.Item1.ToString(), text = p.Item2,
                    value = value is DateTime dt ? dt.ToString("yyyy-MM-dd'T'HH:mm:ss.fffffff", CultureInfo.InvariantCulture)
                        : value is TimeSpan ts ? ts.Ticks.ToString(CultureInfo.InvariantCulture)
                        : Convert.ToString(value, CultureInfo.InvariantCulture) };
            }).ToArray(),
            scanners = pairs.Select(p => new {
                pattern = p.Item1, text = p.Item2,
                result = Parsers<char>.Match(p.Item1, true).Scan(new ArraySource<char>(p.Item2.ToCharArray()), 0)
            }).ToArray(),
            textSources = new[] { (0, 1, 0, "ab"), (0, 1, 0, "A"), (1, 1, 0, "bc"), (1, 1, 0, "B"), (0, 1, 1, ""), (0, 1, 0, "") }
                .Select(p => new { offset = p.Item1, length = p.Item2, start = p.Item3, pattern = p.Item4,
                    sensitive = new TextSource("abc", p.Item1, p.Item2).Matches(p.Item3, p.Item4),
                    insensitive = new TextSource("abc", p.Item1, p.Item2).Matches(p.Item3, p.Item4, true) }).ToArray()
        };
    }
}
