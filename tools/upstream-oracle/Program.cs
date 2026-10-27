using System;
using System.IO;
using System.Linq;
using System.Text.Json;
using Kusto.Language;
using Kusto.Language.Symbols;

if (args.Length > 2)
{
    System.Globalization.CultureInfo.CurrentCulture = System.Globalization.CultureInfo.GetCultureInfo(args[2]);
}
if (args[0] == "--compatibility")
{
    File.WriteAllText(args[1], JsonSerializer.Serialize(Compatibility.Build(), new JsonSerializerOptions { WriteIndented = true }));
    return;
}
var queries = JsonSerializer.Deserialize<string[]>(File.ReadAllText(args[0]));
var results = queries.Select(query => {
    var code = KustoCode.ParseAndAnalyze(query);
    var table = code.ResultType as TableSymbol;
    return new {
        query,
        diagnostics = code.GetDiagnostics().Select(d => new { d.Code, d.Message, start = d.Start, length = d.Length }).ToArray(),
        columns = table?.Columns.OrderBy(c => c.Name, StringComparer.Ordinal).Select(c => new { c.Name, type = c.Type.Name }).ToArray(),
        isOpen = table == null ? (bool?)null : table.IsOpen,
        resultType = code.ResultType?.Name
    };
}).ToArray();
File.WriteAllText(args[1], JsonSerializer.Serialize(results));
