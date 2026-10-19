package kustolang

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// NewComputedSignatureInferred creates a computed signature whose tabularity
// can be resolved with Signature.ComputeTabularity.
func NewComputedSignatureInferred(body string, parameters ...*Parameter) *Signature {
	return internal.NewComputedSignatureInferred(body, parameters...)
}

// NewComputedFunctionSymbolInferred creates the source-equivalent computed-body
// function form that does not require an explicit tabularity.
func NewComputedFunctionSymbolInferred(name, body string, parameters ...*Parameter) *FunctionSymbol {
	return internal.NewComputedFunctionSymbolInferred(name, body, parameters...)
}
