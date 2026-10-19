package kustolang_test

import (
	"reflect"
	"testing"

	"github.com/olafhartong/kusto-lang-go/kustolang"
)

func TestPublicEnvironmentSymbolConstruction(t *testing.T) {
	account := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type())
	table := kustolang.NewTableSymbolWithDescription("Events", []*kustolang.ColumnSymbol{account}, "events")
	members := []kustolang.Symbol{table}
	database := kustolang.NewDatabaseSymbolWithAlternateName("Samples", "SamplesAlias", members, true)
	members[0] = kustolang.EmptyTableSymbol
	if !database.IsOpen() || database.AlternateName() != "SamplesAlias" || database.GetTable("Events") != table {
		t.Fatalf("database = %#v members=%#v", database, database.Members())
	}

	databases := []*kustolang.DatabaseSymbol{database}
	cluster := kustolang.NewClusterSymbolWithDatabases("help", databases, true)
	databases[0] = kustolang.UnknownDatabaseSymbol
	if !cluster.IsOpen() || cluster.GetDatabase("SamplesAlias") != database {
		t.Fatalf("cluster = %#v databases=%#v", cluster, cluster.Databases())
	}

	types := []*kustolang.ScalarSymbol{kustolang.ScalarTypes.String}
	examples := []string{"'value'"}
	option := kustolang.NewOptionSymbol("query_language", "language", types, examples)
	types[0] = kustolang.ScalarTypes.Long
	examples[0] = "changed"
	if option.Description() != "language" || !reflect.DeepEqual(option.Types(), []*kustolang.ScalarSymbol{kustolang.ScalarTypes.String}) || !reflect.DeepEqual(option.Examples(), []string{"'value'"}) {
		t.Fatalf("option = %#v types=%#v examples=%#v", option, option.Types(), option.Examples())
	}

	entityGroup := kustolang.NewEntityGroupSymbolFromDefinition("Targets", `[cluster("help").database("Samples")]`, "targets")
	if entityGroup.Definition() == "" || entityGroup.Description() != "targets" || entityGroup.Signature() == nil {
		t.Fatalf("entity group = %#v", entityGroup)
	}
	if body := kustolang.EntityGroupBodyFromDefinition(`"cluster('help').database('Samples')"`); body != `entity_group [cluster('help').database('Samples')]` {
		t.Fatalf("entity group body = %q", body)
	}
	element := kustolang.NewEntityGroupElementSymbol("target", kustolang.NewEntityGroupSymbol("Targets", database), nil)
	if element.EntityGroup == nil || element.UnderlyingSymbol != database || element.SymbolTabularity() != database.SymbolTabularity() {
		t.Fatalf("entity group element = %#v", element)
	}
}

func TestPublicGraphTupleAndCommandConstruction(t *testing.T) {
	graph, err := kustolang.NewGraphSymbolFromSchema("Connections", "Source:string, Target:string", "Id:long")
	if err != nil {
		t.Fatal(err)
	}
	if graph.SymbolName() != "Connections" || graph.EdgeShape.GetColumn("Source") == nil || graph.NodeShape.GetColumn("Id") == nil {
		t.Fatalf("graph = %#v edge=%#v node=%#v", graph, graph.EdgeShape, graph.NodeShape)
	}
	tuple, err := kustolang.TupleSymbolFromSchema("Name:string, Count:long")
	if err != nil {
		t.Fatal(err)
	}
	if len(tuple.Columns()) != 2 || tuple.Columns()[0].SymbolName() != "Name" {
		t.Fatalf("tuple = %#v columns=%#v", tuple, tuple.Columns())
	}
	recoveredGraph, err := kustolang.GraphSymbolFromSchema("Name:")
	if err != nil || len(recoveredGraph.EdgeShape.Columns()) != 1 || recoveredGraph.EdgeShape.Columns()[0].Type == nil {
		t.Fatalf("recovered graph schema = graph:%#v columns:%#v err:%v", recoveredGraph, recoveredGraph.EdgeShape.Columns(), err)
	}
	recoveredTuple, err := kustolang.TupleSymbolFromSchema("Name:")
	if err != nil || len(recoveredTuple.Columns()) != 1 || recoveredTuple.Columns()[0].Type == nil {
		t.Fatalf("recovered tuple schema = tuple:%#v columns:%#v err:%v", recoveredTuple, recoveredTuple.Columns(), err)
	}

	second, err := kustolang.GraphSymbolFromSchema("Weight:real")
	if err != nil {
		t.Fatal(err)
	}
	merged := kustolang.MergeGraphSymbols(graph, second)
	if merged == nil || merged.EdgeShape.GetColumn("Source") == nil || merged.EdgeShape.GetColumn("Weight") == nil {
		t.Fatalf("merged graph = %#v", merged)
	}

	snapshot := kustolang.NewGraphSnapshotSymbol("Latest")
	edges := []*kustolang.Signature{kustolang.NewComputedSignature("Edges", kustolang.TabularityTabular)}
	model := kustolang.NewGraphModelSymbol("Network", edges, nil, []*kustolang.GraphSnapshotSymbol{snapshot})
	got, ok := model.TryGetSnapshot("Latest")
	if snapshot.Model != model || edges[0].Symbol != model || !ok || got != snapshot {
		t.Fatalf("graph model = %#v snapshot=%#v", model, snapshot)
	}
	group := kustolang.NewGroupSymbol(graph, model)
	if len(group.Members()) != 2 || group.SymbolTabularity() != graph.SymbolTabularity() {
		t.Fatalf("group = %#v members=%#v", group, group.Members())
	}

	command := kustolang.NewCommandSymbolWithResultType(".show sample", graph.EdgeShape, "show_sample")
	if command.ResultType() != graph.EdgeShape || command.Construction() != "show_sample" || command.ResultSchema() == "" {
		t.Fatalf("command = %#v schema=%q", command, command.ResultSchema())
	}
}

