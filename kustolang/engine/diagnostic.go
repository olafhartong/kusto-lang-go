package kustolang

// Ported from src/Kusto.Language/Diagnostics/Diagnostic.cs in
// microsoft/Kusto-Query-Language. The Go shape is immutable-by-value instead
// of reference implementation reference-immutable, but the defaults and With* behavior match the
// upstream public API.

const (
	DiagnosticCategoryGeneral     = "General"
	DiagnosticCategoryCorrectness = "Correctness"
	DiagnosticCategoryPerformance = "Performance"
)

const (
	DiagnosticSeverityError       = "Error"
	DiagnosticSeverityWarning     = "Warning"
	DiagnosticSeveritySuggestion  = "Suggestion"
	DiagnosticSeverityInformation = "Information"
	DiagnosticSeverityHidden      = "Hidden"
)

type DiagnosticLocationKind string

const (
	DiagnosticLocationRelative    DiagnosticLocationKind = "Relative"
	DiagnosticLocationAbsolute    DiagnosticLocationKind = "Absolute"
	DiagnosticLocationRelativeEnd DiagnosticLocationKind = "RelativeEnd"
)

type Diagnostic struct {
	Code         string                 `json:"code"`
	Category     string                 `json:"category,omitempty"`
	Severity     string                 `json:"severity,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Message      string                 `json:"message"`
	LocationKind DiagnosticLocationKind `json:"location_kind,omitempty"`
	Start        int                    `json:"start"`
	Length       int                    `json:"length"`
	locationText string
}

type diagnosticIdentity struct {
	Code    string
	Message string
	Start   int
	Length  int
}

func diagnosticIdentityForEquality(diagnostic Diagnostic) (diagnosticIdentity, bool) {
	if diagnostic.LocationKind != DiagnosticLocationAbsolute {
		return diagnosticIdentity{}, false
	}
	return diagnosticIdentity{
		Code:    diagnostic.Code,
		Message: diagnostic.Message,
		Start:   diagnostic.Start,
		Length:  diagnostic.Length,
	}, true
}

func NewDiagnostic(code, message string) Diagnostic {
	return newDiagnostic(code, DiagnosticCategoryGeneral, DiagnosticSeverityError, message, message)
}

func NewDiagnosticWithDescription(code, category, severity, description string) Diagnostic {
	if category == "" {
		category = DiagnosticCategoryGeneral
	}
	if severity == "" {
		severity = DiagnosticSeverityError
	}
	return newDiagnostic(code, category, severity, description, description)
}

func NewDiagnosticDetailed(code, category, severity, description, message string) Diagnostic {
	if category == "" {
		category = DiagnosticCategoryGeneral
	}
	if severity == "" {
		severity = DiagnosticSeverityError
	}
	if description == "" {
		description = message
	}
	if message == "" {
		message = description
	}
	return newDiagnostic(code, category, severity, description, message)
}

func newDiagnostic(code, category, severity, description, message string) Diagnostic {
	return Diagnostic{
		Code:         code,
		Category:     category,
		Severity:     severity,
		Description:  description,
		Message:      message,
		LocationKind: DiagnosticLocationRelative,
	}
}

func (d Diagnostic) HasLocation() bool {
	return d.LocationKind == DiagnosticLocationAbsolute
}

func (d Diagnostic) End() int {
	return d.Start + d.Length
}

func (d Diagnostic) WithCode(code string) Diagnostic {
	d.Code = code
	return d
}

func (d Diagnostic) WithCategory(category string) Diagnostic {
	d.Category = category
	return d
}

func (d Diagnostic) WithSeverity(severity string) Diagnostic {
	d.Severity = severity
	return d
}

func (d Diagnostic) WithDescription(description string) Diagnostic {
	d.Description = description
	return d
}

func (d Diagnostic) WithMessage(message string) Diagnostic {
	d.Message = message
	return d
}

func (d Diagnostic) WithLocation(start, length int) Diagnostic {
	return d.withLocation(start, length)
}

// WithByteLocation records an internal byte-coordinate location. Native parser,
// binder, and analyzer code uses this until the public UTF-16 boundary.
func (d Diagnostic) WithByteLocation(start, length int) Diagnostic {
	return d.withLocation(start, length)
}

func (d Diagnostic) withLocation(start, length int) Diagnostic {
	if start < 0 {
		start = d.Start
	}
	if length < 0 {
		length = d.Length
	}
	d.LocationKind = DiagnosticLocationAbsolute
	d.Start = start
	d.Length = length
	return d
}

func (d Diagnostic) WithElementLocation(location SyntaxElement) Diagnostic {
	if location == nil {
		return d
	}
	return d.WithLocation(location.UTF16TextStart(), location.UTF16Width())
}

func (d Diagnostic) WithLocationKind(kind DiagnosticLocationKind) Diagnostic {
	d.LocationKind = kind
	return d
}

// DiagnosticsToUTF16 converts byte-based absolute diagnostics. Callers that
// combine relative and already-absolute public syntax diagnostics use the
// coordinate-aware syntax collector instead.
func DiagnosticsToUTF16(text string, diagnostics []Diagnostic) []Diagnostic {
	if len(diagnostics) == 0 {
		return []Diagnostic{}
	}
	converted := make([]Diagnostic, len(diagnostics))
	copy(converted, diagnostics)
	for i := range converted {
		diagnostic := converted[i]
		if diagnostic.LocationKind != DiagnosticLocationAbsolute {
			continue
		}
		start := byteOffsetToUTF16Position(text, diagnostic.Start)
		end := byteOffsetToUTF16Position(text, diagnostic.End())
		converted[i] = diagnostic.withLocation(start, end-start)
	}
	return converted
}

func (d Diagnostic) WithLocationText(text string) Diagnostic {
	d.locationText = text
	return d
}

func (d Diagnostic) LocationText() string {
	return d.locationText
}

func (d Diagnostic) Equals(other Diagnostic) bool {
	left, leftOK := diagnosticIdentityForEquality(d)
	right, rightOK := diagnosticIdentityForEquality(other)
	return leftOK && rightOK && left == right
}
