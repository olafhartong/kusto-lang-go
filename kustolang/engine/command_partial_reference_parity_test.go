package kustolang

import "testing"

func TestGeneratedCommandPartialParsingUsesLongestDescriptorPrefix(t *testing.T) {
	block := ParseCommandBlock(`.show materialized-view MV schema csl`)
	wrapped, ok := firstCommandExpression(t, block).(*CommandAndSkippedTokens)
	if !ok {
		t.Fatalf("command = %T, want skipped partial", firstCommandExpression(t, block))
	}
	partial, ok := wrapped.Command().(*PartialCommand)
	if !ok {
		t.Fatalf("inner command = %T, want partial", wrapped.Command())
	}
	if got := partial.Text(IncludeTriviaAll); got != `.show materialized-view MV schema` {
		t.Fatalf("partial boundary = %q", got)
	}
	if got := wrapped.SkippedTokens().Text(IncludeTriviaAll); got != ` csl` {
		t.Fatalf("skipped boundary = %q", got)
	}
	if len(partial.CommandKinds()) != 1 || partial.CommandKinds()[0] != "ShowMaterializedViewSchemaAsJson" {
		t.Fatalf("partial kinds = %#v", partial.CommandKinds())
	}
	assertSingleCommandDiagnostic(t, block, "KS005", "Expected: as")
}

func TestGeneratedCommandSeparatedListRecoveryIsTransactional(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		partialText string
		skippedText string
		message     string
	}{
		{"trailing comma", `.set-or-append Target with (folder="Ops",) <| Source | project x`, `.set-or-append Target with (folder="Ops",`, `)`, "Expected one of: <name>, creationTime, distributed, docstring, extend_schema, folder, format, ignoreFirstRecord, ingestIfNotExists, ingestionMapping, ingestionMappingReference, persistDetails, policy_ingestionTime, recreate_schema, small_dimension_table, tags, validationPolicy, zipPattern"},
		{"double comma", `.set-or-append Target with (folder="Ops",, format="csv") <| Source | project x`, `.set-or-append Target with (folder="Ops",`, `, format="csv")`, "Expected one of: <name>, creationTime, distributed, docstring, extend_schema, folder, format, ignoreFirstRecord, ingestIfNotExists, ingestionMapping, ingestionMappingReference, persistDetails, policy_ingestionTime, recreate_schema, small_dimension_table, tags, validationPolicy, zipPattern"},
		{"missing comma", `.set-or-append Target with (folder="Ops" format="csv") <| Source | project x`, `.set-or-append Target with (folder="Ops"`, ` format="csv")`, "Expected: ')'"},
		{"script trailing comma", `.execute database script with (ContinueOnErrors=true,) <| .show table Source details`, `.execute database script with (ContinueOnErrors=true,`, `)`, "Expected one of: <name>, ContinueOnErrors, ThrowOnErrors"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := ParseCommandBlock(tt.text)
			wrapped, ok := firstCommandExpression(t, block).(*CommandAndSkippedTokens)
			if !ok {
				t.Fatalf("command = %T, want skipped partial", firstCommandExpression(t, block))
			}
			partial, ok := wrapped.Command().(*PartialCommand)
			if !ok {
				t.Fatalf("inner command = %T, want partial", wrapped.Command())
			}
			if got := partial.Text(IncludeTriviaAll); got != tt.partialText {
				t.Fatalf("partial boundary = %q, want %q", got, tt.partialText)
			}
			if got := wrapped.SkippedTokens().Text(IncludeTriviaAll); got != tt.skippedText {
				t.Fatalf("skipped boundary = %q, want %q", got, tt.skippedText)
			}
			if diagnosticsContain(partial.ContainedDiagnostics(), "KS006") {
				t.Fatalf("partial fabricated a missing complex list element: %#v", partial.ContainedDiagnostics())
			}
			if !diagnosticsContainMessage(partial.ContainedDiagnostics(), "KS005", tt.message) {
				t.Fatalf("partial diagnostics = %#v", partial.ContainedDiagnostics())
			}
		})
	}
}