func TestPublicCanonicalSymbols(t *testing.T) {
	if !kustolang.UnknownClusterSymbol.IsOpen() || !kustolang.UnknownDatabaseSymbol.IsOpen() {
		t.Fatalf("unknown symbols = cluster:%#v database:%#v", kustolang.UnknownClusterSymbol, kustolang.UnknownDatabaseSymbol)
	}
	if len(kustolang.EmptyTableSymbol.Columns()) != 0 || len(kustolang.EmptyTupleSymbol.Columns()) != 0 || len(kustolang.EmptyEntityGroupSymbol.Members()) != 0 {
		t.Fatalf("empty symbols = table:%#v tuple:%#v entity-group:%#v", kustolang.EmptyTableSymbol, kustolang.EmptyTupleSymbol, kustolang.EmptyEntityGroupSymbol)
	}
	if kustolang.EmptyGraphSymbol.EdgeShape != kustolang.EmptyTableSymbol || kustolang.Error != kustolang.ErrorSymbolInstance || kustolang.Void != kustolang.VoidSymbolInstance {
		t.Fatalf("canonical identities = graph:%#v error:%#v void:%#v", kustolang.EmptyGraphSymbol, kustolang.Error, kustolang.Void)
	}
	if !kustolang.ErrorSymbolInstance.IsError() || kustolang.UnknownParameter == nil || kustolang.UnknownParameter.Name != "" {
		t.Fatalf("canonical error/parameter = error:%#v parameter:%#v", kustolang.ErrorSymbolInstance, kustolang.UnknownParameter)
	}
}

func TestTableSymbolWithDescriptonCompatibilityAlias(t *testing.T) {
	table := kustolang.NewTableSymbol("Events")
	described := table.WithDescripton("event records")
	if got := described.Description(); got != "event records" {
		t.Fatalf("WithDescripton description = %q, want %q", got, "event records")
	}
	if table.Description() != "" {
		t.Fatalf("WithDescripton mutated original: %#v", table)
	}
}

func TestPublicSymbolLookupAssignabilityAndEquivalence(t *testing.T) {
	account := kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type())
	count := kustolang.NewColumnSymbol("Count", kustolang.ScalarTypes.Long.Type())
	table := kustolang.NewTableSymbol("Events", account, count)
	if got := kustolang.FindFirstMember(table, "account", true); got != account {
		t.Fatalf("case-insensitive member = %#v", got)
	}
	if got := kustolang.FindFirstMemberMatch(table, "Count", kustolang.SymbolMatchColumn, false); got != count {
		t.Fatalf("matched member = %#v", got)
	}
	if !kustolang.SymbolAssignableTo(kustolang.ScalarTypes.Int.Type(), kustolang.ScalarTypes.Long.Type(), kustolang.ConversionPromotable) {
		t.Fatal("int was not promotable to long")
	}
	if !kustolang.SymbolAssignableToAny(kustolang.ScalarTypes.Int.Type(), []kustolang.Symbol{kustolang.ScalarTypes.String.Type(), kustolang.ScalarTypes.Real.Type()}, kustolang.ConversionPromotable) {
		t.Fatal("int was not assignable to any promotable target")
	}

	equivalent := kustolang.NewTableSymbol("Events",
		kustolang.NewColumnSymbol("Account", kustolang.ScalarTypes.String.Type()),
		kustolang.NewColumnSymbol("Count", kustolang.ScalarTypes.Long.Type()),
	)
	columnsOnly := equivalent.WithName("Other")
	if !kustolang.AreTablesEquivalent(table, equivalent) || !kustolang.AreTableResultsEquivalent(table, columnsOnly) || !kustolang.AreTableColumnsEquivalent(table, columnsOnly) {
		t.Fatalf("table equivalence = full:%t result:%t columns:%t", kustolang.AreTablesEquivalent(table, equivalent), kustolang.AreTableResultsEquivalent(table, columnsOnly), kustolang.AreTableColumnsEquivalent(table, columnsOnly))
	}
}

