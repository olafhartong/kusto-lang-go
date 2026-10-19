package kustolang

import "strings"

// TryGetDirectiveClusterAndDatabase extracts the database context encoded by a
// client directive. A nil name means the directive did not specify that part;
// a non-nil pointer may contain an empty parsed value.
func TryGetDirectiveClusterAndDatabase(directive ClientDirective) (clusterName, databaseName *string, ok bool) {
	stringValue := func(value string) *string { return &value }
	nonEmptyValue := func(value string) *string {
		if value == "" {
			return nil
		}
		return stringValue(value)
	}

	switch directive.Name {
	case "database":
		if len(directive.Arguments) > 1 {
			return stringValue(directiveArgumentString(directive.Arguments[0])), stringValue(directiveArgumentString(directive.Arguments[1])), true
		}
		if len(directive.Arguments) == 1 {
			value := directiveArgumentString(directive.Arguments[0])
			host, path := kustoDirectiveHostAndPath(value)
			if host != "" && path != "" {
				return stringValue(host), stringValue(path), true
			}
			if host != "" {
				return nil, stringValue(host), true
			}
		}

	case "connect":
		if len(directive.Arguments) == 0 {
			return nil, nil, false
		}
		argument := directive.Arguments[0]
		if strings.HasPrefix(argument.Text, "cluster") {
			cluster, database, parsed := kustoDirectiveEntityPath(argument.Text)
			if !parsed {
				return nil, nil, false
			}
			host, path := kustoDirectiveHostAndPath(cluster)
			if host != "" {
				cluster = host
			}
			if database == "" {
				database = path
			}
			return nonEmptyValue(cluster), nonEmptyValue(database), cluster != ""
		}
		if !strings.HasPrefix(argument.Text, "@") {
			connection := ParseConnectionInfo(directiveArgumentString(argument))
			host, path := kustoDirectiveHostAndPath(connection.DataSource())
			return nonEmptyValue(host), nonEmptyValue(path), true
		}
	}

	return nil, nil, false
}

func (b *Binder) bindDirective(node *Directive) *SemanticInfo {
	diagnostics := b.applyDirective(node)
	return semanticInfoPtr(NewSemanticInfoWithDiagnostics(diagnostics...))
}

func (b *Binder) applyDirective(node *Directive) []Diagnostic {
	if b == nil || node == nil || node.Token() == nil {
		return nil
	}
	directive, parsed := TryParseClientDirective(NewEditString(node.Token().RawText()))
	if !parsed {
		return nil
	}

	switch directive.Name {
	case "connect", "database":
		return b.applyConnectOrDatabaseDirective(directive, node.Token())
	default:
		if !directiveNameIsKnown(directive.Name) {
			diagnostic := DiagnosticUnknownDirective(directive.Name)
			if location, found := b.syntaxTokenLocations[node.Token()]; found {
				diagnostic = diagnostic.WithByteLocation(location.Start+1, len(directive.Name))
			}
			return []Diagnostic{diagnostic}
		}
	}
	return nil
}

func (b *Binder) applyConnectOrDatabaseDirective(directive ClientDirective, token *SyntaxToken) []Diagnostic {
	clusterName, databaseName, parsed := TryGetDirectiveClusterAndDatabase(directive)
	if !parsed {
		return nil
	}

	var diagnostics []Diagnostic
	if clusterName != nil {
		b.currentCluster = b.initialGlobals.GetCluster(*clusterName)
		if b.currentCluster == nil {
			b.currentCluster = UnknownClusterSymbol
			diagnostic := DiagnosticNameDoesNotReferToAnyKnownCluster(*clusterName).WithSeverity(DiagnosticSeverityError)
			diagnostics = append(diagnostics, b.diagnosticAtSyntaxElement(diagnostic, token))
		}
	}
	if databaseName != nil {
		b.currentDatabase = b.currentCluster.GetDatabase(*databaseName)
		if b.currentDatabase == nil {
			b.currentDatabase = UnknownDatabaseSymbol
			diagnostic := DiagnosticNameDoesNotReferToAnyKnownDatabase(*databaseName)
			diagnostics = append(diagnostics, b.diagnosticAtSyntaxElement(diagnostic, token))
		}
	}
	b.updateCurrentDatabaseContext()
	return diagnostics
}

func directiveNameIsKnown(name string) bool {
	for _, known := range KustoFacts.Directives {
		if name == known {
			return true
		}
	}
	return false
}

// BindDirectives binds only the leading directive list rooted at root and
// returns the number of directives applied. Other query syntax is untouched.
func (b *Binder) BindDirectives(root SyntaxElement) int {
	if b == nil || root == nil {
		return 0
	}
	b.resetCurrentDatabaseContext()
	b.captureSyntaxLocations(root)

	bind := func(directive *Directive) bool {
		if directive == nil {
			return false
		}
		if info := b.bindDirective(directive); info != nil {
			directive.SetSemanticInfo(*info)
		}
		return true
	}

	count := 0
	switch syntax := WrapGeneratedSyntaxElement(root).(type) {
	case *Directive:
		if bind(syntax) {
			count++
		}
	case *QueryBlock:
		for _, directive := range syntax.Directives().Elements() {
			if bind(directive) {
				count++
			}
		}
	case *DirectiveBlock:
		for _, directive := range syntax.Directives().Elements() {
			if bind(directive) {
				count++
			}
		}
	case *CommandBlock:
		for _, directive := range syntax.Directives().Elements() {
			if bind(directive) {
				count++
			}
		}
	default:
		for i := 0; i < root.ChildCount(); i++ {
			directive, isDirective := WrapGeneratedSyntaxElement(root.GetChild(i)).(*Directive)
			if !isDirective {
				break
			}
			if bind(directive) {
				count++
			}
		}
	}
	return count
}
