package kustolang

import (
	"strings"
	"testing"
)

func TestBuiltInOptionsCatalog(t *testing.T) {
	if len(Options.All) != 58 {
		t.Fatalf("built-in option count = %d, want 58", len(Options.All))
	}
	if Options.All[0] != Options.BestEffort || Options.All[len(Options.All)-1] != Options.ValidatePermissions {
		t.Fatalf("built-in option order does not match upstream")
	}

	seen := map[string]bool{}
	for _, option := range Options.All {
		if option == nil {
			t.Fatal("built-in options must not contain nil entries")
		}
		name := option.SymbolName()
		if seen[name] {
			t.Fatalf("duplicate built-in option %q", name)
		}
		seen[name] = true
	}

	if Options.QueryNow.SymbolName() != "query_now" || !strings.Contains(Options.QueryNow.Description(), "now(0s)") {
		t.Fatalf("query_now option metadata was not ported")
	}
	if got := Options.DebugPython.Types(); len(got) != 2 || got[0] != ScalarTypes.Bool || got[1] != ScalarTypes.Int {
		t.Fatalf("query_python_debug types = %#v", got)
	}
	if got := Options.QueryLanguage.Examples(); len(got) != 3 || got[0] != "'csl'" || got[1] != "'kql'" || got[2] != "'sql'" {
		t.Fatalf("query_language examples = %#v", got)
	}
	if got := Options.ResultsProgressiveEnabled.Types(); len(got) != 0 {
		t.Fatalf("typeless upstream option should stay typeless, got %#v", got)
	}
}

func TestDefaultGlobalStateIncludesBuiltInOptions(t *testing.T) {
	if DefaultGlobalState.GetOption("query_now") != Options.QueryNow {
		t.Fatal("default global state should expose built-in query options")
	}
	globals := NewGlobalState(nil)
	if globals.GetOption("query_language") != Options.QueryLanguage {
		t.Fatal("new global state should expose built-in query options")
	}

	globals.Options[0] = nil
	if DefaultGlobalState.GetOption("best_effort") != Options.BestEffort {
		t.Fatal("new global state option slice should not alias DefaultGlobalState")
	}
}