func TestPublicVariableConstantValueInfoIsTyped(t *testing.T) {
	typedInfo := kustolang.ParseLiteral("42").LiteralValueInfo()
	typed := kustolang.NewVariableSymbolWithValueInfo("Limit", kustolang.ScalarTypes.Long.Type(), true, typedInfo, nil)
	if typed.ConstantValueInfo != typedInfo || typed.ConstantValue() != int64(42) {
		t.Fatalf("typed variable constant = info:%#v value:%#v", typed.ConstantValueInfo, typed.ConstantValue())
	}

	compatibility := kustolang.NewVariableSymbol("Mode", kustolang.ScalarTypes.String.Type(), true, "fast", nil)
	if compatibility.ConstantValueInfo == nil || compatibility.ConstantValueInfo.ValueText != "fast" || compatibility.ConstantValue() != "fast" {
		t.Fatalf("compatibility variable constant = info:%#v value:%#v", compatibility.ConstantValueInfo, compatibility.ConstantValue())
	}
}

func TestPublicPatternSymbolNameOnlyConstruction(t *testing.T) {
	pattern := kustolang.NewNamedPatternSymbol("detections")
	if pattern.SymbolName() != "detections" || pattern.SymbolKind() != kustolang.SymbolKindPattern || pattern.SymbolTabularity() != kustolang.TabularityTabular {
		t.Fatalf("pattern metadata = name:%q kind:%s tabularity:%s", pattern.SymbolName(), pattern.SymbolKind(), pattern.SymbolTabularity())
	}
	if len(pattern.Parameters()) != 0 || pattern.PathParameter() != nil || len(pattern.Signatures()) != 0 {
		t.Fatalf("name-only pattern = parameters:%#v path:%#v signatures:%#v", pattern.Parameters(), pattern.PathParameter(), pattern.Signatures())
	}
}

func TestPublicPatternSignatureBodyOverloads(t *testing.T) {
	parameter := kustolang.NewParameter("source", kustolang.ScalarTypes.String.Type())
	pathParameter := kustolang.NewParameter("path", kustolang.ScalarTypes.String.Type())

	emptyText := kustolang.NewPatternSignature([]string{"SecurityEvent"}, "", "", true)
	textBody := kustolang.NewPatternSignature(nil, "", "{ print source }")
	declaration := kustolang.ParseFunctionBody("{ print source }")
	syntaxBody := kustolang.NewPatternSignatureFromFunctionBody(nil, "", declaration)
	absentBody := kustolang.NewPatternSignatureWithFunctionBody(nil, "", nil)
	pattern := kustolang.NewPatternSymbol(
		"detections",
		[]*kustolang.Parameter{parameter},
		pathParameter,
		[]*kustolang.PatternSignature{emptyText, textBody, syntaxBody, absentBody},
	)

	if !emptyText.HasPathValue() || emptyText.BodyText() != "" {
		t.Fatalf("empty text body = hasPath:%t body:%q", emptyText.HasPathValue(), emptyText.BodyText())
	}
	emptySignature := emptyText.Signature()
	if emptySignature == nil || emptySignature.Body != "" || emptySignature.Declaration != nil {
		t.Fatalf("empty text signature = %#v", emptySignature)
	}
	if textSignature := textBody.Signature(); textSignature == nil || textSignature.Body != "{ print source }" || textSignature.Declaration != nil {
		t.Fatalf("text signature = %#v", textSignature)
	}
	declarationSignature := syntaxBody.Signature()
	if declarationSignature == nil || declarationSignature.Declaration != declaration {
		t.Fatalf("declaration signature = %#v, want declaration %#v", declarationSignature, declaration)
	}
	if declarationSignature.Symbol != pattern || len(declarationSignature.Parameters) != 2 || declarationSignature.Parameters[0] != parameter || declarationSignature.Parameters[1] != pathParameter {
		t.Fatalf("declaration signature ownership = symbol:%#v parameters:%#v", declarationSignature.Symbol, declarationSignature.Parameters)
	}
	if signature := absentBody.Signature(); signature != nil {
		t.Fatalf("absent body signature = %#v, want nil", signature)
	}
}
