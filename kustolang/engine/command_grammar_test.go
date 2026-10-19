package kustolang

import (
	"strings"
	"testing"
)

func TestCommandGrammarParsesUnknownCommandBlock(t *testing.T) {
	block := ParseCommandBlock(".not_a_real_command version")
	if block == nil || block.Kind() != SyntaxKindCommandBlock || block.Statements().Count() != 1 {
		t.Fatalf("command block = %#v", block)
	}
	command := firstCommandExpression(t, block).(*UnknownCommand)
	if command.DotToken().Kind() != SyntaxKindDotToken || command.Parts().Count() != 2 {
		t.Fatalf("unknown command = dot:%#v parts:%d", command.DotToken(), command.Parts().Count())
	}
	if command.Parts().At(0).(*SyntaxToken).RawText() != "not_a_real_command" || command.Parts().At(1).(*SyntaxToken).RawText() != "version" {
		t.Fatalf("unknown command parts = %#v", command.Parts())
	}
}

func TestCommandGrammarParsesGeneratedEngineSimpleCommands(t *testing.T) {
	if got := len(engineCommandParserInfos()); got != 637 {
		t.Fatalf("engine simple command parser count = %d", got)
	}

	for text, kind := range map[string]string{
		".show version":          "ShowVersion",
		".show queries":          "ShowQueries",
		".show clockdiagnostics": "ShowClockDiagnostics",
		`.show clockdiagnostics with (scope="cluster", attempts=3)`:                                        "ShowClockDiagnostics",
		".show commands-and-queries":                                                                       "ShowCommandsAndQueries",
		".clear cluster cache credstore":                                                                   "ClearClusterCredStoreCache",
		".show cluster details":                                                                            "ShowClusterDetails",
		".enable cluster maintenance_mode":                                                                 "SetClusterMaintenanceMode",
		".disable cluster maintenance_mode":                                                                "SetClusterMaintenanceMode",
		".drop cluster metadata storage artifacts cleanup state":                                           "ClusterDropStorageArtifactsCleanupState",
		".show materialized-view MV kqlschema":                                                             "ShowMaterializedViewCslSchema",
		".show materialized-view MV cslschema":                                                             "ShowMaterializedViewCslSchema",
		".show external table Ext policy row_level_security":                                               "ShowExternalTablePolicyRowLevelSecurity",
		".show external table * policy row_level_security":                                                 "ShowExternalTableStarPolicyRowLevelSecurity",
		`.alter external table Ext policy row_level_security enable with (audit=true) "T | where Allowed"`: "AlterExternalTablePolicyRowLevelSecurity",
		".delete external table Ext policy row_level_security":                                             "DeleteExternalTablePolicyRowLevelSecurity",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
	}

	malformedExternalRLS := ParseCommandBlock(`.alter external table Ext policy row_level_security enable`)
	if !diagnosticsContain(malformedExternalRLS.ContainedDiagnostics(), "KS005") {
		t.Fatalf("malformed external-table RLS diagnostics = %#v", malformedExternalRLS.ContainedDiagnostics())
	}

	malformedClockDiagnostics := ParseCommandBlock(`.show clockdiagnostics with (scope=)`)
	if !diagnosticsContain(malformedClockDiagnostics.ContainedDiagnostics(), "KS005") {
		t.Fatalf("malformed clockdiagnostics diagnostics = %#v", malformedClockDiagnostics.ContainedDiagnostics())
	}
}

func TestCommandGrammarDatabaseRowIDPolicy(t *testing.T) {
	for _, test := range []struct {
		verb   string
		tail   string
		symbol *CommandSymbol
	}{
		{"show", "", EngineCommands.ShowDatabasePolicyRowId},
		{"alter", " true", EngineCommands.AlterDatabasePolicyRowId},
		{"alter", " false", EngineCommands.AlterDatabasePolicyRowId},
		{"delete", "", EngineCommands.DeleteDatabasePolicyRowId},
	} {
		names := []string{"", " MyDb", " ['policy']"}
		if test.verb == "show" {
			names = append(names, " *")
		}
		for _, name := range names {
			query := "." + test.verb + " database" + name + " policy rowid" + test.tail
			t.Run(query, func(t *testing.T) {
				block := ParseCommandBlock(query)
				if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
					t.Fatalf("diagnostics = %#v", diagnostics)
				}
				command, ok := firstCommandExpression(t, block).(*CustomCommand)
				if !ok || command.CommandKind() != test.symbol.SymbolName() || command.ReferencedSymbol() != test.symbol {
					t.Fatalf("command = %#v", command)
				}
				custom := command.Custom().(*CustomNode)
				if !custom.IsOptional(2) || (custom.GetChild(2) == nil) != (name == "") {
					t.Fatalf("optional database = %#v", custom.GetChild(2))
				}
			})
		}
	}
	for _, query := range []string{
		".alter database policy rowid",
		".alter database MyDb policy rowid",
		".alter database policy rowid 1",
		".alter database policy rowid enable",
		".alter database * policy rowid true",
		".delete database * policy rowid",
	} {
		t.Run(query, func(t *testing.T) {
			if diagnostics := ParseCommandBlock(query).ContainedDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("expected diagnostics for invalid row-ID policy command")
			}
		})
	}
}

func TestCommandGrammarRowStoreNamesDoNotShadowSiblingCommands(t *testing.T) {
	tests := []struct {
		text   string
		kind   string
		symbol *CommandSymbol
	}{
		{text: `.show rowstore StoreA`, kind: "ShowRowStore", symbol: EngineCommands.ShowRowStore},
		{text: `.show rowstore transactions`, kind: "ShowRowStoreTransactions", symbol: EngineCommands.ShowRowStoreTransactions},
		{text: `.show rowstore seals "T"`, kind: "ShowRowStoreSeals", symbol: EngineCommands.ShowRowStoreSeals},
	}
	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			block := ParseCommandBlock(test.text)
			command, ok := firstCommandExpression(t, block).(*CustomCommand)
			if !ok {
				t.Fatalf("%s command = %#v", test.text, firstCommandExpression(t, block))
			}
			if command.CommandKind() != test.kind || command.ReferencedSymbol() != test.symbol {
				t.Fatalf("%s command = %#v kind:%q symbol:%#v", test.text, firstCommandExpression(t, block), command.CommandKind(), command.ReferencedSymbol())
			}
			if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("%s diagnostics = %#v", test.text, diagnostics)
			}
		})
	}

	missingSealsTable := ParseCommandBlock(`.show rowstore seals`)
	if command, ok := firstCommandExpression(t, missingSealsTable).(*CustomCommand); ok && command.CommandKind() == "ShowRowStore" {
		t.Fatalf("missing seals table fell back to ShowRowStore: %#v", command)
	}
	if !diagnosticsContain(missingSealsTable.ContainedDiagnostics(), "KS005") {
		t.Fatalf("missing seals table diagnostics = %#v", missingSealsTable.ContainedDiagnostics())
	}
}

func TestCommandGrammarParsesOptionalDatabasePolicyCommandNames(t *testing.T) {
	for text, kind := range map[string]string{
		".show database policy streamingingestion":      "ShowDatabasePolicyStreamingIngestion",
		".show database MyDb policy streamingingestion": "ShowDatabasePolicyStreamingIngestion",
		".show database policy managed_identity":        "ShowDatabasePolicyManagedIdentity",
		".show database MyDb policy managed_identity":   "ShowDatabasePolicyManagedIdentity",
		".show database policy diagnostics":             "ShowDatabasePolicyDiagnostics",
		".show database MyDb policy diagnostics":        "ShowDatabasePolicyDiagnostics",
	} {
		block := ParseCommandBlock(text)
		if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("%s diagnostics = %#v", text, diagnostics)
		}
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
	}
}

