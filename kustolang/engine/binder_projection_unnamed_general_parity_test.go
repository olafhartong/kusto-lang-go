package kustolang

import "testing"

type next194ProjectionProfile struct {
	name      string
	parse     func(string) *QueryBlock
	errorType *TypeSymbol
}

func next194ProjectionProfiles() []next194ProjectionProfile {
	return []next194ProjectionProfile{{name: "current", parse: ParseQuery, errorType: ScalarTypes.Unknown.Type()}}
}

func next194ProjectionGlobals(columns ...*ColumnSymbol) (*GlobalState, *TableSymbol) {
	if len(columns) == 0 {
		columns = []*ColumnSymbol{
			NewColumnSymbol("A", ScalarTypes.String.Type()),
			NewColumnSymbol("N", ScalarTypes.Long.Type()),
			NewColumnSymbol("D", ScalarTypes.Dynamic.Type()),
		}
	}
	table := NewTableSymbol("T", columns...)
	return NewGlobalState(NewDatabaseSymbol("db", table)), table
}

func next194Bind(t *testing.T, profile next194ProjectionProfile, globals *GlobalState, text string) *QueryBlock {
	t.Helper()
	query := profile.parse(text)
	NewBinder(globals).TryBind(query)
	return query
}

func next194ProjectionTable(t *testing.T, query SyntaxElement, operator string) *TableSymbol {
	t.Helper()
	switch operator {
	case "project":
		return semanticTable(t, firstBoundNode[*ProjectOperator](t, query))
	case "extend":
		return semanticTable(t, firstBoundNode[*ExtendOperator](t, query))
	case "serialize":
		return semanticTable(t, firstBoundNode[*SerializeOperator](t, query))
	default:
		t.Fatalf("unknown projection operator %q", operator)
		return nil
	}
}

func next194AssertType(t *testing.T, table *TableSymbol, name string, want *TypeSymbol) {
	t.Helper()
	column := table.GetColumn(name)
	if column == nil || column.Type != want {
		t.Fatalf("column %q = %#v, want type %s", name, column, typeName(want))
	}
}

func next194AssertContainedDiagnostic(t *testing.T, root SyntaxElement, code string) {
	t.Helper()
	if !diagnosticsContain(GetContainedDiagnostics(root, DiagnosticsIncludeSemantic), code) {
		t.Fatalf("semantic diagnostics = %#v, want %s", GetContainedDiagnostics(root, DiagnosticsIncludeSemantic), code)
	}
}

func TestNext194UnnamedProjectionCoreProfileParity(t *testing.T) {
	for _, profile := range next194ProjectionProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			globals, _ := next194ProjectionGlobals()

			t.Run("project function name", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | project strlen(A)")
				result := next194ProjectionTable(t, query, "project")
				assertOrderedColumns(t, result, "strlen_A")
				next194AssertType(t, result, "strlen_A", ScalarTypes.Long.Type())
			})

			t.Run("path error name", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | project A.foo")
				result := next194ProjectionTable(t, query, "project")
				assertOrderedColumns(t, result, "foo")
				next194AssertType(t, result, "foo", profile.errorType)
				assertSemanticDiagnostic(t, firstNamedReference(t, query, "foo"), "KS142")
			})

			t.Run("element error name", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | extend A[0]")
				result := next194ProjectionTable(t, query, "extend")
				assertOrderedColumns(t, result, "A", "N", "D", "0")
				next194AssertType(t, result, "0", profile.errorType)
				next194AssertContainedDiagnostic(t, query, "KS147")
			})

			t.Run("missing function default", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | project missing_function(A)")
				result := next194ProjectionTable(t, query, "project")
				assertOrderedColumns(t, result, "Column1")
				next194AssertType(t, result, "Column1", profile.errorType)
				assertSemanticDiagnostic(t, firstNamedReference(t, query, "missing_function"), "KS211")
			})

			t.Run("invalid addition keeps inferred result", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | extend A + N")
				result := next194ProjectionTable(t, query, "extend")
				assertOrderedColumns(t, result, "A", "N", "D", "Column1")
				next194AssertType(t, result, "Column1", ScalarTypes.Long.Type())
				assertSemanticDiagnostic(t, firstBoundNode[*BinaryExpression](t, query), "KS106")
			})

			t.Run("parenthesized and literal defaults", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | project (N + 1), 1")
				result := next194ProjectionTable(t, query, "project")
				assertOrderedColumns(t, result, "Column1", "Column2")
				next194AssertType(t, result, "Column1", ScalarTypes.Long.Type())
				next194AssertType(t, result, "Column2", ScalarTypes.Long.Type())
			})

			t.Run("serialize errors", func(t *testing.T) {
				query := next194Bind(t, profile, globals, "T | serialize A[0], missing_function(A)")
				serialize := firstBoundNode[*SerializeOperator](t, query)
				result := semanticTable(t, serialize)
				assertOrderedColumns(t, result, "A", "N", "D", "0", "Column1")
				next194AssertType(t, result, "0", profile.errorType)
				next194AssertType(t, result, "Column1", profile.errorType)
				if !result.IsSerialized() {
					t.Fatal("serialize result lost serialized state")
				}
				next194AssertContainedDiagnostic(t, query, "KS147")
				next194AssertContainedDiagnostic(t, query, "KS211")
			})
		})
	}
}

