package kustolang

// Ported from src/Kusto.Language/Symbols/ParameterLayout.cs and
// src/Kusto.Language/Symbols/ParameterLayouts.cs at
// c94c9e761b3d095e2b9c21e3ec3e92afa0af4856.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

import internal "github.com/olafhartong/kusto-lang-go/kustolang/engine"

// ParameterLayoutDefinition is the object-form layout API. ParameterLayout is
// retained as the comparable enum stored by Signature for compatibility.
type ParameterLayoutDefinition = internal.ParameterLayoutDefinition
type ParameterLayoutStrategy = internal.ParameterLayoutStrategy
type ParameterLayoutsAPI = internal.ParameterLayoutsAPI

// ParameterLayouts mirrors the upstream static layout catalog.
var ParameterLayouts = internal.ParameterLayouts
