package kustolang

import (
	"sort"
	"strings"
)

// Ported from:
//   src/Kusto.Language/Editor/DisabledDiagnostics.cs
//   src/Kusto.Language/Editor/QuickInfoOptions.cs
//   src/Kusto.Language/Editor/Kusto/Analyzers/KustoAnalyzerCodes.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

const (
	KustoAnalyzerCodeAvoidUsingContains                   = "KS500"
	KustoAnalyzerCodeAvoidUsingIsNullStringComparision    = "KS501"
	KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison  = "KS502"
	KustoAnalyzerCodeAvoidUsingShortStringComparision     = "KS503"
	KustoAnalyzerCodeAvoidUsingToBoolOnNumerics           = "KS504"
	KustoAnalyzerCodeNullAggregation                      = "KS505"
	KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate  = "KS506"
	KustoAnalyzerCodeAvoidUsingObsoleteFunctions          = "KS507"
	KustoAnalyzerCodeAvoidJoinWithoutKind                 = "KS508"
	KustoAnalyzerCodeStdevTimespanConversion              = "KS509"
	KustoAnalyzerCodeColumnHasSameNameAsVariable          = "KS510"
	KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic = "KS511"
	KustoAnalyzerCodeCalledFunctionHasErrors              = "KS512"
	KustoAnalyzerCodeAvoidUsingLegacyPartition            = "KS513"
	KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings         = "KS514"
	KustoAnalyzerCodePreferUsingOptimizedAlternative      = "KS515"
	KustoAnalyzerCodeAvoidStrlenWithDynamic               = "KS516"
)

const disabledDiagnosticsSeverityPrefix = "sev:"

type DisabledDiagnostics struct {
	severities map[string]struct{}
	codes      map[string]struct{}
}

var DisabledDiagnosticsDefault = DisabledDiagnostics{}

func NewDisabledDiagnostics(text ...string) DisabledDiagnostics {
	if len(text) == 0 || text[0] == "" {
		return DisabledDiagnosticsDefault
	}

	severities := map[string]struct{}{}
	codes := map[string]struct{}{}
	for _, raw := range strings.Split(text[0], ",") {
		if raw == "" {
			continue
		}
		item := strings.TrimSpace(raw)
		if strings.HasPrefix(item, disabledDiagnosticsSeverityPrefix) {
			severities[item[len(disabledDiagnosticsSeverityPrefix):]] = struct{}{}
		} else {
			codes[item] = struct{}{}
		}
	}
	return DisabledDiagnostics{severities: severities, codes: codes}
}

func ParseDisabledDiagnostics(text string) DisabledDiagnostics {
	return NewDisabledDiagnostics(text)
}

func (d DisabledDiagnostics) String() string {
	if d.severities == nil || d.codes == nil {
		return ""
	}

	severities := make([]string, 0, len(d.severities))
	for severity := range d.severities {
		severities = append(severities, disabledDiagnosticsSeverityPrefix+severity)
	}
	sort.Strings(severities)

	codes := make([]string, 0, len(d.codes))
	for code := range d.codes {
		codes = append(codes, code)
	}
	sort.Strings(codes)

	items := append(severities, codes...)
	return strings.Join(items, ",")
}

func (d DisabledDiagnostics) IsSeverityEnabled(severity string) bool {
	return !d.IsServerityDisabled(severity)
}

func (d DisabledDiagnostics) IsServerityDisabled(severity string) bool {
	return d.severities != nil && setContains(d.severities, severity)
}

func (d DisabledDiagnostics) IsSeverityDisabled(severity string) bool {
	return d.IsServerityDisabled(severity)
}

func (d DisabledDiagnostics) IsCodeEnabled(code string) bool {
	return !d.IsCodeDisabled(code)
}

func (d DisabledDiagnostics) IsCodeDisabled(code string) bool {
	return d.codes != nil && setContains(d.codes, code)
}

func (d DisabledDiagnostics) IsDiagnosticEnabled(diagnostic Diagnostic) bool {
	return !d.IsDiagnosticDisabled(diagnostic)
}

func (d DisabledDiagnostics) IsDiagnosticDisabled(diagnostic Diagnostic) bool {
	return d.IsServerityDisabled(diagnostic.Severity) || d.IsCodeDisabled(diagnostic.Code)
}

func (d DisabledDiagnostics) ToBuilder() *DisabledDiagnosticsBuilder {
	return NewDisabledDiagnosticsBuilder(d)
}

func CanDisableSeverity(severity string) bool {
	return severity == DiagnosticSeverityWarning ||
		severity == DiagnosticSeveritySuggestion ||
		severity == DiagnosticSeverityInformation
}