func TestNext194UnnamedProjectionMultipleNamesAndReplacementParity(t *testing.T) {
	for _, profile := range next194ProjectionProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			globals, _ := next194ProjectionGlobals()
			expressions := "strlen(A), strlen(A), A.foo, A[0], (N + 1), 1"

			project := next194Bind(t, profile, globals, "T | project "+expressions)
			assertOrderedColumns(t, next194ProjectionTable(t, project, "project"),
				"strlen_A", "strlen_A1", "foo", "0", "Column1", "Column2")

			extend := next194Bind(t, profile, globals, "T | extend "+expressions)
			assertOrderedColumns(t, next194ProjectionTable(t, extend, "extend"),
				"A", "N", "D", "strlen_A", "foo", "0", "Column1", "Column2")

			serialize := next194Bind(t, profile, globals, "T | serialize "+expressions)
			assertOrderedColumns(t, next194ProjectionTable(t, serialize, "serialize"),
				"A", "N", "D", "strlen_A", "strlen_A1", "foo", "0", "Column1", "Column2")

			for _, operator := range []string{"project", "extend", "serialize"} {
				t.Run(operator+" repeated missing", func(t *testing.T) {
					query := next194Bind(t, profile, globals, "T | "+operator+" missing_function(A), missing_function(A)")
					result := next194ProjectionTable(t, query, operator)
					if operator == "project" {
						assertOrderedColumns(t, result, "Column1", "Column2")
					} else {
						assertOrderedColumns(t, result, "A", "N", "D", "Column1", "Column2")
					}
					next194AssertType(t, result, "Column1", profile.errorType)
					next194AssertType(t, result, "Column2", profile.errorType)
				})
			}
		})
	}
}

func TestNext194UnnamedProjectionSuffixScopeAndCollisions(t *testing.T) {
	for _, profile := range next194ProjectionProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			globals, table := next194ProjectionGlobals()

			pipeline := next194Bind(t, profile, globals,
				"T | project 1 | serialize 2 | mv-expand dynamic([1]) | distinct 3 | summarize 4 | extend 5")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ProjectOperator](t, pipeline)), "Column1")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*SerializeOperator](t, pipeline)), "Column1", "Column2")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*MvExpandOperator](t, pipeline)), "Column1", "Column2", "Column3")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*DistinctOperator](t, pipeline)), "Column3")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*SummarizeOperator](t, pipeline)), "Column4")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*ExtendOperator](t, pipeline)), "Column4", "Column5")

			collisionPipeline := next194Bind(t, profile, globals,
				"T | project Column1=1, 2 | serialize Column3=3, 4 | distinct Column5=5, 6")
			collisionProjects := boundNodes[*ProjectOperator](collisionPipeline)
			assertOrderedColumns(t, semanticTable(t, collisionProjects[0]), "Column1", "Column2")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*SerializeOperator](t, collisionPipeline)),
				"Column1", "Column2", "Column3", "Column4")
			assertOrderedColumns(t, semanticTable(t, firstBoundNode[*DistinctOperator](t, collisionPipeline)), "Column5", "Column6")

			project := next194Bind(t, profile, globals, "T | project Column1=N, 1")
			assertOrderedColumns(t, next194ProjectionTable(t, project, "project"), "Column1", "Column2")

			rowColumn1 := NewColumnSymbol("Column1", ScalarTypes.String.Type())
			rowGlobals, _ := next194ProjectionGlobals(rowColumn1, NewColumnSymbol("N", ScalarTypes.Long.Type()))
			rowProject := next194Bind(t, profile, rowGlobals, "T | project 1")
			assertOrderedColumns(t, next194ProjectionTable(t, rowProject, "project"), "Column1")
			rowExtend := next194Bind(t, profile, rowGlobals, "T | extend 1")
			assertOrderedColumns(t, next194ProjectionTable(t, rowExtend, "extend"), "Column1", "N", "Column2")

			globalCollision := NewTableSymbol("Column1", NewColumnSymbol("X", ScalarTypes.Long.Type()))
			collisionGlobals := NewGlobalState(NewDatabaseSymbol("db", table, globalCollision))
			globalProject := next194Bind(t, profile, collisionGlobals, "T | project 1")
			assertOrderedColumns(t, next194ProjectionTable(t, globalProject, "project"), "Column2")
		})
	}
}

