package kustolang

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unicode/utf16"
)

const nov12Commit = "3a7d2d690172a4ac73d6b2e2ae5b2f6a37cbe64f"

type nov12Fixture struct {
	Commit, Construction, Schema string
	CommandCount                 int
	Commands                     []attestedLanguageResult
	Completions                  []struct {
		Text  string
		Items []string
	}
	Ranges []struct {
		Text          string
		Start, Length int
		Value         string
	}
	RepeatedRangeSameInstance bool
}

func readNov12Fixture(t *testing.T) nov12Fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/upstream_3a7d2d6_commands.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture nov12Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Commit != nov12Commit || fixture.CommandCount != 646 || len(fixture.Commands) != 22 || len(fixture.Completions) != 5 || len(fixture.Ranges) != 6 {
		t.Fatal("incorrectly pinned or incomplete November 12 oracle")
	}
	return fixture
}

func TestUpstreamNov12Commands(t *testing.T) {
	fixture := readNov12Fixture(t)
	for _, want := range fixture.Commands {
		t.Run(want.Query, func(t *testing.T) {
			got := currentAttestedLanguageResult(want.Query)
			if len(got.Columns) == 0 {
				got.Columns = nil
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("command = %#v, want %#v", got, want)
			}
		})
	}
	command := EngineCommands.AlterDatabaseExtentContainerPath
	if len(EngineCommands.All) != fixture.CommandCount || command == nil || command.Construction() != fixture.Construction || command.ResultSchema() != fixture.Schema {
		t.Fatal("new command catalog metadata does not match upstream")
	}
	if DefaultGlobalState.GetCommand(command.SymbolName()) != command {
		t.Error("new command is not registered in globals")
	}
	for text, kind := range map[string]string{
		".show capacity async-update-policy":                   "ShowCapacity",
		".alter database db extentcontainer with (path='new')": "AlterDatabaseExtentContainerPath",
	} {
		command, ok := firstCommandExpression(t, ParseCommandBlock(text)).(*CustomCommand)
		if !ok || command.CommandKind() != kind {
			t.Errorf("%q parsed as %#v, want %s", text, command, kind)
		}
	}
}

func TestUpstreamNov12CommandCompletion(t *testing.T) {
	for _, want := range readNov12Fixture(t).Completions {
		items := NewKustoCodeService(want.Text).GetCompletionItems(utf16TextLength(want.Text), CompletionOptionsDefault.WithIncludeSymbols(false).WithIncludeExtendedSyntax(false), CancellationToken{}).Items
		got := completionDisplayTexts(items)
		// Compare the vocabulary affected by this sync. The complete source
		// list remains in the fixture for future completion-parity work.
		for _, name := range []string{"async-update-policy", "update-policy-over-external-table", "query-acceleration", "extentcontainer", "with"} {
			has := func(items []string) bool {
				for _, item := range items {
					if item == name {
						return true
					}
				}
				return false
			}
			if has(got) != has(want.Items) {
				t.Errorf("%q completion %q present=%v, want %v", want.Text, name, has(got), has(want.Items))
			}
		}
	}
}

func TestUpstreamNov12StringTableRanges(t *testing.T) {
	table := NewStringTable()
	for _, want := range readNov12Fixture(t).Ranges {
		units := utf16.Encode([]rune(want.Text))
		start := len(string(utf16.Decode(units[:want.Start])))
		end := len(string(utf16.Decode(units[:want.Start+want.Length])))
		got := table.AddRange(want.Text, start, end-start)
		if got != want.Value {
			t.Errorf("range %q = %q, want %q", want.Text, got, want.Value)
		}
	}
	if len(table.Values()) != 4 {
		t.Errorf("interned values = %q, want four distinct strings", table.Values())
	}
	if !readNov12Fixture(t).RepeatedRangeSameInstance {
		t.Fatal("source oracle did not reuse the interned range")
	}
}

func TestUpstreamNov12HistoricalBehavior(t *testing.T) {
	for _, suffix := range []string{"behavior", "caching"} {
		oldData, err := os.ReadFile("testdata/upstream_dbd7ca7_" + suffix + ".json")
		if err != nil {
			t.Fatal(err)
		}
		currentData, err := os.ReadFile("testdata/upstream_3a7d2d6_" + suffix + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var old, current map[string]any
		if err := json.Unmarshal(oldData, &old); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(currentData, &current); err != nil {
			t.Fatal(err)
		}
		if current["commit"] != nov12Commit {
			t.Fatal("current historical-behavior fixture is not pinned")
		}
		old["commit"] = nov12Commit
		if !reflect.DeepEqual(old, current) {
			t.Errorf("%s source behavior changed from the replayed November 8 fixture", suffix)
		}
	}
}
