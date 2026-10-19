// Command langsmoke is an out-of-tree consumer of the reusable language layer
// (kustolang + kustolang/environment). It exists to prove, in CI, that the
// language layer compiles and works without the analyzer packages. It doubles
// as a minimal usage example for external projects.
package main

import (
	"fmt"
	"os"

	"github.com/olafhartong/kusto-lang-go/kustolang"
	"github.com/olafhartong/kusto-lang-go/kustolang/environment"
)

func main() {
	env := &environment.EnvironmentDefinition{
		TableDetails: map[string]environment.TableDetails{
			"MyLogs": {"Timestamp": "datetime", "Computer": "string", "EventID": "int"},
		},
	}

	code, err := kustolang.ParseAndAnalyzeForEnvironment(
		"MyLogs | where EventID == 4624 | project Timestamp, Computer",
		env, environment.LocalData{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "analyze:", err)
		os.Exit(1)
	}
	diagnostics := code.GetDiagnostics()
	fmt.Printf("diagnostics: %d\n", len(diagnostics))
	for _, d := range diagnostics {
		fmt.Printf("  %s: %s\n", d.Code, d.Message)
	}

	service, err := kustolang.NewCodeServiceForEnvironment(
		"MyLogs | where ", env, environment.LocalData{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "code service:", err)
		os.Exit(1)
	}
	completion := service.GetCompletionItemsDefault(15, kustolang.CancellationToken{})
	fmt.Printf("completions at |where: %d\n", len(completion.Items))

	if len(diagnostics) != 0 || len(completion.Items) == 0 {
		fmt.Fprintln(os.Stderr, "unexpected smoke result")
		os.Exit(1)
	}
	fmt.Println("language layer smoke OK")
}
