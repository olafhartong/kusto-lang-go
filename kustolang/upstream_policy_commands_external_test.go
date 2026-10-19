package kustolang_test

import (
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicChangeDataCaptureAndPrioritizationCommands(t *testing.T) {
	for _, command := range []*kustolang.CommandSymbol{
		kustolang.EngineCommands.ShowTablePolicyChangeDataCapture,
		kustolang.EngineCommands.AlterTablePolicyChangeDataCapture,
		kustolang.EngineCommands.AlterTablesPolicyChangeDataCapture,
		kustolang.EngineCommands.DeleteTablePolicyChangeDataCapture,
		kustolang.EngineCommands.ShowClusterPolicyMaterializedViewPrioritization,
		kustolang.EngineCommands.AlterClusterPolicyMaterializedViewPrioritization,
		kustolang.EngineCommands.AlterMergeClusterPolicyMaterializedViewPrioritization,
		kustolang.EngineCommands.DropClusterPolicyMaterializedViewPrioritization,
	} {
		if command == nil {
			t.Fatal("new upstream policy command is missing")
		}
		if got := kustolang.DefaultGlobalState.GetCommand(command.SymbolName()); got != command {
			t.Fatalf("%s not registered in public default catalog", command.SymbolName())
		}
		want := 5
		if command == kustolang.EngineCommands.ShowClusterPolicyMaterializedViewPrioritization {
			want = 6
		}
		if table := command.ResultType(); table == nil || len(table.Columns()) != want {
			t.Fatalf("%s result = %#v, want %d columns", command.SymbolName(), table, want)
		}
	}
}

func TestPublicIsCommonArgumentTypeKind(t *testing.T) {
	for _, kind := range []kustolang.ParameterTypeKind{
		kustolang.ParameterTypeCommonScalar, kustolang.ParameterTypeCommonNumber,
		kustolang.ParameterTypeCommonSummable, kustolang.ParameterTypeCommonOrderable,
		kustolang.ParameterTypeCommonScalarOrDynamic,
	} {
		if !kustolang.IsCommonArgumentTypeKind(kind) {
			t.Fatalf("common type kind %v returned false", kind)
		}
	}
	for _, kind := range []kustolang.ParameterTypeKind{
		kustolang.ParameterTypeDeclared, kustolang.ParameterTypeScalar,
		kustolang.ParameterTypeNumber, kustolang.ParameterTypeKind("invalid"),
	} {
		if kustolang.IsCommonArgumentTypeKind(kind) {
			t.Fatalf("non-common type kind %v returned true", kind)
		}
	}
}
