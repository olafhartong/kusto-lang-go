// Package kustolang is an incremental Go port of the public Microsoft
// Kusto.Language implementation.
//
// Portions of this package are ported from:
// https://github.com/microsoft/Kusto-Query-Language
// Upstream commit tracked in third_party/kusto/UPSTREAM.json.
//
// The upstream project is licensed under Apache-2.0. See
// third_party/kusto/LICENSE for the preserved license text.
//
// This low-level API is shared by the kustolang facade, environment schemas,
// and external consumers such as KustoKITT. Native parsing and syntax spans
// use UTF-8 bytes; editor services use UTF-16 positions. The higher-level
// kustolang facade converts document diagnostics at its public boundary.
package kustolang