func TestCommandGrammarGeneratedFunctionDeclarationCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.create function F(x:string) { x }`:                             "CreateFunction",
		`.create function with (folder="Ops") F() { Fallback }`:          "CreateFunction",
		`.alter function F(x:string) { x }`:                              "AlterFunction",
		`.create-or-alter function with (folder="Ops") F() { Fallback }`: "CreateOrAlterFunction",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	create := firstCommandExpression(t, ParseCommandBlock(`.create function F(x:string) { x }`)).(*CustomCommand)
	createCustom := create.Custom().(*CustomNode)
	declaration, ok := createCustom.GetChild(5).(*FunctionDeclaration)
	if !ok {
		t.Fatalf("create function declaration = %#v", createCustom.GetChild(5))
	}
	if declaration.Parameters().Parameters().Count() != 1 || declaration.Body().Expression() == nil {
		t.Fatalf("create function declaration parameters=%d body=%#v", declaration.Parameters().Parameters().Count(), declaration.Body())
	}

	createOrAlter := firstCommandExpression(t, ParseCommandBlock(`.create-or-alter function with (folder="Ops") F() { Fallback }`)).(*CustomCommand)
	createOrAlterCustom := createOrAlter.Custom().(*CustomNode)
	optionsNode, ok := createOrAlterCustom.GetChild(2).(*CustomNode)
	if !ok {
		t.Fatalf("create-or-alter function options node = %#v", createOrAlterCustom.GetChild(2))
	}
	options, ok := optionsNode.GetChild(2).(*SyntaxList)
	if !ok || options.Count() != 1 {
		t.Fatalf("create-or-alter function options = %#v", optionsNode.GetChild(2))
	}
	declaration, ok = createOrAlterCustom.GetChild(4).(*FunctionDeclaration)
	if !ok {
		t.Fatalf("create-or-alter function declaration = %#v", createOrAlterCustom.GetChild(4))
	}
	if declaration.Parameters().Parameters().Count() != 0 || declaration.Body().Expression() == nil {
		t.Fatalf("create-or-alter function declaration parameters=%d body=%#v", declaration.Parameters().Parameters().Count(), declaration.Body())
	}
}

func TestCommandGrammarRejectsOptionalFunctionTermsAsFunctionName(t *testing.T) {
	block := ParseCommandBlock(`.create function with () {}`)
	if _, ok := firstCommandExpression(t, block).(*CustomCommand); ok {
		t.Fatalf("malformed create function resolved as a complete command: %#v", firstCommandExpression(t, block))
	}
	if len(block.ContainedDiagnostics()) == 0 {
		t.Fatal("malformed create function has no syntax diagnostics")
	}
}

func TestCommandGrammarGeneratedFunctionBodyCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.create materialized-view MV on table T { T | count }`:                            "CreateMaterializedView",
		`.alter materialized-view MV on table T { T | summarize Count=count() }`:           "AlterMaterializedView",
		`.create-or-alter materialized-view MV on materialized-view SourceMV { SourceMV }`: "CreateOrAlterMaterializedViewOverMaterializedView",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	create := firstCommandExpression(t, ParseCommandBlock(`.create materialized-view MV on table T { T | count }`)).(*CustomCommand)
	createCustom := create.Custom().(*CustomNode)
	body, ok := createCustom.GetChild(9).(*FunctionBody)
	if !ok {
		t.Fatalf("create materialized-view body = %#v", createCustom.GetChild(9))
	}
	if body.Expression() == nil || body.CloseBrace() == nil {
		t.Fatalf("create materialized-view body expression=%#v close=%#v", body.Expression(), body.CloseBrace())
	}

	createOrAlter := firstCommandExpression(t, ParseCommandBlock(`.create-or-alter materialized-view MV on materialized-view SourceMV { SourceMV }`)).(*CustomCommand)
	createOrAlterCustom := createOrAlter.Custom().(*CustomNode)
	body, ok = createOrAlterCustom.GetChild(7).(*FunctionBody)
	if !ok {
		t.Fatalf("create-or-alter materialized-view body = %#v", createOrAlterCustom.GetChild(7))
	}
	if body.Expression() == nil || body.CloseBrace() == nil {
		t.Fatalf("create-or-alter materialized-view body expression=%#v close=%#v", body.Expression(), body.CloseBrace())
	}
}

func TestCommandGrammarGeneratedInputTextCommands(t *testing.T) {
	for text := range map[string]struct{}{
		".ingest inline into table T [a,b\nc,d]":                 {},
		".ingest inline into table T <| a,b\nc,d":                {},
		`.ingest inline into table T with (format="csv") <| a,b`: {},
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != "IngestInlineIntoTable" {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	bracketed := firstCommandExpression(t, ParseCommandBlock(".ingest inline into table T [a,b\nc,d]")).(*CustomCommand)
	bracketedCustom := bracketed.Custom().(*CustomNode)
	bracketedBranch, ok := bracketedCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("bracketed inline branch = %#v", bracketedCustom.GetChild(5))
	}
	bracketedInput, ok := bracketedBranch.GetChild(1).(*SyntaxToken)
	if !ok || bracketedInput.Kind() != SyntaxKindInputTextToken || bracketedInput.RawText() != "a,b\nc,d" {
		t.Fatalf("bracketed input text = %#v", bracketedBranch.GetChild(1))
	}

	recovered := ParseCommandBlock(".ingest inline into table T [a,b")
	if !diagnosticsContainMessage(recovered.ContainedDiagnostics(), "KS005", "Expected: ]") {
		t.Fatalf("missing bracketed inline close diagnostic = %#v", recovered.ContainedDiagnostics())
	}
	recoveredCommand, ok := firstCommandExpression(t, recovered).(*CustomCommand)
	if !ok || recoveredCommand.CommandKind() != "IngestInlineIntoTable" {
		t.Fatalf("recovered bracketed inline command = %#v", firstCommandExpression(t, recovered))
	}
	recoveredCustom := recoveredCommand.Custom().(*CustomNode)
	recoveredBranch, ok := recoveredCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("recovered bracketed inline branch = %#v", recoveredCustom.GetChild(5))
	}
	recoveredInput, ok := recoveredBranch.GetChild(1).(*SyntaxToken)
	if !ok || recoveredInput.Kind() != SyntaxKindInputTextToken || recoveredInput.RawText() != "a,b" {
		t.Fatalf("recovered bracketed input text = %#v", recoveredBranch.GetChild(1))
	}

	plain := firstCommandExpression(t, ParseCommandBlock(".ingest inline into table T <| a,b\nc,d")).(*CustomCommand)
	plainCustom := plain.Custom().(*CustomNode)
	plainBranch, ok := plainCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("plain inline branch = %#v", plainCustom.GetChild(5))
	}
	plainInput, ok := plainBranch.GetChild(1).(*SyntaxToken)
	if !ok || plainInput.Kind() != SyntaxKindInputTextToken || plainInput.RawText() != "a,b\nc,d" {
		t.Fatalf("plain input text = %#v", plainBranch.GetChild(1))
	}

	withOptions := firstCommandExpression(t, ParseCommandBlock(`.ingest inline into table T with (format="csv") <| a,b`)).(*CustomCommand)
	withOptionsCustom := withOptions.Custom().(*CustomNode)
	withOptionsBranch, ok := withOptionsCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("with-options inline branch = %#v", withOptionsCustom.GetChild(5))
	}
	options, ok := withOptionsBranch.GetChild(2).(*SyntaxList)
	if !ok || options.Count() != 1 {
		t.Fatalf("with-options list = %#v", withOptionsBranch.GetChild(2))
	}
	withOptionsInput, ok := withOptionsBranch.GetChild(5).(*SyntaxToken)
	if !ok || withOptionsInput.Kind() != SyntaxKindInputTextToken || withOptionsInput.RawText() != "a,b" {
		t.Fatalf("with-options input text = %#v", withOptionsBranch.GetChild(5))
	}
}

