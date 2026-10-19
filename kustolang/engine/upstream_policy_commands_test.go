package kustolang

import "testing"

// Covers the command additions at upstream revision 6b5c2f7.
func TestUpstreamChangeDataCaptureAndPrioritizationCommands(t *testing.T) {
	tests := []struct {
		query  string
		symbol *CommandSymbol
	}{
		{".show table T policy changedatacapture", EngineCommands.ShowTablePolicyChangeDataCapture},
		{".show table db.T policy changedatacapture", EngineCommands.ShowTablePolicyChangeDataCapture},
		{".alter table T policy changedatacapture (All)", EngineCommands.AlterTablePolicyChangeDataCapture},
		{".alter table ['policy'] policy changedatacapture (All, ['Other'])", EngineCommands.AlterTablePolicyChangeDataCapture},
		{".alter tables (T, U) policy changedatacapture (All, Other)", EngineCommands.AlterTablesPolicyChangeDataCapture},
		{".delete table T policy changedatacapture", EngineCommands.DeleteTablePolicyChangeDataCapture},
		{".show cluster policy materialized-view-prioritization", EngineCommands.ShowClusterPolicyMaterializedViewPrioritization},
		{`.alter cluster policy materialized-view-prioritization '{"Enabled":true}'`, EngineCommands.AlterClusterPolicyMaterializedViewPrioritization},
		{`.alter-merge cluster policy materialized-view-prioritization '{}'`, EngineCommands.AlterMergeClusterPolicyMaterializedViewPrioritization},
		{".drop cluster policy materialized-view-prioritization", EngineCommands.DropClusterPolicyMaterializedViewPrioritization},
	}
	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			block := ParseCommandBlock(test.query)
			if diagnostics := block.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			command, ok := firstCommandExpression(t, block).(*CustomCommand)
			if !ok || command.CommandKind() != test.symbol.SymbolName() || command.ReferencedSymbol() != test.symbol {
				t.Fatalf("command = %#v, want %s", command, test.symbol.SymbolName())
			}
			columns := test.symbol.ResultType().Columns()
			want := []string{"PolicyName", "EntityName", "Policy", "ChildEntities", "EntityType"}
			if test.symbol == EngineCommands.ShowClusterPolicyMaterializedViewPrioritization {
				want = append(want, "EffectivePolicy")
			}
			if len(columns) != len(want) {
				t.Fatalf("columns = %#v, want %v", columns, want)
			}
			for i, name := range want {
				if columns[i].SymbolName() != name || columns[i].Type != ScalarTypes.String.Type() {
					t.Fatalf("column %d = %#v, want %s:string", i, columns[i], name)
				}
			}
		})
	}
	for _, query := range []string{
		".alter table T policy changedatacapture",
		".alter table T policy changedatacapture ()",
		".alter tables () policy changedatacapture (All)",
		".alter tables (T, U) policy changedatacapture ()",
		".alter table T policy changedatacapture (1)",
		".alter table T policy changedatacapture (All Other)",
		".alter cluster policy materialized-view-prioritization",
		".alter-merge cluster policy materialized-view-prioritization 1",
		".delete cluster policy materialized-view-prioritization",
	} {
		t.Run(query, func(t *testing.T) {
			if diagnostics := ParseCommandBlock(query).ContainedDiagnostics(); len(diagnostics) == 0 {
				t.Fatal("expected diagnostics for malformed policy command")
			}
		})
	}
}

func TestUpstreamPolicyCommandCompletions(t *testing.T) {
	for _, test := range []struct{ query, want string }{
		{".show table T policy ", "changedatacapture"},
		{".alter table T policy ", "changedatacapture"},
		{".show cluster policy ", "materialized-view-prioritization"},
		{".drop cluster policy ", "materialized-view-prioritization"},
	} {
		items := NewKustoCodeService(test.query).GetCompletionItems(len(test.query), CompletionOptionsDefault, CancellationToken{}).Items
		assertCompletionIncludes(t, items, test.want)
	}
}