var kustoAnalyzerCodesThatCanBeDisabled = map[string]struct{}{
	KustoAnalyzerCodeAvoidUsingContains:                   {},
	KustoAnalyzerCodeAvoidUsingIsNullStringComparision:    {},
	KustoAnalyzerCodeAvoidUsingIsNotNullStringComparison:  {},
	KustoAnalyzerCodeAvoidUsingToBoolOnNumerics:           {},
	KustoAnalyzerCodeNullAggregation:                      {},
	KustoAnalyzerCodeAvoidUsingFormatDateTimeInPredicate:  {},
	KustoAnalyzerCodeAvoidUsingObsoleteFunctions:          {},
	KustoAnalyzerCodeAvoidJoinWithoutKind:                 {},
	KustoAnalyzerCodeStdevTimespanConversion:              {},
	KustoAnalyzerCodeColumnHasSameNameAsVariable:          {},
	KustoAnalyzerCodePreferUsingMaterializedViewIntrinsic: {},
	KustoAnalyzerCodeCalledFunctionHasErrors:              {},
	KustoAnalyzerCodeAvoidUsingLegacyPartition:            {},
	KustoAnalyzerCodeAvoidUsingHasWithIPv4Strings:         {},
	KustoAnalyzerCodePreferUsingOptimizedAlternative:      {},
	KustoAnalyzerCodeAvoidStrlenWithDynamic:               {},
}

func CanDisableCode(code string) bool {
	return setContains(kustoAnalyzerCodesThatCanBeDisabled, code)
}

func (d DisabledDiagnostics) WithCodeEnabled(code string, enabled bool) DisabledDiagnostics {
	builder := d.ToBuilder()
	builder.SetCodeEnabled(code, enabled)
	return builder.ToDisabledDiagnostics()
}

func (d DisabledDiagnostics) WithSeverityEnabled(severity string, enabled bool) DisabledDiagnostics {
	builder := d.ToBuilder()
	builder.SetSeverityEnabled(severity, enabled)
	return builder.ToDisabledDiagnostics()
}

type DisabledDiagnosticsBuilder struct {
	source     *DisabledDiagnostics
	severities map[string]struct{}
	codes      map[string]struct{}
}

func NewDisabledDiagnosticsBuilder(source ...DisabledDiagnostics) *DisabledDiagnosticsBuilder {
	builder := &DisabledDiagnosticsBuilder{}
	if len(source) > 0 {
		value := source[0]
		builder.source = &value
	}
	return builder
}

func (b *DisabledDiagnosticsBuilder) ToDisabledDiagnostics() DisabledDiagnostics {
	if b == nil {
		return DisabledDiagnosticsDefault
	}
	if b.source != nil {
		return *b.source
	}
	result := DisabledDiagnostics{
		severities: copyStringSet(b.severities),
		codes:      copyStringSet(b.codes),
	}
	b.source = &result
	b.severities = nil
	b.codes = nil
	return result
}

func (b *DisabledDiagnosticsBuilder) String() string {
	return b.ToDisabledDiagnostics().String()
}

func (b *DisabledDiagnosticsBuilder) tryInit() {
	if b.severities == nil {
		if b.source != nil && b.source.severities != nil {
			b.severities = copyStringSet(b.source.severities)
		} else {
			b.severities = map[string]struct{}{}
		}
	}
	if b.codes == nil {
		if b.source != nil && b.source.codes != nil {
			b.codes = copyStringSet(b.source.codes)
		} else {
			b.codes = map[string]struct{}{}
		}
	}
	b.source = nil
}

func (b *DisabledDiagnosticsBuilder) SetSeverityEnabled(severity string, enabled bool) {
	if b == nil {
		return
	}
	b.tryInit()
	if enabled {
		delete(b.severities, severity)
	} else if CanDisableSeverity(severity) {
		b.severities[severity] = struct{}{}
	}
}

func (b *DisabledDiagnosticsBuilder) SetCodeEnabled(code string, enabled bool) {
	if b == nil {
		return
	}
	b.tryInit()
	if enabled {
		delete(b.codes, code)
	} else if CanDisableCode(code) {
		b.codes[code] = struct{}{}
	}
}

func copyStringSet(source map[string]struct{}) map[string]struct{} {
	if source == nil {
		return nil
	}
	copied := make(map[string]struct{}, len(source))
	for key := range source {
		copied[key] = struct{}{}
	}
	return copied
}

func setContains(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}

type QuickInfoOptions struct {
	ShowDescriptions bool
	ShowDiagnostics  bool
	DiagnosticFilter DisabledDiagnostics
}

var QuickInfoOptionsDefault = NewQuickInfoOptions(true, true, nil)

func NewQuickInfoOptions(showDescriptions bool, showDiagnostics bool, filter *DisabledDiagnostics) QuickInfoOptions {
	diagnosticFilter := DisabledDiagnosticsDefault
	if filter != nil {
		diagnosticFilter = *filter
	}
	return QuickInfoOptions{
		ShowDescriptions: showDescriptions,
		ShowDiagnostics:  showDiagnostics,
		DiagnosticFilter: diagnosticFilter,
	}
}

func (options QuickInfoOptions) WithShowDescriptions(show bool) QuickInfoOptions {
	options.ShowDescriptions = show
	return options
}

func (options QuickInfoOptions) WithShowDiagnostics(show bool) QuickInfoOptions {
	options.ShowDiagnostics = show
	return options
}

func (options QuickInfoOptions) WithDiagnosticFilter(filter *DisabledDiagnostics) QuickInfoOptions {
	if filter == nil {
		options.DiagnosticFilter = DisabledDiagnosticsDefault
	} else {
		options.DiagnosticFilter = *filter
	}
	return options
}