func TestCommandGrammarGeneratedDescriptorOnlyCustomListCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.export async compressed to csv ("https://storage/a", "https://storage/b") with (sizeLimit=1000) <| T | take 1`: "ExportToStorage",
		`.alter follower database Db from "Leader" external tables exclude drop (Ext*, Logs*)`:                           "AlterFollowerDatabaseChildEntities",
		`.delete policies of retention ("database('Db').table('T')", "database('Db').materialized-view('MV')")`:          "DeletePoliciesOfRetention",
		`.clear cluster cache external-artifacts ("https://storage/a", "https://storage/b")`:                             "ClearExternalArtifactsCache",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	export := firstCommandExpression(t, ParseCommandBlock(`.export async compressed to csv ("https://storage/a", "https://storage/b") with (sizeLimit=1000) <| T | take 1`)).(*CustomCommand)
	exportCustom := export.Custom().(*CustomNode)
	exportConnections, ok := exportCustom.GetChild(6).(*SyntaxList)
	if !ok || exportConnections.Count() != 2 {
		t.Fatalf("export connection list = %#v", exportCustom.GetChild(6))
	}
	firstConnection, ok := exportConnections.At(0).(*SeparatedElement).Element().(*CustomNode)
	if !ok {
		t.Fatalf("export first connection = %#v", exportConnections.At(0))
	}
	if firstConnection.GetName(0) != "DataConnectionString" || firstConnection.CompletionHint(0) != CompletionHintLiteral {
		t.Fatalf("export first connection descriptor name=%q hint=%v", firstConnection.GetName(0), firstConnection.CompletionHint(0))
	}
	exportOptionsNode, ok := exportCustom.GetChild(8).(*CustomNode)
	if !ok {
		t.Fatalf("export option node = %#v", exportCustom.GetChild(8))
	}
	exportOptions, ok := exportOptionsNode.GetChild(2).(*SyntaxList)
	if !ok || exportOptions.Count() != 1 {
		t.Fatalf("export option list = %#v", exportOptionsNode.GetChild(2))
	}
	exportQuery, ok := exportCustom.GetChild(10).(*SyntaxList)
	if !ok || exportQuery.Count() == 0 {
		t.Fatalf("export query input = %#v", exportCustom.GetChild(10))
	}

	follower := firstCommandExpression(t, ParseCommandBlock(`.alter follower database Db from "Leader" external tables exclude drop (Ext*, Logs*)`)).(*CustomCommand)
	followerCustom := follower.Custom().(*CustomNode)
	leader, ok := followerCustom.GetChild(4).(*CustomNode)
	if !ok || leader.GetName(1) != "leaderClusterMetadataPath" || leader.CompletionHint(1) != CompletionHintLiteral {
		t.Fatalf("follower leader descriptor node=%#v", followerCustom.GetChild(4))
	}
	entityKind, ok := followerCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("follower entity kind branch = %#v", followerCustom.GetChild(5))
	}
	if firstToken, ok := entityKind.GetChild(0).(*SyntaxToken); !ok || firstToken.RawText() != "external" {
		t.Fatalf("follower entity kind first token = %#v", entityKind.GetChild(0))
	}
	followerEntities, ok := followerCustom.GetChild(9).(*SyntaxList)
	if !ok || followerEntities.Count() != 2 {
		t.Fatalf("follower entity list = %#v", followerCustom.GetChild(9))
	}
	firstEntity, ok := followerEntities.At(0).(*SeparatedElement).Element().(*CustomNode)
	if !ok {
		t.Fatalf("follower first entity = %#v", followerEntities.At(0))
	}
	if firstEntity.GetName(0) != "ename" || firstEntity.CompletionHint(0) != CompletionHintNone {
		t.Fatalf("follower first entity descriptor name=%q hint=%v", firstEntity.GetName(0), firstEntity.CompletionHint(0))
	}

	policies := firstCommandExpression(t, ParseCommandBlock(`.delete policies of retention ("database('Db').table('T')", "database('Db').materialized-view('MV')")`)).(*CustomCommand)
	policiesCustom := policies.Custom().(*CustomNode)
	policyList, ok := policiesCustom.GetChild(5).(*SyntaxList)
	if !ok || policyList.Count() != 2 {
		t.Fatalf("policy retention list = %#v", policiesCustom.GetChild(5))
	}

	clear := firstCommandExpression(t, ParseCommandBlock(`.clear cluster cache external-artifacts ("https://storage/a", "https://storage/b")`)).(*CustomCommand)
	clearCustom := clear.Custom().(*CustomNode)
	clearNested, ok := clearCustom.GetChild(4).(*CustomNode)
	if !ok {
		t.Fatalf("clear external-artifacts nested = %#v", clearCustom.GetChild(4))
	}
	clearList, ok := clearNested.GetChild(1).(*SyntaxList)
	if !ok || clearList.Count() != 2 {
		t.Fatalf("clear external-artifacts list = %#v", clearNested.GetChild(1))
	}
}

func TestCommandGrammarGeneratedScriptInputCommands(t *testing.T) {
	for text, kind := range map[string]string{
		".execute database script <| .show version; .show queries":                             "ExecuteDatabaseScript",
		".execute cluster script with (ContinueOnErrors=true) <| .show version\n.show queries": "ExecuteClusterScript",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	database := firstCommandExpression(t, ParseCommandBlock(".execute database script <| .show version; .show queries")).(*CustomCommand)
	databaseCustom := database.Custom().(*CustomNode)
	databaseScript, ok := databaseCustom.GetChild(5).(*SyntaxList)
	if !ok {
		t.Fatalf("database script input = %#v", databaseCustom.GetChild(5))
	}
	if databaseScript.Count() != 2 {
		t.Fatalf("database script input count = %d", databaseScript.Count())
	}
	first := databaseScript.At(0).(*SeparatedElement)
	if first.Separator() == nil || first.Separator().Kind() != SyntaxKindSemicolonToken {
		t.Fatalf("database script separator = %#v", first.Separator())
	}
	firstCommand, ok := first.Element().(*CustomCommand)
	if !ok || firstCommand.CommandKind() != "ShowVersion" {
		t.Fatalf("database script first command = %#v", first.Element())
	}
	second := databaseScript.At(1).(*SeparatedElement)
	secondCommand, ok := second.Element().(*CustomCommand)
	if !ok || secondCommand.CommandKind() != "ShowQueries" {
		t.Fatalf("database script second command = %#v", second.Element())
	}

	cluster := firstCommandExpression(t, ParseCommandBlock(".execute cluster script with (ContinueOnErrors=true) <| .show version\n.show queries")).(*CustomCommand)
	clusterCustom := cluster.Custom().(*CustomNode)
	optionsNode, ok := clusterCustom.GetChild(3).(*CustomNode)
	if !ok {
		t.Fatalf("cluster script options node = %#v", clusterCustom.GetChild(3))
	}
	options, ok := optionsNode.GetChild(2).(*SyntaxList)
	if !ok || options.Count() != 1 {
		t.Fatalf("cluster script options = %#v", optionsNode.GetChild(2))
	}
	clusterScript, ok := clusterCustom.GetChild(5).(*SyntaxList)
	if !ok {
		t.Fatalf("cluster script input = %#v", clusterCustom.GetChild(5))
	}
	if clusterScript.Count() != 2 {
		t.Fatalf("cluster script input count = %d", clusterScript.Count())
	}
	if first := clusterScript.At(0).(*SeparatedElement); first.Separator() != nil {
		t.Fatalf("cluster newline script separator = %#v", first.Separator())
	}

	recoveredOptions := ParseCommandBlock(".execute database script with (ContinueOnErrors=true,) <| .show table T details")
	assertSingleCommandDiagnostic(t, recoveredOptions, "KS005", "Expected one of: <name>, ContinueOnErrors, ThrowOnErrors")
	if _, ok := firstCommandExpression(t, recoveredOptions).(*CustomCommand); ok {
		t.Fatalf("script command trailing option comma should not bind a complete command: %#v", firstCommandExpression(t, recoveredOptions))
	}
}

func TestCommandGrammarGeneratedQueryInputCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.set-or-append T <| print x=1`:       "SetOrAppendTable",
		`.append T <| T | take 1`:             "AppendTable",
		`.show queryexecution <| print x = 1`: "ShowQueryExecution",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	set := firstCommandExpression(t, ParseCommandBlock(`.set-or-append T <| print x=1`)).(*CustomCommand)
	setCustom := set.Custom().(*CustomNode)
	queryInput, ok := setCustom.GetChild(5).(*SyntaxList)
	if !ok {
		t.Fatalf("set-or-append query input = %#v", setCustom.GetChild(5))
	}
	if queryInput.Count() == 0 {
		t.Fatalf("set-or-append query input count = %d", queryInput.Count())
	}

	trailingCommaBlock := ParseCommandBlock(`.set-or-append T with (folder="Ops",) <| print x=1`)
	assertSingleCommandDiagnostic(t, trailingCommaBlock, "KS005", "Expected one of: <name>, creationTime, distributed, docstring, extend_schema, folder, format, ignoreFirstRecord, ingestIfNotExists, ingestionMapping, ingestionMappingReference, persistDetails, policy_ingestionTime, recreate_schema, small_dimension_table, tags, validationPolicy, zipPattern")
	if _, ok := firstCommandExpression(t, trailingCommaBlock).(*CustomCommand); ok {
		t.Fatalf("set-or-append trailing option comma should not bind a complete command: %#v", firstCommandExpression(t, trailingCommaBlock))
	}

	doubleCommaBlock := ParseCommandBlock(`.set-or-append T with (folder="Ops",, format="csv") <| print x=1`)
	assertSingleCommandDiagnostic(t, doubleCommaBlock, "KS005", "Expected one of: <name>, creationTime, distributed, docstring, extend_schema, folder, format, ignoreFirstRecord, ingestIfNotExists, ingestionMapping, ingestionMappingReference, persistDetails, policy_ingestionTime, recreate_schema, small_dimension_table, tags, validationPolicy, zipPattern")
	if _, ok := firstCommandExpression(t, doubleCommaBlock).(*CustomCommand); ok {
		t.Fatalf("set-or-append double option comma should not bind a complete command: %#v", firstCommandExpression(t, doubleCommaBlock))
	}

	missingCommaBlock := ParseCommandBlock(`.set-or-append T with (folder="Ops" format="csv") <| print x=1`)
	assertSingleCommandDiagnostic(t, missingCommaBlock, "KS005", "Expected: ')'")
	if _, ok := firstCommandExpression(t, missingCommaBlock).(*CustomCommand); ok {
		t.Fatalf("set-or-append missing option comma should not bind a complete command: %#v", firstCommandExpression(t, missingCommaBlock))
	}
}

