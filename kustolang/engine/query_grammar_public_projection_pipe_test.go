package kustolang

import "testing"

func TestQueryGrammarPublicPostPipeOperatorKinds(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name string
		text string
		want SyntaxKind
	}{
		{name: "serialize", text: "serialize Copy=Value", want: SyntaxKindSerializeOperator},
		{name: "sample", text: "sample 10", want: SyntaxKindSampleOperator},
		{name: "sample distinct", text: "sample-distinct 3 of User", want: SyntaxKindSampleDistinctOperator},
		{name: "get schema", text: "getschema kind=csl", want: SyntaxKindGetSchemaOperator},
		{name: "execute and cache", text: "__executeAndCache", want: SyntaxKindExecuteAndCacheOperator},
		{name: "project away", text: "project-away Account*", want: SyntaxKindProjectAwayOperator},
		{name: "project keep", text: "project-keep Account", want: SyntaxKindProjectKeepOperator},
		{name: "project rename", text: "project-rename Alias=Account", want: SyntaxKindProjectRenameOperator},
		{name: "project reorder", text: "project-reorder Account desc", want: SyntaxKindProjectReorderOperator},
		{name: "project by names", text: `project-by-names dynamic(["Account"])`, want: SyntaxKindProjectByNamesOperator},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			if operator.Kind() != test.want {
				t.Fatalf("operator %q kind = %v, want %v", test.text, operator.Kind(), test.want)
			}
			if diagnostics := operator.ContainedDiagnostics(); len(diagnostics) != 0 {
				t.Fatalf("operator %q diagnostics = %#v", test.text, diagnostics)
			}
		})
	}
}

func TestQueryGrammarPublicPostPipeMultiStageAST(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		"T | serialize Copy=Value | sample 10 | sample-distinct 3 of ['User Name'] | __executeAndCache | getschema kind=csl")

	getSchemaPipe, getSchemaNode := queryGrammarPipeStage(t, expression, SyntaxKindGetSchemaOperator)
	executePipe, executeNode := queryGrammarPipeStage(t, getSchemaPipe.Expression(), SyntaxKindExecuteAndCacheOperator)
	sampleDistinctPipe, sampleDistinctNode := queryGrammarPipeStage(t, executePipe.Expression(), SyntaxKindSampleDistinctOperator)
	samplePipe, sampleNode := queryGrammarPipeStage(t, sampleDistinctPipe.Expression(), SyntaxKindSampleOperator)
	serializePipe, serializeNode := queryGrammarPipeStage(t, samplePipe.Expression(), SyntaxKindSerializeOperator)
	if source, ok := serializePipe.Expression().(*NameReference); !ok || NameReferenceSimpleName(source) != "T" {
		t.Fatalf("pipe source = %#v, want T", serializePipe.Expression())
	}

	serialize := serializeNode.(*SerializeOperator)
	if serialize.Expressions().Count() != 1 || serialize.Expressions().At(0).Element().Kind() != SyntaxKindSimpleNamedExpression {
		t.Fatalf("serialize expressions = %#v, want one named expression", serialize.Expressions())
	}
	sample := sampleNode.(*SampleOperator)
	if sample.Expression().Kind() != SyntaxKindLongLiteralExpression {
		t.Fatalf("sample expression kind = %v, want long literal", sample.Expression().Kind())
	}
	sampleDistinct := sampleDistinctNode.(*SampleDistinctOperator)
	if sampleDistinct.Expression().Kind() != SyntaxKindLongLiteralExpression || sampleDistinct.OfKeyword().Kind() != SyntaxKindOfKeyword {
		t.Fatalf("sample-distinct count/of = %v/%v", sampleDistinct.Expression().Kind(), sampleDistinct.OfKeyword().Kind())
	}
	ofName, ok := sampleDistinct.OfExpression().(*NameReference)
	if !ok || ofName.Name().Kind() != SyntaxKindBracketedName || NameReferenceSimpleName(ofName) != "User Name" {
		t.Fatalf("sample-distinct of expression = %#v, want bracketed User Name", sampleDistinct.OfExpression())
	}
	if executeNode.(*ExecuteAndCacheOperator).ExecuteAndCacheKeyword().Kind() != SyntaxKindExecuteAndCacheKeyword {
		t.Fatalf("execute-and-cache keyword = %#v", executeNode)
	}
	getSchema := getSchemaNode.(*GetSchemaOperator)
	if getSchema.KindParameter() == nil || getSchema.KindParameter().Expression().Kind() != SyntaxKindTokenLiteralExpression {
		t.Fatalf("getschema kind parameter = %#v", getSchema.KindParameter())
	}
}

