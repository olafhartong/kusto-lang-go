package kustolang

import (
	"net/url"
	"strings"
)

// Ported from the directive branch of
// src/Kusto.Language/Editor/Kusto/KustoReferenceFinder.cs and
// src/Kusto.Language/Binder/Binder_Misc.cs.

func kustoDirectiveClusterReferences(code *KustoCode) []ClusterReference {
	if code == nil || code.Text == "" {
		return EmptyReadOnlyList[ClusterReference]()
	}
	info := ParseClientDirectiveInfo(NewEditString(code.Text))
	refs := make([]ClusterReference, 0, len(info.Directives))
	for _, directive := range info.Directives {
		cluster, _, ok := kustoDirectiveClusterAndDatabase(directive)
		if !ok || cluster == "" {
			continue
		}
		start, length := kustoDirectivePublicRange(code.Text, directive)
		refs = append(refs, NewClusterReference(cluster, start, length))
	}
	return refs
}

func kustoDirectiveDatabaseReferences(code *KustoCode, currentCluster *ClusterSymbol) []DatabaseReference {
	if code == nil || code.Text == "" {
		return EmptyReadOnlyList[DatabaseReference]()
	}
	info := ParseClientDirectiveInfo(NewEditString(code.Text))
	refs := make([]DatabaseReference, 0, len(info.Directives))
	for i, directive := range info.Directives {
		cluster, database, ok := kustoDirectiveClusterAndDatabase(directive)
		if !ok {
			continue
		}
		if cluster == "" {
			cluster = kustoRelatedDirectiveCluster(info.Directives, i, currentCluster)
		}
		start, length := kustoDirectivePublicRange(code.Text, directive)
		refs = append(refs, NewDatabaseReference(database, cluster, start, length))
	}
	return refs
}

func kustoDirectiveClusterAndDatabase(directive ClientDirective) (cluster string, database string, ok bool) {
	switch directive.Name {
	case "database":
		if len(directive.Arguments) > 1 {
			return directiveArgumentString(directive.Arguments[0]), directiveArgumentString(directive.Arguments[1]), true
		}
		if len(directive.Arguments) == 1 {
			value := directiveArgumentString(directive.Arguments[0])
			host, path := kustoDirectiveHostAndPath(value)
			if host != "" && path != "" {
				return host, path, true
			}
			if host != "" {
				return "", host, true
			}
		}
	case "connect":
		if len(directive.Arguments) == 0 {
			return "", "", false
		}
		argument := directive.Arguments[0]
		if strings.HasPrefix(argument.Text, "cluster") {
			cluster, database, ok = kustoDirectiveEntityPath(argument.Text)
			if ok {
				host, path := kustoDirectiveHostAndPath(cluster)
				if host != "" {
					cluster = host
				}
				if database == "" {
					database = path
				}
				return cluster, database, cluster != ""
			}
		} else if !strings.HasPrefix(argument.Text, "@") {
			connection := ParseConnectionInfo(directiveArgumentString(argument))
			host, path := kustoDirectiveHostAndPath(connection.DataSource())
			return host, path, true
		}
	}
	return "", "", false
}

func kustoDirectiveEntityPath(text string) (cluster string, database string, ok bool) {
	expression := ParseEntityPath(text)
	if expression == nil {
		return "", "", false
	}
	walkSyntaxWithTriviaStartAndAncestors(simpleBinderSyntaxNode(expression), 0, func(element SyntaxElement, _ int, _ []SyntaxElement) {
		call, callOK := WrapGeneratedSyntaxElement(element).(*FunctionCallExpression)
		if !callOK {
			return
		}
		argument := functionCallArgumentAt(call, 0)
		value, valueOK := ExpressionLiteralValue(argument).(string)
		if !valueOK {
			return
		}
		switch NameReferenceSimpleName(call.Name()) {
		case "cluster":
			cluster = value
		case "database":
			database = value
		}
	})
	return cluster, database, cluster != ""
}

func kustoDirectiveHostAndPath(value string) (host string, path string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ""
	}
	parsed, err := url.Parse(value)
	if err == nil && parsed.Hostname() != "" {
		return parsed.Hostname(), strings.Trim(parsed.Path, "/")
	}
	if strings.Contains(value, "/") {
		parsed, err = url.Parse("https://" + strings.TrimPrefix(value, "//"))
		if err == nil && parsed.Hostname() != "" {
			return parsed.Hostname(), strings.Trim(parsed.Path, "/")
		}
	}
	return value, ""
}

func kustoRelatedDirectiveCluster(directives []ClientDirective, index int, currentCluster *ClusterSymbol) string {
	for i := index - 1; i >= 0; i-- {
		if directives[i].Name == "cluster" && len(directives[i].Arguments) > 0 {
			return directiveArgumentString(directives[i].Arguments[0])
		}
	}
	if currentCluster != nil {
		return currentCluster.SymbolName()
	}
	return ""
}

func directiveArgumentString(argument ClientDirectiveArgument) string {
	if value, ok := argument.Value.(string); ok {
		return value
	}
	return ""
}

func kustoDirectivePublicRange(text string, directive ClientDirective) (start int, length int) {
	byteStart := directive.Text.GetOriginalPosition(0)
	byteEnd := byteStart + len(directive.Text.CurrentText)
	start = byteOffsetToUTF16Position(text, byteStart)
	return start, byteOffsetToUTF16Position(text, byteEnd) - start
}