func TestCommandGrammarGeneratedExpandedRootStepShape(t *testing.T) {
	command, ok := firstCommandExpression(t, ParseCommandBlock(`.delete table T records <| T | take 1`)).(*CustomCommand)
	if !ok || command.CommandKind() != "DeleteTableRecords" {
		t.Fatalf("delete records command = %#v", command)
	}
	custom := command.Custom().(*CustomNode)
	if custom.ChildCount() != 7 {
		t.Fatalf("root child count = %d, want 7", custom.ChildCount())
	}
	if custom.GetChild(5) != nil || !custom.IsOptional(5) {
		t.Fatalf("optional property-list child = %#v optional=%t", custom.GetChild(5), custom.IsOptional(5))
	}
	if got := custom.GetName(6); got != "csl" {
		t.Fatalf("csl child name = %q", got)
	}
	csl, ok := custom.GetChild(6).(*CustomNode)
	if !ok || csl.ChildCount() != 2 {
		t.Fatalf("csl child = %#v", custom.GetChild(6))
	}
	if token, ok := csl.GetChild(0).(*SyntaxToken); !ok || token.RawText() != "<|" {
		t.Fatalf("csl pipe = %#v", csl.GetChild(0))
	}
	query, ok := csl.GetChild(1).(*SyntaxList)
	if !ok || query.Count() == 0 {
		t.Fatalf("csl query = %#v", csl.GetChild(1))
	}

	withProperties := firstCommandExpression(t, ParseCommandBlock(`.delete table T records with (folder="Ops") <| T | take 1`)).(*CustomCommand).Custom().(*CustomNode)
	properties, ok := withProperties.GetChild(5).(*CustomNode)
	if !ok || properties.ChildCount() != 4 || !withProperties.IsOptional(5) {
		t.Fatalf("present property-list child = %#v optional=%t", withProperties.GetChild(5), withProperties.IsOptional(5))
	}

	missingOpen := ParseCommandBlock(`.delete table T records with`)
	assertSingleCommandDiagnostic(t, missingOpen, "KS005", "Expected: '('")
	if _, ok := firstCommandExpression(t, missingOpen).(*CustomCommand); ok {
		t.Fatalf("missing property-list open resolved as complete command: %#v", firstCommandExpression(t, missingOpen))
	}
	trailingComma := ParseCommandBlock(`.delete table T records with (folder="Ops",) <| T | take 1`)
	assertSingleCommandDiagnostic(t, trailingComma, "KS005", "Expected: <name>")
	if _, ok := firstCommandExpression(t, trailingComma).(*CustomCommand); ok {
		t.Fatalf("trailing property comma resolved as complete command: %#v", firstCommandExpression(t, trailingComma))
	}
}

func TestCommandGrammarQueryInputCommaTailLeavesOuterPipe(t *testing.T) {
	block := ParseCommandBlock(`.set-or-append Target <| Source, Missing | project x`)
	expression := firstCommandExpression(t, block)
	pipe, ok := expression.(*PipeExpression)
	if !ok {
		t.Fatalf("command expression = %T, want outer PipeExpression", expression)
	}
	if _, ok := pipe.Operator().(*ProjectOperator); !ok {
		t.Fatalf("outer operator = %T, want ProjectOperator", pipe.Operator())
	}
	if _, ok := pipe.Expression().(Command); !ok {
		t.Fatalf("pipe source = %T, want recovered command", pipe.Expression())
	}
	foundSource := false
	foundMissing := false
	WalkSyntax(pipe.Expression(), func(element SyntaxElement) {
		reference, ok := WrapGeneratedSyntaxElement(element).(*NameReference)
		if !ok {
			return
		}
		switch NameReferenceSimpleName(reference) {
		case "Source":
			foundSource = true
		case "Missing":
			foundMissing = true
		}
	})
	if !foundSource || foundMissing {
		t.Fatalf("query-input references: Source=%t Missing=%t", foundSource, foundMissing)
	}
	if diagnosticsContain(block.ContainedDiagnostics(), "KS198") {
		t.Fatalf("query-input comma tail should not collapse to KS198: %#v", block.ContainedDiagnostics())
	}
}

