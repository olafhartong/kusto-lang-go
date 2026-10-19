package kustolang

import "strings"

// Ported from:
//   src/Kusto.Language/Editor/Kusto/KustoQualifier.cs
//   src/Kusto.Language/Editor/Kusto/Actors/ActorUtilities.cs
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

// KustoQualifierFullyQualify returns query text with database entities fully
// qualified against globals.
func KustoQualifierFullyQualify(query string, globals *GlobalState) string {
	return KustoQualifierRequalify(query, globals, nil, nil)
}

// KustoQualifierMinimallyQualify returns query text minimally qualified for
// the current cluster and database.
func KustoQualifierMinimallyQualify(query string, globals *GlobalState) string {
	if globals == nil {
		globals = DefaultGlobalState
	}
	return KustoQualifierRequalify(query, globals, globals.Cluster, globals.Database)
}

// KustoQualifierRequalify minimally qualifies database entity expressions for
// a new default cluster and database.
func KustoQualifierRequalify(query string, globals *GlobalState, cluster *ClusterSymbol, database *DatabaseSymbol) string {
	if CodeKindOf(query) != CodeKindQuery {
		return query
	}
	code := ParseAndAnalyze(query, globals)
	if code == nil || code.Syntax() == nil || !code.HasSemantics() {
		return query
	}
	return kustoQualifierRequalifyEdit(NewEditString(query), code.Syntax(), code.Globals, cluster, database).CurrentText
}

func kustoQualifierRequalifyEdit(text EditString, root SyntaxElement, globals *GlobalState, cluster *ClusterSymbol, database *DatabaseSymbol) EditString {
	if root == nil || globals == nil {
		return text
	}
	edits := make([]TextEdit, 0)
	for _, candidate := range GetDescendants[Expression](root) {
		entity := actorGetEntityExpression(candidate, globals)
		if entity == nil || !extractSameElement(entity, candidate) || kustoQualifierEntityGroupPart(entity) != nil {
			continue
		}

		clusterPart := kustoQualifierClusterPart(entity)
		databasePart := kustoQualifierDatabasePart(entity)
		entityPart := kustoQualifierEntityPart(entity, globals)
		entityPartStart := entity.End()
		if entityPart != nil {
			entityPartStart = entityPart.TextStart()
		}

		entitySymbol := kustoQualifierEntityReference(entityPart)
		entityDatabase := kustoQualifierDatabaseResult(databasePart)
		if databasePart == nil {
			if entitySymbol == nil {
				entityDatabase = globals.Database
			} else {
				entityDatabase = globals.GetDatabase(entitySymbol)
			}
		}
		if entityDatabase != nil && entityDatabase.SymbolName() == "" {
			entityDatabase = nil
		}

		entityCluster := kustoQualifierClusterResult(clusterPart)
		if clusterPart == nil {
			if entityDatabase == nil {
				if databasePart != nil {
					entityCluster = globals.Cluster
				}
			} else {
				entityCluster = globals.GetClusterForDatabase(entityDatabase)
			}
		}
		if entityCluster != nil && entityCluster.SymbolName() == "" {
			entityCluster = nil
		}

		switch {
		case entityCluster == cluster && entityDatabase == database && cluster != nil && database != nil && entityPart != nil:
			edits = append(edits, NewTextEdit(entity.TextStart(), entityPartStart-entity.TextStart(), ""))
		case entityCluster == cluster && cluster != nil:
			dbQualifier := kustoQualifierDatabaseText(databasePart, entityDatabase)
			entityText := ""
			if entityPart != nil {
				entityText = entityPart.Text(IncludeTriviaInterior)
			}
			edits = append(edits, NewTextEdit(entity.TextStart(), entity.Width(), kustoQualifierJoinPath(dbQualifier, entityText)))
		default:
			dbQualifier := kustoQualifierDatabaseText(databasePart, entityDatabase)
			clusterQualifier := kustoQualifierClusterText(clusterPart, entityCluster)
			entityText := ""
			if entityPart != nil {
				entityText = entityPart.Text(IncludeTriviaInterior)
			}
			edits = append(edits, NewTextEdit(entity.TextStart(), entity.Width(), kustoQualifierJoinPath(clusterQualifier, dbQualifier, entityText)))
		}
	}
	return text.ApplyAll(edits)
}

func kustoQualifierExpressionPart(entity Expression, predicate func(Expression) bool) Expression {
	if entity == nil {
		return nil
	}
	if predicate(entity) {
		return entity
	}
	for {
		path, ok := entity.(*PathExpression)
		if !ok {
			return nil
		}
		if predicate(path.Expression()) {
			return path.Expression()
		}
		if predicate(path.Selector()) {
			return path.Selector()
		}
		entity = path.Expression()
	}
}

func kustoQualifierClusterPart(entity Expression) Expression {
	return kustoQualifierExpressionPart(entity, actorIsClusterPart)
}

func kustoQualifierDatabasePart(entity Expression) Expression {
	return kustoQualifierExpressionPart(entity, actorIsDatabasePart)
}

func kustoQualifierEntityPart(entity Expression, globals *GlobalState) Expression {
	return kustoQualifierExpressionPart(entity, func(expression Expression) bool {
		return actorIsEntityPart(expression, globals)
	})
}

func kustoQualifierEntityGroupPart(entity Expression) Expression {
	return kustoQualifierExpressionPart(entity, actorIsEntityGroupElementPart)
}

func kustoQualifierEntityReference(part Expression) Symbol {
	if call, ok := part.(*FunctionCallExpression); ok {
		function, _ := call.ReferencedSymbol().(*FunctionSymbol)
		if !actorIsEntityFunction(function) {
			return nil
		}
		if result, found := call.Payload(functionCallResultSymbolPayload); found {
			if symbol, symbolOK := result.(Symbol); symbolOK {
				return symbol
			}
		}
	}
	if reference, ok := part.(*NameReference); ok {
		if _, isCallName := WrapGeneratedSyntaxElement(reference.Parent()).(*FunctionCallExpression); !isCallName {
			return reference.ReferencedSymbol()
		}
	}
	return nil
}

func kustoQualifierDatabaseResult(part Expression) *DatabaseSymbol {
	if call, ok := part.(*FunctionCallExpression); ok {
		if result, found := call.Payload(functionCallResultSymbolPayload); found {
			database, _ := result.(*DatabaseSymbol)
			return database
		}
	}
	return nil
}

func kustoQualifierClusterResult(part Expression) *ClusterSymbol {
	if call, ok := part.(*FunctionCallExpression); ok {
		if result, found := call.Payload(functionCallResultSymbolPayload); found {
			cluster, _ := result.(*ClusterSymbol)
			return cluster
		}
	}
	return nil
}

func kustoQualifierDatabaseText(part Expression, database *DatabaseSymbol) string {
	if part != nil {
		return part.Text(IncludeTriviaInterior)
	}
	if database == nil {
		return ""
	}
	name := database.AlternateName()
	if name == "" {
		name = database.SymbolName()
	}
	return "database(" + GetSingleQuotedStringLiteral(name) + ")"
}

func kustoQualifierClusterText(part Expression, cluster *ClusterSymbol) string {
	if part != nil {
		return part.Text(IncludeTriviaInterior)
	}
	if cluster == nil {
		return ""
	}
	return "cluster(" + GetSingleQuotedStringLiteral(cluster.SymbolName()) + ")"
}

func kustoQualifierJoinPath(elements ...string) string {
	parts := make([]string, 0, len(elements))
	for _, element := range elements {
		if strings.TrimSpace(element) != "" {
			parts = append(parts, element)
		}
	}
	return strings.Join(parts, ".")
}