func TestQueryGrammarPublicProjectionOperatorsAST(t *testing.T) {
	expression := parseQueryGrammarProduction(t, QueryGrammarFrom(nil).Expression,
		`T | project-away Temp*, ['Debug Column'] | project-keep Account*, ['Display Name'] | project-rename ['Friendly Name']=['Display Name'] | project-reorder ['Friendly Name'] desc, Account* asc | project-by-names pack_array("Friendly Name", "Account")`)

	byNamesPipe, byNamesNode := queryGrammarPipeStage(t, expression, SyntaxKindProjectByNamesOperator)
	reorderPipe, reorderNode := queryGrammarPipeStage(t, byNamesPipe.Expression(), SyntaxKindProjectReorderOperator)
	renamePipe, renameNode := queryGrammarPipeStage(t, reorderPipe.Expression(), SyntaxKindProjectRenameOperator)
	keepPipe, keepNode := queryGrammarPipeStage(t, renamePipe.Expression(), SyntaxKindProjectKeepOperator)
	_, awayNode := queryGrammarPipeStage(t, keepPipe.Expression(), SyntaxKindProjectAwayOperator)

	away := awayNode.(*ProjectAwayOperator)
	assertQueryGrammarProjectionNameKinds(t, away.Expressions(), SyntaxKindWildcardedName, SyntaxKindBracketedName)
	keep := keepNode.(*ProjectKeepOperator)
	assertQueryGrammarProjectionNameKinds(t, keep.Expressions(), SyntaxKindWildcardedName, SyntaxKindBracketedName)

	rename := renameNode.(*ProjectRenameOperator)
	if rename.Expressions().Count() != 1 {
		t.Fatalf("project-rename expressions = %d, want 1", rename.Expressions().Count())
	}
	assignment, ok := rename.Expressions().At(0).Element().(*SimpleNamedExpression)
	if !ok || NameDeclarationSimpleName(assignment.Name()) != "Friendly Name" {
		t.Fatalf("project-rename assignment = %#v", rename.Expressions().At(0).Element())
	}
	renameSource, ok := assignment.Expression().(*NameReference)
	if !ok || renameSource.Name().Kind() != SyntaxKindBracketedName || NameReferenceSimpleName(renameSource) != "Display Name" {
		t.Fatalf("project-rename source = %#v", assignment.Expression())
	}

	reorder := reorderNode.(*ProjectReorderOperator)
	if reorder.Expressions().Count() != 2 {
		t.Fatalf("project-reorder expressions = %d, want 2", reorder.Expressions().Count())
	}
	for i, want := range []SyntaxKind{SyntaxKindDescKeyword, SyntaxKindAscKeyword} {
		ordered, ok := reorder.Expressions().At(i).Element().(*OrderedExpression)
		if !ok || ordered.Ordering().AscOrDescKeyword().Kind() != want || ordered.Ordering().NullsClause() != nil {
			t.Fatalf("project-reorder expression %d = %#v, want ordering %v without nulls clause", i, reorder.Expressions().At(i).Element(), want)
		}
	}
	if wildcard := reorder.Expressions().At(1).Element().(*OrderedExpression).Expression().(*NameReference); wildcard.Name().Kind() != SyntaxKindWildcardedName {
		t.Fatalf("project-reorder wildcard kind = %v", wildcard.Name().Kind())
	}

	byNames := byNamesNode.(*ProjectByNamesOperator)
	if byNames.Expressions().Count() != 1 || byNames.Expressions().At(0).Element().Kind() != SyntaxKindFunctionCallExpression {
		t.Fatalf("project-by-names expressions = %#v, want one function call", byNames.Expressions())
	}
}

func TestQueryGrammarPublicPostPipeRecovery(t *testing.T) {
	grammar := QueryGrammarFrom(nil)
	tests := []struct {
		name  string
		text  string
		kind  SyntaxKind
		codes []string
	}{
		{name: "serialize trailing comma", text: "serialize Copy=Value,", kind: SyntaxKindSerializeOperator, codes: []string{"KS006"}},
		{name: "sample missing expression", text: "sample", kind: SyntaxKindSampleOperator, codes: []string{"KS006"}},
		{name: "sample distinct missing of operand", text: "sample-distinct 3", kind: SyntaxKindSampleDistinctOperator, codes: []string{"KS005", "KS006"}},
		{name: "getschema missing kind value", text: "getschema kind=", kind: SyntaxKindGetSchemaOperator, codes: []string{"KS005"}},
		{name: "project away trailing comma", text: "project-away Account,", kind: SyntaxKindProjectAwayOperator, codes: []string{"KS006"}},
		{name: "project keep leading comma", text: "project-keep ,Account", kind: SyntaxKindProjectKeepOperator, codes: []string{"KS006"}},
		{name: "project rename missing value", text: "project-rename Alias=", kind: SyntaxKindProjectRenameOperator, codes: []string{"KS006"}},
		{name: "project reorder trailing comma", text: "project-reorder Account desc,", kind: SyntaxKindProjectReorderOperator, codes: []string{"KS006"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator := parseQueryGrammarProduction(t, grammar.QueryOperator, test.text)
			if operator.Kind() != test.kind {
				t.Fatalf("operator kind = %v, want %v", operator.Kind(), test.kind)
			}
			diagnostics := operator.ContainedDiagnostics()
			for _, code := range test.codes {
				if !diagnosticsContain(diagnostics, code) {
					t.Fatalf("operator %q missing %s diagnostic: %#v", test.text, code, diagnostics)
				}
			}
		})
	}
	emptyByNames := parseQueryGrammarProduction(t, grammar.QueryOperator, "project-by-names").(*ProjectByNamesOperator)
	if emptyByNames.Expressions().Count() != 0 || len(emptyByNames.ContainedDiagnostics()) != 0 {
		t.Fatalf("empty project-by-names = count:%d diagnostics:%#v", emptyByNames.Expressions().Count(), emptyByNames.ContainedDiagnostics())
	}

	assertQueryGrammarProductionLeavesTail(t, grammar.QueryOperator, "__executeAndCache Extra", 1)
}

func assertQueryGrammarProjectionNameKinds(t *testing.T, expressions *SeparatedSyntaxListOf[Expression], want ...SyntaxKind) {
	t.Helper()
	if expressions.Count() != len(want) {
		t.Fatalf("projection expressions = %d, want %d", expressions.Count(), len(want))
	}
	for i, kind := range want {
		reference, ok := expressions.At(i).Element().(*NameReference)
		if !ok || reference.Name().Kind() != kind {
			t.Fatalf("projection expression %d = %#v, want name kind %v", i, expressions.At(i).Element(), kind)
		}
	}
}