func TestCommandGrammarGeneratedZeroOrMoreListQueryInputCommands(t *testing.T) {
	for text := range map[string]struct{}{
		`.attach extents by metadata <| print ExtentId="abc"`:                               {},
		`.attach async extents into table T1 into table T2 by metadata <| print ExtentId=1`: {},
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != "AttachExtentsIntoTableByMetadata" {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	emptyCommand := firstCommandExpression(t, ParseCommandBlock(`.attach extents by metadata <| print ExtentId="abc"`)).(*CustomCommand)
	emptyCustom := emptyCommand.Custom().(*CustomNode)
	emptyList, ok := emptyCustom.GetChild(3).(*SyntaxList)
	if !ok {
		t.Fatalf("empty metadata target list = %#v", emptyCustom.GetChild(3))
	}
	if emptyList.Count() != 0 {
		t.Fatalf("empty metadata target list count = %d", emptyList.Count())
	}
	emptyCsl, ok := emptyCustom.GetChild(6).(*CustomNode)
	if !ok || emptyCustom.GetName(6) != "csl" {
		t.Fatalf("empty metadata csl = %#v name=%q", emptyCustom.GetChild(6), emptyCustom.GetName(6))
	}
	emptyQuery, ok := emptyCsl.GetChild(1).(*SyntaxList)
	if !ok || emptyQuery.Count() == 0 {
		t.Fatalf("empty metadata query input = %#v", emptyCsl.GetChild(1))
	}

	populatedCommand := firstCommandExpression(t, ParseCommandBlock(`.attach async extents into table T1 into table T2 by metadata <| print ExtentId=1`)).(*CustomCommand)
	populatedCustom := populatedCommand.Custom().(*CustomNode)
	populatedList, ok := populatedCustom.GetChild(3).(*SyntaxList)
	if !ok {
		t.Fatalf("populated metadata target list = %#v", populatedCustom.GetChild(3))
	}
	if populatedList.Count() != 2 {
		t.Fatalf("populated metadata target list count = %d", populatedList.Count())
	}
	for i := 0; i < populatedList.Count(); i++ {
		if _, ok := populatedList.At(i).(*CustomNode); !ok {
			t.Fatalf("metadata target item %d = %#v", i, populatedList.At(i))
		}
	}
}

func TestCommandGrammarGeneratedOneOrMoreListCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.attach extents into table T by container "https://storage/container" "extent-a" "extent-b"`: "AttachExtentsIntoTableByContainer",
		`.drop storedqueryresultcontainers MyDb "container-a" "container-b"`:                          "DropStoredQueryResultContainers",
		`.show cluster extents where tags has "hot" and tags contains "prod"`:                         "ShowClusterExtents",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	attach := firstCommandExpression(t, ParseCommandBlock(`.attach extents into table T by container "https://storage/container" "extent-a" "extent-b"`)).(*CustomCommand)
	attachCustom := attach.Custom().(*CustomNode)
	extentList, ok := attachCustom.GetChild(8).(*SyntaxList)
	if !ok {
		t.Fatalf("attach extent list = %#v", attachCustom.GetChild(8))
	}
	if extentList.Count() != 2 {
		t.Fatalf("attach extent list count = %d", extentList.Count())
	}
	for i := 0; i < extentList.Count(); i++ {
		if _, ok := extentList.At(i).(*CustomNode); !ok {
			t.Fatalf("attach extent item %d = %#v", i, extentList.At(i))
		}
	}

	show := firstCommandExpression(t, ParseCommandBlock(`.show cluster extents where tags has "hot" and tags contains "prod"`)).(*CustomCommand)
	showCustom := show.Custom().(*CustomNode)
	whereNode, ok := showCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("show extents where node = %#v", showCustom.GetChild(5))
	}
	tagList, ok := whereNode.GetChild(1).(*SyntaxList)
	if !ok {
		t.Fatalf("show extents tag list = %#v", whereNode.GetChild(1))
	}
	if tagList.Count() != 2 {
		t.Fatalf("show extents tag list count = %d", tagList.Count())
	}
	firstTag := tagList.At(0).(*SeparatedElement)
	if firstTag.Separator() == nil || firstTag.Separator().RawText() != "and" {
		t.Fatalf("show extents tag separator = %#v", firstTag.Separator())
	}
	firstTagNode, ok := firstTag.Element().(*CustomNode)
	if !ok {
		t.Fatalf("show extents first tag = %#v", firstTag.Element())
	}
	if token, ok := firstTagNode.GetChild(1).(*SyntaxToken); !ok || token.RawText() != "has" {
		t.Fatalf("show extents first tag operator = %#v", firstTagNode.GetChild(1))
	}
}

func TestCommandGrammarGeneratedZeroOrMoreCommaListCommands(t *testing.T) {
	for text := range map[string]struct{}{
		`.alter table T column statistics`:                         {},
		`.alter table T column statistics C1 "{}", C2 "{\"a\":1}"`: {},
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != "AlterTableColumnStatistics" {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	emptyCommand := firstCommandExpression(t, ParseCommandBlock(".alter table T column statistics")).(*CustomCommand)
	emptyCustom := emptyCommand.Custom().(*CustomNode)
	emptyList, ok := emptyCustom.GetChild(5).(*SyntaxList)
	if !ok {
		t.Fatalf("empty statistics list = %#v", emptyCustom.GetChild(5))
	}
	if emptyList.Count() != 0 {
		t.Fatalf("empty statistics list count = %d", emptyList.Count())
	}

	populatedCommand := firstCommandExpression(t, ParseCommandBlock(`.alter table T column statistics C1 "{}", C2 "{\"a\":1}"`)).(*CustomCommand)
	populatedCustom := populatedCommand.Custom().(*CustomNode)
	populatedList, ok := populatedCustom.GetChild(5).(*SyntaxList)
	if !ok {
		t.Fatalf("populated statistics list = %#v", populatedCustom.GetChild(5))
	}
	if populatedList.Count() != 2 {
		t.Fatalf("populated statistics list count = %d", populatedList.Count())
	}
	first := populatedList.At(0).(*SeparatedElement)
	if first.Separator() == nil || first.Separator().RawText() != "," {
		t.Fatalf("statistics first separator = %#v", first.Separator())
	}
	firstStat, ok := first.Element().(*CustomNode)
	if !ok {
		t.Fatalf("statistics first element = %#v", first.Element())
	}
	if _, ok := firstStat.GetChild(0).(*NameDeclaration); !ok {
		t.Fatalf("statistics first column = %#v", firstStat.GetChild(0))
	}
	if literal, ok := firstStat.GetChild(1).(*LiteralExpression); !ok || literal.Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("statistics first payload = %#v", firstStat.GetChild(1))
	}
}

func TestCommandGrammarGeneratedCommaListCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.create tables T1 (C1:string, C2:long), T2 (D1:string)`:                 "CreateTables",
		`.alter table T policy roworder (C1 asc, C2 desc)`:                       "AlterTablePolicyRowOrder",
		`.add cluster admins ("aaduser=user@example.com", "aaduser=other@test")`: "AddClusterRole",
		`.show databases (Db1, Db2) schema`:                                      "ShowDatabasesSchema",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	createTables := firstCommandExpression(t, ParseCommandBlock(".create tables T1 (C1:string, C2:long), T2 (D1:string)")).(*CustomCommand)
	createCustom := createTables.Custom().(*CustomNode)
	tableList, ok := createCustom.GetChild(2).(*SyntaxList)
	if !ok {
		t.Fatalf("create tables list child = %#v", createCustom.GetChild(2))
	}
	if tableList.Count() != 2 {
		t.Fatalf("create tables list count = %d", tableList.Count())
	}
	firstTable := tableList.At(0).(*SeparatedElement)
	if firstTable.Separator() == nil || firstTable.Separator().RawText() != "," {
		t.Fatalf("create tables first separator = %#v", firstTable.Separator())
	}
	firstTableNode, ok := firstTable.Element().(*CustomNode)
	if !ok {
		t.Fatalf("create tables first element = %#v", firstTable.Element())
	}
	columns, ok := firstTableNode.GetChild(2).(*SyntaxList)
	if !ok || columns.Count() != 2 {
		count := -1
		if columns != nil {
			count = columns.Count()
		}
		t.Fatalf("create tables columns = %#v count:%d", firstTableNode.GetChild(2), count)
	}

	rowOrder := firstCommandExpression(t, ParseCommandBlock(".alter table T policy roworder (C1 asc, C2 desc)")).(*CustomCommand)
	rowCustom := rowOrder.Custom().(*CustomNode)
	rowList, ok := rowCustom.GetChild(6).(*SyntaxList)
	if !ok || rowList.Count() != 2 {
		count := -1
		if rowList != nil {
			count = rowList.Count()
		}
		t.Fatalf("roworder list = %#v count:%d", rowCustom.GetChild(6), count)
	}
	firstRowColumn := rowList.At(0).(*SeparatedElement).Element().(*CustomNode)
	if token, ok := firstRowColumn.GetChild(1).(*SyntaxToken); !ok || token.RawText() != "asc" {
		t.Fatalf("roworder first direction = %#v", firstRowColumn.GetChild(1))
	}
	secondRowColumn := rowList.At(1).(*SeparatedElement).Element().(*CustomNode)
	if token, ok := secondRowColumn.GetChild(1).(*SyntaxToken); !ok || token.RawText() != "desc" {
		t.Fatalf("roworder second direction = %#v", secondRowColumn.GetChild(1))
	}
}

func TestCommandGrammarGeneratedNestedBestCustomCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.show database MyDb schema as json`:                     "ShowDatabaseSchemaAsJson",
		`.attach database MyDb from "https://storage"`:           "AttachDatabase",
		`.alter table T policy caching hot = 1`:                  "AlterTablePolicyCaching",
		`.alter table T policy caching hotdata = 1 hotindex = 2`: "AlterTablePolicyCaching",
		`.alter cluster policy caching hotdata = 1 hotindex = 2`: "AlterClusterPolicyCaching",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	schemaCommand := firstCommandExpression(t, ParseCommandBlock(".show database MyDb schema as json")).(*CustomCommand)
	schemaCustom := schemaCommand.Custom().(*CustomNode)
	schemaChoice, ok := schemaCustom.GetChild(2).(*CustomNode)
	if !ok {
		t.Fatalf("database schema best choice = %#v", schemaCustom.GetChild(2))
	}
	if _, ok := schemaChoice.GetChild(0).(*NameReference); !ok {
		t.Fatalf("database schema choice database child = %#v", schemaChoice.GetChild(0))
	}
	if token, ok := schemaChoice.GetChild(1).(*SyntaxToken); !ok || token.RawText() != "schema" {
		t.Fatalf("database schema choice token = %#v", schemaChoice.GetChild(1))
	}

	attachCommand := firstCommandExpression(t, ParseCommandBlock(`.attach database MyDb from "https://storage"`)).(*CustomCommand)
	attachCustom := attachCommand.Custom().(*CustomNode)
	attachDatabase, ok := attachCustom.GetChild(1).(*CustomNode)
	if !ok {
		t.Fatalf("attach database optional custom = %#v", attachCustom.GetChild(1))
	}
	if token, ok := attachDatabase.GetChild(0).(*SyntaxToken); !ok || token.RawText() != "database" {
		t.Fatalf("attach database token = %#v", attachDatabase.GetChild(0))
	}
	if attachDatabase.IsOptional(1) != true {
		t.Fatalf("attach database all/metadata descriptor optional = %v", attachDatabase.IsOptional(1))
	}
	if attachDatabase.GetName(2) != "DatabaseName" || attachDatabase.CompletionHint(2) != CompletionHintDatabase {
		t.Fatalf("attach database descriptor name=%q hint=%v", attachDatabase.GetName(2), attachDatabase.CompletionHint(2))
	}
	if _, ok := attachDatabase.GetChild(2).(*NameReference); !ok {
		t.Fatalf("attach database name = %#v", attachDatabase.GetChild(2))
	}

	cachingCommand := firstCommandExpression(t, ParseCommandBlock(".alter table T policy caching hotdata = 1 hotindex = 2")).(*CustomCommand)
	cachingCustom := cachingCommand.Custom().(*CustomNode)
	cachingChoice, ok := cachingCustom.GetChild(5).(*CustomNode)
	if !ok {
		t.Fatalf("caching best choice = %#v", cachingCustom.GetChild(5))
	}
	if cachingChoice.ChildCount() != 6 {
		t.Fatalf("caching hotdata child count = %d", cachingChoice.ChildCount())
	}
	if token, ok := cachingChoice.GetChild(0).(*SyntaxToken); !ok || token.RawText() != "hotdata" {
		t.Fatalf("caching hotdata token = %#v", cachingChoice.GetChild(0))
	}
	if token, ok := cachingChoice.GetChild(3).(*SyntaxToken); !ok || token.RawText() != "hotindex" {
		t.Fatalf("caching hotindex token = %#v", cachingChoice.GetChild(3))
	}
}

func TestCommandGrammarGeneratedBestElementCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.attach from "https://storage/account"`:                 "AttachDatabase",
		`.load from "https://storage/account"`:                   "AttachDatabase",
		`.load database all MyDb from "https://storage/account"`: "AttachDatabase",
		`.detach database MyDb`:                                  "DetachDatabase",
		`.drop database MyDb`:                                    "DetachDatabase",
		`.create database MyDb volatile`:                         "CreateDatabase",
		`.show database schema as json`:                          "ShowDatabaseSchemaAsJson",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	loadCommand := firstCommandExpression(t, ParseCommandBlock(`.load from "https://storage/account"`)).(*CustomCommand)
	loadCustom := loadCommand.Custom().(*CustomNode)
	if token, ok := loadCustom.GetChild(0).(*SyntaxToken); !ok || token.RawText() != "load" {
		t.Fatalf("load best-choice child = %#v", loadCustom.GetChild(0))
	}

	loadAllCommand := firstCommandExpression(t, ParseCommandBlock(`.load database all MyDb from "https://storage/account"`)).(*CustomCommand)
	loadAllCustom := loadAllCommand.Custom().(*CustomNode)
	loadAllDatabase, ok := loadAllCustom.GetChild(1).(*CustomNode)
	if !ok {
		t.Fatalf("load all database optional custom = %#v", loadAllCustom.GetChild(1))
	}
	if token, ok := loadAllDatabase.GetChild(1).(*SyntaxToken); !ok || token.RawText() != "all" {
		t.Fatalf("load all hidden selector = %#v", loadAllDatabase.GetChild(1))
	}
	if _, ok := loadAllDatabase.GetChild(2).(*NameReference); !ok {
		t.Fatalf("load all database name = %#v", loadAllDatabase.GetChild(2))
	}

	dropCommand := firstCommandExpression(t, ParseCommandBlock(".drop database MyDb")).(*CustomCommand)
	dropCustom := dropCommand.Custom().(*CustomNode)
	if token, ok := dropCustom.GetChild(0).(*SyntaxToken); !ok || token.RawText() != "drop" {
		t.Fatalf("drop best-choice child = %#v", dropCustom.GetChild(0))
	}
}

