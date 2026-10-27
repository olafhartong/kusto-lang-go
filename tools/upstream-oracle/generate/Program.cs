using System;
using System.IO;
using Kusto.Language.Generator;
using Kusto.Language.Generators;

// Execute the same generator entry points as upstream's checked-in T4 templates.
Directory.CreateDirectory(args[0]);
File.WriteAllText(Path.Combine(args[0], "GeneratedSyntaxNodes.cs"),
    SyntaxNodeGenerator.Generate(SyntaxNodeInfos.All, SyntaxNodeInfos.KnownTypes));
foreach (var group in new[] { typeof(EngineCommandInfos), typeof(DataManagerCommandInfos),
                             typeof(ClusterManagerCommandInfos), typeof(AriaBridgeCommandInfos) })
{
    var name = group.Name.Replace("CommandInfos", "");
    File.WriteAllText(Path.Combine(args[0], name + "Commands.cs"),
        new CommandGenerator().GenerateSymbols(name + "Commands", group));
    File.WriteAllText(Path.Combine(args[0], name + "CommandGrammar.cs"),
        new CommandGenerator().GenerateParser(name + "CommandGrammar", group));
}