func TestNext194UnnamedProjectionTabularAndRecoveryParity(t *testing.T) {
	for _, profile := range next194ProjectionProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			globals, _ := next194ProjectionGlobals()

			tabular := next194Bind(t, profile, globals, "T | project datatable(X:long)[1]")
			project := firstBoundNode[*ProjectOperator](t, tabular)
			if !project.IsBound() {
				t.Fatal("tabular unnamed projection rolled back")
			}
			result := semanticTable(t, project)
			assertOrderedColumns(t, result, "Column1")
			next194AssertType(t, result, "Column1", ScalarTypes.Unknown.Type())
			assertDiagnosticCodes(t, project, "KS108")

			malformed := next194Bind(t, profile, globals, "T | project strlen(A")
			malformedProject := firstBoundNode[*ProjectOperator](t, malformed)
			if !malformedProject.IsBound() {
				t.Fatal("parser-recovered missing function close rolled back")
			}
			malformedResult := semanticTable(t, malformedProject)
			assertOrderedColumns(t, malformedResult, "strlen_A")
			next194AssertType(t, malformedResult, "strlen_A", ScalarTypes.Long.Type())
			diagnostics := GetContainedDiagnostics(malformed, DiagnosticsIncludeSyntactic)
			if len(diagnostics) != 1 || diagnostics[0].Code != "KS005" {
				t.Fatalf("recovery diagnostics = %#v, want one KS005", diagnostics)
			}
		})
	}
}

func TestNext194UnnamedProjectionDownstreamAndDirectIdentityParity(t *testing.T) {
	for _, profile := range next194ProjectionProfiles() {
		t.Run(profile.name, func(t *testing.T) {
			a := NewColumnSymbol("A", ScalarTypes.String.Type())
			n := NewColumnSymbol("N", ScalarTypes.Long.Type())
			globals, _ := next194ProjectionGlobals(a, n, NewColumnSymbol("D", ScalarTypes.Dynamic.Type()))

			direct := next194Bind(t, profile, globals, "T | project A, N")
			directResult := next194ProjectionTable(t, direct, "project")
			if directResult.GetColumn("A") != a || directResult.GetColumn("N") != n {
				t.Fatal("direct projection did not preserve source column identities")
			}

			query := next194Bind(t, profile, globals,
				"T | project strlen(A), missing_function(A) | project strlen_A, Column1")
			projects := boundNodes[*ProjectOperator](query)
			if len(projects) != 2 || !projects[0].IsBound() || !projects[1].IsBound() {
				t.Fatalf("project chain binding = %#v", projects)
			}
			upstream := semanticTable(t, projects[0])
			downstream := semanticTable(t, projects[1])
			for _, name := range []string{"strlen_A", "Column1"} {
				if downstream.GetColumn(name) != upstream.GetColumn(name) || firstNamedReference(t, projects[1], name).ReferencedSymbol() != upstream.GetColumn(name) {
					t.Fatalf("downstream %s did not preserve committed column identity", name)
				}
			}
		})
	}
}