func TestCommandGrammarParsesGeneratedEngineRuleBackedCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.show functions`:                                        "ShowFunctions",
		`.drop function MyFunction ifexists`:                     "DropFunction",
		`.create-or-alter workload_group Wg "{}"`:                "CreateOrAleterWorkloadGroup",
		`.alter-merge workload_group Wg "{}"`:                    "AlterMergeWorkloadGroup",
		`.create database MyDb ingestion csv mapping "Map" "[]"`: "CreateDatabaseIngestionMapping",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}
}

func TestCommandGrammarGeneratedGuardedRuleCommands(t *testing.T) {
	for text, kind := range map[string]string{
		`.show database details`:                              "ShowDatabaseDetails",
		`.show table T`:                                       "ShowTable",
		`.show table T details`:                               "ShowTableDetails",
		`.show table T kqlschema`:                             "ShowTableCslSchema",
		`.show table T policy caching`:                        "ShowTablePolicyCaching",
		`.show table * policy caching`:                        "ShowTableStarPolicyCaching",
		`.show external table Ext`:                            "ShowExternalTable",
		`.show external table Ext details`:                    "ShowExternalTableDetails",
		`.show external table Ext artifacts limit 10`:         "ShowExternalTableArtifacts",
		`.create table NewT based-on OldT`:                    "CreateTableBasedOnAnother",
		`.alter table T policy mirroring kind = delta`:        "AlterTablePolicyMirroring",
		`.alter-merge table T policy mirroring kind = delta`:  "AlterMergeTablePolicyMirroring",
		`.alter database MyDb ingestion csv mapping "M" "[]"`: "AlterDatabaseIngestionMapping",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	showTable := firstCommandExpression(t, ParseCommandBlock(".show table T")).(*CustomCommand)
	custom := showTable.Custom().(*CustomNode)
	if custom.ChildCount() != 3 {
		t.Fatalf("show table custom child count = %d", custom.ChildCount())
	}
	if _, ok := custom.GetChild(2).(*NameReference); !ok {
		t.Fatalf("show table guarded child = %#v", custom.GetChild(2))
	}

	for _, text := range []string{
		`.show table *`,
		`.show table * details`,
	} {
		block := ParseCommandBlock(text)
		expression := firstCommandExpression(t, block)
		if wrapped, ok := expression.(*CommandAndSkippedTokens); ok {
			expression, _ = wrapped.Command().(Expression)
		}
		if _, ok := expression.(*PartialCommand); !ok {
			t.Fatalf("%s should retain the recognized command prefix as partial syntax, got %#v", text, expression)
		}
	}
}

func TestCommandGrammarGeneratedParserArrayCommands(t *testing.T) {
	command := firstCommandExpression(t, ParseCommandBlock(".alter table T policy mirroring kind = delta")).(*CustomCommand)
	if command.CommandKind() != "AlterTablePolicyMirroring" {
		t.Fatalf("parser-array command kind = %q", command.CommandKind())
	}
	custom := command.Custom().(*CustomNode)
	if custom.ChildCount() != 11 {
		t.Fatalf("parser-array custom child count = %d", custom.ChildCount())
	}
	if custom.GetChild(5) != nil || !custom.IsOptional(5) {
		t.Fatalf("partition optional fragment = %#v optional:%t", custom.GetChild(5), custom.IsOptional(5))
	}
	for i, text := range map[int]string{6: "kind", 7: "=", 8: "delta"} {
		token, ok := custom.GetChild(i).(*SyntaxToken)
		if !ok || token.RawText() != text {
			t.Fatalf("parser-array child %d = %#v, want token %q", i, custom.GetChild(i), text)
		}
	}
	for i := 9; i <= 10; i++ {
		if custom.GetChild(i) != nil || !custom.IsOptional(i) {
			t.Fatalf("trailing optional placeholder %d = %#v optional:%t", i, custom.GetChild(i), custom.IsOptional(i))
		}
	}
}

func TestCommandGrammarGeneratedOptionalFragmentsParseAtomically(t *testing.T) {
	omitted := firstCommandExpression(t, ParseCommandBlock(".show external table Ext artifacts")).(*CustomCommand)
	if omitted.CommandKind() != "ShowExternalTableArtifacts" {
		t.Fatalf("omitted optional fragment command kind = %q", omitted.CommandKind())
	}
	omittedCustom := omitted.Custom().(*CustomNode)
	if omittedCustom.ChildCount() != 6 || !omittedCustom.IsOptional(5) || omittedCustom.GetChild(5) != nil {
		t.Fatalf("omitted artifact limit fragment = count:%d child5:%#v/%t",
			omittedCustom.ChildCount(), omittedCustom.GetChild(5), omittedCustom.IsOptional(5))
	}

	present := firstCommandExpression(t, ParseCommandBlock(".show external table Ext artifacts limit 10")).(*CustomCommand)
	if present.CommandKind() != "ShowExternalTableArtifacts" {
		t.Fatalf("present optional fragment command kind = %q", present.CommandKind())
	}
	presentCustom := present.Custom().(*CustomNode)
	limit, ok := presentCustom.GetChild(5).(*CustomNode)
	if !ok || !presentCustom.IsOptional(5) {
		t.Fatalf("present artifact limit fragment = %#v optional:%t", presentCustom.GetChild(5), presentCustom.IsOptional(5))
	}
	if token, ok := limit.GetChild(0).(*SyntaxToken); !ok || token.RawText() != "limit" {
		t.Fatalf("present artifact limit token = %#v", limit.GetChild(0))
	}
	if literal, ok := limit.GetChild(1).(*LiteralExpression); !ok || literal.Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("present artifact limit value = %#v", limit.GetChild(1))
	}

	partial := firstCommandExpression(t, ParseCommandBlock(".show external table Ext artifacts limit"))
	if _, ok := partial.(*PartialCommand); !ok || !diagnosticsContain(partial.ContainedDiagnostics(), "KS005") {
		t.Fatalf("partial optional fragment = %#v diagnostics=%#v", partial, partial.ContainedDiagnostics())
	}
}

func TestCommandGrammarGeneratedRuleElementsPreserveCustomChildren(t *testing.T) {
	block := ParseCommandBlock(`.create-or-alter workload_group Wg "{}"`)
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "CreateOrAleterWorkloadGroup" {
		t.Fatalf("workload-group command = %#v", firstCommandExpression(t, block))
	}
	custom := command.Custom().(*CustomNode)
	if custom.ChildCount() != 4 {
		t.Fatalf("workload-group custom child count = %d", custom.ChildCount())
	}
	if _, ok := custom.GetChild(2).(*NameDeclaration); !ok {
		t.Fatalf("workload-group name child = %#v", custom.GetChild(2))
	}
	if literal, ok := custom.GetChild(3).(*LiteralExpression); !ok || literal.Kind() != SyntaxKindStringLiteralExpression {
		t.Fatalf("workload-group string child = %#v", custom.GetChild(3))
	}
}

func TestCommandGrammarGeneratedOptionalFragmentsCanBeOmitted(t *testing.T) {
	for text, kind := range map[string]string{
		".show tables":              "ShowTables",
		".show tables details":      "ShowTablesDetails",
		".show function MyFunction": "ShowFunction",
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if command.CommandKind() != kind {
			t.Fatalf("%s command kind = %q", text, command.CommandKind())
		}
		if command.ReferencedSymbol() == nil {
			t.Fatalf("%s command referenced symbol = %#v", text, command.ReferencedSymbol())
		}
	}

	showTables := firstCommandExpression(t, ParseCommandBlock(".show tables")).(*CustomCommand)
	custom := showTables.Custom().(*CustomNode)
	if custom.ChildCount() != 3 || !custom.IsOptional(2) || custom.GetChild(2) != nil {
		t.Fatalf("show tables optional fragment child = count:%d optional:%t child:%#v", custom.ChildCount(), custom.IsOptional(2), custom.GetChild(2))
	}

	showTablesDetails := firstCommandExpression(t, ParseCommandBlock(".show tables details")).(*CustomCommand)
	detailsCustom := showTablesDetails.Custom().(*CustomNode)
	if detailsCustom.ChildCount() != 4 || !detailsCustom.IsOptional(2) || detailsCustom.GetChild(2) != nil {
		t.Fatalf("show tables details optional fragment child = count:%d optional:%t child:%#v", detailsCustom.ChildCount(), detailsCustom.IsOptional(2), detailsCustom.GetChild(2))
	}
	if token, ok := detailsCustom.GetChild(3).(*SyntaxToken); !ok || token.RawText() != "details" {
		t.Fatalf("show tables details token child = %#v", detailsCustom.GetChild(3))
	}

	showFunction := firstCommandExpression(t, ParseCommandBlock(".show function MyFunction")).(*CustomCommand)
	functionCustom := showFunction.Custom().(*CustomNode)
	if functionCustom.ChildCount() != 4 || !functionCustom.IsOptional(3) || functionCustom.GetChild(3) != nil {
		t.Fatalf("show function optional fragment child = count:%d optional:%t child:%#v", functionCustom.ChildCount(), functionCustom.IsOptional(3), functionCustom.GetChild(3))
	}
}

func TestCommandGrammarGeneratedEngineOptionalTokenCommand(t *testing.T) {
	block := ParseCommandBlock(".show memory")
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "ShowMemory" {
		t.Fatalf("show memory command = %#v kind=%q", firstCommandExpression(t, block), command.CommandKind())
	}
	custom := command.Custom().(*CustomNode)
	if custom.ChildCount() != 3 || !custom.IsOptional(2) || custom.GetChild(2) != nil {
		t.Fatalf("show memory custom optional child = count:%d optional:%t child:%#v", custom.ChildCount(), custom.IsOptional(2), custom.GetChild(2))
	}

	details := ParseCommandBlock(".show memory details")
	detailsCommand, ok := firstCommandExpression(t, details).(*CustomCommand)
	if !ok || detailsCommand.CommandKind() != "ShowMemory" {
		t.Fatalf("show memory details command = %#v", firstCommandExpression(t, details))
	}
	detailsCustom := detailsCommand.Custom().(*CustomNode)
	if detailsCustom.GetChild(2) == nil || detailsCustom.GetChild(2).(*SyntaxToken).RawText() != "details" {
		t.Fatalf("show memory details optional child = %#v", detailsCustom.GetChild(2))
	}
}

func TestCommandGrammarShowCapacityUpdatePolicyOverExternalTable(t *testing.T) {
	block := ParseCommandBlock(".show capacity update-policy-over-external-table")
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "ShowCapacity" {
		t.Fatalf("show capacity command = %#v", firstCommandExpression(t, block))
	}
	custom := command.Custom().(*CustomNode)
	resource, ok := custom.GetChild(2).(*SyntaxToken)
	if !ok || resource.RawText() != "update-policy-over-external-table" {
		t.Fatalf("show capacity resource = %#v", custom.GetChild(2))
	}
	if command.ReferencedSymbol() != EngineCommands.ShowCapacity {
		t.Fatalf("show capacity symbol = %#v", command.ReferencedSymbol())
	}
}

func TestCommandGrammarGeneratedNestedAlternativeNameGuards(t *testing.T) {
	for text, kind := range map[string]string{
		`.show database schema as json`:    "ShowDatabaseSchemaAsJson",
		`.show database db schema as json`: "ShowDatabaseSchemaAsJson",
	} {
		command, ok := firstCommandExpression(t, ParseCommandBlock(text)).(*CustomCommand)
		if !ok || command.CommandKind() != kind {
			t.Fatalf("%q command = %#v", text, command)
		}
	}

	for _, text := range []string{
		`.show database schema schema as json`,
		`.show database schema schema as csl`,
	} {
		block := ParseCommandBlock(text)
		wrapped, ok := firstCommandExpression(t, block).(*CommandAndSkippedTokens)
		if !ok {
			t.Fatalf("%q command = %T, want command with skipped tokens", text, firstCommandExpression(t, block))
		}
		command, ok := wrapped.Command().(*CustomCommand)
		if !ok || command.CommandKind() != "ShowDatabaseSchema" {
			t.Fatalf("%q inner command = %#v", text, wrapped.Command())
		}
		if got := wrapped.SkippedTokens().Text(IncludeTriviaAll); got != ` schema as `+strings.TrimPrefix(text[strings.LastIndex(text, " "):], " ") {
			t.Fatalf("%q skipped text = %q", text, got)
		}
		if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", text, diagnostics)
		}
	}
}

func TestCommandGrammarGeneratedCrossCommandOptionalNameGuards(t *testing.T) {
	for text, kind := range map[string]string{
		`.show database principals`: "ShowDatabasePrincipals",
		`.show database extents`:    "ShowDatabaseExtents",
	} {
		command, ok := firstCommandExpression(t, ParseCommandBlock(text)).(*CustomCommand)
		if !ok || command.CommandKind() != kind {
			t.Fatalf("%q command = %#v", text, command)
		}
	}
	block := ParseCommandBlock(`.show database ingestion`)
	partial, ok := firstCommandExpression(t, block).(*PartialCommand)
	if !ok {
		t.Fatalf("ingestion command = %T, want partial", firstCommandExpression(t, block))
	}
	assertSingleCommandDiagnostic(t, block, "KS005", "Expected one of: apacheavro, avro, azmonstream, csv, json, mappings, orc, parquet, w3clogfile")
	if got := partial.Text(IncludeTriviaAll); got != `.show database ingestion` {
		t.Fatalf("ingestion partial text = %q", got)
	}
}

func TestCommandGrammarBindsCustomCommandSymbolAndResultType(t *testing.T) {
	block := ParseCommandBlock(".show queries")
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "ShowQueries" {
		t.Fatalf("show queries command = %#v", firstCommandExpression(t, block))
	}
	if command.ReferencedSymbol() != EngineCommands.ShowQueries {
		t.Fatalf("show queries referenced symbol = %#v", command.ReferencedSymbol())
	}
	table := EngineCommands.ShowQueries.ResultType()
	if table == nil || len(table.Columns()) != 21 {
		t.Fatalf("show queries command result table = %#v", table)
	}
	if command.RawResultType() != &table.TypeSymbol {
		t.Fatalf("show queries raw result type = %#v, want %#v", command.RawResultType(), &table.TypeSymbol)
	}

	trailingBlock := ParseCommandBlock(".show queries extra")
	withSkipped, ok := firstCommandExpression(t, trailingBlock).(*CommandAndSkippedTokens)
	if !ok {
		t.Fatalf("show queries with skipped tokens = %#v", firstCommandExpression(t, trailingBlock))
	}
	trailingCommand, ok := withSkipped.Command().(*CustomCommand)
	if !ok || trailingCommand.CommandKind() != "ShowQueries" {
		t.Fatalf("inner show queries command = %#v", withSkipped.Command())
	}
	if trailingCommand.ReferencedSymbol() != EngineCommands.ShowQueries {
		t.Fatalf("show queries trailing referenced symbol = %#v", trailingCommand.ReferencedSymbol())
	}
	if withSkipped.SkippedTokens() == nil ||
		withSkipped.SkippedTokens().Tokens().Count() != 1 ||
		withSkipped.SkippedTokens().Tokens().At(0).RawText() != "extra" {
		t.Fatalf("show queries skipped tokens = %#v", withSkipped.SkippedTokens())
	}
	if len(trailingBlock.ContainedDiagnostics()) != 0 {
		t.Fatalf("show queries trailing diagnostics = %#v", trailingBlock.ContainedDiagnostics())
	}
}

func TestCommandGrammarBindsServerSpecificCustomCommandSymbol(t *testing.T) {
	globals := NewGlobalState(nil).WithServerKind(ServerKindDataManager)
	block := CommandGrammarFrom(globals).ParseCommandBlock(".show version extra")
	withSkipped, ok := firstCommandExpression(t, block).(*CommandAndSkippedTokens)
	if !ok {
		t.Fatalf("command with skipped tokens = %#v", firstCommandExpression(t, block))
	}
	if withSkipped.ReferencedSymbol() != nil || withSkipped.RawResultType() != nil {
		t.Fatalf("command skipped wrapper semantics = %#v/%#v", withSkipped.ReferencedSymbol(), withSkipped.RawResultType())
	}
	command, ok := withSkipped.Command().(*CustomCommand)
	if !ok || command.CommandKind() != "ShowVersion" {
		t.Fatalf("inner show version command = %#v", withSkipped.Command())
	}
	if command.ReferencedSymbol() != DataManagerCommands.ShowVersion {
		t.Fatalf("data-manager show version referenced symbol = %#v", command.ReferencedSymbol())
	}
	table := DataManagerCommands.ShowVersion.ResultType()
	if table == nil || len(table.Columns()) != 4 {
		t.Fatalf("data-manager show version result table = %#v", table)
	}
	if command.RawResultType() != &table.TypeSymbol {
		t.Fatalf("data-manager show version raw result type = %#v, want %#v", command.RawResultType(), &table.TypeSymbol)
	}
}

func TestCommandGrammarUnknownCommandHasNoSemanticInfo(t *testing.T) {
	command := firstCommandExpression(t, ParseCommandBlock(".not_a_real_command version")).(*UnknownCommand)
	if command.ReferencedSymbol() != nil || command.RawResultType() != nil {
		t.Fatalf("unknown command semantics = %#v/%#v", command.ReferencedSymbol(), command.RawResultType())
	}
}

func TestCommandGrammarEngineSubsetRecoversRequiredFragments(t *testing.T) {
	partialBlock := ParseCommandBlock(".show")
	if _, ok := firstCommandExpression(t, partialBlock).(*PartialCommand); !ok {
		t.Fatalf("engine incomplete command should retain the recognized prefix, got %#v", firstCommandExpression(t, partialBlock))
	}
}

func TestCommandGrammarParsesGeneratedShowVersionCommand(t *testing.T) {
	for name, globals := range map[string]*GlobalState{
		"aria":    NewGlobalState(nil).WithServerKind(ServerKindAriaBridge),
		"cluster": NewGlobalState(nil).WithServerKind(ServerKindClusterManager),
		"data":    NewGlobalState(nil).WithServerKind(ServerKindDataManager),
	} {
		block := CommandGrammarFrom(globals).ParseCommandBlock(".SHOW VERSION")
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok {
			t.Fatalf("%s command = %#v", name, firstCommandExpression(t, block))
		}
		if command.CommandKind() != "ShowVersion" {
			t.Fatalf("%s command kind = %q", name, command.CommandKind())
		}
		custom, ok := command.Custom().(*CustomNode)
		if !ok || custom.ChildCount() != 2 {
			t.Fatalf("%s custom node = %#v", name, command.Custom())
		}
		if custom.GetChild(0).(*SyntaxToken).RawText() != "SHOW" || custom.GetChild(1).(*SyntaxToken).RawText() != "VERSION" {
			t.Fatalf("%s custom text = %q", name, custom.Text(IncludeTriviaAll))
		}
	}
}

func TestCommandGrammarGeneratedShowVersionRecovery(t *testing.T) {
	globals := NewGlobalState(nil).WithServerKind(ServerKindDataManager)

	partialBlock := CommandGrammarFrom(globals).ParseCommandBlock(".show")
	partial, ok := firstCommandExpression(t, partialBlock).(*PartialCommand)
	if !ok {
		t.Fatalf("partial command = %#v", firstCommandExpression(t, partialBlock))
	}
	if len(partial.CommandKinds()) != 1 || partial.CommandKinds()[0] != "ShowVersion" {
		t.Fatalf("partial command kinds = %#v", partial.CommandKinds())
	}
	if partial.Parts().Count() != 2 || !diagnosticsContain(partial.Parts().At(1).Diagnostics(), "KS005") {
		t.Fatalf("partial parts/diagnostics = %#v/%#v", partial.Parts(), partial.Parts().At(1).Diagnostics())
	}

	trailingBlock := CommandGrammarFrom(globals).ParseCommandBlock(".show version extra")
	withSkipped, ok := firstCommandExpression(t, trailingBlock).(*CommandAndSkippedTokens)
	if !ok {
		t.Fatalf("command with skipped tokens = %#v", firstCommandExpression(t, trailingBlock))
	}
	if command, ok := withSkipped.Command().(*CustomCommand); !ok || command.CommandKind() != "ShowVersion" {
		t.Fatalf("inner command = %#v", withSkipped.Command())
	}
	if withSkipped.SkippedTokens() == nil || withSkipped.SkippedTokens().Tokens().Count() != 1 {
		t.Fatalf("inner skipped tokens = %#v", withSkipped.SkippedTokens())
	}
	if len(trailingBlock.ContainedDiagnostics()) != 0 {
		t.Fatalf("trailing command tokens diagnostics = %#v", trailingBlock.ContainedDiagnostics())
	}

	scriptBlock := ParseCommandBlock(".execute database script <| .show table T details extra")
	if len(scriptBlock.ContainedDiagnostics()) != 0 {
		t.Fatalf("script-input trailing command diagnostics = %#v", scriptBlock.ContainedDiagnostics())
	}

	pipedBlock := CommandGrammarFrom(globals).ParseCommandBlock(".show version | count")
	pipe, ok := firstCommandExpression(t, pipedBlock).(*PipeExpression)
	if !ok {
		t.Fatalf("piped known command = %#v", firstCommandExpression(t, pipedBlock))
	}
	if command, ok := pipe.Expression().(*CustomCommand); !ok || command.CommandKind() != "ShowVersion" {
		t.Fatalf("piped command expression = %#v", pipe.Expression())
	}
	if _, ok := pipe.Operator().(*CountOperator); !ok {
		t.Fatalf("piped command operator = %#v", pipe.Operator())
	}
}

func TestCommandGrammarParsesBadCommandAndCommandPipe(t *testing.T) {
	block := ParseCommandBlock(". | count")
	pipe, ok := firstCommandExpression(t, block).(*PipeExpression)
	if !ok {
		t.Fatalf("command pipe = %#v", firstCommandExpression(t, block))
	}
	bad := pipe.Expression().(*BadCommand)
	if !diagnosticsContain(bad.Diagnostics(), "KS300") {
		t.Fatalf("bad command diagnostics = %#v", bad.Diagnostics())
	}
	if _, ok := pipe.Operator().(*CountOperator); !ok {
		t.Fatalf("pipe operator = %#v", pipe.Operator())
	}
}

func TestCommandGrammarCommandBlockStatementsAndSkippedTokens(t *testing.T) {
	block := ParseCommandBlock(".show version; let x = 1; print x")
	if block.Statements().Count() != 3 || block.SkippedTokens() != nil {
		t.Fatalf("command block statements/skipped = %d/%#v", block.Statements().Count(), block.SkippedTokens())
	}
	if _, ok := block.Statements().At(1).Element().(*LetStatement); !ok {
		t.Fatalf("second statement = %#v", block.Statements().At(1).Element())
	}
	if _, ok := block.Statements().At(2).Element().(*ExpressionStatement); !ok {
		t.Fatalf("third statement = %#v", block.Statements().At(2).Element())
	}

	skipped := ParseCommandBlock(".show <| T")
	if skipped.SkippedTokens() == nil || skipped.SkippedTokens().Tokens().Count() != 2 {
		t.Fatalf("skipped tokens = %#v", skipped.SkippedTokens())
	}
}

func TestCommandGrammarStartLimiterAndCache(t *testing.T) {
	tokens := Lex(".show version\n.show tables")
	source := NewArraySource(tokens)
	if got := CommandLimiterParser().Scan(source, 0); got != 3 {
		t.Fatalf("command limiter length = %d, want 3", got)
	}
	if !IsPossibleStartOfCommand(tokens[3]) {
		t.Fatalf("second dot should start command: %#v", tokens[3])
	}
	if !IsPossibleStartOfCommand(Lex(".show")[0], true) {
		t.Fatal("first dot without trivia should start command")
	}

	defaultOne := CommandGrammarFrom(nil)
	defaultTwo := CommandGrammarFrom(DefaultGlobalState)
	if defaultOne == nil || defaultOne != defaultTwo {
		t.Fatalf("default command grammar cache = %#v / %#v", defaultOne, defaultTwo)
	}

	globals := NewGlobalState(nil).WithServerKind(ServerKindDataManager).WithCache()
	first := CommandGrammarFrom(globals)
	second := CommandGrammarFrom(globals)
	if first == nil || first != second {
		t.Fatalf("global command grammar cache = %#v / %#v", first, second)
	}
}

func TestKustoCodeParsesCommandTextAsCommandBlock(t *testing.T) {
	code := Parse(".show version")
	if code.Kind != CodeKindCommand || code.Query != nil {
		t.Fatalf("command code kind/query = %s/%#v", code.Kind, code.Query)
	}
	block, ok := WrapGeneratedSyntaxNode(code.Syntax()).(*CommandBlock)
	if !ok || block.Statements().Count() != 1 {
		t.Fatalf("command syntax = %#v", code.Syntax())
	}
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "ShowVersion" {
		t.Fatalf("command expression = %#v", firstCommandExpression(t, block))
	}
}

func TestKustoCodeUsesServerSpecificCommandGrammar(t *testing.T) {
	code := ParseWithGlobals(".show version", NewGlobalState(nil).WithServerKind(ServerKindDataManager))
	if code.Kind != CodeKindCommand || code.Query != nil {
		t.Fatalf("command code kind/query = %s/%#v", code.Kind, code.Query)
	}
	block, ok := WrapGeneratedSyntaxNode(code.Syntax()).(*CommandBlock)
	if !ok || block.Statements().Count() != 1 {
		t.Fatalf("command syntax = %#v", code.Syntax())
	}
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "ShowVersion" {
		t.Fatalf("command expression = %#v", firstCommandExpression(t, block))
	}
}

func firstCommandExpression(t *testing.T, block *CommandBlock) Expression {
	t.Helper()
	statement, ok := block.Statements().At(0).Element().(*ExpressionStatement)
	if !ok {
		t.Fatalf("first command statement = %#v", block.Statements().At(0).Element())
	}
	return statement.Expression()
}
