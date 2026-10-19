package main

import "testing"

func TestCSharpDiagnosticFactoryInventory(t *testing.T) {
	source := []byte(`
        // public static Diagnostic GetIgnored()
        public static Diagnostic GetPlain(string text) => new Diagnostic("public static Diagnostic GetIgnored()");
        public static Diagnostic GetGeneric<T>(params T[] values) => default;
        /* public static Diagnostic GetAlsoIgnored() */
    `)
	tokens, err := csharpTokens(source)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for i := range tokens {
		if isDiagnosticFactoryDeclaration(tokens, i) {
			names = append(names, tokens[i+3])
		}
	}
	if len(names) != 2 || names[0] != "GetPlain" || names[1] != "GetGeneric" {
		t.Fatalf("factory names = %#v", names)
	}
}

func TestCSharpTokensRejectUnterminatedInput(t *testing.T) {
	for _, source := range []string{"/* missing", `"missing`} {
		if _, err := csharpTokens([]byte(source)); err == nil {
			t.Fatalf("csharpTokens(%q) succeeded", source)
		}
	}
}
