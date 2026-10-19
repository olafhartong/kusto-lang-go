package kustolang

import "testing"

func TestParameterDeclarationFormatsScalarAndSpecialKinds(t *testing.T) {
	tests := []struct {
		name      string
		parameter *Parameter
		want      string
	}{
		{
			name:      "scalar",
			parameter: NewParameter("Account", ScalarTypes.String.Type()),
			want:      "Account: string",
		},
		{
			name:      "bracketed name",
			parameter: NewParameter("bad name", ScalarTypes.Long.Type()),
			want:      "['bad name']: long",
		},
		{
			name:      "tabular kind",
			parameter: NewParameterKind("rows", ParameterTypeTabular),
			want:      "rows: (*)",
		},
		{
			name:      "dependent kind",
			parameter: NewParameterKind("value", ParameterTypeScalar),
			want:      "value: dynamic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParameterDeclaration(tt.parameter); got != tt.want {
				t.Fatalf("ParameterDeclaration = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParameterDeclarationFormatsTableSchema(t *testing.T) {
	table := NewTableSymbol("T",
		NewColumnSymbol("Account", ScalarTypes.String.Type()),
		NewColumnSymbol("bad name", ScalarTypes.Long.Type()),
	)

	if got, want := ParameterDeclaration(NewParameterSymbolType("T", table)), "T: (Account: string, ['bad name']: long)"; got != want {
		t.Fatalf("table parameter declaration = %q, want %q", got, want)
	}

	empty := NewTableSymbol("Empty")
	if got, want := ParameterDeclaration(NewParameterSymbolType("Rows", empty)), "Rows: (*)"; got != want {
		t.Fatalf("empty table parameter declaration = %q, want %q", got, want)
	}
}

func TestParameterListDeclarationAndDefaultValue(t *testing.T) {
	value := ParseExpression(`"guest"`)
	parameter := NewParameter("Account", ScalarTypes.String.Type(), WithParameterDefaultValue(value))

	if !parameter.IsOptional() || parameter.IsRepeatable() || parameter.MinOccurring != 0 || parameter.MaxOccurring != 1 {
		t.Fatalf("parameter with default value occurrence = %d..%d repeat=%t", parameter.MinOccurring, parameter.MaxOccurring, parameter.IsRepeatable())
	}
	if got, want := ParameterDeclaration(parameter), `Account: string = "guest"`; got != want {
		t.Fatalf("default parameter declaration = %q, want %q", got, want)
	}
	defaultThenOccurrence := NewParameter("Account", ScalarTypes.String.Type(),
		WithParameterDefaultValue(value),
		WithParameterOccurrence(2, MaxRepeat),
	)
	if defaultThenOccurrence.MinOccurring != 0 || defaultThenOccurrence.MaxOccurring != 1 {
		t.Fatalf("default value should override later occurrence bounds, got %d..%d", defaultThenOccurrence.MinOccurring, defaultThenOccurrence.MaxOccurring)
	}
	occurrenceThenDefault := NewParameterKind("Account", ParameterTypeScalar,
		WithParameterOccurrence(2, MaxRepeat),
		WithParameterDefaultValue(value),
	)
	if occurrenceThenDefault.MinOccurring != 0 || occurrenceThenDefault.MaxOccurring != 1 {
		t.Fatalf("default value should override earlier occurrence bounds, got %d..%d", occurrenceThenDefault.MinOccurring, occurrenceThenDefault.MaxOccurring)
	}

	parameters := []*Parameter{
		NewParameter("Account", ScalarTypes.String.Type()),
		NewParameterKind("Rows", ParameterTypeTabular),
	}
	if got, want := ParameterListDeclaration(parameters), "(Account: string, Rows: (*))"; got != want {
		t.Fatalf("ParameterListDeclaration = %q, want %q", got, want)
	}
}

func TestNewParameterFromSymbolPreservesMetadata(t *testing.T) {
	symbol := NewParameterSymbol("TenantId", ScalarTypes.String.Type(), "Tenant identifier")
	parameter := NewParameterFromSymbol(symbol, true, nil)

	if parameter.Name != "TenantId" || parameter.Description != "Tenant identifier" {
		t.Fatalf("parameter metadata = name:%q desc:%q", parameter.Name, parameter.Description)
	}
	if !parameter.IsOptional() {
		t.Fatalf("optional parameter should have MinOccurring 0")
	}
	if got, want := ParameterDeclaration(parameter), "TenantId: string"; got != want {
		t.Fatalf("symbol parameter declaration = %q, want %q", got, want)
	}
}

func TestParameterTabularityAndTypeDependencyHelpers(t *testing.T) {
	scalar := NewParameter("value", ScalarTypes.Long.Type())
	table := NewParameterSymbolType("rows", NewTableSymbol("T", NewColumnSymbol("x", ScalarTypes.Long.Type())))
	tabularKind := NewParameterKind("rows", ParameterTypeTabular)

	if scalar.TypeDependsOnArguments() || !scalar.IsScalar() || scalar.IsTabular() {
		t.Fatalf("scalar helpers = depends:%t scalar:%t tabular:%t", scalar.TypeDependsOnArguments(), scalar.IsScalar(), scalar.IsTabular())
	}
	if !table.TypeDependsOnArguments() || table.Tabularity() != TabularityTabular || table.IsScalar() || !table.IsTabular() {
		t.Fatalf("table helpers = depends:%t tabularity:%s scalar:%t tabular:%t", table.TypeDependsOnArguments(), table.Tabularity(), table.IsScalar(), table.IsTabular())
	}
	if !tabularKind.TypeDependsOnArguments() || tabularKind.Tabularity() != TabularityTabular {
		t.Fatalf("tabular-kind helpers = depends:%t tabularity:%s", tabularKind.TypeDependsOnArguments(), tabularKind.Tabularity())
	}
}

func TestSchemaDisplayUsesDeclaredTableParameter(t *testing.T) {
	fn := NewFunctionSymbol("f",
		NewSymbolSignature(NewTableSymbol("Out", NewColumnSymbol("x", ScalarTypes.Long.Type())),
			NewParameterSymbolType("input", NewTableSymbol("In", NewColumnSymbol("Account", ScalarTypes.String.Type()))),
		),
	)

	if got, want := SchemaDisplayText(fn), "(input: (Account: string))"; got != want {
		t.Fatalf("SchemaDisplayText(function table parameter) = %q, want %q", got, want)
	}
}

func TestParseParameterListConvertsFunctionParameters(t *testing.T) {
	parameters := ParseParameterList(`(User:string="admin", Count:int64, Rows:(Account:string, ['bad name']:real), AnyRows:(*))`)
	if len(parameters) != 4 {
		t.Fatalf("parameter count = %d", len(parameters))
	}

	if got, want := ParameterDeclaration(parameters[0]), `User: string = "admin"`; got != want {
		t.Fatalf("parameter 0 declaration = %q, want %q", got, want)
	}
	if !parameters[0].IsOptional() || parameters[0].DefaultValue == nil {
		t.Fatalf("parameter 0 optional/default = %t/%#v", parameters[0].IsOptional(), parameters[0].DefaultValue)
	}
	if parameters[1].DeclaredTypes[0] != ScalarTypes.Long.Type() {
		t.Fatalf("int64 alias declared type = %#v, want long", parameters[1].DeclaredTypes[0])
	}
	if got, want := ParameterDeclaration(parameters[2]), "Rows: (Account: string, ['bad name']: real)"; got != want {
		t.Fatalf("parameter 2 declaration = %q, want %q", got, want)
	}
	if parameters[2].Tabularity() != TabularityTabular || !parameters[2].TypeDependsOnArguments() {
		t.Fatalf("parameter 2 tabularity/depends = %s/%t", parameters[2].Tabularity(), parameters[2].TypeDependsOnArguments())
	}
	if got, want := ParameterDeclaration(parameters[3]), "AnyRows: (*)"; got != want {
		t.Fatalf("parameter 3 declaration = %q, want %q", got, want)
	}
}

func TestNewParameterFromFunctionParameterConvertsSingleParameter(t *testing.T) {
	syntax := ParseFunctionParameter(`['tenant id']:guid=00000000-0000-0000-0000-000000000001`)
	parameter := NewParameterFromFunctionParameter(syntax)

	if parameter.Name != "tenant id" || parameter.DeclaredTypes[0] != ScalarTypes.Guid.Type() {
		t.Fatalf("single parameter metadata = name:%q type:%#v", parameter.Name, parameter.DeclaredTypes[0])
	}
	if got, want := ParameterDeclaration(parameter), "['tenant id']: guid = 00000000-0000-0000-0000-000000000001"; got != want {
		t.Fatalf("single parameter declaration = %q, want %q", got, want)
	}
}

func TestParameterFromFunctionParameterBindsInvalidTypeToError(t *testing.T) {
	syntax := ParseFunctionParameter(`value:notatype`)
	parameter := NewParameterFromFunctionParameter(syntax)

	if parameter.DeclaredTypes[0] != &Error.TypeSymbol {
		t.Fatalf("invalid type declared symbol = %#v, want error", parameter.DeclaredTypes[0])
	}
	if got, want := ParameterDeclaration(parameter), "value: error"; got != want {
		t.Fatalf("invalid type declaration = %q, want %q", got, want)
	}
}
