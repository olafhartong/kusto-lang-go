package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestCommandFactsUsesUTF16CommandBlockCoordinates(t *testing.T) {
	text := "😀 preface\n// first\n.show tables\n\n😀 value\n// second\n.show databases"
	lineStarts := kustolang.TextFacts.GetLineStarts(text)
	wantStarts := []int{0, 43}

	if got := kustolang.CommandFacts.GetCommandBlockStarts(text); !reflect.DeepEqual(got, wantStarts) {
		t.Fatalf("GetCommandBlockStarts() = %v, want %v", got, wantStarts)
	}
	if got := kustolang.CommandFacts.GetCommandBlockStarts(text, lineStarts); !reflect.DeepEqual(got, wantStarts) {
		t.Fatalf("GetCommandBlockStarts(lineStarts) = %v, want %v", got, wantStarts)
	}
	if !kustolang.CommandFacts.IsCommandStartLine(text, lineStarts[2]) {
		t.Fatal("IsCommandStartLine() = false for first command")
	}
	if kustolang.CommandFacts.IsCommandStartLine(text, lineStarts[1]) {
		t.Fatal("IsCommandStartLine() = true for comment")
	}
}

func TestCommandFactsReturnsIndependentBlocksAndTrimmedCommands(t *testing.T) {
	text := "// lead\n.show tables\n\n// next\n.show databases\n// tail"
	blocks := kustolang.CommandFacts.GetCommandBlockTexts(text)
	if len(blocks) != 2 || blocks[0] != "// lead\n.show tables\n\n" || blocks[1] != "// next\n.show databases\n// tail" {
		t.Fatalf("GetCommandBlockTexts() = %#v", blocks)
	}
	commands := kustolang.CommandFacts.GetCommandTexts(text)
	want := []string{".show tables", ".show databases"}
	if !reflect.DeepEqual(commands, want) {
		t.Fatalf("GetCommandTexts() = %#v, want %#v", commands, want)
	}

	blocks[0] = "changed"
	if got := kustolang.CommandFacts.GetCommandBlockTexts(text)[0]; got != "// lead\n.show tables\n\n" {
		t.Fatalf("GetCommandBlockTexts() retained caller mutation: %q", got)
	}
}

func TestCommandFactsNoCommandReturnsInitialBlock(t *testing.T) {
	if got := kustolang.CommandFacts.GetCommandBlockStarts("T | count"); !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("GetCommandBlockStarts(no command) = %v", got)
	}
	if got := kustolang.CommandFacts.GetCommandTexts("T | count"); len(got) != 0 {
		t.Fatalf("GetCommandTexts(no command) = %#v", got)
	}
}