func TestCurrentCommandPartialParserGraphOwnsGuardedListBoundaries(t *testing.T) {
	grammar := CommandGrammarFrom(nil)
	tests := []struct {
		text string
		want string
	}{
		{`.set-or-append Target with (folder="Ops",) <| Source | project x`, `.set-or-append Target with (folder="Ops",`},
		{`.set-or-append Target with (folder="Ops",, format="csv") <| Source | project x`, `.set-or-append Target with (folder="Ops",`},
		{`.set-or-append Target with (folder="Ops" format="csv") <| Source | project x`, `.set-or-append Target with (folder="Ops"`},
		{`.show table T schema`, `.show table T schema`},
		{`.show materialized-view MV schema`, `.show materialized-view MV schema`},
		{`.show tables (T1, T2,`, ``},
		{`.attach extents into table T by container "uri" "id1" (`, ``},
	}
	for _, test := range tests {
		parser := newCommandParserState(Lex(test.text), grammar)
		partial := parser.parsePartialKnownCommandWithParsers()
		if partial == nil {
			t.Fatalf("%q generic parser returned nil", test.text)
		}
		got := partial.Text(IncludeTriviaAll)
		if test.want != "" && got != test.want {
			t.Fatalf("%q generic boundary = %q, want %q", test.text, got, test.want)
		}
		genericEnd := parser.pos
		parser.pos = 0
		descriptor := parser.parsePartialKnownCommandWithDescriptors(grammar.commandParsers)
		if descriptor == nil {
			t.Fatalf("%q descriptor parser returned nil", test.text)
		}
		if descriptorEnd := parser.pos; genericEnd < descriptorEnd {
			t.Fatalf("%q generic end %d before descriptor end %d (generic %q, descriptor %q)", test.text, genericEnd, descriptorEnd, got, descriptor.Text(IncludeTriviaAll))
		}
	}
}

func TestGeneratedCommandSeparatedListValidControls(t *testing.T) {
	for _, text := range []string{
		`.set-or-append Target with (folder="Ops") <| Source | project x`,
		`.set-or-append Target with (folder="Ops", format="csv") <| Source | project x`,
	} {
		block := ParseCommandBlock(text)
		command, ok := firstCommandExpression(t, block).(*CustomCommand)
		if !ok || command.CommandKind() != "SetOrAppendTable" {
			t.Fatalf("%s command = %#v", text, firstCommandExpression(t, block))
		}
		if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
			t.Fatalf("%s diagnostics = %#v", text, diagnostics)
		}
	}
}

func TestGeneratedCommandMissingCloseBracketUsesExpectedDiagnostic(t *testing.T) {
	block := ParseCommandBlock(`.ingest inline into table SigninLogs [a,b`)
	command, ok := firstCommandExpression(t, block).(*CustomCommand)
	if !ok || command.CommandKind() != "IngestInlineIntoTable" {
		t.Fatalf("command = %#v", firstCommandExpression(t, block))
	}
	assertSingleCommandDiagnostic(t, block, "KS005", "Expected: ]")
}

func TestCurrentCommandPartialFailureUsesParserAnnotationsAndTags(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		message string
	}{
		{"wildcard rule tag", `.show cluster databases (`, "Expected one of: <database>, <wildcard>"},
		{"string literal rule tag", `.create database Db persist (`, "Expected: <string-literal>"},
		{"type completion alternatives", `.create table T (A:`, "Expected one of: bool, boolean, date, datetime, decimal, double, dynamic, float, guid, int, int16, int32, int64, int8, long, real, string, time, timespan, uint, uint16, uint32, uint64, uint8, ulong, uniqueid"},
		{"pipe tag splitting", `.set T`, "Expected one of: ', '<, with"},
		{"graph model rule tag", `.show graph_model`, "Expected one of: *, <graphmodel>"},
		{"tied command alternatives", `.show database schema as`, "Expected one of: csl, json, kql"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			block := ParseCommandBlock(test.text)
			assertSingleCommandDiagnostic(t, block, "KS005", test.message)
		})
	}
}

func assertSingleCommandDiagnostic(t *testing.T, element SyntaxElement, code, message string) {
	t.Helper()
	diagnostics := element.ContainedDiagnostics()
	if len(diagnostics) != 1 || diagnostics[0].Code != code || diagnostics[0].Message != message {
		t.Fatalf("diagnostics = %#v, want %s %q", diagnostics, code, message)
	}
}
