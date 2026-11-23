using System;
using System.Linq;
using Kusto.Language;
using Kusto.Language.Editor;
using Kusto.Language.Symbols;
using Kusto.Language.Syntax;
using Kusto.Language.Utils;

static class November12
{
    public static object Build(string commit)
    {
        var queries = new[] {
            ".show capacity", ".show capacity async-update-policy",
            ".show capacity async-update-policy with (scope=cluster)",
            ".show capacity async-update-policy with (scope=workloadgroup)",
            ".show capacity update-policy-over-external-table", ".show capacity query-acceleration",
            ".show capacity async-update-policy with (scope=)", ".show capacity async-update-policy extra",
            ".alter database db extentcontainer with (path='https://example.test/container')",
            ".alter database ['database name'] extentcontainer with (path='new')",
            ".alter database db extentcontainer with (path='one', enabled=true, value=1)",
            ".alter database db extentcontainer with (path=dynamic({'a':1}))",
            ".alter database db extentcontainer", ".alter database db extentcontainer with",
            ".alter database db extentcontainer with (", ".alter database db extentcontainer with ()",
            ".alter database db extentcontainer with (path=)", ".alter database db extentcontainer with (path='new'",
            ".alter database db extentcontainer with (path='new',)",
            ".alter database db extentcontainer with (path='one',, value=1)",
            ".alter database extentcontainer with (path='new')",
            ".alter database db extentcontainer with (path='new') extra",
        };
        var commands = queries.Select(query => {
            var code = KustoCode.ParseAndAnalyze(query);
            var table = code.ResultType as TableSymbol;
            return new { query,
                diagnostics = code.GetDiagnostics().Select(d => new { d.Code, d.Message, start = d.Start, length = d.Length }).ToArray(),
                columns = table?.Columns.OrderBy(c => c.Name, StringComparer.Ordinal).Select(c => new { c.Name, type = c.Type.Name }).ToArray(),
                isOpen = table == null ? (bool?)null : table.IsOpen, resultType = code.ResultType?.Name };
        }).ToArray();
        var completions = new[] { ".show capacity ", ".show capacity async-update-policy ", ".alter database db ", ".alter database db extentcontainer ", ".alter database db extentcontainer with (" }
            .Select(text => new { text, items = new KustoCodeService(text).GetCompletionItems(text.Length,
                CompletionOptions.Default.WithIncludeSymbols(false).WithIncludeExtendedSyntax(false)).Items.Select(i => i.DisplayText).OrderBy(x => x, StringComparer.Ordinal).ToArray() }).ToArray();
        var strings = new StringTable();
        var operations = new[] { ("Alpha",0,5), ("__Alpha__",2,5), ("__Beta__",2,4), ("",0,0), ("😀β",2,1), ("--β--",2,1) };
        var ranges = operations.Select(o => new { text = o.Item1, start = o.Item2, length = o.Item3, value = strings.Add(o.Item1,o.Item2,o.Item3) }).ToArray();
        var command = EngineCommands.AlterDatabaseExtentContainerPath;
        return new { commit, commands, completions, ranges,
            commandCount = EngineCommands.All.Count,
            construction = command.Construction, schema = command.ResultSchema,
            repeatedRangeSameInstance = ReferenceEquals(strings.Add("Alpha"), strings.Add("__Alpha__",2,5)) };
    }
}
